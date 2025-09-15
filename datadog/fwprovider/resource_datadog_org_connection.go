package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &OrgConnectionResource{}
	_ resource.ResourceWithImportState = &OrgConnectionResource{}
)

type OrgConnectionResource struct {
	API  *datadogV2.OrgConnectionsApi
	Auth context.Context
}

type OrgConnectionModel struct {
	ID              types.String      `tfsdk:"id"`
	ConnectionTypes types.Set         `tfsdk:"connection_types"`
	SinkOrgID       types.String      `tfsdk:"sink_org_id"`
	SourceOrgID     types.String      `tfsdk:"source_org_id"`
	SinkOrgName     types.String      `tfsdk:"sink_org_name"`
	SourceOrgName   types.String      `tfsdk:"source_org_name"`
	CreatedAt       timetypes.RFC3339 `tfsdk:"created_at"`
	CreatedBy       types.String      `tfsdk:"created_by"`
}

func NewOrgConnectionResource() resource.Resource {
	return &OrgConnectionResource{}
}

func (r *OrgConnectionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetOrgConnectionsApiV2()
	r.Auth = providerData.Auth
}

func (r *OrgConnectionResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "org_connection"
}

func (r *OrgConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Org Connection resource. Org Connections make data from a source org available in the sink org for selected connection data types (e.g., logs or metrics). Org Connections can only be created from a source org to a specified sink org",

		Attributes: map[string]schema.Attribute{
			"connection_types": schema.SetAttribute{
				Description: "Set of connection types to enable for this connection (e.g., metrics, logs).",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},

			"sink_org_id": schema.StringAttribute{
				Description: "UUID of the sink (destination) organization.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F-]{36}$`),
						"must be a valid UUID",
					),
				},
			},

			// Computed fields from response
			"id": utils.ResourceIDAttribute(),

			"created_at": schema.StringAttribute{
				Description: "Timestamp when the connection was created (RFC 3339).",
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
			},

			"created_by": schema.StringAttribute{
				Description: "Creator user ID (UUID).",
				Computed:    true,
			},
			"source_org_id": schema.StringAttribute{
				Description: "UUID of the source (current) organization.",
				Computed:    true,
			},
			"source_org_name": schema.StringAttribute{
				Description: "Name of the source organization.",
				Computed:    true,
			},

			"sink_org_name": schema.StringAttribute{
				Description: "Name of the sink (destination) organization.",
				Computed:    true,
			},
		},
	}
}

func (r *OrgConnectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *OrgConnectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data OrgConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	queryParams := datadogV2.ListOrgConnectionsOptionalParameters{}
	queryParams.WithSinkOrgId(data.SinkOrgID.ValueString())
	resp, httpResp, err := r.API.ListOrgConnections(r.Auth, queryParams)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving org connections"))
		return
	}

	if len(resp.GetData()) == 0 {
		response.State.RemoveResource(ctx)
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &data, &resp.GetData()[0])
}

func (r *OrgConnectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data OrgConnectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildCreateRequestBody(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.CreateOrgConnections(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating org connection"))
		return
	}
	if err = utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &data, &resp.Data)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *OrgConnectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data OrgConnectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	connectionID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "connection ID must be a valid UUID"))
		return
	}
	body, diags := r.buildUpdateRequestBody(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.UpdateOrgConnections(r.Auth, connectionID, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating org connection"))
		return
	}

	if err = utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &data, &resp.Data)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *OrgConnectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data OrgConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		return
	}

	httpResp, err := r.API.DeleteOrgConnections(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting org connection"))
	}
}

