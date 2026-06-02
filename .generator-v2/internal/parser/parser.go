package parser

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// LoadSpec reads and parses the OpenAPI v3 specification at path and projects
// it into the generator's internal model. Every operation across all paths and
// methods is enumerated and the resulting slice is sorted by (path, method) so
// downstream iteration and generated output is deterministic.
//
// LoadSpec populates only what it can read straight from the document: Path,
// Method, OperationId and Tag.
func LoadSpec(path string) (*model.Spec, error) {
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
