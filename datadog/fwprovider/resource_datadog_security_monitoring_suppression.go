package fwprovider

import (
	"context"
	"sync"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &securityMonitoringSuppressionResource{}
	_ resource.ResourceWithImportState = &securityMonitoringSuppressionResource{}
)

type securityMonitoringSuppressionModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	StartDate          types.String `tfsdk:"start_date"`
	ExpirationDate     types.String `tfsdk:"expiration_date"`
	RuleQuery          types.String `tfsdk:"rule_query"`
	SuppressionQuery   types.String `tfsdk:"suppression_query"`
	DataExclusionQuery types.String `tfsdk:"data_exclusion_query"`
}

type securityMonitoringSuppressionResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

var suppressionWriteMutex = sync.Mutex{}

func NewSecurityMonitoringSuppressionResource() resource.Resource {
	return &securityMonitoringSuppressionResource{}
}

func (r *securityMonitoringSuppressionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_suppression"
}

func (r *securityMonitoringSuppressionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (r *securityMonitoringSuppressionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Suppression API resource. It can be used to create and manage Datadog security monitoring suppression rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the suppression rule.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description for the suppression rule.",
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the suppression rule is enabled.",
			},
			"start_date": schema.StringAttribute{
				Optional:    true,
				Description: "A RFC3339 timestamp giving a start date for the suppression rule. Before this date, it doesn't suppress signals.",
			},
			"expiration_date": schema.StringAttribute{
				Optional:    true,
				Description: "A RFC3339 timestamp giving an expiration date for the suppression rule. After this date, it won't suppress signals anymore.",
			},
			"rule_query": schema.StringAttribute{
				Required:    true,
				Description: "The rule query of the suppression rule, with the same syntax as the search bar for detection rules.",
			},
			"suppression_query": schema.StringAttribute{
				Optional:    true,
				Description: "The suppression query of the suppression rule. If a signal matches this query, it is suppressed and is not triggered. It uses the same syntax as the queries to search signals in the Signals Explorer.",
			},
			"data_exclusion_query": schema.StringAttribute{
				Optional:    true,
				Description: "An exclusion query on the input data of the security rules, which could be logs, Agent events, or other types of data based on the security rule. Events matching this query are ignored by any detection rules referenced in the suppression rule.",
			},
		},
	}
}

func (r *securityMonitoringSuppressionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *securityMonitoringSuppressionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state securityMonitoringSuppressionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	suppressionPayload, err := r.buildCreateSecurityMonitoringSuppressionPayload(&state)

	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
		return
	}

	suppressionWriteMutex.Lock()
	defer suppressionWriteMutex.Unlock()

	res, _, err := r.api.CreateSecurityMonitoringSuppression(r.auth, *suppressionPayload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating security monitoring suppression"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityMonitoringSuppressionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityMonitoringSuppressionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	suppressionId := state.Id.ValueString()

	res, httpResponse, err := r.api.GetSecurityMonitoringSuppression(r.auth, suppressionId)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching security monitoring suppression"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityMonitoringSuppressionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state securityMonitoringSuppressionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	suppressionPayload, err := r.buildUpdateSecurityMonitoringSuppressionPayload(&state)

	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
		return
	}

	suppressionWriteMutex.Lock()
	defer suppressionWriteMutex.Unlock()

	res, _, err := r.api.UpdateSecurityMonitoringSuppression(r.auth, state.Id.ValueString(), *suppressionPayload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating security monitoring suppression"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityMonitoringSuppressionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state securityMonitoringSuppressionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	suppressionWriteMutex.Lock()
	defer suppressionWriteMutex.Unlock()

	httpResp, err := r.api.DeleteSecurityMonitoringSuppression(r.auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting suppression"))
		return
	}
}

func (r *securityMonitoringSuppressionResource) buildCreateSecurityMonitoringSuppressionPayload(state *securityMonitoringSuppressionModel) (*datadogV2.SecurityMonitoringSuppressionCreateRequest, error) {
	name, description, enabled, startDate, expirationDate, ruleQuery, suppressionQuery, dataExclusionQuery, err := r.extractSuppressionAttributesFromResource(state)

	if err != nil {
		return nil, err
	}

	attributes := datadogV2.NewSecurityMonitoringSuppressionCreateAttributes(enabled, name, ruleQuery)
	attributes.SuppressionQuery = suppressionQuery
	attributes.DataExclusionQuery = dataExclusionQuery
	attributes.Description = description
	attributes.StartDate = startDate
	attributes.ExpirationDate = expirationDate

	data := datadogV2.NewSecurityMonitoringSuppressionCreateData(*attributes, datadogV2.SECURITYMONITORINGSUPPRESSIONTYPE_SUPPRESSIONS)
	return datadogV2.NewSecurityMonitoringSuppressionCreateRequest(*data), nil
}

