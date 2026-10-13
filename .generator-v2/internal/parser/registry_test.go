package parser

import (
	"errors"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	yaml "go.yaml.in/yaml/v4"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// tagExt builds the extensions map libopenapi hands back for a tag
// carrying a single vendor extension under key, whose value is the YAML mapping
// in body.
func tagExt(t *testing.T, key, body string) *orderedmap.Map[string, *yaml.Node] {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatalf("unmarshal registry body: %v", err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		t.Fatalf("expected a YAML document node with content, got kind %d", doc.Kind)
	}
	return orderedmap.ToOrderedMap(map[string]*yaml.Node{key: doc.Content[0]})
}

func decodeRegistry(t *testing.T, body string) []*model.TrackingFieldMetadata {
	t.Helper()
	metas, err := DecodeRegistry(tagExt(t, DefaultTrackingFieldName, body), DefaultTrackingFieldName)
	if err != nil {
		t.Fatalf("DecodeRegistry: %v", err)
	}
	return metas
}

func decodeRegistryErr(t *testing.T, body string) error {
	t.Helper()
	metas, err := DecodeRegistry(tagExt(t, DefaultTrackingFieldName, body), DefaultTrackingFieldName)
	if err == nil {
		t.Fatalf("DecodeRegistry: want error, got %d metadata", len(metas))
	}
	return err
}

// The full family: one entry declares three artifacts, and the block key is
// what supplies artifact_kind and cardinality — neither appears in the source.
const databricksRegistry = `
integration_databricks_account:
  resource:
    tf_description: Provides a Datadog Databricks integration account resource.
    group:
      create: CreateDatabricksIntegrationAccount
      read: GetDatabricksIntegrationAccount
      update: UpdateDatabricksIntegrationAccount
      delete: DeleteDatabricksIntegrationAccount
  datasource_singular:
    tf_description: Use this data source to retrieve an existing account.
    group:
      read: GetDatabricksIntegrationAccount
  datasource_list:
    tf_description: Use this data source to list accounts.
    group:
      read: ListDatabricksIntegrationAccounts
`

func TestDecodeRegistryOutOfScope(t *testing.T) {
	t.Run("nil extensions", func(t *testing.T) {
		got, err := DecodeRegistry(nil, DefaultTrackingFieldName)
		if got != nil || err != nil {
			t.Fatalf("got (%v, %v), want (nil, nil)", got, err)
		}
	})
	// A document annotated only the legacy way must read as having no registry,
	// which is what lets both shapes coexist during the migration.
	t.Run("different extension key present", func(t *testing.T) {
		got, err := DecodeRegistry(tagExt(t, "x-some-other-extension", "foo: bar\n"), DefaultTrackingFieldName)
		if got != nil || err != nil {
			t.Fatalf("got (%v, %v), want (nil, nil)", got, err)
		}
	})
}

func TestDecodeRegistryFullFamily(t *testing.T) {
	metas := decodeRegistry(t, databricksRegistry)
	if len(metas) != 3 {
		t.Fatalf("got %d artifacts, want 3", len(metas))
	}

	type got struct {
		block       string
		name        string
		kind        model.ArtifactKind
		cardinality model.Cardinality
	}
	want := []got{
		{BlockResource, "integration_databricks_account", model.ArtifactKindResource, ""},
		{BlockDatasourceSingular, "integration_databricks_account", model.ArtifactKindDataSource, model.CardinalitySingular},
		{BlockDatasourceList, "integration_databricks_accounts", model.ArtifactKindDataSource, model.CardinalityPlural},
	}
	for i, w := range want {
		m := metas[i]
		g := got{m.Block, m.ArtifactName, m.ArtifactKind, m.Cardinality}
		if g != w {
			t.Errorf("artifact %d: got %+v, want %+v", i, g, w)
		}
		if m.Family != "integration_databricks_account" {
			t.Errorf("artifact %d: family = %q, want the registry key", i, m.Family)
		}
		// Unset in the source, so every artifact must carry the schema default
		// rather than an empty strategy the emitters would have to interpret.
		if m.IdStrategy != model.IdStrategyDataID {
			t.Errorf("artifact %d: id_strategy = %q, want the %q default", i, m.IdStrategy, model.IdStrategyDataID)
		}
	}

	// The resource and the singular data source share a name deliberately:
	// Terraform keeps resources and data sources in separate namespaces.
	if metas[0].ArtifactName != metas[1].ArtifactName {
		t.Errorf("resource and singular data source should share the family name, got %q and %q",
			metas[0].ArtifactName, metas[1].ArtifactName)
	}
	if metas[2].Group == nil || metas[2].Group.Read != "ListDatabricksIntegrationAccounts" {
		t.Errorf("list group.read = %+v, want the list operationId", metas[2].Group)
	}
}

// Families are emitted in name order and blocks in a fixed order, so a run over
// an unchanged document is deterministic however the document was written.
func TestDecodeRegistryIsDeterministic(t *testing.T) {
	const body = `
zebra:
  datasource_list:
    group: {read: ListZebras}
  resource:
    group: {create: CreateZebra, read: GetZebra, delete: DeleteZebra}
alpha:
  datasource_singular:
    group: {read: GetAlpha}
`
	var first []string
	for range 5 {
		var addresses []string
		for _, m := range decodeRegistry(t, body) {
			addresses = append(addresses, m.Family+"."+m.Block)
		}
		if first == nil {
			first = addresses
			continue
		}
		if strings.Join(addresses, ",") != strings.Join(first, ",") {
			t.Fatalf("order varies between runs: %v then %v", first, addresses)
		}
	}
	want := "alpha.datasource_singular,zebra.resource,zebra.datasource_list"
	if got := strings.Join(first, ","); got != want {
		t.Errorf("order = %q, want %q", got, want)
	}
}

func TestDecodeRegistrySkip(t *testing.T) {
	t.Run("family skip drops every block", func(t *testing.T) {
		const body = `
thing:
  skip: true
  resource:
    group: {create: CreateThing, read: GetThing, delete: DeleteThing}
  datasource_singular:
    group: {read: GetThing}
`
		if metas := decodeRegistry(t, body); len(metas) != 0 {
			t.Fatalf("got %d artifacts, want 0", len(metas))
		}
	})
	t.Run("block skip drops only that block", func(t *testing.T) {
		const body = `
thing:
  resource:
    skip: true
    group: {create: CreateThing, read: GetThing, delete: DeleteThing}
  datasource_singular:
    group: {read: GetThing}
`
		metas := decodeRegistry(t, body)
		if len(metas) != 1 || metas[0].Block != BlockDatasourceSingular {
			t.Fatalf("got %+v, want only the singular data source", metas)
		}
	})
}

func TestDecodeRegistryArtifactNameResolution(t *testing.T) {
	t.Run("explicit override wins over derivation", func(t *testing.T) {
		const body = `
monitor_config_policy:
  datasource_list:
    artifact_name: monitor_config_policies_v2
    group: {read: ListPolicies}
`
		metas := decodeRegistry(t, body)
		if got := metas[0].ArtifactName; got != "monitor_config_policies_v2" {
			t.Errorf("name = %q, want the override", got)
		}
	})
	t.Run("override applies to a resource too", func(t *testing.T) {
		const body = `
thing:
  resource:
    artifact_name: other_thing
    group: {create: CreateThing, read: GetThing, delete: DeleteThing}
`
		if got := decodeRegistry(t, body)[0].ArtifactName; got != "other_thing" {
			t.Errorf("name = %q, want the override", got)
		}
	})
	// A mass noun inflects regularly rather than colliding, so a poor-reading
	// plural is a quality problem the override exists for, not an error. What is
	// an error is a derived name the contract would have rejected had an author
	// written it out.
	t.Run("derived name over the length limit is refused", func(t *testing.T) {
		body := "\n" + strings.Repeat("a", maxArtifactNameLen) + ":\n  datasource_list:\n    group: {read: ListThings}\n"
		err := decodeRegistryErr(t, body)
		var regErr *RegistryError
		if !errors.As(err, &regErr) {
			t.Fatalf("error %v is not a *RegistryError", err)
		}
		if regErr.Block != BlockDatasourceList {
			t.Errorf("block = %q, want %q", regErr.Block, BlockDatasourceList)
		}
		if !strings.Contains(err.Error(), "artifact_name") {
			t.Errorf("error should name the remedy, got: %v", err)
		}
	})
}

// The per-block group constraints are the contract's, so they fail at parse
// time with the address the author wrote rather than partway through a build.
func TestDecodeRegistryRejectsMalformed(t *testing.T) {
	cases := map[string]string{
		"unknown field in a block": `
thing:
  resource:
    artifact_kind: resource
    group: {create: CreateThing, read: GetThing, delete: DeleteThing}
`,
		"unknown block": `
thing:
  datasource_cursed:
    group: {read: GetThing}
`,
		"resource missing delete": `
thing:
  resource:
    group: {create: CreateThing, read: GetThing}
`,
		"resource declaring search": `
thing:
  resource:
    group: {create: CreateThing, read: GetThing, delete: DeleteThing, search: ListThings}
`,
		"singular with neither read nor search": `
thing:
  datasource_singular:
    group: {}
`,
		"list declaring search instead of read": `
thing:
  datasource_list:
    group: {search: ListThings}
`,
		"family name not snake_case": `
Thing:
  datasource_singular:
    group: {read: GetThing}
`,
		"empty family": `
thing: {}
`,
		"registry is not a mapping": `
- thing
`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			decodeRegistryErr(t, body)
		})
	}
}

