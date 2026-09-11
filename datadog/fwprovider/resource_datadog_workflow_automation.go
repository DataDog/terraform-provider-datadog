package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2" // v0.1.0, else breaking
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure      = &workflowAutomationResource{}
	_ resource.ResourceWithImportState    = &workflowAutomationResource{}
	_ resource.ResourceWithValidateConfig = &workflowAutomationResource{}
)

type workflowAutomationResource struct {
	Api  *datadogV2.WorkflowAutomationApi
	Auth context.Context
}

type workflowAutomationResourceModel struct {
	ID            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	Tags          []types.String       `tfsdk:"tags"`
	Published     types.Bool           `tfsdk:"published"`
	SpecJson      jsontypes.Normalized `tfsdk:"spec_json"`
	WebhookSecret types.String         `tfsdk:"webhook_secret"`
	RunAs         types.Object         `tfsdk:"run_as"`
}

type workflowAutomationRunAsModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

var workflowAutomationRunAsAttributeTypes = map[string]attr.Type{
	"type": types.StringType,
	"id":   types.StringType,
}

func NewWorkflowAutomationResource() resource.Resource {
	return &workflowAutomationResource{}
}

func (r *workflowAutomationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetWorkflowAutomationApiV2()
	// Used to identify requests made from Terraform
	r.Api.Client.Cfg.AddDefaultHeader("X-Datadog-Workflow-Automation-Source", "terraform")
	r.Auth = providerData.Auth
}

func (r *workflowAutomationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "workflow_automation"
}

func (r *workflowAutomationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Enables the creation and management of Datadog workflows using Workflow Automation. To easily export a workflow for use with Terraform, use the export button in the Datadog Workflow Automation UI. This resource requires a [registered application key](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/app_key_registration).",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the workflow.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description of the workflow.",
			},
			"tags": schema.SetAttribute{
				// we use TypeSet to represent tags to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Required:    true,
				Description: "Tags of the workflow.",
				ElementType: types.StringType,
			},
			"published": schema.BoolAttribute{
				Required:    true,
				Description: "Set the workflow to published or unpublished. Workflows in an unpublished state are only executable through manual runs. Automatic triggers such as Schedule do not execute the workflow until it is published.",
			},
			"spec_json": schema.StringAttribute{
				Required:    true,
				Description: "The spec defines what the workflow does.",
				CustomType:  jsontypes.NormalizedType{},
			},
			"webhook_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "If a webhook trigger is defined on this workflow, a webhookSecret is required and should be provided here.",
				// BE validation requires 16 characters
				Validators: []validator.String{stringvalidator.LengthAtLeast(16)},
			},
			"run_as": schema.SingleNestedAttribute{
				Description: "Identity used to run the workflow. When omitted, the server-managed value is preserved.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseNonNullStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Type of identity used to run the workflow. `owner` uses the workflow owner, `initiator` uses the user who starts the execution, and `service_account` uses the account specified by `id`. Required when `run_as` is configured.",
						Validators: []validator.String{
							stringvalidator.OneOf("owner", "service_account", "initiator"),
						},
					},
					"id": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Service account identifier. Required when `type` is `service_account` and omitted otherwise.",
					},
				},
			},
		},
	}
}

