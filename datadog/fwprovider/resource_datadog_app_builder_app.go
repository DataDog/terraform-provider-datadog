package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &appBuilderAppResource{}
	_ resource.ResourceWithImportState = &appBuilderAppResource{}
	_ resource.ResourceWithModifyPlan  = &appBuilderAppResource{}
)

const ErrDeploymentExists = "this version of the app has already been published"
const ErrAlreadyUnpublished = "this version of the app has already been unpublished"

type appBuilderAppResource struct {
	Api  *datadogV2.AppBuilderApi
	Auth context.Context
}

// try single property JSON input -> validation will be handled on the API side
type appBuilderAppResourceModel struct {
	ID                              types.String `tfsdk:"id"`
	AppJson                         types.String `tfsdk:"app_json"`
	ActionQueryNamesToConnectionIDs types.Map    `tfsdk:"action_query_names_to_connection_ids"`
	Name                            types.String `tfsdk:"name"`
	Description                     types.String `tfsdk:"description"`
	RootInstanceName                types.String `tfsdk:"root_instance_name"`
	Published                       types.Bool   `tfsdk:"published"`
}

func NewAppBuilderAppResource() resource.Resource {
	return &appBuilderAppResource{}
}

func (r *appBuilderAppResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAppBuilderApiV2()
	r.Auth = providerData.Auth
}

func (r *appBuilderAppResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "app_builder_app"
}

func (r *appBuilderAppResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog App resource for creating and managing Datadog Apps from App Builder using the JSON definition. To easily export an App for use with Terraform, use the export button in the Datadog App Builder UI. This resource requires a [registered application key](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/app_key_registration).",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"app_json": schema.StringAttribute{
				Required:    true,
				Description: "The JSON representation of the App.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"action_query_names_to_connection_ids": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "If specified, this will override the Action Connection IDs for the specified Action Query Names in the App JSON. Otherwise, a map of the App's Action Query Names to Action Connection IDs will be returned in output.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If specified, this will override the name of the App in the App JSON.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If specified, this will override the human-readable description of the App in the App JSON.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"root_instance_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the root component of the app. This must be a grid component that contains all other components. If specified, this will override the root instance name of the App in the App JSON.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"published": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Set the app to published or unpublished. Published apps are available to other users. To ensure the app is accessible to the correct users, you also need to set a [Restriction Policy](https://docs.datadoghq.com/api/latest/restriction-policies/) on the app if a policy does not yet exist.",
			},
		},
	}
}

func (r *appBuilderAppResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *appBuilderAppResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan appBuilderAppResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	createReq, err := appBuilderAppModelToCreateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to create app request",
			err.Error(),
		)
		return
	}

	// create the app
	createResp, httpResp, err := r.Api.CreateApp(r.Auth, *createReq)
	if err != nil {
		if httpResp != nil {
			body, _ := io.ReadAll(httpResp.Body)
			response.Diagnostics.AddError("error creating app", string(body))
		} else {
			response.Diagnostics.AddError("error creating app", err.Error())
		}
		return
	}

	appID := createResp.Data.GetId()

	if ok := handleAppBuilderPublishState(r.Auth, &response.Diagnostics, r.Api, &plan, appID); !ok {
		return
	}

	// set computed values
	plan.ID = types.StringValue(appID.String())
	attributes := createReq.GetData().Attributes

	// if optional fields are not set, compute them from the app json
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		plan.Name = types.StringValue(*attributes.Name)
	}
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue(*attributes.Description)
	}
	if plan.RootInstanceName.IsNull() || plan.RootInstanceName.IsUnknown() {
		plan.RootInstanceName = types.StringValue(*attributes.RootInstanceName)
	}
	if plan.ActionQueryNamesToConnectionIDs.IsNull() || plan.ActionQueryNamesToConnectionIDs.IsUnknown() {
		plan.ActionQueryNamesToConnectionIDs, err = buildActionQueryNamesToConnectionIDsMap(attributes.GetQueries())
		if err != nil {
			response.Diagnostics.AddError("error building action_query_names_to_connection_ids map", err.Error())
			return
		}
	}

	// Save data into Terraform state
	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *appBuilderAppResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state appBuilderAppResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("error parsing ID as UUID", err.Error())
		return
	}

	// read the app
	appBuilderAppModel, err := readAppBuilderApp(r.Auth, r.Api, id)
	if err != nil {
		response.Diagnostics.AddError("error reading app", err.Error())
		return
	}

	// Save data into Terraform state
	diags = response.State.Set(ctx, appBuilderAppModel)
	response.Diagnostics.Append(diags...)
}

