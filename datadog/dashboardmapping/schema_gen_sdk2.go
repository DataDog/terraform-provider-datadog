package dashboardmapping

// schema_gen_sdk2.go
//
// SDKv2 schema generation functions that convert FieldSpec declarations into
// terraform-plugin-sdk/v2 schema.Schema values suitable for Terraform provider registration.
//
// This is the SDKv2 parallel of schema_gen.go (framework schema generation).
// All FieldSpec/WidgetSpec declarations are shared with the framework version.
//
// Primary entry points:
//   - FieldSpecsToSDKv2Schema: convert []FieldSpec → map[string]*schema.Schema
//   - FieldSpecToSDKv2: convert one FieldSpec → *schema.Schema
//   - AllWidgetSDKv2Schema: all widget definition blocks as map[string]*schema.Schema

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// FieldSpecsToSDKv2Schema converts a []FieldSpec into a map[string]*schema.Schema.
// All field types (including TypeBlock, TypeBlockList, TypeOneOf) are handled.
func FieldSpecsToSDKv2Schema(fields []FieldSpec) map[string]*schema.Schema {
	result := make(map[string]*schema.Schema, len(fields))
	for _, f := range fields {
		result[f.HCLKey] = FieldSpecToSDKv2(f)
	}
	return result
}

// FieldSpecToSDKv2 converts a single FieldSpec to a *schema.Schema.
func FieldSpecToSDKv2(f FieldSpec) *schema.Schema {
	desc := enrichDesc(f.Description, f.ValidValues)

	s := &schema.Schema{
		Description: desc,
		Computed:    f.Computed,
		Sensitive:   f.Sensitive,
		ForceNew:    f.ForceNew,
	}

	if f.Required {
		s.Required = true
	} else {
		s.Optional = true
	}

	if f.Deprecated != "" {
		s.Deprecated = f.Deprecated
	}

	if f.Default != nil {
		switch v := f.Default.(type) {
		case string:
			if v != "" {
				s.Default = v
			}
		default:
			s.Default = f.Default
		}
	}

	// ConflictsWith is omitted: SDKv2 InternalValidate rejects paths that
	// reference sibling attributes inside nested blocks, which breaks for
	// all widget definition fields. Mutual-exclusivity is enforced at the
	// API level and is validated by the framework version of this resource.

	switch f.Type {
	case TypeString:
		s.Type = schema.TypeString
		if len(f.ValidValues) > 0 {
			s.ValidateDiagFunc = validation.ToDiagFunc(
				validation.StringInSlice(f.ValidValues, false),
			)
		}

	case TypeBool:
		s.Type = schema.TypeBool

	case TypeInt:
		s.Type = schema.TypeInt

	case TypeFloat:
		s.Type = schema.TypeFloat

	case TypeStringList:
		if f.UseSet {
			s.Type = schema.TypeSet
		} else {
			s.Type = schema.TypeList
		}
		s.Elem = &schema.Schema{Type: schema.TypeString}
		if f.MaxItems > 0 {
			s.MaxItems = f.MaxItems
		}

	case TypeIntList:
		if f.UseSet {
			s.Type = schema.TypeSet
		} else {
			s.Type = schema.TypeList
		}
		s.Elem = &schema.Schema{Type: schema.TypeInt}
		if f.MaxItems > 0 {
			s.MaxItems = f.MaxItems
		}

	case TypeBlock:
		s.Type = schema.TypeList
		s.MaxItems = 1
		s.Elem = &schema.Resource{
			Schema: FieldSpecsToSDKv2Schema(f.Children),
		}

	case TypeBlockList:
		s.Type = schema.TypeList
		s.Elem = &schema.Resource{
			Schema: FieldSpecsToSDKv2Schema(f.Children),
		}
		if f.MaxItems > 0 {
			s.MaxItems = f.MaxItems
		}

	case TypeOneOf:
		// Flatten all children variants into a single TypeList, MaxItems:1 block.
		// Each child variant is a TypeBlock, so we merge their children schemas.
		merged := make(map[string]*schema.Schema)
		for _, child := range f.Children {
			for key, childSchema := range FieldSpecsToSDKv2Schema([]FieldSpec{child}) {
				// In a TypeOneOf, all variant blocks are optional
				childSchema.Required = false
				childSchema.Optional = true
				merged[key] = childSchema
			}
		}
		s.Type = schema.TypeList
		s.MaxItems = 1
		s.Elem = &schema.Resource{
			Schema: merged,
		}
	}

	return s
}