func (r *workflowAutomationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *workflowAutomationResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config workflowAutomationResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() || config.RunAs.IsNull() || config.RunAs.IsUnknown() {
		return
	}

	runAs := workflowAutomationRunAsFromObject(config.RunAs)
	if runAs.Type.IsUnknown() {
		return
	}
	if runAs.Type.IsNull() {
		response.Diagnostics.AddAttributeError(
			frameworkPath.Root("run_as").AtName("type"),
			"Missing run_as type",
			"run_as.type must be provided when run_as is configured.",
		)
		return
	}

	runAsType := runAs.Type.ValueString()
	if runAsType == "service_account" {
		if runAs.ID.IsUnknown() {
			return
		}
		if runAs.ID.IsNull() || runAs.ID.ValueString() == "" {
			response.Diagnostics.AddAttributeError(
				frameworkPath.Root("run_as").AtName("id"),
				"Missing run_as service account ID",
				"run_as.id must be a non-empty value when run_as.type is service_account.",
			)
			return
		}
		if _, err := parseWorkflowServiceAccountID(runAs.ID.ValueString()); err != nil {
			response.Diagnostics.AddAttributeError(
				frameworkPath.Root("run_as").AtName("id"),
				"Invalid run_as service account ID",
				"run_as.id must be a canonical UUID when run_as.type is service_account.",
			)
		}
		return
	}

	if (runAsType == "owner" || runAsType == "initiator") && !runAs.ID.IsNull() && !runAs.ID.IsUnknown() {
		response.Diagnostics.AddAttributeError(
			frameworkPath.Root("run_as").AtName("id"),
			"Invalid run_as ID",
			"run_as.id must be omitted when run_as.type is owner or initiator.",
		)
	}
}

func (r *workflowAutomationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan workflowAutomationResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var config workflowAutomationResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	requestModel := plan
	requestModel.RunAs = workflowAutomationRunAsForRequest(
		config.RunAs,
		plan.RunAs,
		types.ObjectNull(workflowAutomationRunAsAttributeTypes),
	)
	createRequest, err := workflowAutomationModelToCreateApiRequest(requestModel)
	if err != nil {
		response.Diagnostics.AddError("Error building create workflow request", err.Error())
		return
	}

	createResp, httpResp, err := r.Api.CreateWorkflow(r.Auth, *createRequest)
	if err != nil {
		if httpResp != nil {
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("Error reading error response", err.Error())
				return
			}
			response.Diagnostics.AddError("Error creating workflow", string(body))
		} else {
			response.Diagnostics.AddError("Error creating workflow", err.Error())
		}
		return
	}

	// Set computed values
	plan.ID = types.StringPointerValue(createResp.Data.Id)
	plan.RunAs, err = apiWorkflowRunAsToModel(
		createResp.Data.Attributes.RunAsUserMode,
		createResp.Data.Relationships,
	)
	if err != nil {
		response.Diagnostics.AddError("Error reading run_as from create workflow response", err.Error())
		return
	}

	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *workflowAutomationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state workflowAutomationResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	readResp, err, httpStatusCode := readWorkflow(r.Auth, r.Api, state.ID.ValueString())
	if err != nil {
		if httpStatusCode == http.StatusNotFound {
			// If the workflow is not found, we log a warning and remove the resource from state. This may be due to changes in the UI.
			response.Diagnostics.AddWarning("The workflow with ID '"+state.ID.ValueString()+"' is not found. It may have been deleted outside of Terraform.", err.Error())
			response.State.RemoveResource(ctx)
			return
		}

		response.Diagnostics.AddError("Could not read workflow", err.Error())
		return
	}

	workflowModel, err := apiResponseToWorkflowAutomationResourceModel(readResp)
	if err != nil {
		response.Diagnostics.AddError("Could not create workflow resource model", err.Error())
		return
	}

	// Set webhookSecret to current state as it is never returned by the API
	workflowModel.WebhookSecret = state.WebhookSecret

	diags = response.State.Set(ctx, workflowModel)
	response.Diagnostics.Append(diags...)
}

