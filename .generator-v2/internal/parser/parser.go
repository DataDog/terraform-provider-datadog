package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// DefaultMaxDepth bounds recursive $ref expansion when no override is supplied.
// It guards against absurd-but-acyclic nesting only; genuine cycles are found
// by DetectRefCycles regardless of depth, so the bound need not be tight. 20 is
// twice the deepest chain measured across the Datadog v2 slices in
// internal/testdata/mini-oas — re-measure there before lowering it.
const DefaultMaxDepth = 20

// Option configures LoadSpec.
type Option func(*loadConfig)

type loadConfig struct {
	maxDepth          int
	trackingFieldName string
}

// WithMaxDepth sets the recursive $ref expansion limit, in $ref edges followed
// on one path. A value <= 0 disables the bound.
func WithMaxDepth(n int) Option {
	return func(c *loadConfig) { c.maxDepth = n }
}

// WithTrackingFieldName overrides the OpenAPI extension key tracking metadata
// is decoded from. An empty name keeps DefaultTrackingFieldName.
func WithTrackingFieldName(name string) Option {
	return func(c *loadConfig) {
		if name != "" {
			c.trackingFieldName = name
		}
	}
}

// LoadSpec parses the OpenAPI v3 spec at path into a *model.Spec: it
// enumerates every path/method operation, sorts them by (path, method) for
// determinism, then resolves operation groups and normalizes their schemas. A
// component chain deeper than the max-depth bound fails the load; a circular
// $ref does not, being classified per node during normalization instead.
func LoadSpec(path string, opts ...Option) (*model.Spec, error) {
	cfg := loadConfig{maxDepth: DefaultMaxDepth, trackingFieldName: DefaultTrackingFieldName}
	for _, opt := range opts {
		opt(&cfg)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec %q: %w", path, err)
	}

	// SkipCircularReferenceCheck hands cycle handling to the normalizer: left on,
	// libopenapi refuses to build the model at all when any component in the
	// document is cyclic, even ones no annotated operation reaches. Start from
	// NewDocumentConfiguration, not a bare &DocumentConfiguration{}: the zero
	// value disables TransformSiblingRefs, MergeReferencedProperties and
	// PreserveLocal, which turn a $ref with siblings into the synthetic allOf.
	docCfg := datamodel.NewDocumentConfiguration()
	docCfg.SkipCircularReferenceCheck = true
	doc, err := libopenapi.NewDocumentWithConfiguration(data, docCfg)
	if err != nil {
		return nil, fmt.Errorf("parsing spec %q: %w", path, err)
	}

	v3doc, err := doc.BuildV3Model()
	if err != nil {
		// A $ref with a missing target is flagged during indexing; report it
		// as a typed *UnresolvableRefError naming the ref.
		if refErr := asUnresolvableRefError(err); refErr != nil {
			return nil, refErr
		}
		return nil, fmt.Errorf("building OpenAPI v3 model for %q: %w", path, err)
	}

	spec := &model.Spec{
		Source:     path,
		Components: v3doc.Model.Components,
		Hash:       specHash(data),
	}

	// Cycles found here are not errors: the walker's stack check stops the
	// recursion and a re-entered $ref is classified per node during
	// normalization. Only a chain longer than maxDepth is fatal.
	if _, err := DetectComponentRefCycles(spec.Components, cfg.maxDepth); err != nil {
		return nil, err
	}

	// rawOps maps each projected operation back to the libopenapi operation it
	// came from, so NormalizeSchemas can reach request/response bodies, which
	// the model itself does not retain.
	rawOps := make(map[*model.Operation]*v3.Operation)
	if paths := v3doc.Model.Paths; paths != nil && paths.PathItems != nil {
		for opPath, item := range paths.PathItems.FromOldest() {
			if item == nil {
				continue
			}
			for method, op := range item.GetOperations().FromOldest() {
				if op == nil {
					continue
				}
				upperMethod := strings.ToUpper(method)
				tracking, err := DecodeTracking(op, opPath, upperMethod, cfg.trackingFieldName)
				if err != nil {
					return nil, err
				}
				mop := &model.Operation{
					Path:        opPath,
					Method:      upperMethod,
					OperationId: op.OperationId,
					Tag:         firstTag(op.Tags),
					Tracking:    tracking,
					Unstable:    declaresUnstable(op),
				}
				spec.Operations = append(spec.Operations, mop)
				rawOps[mop] = op
			}
		}
	}

	// The tag-scoped registry is read after enumeration because a family
	// block is anchored on the operation it names, and resolved before the sort
	// so the synthetic operations it mints sort alongside the real ones.
	if err := fanOutRegistry(spec, rawOps, v3doc.Model.Tags, cfg.trackingFieldName); err != nil {
		return nil, err
	}

	sort.Slice(spec.Operations, func(i, j int) bool {
		a, b := spec.Operations[i], spec.Operations[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Method < b.Method
	})

	if err := CheckDuplicateArtifactNames(spec); err != nil {
		return nil, err
	}

	// Must precede normalization, which walks the operations a group points at:
	// those pointers only exist once every operation is enumerated.
	ResolveOperationGroups(spec)

	if err := NormalizeSchemas(spec, rawOps, cfg.maxDepth, cfg.trackingFieldName); err != nil {
		return nil, err
	}

	return spec, nil
}

// fanOutRegistry decodes the tag-scoped family registry and appends one
// synthetic operation per artifact it declares, each carrying the desugared
// tracking metadata. A document with no registry is left untouched, which is
// what keeps the legacy operation-level annotation working unchanged.
//
// Each artifact is anchored on its primary operation — create for a resource,
// read (else search) for a singular data source, read for a list — so a
// diagnostic still names a real endpoint, and the x-unstable gate set stays
// what it was when the annotation sat on that operation. The anchor is a clone
// rather than the operation itself because one operation can be primary for
// more than one artifact: a family's resource and its singular data source
// both read through the same endpoint. The clone inherits the raw handle so
// schema normalization can still reach its request and response bodies.
func fanOutRegistry(spec *model.Spec, rawOps map[*model.Operation]*v3.Operation, tags []*base.Tag, extensionName string) error {
	metas, err := decodeTagRegistries(tags, extensionName)
	if err != nil {
		return err
	}
	if len(metas) == 0 {
		return nil
	}
	byID := indexOperationsByID(spec.Operations)
	for _, meta := range metas {
		role, operationId := primaryOperation(meta)
		primary := byID[operationId]
		if primary == nil {
			return &RegistryError{Family: meta.Family, Block: meta.Block, Cause: fmt.Errorf(
				"group.%s names operationId %q, which the document does not declare", role, operationId)}
		}
		clone := &model.Operation{
			Path:        primary.Path,
			Method:      primary.Method,
			OperationId: primary.OperationId,
			Tag:         primary.Tag,
			Tracking:    meta,
			Unstable:    primary.Unstable,
			Synthetic:   true,
		}
		spec.Operations = append(spec.Operations, clone)
		rawOps[clone] = rawOps[primary]
	}
	return nil
}

// primaryOperation names the group role an artifact is anchored on, and the
// operationId filling it. The schema guarantees the role is present: a resource
// group requires create, a list group requires read, and a singular group
// requires at least one of read and search.
func primaryOperation(meta *model.TrackingFieldMetadata) (role model.GroupRole, operationId string) {
	g := meta.Group
	switch {
	case meta.ArtifactKind == model.ArtifactKindResource:
		return model.GroupRoleCreate, g.Create
	case meta.Cardinality == model.CardinalityPlural, g.Read != "":
		return model.GroupRoleRead, g.Read
	default:
		return model.GroupRoleSearch, g.Search
	}
}

// firstTag returns the operation's first OpenAPI tag, or "" when untagged. The
// client generator keys package selection on the same tag.
func firstTag(tags []string) string {
	if len(tags) > 0 {
		return tags[0]
	}
	return ""
}

func specHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// unstableExtension is the OpenAPI extension marking a beta endpoint. It is
// the spec's own vocabulary, so unlike the tracking field it is fixed.
const unstableExtension = "x-unstable"

// declaresUnstable reports whether op carries x-unstable. Only presence matters
// — the value is a prose beta notice — so the node is never decoded.
func declaresUnstable(op *v3.Operation) bool {
	if op == nil || op.Extensions == nil {
		return false
	}
	return op.Extensions.GetOrZero(unstableExtension) != nil
}
