package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid" // v0.1.0, else breaking
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/customtypes"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/planmodifiers"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &appBuilderAppJSONResource{}
	_ resource.ResourceWithImportState = &appBuilderAppJSONResource{}
)

type appBuilderAppJSONResource struct {
	Api  *datadogV2.AppBuilderApi
	Auth context.Context
}

// try single property JSON input -> validation will be handled on the API side
type appBuilderAppJSONResourceModel struct {
	ID                                      types.String                             `tfsdk:"id"`
	AppJson                                 customtypes.AppBuilderAppJSONStringValue `tfsdk:"app_json"`
	OverrideActionQueryNamesToConnectionIDs types.Map                                `tfsdk:"override_action_query_names_to_connection_ids"`
	ActionQueryNamesToConnectionIDs         types.Map                                `tfsdk:"action_query_names_to_connection_ids"`
	Name                                    types.String                             `tfsdk:"name"`
	Description                             types.String                             `tfsdk:"description"`
	RootInstanceName                        types.String                             `tfsdk:"root_instance_name"`
	Tags                                    types.Set                                `tfsdk:"tags"`
	PublishStatusUpdate                     types.String                             `tfsdk:"publish_status_update"`
}

func NewAppBuilderAppJSONResource() resource.Resource {
	return &appBuilderAppJSONResource{}
}

func (r *appBuilderAppJSONResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAppBuilderApiV2()
	r.Auth = providerData.Auth
}

func (r *appBuilderAppJSONResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "app_builder_app_json"
}

func (r *appBuilderAppJSONResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog App JSON resource for creating and managing Datadog Apps from App Builder using the JSON definition.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"app_json": schema.StringAttribute{
				Required:    true,
				Description: "The JSON representation of the App.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				CustomType: customtypes.AppBuilderAppJSONStringType{},
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
				Description: "If specified, this will override the name of the App in the App JSON.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "If specified, this will override the human-readable description of the App in the App JSON.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"root_instance_name": schema.StringAttribute{
				Optional:    true,
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
					// validators.TagsSetIsNormalized(),
				},
				PlanModifiers: []planmodifier.Set{
					planmodifiers.NormalizeTagSet(),
				},
			},
			"publish_status_update": schema.StringAttribute{
				Optional:    true,
				Description: "If `publish`, the latest app version will be published and available to other users. To ensure the app is accessible to the correct users, you also need to set a [Restriction Policy](https://docs.datadoghq.com/api/latest/restriction-policies/) on the app if a policy does not yet exist. If `unpublish`, the app will be unpublished, removing the live version of the app. If unspecified, the publish status will not be updated.",
				Validators:  []validator.String{validators.NewEnumValidator[validator.String](NewPublishStatusUpdateFromValue)},
			},
			// TODO: update CRUD operations to handle the new optional fields
		},
	}
}

func (r *appBuilderAppJSONResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *appBuilderAppJSONResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan appBuilderAppJSONResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	createRequest, err := appBuilderAppJSONModelToCreateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("error building create app request", err.Error())
		return
	}

	resp, httpResp, err := r.Api.CreateApp(r.Auth, *createRequest)
	if err != nil {
		if httpResp != nil {
			// error body may have useful info for the user
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("error reading error response", err.Error())
				return
			}
			response.Diagnostics.AddError("error creating app", string(body))
		} else {
			response.Diagnostics.AddError("error creating app", err.Error())
		}
		return
	}

	// set computed values
	plan.ID = types.StringValue(resp.Data.GetId().String())
	plan.ActionQueryNamesToConnectionIDs, err = buildActionQueryNamesToConnectionIDsMap(createRequest.GetData().Attributes.GetQueries())
	if err != nil {
		response.Diagnostics.AddError("error building action_query_names_to_connection_ids map", err.Error())
		return
	}

	// Save data into Terraform state
	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *appBuilderAppJSONResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state appBuilderAppJSONResourceModel
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

	appBuilderAppJSONModel, err := readAppBuilderAppJSON(r.Auth, r.Api, id)
	if err != nil {
		response.Diagnostics.AddError("error reading app", err.Error())
		return
	}

	// Save data into Terraform state
	diags = response.State.Set(ctx, appBuilderAppJSONModel)
	response.Diagnostics.Append(diags...)
}

func (r *appBuilderAppJSONResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan appBuilderAppJSONResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("error parsing ID as UUID", err.Error())
		return
	}

	updateRequest, err := appBuilderAppJSONModelToUpdateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("error building update app request", err.Error())
		return
	}

	_, httpResp, err := r.Api.UpdateApp(r.Auth, id, *updateRequest)
	if err != nil {
		if httpResp != nil {
			// error body may have useful info for the user
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("error reading error response", err.Error())
				return
			}
			response.Diagnostics.AddError("error updating app", string(body))
		} else {
			response.Diagnostics.AddError("error updating app", err.Error())
		}
		return
	}

	// set computed values
	plan.ActionQueryNamesToConnectionIDs, err = buildActionQueryNamesToConnectionIDsMap(updateRequest.GetData().Attributes.GetQueries())
	if err != nil {
		response.Diagnostics.AddError("error building action_query_names_to_connection_ids map", err.Error())
		return
	}

	// Save data into Terraform state
	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *appBuilderAppJSONResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state appBuilderAppJSONResourceModel
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

