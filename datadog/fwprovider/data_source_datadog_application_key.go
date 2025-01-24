package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSource = &applicationKeyDataSource{}

type applicationKeyDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ExactMatch types.Bool   `tfsdk:"exact_match"`
	Key        types.String `tfsdk:"key"`
}

type applicationKeyDataSource struct {
	Api  *datadogV2.KeyManagementApi
	Auth context.Context
}

func NewApplicationKeyDataSource() datasource.DataSource {
	return &applicationKeyDataSource{}
}

// Metadata implements datasource.DataSource.
func (d *applicationKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "application_key"
}

// Schema implements datasource.DataSource.
func (d *applicationKeyDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing application key. Deprecated. This will be removed in a future release with prior notice. Securely store your application keys using a secret management system or use the datadog_application_key resource to manage application keys in your Datadog account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id for Application Key.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name for Application Key.",
				Optional:    true,
			},
			"exact_match": schema.BoolAttribute{
				Description: "Whether to use exact match when searching by name.",
				Optional:    true,
			},
			"key": schema.StringAttribute{
				Description: "The value of the Application Key.",
				Computed:    true,
				Sensitive:   true,
			},
		},
		DeprecationMessage: "The datadog_application_key data source is deprecated and will be removed in a future release with prior notice. Securely store your application key using a secret management system or use the datadog_application_key resource to manage application keys in your Datadog account.",
	}
}

func (r *applicationKeyDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetKeyManagementApiV2()
	r.Auth = providerData.Auth
}

// Read implements datasource.DataSource.
func (d *applicationKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state applicationKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !state.Id.IsNull() {
		ddResp, _, err := d.Api.GetCurrentUserApplicationKey(d.Auth, state.Id.ValueString())
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting application key"))
			return
		}
		appKeyData := ddResp.GetData()
		if !d.checkAPIDeprecated(&appKeyData, resp) {
			d.updateState(&state, &appKeyData)
		}
	} else if !state.Name.IsNull() {
		optionalParams := datadogV2.NewListCurrentUserApplicationKeysOptionalParameters()
		optionalParams.WithFilter(state.Name.ValueString())
		applicationKeysResponse, _, err := d.Api.ListCurrentUserApplicationKeys(d.Auth, *optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting application keys"))
			return
		}
		applicationKeysData := applicationKeysResponse.GetData()
		if len(applicationKeysData) > 1 && !state.ExactMatch.ValueBool() {
			resp.Diagnostics.AddError("your query returned more than one result, please try a more specific search criteria", "")
			return
		}
		if len(applicationKeysData) == 0 {
			resp.Diagnostics.AddError("your query returned no result, please try a less specific search criteria", "")
			return
		}
		if state.ExactMatch.ValueBool() {
			exact_matches := 0
			var applicationKeyData datadogV2.FullApplicationKey
			for _, appKeyPartialData := range applicationKeysData {
				appKeyAttributes := appKeyPartialData.GetAttributes()
				if state.Name.ValueString() == appKeyAttributes.GetName() {
					exact_matches++
					id := appKeyPartialData.GetId()
					ddResp, _, err := d.Api.GetCurrentUserApplicationKey(d.Auth, id)
					if err != nil {
						resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting application key"))
						return
					}
					applicationKeyData = ddResp.GetData()
				}
			}
			if exact_matches > 1 {
				resp.Diagnostics.AddError("your query returned more than one exact match, please try a more specific search criteria", "")
				return
			}
			if exact_matches == 0 {
				resp.Diagnostics.AddError("your query returned no exact matches, please try a less specific search criteria", "")
				return
			}
			if !d.checkAPIDeprecated(&applicationKeyData, resp) {
				d.updateState(&state, &applicationKeyData)
			}
		} else {
			id := applicationKeysData[0].GetId()
			applicationKeyResponse, _, err := d.Api.GetCurrentUserApplicationKey(d.Auth, id)
			if err != nil {
				resp.Diagnostics.AddError("error getting application key", "")
				return
			}
			applicationKeyFullData := applicationKeyResponse.GetData()
			if !d.checkAPIDeprecated(&applicationKeyFullData, resp) {
				d.updateState(&state, &applicationKeyFullData)
			}
		}
	} else {
		resp.Diagnostics.AddError("missing id or name parameter", "")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationKeyDataSource) updateState(state *applicationKeyDataSourceModel, applicationKeyData *datadogV2.FullApplicationKey) {
	applicationKeyAttributes := applicationKeyData.GetAttributes()

	state.Id = types.StringValue(applicationKeyData.GetId())
	state.Name = types.StringValue(applicationKeyAttributes.GetName())
	state.Key = types.StringValue(applicationKeyAttributes.GetKey())
}

func (r *applicationKeyDataSource) checkAPIDeprecated(applicationKeyData *datadogV2.FullApplicationKey, resp *datasource.ReadResponse) bool {
	applicationKeyAttributes := applicationKeyData.GetAttributes()
	if !applicationKeyAttributes.HasKey() {
		resp.Diagnostics.AddError("Deprecated", "The datadog_application_key data source is deprecated and will be removed in a future release. Securely store your application key using a secret management system or use the datadog_application_key resource to manage application keys in your Datadog account.")
		return true
	}
	return false
}
