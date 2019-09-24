package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.skia.org/infra/gold-client/go/goldclient"
	"go.skia.org/infra/golden/go/jsonio"
	"go.skia.org/infra/golden/go/shared"
	"go.skia.org/infra/golden/go/types"
)

// imgTest is the state for the imgtest command and its sub-commands.
// Specifically, it houses the flags.
type imgTest struct {
	// common flags
	commitHash   string
	corpus       string
	failureFile  string
	instanceID   string
	changeListID string
	tryJobID     string
	keysFile     string
	passFailStep bool
	patchSetID   string
	uploadOnly   bool
	urlOverride  string
	workDir      string

	testName string
	pngFile  string
	// a file to a json dictionary of key pairs that will be added to this test
	// after read into a map[string]string
	testKeysFile string
	// a slice of strings like foo:bar that will be split on the first ':' into
	// key value pairs that will go into a map[string]string
	testKeysStrings []string
}

// getImgTestCmd returns the definition of the imgtest command.
func getImgTestCmd() *cobra.Command {
	env := &imgTest{}

	// imgtest command and its sub commands
	imgTestCmd := &cobra.Command{
		Use:   "imgtest",
		Short: "Collect  and upload test results as images",
		Long: `
Collect and upload test results to the Gold backend.`,
	}

	// cmd: imgtest init
	imgTestInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a  testing environment",
		Long: `
Start a testing session during which tests are added. This initializes the environment.
It gathers whether the 'add' command returns a pass/fail value and the common
keys shared by all tests that are added via 'add'.
`,
		PreRunE: env.validate,
		Run:     env.runImgTestInitCmd,
	}
	env.addCommonFlags(imgTestInitCmd, false)

	imgTestAddCmd := &cobra.Command{
		Use:   "add",
		Short: "Adds a test image to the results.",
		Long: `
Add images generated by the tests to the test results. This requires two arguments:
			 - The test name
			 - The path to the resulting PNG.
`,
		PreRunE: env.validate,
		Run:     env.runImgTestAddCmd,
		Args:    cobra.NoArgs,
	}
	env.addCommonFlags(imgTestAddCmd, true)
	imgTestAddCmd.Flags().StringVar(&env.testName, "test-name", "", "Unique name of the test, must not contain spaces.")
	imgTestAddCmd.Flags().StringVar(&env.pngFile, "png-file", "", "Path to the PNG file that contains the test results.")
	imgTestAddCmd.Flags().StringVar(&env.testKeysFile, "add-test-key-file", "", "A JSON file containing keys and values that should be applied to this test only.")
	imgTestAddCmd.Flags().StringSliceVar(&env.testKeysStrings, "add-test-key", []string{}, "Any amount of key:value paris that will be added to this test only.")

	Must(imgTestAddCmd.MarkFlagRequired("test-name"))
	Must(imgTestAddCmd.MarkFlagRequired("png-file"))

	imgTestFinalizeCmd := &cobra.Command{
		Use:   "finalize",
		Short: "Finish adding tests and process results.",
		Long: `
All tests have been added. Upload images and generate and upload the JSON file that captures
test results.`,
		Run: env.runImgTestFinalizeCmd,
	}
	imgTestFinalizeCmd.Flags().StringVar(&env.workDir, fstrWorkDir, "", "Work directory for intermediate results")
	Must(imgTestFinalizeCmd.MarkFlagRequired(fstrWorkDir))

	imgTestCheckCmd := &cobra.Command{
		Use:   "check",
		Short: "Checks whether the results match expectations",
		Long: `Check against Gold's baseline whether the results match the expectations.
Does not upload anything nor queue anything for upload.`,
		Run: env.runImgTestCheckCmd,
	}
	imgTestCheckCmd.Flags().StringVar(&env.workDir, fstrWorkDir, "", "Work directory for intermediate results")
	imgTestCheckCmd.Flags().StringVar(&env.testName, "test-name", "", "Unique name of the test, must not contain spaces.")
	imgTestCheckCmd.Flags().StringVar(&env.pngFile, "png-file", "", "Path to the PNG file that contains the test results.")
	imgTestCheckCmd.Flags().StringVar(&env.instanceID, "instance", "", "ID of the Gold instance.")

	imgTestCheckCmd.Flags().StringVar(&env.changeListID, "changelist", "", "If provided, the ChangeListExpectations matching this will apply.")
	imgTestCheckCmd.Flags().StringVar(&env.urlOverride, "url", "", "URL of the Gold instance. Used for testing, if empty the URL will be derived from the value of 'instance'")

	Must(imgTestCheckCmd.MarkFlagRequired(fstrWorkDir))
	Must(imgTestCheckCmd.MarkFlagRequired("test-name"))
	Must(imgTestCheckCmd.MarkFlagRequired("png-file"))
	Must(imgTestCheckCmd.MarkFlagRequired("instance"))

	// assemble the imgtest command.
	imgTestCmd.AddCommand(
		imgTestInitCmd,
		imgTestAddCmd,
		imgTestFinalizeCmd,
		imgTestCheckCmd,
	)
	return imgTestCmd
}

