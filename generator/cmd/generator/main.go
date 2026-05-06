package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DataDog/terraform-provider-datadog/generator/internal/codegen"
	"github.com/DataDog/terraform-provider-datadog/generator/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dd-tf-generator",
		Short: "Generate Terraform data sources from OpenAPI specs",
	}

	var configPath string
	var outputDir string
	var dryRun bool
	var maxDepth int

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Terraform data source files",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			configDir := filepath.Dir(configPath)
			if err := codegen.Generate(cfg, configDir, outputDir, dryRun, maxDepth); err != nil {
				return fmt.Errorf("generation failed: %w", err)
			}

			if !dryRun {
				fmt.Println("Generation complete.")
			}
			return nil
		},
	}

	generateCmd.Flags().StringVar(&configPath, "config", "", "Path to YAML configuration file (required)")
	generateCmd.Flags().StringVar(&outputDir, "output", "", "Output directory for generated files (required)")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview output without writing files")
	generateCmd.Flags().IntVar(&maxDepth, "max-depth", 5, "Maximum recursion depth for nested block generation")
	_ = generateCmd.MarkFlagRequired("config")
	_ = generateCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(generateCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
