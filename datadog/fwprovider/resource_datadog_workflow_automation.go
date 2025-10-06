package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2" // v0.1.0, else breaking
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

type workflowAutomationResourceModel struct {
	ID            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	Tags          []types.String       `tfsdk:"tags"`
	Published     types.Bool           `tfsdk:"published"`
	SpecJson      jsontypes.Normalized `tfsdk:"spec_json"`
	WebhookSecret types.String         `tfsdk:"webhook_secret"`
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

	updateRequest, err := workflowAutomationModelToUpdateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("Error building update workflow request", err.Error())
		return
	}

	_, httpResp, err := r.Api.UpdateWorkflow(r.Auth, plan.ID.ValueString(), *updateRequest)
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

	err := json.Unmarshal([]byte(workflowAutomationModel.SpecJson.ValueString()), &attributes.Spec)
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
