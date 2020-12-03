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
	"go.skia.org/infra/golden/go/types"
)

// imgTest is the state for the imgtest command and its sub-commands.
// Specifically, it houses the flags.
type imgTest struct {
	// Common flags.
	codeReviewSystem            string
	continuousIntegrationSystem string
	commitHash                  string
	corpus                      string
	failureFile                 string
	instanceID                  string
	changelistID                string
	tryJobID                    string
	keysFile                    string
	passFailStep                bool
	patchsetOrder               int
	patchsetID                  string
	uploadOnly                  bool
	urlOverride                 string
	workDir                     string

	testName  string
	pngFile   string
	pngDigest string

	testKeysFile    string   // File with a JSON dictionary of test-specific keys.
	testKeysStrings []string // Test-specific keys represented as key:value pairs.

	testOptionalKeysFile    string   // File with a JSON dictionary of test-specific optional keys.
	testOptionalKeysStrings []string // Test-specific optional keys represented as key:value pairs.
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
	imgTestInitCmd.Flags().StringSliceVar(&env.testKeysStrings, "key", []string{}, "Any number of keys represented as key:value pairs. These keys will be applied to every test in this run. If specified, keys-file will be ignored.")

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
	env.addKeysFlags(imgTestAddCmd, "add-test-" /* =flagsPrefix */)
	imgTestAddCmd.Flags().StringVar(&env.testName, "test-name", "", "Unique name of the test, must not contain spaces.")
	imgTestAddCmd.Flags().StringVar(&env.pngFile, "png-file", "", "Path to the PNG file that contains the test results. png-file or png-digest must be provided")
	imgTestAddCmd.Flags().StringVar(&env.pngDigest, "png-digest", "", "If provided, will be used as the digest for the given image. If omitted, an md5 hash of the pixel content will be done and used.")

	Must(imgTestAddCmd.MarkFlagRequired("test-name"))

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
		PreRunE: env.validate,
		Run:     env.runImgTestCheckCmd,
	}
	env.addKeysFlags(imgTestCheckCmd, "" /* =flagsPrefix */)
	imgTestCheckCmd.Flags().StringVar(&env.workDir, fstrWorkDir, "", "Work directory for intermediate results")
	imgTestCheckCmd.Flags().StringVar(&env.testName, "test-name", "", "Unique name of the test, must not contain spaces.")
	imgTestCheckCmd.Flags().StringVar(&env.pngFile, "png-file", "", "Path to the PNG file that contains the test results.")
	imgTestCheckCmd.Flags().StringVar(&env.instanceID, "instance", "", "ID of the Gold instance.")

	imgTestCheckCmd.Flags().StringVar(&env.changelistID, "changelist", "", "If provided, the ChangelistExpectations matching this will apply.")
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

	cmd.Flags().StringVar(&i.changelistID, "changelist", "", "Changelist ID if this is run as a TryJob.")
	cmd.Flags().StringVar(&i.codeReviewSystem, "crs", "", "CodeReviewSystem, if any (e.g. 'gerrit', 'github')")
	cmd.Flags().StringVar(&i.commitHash, "commit", "", "Git commit hash")
	cmd.Flags().StringVar(&i.continuousIntegrationSystem, "cis", "", "ContinuousIntegrationSystem, if any (e.g. 'buildbucket')")
	cmd.Flags().StringVar(&i.corpus, "corpus", "", "Gold Corpus Name. Overrides any other values (e.g. from keys-file or add-test-key)")
	cmd.Flags().StringVar(&i.failureFile, "failure-file", "", "Path to the file where to write failure information")
	cmd.Flags().StringVar(&i.keysFile, "keys-file", "", "JSON file containing key/value pairs commmon to all tests")
	cmd.Flags().IntVar(&i.patchsetOrder, "patchset", 0, "Patchset number if this is run as a TryJob.")
	cmd.Flags().StringVar(&i.patchsetID, "patchset_id", "", "Patchset id (e.g. githash) if this is run as a TryJob.")
	cmd.Flags().StringVar(&i.tryJobID, "jobid", "", "TryJob ID if this is a TryJob run.")
	cmd.Flags().StringVar(&i.urlOverride, "url", "", "URL of the Gold instance. Used for testing, if empty the URL will be derived from the value of 'instance'")

	cmd.Flags().StringVar(&i.changelistID, "issue", "", "[deprecated] Gerrit issue if this is trybot run. ")
	Must(cmd.MarkFlagRequired(fstrWorkDir))
	if !optional {
		Must(cmd.MarkFlagRequired("instance"))
		Must(cmd.MarkFlagRequired("commit"))
	}
}

func (i *imgTest) addKeysFlags(cmd *cobra.Command, flagsPrefix string) {
	cmd.Flags().StringVar(&i.testKeysFile, flagsPrefix+"keys-file", "", "File with a JSON dictionary of test-specific keys.")
	cmd.Flags().StringSliceVar(&i.testKeysStrings, flagsPrefix+"key", []string{}, "Any number of test-specific keys represented as key:value pairs.")
	cmd.Flags().StringVar(&i.testOptionalKeysFile, flagsPrefix+"optional-keys-file", "", "File with a JSON dictionary of test-specific optional keys.")
	cmd.Flags().StringSliceVar(&i.testOptionalKeysStrings, flagsPrefix+"optional-key", []string{}, "Any number of test-specific optional keys represented as key:value pairs.")
}

