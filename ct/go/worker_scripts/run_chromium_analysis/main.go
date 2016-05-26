// run_chromium_analysis is an application that runs the specified benchmark over
// CT's webpage archives. It is intended to be run on swarming bots.
package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"github.com/skia-dev/glog"

	"strings"

	"go.skia.org/infra/ct/go/util"
	"go.skia.org/infra/ct/go/worker_scripts/worker_common"
	"go.skia.org/infra/go/common"
	skutil "go.skia.org/infra/go/util"
)

const (
	// The number of goroutines that will run in parallel to run benchmarks.
	WORKER_POOL_SIZE = 10
)

var (
	startRange         = flag.Int("start_range", 1, "The number this worker will run benchmarks from.")
	num                = flag.Int("num", 100, "The total number of benchmarks to run starting from the start_range.")
	pagesetType        = flag.String("pageset_type", util.PAGESET_TYPE_MOBILE_10k, "The type of pagesets to analyze. Eg: 10k, Mobile10k, All.")
	chromiumBuild      = flag.String("chromium_build", "", "The chromium build to use.")
	runID              = flag.String("run_id", "", "The unique run id (typically requester + timestamp).")
	benchmarkName      = flag.String("benchmark_name", "", "The telemetry benchmark to run on this worker.")
	benchmarkExtraArgs = flag.String("benchmark_extra_args", "", "The extra arguments that are passed to the specified benchmark.")
	browserExtraArgs   = flag.String("browser_extra_args", "", "The extra arguments that are passed to the browser while running the benchmark.")
	chromeCleanerTimer = flag.Duration("cleaner_timer", 15*time.Minute, "How often all chrome processes will be killed on this slave.")
)

