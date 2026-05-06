package codegen

import (
	"fmt"
	"log"
	"strings"

	"github.com/DataDog/terraform-provider-datadog/generator/internal/model"
	"github.com/DataDog/terraform-provider-datadog/generator/internal/openapi"
	"github.com/DataDog/terraform-provider-datadog/generator/internal/templates"
	"github.com/DataDog/terraform-provider-datadog/generator/internal/types"
)

// BuildTemplateData assembles a TemplateData from a parsed operation and schema.
// When listOp is non-nil, filter-fallback lookup is enabled.
func BuildTemplateData(name string, op *openapi.ParsedOperation, schema *openapi.SchemaObject, isJSONAPI bool, listOp *openapi.ParsedOperation) (*model.TemplateData, error) {
	typeName := templates.ToPascalCase(name)
	tag := op.Tag
	if tag == "" {
		return nil, fmt.Errorf("operation %s %s has no tag; a tag is required for SDK client derivation (e.g., \"Teams\")", op.Method, op.Path)
	}

	// Derive SDK names from the tag (sanitize spaces for valid Go identifiers)
	sdkImportPath := deriveSDKImportPath(op.Path)
	sanitizedTag := templates.ToPascalCase(tag)
	sdkApiType := sanitizedTag + "Api"
	sdkApiAccessor := "Get" + sanitizedTag + "Api" + deriveVersion(op.Path)
	sdkReadMethod := op.OperationID

	// Derive response type from $ref (Phase 8: 11a)
	// Prefer the $ref name; fall back to OperationID+"Response" for inline schemas.
	sdkResponseType := op.ResponseTypeName
	if sdkResponseType == "" {
		sdkResponseType = op.OperationID + "Response"
		log.Printf("Warning: response schema for %s %s has no $ref; falling back to %q", op.Method, op.Path, sdkResponseType)
	}

	data := &model.TemplateData{
		PackageName:     "fwprovider",
		ResourceName:    name,
		TypeName:        typeName,
		Description:     op.Description,
		SDKImportPath:   sdkImportPath,
		SDKApiType:      sdkApiType,
		SDKApiAccessor:  sdkApiAccessor,
		SDKReadMethod:   sdkReadMethod,
		SDKResponseType: sdkResponseType,
	}

	// Handle list fallback (Phase 7)
	if listOp != nil {
		data.HasListFallback = true
		data.SDKListMethod = listOp.OperationID
		data.SDKListOptParams = listOp.OperationID + "OptionalParameters"
	}

	// Filter schema fields: exclude date-time fields (Phase 8: 11c)
	filteredFields := filterFields(schema.Fields)

	// Build model fields
	modelFields, nestedStructs := buildModelFields(name, filteredFields, op, listOp)
	data.Model = model.ModelData{
		StructName:    "datadog" + typeName + "DataSourceModel",
		Fields:        modelFields,
		NestedStructs: nestedStructs,
	}

	// Build schema
	data.Schema = buildSchemaData(filteredFields, op, listOp)

	// Build read data
	data.Read = buildReadData(op, isJSONAPI, listOp)

	// Build state data
	data.State = buildStateData(filteredFields, isJSONAPI)

	return data, nil
}

// filterFields removes date-time fields from schema fields (Phase 8: 11c).
func filterFields(fields []openapi.SchemaField) []openapi.SchemaField {
	var filtered []openapi.SchemaField
	for _, field := range fields {
		if field.Format == "date-time" {
			log.Printf("Skipping date-time field: %s", field.Name)
			continue
		}
		filtered = append(filtered, field)
	}
	return filtered
}

// deriveSDKImportPath extracts the API version from the path.
func deriveSDKImportPath(path string) string {
	if strings.Contains(path, "/v2/") {
		return "datadogV2"
	}
	return "datadogV1"
}

// deriveVersion extracts the version string from an API path.
func deriveVersion(path string) string {
	if strings.Contains(path, "/v2/") {
		return "V2"
	}
	return "V1"
}

