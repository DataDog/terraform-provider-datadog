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
	var spec          string
	var outputRoot    string
	var hooksRoot     string
	var trackingField string
	var maxDepth      int
	var report        string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Terraform artifacts from the OpenAPI spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := parser.LoadSpec(spec,
				parser.WithMaxDepth(maxDepth),
				parser.WithTrackingFieldName(trackingField))
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

	cmd.PersistentFlags().BoolVar(&check, "check", false, "Read-only mode: exit 3 if any file would change")
	cmd.PersistentFlags().IntVar(&maxDepth, "max-depth", parser.DefaultMaxDepth, "Hard limit on recursive $ref expansion")
	cmd.PersistentFlags().StringVar(&spec, "spec", ".generator/V2/openapi.yaml", "OpenAPI spec to read")
	cmd.PersistentFlags().StringVar(&include, "include", "", "Comma-separated artifact names to generate (empty = all)")
	cmd.PersistentFlags().StringVar(&outputRoot, "output-root", "datadog/fwprovider", "Root directory for generated artifacts")
	cmd.PersistentFlags().StringVar(&hooksRoot, "hooks-root", "datadog/fwprovider/hooks", "Root directory for hook subpackages")
	cmd.PersistentFlags().StringVar(&trackingField, "tracking-field", "x-datadog-tf-generator", "OpenAPI extension name for the tracking field")
	cmd.PersistentFlags().StringVar(&report, "report", "-", "Where to write the run report (\"-\" = stdout)")

	return cmd
}