func apiResponseToAppBuilderAppJSONModel(resp datadogV2.GetAppResponse) (*appBuilderAppJSONResourceModel, error) {
	appBuilderAppJSONModel := &appBuilderAppJSONResourceModel{
		ID: types.StringValue(resp.Data.GetId().String()),
	}

	data := resp.GetData()
	attributes := data.GetAttributes()

	// marshal attributes into JSON string and set it as the app_json value
	marshalledBytes, err := json.Marshal(attributes)
	if err != nil {
		err = fmt.Errorf("error marshaling attributes: %s", err)
		return nil, err
	}
	// we use AppBuilderAppJSONString type and value to ignore inconsequential differences in the JSON strings
	// and also ignore other differences such as the App's ID, which is ignored in the App Builder API
	appBuilderAppJSONModel.AppJson = customtypes.NewAppBuilderAppJSONStringValue(string(marshalledBytes))

	// build action query ids to connection ids map
	queries := attributes.GetQueries()
	actionQueryNamesToConnectionIDs, err := buildActionQueryNamesToConnectionIDsMap(queries)
	if err != nil {
		err = fmt.Errorf("error building action_query_names_to_connection_ids map: %s", err)
		return nil, err
	}
	appBuilderAppJSONModel.ActionQueryNamesToConnectionIDs = actionQueryNamesToConnectionIDs

	return appBuilderAppJSONModel, nil
}

func appBuilderAppJSONModelToCreateApiRequest(plan appBuilderAppJSONResourceModel) (*datadogV2.CreateAppRequest, error) {
	attributes := datadogV2.NewCreateAppRequestDataAttributesWithDefaults()

	// decode encoded json into the attributes struct
	err := json.Unmarshal([]byte(plan.AppJson.ValueString()), attributes)
	if err != nil {
		err = fmt.Errorf("error unmarshalling app JSON string to attributes struct: %s", err)
		return nil, err
	}

	// replace connection ids in the create request attributes with the ones provided in the plan
	err = replaceConnectionIDsInActionQueries(plan.OverrideActionQueryNamesToConnectionIDs, attributes.GetQueries())
	if err != nil {
		err = fmt.Errorf("error replacing connection IDs in queries: %s", err.Error())
		return nil, err
	}

	req := datadogV2.NewCreateAppRequestWithDefaults()
	req.Data = datadogV2.NewCreateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, nil
}

func appBuilderAppJSONModelToUpdateApiRequest(plan appBuilderAppJSONResourceModel) (*datadogV2.UpdateAppRequest, error) {
	attributes := datadogV2.NewUpdateAppRequestDataAttributesWithDefaults()

	// decode encoded json into the attributes struct
	err := json.Unmarshal([]byte(plan.AppJson.ValueString()), attributes)
	if err != nil {
		err = fmt.Errorf("error unmarshalling app JSON string to attributes struct: %s", err)
		return nil, err
	}

	// replace connection ids in the update request attributes with the ones provided in the plan
	err = replaceConnectionIDsInActionQueries(plan.OverrideActionQueryNamesToConnectionIDs, attributes.GetQueries())
	if err != nil {
		err = fmt.Errorf("error replacing connection IDs in queries: %s", err.Error())
		return nil, err
	}

	req := datadogV2.NewUpdateAppRequestWithDefaults()
	req.Data = datadogV2.NewUpdateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, nil
}

// Read logic is shared between data source and resource
func readAppBuilderAppJSON(ctx context.Context, api *datadogV2.AppBuilderApi, id uuid.UUID) (*appBuilderAppJSONResourceModel, error) {
	resp, httpResp, err := api.GetApp(ctx, id)
	if err != nil {
		if httpResp != nil {
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, fmt.Errorf("could not read error response")
			}
			return nil, fmt.Errorf("%s", body)
		}
		return nil, err
	}

	appModel, err := apiResponseToAppBuilderAppJSONModel(resp)
	if err != nil {
		return nil, err
	}

	return appModel, nil
}

