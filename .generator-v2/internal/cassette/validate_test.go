package cassette

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// requestKey is a stable key for a create request body, which is where every
// case below declares its values.
func requestKey() SetKey {
	return SetKey{OperationId: "CreateThing", Role: SetRoleRequest}
}

// setWith builds a materialized set carrying just the leaves a case needs.
// Body is left nil: validation reads Values, because those are the paths
// materialization already resolved against the schema.
func setWith(values ...model.MaterializedValue) MaterializedSet {
	return MaterializedSet{Values: values}
}

// objectSchema wraps leaf schemas in the one-level object the paths below
// address, so a case states only the leaf it is about.
func objectSchema(properties map[string]*model.Schema, required ...string) *model.Schema {
	return &model.Schema{
		Kind:       model.SchemaKindObject,
		Properties: properties,
		Required:   required,
	}
}

// violationPaths is the paths a ConformanceError names, for asserting which
// leaves were caught without coupling to the reason wording.
func violationPaths(err error) []string {
	ExpectWithOffset(1, err).To(HaveOccurred())
	conformance, ok := err.(*ConformanceError)
	ExpectWithOffset(1, ok).To(BeTrue(), "expected a *ConformanceError, got %T", err)
	paths := make([]string, 0, len(conformance.Violations))
	for _, v := range conformance.Violations {
		paths = append(paths, v.Path)
	}
	return paths
}

