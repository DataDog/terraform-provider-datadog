package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes" // v0.1.0, else breaking
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
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
	ID                            types.String         `tfsdk:"id"`
	AppJson                       jsontypes.Normalized `tfsdk:"app_json"`
	ActionQueryIDsToConnectionIDs types.Map            `tfsdk:"action_query_ids_to_connection_ids"`
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
				CustomType: jsontypes.NormalizedType{},
			},
			"action_query_ids_to_connection_ids": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A map of the App's Action Query IDs to Action Connection IDs. If specified, this will override the Action Connection IDs in the App JSON.",
				ElementType: types.StringType,
			},
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
	if plan.ActionQueryIDsToConnectionIDs.IsUnknown() {
		plan.ActionQueryIDsToConnectionIDs = types.MapNull(types.StringType)
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
	if plan.ActionQueryIDsToConnectionIDs.IsUnknown() {
		plan.ActionQueryIDsToConnectionIDs = types.MapNull(types.StringType)
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
	appBuilderAppJSONModel.AppJson = jsontypes.NewNormalizedValue(string(marshalledBytes))

	// build action query ids to connection ids map
	queries := attributes.GetQueries()
	actionQueryIDsToConnectionIDs, err := buildActionQueryIDsToConnectionIDsMap(queries)
	if err != nil {
		err = fmt.Errorf("error building action_query_ids_to_connection_ids map: %s", err)
		return nil, err
	}
	appBuilderAppJSONModel.ActionQueryIDsToConnectionIDs = actionQueryIDsToConnectionIDs

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
	err = replaceConnectionIDsInActionQueries(plan.ActionQueryIDsToConnectionIDs, attributes.GetQueries())
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
	err = replaceConnectionIDsInActionQueries(plan.ActionQueryIDsToConnectionIDs, attributes.GetQueries())
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

// replace the connection ids in the queries with the ones provided in the plan, as specified by {action_query_id: connection_id}
func replaceConnectionIDsInActionQueries(actionQueryIDsToConnectionIDs types.Map, queries []datadogV2.Query) error {
	mapElements := actionQueryIDsToConnectionIDs.Elements()

	// skip if no action query ids to connection ids are provided
	if len(mapElements) == 0 {
		return nil
	}

	// keep track of specified query ids that have not been used
	unusedQueryIDs := make(map[string]struct{})
	for key := range mapElements {
		unusedQueryIDs[key] = struct{}{}
	}

	// loop over queries list and replace connection ids if the query id is in the map
	for i, query := range queries {
		queryID := query.GetId().String()
		if connectionID, ok := mapElements[queryID]; ok {
			// must be an action query
			queryType := query.GetType()
			if queryType != datadogV2.QUERYTYPE_ACTION {
				return fmt.Errorf("query with ID %s is not an Action Query: %s", queryID, queryType)
			}

			connectionIDString := connectionID.(types.String).ValueString()
			err := setConnectionIDForActionQuery(queryID, &queries[i], connectionIDString)
			if err != nil {
				return err
			}
			delete(unusedQueryIDs, queryID)
		}
	}

	// return err if any query ids specified in the map were not found in the queries list
	if len(unusedQueryIDs) > 0 {
		return fmt.Errorf("action Query IDs not found in the App's queries: %v", slices.Collect(maps.Keys(unusedQueryIDs)))
	}

	return nil
}

func setConnectionIDForActionQuery(queryID string, query *datadogV2.Query, connectionIDString string) error {
	// TODO: update this once the API strictly types the Action Query schema
	properties := query.GetProperties().(map[string]any)
	spec, ok := properties["spec"]
	if !ok {
		return fmt.Errorf("action Query with ID %s is missing a spec", queryID)
	}

	specMap, ok := spec.(map[string]any)
	if !ok {
		return fmt.Errorf("action Query with ID %s has invalid spec: %s", queryID, specMap)
	}

	// UUID validation also happens in API, but doing it here can help Terraform users catch errors earlier
	_, err := uuid.Parse(connectionIDString)
	if err != nil {
		err = fmt.Errorf("specified Connection ID %s is not a valid UUID: %s", connectionIDString, err)
		return err
	}

	specMap["connectionId"] = connectionIDString
	return nil
}

func buildActionQueryIDsToConnectionIDsMap(queries []datadogV2.Query) (types.Map, error) {
	elementsMap := map[string]attr.Value{}

	for _, query := range queries {
		queryID := query.GetId().String()

		// must be an action query
		queryType := query.GetType()
		if queryType != datadogV2.QUERYTYPE_ACTION {
			continue
		}

		// TODO: update this once the API strictly types the Action Query schema
		// since we are reading the response from the API, we can ignore validation errors
		properties, ok := query.GetProperties().(map[string]any)
		if !ok {
			continue
		}
		spec, ok := properties["spec"]
		if !ok {
			continue
		}
		specMap, ok := spec.(map[string]any)
		if !ok {
			continue
		}

		connectionID, ok := specMap["connectionId"]
		if !ok {
			continue
		}
		if connectionIDStr, ok := connectionID.(string); ok {
			elementsMap[queryID] = types.StringValue(connectionIDStr)
		}
	}

	// convert map to types.Map
	resultMap, diags := types.MapValue(types.StringType, elementsMap)
	if diags != nil {
		return types.MapNull(types.StringType), fmt.Errorf("error converting map to types.Map: %v", diags)
	}

	return resultMap, nil
}
