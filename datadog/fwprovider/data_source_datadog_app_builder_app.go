package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/customtypes"
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
	// Used to identify requests made from Terraform
	d.Api.Client.Cfg.AddDefaultHeader("X-Datadog-App-Builder-Source", "terraform")
	d.Auth = providerData.Auth
}

func (d *appBuilderAppDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "app_builder_app"
}

func (d *appBuilderAppDataSource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "This data source retrieves the definition of an existing Datadog App from App Builder for use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID for the App.",
				Required:    true,
			},
			"app_json": schema.StringAttribute{
				Computed:    true,
				Description: "The JSON representation of the App.",
				CustomType:  customtypes.AppBuilderAppStringType{},
			},
			"action_query_names_to_connection_ids": schema.MapAttribute{
				Computed:    true,
				Description: "A computed map of the App's Action Query Names to Action Connection IDs.",
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
			"tags": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A list of tags for the app, which can be used to filter apps.",
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
	appBuilderAppModel := &appBuilderAppResourceModel{
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
	// we use AppBuilderAppString type and value to ignore inconsequential differences in the JSON strings
	// and also ignore other differences such as the App's ID, which is ignored in the App Builder API
	appBuilderAppModel.AppJson = customtypes.NewAppBuilderAppStringValue(string(marshalledBytes))

	// build action_query_names_to_connection_ids map
	queries := attributes.GetQueries()
	actionQueryNamesToConnectionIDs, err := buildActionQueryNamesToConnectionIDsMap(queries)
	if err != nil {
		err = fmt.Errorf("error building action_query_names_to_connection_ids map: %s", err)
		return nil, err
	}
	appBuilderAppModel.ActionQueryNamesToConnectionIDs = actionQueryNamesToConnectionIDs

	// get other attributes like name, description, root_instance_name, tags, published
	appBuilderAppModel.Name = types.StringValue(attributes.GetName())
	appBuilderAppModel.Description = types.StringValue(attributes.GetDescription())
	appBuilderAppModel.RootInstanceName = types.StringValue(attributes.GetRootInstanceName())

	// tags is a set of strings -> need to convert []string to []attr.Value
	attrTags := convertTagsToAttrValues(attributes.GetTags())
	appBuilderAppModel.Tags = types.SetValueMust(types.StringType, attrTags)

	// published is a bool -> fetch published status from response included object (deployment)
	if included, ok := resp.GetIncludedOk(); ok {
		deployment := (*included)[0]
		if deployment.Attributes.GetAppVersionId() != uuid.Nil {
			appBuilderAppModel.Published = types.BoolValue(true)
		} else {
			appBuilderAppModel.Published = types.BoolValue(false)
		}
	}

	return appBuilderAppModel, nil
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

func convertTagsToAttrValues(tags []string) []attr.Value {
	attrTags := []attr.Value{}
	for _, tag := range tags {
		attrTags = append(attrTags, types.StringValue(tag))
	}
	return attrTags
}
