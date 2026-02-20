package dashboardmapping

// schema_gen.go
//
// Framework schema generation functions that convert FieldSpec declarations into
// terraform-plugin-framework schema.Attribute and schema.Block values suitable
// for Terraform provider registration.
//
// Primary entry points:
//   - FieldSpecsToFWSchema: convert []FieldSpec → (attrs, blocks)
//   - WidgetSpecToFWBlock: convert WidgetSpec → schema.ListNestedBlock
//   - AllWidgetFWBlocks: all widget definition blocks as (attrs, blocks)
//   - FieldSpecsToAttrTypes: compute attr.Type map (for state conversion)

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// enrichDesc appends "Valid values are `a`, `b`." to a description when ValidValues are set.
func enrichDesc(desc string, validValues []string) string {
	if len(validValues) == 0 || strings.Contains(desc, "Valid values") {
		return desc
	}
	quoted := make([]string, len(validValues))
	for i, v := range validValues {
		quoted[i] = fmt.Sprintf("`%s`", v)
	}
	return desc + " Valid values are " + strings.Join(quoted, ", ") + "."
}

// FieldSpecToFWAttribute converts a non-block FieldSpec into a framework schema.Attribute.
// Only valid for TypeString, TypeBool, TypeInt, TypeFloat, TypeStringList, TypeIntList.
// TypeBlock and TypeBlockList are handled as blocks (see FieldSpecsToFWSchema).
func FieldSpecToFWAttribute(f FieldSpec) schema.Attribute {
	desc := enrichDesc(f.Description, f.ValidValues)
	isRequired := f.Required
	isComputed := f.Computed
	isOptional := !f.Required

	switch f.Type {
	case TypeString:
		// Optional string fields are always Optional+Computed+UseStateForUnknown to avoid
		// "inconsistent plan" errors when the API omits empty optional strings.
		// This matches SDKv2 behavior where TypeString defaults to "".
		isEffectivelyComputed := isComputed || (!isRequired && !isComputed)
		a := schema.StringAttribute{
			Description:        desc,
			Required:           isRequired,
			Optional:           isOptional,
			Computed:           isEffectivelyComputed,
			Sensitive:          f.Sensitive,
			DeprecationMessage: f.Deprecated,
		}
		if len(f.ValidValues) > 0 {
			a.Validators = []validator.String{
				stringvalidator.OneOf(f.ValidValues...),
			}
		}
		if f.Default != nil {
			if sv, ok := f.Default.(string); ok {
				a.Default = stringdefault.StaticString(sv)
				a.Computed = true
			}
		}
		if f.ForceNew {
			a.PlanModifiers = append(a.PlanModifiers, stringplanmodifier.RequiresReplace())
		}
		// UseStateForUnknown prevents "(known after apply)" for computed optional strings
		if !isRequired && f.Default == nil {
			a.PlanModifiers = append(a.PlanModifiers, stringplanmodifier.UseStateForUnknown())
		}
		return a

	case TypeBool:
		// Like TypeString, Optional bool attributes are Optional+Computed to match SDKv2 defaults.
		isEffectivelyComputed := isComputed || (!isRequired && !isComputed)
		a := schema.BoolAttribute{
			Description:        desc,
			Required:           isRequired,
			Optional:           isOptional,
			Computed:           isEffectivelyComputed,
			Sensitive:          f.Sensitive,
			DeprecationMessage: f.Deprecated,
		}
		if f.Default != nil {
			if bv, ok := f.Default.(bool); ok {
				a.Default = booldefault.StaticBool(bv)
				a.Computed = true
			}
		}
		if f.ForceNew {
			a.PlanModifiers = append(a.PlanModifiers, boolplanmodifier.RequiresReplace())
		}
		if !isRequired && f.Default == nil {
			a.PlanModifiers = append(a.PlanModifiers, boolplanmodifier.UseStateForUnknown())
		}
		return a

	case TypeInt:
		isEffectivelyComputed := isComputed || (!isRequired && !isComputed)
		a := schema.Int64Attribute{
			Description:        desc,
			Required:           isRequired,
			Optional:           isOptional,
			Computed:           isEffectivelyComputed,
			Sensitive:          f.Sensitive,
			DeprecationMessage: f.Deprecated,
		}
		if f.ForceNew {
			a.PlanModifiers = append(a.PlanModifiers, int64planmodifier.RequiresReplace())
		}
		if !isRequired {
			a.PlanModifiers = append(a.PlanModifiers, int64planmodifier.UseStateForUnknown())
		}
		return a

	case TypeFloat:
		isEffectivelyComputed := isComputed || (!isRequired && !isComputed)
		a := schema.Float64Attribute{
			Description:        desc,
			Required:           isRequired,
			Optional:           isOptional,
			Computed:           isEffectivelyComputed,
			Sensitive:          f.Sensitive,
			DeprecationMessage: f.Deprecated,
		}
		if f.ForceNew {
			a.PlanModifiers = append(a.PlanModifiers, float64planmodifier.RequiresReplace())
		}
		if !isRequired {
			a.PlanModifiers = append(a.PlanModifiers, float64planmodifier.UseStateForUnknown())
		}
		return a

	case TypeStringList:
		// Optional list fields are also Computed to allow provider to set empty list
		// when user hasn't configured the field (API always returns []).
		isEffectivelyComputed := isComputed || (!isRequired && !isComputed)
		a := schema.ListAttribute{
			Description:        desc,
			Required:           isRequired,
			Optional:           isOptional,
			Computed:           isEffectivelyComputed,
			Sensitive:          f.Sensitive,
			DeprecationMessage: f.Deprecated,
			ElementType:        types.StringType,
		}
		var listValidators []validator.List
		if f.MaxItems > 0 {
			listValidators = append(listValidators, listvalidator.SizeAtMost(f.MaxItems))
		}
		if len(f.ValidValues) > 0 {
			listValidators = append(listValidators, listvalidator.ValueStringsAre(
				stringvalidator.OneOf(f.ValidValues...),
			))
		}
		if len(listValidators) > 0 {
			a.Validators = listValidators
		}
		if f.ForceNew {
			a.PlanModifiers = append(a.PlanModifiers, listplanmodifier.RequiresReplace())
		}
		if !isRequired {
			a.PlanModifiers = append(a.PlanModifiers, listplanmodifier.UseStateForUnknown())
		}
		return a

	case TypeIntList:
		// TypeIntList: only Computed when FieldSpec explicitly says so (e.g. dashboard_lists_removed).
		// dashboard_lists is SchemaOnly/Optional-only; making it Computed would cause unknown issues.
		if f.UseSet {
			a := schema.SetAttribute{
				Description:        desc,
				Required:           isRequired,
				Optional:           isOptional && !isComputed,
				Computed:           isComputed,
				Sensitive:          f.Sensitive,
				DeprecationMessage: f.Deprecated,
				ElementType:        types.Int64Type,
			}
			if isOptional && isComputed {
				a.Optional = true
			}
			if isComputed {
				a.PlanModifiers = append(a.PlanModifiers, setplanmodifier.UseStateForUnknown())
			}
			return a
		}
		a := schema.ListAttribute{
			Description:        desc,
			Required:           isRequired,
			Optional:           isOptional && !isComputed,
			Computed:           isComputed,
			Sensitive:          f.Sensitive,
			DeprecationMessage: f.Deprecated,
			ElementType:        types.Int64Type,
		}
		if isOptional && isComputed {
			a.Optional = true
		}
		return a
	}

	// Default: return empty StringAttribute
	return schema.StringAttribute{Description: desc}
}

