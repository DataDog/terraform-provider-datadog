package dashboardmapping

// schema_gen.go
//
// Schema generation functions that convert FieldSpec declarations into
// *schema.Schema values suitable for Terraform provider registration.
//
// The primary entry point is WidgetSpecToSchemaBlock, which produces the
// TypeList/MaxItems:1 block schema for a widget definition.

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ValidateStringIsNotEmpty is a SchemaValidateDiagFunc that rejects empty strings.
// Use this as ValidateDiag in FieldSpec when the field must be non-empty.
var ValidateStringIsNotEmpty = validation.ToDiagFunc(validation.StringIsNotEmpty)

// FieldSpecToSchemaElem converts a single FieldSpec to *schema.Schema.
func FieldSpecToSchemaElem(f FieldSpec) *schema.Schema {
	s := &schema.Schema{
		Description: f.Description,
		Deprecated:  f.Deprecated,
		Sensitive:   f.Sensitive,
	}

	// Optionality
	switch {
	case f.Required:
		s.Required = true
	case f.Computed && !f.Required:
		s.Optional = true
		s.Computed = true
	default:
		s.Optional = true
	}

	if f.Default != nil {
		s.Default = f.Default
	}
	if len(f.ConflictsWith) > 0 {
		s.ConflictsWith = f.ConflictsWith
	}
	// Type and Elem
	listType := schema.TypeList
	if f.UseSet {
		listType = schema.TypeSet
	}

	switch f.Type {
	case TypeString:
		s.Type = schema.TypeString
		if len(f.ValidValues) > 0 {
			s.ValidateDiagFunc = validation.ToDiagFunc(validation.StringInSlice(f.ValidValues, false))
		} else if f.ValidateDiag != nil {
			s.ValidateDiagFunc = f.ValidateDiag
		}
	case TypeBool:
		s.Type = schema.TypeBool
	case TypeInt:
		s.Type = schema.TypeInt
	case TypeFloat:
		s.Type = schema.TypeFloat
	case TypeStringList:
		s.Type = listType
		elem := &schema.Schema{Type: schema.TypeString}
		if len(f.ValidValues) > 0 {
			elem.ValidateDiagFunc = validation.ToDiagFunc(validation.StringInSlice(f.ValidValues, false))
		} else if f.ValidateDiag != nil {
			elem.ValidateDiagFunc = f.ValidateDiag
		}
		s.Elem = elem
		if f.MaxItems > 0 {
			s.MaxItems = f.MaxItems
		}
	case TypeIntList:
		s.Type = listType
		s.Elem = &schema.Schema{Type: schema.TypeInt}
	case TypeBlock:
		s.Type = schema.TypeList
		s.MaxItems = 1
		s.Elem = &schema.Resource{Schema: FieldSpecsToSchema(f.Children)}
	case TypeBlockList:
		s.Type = listType
		if f.MaxItems > 0 {
			s.MaxItems = f.MaxItems
		}
		s.Elem = &schema.Resource{Schema: FieldSpecsToSchema(f.Children)}
	}
	return s
}

// FieldSpecsToSchema converts []FieldSpec to map[string]*schema.Schema.
func FieldSpecsToSchema(fields []FieldSpec) map[string]*schema.Schema {
	result := make(map[string]*schema.Schema, len(fields))
	for _, f := range fields {
		result[f.HCLKey] = FieldSpecToSchemaElem(f)
	}
	return result
}

// WidgetSpecToSchemaBlock generates the *schema.Schema for a widget definition block.
// Merges CommonWidgetFields with the widget's own fields.
// The outer block is Optional, TypeList, MaxItems: 1.
func WidgetSpecToSchemaBlock(ws WidgetSpec) *schema.Schema {
	allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(ws.Fields))
	allFields = append(allFields, CommonWidgetFields...)
	allFields = append(allFields, ws.Fields...)
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: ws.Description,
		Elem:        &schema.Resource{Schema: FieldSpecsToSchema(allFields)},
	}
}
