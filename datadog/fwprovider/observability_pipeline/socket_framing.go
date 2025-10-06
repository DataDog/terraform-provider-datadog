package observability_pipeline

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SocketFramingModel represents socket framing configuration
type SocketFramingModel struct {
	Method             types.String                          `tfsdk:"method"`
	CharacterDelimited *SocketFramingCharacterDelimitedModel `tfsdk:"character_delimited"`
}

// SocketFramingCharacterDelimitedModel represents character delimited framing
type SocketFramingCharacterDelimitedModel struct {
	Delimiter types.String `tfsdk:"delimiter"`
}
