package emit

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("BuildDataSourceView", func() {
	It("resolves the SDK call bindings onto the view", func() {
		view, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())
		Expect(view.TypeName).To(Equal("incident_type"))
		Expect(view.GoName).To(Equal("datadogIncidentType"))
		Expect(view.Description).To(Equal("Use this data source to retrieve information about an existing incident type."))
		Expect(view.SDKPackage).To(Equal("datadogV2"))
		Expect(view.APIStruct).To(Equal("IncidentsApi"))
		Expect(view.APIAccessor).To(Equal("GetIncidentsApiV2"))
		Expect(view.Read.Method).To(Equal("GetIncidentType"))
		Expect(view.Read.ResponseType).To(Equal("IncidentTypeResponse"))
	})

	It("flattens the envelope: data.attributes.* become top-level computed attributes, sorted, with no nested blocks", func() {
		view, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Schema.Blocks).To(BeEmpty())

		var names []string
		for _, a := range view.Schema.Attributes {
			names = append(names, a.TFName)
			Expect(a.Computed).To(BeTrue(), "attribute %q should be computed", a.TFName)
		}
		Expect(names).To(Equal([]string{"description", "is_default", "name"}))
	})

	It("prepends the lookup id and maps state off resp.Data through guarded optional getters", func() {
		view, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())

		Expect(view.Models).To(HaveLen(1))
		var fields []string
		for _, f := range view.Models[0].Fields {
			fields = append(fields, f.GoField)
		}
		Expect(fields).To(Equal([]string{"ID", "Description", "IsDefault", "Name"}))

		Expect(view.State.ParamName).To(Equal("resp"))
		Expect(view.State.ParamType).To(Equal("*datadogV2.IncidentTypeResponse"))
		Expect(view.State.Preamble).To(Equal([]string{"attributes := resp.Data.GetAttributes()"}))
		// Guarded assignments: an absent field stays null rather than a zero value.
		Expect(view.State.Assignments).To(Equal([]StateAssignment{
			{Var: "id", GetterOk: "resp.Data.GetIdOk()", LHS: "state.ID", RHS: "types.StringValue(*id)"},
			{Var: "description", GetterOk: "attributes.GetDescriptionOk()", LHS: "state.Description", RHS: "types.StringValue(*description)"},
			{Var: "isDefault", GetterOk: "attributes.GetIsDefaultOk()", LHS: "state.IsDefault", RHS: "types.BoolValue(*isDefault)"},
			{Var: "name", GetterOk: "attributes.GetNameOk()", LHS: "state.Name", RHS: "types.StringValue(*name)"},
		}))
	})

	It("renders date-time and UUID strings via .String(), a named enum via a string() cast, and avoids shadowing state", func() {
		op := incidentTypeOperation()
		attrs := op.ResponseSchema.Properties["data"].Properties["attributes"].Properties
		attrs["resolved_at"] = &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time"}
		attrs["owner_id"] = &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Format: "uuid"}
		attrs["state"] = &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"active", "archived"}}
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())

		assign := map[string]StateAssignment{}
		for _, a := range view.State.Assignments {
			assign[a.LHS] = a
		}
		Expect(assign["state.ResolvedAt"].RHS).To(Equal("types.StringValue(resolvedAt.String())"))
		Expect(assign["state.ResolvedAt"].GetterOk).To(Equal("attributes.GetResolvedAtOk()"))
		Expect(assign["state.OwnerId"].RHS).To(Equal("types.StringValue(ownerId.String())"))
		// "state" would shadow the updateState receiver, so its local is suffixed.
		Expect(assign["state.State"].Var).To(Equal("stateValue"))
		Expect(assign["state.State"].RHS).To(Equal("types.StringValue(string(*stateValue))"))
	})

	It("renders a UUID envelope id through its canonical string form", func() {
		op := incidentTypeOperation()
		op.ResponseSchema.Properties["data"].Properties["id"].Format = "uuid"

		view := mustView(op)
		Expect(view.State.Assignments[0]).To(Equal(StateAssignment{
			Var: "id", GetterOk: "resp.Data.GetIdOk()", LHS: "state.ID", RHS: "types.StringValue(id.String())",
		}))
	})

	It("keeps the Terraform identity without calling an absent response ID getter", func() {
		op := incidentTypeOperation()
		delete(op.ResponseSchema.Properties["data"].Properties, "id")

		view := mustView(op)
		Expect(view.Models[0].Fields[0]).To(Equal(ModelFieldView{
			GoField: "ID", GoType: "types.String", TFName: "id",
		}))
		Expect(view.State.Assignments).NotTo(ContainElement(HaveField("LHS", "state.ID")))

		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		src := string(rendered)
		Expect(src).NotTo(ContainSubstring("GetIdOk()"))
		Expect(src).NotTo(ContainSubstring("state.ID ="))
	})

	It("omits the attributes preamble when no response assignments use it", func() {
		op := incidentTypeOperation()
		op.ResponseSchema.Properties["data"].Properties["attributes"].Properties = map[string]*model.Schema{}

		view := mustView(op)
		Expect(view.State.Preamble).To(BeEmpty())
		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(rendered)).NotTo(ContainSubstring("attributes :="))
	})

	It("renders a Go-keyword attribute through a safe guarded local", func() {
		op := incidentTypeOperation()
		op.ResponseSchema.Properties["data"].Properties["attributes"].
			Properties["type"] = prim("string", "The incident type category.")

		got, err := RenderDataSource(mustView(op))
		Expect(err).NotTo(HaveOccurred())
		src := string(got)
		Expect(src).To(ContainSubstring(
			"if typeVar, ok := attributes.GetTypeOk(); ok && typeVar != nil {",
		))
		Expect(src).To(ContainSubstring("state.Type = types.StringValue(*typeVar)"))
		Expect(src).NotTo(ContainSubstring("if type, ok :="))
	})

	It("casts a named string envelope id before storing it in Terraform state", func() {
		op := incidentTypeOperation()
		op.ResponseSchema.Properties["data"].Properties["id"].Enum = []string{"permanent"}
		view := mustView(op)
		Expect(view.State.Assignments[0]).To(Equal(StateAssignment{
			Var: "id", GetterOk: "resp.Data.GetIdOk()", LHS: "state.ID", RHS: "types.StringValue(string(*id))",
		}))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})

	It("renders resolved path arguments in SDK order, including UUID conversion", func() {
		op := incidentTypeOperation()
		op.Path = "/api/v2/accounts/{account_id}/incident-types/{incident_type_id}"
		op.SDKBinding = &model.SDKOperationBinding{Required: []model.SDKArgument{
			{Name: "account_id", GoName: "accountId", GoType: "int64", Location: "path", Schema: prim("integer", "The account ID.")},
			{Name: "incident_type_id", GoName: "incidentTypeId", GoType: "uuid.UUID", Location: "path", Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Format: "uuid"}},
		}}

		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.UsesUUID).To(BeTrue())
		Expect(view.Read.Arguments).To(Equal([]SDKArgumentView{
			{Expression: "state.AccountId.ValueInt64()", TFName: "account_id"},
			{Expression: "parsedId", UUIDVar: "parsedId", UUIDSource: "state.ID.ValueString()", TFName: "id"},
		}))

		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		src := string(rendered)
		Expect(src).To(ContainSubstring(`parsedId, err := uuid.Parse(state.ID.ValueString())`))
		Expect(src).To(ContainSubstring(`GetIncidentType(d.Auth, state.AccountId.ValueInt64(), parsedId)`))
	})

	It("coalesces a required path lookup with the returned field of the same name", func() {
		op := incidentTypeOperation()
		op.Path = "/api/v2/accounts/{account_id}/incident-types/{incident_type_id}"
		op.ResponseSchema.Properties["data"].Properties["attributes"].Properties["account_id"] =
			prim("integer", "The returned account ID.")
		op.SDKBinding = &model.SDKOperationBinding{Required: []model.SDKArgument{
			{Name: "account_id", GoName: "accountId", GoType: "int64", Location: "path", Schema: prim("integer", "The lookup account ID.")},
			{Name: "incident_type_id", GoName: "incidentTypeId", GoType: "string", Location: "path", Schema: prim("string", "The incident type ID.")},
		}}

		view := mustView(op)
		var accountAttrs []AttrView
		for _, attr := range view.Schema.Attributes {
			if attr.TFName == "account_id" {
				accountAttrs = append(accountAttrs, attr)
			}
		}
		Expect(accountAttrs).To(Equal([]AttrView{{
			TFName: "account_id", TFType: "schema.Int64Attribute", Description: "The account_id argument.", Required: true,
		}}))
		var accountFields []ModelFieldView
		for _, field := range view.Models[0].Fields {
			if field.TFName == "account_id" {
				accountFields = append(accountFields, field)
			}
		}
		Expect(accountFields).To(HaveLen(1))
		Expect(view.State.Assignments).To(ContainElement(HaveField("LHS", "state.AccountId")))
	})

	It("parses a string-valued Terraform id for a numeric SDK path argument", func() {
		op := incidentTypeOperation()
		op.SDKBinding = &model.SDKOperationBinding{Required: []model.SDKArgument{{
			Name: "incident_type_id", GoName: "incidentTypeId", GoType: "int64", Location: "path",
			Schema: prim("integer", "The numeric incident type ID."),
		}}}

		view := mustView(op)
		Expect(view.UsesStrconv).To(BeTrue())
		Expect(view.Read.Arguments).To(Equal([]SDKArgumentView{{
			Expression: "parsedId", IntVar: "parsedId", IntSource: "state.ID.ValueString()", IntBits: 64, TFName: "id",
		}}))

		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		src := string(rendered)
		Expect(src).To(ContainSubstring(`"strconv"`))
		Expect(src).To(ContainSubstring(`parsedId, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)`))
		Expect(src).To(ContainSubstring(`response.Diagnostics.AddError("Invalid id", err.Error())`))
		Expect(src).To(ContainSubstring(`GetIncidentType(d.Auth, parsedId)`))
	})

	It("renders a resolved singleton call without inventing an id argument", func() {
		op := incidentTypeOperation()
		op.Path = "/api/v2/incidents/config/types/default"
		op.SDKBinding = &model.SDKOperationBinding{}

		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.ByID).To(BeFalse())
		Expect(view.Read.Arguments).To(BeEmpty())

		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(rendered)).To(ContainSubstring(`GetIncidentType(d.Auth)`))
	})

	DescribeTable("fail-slows anything outside the recognized envelope",
		func(mutate func(*model.Operation), wantReason string) {
			op := incidentTypeOperation()
			mutate(op)
			art, err := model.BuildArtifact(op)
			Expect(err).NotTo(HaveOccurred())

			view, err := BuildDataSourceView(art)
			var uerr *UnsupportedEmitError
			Expect(errors.As(err, &uerr)).To(BeTrue(), "expected an UnsupportedEmitError, got %v", err)
			Expect(uerr.Error()).To(ContainSubstring(wantReason))
			Expect(view).To(Equal(DataSourceView{}), "no view should be produced on failure")
		},
		Entry("a response root with no data object",
			func(op *model.Operation) {
				op.ResponseSchema = obj(map[string]*model.Schema{"name": prim("string", "")})
			},
			"expected a JSON:API envelope with a data object"),
		Entry("an id_strategy other than data.id",
			func(op *model.Operation) { op.Tracking.IdStrategy = model.IdStrategyDataAttributesUID },
			"id_strategy"),
		Entry("a missing response type name",
			func(op *model.Operation) { op.ResponseRefName = "" },
			"missing response type name"),
	)

	It("drops a data member outside {id, type, attributes} and records it", func() {
		op := incidentTypeOperation()
		op.ResponseSchema.Properties["data"].Properties["relationships"] =
			obj(map[string]*model.Schema{"created_by": prim("string", "")})
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Dropped).To(ContainElement(And(
			HaveField("Message", ContainSubstring("relationships")),
			HaveField("Severity", model.SeverityInfo),
		)))
	})
})