func (r *securityMonitoringSuppressionResource) buildUpdateSecurityMonitoringSuppressionPayload(state *securityMonitoringSuppressionModel) (*datadogV2.SecurityMonitoringSuppressionUpdateRequest, error) {
	name, description, enabled, startDate, expirationDate, ruleQuery, suppressionQuery, dataExclusionQuery, err := r.extractSuppressionAttributesFromResource(state)

	if err != nil {
		return nil, err
	}

	attributes := datadogV2.NewSecurityMonitoringSuppressionUpdateAttributes()
	attributes.SetName(name)
	attributes.Description = description
	attributes.SetEnabled(enabled)
	if startDate != nil {
		attributes.SetStartDate(*startDate)
	} else {
		attributes.SetStartDateNil()
	}
	if expirationDate != nil {
		attributes.SetExpirationDate(*expirationDate)
	} else {
		attributes.SetExpirationDateNil()
	}
	attributes.SetRuleQuery(ruleQuery)

	if suppressionQuery != nil {
		attributes.SuppressionQuery = suppressionQuery
	} else {
		attributes.SetSuppressionQuery("")
	}

	if dataExclusionQuery != nil {
		attributes.DataExclusionQuery = dataExclusionQuery
	} else {
		attributes.SetDataExclusionQuery("")
	}

	data := datadogV2.NewSecurityMonitoringSuppressionUpdateData(*attributes, datadogV2.SECURITYMONITORINGSUPPRESSIONTYPE_SUPPRESSIONS)
	return datadogV2.NewSecurityMonitoringSuppressionUpdateRequest(*data), nil
}

func (r *securityMonitoringSuppressionResource) extractSuppressionAttributesFromResource(state *securityMonitoringSuppressionModel) (string, *string, bool, *int64, *int64, string, *string, *string, error) {
	// Mandatory fields

	name := state.Name.ValueString()
	enabled := state.Enabled.ValueBool()
	ruleQuery := state.RuleQuery.ValueString()

	// Optional fields

	description := state.Description.ValueStringPointer()
	suppressionQuery := state.SuppressionQuery.ValueStringPointer()
	dataExclusionQuery := state.DataExclusionQuery.ValueStringPointer()

	var startDate *int64

	if tfStartDate := state.StartDate.ValueStringPointer(); tfStartDate != nil {
		startDateTime, err := time.Parse(time.RFC3339, *tfStartDate)

		if err != nil {
			return "", nil, false, nil, nil, "", nil, nil, err
		}

		startDateTimestamp := startDateTime.UnixMilli()
		startDate = &startDateTimestamp

	}

	var expirationDate *int64

	if tfExpirationDate := state.ExpirationDate.ValueStringPointer(); tfExpirationDate != nil {
		expirationDateTime, err := time.Parse(time.RFC3339, *tfExpirationDate)

		if err != nil {
			return "", nil, false, nil, nil, "", nil, nil, err
		}

		expirationDateTimestamp := expirationDateTime.UnixMilli()
		expirationDate = &expirationDateTimestamp

	}

	return name, description, enabled, startDate, expirationDate, ruleQuery, suppressionQuery, dataExclusionQuery, nil
}

func (r *securityMonitoringSuppressionResource) updateStateFromResponse(ctx context.Context, state *securityMonitoringSuppressionModel, res *datadogV2.SecurityMonitoringSuppressionResponse) {
	state.Id = types.StringValue(res.Data.GetId())

	attributes := res.Data.Attributes

	state.Name = types.StringValue(attributes.GetName())

	// Only update the state if the description is not empty, or if it's not null in the plan
	// If the description is null in the TF config, it is omitted from the API call
	// The API returns an empty string, which, if put in the state, would result in a mismatch between state and config
	if description := attributes.GetDescription(); description != "" || !state.Description.IsNull() {
		state.Description = types.StringValue(description)
	}

	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.RuleQuery = types.StringValue(attributes.GetRuleQuery())
	state.SuppressionQuery = types.StringValue(attributes.GetSuppressionQuery())

	if suppressionQuery := attributes.GetSuppressionQuery(); suppressionQuery != "" {
		state.SuppressionQuery = types.StringValue(suppressionQuery)
	} else {
		state.SuppressionQuery = types.StringNull()
	}

	if dataExclusionQuery := attributes.GetDataExclusionQuery(); dataExclusionQuery != "" {
		state.DataExclusionQuery = types.StringValue(dataExclusionQuery)
	} else {
		state.DataExclusionQuery = types.StringNull()
	}

	// For the StartDate and the ExpirationDate
	// The API only requires a millisecond timestamp, it does not care about timezones.
	// If the timestamp string written by the user has the same millisecond value as the one returned by the API,
	// we keep the user-defined one in the state.
	if attributes.StartDate != nil {
		responseStartDate := time.UnixMilli(*attributes.StartDate).UTC()
		startDate := responseStartDate.Format(time.RFC3339)

		if userStartDateStr := state.StartDate.ValueString(); userStartDateStr != "" {
			if userStartDate, err := time.Parse(time.RFC3339, userStartDateStr); err == nil {
				if userStartDate.UnixMilli() == responseStartDate.UnixMilli() {
					startDate = userStartDateStr
				}
			}
		}
		state.StartDate = types.StringValue(startDate)
	}

	if attributes.ExpirationDate != nil {
		responseExpirationDate := time.UnixMilli(*attributes.ExpirationDate).UTC()
		expirationDate := responseExpirationDate.Format(time.RFC3339)

		if userExpirationDateStr := state.ExpirationDate.ValueString(); userExpirationDateStr != "" {
			if userExpirationDate, err := time.Parse(time.RFC3339, userExpirationDateStr); err == nil {
				if userExpirationDate.UnixMilli() == responseExpirationDate.UnixMilli() {
					expirationDate = userExpirationDateStr
				}
			}
		}
		state.ExpirationDate = types.StringValue(expirationDate)
	}
}