func TestPluralize(t *testing.T) {
	// The left column is spelled as the provider already ships it, so a
	// regression here is a regression against a real Terraform address.
	cases := map[string]string{
		"integration_databricks_account":          "integration_databricks_accounts",
		"monitor_config_policy":                   "monitor_config_policies",
		"restriction_policy":                      "restriction_policies",
		"logs_index":                              "logs_indexes",
		"service_level_objective":                 "service_level_objectives",
		"api_key":                                 "api_keys",
		"team_membership":                         "team_memberships",
		"sensitive_data_scanner_standard_pattern": "sensitive_data_scanner_standard_patterns",
		"team":   "teams",
		"status": "statuses",
		// datastore is a live case: datadog-api-spec's actions_datastore family
		// relies on this derivation for the datadog_datastores list data source.
		"datastore":    "datastores",
		"powerpack":    "powerpacks",
		"logs_archive": "logs_archives",
		// A mass noun inflects regularly: legal, reads poorly, overridable.
		"metadata": "metadatas",
	}
	for in, want := range cases {
		if got := Pluralize(in); got != want {
			t.Errorf("Pluralize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTagRegistries(t *testing.T) {
	tag := func(name, body string) *base.Tag {
		return &base.Tag{Name: name, Extensions: tagExt(t, DefaultTrackingFieldName, body)}
	}
	t.Run("multiple tags and stable ordering", func(t *testing.T) {
		tags := []*base.Tag{
			tag("Zebra", "zebra:\n  datasource_list:\n    group: {read: ListZebras}\n"),
			tag("Things", "thing:\n  datasource_singular:\n    group: {read: GetThing}\n  datasource_list:\n    group: {read: ListThings}\n"),
		}
		metas, err := decodeTagRegistries(tags, DefaultTrackingFieldName)
		if err != nil {
			t.Fatal(err)
		}
		if len(metas) != 3 || metas[0].ArtifactName != "thing" || metas[1].ArtifactName != "things" || metas[2].ArtifactName != "zebras" {
			t.Fatalf("unexpected artifacts: %+v", metas)
		}
	})
	t.Run("duplicate family even when skipped", func(t *testing.T) {
		_, err := decodeTagRegistries([]*base.Tag{tag("First", "thing: {skip: true}"), tag("Second", "thing: {skip: true}")}, DefaultTrackingFieldName)
		if err == nil || !strings.Contains(err.Error(), "already declared") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("duplicate tag declaration", func(t *testing.T) {
		_, err := decodeTagRegistries([]*base.Tag{tag("Things", "thing: {skip: true}"), tag("Things", "other: {skip: true}")}, DefaultTrackingFieldName)
		if err == nil || !strings.Contains(err.Error(), "more than once") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("malformed registry names its tag", func(t *testing.T) {
		_, err := decodeTagRegistries([]*base.Tag{tag("Things", "[]")}, DefaultTrackingFieldName)
		if err == nil || !strings.Contains(err.Error(), `tag "Things"`) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("skip false alone is not an artifact", func(t *testing.T) {
		decodeRegistryErr(t, "thing: {skip: false}")
	})
}