var _ = Describe("BuildDataSourceView envelope id collision", func() {
	It("drops an id under singular attributes in favour of the envelope id and warns", func() {
		op := incidentTypeOperation()
		op.ResponseSchema.Properties["data"].Properties["attributes"].
			Properties["id"] = prim("string", "The incident type's config ID.")
		view := mustView(op)

		Expect(view.Models).To(HaveLen(1))
		var fields []string
		for _, f := range view.Models[0].Fields {
			fields = append(fields, f.TFName)
		}
		Expect(fields).To(Equal([]string{"id", "description", "is_default", "name"}))
		Expect(view.Dropped).To(ContainElement(And(
			HaveField("Message", ContainSubstring("attributes.id")),
			HaveField("Message", ContainSubstring(`collides with the envelope id`)),
			HaveField("Severity", model.SeverityWarning),
		)))
	})

	It("drops an id under plural item attributes in favour of the item's envelope id and warns", func() {
		op := datastoresOperation()
		op.ResponseSchema.Properties["data"].Items.
			Properties["attributes"].Properties["id"] = prim("string", "The datastore's config ID.")
		view := mustView(op)

		var itemAttrs []string
		for _, b := range view.Schema.Blocks {
			for _, a := range b.Attributes {
				itemAttrs = append(itemAttrs, a.TFName)
			}
		}
		Expect(itemAttrs).To(Equal([]string{
			"creator_user_id", "creator_user_uuid", "description", "id", "name",
			"org_id", "primary_column_name", "primary_key_generation_strategy",
		}))
		Expect(view.Dropped).To(ContainElement(And(
			HaveField("Message", ContainSubstring("attributes.id")),
			HaveField("Severity", model.SeverityWarning),
		)))
	})

	It("keeps an id under attributes when the envelope has no id of its own", func() {
		op := datastoresOperation()
		items := op.ResponseSchema.Properties["data"].Items
		delete(items.Properties, "id")
		items.Properties["attributes"].Properties["id"] = prim("string", "The datastore's config ID.")
		view := mustView(op)

		var itemAttrs []string
		for _, b := range view.Schema.Blocks {
			for _, a := range b.Attributes {
				itemAttrs = append(itemAttrs, a.TFName)
			}
		}
		Expect(itemAttrs).To(Equal([]string{
			"creator_user_id", "creator_user_uuid", "description", "id", "name",
			"org_id", "primary_column_name", "primary_key_generation_strategy",
		}))
		Expect(view.Dropped).NotTo(ContainElement(
			HaveField("Message", ContainSubstring("collides with the envelope id"))))
	})

	It("renders a colliding singular envelope to compiling Go with a single id attribute", func() {
		op := incidentTypeOperation()
		op.ResponseSchema.Properties["data"].Properties["attributes"].
			Properties["id"] = prim("string", "The incident type's config ID.")

		src, err := RenderDataSource(mustView(op))
		Expect(err).NotTo(HaveOccurred())
		Expect(bytes.Count(src, []byte("`tfsdk:\"id\"`"))).To(Equal(1))
	})

	It("produces byte-identical output across two runs of the same input", func() {
		build := func() []byte {
			op := datastoresOperation()
			op.ResponseSchema.Properties["data"].Items.
				Properties["attributes"].Properties["id"] = prim("string", "The datastore's config ID.")
			src, err := RenderDataSource(mustView(op))
			Expect(err).NotTo(HaveOccurred())
			return src
		}
		Expect(build()).To(Equal(build()))
	})
})

var _ = Describe("leafVar", func() {
	It("suffixes every Go keyword and preserves ordinary identifiers", func() {
		keywords := []string{
			"break", "default", "func", "interface", "select",
			"case", "defer", "go", "map", "struct",
			"chan", "else", "goto", "package", "switch",
			"const", "fallthrough", "if", "range", "type",
			"continue", "for", "import", "return", "var",
		}
		for _, keyword := range keywords {
			Expect(leafVar(keyword)).To(Equal(keyword+"Var"), keyword)
		}

		Expect(leafVar("display_name")).To(Equal("displayName"))
		Expect(leafVar("state")).To(Equal("stateValue"))
	})
})

var _ = Describe("RenderDataSource duplicate field guard", func() {
	It("fails instead of writing a model that declares the same tfsdk tag twice", func() {
		view := mustView(incidentTypeOperation())
		view.Models[0].Fields = append(view.Models[0].Fields,
			ModelFieldView{GoField: "ID", GoType: "types.String", TFName: "id"})

		_, err := RenderDataSource(view)
		Expect(err).To(MatchError(ContainSubstring(`declares tfsdk:"id" twice`)))
	})

	DescribeTable("drops a JSON:API sibling of data rather than failing on the response shape",
		func(member string, schema *model.Schema) {
			op := incidentTypeOperation()
			op.ResponseSchema.Properties[member] = schema
			view := mustView(op)

			Expect(view.Dropped).To(HaveLen(1))
			Expect(view.Dropped[0].Message).To(ContainSubstring("dropped \"response." + member + "\": not part of the surfaced {id, type, attributes} envelope"))

			for _, a := range view.Schema.Attributes {
				Expect(a.TFName).NotTo(Equal(member))
			}
			for _, blk := range view.Schema.Blocks {
				Expect(blk.TFName).NotTo(Equal(member))
			}
		},
		Entry("meta", "meta", obj(map[string]*model.Schema{"page": prim("string", "")})),
		Entry("links", "links", obj(map[string]*model.Schema{"next": prim("string", "")})),
		Entry("included", "included", &model.Schema{
			Kind:  model.SchemaKindArray,
			Items: obj(map[string]*model.Schema{"handle": prim("string", "")}),
		}),
	)

	It("drops included, keeping its element union out of the view entirely", func() {
		// The real shape: included is an array whose element is a oneOf, since it
		// sideloads more than one resource type (User | LeakedKey for an API key).
		op := incidentTypeOperation()
		op.ResponseSchema.Properties["included"] = &model.Schema{
			Kind: model.SchemaKindArray,
			Items: &model.Schema{
				Kind: model.SchemaKindOneOf,
				OneOf: &model.OneOfSpec{
					Name: "IncidentTypeResponseIncludedItem",
					Path: "response.included[]",
					Variants: []model.OneOfVariant{{
						TFName: "user",
						GoName: "User",
						Schema: obj(map[string]*model.Schema{"handle": prim("string", "")}),
					}},
				},
			},
		}
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		// The model still projects the union faithfully; only the view drops it.
		Expect(art.Schema.Attributes).To(ContainElement(HaveField("Path", "response.included")))

		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred(), "a union under a dropped member must not fail the artifact")
		Expect(view.Dropped).To(HaveLen(1))
		Expect(view.Dropped[0].Message).To(ContainSubstring("dropped \"response.included\": not part of the surfaced {id, type, attributes} envelope"))
	})

	It("fails instead of writing two model types with the same Go name", func() {
		view := mustView(incidentTypeOperation())
		view.Models = append(view.Models, ModelStructView{Name: view.Models[0].Name})

		_, err := RenderDataSource(view)
		Expect(err).To(MatchError(ContainSubstring(`model datadogIncidentTypeDataSourceModel is declared twice`)))
	})
})

