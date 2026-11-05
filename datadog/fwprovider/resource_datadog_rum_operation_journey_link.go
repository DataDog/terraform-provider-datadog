package fwprovider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const (
	rumOperationJourneyLinkIdDelimiter = ":"
	rumOperationJourneyLinkPageLimit   = 200
)

var (
	_ resource.ResourceWithConfigure   = &rumOperationJourneyLinkResource{}
	_ resource.ResourceWithImportState = &rumOperationJourneyLinkResource{}
)

type rumOperationJourneyLinkResource struct {
	Api  *datadogV2.RUMOperationsApi
	Auth context.Context
}

type rumOperationJourneyLinkModel struct {
	ID          types.String `tfsdk:"id"`
	OperationID types.String `tfsdk:"operation_id"`
	JourneyID   types.String `tfsdk:"journey_id"`
}

func NewRumOperationJourneyLinkResource() resource.Resource {
	return &rumOperationJourneyLinkResource{}
}

func (r *rumOperationJourneyLinkResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRUMOperationsApiV2()
	r.Auth = providerData.Auth
}

func (r *rumOperationJourneyLinkResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "rum_operation_journey_link"
}

func (r *rumOperationJourneyLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RUM Operation journey link resource. This links a RUM operation to a journey. " +
			"Links managed by this resource are always confirmed; an existing draft or rejected link between the same " +
			"operation and journey is confirmed on creation.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"operation_id": schema.StringAttribute{
				Description: "ID of the RUM operation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"journey_id": schema.StringAttribute{
				Description: "ID of the journey.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *rumOperationJourneyLinkResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	operationID, journeyID, err := parseRumOperationJourneyLinkId(request.ID)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), "")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("operation_id"), operationID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("journey_id"), journeyID)...)
}

func (r *rumOperationJourneyLinkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state rumOperationJourneyLinkModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	confirmed, err := r.isLinkConfirmed(state.OperationID.ValueString(), state.JourneyID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RUM operation journey link"))
		return
	}
	// A link rejected or deleted outside of Terraform is gone; the next apply re-confirms it.
	if !confirmed {
		response.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(rumOperationJourneyLinkId(state.OperationID.ValueString(), state.JourneyID.ValueString()))
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumOperationJourneyLinkResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state rumOperationJourneyLinkModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	operationID := state.OperationID.ValueString()
	journeyID := state.JourneyID.ValueString()

	attributes := datadogV2.NewRUMOperationStrongLinkCreateRequestAttributes(journeyID)
	attributes.SetOperationId(operationID)
	attributes.SetStatus(datadogV2.RUMOPERATIONSTRONGLINKSTATUS_CONFIRMED)
	body := datadogV2.NewRUMOperationStrongLinkCreateRequest(
		*datadogV2.NewRUMOperationStrongLinkCreateRequestData(*attributes, datadogV2.RUMOPERATIONSTRONGLINKTYPE_STRONG_LINKS))

	_, httpResp, err := r.Api.CreateRUMOperationStrongLink(r.Auth, *body)
	if err != nil {
		// A draft or rejected link already holds the operation/journey pair: confirm it instead.
		if httpResp == nil || httpResp.StatusCode != 409 {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating RUM operation journey link"))
			return
		}
		update := datadogV2.NewRUMOperationStrongLinkUpdateRequest(*datadogV2.NewRUMOperationStrongLinkUpdateRequestData(
			*datadogV2.NewRUMOperationStrongLinkUpdateRequestAttributes(datadogV2.RUMOPERATIONSTRONGLINKUPDATESTATUS_CONFIRMED),
			datadogV2.RUMOPERATIONSTRONGLINKTYPE_STRONG_LINKS))
		if _, _, err := r.Api.UpdateRUMOperationStrongLink(r.Auth, operationID, journeyID, *update); err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error confirming existing RUM operation journey link"))
			return
		}
	}

	state.ID = types.StringValue(rumOperationJourneyLinkId(operationID, journeyID))
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Update is never called with a change: every attribute requires replacement.
func (r *rumOperationJourneyLinkResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state rumOperationJourneyLinkModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumOperationJourneyLinkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state rumOperationJourneyLinkModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteRUMOperationStrongLink(r.Auth, state.OperationID.ValueString(), state.JourneyID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting RUM operation journey link"))
	}
}

// isLinkConfirmed pages through the operation's active links. The single-link lookup is not used because
// it also returns deleted links and the response does not say whether a link is deleted.
func (r *rumOperationJourneyLinkResource) isLinkConfirmed(operationID, journeyID string) (bool, error) {
	for offset := int64(0); ; offset += rumOperationJourneyLinkPageLimit {
		params := datadogV2.NewListRUMOperationStrongLinksOptionalParameters().
			WithOperationId(operationID).
			WithPageOffset(offset).
			WithPageLimit(rumOperationJourneyLinkPageLimit)
		resp, _, err := r.Api.ListRUMOperationStrongLinks(r.Auth, *params)
		if err != nil {
			return false, err
		}
		if err := utils.CheckForUnparsed(resp); err != nil {
			return false, err
		}
		for _, link := range resp.GetData() {
			attributes := link.GetAttributes()
			if attributes.GetFeatureId() == journeyID {
				return attributes.GetStatus() == datadogV2.RUMOPERATIONSTRONGLINKSTATUS_CONFIRMED, nil
			}
		}
		if len(resp.GetData()) < rumOperationJourneyLinkPageLimit {
			return false, nil
		}
		if total, ok := resp.Meta.GetTotalOk(); ok && offset+rumOperationJourneyLinkPageLimit >= *total {
			return false, nil
		}
	}
}

func rumOperationJourneyLinkId(operationID, journeyID string) string {
	return fmt.Sprintf("%s%s%s", operationID, rumOperationJourneyLinkIdDelimiter, journeyID)
}

func parseRumOperationJourneyLinkId(id string) (operationID string, journeyID string, err error) {
	result := strings.SplitN(id, rumOperationJourneyLinkIdDelimiter, 2)
	if len(result) != 2 || result[0] == "" || result[1] == "" {
		return "", "", errors.New("error parsing id into operation_id and journey_id, expected <operation_id>:<journey_id>")
	}
	return result[0], result[1], nil
}
