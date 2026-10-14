package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	yaml "go.yaml.in/yaml/v4"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/contracts"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Tag-scoped family registry
//
// The registry is the tag-level form of the tracking extension: a map of
// family name to the artifacts generated over that API. Each block becomes one
// model.TrackingFieldMetadata consumed by artifact generation.
//
// The block key carries what artifact_kind and cardinality carried before, and
// the map key carries artifact_name, so neither can be stated inconsistently
// across one family.
// ----------------------------------------------------------------------------

// Block names in the registry. The order is the order artifacts are emitted in,
// which keeps a run over an unchanged document deterministic.
const (
	BlockResource           = "resource"
	BlockDatasourceSingular = "datasource_singular"
	BlockDatasourceList     = "datasource_list"
)

// blockOrder fixes the emission order within one family.
var blockOrder = []string{BlockResource, BlockDatasourceSingular, BlockDatasourceList}

// compiledRegistrySchema compiles the embedded family-registry schema exactly
// once. A failure here indicts the committed
// contract rather than the document under test.
var compiledRegistrySchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compileEmbedded(contracts.FamilyRegistrySchema, "family-registry")
})

// RegistryError reports a malformed or unsatisfiable family registry. It names
// the address the author wrote — the family key and, when the fault is inside
// one, the block — rather than an OpenAPI path, because the registry lives at
// a Tag Object and mentions no path.
type RegistryError struct {
	Tag    string
	Family string
	Block  string
	Cause  error
}

func (e *RegistryError) Error() string {
	if e.Tag != "" {
		inner := *e
		inner.Tag = ""
		return fmt.Sprintf("tag %q: %s", e.Tag, inner.Error())
	}
	switch {
	case e.Family == "":
		return fmt.Sprintf("parser: invalid family registry on a tag: %v", e.Cause)
	case e.Block == "":
		return fmt.Sprintf("parser: invalid family registry entry %q: %v", e.Family, e.Cause)
	default:
		return fmt.Sprintf("parser: invalid family registry entry %q, block %q: %v", e.Family, e.Block, e.Cause)
	}
}

func (e *RegistryError) Unwrap() error { return e.Cause }

// familyDecl is one registry entry: up to three artifact blocks over one API.
type familyDecl struct {
	Skip               bool           `json:"skip,omitempty"`
	Resource           *artifactBlock `json:"resource,omitempty"`
	DatasourceSingular *artifactBlock `json:"datasource_singular,omitempty"`
	DatasourceList     *artifactBlock `json:"datasource_list,omitempty"`
}

// block returns the declaration under name, or nil when the family omits it.
func (f familyDecl) block(name string) *artifactBlock {
	switch name {
	case BlockResource:
		return f.Resource
	case BlockDatasourceSingular:
		return f.DatasourceSingular
	case BlockDatasourceList:
		return f.DatasourceList
	}
	return nil
}

// artifactBlock is one artifact's declaration. Every field is per-artifact by
// construction: the block is the only place they can be written.
type artifactBlock struct {
	ArtifactName  string                `json:"artifact_name,omitempty"`
	TfDescription string                `json:"tf_description,omitempty"`
	IdStrategy    model.IdStrategy      `json:"id_strategy,omitempty"`
	Overwrites    string                `json:"overwrites,omitempty"`
	Skip          bool                  `json:"skip,omitempty"`
	Group         *model.OperationGroup `json:"group,omitempty"`
}