var _ = Describe("BuildDataSourceView audit fields", func() {
	It("drops server-managed created_at/updated_at/created_by/updated_by/modified_at from a singular record and records them", func() {
		op := incidentTypeOperation()
		attrs := op.ResponseSchema.Properties["data"].Properties["attributes"].Properties
		for _, f := range []string{"created_at", "updated_at", "created_by", "updated_by", "modified_at"} {
			attrs[f] = prim("string", "")
		}
		view := mustView(op)

		for _, a := range view.Schema.Attributes {
			Expect(a.TFName).NotTo(BeElementOf("created_at", "updated_at", "created_by", "updated_by", "modified_at"))
		}
		Expect(view.Dropped).To(ContainElement(And(
			HaveField("Message", ContainSubstring("created_at")),
			HaveField("Severity", model.SeverityInfo),
		)))
	})

	It("keeps an audit-named field nested inside an object", func() {
		view := mustView(retentionFilterOperation())

		blocks := map[string]AttrView{}
		for _, b := range view.Schema.Blocks {
			blocks[b.TFName] = b
		}
		metadata := map[string]AttrView{}
		for _, b := range blocks["filter"].Blocks {
			metadata[b.TFName] = b
		}
		var nested []string
		for _, a := range metadata["metadata"].Attributes {
			nested = append(nested, a.TFName)
		}
		Expect(nested).To(ContainElement("created_by"))
	})

	It("drops top-level audit fields from plural item attributes but keeps non-audit siblings", func() {
		view := mustView(datastoresOperation())

		var itemAttrs []string
		for _, b := range view.Schema.Blocks {
			for _, a := range b.Attributes {
				itemAttrs = append(itemAttrs, a.TFName)
			}
		}
		Expect(itemAttrs).NotTo(ContainElement("created_at"))
		Expect(itemAttrs).NotTo(ContainElement("modified_at"))
		Expect(itemAttrs).To(ContainElement("description"))
	})
})

var _ = Describe("BuildDataSourceView singular search", func() {
	Context("search only", func() {
		It("binds the list call as Search, derives filters, and computes the id", func() {
			view, err := BuildDataSourceView(mustArtifact(powerpackSearchOperation()))
			Expect(err).NotTo(HaveOccurred())

			Expect(view.ByID).To(BeFalse())
			Expect(view.Searchable).To(BeTrue())
			Expect(view.Search.Method).To(Equal("ListPowerpacks"))
			Expect(view.Search.ItemType).To(Equal("PowerpackData"))
			Expect(view.Search.Paginated).To(BeTrue())

			// The lone filter becomes both an Optional schema attr and a list param.
			Expect(view.Search.Filters).To(Equal([]FilterParamView{
				{StateField: "FilterName", ParamField: "FilterName", ValueExpr: "ValueStringPointer()"},
			}))
			Expect(view.Schema.Blocks).To(BeEmpty(), "singular output has no list/items block")

			// The record reads off the list element by value, through guarded getters.
			Expect(view.State.ParamName).To(Equal("data"))
			Expect(view.State.ParamType).To(Equal("datadogV2.PowerpackData"))
			Expect(view.State.Preamble).To(Equal([]string{"attributes := data.GetAttributes()"}))
		})

		It("coalesces an optional search lookup with its computed returned field", func() {
			op := powerpackSearchOperation()
			op.QueryParams[0].Name = "name"

			view := mustView(op)
			var nameAttrs []AttrView
			for _, attr := range view.Schema.Attributes {
				if attr.TFName == "name" {
					nameAttrs = append(nameAttrs, attr)
				}
			}
			Expect(nameAttrs).To(Equal([]AttrView{{
				TFName: "name", TFType: "schema.StringAttribute", Description: "The name of the Powerpack to search for.",
				Optional: true, Computed: true,
			}}))
			var nameFields []ModelFieldView
			for _, field := range view.Models[0].Fields {
				if field.TFName == "name" {
					nameFields = append(nameFields, field)
				}
			}
			Expect(nameFields).To(HaveLen(1))
			Expect(view.State.Assignments).To(ContainElement(HaveField("LHS", "state.Name")))
		})

		It("parses an optional UUID filter before calling its pinned SDK setter", func() {
			op := powerpackSearchOperation()
			op.QueryParams[0].Schema.Format = "uuid"
			op.SDKBinding = &model.SDKOperationBinding{
				OptionalParamsType: "ListPowerpacksOptionalParameters",
				Optional: []model.SDKArgument{{
					Name: "filter[name]", GoName: "filterName", GoType: "uuid.UUID", Location: "query",
					Schema: op.QueryParams[0].Schema, Setter: "WithFilterName",
				}},
			}

			view := mustView(op)
			Expect(view.UsesUUID).To(BeTrue())
			Expect(view.Search.Filters).To(Equal([]FilterParamView{{
				StateField: "FilterName", ParamField: "FilterName", ValueExpr: "parsedFilterName", Setter: "WithFilterName",
				UUIDVar: "parsedFilterName", UUIDSource: "state.FilterName.ValueString()", TFName: "filter_name",
			}}))

			rendered, err := RenderDataSource(view)
			Expect(err).NotTo(HaveOccurred())
			src := string(rendered)
			Expect(src).To(ContainSubstring(`parsedFilterName, err := uuid.Parse(state.FilterName.ValueString())`))
			Expect(src).To(ContainSubstring(`response.Diagnostics.AddError("Invalid filter_name", err.Error())`))
			Expect(src).To(ContainSubstring(`optionalParams.WithFilterName(parsedFilterName)`))
		})

		It("parses an RFC3339 filter before calling a time.Time SDK setter", func() {
			op := powerpackSearchOperation()
			op.QueryParams[0].Schema.Format = "date-time"
			op.SDKBinding = &model.SDKOperationBinding{
				OptionalParamsType: "ListPowerpacksOptionalParameters",
				Optional: []model.SDKArgument{{
					Name: "filter[name]", GoName: "filterName", GoType: "time.Time", Location: "query",
					Schema: op.QueryParams[0].Schema, Setter: "WithFilterName",
				}},
			}

			view := mustView(op)
			Expect(view.UsesTime).To(BeTrue())
			Expect(view.Search.Filters).To(Equal([]FilterParamView{{
				StateField: "FilterName", ParamField: "FilterName", ValueExpr: "parsedFilterName", Setter: "WithFilterName",
				TimeVar: "parsedFilterName", TimeSource: "state.FilterName.ValueString()", TimeLayout: "time.RFC3339", TFName: "filter_name",
			}}))

			rendered, err := RenderDataSource(view)
			Expect(err).NotTo(HaveOccurred())
			src := string(rendered)
			Expect(src).To(ContainSubstring(`"time"`))
			Expect(src).To(ContainSubstring(`parsedFilterName, err := time.Parse(time.RFC3339, state.FilterName.ValueString())`))
			Expect(src).To(ContainSubstring(`optionalParams.WithFilterName(parsedFilterName)`))
		})

		It("hashes the search inputs when the selected record has no id", func() {
			op := powerpackSearchOperation()
			delete(op.ResponseSchema.Properties["data"].Items.Properties, "id")

			view := mustView(op)
			Expect(view.Search.HashInputs).To(Equal([]FilterParamView{
				{StateField: "FilterName", ValueExpr: "ValueStringPointer()"},
			}))
			Expect(view.State.Assignments).NotTo(ContainElement(HaveField("LHS", "state.ID")))

			rendered, err := RenderDataSource(view)
			Expect(err).NotTo(HaveOccurred())
			src := string(rendered)
			Expect(src).To(ContainSubstring(`hashingData := fmt.Sprintf("%s", state.FilterName.ValueString())`))
			Expect(src).To(ContainSubstring(`state.ID = types.StringValue(utils.ConvertToSha256(hashingData))`))
		})

	})

	Context("both", func() {
		It("binds the by-id Read and the list Search and makes the id optional+computed", func() {
			view, err := BuildDataSourceView(mustArtifact(datastoreBothOperation()))
			Expect(err).NotTo(HaveOccurred())

			Expect(view.ByID).To(BeTrue())
			Expect(view.Searchable).To(BeTrue())
			Expect(view.Read.Method).To(Equal("GetDatastore"))
			Expect(view.Search.Method).To(Equal("ListDatastores"))
			Expect(view.State.ParamType).To(Equal("datadogV2.DatastoreData"))
		})

		It("partitions the list op's scalar filter into an Optional attribute and a search param alongside the by-id id", func() {
			view, err := BuildDataSourceView(mustArtifact(datastoreBothWithFilterOperation()))
			Expect(err).NotTo(HaveOccurred())

			Expect(view.ByID).To(BeTrue())
			Expect(view.Searchable).To(BeTrue())

			// The filter binds the list call's optional parameter...
			Expect(view.Search.Filters).To(Equal([]FilterParamView{
				{StateField: "FilterKeyword", ParamField: "FilterKeyword", ValueExpr: "ValueStringPointer()"},
			}))
			// ...and surfaces as an Optional schema attribute next to the record.
			Expect(view.Schema.Attributes).To(ContainElement(AttrView{
				TFName:      "filter_keyword",
				TFType:      "schema.StringAttribute",
				Description: "Filter datastores by keyword.",
				Optional:    true,
			}))
		})

		It("hashes a fixed seed on the search path when the selected record has no id", func() {
			op := datastoreBothOperation()
			delete(op.ResponseSchema.Properties["data"].Properties, "id")
			delete(op.SearchOp.ResponseSchema.Properties["data"].Items.Properties, "id")

			view := mustView(op)
			Expect(view.Search.HashInputs).To(BeEmpty())
			Expect(view.State.Assignments).NotTo(ContainElement(HaveField("LHS", "state.ID")))

			rendered, err := RenderDataSource(view)
			Expect(err).NotTo(HaveOccurred())
			src := string(rendered)
			Expect(src).To(ContainSubstring(`hashingData := "datastore"`))
			Expect(src).To(ContainSubstring(`state.ID = types.StringValue(utils.ConvertToSha256(hashingData))`))
		})

	})

	DescribeTable("the emitted Read guards the result count and indexes only the single match",
		func(fixture func() *model.Operation) {
			got, err := RenderDataSource(mustView(fixture()))
			Expect(err).NotTo(HaveOccurred())
			src := string(got)
			Expect(src).To(ContainSubstring(`if len(items) == 0 {`))
			Expect(src).To(ContainSubstring(`response.Diagnostics.AddError("filters returned no results", "")`))
			Expect(src).To(ContainSubstring(`if len(items) > 1 {`))
			Expect(src).To(ContainSubstring(`use more specific search criteria`))
			Expect(src).To(ContainSubstring(`response.Diagnostics.Append(d.updateState(ctx, &state, items[0])...)`))
		},
		Entry("search only", powerpackSearchOperation),
		Entry("both", datastoreBothOperation),
	)

	It("absent fields stay null: every record assignment is a guarded optional getter", func() {
		got, err := RenderDataSource(mustView(datastoreBothOperation()))
		Expect(err).NotTo(HaveOccurred())
		src := string(got)
		Expect(src).To(ContainSubstring(`if name, ok := attributes.GetNameOk(); ok && name != nil {`))
		Expect(src).NotTo(ContainSubstring(`types.StringValue(attributes.GetName())`), "must not write the unguarded zero value")
	})
})

