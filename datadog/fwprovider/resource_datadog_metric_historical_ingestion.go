package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const historicalIngestionTimeout = 60 * time.Second

var (
	_ resource.ResourceWithConfigure   = &metricHistoricalIngestionResource{}
	_ resource.ResourceWithImportState = &metricHistoricalIngestionResource{}
)

type metricHistoricalIngestionResource struct {
	api  *datadogV2.MetricsApi
	auth context.Context
	// Override only in deterministic tests.
	timeout time.Duration
}

type metricHistoricalIngestionModel struct {
	ID         types.String `tfsdk:"id"`
	MetricName types.String `tfsdk:"metric_name"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func NewMetricHistoricalIngestionResource() resource.Resource {
	return &metricHistoricalIngestionResource{}
}

func (r *metricHistoricalIngestionResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "metric_historical_ingestion"
}

func (r *metricHistoricalIngestionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enables Historical Metrics Ingestion for an existing metric. This resource does not submit or backfill metric data. The API enforces eligibility: supported custom metrics are count, rate, or gauge; distributions, metrics generated from other Datadog data types, and metrics with existing tag or aggregation configurations are incompatible. Reading requires the metrics_read permission and changes require late_metrics_config_write. Import reads an already-enabled metric without enabling it; removing this resource disables ingestion. Create and delete wait up to 60 seconds for visibility; a timeout does not mean the write failed.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"metric_name": schema.StringAttribute{
				Description:   "Name of an existing eligible custom metric. Changing this value replaces the resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 200), stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._]*$`), "must start with a letter and contain only letters, digits, periods, and underscores")},
			},
			"created_at": schema.StringAttribute{Description: "Server-provided creation timestamp in RFC3339 format, if available.", Computed: true},
		},
	}
}

func (r *metricHistoricalIngestionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	p, ok := req.ProviderData.(*FrameworkProvider)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *FrameworkProvider, got %T.", req.ProviderData))
		return
	}
	r.api, r.auth = p.DatadogApiInstances.GetMetricsApiV2(), p.Auth
}

