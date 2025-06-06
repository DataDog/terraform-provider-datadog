package fwutils

import (
	"regexp"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

func TestEnrichSchemaAttributes(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		schema              schema.Schema
		expectedDescription string
	}
	testCases := map[string]testStruct{
		"description with string enum validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.StringAttribute{
						Description: "Example description.",
						Required:    true,
						Validators: []validator.String{
							validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`.",
		},
		"description with string oneOf validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.StringAttribute{
						Required:    true,
						Description: "Example description.",
						Validators: []validator.String{
							stringvalidator.OneOf("asc", "desc"),
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `asc`, `desc`.",
		},
		"description with string regexMatches with message validator": {
			schema: schema.Schema{

				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.StringAttribute{
						Required:    true,
						Description: "Example description.",
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
						},
					},
				},
			},
			expectedDescription: "Example description. Must be all uppercase with underscores.",
		},
		"description with string regexMatches without message validator": {
			schema: schema.Schema{

				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.StringAttribute{
						Required:    true,
						Description: "Example description.",
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), ""),
						},
					},
				},
			},
			expectedDescription: "Example description. Value must match regular expression '^[A-Z][A-Z0-9_]+[A-Z0-9]$'.",
		},
		"description with string entity YAML validator": {
			schema: schema.Schema{

				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.StringAttribute{
						Required:    true,
						Description: "Example description.",
						Validators:  []validator.String{validators.ValidEntityYAMLValidator()},
					},
				},
			},
			expectedDescription: "Example description. Entity must be a valid entity YAML/JSON structure.",
		},
		"description with string CIDR IP validator": {
			schema: schema.Schema{

				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.StringAttribute{
						Required:    true,
						Description: "Example description.",
						Validators:  []validator.String{validators.CidrIpValidator()},
					},
				},
			},
			expectedDescription: "Example description. String must be a valid CIDR block or IP address.",
		},
		"description with string lengthAtLeast validator": {
			schema: schema.Schema{

				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.StringAttribute{
						Required:    true,
						Description: "Example description.",
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
				},
			},
			expectedDescription: "Example description. String length must be at least 1.",
		},
		"description with string betweenValidator|float validator": {
			schema: schema.Schema{

				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.StringAttribute{
						Required:    true,
						Description: "Example description.",
						Validators: []validator.String{
							validators.Float64Between(0, 1),
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 0.00 and 1.00.",
		},

		// Int validators
		"description with int64 enum validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.Int64Attribute{
						Description: "Example description.",
						Required:    true,
						Validators: []validator.Int64{
							validators.NewEnumValidator[validator.Int64](datadogV1.NewSyntheticsPlayingTabFromValue),
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `-1`, `0`, `1`, `2`, `3`.",
		},
		"description with int64 betweenValidator|int validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.Int64Attribute{
						Required:    true,
						Description: "Example description.",
						Validators: []validator.Int64{
							int64validator.Between(4, 10),
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 4 and 10.",
		},
		"description with int64 atLeast validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.Int64Attribute{
						Required:    true,
						Description: "Example description.",
						Validators: []validator.Int64{
							int64validator.AtLeast(1),
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be at least 1.",
		},
		"description without validators": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test_attribute": schema.StringAttribute{
						Description: "Example description.",
						Required:    true,
					},
				},
			},
			expectedDescription: "Example description.",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			EnrichFrameworkResourceSchema(&testCase.schema)
			description := testCase.schema.Attributes["test_attribute"].GetDescription()
			if description != testCase.expectedDescription {
				t.Errorf("expected description '%s', got '%s' instead.", testCase.expectedDescription, description)
			}
		})
	}
}

func TestEnrichSchemaListNestedBlock(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		schema              schema.Schema
		expectedDescription string
	}
	testCases := map[string]testStruct{
		// String validators
		"description with string enum validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.String{
										validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`.",
		},
		"description with string oneOf validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.String{
										stringvalidator.OneOf("asc", "desc"),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `asc`, `desc`.",
		},
		"description with string regexMatches with message validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.String{
										stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Must be all uppercase with underscores.",
		},
		"description with string regexMatches without message validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.String{
										stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), ""),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must match regular expression '^[A-Z][A-Z0-9_]+[A-Z0-9]$'.",
		},
		"description with string entity YAML validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
									Validators:  []validator.String{validators.ValidEntityYAMLValidator()},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Entity must be a valid entity YAML/JSON structure.",
		},
		"description with string CIDR IP validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
									Validators:  []validator.String{validators.CidrIpValidator()},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. String must be a valid CIDR block or IP address.",
		},
		"description with string lengthAtLeast validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. String length must be at least 1.",
		},
		"description with string betweenValidator|float validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.String{
										validators.Float64Between(0, 1),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 0.00 and 1.00.",
		},

		// Int validators
		"description with int64 betweenValidator|int validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.Int64Attribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.Int64{
										int64validator.Between(4, 10),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 4 and 10.",
		},
		"description with int64 atLeast validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.Int64Attribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.Int64{
										int64validator.AtLeast(1),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be at least 1.",
		},
		"description without validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description.",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			EnrichFrameworkResourceSchema(&testCase.schema)
			description := testCase.schema.Blocks["nested_block"].GetNestedObject().GetAttributes()["test_attribute"].GetDescription()
			if description != testCase.expectedDescription {
				t.Errorf("expected description '%s', got '%s' instead.", testCase.expectedDescription, description)
			}
		})
	}
}