// DecodeRegistry validates and decodes a tag-scoped family registry into artifact metadata.
// An absent registry declares no artifacts.
//
// Results are ordered by family name, then by block (resource, singular, list),
// so a run over an unchanged document is deterministic regardless of how the
// document was written. A family or block setting skip:true is omitted, the
// documented opt-out equivalent to deleting it.
func DecodeRegistry(ext *orderedmap.Map[string, *yaml.Node], extensionName string) ([]*model.TrackingFieldMetadata, error) {
	if ext == nil {
		return nil, nil
	}
	node := ext.GetOrZero(extensionName)
	if node == nil {
		return nil, nil // absent → the document declares no registry
	}

	var raw any
	if err := node.Decode(&raw); err != nil {
		return nil, &RegistryError{Cause: fmt.Errorf("decoding registry node: %w", err)}
	}
	sch, err := compiledRegistrySchema()
	if err != nil {
		return nil, err // committed-contract bug, not author-attributable
	}
	if err := sch.Validate(raw); err != nil {
		return nil, &RegistryError{Cause: err}
	}

	// Bridge YAML→JSON rather than decoding the node straight into the structs:
	// the declarations carry json tags (the snake_case keys), and go.yaml.in/yaml
	// honors yaml tags, which they do not declare.
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, &RegistryError{Cause: fmt.Errorf("re-encoding registry: %w", err)}
	}
	var registry map[string]familyDecl
	if err := json.Unmarshal(jsonBytes, &registry); err != nil {
		return nil, &RegistryError{Cause: fmt.Errorf("decoding registry into declarations: %w", err)}
	}

	families := make([]string, 0, len(registry))
	for name := range registry {
		families = append(families, name)
	}
	sort.Strings(families)

	var metas []*model.TrackingFieldMetadata
	for _, family := range families {
		decl := registry[family]
		if decl.Skip {
			continue
		}
		for _, blockName := range blockOrder {
			block := decl.block(blockName)
			if block == nil || block.Skip {
				continue
			}
			meta, err := desugarBlock(family, blockName, block)
			if err != nil {
				return nil, err
			}
			metas = append(metas, meta)
		}
	}
	return metas, nil
}

// desugarBlock projects one artifact block onto the tracking metadata the rest
// of the generator consumes, synthesizing artifact_kind and cardinality from
// the block key and resolving the artifact's name.
func desugarBlock(family, blockName string, block *artifactBlock) (*model.TrackingFieldMetadata, error) {
	meta := &model.TrackingFieldMetadata{
		Family:        family,
		Block:         blockName,
		TfDescription: block.TfDescription,
		Group:         block.Group,
		IdStrategy:    block.IdStrategy,
		Overwrites:    block.Overwrites,
	}
	switch blockName {
	case BlockResource:
		meta.ArtifactKind = model.ArtifactKindResource
	case BlockDatasourceSingular:
		meta.ArtifactKind = model.ArtifactKindDataSource
		meta.Cardinality = model.CardinalitySingular
	case BlockDatasourceList:
		meta.ArtifactKind = model.ArtifactKindDataSource
		meta.Cardinality = model.CardinalityPlural
	default:
		// Unreachable: the schema's additionalProperties:false admits no other
		// block, so this guards a future block added to the contract alone.
		return nil, &RegistryError{Family: family, Block: blockName, Cause: fmt.Errorf("unknown block")}
	}
	if meta.IdStrategy == "" {
		meta.IdStrategy = model.IdStrategyDataID // schema default "data.id"
	}

	name, err := resolveArtifactName(family, blockName, block)
	if err != nil {
		return nil, err
	}
	meta.ArtifactName = name
	return meta, nil
}

// resolveArtifactName returns the Terraform name one block takes: its explicit
// artifact_name when set, the family name for a resource or singular data
// source, and the pluralized family name for a list.
//
// A derived name is validated against the same constraints the contract puts on
// a written one, because it becomes a Terraform address and a filename either
// way — and inflection is the one way to arrive at a name no author reviewed.
//
// What it deliberately does not try to detect is a bad-but-legal plural. A mass
// noun inflects regularly here ("metadata" → "metadatas"), which reads poorly
// but cannot collide with anything, since every rule either appends or replaces
// a suffix; that is what artifact_name overrides. A derived name that does
// collide — with another family, or with a name written by hand — is caught
// spec-wide by CheckDuplicateArtifactNames, which sees every artifact rather
// than one family at a time.
func resolveArtifactName(family, blockName string, block *artifactBlock) (string, error) {
	if block.ArtifactName != "" {
		return block.ArtifactName, nil
	}
	if blockName != BlockDatasourceList {
		return family, nil
	}
	plural := Pluralize(family)
	fail := func(why string) error {
		return &RegistryError{Family: family, Block: blockName, Cause: fmt.Errorf(
			"cannot derive a list name from family %q (%s): set artifact_name on this block", family, why)}
	}
	switch {
	case !artifactNamePattern.MatchString(plural):
		return "", fail(fmt.Sprintf("the derived name %q is not lowercase snake_case", plural))
	case len(plural) > maxArtifactNameLen:
		return "", fail(fmt.Sprintf("the derived name %q exceeds the %d-character limit", plural, maxArtifactNameLen))
	}
	return plural, nil
}

