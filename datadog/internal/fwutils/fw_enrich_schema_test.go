package fwutils

import (
	"regexp"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	ephemeralSchema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

func TestEnrichSchemaSingleNestedAttribute(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		schema              schema.Schema
		expectedDescription string
	}
	testCases := map[string]testStruct{
		"description with string enum validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SingleNestedAttribute{
						Required: true,
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
		"description with int64 between validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.Int64Attribute{
								Required:    true,
								Description: "Example description.",
								Validators: []validator.Int64{
									int64validator.Between(1, 500),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 1 and 500.",
		},
		"description with bool default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Example description.",
								Default:     booldefault.StaticBool(true),
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Defaults to `true`.",
		},
		"description with enum validator and default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"test_attribute": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Example description.",
								Default:     stringdefault.StaticString("admins"),
								Validators: []validator.String{
									validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`. Defaults to `\"admins\"`.",
		},
		"description without validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SingleNestedAttribute{
						Required: true,
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
			description := testCase.schema.Attributes["nested_attribute"].(schema.SingleNestedAttribute).Attributes["test_attribute"].GetDescription()
			if description != testCase.expectedDescription {
				t.Errorf("expected description '%s', got '%s' instead.", testCase.expectedDescription, description)
			}
		})
	}
}

func TestEnrichSchemaListNestedAttribute(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		schema              schema.Schema
		expectedDescription string
	}
	testCases := map[string]testStruct{
		"description with string enum validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
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
		"description with int64 between validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.Int64Attribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.Int64{
										int64validator.Between(1, 500),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 1 and 500.",
		},
		"description with bool default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Example description.",
									Default:     booldefault.StaticBool(true),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Defaults to `true`.",
		},
		"description with enum validator and default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Example description.",
									Default:     stringdefault.StaticString("admins"),
									Validators: []validator.String{
										validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`. Defaults to `\"admins\"`.",
		},
		"description without validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
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
			description := testCase.schema.Attributes["nested_attribute"].(schema.ListNestedAttribute).NestedObject.Attributes["test_attribute"].GetDescription()
			if description != testCase.expectedDescription {
				t.Errorf("expected description '%s', got '%s' instead.", testCase.expectedDescription, description)
			}
		})
	}
}

func TestEnrichSchemaSetNestedAttribute(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		schema              schema.Schema
		expectedDescription string
	}
	testCases := map[string]testStruct{
		"description with string enum validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SetNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
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
		"description with int64 between validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SetNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.Int64Attribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.Int64{
										int64validator.Between(1, 500),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 1 and 500.",
		},
		"description with bool default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SetNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Example description.",
									Default:     booldefault.StaticBool(true),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Defaults to `true`.",
		},
		"description with enum validator and default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SetNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Example description.",
									Default:     stringdefault.StaticString("admins"),
									Validators: []validator.String{
										validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`. Defaults to `\"admins\"`.",
		},
		"description without validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.SetNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
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
			description := testCase.schema.Attributes["nested_attribute"].(schema.SetNestedAttribute).NestedObject.Attributes["test_attribute"].GetDescription()
			if description != testCase.expectedDescription {
				t.Errorf("expected description '%s', got '%s' instead.", testCase.expectedDescription, description)
			}
		})
	}
}

func TestEnrichSchemaMapNestedAttribute(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		schema              schema.Schema
		expectedDescription string
	}
	testCases := map[string]testStruct{
		"description with string enum validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.MapNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
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
		"description with int64 between validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.MapNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.Int64Attribute{
									Required:    true,
									Description: "Example description.",
									Validators: []validator.Int64{
										int64validator.Between(1, 500),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Value must be between 1 and 500.",
		},
		"description with bool default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.MapNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Example description.",
									Default:     booldefault.StaticBool(true),
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Defaults to `true`.",
		},
		"description with enum validator and default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.MapNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"test_attribute": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Example description.",
									Default:     stringdefault.StaticString("admins"),
									Validators: []validator.String{
										validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
									},
								},
							},
						},
					},
				},
			},
			expectedDescription: "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`. Defaults to `\"admins\"`.",
		},
		"description without validator": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"nested_attribute": schema.MapNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
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
			description := testCase.schema.Attributes["nested_attribute"].(schema.MapNestedAttribute).NestedObject.Attributes["test_attribute"].GetDescription()
			if description != testCase.expectedDescription {
				t.Errorf("expected description '%s', got '%s' instead.", testCase.expectedDescription, description)
			}
		})
	}
}

func TestEnrichDatasourceSchemaSingleNestedAttribute(t *testing.T) {
	t.Parallel()

	s := datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"nested_attribute": datasourceSchema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]datasourceSchema.Attribute{
					"test_attribute": datasourceSchema.StringAttribute{
						Required:    true,
						Description: "Example description.",
						Validators: []validator.String{
							validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
						},
					},
				},
			},
		},
	}

	EnrichFrameworkDatasourceSchema(&s)

	expected := "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`."
	description := s.Attributes["nested_attribute"].(datasourceSchema.SingleNestedAttribute).Attributes["test_attribute"].GetDescription()
	if description != expected {
		t.Errorf("expected description '%s', got '%s' instead.", expected, description)
	}
}

