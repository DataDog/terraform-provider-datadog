package cli

import (
	"github.com/spf13/cobra"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

func newGenerateCmd(flags *globalFlags) *cobra.Command {
	var check bool
	var include string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Terraform artifacts from the OpenAPI spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := parser.LoadSpec(flags.spec, parser.WithMaxDepth(flags.maxDepth))
			if err != nil {
				return err
			}

			// TODO: model -> emit -> writer -> report, honoring --check and --include.
			_ = spec
			_ = check
			_ = include
			return nil
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "Read-only mode: exit 3 if any file would change")
	cmd.Flags().StringVar(&include, "include", "", "Comma-separated artifact names to generate (empty = all)")

	return cmd
}