// buildModelFields creates model field definitions from schema and parameters.
func buildModelFields(name string, fields []openapi.SchemaField, op *openapi.ParsedOperation, listOp *openapi.ParsedOperation) ([]model.ModelField, []model.ModelData) {
	var modelFields []model.ModelField
	var nestedStructs []model.ModelData

	// Add path parameters first (they're part of the model)
	for _, param := range op.PathParams {
		modelFields = append(modelFields, model.ModelField{
			GoName:   templates.ToPascalCase(param.Name),
			TfsdkTag: param.Name,
			GoType:   "types.String",
		})
	}

	// Add query parameters
	for _, param := range op.QueryParams {
		modelFields = append(modelFields, model.ModelField{
			GoName:   templates.ToPascalCase(param.Name),
			TfsdkTag: param.Name,
			GoType:   "types.String",
		})
	}

	// Add filter parameters from list operation (Phase 7)
	if listOp != nil {
		for _, param := range listOp.QueryParams {
			if openapi.IsFilterParam(param) {
				goName := filterParamGoName(param.Name)
				tfsdkName := templates.ToSnakeCase(goName)
				goType := filterParamModelType(param)
				modelFields = append(modelFields, model.ModelField{
					GoName:   goName,
					TfsdkTag: tfsdkName,
					GoType:   goType,
				})
			}
		}
	}

	// Add ID field for JSON:API resources (skip if already present from path params)
	hasID := false
	for _, f := range modelFields {
		if f.TfsdkTag == "id" {
			hasID = true
			break
		}
	}
	if !hasID {
		modelFields = append(modelFields, model.ModelField{
			GoName:   "ID",
			TfsdkTag: "id",
			GoType:   "types.String",
		})
	}

	// Add schema fields
	for _, field := range fields {
		goType := types.ToModelType(field)

		switch field.Type {
		case openapi.FieldTypeObject:
			goType = "*" + templates.ToPascalCase(name) + templates.ToPascalCase(field.Name) + "Model"
			nested := buildNestedModelData(name, field)
			nestedStructs = append(nestedStructs, nested)
		case openapi.FieldTypeArrayOfObjects:
			goType = "[]" + templates.ToPascalCase(name) + templates.ToPascalCase(field.Name) + "Model"
			nested := buildNestedModelData(name, field)
			nestedStructs = append(nestedStructs, nested)
		}

		modelFields = append(modelFields, model.ModelField{
			GoName:   templates.ToPascalCase(field.Name),
			TfsdkTag: field.Name,
			GoType:   goType,
		})
	}

	return modelFields, nestedStructs
}

// buildNestedModelData creates a ModelData for a nested struct.
func buildNestedModelData(parentName string, field openapi.SchemaField) model.ModelData {
	structName := templates.ToPascalCase(parentName) + templates.ToPascalCase(field.Name) + "Model"
	var fields []model.ModelField

	if field.Children != nil {
		for _, child := range field.Children.Fields {
			fields = append(fields, model.ModelField{
				GoName:   templates.ToPascalCase(child.Name),
				TfsdkTag: child.Name,
				GoType:   types.ToModelType(child),
			})
		}
	}

	return model.ModelData{
		StructName: structName,
		Fields:     fields,
	}
}

