// capture_archives_on_workers is an application that captures archives on all CT
// workers and uploads it to Google Storage. The requester is emailed when the task
// is done.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.skia.org/infra/ct/go/ctfe/admin_tasks"
	"go.skia.org/infra/ct/go/ctfe/task_common"
	"go.skia.org/infra/ct/go/master_scripts/master_common"
	"go.skia.org/infra/ct/go/util"
	"go.skia.org/infra/go/sklog"
	skutil "go.skia.org/infra/go/util"
)

var (
	emails      = flag.String("emails", "", "The comma separated email addresses to notify when the task is picked up and completes.")
	taskID      = flag.Int64("task_id", -1, "The key of the CT task in CTFE. The task will be updated when it is started and also when it completes.")
	pagesetType = flag.String("pageset_type", "", "The type of pagesets to use. Eg: 10k, Mobile10k, All.")
	runOnGCE    = flag.Bool("run_on_gce", true, "Run on Linux GCE instances.")
	runID       = flag.String("run_id", "", "The unique run id (typically requester + timestamp).")

	taskCompletedSuccessfully = new(bool)
)

const (
	MAX_PAGES_PER_SWARMING_BOT = 100
)

func sendEmail(recipients []string) {
	// Send completion email.
	emailSubject := "Capture archives Cluster telemetry task has completed"
	failureHtml := ""
	if !*taskCompletedSuccessfully {
		emailSubject += " with failures"
		failureHtml = util.GetFailureEmailHtml(*runID)
	}
	bodyTemplate := `
	The Cluster telemetry queued task to capture archives of %s pagesets has completed. %s.<br/>
	%s
	You can schedule more runs <a href="%s">here</a>.<br/><br/>
	Thanks!
	`
	emailBody := fmt.Sprintf(bodyTemplate, *pagesetType, util.GetSwarmingLogsLink(*runID), failureHtml, master_common.AdminTasksWebapp)
	if err := util.SendEmail(recipients, emailSubject, emailBody); err != nil {
		sklog.Errorf("Error while sending email: %s", err)
		return
	}
}

func updateTaskInDatastore(ctx context.Context) {
	vars := admin_tasks.RecreateWebpageArchivesUpdateVars{}
	vars.Id = *taskID
	vars.SetCompleted(*taskCompletedSuccessfully)
	skutil.LogErr(task_common.FindAndUpdateTask(ctx, &vars))
}

func captureArchivesOnWorkers() error {
	master_common.Init("capture_archives")

	ctx := context.Background()

	// Send start email.
	emailsArr := util.ParseEmails(*emails)
	emailsArr = append(emailsArr, util.CtAdmins...)
	if len(emailsArr) == 0 {
		return errors.New("At least one email address must be specified")
	}
	skutil.LogErr(task_common.UpdateTaskSetStarted(ctx, &admin_tasks.RecreateWebpageArchivesUpdateVars{}, *taskID, *runID))
	skutil.LogErr(util.SendTaskStartEmail(*taskID, emailsArr, "Capture archives", *runID, "", ""))
	// Ensure task is updated and completion email is sent even if task fails.
	defer updateTaskInDatastore(ctx)
	defer sendEmail(emailsArr)
	// Finish with glog flush and how long the task took.
	defer util.TimeTrack(time.Now(), "Capture archives on Workers")
	defer sklog.Flush()

	if *pagesetType == "" {
		return errors.New("Must specify --pageset_type")
	}

	// Empty the remote dir before the workers upload to it.
	gs, err := util.NewGcsUtil(nil)
	if err != nil {
		return err
	}
	gsBaseDir := filepath.Join(util.SWARMING_DIR_NAME, util.WEB_ARCHIVES_DIR_NAME, *pagesetType)
	skutil.LogErr(gs.DeleteRemoteDir(gsBaseDir))

	// Find which chromium hash the workers should use.
	chromiumHash, err := util.GetChromiumHash(ctx)
	if err != nil {
		return fmt.Errorf("Could not find the latest chromium hash: %s", err)
	}

	// Trigger task to return hash of telemetry isolates.
	telemetryHash, err := util.TriggerIsolateTelemetrySwarmingTask(ctx, "isolate_telemetry", *runID, chromiumHash, "", util.PLATFORM_LINUX, []string{}, 1*time.Hour, 1*time.Hour, *master_common.Local)
	if err != nil {
		return fmt.Errorf("Error encountered when swarming isolate telemetry task: %s", err)
	}
	if telemetryHash == "" {
		return errors.New("Found empty telemetry hash!")
	}
	isolateDeps := []string{telemetryHash}

	// Archive, trigger and collect swarming tasks.
	if _, err := util.TriggerSwarmingTask(ctx, *pagesetType, "capture_archives", util.CAPTURE_ARCHIVES_ISOLATE, *runID, "", util.PLATFORM_LINUX, 4*time.Hour, 1*time.Hour, util.TASKS_PRIORITY_LOW, MAX_PAGES_PER_SWARMING_BOT, util.PagesetTypeToInfo[*pagesetType].NumPages, map[string]string{}, *runOnGCE, *master_common.Local, 1, isolateDeps); err != nil {
		return fmt.Errorf("Error encountered when swarming tasks: %s", err)
	}

	*taskCompletedSuccessfully = true
	return nil
}

func main() {
	retCode := 0
	if err := captureArchivesOnWorkers(); err != nil {
		sklog.Errorf("Error while running capture archives on workers: %s", err)
		retCode = 255
	}
	os.Exit(retCode)
}
