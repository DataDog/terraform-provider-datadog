package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes" // v0.1.0, else breaking
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSource = &workflowAutomationDatasource{}

type workflowAutomationDatasource struct {
	Api  *datadogV2.WorkflowAutomationApi
	Auth context.Context
}

func NewDatadogWorkflowAutomationDataSource() datasource.DataSource {
	return &workflowAutomationDatasource{}
}

func (d *workflowAutomationDatasource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetWorkflowAutomationApiV2()
	d.Auth = providerData.Auth
}

func (d *workflowAutomationDatasource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "workflow_automation"
}

func (d *workflowAutomationDatasource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "TODO",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the workflow.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the workflow.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				Description: "Tags of the workflow.",
				ElementType: types.StringType,
			},
			"published": schema.BoolAttribute{
				Computed:    true,
				Description: "Set the workflow to published or unpublished. Workflows in an unpublished state will only be executable via manual runs. Automatic triggers such as Schedule will not execute the workflow until it is published.",
			},
			"spec_json": schema.StringAttribute{
				Computed:    true,
				Description: "The spec defines what the workflow does.",
				CustomType:  jsontypes.NormalizedType{},
			},
			"webhook_secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "If a Webhook trigger is defined on this workflow, a webhookSecret is required and should be provided here.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the workflow was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the workflow was last updated.",
			},
		},
	}
}

func (d *workflowAutomationDatasource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state workflowAutomationResourceModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	workflowModel, err := readWorkflow(d.Auth, d.Api, state.ID.ValueString(), state)
	if err != nil {
		response.Diagnostics.AddError("Could not read workflow", err.Error())
		return
	}

	diags = response.State.Set(ctx, workflowModel)
	response.Diagnostics.Append(diags...)
}
