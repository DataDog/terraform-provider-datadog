package cli

import (
	"github.com/spf13/cobra"
)

type globalFlags struct {
	spec          string
	outputRoot    string
	hooksRoot     string
	trackingField string
	maxDepth      int
	report        string
	quiet         bool
}

func newRootCmd(version string, flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tfgen",
		Short:        "Datadog Terraform Provider Generator",
		Version:      version,
		SilenceUsage: true,
	}

	cmd.PersistentFlags().StringVar(&flags.spec, "spec", ".generator/V2/openapi.yaml", "OpenAPI spec to read")
	cmd.PersistentFlags().StringVar(&flags.outputRoot, "output-root", "datadog/fwprovider", "Root directory for generated artifacts")
	cmd.PersistentFlags().StringVar(&flags.hooksRoot, "hooks-root", "datadog/fwprovider/hooks", "Root directory for hook subpackages")
	cmd.PersistentFlags().StringVar(&flags.trackingField, "tracking-field", "x-datadog-tf-generator", "OpenAPI extension name for the tracking field")
	cmd.PersistentFlags().IntVar(&flags.maxDepth, "max-depth", 8, "Hard limit on recursive $ref expansion")
	cmd.PersistentFlags().StringVar(&flags.report, "report", "-", "Where to write the run report (\"-\" = stdout)")
	cmd.PersistentFlags().BoolVar(&flags.quiet, "quiet", false, "Suppress informational logging")

	return cmd
}

// Execute is the entry point called by main. Returns an exit code.
func Execute(version string) int {
	flags := &globalFlags{}
	root := newRootCmd(version, flags)
	root.AddCommand(newGenerateCmd(flags))
	root.AddCommand(newVerifyCmd(flags))

	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}
