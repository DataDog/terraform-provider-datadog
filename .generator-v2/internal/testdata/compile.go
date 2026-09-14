// Package testdata holds the generator's end-to-end test infrastructure: the
// fixture catalogue, the golden-output assertions, and the compile gate below.
// The directory name makes the go tool skip it when expanding ./..., while an
// explicit import still resolves it normally.
package testdata

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/providermod"
)

// StagedFile places one generated file into the provider module for the
// duration of a compile.
type StagedFile struct {
	// ProviderPath is the destination, relative to the provider module root,
	// e.g. "datadog/fwprovider/resource_datadog_okta_account.go". Staging at
	// the artifact's real path compiles exactly the package a practitioner
	// would get, including one where the artifact replaces a hand-written file.
	ProviderPath string
	// SourcePath is the generated file on disk, as written by the generator.
	SourcePath string
}

// CompileResult reports what the Go compiler made of the staged files.
type CompileResult struct {
	// Diagnostics is the compiler's own output, empty when the build succeeded.
	// Returned rather than asserted on here, since a caller may be checking
	// that a build fails with a particular message.
	Diagnostics string
}

// OK reports whether the provider packages built with the staged files in
// place. Derived from Diagnostics rather than stored beside it, so the two
// cannot disagree.
func (r CompileResult) OK() bool { return r.Diagnostics == "" }

// Compile type-checks the staged files against the provider module with
// `go build -overlay`, which maps each staged path onto its generated content
// for one invocation: the code compiles as a real member of its provider
// package, resolves the SDK and plugin framework through the provider's go.mod,
// and writes nothing into the checkout. A control build of the same packages
// runs first, so an error — no provider module, no go tool, an unreadable
// staged file, a checkout that does not build on its own — means the gate could
// not run; a failure from the staged files is a non-OK CompileResult instead.
func Compile(files []StagedFile) (CompileResult, error) {
	if len(files) == 0 {
		return CompileResult{}, fmt.Errorf("compile gate: no files staged")
	}

	providerRoot, err := providermod.Root("")
	if err != nil {
		return CompileResult{}, fmt.Errorf("compile gate: %w", err)
	}

	packages, err := stagedPackages(files)
	if err != nil {
		return CompileResult{}, fmt.Errorf("compile gate: %w", err)
	}
	if err := baseline(providerRoot, packages); err != nil {
		return CompileResult{}, fmt.Errorf("compile gate: %w", err)
	}

	// The overlay file goes to a temp dir, never into the checkout.
	dir, err := os.MkdirTemp("", "tfgen-compile-overlay-")
	if err != nil {
		return CompileResult{}, fmt.Errorf("compile gate: create overlay dir: %w", err)
	}
	defer os.RemoveAll(dir)

	overlay, err := writeOverlay(dir, providerRoot, files)
	if err != nil {
		return CompileResult{}, fmt.Errorf("compile gate: %w", err)
	}

	out, err := build(providerRoot, packages, "-overlay", overlay)
	if err == nil {
		return CompileResult{}, nil
	}
	if out == "" {
		// The go tool failed without saying anything about the code — a
		// toolchain or environment fault, not a verdict on the artifact.
		return CompileResult{}, fmt.Errorf("compile gate: go build in %s: %w", providerRoot, err)
	}
	return CompileResult{Diagnostics: out}, nil
}

// stagedPackages reduces the staged files to the sorted set of package patterns
// to build, one per directory they land in.
func stagedPackages(files []StagedFile) ([]string, error) {
	seen := map[string]struct{}{}
	for _, f := range files {
		if f.ProviderPath == "" || f.SourcePath == "" {
			return nil, fmt.Errorf("staged file needs both a provider path and a source path, got %+v", f)
		}
		if filepath.IsAbs(f.ProviderPath) {
			return nil, fmt.Errorf("provider path must be relative to the provider module root, got %q", f.ProviderPath)
		}
		seen["./"+filepath.ToSlash(filepath.Dir(f.ProviderPath))] = struct{}{}
	}
	return slices.Sorted(maps.Keys(seen)), nil
}

// baselineDone memoizes the control build per (provider root, package set), so
// a run of many fixtures staging into the same package type-checks the provider
// once. Failures are cached too, so a broken checkout fails every fixture fast.
// The mutex also serialises concurrent first calls into one control build.
var (
	baselineMu   sync.Mutex
	baselineDone = map[string]error{}
)

// baseline proves the provider packages build before anything is staged, so a
// failure afterwards is attributable to the staged files.
func baseline(providerRoot string, packages []string) error {
	key := providerRoot + "\x00" + strings.Join(packages, "\x00")

	baselineMu.Lock()
	defer baselineMu.Unlock()
	if err, done := baselineDone[key]; done {
		return err
	}

	var err error
	if out, buildErr := build(providerRoot, packages); buildErr != nil {
		err = fmt.Errorf(
			"the provider checkout at %s does not build %s on its own, so it cannot judge generated code. "+
				"Populate the module cache first (go mod download in the provider module) and fix any unrelated build error: %w\n%s",
			providerRoot, strings.Join(packages, " "), buildErr, out)
	}
	baselineDone[key] = err
	return err
}

// writeOverlay materializes the go build overlay in dir: a JSON file mapping
// each staged provider path to the generated file that supplies its content.
func writeOverlay(dir, providerRoot string, files []StagedFile) (string, error) {
	replace := make(map[string]string, len(files))
	for _, f := range files {
		source, absErr := filepath.Abs(f.SourcePath)
		if absErr != nil {
			return "", fmt.Errorf("resolve staged source %s: %w", f.SourcePath, absErr)
		}
		// Stat it rather than trusting the path: go build reports a missing
		// overlay source as an error against the package, which the gate would
		// then misattribute to the artifact.
		if _, statErr := os.Stat(source); statErr != nil {
			return "", fmt.Errorf("staged source %s: %w", f.SourcePath, statErr)
		}
		replace[filepath.Join(providerRoot, f.ProviderPath)] = source
	}

	encoded, err := json.Marshal(struct {
		Replace map[string]string
	}{Replace: replace})
	if err != nil {
		return "", fmt.Errorf("encode overlay: %w", err)
	}

	path := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		return "", fmt.Errorf("write overlay: %w", err)
	}
	return path, nil
}

// build runs go build in the provider module and returns its combined output.
// GOPROXY=off keeps the gate hermetic: a dependency missing from the module
// cache fails loudly here instead of being fetched from the network mid-test.
func build(providerRoot string, packages []string, extraArgs ...string) (string, error) {
	args := append([]string{"build"}, extraArgs...)
	args = append(args, packages...)
	cmd := exec.Command("go", args...)
	cmd.Dir = providerRoot
	cmd.Env = append(os.Environ(), "GOPROXY=off")
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