// mustView builds an Artifact and its view from op or fails the test.
func mustView(op *model.Operation) DataSourceView {
	GinkgoHelper()
	view, err := BuildDataSourceView(mustArtifact(op))
	Expect(err).NotTo(HaveOccurred())
	return view
}

// powerpackSearchOperation is a search-only singular data source: the list GET is
// the tracked op, paginated, with one scalar filter. (A representative server-side
// search shape; the real powerpack matches client-side, which is out of scope.)
func powerpackSearchOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/powerpacks",
		Method:          "GET",
		OperationId:     "ListPowerpacks",
		Tag:             "Powerpacks",
		ResponseRefName: "PowerpacksResponse",
		ItemRefName:     "PowerpackData",
		Pagination:      &model.Pagination{LimitParam: "page[limit]", PageParam: "page[offset]", ResultsPath: "data"},
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "powerpack",
			TfDescription: "Use this data source to retrieve information about an existing Datadog Powerpack.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Search: "ListPowerpacks"},
		},
		QueryParams: []model.QueryParam{
			{Name: "filter[name]", Schema: prim("string", ""), Description: "The name of the Powerpack to search for."},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind: model.SchemaKindArray,
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The ID of the Powerpack."),
					"type": prim("string", "Type of widget, must be `powerpack`."),
					"attributes": obj(map[string]*model.Schema{
						"description": prim("string", "Description of the powerpack."),
						"name":        prim("string", "The name of the powerpack."),
					}),
				}),
			},
		}),
	}
}

// datastoreBothOperation is an id-optional singular data source: the tracked op is
// the by-id GET, and SearchOp points at the list GET (no query params, matching the
// real ListDatastores). Its element mirrors data_source_datadog_datastore.go.
func datastoreBothOperation() *model.Operation {
	listOp := &model.Operation{
		Path:            "/api/v2/actions-datastores",
		Method:          "GET",
		OperationId:     "ListDatastores",
		Tag:             "Actions Datastores",
		ResponseRefName: "DatastoreArray",
		ItemRefName:     "DatastoreData",
		ResponseSchema:  obj(map[string]*model.Schema{"data": {Kind: model.SchemaKindArray, Items: datastoreElement()}}),
	}
	return &model.Operation{
		Path:                "/api/v2/actions-datastores/{datastore_id}",
		Method:              "GET",
		OperationId:         "GetDatastore",
		Tag:                 "Actions Datastores",
		ResponseRefName:     "Datastore",
		ResponseDataRefName: "DatastoreData", // matches the list element → stays "both"
		SearchOp:            listOp,
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "datastore",
			TfDescription: "Use this data source to retrieve information about an existing Datadog datastore.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetDatastore", Search: "ListDatastores"},
		},
		ResponseSchema: obj(map[string]*model.Schema{"data": datastoreElement()}),
	}
}

// datastoreBothWithFilterOperation is datastoreBothOperation whose list op carries
// one scalar filter, exercising filter partitioning in the "both" shape: the by-id
// id field and the search filter coexist.
func datastoreBothWithFilterOperation() *model.Operation {
	op := datastoreBothOperation()
	op.SearchOp.QueryParams = []model.QueryParam{
		{Name: "filter[keyword]", Schema: prim("string", ""), Description: "Filter datastores by keyword."},
	}
	return op
}

// datastoreElement is the JSON:API datastore element ({id,type,attributes}) shared
// by the by-id and list responses, mirroring data_source_datadog_datastore.go.
func datastoreElement() *model.Schema {
	return obj(map[string]*model.Schema{
		"id":   prim("string", "The unique identifier of the datastore."),
		"type": prim("string", "The resource type for datastores."),
		"attributes": obj(map[string]*model.Schema{
			"created_at":                      {Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time", Description: "Timestamp when the datastore was created."},
			"creator_user_id":                 prim("integer", "The numeric ID of the user who created the datastore."),
			"creator_user_uuid":               prim("string", "The UUID of the user who created the datastore."),
			"description":                     prim("string", "A human-readable description about the datastore."),
			"modified_at":                     {Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time", Description: "Timestamp when the datastore was last modified."},
			"name":                            prim("string", "The display name of the datastore."),
			"org_id":                          prim("integer", "The ID of the organization that owns this datastore."),
			"primary_column_name":             prim("string", "The name of the primary key column for this datastore."),
			"primary_key_generation_strategy": {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"none", "uuid"}, Description: "Strategy for generating primary keys."},
		}),
	})
}

// prim and obj build model.Schema nodes for the emit fixtures (the model package
// keeps its own equivalents; these avoid a cross-package test dependency).
func prim(typ, desc string) *model.Schema {
	return &model.Schema{Kind: model.SchemaKindPrimitive, Type: typ, Description: desc}
}

func obj(props map[string]*model.Schema) *model.Schema {
	return &model.Schema{Kind: model.SchemaKindObject, Properties: props}
}

// incidentTypeOperation is the incident_type GET as a parser-shaped Operation: a
// JSON:API envelope ({data:{id,type,attributes:{description,is_default,name}}}).
func incidentTypeOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/incidents/config/types/{incident_type_id}",
		Method:          "GET",
		OperationId:     "GetIncidentType",
		Tag:             "Incidents",
		ResponseRefName: "IncidentTypeResponse",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "incident_type",
			TfDescription: "Use this data source to retrieve information about an existing incident type.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetIncidentType"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The incident type's ID."),
				"type": prim("string", "Incident type resource type."),
				"attributes": obj(map[string]*model.Schema{
					"name":        prim("string", "Name of the incident type."),
					"description": prim("string", "Description of the incident type."),
					"is_default":  prim("boolean", "Whether this incident type is the default type."),
				}),
			}),
		}),
	}
}

// incidentTypeArtifact resolves incidentTypeOperation into an *model.Artifact.
func incidentTypeArtifact() *model.Artifact {
	GinkgoHelper()
	art, err := model.BuildArtifact(incidentTypeOperation())
	Expect(err).NotTo(HaveOccurred())
	return art
}

// teamSingularOperation is the team GET-by-id as a parser-shaped Operation: a
// JSON:API envelope whose attributes carry scalar leaves plus two string arrays
// (visible_modules/hidden_modules), exercising collection-of-primitive hoisting.
func teamSingularOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/team/{team_id}",
		Method:          "GET",
		OperationId:     "GetTeam",
		Tag:             "Teams",
		ResponseRefName: "TeamResponse",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "team",
			TfDescription: "Use this data source to retrieve information about an existing Datadog team.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetTeam"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The team's identifier."),
				"type": prim("string", "Team resource type."),
				"attributes": obj(map[string]*model.Schema{
					"handle":          prim("string", "The team's handle."),
					"name":            prim("string", "The name of the team."),
					"visible_modules": {Kind: model.SchemaKindArray, Description: "Collection of visible modules for the team.", Items: prim("string", "String identifier of the module.")},
					"hidden_modules":  {Kind: model.SchemaKindArray, Description: "Collection of hidden modules for the team.", Items: prim("string", "String identifier of the module.")},
				}),
			}),
		}),
	}
}

