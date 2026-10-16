package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &rumOperationResource{}
	_ resource.ResourceWithImportState = &rumOperationResource{}
)

type rumOperationResource struct {
	Api  *datadogV2.RUMOperationsApi
	Auth context.Context
}

type rumOperationModel struct {
	ID            types.String            `tfsdk:"id"`
	Name          types.String            `tfsdk:"name"`
	DisplayName   types.String            `tfsdk:"display_name"`
	Description   types.String            `tfsdk:"description"`
	Category      types.String            `tfsdk:"category"`
	ApplicationID types.String            `tfsdk:"application_id"`
	Tags          types.Set               `tfsdk:"tags"`
	Steps         []rumOperationStepModel `tfsdk:"steps"`
}

type rumOperationStepModel struct {
	Type      types.String                `tfsdk:"type"`
	Nodes     []rumOperationNodeModel     `tfsdk:"nodes"`
	Composite *rumOperationCompositeModel `tfsdk:"composite"`
}

type rumOperationNodeModel struct {
	Query types.String `tfsdk:"query"`
}

type rumOperationCompositeModel struct {
	Kind        types.String            `tfsdk:"kind"`
	MaxWindowMs types.Int64             `tfsdk:"max_window_ms"`
	Predicates  []rumOperationNodeModel `tfsdk:"predicates"`
}

func NewRumOperationResource() resource.Resource {
	return &rumOperationResource{}
}

func (r *rumOperationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRUMOperationsApiV2()
	r.Auth = providerData.Auth
}

func (r *rumOperationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "rum_operation"
}

func (r *rumOperationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	queryObject := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"query": schema.StringAttribute{
				Description: "RUM event search query.",
				Required:    true,
			},
		},
	}

	response.Schema = schema.Schema{
		Description: "Provides a Datadog RUM Operation resource. This can be used to create and manage RUM operations. " +
			"When automatic SLO creation is enabled for the organization, creating an operation also creates an availability SLO " +
			"tagged `feature_operation:<name>`, and deleting it deletes the SLOs carrying that tag for the operation's application. " +
			"Renaming the operation or changing its application leaves the previous SLO in place.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Unique name of the operation within the organization. Must not contain spaces. Changing it recreates the operation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human readable name of the operation. Defaults to `name`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the operation.",
				Optional:    true,
			},
			"category": schema.StringAttribute{
				Description: "Category of the operation.",
				Optional:    true,
			},
			"application_id": schema.StringAttribute{
				Description: "ID of the RUM application the operation belongs to.",
				Required:    true,
			},
			"tags": schema.SetAttribute{
				Description: "Tags attached to the operation.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
			"steps": schema.ListNestedAttribute{
				Description: "Ordered steps defining how RUM events drive the operation lifecycle.",
				Required:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Lifecycle transition triggered by this step.",
							Required:    true,
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewRUMOperationJourneyStepTypeFromValue),
							},
						},
						"nodes": schema.ListNestedAttribute{
							Description:  "Queries matched by this step. Exactly one of `nodes` or `composite` must be set.",
							Optional:     true,
							Validators:   []validator.List{listvalidator.SizeAtLeast(1)},
							NestedObject: queryObject,
						},
						"composite": schema.SingleNestedAttribute{
							Description: "Rule combining several predicates into this step. Exactly one of `nodes` or `composite` must be set.",
							Optional:    true,
							Validators: []validator.Object{
								objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("nodes")),
							},
							Attributes: map[string]schema.Attribute{
								"kind": schema.StringAttribute{
									Description: "How the predicates combine.",
									Required:    true,
									Validators: []validator.String{
										validators.NewEnumValidator[validator.String](datadogV2.NewRUMOperationJourneyCompositeRuleKindFromValue),
									},
								},
								"max_window_ms": schema.Int64Attribute{
									Description: "Maximum time window, in milliseconds, in which all predicates must match.",
									Optional:    true,
									Validators:  []validator.Int64{int64validator.Between(0, 30*60*1000)},
								},
								"predicates": schema.ListNestedAttribute{
									Description:  "Predicates combined by the rule.",
									Required:     true,
									Validators:   []validator.List{listvalidator.SizeBetween(2, 16)},
									NestedObject: queryObject,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *rumOperationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *rumOperationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state rumOperationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetRUMOperation(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RUM operation"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumOperationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state rumOperationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes, diags := buildRumOperationRequestAttributes(ctx, &state, nil)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body := datadogV2.NewRUMOperationCreateRequest(*datadogV2.NewRUMOperationCreateRequestData(*attributes, datadogV2.RUMOPERATIONTYPE_OPERATIONS))

	resp, _, err := r.Api.CreateRUMOperation(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating RUM operation"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumOperationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state rumOperationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	// The API regenerates a composite rule id unless the update echoes it back, and ids are not part
	// of the schema, so fetch the stored ones to keep rules stable across applies.
	current, _, err := r.Api.GetRUMOperation(r.Auth, id)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RUM operation"))
		return
	}
	currentAttributes := current.Data.Attributes

	attributes, diags := buildRumOperationRequestAttributes(ctx, &state, &currentAttributes)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body := datadogV2.NewRUMOperationUpdateRequest(*datadogV2.NewRUMOperationUpdateRequestData(*attributes, id, datadogV2.RUMOPERATIONTYPE_OPERATIONS))

	resp, _, err := r.Api.UpdateRUMOperation(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating RUM operation"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumOperationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state rumOperationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteRUMOperation(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting RUM operation"))
	}
}

