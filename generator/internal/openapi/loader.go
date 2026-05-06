package openapi

import (
	"fmt"
	"os"

	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// LoadSpec loads an OpenAPI 3.x spec from a YAML or JSON file and returns the
// high-level v3 document model.
func LoadSpec(path string) (*libopenapi.DocumentModel[v3high.Document], error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec file %s: %w", path, err)
	}

	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, fmt.Errorf("parsing spec file %s: %w", path, err)
	}

	model, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("building v3 model for %s: %w", path, err)
	}

	return model, nil
}