// costBudgetOperation is the cost budget GET-by-id as a parser-shaped Operation: a
// JSON:API envelope whose attributes carry a name plus an entries array of objects,
// each holding scalars and a nested tag_filters array of objects — exercising
// recursive array-of-object hoisting.
func costBudgetOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/cost/budget/{budget_id}",
		Method:          "GET",
		OperationId:     "GetBudget",
		Tag:             "Cloud Cost Management",
		ResponseRefName: "BudgetWithEntries",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "cost_budget",
			TfDescription: "Use this data source to retrieve information about an existing Datadog cost budget.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetBudget"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The budget's identifier."),
				"type": prim("string", "Budget resource type."),
				"attributes": obj(map[string]*model.Schema{
					"name": prim("string", "The name of the budget."),
					"entries": {Kind: model.SchemaKindArray, Description: "The list of monthly budget entries.", Items: obj(map[string]*model.Schema{
						"amount": prim("number", "The budgeted amount for this entry."),
						"month":  prim("integer", "The month this budget entry applies to."),
						"tag_filters": {Kind: model.SchemaKindArray, Description: "The list of tag filters scoping this entry.", Items: obj(map[string]*model.Schema{
							"tag_key":   prim("string", "The tag key to filter on."),
							"tag_value": prim("string", "The tag value to filter on."),
						})},
					})},
				}),
			}),
		}),
	}
}

var _ = Describe("BuildDataSourceView singular nested arrays", func() {
	It("hoists an object array into a ListNestedBlock and recurses into nested object arrays", func() {
		view, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())

		Expect(view.Schema.Blocks).To(HaveLen(1))
		entries := view.Schema.Blocks[0]
		Expect(entries.TFName).To(Equal("entries"))
		Expect(entries.ListBlock).To(BeTrue())

		var entryAttrs []string
		for _, a := range entries.Attributes {
			entryAttrs = append(entryAttrs, a.TFName)
		}
		Expect(entryAttrs).To(Equal([]string{"amount", "month"}))
		Expect(entries.Blocks).To(HaveLen(1))
		Expect(entries.Blocks[0].TFName).To(Equal("tag_filters"))
		Expect(entries.Blocks[0].ListBlock).To(BeTrue())
	})

	It("generates a nested model struct per object level, parent first", func() {
		view, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())

		var names []string
		for _, m := range view.Models {
			names = append(names, m.Name)
		}
		Expect(names).To(Equal([]string{"datadogCostBudgetDataSourceModel", "datadogCostBudgetEntriesModel", "datadogCostBudgetEntriesTagFiltersModel"}))
	})

	It("maps each element through a guarded loop, recursing for nested arrays", func() {
		view, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())

		Expect(view.State.Lists).To(HaveLen(1))
		entries := view.State.Lists[0]
		Expect(entries.Kind).To(Equal("object"))
		Expect(entries.LHS).To(Equal("state.Entries"))
		Expect(entries.GetterOk).To(Equal("attributes.GetEntriesOk()"))
		Expect(entries.LoopVar).To(Equal("entriesItem"))
		Expect(entries.ElemVar).To(Equal("entriesModel"))
		Expect(entries.ElemStruct).To(Equal("datadogCostBudgetEntriesModel"))
		Expect(entries.Scalars).To(ContainElement(StateAssignment{
			Var: "amount", GetterOk: "entriesItem.GetAmountOk()",
			LHS: "entriesModel.Amount", RHS: "types.Float64Value(*amount)",
		}))

		Expect(entries.Lists).To(HaveLen(1))
		tagFilters := entries.Lists[0]
		Expect(tagFilters.Kind).To(Equal("object"))
		Expect(tagFilters.LHS).To(Equal("entriesModel.TagFilters"))
		Expect(tagFilters.GetterOk).To(Equal("entriesItem.GetTagFiltersOk()"))
		Expect(tagFilters.LoopVar).To(Equal("tagFiltersItem"))
		Expect(tagFilters.ElemStruct).To(Equal("datadogCostBudgetEntriesTagFiltersModel"))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})
})

var _ = Describe("BuildDataSourceView singular arrays", func() {
	It("renders unconstrained response values as normalized JSON", func() {
		op := teamSingularOperation()
		attrs := op.ResponseSchema.Properties["data"].Properties["attributes"].Properties
		attrs["metadata"] = &model.Schema{
			Kind: model.SchemaKindJSON, Description: "Arbitrary metadata.",
		}

		view := mustView(op)
		Expect(view.UsesJSON).To(BeTrue())
		var metadata AttrView
		for _, attr := range view.Schema.Attributes {
			if attr.TFName == "metadata" {
				metadata = attr
			}
		}
		Expect(metadata.TFType).To(Equal("schema.StringAttribute"))
		Expect(metadata.CustomType).To(Equal("jsontypes.NormalizedType{}"))
		Expect(view.State.Lists).To(ContainElement(ListAssignment{
			Kind: "json", LHS: "state.Metadata", GetterOk: "attributes.GetMetadataOk()",
			Var: "metadata", Path: "response.metadata",
		}))

		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		src := string(rendered)
		Expect(src).To(ContainSubstring(`"encoding/json"`))
		Expect(src).To(ContainSubstring(`"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"`))
		Expect(src).To(ContainSubstring("encoded, err := json.Marshal(metadata)"))
		Expect(src).To(ContainSubstring("state.Metadata = jsontypes.NewNormalizedValue(string(encoded))"))
	})

	It("hoists a string array under attributes into a ListAttribute carrying its element type", func() {
		view, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())

		attrs := map[string]AttrView{}
		for _, a := range view.Schema.Attributes {
			attrs[a.TFName] = a
		}
		Expect(attrs["visible_modules"].TFType).To(Equal("schema.ListAttribute"))
		Expect(attrs["visible_modules"].ElementType).To(Equal("types.StringType"))
		Expect(attrs["visible_modules"].Computed).To(BeTrue())
		Expect(view.Schema.Blocks).To(BeEmpty(), "a collection-of-primitive is a leaf attribute, not a block")
	})

	It("declares the list field as a types.List in the model", func() {
		view, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())

		goTypes := map[string]string{}
		for _, f := range view.Models[0].Fields {
			goTypes[f.TFName] = f.GoType
		}
		Expect(goTypes["visible_modules"]).To(Equal("types.List"))
		Expect(goTypes["hidden_modules"]).To(Equal("types.List"))
	})

	It("maps each list through a guarded ListValueFrom assignment", func() {
		view, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())

		// Sorted by attribute name, both string arrays become guarded primitive lists.
		Expect(view.State.Lists).To(Equal([]ListAssignment{
			{Kind: "primitive", ContainerKind: "list", LHS: "state.HiddenModules", GetterOk: "attributes.GetHiddenModulesOk()", Var: "hiddenModules", ElementType: "types.StringType"},
			{Kind: "primitive", ContainerKind: "list", LHS: "state.VisibleModules", GetterOk: "attributes.GetVisibleModulesOk()", Var: "visibleModules", ElementType: "types.StringType"},
		}))
	})

	It("emits recursive list/map types and the matching state constructors", func() {
		op := teamSingularOperation()
		attrs := op.ResponseSchema.Properties["data"].Properties["attributes"].Properties
		attrs["host_tags_lists"] = &model.Schema{
			Kind: model.SchemaKindArray, Description: "Host tag conjunctions.",
			Items: &model.Schema{Kind: model.SchemaKindArray, Items: prim("string", "A host tag.")},
		}
		attrs["group_bys"] = &model.Schema{
			Kind: model.SchemaKindMap, Description: "Named grouping dimensions.",
			Items: &model.Schema{Kind: model.SchemaKindArray, Items: prim("string", "A grouping dimension.")},
		}
		attrs["cloud_provider"] = &model.Schema{
			Kind: model.SchemaKindMap, Description: "Cloud provider filters.",
			Items: &model.Schema{Kind: model.SchemaKindMap, Items: &model.Schema{
				Kind: model.SchemaKindArray, Items: prim("string", "A filter value."),
			}},
		}

		view := mustView(op)
		attributes := map[string]AttrView{}
		for _, attr := range view.Schema.Attributes {
			attributes[attr.TFName] = attr
		}
		Expect(attributes["host_tags_lists"].TFType).To(Equal("schema.ListAttribute"))
		Expect(attributes["host_tags_lists"].ElementType).To(Equal("types.ListType{ElemType: types.StringType}"))
		Expect(attributes["group_bys"].TFType).To(Equal("schema.MapAttribute"))
		Expect(attributes["group_bys"].ElementType).To(Equal("types.ListType{ElemType: types.StringType}"))
		Expect(attributes["cloud_provider"].TFType).To(Equal("schema.MapAttribute"))
		Expect(attributes["cloud_provider"].ElementType).To(Equal("types.MapType{ElemType: types.ListType{ElemType: types.StringType}}"))

		assignments := map[string]ListAssignment{}
		for _, assignment := range view.State.Lists {
			assignments[assignment.LHS] = assignment
		}
		Expect(assignments["state.HostTagsLists"].ContainerKind).To(Equal("list"))
		Expect(assignments["state.GroupBys"].ContainerKind).To(Equal("map"))
		Expect(assignments["state.CloudProvider"].ContainerKind).To(Equal("map"))

		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		src := string(rendered)
		Expect(src).To(ContainSubstring("types.ListValueFrom(ctx, types.ListType{ElemType: types.StringType}, *hostTagsLists)"))
		Expect(src).To(ContainSubstring("types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, *groupBys)"))
		Expect(src).To(ContainSubstring("types.MapNull(types.MapType{ElemType: types.ListType{ElemType: types.StringType}})"))
	})

	It("emits a typed map nested inside an object-list element", func() {
		op := costBudgetOperation()
		entry := op.ResponseSchema.Properties["data"].Properties["attributes"].Properties["entries"].Items
		entry.Properties["metadata"] = &model.Schema{
			Kind: model.SchemaKindMap, Description: "Entry metadata.",
			Items: prim("string", "A metadata value."),
		}

		view := mustView(op)
		entries := view.State.Lists[0]
		nested := map[string]ListAssignment{}
		for _, assignment := range entries.Lists {
			nested[assignment.LHS] = assignment
		}
		Expect(nested["entriesModel.Metadata"].Kind).To(Equal("primitive"))
		Expect(nested["entriesModel.Metadata"].ContainerKind).To(Equal("map"))
		Expect(nested["entriesModel.Metadata"].ElementType).To(Equal("types.StringType"))
		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(rendered)).To(ContainSubstring("types.MapValueFrom(ctx, types.StringType, *metadata)"))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})
})

