package fwprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &incidentPostmortemTemplateResource{}
	_ resource.ResourceWithImportState    = &incidentPostmortemTemplateResource{}
	_ resource.ResourceWithValidateConfig = &incidentPostmortemTemplateResource{}
)

type incidentPostmortemTemplateResource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentPostmortemTemplateModel struct {
	ID                 types.String                               `tfsdk:"id"`
	Name               types.String                               `tfsdk:"name"`
	Content            types.String                               `tfsdk:"content"`
	IsDefault          types.Bool                                 `tfsdk:"is_default"`
	Location           types.String                               `tfsdk:"location"`
	IncidentType       types.String                               `tfsdk:"incident_type"`
	Confluence         *incidentPostmortemTemplateConfluenceModel `tfsdk:"confluence_postmortem_settings"`
	GoogleDocs         *incidentPostmortemTemplateGoogleDocsModel `tfsdk:"google_docs_postmortem_settings"`
	Created            types.String                               `tfsdk:"created"`
	Modified           types.String                               `tfsdk:"modified"`
	LastModifiedByUser types.String                               `tfsdk:"last_modified_by_user"`
}

type incidentPostmortemTemplateConfluenceModel struct {
	AccountID types.String `tfsdk:"account_id"`
	SpaceID   types.String `tfsdk:"space_id"`
	ParentID  types.String `tfsdk:"parent_id"`
}

type incidentPostmortemTemplateGoogleDocsModel struct {
	AccountID      types.String `tfsdk:"account_id"`
	ParentFolderID types.String `tfsdk:"parent_folder_id"`
}

func NewIncidentPostmortemTemplateResource() resource.Resource {
	return &incidentPostmortemTemplateResource{}
}

func (r *incidentPostmortemTemplateResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "incident_postmortem_template"
}

func (r *incidentPostmortemTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog incident postmortem template resource. This can be used to create and manage Datadog incident postmortem templates.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the postmortem template.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the template.",
				Required:    true,
			},
			"content": schema.StringAttribute{
				Description: "The templated content of the postmortem, supporting Markdown and incident template variables.",
				Optional:    true,
				Computed:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this template is a default for its incident type. The API stores a timestamp; the effective default for an incident type is the template with the most recent default timestamp.",
				Optional:    true,
				Computed:    true,
			},
			"location": schema.StringAttribute{
				Description: "The location where the postmortem is created and stored. Valid values are: datadog_notebooks, confluence, google_docs. Defaults to datadog_notebooks. The confluence and google_docs locations are gated behind their respective integrations and feature flags.",
				Optional:    true,
				Computed:    true,
			},
			"incident_type": schema.StringAttribute{
				Description: "The ID of the incident type this template is associated with. Immutable after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.StringAttribute{
				Description: "Timestamp when the template was created.",
				Computed:    true,
			},
			"modified": schema.StringAttribute{
				Description: "Timestamp when the template was last modified.",
				Computed:    true,
			},
			"last_modified_by_user": schema.StringAttribute{
				Description: "The ID of the user who last modified the template.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"confluence_postmortem_settings": schema.SingleNestedBlock{
				Description: "Settings for a postmortem template stored in Confluence. Required when location is confluence.",
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						Description: "The ID of the Confluence integration account.",
						Optional:    true,
					},
					"space_id": schema.StringAttribute{
						Description: "The ID of the Confluence space where postmortems are created.",
						Optional:    true,
					},
					"parent_id": schema.StringAttribute{
						Description: "The ID of the parent Confluence page under which postmortems are created.",
						Optional:    true,
					},
				},
			},
			"google_docs_postmortem_settings": schema.SingleNestedBlock{
				Description: "Settings for a postmortem template stored in Google Docs. Required when location is google_docs.",
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						Description: "The ID of the Google Drive integration account.",
						Optional:    true,
					},
					"parent_folder_id": schema.StringAttribute{
						Description: "The ID of the Google Drive folder where postmortems are created.",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (r *incidentPostmortemTemplateResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}
	r.Api = providerData.DatadogApiInstances.GetIncidentsApiV2()
	r.Auth = providerData.Auth
}

func (r *incidentPostmortemTemplateResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var cfg incidentPostmortemTemplateModel
	response.Diagnostics.Append(request.Config.Get(ctx, &cfg)...)
	if response.Diagnostics.HasError() {
		return
	}

	// location is optional and defaults to datadog_notebooks server-side. If it is
	// unknown at plan time we can't validate the pairing yet.
	if cfg.Location.IsUnknown() {
		return
	}
	location := "datadog_notebooks"
	if !cfg.Location.IsNull() {
		location = cfg.Location.ValueString()
	}

	if cfg.Confluence != nil && location != "confluence" {
		response.Diagnostics.AddAttributeError(
			path.Root("confluence_postmortem_settings"),
			"Invalid postmortem template configuration",
			fmt.Sprintf("confluence_postmortem_settings may only be set when location is \"confluence\", got %q.", location),
		)
	}
	if cfg.GoogleDocs != nil && location != "google_docs" {
		response.Diagnostics.AddAttributeError(
			path.Root("google_docs_postmortem_settings"),
			"Invalid postmortem template configuration",
			fmt.Sprintf("google_docs_postmortem_settings may only be set when location is \"google_docs\", got %q.", location),
		)
	}
	if location == "confluence" && cfg.Confluence == nil {
		response.Diagnostics.AddAttributeError(
			path.Root("confluence_postmortem_settings"),
			"Invalid postmortem template configuration",
			"confluence_postmortem_settings is required when location is \"confluence\".",
		)
	}
	if location == "google_docs" && cfg.GoogleDocs == nil {
		response.Diagnostics.AddAttributeError(
			path.Root("google_docs_postmortem_settings"),
			"Invalid postmortem template configuration",
			"google_docs_postmortem_settings is required when location is \"google_docs\".",
		)
	}
}

func (r *incidentPostmortemTemplateResource) buildAttributes(plan *incidentPostmortemTemplateModel) datadogV2.PostmortemTemplateAttributesRequest {
	attributes := datadogV2.PostmortemTemplateAttributesRequest{
		Name: plan.Name.ValueString(),
	}

	if !plan.Content.IsNull() && !plan.Content.IsUnknown() {
		content := plan.Content.ValueString()
		attributes.Content = &content
	}

	if !plan.IsDefault.IsNull() && !plan.IsDefault.IsUnknown() {
		if plan.IsDefault.ValueBool() {
			attributes.SetIsDefault(time.Now().UTC())
		} else {
			attributes.SetIsDefaultNil()
		}
	}

	if !plan.Location.IsNull() && !plan.Location.IsUnknown() {
		location := datadogV2.PostmortemTemplateLocation(plan.Location.ValueString())
		attributes.Location = &location
	}

	if plan.Confluence != nil {
		settings := datadogV2.NewConfluencePostmortemSettings(
			plan.Confluence.AccountID.ValueString(),
			plan.Confluence.SpaceID.ValueString(),
		)
		if !plan.Confluence.ParentID.IsNull() && !plan.Confluence.ParentID.IsUnknown() {
			settings.SetParentId(plan.Confluence.ParentID.ValueString())
		}
		attributes.ConfluencePostmortemSettings = settings
	}

	if plan.GoogleDocs != nil {
		attributes.GoogleDocsPostmortemSettings = datadogV2.NewGoogleDocsPostmortemSettings(
			plan.GoogleDocs.AccountID.ValueString(),
			plan.GoogleDocs.ParentFolderID.ValueString(),
		)
	}

	return attributes
}