// replace the connection ids in the queries with the ones provided in the plan, as specified by {action_query_name: connection_id}
func replaceConnectionIDsInActionQueries(overrideActionQueryNamesToConnectionIDs types.Map, queries []datadogV2.Query) error {
	mapElements := overrideActionQueryNamesToConnectionIDs.Elements()

	// skip if no action query names to connection ids are provided
	if len(mapElements) == 0 {
		return nil
	}

	// keep track of specified query names that have not been used
	unusedQueryNames := make(map[string]struct{})
	for key := range mapElements {
		unusedQueryNames[key] = struct{}{}
	}

	// loop over queries list and replace connection ids if the query name is in the map
	for _, query := range queries {
		// get the query name
		queryName := getQueryName(query)
		actionQuery := query.ActionQuery

		if connectionID, ok := mapElements[queryName]; ok {
			// must be an action query
			if actionQuery == nil {
				return fmt.Errorf("query with Name %s is not an Action Query", queryName)
			}

			connectionIDString := connectionID.(types.String).ValueString()
			err := setConnectionIDForActionQuery(actionQuery, connectionIDString)
			if err != nil {
				return err
			}
			delete(unusedQueryNames, queryName)
		}
	}

	// return err if any query names specified in the map were not found in the queries list
	if len(unusedQueryNames) > 0 {
		return fmt.Errorf("action Query Names not found in the App's queries: %v", slices.Collect(maps.Keys(unusedQueryNames)))
	}

	return nil
}

func setConnectionIDForActionQuery(actionQuery *datadogV2.ActionQuery, connectionIDString string) error {
	// API strictly types the Action Query schema so invalid queries should be caught when unmarshaling, before this step
	specObj := actionQuery.Properties.GetSpec().ActionQuerySpecObject

	// UUID validation also happens in API, but doing it here can help Terraform users catch errors earlier
	_, err := uuid.Parse(connectionIDString)
	if err != nil {
		err = fmt.Errorf("specified Connection ID %s is not a valid UUID: %s", connectionIDString, err)
		return err
	}

	specObj.SetConnectionId(connectionIDString)
	return nil
}

func buildActionQueryNamesToConnectionIDsMap(queries []datadogV2.Query) (types.Map, error) {
	elementsMap := map[string]attr.Value{}

	for _, query := range queries {
		// must be an action query
		actionQuery := query.ActionQuery
		if actionQuery == nil {
			continue
		}

		queryName := actionQuery.GetName()

		// since we are reading the response from the API, we can ignore validation errors
		specObj := actionQuery.Properties.GetSpec().ActionQuerySpecObject
		if specObj.HasConnectionId() {
			connectionID := specObj.GetConnectionId()
			elementsMap[queryName] = types.StringValue(connectionID)
		}
	}

	// convert map to types.Map
	resultMap, diags := types.MapValue(types.StringType, elementsMap)
	if diags != nil {
		return types.MapNull(types.StringType), fmt.Errorf("error converting map to types.Map: %v", diags)
	}

	return resultMap, nil
}

func getQueryName(query datadogV2.Query) string {
	if query.ActionQuery != nil {
		return query.ActionQuery.GetName()
	} else if query.DataTransform != nil {
		return query.DataTransform.GetName()
	} else if query.StateVariable != nil {
		return query.StateVariable.GetName()
	}
	return ""
}

// create enum for App Builder App's PublishStatusUpdate TF field
type PublishStatusUpdate string

// List of PublishStatusUpdate.
const (
	PUBLISHSTATUSUPDATE_PUBLISH   PublishStatusUpdate = "publish"
	PUBLISHSTATUSUPDATE_UNPUBLISH PublishStatusUpdate = "unpublish"
)

var allowedPublishStatusUpdateEnumValues = []PublishStatusUpdate{
	PUBLISHSTATUSUPDATE_PUBLISH,
	PUBLISHSTATUSUPDATE_UNPUBLISH,
}

// GetAllowedValues returns the list of possible values.
func (v *PublishStatusUpdate) GetAllowedValues() []PublishStatusUpdate {
	return allowedPublishStatusUpdateEnumValues
}

// UnmarshalJSON deserializes the given payload.
func (v *PublishStatusUpdate) UnmarshalJSON(src []byte) error {
	var value string
	err := datadog.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	*v = PublishStatusUpdate(value)
	return nil
}

// NewPublishStatusUpdateFromValue returns a pointer to a valid PublishStatusUpdate
// for the value passed as argument, or an error if the value passed is not allowed by the enum.
func NewPublishStatusUpdateFromValue(v string) (*PublishStatusUpdate, error) {
	ev := PublishStatusUpdate(v)
	if ev.IsValid() {
		return &ev, nil
	}
	return nil, fmt.Errorf("invalid value '%v' for PublishStatusUpdate: valid values are %v", v, allowedPublishStatusUpdateEnumValues)
}

// IsValid return true if the value is valid for the enum, false otherwise.
func (v PublishStatusUpdate) IsValid() bool {
	for _, existing := range allowedPublishStatusUpdateEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to PublishStatusUpdate value.
func (v PublishStatusUpdate) Ptr() *PublishStatusUpdate {
	return &v
}