var _ = Describe("BuildDataSourceView plural", func() {
	It("builds the teams plural view end-to-end, matching the golden-backing fixture", func() {
		art, err := model.BuildArtifact(teamsOperation())
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(view).To(Equal(pluralFixture()))
	})

	It("renders unconstrained result-item values as normalized JSON", func() {
		op := teamsOperation()
		attributes := op.ResponseSchema.Properties["data"].Items.Properties["attributes"].Properties
		attributes["metadata"] = &model.Schema{
			Kind: model.SchemaKindJSON, Description: "Arbitrary metadata.",
		}

		view := mustView(op)
		Expect(view.UsesJSON).To(BeTrue())
		Expect(view.State.ItemLists).To(ContainElement(ListAssignment{
			Kind: "json", LHS: "r.Metadata", GetterOk: "item.Attributes.GetMetadataOk()",
			Var: "metadata", Path: "response.data[].attributes.metadata",
		}))

		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		src := string(rendered)
		Expect(src).To(ContainSubstring("r.Metadata = jsontypes.NewNormalizedValue(string(encoded))"))
	})

	It("emits a recursively typed map on a plural result item", func() {
		op := teamsOperation()
		attributes := op.ResponseSchema.Properties["data"].Items.Properties["attributes"].Properties
		attributes["group_bys"] = &model.Schema{
			Kind: model.SchemaKindMap, Description: "Named grouping dimensions.",
			Items: &model.Schema{Kind: model.SchemaKindArray, Items: prim("string", "A grouping dimension.")},
		}

		view := mustView(op)
		itemCollections := map[string]ListAssignment{}
		for _, assignment := range view.State.ItemLists {
			itemCollections[assignment.LHS] = assignment
		}
		Expect(itemCollections["r.GroupBys"].Kind).To(Equal("primitive"))
		Expect(itemCollections["r.GroupBys"].ContainerKind).To(Equal("map"))
		Expect(itemCollections["r.GroupBys"].ElementType).To(Equal("types.ListType{ElemType: types.StringType}"))
		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(rendered)).To(ContainSubstring("types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, *groupBys)"))
	})

	It("drops array and enum query params from the filter set", func() {
		art, err := model.BuildArtifact(teamsOperation())
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())

		var names []string
		for _, f := range view.Schema.Attributes {
			names = append(names, f.TFName)
		}
		Expect(names).To(Equal([]string{"filter_keyword", "filter_me"}))
	})

	It("renders UUID response values at the item root, in attributes, and in nested objects", func() {
		op := pluralNestedOperation()
		item := op.ResponseSchema.Properties["data"].Items
		item.Properties["id"].Format = "uuid"
		item.Properties["attributes"].Properties["owner_id"] = &model.Schema{
			Kind: model.SchemaKindPrimitive, Type: "string", Format: "uuid", Description: "The owner UUID.",
		}
		item.Properties["attributes"].Properties["parts"].Items.Properties["label"].Format = "uuid"

		view := mustView(op)
		fields := map[string]string{}
		for _, assignment := range view.State.ItemFields {
			fields[assignment.LHS] = assignment.RHS
		}
		Expect(fields["ID"]).To(Equal("types.StringValue(item.GetId().String())"))
		Expect(fields["OwnerId"]).To(Equal("types.StringValue(item.Attributes.GetOwnerId().String())"))
		Expect(view.State.ItemLists[0].Scalars).To(ContainElement(StateAssignment{
			Var: "label", GetterOk: "partsItem.GetLabelOk()", LHS: "partsModel.Label", RHS: "types.StringValue(label.String())",
		}))
	})

	It("parses an optional UUID filter in the plural Read path", func() {
		op := teamsOperation()
		filter := op.QueryParams[0]
		filter.Schema.Format = "uuid"
		op.SDKBinding = &model.SDKOperationBinding{
			OptionalParamsType: "ListTeamsOptionalParameters",
			Optional: []model.SDKArgument{
				{Name: filter.Name, GoName: "filterKeyword", GoType: "uuid.UUID", Location: "query", Schema: filter.Schema, Setter: "WithFilterKeyword"},
				{Name: "filter[me]", GoName: "filterMe", GoType: "bool", Location: "query", Schema: op.QueryParams[1].Schema, Setter: "WithFilterMe"},
			},
		}

		view := mustView(op)
		Expect(view.UsesUUID).To(BeTrue())
		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		src := string(rendered)
		Expect(src).To(ContainSubstring(`parsedFilterKeyword, err := uuid.Parse(state.FilterKeyword.ValueString())`))
		Expect(src).To(ContainSubstring(`optionalParams.WithFilterKeyword(parsedFilterKeyword)`))
	})

	It("parses a date filter in the plural Read path", func() {
		op := teamsOperation()
		filter := op.QueryParams[0]
		filter.Schema.Format = "date"
		op.SDKBinding = &model.SDKOperationBinding{
			OptionalParamsType: "ListTeamsOptionalParameters",
			Optional: []model.SDKArgument{
				{Name: filter.Name, GoName: "filterKeyword", GoType: "time.Time", Location: "query", Schema: filter.Schema, Setter: "WithFilterKeyword"},
				{Name: "filter[me]", GoName: "filterMe", GoType: "bool", Location: "query", Schema: op.QueryParams[1].Schema, Setter: "WithFilterMe"},
			},
		}

		view := mustView(op)
		Expect(view.UsesTime).To(BeTrue())
		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		src := string(rendered)
		Expect(src).To(ContainSubstring(`parsedFilterKeyword, err := time.Parse(time.DateOnly, state.FilterKeyword.ValueString())`))
		Expect(src).To(ContainSubstring(`optionalParams.WithFilterKeyword(parsedFilterKeyword)`))
	})

	It("uses the pinned setter type rather than UUID schema format for input syntax", func() {
		op := teamsOperation()
		filter := op.QueryParams[0]
		filter.Schema.Format = "uuid"
		op.SDKBinding = &model.SDKOperationBinding{
			OptionalParamsType: "ListTeamsOptionalParameters",
			Optional: []model.SDKArgument{
				{Name: filter.Name, GoName: "filterKeyword", GoType: "string", Location: "query", Schema: filter.Schema, Setter: "WithFilterKeyword"},
				{Name: "filter[me]", GoName: "filterMe", GoType: "bool", Location: "query", Schema: op.QueryParams[1].Schema, Setter: "WithFilterMe"},
			},
		}

		view := mustView(op)
		Expect(view.UsesUUID).To(BeFalse())
		Expect(view.Read.Filters[0]).To(Equal(FilterParamView{
			StateField: "FilterKeyword", ParamField: "FilterKeyword",
			ValueExpr: "state.FilterKeyword.ValueString()", Setter: "WithFilterKeyword",
		}))
		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(rendered)).NotTo(ContainSubstring("github.com/google/uuid"))
	})

	It("hashes a fixed seed when an endpoint has no filters", func() {
		op := teamsOperation()
		op.QueryParams = nil
		op.Pagination = nil
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Read.Filters).To(BeEmpty())
		Expect(view.Read.OptionalParamsType).To(BeEmpty())
		Expect(view.Schema.Attributes).To(BeEmpty())
	})

	It("includes required positional SDK inputs in the plural identity hash", func() {
		op := teamsOperation()
		op.Path = "/api/v2/accounts/{account_id}/team"
		op.SDKBinding = &model.SDKOperationBinding{
			Required: []model.SDKArgument{
				{Name: "account_id", GoName: "accountId", GoType: "int64", Location: "path", Schema: prim("integer", "The account ID.")},
			},
			OptionalParamsType: "ListTeamsOptionalParameters",
			Optional: []model.SDKArgument{
				{Name: "filter[keyword]", GoName: "filterKeyword", GoType: "string", Location: "query", Schema: prim("string", ""), Setter: "WithFilterKeyword"},
				{Name: "filter[me]", GoName: "filterMe", GoType: "bool", Location: "query", Schema: prim("boolean", ""), Setter: "WithFilterMe"},
			},
		}

		view, err := BuildDataSourceView(mustArtifact(op))
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Read.HashInputs).To(Equal([]FilterParamView{
			{StateField: "AccountId", ValueExpr: "ValueInt64Pointer()"},
			{StateField: "FilterKeyword", ValueExpr: "ValueStringPointer()"},
			{StateField: "FilterMe", ValueExpr: "ValueBoolPointer()"},
		}))
		rendered, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(rendered)).To(ContainSubstring(`fmt.Sprintf("%d:%s:%t", state.AccountId.ValueInt64(), state.FilterKeyword.ValueString(), state.FilterMe.ValueBool())`))
	})

	It("produces a deeply-equal plural view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(teamsOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(teamsOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})

	DescribeTable("fail-slows unsupported item-element nodes into one UnsupportedEmitError",
		func(mutate func(*model.Operation), wantReason string) {
			op := teamsOperation()
			mutate(op)
			view, err := BuildDataSourceView(mustArtifact(op))
			var uerr *UnsupportedEmitError
			Expect(errors.As(err, &uerr)).To(BeTrue(), "expected an UnsupportedEmitError, got %v", err)
			Expect(uerr.Error()).To(ContainSubstring(wantReason))
			Expect(view).To(Equal(DataSourceView{}), "no view should be produced on failure")
		},
		Entry("a missing item element type",
			func(op *model.Operation) { op.ItemRefName = "" },
			"missing list item type"),
	)
})

