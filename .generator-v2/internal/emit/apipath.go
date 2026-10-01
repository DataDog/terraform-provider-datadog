package emit

import (
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// API path correspondence
//
// An attribute's Path is the Terraform path: SnakeCase-normalized and rooted at
// the artifact. The values a cassette scenario materializes are keyed by the
// path they occupy in the API body. The two are not derivable from one another
// — SnakeCase collapses several spellings onto one, which is exactly why
// model.Attribute.OpenAPIName exists — so the correspondence is recorded here
// while both are in hand, and read back as an exact lookup.
//
// Matching the two by string suffix instead is wrong, not merely approximate:
// `background_color` never matches `backgroundColor`, and `twilio_messages_logs`
// never matches `twilio-messages-logs`, so the value is silently dropped from a
// generated configuration while the cassette still sends it.
// ----------------------------------------------------------------------------

// apiPathIndex maps each attribute's Terraform path to the dotted path it
// occupies in the API body.
//
// Two kinds of node contribute nothing to the API path. The artifact root
// ("resource", "data_source") is a Terraform-side anchor with no counterpart.
// And a synthetic union wrapper — the node a oneOf envelope introduces, which
// carries no OpenAPIName of its own — is a generated Terraform block, not a
// JSON member: a declared body nests `authentication.password`, never
// `authentication.<variant>.password`.
func apiPathIndex(attributes []*model.Attribute) map[string]string {
	index := map[string]string{}
	indexAPIPaths(attributes, "", index)
	return index
}

func indexAPIPaths(attributes []*model.Attribute, prefix string, index map[string]string) {
	for _, attr := range attributes {
		path := prefix
		if attr.OpenAPIName != "" {
			path = model.ChildPath(prefix, attr.OpenAPIName)
		}
		// A node with no OpenAPIName passes its parent's path down unchanged,
		// so its children land where the body actually nests them.
		if path != "" {
			index[attr.Path] = path
		}
		indexAPIPaths(attr.Children, path, index)
	}
}

// trimAPIRoot drops the artifact-root segment an attribute path carries, so a
// caller comparing against an API-rooted key is not tripped by it.
func trimAPIRoot(path string) string {
	if _, rest, found := strings.Cut(path, "."); found {
		return rest
	}
	return path
}
