package cli

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

// runTfgen builds the root command exactly as Execute does, but with explicit
// args and errors returned rather than printed, so tests can assert on them.
func runTfgen(args ...string) error {
	flags := &globalFlags{}
	root := newRootCmd("test", flags)
	root.AddCommand(newGenerateCmd(flags))
	root.AddCommand(newVerifyCmd(flags))
	root.SetArgs(args)
	root.SilenceErrors = true
	return root.Execute()
}

// TestGenerateWiresMaxDepth proves the --max-depth flag value reaches LoadSpec:
// the same deep-but-acyclic spec loads at a high bound and fails at a low one.
// If the flag were ignored, both runs would behave identically.
func TestGenerateWiresMaxDepth(t *testing.T) {
	deep := filepath.Join("..", "testdata", "parser", "deep_chain.yaml")

	if err := runTfgen("generate", "--spec", deep, "--max-depth", "20"); err != nil {
		t.Errorf("--max-depth 20 should load the 8-deep chain, got: %v", err)
	}
	if err := runTfgen("generate", "--spec", deep, "--max-depth", "4"); err == nil {
		t.Error("--max-depth 4 should fail on the 8-deep chain, got nil")
	}
}

// TestGenerateSurfacesCycleError proves cycle detection is reachable through the
// command and surfaces the typed error.
func TestGenerateSurfacesCycleError(t *testing.T) {
	self := filepath.Join("..", "testdata", "parser", "cycle_self.yaml")

	err := runTfgen("generate", "--spec", self)
	var cycleErr *parser.RefCycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("error %v (%T) is not a *parser.RefCycleError", err, err)
	}
}