func (r *metricHistoricalIngestionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Use the operation's deadline and cancellation while retaining all provider auth,
// server, and other SDK context values, including future auth methods.
type historicalIngestionContext struct {
	context.Context
	auth context.Context
}

func (c historicalIngestionContext) Value(key interface{}) interface{} {
	if c.auth != nil {
		if v := c.auth.Value(key); v != nil {
			return v
		}
	}
	return c.Context.Value(key)
}
func (r *metricHistoricalIngestionResource) requestContext(ctx context.Context) context.Context {
	return historicalIngestionContext{Context: ctx, auth: r.auth}
}
func (r *metricHistoricalIngestionResource) budget() time.Duration {
	if r.timeout > 0 {
		return r.timeout
	}
	return historicalIngestionTimeout
}

// Do not blindly replay a mutation after an ambiguous SDK-level 5xx/429. The
// provider's client remains unchanged and its bounded retry policy still applies
// to GETs. Both clients use the same transport, auth and server configuration.
func (r *metricHistoricalIngestionResource) mutationAPI() *datadogV2.MetricsApi {
	cfg := *r.api.Client.Cfg
	cfg.RetryConfiguration.EnableRetry = false
	return datadogV2.NewMetricsApi(datadog.NewAPIClient(&cfg))
}

// Successful responses must identify the requested object and type. The timestamp
// and its parent attributes are optional, not evidence of presence.
func historicalIngestionState(state *metricHistoricalIngestionModel, result datadogV2.HistoricalMetricsConfigurationResponse, name string) error {
	if err := utils.CheckForUnparsed(result); err != nil {
		return err
	}
	data, ok := result.GetDataOk()
	if !ok || data == nil || data.Id == nil || *data.Id != name || data.Type == nil || *data.Type != datadogV2.HISTORICALMETRICSCONFIGURATIONTYPE_HISTORICAL_METRICS_CONFIGURATIONS {
		return fmt.Errorf("response is missing or contains a mismatched resource identity or type for metric %q", name)
	}
	state.ID, state.MetricName = types.StringValue(name), types.StringValue(name)
	state.CreatedAt = types.StringNull()
	if data.Attributes != nil && data.Attributes.CreatedAt != nil {
		state.CreatedAt = types.StringValue(data.Attributes.CreatedAt.UTC().Format(time.RFC3339Nano))
	}
	return nil
}

func (r *metricHistoricalIngestionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state metricHistoricalIngestionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := state.MetricName.ValueString()
	data := datadogV2.NewHistoricalMetricsConfigurationCreateData(name, datadogV2.HISTORICALMETRICSCONFIGURATIONTYPE_HISTORICAL_METRICS_CONFIGURATIONS)
	body := datadogV2.NewHistoricalMetricsConfigurationCreateRequest(*data)
	opCtx, cancel := context.WithTimeout(ctx, r.budget())
	defer cancel()
	result, httpResp, err := r.mutationAPI().CreateHistoricalMetricsConfiguration(r.requestContext(opCtx), *body)
	// The endpoint identifies the object by the requested metric. Once the server
	// acknowledges a successful write, retain that identity even if decoding or
	// later verification fails, so a subsequent run can recover it.
	if httpResp != nil && (httpResp.StatusCode == http.StatusOK || httpResp.StatusCode == http.StatusCreated) {
		state.ID = types.StringValue(name)
		state.CreatedAt = types.StringNull()
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error enabling historical ingestion"), "Error enabling historical ingestion"))
		return
	}
	if httpResp == nil || (httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated) {
		resp.Diagnostics.AddError("Invalid historical ingestion response", "The create request returned an unexpected success status; the resource was not verified.")
		return
	}
	if err := historicalIngestionState(&state, result, name); err != nil {
		resp.Diagnostics.AddError("Invalid historical ingestion response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, httpResp, err = r.wait(opCtx, name, true)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error waiting for historical ingestion visibility"), "Error verifying historical ingestion"))
		return
	}
	if err := historicalIngestionState(&state, result, name); err != nil {
		resp.Diagnostics.AddError("Invalid historical ingestion response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *metricHistoricalIngestionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state metricHistoricalIngestionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := state.ID.ValueString()
	result, httpResp, err := r.api.GetHistoricalMetricsConfiguration(r.requestContext(ctx), name)
	if err != nil {
		if ctx.Err() == nil && httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error reading historical ingestion"), "Error reading historical ingestion"))
		return
	}
	if httpResp == nil || httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Invalid historical ingestion response", "The read request returned an unexpected success status.")
		return
	}
	if err := historicalIngestionState(&state, result, name); err != nil {
		resp.Diagnostics.AddError("Invalid historical ingestion response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *metricHistoricalIngestionResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unexpected update", "Historical ingestion has no updatable arguments; changing metric_name replaces the resource.")
}

func (r *metricHistoricalIngestionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state metricHistoricalIngestionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := state.ID.ValueString()
	opCtx, cancel := context.WithTimeout(ctx, r.budget())
	defer cancel()
	httpResp, err := r.mutationAPI().DeleteHistoricalMetricsConfiguration(r.requestContext(opCtx), name)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error disabling historical ingestion"), "Error disabling historical ingestion"))
		return
	}
	if httpResp == nil || httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError("Invalid historical ingestion response", "The delete request returned an unexpected success status; absence was not verified.")
		return
	}
	_, httpResp, err = r.wait(opCtx, name, false)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error waiting for historical ingestion absence"), "Error verifying historical ingestion deletion"))
	}
}

func (r *metricHistoricalIngestionResource) wait(ctx context.Context, name string, present bool) (datadogV2.HistoricalMetricsConfigurationResponse, *http.Response, error) {
	var result datadogV2.HistoricalMetricsConfigurationResponse
	var httpResp *http.Response
	var mu sync.Mutex // RetryContext may return while the callback is still exiting.
	err := retry.RetryContext(ctx, r.budget(), func() *retry.RetryError {
		got, h, err := r.api.GetHistoricalMetricsConfiguration(r.requestContext(ctx), name)
		mu.Lock()
		result, httpResp = got, h
		mu.Unlock()
		if err != nil {
			if ctx.Err() == nil && h != nil && h.StatusCode == http.StatusNotFound {
				if !present {
					return nil
				}
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		if h == nil || h.StatusCode != http.StatusOK {
			return retry.NonRetryableError(fmt.Errorf("historical ingestion read returned an unexpected success status"))
		}
		var tmp metricHistoricalIngestionModel
		if err := historicalIngestionState(&tmp, got, name); err != nil {
			return retry.NonRetryableError(err)
		}
		if !present {
			return retry.RetryableError(fmt.Errorf("historical ingestion is still present for %q", name))
		}
		return nil
	})
	mu.Lock()
	defer mu.Unlock()
	return result, httpResp, err
}
