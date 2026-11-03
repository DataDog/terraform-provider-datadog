package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const metricTagConfigurationVisibilityTimeout = time.Minute

var (
	_ resource.ResourceWithConfigure      = &metricTagConfigurationV2Resource{}
	_ resource.ResourceWithImportState    = &metricTagConfigurationV2Resource{}
	_ resource.ResourceWithValidateConfig = &metricTagConfigurationV2Resource{}
)

type metricTagConfigurationV2Resource struct {
	API               *datadogV2.MetricsApi
	Auth              context.Context
	visibilityTimeout time.Duration
}

type metricTagConfigurationV2Model struct {
	ID              types.String `tfsdk:"id"`
	MetricName      types.String `tfsdk:"metric_name"`
	MetricType      types.String `tfsdk:"metric_type"`
	Tags            types.Set    `tfsdk:"tags"`
	ExcludeTagsMode types.Bool   `tfsdk:"exclude_tags_mode"`
}

func NewMetricTagConfigurationV2Resource() resource.Resource {
	return &metricTagConfigurationV2Resource{visibilityTimeout: metricTagConfigurationVisibilityTimeout}
}

func (r *metricTagConfigurationV2Resource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "metric_tag_configuration_v2"
}

func (r *metricTagConfigurationV2Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	provider, ok := req.ProviderData.(*FrameworkProvider)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider configuration", fmt.Sprintf("Expected *FrameworkProvider, got %T.", req.ProviderData))
		return
	}
	r.API = provider.DatadogApiInstances.GetMetricsApiV2()
	r.Auth = provider.Auth
}

