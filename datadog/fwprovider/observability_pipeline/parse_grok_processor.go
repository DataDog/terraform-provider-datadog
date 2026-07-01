package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ParseGrokProcessorModel struct {
	DisableLibraryRules types.Bool                    `tfsdk:"disable_library_rules"`
	Rules               []ParseGrokProcessorRuleModel `tfsdk:"rule"`
}

type ParseGrokProcessorRuleModel struct {
	Source       types.String    `tfsdk:"source"`
	MatchRules   []GrokRuleModel `tfsdk:"match_rule"`
	SupportRules []GrokRuleModel `tfsdk:"support_rule"`
}

type GrokRuleModel struct {
	Name types.String `tfsdk:"name"`
	Rule types.String `tfsdk:"rule"`
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
					Description: "If set to `true`, disables the default Grok rules provided by Datadog.",
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
							"match_rule": schema.ListNestedBlock{
								Description: "A list of Grok parsing rules that define how to extract fields from the source field. Each rule must contain a name and a valid Grok pattern.",
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
							},
							"support_rule": schema.ListNestedBlock{
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
							},
						},
					},
				},
			},
		},
	}
}

func ExpandParseGrokProcessor(common BaseProcessorFields, src *ParseGrokProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineParseGrokProcessorWithDefaults()
	common.ApplyTo(proc)

	if !src.DisableLibraryRules.IsNull() {
		proc.SetDisableLibraryRules(src.DisableLibraryRules.ValueBool())
	}

	var rules []datadogV2.ObservabilityPipelineParseGrokProcessorRule
	for _, r := range src.Rules {
		rule := datadogV2.ObservabilityPipelineParseGrokProcessorRule{
			Source: r.Source.ValueString(),
		}
		var matchRules []datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule
		for _, m := range r.MatchRules {
			matchRules = append(matchRules, datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule{
				Name: m.Name.ValueString(),
				Rule: m.Rule.ValueString(),
			})
		}
		rule.SetMatchRules(matchRules)
		var supportRules []datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule
		for _, s := range r.SupportRules {
			supportRules = append(supportRules, datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule{
				Name: s.Name.ValueString(),
				Rule: s.Rule.ValueString(),
			})
		}
		rule.SetSupportRules(supportRules)
		rules = append(rules, rule)
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
	}
	for _, rule := range src.GetRules() {
		r := ParseGrokProcessorRuleModel{
			Source: types.StringValue(rule.GetSource()),
		}
		for _, m := range rule.GetMatchRules() {
			r.MatchRules = append(r.MatchRules, GrokRuleModel{
				Name: types.StringValue(m.GetName()),
				Rule: types.StringValue(m.GetRule()),
			})
		}
		for _, s := range rule.GetSupportRules() {
			r.SupportRules = append(r.SupportRules, GrokRuleModel{
				Name: types.StringValue(s.GetName()),
				Rule: types.StringValue(s.GetRule()),
			})
		}
		grok.Rules = append(grok.Rules, r)
	}
	return grok
}
