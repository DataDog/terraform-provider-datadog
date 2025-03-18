package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/customtypes"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &appBuilderAppResource{}
	_ resource.ResourceWithImportState = &appBuilderAppResource{}
)

type appBuilderAppResource struct {
	Api  *datadogV2.AppBuilderApi
	Auth context.Context
}

// try single property JSON input -> validation will be handled on the API side
type appBuilderAppResourceModel struct {
	ID                                      types.String                         `tfsdk:"id"`
	AppJson                                 customtypes.AppBuilderAppStringValue `tfsdk:"app_json"`
	OverrideActionQueryNamesToConnectionIDs types.Map                            `tfsdk:"override_action_query_names_to_connection_ids"`
	ActionQueryNamesToConnectionIDs         types.Map                            `tfsdk:"action_query_names_to_connection_ids"`
	Name                                    types.String                         `tfsdk:"name"`
	Description                             types.String                         `tfsdk:"description"`
	RootInstanceName                        types.String                         `tfsdk:"root_instance_name"`
	Tags                                    types.Set                            `tfsdk:"tags"`
	Published                               types.Bool                           `tfsdk:"published"`
}

func NewAppBuilderAppResource() resource.Resource {
	return &appBuilderAppResource{}
}

func (r *appBuilderAppResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAppBuilderApiV2()
	// Used to identify requests made from Terraform
	r.Api.Client.Cfg.AddDefaultHeader("X-Datadog-App-Builder-Source", "terraform")
	r.Auth = providerData.Auth
}

func (r *appBuilderAppResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "app_builder_app"
}

