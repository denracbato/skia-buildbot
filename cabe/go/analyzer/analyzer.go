package analyzer

import (
	"context"
	"fmt"
	"sort"
	"strings"

	stat "github.com/aclements/go-moremath/stats"
	"github.com/bazelbuild/remote-apis-sdks/go/pkg/digest"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.skia.org/infra/go/sklog"
	"golang.org/x/sync/errgroup"

	"go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.skia.org/infra/cabe/go/backends"
	"go.skia.org/infra/cabe/go/perfresults"
	cpb "go.skia.org/infra/cabe/go/proto"
	cabe_stats "go.skia.org/infra/cabe/go/stats"
)

const maxReadCASPoolWorkers = 100

// Options configure one or more fields of an Analyzer instance.
type Options func(*Analyzer)

// WithCASResultReader configures an Analyzer instance to use the given CASResultReader.
func WithCASResultReader(r backends.CASResultReader) Options {
	return func(e *Analyzer) {
		e.readCAS = r
	}
}

// WithTaskResultsReader configures an Analyzer instance to use the given TaskResultsReader.
func WithSwarmingTaskReader(r backends.SwarmingTaskReader) Options {
	return func(e *Analyzer) {
		e.readSwarmingTasks = r
	}
}

// WithExperimentSpec configures an Analyzer instance to use the given ExperimentSpec.
func WithExperimentSpec(s *cpb.ExperimentSpec) Options {
	return func(e *Analyzer) {
		e.experimentSpec = s
	}
}