// FieldSpecsToFWSchema converts a []FieldSpec into separate framework (attrs, blocks) maps.
// TypeBlock, TypeBlockList, and TypeOneOf become blocks; all others become attributes.
func FieldSpecsToFWSchema(fields []FieldSpec) (map[string]schema.Attribute, map[string]schema.Block) {
	attrs := make(map[string]schema.Attribute)
	blocks := make(map[string]schema.Block)
	for _, f := range fields {
		switch f.Type {
		case TypeBlock:
			blocks[f.HCLKey] = fieldSpecToFWBlock(f)
		case TypeBlockList:
			blocks[f.HCLKey] = fieldSpecToFWBlockList(f)
		case TypeOneOf:
			blocks[f.HCLKey] = fieldSpecToFWOneOf(f)
		default:
			attrs[f.HCLKey] = FieldSpecToFWAttribute(f)
		}
	}
	return attrs, blocks
}

// fieldSpecToFWOneOf converts a TypeOneOf FieldSpec into a ListNestedBlock with SizeAtMost(1).
// The nested block contains all variant children as optional attributes/blocks.
// Exactly one child variant should be set; add ConflictsWith to the FieldSpec declarations
// to enforce mutual exclusivity at the schema level.
func fieldSpecToFWOneOf(f FieldSpec) schema.ListNestedBlock {
	childAttrs, childBlocks := FieldSpecsToFWSchema(f.Children)
	desc := enrichDesc(f.Description, f.ValidValues)
	validators := []validator.List{
		listvalidator.SizeAtMost(1),
	}
	if f.Required {
		validators = append(validators, listvalidator.IsRequired())
	}
	return schema.ListNestedBlock{
		Description: desc,
		Validators:  validators,
		NestedObject: schema.NestedBlockObject{
			Attributes: childAttrs,
			Blocks:     childBlocks,
		},
	}
}