func main() {
	defer common.LogPanic()
	worker_common.Init()
	defer util.TimeTrack(time.Now(), "Running Chromium Analysis")
	defer glog.Flush()

	// Validate required arguments.
	if *chromiumBuild == "" {
		glog.Fatal("Must specify --chromium_build")
	}
	if *runID == "" {
		glog.Fatal("Must specify --run_id")
	}
	if *benchmarkName == "" {
		glog.Fatal("Must specify --benchmark_name")
	}

	// Reset the local chromium checkout.
	if err := util.ResetCheckout(util.ChromiumSrcDir); err != nil {
		glog.Fatalf("Could not reset %s: %s", util.ChromiumSrcDir, err)
	}
	// Sync the local chromium checkout.
	if err := util.SyncDir(util.ChromiumSrcDir); err != nil {
		glog.Fatalf("Could not gclient sync %s: %s", util.ChromiumSrcDir, err)
	}

	// Instantiate GsUtil object.
	gs, err := util.NewGsUtil(nil)
	if err != nil {
		glog.Fatal(err)
	}

	// Download the benchmark patch for this run from Google storage.
	benchmarkPatchName := *runID + ".benchmark.patch"
	tmpDir, err := ioutil.TempDir("", "patches")
	if err != nil {
		glog.Fatalf("Could not create a temp dir: %s", err)
	}
	defer skutil.RemoveAll(tmpDir)
	benchmarkPatchLocalPath := filepath.Join(tmpDir, benchmarkPatchName)
	remotePatchesDir := filepath.Join(util.ChromiumAnalysisRunsDir, *runID)
	benchmarkPatchRemotePath := filepath.Join(remotePatchesDir, benchmarkPatchName)
	respBody, err := gs.GetRemoteFileContents(benchmarkPatchRemotePath)
	if err != nil {
		glog.Fatalf("Could not fetch %s: %s", benchmarkPatchRemotePath, err)
	}
	defer skutil.Close(respBody)
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(respBody); err != nil {
		glog.Fatalf("Could not read from %s: %s", benchmarkPatchRemotePath, err)
	}
	if err := ioutil.WriteFile(benchmarkPatchLocalPath, buf.Bytes(), 0666); err != nil {
		glog.Fatalf("Unable to create file %s: %s", benchmarkPatchLocalPath, err)
	}
	// Apply benchmark patch to the local chromium checkout.
	if buf.Len() > 10 {
		if err := util.ApplyPatch(benchmarkPatchLocalPath, util.ChromiumSrcDir); err != nil {
			glog.Fatalf("Could not apply Telemetry's patch in %s: %s", util.ChromiumSrcDir, err)
		}
	}

	// Download the specified chromium build.
	if err := gs.DownloadChromiumBuild(*chromiumBuild); err != nil {
		glog.Fatal(err)
	}
	//Delete the chromium build to save space when we are done.
	defer skutil.RemoveAll(filepath.Join(util.ChromiumBuildsDir, *chromiumBuild))

	chromiumBinary := filepath.Join(util.ChromiumBuildsDir, *chromiumBuild, util.BINARY_CHROME)

	// Download pagesets if they do not exist locally.
	pathToPagesets := filepath.Join(util.PagesetsDir, *pagesetType)
	if _, err := gs.DownloadSwarmingArtifacts(pathToPagesets, util.PAGESETS_DIR_NAME, *pagesetType, *startRange, *num); err != nil {
		glog.Fatal(err)
	}
	defer skutil.RemoveAll(pathToPagesets)

	// Download archives if they do not exist locally.
	pathToArchives := filepath.Join(util.WebArchivesDir, *pagesetType)
	if _, err := gs.DownloadSwarmingArtifacts(pathToArchives, util.WEB_ARCHIVES_DIR_NAME, *pagesetType, *startRange, *num); err != nil {
		glog.Fatal(err)
	}
	defer skutil.RemoveAll(pathToArchives)

	// Establish nopatch output paths.
	localOutputDir := filepath.Join(util.StorageDir, util.BenchmarkRunsDir, *runID)
	skutil.RemoveAll(localOutputDir)
	skutil.MkdirAll(localOutputDir, 0700)
	defer skutil.RemoveAll(localOutputDir)
	remoteDir := filepath.Join(util.BenchmarkRunsDir, *runID)

	// Construct path to CT's python scripts.
	pathToPyFiles := util.GetPathToPyFiles(!*worker_common.Local)

	fileInfos, err := ioutil.ReadDir(pathToPagesets)
	if err != nil {
		glog.Fatalf("Unable to read the pagesets dir %s: %s", pathToPagesets, err)
	}

	glog.Infoln("===== Going to run the task with parallel chrome processes =====")

	// Create channel that contains all pageset file names. This channel will
	// be consumed by the worker pool.
	pagesetRequests := util.GetClosedChannelOfPagesets(fileInfos)

	var wg sync.WaitGroup
	// Use a RWMutex for the chromeProcessesCleaner goroutine to communicate to
	// the workers (acting as "readers") when it wants to be the "writer" and
	// kill all zombie chrome processes.
	var mutex sync.RWMutex

	// Loop through workers in the worker pool.
	for i := 0; i < WORKER_POOL_SIZE; i++ {
		// Increment the WaitGroup counter.
		wg.Add(1)

		// Create and run a goroutine closure that runs the benchmark.
		go func() {
			// Decrement the WaitGroup counter when the goroutine completes.
			defer wg.Done()

			for pagesetName := range pagesetRequests {

				mutex.RLock()
				if err := util.RunBenchmark(pagesetName, pathToPagesets, pathToPyFiles, localOutputDir, *chromiumBuild, chromiumBinary, *runID, *browserExtraArgs, *benchmarkName, "Linux", *benchmarkExtraArgs, *pagesetType, -1); err != nil {
					glog.Errorf("Error while running withpatch benchmark: %s", err)
					continue
				}
				mutex.RUnlock()
			}
		}()
	}

	if !*worker_common.Local {
		// Start the cleaner.
		go util.ChromeProcessesCleaner(&mutex, *chromeCleanerTimer)
	}

	// Wait for all spawned goroutines to complete.
	wg.Wait()

	// If "--output-format=csv-pivot-table" was specified then merge all CSV files and upload.
	if strings.Contains(*benchmarkExtraArgs, "--output-format=csv-pivot-table") {
		if err := util.MergeUploadCSVFilesOnWorkers(localOutputDir, pathToPyFiles, *runID, remoteDir, gs, *startRange); err != nil {
			glog.Fatalf("Error while processing withpatch CSV files: %s", err)
		}
	}
}
