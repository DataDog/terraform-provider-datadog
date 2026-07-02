package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The `type` attribute is exposed as a human-readable name (avoiding magic
// numbers) and translated to/from the API's numeric field type. The three
// collections below are kept in sync across the eight supported field types.
var incidentUserDefinedFieldTypeNames = map[int32]string{
	1: "dropdown",
	2: "multiselect",
	3: "textbox",
	4: "textarray",
	5: "metrictag",
	6: "autocomplete",
	7: "number",
	8: "datetime",
}

var incidentUserDefinedFieldTypeValues = map[string]datadogV2.IncidentUserDefinedFieldFieldType{
	"dropdown":     1,
	"multiselect":  2,
	"textbox":      3,
	"textarray":    4,
	"metrictag":    5,
	"autocomplete": 6,
	"number":       7,
	"datetime":     8,
}

// incidentUserDefinedFieldTypeNameList lists the valid `type` values in numeric order.
var incidentUserDefinedFieldTypeNameList = []string{
	"dropdown", "multiselect", "textbox", "textarray",
	"metrictag", "autocomplete", "number", "datetime",
}

var (
	_ resource.ResourceWithConfigure   = &incidentUserDefinedFieldResource{}
	_ resource.ResourceWithImportState = &incidentUserDefinedFieldResource{}
)

type incidentUserDefinedFieldResource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentUserDefinedFieldModel struct {
	ID           types.String                              `tfsdk:"id"`
	Name         types.String                              `tfsdk:"name"`
	Type         types.String                              `tfsdk:"type"`
	IncidentType types.String                              `tfsdk:"incident_type"`
	DisplayName  types.String                              `tfsdk:"display_name"`
	Category     types.String                              `tfsdk:"category"`
	Collected    types.String                              `tfsdk:"collected"`
	DefaultValue types.String                              `tfsdk:"default_value"`
	Ordinal      types.String                              `tfsdk:"ordinal"`
	Required     types.Bool                                `tfsdk:"required"`
	TagKey       types.String                              `tfsdk:"tag_key"`
	AttachedTo   types.String                              `tfsdk:"attached_to"`
	Reserved     types.Bool                                `tfsdk:"reserved"`
	Prerequisite types.String                              `tfsdk:"prerequisite"`
	TableID      types.Int64                               `tfsdk:"table_id"`
	Created      types.String                              `tfsdk:"created"`
	Modified     types.String                              `tfsdk:"modified"`
	Deleted      types.String                              `tfsdk:"deleted"`
	Metadata     *incidentUserDefinedFieldMetadataModel    `tfsdk:"metadata"`
	ValidValues  []incidentUserDefinedFieldValidValueModel `tfsdk:"valid_values"`
}

type incidentUserDefinedFieldValidValueModel struct {
	DisplayName      types.String `tfsdk:"display_name"`
	Value            types.String `tfsdk:"value"`
	Description      types.String `tfsdk:"description"`
	ShortDescription types.String `tfsdk:"short_description"`
}

type incidentUserDefinedFieldMetadataModel struct {
	Category         types.String `tfsdk:"category"`
	SearchURL        types.String `tfsdk:"search_url"`
	SearchQueryParam types.String `tfsdk:"search_query_param"`
	SearchLimitParam types.String `tfsdk:"search_limit_param"`
	SearchResultPath types.String `tfsdk:"search_result_path"`
	SearchParams     types.Map    `tfsdk:"search_params"`
}

func NewIncidentUserDefinedFieldResource() resource.Resource {
	return &incidentUserDefinedFieldResource{}
}

func (r *incidentUserDefinedFieldResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "incident_user_defined_field"
}

func (r *incidentUserDefinedFieldResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *incidentUserDefinedFieldResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog incident user-defined field resource. This can be used to create and manage custom fields on Datadog incidents. **Note**: This resource targets an endpoint that is in preview and is subject to change.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the user-defined field.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The unique identifier of the field. Must start with a letter or digit and contain only letters, digits, underscores, or periods. Changing the name forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The data type of the field. Changing the type forces a new resource.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(incidentUserDefinedFieldTypeNameList...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"incident_type": schema.StringAttribute{
				Description: "The ID of the incident type this field is associated with. Changing the incident type forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The human-readable name shown in the UI. Defaults to a formatted version of the name if not provided.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"category": schema.StringAttribute{
				Description: "The section in which the field appears: `what_happened` or `why_it_happened`. When unset, the field appears in the Attributes section.",
				Optional:    true,
			},
			"collected": schema.StringAttribute{
				Description: "The lifecycle stage at which the app prompts users to fill out this field. One of `active`, `stable`, `resolved`, or `completed`. Cannot be set on required fields.",
				Optional:    true,
			},
			"default_value": schema.StringAttribute{
				Description: "The default value for the field. Must be one of the valid values when `valid_values` is set.",
				Optional:    true,
			},
			"ordinal": schema.StringAttribute{
				Description: "A decimal string representing the field's display order in the UI. Assigned by the server when not provided.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"required": schema.BoolAttribute{
				Description: "When true, users must fill out this field on incidents.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"tag_key": schema.StringAttribute{
				Description: "For metric tag-type fields only, the metric tag key that powers the autocomplete options. Changing the tag key forces a new resource.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attached_to": schema.StringAttribute{
				Description: "The resource type this field is attached to. Always `incidents`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reserved": schema.BoolAttribute{
				Description: "When true, this field is reserved for system use and cannot be deleted.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"prerequisite": schema.StringAttribute{
				Description: "Reserved for future use. Always null.",
				Computed:    true,
			},
			"table_id": schema.Int64Attribute{
				Description: "Reserved for internal use. Always 0.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Description: "Timestamp when the field was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified": schema.StringAttribute{
				Description: "Timestamp when the field was last modified.",
				Computed:    true,
			},
			"deleted": schema.StringAttribute{
				Description: "Timestamp when the field was soft-deleted, or null if not deleted.",
				Computed:    true,
			},
			"metadata": schema.SingleNestedAttribute{
				Description: "Metadata for autocomplete-type fields, describing how to populate autocomplete options. Populated by the server for supported fields.",
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"category": schema.StringAttribute{
						Description: "The category of the autocomplete source.",
						Computed:    true,
					},
					"search_url": schema.StringAttribute{
						Description: "The URL used to populate autocomplete options.",
						Computed:    true,
					},
					"search_query_param": schema.StringAttribute{
						Description: "The query parameter used to pass typed input to the search URL.",
						Computed:    true,
					},
					"search_limit_param": schema.StringAttribute{
						Description: "The query parameter used to limit the number of autocomplete results.",
						Computed:    true,
					},
					"search_result_path": schema.StringAttribute{
						Description: "The JSON path to the results in the response body.",
						Computed:    true,
					},
					"search_params": schema.MapAttribute{
						Description: "Additional query parameters to include in the search URL.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"valid_values": schema.ListNestedBlock{
				Description: "The list of allowed values for dropdown, multiselect, and autocomplete fields. Limited to 1000 values.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{
							Description: "The human-readable display name for this value.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "The identifier that is stored when this option is selected.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "A detailed description of the valid value.",
							Optional:    true,
						},
						"short_description": schema.StringAttribute{
							Description: "A short description of the valid value.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *incidentUserDefinedFieldResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan incidentUserDefinedFieldModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes := datadogV2.IncidentUserDefinedFieldAttributesCreateRequest{
		Name: plan.Name.ValueString(),
		Type: incidentUserDefinedFieldTypeValues[plan.Type.ValueString()],
	}

	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		attributes.SetDisplayName(plan.DisplayName.ValueString())
	}
	if !plan.Category.IsNull() {
		attributes.SetCategory(datadogV2.IncidentUserDefinedFieldCategory(plan.Category.ValueString()))
	}
	if !plan.Collected.IsNull() {
		attributes.SetCollected(datadogV2.IncidentUserDefinedFieldCollected(plan.Collected.ValueString()))
	}
	if !plan.DefaultValue.IsNull() {
		attributes.SetDefaultValue(plan.DefaultValue.ValueString())
	}
	if !plan.Ordinal.IsNull() && !plan.Ordinal.IsUnknown() {
		attributes.SetOrdinal(plan.Ordinal.ValueString())
	}
	if !plan.Required.IsNull() && !plan.Required.IsUnknown() {
		attributes.SetRequired(plan.Required.ValueBool())
	}
	if !plan.TagKey.IsNull() {
		attributes.SetTagKey(plan.TagKey.ValueString())
	}
	if validValues := buildValidValues(plan.ValidValues); len(validValues) > 0 {
		attributes.SetValidValues(validValues)
	}

	body := datadogV2.IncidentUserDefinedFieldCreateRequest{
		Data: datadogV2.IncidentUserDefinedFieldCreateData{
			Type:       datadogV2.INCIDENTUSERDEFINEDFIELDTYPE_USER_DEFINED_FIELD,
			Attributes: attributes,
			Relationships: datadogV2.IncidentUserDefinedFieldCreateRelationships{
				IncidentType: datadogV2.RelationshipToIncidentType{
					Data: datadogV2.RelationshipToIncidentTypeData{
						Id:   plan.IncidentType.ValueString(),
						Type: datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
					},
				},
			},
		},
	}

	resp, httpResp, err := r.Api.CreateIncidentUserDefinedField(r.Auth, body)
	if err != nil {
		errorMsg := "Could not create incident user-defined field, unexpected error: " + err.Error()
		if httpResp != nil {
			errorMsg += fmt.Sprintf(" (Status: %d)", httpResp.StatusCode)
		}
		response.Diagnostics.AddError("Error creating incident user-defined field", errorMsg)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating incident user-defined field",
			fmt.Sprintf("Could not create incident user-defined field, status code: %d", httpResp.StatusCode),
		)
		return
	}

	r.updateStateFromResponse(ctx, &plan, &resp, &response.Diagnostics)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *incidentUserDefinedFieldResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state incidentUserDefinedFieldModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetIncidentUserDefinedField(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading incident user-defined field",
			"Could not read incident user-defined field ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.updateStateFromResponse(ctx, &state, &resp, &response.Diagnostics)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentUserDefinedFieldResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan incidentUserDefinedFieldModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	attributes := datadogV2.IncidentUserDefinedFieldAttributesUpdateRequest{}

	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		attributes.SetDisplayName(plan.DisplayName.ValueString())
	}
	if !plan.Category.IsNull() {
		attributes.SetCategory(datadogV2.IncidentUserDefinedFieldCategory(plan.Category.ValueString()))
	} else {
		attributes.SetCategoryNil()
	}
	if !plan.Collected.IsNull() {
		attributes.SetCollected(datadogV2.IncidentUserDefinedFieldCollected(plan.Collected.ValueString()))
	} else {
		attributes.SetCollectedNil()
	}
	if !plan.DefaultValue.IsNull() {
		attributes.SetDefaultValue(plan.DefaultValue.ValueString())
	}
	if !plan.Ordinal.IsNull() && !plan.Ordinal.IsUnknown() {
		attributes.SetOrdinal(plan.Ordinal.ValueString())
	}
	if !plan.Required.IsNull() && !plan.Required.IsUnknown() {
		attributes.SetRequired(plan.Required.ValueBool())
	}
	attributes.SetValidValues(buildValidValues(plan.ValidValues))

	body := datadogV2.IncidentUserDefinedFieldUpdateRequest{
		Data: datadogV2.IncidentUserDefinedFieldUpdateData{
			Id:         id,
			Type:       datadogV2.INCIDENTUSERDEFINEDFIELDTYPE_USER_DEFINED_FIELD,
			Attributes: attributes,
		},
	}

	resp, httpResp, err := r.Api.UpdateIncidentUserDefinedField(r.Auth, id, body)
	if err != nil {
		errorMsg := "Could not update incident user-defined field, unexpected error: " + err.Error()
		if httpResp != nil {
			errorMsg += fmt.Sprintf(" (Status: %d)", httpResp.StatusCode)
		}
		response.Diagnostics.AddError("Error updating incident user-defined field", errorMsg)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating incident user-defined field",
			fmt.Sprintf("Could not update incident user-defined field, status code: %d", httpResp.StatusCode),
		)
		return
	}

	r.updateStateFromResponse(ctx, &plan, &resp, &response.Diagnostics)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *incidentUserDefinedFieldResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state incidentUserDefinedFieldModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteIncidentUserDefinedField(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		errorMsg := "Could not delete incident user-defined field, unexpected error: " + err.Error()
		if httpResp != nil {
			errorMsg += fmt.Sprintf(" (Status: %d)", httpResp.StatusCode)
		}
		response.Diagnostics.AddError("Error deleting incident user-defined field", errorMsg)
		return
	}
}

func (r *incidentUserDefinedFieldResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func buildValidValues(values []incidentUserDefinedFieldValidValueModel) []datadogV2.IncidentUserDefinedFieldValidValue {
	validValues := make([]datadogV2.IncidentUserDefinedFieldValidValue, len(values))
	for i, vv := range values {
		item := datadogV2.IncidentUserDefinedFieldValidValue{
			DisplayName: vv.DisplayName.ValueString(),
			Value:       vv.Value.ValueString(),
		}
		if !vv.Description.IsNull() {
			item.SetDescription(vv.Description.ValueString())
		}
		if !vv.ShortDescription.IsNull() {
			item.SetShortDescription(vv.ShortDescription.ValueString())
		}
		validValues[i] = item
	}
	return validValues
}

func (r *incidentUserDefinedFieldResource) updateStateFromResponse(ctx context.Context, state *incidentUserDefinedFieldModel, resp *datadogV2.IncidentUserDefinedFieldResponse, diags *diag.Diagnostics) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	attributes, ok := data.GetAttributesOk()
	if ok && attributes != nil {
		state.Name = types.StringValue(attributes.GetName())
		state.DisplayName = types.StringValue(attributes.GetDisplayName())
		state.AttachedTo = types.StringValue(attributes.GetAttachedTo())
		state.Required = types.BoolValue(attributes.GetRequired())
		state.Reserved = types.BoolValue(attributes.GetReserved())
		state.TableID = types.Int64Value(attributes.GetTableId())

		if v, ok := attributes.GetTypeOk(); ok && v != nil {
			if name, found := incidentUserDefinedFieldTypeNames[*v]; found {
				state.Type = types.StringValue(name)
			}
		}
		if v, ok := attributes.GetCategoryOk(); ok && v != nil {
			state.Category = types.StringValue(string(*v))
		} else {
			state.Category = types.StringNull()
		}
		if v, ok := attributes.GetCollectedOk(); ok && v != nil {
			state.Collected = types.StringValue(string(*v))
		} else {
			state.Collected = types.StringNull()
		}
		if v, ok := attributes.GetDefaultValueOk(); ok && v != nil {
			state.DefaultValue = types.StringValue(*v)
		} else {
			state.DefaultValue = types.StringNull()
		}
		if v, ok := attributes.GetOrdinalOk(); ok && v != nil {
			state.Ordinal = types.StringValue(*v)
		} else {
			state.Ordinal = types.StringNull()
		}
		if v, ok := attributes.GetTagKeyOk(); ok && v != nil {
			state.TagKey = types.StringValue(*v)
		} else {
			state.TagKey = types.StringNull()
		}
		if v, ok := attributes.GetPrerequisiteOk(); ok && v != nil {
			state.Prerequisite = types.StringValue(*v)
		} else {
			state.Prerequisite = types.StringNull()
		}
		if v, ok := attributes.GetCreatedOk(); ok && v != nil {
			state.Created = types.StringValue(v.Format("2006-01-02T15:04:05Z"))
		}
		if v, ok := attributes.GetModifiedOk(); ok && v != nil {
			state.Modified = types.StringValue(v.Format("2006-01-02T15:04:05Z"))
		} else {
			state.Modified = types.StringNull()
		}
		if v, ok := attributes.GetDeletedOk(); ok && v != nil {
			state.Deleted = types.StringValue(v.Format("2006-01-02T15:04:05Z"))
		} else {
			state.Deleted = types.StringNull()
		}

		if vv, ok := attributes.GetValidValuesOk(); ok && vv != nil {
			state.ValidValues = make([]incidentUserDefinedFieldValidValueModel, len(*vv))
			for i, v := range *vv {
				m := incidentUserDefinedFieldValidValueModel{
					DisplayName: types.StringValue(v.GetDisplayName()),
					Value:       types.StringValue(v.GetValue()),
				}
				if d, ok := v.GetDescriptionOk(); ok && d != nil {
					m.Description = types.StringValue(*d)
				} else {
					m.Description = types.StringNull()
				}
				if sd, ok := v.GetShortDescriptionOk(); ok && sd != nil {
					m.ShortDescription = types.StringValue(*sd)
				} else {
					m.ShortDescription = types.StringNull()
				}
				state.ValidValues[i] = m
			}
		} else {
			state.ValidValues = nil
		}

		if md, ok := attributes.GetMetadataOk(); ok && md != nil {
			metadata := &incidentUserDefinedFieldMetadataModel{
				Category:         types.StringValue(md.GetCategory()),
				SearchURL:        types.StringValue(md.GetSearchUrl()),
				SearchQueryParam: types.StringValue(md.GetSearchQueryParam()),
				SearchLimitParam: types.StringValue(md.GetSearchLimitParam()),
				SearchResultPath: types.StringValue(md.GetSearchResultPath()),
				SearchParams:     types.MapNull(types.StringType),
			}
			if sp, ok := md.GetSearchParamsOk(); ok && sp != nil {
				strMap := make(map[string]string, len(*sp))
				for k, val := range *sp {
					strMap[k] = fmt.Sprintf("%v", val)
				}
				mapVal, d := types.MapValueFrom(ctx, types.StringType, strMap)
				diags.Append(d...)
				metadata.SearchParams = mapVal
			}
			state.Metadata = metadata
		} else {
			state.Metadata = nil
		}
	}

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if incidentType, ok := relationships.GetIncidentTypeOk(); ok && incidentType != nil {
			if incidentTypeData, ok := incidentType.GetDataOk(); ok && incidentTypeData != nil {
				state.IncidentType = types.StringValue(incidentTypeData.GetId())
			}
		}
	}
}