var _ = Describe("scenario validation", func() {
	Describe("a value its schema contradicts", func() {
		// The premise of the feature is that the values come from the
		// description. An example naming an enum member the schema does not
		// declare assembles into a perfectly complete value and is still a
		// value the API will reject.
		It("rejects a value outside its schema enum", func() {
			schema := objectSchema(map[string]*model.Schema{
				"state": {Kind: model.SchemaKindPrimitive, Type: "string",
					Enum: []string{"active", "paused"}},
			}, "state")
			set := setWith(model.MaterializedValue{Path: "state", Value: "archived"})

			err := ValidateSet(requestKey(), set, schema)
			Expect(err).To(HaveOccurred())
			Expect(violationPaths(err)).To(ConsistOf("state"))
		})

		It("accepts a value that is an enum member", func() {
			schema := objectSchema(map[string]*model.Schema{
				"state": {Kind: model.SchemaKindPrimitive, Type: "string",
					Enum: []string{"active", "paused"}},
			}, "state")
			set := setWith(model.MaterializedValue{Path: "state", Value: "paused"})

			Expect(ValidateSet(requestKey(), set, schema)).To(Succeed())
		})

		It("rejects a value whose Go type contradicts the schema type", func() {
			schema := objectSchema(map[string]*model.Schema{
				"retries": {Kind: model.SchemaKindPrimitive, Type: "integer"},
			}, "retries")
			set := setWith(model.MaterializedValue{Path: "retries", Value: "three"})

			err := ValidateSet(requestKey(), set, schema)
			Expect(err).To(HaveOccurred())
			Expect(violationPaths(err)).To(ConsistOf("retries"))
		})

		It("rejects a string that contradicts its declared format", func() {
			schema := objectSchema(map[string]*model.Schema{
				"created_at": {Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time"},
			}, "created_at")
			set := setWith(model.MaterializedValue{Path: "created_at", Value: "last Tuesday"})

			err := ValidateSet(requestKey(), set, schema)
			Expect(err).To(HaveOccurred())
			Expect(violationPaths(err)).To(ConsistOf("created_at"))
		})

		// A format the normalized model does not know how to check must not
		// become a refusal: the generator would reject descriptions for using
		// a vocabulary it simply has not learned.
		It("accepts a string whose format it cannot check", func() {
			schema := objectSchema(map[string]*model.Schema{
				"selector": {Kind: model.SchemaKindPrimitive, Type: "string", Format: "css-selector"},
			}, "selector")
			set := setWith(model.MaterializedValue{Path: "selector", Value: "div > p"})

			Expect(ValidateSet(requestKey(), set, schema)).To(Succeed())
		})

		It("names every violating leaf, not only the first", func() {
			schema := objectSchema(map[string]*model.Schema{
				"state":   {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"active"}},
				"retries": {Kind: model.SchemaKindPrimitive, Type: "integer"},
			}, "state", "retries")
			set := setWith(
				model.MaterializedValue{Path: "state", Value: "archived"},
				model.MaterializedValue{Path: "retries", Value: "three"},
			)

			err := ValidateSet(requestKey(), set, schema)
			Expect(violationPaths(err)).To(ConsistOf("state", "retries"))
		})
	})

	Describe("an unsupported schema node", func() {
		// The normalizer marks a node it cannot represent and records why, so
		// the affected artifact can fail locally instead of aborting the spec
		// load. The cassette package handles no such kind today: the
		// materializer's switch covers object/oneOf and array/map, and an
		// unsupported node falls through to the leaf path, where it quietly
		// acquires a scalar value for something that has no representation.
		It("rejects a value standing on a node the normalizer could not represent", func() {
			schema := objectSchema(map[string]*model.Schema{
				"payload": {
					Kind:              model.SchemaKindUnsupported,
					UnsupportedReason: "recursive $ref exceeded the depth limit",
				},
			}, "payload")
			set := setWith(model.MaterializedValue{Path: "payload", Value: "dummy-payload"})

			err := ValidateSet(requestKey(), set, schema)
			Expect(err).To(HaveOccurred())
			Expect(violationPaths(err)).To(ConsistOf("payload"))
		})

		// The reason the normalizer recorded is the only actionable part of
		// the diagnostic, so it has to survive into the violation.
		It("carries the normalizer's own reason into the violation", func() {
			schema := objectSchema(map[string]*model.Schema{
				"payload": {
					Kind:              model.SchemaKindUnsupported,
					UnsupportedReason: "recursive $ref exceeded the depth limit",
				},
			}, "payload")
			set := setWith(model.MaterializedValue{Path: "payload", Value: "dummy-payload"})

			err := ValidateSet(requestKey(), set, schema)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("recursive $ref exceeded the depth limit"))
		})
	})

	Describe("what it must not report", func() {
		// A synthesized leaf is schema-valid by construction and is already
		// reported on SynthesizedPaths. Treating it as a violation would make
		// every artifact whose description omits a required example ineligible
		// again, which is the refusal the synthesis change deliberately
		// removed.
		It("accepts a synthesized value and leaves it to SynthesizedPaths", func() {
			schema := objectSchema(map[string]*model.Schema{
				"state": {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"active"}},
			}, "state")
			set := MaterializedSet{
				Values:           []model.MaterializedValue{{Path: "state", Value: "active"}},
				SynthesizedPaths: []string{"state"},
			}

			Expect(ValidateSet(requestKey(), set, schema)).To(Succeed())
		})

		// A replaced secret holds a safe stand-in, not what the description
		// declared, so checking it against the schema checks the replacement
		// rather than the example.
		It("accepts a replaced sensitive value", func() {
			schema := objectSchema(map[string]*model.Schema{
				"token": {Kind: model.SchemaKindPrimitive, Type: "string", Format: "uuid",
					WriteOnlySecret: true},
			}, "token")
			set := MaterializedSet{
				Values: []model.MaterializedValue{
					{Path: "token", Value: "[redacted]", Sensitive: true},
				},
				SensitiveReplacements: map[string]string{"token": "[redacted]"},
			}

			Expect(ValidateSet(requestKey(), set, schema)).To(Succeed())
		})

		// Nil is a declared null rather than absence, and a nullable field is
		// not something the normalized model distinguishes, so a null must not
		// be read as a type violation.
		It("accepts a declared null", func() {
			schema := objectSchema(map[string]*model.Schema{
				"description": {Kind: model.SchemaKindPrimitive, Type: "string"},
			})
			set := setWith(model.MaterializedValue{Path: "description", Value: nil})

			Expect(ValidateSet(requestKey(), set, schema)).To(Succeed())
		})

		// A value whose path the schema does not describe is materialization's
		// business, not validation's. Reporting it here would duplicate a
		// check that already failed upstream, with a worse message.
		It("ignores a path the schema does not describe", func() {
			schema := objectSchema(map[string]*model.Schema{
				"state": {Kind: model.SchemaKindPrimitive, Type: "string"},
			})
			set := setWith(model.MaterializedValue{Path: "nonexistent", Value: "whatever"})

			Expect(ValidateSet(requestKey(), set, schema)).To(Succeed())
		})

		It("accepts an empty set", func() {
			Expect(ValidateSet(requestKey(), MaterializedSet{}, objectSchema(nil))).To(Succeed())
		})

		It("accepts a nil schema rather than panicking", func() {
			set := setWith(model.MaterializedValue{Path: "state", Value: "active"})
			Expect(ValidateSet(requestKey(), set, nil)).To(Succeed())
		})
	})

	Describe("the diagnostic it produces", func() {
		// Every error type in this package names paths and never values,
		// because a declared value may be a credential and a diagnostic is
		// the one place it must not appear.
		It("never quotes the offending value", func() {
			schema := objectSchema(map[string]*model.Schema{
				"token": {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"a", "b"}},
			}, "token")
			set := setWith(model.MaterializedValue{Path: "token", Value: "sk-live-deadbeef"})

			err := ValidateSet(requestKey(), set, schema)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).NotTo(ContainSubstring("sk-live-deadbeef"))
		})

		It("names the set the violations belong to", func() {
			schema := objectSchema(map[string]*model.Schema{
				"state": {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"active"}},
			}, "state")
			set := setWith(model.MaterializedValue{Path: "state", Value: "archived"})

			err := ValidateSet(requestKey(), set, schema)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("CreateThing"))
		})
	})
})