// fieldSpecToFWBlock converts a TypeBlock FieldSpec into a ListNestedBlock with SizeAtMost(1).
func fieldSpecToFWBlock(f FieldSpec) schema.ListNestedBlock {
	childAttrs, childBlocks := FieldSpecsToFWSchema(f.Children)
	desc := enrichDesc(f.Description, f.ValidValues)
	validators := []validator.List{
		listvalidator.SizeAtMost(1),
	}
	if f.Required {
		validators = append(validators, listvalidator.IsRequired())
	}
	return schema.ListNestedBlock{
		Description: desc,
		Validators:  validators,
		NestedObject: schema.NestedBlockObject{
			Attributes: childAttrs,
			Blocks:     childBlocks,
		},
	}
}

// fieldSpecToFWBlockList converts a TypeBlockList FieldSpec into a ListNestedBlock.
func fieldSpecToFWBlockList(f FieldSpec) schema.ListNestedBlock {
	childAttrs, childBlocks := FieldSpecsToFWSchema(f.Children)
	desc := enrichDesc(f.Description, f.ValidValues)
	var validators []validator.List
	if f.Required {
		validators = append(validators, listvalidator.IsRequired())
	}
	if f.MaxItems > 0 {
		validators = append(validators, listvalidator.SizeAtMost(f.MaxItems))
	}
	return schema.ListNestedBlock{
		Description: desc,
		Validators:  validators,
		NestedObject: schema.NestedBlockObject{
			Attributes: childAttrs,
			Blocks:     childBlocks,
		},
	}
}

// WidgetSpecToFWBlock generates the schema.ListNestedBlock for a widget definition block.
// Merges CommonWidgetFields with the widget's own fields.
// The outer block is Optional, TypeList, MaxItems: 1.
func WidgetSpecToFWBlock(ws WidgetSpec) schema.ListNestedBlock {
	allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(ws.Fields))
	allFields = append(allFields, CommonWidgetFields...)
	allFields = append(allFields, ws.Fields...)
	childAttrs, childBlocks := FieldSpecsToFWSchema(allFields)
	return schema.ListNestedBlock{
		Description: ws.Description,
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: childAttrs,
			Blocks:     childBlocks,
		},
	}
}

