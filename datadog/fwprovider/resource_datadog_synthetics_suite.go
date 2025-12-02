package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &syntheticsSuiteResource{}
	_ resource.ResourceWithImportState = &syntheticsSuiteResource{}
)

func NewSyntheticsSuiteResource() resource.Resource {
	return &syntheticsSuiteResource{}
}

type SyntheticsSuiteModel struct {
	PublicID types.String `tfsdk:"public_id"`
	Name     types.String `tfsdk:"name"`
	Message  types.String `tfsdk:"message"`
	Tags     types.List   `tfsdk:"tags"`
	Options  types.List   `tfsdk:"options"`
	Tests    types.List   `tfsdk:"tests"`
}

type SyntheticsSuiteOptionsModel struct {
	AlertingThreshold types.Float64 `tfsdk:"alerting_threshold"`
}

type SyntheticsSuiteTestModel struct {
	PublicID            types.String `tfsdk:"public_id"`
	AlertingCriticality types.String `tfsdk:"alerting_criticality"`
}

var syntheticsSuiteOptionsAttrType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"alerting_threshold": types.Float64Type,
	},
}

var syntheticsSuiteTestAttrType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"public_id":            types.StringType,
		"alerting_criticality": types.StringType,
	},
}

type syntheticsSuiteResource struct {
	Api  *datadogV2.SyntheticsApi
	Auth context.Context
}

func (r *syntheticsSuiteResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSyntheticsApiV2()
	r.Auth = providerData.Auth
}

func (r *syntheticsSuiteResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "synthetics_suite"
}

func (r *syntheticsSuiteResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Synthetics Suite resource. This can be used to create and manage Synthetics test suites.",
		Attributes: map[string]schema.Attribute{
			"public_id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of the Synthetics suite.",
				Required:    true,
			},
			"message": schema.StringAttribute{
				Description: "Message of the Synthetics suite.",
				Optional:    true,
			},
			"tags": schema.ListAttribute{
				Description: "A list of tags to associate with your synthetics suite.",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"options": schema.ListNestedBlock{
				Description: "Options for the Synthetics suite.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alerting_threshold": schema.Float64Attribute{
							Description: "Alerting threshold for the suite. Must be between 0 and 1.",
							Required:    true,
							Validators: []validator.Float64{
								float64validator.Between(0, 1),
							},
						},
					},
				},
			},
			"tests": schema.ListNestedBlock{
				Description: "List of tests in the Synthetics suite. Can be empty but the field is always sent to the API.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"public_id": schema.StringAttribute{
							Description: "Public ID of the test.",
							Required:    true,
						},
						"alerting_criticality": schema.StringAttribute{
							Description: "Alerting criticality for the test. Valid values are `ignore`, `critical`.",
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("ignore", "critical"),
							},
						},
					},
				},
			},
		},
	}
}

func (r *syntheticsSuiteResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state SyntheticsSuiteModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildSyntheticsSuiteBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResponse, err := r.Api.CreateSyntheticsSuite(r.Auth, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error creating synthetics suite: %v", httpResponse)))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsSuiteResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state SyntheticsSuiteModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	publicID := state.PublicID.ValueString()
	resp, httpResponse, err := r.Api.GetSyntheticsSuite(r.Auth, publicID)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reading synthetics suite: %v", httpResponse)))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsSuiteResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state SyntheticsSuiteModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	publicID := state.PublicID.ValueString()
	body, diags := r.buildSyntheticsSuiteBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResponse, err := r.Api.EditSyntheticsSuite(r.Auth, publicID, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error updating synthetics suite: %v", httpResponse)))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsSuiteResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state SyntheticsSuiteModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	publicID := state.PublicID.ValueString()

	// Build delete request with suite public_id
	deleteAttrs := datadogV2.NewDeletedSuitesRequestDeleteAttributes([]string{publicID})
	deleteData := datadogV2.NewDeletedSuitesRequestDelete(*deleteAttrs)
	deleteBody := datadogV2.NewDeletedSuitesRequestDeleteRequest(*deleteData)

	_, httpResponse, err := r.Api.DeleteSyntheticsSuites(r.Auth, *deleteBody)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error deleting synthetics suite: %v", httpResponse)))
		return
	}
}

func (r *syntheticsSuiteResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("public_id"), request, response)
}