func (r *OrgConnectionResource) updateState(ctx context.Context, state *OrgConnectionModel, orgConnectionData *datadogV2.OrgConnection) {
	state.ID = types.StringValue(orgConnectionData.GetId().String())
	orgConnectionAttributes := orgConnectionData.GetAttributes()

	// Update State from Attributes of Response
	if connTypes, ok := orgConnectionAttributes.GetConnectionTypesOk(); ok {
		state.ConnectionTypes, _ = types.SetValueFrom(ctx, types.StringType, connTypes)
	} else {
		state.ConnectionTypes, _ = types.SetValue(types.StringType, []attr.Value{})
	}

	if createdAt, ok := orgConnectionAttributes.GetCreatedAtOk(); ok {
		state.CreatedAt = timetypes.NewRFC3339TimeValue(*createdAt)
	} else {
		state.CreatedAt = timetypes.NewRFC3339Null()
	}

	// Update State from Data of Response
	orgConnectionRelationships := orgConnectionData.GetRelationships()
	if sinkOrg, ok := orgConnectionRelationships.GetSinkOrgOk(); ok {
		sinkOrgData := sinkOrg.GetData()
		state.SinkOrgID = types.StringValue(sinkOrgData.GetId())
		state.SinkOrgName = types.StringValue(sinkOrgData.GetName())
	} else {
		state.SinkOrgID = types.StringValue("")
		state.SinkOrgName = types.StringValue("")
	}

	if sourceOrg, ok := orgConnectionRelationships.GetSourceOrgOk(); ok {
		sourceOrgData := sourceOrg.GetData()
		state.SourceOrgID = types.StringValue(sourceOrgData.GetId())
		state.SourceOrgName = types.StringValue(sourceOrgData.GetName())
	} else {
		state.SourceOrgID = types.StringValue("")
		state.SourceOrgName = types.StringValue("")
	}

	if createdBy, ok := orgConnectionRelationships.GetCreatedByOk(); ok {
		createdByData := createdBy.GetData()
		state.CreatedBy = types.StringValue(createdByData.GetId())
	} else {
		state.CreatedBy = types.StringValue("")
	}
}

func (r *OrgConnectionResource) buildCreateRequestBody(ctx context.Context, data *OrgConnectionModel) (*datadogV2.OrgConnectionCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	body := datadogV2.NewOrgConnectionCreateWithDefaults()

	connectionTypes := []datadogV2.OrgConnectionTypeEnum{}
	for _, val := range data.ConnectionTypes.Elements() {
		strVal := val.(types.String)
		enumVal, err := datadogV2.NewOrgConnectionTypeEnumFromValue(strVal.ValueString())
		if err != nil {
			diags.AddError(
				fmt.Sprintf("invalid value found for connection_types: %s", strVal),
				"`connection_types` is invalid; provide a valid value.",
			)
			return nil, diags
		}
		connectionTypes = append(connectionTypes, *enumVal)
	}

	attributes := datadogV2.OrgConnectionCreateAttributes{
		ConnectionTypes: connectionTypes,
	}

	relationshipType := datadogV2.ORGCONNECTIONORGRELATIONSHIPDATATYPE_ORGS
	relationships := datadogV2.OrgConnectionCreateRelationships{
		SinkOrg: datadogV2.OrgConnectionOrgRelationship{
			Data: &datadogV2.OrgConnectionOrgRelationshipData{
				Id:   data.SinkOrgID.ValueStringPointer(),
				Name: data.SinkOrgName.ValueStringPointer(),
				Type: &relationshipType,
			},
		},
	}

	body.SetAttributes(attributes)
	body.SetRelationships(relationships)
	body.SetType(datadogV2.ORGCONNECTIONTYPE_ORG_CONNECTION)
	req := datadogV2.NewOrgConnectionCreateRequest(*body)
	return req, diags
}

func (r *OrgConnectionResource) buildUpdateRequestBody(ctx context.Context, data *OrgConnectionModel) (*datadogV2.OrgConnectionUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	body := datadogV2.NewOrgConnectionUpdateWithDefaults()

	connectionTypes := []datadogV2.OrgConnectionTypeEnum{}
	for _, val := range data.ConnectionTypes.Elements() {
		strVal := val.(types.String)
		enumVal, err := datadogV2.NewOrgConnectionTypeEnumFromValue(strVal.ValueString())
		if err != nil {
			diags.AddError(
				fmt.Sprintf("invalid value found for connection_types: %s", strVal),
				"`connection_types` is invalid; provide a valid value.",
			)
			return nil, diags
		}
		connectionTypes = append(connectionTypes, *enumVal)
	}

	attributes := datadogV2.OrgConnectionUpdateAttributes{
		ConnectionTypes: connectionTypes,
	}

	connectionID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		diags.AddError(
			fmt.Sprintf("invalid value found for connection_id: %s", data.ID),
			"`connection_id` is invalid; provide a valid uuid value.",
		)
		return nil, diags
	}
	body.SetId(connectionID)
	body.SetAttributes(attributes)
	body.SetType(datadogV2.ORGCONNECTIONTYPE_ORG_CONNECTION)
	req := datadogV2.NewOrgConnectionUpdateRequest(*body)
	return req, diags
}
