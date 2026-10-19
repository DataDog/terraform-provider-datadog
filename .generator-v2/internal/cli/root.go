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

// RootCommand assembles the full tfgen command tree. Execute runs it, and the
// docs generator (cmd/tfgen-docs) walks it to render docs/reference/cli.md, so
// that page cannot drift from the flags actually registered here.
func RootCommand(version string) *cobra.Command {
	flags := &globalFlags{}
	root := newRootCmd(version, flags)
	root.AddCommand(newGenerateCmd(flags))
	root.AddCommand(newVerifyCmd(flags))
	root.AddCommand(newSplitCmd(flags))
	return root
}

// Execute is the entry point called by main. Returns an exit code.
func Execute(version string) int {
	root := RootCommand(version)

	if err := root.Execute(); err != nil {
		if err == errCheckFailed {
			return 3
		}
		return 1
	}
	return 0
}
