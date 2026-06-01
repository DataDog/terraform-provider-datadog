package cli

import (
	"github.com/spf13/cobra"
)

func newGenerateCmd(flags *globalFlags) *cobra.Command {
	var check bool
	var include string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Terraform artifacts from the OpenAPI spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement in T033
			return nil
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "Read-only mode: exit 3 if any file would change")
	cmd.Flags().StringVar(&include, "include", "", "Comma-separated artifact names to generate (empty = all)")
	_ = check
	_ = include

	return cmd
}