// mustArtifact builds an Artifact from op or fails the test.
func mustArtifact(op *model.Operation) *model.Artifact {
	GinkgoHelper()
	art, err := model.BuildArtifact(op)
	Expect(err).NotTo(HaveOccurred())
	return art
}

// teamsOperation is the teams list GET as a parser-shaped Operation: a paginated
// JSON:API collection whose response carries metadata siblings (meta/links/
// included) the builder must drop, plus array and enum query params it must drop
// from the filter set. Descriptions mirror the golden-backing pluralFixture.
func teamsOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/team",
		Method:          "GET",
		OperationId:     "ListTeams",
		Tag:             "Teams",
		ResponseRefName: "TeamsResponse",
		ItemRefName:     "Team",
		Pagination:      &model.Pagination{LimitParam: "page[size]", PageParam: "page[number]", ResultsPath: "data"},
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "teams",
			Cardinality:   model.CardinalityPlural,
			TfDescription: "Use this data source to retrieve information about existing teams for use in other resources.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "ListTeams"},
		},
		QueryParams: []model.QueryParam{
			{Name: "filter[keyword]", Schema: prim("string", ""), Description: "Search query. Can be team name, team handle, or email of team member."},
			{Name: "filter[me]", Schema: prim("boolean", ""), Description: "When true, only returns teams the current user belongs to."},
			{Name: "include", Schema: &model.Schema{Kind: model.SchemaKindArray, Items: prim("string", "")}},
			{Name: "page[number]", Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "integer", Format: "int64"}},
			{Name: "page[size]", Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "integer", Format: "int64"}},
			{Name: "sort", Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"name"}}},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind:        model.SchemaKindArray,
				Description: "List of teams",
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The team's identifier."),
					"type": prim("string", "Team resource type."),
					"attributes": obj(map[string]*model.Schema{
						"description":     prim("string", "Free-form markdown description/content for the team's homepage."),
						"handle":          prim("string", "The team's handle."),
						"hidden_modules":  {Kind: model.SchemaKindArray, Description: "Collection of hidden modules for the team.", Items: prim("string", "String identifier of the module.")},
						"link_count":      prim("integer", "The number of links belonging to the team."),
						"name":            prim("string", "The name of the team."),
						"summary":         prim("string", "A brief summary of the team, derived from the `description`."),
						"user_count":      prim("integer", "The number of users belonging to the team."),
						"visible_modules": {Kind: model.SchemaKindArray, Description: "Collection of visible modules for the team.", Items: prim("string", "String identifier of the module.")},
					}),
				}),
			},
			// Response metadata siblings: the model keeps only the results array.
			"meta":     obj(map[string]*model.Schema{"x": prim("string", "")}),
			"links":    obj(map[string]*model.Schema{"x": prim("string", "")}),
			"included": {Kind: model.SchemaKindArray, Items: obj(map[string]*model.Schema{"x": prim("string", "")})},
		}),
	}
}

// datastoresOperation is the ListDatastores GET: a non-paginated list with no
// query parameters, exercising the no-optional-params call form and the
// zero-filter id seed. Its element attributes include a date-time and an enum,
// mirroring the singular data_source_datadog_datastore.go mapping.
func datastoresOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/actions-datastores",
		Method:          "GET",
		OperationId:     "ListDatastores",
		Tag:             "Actions Datastores",
		ResponseRefName: "DatastoreArray",
		ItemRefName:     "DatastoreData",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "datastores",
			Cardinality:   model.CardinalityPlural,
			TfDescription: "Use this data source to retrieve information about existing Datadog datastores.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "ListDatastores"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind:        model.SchemaKindArray,
				Description: "An array of datastore objects containing their configurations and metadata.",
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The unique identifier of the datastore."),
					"type": prim("string", "The resource type for datastores."),
					"attributes": obj(map[string]*model.Schema{
						"created_at":                      {Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time", Description: "Timestamp when the datastore was created."},
						"creator_user_id":                 prim("integer", "The numeric ID of the user who created the datastore."),
						"creator_user_uuid":               prim("string", "The UUID of the user who created the datastore."),
						"description":                     prim("string", "A human-readable description about the datastore."),
						"modified_at":                     {Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time", Description: "Timestamp when the datastore was last modified."},
						"name":                            prim("string", "The display name of the datastore."),
						"org_id":                          prim("integer", "The ID of the organization that owns this datastore."),
						"primary_column_name":             prim("string", "The name of the primary key column for this datastore."),
						"primary_key_generation_strategy": {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"none", "uuid"}, Description: "Strategy for generating primary keys."},
					}),
				}),
			},
		}),
	}
}

// pluralNestedOperation is a synthetic plural list whose item attributes carry an
// object array (parts), each part an object with scalars — exercising the
// array-of-object item path that walks an element struct inside buildPluralView.
func pluralNestedOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/widgets",
		Method:          "GET",
		OperationId:     "ListWidgets",
		Tag:             "Widgets",
		ResponseRefName: "WidgetsResponse",
		ItemRefName:     "Widget",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "widgets",
			Cardinality:   model.CardinalityPlural,
			TfDescription: "Use this data source to retrieve information about existing widgets.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "ListWidgets"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind:        model.SchemaKindArray,
				Description: "List of widgets",
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The widget's identifier."),
					"type": prim("string", "Widget resource type."),
					"attributes": obj(map[string]*model.Schema{
						"name": prim("string", "The name of the widget."),
						"parts": {Kind: model.SchemaKindArray, Description: "The parts that make up the widget.", Items: obj(map[string]*model.Schema{
							"label":    prim("string", "The label of the part."),
							"quantity": prim("integer", "How many of this part the widget uses."),
						})},
					}),
				}),
			},
		}),
	}
}

