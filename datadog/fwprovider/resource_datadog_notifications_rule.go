package fwprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationsRulesResource struct{}

type notificationsRuleModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Tags        types.List   `tfsdk:"tags"`
	Query       types.String `tfsdk:"query"`
	Sources     types.List   `tfsdk:"sources"`
	Source      types.String `tfsdk:"source"`
	Targets     types.List   `tfsdk:"targets"`
	RuleTargets types.List   `tfsdk:"rule_targets"`
	Active      types.Bool   `tfsdk:"active"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ModifiedAt  types.String `tfsdk:"modified_at"`
}

func (r *NotificationsRulesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "notifications_rule"
}

func (r *NotificationsRulesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Datadog Notification Rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the notification rule.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the rule.",
				Required:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with this rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"query": schema.StringAttribute{
				Description: "Rule query.",
				Optional:    true,
			},
			"sources": schema.ListAttribute{
				Description: "List of sources.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"source": schema.StringAttribute{
				Description: "Single source.",
				Optional:    true,
			},
			"targets": schema.ListAttribute{
				Description: "List of handle targets.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"rule_targets": schema.ListAttribute{
				Description: "Rule targets configurations.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the rule is active.",
				Required:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When this rule was created.",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "When this rule was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *NotificationsRulesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *NotificationsRulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationsRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateNotificationsRule(plan); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "notification_rule",
			"attributes": map[string]interface{}{
				"name":         plan.Name.ValueString(),
				"tags":         expandList(plan.Tags),
				"query":        plan.Query.ValueString(),
				"sources":      expandList(plan.Sources),
				"source":       plan.Source.ValueString(),
				"targets":      expandList(plan.Targets),
				"rule_targets": expandList(plan.RuleTargets),
				"enabled":      plan.Active.ValueBool(),
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	response, err := upsertNotificationsRuleHTTP("POST", "/notification_rules", payloadBytes)
	if err != nil {
		resp.Diagnostics.AddError("Error creating notification rule", err.Error())
		return
	}

	plan.ID = types.StringValue(response.Data.ID)
	plan.CreatedAt = types.StringValue(response.Data.Attributes.CreationDate)
	plan.ModifiedAt = types.StringValue(response.Data.Attributes.UpdateDate)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NotificationsRulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationsRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("/notification_rules/%s", state.ID.ValueString())
	response, err := readNotificationsRuleHTTP(url)
	if err != nil {
		resp.Diagnostics.AddError("Error reading notification rule", err.Error())
		return
	}

	state.Name = types.StringValue(response.Data.Attributes.Name)
	state.CreatedAt = types.StringValue(response.Data.Attributes.CreationDate)
	state.ModifiedAt = types.StringValue(response.Data.Attributes.UpdateDate)

	tags, _ := types.ListValue(types.StringType, stringSliceToTerraformValues(response.Data.Attributes.Targets))
	state.Tags = tags
	state.Active = types.BoolValue(response.Data.Attributes.Enabled)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *NotificationsRulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationsRuleModel
	var state notificationsRuleModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateNotificationsRule(plan); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	url := fmt.Sprintf("/notification_rules/%s", state.ID.ValueString())
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "notification_rule",
			"attributes": map[string]interface{}{
				"name":         plan.Name.ValueString(),
				"tags":         expandList(plan.Tags),
				"query":        plan.Query.ValueString(),
				"sources":      expandList(plan.Sources),
				"source":       plan.Source.ValueString(),
				"targets":      expandList(plan.Targets),
				"rule_targets": expandList(plan.RuleTargets),
				"enabled":      plan.Active.ValueBool(),
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	response, err := upsertNotificationsRuleHTTP("PATCH", url, payloadBytes)
	if err != nil {
		resp.Diagnostics.AddError("Error updating notification rule", err.Error())
		return
	}

	plan.ModifiedAt = types.StringValue(response.Data.Attributes.UpdateDate)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NotificationsRulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationsRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("/notification_rules/%s", state.ID.ValueString())
	err := deleteNotificationsRuleHTTP(url)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting notification rule", err.Error())
		return
	}
}

func (r *NotificationsRulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func validateNotificationsRule(plan notificationsRuleModel) error {
	if plan.Name.ValueString() == "" {
		return fmt.Errorf("rule name must not be empty")
	}
	if plan.Name.ValueString() != "" && len(plan.Name.ValueString()) > 255 {
		return fmt.Errorf("rule name must not exceed 255 characters")
	}
	if plan.Query.ValueString() != "" && len(plan.Query.ValueString()) > 2000 {
		return fmt.Errorf("rule query must not exceed 2000 characters")
	}
	if plan.Query.ValueString() == "" {
		return fmt.Errorf("rule query must not be empty")
	}
	if plan.Tags.IsUnknown() == false && plan.Tags.IsNull() == false {
		if len(expandList(plan.Tags)) > 20 {
			return fmt.Errorf("number of tags must be <= 20")
		}
	}
	if plan.Sources.IsNull() && plan.Source.IsNull() {
		return fmt.Errorf("there must be at least one source")
	}
	if !plan.Targets.IsNull() && !plan.RuleTargets.IsNull() {
		if len(expandList(plan.Targets)) > 0 && len(expandList(plan.RuleTargets)) > 0 {
			return fmt.Errorf("either targets or rule_targets can be used but not both")
		}
	}
	return nil
}

type notificationRuleAPIResponse struct {
	Data struct {
		ID         string                        `json:"id"`
		Type       string                        `json:"type"`
		Attributes notificationRuleAPIAttributes `json:"attributes"`
	} `json:"data"`
}

type notificationRuleAPIAttributes struct {
	Name         string   `json:"name"`
	Targets      []string `json:"targets"`
	Enabled      bool     `json:"enabled"`
	CreationDate string   `json:"creation_date"`
	UpdateDate   string   `json:"update_date"`
}

func upsertNotificationsRuleHTTP(method, url string, payload []byte) (notificationRuleAPIResponse, error) {
	req, err := http.NewRequest(method, "https://dd.datad0g.com/api/v2"+url, bytes.NewReader(payload))
	if err != nil {
		return notificationRuleAPIResponse{}, err
	}
	addHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return notificationRuleAPIResponse{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return notificationRuleAPIResponse{}, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return notificationRuleAPIResponse{}, fmt.Errorf("status %s: %s", resp.Status, string(body))
	}
	var response notificationRuleAPIResponse
	json.Unmarshal(body, &response)
	return response, nil
}

func readNotificationsRuleHTTP(url string) (notificationRuleAPIResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "https://dd.datad0g.com/api/v2"+url, nil)
	if err != nil {
		return notificationRuleAPIResponse{}, err
	}
	addHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return notificationRuleAPIResponse{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return notificationRuleAPIResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return notificationRuleAPIResponse{}, fmt.Errorf("status %s: %s", resp.Status, string(body))
	}
	var r notificationRuleAPIResponse
	json.Unmarshal(body, &r)
	return r, nil
}

func deleteNotificationsRuleHTTP(url string) error {
	req, err := http.NewRequest(http.MethodDelete, "https://dd.datad0g.com/api/v2"+url, nil)
	if err != nil {
		return err
	}
	addHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %s: %s", resp.Status, string(body))
	}
	return nil
}

func NewNotificationsRulesResource() resource.Resource {
	return &NotificationsRulesResource{}
}