func (r *workflowAutomationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan workflowAutomationResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var config workflowAutomationResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state workflowAutomationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	requestModel := plan
	requestModel.RunAs = workflowAutomationRunAsForRequest(config.RunAs, plan.RunAs, state.RunAs)
	updateRequest, err := workflowAutomationModelToUpdateApiRequest(requestModel)
	if err != nil {
		response.Diagnostics.AddError("Error building update workflow request", err.Error())
		return
	}

	updateResp, httpResp, err := r.Api.UpdateWorkflow(r.Auth, plan.ID.ValueString(), *updateRequest)
	if err != nil {
		if httpResp != nil {
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("Error reading error response", err.Error())
				return
			}
			response.Diagnostics.AddError("Error updating workflow", string(body))
		} else {
			response.Diagnostics.AddError("Error updating workflow", err.Error())
		}
		return
	}

	if updateResp.Data == nil {
		response.Diagnostics.AddError("Error reading run_as from update workflow response", "workflow response is missing data")
		return
	}
	plan.RunAs, err = apiWorkflowRunAsToModel(
		updateResp.Data.Attributes.RunAsUserMode,
		updateResp.Data.Relationships,
	)
	if err != nil {
		response.Diagnostics.AddError("Error reading run_as from update workflow response", err.Error())
		return
	}

	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *workflowAutomationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state workflowAutomationResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	deleteResp, err := r.Api.DeleteWorkflow(r.Auth, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Delete workflow failed", err.Error())
		return
	}

	if deleteResp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(deleteResp.Body)
		if err != nil {
			response.Diagnostics.AddError("Delete workflow failed", "Failed to read error")
		} else {
			response.Diagnostics.AddError("Delete workflow failed", string(body))
		}
	}
}

func parseWorkflowServiceAccountID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, err
	}
	if id.String() != value {
		return uuid.Nil, fmt.Errorf("must use canonical UUID format")
	}
	return id, nil
}

func workflowAutomationModelToApiRunAs(runAs types.Object) (*datadogV2.WorkflowRunAs, error) {
	if runAs.IsNull() || runAs.IsUnknown() {
		return nil, nil
	}

	runAsModel := workflowAutomationRunAsFromObject(runAs)
	runAsType := runAsModel.Type.ValueString()
	if (runAsType == "owner" || runAsType == "initiator") && !runAsModel.ID.IsNull() && !runAsModel.ID.IsUnknown() {
		return nil, fmt.Errorf("run_as.id must be omitted when run_as.type is owner or initiator")
	}

	var apiRunAs datadogV2.WorkflowRunAs
	switch runAsType {
	case "owner":
		apiRunAs = datadogV2.WorkflowRunAsOwnerAsWorkflowRunAs(
			datadogV2.NewWorkflowRunAsOwner(datadogV2.WORKFLOWRUNASOWNERTYPE_OWNER),
		)
	case "service_account":
		id, err := parseWorkflowServiceAccountID(runAsModel.ID.ValueString())
		if err != nil {
			return nil, fmt.Errorf("invalid run_as service account ID: %w", err)
		}
		apiRunAs = datadogV2.WorkflowRunAsServiceAccountAsWorkflowRunAs(
			datadogV2.NewWorkflowRunAsServiceAccount(id.String(), datadogV2.WORKFLOWRUNASSERVICEACCOUNTTYPE_SERVICE_ACCOUNT),
		)
	case "initiator":
		apiRunAs = datadogV2.WorkflowRunAsInitiatorAsWorkflowRunAs(
			datadogV2.NewWorkflowRunAsInitiator(datadogV2.WORKFLOWRUNASINITIATORTYPE_INITIATOR),
		)
	default:
		return nil, fmt.Errorf("invalid run_as type %q", runAsModel.Type.ValueString())
	}
	return &apiRunAs, nil
}

