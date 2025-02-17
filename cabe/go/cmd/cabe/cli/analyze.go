package cli

import (
	"fmt"
	"os"
	"strings"

	"go.skia.org/infra/cabe/go/analyzer"
	"go.skia.org/infra/go/sklog"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

// analyzeCmd holds the flag values and any internal state necessary for
// executing the `analyze` subcommand.
type analyzeCmd struct {
	commonCmd
}

// AnalyzeCommand returns a [*cli.Command] for running the analyzer process locally.
func AnalyzeCommand() *cli.Command {
	cmd := &analyzeCmd{}
	return &cli.Command{
		Name:        "analyze",
		Description: "analyze runs the analyzer process locally.",
		Usage:       "cabe analyze -- --pinpoint-job <pinpoint-job>",
		Flags:       cmd.flags(),
		Action:      cmd.action,
		After:       cmd.cleanup,
	}
}

// action runs the analyzer process locally.
func (cmd *analyzeCmd) action(cliCtx *cli.Context) error {
	ctx := cliCtx.Context
	if err := cmd.dialBackends(ctx); err != nil {
		return err
	}

	var analyzerOpts = []analyzer.Options{
		analyzer.WithCASResultReader(cmd.casResultReader),
		analyzer.WithSwarmingTaskReader(cmd.swarmingTaskReader),
		analyzer.WithExperimentSpec(cmd.experimentSpecFromFlags()),
	}

	a := analyzer.New(cmd.pinpointJobID, analyzerOpts...)

	results, err := a.Run(ctx)
	if err != nil {
		sklog.Errorf("failed to analyze %s: %+v", cmd.pinpointJobID, err)
		return err
	}

	printAnalysisResultsTable(results)

	return nil
}

var (
	defaultRowColor = []int{tablewriter.Normal, tablewriter.FgWhiteColor, tablewriter.BgBlackColor}
	rowColorPos     = []int{tablewriter.Bold, tablewriter.FgGreenColor, tablewriter.BgBlackColor}
	rowColorNeg     = []int{tablewriter.Bold, tablewriter.FgRedColor, tablewriter.BgBlackColor}
	headerColor     = []int{tablewriter.Bold, tablewriter.FgWhiteColor, tablewriter.BgBlackColor}
)

func printAnalysisResultsTable(a []analyzer.Results) {
	w := tablewriter.NewWriter(os.Stdout)
	headers := []string{"Benchmark", "Workload", "Control median", "Treatment median", "CI Lower", "CI Upper", "P Value"}
	w.SetHeader(headers)
	headerColors := []tablewriter.Colors{}
	for range headers {
		headerColors = append(headerColors, headerColor)
	}
	w.SetHeaderColor(headerColors...)
	w.SetAutoFormatHeaders(false)
	w.SetCenterSeparator("")
	w.SetColumnSeparator("")
	w.SetRowSeparator("")
	w.SetBorder(false)
	w.SetTablePadding("\t")
	w.SetNoWhiteSpace(false)

	for _, ar := range a {
		bmark := ar.Benchmark
		workload := ar.WorkLoad
		s := ar.Statistics
		rc := defaultRowColor
		if s.PValue <= 0.05 {
			if s.LowerCi < 0 && s.UpperCi < 0 {
				if upIsBetter(bmark, workload) {
					rc = rowColorNeg
				} else {
					rc = rowColorPos
				}
			} else if s.LowerCi > 0 && s.UpperCi > 0 {
				if upIsBetter(bmark, workload) {
					rc = rowColorPos
				} else {
					rc = rowColorNeg
				}
			}
		}

		row := []string{
			bmark,
			workload,
			fmt.Sprintf("%10.6f", s.YMedian),
			fmt.Sprintf("%10.6f", s.XMedian),
			fmt.Sprintf("%10.6f", s.LowerCi),
			fmt.Sprintf("%10.6f", s.UpperCi),
			fmt.Sprintf("%10.6f", s.PValue),
		}
		rowColors := []tablewriter.Colors{}
		for range row {
			rowColors = append(rowColors, rc)
		}
		w.Rich(row, rowColors)
	}
	w.Render()
}

/*
Ugly hack: Allows us to customize the color-coding for significant changes, which depends on
external client-specific details that we don't and shouldn't know about in shared code like this.
Until we have some kind of central store for authoritative metrics metadata, we'll have
to continue hard-coding special cases that would be better managed as configuration/data than code.
*/
func upIsBetter(benchmark, workload string) bool {
	if benchmark == "motionmark" || benchmark == "jetstream2" || strings.Contains(workload, "RunsPerMinute") {
		return true
	}
	return false
}
