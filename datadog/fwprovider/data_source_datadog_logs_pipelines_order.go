package fwprovider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSourceWithConfigure = &logsPipelinesOrderDataSource{}
)

type logsPipelinesOrderDataSource struct {
	api  *datadogV1.LogsPipelinesApi
	auth context.Context
}

type logsPipelinesOrderDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	PipelineIds []string     `tfsdk:"pipeline_ids"`
}

func NewLogsPipelinesOrderDataSource() datasource.DataSource {
	return &logsPipelinesOrderDataSource{}
}

func (r *logsPipelinesOrderDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetLogsPipelinesApiV1()
	r.auth = providerData.Auth
}

func (*logsPipelinesOrderDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "logs_pipelines_order"
}

func (r *logsPipelinesOrderDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state logsPipelinesOrderDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, _, err := r.api.GetLogsPipelineOrder(r.auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while fetching logs pipelines order"))
		return
	}

	stateIdHash := sha256.Sum256([]byte(strings.Join(res.GetPipelineIds(), ";")))
	state.Id = types.StringValue(hex.EncodeToString(stateIdHash[:]))

	state.PipelineIds = res.GetPipelineIds()
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (*logsPipelinesOrderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve the current order of your log pipelines.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"pipeline_ids": schema.ListAttribute{
				Computed:    true,
				Description: "Array of strings identifying by their id(s) the pipeline(s) of your organization. For each pipeline, following the order of the array, logs are tested against the query filter and processed if matching.",
				ElementType: types.StringType,
			},
		},
	}
}