func workflowAutomationModelToCreateApiRequest(workflowAutomationModel workflowAutomationResourceModel) (*datadogV2.CreateWorkflowRequest, error) {
	attributes := datadogV2.NewWorkflowDataAttributesWithDefaults()
	attributes.SetName(workflowAutomationModel.Name.ValueString())
	attributes.SetDescription(workflowAutomationModel.Description.ValueString())
	tags := make([]string, len(workflowAutomationModel.Tags))
	for i, tag := range workflowAutomationModel.Tags {
		tags[i] = tag.ValueString()
	}
	sort.Strings(tags)
	attributes.SetTags(tags)
	attributes.SetPublished(workflowAutomationModel.Published.ValueBool())
	attributes.SetWebhookSecret(workflowAutomationModel.WebhookSecret.ValueString())
	runAs, err := workflowAutomationModelToApiRunAs(workflowAutomationModel.RunAs)
	if err != nil {
		return nil, err
	}
	if runAs != nil {
		attributes.SetRunAs(*runAs)
	}

	err = json.Unmarshal([]byte(workflowAutomationModel.SpecJson.ValueString()), &attributes.Spec)
	if err != nil {
		err = fmt.Errorf("error unmarshalling spec json string to attributes.Spec struct: %s", err)
		return nil, err
	}
	// Enforce strict decoding
	err = utils.CheckForAdditionalProperties(attributes.Spec)
	if err != nil {
		return nil, fmt.Errorf("unknown field in spec, this could be due to misspelled field, using a version of the Go client that is out of date, or support for this field has not been added. Check the [API](https://docs.datadoghq.com/api/latest/workflow-automation/#create-a-workflow) documentation for what fields are currently supported. Error: %s", err)
	}

	data := datadogV2.NewWorkflowData(*attributes, datadogV2.WORKFLOWDATATYPE_WORKFLOWS)
	req := datadogV2.NewCreateWorkflowRequest(*data)

	return req, nil
}

func workflowAutomationModelToUpdateApiRequest(workflowAutomationModel workflowAutomationResourceModel) (*datadogV2.UpdateWorkflowRequest, error) {
	attributes := datadogV2.NewWorkflowDataUpdateAttributesWithDefaults()
	attributes.SetName(workflowAutomationModel.Name.ValueString())
	attributes.SetDescription(workflowAutomationModel.Description.ValueString())
	tags := make([]string, len(workflowAutomationModel.Tags))
	for i, tag := range workflowAutomationModel.Tags {
		tags[i] = tag.ValueString()
	}
	sort.Strings(tags)
	attributes.SetTags(tags)
	attributes.SetPublished(workflowAutomationModel.Published.ValueBool())
	attributes.SetWebhookSecret(workflowAutomationModel.WebhookSecret.ValueString())
	runAs, err := workflowAutomationModelToApiRunAs(workflowAutomationModel.RunAs)
	if err != nil {
		return nil, err
	}
	if runAs != nil {
		attributes.SetRunAs(*runAs)
	}

	err = json.Unmarshal([]byte(workflowAutomationModel.SpecJson.ValueString()), &attributes.Spec)
	if err != nil {
		err = fmt.Errorf("error unmarshalling spec json string to attributes.Spec struct: %s", err)
		return nil, err
	}
	// Enforce strict decoding
	err = utils.CheckForAdditionalProperties(attributes.Spec)
	if err != nil {
		return nil, fmt.Errorf("unknown field in spec, this could be due to misspelled field, using a version of the Go client that is out of date, or support for this field has not been added. Check the [API](https://docs.datadoghq.com/api/latest/workflow-automation/#create-a-workflow) documentation for what fields are currently supported. Error: %s", err)
	}

	data := datadogV2.NewWorkflowDataUpdate(*attributes, datadogV2.WORKFLOWDATATYPE_WORKFLOWS)
	req := datadogV2.NewUpdateWorkflowRequest(*data)

	return req, nil
}

func apiResponseToWorkflowAutomationResourceModel(workflow *datadogV2.GetWorkflowResponse) (*workflowAutomationResourceModel, error) {
	workflowModel := &workflowAutomationResourceModel{
		ID: types.StringPointerValue(workflow.Data.Id),
	}

	attributes := workflow.Data.Attributes

	workflowModel.Name = types.StringValue(attributes.Name)

	if attributes.Description == nil {
		workflowModel.Description = types.StringValue("")
	} else {
		workflowModel.Description = types.StringPointerValue(attributes.Description)
	}

	workflowModel.Published = types.BoolPointerValue(attributes.Published)
	var err error
	workflowModel.RunAs, err = apiWorkflowRunAsToModel(attributes.RunAsUserMode, workflow.Data.Relationships)
	if err != nil {
		return nil, err
	}

	sort.Strings(attributes.Tags)
	var tags []types.String = make([]types.String, 0, len(attributes.Tags))
	for _, tag := range attributes.Tags {
		tags = append(tags, types.StringValue(tag))
	}
	workflowModel.Tags = tags

	marshalledBytes, err := json.Marshal(attributes.Spec)
	if err != nil {
		err = fmt.Errorf("error marshaling attributes.Spec: %s", err)
		return nil, err
	}
	workflowModel.SpecJson = jsontypes.NewNormalizedValue(string(marshalledBytes))

	return workflowModel, nil
}

