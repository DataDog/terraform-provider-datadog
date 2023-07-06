package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &syntheticsConcurrencyCap{}
	_ resource.ResourceWithImportState = &syntheticsConcurrencyCap{}
)

func NewSyntheticsConcurrencyCapResource() resource.Resource {
	return &syntheticsConcurrencyCap{}
}

type SyntheticsConcurrencyCapModel struct {
	ID                     types.String `tfsdk:"id"`
	OnDemandConcurrencyCap types.Int64  `tfsdk:"on_demand_concurrency_cap"`
}

type syntheticsConcurrencyCap struct {
	Api  *datadogV2.SyntheticsApi
	Auth context.Context
}

func (r *syntheticsConcurrencyCap) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSyntheticsApiV2()
	r.Auth = providerData.Auth
}

func (r *syntheticsConcurrencyCap) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "synthetics_concurrency_cap"
}

func (r *syntheticsConcurrencyCap) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Synthetics On Demand Concurrency Cap API resource. This can be used to manage the Concurrency Cap for Synthetic tests.",
		Attributes: map[string]schema.Attribute{
			"on_demand_concurrency_cap": schema.Int64Attribute{
				Description: "Value of the on-demand concurrency cap, customizing the number of Synthetic tests run in parallel.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			// Resource ID
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *syntheticsConcurrencyCap) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state SyntheticsConcurrencyCapModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateCap(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsConcurrencyCap) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state SyntheticsConcurrencyCapModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResponse, err := r.Api.GetOnDemandConcurrencyCap(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reading synthetics concurrency cap. http response: %v", httpResponse)))
		return
	}
	if respData, ok := resp.GetDataOk(); ok {
		if respAttributes, ok := respData.GetAttributesOk(); ok {
			if respConcurrencyCap, ok := respAttributes.GetOnDemandConcurrencyCapOk(); ok {
				state.OnDemandConcurrencyCap = types.Int64Value(int64(*respConcurrencyCap))
			}
		}
	}

	state.ID = types.StringValue("synthetics-concurrency-cap")
	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsConcurrencyCap) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state SyntheticsConcurrencyCapModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateCap(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsConcurrencyCap) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}

func (r *syntheticsConcurrencyCap) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *syntheticsConcurrencyCap) updateCap(state *SyntheticsConcurrencyCapModel, diag *diag.Diagnostics) {
	ddConcurrencyCap := datadogV2.NewOnDemandConcurrencyCapAttributesWithDefaults()
	ddConcurrencyCap.SetOnDemandConcurrencyCap(float64(state.OnDemandConcurrencyCap.ValueInt64()))

	updatedCap, httpResponse, err := r.Api.SetOnDemandConcurrencyCap(r.Auth, *ddConcurrencyCap)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error updating synthetics concurrency cap: %v", httpResponse)))
	}
	if err := utils.CheckForUnparsed(updatedCap); err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, ""))
	}

	if respData, ok := updatedCap.GetDataOk(); ok {
		if respAttributes, ok := respData.GetAttributesOk(); ok {
			if respConcurrencyCap, ok := respAttributes.GetOnDemandConcurrencyCapOk(); ok {
				state.OnDemandConcurrencyCap = types.Int64Value(int64(*respConcurrencyCap))
			}
		}
	}

	state.ID = types.StringValue("synthetics-concurrency-cap")
}