// buildSchemaData creates schema attribute/block definitions.
func buildSchemaData(fields []openapi.SchemaField, op *openapi.ParsedOperation, listOp *openapi.ParsedOperation) model.SchemaData {
	var attrs []model.AttributeData
	var blocks []model.BlockData

	hasListFallback := listOp != nil

	// Always add an "id" attribute. If a path param is named "id", it takes
	// the role (Required or Optional+Computed). Otherwise, add a Computed "id".
	hasIDParam := false
	for _, param := range op.PathParams {
		if param.Name == "id" {
			hasIDParam = true
		}
	}

	// Path parameters -> Required (or Optional+Computed with list fallback)
	for _, param := range op.PathParams {
		attr := model.AttributeData{
			Name:        param.Name,
			Type:        "schema.StringAttribute",
			Description: param.Description,
		}
		if hasListFallback {
			attr.Optional = true
			attr.Computed = true
		} else {
			attr.Required = true
		}
		attrs = append(attrs, attr)
	}

	// Add "id" as Computed if no path param provides it
	if !hasIDParam {
		attrs = append(attrs, model.AttributeData{
			Name:        "id",
			Type:        "schema.StringAttribute",
			Description: "The ID of this resource.",
			Computed:    true,
		})
	}

	// Query parameters -> Optional attributes
	for _, param := range op.QueryParams {
		attrs = append(attrs, model.AttributeData{
			Name:        param.Name,
			Type:        "schema.StringAttribute",
			Description: param.Description,
			Optional:    true,
		})
	}

	// Filter parameters from list operation -> Optional attributes (Phase 7)
	if listOp != nil {
		for _, param := range listOp.QueryParams {
			if openapi.IsFilterParam(param) {
				goName := filterParamGoName(param.Name)
				tfsdkName := templates.ToSnakeCase(goName)
				attrs = append(attrs, model.AttributeData{
					Name:        tfsdkName,
					Type:        filterParamSchemaType(param),
					Description: param.Description,
					Optional:    true,
				})
			}
		}
	}

	// Schema fields -> Computed attributes or blocks
	for _, field := range fields {
		switch field.Type {
		case openapi.FieldTypeObject:
			blocks = append(blocks, buildBlockData(field))
		case openapi.FieldTypeArrayOfObjects:
			blocks = append(blocks, buildBlockData(field))
		case openapi.FieldTypeMapOfObjects:
			attr := model.AttributeData{
				Name:        field.Name,
				Type:        "schema.MapNestedAttribute",
				Description: field.Description,
				Computed:    true,
				IsNestedMap: true,
			}
			if field.Children != nil {
				for _, child := range field.Children.Fields {
					attr.NestedAttributes = append(attr.NestedAttributes, model.AttributeData{
						Name:        child.Name,
						Type:        types.ToTerraformSchemaType(child),
						Description: child.Description,
						Computed:    true,
					})
				}
			}
			attrs = append(attrs, attr)
		default:
			attr := model.AttributeData{
				Name:        field.Name,
				Type:        types.ToTerraformSchemaType(field),
				Description: field.Description,
				Computed:    true,
			}
			if isListType(field) {
				attr.ElementType = types.ListElementType(field)
			}
			if types.IsMapType(field) {
				attr.ElementType = types.MapElementType(field)
			}
			attrs = append(attrs, attr)
		}
	}

	return model.SchemaData{
		Attributes:      attrs,
		Blocks:          blocks,
		HasListFallback: hasListFallback,
	}
}

// buildBlockData creates a block definition from a nested object or array of objects.
func buildBlockData(field openapi.SchemaField) model.BlockData {
	blockType := "SingleNestedBlock"
	if field.Type == openapi.FieldTypeArrayOfObjects {
		blockType = "ListNestedBlock"
	}

	block := model.BlockData{
		Name:        field.Name,
		Type:        blockType,
		Description: field.Description,
	}

	if field.Children != nil {
		for _, child := range field.Children.Fields {
			block.Attributes = append(block.Attributes, model.AttributeData{
				Name:        child.Name,
				Type:        types.ToTerraformSchemaType(child),
				Description: child.Description,
				Computed:    true,
			})
		}
	}

	return block
}

// isListType returns true if the field is a list type.
func isListType(field openapi.SchemaField) bool {
	switch field.Type {
	case openapi.FieldTypeArrayOfStrings, openapi.FieldTypeArrayOfInts,
		openapi.FieldTypeArrayOfFloats, openapi.FieldTypeArrayOfBools:
		return true
	}
	return false
}

// buildReadData creates read method parameter data.
func buildReadData(op *openapi.ParsedOperation, isJSONAPI bool, listOp *openapi.ParsedOperation) model.ReadData {
	var params []model.ParamData
	for _, param := range op.PathParams {
		params = append(params, model.ParamData{
			Name:        param.Name,
			GoName:      templates.ToPascalCase(param.Name),
			ValueMethod: "ValueString()",
		})
	}

	readData := model.ReadData{
		PathParams: params,
		IsJSONAPI:  isJSONAPI,
	}

	// Add filter fallback data (Phase 7)
	if listOp != nil {
		readData.HasListFallback = true
		readData.SDKListMethod = listOp.OperationID
		readData.SDKListOptParams = listOp.OperationID + "OptionalParameters"

		for _, param := range listOp.QueryParams {
			if openapi.IsFilterParam(param) {
				goName := filterParamGoName(param.Name)
				readData.FilterParams = append(readData.FilterParams, model.ParamData{
					Name:        param.Name,
					GoName:      goName,
					ValueMethod: filterParamValueMethod(param),
				})
			}
		}
	}

	return readData
}