func TestEnrichDatasourceSchemaListNestedAttribute(t *testing.T) {
	t.Parallel()

	s := datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"nested_attribute": datasourceSchema.ListNestedAttribute{
				Required: true,
				NestedObject: datasourceSchema.NestedAttributeObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"test_attribute": datasourceSchema.StringAttribute{
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
	}

	EnrichFrameworkDatasourceSchema(&s)

	expected := "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`."
	description := s.Attributes["nested_attribute"].(datasourceSchema.ListNestedAttribute).NestedObject.Attributes["test_attribute"].GetDescription()
	if description != expected {
		t.Errorf("expected description '%s', got '%s' instead.", expected, description)
	}
}

func TestEnrichDatasourceSchemaSetNestedAttribute(t *testing.T) {
	t.Parallel()

	s := datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"nested_attribute": datasourceSchema.SetNestedAttribute{
				Required: true,
				NestedObject: datasourceSchema.NestedAttributeObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"test_attribute": datasourceSchema.StringAttribute{
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
	}

	EnrichFrameworkDatasourceSchema(&s)

	expected := "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`."
	description := s.Attributes["nested_attribute"].(datasourceSchema.SetNestedAttribute).NestedObject.Attributes["test_attribute"].GetDescription()
	if description != expected {
		t.Errorf("expected description '%s', got '%s' instead.", expected, description)
	}
}

func TestEnrichDatasourceSchemaMapNestedAttribute(t *testing.T) {
	t.Parallel()

	s := datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"nested_attribute": datasourceSchema.MapNestedAttribute{
				Required: true,
				NestedObject: datasourceSchema.NestedAttributeObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"test_attribute": datasourceSchema.StringAttribute{
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
	}

	EnrichFrameworkDatasourceSchema(&s)

	expected := "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`."
	description := s.Attributes["nested_attribute"].(datasourceSchema.MapNestedAttribute).NestedObject.Attributes["test_attribute"].GetDescription()
	if description != expected {
		t.Errorf("expected description '%s', got '%s' instead.", expected, description)
	}
}

func TestEnrichEphemeralSchemaSingleNestedAttribute(t *testing.T) {
	t.Parallel()

	s := ephemeralSchema.Schema{
		Attributes: map[string]ephemeralSchema.Attribute{
			"nested_attribute": ephemeralSchema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]ephemeralSchema.Attribute{
					"test_attribute": ephemeralSchema.StringAttribute{
						Required:    true,
						Description: "Example description.",
						Validators: []validator.String{
							validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
						},
					},
				},
			},
		},
	}

	EnrichFrameworkEphemeralResourceSchema(&s)

	expected := "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`."
	description := s.Attributes["nested_attribute"].(ephemeralSchema.SingleNestedAttribute).Attributes["test_attribute"].GetDescription()
	if description != expected {
		t.Errorf("expected description '%s', got '%s' instead.", expected, description)
	}
}

func TestEnrichEphemeralSchemaListNestedAttribute(t *testing.T) {
	t.Parallel()

	s := ephemeralSchema.Schema{
		Attributes: map[string]ephemeralSchema.Attribute{
			"nested_attribute": ephemeralSchema.ListNestedAttribute{
				Required: true,
				NestedObject: ephemeralSchema.NestedAttributeObject{
					Attributes: map[string]ephemeralSchema.Attribute{
						"test_attribute": ephemeralSchema.StringAttribute{
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
	}

	EnrichFrameworkEphemeralResourceSchema(&s)

	expected := "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`."
	description := s.Attributes["nested_attribute"].(ephemeralSchema.ListNestedAttribute).NestedObject.Attributes["test_attribute"].GetDescription()
	if description != expected {
		t.Errorf("expected description '%s', got '%s' instead.", expected, description)
	}
}

func TestEnrichEphemeralSchemaSetNestedAttribute(t *testing.T) {
	t.Parallel()

	s := ephemeralSchema.Schema{
		Attributes: map[string]ephemeralSchema.Attribute{
			"nested_attribute": ephemeralSchema.SetNestedAttribute{
				Required: true,
				NestedObject: ephemeralSchema.NestedAttributeObject{
					Attributes: map[string]ephemeralSchema.Attribute{
						"test_attribute": ephemeralSchema.StringAttribute{
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
	}

	EnrichFrameworkEphemeralResourceSchema(&s)

	expected := "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`."
	description := s.Attributes["nested_attribute"].(ephemeralSchema.SetNestedAttribute).NestedObject.Attributes["test_attribute"].GetDescription()
	if description != expected {
		t.Errorf("expected description '%s', got '%s' instead.", expected, description)
	}
}

func TestEnrichEphemeralSchemaMapNestedAttribute(t *testing.T) {
	t.Parallel()

	s := ephemeralSchema.Schema{
		Attributes: map[string]ephemeralSchema.Attribute{
			"nested_attribute": ephemeralSchema.MapNestedAttribute{
				Required: true,
				NestedObject: ephemeralSchema.NestedAttributeObject{
					Attributes: map[string]ephemeralSchema.Attribute{
						"test_attribute": ephemeralSchema.StringAttribute{
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
	}

	EnrichFrameworkEphemeralResourceSchema(&s)

	expected := "Example description. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`."
	description := s.Attributes["nested_attribute"].(ephemeralSchema.MapNestedAttribute).NestedObject.Attributes["test_attribute"].GetDescription()
	if description != expected {
		t.Errorf("expected description '%s', got '%s' instead.", expected, description)
	}
}

func TestEnrichSchemaNestedAttributeWithinNestedAttribute(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"action": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"levels": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"team": schema.StringAttribute{
									Required:    true,
									Description: "A team level.",
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
	}

	EnrichFrameworkResourceSchema(&s)

	expected := "A team level. Valid values are `admins`, `members`, `organization`, `user_access_manage`, `teams_manage`."
	description := s.Attributes["action"].(schema.SingleNestedAttribute).
		Attributes["levels"].(schema.ListNestedAttribute).
		NestedObject.Attributes["team"].GetDescription()
	if description != expected {
		t.Errorf("expected description '%s', got '%s' instead.", expected, got)
	}
}