func TestNestedNestedBlock(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		schema              schema.Schema
		expectedDescription string
	}
	testCases := map[string]testStruct{
		// String validators
		"description with string enum validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.StringAttribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.String{
												validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`.",
		},
		"description with string oneOf validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.StringAttribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.String{
												stringvalidator.OneOf("asc", "desc"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. Valid values are `asc`, `desc`.",
		},
		"description with string regexMatches with message validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.StringAttribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.String{
												stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. Must be all uppercase with underscores.",
		},
		"description with string regexMatches without message validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.StringAttribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.String{
												stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), ""),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. Value must match regular expression '^[A-Z][A-Z0-9_]+[A-Z0-9]$'.",
		},
		"description with string entity YAML validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.StringAttribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.String{
												validators.ValidEntityYAMLValidator(),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. Entity must be a valid entity YAML/JSON structure.",
		},
		"description with string CIDR IP validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.StringAttribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.String{
												validators.CidrIpValidator(),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. String must be a valid CIDR block or IP address.",
		},
		"description with string lengthAtLeast validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.StringAttribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.String{
												stringvalidator.LengthAtLeast(1),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. String length must be at least 1.",
		},
		"description with string betweenValidator|float validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.StringAttribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.String{
												validators.Float64Between(0, 1),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. Value must be between 0.00 and 1.00.",
		},

		// Int validators
		"description with int64 betweenValidator|int validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.Int64Attribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.Int64Attribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.Int64{
												int64validator.Between(4, 10),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. Value must be between 4 and 10.",
		},
		"description with int64 atLeast validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.Int64Attribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.Int64Attribute{
											Required:    true,
											Description: "Nested test attribute.",
											Validators: []validator.Int64{
												int64validator.AtLeast(1),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute. Value must be at least 1.",
		},
		"description without validator in nested nested block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Required:    true,
									Description: "Example description.",
								},
							},
							Blocks: map[string]schema.Block{
								"block": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"test_attribute": schema.StringAttribute{
											Required:    true,
											Description: "Nested test attribute.",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Nested test attribute.",
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			EnrichFrameworkResourceSchema(&testCase.schema)
			description := testCase.schema.Blocks["nested_block"].GetNestedObject().GetBlocks()["block"].GetNestedObject().GetAttributes()["test_attribute"].GetDescription()
			if description != testCase.expectedDescription {
				t.Errorf("expected description '%s', got '%s' instead.", testCase.expectedDescription, description)
			}
		})
	}
}

func TestEnrichSchemaSingleNestedBlock(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		schema              schema.Schema
		expectedDescription string
	}
	testCases := map[string]testStruct{
		"description with string enum validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required:    true,
								Description: "Example description.",
								Validators: []validator.String{
									validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`.",
		},
		"description with string oneOf validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required:    true,
								Description: "Example description.",
								Validators: []validator.String{
									stringvalidator.OneOf("asc", "desc"),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `asc`, `desc`.",
		},
		"description with string regexMatches with message validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required:    true,
								Description: "Example description.",
								Validators: []validator.String{
									stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Must be all uppercase with underscores.",
		},
		"description with string regexMatches without message validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required:    true,
								Description: "Example description.",
								Validators: []validator.String{
									stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), ""),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must match regular expression '^[A-Z][A-Z0-9_]+[A-Z0-9]$'.",
		},
		"description with string entity YAML validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required:    true,
								Description: "Example description.",
								Validators:  []validator.String{validators.ValidEntityYAMLValidator()},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Entity must be a valid entity YAML/JSON structure.",
		},
		"description with string CIDR IP validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required:    true,
								Description: "Example description.",
								Validators:  []validator.String{validators.CidrIpValidator()},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. String must be a valid CIDR block or IP address.",
		},
		"description with string lengthAtLeast validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required:    true,
								Description: "Example description.",
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(1),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. String length must be at least 1.",
		},
		"description with string betweenValidator|float validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required:    true,
								Description: "Example description.",
								Validators: []validator.String{
									validators.Float64Between(0, 1),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 0.00 and 1.00.",
		},

		// Int validators
		"description with int64 betweenValidator|int validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.Int64Attribute{
								Required:    true,
								Description: "Example description.",
								Validators: []validator.Int64{
									int64validator.Between(4, 10),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 4 and 10.",
		},
		"description with int64 atLeast validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.Int64Attribute{
								Required:    true,
								Description: "Example description.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be at least 1.",
		},
		"description without validator": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Required:    true,
								Description: "Example description.",
							},
						},
					},
				},
			},
			expectedDescription: "Example description.",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			EnrichFrameworkResourceSchema(&testCase.schema)
			description := testCase.schema.Blocks["nested_block"].GetNestedObject().GetAttributes()["test_attribute"].GetDescription()
			if description != testCase.expectedDescription {
				t.Errorf("expected description '%s', got '%s' instead.", testCase.expectedDescription, description)
			}
		})
	}
}