// buildStateData creates state field mapping data.
func buildStateData(fields []openapi.SchemaField, isJSONAPI bool) model.StateData {
	state := model.StateData{}

	var prefix string
	if isJSONAPI {
		prefix = "attributes."
	} else {
		prefix = "resp."
	}

	for _, field := range fields {
		if field.Type == openapi.FieldTypeObject || field.Type == openapi.FieldTypeArrayOfObjects {
			continue // Skip complex types in state mapping for now
		}
		if field.Type == openapi.FieldTypeMapOfObjects {
			// MapOfObjects state mapping handled via MapValueFrom with types.StringType fallback
			mapping := model.FieldMapping{
				ModelField:     fmt.Sprintf("state.%s", templates.ToPascalCase(field.Name)),
				SDKAccessor:    fmt.Sprintf("%sGet%s()", prefix, templates.ToSDKPascalCase(field.Name)),
				IsMap:          true,
				MapElementType: "types.StringType",
			}
			state.FieldMappings = append(state.FieldMappings, mapping)
			continue
		}

		mapping := model.FieldMapping{
			ModelField:    fmt.Sprintf("state.%s", templates.ToPascalCase(field.Name)),
			SDKAccessor:   fmt.Sprintf("%sGet%s()", prefix, templates.ToSDKPascalCase(field.Name)),
			TypeConverter: types.TypeValueConstructor(field),
			NeedsCast:     types.NeedsCast(field),
		}

		// Handle map fields (Phase 10: additionalProperties)
		if types.IsMapType(field) {
			mapping.IsMap = true
			mapping.MapElementType = types.MapElementType(field)
			state.FieldMappings = append(state.FieldMappings, mapping)
			continue
		}

		// Handle nullable fields (Phase 8: 11b)
		if field.Nullable {
			mapping.Nullable = true
			mapping.SDKOkAccessor = fmt.Sprintf("%sGet%sOk()", prefix, templates.ToSDKPascalCase(field.Name))
			mapping.NullValue = nullValueConstructor(field)
		}

		// Handle list/array fields (Phase 8: 11d)
		if isListType(field) {
			mapping.IsList = true
			mapping.ListElementType = types.ListElementType(field)
		}

		state.FieldMappings = append(state.FieldMappings, mapping)
	}

	return state
}

// nullValueConstructor returns the types.XxxNull() constructor for a field type.
func nullValueConstructor(field openapi.SchemaField) string {
	switch field.Type {
	case openapi.FieldTypeString:
		return "types.StringNull()"
	case openapi.FieldTypeInt64, openapi.FieldTypeInt32:
		return "types.Int64Null()"
	case openapi.FieldTypeFloat64, openapi.FieldTypeFloat32:
		return "types.Float64Null()"
	case openapi.FieldTypeBool:
		return "types.BoolNull()"
	default:
		return "types.StringNull()"
	}
}

// filterParamGoName converts a bracket-notation filter param to a Go name.
// e.g., "filter[keyword]" -> "FilterKeyword"
func filterParamGoName(name string) string {
	// Strip "filter[" prefix and "]" suffix
	inner := name
	if strings.HasPrefix(name, "filter[") && strings.HasSuffix(name, "]") {
		inner = name[len("filter[") : len(name)-1]
	}
	return "Filter" + templates.ToPascalCase(inner)
}

// filterParamModelType returns the model type for a filter parameter.
func filterParamModelType(param openapi.Parameter) string {
	if param.Schema != nil {
		schema, err := param.Schema.BuildSchema()
		if err == nil && len(schema.Type) > 0 {
			switch schema.Type[0] {
			case "boolean":
				return "types.Bool"
			case "integer":
				return "types.Int64"
			}
		}
	}
	return "types.String"
}

// filterParamSchemaType returns the schema attribute type for a filter parameter.
func filterParamSchemaType(param openapi.Parameter) string {
	if param.Schema != nil {
		schema, err := param.Schema.BuildSchema()
		if err == nil && len(schema.Type) > 0 {
			switch schema.Type[0] {
			case "boolean":
				return "schema.BoolAttribute"
			case "integer":
				return "schema.Int64Attribute"
			}
		}
	}
	return "schema.StringAttribute"
}

// filterParamValueMethod returns the value extraction method for a filter parameter.
func filterParamValueMethod(param openapi.Parameter) string {
	if param.Schema != nil {
		schema, err := param.Schema.BuildSchema()
		if err == nil && len(schema.Type) > 0 {
			switch schema.Type[0] {
			case "boolean":
				return "ValueBool()"
			case "integer":
				return "ValueInt64()"
			}
		}
	}
	return "ValueString()"
}
