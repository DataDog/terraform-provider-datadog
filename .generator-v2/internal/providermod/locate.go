// Package providermod finds the terraform-provider-datadog module on disk and
// queries it about the pinned datadog-api-client-go. This generator is a
// separate module requiring neither, so anything needing the provider's own
// dependency graph — resolving the datadogV2 package, compiling a generated
// artifact against it — must locate that checkout first.
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
	sdkModule  = "github.com/DataDog/datadog-api-client-go/v2"
	sdkPackage = sdkModule + "/api/datadogV2"
)

// moduleDecl is one of the two markers a candidate go.mod must carry; the
// sdkModule requirement is the other. Both are needed because this generator's
// own go.mod declares "module <provider path>/generator", which contains
// moduleDecl as a prefix — only the SDK requirement tells the two apart.
const moduleDecl = "module github.com/terraform-providers/terraform-provider-datadog"

// Root returns the provider module's root directory, searching upward from hint
// (which may be empty) and then from the working directory, so a caller
// rendering into a temporary output root still resolves the real checkout.
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
			bytes.Contains(goMod, []byte(sdkModule)) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// SDKPackageDir returns the directory of the pinned datadogV2 package, as
// resolved by the provider module found from hint.
func SDKPackageDir(hint string) (string, error) {
	return goList(hint, "-f={{.Dir}}", sdkPackage)
}

// SDKModuleDir returns the root of the pinned datadog-api-client-go module,
// where the SDK's own bundled generator lives.
func SDKModuleDir(hint string) (string, error) {
	return goList(hint, "-m", "-f={{.Dir}}", sdkModule)
}

// goList runs `go list` with cmd.Dir at the provider module root and returns
// its trimmed single-line output, erroring on an empty result. Running it there
// keeps the query off the generator module, which requires no SDK.
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