func (r *appBuilderAppResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state appBuilderAppResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	diags = request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	updateReq, err := appBuilderAppModelToUpdateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to create app update request",
			err.Error(),
		)
		return
	}

	appID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to parse app ID",
			err.Error(),
		)
		return
	}

	updateResp, httpResp, err := r.Api.UpdateApp(r.Auth, appID, *updateReq)
	if err != nil {
		if httpResp != nil {
			body, _ := io.ReadAll(httpResp.Body)
			response.Diagnostics.AddError("error updating app", string(body))
		} else {
			response.Diagnostics.AddError("error updating app", err.Error())
		}
		return
	}

	if ok := handleAppBuilderPublishState(r.Auth, &response.Diagnostics, r.Api, &plan, appID); !ok {
		return
	}

	// set computed values
	attributes := updateResp.GetData().Attributes

	// if optional fields are not set, compute them from the app json
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		plan.Name = types.StringValue(*attributes.Name)
	}
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue(*attributes.Description)
	}
	if plan.RootInstanceName.IsNull() || plan.RootInstanceName.IsUnknown() {
		plan.RootInstanceName = types.StringValue(*attributes.RootInstanceName)
	}
	if plan.ActionQueryNamesToConnectionIDs.IsNull() || plan.ActionQueryNamesToConnectionIDs.IsUnknown() {
		plan.ActionQueryNamesToConnectionIDs, err = buildActionQueryNamesToConnectionIDsMap(attributes.GetQueries())
		if err != nil {
			response.Diagnostics.AddError("error building action_query_names_to_connection_ids map", err.Error())
			return
		}
	}

	// Save data into Terraform state
	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *appBuilderAppResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state appBuilderAppResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("error parsing ID as UUID", err.Error())
		return
	}

	// delete the app
	_, httpResp, err := r.Api.DeleteApp(r.Auth, id)
	if err != nil {
		if httpResp != nil {
			// error body may have useful info for the user
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("error reading error response", err.Error())
				return
			}
			response.Diagnostics.AddError("error deleting app", string(body))
		} else {
			response.Diagnostics.AddError("error deleting app", err.Error())
		}
		return
	}
}

func (r *appBuilderAppResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If the plan is null (resource is being destroyed) or no state exists yet, return early
	// as there's nothing to modify
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var plan, config, state appBuilderAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If app_json isn't set in either plan or config, nothing to do
	if plan.AppJson.IsNull() || config.AppJson.IsNull() {
		return
	}

	// Unmarshal plan and state JSONs to compare their structures
	var planJSON, stateJSON map[string]any
	if err := json.Unmarshal([]byte(plan.AppJson.ValueString()), &planJSON); err != nil {
		resp.Diagnostics.AddError("Error unmarshaling plan JSON", err.Error())
		return
	}
	if err := json.Unmarshal([]byte(state.AppJson.ValueString()), &stateJSON); err != nil {
		resp.Diagnostics.AddError("Error unmarshaling state JSON", err.Error())
		return
	}

	// Create copies of the JSON structures to modify without affecting originals
	planCopy := make(map[string]any)
	stateCopy := make(map[string]any)
	maps.Copy(planCopy, planJSON)
	maps.Copy(stateCopy, stateJSON)

	// Remove fields that match the state values - these don't need to trigger updates
	if !plan.Name.IsNull() && state.Name.ValueString() == stateCopy["name"] {
		delete(planCopy, "name")
		delete(stateCopy, "name")
	}
	if !plan.Description.IsNull() && state.Description.ValueString() == stateCopy["description"] {
		delete(planCopy, "description")
		delete(stateCopy, "description")
	}
	if !plan.RootInstanceName.IsNull() && state.RootInstanceName.ValueString() == stateCopy["rootInstanceName"] {
		delete(planCopy, "rootInstanceName")
		delete(stateCopy, "rootInstanceName")
	}

	// Handle connection ID overrides
	if !plan.ActionQueryNamesToConnectionIDs.IsNull() {
		overrides := plan.ActionQueryNamesToConnectionIDs.Elements()

		// Process queries in all JSONs to remove connectionId fields that will be overridden
		processQueries := func(queries []any, isState bool) {
			for _, q := range queries {
				if query, ok := q.(map[string]any); ok {
					if queryName, ok := query["name"].(string); ok {
						if _, exists := overrides[queryName]; exists {
							if properties, ok := query["properties"].(map[string]any); ok {
								if spec, ok := properties["spec"].(map[string]any); ok {
									delete(spec, "connectionId")
								}
							}
						}
					}
					// Remove empty events array as it's not meaningful for comparison
					if _, hasEvents := query["events"]; hasEvents {
						delete(query, "events")
					}
				}
			}
		}

		if queries, ok := planCopy["queries"].([]any); ok {
			processQueries(queries, false)
		}
		if queries, ok := stateCopy["queries"].([]any); ok {
			processQueries(queries, true)
		}
	}

	// Clean up any remaining empty values that might have been created during processing
	utils.RemoveEmptyValuesInMap(planCopy)
	utils.RemoveEmptyValuesInMap(stateCopy)

	// Marshal back to JSON for final comparison
	planBytes, _ := json.Marshal(planCopy)
	stateBytes, _ := json.Marshal(stateCopy)

	planAndStateEq, _ := utils.AppJSONStringSemanticEquals(string(planBytes), string(stateBytes))

	// if the plan and state are equal, set the plan to the state
	if planAndStateEq {
		plan.AppJson = state.AppJson

		// Keep existing computed values logic
		if !state.ActionQueryNamesToConnectionIDs.IsNull() && config.ActionQueryNamesToConnectionIDs.IsNull() {
			plan.ActionQueryNamesToConnectionIDs = state.ActionQueryNamesToConnectionIDs
		}
		if !state.Name.IsNull() && config.Name.IsNull() {
			plan.Name = state.Name
		}
		if !state.Description.IsNull() && config.Description.IsNull() {
			plan.Description = state.Description
		}
		if !state.RootInstanceName.IsNull() && config.RootInstanceName.IsNull() {
			plan.RootInstanceName = state.RootInstanceName
		}
		if !state.Published.IsNull() && config.Published.IsNull() {
			plan.Published = state.Published
		}

		// Keep existing action_query_names_to_connection_ids map logic
		if !config.ActionQueryNamesToConnectionIDs.IsNull() && !state.ActionQueryNamesToConnectionIDs.IsNull() {
			plan.ActionQueryNamesToConnectionIDs = state.ActionQueryNamesToConnectionIDs
		}

		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}
}

