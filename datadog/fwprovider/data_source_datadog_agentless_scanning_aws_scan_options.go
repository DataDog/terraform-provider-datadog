package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &agentlessScanningAwsScanOptionsDataSource{}
)

type AwsScanOptionsModel struct {
	AwsAccountId     types.String `tfsdk:"aws_account_id"`
	Lambda           types.Bool   `tfsdk:"lambda"`
	SensitiveData    types.Bool   `tfsdk:"sensitive_data"`
	VulnContainersOs types.Bool   `tfsdk:"vuln_containers_os"`
	VulnHostOs       types.Bool   `tfsdk:"vuln_host_os"`
}

type agentlessScanningAwsScanOptionsDataSourceModel struct {
	ID             types.String           `tfsdk:"id"`
	AwsScanOptions []*AwsScanOptionsModel `tfsdk:"aws_scan_options"`
}

type agentlessScanningAwsScanOptionsDataSource struct {
	Api  *datadogV2.AgentlessScanningApi
	Auth context.Context
}

func NewAgentlessScanningAwsScanOptionsDataSource() datasource.DataSource {
	return &agentlessScanningAwsScanOptionsDataSource{}
}

func (d *agentlessScanningAwsScanOptionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAgentlessScanningApiV2()
	d.Auth = providerData.Auth
}

func (d *agentlessScanningAwsScanOptionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "agentless_scanning_aws_scan_options"
}

func (d *agentlessScanningAwsScanOptionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all AWS scan options for Datadog Agentless Scanning.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"aws_scan_options": schema.ListAttribute{
				Computed:    true,
				Description: "List of AWS scan options.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"aws_account_id":     types.StringType,
						"lambda":             types.BoolType,
						"sensitive_data":     types.BoolType,
						"vuln_containers_os": types.BoolType,
						"vuln_host_os":       types.BoolType,
					},
				},
			},
		},
	}
}

func (d *agentlessScanningAwsScanOptionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state agentlessScanningAwsScanOptionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddResp, _, err := d.Api.ListAwsScanOptions(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing AWS scan options"))
		return
	}

	d.updateState(&state, ddResp.GetData())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *agentlessScanningAwsScanOptionsDataSource) updateState(state *agentlessScanningAwsScanOptionsDataSourceModel, scanOptionsData []datadogV2.AwsScanOptionsData) {
	var awsScanOptions []*AwsScanOptionsModel
	for _, scanOption := range scanOptionsData {
		attributes := scanOption.GetAttributes()
		opt := AwsScanOptionsModel{
			AwsAccountId:     types.StringValue(scanOption.GetId()),
			Lambda:           types.BoolValue(attributes.GetLambda()),
			SensitiveData:    types.BoolValue(attributes.GetSensitiveData()),
			VulnContainersOs: types.BoolValue(attributes.GetVulnContainersOs()),
			VulnHostOs:       types.BoolValue(attributes.GetVulnHostOs()),
		}
		awsScanOptions = append(awsScanOptions, &opt)
	}

	state.ID = types.StringValue(utils.ConvertToSha256("agentless_scanning_aws_scan_options"))
	state.AwsScanOptions = awsScanOptions
}
