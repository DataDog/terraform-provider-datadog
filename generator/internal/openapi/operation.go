package openapi

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Parameter represents a parsed API parameter (path or query).
type Parameter struct {
	Name        string
	Description string
	Required    bool
	Schema      *base.SchemaProxy
}

// ParsedOperation holds extracted details from an OpenAPI operation.
type ParsedOperation struct {
	Name                string
	OperationID         string
	Tag                 string
	Description         string
	Path                string
	Method              string
	PathParams          []Parameter
	QueryParams         []Parameter
	ResponseSchemaProxy *base.SchemaProxy
	ResponseTypeName    string // Schema name from $ref (e.g., "TeamResponse")
}

// ExtractOperation looks up an operation by API path and HTTP method
// and extracts its metadata, parameters, and response schema proxy.
func ExtractOperation(doc *v3high.Document, apiPath, method string) (*ParsedOperation, error) {
	if doc.Paths == nil || doc.Paths.PathItems == nil {
		return nil, fmt.Errorf("spec has no paths defined")
	}

	pathItem, ok := doc.Paths.PathItems.Get(apiPath)
	if !ok {
		return nil, fmt.Errorf("path %q not found in spec", apiPath)
	}

	op := getOperationByMethod(pathItem, method)
	if op == nil {
		return nil, fmt.Errorf("method %q not found for path %q", method, apiPath)
	}

	parsed := &ParsedOperation{
		OperationID: op.OperationId,
		Description: op.Description,
		Path:        apiPath,
		Method:      strings.ToLower(method),
	}

	if len(op.Tags) > 0 {
		parsed.Tag = op.Tags[0]
	}

	for _, param := range op.Parameters {
		p := Parameter{
			Name:        param.Name,
			Description: param.Description,
			Required:    param.Required != nil && *param.Required,
			Schema:      param.Schema,
		}
		switch param.In {
		case "path":
			parsed.PathParams = append(parsed.PathParams, p)
		case "query":
			parsed.QueryParams = append(parsed.QueryParams, p)
		}
	}

	responseProxy, err := extractResponseSchema(op)
	if err != nil {
		return nil, fmt.Errorf("extracting response schema for %s %s: %w", method, apiPath, err)
	}
	parsed.ResponseSchemaProxy = responseProxy

	// Extract response type name from $ref
	parsed.ResponseTypeName = extractRefName(responseProxy)

	return parsed, nil
}

// extractRefName extracts the schema name from a SchemaProxy's $ref string.
// Returns an empty string if no $ref is present (inline schema).
func extractRefName(proxy *base.SchemaProxy) string {
	ref := proxy.GetReference()
	if ref == "" {
		return ""
	}
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

// IsFilterParam returns true when the parameter name starts with "filter[".
func IsFilterParam(param Parameter) bool {
	return strings.HasPrefix(param.Name, "filter[")
}

// IsExcludedParam returns true for pagination, fields, sort, and include params.
func IsExcludedParam(param Parameter) bool {
	name := strings.ToLower(param.Name)
	return strings.HasPrefix(name, "page[") ||
		strings.HasPrefix(name, "fields[") ||
		name == "sort" ||
		name == "include"
}

// getOperationByMethod returns the operation for the given HTTP method.
func getOperationByMethod(pi *v3high.PathItem, method string) *v3high.Operation {
	switch strings.ToLower(method) {
	case "get":
		return pi.Get
	case "post":
		return pi.Post
	case "put":
		return pi.Put
	case "patch":
		return pi.Patch
	case "delete":
		return pi.Delete
	default:
		return nil
	}
}

// extractResponseSchema extracts the schema proxy from the 200 response's
// application/json content type.
func extractResponseSchema(op *v3high.Operation) (*base.SchemaProxy, error) {
	if op.Responses == nil || op.Responses.Codes == nil {
		return nil, fmt.Errorf("no responses defined")
	}

	resp, ok := op.Responses.Codes.Get("200")
	if !ok {
		return nil, fmt.Errorf("no 200 response defined")
	}

	if resp.Content == nil {
		return nil, fmt.Errorf("200 response has no content")
	}

	mt, ok := resp.Content.Get("application/json")
	if !ok {
		return nil, fmt.Errorf("200 response has no application/json content")
	}

	if mt.Schema == nil {
		return nil, fmt.Errorf("200 response application/json has no schema")
	}

	return mt.Schema, nil
}