func (i *imgTest) addCommonFlags(cmd *cobra.Command, optional bool) {
	cmd.Flags().StringVar(&i.instanceID, "instance", "", "ID of the Gold instance.")
	cmd.Flags().StringVar(&i.workDir, fstrWorkDir, "", "Work directory for intermediate results")
	cmd.Flags().BoolVar(&i.passFailStep, "passfail", false, "Whether the 'add' call returns a pass/fail for each test.")
	cmd.Flags().BoolVar(&i.uploadOnly, "upload-only", false, "Skip reading expectations from the server. Incompatible with passfail=true.")

	cmd.Flags().StringVar(&i.commitHash, "commit", "", "Git commit hash")
	cmd.Flags().StringVar(&i.corpus, "corpus", "", "Gold Corpus Name. Overrides any other values (e.g. from keys-file or add-test-key)")
	cmd.Flags().StringVar(&i.failureFile, "failure-file", "", "Path to the file where to write failure information")
	cmd.Flags().StringVar(&i.changeListID, "changelist", "", "ChangeList ID if this is run as a TryJob.")
	cmd.Flags().StringVar(&i.tryJobID, "jobid", "", "TryJob ID if this is a TryJob run.")
	cmd.Flags().StringVar(&i.keysFile, "keys-file", "", "JSON file containing key/value pairs commmon to all tests")
	cmd.Flags().StringVar(&i.patchSetID, "patchset", "", "Gerrit patchset number if this is a trybot run. ")
	cmd.Flags().StringVar(&i.urlOverride, "url", "", "URL of the Gold instance. Used for testing, if empty the URL will be derived from the value of 'instance'")

	cmd.Flags().StringVar(&i.changeListID, "issue", "", "[deprecated] Gerrit issue if this is trybot run. ")
	Must(cmd.MarkFlagRequired(fstrWorkDir))
	if !optional {
		Must(cmd.MarkFlagRequired("instance"))
		Must(cmd.MarkFlagRequired("commit"))
		Must(cmd.MarkFlagRequired("keys-file"))
	}
}

func (i *imgTest) validate(cmd *cobra.Command, args []string) error {
	if i.uploadOnly && i.passFailStep {
		return errors.New("Cannot have --upload-only and --passfail both be true.")
	}
	if i.testKeysFile != "" && len(i.testKeysStrings) > 0 {
		return errors.New("Cannot have both --add-test-key and --add-test-key-file.")
	}
	return nil
}