func appBuilderAppModelToCreateApiRequest(plan appBuilderAppResourceModel) (*datadogV2.CreateAppRequest, error) {
	// unmarshal app JSON into a map
	var appJson map[string]any
	if err := json.Unmarshal([]byte(plan.AppJson.ValueString()), &appJson); err != nil {
		return nil, fmt.Errorf("error unmarshalling app JSON: %s", err)
	}

	// apply plan overrides to the app JSON map
	err := overrideAppBuilderAppAttributesInCreateRequestAttributes(plan, appJson)
	if err != nil {
		err = fmt.Errorf("error overriding app JSON attributes: %s", err.Error())
		return nil, err
	}

	// marshal the modified map into attributes
	marshalledBytes, err := json.Marshal(appJson)
	if err != nil {
		return nil, fmt.Errorf("error marshalling modified app JSON: %s", err)
	}

	attributes := datadogV2.NewCreateAppRequestDataAttributesWithDefaults()
	if err := json.Unmarshal(marshalledBytes, attributes); err != nil {
		return nil, fmt.Errorf("error unmarshalling to attributes: %s", err)
	}

	req := datadogV2.NewCreateAppRequestWithDefaults()
	req.Data = datadogV2.NewCreateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, nil
}

func appBuilderAppModelToUpdateApiRequest(plan appBuilderAppResourceModel) (*datadogV2.UpdateAppRequest, error) {
	// unmarshal app JSON into a map
	var appJson map[string]any
	if err := json.Unmarshal([]byte(plan.AppJson.ValueString()), &appJson); err != nil {
		return nil, fmt.Errorf("error unmarshalling app JSON: %s", err)
	}

	// apply plan overrides to the app JSON map
	err := overrideAppBuilderAppAttributesInUpdateRequestAttributes(plan, appJson)
	if err != nil {
		err = fmt.Errorf("error overriding app JSON attributes: %s", err.Error())
		return nil, err
	}

	// marshal the modified map into attributes
	marshalledBytes, err := json.Marshal(appJson)
	if err != nil {
		return nil, fmt.Errorf("error marshalling modified app JSON: %s", err)
	}
	attributes := datadogV2.NewUpdateAppRequestDataAttributesWithDefaults()
	if err := json.Unmarshal(marshalledBytes, attributes); err != nil {
		return nil, fmt.Errorf("error unmarshalling to attributes: %s", err)
	}

	req := datadogV2.NewUpdateAppRequestWithDefaults()
	req.Data = datadogV2.NewUpdateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, nil
}

func overrideAppBuilderAppAttributesInCreateRequestAttributes(plan appBuilderAppResourceModel, appJsonMap map[string]any) error {
	// name, description, root_instance_name are straightforward string replacements
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		appJsonMap["name"] = plan.Name.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		appJsonMap["description"] = plan.Description.ValueString()
	}
	if !plan.RootInstanceName.IsNull() && !plan.RootInstanceName.IsUnknown() {
		appJsonMap["rootInstanceName"] = plan.RootInstanceName.ValueString()
	}

	// Using action_query_names_to_connection_ids, replace connection ids in the update request attributes with the ones provided in the plan
	err := replaceConnectionIDsInActionQueries(plan.ActionQueryNamesToConnectionIDs, appJsonMap)
	if err != nil {
		err = fmt.Errorf("error replacing connection IDs in queries: %s", err.Error())
		return err
	}

	return nil
}

