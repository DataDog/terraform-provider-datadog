package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ParseGrokProcessorModel struct {
	DisableLibraryRules types.Bool                           `tfsdk:"disable_library_rules"`
	Field               types.String                         `tfsdk:"field"`
	Rules               []ParseGrokProcessorRuleModel        `tfsdk:"rule"`
	IncludeRules         []ParseGrokProcessorIncludeRuleModel `tfsdk:"include_rule"`
}

type ParseGrokProcessorRuleModel struct {
	Source       types.String    `tfsdk:"source"`
	MatchRules   []GrokRuleModel `tfsdk:"match_rule"`
	SupportRules []GrokRuleModel `tfsdk:"support_rule"`
}

type ParseGrokProcessorIncludeRuleModel struct {
	Include      types.String    `tfsdk:"include"`
	MatchRules   []GrokRuleModel `tfsdk:"match_rule"`
	SupportRules []GrokRuleModel `tfsdk:"support_rule"`
}

type GrokRuleModel struct {
	Name types.String `tfsdk:"name"`
	Rule types.String `tfsdk:"rule"`
}

func grokMatchRuleBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "A list of Grok parsing rules that define how to extract fields. Each rule must contain a name and a valid Grok pattern.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "The name of the rule.",
				},
				"rule": schema.StringAttribute{
					Required:    true,
					Description: "The definition of the Grok rule.",
				},
			},
		},
	}
}

func grokSupportRuleBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "A list of helper Grok rules that can be referenced by the parsing rules.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "The name of the helper Grok rule.",
				},
				"rule": schema.StringAttribute{
					Required:    true,
					Description: "The definition of the helper Grok rule.",
				},
			},
		},
	}
}

func ParseGrokProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `parse_grok` processor extracts structured fields from unstructured log messages using Grok patterns.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"disable_library_rules": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "If set to `true`, disables the default Grok rules provided by Datadog.",
				},
				"field": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("message"),
					Description: "The log field to parse with the Grok rules.",
				},
			},
			Blocks: map[string]schema.Block{
				"rule": schema.ListNestedBlock{
					Description: "The list of Grok parsing rules. If multiple parsing rules are provided, they are evaluated in order. The first successful match is applied.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"source": schema.StringAttribute{
								Required:    true,
								Description: "The value of the source field in log events which should be processed by the Grok rules.",
							},
						},
						Blocks: map[string]schema.Block{
							"match_rule":   grokMatchRuleBlock(),
							"support_rule": grokSupportRuleBlock(),
						},
					},
				},
				"include_rule": schema.ListNestedBlock{
					Description: "A Grok parsing rule that targets logs matching a Datadog search query.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"include": schema.StringAttribute{
								Required:    true,
								Description: "A Datadog search query used to determine which logs this Grok rule targets.",
							},
						},
						Blocks: map[string]schema.Block{
							"match_rule":   grokMatchRuleBlock(),
							"support_rule": grokSupportRuleBlock(),
						},
					},
				},
			},
		},
	}
}

func expandGrokMatchRules(src []GrokRuleModel) []datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule {
	out := make([]datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule, 0, len(src))
	for _, m := range src {
		out = append(out, datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule{
			Name: m.Name.ValueString(),
			Rule: m.Rule.ValueString(),
		})
	}
	return out
}

func expandGrokSupportRules(src []GrokRuleModel) []datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule {
	out := make([]datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule, 0, len(src))
	for _, s := range src {
		out = append(out, datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule{
			Name: s.Name.ValueString(),
			Rule: s.Rule.ValueString(),
		})
	}
	return out
}

func flattenGrokMatchRules(src []datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule) []GrokRuleModel {
	out := make([]GrokRuleModel, 0, len(src))
	for _, m := range src {
		out = append(out, GrokRuleModel{
			Name: types.StringValue(m.GetName()),
			Rule: types.StringValue(m.GetRule()),
		})
	}
	return out
}

func flattenGrokSupportRules(src []datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule) []GrokRuleModel {
	out := make([]GrokRuleModel, 0, len(src))
	for _, s := range src {
		out = append(out, GrokRuleModel{
			Name: types.StringValue(s.GetName()),
			Rule: types.StringValue(s.GetRule()),
		})
	}
	return out
}

func ExpandParseGrokProcessor(common BaseProcessorFields, src *ParseGrokProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineParseGrokProcessorWithDefaults()
	common.ApplyTo(proc)

	if !src.DisableLibraryRules.IsNull() {
		proc.SetDisableLibraryRules(src.DisableLibraryRules.ValueBool())
	}
	if !src.Field.IsNull() && !src.Field.IsUnknown() {
		proc.SetField(src.Field.ValueString())
	}

	var rules []datadogV2.ObservabilityPipelineParseGrokProcessorRuleItem
	for _, r := range src.Rules {
		rule := &datadogV2.ObservabilityPipelineParseGrokProcessorRule{
			Source:       r.Source.ValueString(),
			MatchRules:   expandGrokMatchRules(r.MatchRules),
			SupportRules: expandGrokSupportRules(r.SupportRules),
		}
		rules = append(rules, datadogV2.ObservabilityPipelineParseGrokProcessorRuleAsObservabilityPipelineParseGrokProcessorRuleItem(rule))
	}
	for _, r := range src.IncludeRules {
		rule := &datadogV2.ObservabilityPipelineParseGrokProcessorIncludeRule{
			Include:      r.Include.ValueString(),
			MatchRules:   expandGrokMatchRules(r.MatchRules),
			SupportRules: expandGrokSupportRules(r.SupportRules),
		}
		rules = append(rules, datadogV2.ObservabilityPipelineParseGrokProcessorIncludeRuleAsObservabilityPipelineParseGrokProcessorRuleItem(rule))
	}
	proc.SetRules(rules)

	return datadogV2.ObservabilityPipelineParseGrokProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func FlattenParseGrokProcessor(src *datadogV2.ObservabilityPipelineParseGrokProcessor) *ParseGrokProcessorModel {
	if src == nil {
		return nil
	}

	grok := &ParseGrokProcessorModel{
		DisableLibraryRules: types.BoolValue(src.GetDisableLibraryRules()),
		Field:               types.StringValue(src.GetField()),
	}
	for _, item := range src.GetRules() {
		if r := item.ObservabilityPipelineParseGrokProcessorRule; r != nil {
			grok.Rules = append(grok.Rules, ParseGrokProcessorRuleModel{
				Source:       types.StringValue(r.GetSource()),
				MatchRules:   flattenGrokMatchRules(r.GetMatchRules()),
				SupportRules: flattenGrokSupportRules(r.GetSupportRules()),
			})
		} else if r := item.ObservabilityPipelineParseGrokProcessorIncludeRule; r != nil {
			grok.IncludeRules = append(grok.IncludeRules, ParseGrokProcessorIncludeRuleModel{
				Include:      types.StringValue(r.GetInclude()),
				MatchRules:   flattenGrokMatchRules(r.GetMatchRules()),
				SupportRules: flattenGrokSupportRules(r.GetSupportRules()),
			})
		}
	}
	return grok
}