func (r *incidentPostmortemTemplateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan incidentPostmortemTemplateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	incidentTypeID, err := uuid.Parse(plan.IncidentType.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error parsing incident type ID", err.Error())
		return
	}

	body := datadogV2.PostmortemTemplateRequest{
		Data: datadogV2.PostmortemTemplateDataRequest{
			Type:       datadogV2.POSTMORTEMTEMPLATETYPE_POSTMORTEM_TEMPLATES,
			Attributes: r.buildAttributes(&plan),
			Relationships: &datadogV2.PostmortemTemplateCreateRelationships{
				IncidentType: &datadogV2.PostmortemTemplateIncidentTypeRelationship{
					Data: datadogV2.PostmortemTemplateIncidentTypeRelationshipData{
						Id:   incidentTypeID,
						Type: "incident_types",
					},
				},
			},
		},
	}

	resp, httpResp, err := r.Api.CreateIncidentPostmortemTemplate(r.Auth, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating incident postmortem template",
			fmt.Sprintf("Could not create incident postmortem template, unexpected error: %s. HTTP Response: %v", err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating incident postmortem template",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	var state incidentPostmortemTemplateModel
	r.updateStateFromResponse(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentPostmortemTemplateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state incidentPostmortemTemplateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetIncidentPostmortemTemplate(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading incident postmortem template",
			"Could not read incident postmortem template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.updateStateFromResponse(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentPostmortemTemplateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan incidentPostmortemTemplateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	// incident_type is immutable (ForceNew) and is not sent on update.
	body := datadogV2.PostmortemTemplateRequest{
		Data: datadogV2.PostmortemTemplateDataRequest{
			Id:         &id,
			Type:       datadogV2.POSTMORTEMTEMPLATETYPE_POSTMORTEM_TEMPLATES,
			Attributes: r.buildAttributes(&plan),
		},
	}

	resp, httpResp, err := r.Api.UpdateIncidentPostmortemTemplate(r.Auth, id, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating incident postmortem template",
			fmt.Sprintf("Could not update incident postmortem template ID %s, unexpected error: %s. HTTP Response: %v", id, err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating incident postmortem template",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	r.updateStateFromResponse(&plan, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *incidentPostmortemTemplateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state incidentPostmortemTemplateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteIncidentPostmortemTemplate(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting incident postmortem template",
			"Could not delete incident postmortem template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *incidentPostmortemTemplateResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *incidentPostmortemTemplateResource) updateStateFromResponse(state *incidentPostmortemTemplateModel, resp *datadogV2.PostmortemTemplateResponse) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		state.Name = types.StringValue(attributes.GetName())
		state.Content = types.StringValue(attributes.GetContent())

		if isDefault, isDefaultOk := attributes.GetIsDefaultOk(); isDefaultOk && isDefault != nil {
			state.IsDefault = types.BoolValue(true)
		} else {
			state.IsDefault = types.BoolValue(false)
		}

		state.Location = types.StringValue(string(attributes.GetLocation()))
		state.Created = types.StringValue(attributes.GetCreatedAt().Format("2006-01-02T15:04:05Z"))
		state.Modified = types.StringValue(attributes.GetModifiedAt().Format("2006-01-02T15:04:05Z"))

		if confluence, ok := attributes.GetConfluencePostmortemSettingsOk(); ok && confluence != nil {
			model := &incidentPostmortemTemplateConfluenceModel{
				AccountID: types.StringValue(confluence.GetAccountId()),
				SpaceID:   types.StringValue(confluence.GetSpaceId()),
			}
			if parentID, parentOk := confluence.GetParentIdOk(); parentOk && parentID != nil {
				model.ParentID = types.StringValue(*parentID)
			} else {
				model.ParentID = types.StringNull()
			}
			state.Confluence = model
		}

		if googleDocs, ok := attributes.GetGoogleDocsPostmortemSettingsOk(); ok && googleDocs != nil {
			state.GoogleDocs = &incidentPostmortemTemplateGoogleDocsModel{
				AccountID:      types.StringValue(googleDocs.GetAccountId()),
				ParentFolderID: types.StringValue(googleDocs.GetParentFolderId()),
			}
		}
	}

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if incidentType, ok := relationships.GetIncidentTypeOk(); ok && incidentType != nil {
			if incidentTypeData, ok := incidentType.GetDataOk(); ok && incidentTypeData != nil {
				state.IncidentType = types.StringValue(incidentTypeData.GetId().String())
			}
		}
		if lastModifiedBy, ok := relationships.GetLastModifiedByUserOk(); ok && lastModifiedBy != nil {
			if userData, ok := lastModifiedBy.GetDataOk(); ok && userData != nil {
				state.LastModifiedByUser = types.StringValue(userData.GetId().String())
			}
		}
	}
}