func (i *imgTest) runImgTestCheckCmd(cmd *cobra.Command, args []string) {
	auth, err := goldclient.LoadAuthOpt(i.workDir)
	ifErrLogExit(cmd, err)

	if auth == nil {
		logErrf(cmd, "Auth is empty - did you call goldctl auth first?")
		exitProcess(cmd, 1)
	}

	goldClient, err := goldclient.LoadCloudClient(auth, i.workDir)
	if err != nil {
		fmt.Printf("Could not load existing run, trying to initialize %s\n", i.workDir)
		config := goldclient.GoldClientConfig{
			WorkDir:         i.workDir,
			InstanceID:      i.instanceID,
			OverrideGoldURL: i.urlOverride,
		}
		goldClient, err = goldclient.NewCloudClient(auth, config)
		ifErrLogExit(cmd, err)

		if i.changeListID != "" {
			gr := jsonio.GoldResults{
				ChangeListID: i.changeListID,
				GitHash:      "HEAD",
			}
			err = goldClient.SetSharedConfig(gr) // this will load the baseline
			ifErrLogExit(cmd, err)
		}
	}

	pass, err := goldClient.Check(types.TestName(i.testName), i.pngFile)
	ifErrLogExit(cmd, err)

	if !pass {
		logErrf(cmd, "Test: %s FAIL\n", i.testName)
		exitProcess(cmd, 1)
	}
	logInfof(cmd, "Test: %s PASS\n", i.testName)
	exitProcess(cmd, 0)
}

func (i *imgTest) runImgTestInitCmd(cmd *cobra.Command, args []string) {
	auth, err := goldclient.LoadAuthOpt(i.workDir)
	ifErrLogExit(cmd, err)

	if auth == nil {
		logErrf(cmd, "Auth is empty - did you call goldctl auth first?")
		exitProcess(cmd, 1)
	}

	auth.SetDryRun(flagDryRun)

	keyMap, err := readKeysFile(i.keysFile)
	ifErrLogExit(cmd, err)

	if i.corpus != "" {
		keyMap[types.CORPUS_FIELD] = i.corpus
	}

	validation := shared.Validation{}
	issueID := validation.Int64Value("issue", i.changeListID, types.MasterBranch)
	patchsetID := validation.Int64Value("patchset", i.patchSetID, 0)
	jobID := validation.Int64Value("jobid", i.tryJobID, 0)
	ifErrLogExit(cmd, validation.Errors())

	config := goldclient.GoldClientConfig{
		FailureFile:     i.failureFile,
		InstanceID:      i.instanceID,
		OverrideGoldURL: i.urlOverride,
		PassFailStep:    i.passFailStep,
		UploadOnly:      i.uploadOnly,
		WorkDir:         i.workDir,
	}
	goldClient, err := goldclient.NewCloudClient(auth, config)
	ifErrLogExit(cmd, err)

	// Define the meta data of the result that is shared by all tests.
	// TODO(kjlubick): make the CodeReviewSystem (e.g. gerrit) and
	// ContinuousIntegrationSystem (e.g. buildbucket) configurable
	// via command line args.
	// See https://bugs.chromium.org/p/skia/issues/detail?id=9340
	gr := jsonio.GoldResults{
		GitHash:            i.commitHash,
		Key:                keyMap,
		GerritChangeListID: issueID,
		GerritPatchSet:     patchsetID,
		BuildBucketID:      jobID,
	}

	logVerbose(cmd, "Loading hashes and baseline from Gold\n")
	err = goldClient.SetSharedConfig(gr)
	ifErrLogExit(cmd, err)

	logInfof(cmd, "Directory %s successfully loaded with configuration\n", i.workDir)
}