// splitGraphSourceWidgetFWBlocks returns the blocks map for source_widget_definition
// inside split_graph_definition. Only the 9 widget types supported by the API
// as split graph sources are included.
func splitGraphSourceWidgetFWBlocks() map[string]schema.Block {
	blocks := make(map[string]schema.Block)
	for _, spec := range allWidgetSpecs {
		if !splitGraphSourceWidgetTypes[spec.JSONType] {
			continue
		}
		blocks[spec.HCLKey] = WidgetSpecToFWBlock(spec)
	}
	return blocks
}

// AllWidgetFWBlocksNoGroup returns widget (attrs, blocks) for all widget types EXCEPT group.
// Used for widgets nested inside a group_definition block (no recursive groups).
func AllWidgetFWBlocksNoGroup() (map[string]schema.Attribute, map[string]schema.Block) {
	attrs := map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the widget.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
	}
	blocks := map[string]schema.Block{
		"widget_layout": schema.ListNestedBlock{
			Description: "The layout of the widget on a 'free' dashboard.",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: func() map[string]schema.Attribute {
					a, _ := FieldSpecsToFWSchema(widgetLayoutFieldSpecs)
					return a
				}(),
			},
		},
	}
	for _, spec := range allWidgetSpecs {
		if spec.JSONType == "group" || spec.JSONType == "powerpack" || spec.JSONType == "split_group" {
			continue
		}
		blocks[spec.HCLKey] = WidgetSpecToFWBlock(spec)
	}
	return attrs, blocks
}

// AllWidgetFWBlocks returns the (attrs, blocks) maps for all widget definition types,
// including widget_layout and id wrapper fields. If excludePowerpackOnly is true,
// powerpack and split_graph definitions are excluded (for use by the powerpack resource).
func AllWidgetFWBlocks(excludePowerpackOnly bool) (map[string]schema.Attribute, map[string]schema.Block) {
	attrs := map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the widget.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
	}
	blocks := map[string]schema.Block{
		"widget_layout": schema.ListNestedBlock{
			Description: "The layout of the widget on a 'free' dashboard.",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: func() map[string]schema.Attribute {
					a, _ := FieldSpecsToFWSchema(widgetLayoutFieldSpecs)
					return a
				}(),
			},
		},
	}
	for _, spec := range allWidgetSpecs {
		if excludePowerpackOnly && (spec.JSONType == "powerpack" || spec.JSONType == "split_group") {
			continue
		}
		block := WidgetSpecToFWBlock(spec)
		if block.NestedObject.Blocks == nil {
			block.NestedObject.Blocks = make(map[string]schema.Block)
		}
		// Inject source_widget_definition into split_graph_definition.
		if spec.JSONType == "split_group" {
			srcWidgetBlocks := splitGraphSourceWidgetFWBlocks()
			block.NestedObject.Blocks["source_widget_definition"] = schema.ListNestedBlock{
				Description: "The original widget we are splitting on.",
				Validators:  []validator.List{listvalidator.SizeAtMost(1), listvalidator.IsRequired()},
				NestedObject: schema.NestedBlockObject{
					Blocks: srcWidgetBlocks,
				},
			}
		}
		// Inject nested widget list into group_definition.
		if spec.JSONType == "group" {
			innerWidgetAttrs, innerWidgetBlocks := AllWidgetFWBlocksNoGroup()
			block.NestedObject.Blocks["widget"] = schema.ListNestedBlock{
				Description: "The list of widgets in this group.",
				NestedObject: schema.NestedBlockObject{
					Attributes: innerWidgetAttrs,
					Blocks:     innerWidgetBlocks,
				},
			}
		}
		blocks[spec.HCLKey] = block
	}
	return attrs, blocks
}

// ============================================================
// AttrType computation for state conversion
// ============================================================

// FieldSpecsToAttrTypes returns the attr.Type map for a given []FieldSpec.
// Used to create types.Object values when setting framework state.
func FieldSpecsToAttrTypes(fields []FieldSpec) map[string]attr.Type {
	result := make(map[string]attr.Type, len(fields))
	for _, f := range fields {
		result[f.HCLKey] = FieldSpecToAttrType(f)
	}
	return result
}

