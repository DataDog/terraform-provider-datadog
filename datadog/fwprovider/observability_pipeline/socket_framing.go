package observability_pipeline

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SocketFramingModel represents socket framing configuration
type SocketFramingModel struct {
	Method             types.String                           `tfsdk:"method"`
	CharacterDelimited []SocketFramingCharacterDelimitedModel `tfsdk:"character_delimited"`
}

// SocketFramingCharacterDelimitedModel represents character delimited framing
type SocketFramingCharacterDelimitedModel struct {
	Delimiter types.String `tfsdk:"delimiter"`
}

// SocketFramingValidator validates that character_delimited block is present if and only if method is "character_delimited"
type SocketFramingValidator struct{}

func (v SocketFramingValidator) Description(ctx context.Context) string {
	return "validates that character_delimited block is required when and only when method is 'character_delimited'"
}

func (v SocketFramingValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v SocketFramingValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()

	methodAttr, ok := attrs["method"]
	if !ok || methodAttr.IsNull() || methodAttr.IsUnknown() {
		return
	}
	method := methodAttr.(types.String).ValueString()

	charDelimitedAttr, hasCharDelimitedAttr := attrs["character_delimited"]
	hasCharacterDelimited := false
	if hasCharDelimitedAttr && !charDelimitedAttr.IsNull() {
		if list, ok := charDelimitedAttr.(types.List); ok {
			hasCharacterDelimited = len(list.Elements()) > 0
		}
	}

	if method == "character_delimited" && !hasCharacterDelimited {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Missing Required Block",
			"The 'character_delimited' block is required when 'method' is set to 'character_delimited'.",
		))
	}

	if method != "character_delimited" && hasCharacterDelimited {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid Block",
			"The 'character_delimited' block must not be specified when 'method' is not 'character_delimited'.",
		))
	}
}
