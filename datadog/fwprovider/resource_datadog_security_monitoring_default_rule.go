package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &securityMonitoringDefaultRuleResource{}
	_ resource.ResourceWithImportState = &securityMonitoringDefaultRuleResource{}
)

type securityMonitoringDefaultRuleResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

type securityMonitoringDefaultRuleResourceModel struct {
	ID            types.String `tfsdk:"id"`
	CustomMessage types.String `tfsdk:"custom_message"`
	CustomName    types.String `tfsdk:"custom_name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Type          types.String `tfsdk:"type"`
	CustomTags    types.Set    `tfsdk:"custom_tags"`

	Cases   []defaultRuleCaseModel    `tfsdk:"case"`
	Queries []defaultRuleQueryModel   `tfsdk:"query"`
	Filters []ruleFilterModel         `tfsdk:"filter"`
	Options []defaultRuleOptionsModel `tfsdk:"options"`
}

type defaultRuleCaseModel struct {
	Status        types.String `tfsdk:"status"`
	CustomStatus  types.String `tfsdk:"custom_status"`
	Notifications types.List   `tfsdk:"notifications"`
}

type defaultRuleQueryModel struct {
	AgentRules               []defaultRuleQueryAgentRuleModel `tfsdk:"agent_rule"`
	Aggregation              types.String                     `tfsdk:"aggregation"`
	DistinctFields           types.List                       `tfsdk:"distinct_fields"`
	GroupByFields            types.List                       `tfsdk:"group_by_fields"`
	HasOptionalGroupByFields types.Bool                       `tfsdk:"has_optional_group_by_fields"`
	DataSource               types.String                     `tfsdk:"data_source"`
	Metric                   types.String                     `tfsdk:"metric"`
	Metrics                  types.List                       `tfsdk:"metrics"`
	Name                     types.String                     `tfsdk:"name"`
	Query                    types.String                     `tfsdk:"query"`
	CustomQueryExtension     types.String                     `tfsdk:"custom_query_extension"`
}

type defaultRuleQueryAgentRuleModel struct {
	AgentRuleID types.String `tfsdk:"agent_rule_id"`
	Expression  types.String `tfsdk:"expression"`
}

type defaultRuleOptionsModel struct {
	DecreaseCriticalityBasedOnEnv types.Bool `tfsdk:"decrease_criticality_based_on_env"`
}

func NewSecurityMonitoringDefaultRuleResource() resource.Resource {
	return &securityMonitoringDefaultRuleResource{}
}

func (r *securityMonitoringDefaultRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (r *securityMonitoringDefaultRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_default_rule"
}

func (r *securityMonitoringDefaultRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Rule API resource for default rules. It can only be imported, you can't create a default rule.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"custom_message": schema.StringAttribute{
				Optional:    true,
				Description: "Custom Message (will override default message) for generated signals.",
			},
			"custom_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name (will override default name) of the rule.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Enable the rule.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The rule type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_tags": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Custom tags for generated signals.",
			},
		},
		Blocks: map[string]schema.Block{
			"case": schema.ListNestedBlock{
				Description: "Cases of the rule, this is used to update notifications.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"status": schema.StringAttribute{
							Required:    true,
							Description: "Status of the rule case to match.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
							},
						},
						"custom_status": schema.StringAttribute{
							Optional:    true,
							Description: "Status of the rule case to override.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
							},
						},
						"notifications": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Notification targets for each rule case.",
						},
					},
				},
			},
			"query": schema.ListNestedBlock{
				Description: "Queries for selecting logs which are part of the rule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"aggregation": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The aggregation type. For Signal Correlation rules, it must be event_count.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleQueryAggregationFromValue),
							},
						},
						"distinct_fields": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "Field for which the cardinality is measured. Sent as an array.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.List{
								listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
							},
						},
						"group_by_fields": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "Fields to group by.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.List{
								listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
							},
						},
						"has_optional_group_by_fields": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "When false, events without a group-by value are ignored by the rule. When true, events with missing group-by fields are processed with `N/A`, replacing the missing values.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"data_source": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Source of events.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								validators.SecurityMonitoringDataSourceWarningValidator(),
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringStandardDataSourceFromValue),
							},
						},
						"metric": schema.StringAttribute{
							Optional:           true,
							Computed:           true,
							Description:        "The target field to aggregate over when using the `sum`, `max`, or `geo_data` aggregations.",
							DeprecationMessage: "Configure `metrics` instead. This attribute will be removed in the next major version of the provider.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"metrics": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "Group of target fields to aggregate over when using the `sum`, `max`, `geo_data`, or `new_value` aggregations. The `sum`, `max`, and `geo_data` aggregations only accept one value in this list, whereas the `new_value` aggregation accepts up to five values.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Name of the query. Not compatible with `new_value` aggregations.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"query": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Query to run on logs.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"custom_query_extension": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Query extension to append to the logs query.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"agent_rule": schema.ListNestedBlock{
							Description:        "**Deprecated**. It won't be applied anymore.",
							DeprecationMessage: "`agent_rule` has been deprecated in favor of new Agent Rule resource.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"agent_rule_id": schema.StringAttribute{
										Required:    true,
										Description: "**Deprecated**. It won't be applied anymore.",
									},
									"expression": schema.StringAttribute{
										Required:    true,
										Description: "**Deprecated**. It won't be applied anymore.",
									},
								},
							},
						},
					},
				},
			},
			"filter": schema.ListNestedBlock{
				Description: "Additional queries to filter matched events before they are processed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Required:    true,
							Description: "The type of filtering action. Allowed enum values: require, suppress",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringFilterActionFromValue),
							},
						},
						"query": schema.StringAttribute{
							Required:    true,
							Description: "Query for selecting logs to apply the filtering action.",
						},
					},
				},
			},
			"options": schema.ListNestedBlock{
				Description: "Options on default rules. Note that only a subset of fields can be updated on default rule options.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"decrease_criticality_based_on_env": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "If true, signals in non-production environments have a lower severity than what is defined by the rule case, which can reduce noise. The decrement is applied when the environment tag of the signal starts with `staging`, `test`, or `dev`. Only available when the rule type is `log_detection`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *securityMonitoringDefaultRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	response.Diagnostics.AddError(
		"Default rule cannot be created",
		"cannot create a default rule, please import it first before making changes",
	)
}

func (r *securityMonitoringDefaultRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	response.Diagnostics.AddError("not implemented", "Read is not implemented yet for the framework default rule resource")
}

func (r *securityMonitoringDefaultRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("not implemented", "Update is not implemented yet for the framework default rule resource")
}

func (r *securityMonitoringDefaultRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// no-op
}

func (r *securityMonitoringDefaultRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
