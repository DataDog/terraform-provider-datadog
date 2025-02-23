package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes" // v0.1.0, else breaking
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &workflowAutomationResource{}
	_ resource.ResourceWithImportState = &workflowAutomationResource{}
)

type workflowAutomationResource struct {
	Api  *datadogV2.WorkflowAutomationApi
	Auth context.Context
}

// try single property JSON input -> validation will be handled on the API side
type workflowAutomationResourceModel struct {
	ID            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	Tags          []types.String       `tfsdk:"tags"`
	Published     types.Bool           `tfsdk:"published"`
	SpecJson      jsontypes.Normalized `tfsdk:"spec_json"`
	WebhookSecret types.String         `tfsdk:"webhook_secret"`
}

func NewDatadogWorkflowAutomationResource() resource.Resource {
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
		Description: "TODO",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the workflow.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the workflow.",
			},
			"tags": schema.SetAttribute{
				// we use TypeSet to represent tags to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Optional:    true,
				Description: "Tags of the workflow.",
				ElementType: types.StringType,
			},
			"published": schema.BoolAttribute{
				Optional:    true,
				Description: "Set the workflow to published or unpublished. Workflows in an unpublished state will only be executable via manual runs. Automatic triggers such as Schedule will not execute the workflow until it is published.",
			},
			"spec_json": schema.StringAttribute{
				Required:    true,
				Description: "The spec defines what the workflow does.",
				CustomType:  jsontypes.NormalizedType{},
			},
			"webhook_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "If a Webhook trigger is defined on this workflow, a webhookSecret is required and should be provided here.",
			},
		},
	}
}

func (r *workflowAutomationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *workflowAutomationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan workflowAutomationResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	createRequest, err := workflowAutomationModelToCreateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("Error building create workflow request", err.Error())
		return
	}

	workflow, httpResp, err := r.Api.CreateWorkflow(r.Auth, *createRequest)
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
	plan.ID = types.StringPointerValue(workflow.Data.Id)

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

	workflowAutomationModel, err := readWorkflow(r.Auth, r.Api, state.ID.ValueString(), state)
	if err != nil {
		response.Diagnostics.AddError("Could not read workflow", err.Error())
		return
	}

	diags = response.State.Set(ctx, workflowAutomationModel)
	response.Diagnostics.Append(diags...)
}

func (r *workflowAutomationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan workflowAutomationResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error parsing id as uuid", err.Error())
		return
	}

	updateRequest, err := workflowAutomationModelToUpdateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("Error building update workflow request", err.Error())
		return
	}

	_, httpResp, err := r.Api.UpdateWorkflow(r.Auth, id.String(), *updateRequest)
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

	res, err := r.Api.DeleteWorkflow(r.Auth, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Delete workflow failed", err.Error())
		return
	}

	if res.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			response.Diagnostics.AddError("Delete workflow failed", "Failed to read error")
		} else {
			response.Diagnostics.AddError("Delete workflow failed", string(body))
		}
	}
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

	err := json.Unmarshal([]byte(workflowAutomationModel.SpecJson.ValueString()), &attributes.Spec)
	if err != nil {
		err = fmt.Errorf("error unmarshalling spec json string to attributes.Spec struct: %s", err)
		return nil, err
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

	err := json.Unmarshal([]byte(workflowAutomationModel.SpecJson.ValueString()), &attributes.Spec)
	if err != nil {
		err = fmt.Errorf("error unmarshalling spec json string to attributes.Spec struct: %s", err)
		return nil, err
	}

	data := datadogV2.NewWorkflowDataUpdate(*attributes, datadogV2.WORKFLOWDATATYPE_WORKFLOWS)

	req := datadogV2.NewUpdateWorkflowRequest(*data)

	return req, nil
}

func apiResponseToWorkflowAutomationModel(workflow datadogV2.GetWorkflowResponse) (*workflowAutomationResourceModel, error) {
	workflowModel := &workflowAutomationResourceModel{
		ID: types.StringPointerValue(workflow.Data.Id),
	}

	attributes := workflow.Data.Attributes
	workflowModel.Name = types.StringValue(attributes.Name)
	workflowModel.Description = types.StringPointerValue(attributes.Description)
	workflowModel.Published = types.BoolPointerValue(attributes.Published)

	sort.Strings(attributes.Tags)
	var tags []types.String
	for _, tag := range attributes.Tags {
		tags = append(tags, types.StringValue(tag))
	}

	workflowModel.Tags = tags
	workflowModel.WebhookSecret = types.StringPointerValue(attributes.WebhookSecret)

	marshalledBytes, err := json.Marshal(attributes.Spec)
	if err != nil {
		err = fmt.Errorf("error marshaling attributes.Spec: %s", err)
		return nil, err
	}
	workflowModel.SpecJson = jsontypes.NewNormalizedValue(string(marshalledBytes))

	return workflowModel, nil
}

// Read logic is shared between data source and resource
func readWorkflow(authCtx context.Context, api *datadogV2.WorkflowAutomationApi, id string, currentState workflowAutomationResourceModel) (*workflowAutomationResourceModel, error) {
	workflow, httpResponse, err := api.GetWorkflow(authCtx, id)
	if err != nil {
		if httpResponse != nil {
			body, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				return nil, fmt.Errorf("could not read error response")
			}
			return nil, fmt.Errorf("%s", body)
		}
		return nil, err
	}

	if _, ok := workflow.GetDataOk(); !ok {
		return nil, fmt.Errorf("workflow not found")
	}

	workflowModel, err := apiResponseToWorkflowAutomationModel(workflow)
	if err != nil {
		return nil, err
	}

	return workflowModel, nil
}