func (r *metricTagConfigurationV2Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a metric's tag configuration independently of percentile aggregations. Creating, updating, or deleting this resource does not configure percentiles. Do not manage the same tag configuration with this resource and datadog_metric_tag_configuration simultaneously. Historical Metrics Ingestion is incompatible with a tag configuration. See the [migration guide](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/guides/metric_tag_configuration_v2_migration) before migrating an existing resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"metric_name": schema.StringAttribute{
				Description: "The name of an existing metric. Also used as the resource ID and import identifier. Changing the name replaces the resource.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._]*$`), "must start with a letter and contain only letters, numbers, dots, and underscores"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"metric_type": schema.StringAttribute{
				Description:   "The metric type: count, gauge, rate, or distribution. Changing the type replaces the resource.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.OneOf("count", "gauge", "rate", "distribution")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"tags": schema.SetAttribute{
				Description: "Tag keys that remain queryable. An empty set keeps no tags when exclude_tags_mode is false. This resource does not modify the metric's percentile setting.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{setvalidator.ValueStringsAre(
					stringvalidator.LengthBetween(1, 200),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z][A-Za-z0-9.\-_:\/]*$`), "must be a valid metric tag key"),
				)},
			},
			"exclude_tags_mode": schema.BoolAttribute{
				Description: "When true, tags lists the keys to exclude rather than include. Requires at least one tag. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *metricTagConfigurationV2Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config metricTagConfigurationV2Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.Tags.IsNull() || config.Tags.IsUnknown() || config.ExcludeTagsMode.IsUnknown() {
		return
	}
	if config.ExcludeTagsMode.ValueBool() && len(config.Tags.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(path.Root("exclude_tags_mode"), "Invalid tag exclusion", "exclude_tags_mode requires at least one tag.")
	}
}

func (r *metricTagConfigurationV2Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Keep all provider authentication, server-selection, and delegated-token values,
// while requests and retries use the Terraform operation's cancellation/deadline.
// Copying only API keys would lose the provider's other authentication modes.
type metricTagConfigurationContext struct {
	context.Context
	auth context.Context
}

func (c metricTagConfigurationContext) Value(key any) any {
	if value := c.Context.Value(key); value != nil {
		return value
	}
	if c.auth != nil {
		return c.auth.Value(key)
	}
	return nil
}

func (r *metricTagConfigurationV2Resource) operationContext(ctx context.Context) context.Context {
	return metricTagConfigurationContext{Context: ctx, auth: r.Auth}
}

func (r *metricTagConfigurationV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state metricTagConfigurationV2Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Passthrough import initially supplies only ID, not metric_name.
	result, httpResp, err := r.API.ListTagConfigurationByName(r.operationContext(ctx), state.ID.ValueString())
	if err != nil {
		if ctx.Err() == nil && httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "retrieving metric tag configuration"), "Error reading metric tag configuration"))
		return
	}
	if err := utils.CheckForUnparsed(result); err != nil {
		resp.Diagnostics.AddError("Unparsed metric tag configuration", err.Error())
		return
	}
	resp.Diagnostics.Append(r.updateState(ctx, &state, result.GetData())...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *metricTagConfigurationV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan metricTagConfigurationV2Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tags, diags := metricTagConfigurationV2Tags(ctx, plan.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	metricType, err := datadogV2.NewMetricTagConfigurationMetricTypesFromValue(plan.MetricType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid metric type", err.Error())
		return
	}
	attrs := datadogV2.NewMetricTagConfigurationCreateAttributesWithDefaults()
	attrs.SetMetricType(*metricType)
	attrs.SetTags(tags)
	attrs.SetExcludeTagsMode(plan.ExcludeTagsMode.ValueBool())
	// Be explicit about omission, including with SDK default constructors.
	attrs.IncludePercentiles = nil
	data := datadogV2.NewMetricTagConfigurationCreateDataWithDefaults()
	data.SetId(plan.MetricName.ValueString())
	data.SetAttributes(*attrs)
	body := datadogV2.NewMetricTagConfigurationCreateRequestWithDefaults()
	body.SetData(*data)

	result, httpResp, err := r.API.CreateTagConfiguration(r.operationContext(ctx), plan.MetricName.ValueString(), *body)
	// A 2xx response acknowledges the write even if the SDK cannot decode its
	// body. The requested metric name is enough to retain recoverable ownership.
	state := plan
	state.ID = state.MetricName
	if err == nil || (httpResp != nil && httpResp.StatusCode >= 200 && httpResp.StatusCode < 300) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "creating metric tag configuration"), "Error creating metric tag configuration"))
		return
	}
	if err := utils.CheckForUnparsed(result); err != nil {
		resp.Diagnostics.AddError("Unparsed metric tag configuration", err.Error())
		return
	}
	resp.Diagnostics.Append(r.updateState(ctx, &state, result.GetData())...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if !resp.Diagnostics.HasError() {
		if err := r.waitForVisibility(ctx, &plan, false); err != nil {
			resp.Diagnostics.AddError("Metric tag configuration created but readback did not converge", err.Error())
		}
	}
}

func (r *metricTagConfigurationV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan metricTagConfigurationV2Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tags, diags := metricTagConfigurationV2Tags(ctx, plan.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	attrs := datadogV2.NewMetricTagConfigurationUpdateAttributesWithDefaults()
	attrs.SetTags(tags)
	attrs.SetExcludeTagsMode(plan.ExcludeTagsMode.ValueBool())
	attrs.IncludePercentiles = nil
	data := datadogV2.NewMetricTagConfigurationUpdateDataWithDefaults()
	data.SetId(plan.MetricName.ValueString())
	data.SetAttributes(*attrs)
	body := datadogV2.NewMetricTagConfigurationUpdateRequestWithDefaults()
	body.SetData(*data)

	result, httpResp, err := r.API.UpdateTagConfiguration(r.operationContext(ctx), plan.MetricName.ValueString(), *body)
	state := plan
	state.ID = state.MetricName
	if err == nil || (httpResp != nil && httpResp.StatusCode >= 200 && httpResp.StatusCode < 300) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "updating metric tag configuration"), "Error updating metric tag configuration"))
		return
	}
	if err := utils.CheckForUnparsed(result); err != nil {
		resp.Diagnostics.AddError("Unparsed metric tag configuration", err.Error())
		return
	}
	resp.Diagnostics.Append(r.updateState(ctx, &state, result.GetData())...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if !resp.Diagnostics.HasError() {
		if err := r.waitForVisibility(ctx, &plan, false); err != nil {
			resp.Diagnostics.AddError("Metric tag configuration updated but readback did not converge", err.Error())
		}
	}
}

func (r *metricTagConfigurationV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state metricTagConfigurationV2Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	httpResp, err := r.API.DeleteTagConfiguration(r.operationContext(ctx), state.ID.ValueString())
	if err != nil {
		if ctx.Err() == nil && httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "deleting metric tag configuration"), "Error deleting metric tag configuration"))
		return
	}
	if err := r.waitForVisibility(ctx, &state, true); err != nil {
		resp.Diagnostics.AddError("Metric tag configuration delete readback did not converge", err.Error())
	}
}

func metricTagConfigurationV2Tags(ctx context.Context, value types.Set) ([]string, diag.Diagnostics) {
	// Preserve an explicit empty array, not JSON null.
	tags := make([]string, 0, len(value.Elements()))
	diags := value.ElementsAs(ctx, &tags, false)
	slices.Sort(tags)
	return tags, diags
}

func (r *metricTagConfigurationV2Resource) updateState(ctx context.Context, state *metricTagConfigurationV2Model, data datadogV2.MetricTagConfiguration) diag.Diagnostics {
	expectedID := state.MetricName.ValueString()
	if expectedID == "" {
		expectedID = state.ID.ValueString()
	}
	if data.GetId() == "" || (expectedID != "" && data.GetId() != expectedID) {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Invalid metric tag configuration response", "The response must identify the requested metric.")}
	}
	state.ID = types.StringValue(data.GetId())
	state.MetricName = state.ID
	attrs := data.GetAttributes()
	state.MetricType = types.StringValue(string(attrs.GetMetricType()))
	state.ExcludeTagsMode = types.BoolValue(attrs.GetExcludeTagsMode())
	tags := attrs.GetTags()
	if tags == nil {
		tags = []string{}
	}
	var diags diag.Diagnostics
	state.Tags, diags = types.SetValueFrom(ctx, types.StringType, tags)
	return diags
}

func (r *metricTagConfigurationV2Resource) waitForVisibility(ctx context.Context, wanted *metricTagConfigurationV2Model, absent bool) error {
	timeout := r.visibilityTimeout
	if timeout <= 0 {
		timeout = metricTagConfigurationVisibilityTimeout
	}
	// Bound the actual SDK request as well as RetryContext. Callback locals are
	// not shared with the caller: RetryContext may return during cancellation.
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	expectedTags, diags := metricTagConfigurationV2Tags(waitCtx, wanted.Tags)
	if diags.HasError() {
		return fmt.Errorf("invalid expected tags: %v", diags.Errors())
	}
	metricName := wanted.MetricName.ValueString()
	if metricName == "" {
		metricName = wanted.ID.ValueString()
	}
	return retry.RetryContext(waitCtx, timeout, func() *retry.RetryError {
		result, httpResp, err := r.API.ListTagConfigurationByName(r.operationContext(waitCtx), metricName)
		if err != nil {
			if waitCtx.Err() != nil {
				return retry.NonRetryableError(waitCtx.Err())
			}
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				if absent {
					return nil
				}
				return retry.RetryableError(fmt.Errorf("tag configuration for %q is not visible yet", metricName))
			}
			return retry.NonRetryableError(utils.TranslateClientError(err, httpResp, "verifying metric tag configuration"))
		}
		if err := utils.CheckForUnparsed(result); err != nil {
			return retry.NonRetryableError(err)
		}
		if absent {
			return retry.RetryableError(fmt.Errorf("tag configuration for %q is still visible after deletion", metricName))
		}
		data := result.GetData()
		attrs := data.GetAttributes()
		actualTags := slices.Clone(attrs.GetTags())
		slices.Sort(actualTags)
		if data.GetId() != metricName || string(attrs.GetMetricType()) != wanted.MetricType.ValueString() || attrs.GetExcludeTagsMode() != wanted.ExcludeTagsMode.ValueBool() || !slices.Equal(actualTags, expectedTags) {
			return retry.RetryableError(fmt.Errorf("tag configuration for %q has not reached the requested state", metricName))
		}
		return nil
	})
}
