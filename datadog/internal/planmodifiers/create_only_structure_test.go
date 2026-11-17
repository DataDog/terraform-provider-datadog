package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// Exercise Framework's unknown marking and parent-before-child modifier order,
// rather than passing hand-normalized values directly to the modifiers.
func TestCreateOnlyStructurePlan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for _, test := range []struct {
		name, priorScope, scope   string
		priorSelectors, selectors []string
		wantReplace               bool
	}{
		{name: "mutable sibling", priorScope: "team", scope: "team", priorSelectors: []string{"region"}, selectors: []string{"region"}},
		{name: "object input changes", priorScope: "team", scope: "other", priorSelectors: []string{"region"}, selectors: []string{"region"}, wantReplace: true},
		{name: "object added", scope: "team", priorSelectors: []string{"region"}, selectors: []string{"region"}, wantReplace: true},
		{name: "object input unknown", priorScope: "team", scope: "<unknown>", priorSelectors: []string{"region"}, selectors: []string{"region"}, wantReplace: true},
		{name: "list input changes", priorScope: "team", scope: "team", priorSelectors: []string{"region"}, selectors: []string{"service"}, wantReplace: true},
		{name: "list element unknown", priorScope: "team", scope: "team", priorSelectors: []string{"region"}, selectors: []string{"<unknown>"}, wantReplace: true},
		{name: "list element null", priorScope: "team", scope: "team", priorSelectors: []string{"region"}, selectors: []string{""}, wantReplace: true},
		{name: "list grows", priorScope: "team", scope: "team", priorSelectors: []string{"region"}, selectors: []string{"region", "service"}, wantReplace: true},
		{name: "list shrinks", priorScope: "team", scope: "team", priorSelectors: []string{"region", "service"}, selectors: []string{"region"}, wantReplace: true},
		{name: "list emptied", priorScope: "team", scope: "team", priorSelectors: []string{"region"}, selectors: []string{}, wantReplace: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := providerserver.NewProtocol6(&createOnlyStructureProvider{})()
			prior := createOnlyStructureState(test.priorScope, test.priorSelectors, 1, false)
			config := createOnlyStructureState(test.scope, test.selectors, 2, true)
			proposed := createOnlyStructureState(test.scope, test.selectors, 2, false)
			response, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName: "structure_test", PriorState: createOnlyStructureDynamicValue(t, prior),
				Config: createOnlyStructureDynamicValue(t, config), ProposedNewState: createOnlyStructureDynamicValue(t, proposed),
			})
			require.NoError(t, err)
			for _, diagnostic := range response.Diagnostics {
				require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s: %s", diagnostic.Summary, diagnostic.Detail)
			}
			if test.wantReplace {
				require.NotEmpty(t, response.RequiresReplace)
			} else {
				require.Empty(t, response.RequiresReplace)
				// Computed output values may change when capacity updates. The
				// modifier must ignore them, without freezing them to prior state.
				planned, err := response.PlannedState.Unmarshal(prior.Type())
				require.NoError(t, err)
				output, _, err := tftypes.WalkAttributePath(planned, tftypes.NewAttributePath().WithAttributeName("scope").WithAttributeName("output"))
				require.NoError(t, err)
				require.False(t, output.(tftypes.Value).IsKnown())
			}
		})
	}

	t.Run("whole-value comparison reproduces replacement", func(t *testing.T) {
		server := providerserver.NewProtocol6(&createOnlyStructureProvider{wholeValue: true})()
		prior := createOnlyStructureState("team", []string{"region"}, 1, false)
		config := createOnlyStructureState("team", []string{"region"}, 2, true)
		proposed := createOnlyStructureState("team", []string{"region"}, 2, false)
		response, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
			TypeName: "structure_test", PriorState: createOnlyStructureDynamicValue(t, prior),
			Config: createOnlyStructureDynamicValue(t, config), ProposedNewState: createOnlyStructureDynamicValue(t, proposed),
		})
		require.NoError(t, err)
		for _, diagnostic := range response.Diagnostics {
			require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s: %s", diagnostic.Summary, diagnostic.Detail)
		}
		require.Len(t, response.RequiresReplace, 2, "both whole-value comparisons mistake computed unknowns for input changes")
	})
}

