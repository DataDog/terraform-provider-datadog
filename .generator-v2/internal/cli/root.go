package cli

import (
	"github.com/spf13/cobra"
)

type globalFlags struct {
	quiet bool
}

func newRootCmd(version string, flags *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tfgen",
		Short:        "Datadog Terraform Provider Generator",
		Version:      version,
		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVar(&flags.quiet, "quiet", false, "Suppress informational logging")

	return cmd
}

// Execute is the entry point called by main. Returns an exit code.
func Execute(version string) int {
	flags := &globalFlags{}
	root := newRootCmd(version, flags)
	root.AddCommand(newGenerateCmd(flags))
	root.AddCommand(newVerifyCmd(flags))
	root.AddCommand(newSplitCmd(flags))

	if err := root.Execute(); err != nil {
		if err == errCheckFailed {
			return 3
		}
		return 1
	}
	return 0
}
