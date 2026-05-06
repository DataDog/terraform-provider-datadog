package model

// TemplateData is the top-level input passed to Go templates.
type TemplateData struct {
	PackageName      string
	ResourceName     string // e.g., "team"
	TypeName         string // e.g., "Team" (PascalCase)
	Description      string // Data source description
	SDKImportPath    string // e.g., "datadogV2"
	SDKApiType       string // e.g., "TeamsApi"
	SDKApiAccessor   string // e.g., "GetTeamsApiV2"
	SDKReadMethod    string // e.g., "GetTeam"
	SDKResponseType  string // e.g., "TeamResponse" (from OpenAPI $ref)
	HasListFallback  bool   // true when list config is present
	SDKListMethod    string // e.g., "ListTeams"
	SDKListOptParams string // e.g., "ListTeamsOptionalParameters"
	Model            ModelData
	Schema           SchemaData
	Read             ReadData
	State            StateData
}

// ModelData describes the model struct(s) for the template.
type ModelData struct {
	StructName    string
	Fields        []ModelField
	NestedStructs []ModelData
}

// ModelField describes a single field in a model struct.
type ModelField struct {
	GoName    string // PascalCase field name
	TfsdkTag  string // snake_case tfsdk tag
	GoType    string // e.g., "types.String"
	IsPointer bool   // For optional nested blocks
}

// SchemaData describes the Terraform schema for the template.
type SchemaData struct {
	Attributes      []AttributeData
	Blocks          []BlockData
	HasListFallback bool // Controls whether path params are Optional+Computed
}

// AttributeData describes a single Terraform schema attribute.
type AttributeData struct {
	Name             string // tfsdk name
	Type             string // e.g., "schema.StringAttribute"
	ElementType      string // For ListAttribute, e.g., "types.StringType"
	Description      string
	Required         bool
	Optional         bool
	Computed         bool
	IsNestedMap      bool            // true for FieldTypeMapOfObjects (renders MapNestedAttribute)
	NestedAttributes []AttributeData // children for MapNestedAttribute
}

// BlockData describes a nested Terraform schema block.
type BlockData struct {
	Name           string
	Type           string // "SingleNestedBlock" or "ListNestedBlock"
	Description    string
	Attributes     []AttributeData
	Blocks         []BlockData
	IsVariantBlock bool // true for oneOf/anyOf variant blocks
}

// NeedsNestedObject returns true for block types that require a NestedObject wrapper
// (ListNestedBlock, SetNestedBlock). SingleNestedBlock places Attributes directly.
func (b BlockData) NeedsNestedObject() bool {
	return b.Type == "ListNestedBlock" || b.Type == "SetNestedBlock"
}

// ReadData describes parameters for the Read method template.
type ReadData struct {
	PathParams       []ParamData
	ResponseAccessor string // e.g., ".GetData()"
	IsJSONAPI        bool
	HasListFallback  bool
	FilterParams     []ParamData // Filter query params from list endpoint
	SDKListMethod    string      // e.g., "ListTeams"
	SDKListOptParams string      // e.g., "ListTeamsOptionalParameters"
}

// ParamData describes a single parameter for method calls.
type ParamData struct {
	Name        string // snake_case
	GoName      string // PascalCase
	ValueMethod string // e.g., "ValueString()"
}

// StateData describes field mappings for the updateState method.
type StateData struct {
	IDAccessor    string // e.g., "resp.Data.GetId()"
	FieldMappings []FieldMapping
	VariantBlocks []VariantBlockMapping
}

// VariantBlockMapping describes oneOf/anyOf variant probing in updateState.
type VariantBlockMapping struct {
	ParentField string
	Variants    []VariantMapping
}

// VariantMapping describes a single variant's state mapping.
type VariantMapping struct {
	BlockName     string // e.g., "config_type_a"
	SDKOkAccessor string // e.g., "GetConfigTypeAOk()"
	HelperFunc    string // e.g., "updateConfigTypeAState"
	ModelField    string // e.g., "ConfigTypeA"
}

// FieldMapping describes how to map an SDK response field to a model field.
type FieldMapping struct {
	ModelField      string // e.g., "state.Name"
	SDKAccessor     string // e.g., "attributes.GetName()"
	SDKOkAccessor   string // e.g., "attributes.GetNameOk()" (for nullable)
	TypeConverter   string // e.g., "types.StringValue"
	NullValue       string // e.g., "types.StringNull()" (for nullable)
	NeedsCast       string // e.g., "int64" or ""
	Nullable        bool   // true when OpenAPI field has nullable: true
	IsList          bool   // true for array-of-primitives
	ListElementType string // e.g., "types.StringType" (for list fields)
	IsMap           bool   // true for additionalProperties fields
	MapElementType  string // e.g., "types.StringType" (for map fields)
}
