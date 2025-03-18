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
	"github.com/google/uuid" // v0.1.0, else breaking
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/customtypes"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/planmodifiers"
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
				},
				PlanModifiers: []planmodifier.Set{
					planmodifiers.NormalizeTagSet(),
				},
			},
			"published": schema.BoolAttribute{
				Optional:    true,
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

	// build the create request
	createRequest, err := appBuilderAppModelToCreateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("error building create app request", err.Error())
		return
	}

	// create the app
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

	appID := resp.Data.GetId()

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
	attributes := createRequest.GetData().Attributes
	plan.ActionQueryNamesToConnectionIDs, err = buildActionQueryNamesToConnectionIDsMap(attributes.GetQueries())
	if err != nil {
		response.Diagnostics.AddError("error building action_query_names_to_connection_ids map", err.Error())
		return
	}
	attrTags := convertTagsToAttrValues(attributes.GetTags())
	plan.Tags = types.SetValueMust(types.StringType, attrTags)

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
	var plan appBuilderAppResourceModel
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

	// build the update request
	updateRequest, err := appBuilderAppModelToUpdateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("error building update app request", err.Error())
		return
	}

	// update the app
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

	// publish the app if the published attribute is true
	if plan.Published.ValueBool() {
		_, httpResp, err := r.Api.PublishApp(r.Auth, id)
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
		_, httpResp, err := r.Api.UnpublishApp(r.Auth, id)
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
	attributes := updateRequest.GetData().Attributes
	plan.ActionQueryNamesToConnectionIDs, err = buildActionQueryNamesToConnectionIDsMap(attributes.GetQueries())
	if err != nil {
		response.Diagnostics.AddError("error building action_query_names_to_connection_ids map", err.Error())
		return
	}
	attrTags := convertTagsToAttrValues(attributes.GetTags())
	plan.Tags = types.SetValueMust(types.StringType, attrTags)

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
	attributes := datadogV2.NewCreateAppRequestDataAttributesWithDefaults()

	// decode encoded json into the attributes struct
	err := json.Unmarshal([]byte(plan.AppJson.ValueString()), attributes)
	if err != nil {
		err = fmt.Errorf("error unmarshalling app JSON string to attributes struct: %s", err)
		return nil, err
	}

	// override the attributes with the ones provided in the plan
	err = overrideAppBuilderAppAttributesInCreateRequestAttributes(plan, attributes)
	if err != nil {
		err = fmt.Errorf("error overriding app JSON attributes: %s", err.Error())
		return nil, err
	}

	req := datadogV2.NewCreateAppRequestWithDefaults()
	req.Data = datadogV2.NewCreateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, nil
}

func appBuilderAppModelToUpdateApiRequest(plan appBuilderAppResourceModel) (*datadogV2.UpdateAppRequest, error) {
	attributes := datadogV2.NewUpdateAppRequestDataAttributesWithDefaults()

	// decode encoded json into the attributes struct
	err := json.Unmarshal([]byte(plan.AppJson.ValueString()), attributes)
	if err != nil {
		err = fmt.Errorf("error unmarshalling app JSON string to attributes struct: %s", err)
		return nil, err
	}

	// override the attributes with the ones provided in the plan
	err = overrideAppBuilderAppAttributesInUpdateRequestAttributes(plan, attributes)
	if err != nil {
		err = fmt.Errorf("error overriding app JSON attributes: %s", err.Error())
		return nil, err
	}

	req := datadogV2.NewUpdateAppRequestWithDefaults()
	req.Data = datadogV2.NewUpdateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, nil
}

func overrideAppBuilderAppAttributesInCreateRequestAttributes(plan appBuilderAppResourceModel, attributes *datadogV2.CreateAppRequestDataAttributes) error {

	// name, description, root_instance_name are straightforward string replacements
	if plan.Name.ValueString() != "" {
		attributes.Name = plan.Name.ValueStringPointer()
	}
	if plan.Description.ValueString() != "" {
		attributes.Description = plan.Description.ValueStringPointer()
	}
	if plan.RootInstanceName.ValueString() != "" {
		attributes.RootInstanceName = plan.RootInstanceName.ValueStringPointer()
	}

	// tags are a bit more complex, we need to convert the types.Set to a list of strings and then replace the tags from the app json
	setElements := plan.Tags.Elements()
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() || len(setElements) > 0 {
		tags := []string{}
		for _, element := range setElements {
			tags = append(tags, element.(types.String).ValueString())
		}
		// sort the tags
		sort.Strings(tags)
		attributes.Tags = tags
	}

	// Using override_action_query_names_to_connection_ids, replace connection ids in the update request attributes with the ones provided in the plan
	err := replaceConnectionIDsInActionQueries(plan.OverrideActionQueryNamesToConnectionIDs, attributes.GetQueries())
	if err != nil {
		err = fmt.Errorf("error replacing connection IDs in queries: %s", err.Error())
		return err
	}

	return nil
}

func overrideAppBuilderAppAttributesInUpdateRequestAttributes(plan appBuilderAppResourceModel, attributes *datadogV2.UpdateAppRequestDataAttributes) error {

	// name, description, root_instance_name are straightforward string replacements
	if plan.Name.ValueString() != "" {
		attributes.Name = plan.Name.ValueStringPointer()
	}
	if plan.Description.ValueString() != "" {
		attributes.Description = plan.Description.ValueStringPointer()
	}
	if plan.RootInstanceName.ValueString() != "" {
		attributes.RootInstanceName = plan.RootInstanceName.ValueStringPointer()
	}

	// tags are a bit more complex, we need to convert the types.Set to a list of strings and then replace the tags from the app json
	setElements := plan.Tags.Elements()
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() || len(setElements) > 0 {
		tags := []string{}
		for _, element := range setElements {
			tags = append(tags, element.(types.String).ValueString())
		}
		// sort the tags
		sort.Strings(tags)
		attributes.Tags = tags
	}

	// Using override_action_query_names_to_connection_ids, replace connection ids in the update request attributes with the ones provided in the plan
	err := replaceConnectionIDsInActionQueries(plan.OverrideActionQueryNamesToConnectionIDs, attributes.GetQueries())
	if err != nil {
		err = fmt.Errorf("error replacing connection IDs in queries: %s", err.Error())
		return err
	}

	return nil
}

// replace the connection ids in the queries with the ones provided in the plan, as specified by {action_query_name: connection_id}
func replaceConnectionIDsInActionQueries(overrideActionQueryNamesToConnectionIDs types.Map, queries []datadogV2.Query) error {
	mapElements := overrideActionQueryNamesToConnectionIDs.Elements()

	// skip if overrideActionQueryNamesToConnectionIDs is empty
	if overrideActionQueryNamesToConnectionIDs.IsNull() || overrideActionQueryNamesToConnectionIDs.IsUnknown() || mapElements == nil {
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
