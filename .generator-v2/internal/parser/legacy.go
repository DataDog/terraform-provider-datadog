package parser

import (
	"fmt"
	"strings"

	yaml "go.yaml.in/yaml/v4"
)

// legacySchemaWarnings inspects source nodes rather than normalized schemas so
// ignored annotations are visible even when no family references their schema.
// It never prints extension values, which may contain credentials.
func legacySchemaWarnings(root *yaml.Node, extension string) []string {
	var warnings []string
	seen := map[*yaml.Node]bool{}
	var walkDocument, walkSchema func(*yaml.Node)
	visit := func(node *yaml.Node) *yaml.Node {
		for node != nil && node.Kind == yaml.AliasNode {
			if seen[node] {
				return nil
			}
			seen[node] = true
			node = node.Alias
		}
		if node == nil || seen[node] {
			return nil
		}
		seen[node] = true
		return node
	}
	walkSchema = func(node *yaml.Node) {
		node = visit(node)
		if node == nil || node.Kind != yaml.MappingNode {
			return
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			if key.Value == extension {
				if value.Kind == yaml.MappingNode {
					for j := 0; j+1 < len(value.Content); j += 2 {
						if value.Content[j].Value == "sensitive" {
							warnings = append(warnings, fmt.Sprintf("legacy schema %s.sensitive at line %d is ignored; sensitivity comes only from writeOnly and x-secret", extension, key.Line))
						}
					}
				}
				continue
			}
			switch key.Value {
			case "properties", "patternProperties", "dependentSchemas", "$defs":
				if value.Kind == yaml.MappingNode {
					for j := 1; j < len(value.Content); j += 2 {
						walkSchema(value.Content[j])
					}
				}
			case "allOf", "anyOf", "oneOf", "prefixItems":
				if value.Kind == yaml.SequenceNode {
					for _, child := range value.Content {
						walkSchema(child)
					}
				}
			case "items", "additionalProperties", "not", "contains", "propertyNames", "unevaluatedProperties", "if", "then", "else":
				walkSchema(value)
			}
		}
	}
	walkDocument = func(node *yaml.Node) {
		node = visit(node)
		if node == nil {
			return
		}
		if node.Kind != yaml.MappingNode {
			for _, child := range node.Content {
				walkDocument(child)
			}
			return
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			key, value := node.Content[i].Value, node.Content[i+1]
			switch {
			case key == "schema":
				walkSchema(value)
			case key == "schemas":
				if value.Kind == yaml.MappingNode {
					for j := 1; j < len(value.Content); j += 2 {
						walkSchema(value.Content[j])
					}
				}
			case strings.HasPrefix(strings.ToLower(key), "x-"), key == "example", key == "examples", key == "default", key == "enum":
				// These values are arbitrary user data, not OpenAPI schema declarations.
			default:
				walkDocument(value)
			}
		}
	}
	walkDocument(root)
	return warnings
}