// WidgetSpecToSDKv2Schema generates the schema map for a widget definition block.
// Merges CommonWidgetFields with the widget's own fields.
func WidgetSpecToSDKv2Schema(ws WidgetSpec) *schema.Schema {
	allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(ws.Fields))
	allFields = append(allFields, CommonWidgetFields...)
	allFields = append(allFields, ws.Fields...)
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: ws.Description,
		Elem: &schema.Resource{
			Schema: FieldSpecsToSDKv2Schema(allFields),
		},
	}
}

// splitGraphSourceWidgetSDKv2Schemas returns the schema map for source_widget_definition
// inside split_graph_definition. Only the widget types supported as split graph sources are included.
func splitGraphSourceWidgetSDKv2Schemas() map[string]*schema.Schema {
	result := make(map[string]*schema.Schema)
	for _, spec := range allWidgetSpecs {
		if !splitGraphSourceWidgetTypes[spec.JSONType] {
			continue
		}
		result[spec.HCLKey] = WidgetSpecToSDKv2Schema(spec)
	}
	return result
}

// AllWidgetSDKv2SchemaNoGroup returns the widget schema map for all widget types EXCEPT group.
// Used for widgets nested inside a group_definition block.
func AllWidgetSDKv2SchemaNoGroup() map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "The ID of the widget.",
		},
		"widget_layout": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The layout of the widget on a 'free' dashboard.",
			Elem: &schema.Resource{
				Schema: FieldSpecsToSDKv2Schema(widgetLayoutFieldSpecs),
			},
		},
	}
	for _, spec := range allWidgetSpecs {
		if spec.JSONType == "group" || spec.JSONType == "powerpack" || spec.JSONType == "split_group" {
			continue
		}
		result[spec.HCLKey] = WidgetSpecToSDKv2Schema(spec)
	}
	return result
}

// AllWidgetSDKv2Schema returns the complete widget schema map for all widget definition types,
// including widget_layout and id wrapper fields.
// If excludePowerpackOnly is true, powerpack and split_graph definitions are excluded.
func AllWidgetSDKv2Schema(excludePowerpackOnly bool) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "The ID of the widget.",
		},
		"widget_layout": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The layout of the widget on a 'free' dashboard.",
			Elem: &schema.Resource{
				Schema: FieldSpecsToSDKv2Schema(widgetLayoutFieldSpecs),
			},
		},
	}
	for _, spec := range allWidgetSpecs {
		if excludePowerpackOnly && (spec.JSONType == "powerpack" || spec.JSONType == "split_group") {
			continue
		}
		widgetSchema := WidgetSpecToSDKv2Schema(spec)
		widgetResource := widgetSchema.Elem.(*schema.Resource)

		// Inject source_widget_definition into split_graph_definition.
		if spec.JSONType == "split_group" {
			widgetResource.Schema["source_widget_definition"] = &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The original widget we are splitting on.",
				Elem: &schema.Resource{
					Schema: splitGraphSourceWidgetSDKv2Schemas(),
				},
			}
		}

		// Inject nested widget list into group_definition.
		if spec.JSONType == "group" {
			widgetResource.Schema["widget"] = &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of widgets in this group.",
				Elem: &schema.Resource{
					Schema: AllWidgetSDKv2SchemaNoGroup(),
				},
			}
		}

		result[spec.HCLKey] = widgetSchema
	}
	return result
}