// FieldSpecToAttrType returns the attr.Type for a single FieldSpec.
func FieldSpecToAttrType(f FieldSpec) attr.Type {
	switch f.Type {
	case TypeString:
		return types.StringType
	case TypeBool:
		return types.BoolType
	case TypeInt:
		return types.Int64Type
	case TypeFloat:
		return types.Float64Type
	case TypeStringList:
		// UseSet uses ListType (not SetType) to preserve insertion order - see FieldSpecToFWAttribute.
		return types.ListType{ElemType: types.StringType}
	case TypeIntList:
		if f.UseSet {
			return types.SetType{ElemType: types.Int64Type}
		}
		return types.ListType{ElemType: types.Int64Type}
	case TypeBlock, TypeBlockList, TypeOneOf:
		childTypes := FieldSpecsToAttrTypes(f.Children)
		objType := types.ObjectType{AttrTypes: childTypes}
		return types.ListType{ElemType: objType}
	}
	return types.StringType
}

// WidgetAttrTypes returns the attr.Type map for a widget definition object.
// Includes all common widget fields + the widget spec's specific fields.
func WidgetAttrTypes(ws WidgetSpec) map[string]attr.Type {
	allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(ws.Fields))
	allFields = append(allFields, CommonWidgetFields...)
	allFields = append(allFields, ws.Fields...)
	return FieldSpecsToAttrTypes(allFields)
}

// AllWidgetAttrTypes returns the complete attr.Type map for a widget container object
// (the object that can hold any of the defined widget definition blocks).
func AllWidgetAttrTypes(excludePowerpackOnly bool) map[string]attr.Type {
	result := map[string]attr.Type{
		"id":            types.Int64Type,
		"widget_layout": types.ListType{ElemType: types.ObjectType{AttrTypes: FieldSpecsToAttrTypes(widgetLayoutFieldSpecs)}},
	}
	for _, spec := range allWidgetSpecs {
		if excludePowerpackOnly && (spec.JSONType == "powerpack" || spec.JSONType == "split_group") {
			continue
		}
		defAttrTypes := WidgetAttrTypes(spec)
		// Add source_widget_definition for split_graph
		if spec.JSONType == "split_group" {
			srcWidgetAttrTypes := map[string]attr.Type{}
			for _, s := range allWidgetSpecs {
				if !splitGraphSourceWidgetTypes[s.JSONType] {
					continue
				}
				srcWidgetAttrTypes[s.HCLKey] = types.ListType{ElemType: types.ObjectType{AttrTypes: WidgetAttrTypes(s)}}
			}
			defAttrTypes["source_widget_definition"] = types.ListType{
				ElemType: types.ObjectType{AttrTypes: srcWidgetAttrTypes},
			}
		}
		// Add nested widget list for group widget
		if spec.JSONType == "group" {
			innerWidgetTypes := AllWidgetAttrTypesNoGroup()
			defAttrTypes["widget"] = types.ListType{
				ElemType: types.ObjectType{AttrTypes: innerWidgetTypes},
			}
		}
		result[spec.HCLKey] = types.ListType{ElemType: types.ObjectType{AttrTypes: defAttrTypes}}
	}
	return result
}

// AllWidgetAttrTypesNoGroup is like AllWidgetAttrTypes but excludes GroupWidget
// (used for the inner widget list inside group_definition).
func AllWidgetAttrTypesNoGroup() map[string]attr.Type {
	result := map[string]attr.Type{
		"id":            types.Int64Type,
		"widget_layout": types.ListType{ElemType: types.ObjectType{AttrTypes: FieldSpecsToAttrTypes(widgetLayoutFieldSpecs)}},
	}
	for _, spec := range allWidgetSpecs {
		if spec.JSONType == "group" || spec.JSONType == "powerpack" || spec.JSONType == "split_group" {
			continue
		}
		defAttrTypes := WidgetAttrTypes(spec)
		result[spec.HCLKey] = types.ListType{ElemType: types.ObjectType{AttrTypes: defAttrTypes}}
	}
	return result
}
