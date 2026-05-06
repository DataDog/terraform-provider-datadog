package codegen

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/DataDog/terraform-provider-datadog/generator/internal/config"
	"github.com/DataDog/terraform-provider-datadog/generator/internal/openapi"
	"github.com/DataDog/terraform-provider-datadog/generator/internal/templates"
)

// Generate runs the full generation pipeline for all data sources in the config.
// maxDepth controls the maximum recursion depth for nested block generation (default: 5).
func Generate(cfg *config.Config, configDir, outputDir string, dryRun bool, maxDepth int) error {
	if maxDepth <= 0 {
		maxDepth = 5
	}

	engine, err := templates.NewEngine()
	if err != nil {
		return fmt.Errorf("initializing template engine: %w", err)
	}

	// Sort data source names for deterministic output
	names := make([]string, 0, len(cfg.DataSources))
	for name := range cfg.DataSources {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		dsCfg := cfg.DataSources[name]
		if err := generateDataSource(engine, cfg, configDir, outputDir, name, dsCfg, dryRun, maxDepth); err != nil {
			return fmt.Errorf("generating data source %q: %w", name, err)
		}
	}

	return nil
}

// generateDataSource generates a single data source.
// maxDepth is reserved for future use by recursive nested block generation (T099a).
func generateDataSource(engine *templates.Engine, cfg *config.Config, configDir, outputDir, name string, dsCfg config.DataSourceConfig, dryRun bool, maxDepth int) error {
	// Resolve the spec
	specName := cfg.ResolveSpec(dsCfg)
	specCfg := cfg.Specs[specName]
	specPath := specCfg.Path
	if !filepath.IsAbs(specPath) {
		specPath = filepath.Join(configDir, specPath)
	}

	// Load the spec
	docModel, err := openapi.LoadSpec(specPath)
	if err != nil {
		return fmt.Errorf("loading spec %s: %w", specPath, err)
	}

	// Extract the read operation
	op, err := openapi.ExtractOperation(&docModel.Model, dsCfg.Read.Path, dsCfg.Read.Method)
	if err != nil {
		return fmt.Errorf("extracting operation: %w", err)
	}
	op.Name = name

	// Extract the list operation if configured (Phase 7)
	var listOp *openapi.ParsedOperation
	if dsCfg.List != nil {
		listOp, err = openapi.ExtractOperation(&docModel.Model, dsCfg.List.Path, dsCfg.List.Method)
		if err != nil {
			return fmt.Errorf("extracting list operation: %w", err)
		}
	}

	// Get the response schema and detect JSON:API
	responseSchema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		return fmt.Errorf("building response schema: %w", err)
	}

	var schemaObj *openapi.SchemaObject
	isJSONAPI := openapi.IsJSONAPIEnvelope(responseSchema)

	if isJSONAPI {
		attrsProxy, _, err := openapi.UnwrapJSONAPI(responseSchema)
		if err != nil {
			return fmt.Errorf("unwrapping JSON:API: %w", err)
		}
		schemaObj, err = openapi.ParseSchema(attrsProxy)
		if err != nil {
			return fmt.Errorf("parsing JSON:API attributes schema: %w", err)
		}
	} else {
		schemaObj, err = openapi.ParseSchema(op.ResponseSchemaProxy)
		if err != nil {
			return fmt.Errorf("parsing response schema: %w", err)
		}
	}

	// Build template data
	data, err := BuildTemplateData(name, op, schemaObj, isJSONAPI, listOp)
	if err != nil {
		return fmt.Errorf("building template data: %w", err)
	}

	// Render the main generated file
	mainContent, err := engine.Render("main.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("rendering main template: %w", err)
	}

	// Render the hooks scaffold
	hooksContent, err := engine.Render("hooks.go.tmpl", data)
	if err != nil {
		return fmt.Errorf("rendering hooks template: %w", err)
	}

	if dryRun {
		fmt.Printf("=== %s (generated) ===\n%s\n", name, string(mainContent))
		fmt.Printf("=== %s (hooks) ===\n%s\n", name, string(hooksContent))
		return nil
	}

	// Write the generated file (always overwrite)
	genPath := filepath.Join(outputDir, fmt.Sprintf("data_source_datadog_%s_generated.go", name))
	if err := WriteGoFile(genPath, mainContent); err != nil {
		return fmt.Errorf("writing generated file: %w", err)
	}

	// Write the hooks file (only if it doesn't exist)
	hooksPath := filepath.Join(outputDir, fmt.Sprintf("data_source_datadog_%s_hooks.go", name))
	if err := WriteIfNotExists(hooksPath, hooksContent); err != nil {
		return fmt.Errorf("writing hooks file: %w", err)
	}

	return nil
}
