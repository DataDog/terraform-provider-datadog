package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogMonitorValidationDataSource{}
)

type datadogMonitorValidationDataSourceModel struct {
	// Query Parameters
	Query types.String `tfsdk:"query"`
	Type  types.String `tfsdk:"type"`

	// Results
	ID               types.String `tfsdk:"id"`
	ValidationErrors types.List   `tfsdk:"validation_errors"`
	Valid            types.Bool   `tfsdk:"valid"`
}

type datadogMonitorValidationDataSource struct {
	Api  *datadogV1.MonitorsApi
	Auth context.Context
}

func NewDatadogMonitorValidationDataSource() datasource.DataSource {
	return &datadogMonitorValidationDataSource{}
}

func (d *datadogMonitorValidationDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetMonitorsApiV1()
	d.Auth = providerData.Auth
}

func (d *datadogMonitorValidationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "monitor_validation"
}

func (d *datadogMonitorValidationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	mt, _ := datadogV1.NewMonitorTypeFromValue("placeholder")
	mts := mt.GetAllowedValues()
	allowedTypes := []string{}
	for _, t := range mts {
		allowedTypes = append(allowedTypes, string(t))
	}

	resp.Schema = schema.Schema{
		Description: "Use this data source to validate a monitor.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"query": schema.StringAttribute{
				Required:    true,
				Description: "The monitor query.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The monitor type.",
				Validators: []validator.String{
					stringvalidator.OneOf(allowedTypes...),
				},
			},

			// computed values
			"valid": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether or not the monitor is valid",
			},
			"validation_errors": schema.ListAttribute{
				Computed:    true,
				Description: "A list of validation errors included in the Datadog API response body",
				ElementType: types.StringType,
			},
		},
	}
}

func (d *datadogMonitorValidationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogMonitorValidationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	valid := true
	validationErrs, _ := types.ListValueFrom(ctx, types.StringType, []string{})

	_, ddResp, err := d.Api.ValidateMonitor(d.Auth, datadogV1.Monitor{
		Type:  datadogV1.MonitorType(state.Type.ValueString()),
		Query: state.Query.ValueString(),
	})
	if err != nil {
		switch v := err.(type) {
		case datadog.GenericOpenAPIError:
			if ddResp.StatusCode == 400 {
				valid = false
				errModel := v.ErrorModel
				apiErr, ok := errModel.(datadogV1.APIErrorResponse)
				if ok {
					validationErrs, _ = types.ListValueFrom(ctx, types.StringType, apiErr.Errors)
				}
			}
		default:
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error validating monitor"))
			return
		}
	}

	hashingData := fmt.Sprintf("%s:%s", state.Query, state.Type)
	state.ID = types.StringValue(utils.ConvertToSha256(hashingData))
	state.Query = state.Query
	state.Type = state.Type
	state.Valid = types.BoolValue(valid)
	state.ValidationErrors = validationErrs
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