func (i *imgTest) validate(cmd *cobra.Command, args []string) error {
	if i.uploadOnly && i.passFailStep {
		return errors.New("Cannot have --upload-only and --passfail both be true.")
	}
	if i.testKeysFile != "" && len(i.testKeysStrings) > 0 {
		return errors.New("Cannot have both --add-test-key and --add-test-keys-file.")
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

		if i.changelistID != "" {
			gr := jsonio.GoldResults{
				ChangelistID: i.changelistID,
				GitHash:      "HEAD",
			}
			err = goldClient.SetSharedConfig(gr, true) // this will load the baseline
			ifErrLogExit(cmd, err)
		}
	}

	// Read test keys. These are only necessary if a non-exact image matching algorithm is specified.
	keys := readKeyValuePairsFromFileOrStringSlice(cmd, i.testKeysFile, i.testKeysStrings)

	// Read optional keys. Only used to specify a non-exact image matching algorithm and parameters.
	optionalKeys := readKeyValuePairsFromFileOrStringSlice(cmd, i.testOptionalKeysFile, i.testOptionalKeysStrings)

	pass, err := goldClient.Check(types.TestName(i.testName), i.pngFile, keys, optionalKeys)
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

	if i.keysFile == "" && len(i.testKeysStrings) == 0 {
		logErrf(cmd, "You must supply --keys-file or at least one --key")
		exitProcess(cmd, 1)
	}
	keyMap := readKeyValuePairsFromFileOrStringSlice(cmd, i.keysFile, i.testKeysStrings)

	if i.corpus != "" {
		keyMap[types.CorpusField] = i.corpus
	}

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
	gr := jsonio.GoldResults{
		GitHash:                     i.commitHash,
		Key:                         keyMap,
		ChangelistID:                i.changelistID,
		PatchsetOrder:               i.patchsetOrder,
		PatchsetID:                  i.patchsetID,
		CodeReviewSystem:            i.codeReviewSystem,
		TryJobID:                    i.tryJobID,
		ContinuousIntegrationSystem: i.continuousIntegrationSystem,
	}

	logVerbose(cmd, "Loading hashes and baseline from Gold\n")
	err = goldClient.SetSharedConfig(gr, false)
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

	if i.pngDigest == "" && i.pngFile == "" {
		logErrf(cmd, "Must supply png-file or png-digest (or both)")
		exitProcess(cmd, 1)
	}

	var goldClient goldclient.GoldClient

	if i.keysFile != "" {
		// user has specified a full set of keys. This happens if they
		// did not (or could not) call init before the start of their test
		keyMap, err := readKeysFile(i.keysFile)
		ifErrLogExit(cmd, err)

		// Define the meta data of the result that is shared by all tests.
		gr := jsonio.GoldResults{
			GitHash:                     i.commitHash,
			Key:                         keyMap,
			ChangelistID:                i.changelistID,
			PatchsetOrder:               i.patchsetOrder,
			PatchsetID:                  i.patchsetID,
			CodeReviewSystem:            i.codeReviewSystem,
			TryJobID:                    i.tryJobID,
			ContinuousIntegrationSystem: i.continuousIntegrationSystem,
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

		err = goldClient.SetSharedConfig(gr, false)
		ifErrLogExit(cmd, err)
	} else {
		// the user is presumed to have called init first, so we can just load it
		goldClient, err = goldclient.LoadCloudClient(auth, i.workDir)
		ifErrLogExit(cmd, err)
	}

	// Read test-specific keys. These will be merged with the shared keys provided via the "init"
	// command.
	additionalKeys := readKeyValuePairsFromFileOrStringSlice(cmd, i.testKeysFile, i.testKeysStrings)
	if i.corpus != "" {
		additionalKeys[types.CorpusField] = i.corpus
	}

	// Read optional keys. Unlike additionalKeys, no shared optional keys are provided via the "init"
	// command.
	optionalKeys := readKeyValuePairsFromFileOrStringSlice(cmd, i.testOptionalKeysFile, i.testOptionalKeysStrings)

	pass, err := goldClient.Test(types.TestName(i.testName), i.pngFile, types.Digest(i.pngDigest), additionalKeys, optionalKeys)
	ifErrLogExit(cmd, err)

	if !pass {
		logErrf(cmd, "Test: %s FAIL\n", i.testName)
		exitProcess(cmd, 1)
	}
	logInfof(cmd, "Test: %s PASS\n", i.testName)
	exitProcess(cmd, 0)
}

// readKeyValuePairsFromFileOrStringSlice reads key/value pairs encoded as a JSON dictionary from
// the given filename, or from the given slice of key:value strings if the filename is empty.
func readKeyValuePairsFromFileOrStringSlice(cmd *cobra.Command, filename string, keyValueStrings []string) map[string]string {
	retval := map[string]string{}

	if filename != "" {
		jsonBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			logErrf(cmd, "Could not read file %s: %s", filename, err)
			exitProcess(cmd, 1)
		}
		if err = json.Unmarshal(jsonBytes, &retval); err != nil {
			logErrf(cmd, "File %s does not contain readable JSON object: %s", filename, err)
			exitProcess(cmd, 1)
		}
	} else {
		for _, pair := range keyValueStrings {
			split := strings.SplitN(pair, ":", 2)
			if len(split) != 2 {
				logInfof(cmd, "Ignoring malformatted key:value pair %s", pair)
			} else {
				retval[split[0]] = split[1]
			}
		}
	}

	return retval
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