func (r *appBuilderAppResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog App resource for creating and managing Datadog Apps from App Builder using the JSON definition.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"app_json": schema.StringAttribute{
				Required:    true,
				Description: "The JSON representation of the App.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				CustomType: customtypes.AppBuilderAppStringType{},
			},
			"override_action_query_names_to_connection_ids": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "If specified, this will override the Action Connection IDs for the specified Action Query Names in the App JSON.",
			},
			"action_query_names_to_connection_ids": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A computed map of the App's Action Query Names to Action Connection IDs.",
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
			// we use SetAttribute to represent tags, paradoxically to be able to maintain them ordered;
			// we order them explicitly in the PlanModifiers of this resource and using
			// SetAttribute makes Terraform ignore differences in order when creating a plan
			"tags": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "A list of tags for the app, which can be used to filter apps. If specified, this will override the list of tags for the App in the App JSON. Otherwise, tags will be returned in output.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				// PlanModifiers: []planmodifier.Set{
				// 	planmodifiers.NormalizeTagSet(),
				// },
			},
			"published": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
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

	// Add debug before create
	fmt.Printf("DEBUG - Create - About to send request:\n%+v\n", createReq)

	// create the app
	createResp, httpResp, err := r.Api.CreateApp(r.Auth, *createReq)
	if err != nil {
		if httpResp != nil {
			body, _ := io.ReadAll(httpResp.Body)
			fmt.Printf("DEBUG - Create - Error response body: %s\n", string(body))
		}
		response.Diagnostics.AddError("error creating app", err.Error())
		return
	}

	// Add these debug statements
	fmt.Printf("DEBUG - Create - Raw API Response:\n%+v\n", createResp)
	if httpResp != nil {
		body, _ := io.ReadAll(httpResp.Body)
		fmt.Printf("DEBUG - Create - Raw HTTP Response body:\n%s\n", string(body))
	}

	appID := createResp.Data.GetId()

	// publish the app if the published attribute is true
	if plan.Published.ValueBool() {
		_, httpResp, err := r.Api.PublishApp(r.Auth, appID)
		if err != nil {
			if httpResp != nil {
				// error body may have useful info for the user
				body, err := io.ReadAll(httpResp.Body)
				if err != nil {
					response.Diagnostics.AddError("error reading error response", err.Error())
					return
				}
				response.Diagnostics.AddError("error publishing app", string(body))
			} else {
				response.Diagnostics.AddError("error publishing app", err.Error())
			}
			return
		}
	}

	// set computed values
	plan.ID = types.StringValue(appID.String())
	attributes := createReq.GetData().Attributes
	plan.ActionQueryNamesToConnectionIDs, err = buildActionQueryNamesToConnectionIDsMap(attributes.GetQueries())
	if err != nil {
		response.Diagnostics.AddError("error building action_query_names_to_connection_ids map", err.Error())
		return
	}

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
	if plan.Tags.IsNull() || plan.Tags.IsUnknown() {
		attrTags := convertTagsToAttrValues(attributes.GetTags())
		tags, newDiags := types.SetValue(types.StringType, attrTags)

		// if there is an error converting the tags to a set, set the tags to an empty set
		if newDiags.HasError() {
			plan.Tags = types.SetValueMust(types.StringType, []attr.Value{})
		} else {
			plan.Tags = tags
		}
	}
	if plan.Published.IsNull() || plan.Published.IsUnknown() {
		plan.Published = types.BoolValue(false)
	}

	// prevent type conversion error when the map is null
	if plan.OverrideActionQueryNamesToConnectionIDs.IsNull() {
		plan.OverrideActionQueryNamesToConnectionIDs = types.MapValueMust(
			types.StringType,
			map[string]attr.Value{},
		)
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

	fmt.Printf("DEBUG - Update - Plan tags: %+v\n", plan.Tags)
	fmt.Printf("DEBUG - Update - State tags: %+v\n", state.Tags)

	updateReq, err := appBuilderAppModelToUpdateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to create app update request",
			err.Error(),
		)
		return
	}

	fmt.Printf("DEBUG - Update Request - Final request:\n%+v\n", updateReq)

	appId, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to parse app ID",
			err.Error(),
		)
		return
	}

	updateResp, httpResp, err := r.Api.UpdateApp(r.Auth, appId, *updateReq)
	if err != nil {
		if httpResp != nil {
			body, _ := io.ReadAll(httpResp.Body)
			fmt.Printf("DEBUG - Update - Error response body: %s\n", string(body))
		}
		response.Diagnostics.AddError(
			"Unable to update app",
			err.Error(),
		)
		return
	}

	fmt.Printf("DEBUG - Update - Raw API Response:\n%+v\n", updateResp)
	if httpResp != nil {
		body, _ := io.ReadAll(httpResp.Body)
		fmt.Printf("DEBUG - Update - Raw HTTP Response body:\n%s\n", string(body))
	}

	// publish the app if the published attribute is true
	if plan.Published.ValueBool() {
		_, httpResp, err := r.Api.PublishApp(r.Auth, appId)
		if err != nil {
			if httpResp != nil {
				// error body may have useful info for the user
				body, err := io.ReadAll(httpResp.Body)
				if err != nil {
					response.Diagnostics.AddError("error reading error response", err.Error())
					return
				}
				response.Diagnostics.AddError("error publishing app", string(body))
			} else {
				response.Diagnostics.AddError("error publishing app", err.Error())
			}
			return
		}
	} else {
		// unpublish the app if the published attribute is false
		_, httpResp, err := r.Api.UnpublishApp(r.Auth, appId)
		if err != nil {
			if httpResp != nil {
				// error body may have useful info for the user
				body, err := io.ReadAll(httpResp.Body)
				if err != nil {
					response.Diagnostics.AddError("error reading error response", err.Error())
					return
				}
				response.Diagnostics.AddError("error unpublishing app", string(body))
			} else {
				response.Diagnostics.AddError("error unpublishing app", err.Error())
			}
			return
		}
	}

	// set computed values
	attributes := updateReq.GetData().Attributes
	plan.ActionQueryNamesToConnectionIDs, err = buildActionQueryNamesToConnectionIDsMap(attributes.GetQueries())
	if err != nil {
		response.Diagnostics.AddError("error building action_query_names_to_connection_ids map", err.Error())
		return
	}

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
	if plan.Tags.IsNull() || plan.Tags.IsUnknown() {
		attrTags := convertTagsToAttrValues(attributes.GetTags())
		tags, newDiags := types.SetValue(types.StringType, attrTags)

		// if there is an error converting the tags to a set, set the tags to an empty set
		if newDiags.HasError() {
			plan.Tags = types.SetValueMust(types.StringType, []attr.Value{})
		} else {
			plan.Tags = tags
		}
	}
	if plan.Published.IsNull() || plan.Published.IsUnknown() {
		plan.Published = types.BoolValue(false)
	}

	// prevent type conversion error when the map is null
	if plan.OverrideActionQueryNamesToConnectionIDs.IsNull() {
		plan.OverrideActionQueryNamesToConnectionIDs = types.MapValueMust(
			types.StringType,
			map[string]attr.Value{},
		)
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

func appBuilderAppModelToCreateApiRequest(plan appBuilderAppResourceModel) (*datadogV2.CreateAppRequest, error) {
	// unmarshal app JSON into a map
	var appJson map[string]interface{}
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
	var appJson map[string]interface{}
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

func overrideAppBuilderAppAttributesInCreateRequestAttributes(plan appBuilderAppResourceModel, appJsonMap map[string]interface{}) error {
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

	// Handle tags: if explicitly set in plan, use those; otherwise use tags from app JSON
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		setElements := plan.Tags.Elements()
		tags := make([]string, 0, len(setElements))
		for _, element := range setElements {
			tags = append(tags, element.(types.String).ValueString())
		}
		sort.Strings(tags)
		appJsonMap["tags"] = tags
	} else if existingTags, ok := appJsonMap["tags"].([]interface{}); ok && len(existingTags) > 0 {
		tags := make([]string, len(existingTags))
		for i, tag := range existingTags {
			if strTag, ok := tag.(string); ok {
				tags[i] = strTag
			}
		}
		sort.Strings(tags)
		appJsonMap["tags"] = tags
	}

	// Using override_action_query_names_to_connection_ids, replace connection ids in the update request attributes with the ones provided in the plan
	err := replaceConnectionIDsInActionQueries(plan.OverrideActionQueryNamesToConnectionIDs, appJsonMap)
	if err != nil {
		err = fmt.Errorf("error replacing connection IDs in queries: %s", err.Error())
		return err
	}

	return nil
}

func overrideAppBuilderAppAttributesInUpdateRequestAttributes(plan appBuilderAppResourceModel, appJsonMap map[string]interface{}) error {
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

	// Handle tags: if explicitly set in plan, use those; otherwise use tags from app JSON
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		setElements := plan.Tags.Elements()
		tags := make([]string, 0, len(setElements))
		for _, element := range setElements {
			tags = append(tags, element.(types.String).ValueString())
		}
		sort.Strings(tags)
		appJsonMap["tags"] = tags
	} else if existingTags := appJsonMap["tags"].([]string); len(existingTags) > 0 {
		// Preserve existing tags from app JSON if no override
		sort.Strings(existingTags)
		appJsonMap["tags"] = existingTags
	}

	// Using override_action_query_names_to_connection_ids, replace connection ids in the update request attributes with the ones provided in the plan
	err := replaceConnectionIDsInActionQueries(plan.OverrideActionQueryNamesToConnectionIDs, appJsonMap)
	if err != nil {
		err = fmt.Errorf("error replacing connection IDs in queries: %s", err.Error())
		return err
	}

	return nil
}

// replace the connection ids in the queries with the ones provided in the plan, as specified by {action_query_name: connection_id}
func replaceConnectionIDsInActionQueries(overrideActionQueryNamesToConnectionIDs types.Map, appJsonMap map[string]interface{}) error {
	// skip if overrideActionQueryNamesToConnectionIDs is empty
	if overrideActionQueryNamesToConnectionIDs.IsNull() || overrideActionQueryNamesToConnectionIDs.IsUnknown() {
		return nil
	}

	// Handle connection ID overrides
	queries, ok := appJsonMap["queries"].([]interface{})
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
		query, ok := q.(map[string]interface{})
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
			properties, ok := query["properties"].(map[string]interface{})
			if !ok {
				continue
			}

			spec, ok := properties["spec"].(map[string]interface{})
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
