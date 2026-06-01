package cli

import (
	"github.com/spf13/cobra"
)

func newVerifyCmd(flags *globalFlags) *cobra.Command {
	var strict bool

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Run post-generation checks without writing files",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement in T065
			return nil
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "Treat orphaned-hook warnings as errors")
	_ = strict

	return cmd
}