func overrideAppBuilderAppAttributesInUpdateRequestAttributes(plan appBuilderAppResourceModel, appJsonMap map[string]any) error {
	// name, description, root_instance_name are straightforward string replacements
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		appJsonMap["name"] = plan.Name.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		appJsonMap["description"] = plan.Description.ValueString()
	}
	if !plan.RootInstanceName.IsNull() && !plan.RootInstanceName.IsUnknown() {
		appJsonMap["rootInstanceName"] = plan.RootInstanceName.ValueString()
	}

	// Using action_query_names_to_connection_ids, replace connection ids in the update request attributes with the ones provided in the plan
	err := replaceConnectionIDsInActionQueries(plan.ActionQueryNamesToConnectionIDs, appJsonMap)
	if err != nil {
		err = fmt.Errorf("error replacing connection IDs in queries: %s", err.Error())
		return err
	}

	return nil
}

// replace the connection ids in the queries with the ones provided in the plan, as specified by {action_query_name: connection_id}
func replaceConnectionIDsInActionQueries(overrideActionQueryNamesToConnectionIDs types.Map, appJsonMap map[string]any) error {
	// skip if overrideActionQueryNamesToConnectionIDs is empty
	if overrideActionQueryNamesToConnectionIDs.IsNull() || overrideActionQueryNamesToConnectionIDs.IsUnknown() {
		return nil
	}

	queries, ok := appJsonMap["queries"].([]any)
	if !ok {
		return nil
	}

	// Create a properly typed map of overrides
	overrides := make(map[string]string)
	for k, v := range overrideActionQueryNamesToConnectionIDs.Elements() {
		strVal, ok := v.(types.String)
		if !ok {
			return fmt.Errorf("expected string value for key %s, got %T", k, v)
		}
		overrides[k] = strVal.ValueString()
	}

	// Track which overrides have been used
	unusedOverrides := maps.Clone(overrides)

	// Update connection IDs in queries
	for _, q := range queries {
		query, ok := q.(map[string]any)
		if !ok {
			continue
		}

		queryName, ok := query["name"].(string)
		if !ok {
			continue
		}

		// Only process action queries
		if queryType, ok := query["type"].(string); !ok || queryType != "action" {
			continue
		}

		// Check if we have an override for this query
		if connectionID, ok := overrides[queryName]; ok {
			properties, ok := query["properties"].(map[string]any)
			if !ok {
				continue
			}

			spec, ok := properties["spec"].(map[string]any)
			if !ok {
				continue
			}

			spec["connectionId"] = connectionID
			delete(unusedOverrides, queryName)
		}
	}

	// Return error if any overrides weren't used
	if len(unusedOverrides) > 0 {
		return fmt.Errorf("action Query Names not found in the App's queries: %v", slices.Collect(maps.Keys(unusedOverrides)))
	}

	return nil
}

func handleAppBuilderPublishState(ctx context.Context, diags *diag.Diagnostics, api *datadogV2.AppBuilderApi, plan *appBuilderAppResourceModel, appID uuid.UUID) (ok bool) {
	// publish the app if the published attribute is true
	if plan.Published.ValueBool() {
		_, httpResp, err := api.PublishApp(ctx, appID)
		if err != nil {
			if httpResp != nil {
				// error body may have useful info for the user
				body, err := io.ReadAll(httpResp.Body)
				if err != nil {
					diags.AddError("error reading error response", err.Error())
					return false
				}

				// if error is related to the app already being published, we can ignore it
				if strings.Contains(string(body), ErrDeploymentExists) {
					return true
				}
				diags.AddError("error publishing app", string(body))
				return false
			}
			diags.AddError("error publishing app", err.Error())
			return false
		}
	} else {
		// unpublish the app if the published attribute is false
		_, httpResp, err := api.UnpublishApp(ctx, appID)
		if err != nil {
			if httpResp != nil {
				// error body may have useful info for the user
				body, err := io.ReadAll(httpResp.Body)
				if err != nil {
					diags.AddError("error reading error response", err.Error())
					return false
				}

				// if error is related to the app already being unpublished, we can ignore it
				if strings.Contains(string(body), ErrAlreadyUnpublished) {
					return true
				}
				diags.AddError("error unpublishing app", string(body))
				return false
			}
			diags.AddError("error unpublishing app", err.Error())
			return false
		}
	}

	return true
}
