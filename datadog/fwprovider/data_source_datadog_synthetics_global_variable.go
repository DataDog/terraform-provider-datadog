package fwprovider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogSyntheticsGlobalVariableDataSource{}
)

func NewDatadogSyntheticsGlobalVariableDataSource() datasource.DataSource {
	return &datadogSyntheticsGlobalVariableDataSource{}
}

type datadogSyntheticsGlobalVariableDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Tags types.List   `tfsdk:"tags"`
}

type datadogSyntheticsGlobalVariableDataSource struct {
	Api  *datadogV1.SyntheticsApi
	Auth context.Context
}

func (d *datadogSyntheticsGlobalVariableDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetSyntheticsApiV1()
	d.Auth = providerData.Auth
}

func (d *datadogSyntheticsGlobalVariableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "synthetics_global_variable"
}

func (d *datadogSyntheticsGlobalVariableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a Datadog Synthetics global variable (to be used in Synthetics tests).",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "The synthetics global variable name to search for. Must only match one global variable.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
				},
			},
			"tags": schema.ListAttribute{
				Description: "A list of tags assigned to the Synthetics global variable.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}

}

func (d *datadogSyntheticsGlobalVariableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogSyntheticsGlobalVariableDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	globalVariables, _, err := d.Api.ListGlobalVariables(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting synthetics global variables"))
		return
	}

	searchedName := state.Name.ValueString()
	var matchedGlobalVariables []datadogV1.SyntheticsGlobalVariable
	for _, globalVariable := range globalVariables.Variables {
		if globalVariable.Name == searchedName {
			matchedGlobalVariables = append(matchedGlobalVariables, globalVariable)
		}
	}

	if len(matchedGlobalVariables) == 0 {
		resp.Diagnostics.AddError(fmt.Sprintf("Couldn't find synthetics global variable named %s", searchedName), "")
		return
	} else if len(matchedGlobalVariables) > 1 {
		resp.Diagnostics.AddError(fmt.Sprintf("Found multiple synthetics global variables named %s", searchedName), "")
		return
	}

	if err := utils.CheckForUnparsed(matchedGlobalVariables); err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error parsing synthetics global variables"))
		return
	}

	d.updateState(ctx, &state, &matchedGlobalVariables[0])

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *datadogSyntheticsGlobalVariableDataSource) updateState(ctx context.Context, state *datadogSyntheticsGlobalVariableDataSourceModel, globalVariable *datadogV1.SyntheticsGlobalVariable) {
	state.Id = types.StringValue(globalVariable.GetId())
	state.Name = types.StringValue(globalVariable.GetName())
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, globalVariable.GetTags())
}
