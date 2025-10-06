package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &appBuilderAppDataSource{}

type appBuilderAppDataSource struct {
	Api  *datadogV2.AppBuilderApi
	Auth context.Context
}

func NewDatadogAppBuilderAppDataSource() datasource.DataSource {
	return &appBuilderAppDataSource{}
}

func (d *appBuilderAppDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAppBuilderApiV2()
	d.Auth = providerData.Auth
}

func (d *appBuilderAppDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "app_builder_app"
}

func (d *appBuilderAppDataSource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "This data source retrieves the definition of an existing Datadog App from App Builder for use in other resources, such as embedding Apps in Dashboards. This data source requires a [registered application key](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/app_key_registration).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID for the App.",
				Required:    true,
			},
			"app_json": schema.StringAttribute{
				Computed:    true,
				Description: "The JSON representation of the App.",
			},
			"action_query_names_to_connection_ids": schema.MapAttribute{
				Computed:    true,
				Description: "A map of the App's Action Query Names to Action Connection IDs.",
				ElementType: types.StringType,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the App.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The human-readable description of the App.",
			},
			"root_instance_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the root component of the app. This is a grid component that contains all other components.",
			},
			"published": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the app is published or unpublished. Published apps are available to other users. To ensure the app is accessible to the correct users, you also need to set a [Restriction Policy](https://docs.datadoghq.com/api/latest/restriction-policies/) on the app if a policy does not yet exist.",
			},
		},
	}
}

func (d *appBuilderAppDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state appBuilderAppResourceModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("error parsing id as uuid", err.Error())
		return
	}

	appBuilderAppModel, err := readAppBuilderApp(d.Auth, d.Api, id)
	if err != nil {
		response.Diagnostics.AddError("error reading app", err.Error())
		return
	}

	diags = response.State.Set(ctx, appBuilderAppModel)
	response.Diagnostics.Append(diags...)
}

// Read logic is shared between data source and resource
func readAppBuilderApp(ctx context.Context, api *datadogV2.AppBuilderApi, id uuid.UUID) (*appBuilderAppResourceModel, error) {
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

	appModel, err := apiResponseToAppBuilderAppModel(resp)
	if err != nil {
		return nil, err
	}

	return appModel, nil
}

func apiResponseToAppBuilderAppModel(resp datadogV2.GetAppResponse) (*appBuilderAppResourceModel, error) {
	data := resp.GetData()
	attributes := data.GetAttributes()

	// Create a copy of the attributes that we can modify
	var appJson map[string]any
	marshalledBytes, err := json.Marshal(attributes)
	if err != nil {
		return nil, fmt.Errorf("error marshaling attributes: %s", err)
	}
	if err := json.Unmarshal(marshalledBytes, &appJson); err != nil {
		return nil, fmt.Errorf("error unmarshaling attributes: %s", err)
	}

	// Initialize the model with the ID
	appBuilderAppModel := &appBuilderAppResourceModel{
		ID: types.StringValue(data.GetId().String()),
	}

	// Set the individual fields from the attributes
	appBuilderAppModel.Name = types.StringValue(attributes.GetName())
	appBuilderAppModel.Description = types.StringValue(attributes.GetDescription())
	appBuilderAppModel.RootInstanceName = types.StringValue(attributes.GetRootInstanceName())

	// Handle published status
	if included, ok := resp.GetIncludedOk(); ok && len(*included) > 0 {
		deployment := (*included)[0]
		appBuilderAppModel.Published = types.BoolValue(deployment.Attributes.GetAppVersionId() != uuid.Nil)
	} else {
		appBuilderAppModel.Published = types.BoolValue(false)
	}

	// Handle action query maps
	actionQueryNamesToConnectionIDs, err := buildActionQueryNamesToConnectionIDsMap(attributes.GetQueries())
	if err != nil {
		return nil, fmt.Errorf("error building action_query_names_to_connection_ids map: %s", err)
	}
	appBuilderAppModel.ActionQueryNamesToConnectionIDs = actionQueryNamesToConnectionIDs

	// Marshal the modified app_json back to string
	marshalledBytes, err = json.Marshal(appJson)
	if err != nil {
		return nil, fmt.Errorf("error marshaling app_json: %s", err)
	}
	appBuilderAppModel.AppJson = types.StringValue(string(marshalledBytes))
	return appBuilderAppModel, nil
}

func buildActionQueryNamesToConnectionIDsMap(queries []datadogV2.Query) (types.Map, error) {
	elementsMap := map[string]string{}

	for _, query := range queries {
		// must be an action query
		actionQuery := query.ActionQuery
		if actionQuery == nil {
			continue
		}

		queryName := actionQuery.GetName()

		specObj := actionQuery.Properties.GetSpec().ActionQuerySpecObject
		if specObj.HasConnectionId() {
			connectionID := specObj.GetConnectionId()
			elementsMap[queryName] = connectionID
		}
	}

	// convert map to types.Map
	ctx := context.Background()
	resultMap, diags := types.MapValueFrom(ctx, types.StringType, elementsMap)
	if diags != nil {
		return types.MapNull(types.StringType), fmt.Errorf("error converting map to types.Map: %v", diags)
	}

	return resultMap, nil
}
