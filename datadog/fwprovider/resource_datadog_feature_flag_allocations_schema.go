package fwprovider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func featureFlagAllocationsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Ordered allocations. The first matching targeting rule determines the variant distribution. Allocation keys are unique across the Datadog organization.",
		Required:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description: "Server-assigned allocation UUID.",
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"key": schema.StringAttribute{
					Description: "Stable allocation key. Keys must be unique across the organization, not only within this flag and environment.",
					Required:    true,
				},
				"name": schema.StringAttribute{
					Description: "Human-readable allocation name.",
					Required:    true,
				},
				"type": schema.StringAttribute{
					Description: "Allocation type: FEATURE_GATE or CANARY.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("FEATURE_GATE"),
					Validators: []validator.String{
						stringvalidator.OneOf("FEATURE_GATE", "CANARY"),
					},
				},
				"order_position": schema.Int64Attribute{
					Description: "Server-assigned evaluation position.",
					Computed:    true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
				"experiment_id": schema.StringAttribute{
					Description: "Optional experiment UUID associated with this allocation.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"targeting_rules":   featureFlagTargetingRulesAttribute(),
				"variant_weights":   featureFlagVariantWeightsAttribute(),
				"guardrail_metrics": featureFlagGuardrailMetricsAttribute(),
				"exposure_schedule": featureFlagExposureScheduleAttribute(),
			},
		},
	}
}

func featureFlagTargetingRulesAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Alternative targeting rules. Rules are ORed; conditions within one rule are ANDed. Use an empty list for an allocation that matches all traffic.",
		Required:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"conditions": schema.ListNestedAttribute{
					Description: "Conditions that must all match for this targeting rule.",
					Required:    true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"attribute": schema.StringAttribute{
								Description: "Evaluation-context attribute to compare.",
								Required:    true,
							},
							"operator": schema.StringAttribute{
								Description: "Comparison operator: EQUALS, ONE_OF, NOT_ONE_OF, MATCHES, NOT_MATCHES, LT, LTE, GT, GTE, or IS_NULL.",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("EQUALS", "ONE_OF", "NOT_ONE_OF", "MATCHES", "NOT_MATCHES", "LT", "LTE", "GT", "GTE", "IS_NULL"),
								},
							},
							"value": schema.ListAttribute{
								Description: "String values used by the operator.",
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
		},
	}
}

func featureFlagVariantWeightsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Variant distribution served when the allocation matches. Percentage weights should sum to 100.",
		Required:    true,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"variant_key": schema.StringAttribute{
					Description: "Key of a variant declared on the parent feature flag.",
					Required:    true,
				},
				"value": schema.Float64Attribute{
					Description: "Percentage weight from 0 through 100.",
					Required:    true,
					Validators: []validator.Float64{
						float64validator.Between(0, 100),
					},
				},
			},
		},
	}
}

func featureFlagGuardrailMetricsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Metrics that pause or abort a progressive rollout when they trigger.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"metric_id": schema.StringAttribute{
					Description: "Datadog metric monitor identifier used as a guardrail.",
					Required:    true,
				},
				"trigger_action": schema.StringAttribute{
					Description: "Action when the metric triggers: PAUSE or ABORT.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.OneOf("PAUSE", "ABORT"),
					},
				},
			},
		},
	}
}

func featureFlagExposureScheduleAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Declarative progressive rollout schedule. Runtime start, pause, resume, and stop actions are not managed by this resource.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Server-assigned exposure schedule UUID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"control_variant_key": schema.StringAttribute{
				Description: "Key of the control variant used for rollout analysis.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"absolute_start_time": schema.StringAttribute{
				Description: "RFC3339 timestamp at which the rollout starts.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rollout_options": schema.SingleNestedAttribute{
				Description: "Rollout cadence and automatic-start settings.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"strategy": schema.StringAttribute{
						Description: "Rollout strategy: UNIFORM_INTERVALS or NO_ROLLOUT.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("UNIFORM_INTERVALS", "NO_ROLLOUT"),
						},
					},
					"autostart": schema.BoolAttribute{
						Description: "Whether Datadog starts the rollout automatically.",
						Optional:    true,
						Computed:    true,
					},
					"selection_interval_ms": schema.Int64Attribute{
						Description: "Interval between rollout steps in milliseconds.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"rollout_steps": schema.ListNestedAttribute{
				Description: "Ordered exposure steps. Exposure ratios range from 0 through 1.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Server-assigned rollout step UUID.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"order_position": schema.Int64Attribute{
							Description: "Server-assigned step position.",
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"exposure_ratio": schema.Float64Attribute{
							Description: "Fraction of matching traffic exposed, from 0 through 1.",
							Required:    true,
							Validators: []validator.Float64{
								float64validator.Between(0, 1),
							},
						},
						"interval_ms": schema.Int64Attribute{
							Description: "Optional interval before this step in milliseconds.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"is_pause_record": schema.BoolAttribute{
							Description: "Whether this entry represents a pause record.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"grouped_step_index": schema.Int64Attribute{
							Description: "Index grouping steps evaluated together. Defaults to the step's list index.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}
