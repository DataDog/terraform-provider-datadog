package model

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The model package is decoupled from the OpenAPI input, with exactly one
// documented exception: Spec.Components keeps a libopenapi handle so schemas
// can be resolved lazily. Everything the cassette pipeline added must describe
// a scenario without retaining libopenapi objects, so a new import here would
// mean the raw document had leaked into the normalized model.
var _ = Describe("model package OpenAPI boundary", func() {
	const libopenapi = "github.com/pb33f/libopenapi"

	It("imports libopenapi only where Spec.Components requires it", func() {
		entries, err := os.ReadDir(".")
		Expect(err).To(Succeed())

		var importers []string
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
				continue
			}
			source, err := os.ReadFile(name)
			Expect(err).To(Succeed())
			if strings.Contains(string(source), libopenapi) {
				importers = append(importers, name)
			}
		}

		Expect(importers).To(ConsistOf("types.go"),
			"only types.go may reference libopenapi (for Spec.Components); a new importer means "+
				"the raw OpenAPI document leaked into the normalized model")
	})
})