func createOnlyStructureState(scope string, selectors []string, capacity int64, config bool) tftypes.Value {
	stringValue := func(value string) tftypes.Value { return tftypes.NewValue(tftypes.String, value) }
	elementType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String, "output": tftypes.String}}
	object := func(name string) tftypes.Value {
		if name == "<unknown>" {
			return tftypes.NewValue(elementType, tftypes.UnknownValue)
		}
		if name == "" {
			return tftypes.NewValue(elementType, nil)
		}
		output := stringValue("server-output")
		if config {
			output = tftypes.NewValue(tftypes.String, nil)
		}
		return tftypes.NewValue(elementType, map[string]tftypes.Value{"name": stringValue(name), "output": output})
	}
	elements := make([]tftypes.Value, len(selectors))
	for index, selector := range selectors {
		elements[index] = object(selector)
	}
	listType := tftypes.List{ElementType: elementType}
	rootType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"scope": elementType, "selectors": listType, "capacity": tftypes.Number}}
	return tftypes.NewValue(rootType, map[string]tftypes.Value{
		"scope": object(scope), "selectors": tftypes.NewValue(listType, elements), "capacity": tftypes.NewValue(tftypes.Number, capacity),
	})
}

func createOnlyStructureDynamicValue(t *testing.T, value tftypes.Value) *tfprotov6.DynamicValue {
	t.Helper()
	dynamic, err := tfprotov6.NewDynamicValue(value.Type(), value)
	require.NoError(t, err)
	return &dynamic
}

type createOnlyStructureProvider struct{ wholeValue bool }

func (p *createOnlyStructureProvider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "structure"
}
func (p *createOnlyStructureProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {
}
func (p *createOnlyStructureProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}
func (p *createOnlyStructureProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}
func (p *createOnlyStructureProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{func() resource.Resource { return &createOnlyStructureResource{wholeValue: p.wholeValue} }}
}

type createOnlyStructureResource struct{ wholeValue bool }

func (r *createOnlyStructureResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "structure_test"
}
func (r *createOnlyStructureResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	objectModifier := RequiresReplaceObjectStructure()
	listModifier := RequiresReplaceListStructure()
	if r.wholeValue {
		objectModifier = objectplanmodifier.RequiresReplace()
		listModifier = listplanmodifier.RequiresReplace()
	}
	children := map[string]schema.Attribute{
		"name":   schema.StringAttribute{Description: "Immutable input.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"output": schema.StringAttribute{Description: "Server output.", Computed: true},
	}
	response.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"capacity": schema.Int64Attribute{Description: "Mutable capacity.", Required: true},
		"scope": schema.SingleNestedAttribute{
			Description: "Immutable scope.", Optional: true, Computed: true, Attributes: children,
			PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown(), objectModifier},
		},
		"selectors": schema.ListNestedAttribute{
			Description: "Immutable selectors.", Optional: true, Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: children},
			PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown(), listModifier},
		},
	}}
}
func (r *createOnlyStructureResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
}
func (r *createOnlyStructureResource) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {
}
func (r *createOnlyStructureResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *createOnlyStructureResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func TestCreateOnlyStructurePresence(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	elementTypes := map[string]attr.Type{"name": types.StringType}
	object := types.ObjectValueMust(elementTypes, map[string]attr.Value{"name": types.StringValue("value")})
	objectNull := types.ObjectNull(elementTypes)
	list := types.ListValueMust(types.ObjectType{AttrTypes: elementTypes}, []attr.Value{object})
	listNull := types.ListNull(types.ObjectType{AttrTypes: elementTypes})
	raw := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{})
	for _, remove := range []bool{false, true} {
		beforeObject, afterObject := objectNull, object
		beforeList, afterList := listNull, list
		if remove {
			beforeObject, afterObject = afterObject, beforeObject
			beforeList, afterList = afterList, beforeList
		}
		objectResponse := planmodifier.ObjectResponse{PlanValue: afterObject}
		RequiresReplaceObjectStructure().PlanModifyObject(ctx, planmodifier.ObjectRequest{
			State: tfsdk.State{Raw: raw}, Plan: tfsdk.Plan{Raw: raw}, StateValue: beforeObject, PlanValue: afterObject,
		}, &objectResponse)
		require.True(t, objectResponse.RequiresReplace, "object removal=%v", remove)
		listResponse := planmodifier.ListResponse{PlanValue: afterList}
		RequiresReplaceListStructure().PlanModifyList(ctx, planmodifier.ListRequest{
			State: tfsdk.State{Raw: raw}, Plan: tfsdk.Plan{Raw: raw}, StateValue: beforeList, PlanValue: afterList,
		}, &listResponse)
		require.True(t, listResponse.RequiresReplace, "list removal=%v", remove)
	}
}