func workflowAutomationRunAsFromObject(runAs types.Object) workflowAutomationRunAsModel {
	attributes := runAs.Attributes()
	return workflowAutomationRunAsModel{
		Type: attributes["type"].(types.String),
		ID:   attributes["id"].(types.String),
	}
}

func workflowAutomationRunAsForRequest(configured, planned, priorState types.Object) types.Object {
	null := types.ObjectNull(workflowAutomationRunAsAttributeTypes)
	if configured.IsNull() {
		return null
	}
	if planned.IsNull() || planned.IsUnknown() || priorState.IsNull() || priorState.IsUnknown() {
		return planned
	}

	plannedRunAs := workflowAutomationRunAsFromObject(planned)
	priorRunAs := workflowAutomationRunAsFromObject(priorState)
	if plannedRunAs.Type.IsNull() || plannedRunAs.Type.IsUnknown() ||
		priorRunAs.Type.IsNull() || priorRunAs.Type.IsUnknown() ||
		!plannedRunAs.Type.Equal(priorRunAs.Type) {
		return planned
	}
	if plannedRunAs.Type.ValueString() != "service_account" {
		if !plannedRunAs.ID.IsNull() && !plannedRunAs.ID.IsUnknown() {
			return planned
		}
		return null
	}
	if plannedRunAs.ID.IsNull() || plannedRunAs.ID.IsUnknown() ||
		priorRunAs.ID.IsNull() || priorRunAs.ID.IsUnknown() ||
		!plannedRunAs.ID.Equal(priorRunAs.ID) {
		return planned
	}
	return null
}

func apiWorkflowRunAsToModel(runAsUserMode *datadogV2.WorkflowRunAsUserMode, relationships *datadogV2.WorkflowDataRelationships) (types.Object, error) {
	if runAsUserMode == nil {
		return types.ObjectNull(workflowAutomationRunAsAttributeTypes), fmt.Errorf("workflow response is missing runAsUserMode")
	}

	runAsType := string(*runAsUserMode)
	if runAsType != "owner" && runAsType != "service_account" && runAsType != "initiator" {
		return types.ObjectNull(workflowAutomationRunAsAttributeTypes), fmt.Errorf("workflow response has unsupported runAsUserMode %q", runAsType)
	}

	id := types.StringNull()
	if runAsType == "service_account" {
		if relationships == nil || relationships.RunAs == nil || relationships.RunAs.Data == nil || relationships.RunAs.Data.Id == "" {
			return types.ObjectNull(workflowAutomationRunAsAttributeTypes), fmt.Errorf("workflow response is missing the runAs relationship for service_account mode")
		}
		id = types.StringValue(relationships.RunAs.Data.Id)
	}
	return types.ObjectValueMust(
		workflowAutomationRunAsAttributeTypes,
		map[string]attr.Value{
			"type": types.StringValue(runAsType),
			"id":   id,
		},
	), nil
}

// Read logic is shared between data source and resource
func readWorkflow(authCtx context.Context, api *datadogV2.WorkflowAutomationApi, id string) (*datadogV2.GetWorkflowResponse, error, int) {
	workflow, httpResponse, err := api.GetWorkflow(authCtx, id)
	if err != nil {
		if httpResponse != nil {
			body, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				return nil, fmt.Errorf("could not read error response"), httpResponse.StatusCode
			}
			return nil, fmt.Errorf("%s", body), httpResponse.StatusCode
		}
		return nil, err, httpResponse.StatusCode
	}

	if _, ok := workflow.GetDataOk(); !ok {
		return nil, fmt.Errorf("workflow not found"), httpResponse.StatusCode
	}

	return &workflow, nil, httpResponse.StatusCode
}
