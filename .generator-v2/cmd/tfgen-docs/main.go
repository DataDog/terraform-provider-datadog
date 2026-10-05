// Command tfgen-docs renders the generated half of .generator-v2/docs/.
//
// These reference pages are projections of something that already exists in the
// repository rather than prose a human keeps in step:
//
//	docs/reference/cli.md         ← the cobra command tree in internal/cli/
//	docs/reference/annotation.md  ← internal/contracts/tracking-field.schema.json
//
// The run report has no machine-readable contract yet. When one lands in
// internal/contracts/, add a page here rather than hand-writing its field table:
// renderSchemaPage is already generic over any of these schemas.
//
// This exists because the hand-maintained versions drifted. An annotation field
// was documented for seven weeks after it failed to merge, and the published
// schema matched neither the provider's contract nor the transformer's model.
// A generated page cannot describe a field the schema does not have.
//
// Usage:
//
//	tfgen-docs --out docs            # write the pages
//	tfgen-docs --out docs --check    # exit 3 if any page is stale
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/cli"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/contracts"
)

// exitStale mirrors tfgen generate --check: 3 means "a generated file would
// change", which lets CI distinguish staleness from a real failure.
const exitStale = 3

type page struct {
	path   string
	render func() (string, error)
}

func main() {
	out := flag.String("out", "docs", "documentation root to write into")
	check := flag.Bool("check", false, "report whether any page is stale without writing (exit 3 if so)")
	flag.Parse()

	pages := []page{
		{
			path: filepath.Join("reference", "cli.md"),
			render: func() (string, error) {
				return renderCLI(cli.RootCommand("dev")), nil
			},
		},
		{
			path: filepath.Join("reference", "annotation.md"),
			render: func() (string, error) {
				return renderAnnotation(contracts.TrackingFieldSchema)
			},
		},
	}

	var stale []string
	for _, p := range pages {
		body, err := p.render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "tfgen-docs: rendering %s: %v\n", p.path, err)
			os.Exit(1)
		}
		full := filepath.Join(*out, p.path)

		existing, readErr := os.ReadFile(full)
		unchanged := readErr == nil && bytes.Equal(existing, []byte(body))
		if unchanged {
			continue
		}
		if *check {
			stale = append(stale, full)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "tfgen-docs: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "tfgen-docs: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", full)
	}

	if len(stale) > 0 {
		fmt.Fprintf(os.Stderr, "tfgen-docs: stale generated docs:\n")
		for _, s := range stale {
			fmt.Fprintf(os.Stderr, "  %s\n", s)
		}
		fmt.Fprintf(os.Stderr, "run `make tfgen-docs` and commit the result\n")
		os.Exit(exitStale)
	}
}

func renderAnnotation(raw []byte) (string, error) {
	doc, err := newSchemaDoc(raw)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	writeFrontMatter(&b, frontMatter{
		Title:            "Annotation reference: `x-datadog-tf-generator`",
		ConfluenceParent: "6733104627",
		Audience:         "anyone annotating an OpenAPI operation to produce a Terraform artifact.",
		Source:           "`internal/contracts/tracking-field.schema.json`",
		Generator:        "cmd/tfgen-docs",
	})
	b.WriteString("The extension opts an OpenAPI operation into Terraform generation. " +
		"tfgen validates every occurrence against the schema this page is rendered from, " +
		"so the fields below are exactly the fields accepted — there are no undocumented ones, " +
		"and anything missing here will fail the run.\n\n")
	b.WriteString("For how to choose between the shapes, see " +
		"[Generating an artifact](../generating/README.md). " +
		"For what the generator can currently do with them, see [Scope](scope.md).\n\n")
	doc.renderBody(&b)
	return b.String(), nil
}

type frontMatter struct {
	Title string
	// ConfluenceParent is the Confluence page these docs are created under:
	// the entry point, docs/README.md, which replaces the old v2 index.
	ConfluenceParent string
	Audience         string
	Source           string
	Generator        string
}

// ddocBlock renders the Confluence-mirroring metadata every page in docs/
// carries, generated and hand-written alike. ddoc is configured per file —
// there is no repository-level config — and its content prefilter requires a
// closed leading YAML block with a column-zero `ddoc:` key, so this has to come
// before the generated-code marker.
//
// Only docs/README.md carries a confluence_id: it replaces the old "Terraform
// Generator v2" index in place. Every other page, these two included, is new and
// is created under it, so they carry confluence_parent instead. The pages they
// supersede ("CLI contract", "OAS extension contract") are deleted separately —
// ddoc never deletes a Confluence page.
func ddocBlock(parentID string) string {
	return fmt.Sprintf(`---
ddoc:
  confluence_space: "API"
  confluence_parent: %q
---

`, parentID)
}

// writeFrontMatter stamps the ddoc metadata and the provenance header every
// generated page carries. It deliberately does not include a timestamp: a date
// would make every regeneration a diff and defeat the --check gate.
func writeFrontMatter(b *strings.Builder, fm frontMatter) {
	b.WriteString(ddocBlock(fm.ConfluenceParent))
	fmt.Fprintf(b, "<!-- Code generated by %s. DO NOT EDIT. -->\n", fm.Generator)
	fmt.Fprintf(b, "<!-- Source of truth: %s -->\n\n", stripBackticks(fm.Source))
	fmt.Fprintf(b, "# %s\n\n", fm.Title)
	fmt.Fprintf(b, "**Who this is for:** %s\n\n", fm.Audience)
	fmt.Fprintf(b, "> Generated from %s by `%s`. Edit that source and run `make tfgen-docs`; "+
		"changes made here are overwritten.\n\n", fm.Source, fm.Generator)
}

func stripBackticks(s string) string {
	return strings.ReplaceAll(s, "`", "")
}
