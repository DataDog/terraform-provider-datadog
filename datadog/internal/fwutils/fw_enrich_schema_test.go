package fwutils

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
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
			updatedSchema := EnrichFrameworkResourceSchema(testCase.schema)
			description := updatedSchema.Attributes["test_attribute"].GetDescription()
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
			updatedSchema := EnrichFrameworkResourceSchema(testCase.schema)
			description := updatedSchema.Blocks["nested_block"].GetNestedObject().GetAttributes()["test_attribute"].GetDescription()
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
			updatedSchema := EnrichFrameworkResourceSchema(testCase.schema)
			description := updatedSchema.Blocks["nested_block"].GetNestedObject().GetAttributes()["test_attribute"].GetDescription()
			if description != testCase.expectedDescription {
				t.Errorf("expected description '%s', got '%s' instead.", testCase.expectedDescription, description)
			}
		})
	}
}
