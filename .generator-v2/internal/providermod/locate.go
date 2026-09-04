// Package providermod locates the terraform-provider-datadog module this
// generator targets.
//
// The generator is a separate module rooted at .generator-v2/ and deliberately
// requires neither the provider nor the pinned datadog-api-client-go (plan.md's
// dependency-placement row; FR-005a's rationale depends on keeping it that
// way). So everything that has to reach the provider's own dependency graph —
// resolving the pinned datadogV2 package to corroborate a derived SDK binding,
// or compiling a generated artifact against it (FR-019) — must first find that
// module on disk. One package owns that lookup, so the generate path and the
// test gate cannot end up disagreeing about which checkout they mean.
package providermod

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// sdkPackage is the pinned SDK's versioned API package, the one every generated
// artifact imports.
const (
	sdkPackage = "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	sdkModule  = "github.com/DataDog/datadog-api-client-go/v2"
)

// moduleDecl and sdkRequirement are the two markers a candidate go.mod must
// carry, and both are load-bearing. The generator's own go.mod declares
// "module <provider path>/generator", which contains moduleDecl as a prefix, so
// a moduleDecl-only walk would stop on .generator-v2 — the nearest go.mod to
// every caller in this module. The SDK requirement is what tells the two apart,
// and it is worth checking on its own account too: a checkout that declares the
// module without requiring the SDK cannot build a generated artifact anyway.
const (
	moduleDecl     = "module github.com/terraform-providers/terraform-provider-datadog"
	sdkRequirement = "github.com/DataDog/datadog-api-client-go/v2"
)

// Root returns the provider module's root directory. It searches upward from
// hint first, then from the working directory, so a caller that renders into a
// temporary output root still resolves against the checkout it was invoked
// from. hint may be empty.
func Root(hint string) (string, error) {
	if hint != "" {
		absoluteHint, err := filepath.Abs(hint)
		if err != nil {
			return "", fmt.Errorf("resolve %s: %w", hint, err)
		}
		if root, ok := findUp(absoluteHint); ok {
			return root, nil
		}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	if root, ok := findUp(workingDir); ok {
		return root, nil
	}
	return "", fmt.Errorf("could not find the terraform-provider-datadog module containing the pinned datadog-api-client-go dependency")
}

// findUp looks for the provider's go.mod at dir and each of its ancestors.
func findUp(dir string) (string, bool) {
	for {
		goMod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil &&
			bytes.Contains(goMod, []byte(moduleDecl)) &&
			bytes.Contains(goMod, []byte(sdkRequirement)) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// SDKPackageDir asks the provider module selected by hint where the pinned
// datadogV2 package is unpacked. It is resolved through that module rather than
// this one precisely because this one does not require the SDK.
func SDKPackageDir(hint string) (string, error) {
	return goList(hint, "-f={{.Dir}}", sdkPackage)
}

// SDKModuleDir is SDKPackageDir's module-level counterpart: the root of the
// pinned datadog-api-client-go checkout, which is where the SDK's own bundled
// generator lives (the source FR-005a permits reading to corroborate a
// derivation).
func SDKModuleDir(hint string) (string, error) {
	return goList(hint, "-m", "-f={{.Dir}}", sdkModule)
}

// goList runs `go list` inside the provider module and returns its single-line
// output. Every query about the pinned SDK goes through here, so none of them
// can accidentally resolve against the generator module — which requires no SDK
// and would answer differently.
func goList(hint string, args ...string) (string, error) {
	moduleRoot, err := Root(hint)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("go", append([]string{"list"}, args...)...)
	cmd.Dir = moduleRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go list %s from %s: %w: %s",
			strings.Join(args, " "), moduleRoot, err, strings.TrimSpace(stderr.String()))
	}
	dir := strings.TrimSpace(string(output))
	if dir == "" {
		return "", fmt.Errorf("go list %s from %s returned an empty directory", strings.Join(args, " "), moduleRoot)
	}
	return dir, nil
}