// New returns a new instance of Analyzer. Set either pinpointJobID, or controlDigests and treatmentDigests.
func New(pinpointJobID string, opts ...Options) *Analyzer {
	ret := &Analyzer{
		pinpointJobID: pinpointJobID,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

// Analyzer encapsulates the state of an Analyzer process execution. Its lifecycle follows a request
// to process all of the output of an A/B benchmark experiment run.
// Users of Analyzer must instantiate and attach the necessary service dependencies.
type Analyzer struct {
	pinpointJobID     string
	readCAS           backends.CASResultReader
	readSwarmingTasks backends.SwarmingTaskReader

	experimentSpec *cpb.ExperimentSpec

	results []Results
}

// Results encapsulates a response from the Go statistical package after it has processed
// swarming task data and verified the experimental setup is valid for analysis.
type Results struct {
	// Benchmark is the name of a perf benchmark suite, such as Speedometer2 or JetStream
	Benchmark string
	// Workload is the name of a benchmark-specific workload, such as TodoMVC-ReactJS
	WorkLoad string
	// BuildConfig is the name of a build configuration, e.g. "Mac arm Builder Perf PGO"
	BuildConfig string
	// RunConfig is the name of a run configuration, e.g. "Macmini9,1_arm64-64-Apple_M1_16384_1_4744421.0"
	RunConfig string
	// Statistics summarizes the difference between the treatment and control arms for the given
	// Benchmark and Workload on the hardware described by RunConfig, using the binary built using
	// the given BuildConfig.
	Statistics *cabe_stats.BerfWilcoxonSignedRankedTestResult
}

// AnalysisResults returns a slice of AnalysisResult protos populated with data from the
// experiment.
func (a *Analyzer) AnalysisResults() []*cpb.AnalysisResult {
	ret := []*cpb.AnalysisResult{}
	// Because ExperimentSpec will have so many identical
	// values across individual results, we'll build a template here
	// then clone and override the distinct per-result values for the
	// response proto below.
	//
	// Note that for most Pinpoint A/B tryjobs, the ExperimentSpec will
	// have a Common RunSpec set, and Treatment and Control will have different BuildSpec values.
	// That is, compare two different builds executing on the same hardware/OS.
	experimentSpecTemplate := a.experimentSpec
	if experimentSpecTemplate.Analysis == nil {
		experimentSpecTemplate.Analysis = &cpb.AnalysisSpec{}
	}
	experimentSpecTemplate.Analysis.Benchmark = nil

	sort.Sort(byBenchmarkAndWorkload(a.results))

	for _, res := range a.results {
		experimentSpec := proto.Clone(experimentSpecTemplate).(*cpb.ExperimentSpec)
		benchmark := []*cpb.Benchmark{
			{
				Name:     res.Benchmark,
				Workload: []string{res.WorkLoad},
			},
		}

		experimentSpec.Analysis.Benchmark = benchmark

		ret = append(ret, &cpb.AnalysisResult{
			ExperimentSpec: experimentSpec,
			Statistic: &cpb.Statistic{
				Upper:           res.Statistics.UpperCi,
				Lower:           res.Statistics.LowerCi,
				PValue:          res.Statistics.PValue,
				ControlMedian:   res.Statistics.YMedian,
				TreatmentMedian: res.Statistics.XMedian,
			},
		})
	}
	return ret
}

type byBenchmarkAndWorkload []Results

func (a byBenchmarkAndWorkload) Len() int      { return len(a) }
func (a byBenchmarkAndWorkload) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byBenchmarkAndWorkload) Less(i, j int) bool {
	if a[i].Benchmark != a[j].Benchmark {
		return a[i].Benchmark < a[j].Benchmark
	}
	return a[i].WorkLoad < a[j].WorkLoad
}

func (a *Analyzer) ExperimentSpec() *cpb.ExperimentSpec {
	return a.experimentSpec
}

func (a *Analyzer) filterCompleteSwarmingTasks(taskInfos []*swarming.SwarmingRpcsTaskRequestMetadata) []*swarming.SwarmingRpcsTaskRequestMetadata {
	ret := []*swarming.SwarmingRpcsTaskRequestMetadata{}
	for _, taskInfo := range taskInfos {
		if taskInfo.TaskResult.State == taskCompletedState {
			ret = append(ret, taskInfo)
		} else {
			sklog.Warningf("ignoring swarming task %s because it is in state %q rather than %q", taskInfo.TaskId, taskInfo.TaskResult.State, taskCompletedState)
		}
	}
	return ret
}

// Run executes the whole Analyzer process for a single, complete experiment.
// TODO(seanmccullough): break this up into distinct, testable stages with one function per stage.
func (a *Analyzer) Run(ctx context.Context) ([]Results, error) {
	var control, treatment []float64
	var benchmark, workload, buildspec, runspec []string
	var replicas []int

	res := []Results{}

	allTaskInfos, err := a.readSwarmingTasks(ctx, a.pinpointJobID)
	if err != nil {
		return res, err
	}

	allTaskInfos = a.filterCompleteSwarmingTasks(allTaskInfos)

	// TODO(seanmccullough): choose different process<Foo>Tasks implementations based on AnalysisSpec values.
	processedArms, err := processPinpointTryjobTasks(allTaskInfos)
	if err != nil {
		return res, err
	}
	err = a.extractTaskOutputs(ctx, processedArms)
	if err != nil {
		return res, err
	}

	// TODO(seanmccullough): include pairing order information so we keep track of which arm executed first in
	// every pairing.
	pairs, err := processedArms.pairedTasks()
	if err != nil {
		return res, err
	}

	if a.experimentSpec == nil {
		ieSpec, err := a.inferExperimentSpec(pairs)
		if err != nil {
			sklog.Errorf("while trying to infer experiment spec: %v", err)
			return res, err
		}

		a.experimentSpec = ieSpec
	}

	for replicaNumber, pair := range pairs {
		// Check task result codes to identify and handle task failures (which are expected; lab hardware is inherently unreliable).
		if pair.hasTaskFailures() {
			sklog.Infof("excluding replica %d from analysis since one or both of its arm's swarming tasks (control: %s, treatment: %s) failed", replicaNumber, pair.control.taskID, pair.treatment.taskID)
			continue
		}
		controlTask, treatmentTask := pair.control, pair.treatment

		runSpecName := pair.control.runConfig
		buildSpecName := pair.control.buildConfig
		for _, benchmarkSpec := range a.experimentSpec.GetAnalysis().GetBenchmark() {
			cRes, ok := controlTask.parsedResults[benchmarkSpec.GetName()]
			if !ok {
				return res, fmt.Errorf("benchmark missing from control task %v: %q", controlTask.taskID, benchmarkSpec.GetName())
			}
			tRes, ok := treatmentTask.parsedResults[benchmarkSpec.GetName()]
			if !ok {
				return res, fmt.Errorf("benchmark missing from treatment task %s: %q", treatmentTask.taskID, benchmarkSpec.GetName())
			}

			cSamples := map[string][]float64{}
			for _, cHist := range cRes.Histograms {
				cSamples[cHist.Name] = cHist.SampleValues
			}

			tSamples := map[string][]float64{}
			for _, tHist := range tRes.Histograms {
				tSamples[tHist.Name] = tHist.SampleValues
			}
			for _, workloadName := range benchmarkSpec.GetWorkload() {
				cValues, ok := cSamples[workloadName]
				if !ok {
					return res, fmt.Errorf("replica %d control task %s is missing %q/%q", replicaNumber, pair.control.taskID, benchmarkSpec.GetName(), workloadName)
				}

				tValues, ok := tSamples[workloadName]
				if !ok {
					return res, fmt.Errorf("replica %d treatment task %s is missing %q/%q", replicaNumber, pair.treatment.taskID, benchmarkSpec.GetName(), workloadName)
				}

				if len(cValues) == 0 || len(tValues) == 0 {
					sklog.Warningf("control (%s) and/or treatment (%s) task had no sample values for %q/%q: %d vs %d", pair.control.taskID, pair.treatment.taskID, benchmarkSpec.GetName(), workloadName, len(cValues), len(tValues))
					continue
				}
				cMean := stat.Mean(cValues)
				tMean := stat.Mean(tValues)

				benchmark = append(benchmark, benchmarkSpec.GetName())
				workload = append(workload, workloadName)
				buildspec = append(buildspec, buildSpecName)
				runspec = append(runspec, runSpecName)
				replicas = append(replicas, replicaNumber)
				control = append(control, cMean)
				treatment = append(treatment, tMean)
			}
		}
	}

	// Group control/treatment pairs by [benchmark, workload, buildspec, runspec], aggregated over [replicas]
	// such that we end up with a map of [benchmark, workload, buildspec, runspec] to lists of [control, treatment] ordered by replica number within the lists
	type groupKey struct{ benchmark, workload, buildspec, runspec string }
	type tcPair struct{ control, treatment float64 }
	aggregateOverReplicas := map[groupKey]map[int]*tcPair{}

	for i, replicaNumber := range replicas {
		gk := groupKey{benchmark[i], workload[i], buildspec[i], runspec[i]}
		tcps := aggregateOverReplicas[gk]
		if tcps == nil {
			tcps = map[int]*tcPair{}
			aggregateOverReplicas[gk] = tcps
		}
		if tcps[replicaNumber] != nil {
			return res, fmt.Errorf("should not have a treatment/control pair aggregate for replica %d yet", replicaNumber)
		}
		tcps[replicaNumber] = &tcPair{control: control[i], treatment: treatment[i]}
	}

	for gk, tcps := range aggregateOverReplicas {
		ctrls, trts := []float64{}, []float64{}

		for _, tcp := range tcps {
			ctrls = append(ctrls, tcp.control)
			trts = append(trts, tcp.treatment)
		}
		r, err := cabe_stats.BerfWilcoxonSignedRankedTest(trts, ctrls, cabe_stats.TwoSided, cabe_stats.LogTransform)
		if err != nil {
			sklog.Errorf("cabe_stats.BerfWilcoxonSignedRankedTest returned an error (%q), "+
				"printing the table of parameters passed to it below:",
				err)
			return res, errors.Wrap(err, "problem reported by cabe_stats.BerfWilcoxonSignedRankedTest")
		}
		res = append(res, Results{
			Benchmark:   gk.benchmark,
			WorkLoad:    gk.workload,
			BuildConfig: gk.buildspec,
			RunConfig:   gk.runspec,
			Statistics:  r,
		})
	}

	a.results = res
	return res, nil
}

// RunChecker verifies some assumptions we need to make about the experiment data input for
// our analyses.
func (a *Analyzer) RunChecker(ctx context.Context, c Checker) error {
	allTaskInfos, err := a.readSwarmingTasks(ctx, a.pinpointJobID)
	if err != nil {
		return err
	}

	for _, taskInfo := range allTaskInfos {
		c.CheckSwarmingTask(taskInfo)
	}

	processedTasks, err := processPinpointTryjobTasks(allTaskInfos)
	if err != nil {
		sklog.Errorf("RunChecker: processPinpointTryjobTasks returned %v", err)
		return err
	}

	err = a.extractTaskOutputs(ctx, processedTasks)
	if err != nil {
		sklog.Errorf("RunChecker: extractTaskOutputs returned %v", err)
		return err
	}

	pairs, err := processedTasks.pairedTasks()
	if err != nil {
		sklog.Errorf("RunChecker: processedTasks.pairedTasks() returned %v", err)
		return err
	}

	if a.experimentSpec == nil {
		ieSpec, err := a.inferExperimentSpec(pairs)
		if err != nil {
			sklog.Errorf("while trying to infer experiment spec: %v", err)
			return err
		}

		a.experimentSpec = ieSpec
	}

	for _, taskInfo := range allTaskInfos {
		c.CheckRunTask(taskInfo)
	}

	c.CheckArmComparability(processedTasks.control, processedTasks.treatment)

	return nil
}

func (a *Analyzer) inferExperimentSpec(pairs []pairedTasks) (*cpb.ExperimentSpec, error) {
	controlTaskResults, treatmentTaskResults := []map[string]perfresults.PerfResults{}, []map[string]perfresults.PerfResults{}
	controlArmSpecs, treatmentArmSpecs := []*cpb.ArmSpec{}, []*cpb.ArmSpec{}
	for replicaNumber, pair := range pairs {
		if pair.hasTaskFailures() {
			sklog.Infof("excluding replica %d from spec inference because it contains failures", replicaNumber)
			continue
		}

		controlTaskResults = append(controlTaskResults, pair.control.parsedResults)
		treatmentTaskResults = append(treatmentTaskResults, pair.treatment.parsedResults)
		cSpec, err := inferArmSpec(pair.control.taskInfo)
		if err != nil {
			return nil, err
		}
		controlArmSpecs = append(controlArmSpecs, cSpec)
		tSpec, err := inferArmSpec(pair.treatment.taskInfo)
		if err != nil {
			return nil, err
		}
		treatmentArmSpecs = append(treatmentArmSpecs, tSpec)
	}

	experimentSpec, err := inferExperimentSpec(controlArmSpecs, treatmentArmSpecs, controlTaskResults, treatmentTaskResults)
	if err != nil {
		return nil, err
	}

	return experimentSpec, nil
}

func (a *Analyzer) extractTaskOutputs(ctx context.Context, processedArms *processedExperimentTasks) error {
	// This is currently un-sliced. As in, it lumps all runconfigs together. This is fine if you only
	// have one runconfig (say, you only asked to analyze Mac results).
	// TODO(seanmccullough): add slicing, which will nest the code below inside an iterator over the slices.
	controlDigests := processedArms.control.outputDigests()
	treatmentDigests := processedArms.treatment.outputDigests()

	if len(controlDigests) != len(treatmentDigests) {
		return fmt.Errorf("control and treatment have different numbers of task output digests: %d vs %d", len(controlDigests), len(treatmentDigests))
	}

	var controlReplicaOutputs, treatmentReplicaOutputs map[string]swarming.SwarmingRpcsTaskResult
	// Fetch outputs from control and treatment arms in parallel since there is no data dependency between them.
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		controlReplicaOutputs, err := a.fetchOutputsFromReplicas(ctx, controlDigests)
		if err != nil {
			return err
		}
		for replica := range controlReplicaOutputs {
			processedArms.control.tasks[replica].parsedResults = controlReplicaOutputs[replica]
		}
		return nil
	})

	g.Go(func() error {
		treatmentReplicaOutputs, err := a.fetchOutputsFromReplicas(ctx, treatmentDigests)
		if err != nil {
			return err
		}
		for replica := range treatmentReplicaOutputs {
			processedArms.treatment.tasks[replica].parsedResults = treatmentReplicaOutputs[replica]
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if len(controlReplicaOutputs) != len(treatmentReplicaOutputs) {
		return fmt.Errorf("control and treatment have different numbers of benchmark outputs: %d vs %d", len(controlReplicaOutputs), len(treatmentReplicaOutputs))
	}

	return nil
}

// returns a slice of maps of perfresults.PerfResults files keyed by benchmark name.
func (a *Analyzer) fetchOutputsFromReplicas(ctx context.Context, outputs []*swarming.SwarmingRpcsCASReference) ([]map[string]perfresults.PerfResults, error) {
	ret := make([]map[string]perfresults.PerfResults, len(outputs))
	g, ctx := errgroup.WithContext(ctx)
	if len(outputs) > maxReadCASPoolWorkers {
		g.SetLimit(maxReadCASPoolWorkers)
	} else {
		g.SetLimit(len(outputs))
	}

	for replica := range outputs {
		replica := replica
		g.Go(func() error {
			casDigest, err := digest.New(outputs[replica].Digest.Hash, outputs[replica].Digest.SizeBytes)
			if err != nil {
				sklog.Errorf("digest.New: %v", err)
				return err
			}
			res, err := a.readCAS(ctx, outputs[replica].CasInstance, casDigest.String())
			if err != nil {
				sklog.Errorf("e.readCAS: %v", err)
				return err
			}
			ret[replica] = res
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		sklog.Errorf("fetchOutputsFromReplicas: %v", err)
		return nil, err
	}

	return ret, nil
}

// Split swarming.SwarmingRpcsTaskRequestMetadatas into experiment arms and pair tasks according to how they should be
// compared in the analysis/hypothesis testing phase.
// The slice of tasks should be a complete list of all the tasks for an experiment, and they should
// all have completed executing successfully.
func processPinpointTryjobTasks(tasks []*swarming.SwarmingRpcsTaskRequestMetadata) (*processedExperimentTasks, error) {
	ret := &processedExperimentTasks{
		treatment: &processedArmTasks{},
		control:   &processedArmTasks{},
	}

	buildTasks := map[string]*buildInfo{}
	for _, task := range tasks {
		// First check if we have a build task. It doesn't run any benchmarks, but it does build the
		// binaries for running benchmarks. So we need to keep track of it for BuildSpec details later.
		buildInfo, err := buildInfoForTask(task)
		if err != nil {
			sklog.Errorf("task.buildInfo(): %v", err)
			return nil, err
		}
		if buildInfo != nil {
			buildTasks[task.TaskId] = buildInfo
			// Just move on to processing the next task now that we know this was a build task.
			continue
		}

		// err should not be nil if the task is not a run task.
		runInfo, err := runInfoForTask(task)
		if err != nil {
			sklog.Errorf("runInfoForTask: %v", err)
			return nil, err
		}

		// If it's not a build task, assume it's a test runner task.
		// Get the CAS digest for the task output so we can fetch it later.
		if task.TaskResult.CasOutputRoot == nil || task.TaskResult.CasOutputRoot.Digest == nil {
			return nil, fmt.Errorf("run task result missing CasOutputRoot: %+v", task)
		}

		t := &armTask{
			taskID:       task.TaskId,
			resultOutput: task.TaskResult.CasOutputRoot,
			runInfo:      runInfo,
			buildInfo:    buildInfo,
			runConfig:    runInfo.String(),
			taskInfo:     task,
		}

		// For pinpoint tryjobs, the following assumptions should hold true:
		// treatment has a tag like "change:exp: chromium@d879632 + 0dd4ae0 (Variant: 1)"
		// control has a tag like "change:base: chromium@d879632 (Variant: 0)"
		change := pinpointChangeTagForTask(task)
		if change == "" {
			return nil, fmt.Errorf("missing pinpoint change tag: %+v", task)
		}
		if strings.HasPrefix(change, "exp:") {
			if ret.treatment.pinpointChangeTag == "" {
				ret.treatment.pinpointChangeTag = change
			}
			if change != ret.treatment.pinpointChangeTag {
				return nil, fmt.Errorf("mismatched change tag for treatment arm. Got %q but expected %q", change, ret.treatment.pinpointChangeTag)
			}
			ret.treatment.tasks = append(ret.treatment.tasks, t)
		} else if strings.HasPrefix(change, "base:") {
			if ret.control.pinpointChangeTag == "" {
				ret.control.pinpointChangeTag = change
			}
			if change != ret.control.pinpointChangeTag {
				return nil, fmt.Errorf("mismatched change tag for control arm. Got %q but expected %q", change, ret.control.pinpointChangeTag)
			}
			ret.control.tasks = append(ret.control.tasks, t)
		} else if len(strings.Split(change, "@")) == 2 {
			// This might be from a bisect job, where control and treatment are builds from two different commit positions on a main branch.
			// TODO(seanmccullough): also check for "comparison_mode:performance" tags on these tasks.
			parts := strings.Split(change, "@")
			repo := parts[0]
			cp := parts[1]
			return nil, fmt.Errorf("unsupported yet: changes that only identify repo and commit position: %v @ %v", repo, cp)
		} else {
			return nil, fmt.Errorf("unrecognized change tag: %q", change)
		}
	}

	// Validate some basic assumptions about the data we should have at this point.

	if len(ret.treatment.tasks) != len(ret.control.tasks) {
		return nil, fmt.Errorf("unequal number of tasks for treatment and control: %d vs %d", len(ret.treatment.tasks), len(ret.control.tasks))
	}

	return ret, nil
}