// Pluralize returns the plural of a snake_case artifact name, inflecting only
// its last word. The rules are deliberately the three that cover the names the
// provider already ships — logs_index → logs_indexes (not "indices"),
// monitor_config_policy → monitor_config_policies, api_key → api_keys — rather
// than a general-purpose inflector: a derived name is validated by the caller
// and an author can always override it, so a wrong guess costs a diagnostic
// rather than a wrong artifact.
func Pluralize(name string) string {
	head, word := "", name
	if i := strings.LastIndex(name, "_"); i >= 0 {
		head, word = name[:i+1], name[i+1:]
	}
	return head + pluralizeWord(word)
}

func pluralizeWord(word string) string {
	if word == "" {
		return word
	}
	// A consonant before a final y turns it into -ies; a vowel keeps it (key →
	// keys), which is why the preceding letter is tested rather than the last.
	if strings.HasSuffix(word, "y") && len(word) >= 2 && !isVowel(word[len(word)-2]) {
		return word[:len(word)-1] + "ies"
	}
	switch last := word[len(word)-1]; {
	case last == 's', last == 'x', last == 'z',
		strings.HasSuffix(word, "ch"), strings.HasSuffix(word, "sh"):
		return word + "es"
	}
	return word + "s"
}

func isVowel(b byte) bool {
	return bytes.IndexByte([]byte("aeiou"), b) >= 0
}

// decodeTagRegistries reads each tag's declaration without merging extension
// mappings. Family names remain unique across the spec, including skipped ones.
func decodeTagRegistries(tags []*base.Tag, extensionName string) ([]*model.TrackingFieldMetadata, error) {
	seenTags := map[string]bool{}
	seenFamilies := map[string]string{}
	var all []*model.TrackingFieldMetadata
	for _, tag := range tags {
		if tag == nil || tag.Extensions == nil || tag.Extensions.GetOrZero(extensionName) == nil {
			continue
		}
		if seenTags[tag.Name] {
			return nil, &RegistryError{Tag: tag.Name, Cause: fmt.Errorf("%s is declared more than once for this tag", extensionName)}
		}
		seenTags[tag.Name] = true
		metas, err := DecodeRegistry(tag.Extensions, extensionName)
		if err != nil {
			return nil, &RegistryError{Tag: tag.Name, Cause: err}
		}
		// DecodeRegistry has already checked that this is a mapping with string keys.
		var families map[string]any
		if err := tag.Extensions.GetOrZero(extensionName).Decode(&families); err != nil {
			return nil, &RegistryError{Tag: tag.Name, Cause: err}
		}
		names := make([]string, 0, len(families))
		for family := range families {
			names = append(names, family)
		}
		sort.Strings(names)
		for _, family := range names {
			if owner, exists := seenFamilies[family]; exists {
				return nil, &RegistryError{Tag: tag.Name, Family: family, Cause: fmt.Errorf("family is already declared on tag %q", owner)}
			}
			seenFamilies[family] = tag.Name
		}
		all = append(all, metas...)
	}
	// Preserve output order if authors reorder tags in the source document.
	sort.SliceStable(all, func(i, j int) bool { return all[i].Family < all[j].Family })
	return all, nil
}