func (r *rumOperationResource) updateState(ctx context.Context, state *rumOperationModel, resp *datadogV2.RUMOperationResponse) diag.Diagnostics {
	data := resp.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId())
	state.Name = types.StringValue(attributes.GetName())
	state.DisplayName = types.StringValue(attributes.GetDisplayName())
	// The API stores a cleared description or category as an empty string.
	state.Description = nonEmptyStringValue(attributes.GetDescription())
	state.Category = nonEmptyStringValue(attributes.GetCategory())
	if applicationID, ok := attributes.GetApplicationIdOk(); ok && applicationID != nil {
		state.ApplicationID = types.StringValue(applicationID.String())
	} else {
		// Legacy operations have no application.
		state.ApplicationID = types.StringNull()
	}

	tags, diags := types.SetValueFrom(ctx, types.StringType, attributes.GetTags())
	state.Tags = tags

	journey := attributes.GetJourneyRum()
	steps := make([]rumOperationStepModel, 0, len(journey.GetRumSteps()))
	for _, step := range journey.GetRumSteps() {
		stepModel := rumOperationStepModel{Type: types.StringValue(string(step.GetType()))}
		for _, node := range step.GetNodes() {
			stepModel.Nodes = append(stepModel.Nodes, rumOperationNodeModel{Query: types.StringValue(node.GetQuery())})
		}
		if composite, ok := step.GetCompositeOk(); ok && composite != nil {
			compositeModel := &rumOperationCompositeModel{
				Kind:        types.StringValue(string(composite.GetKind())),
				MaxWindowMs: types.Int64PointerValue(composite.MaxWindowMs),
			}
			for _, predicate := range composite.GetPredicates() {
				compositeModel.Predicates = append(compositeModel.Predicates, rumOperationNodeModel{Query: types.StringValue(predicate.GetQuery())})
			}
			stepModel.Composite = compositeModel
		}
		steps = append(steps, stepModel)
	}
	state.Steps = steps

	return diags
}

// buildRumOperationRequestAttributes builds the full-replace payload shared by create and update.
// current is nil on create; on update it carries the stored operation so composite rule ids are kept.
func buildRumOperationRequestAttributes(ctx context.Context, state *rumOperationModel, current *datadogV2.RUMOperationResponseAttributes) (*datadogV2.RUMOperationRequestAttributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	var currentSteps []datadogV2.RUMOperationJourneyStep
	if current != nil {
		currentSteps = current.JourneyRum.GetRumSteps()
	}

	tags := []string{}
	diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)

	steps := make([]datadogV2.RUMOperationJourneyStep, 0, len(state.Steps))
	for i, stepModel := range state.Steps {
		step := datadogV2.NewRUMOperationJourneyStep(datadogV2.RUMOperationJourneyStepType(stepModel.Type.ValueString()))
		for _, node := range stepModel.Nodes {
			step.Nodes = append(step.Nodes, *datadogV2.NewRUMOperationJourneyNode(node.Query.ValueString()))
		}
		if c := stepModel.Composite; c != nil {
			predicates := make([]datadogV2.RUMOperationJourneyPredicate, 0, len(c.Predicates))
			for _, p := range c.Predicates {
				predicates = append(predicates, *datadogV2.NewRUMOperationJourneyPredicate(p.Query.ValueString()))
			}
			composite := datadogV2.NewRUMOperationJourneyCompositeRule(datadogV2.RUMOperationJourneyCompositeRuleKind(c.Kind.ValueString()), predicates)
			composite.MaxWindowMs = c.MaxWindowMs.ValueInt64Pointer()
			if i < len(currentSteps) {
				if stored, ok := currentSteps[i].GetCompositeOk(); ok && stored != nil {
					composite.CompositeRuleId = stored.CompositeRuleId
				}
			}
			step.Composite = composite
		}
		steps = append(steps, *step)
	}

	attributes := datadogV2.NewRUMOperationRequestAttributes(*datadogV2.NewRUMOperationJourneyRum(steps), state.Name.ValueString(), tags)

	applicationID, err := uuid.Parse(state.ApplicationID.ValueString())
	if err != nil {
		diags.AddAttributeError(path.Root("application_id"), "invalid application_id", "application_id must be a UUID")
		return nil, diags
	}
	attributes.SetApplicationId(applicationID)

	if !state.DisplayName.IsNull() && !state.DisplayName.IsUnknown() {
		attributes.SetDisplayName(state.DisplayName.ValueString())
	}
	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	// On update an omitted category keeps the stored one, so a removed category is sent as empty.
	if !state.Category.IsNull() {
		attributes.SetCategory(state.Category.ValueString())
	} else if current != nil {
		attributes.SetCategory("")
	}

	return attributes, diags
}

func nonEmptyStringValue(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