// runImgTestCommand processes and uploads test results to Gold.
func (i *imgTest) runImgTestAddCmd(cmd *cobra.Command, args []string) {
	auth, err := goldclient.LoadAuthOpt(i.workDir)
	ifErrLogExit(cmd, err)

	if auth == nil {
		logErrf(cmd, "Auth is empty - did you call goldctl auth first?")
		exitProcess(cmd, 1)
	}

	auth.SetDryRun(flagDryRun)

	var goldClient goldclient.GoldClient

	if i.keysFile != "" {
		// user has specified a full set of keys. This happens if they
		// did not (or could not) call init before the start of their test
		keyMap, err := readKeysFile(i.keysFile)
		ifErrLogExit(cmd, err)

		validation := shared.Validation{}
		issueID := validation.Int64Value("issue", i.changeListID, 0)
		patchsetID := validation.Int64Value("patchset", i.patchSetID, 0)
		jobID := validation.Int64Value("jobid", i.tryJobID, 0)
		ifErrLogExit(cmd, validation.Errors())

		// Define the meta data of the result that is shared by all tests.
		gr := jsonio.GoldResults{
			GitHash:            i.commitHash,
			Key:                keyMap,
			GerritChangeListID: issueID,
			GerritPatchSet:     patchsetID,
			BuildBucketID:      jobID,
		}

		config := goldclient.GoldClientConfig{
			FailureFile:     i.failureFile,
			InstanceID:      i.instanceID,
			OverrideGoldURL: i.urlOverride,
			PassFailStep:    i.passFailStep,
			UploadOnly:      i.uploadOnly,
			WorkDir:         i.workDir,
		}
		goldClient, err = goldclient.NewCloudClient(auth, config)
		ifErrLogExit(cmd, err)

		err = goldClient.SetSharedConfig(gr)
		ifErrLogExit(cmd, err)
	} else {
		// the user is presumed to have called init first, so we can just load it
		goldClient, err = goldclient.LoadCloudClient(auth, i.workDir)
		ifErrLogExit(cmd, err)
	}

	extraKeys := map[string]string{}
	if i.testKeysFile != "" {
		j, err := ioutil.ReadFile(i.testKeysFile)
		if err != nil {
			logErrf(cmd, "Could not read --add-test-key-file: does it exist? %s", err)
			exitProcess(cmd, 1)
		}
		if err = json.Unmarshal(j, &extraKeys); err != nil {
			logErrf(cmd, "--add-test-key-file was not a readable JSON object %s", err)
			exitProcess(cmd, 1)
		}
	} else {
		for _, pair := range i.testKeysStrings {
			split := strings.SplitN(pair, ":", 2)
			if len(split) != 2 {
				logInfof(cmd, "Ignoring malformatted --add-test-key=%s", pair)
			} else {
				extraKeys[split[0]] = split[1]
			}
		}
	}

	if i.corpus != "" {
		extraKeys[types.CORPUS_FIELD] = i.corpus
	}

	pass, err := goldClient.Test(types.TestName(i.testName), i.pngFile, extraKeys)
	ifErrLogExit(cmd, err)

	if !pass {
		logErrf(cmd, "Test: %s FAIL\n", i.testName)
		exitProcess(cmd, 1)
	}
	logInfof(cmd, "Test: %s PASS\n", i.testName)
	exitProcess(cmd, 0)
}

func (i *imgTest) runImgTestFinalizeCmd(cmd *cobra.Command, args []string) {
	auth, err := goldclient.LoadAuthOpt(i.workDir)
	ifErrLogExit(cmd, err)

	if auth == nil {
		logErrf(cmd, "Auth is empty - did you call goldctl auth first?")
		exitProcess(cmd, 1)
	}

	auth.SetDryRun(flagDryRun)

	// the user is presumed to have called init and tests first, so we just
	// have to load it from disk.
	goldClient, err := goldclient.LoadCloudClient(auth, i.workDir)
	ifErrLogExit(cmd, err)

	logVerbose(cmd, "Uploading the final JSON to Gold\n")
	err = goldClient.Finalize()
	ifErrLogExit(cmd, err)
	exitProcess(cmd, 0)
}

// readKeysFile is a helper function to read a JSON file with key/value pairs.
func readKeysFile(keysFile string) (map[string]string, error) {
	reader, err := os.Open(keysFile)
	if err != nil {
		return nil, err
	}

	ret := map[string]string{}
	err = json.NewDecoder(reader).Decode(&ret)
	return ret, err
}
