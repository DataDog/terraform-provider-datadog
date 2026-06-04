package cli

import (
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

func newGenerateCmd(flags *globalFlags) *cobra.Command {
	var check bool
	var include string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Terraform artifacts from the OpenAPI spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := parser.LoadSpec(flags.spec,
				parser.WithMaxDepth(flags.maxDepth),
				parser.WithTrackingFieldName(flags.trackingField))
			if err != nil {
				return err
			}
			runReport := model.RunReport{
				RunId:            uuid.NewString(),   // v4 uuid
				GeneratorVersion: cmd.Root().Version, // Version stamped by main.go
				SpecHash:         spec.Hash,
				StartedAt:        time.Now(),
			}

			// TODO: model -> emit -> writer -> report, honoring --check and --include.
			_ = spec
			_ = check
			_ = include

			runReport.FinishedAt = time.Now()
			return nil
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "Read-only mode: exit 3 if any file would change")
	cmd.Flags().StringVar(&include, "include", "", "Comma-separated artifact names to generate (empty = all)")

	return cmd
}
