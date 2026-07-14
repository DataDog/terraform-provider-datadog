package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/split"
)

// newSplitCmd builds the `tfgen split` subcommand. It fans one aggregate generated
// push (a branch carrying all N data sources) out into one bundle per artifact so
// each can land as its own PR, by diffing the provider and docs paths in a base
// checkout against the pushed-branch checkout. It never runs the generator, the
// spec, or git — it routes emitted files.
func newSplitCmd(flags *globalFlags) *cobra.Command {
	var baseDir, generatedDir, outDir, reportPath string
	var check bool

	cmd := &cobra.Command{
		Use:   "split",
		Short: "Split an aggregate generated push into per-artifact bundles",
		RunE: func(cmd *cobra.Command, args []string) error {
			rep, splitErr := split.Split(split.Options{
				BaseDir:      baseDir,
				GeneratedDir: generatedDir,
				OutDir:       outDir,
				Check:        check,
			})

			// Write the report even on attribution failure, so a human sees exactly
			// what could not be routed.
			if err := writeSplitReport(rep, reportPath, cmd); err != nil {
				return err
			}
			return splitErr
		},
	}

	cmd.Flags().StringVar(&baseDir, "base-dir", "", "Checkout of the base branch the push is diffed against (required)")
	cmd.Flags().StringVar(&generatedDir, "generated-dir", "", "Checkout of the pushed branch carrying the generated artifacts (required)")
	cmd.Flags().StringVar(&outDir, "out", "", "Directory to receive one bundle per artifact (required)")
	cmd.Flags().StringVar(&reportPath, "report", "-", "Where to write the split report (\"-\" = stdout)")
	cmd.Flags().BoolVar(&check, "check", false, "Plan and report without writing the output bundles")
	_ = cmd.MarkFlagRequired("base-dir")
	_ = cmd.MarkFlagRequired("generated-dir")
	_ = cmd.MarkFlagRequired("out")
	_ = flags

	return cmd
}

// writeSplitReport sends the report to reportPath, mapping "-" to the command's
// stdout and anything else to a created file.
func writeSplitReport(rep *split.Result, reportPath string, cmd *cobra.Command) error {
	if reportPath == "-" {
		return rep.WriteJSON(cmd.OutOrStdout())
	}
	f, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("split: opening report %s: %w", reportPath, err)
	}
	defer f.Close()
	return rep.WriteJSON(f)
}
