package observability_pipeline

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BaseProcessor interface defines common fields that all processors have
// Used for both flatten (Get methods) and expand (Set methods) operations
type BaseProcessor interface {
	// Get methods for flatten (API -> Terraform)
	GetId() string
	GetEnabled() bool
	GetInclude() string
	GetDisplayName() string
	GetDisplayNameOk() (*string, bool)
	// Set methods for expand (Terraform -> API)
	SetId(string)
	SetEnabled(bool)
	SetInclude(string)
	SetDisplayName(string)
}

// BaseProcessorFields holds the common fields shared by all processors
type BaseProcessorFields struct {
	Id          string
	Enabled     bool
	Include     string
	DisplayName *string
}

// ApplyTo sets the common fields on any processor
func (c BaseProcessorFields) ApplyTo(proc BaseProcessor) {
	proc.SetId(c.Id)
	proc.SetEnabled(c.Enabled)
	proc.SetInclude(c.Include)
	if c.DisplayName != nil {
		proc.SetDisplayName(*c.DisplayName)
	}
}

// coreProcessorFields are the common processor attributes that are NOT processor types.
// Any other list attribute in the processor block is considered a processor type.
var coreProcessorFields = map[string]struct{}{
	"id":           {},
	"enabled":      {},
	"include":      {},
	"display_name": {},
}

// ExactlyOneProcessorValidator validates that exactly one processor type block is specified
type ExactlyOneProcessorValidator struct{}

func (v ExactlyOneProcessorValidator) Description(ctx context.Context) string {
	return "validates that exactly one processor type block is specified (e.g., filter, sample, quota)"
}

func (v ExactlyOneProcessorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ExactlyOneProcessorValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()

	var specifiedTypes []string
	for attrName, attr := range attrs {
		// Skip core processor fields
		if _, isCoreField := coreProcessorFields[attrName]; isCoreField {
			continue
		}

		if attr.IsNull() || attr.IsUnknown() {
			continue
		}

		// Check if the list has any elements (processor type blocks are list nested blocks)
		if list, ok := attr.(types.List); ok {
			if len(list.Elements()) > 0 {
				specifiedTypes = append(specifiedTypes, attrName)
			}
		}
	}

	if len(specifiedTypes) == 0 {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Missing Processor Type",
			"A processor block must specify exactly one processor type (e.g., filter, sample, quota, parse_json, etc.).",
		))
		return
	}

	if len(specifiedTypes) > 1 {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Multiple Processor Types Specified",
			fmt.Sprintf("A processor block must specify exactly one processor type, but found %d: %v. Each processor should only contain one type.", len(specifiedTypes), specifiedTypes),
		))
	}
}
