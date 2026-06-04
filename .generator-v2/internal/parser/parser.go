package parser

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// DefaultMaxDepth bounds recursive $ref expansion when no override is supplied.
const DefaultMaxDepth = 8

// Option configures LoadSpec.
type Option func(*loadConfig)

type loadConfig struct {
	maxDepth int
}

// WithMaxDepth sets the recursive $ref expansion limit — the value carried by
// the --max-depth CLI flag. A value <= 0 disables the bound.
func WithMaxDepth(n int) Option {
	return func(c *loadConfig) { c.maxDepth = n }
}

// LoadSpec reads and parses the OpenAPI v3 specification at path and projects
// it into the generator's internal model. Every operation across all paths and
// methods is enumerated and the resulting slice is sorted by (path, method) so
// downstream iteration — and generated output — is deterministic.
//
// Before enumerating, LoadSpec resolves the component schema graph and fails
// fast: a circular $ref returns a typed *RefCycleError naming the offending
// $ref, and a $ref chain deeper than the --max-depth bound (WithMaxDepth,
// default DefaultMaxDepth) returns an error. A spec that cannot be resolved must
// not silently produce partial output.
//
// LoadSpec populates Path, Method, OperationId and Tag on each Operation;
// Tracking, RequestSchema and ResponseSchema are filled by later passes.
func LoadSpec(path string, opts ...Option) (*model.Spec, error) {
	cfg := loadConfig{maxDepth: DefaultMaxDepth}
	for _, opt := range opts {
		opt(&cfg)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec %q: %w", path, err)
	}

	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, fmt.Errorf("parsing spec %q: %w", path, err)
	}

	v3doc, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("building OpenAPI v3 model for %q: %w", path, err)
	}

	spec := &model.Spec{
		Source:     path,
		Components: v3doc.Model.Components,
	}

	// Fail fast on an unresolvable schema graph before enumerating anything:
	// circular $refs surface as a typed *RefCycleError, and expansion past
	// --max-depth as a depth error (contracts/cli.md).
	cycles, err := DetectComponentRefCycles(spec.Components, cfg.maxDepth)
	if err != nil {
		return nil, err
	}
	if len(cycles) > 0 {
		return nil, &RefCycleError{Cycles: cycles}
	}

	if paths := v3doc.Model.Paths; paths != nil && paths.PathItems != nil {
		for opPath, item := range paths.PathItems.FromOldest() {
			if item == nil {
				continue
			}
			for method, op := range item.GetOperations().FromOldest() {
				if op == nil {
					continue
				}
				spec.Operations = append(spec.Operations, &model.Operation{
					Path:        opPath,
					Method:      strings.ToUpper(method),
					OperationId: op.OperationId,
					Tag:         firstTag(op.Tags),
				})
			}
		}
	}

	sort.Slice(spec.Operations, func(i, j int) bool {
		a, b := spec.Operations[i], spec.Operations[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Method < b.Method
	})

	return spec, nil
}

// firstTag returns the operation's first OpenAPI tag, or "" when untagged.
// The first tag is what the client generator keys package selection on so
// we must do the same.
func firstTag(tags []string) string {
	if len(tags) > 0 {
		return tags[0]
	}
	return ""
}
