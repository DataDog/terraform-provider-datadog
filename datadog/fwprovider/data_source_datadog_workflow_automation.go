package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes" // v0.1.0, else breaking
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &workflowAutomationDatasource{}

type workflowAutomationDatasource struct {
	Api  *datadogV2.WorkflowAutomationApi
	Auth context.Context
}

type workflowAutomationDatasourceModel struct {
	ID          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Description types.String         `tfsdk:"description"`
	Tags        []types.String       `tfsdk:"tags"`
	Published   types.Bool           `tfsdk:"published"`
	SpecJson    jsontypes.Normalized `tfsdk:"spec_json"`
}

func NewWorkflowAutomationDataSource() datasource.DataSource {
	return &workflowAutomationDatasource{}
}

func (d *workflowAutomationDatasource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetWorkflowAutomationApiV2()
	// Used to identify requests made from Terraform
	d.Api.Client.Cfg.AddDefaultHeader("X-Datadog-Workflow-Automation-Source", "terraform")
	d.Auth = providerData.Auth
}

func (d *workflowAutomationDatasource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "workflow_automation"
}

func (d *workflowAutomationDatasource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "This data source retrieves the definition of an existing Datadog workflow from Workflow Automation for use in other resources. This data source requires a [registered application key](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/app_key_registration).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the workflow.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the workflow.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the workflow.",
			},
			"tags": schema.SetAttribute{
				// we use TypeSet to represent tags to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Computed:    true,
				Description: "Tags of the workflow.",
				ElementType: types.StringType,
			},
			"published": schema.BoolAttribute{
				Computed:    true,
				Description: "Set the workflow to published or unpublished. Workflows in an unpublished state are only executable through manual runs. Automatic triggers such as Schedule do not execute the workflow until it is published.",
			},
			"spec_json": schema.StringAttribute{
				Computed:    true,
				Description: "The spec defines what the workflow does.",
				CustomType:  jsontypes.NormalizedType{},
			},
		},
	}
}

func (d *workflowAutomationDatasource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state workflowAutomationDatasourceModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	readResp, err, httpStatusCode := readWorkflow(d.Auth, d.Api, state.ID.ValueString())
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

	workflowModel, err := apiResponseToWorkflowAutomationDatasourceModel(readResp)
	if err != nil {
		response.Diagnostics.AddError("Could not create workflow data source model", err.Error())
		return
	}

	diags = response.State.Set(ctx, workflowModel)
	response.Diagnostics.Append(diags...)
}

func apiResponseToWorkflowAutomationDatasourceModel(workflow *datadogV2.GetWorkflowResponse) (*workflowAutomationDatasourceModel, error) {
	workflowModel := &workflowAutomationDatasourceModel{
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
