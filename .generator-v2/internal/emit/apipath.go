package emit

import (
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Terraform-to-API path correspondence
//
// A generated schema flattens the JSON:API envelope and normalizes every
// property name, so the attribute `background_color` may sit at
// `data.attributes.backgroundColor` in the body. The two are not derivable from
// one another — SnakeCase collapses several spellings onto one, which is
// exactly why model.Attribute.OpenAPIName exists — so the correspondence is
// recorded here while both are in hand.
//
// Matching the two by string suffix instead is wrong, not merely approximate:
// `background_color` never matches `backgroundColor`, and
// `twilio_messages_logs` never matches `twilio-messages-logs`, so a value is
// silently dropped from a generated configuration while the cassette still
// sends it.
//
// The index is keyed by the Terraform path the generated schema exposes, which
// is what a renderer walking that schema already has in hand. Keying it by the
// model attribute's own path would mean depending on how the view builder roots
// its walk, which differs between the record and response trees.
// ----------------------------------------------------------------------------

// envelopeMembers are the JSON:API members a generated schema flattens away:
// they exist in the body but never as Terraform attributes.
var envelopeMembers = map[string]bool{"data": true, "attributes": true}

// APIPathIndex maps each attribute's Terraform path to the dotted path it
// occupies in the API body.
//
// Three kinds of node contribute nothing to the Terraform path. The artifact
// root is a Terraform-side anchor with no counterpart. The envelope members the
// schema flattens away contribute to the API path only. And a synthetic union
// wrapper — the node a oneOf envelope introduces, carrying no OpenAPIName —
// contributes to neither: a declared body nests `authentication.password`,
// never `authentication.<variant>.password`, and the generated schema nests it
// the same way.
func APIPathIndex(attributes []*model.Attribute) map[string]string {
	index := map[string]string{}
	indexAPIPaths(attributes, "", "", index)
	return index
}

func indexAPIPaths(attributes []*model.Attribute, tfPrefix, apiPrefix string, index map[string]string) {
	for _, attr := range attributes {
		name := attr.OpenAPIName
		apiPath := apiPrefix
		if name != "" {
			apiPath = model.ChildPath(apiPrefix, name)
		}

		tfPath := tfPrefix
		switch {
		case envelopeMembers[name] && tfPrefix == "":
			// data and attributes are flattened away at the root only; a
			// property legitimately called "data" deeper in a body keeps its
			// own name.
		default:
			// A synthetic union wrapper lands here too: it carries no
			// OpenAPIName, so the API path does not advance past it, but the
			// generated schema does expose it as a block — a declared body
			// nests authentication.password while the configuration nests
			// authentication.<variant>.password.
			tfPath = model.ChildPath(tfPrefix, tfNameOf(attr.Path))
		}

		if tfPath != "" && apiPath != "" {
			index[tfPath] = apiPath
		}
		indexAPIPaths(attr.Children, tfPath, apiPath, index)
	}
}