// buildSyntheticsSuiteBody builds the API request body from the Terraform state
func (r *syntheticsSuiteResource) buildSyntheticsSuiteBody(ctx context.Context, state *SyntheticsSuiteModel) (datadogV2.SuiteCreateEditRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	suite := datadogV2.NewSyntheticsSuiteWithDefaults()

	if !state.Name.IsNull() {
		suite.SetName(state.Name.ValueString())
	}
	if !state.Message.IsNull() {
		suite.SetMessage(state.Message.ValueString())
	}

	// Build options
	options := datadogV2.NewSyntheticsSuiteOptions()
	if !state.Options.IsNull() && !state.Options.IsUnknown() {
		var optionsList []SyntheticsSuiteOptionsModel
		diags.Append(state.Options.ElementsAs(ctx, &optionsList, false)...)
		if diags.HasError() {
			return datadogV2.SuiteCreateEditRequest{}, diags
		}

		if len(optionsList) > 0 {
			alerting := datadogV2.NewSyntheticsSuiteOptionsAlerting()
			threshold := optionsList[0].AlertingThreshold.ValueFloat64()
			alerting.SetThreshold(threshold)
			options.SetAlerting(*alerting)
		}
	}
	suite.SetOptions(*options)

	// Build tests - tests is required but can be empty
	var testsList []SyntheticsSuiteTestModel
	if !state.Tests.IsNull() && !state.Tests.IsUnknown() {
		diags.Append(state.Tests.ElementsAs(ctx, &testsList, false)...)
		if diags.HasError() {
			return datadogV2.SuiteCreateEditRequest{}, diags
		}
	}

	suiteTests := make([]datadogV2.SyntheticsSuiteTest, len(testsList))
	for i, test := range testsList {
		suiteTest := datadogV2.NewSyntheticsSuiteTest(test.PublicID.ValueString())
		if !test.AlertingCriticality.IsNull() && !test.AlertingCriticality.IsUnknown() {
			criticality := datadogV2.SyntheticsSuiteTestAlertingCriticality(test.AlertingCriticality.ValueString())
			suiteTest.SetAlertingCriticality(criticality)
		}
		suiteTests[i] = *suiteTest
	}
	suite.SetTests(suiteTests)

	// Build tags
	if !state.Tags.IsNull() && !state.Tags.IsUnknown() {
		var tags []string
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return datadogV2.SuiteCreateEditRequest{}, diags
		}
		suite.SetTags(tags)
	}

	// Build request
	createEdit := datadogV2.NewSuiteCreateEdit(*suite, datadogV2.SYNTHETICSSUITETYPE_SUITE)
	body := datadogV2.NewSuiteCreateEditRequest(*createEdit)

	return *body, diags
}

// updateState updates the Terraform state from the API response
func (r *syntheticsSuiteResource) updateState(ctx context.Context, state *SyntheticsSuiteModel, resp *datadogV2.SyntheticsSuiteResponse, diags *diag.Diagnostics) {
	if data, ok := resp.GetDataOk(); ok {
		if attrs, ok := data.GetAttributesOk(); ok {
			// Update name
			if name, ok := attrs.GetNameOk(); ok {
				state.Name = types.StringValue(*name)
			}

			// Update message
			if message, ok := attrs.GetMessageOk(); ok {
				state.Message = types.StringValue(*message)
			}

			// Update options
			if opts, ok := attrs.GetOptionsOk(); ok {
				if alerting, ok := opts.GetAlertingOk(); ok {
					if threshold, ok := alerting.GetThresholdOk(); ok {
						optionsModel := SyntheticsSuiteOptionsModel{
							AlertingThreshold: types.Float64Value(*threshold),
						}
						optionsList, d := types.ListValueFrom(ctx, syntheticsSuiteOptionsAttrType, []SyntheticsSuiteOptionsModel{optionsModel})
						diags.Append(d...)
						state.Options = optionsList
					}
				}
			}

			// Update tests - always set, even if empty
			var testModels []SyntheticsSuiteTestModel
			if tests, ok := attrs.GetTestsOk(); ok && len(*tests) > 0 {
				testModels = make([]SyntheticsSuiteTestModel, len(*tests))
				for i, test := range *tests {
					testModel := SyntheticsSuiteTestModel{
						PublicID: types.StringValue(test.GetPublicId()),
					}
					if criticality, ok := test.GetAlertingCriticalityOk(); ok {
						testModel.AlertingCriticality = types.StringValue(string(*criticality))
					} else {
						testModel.AlertingCriticality = types.StringNull()
					}
					testModels[i] = testModel
				}
			} else {
				testModels = []SyntheticsSuiteTestModel{}
			}
			testsList, d := types.ListValueFrom(ctx, syntheticsSuiteTestAttrType, testModels)
			diags.Append(d...)
			state.Tests = testsList

			// Update tags
			if tags, ok := attrs.GetTagsOk(); ok {
				tagsList, d := types.ListValueFrom(ctx, types.StringType, *tags)
				diags.Append(d...)
				state.Tags = tagsList
			} else {
				state.Tags = types.ListNull(types.StringType)
			}

			// Handle public_id - Note: The API response might not include the public_id directly
			// in the attributes. We need to check if it's in the AdditionalProperties or
			// use the publicId parameter that was passed to the API call
			// For now, we'll preserve the existing public_id from state if it's already set
			// On Create, we'll need to extract it from the response somehow
			if state.PublicID.IsNull() || state.PublicID.IsUnknown() {
				// Try to get from AdditionalProperties if available
				if data.AdditionalProperties != nil {
					if publicID, ok := data.AdditionalProperties["id"].(string); ok {
						state.PublicID = types.StringValue(publicID)
					} else if publicID, ok := data.AdditionalProperties["public_id"].(string); ok {
						state.PublicID = types.StringValue(publicID)
					}
				}
			}
		}
	}
}