var _ = Describe("BuildDataSourceView plural nested arrays", func() {
	It("renders an object array in an item as a ListNestedBlock with a generated element struct", func() {
		view, err := BuildDataSourceView(mustArtifact(pluralNestedOperation()))
		Expect(err).NotTo(HaveOccurred())

		items := view.Schema.Blocks[0]
		Expect(items.TFName).To(Equal("widgets"))
		Expect(items.Blocks).To(HaveLen(1))
		Expect(items.Blocks[0].TFName).To(Equal("parts"))
		Expect(items.Blocks[0].ListBlock).To(BeTrue())

		var names []string
		for _, m := range view.Models {
			names = append(names, m.Name)
		}
		Expect(names).To(Equal([]string{"datadogWidgetsDataSourceModel", "datadogWidgetsWidgetModel", "datadogWidgetsWidgetPartsModel"}))
	})

	It("maps the object array off item.Attributes into the item accumulator", func() {
		view, err := BuildDataSourceView(mustArtifact(pluralNestedOperation()))
		Expect(err).NotTo(HaveOccurred())

		Expect(view.State.ItemLists).To(HaveLen(1))
		parts := view.State.ItemLists[0]
		Expect(parts.Kind).To(Equal("object"))
		Expect(parts.LHS).To(Equal("r.Parts"))
		Expect(parts.GetterOk).To(Equal("item.Attributes.GetPartsOk()"))
		Expect(parts.LoopVar).To(Equal("partsItem"))
		Expect(parts.ElemStruct).To(Equal("datadogWidgetsWidgetPartsModel"))
		Expect(parts.Scalars).To(ContainElement(StateAssignment{
			Var: "label", GetterOk: "partsItem.GetLabelOk()",
			LHS: "partsModel.Label", RHS: "types.StringValue(*label)",
		}))
	})

	It("renders a nested Go-keyword attribute through a safe guarded local", func() {
		op := pluralNestedOperation()
		op.ResponseSchema.Properties["data"].Items.Properties["attributes"].
			Properties["parts"].Items.Properties["type"] = prim("string", "The part type.")

		got, err := RenderDataSource(mustView(op))
		Expect(err).NotTo(HaveOccurred())
		src := string(got)
		Expect(src).To(ContainSubstring(
			"if typeVar, ok := partsItem.GetTypeOk(); ok && typeVar != nil {",
		))
		Expect(src).To(ContainSubstring("partsModel.Type = types.StringValue(*typeVar)"))
		Expect(src).NotTo(ContainSubstring("if type, ok :="))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(pluralNestedOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(pluralNestedOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})

	It("scopes the shared SDK item type to the Terraform artifact", func() {
		teams := mustView(teamsOperation())
		roleUsersOperation := teamsOperation()
		roleUsersOperation.Tracking.ArtifactName = "role_users"
		roleUsers := mustView(roleUsersOperation)

		Expect(teams.State.ItemStruct).To(Equal("datadogTeamsTeamModel"))
		Expect(roleUsers.State.ItemStruct).To(Equal("datadogRoleUsersTeamModel"))
	})
})

// retentionFilterOperation is the apm retention filter GET-by-id as a parser-shaped
// Operation: a JSON:API envelope whose attributes carry scalars plus a nested filter
// object — itself holding a scalar, a string array, and a nested metadata object —
// exercising bare-object hoisting, recursion, and composition with arrays.
func retentionFilterOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/apm/config/retention-filters/{filter_id}",
		Method:          "GET",
		OperationId:     "GetApmRetentionFilter",
		Tag:             "APM Retention Filters",
		ResponseRefName: "RetentionFilterResponse",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "apm_retention_filter",
			TfDescription: "Use this data source to retrieve information about an existing APM retention filter.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetApmRetentionFilter"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The retention filter's ID."),
				"type": prim("string", "Retention filter resource type."),
				"attributes": obj(map[string]*model.Schema{
					"enabled": prim("boolean", "Whether the retention filter is active."),
					"name":    prim("string", "The name of the retention filter."),
					"filter": obj(map[string]*model.Schema{
						"query": prim("string", "The search query defining the filter."),
						"tags":  {Kind: model.SchemaKindArray, Description: "Tags scoping the filter.", Items: prim("string", "A tag identifier.")},
						"metadata": obj(map[string]*model.Schema{
							"created_by": prim("string", "Handle of the user who created the filter."),
						}),
					}),
				}),
			}),
		}),
	}
}

var _ = Describe("BuildDataSourceView singular nested objects", func() {
	It("hoists a bare object under attributes into a SingleNestedBlock", func() {
		view, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())

		blocks := map[string]AttrView{}
		for _, b := range view.Schema.Blocks {
			blocks[b.TFName] = b
		}
		Expect(blocks).To(HaveKey("filter"))
		Expect(blocks["filter"].ListBlock).To(BeFalse(), "a bare object is a SingleNestedBlock, not a list block")
	})

	It("generates one model struct per object level, parent first", func() {
		view, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())

		var names []string
		for _, m := range view.Models {
			names = append(names, m.Name)
		}
		Expect(names).To(Equal([]string{"datadogApmRetentionFilterDataSourceModel", "datadogApmRetentionFilterFilterModel", "datadogApmRetentionFilterFilterMetadataModel"}))
	})

	It("maps the object through a guarded assignment, recursing into the nested object", func() {
		view, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())

		var filter ListAssignment
		for _, l := range view.State.Lists {
			if l.LHS == "state.Filter" {
				filter = l
			}
		}
		Expect(filter.Kind).To(Equal("object_single"))
		Expect(filter.GetterOk).To(Equal("attributes.GetFilterOk()"))
		Expect(filter.Var).To(Equal("filter"))
		Expect(filter.ElemVar).To(Equal("filterModel"))
		Expect(filter.ElemStruct).To(Equal("datadogApmRetentionFilterFilterModel"))
		Expect(filter.Scalars).To(ContainElement(StateAssignment{
			Var: "query", GetterOk: "filter.GetQueryOk()",
			LHS: "filterModel.Query", RHS: "types.StringValue(*query)",
		}))

		var metadata ListAssignment
		for _, l := range filter.Lists {
			if l.Kind == "object_single" {
				metadata = l
			}
		}
		Expect(metadata.LHS).To(Equal("filterModel.Metadata"))
		Expect(metadata.GetterOk).To(Equal("filter.GetMetadataOk()"))
		Expect(metadata.ElemStruct).To(Equal("datadogApmRetentionFilterFilterMetadataModel"))
	})

	It("renders the guarded object block and the recursive assignment in updateState", func() {
		got, err := RenderDataSource(mustView(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())
		src := string(got)
		Expect(src).To(ContainSubstring("if filter, ok := attributes.GetFilterOk(); ok && filter != nil {"))
		Expect(src).To(ContainSubstring("state.Filter = filterModel"))
		Expect(src).To(ContainSubstring("if metadata, ok := filter.GetMetadataOk(); ok && metadata != nil {"))
		Expect(src).To(ContainSubstring("filterModel.Metadata = metadataModel"))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})

	It("includes the full parent path when two branches reuse the same leaf name", func() {
		op := retentionFilterOperation()
		attrs := op.ResponseSchema.Properties["data"].Properties["attributes"].Properties
		attrs["primary"] = obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{"value": prim("string", "The primary value.")}),
		})
		attrs["secondary"] = obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{"value": prim("string", "The secondary value.")}),
		})

		view := mustView(op)
		names := make([]string, 0, len(view.Models))
		for _, modelView := range view.Models {
			names = append(names, modelView.Name)
		}
		Expect(names).To(Equal([]string{
			"datadogApmRetentionFilterDataSourceModel",
			"datadogApmRetentionFilterFilterModel",
			"datadogApmRetentionFilterFilterMetadataModel",
			"datadogApmRetentionFilterPrimaryModel",
			"datadogApmRetentionFilterPrimaryDataModel",
			"datadogApmRetentionFilterSecondaryModel",
			"datadogApmRetentionFilterSecondaryDataModel",
		}))
	})
})

// pluralObjectOperation is a synthetic plural list whose item attributes carry a
// bare object (spec) of scalars — exercising the single-object item path that walks
// an element struct inside buildPluralView.
func pluralObjectOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/gizmos",
		Method:          "GET",
		OperationId:     "ListGizmos",
		Tag:             "Gizmos",
		ResponseRefName: "GizmosResponse",
		ItemRefName:     "Gizmo",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "gizmos",
			Cardinality:   model.CardinalityPlural,
			TfDescription: "Use this data source to retrieve information about existing gizmos.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "ListGizmos"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind:        model.SchemaKindArray,
				Description: "List of gizmos",
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The gizmo's identifier."),
					"type": prim("string", "Gizmo resource type."),
					"attributes": obj(map[string]*model.Schema{
						"name": prim("string", "The name of the gizmo."),
						"spec": obj(map[string]*model.Schema{
							"shape": prim("string", "The shape of the gizmo."),
							"size":  prim("integer", "The number of segments."),
						}),
					}),
				}),
			},
		}),
	}
}

var _ = Describe("BuildDataSourceView plural nested objects", func() {
	It("renders a bare object in an item as a SingleNestedBlock with a generated struct", func() {
		view, err := BuildDataSourceView(mustArtifact(pluralObjectOperation()))
		Expect(err).NotTo(HaveOccurred())

		items := view.Schema.Blocks[0]
		Expect(items.TFName).To(Equal("gizmos"))
		var spec AttrView
		for _, b := range items.Blocks {
			if b.TFName == "spec" {
				spec = b
			}
		}
		Expect(spec.TFName).To(Equal("spec"))
		Expect(spec.ListBlock).To(BeFalse())

		var names []string
		for _, m := range view.Models {
			names = append(names, m.Name)
		}
		Expect(names).To(Equal([]string{"datadogGizmosDataSourceModel", "datadogGizmosGizmoModel", "datadogGizmosGizmoSpecModel"}))
	})

	It("maps the object off item.Attributes into the item accumulator", func() {
		view, err := BuildDataSourceView(mustArtifact(pluralObjectOperation()))
		Expect(err).NotTo(HaveOccurred())

		Expect(view.State.ItemLists).To(HaveLen(1))
		spec := view.State.ItemLists[0]
		Expect(spec.Kind).To(Equal("object_single"))
		Expect(spec.LHS).To(Equal("r.Spec"))
		Expect(spec.GetterOk).To(Equal("item.Attributes.GetSpecOk()"))
		Expect(spec.Var).To(Equal("spec"))
		Expect(spec.ElemStruct).To(Equal("datadogGizmosGizmoSpecModel"))
		Expect(spec.Scalars).To(ContainElement(StateAssignment{
			Var: "shape", GetterOk: "spec.GetShapeOk()",
			LHS: "specModel.Shape", RHS: "types.StringValue(*shape)",
		}))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(pluralObjectOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(pluralObjectOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})
})
