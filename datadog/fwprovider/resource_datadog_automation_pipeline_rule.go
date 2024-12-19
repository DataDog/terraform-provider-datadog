package fwprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// const apiBaseURL = "http://localhost:8000/vulnerabilities/pipelines"
const apiBaseURL = "https://dd.datad0g.com/api/v2/security/vulnerabilities/pipelines"
const DDAPIKEY = ""
const DDAPPKEY = ""

type automationPipelineRuleResource struct{}

type automationPipelineRuleModel struct {
	Name    types.String       `tfsdk:"name"`
	Enabled types.Bool         `tfsdk:"enabled"`
	Inbox   *inboxBlockModel   `tfsdk:"inbox"`
	Mute    *muteBlockModel    `tfsdk:"mute"`
	DueDate *dueDateBlockModel `tfsdk:"due_date"`
	ID      types.String       `tfsdk:"id"`
}

type actionBlock interface {
	IsNull() bool
	ToPayload() map[string]interface{}
	ValidateRequired() error
}

type ruleBlock interface {
	IsNull() bool
	ToPayload() map[string]interface{}
	ValidateRequired() error
}

type inboxBlockModel struct {
	Rule   *inboxRuleBlockModel   `tfsdk:"rule"`
	Action *inboxActionBlockModel `tfsdk:"action"`
}

type muteBlockModel struct {
	Rule   *muteRuleBlockModel   `tfsdk:"rule"`
	Action *muteActionBlockModel `tfsdk:"action"`
}

type dueDateBlockModel struct {
	Rule   *dueDateRuleBlockModel   `tfsdk:"rule"`
	Action *dueDateActionBlockModel `tfsdk:"action"`
}

type inboxRuleBlockModel struct {
	IssueType  types.String `tfsdk:"issue_type"`
	RuleTypes  types.List   `tfsdk:"rule_types"`
	RuleIds    types.List   `tfsdk:"rule_ids"`
	Severities types.List   `tfsdk:"severities"`
	Query      types.String `tfsdk:"query"`
}

func (r *inboxRuleBlockModel) IsNull() bool {
	return r == nil || (r.IssueType.IsNull() && r.RuleTypes.IsNull() && r.RuleIds.IsNull() && r.Severities.IsNull() && r.Query.IsNull())
}

func (r *inboxRuleBlockModel) ToPayload() map[string]interface{} {
	payload := map[string]interface{}{}
	if !r.IssueType.IsNull() {
		payload["issue_type"] = r.IssueType.ValueString()
	}
	if !r.RuleTypes.IsNull() {
		payload["rule_types"] = expandList(r.RuleTypes)
	}
	if !r.RuleIds.IsNull() {
		payload["rule_ids"] = expandList(r.RuleIds)
	}
	if !r.Severities.IsNull() {
		payload["severities"] = expandList(r.Severities)
	}
	if !r.Query.IsNull() {
		payload["query"] = r.Query.ValueString()
	}
	return payload
}

func (r *inboxRuleBlockModel) ValidateRequired() error {
	if r.IssueType.IsNull() {
		return fmt.Errorf("issue_type is required for inbox rule")
	}
	if r.RuleTypes.IsNull() {
		return fmt.Errorf("rule_types is required for inbox rule")
	}
	return nil
}

type muteRuleBlockModel struct {
	IssueType  types.String `tfsdk:"issue_type"`
	RuleTypes  types.List   `tfsdk:"rule_types"`
	RuleIds    types.List   `tfsdk:"rule_ids"`
	Severities types.List   `tfsdk:"severities"`
	Query      types.String `tfsdk:"query"`
}

func (r *muteRuleBlockModel) IsNull() bool {
	return r == nil || (r.IssueType.IsNull() && r.RuleTypes.IsNull() && r.RuleIds.IsNull() && r.Severities.IsNull() && r.Query.IsNull())
}

func (r *muteRuleBlockModel) ToPayload() map[string]interface{} {
	payload := map[string]interface{}{}
	if !r.IssueType.IsNull() {
		payload["issue_type"] = r.IssueType.ValueString()
	}
	if !r.RuleTypes.IsNull() {
		payload["rule_types"] = expandList(r.RuleTypes)
	}
	if !r.RuleIds.IsNull() {
		payload["rule_ids"] = expandList(r.RuleIds)
	}
	if !r.Severities.IsNull() {
		payload["severities"] = expandList(r.Severities)
	}
	if !r.Query.IsNull() {
		payload["query"] = r.Query.ValueString()
	}
	return payload
}

func (r *muteRuleBlockModel) ValidateRequired() error {
	if r.IssueType.IsNull() {
		return fmt.Errorf("issue_type is required for mute rule")
	}
	if r.RuleTypes.IsNull() {
		return fmt.Errorf("rule_types is required for mute rule")
	}
	return nil
}

type dueDateRuleBlockModel struct {
	IssueType  types.String `tfsdk:"issue_type"`
	RuleTypes  types.List   `tfsdk:"rule_types"`
	RuleIds    types.List   `tfsdk:"rule_ids"`
	Severities types.List   `tfsdk:"severities"`
	Query      types.String `tfsdk:"query"`
}

func (r *dueDateRuleBlockModel) IsNull() bool {
	return r == nil || (r.IssueType.IsNull() && r.RuleTypes.IsNull() && r.RuleIds.IsNull() && r.Severities.IsNull() && r.Query.IsNull())
}

func (r *dueDateRuleBlockModel) ToPayload() map[string]interface{} {
	payload := map[string]interface{}{}
	if !r.IssueType.IsNull() {
		payload["issue_type"] = r.IssueType.ValueString()
	}
	if !r.RuleTypes.IsNull() {
		payload["rule_types"] = expandList(r.RuleTypes)
	}
	if !r.RuleIds.IsNull() {
		payload["rule_ids"] = expandList(r.RuleIds)
	}
	if !r.Severities.IsNull() {
		payload["severities"] = expandList(r.Severities)
	}
	if !r.Query.IsNull() {
		payload["query"] = r.Query.ValueString()
	}
	return payload
}

func (r *dueDateRuleBlockModel) ValidateRequired() error {
	if r.IssueType.IsNull() {
		return fmt.Errorf("issue_type is required for due date rule")
	}
	if r.RuleTypes.IsNull() {
		return fmt.Errorf("rule_types is required for due date rule")
	}
	return nil
}

type inboxActionBlockModel struct {
	ReasonDescription types.String `tfsdk:"reason_description"`
}

func (a *inboxActionBlockModel) IsNull() bool {
	return a == nil || a.ReasonDescription.IsNull()
}

func (a *inboxActionBlockModel) ToPayload() map[string]interface{} {
	payload := map[string]interface{}{}
	if !a.ReasonDescription.IsNull() {
		payload["reason_description"] = a.ReasonDescription.ValueString()
	}
	return payload
}

func (a *inboxActionBlockModel) ValidateRequired() error {
	return nil
}

type muteActionBlockModel struct {
	Reason            types.String `tfsdk:"reason"`
	ReasonDescription types.String `tfsdk:"reason_description"`
	EnabledUntil      types.Int64  `tfsdk:"enabled_until"`
}

func (a *muteActionBlockModel) IsNull() bool {
	return a == nil || (a.Reason.IsNull() && a.ReasonDescription.IsNull() && a.EnabledUntil.IsNull())
}

func (a *muteActionBlockModel) ToPayload() map[string]interface{} {
	payload := map[string]interface{}{}
	if !a.Reason.IsNull() {
		payload["reason"] = a.Reason.ValueString()
	}
	if !a.ReasonDescription.IsNull() {
		payload["reason_description"] = a.ReasonDescription.ValueString()
	}
	if !a.EnabledUntil.IsNull() {
		payload["enabled_until"] = a.EnabledUntil.ValueInt64()
	}
	return payload
}

func (a *muteActionBlockModel) ValidateRequired() error {
	if a.Reason.IsNull() {
		return fmt.Errorf("reason is required for mute action")
	}
	return nil
}

type DueTimePerSeverity struct {
	Severity types.String `tfsdk:"severity"`
	Time     types.String `tfsdk:"time"`
}

type dueDateActionBlockModel struct {
	DueTimePerSeverity []DueTimePerSeverity `tfsdk:"due_time_per_severity"`
	NotifyBeforeDue    types.String         `tfsdk:"notify_before_due"`
}

func (a *dueDateActionBlockModel) IsNull() bool {
	return a == nil || (len(a.DueTimePerSeverity) == 0 && a.NotifyBeforeDue.IsNull())
}

func (a *dueDateActionBlockModel) ToPayload() map[string]interface{} {
	payload := map[string]interface{}{}
	if len(a.DueTimePerSeverity) > 0 {
		var items []map[string]interface{}
		for _, s := range a.DueTimePerSeverity {
			item := map[string]interface{}{}
			if !s.Severity.IsNull() {
				item["severity"] = s.Severity.ValueString()
			}
			if !s.Time.IsNull() {
				item["time"] = s.Time.ValueString()
			}
			items = append(items, item)
		}
		payload["due_time_per_severity"] = items
	}
	if !a.NotifyBeforeDue.IsNull() {
		payload["notify_before_due"] = a.NotifyBeforeDue.ValueString()
	}
	return payload
}

func (a *dueDateActionBlockModel) ValidateRequired() error {
	if len(a.DueTimePerSeverity) == 0 {
		return fmt.Errorf("due_time_per_severity is required for due date action")
	}
	for i, d := range a.DueTimePerSeverity {
		if d.Severity.IsNull() {
			return fmt.Errorf("severity is required for due_time_per_severity at index %d", i)
		}
		if d.Time.IsNull() {
			return fmt.Errorf("time is required for due_time_per_severity at index %d", i)
		}
	}
	return nil
}

func ValidatePlanRequired(plan automationPipelineRuleModel) error {
	var rule ruleBlock
	var action actionBlock

	if plan.Inbox != nil {
		rule = plan.Inbox.Rule
		action = plan.Inbox.Action
	} else if plan.Mute != nil {
		rule = plan.Mute.Rule
		action = plan.Mute.Action
	} else if plan.DueDate != nil {
		rule = plan.DueDate.Rule
		action = plan.DueDate.Action
	} else {
		return fmt.Errorf("at least one rule/action must be specified")
	}

	if rule.IsNull() {
		return fmt.Errorf("rule is required")
	}
	if err := rule.ValidateRequired(); err != nil {
		return err
	}

	if action.IsNull() {
		return fmt.Errorf("action is required")
	}
	if err := action.ValidateRequired(); err != nil {
		return err
	}

	return nil
}
func ValidateOneRuleActionOnly(plan automationPipelineRuleModel) error {
	count := 0
	if plan.Inbox != nil {
		count++
	}
	if plan.Mute != nil {
		count++
	}
	if plan.DueDate != nil {
		count++
	}

	switch count {
	case 0:
		return fmt.Errorf("at least one rule/action must be specified")
	case 1:
		return nil
	default:
		return fmt.Errorf("only one rule/action can be specified")
	}
}

func ValidateTerraform(plan automationPipelineRuleModel) error {
	if err := ValidateOneRuleActionOnly(plan); err != nil {
		return err
	}

	if err := ValidatePlanRequired(plan); err != nil {
		return err
	}
	return nil
}

type ActionType string

const (
	InboxActionType   ActionType = "inbox_rules"
	MuteActionType    ActionType = "mute_rules"
	DueDateActionType ActionType = "due_date_rules"
)

type BaseAction struct {
	actionType ActionType
}

func (b BaseAction) String() string {
	return string(b.actionType)
}

func (b BaseAction) GetSlug() string {
	switch b.actionType {
	case InboxActionType:
		return "inbox_rules"
	case MuteActionType:
		return "mute_rules"
	case DueDateActionType:
		return "due_date_rules"
	default:
		return ""
	}
}

func (r *automationPipelineRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Datadog Automation Pipeline Rule.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the automation pipeline rule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the rule is enabled.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the automation pipeline rule.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"inbox": schema.SingleNestedBlock{
				Description: "Inbox configuration for the pipeline rule.",
				Blocks: map[string]schema.Block{
					"rule":   r.inboxRuleSchema(),
					"action": r.inboxActionSchema(),
				},
			},
			"mute": schema.SingleNestedBlock{
				Description: "Mute configuration for the pipeline rule.",
				Blocks: map[string]schema.Block{
					"rule":   r.muteRuleSchema(),
					"action": r.muteActionSchema(),
				},
			},
			"due_date": schema.SingleNestedBlock{
				Description: "Mute configuration for the pipeline rule.",
				Blocks: map[string]schema.Block{
					"rule":   r.dueDateRuleSchema(),
					"action": r.dueDateActionSchema(),
				},
			},
		},
	}
}

func (r *automationPipelineRuleResource) inboxRuleSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Rule definition for the pipeline.",
		Attributes: map[string]schema.Attribute{
			"issue_type": schema.StringAttribute{
				Description: "The issue type for the rule.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("vulnerability"),
				},
			},
			"rule_types": schema.ListAttribute{
				Description: "The types of rules associated with this automation pipeline rule.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("application_code_vulnerability", "application_library_vulnerability", "attack_path", "container_image_vulnerability", "identity_risk", "misconfiguration", "api_security"),
					),
				},
			},
			"rule_ids": schema.ListAttribute{
				Description: "The IDs of the rules associated with this automation pipeline rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"severities": schema.ListAttribute{
				Description: "The severities of the rules.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("critical", "high", "medium", "low", "info"),
					),
				},
			},
			"query": schema.StringAttribute{
				Description: "The query associated with the rule.",
				Optional:    true,
			},
		},
	}
}

func (r *automationPipelineRuleResource) inboxActionSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Action definition for the inbox configuration.",
		Attributes: map[string]schema.Attribute{
			"reason_description": schema.StringAttribute{
				Description: "The description for the action's reason.",
				Optional:    true,
			},
		},
	}
}

func (r *automationPipelineRuleResource) muteRuleSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Rule definition for the pipeline.",
		Attributes: map[string]schema.Attribute{
			"issue_type": schema.StringAttribute{
				Description: "The issue type for the rule.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("vulnerability"),
				},
			},
			"rule_types": schema.ListAttribute{
				Description: "The types of rules associated with this automation pipeline rule.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("application_code_vulnerability", "application_library_vulnerability", "attack_path", "container_image_vulnerability", "identity_risk", "misconfiguration", "api_security"),
					),
				},
			},
			"rule_ids": schema.ListAttribute{
				Description: "The IDs of the rules associated with this automation pipeline rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"severities": schema.ListAttribute{
				Description: "The severities of the rules.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("critical", "high", "medium", "low", "info"),
					),
				},
			},
			"query": schema.StringAttribute{
				Description: "The query associated with the rule.",
				Optional:    true,
			},
		},
	}
}

func (r *automationPipelineRuleResource) muteActionSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Action definition for the mute configuration.",
		Attributes: map[string]schema.Attribute{
			"reason": schema.StringAttribute{
				Description: "The reason for the action.",
				Optional:    true,
			},
			"reason_description": schema.StringAttribute{
				Description: "The description for the action's reason.",
				Optional:    true,
			},
			"enabled_until": schema.Int64Attribute{
				Description: "The timestamp until the action is enabled.",
				Optional:    true,
				Validators: []validator.Int64{
					NewTimestampValidator(),
				},
			},
		},
	}
}

func (r *automationPipelineRuleResource) dueDateRuleSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Rule definition for the pipeline.",
		Attributes: map[string]schema.Attribute{
			"issue_type": schema.StringAttribute{
				Description: "The issue type for the rule.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("vulnerability"),
				},
			},
			"rule_types": schema.ListAttribute{
				Description: "The types of rules associated with this automation pipeline rule.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("application_code_vulnerability", "application_library_vulnerability", "attack_path", "container_image_vulnerability", "identity_risk", "misconfiguration", "api_security"),
					),
				},
			},
			"rule_ids": schema.ListAttribute{
				Description: "The IDs of the rules associated with this automation pipeline rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"severities": schema.ListAttribute{
				Description: "The severities of the rules.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("critical", "high", "medium", "low", "info"),
					),
				},
			},
			"query": schema.StringAttribute{
				Description: "The query associated with the rule.",
				Optional:    true,
			},
		},
	}
}

func (r *automationPipelineRuleResource) dueDateActionSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Action definition for the due date configuration.",
		Attributes: map[string]schema.Attribute{
			"notify_before_due": schema.StringAttribute{
				Description: "The time to notify before due.",
				Optional:    true,
				Validators: []validator.String{
					ISODuration(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"due_time_per_severity": schema.ListNestedBlock{
				Description: "The due time per severity.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"severity": schema.StringAttribute{
							Description: "The severity for the due time.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("critical", "high", "medium", "low", "info", "unknown"),
							},
						},
						"time": schema.StringAttribute{
							Description: "The time for the due time.",
							Optional:    true,
							Validators: []validator.String{
								ISODuration(),
							},
						},
					},
				},
			},
		},
	}
}

type MicrosecondTimestampValidator struct{}

func (v MicrosecondTimestampValidator) Description(ctx context.Context) string {
	return "Validates that the timestamp is in microseconds and is after the current time."
}

func (v MicrosecondTimestampValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v MicrosecondTimestampValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	unixTime := time.Now().UnixMilli()

	if req.ConfigValue.ValueInt64() <= unixTime {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Timestamp",
			fmt.Sprintf("The provided timestamp (%d) must be after the current time (%d).", req.ConfigValue.ValueInt64(), unixTime),
		)
	}

	if req.ConfigValue.ValueInt64() > unixTime+1000*60*60*24*365*10 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Timestamp",
			fmt.Sprintf("The provided timestamp (%d) must be at most 10 years in the future.", req.ConfigValue.ValueInt64()),
		)
	}
}

func NewTimestampValidator() validator.Int64 {
	return MicrosecondTimestampValidator{}
}

type ISODurationValidator struct{}

func (v ISODurationValidator) Description(_ context.Context) string {
	return "Validates that the string is an ISO 8601 duration format, only accepting Weeks, Days, Hours, Minutes, and Seconds (e.g. PT12H)."
}

func (v ISODurationValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}

func isValidDuration(duration string) error {
	pattern := `^\bP([1-9]\d*)D\b$`
	re := regexp.MustCompile(pattern)

	if !re.MatchString(duration) {
		return fmt.Errorf(
			`invalid duration '%s': %s`,
			duration,
			"does not respect the expected format 'P{integer}D' where {integer} is a positive number of days")
	}

	return nil
}

func (v ISODurationValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	err := isValidDuration(req.ConfigValue.ValueString())

	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ISO 8601 Duration",
			err.Error())
	}
}

func ISODuration() validator.String {
	return ISODurationValidator{}
}

type apiResponse struct {
	Data struct {
		ID         string        `json:"id"`
		Type       string        `json:"type"`
		Attributes apiAttributes `json:"attributes"`
	} `json:"data"`
}

type apiAttributes struct {
	Name       string      `json:"name"`
	Rule       apiRule     `json:"rule"`
	Action     interface{} `json:"action"`
	Enabled    bool        `json:"enabled"`
	CreatedAt  int64       `json:"created_at"`
	CreatedBy  apiUser     `json:"created_by"`
	ModifiedAt int64       `json:"modified_at"`
	ModifiedBy apiUser     `json:"modified_by"`
}

type apiRule struct {
	IssueType  string   `json:"issue_type"`
	RuleTypes  []string `json:"rule_types"`
	RuleIds    []string `json:"rule_ids"`
	Severities []string `json:"severities"`
	Query      string   `json:"query"`
}

type apiUser struct {
	Name   string `json:"name"`
	Handle string `json:"handle"`
}

type apiActionMute struct {
	Reason            string `json:"reason"`
	ReasonDescription string `json:"reason_description"`
	EnabledUntil      int64  `json:"enabled_until"`
}

type apiActionInbox struct {
	ReasonDescription string `json:"reason_description"`
}

type apiActionDueDate struct {
	DueTimePerSeverity []apiDueTimePerSeverity `json:"due_time_per_severity"`
	NotifyBeforeDue    string                  `json:"notify_before_due"`
}

type apiDueTimePerSeverity struct {
	Severity string `json:"severity"`
	Time     string `json:"time"`
}

func addHeaders(req *http.Request) {
	req.Header.Set("DD-API-KEY", DDAPIKEY)
	req.Header.Set("DD-APPLICATION-KEY", DDAPPKEY)
	req.Header.Set("source", "terraform-provider")
}

func ReadHTTP(url string) (apiResponse, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return apiResponse{}, err
	}
	addHeaders(req)

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return apiResponse{}, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return apiResponse{}, err
	}

	if httpResp.StatusCode != http.StatusOK {
		return apiResponse{}, fmt.Errorf("HTTP status %s\ndetail: %s", httpResp.Status, string(body))
	}

	var response apiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return apiResponse{}, err
	}
	return response, nil
}

func UpsertHTTP(url string, method string, payload []byte) (apiResponse, error) {
	f, err := os.OpenFile("/tmp/terraform-provider-datadog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return apiResponse{}, err
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("URL: %s\n", url)); err != nil {
		return apiResponse{}, err
	}
	if _, err := f.WriteString(fmt.Sprintf("Payload: %s\n", string(payload))); err != nil {
		return apiResponse{}, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return apiResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	addHeaders(req)

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return apiResponse{}, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return apiResponse{}, err
	}

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		return apiResponse{}, fmt.Errorf("HTTP status %s\ndetail: %s", httpResp.Status, string(body))
	}

	var response apiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return apiResponse{}, err
	}

	return response, nil
}

func DeleteHTTP(url string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	addHeaders(req)

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("HTTP status %s\n", httpResp.Status)
	}

	return nil
}

func expandList(list types.List) []interface{} {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var items []interface{}
	for _, elem := range list.Elements() {
		if str, ok := elem.(types.String); ok {
			items = append(items, str.ValueString())
		}
	}
	return items
}

func stringSliceToTerraformValues(slice []string) []attr.Value {
	values := make([]attr.Value, len(slice))
	for i, v := range slice {
		cleaned := strings.Trim(v, "\"")
		values[i] = types.StringValue(cleaned)
	}
	return values
}

func NewAutomationPipelineRuleResource() resource.Resource {
	return &automationPipelineRuleResource{}
}

func (r *automationPipelineRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "automation_pipeline_rule"
}

func (r *automationPipelineRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// No-op
}

// Create function
func (r *automationPipelineRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan automationPipelineRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := ValidateTerraform(plan); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	var action actionBlock
	var rule ruleBlock
	var baseAction BaseAction

	if plan.Inbox != nil {
		action = plan.Inbox.Action
		rule = plan.Inbox.Rule
		baseAction = BaseAction{actionType: InboxActionType}
	} else if plan.Mute != nil {
		action = plan.Mute.Action
		rule = plan.Mute.Rule
		baseAction = BaseAction{actionType: MuteActionType}
	} else if plan.DueDate != nil {
		action = plan.DueDate.Action
		rule = plan.DueDate.Rule
		baseAction = BaseAction{actionType: DueDateActionType}
	}

	url := fmt.Sprintf("%s/%s", apiBaseURL, baseAction.GetSlug())

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"type": baseAction.String(),
			"attributes": map[string]interface{}{
				"name":    plan.Name.ValueString(),
				"rule":    rule.ToPayload(),
				"action":  action.ToPayload(),
				"enabled": plan.Enabled.ValueBool(),
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	response, err := UpsertHTTP(url, http.MethodPost, payloadBytes)
	if err != nil {
		resp.Diagnostics.AddError("Error creating action", err.Error())
		return
	}

	// Set the computed fields
	plan.ID = types.StringValue(response.Data.ID)
	plan.Enabled = types.BoolValue(response.Data.Attributes.Enabled)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read function
func (r *automationPipelineRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state automationPipelineRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var baseAction BaseAction

	if state.Inbox != nil {
		baseAction = BaseAction{actionType: InboxActionType}
	} else if state.Mute != nil {
		baseAction = BaseAction{actionType: MuteActionType}
	} else if state.DueDate != nil {
		baseAction = BaseAction{actionType: DueDateActionType}
	}

	url := fmt.Sprintf("%s/%s/%s", apiBaseURL, baseAction.GetSlug(), state.ID.ValueString())
	response, err := ReadHTTP(url)
	if err != nil {
		resp.Diagnostics.AddError("Error reading action", err.Error())
		return
	}

	state.Name = types.StringValue(response.Data.Attributes.Name)
	state.Enabled = types.BoolValue(response.Data.Attributes.Enabled)

	ruleTypes, diag1 := types.ListValue(
		types.StringType,
		stringSliceToTerraformValues(response.Data.Attributes.Rule.RuleTypes),
	)
	resp.Diagnostics.Append(diag1...)

	ruleIds, diag2 := types.ListValue(
		types.StringType,
		stringSliceToTerraformValues(response.Data.Attributes.Rule.RuleIds),
	)
	resp.Diagnostics.Append(diag2...)

	severities, diag3 := types.ListValue(
		types.StringType,
		stringSliceToTerraformValues(response.Data.Attributes.Rule.Severities),
	)
	resp.Diagnostics.Append(diag3...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Clear actions to ensure only one is populated
	state.Mute = nil
	state.Inbox = nil
	state.DueDate = nil
	if baseAction.actionType == InboxActionType {
		state.Inbox = &inboxBlockModel{}
		var actionData apiActionInbox
		actionBytes, _ := json.Marshal(response.Data.Attributes.Action)
		json.Unmarshal(actionBytes, &actionData)
		state.Inbox.Action = &inboxActionBlockModel{
			ReasonDescription: types.StringValue(actionData.ReasonDescription),
		}
		state.Inbox.Rule = &inboxRuleBlockModel{
			IssueType:  types.StringValue(response.Data.Attributes.Rule.IssueType),
			RuleTypes:  ruleTypes,
			RuleIds:    ruleIds,
			Severities: severities,
			Query:      types.StringValue(response.Data.Attributes.Rule.Query),
		}
	} else if baseAction.actionType == MuteActionType {
		state.Mute = &muteBlockModel{}
		var actionData apiActionMute
		actionBytes, _ := json.Marshal(response.Data.Attributes.Action)
		json.Unmarshal(actionBytes, &actionData)
		state.Mute.Action = &muteActionBlockModel{
			Reason:            types.StringValue(actionData.Reason),
			ReasonDescription: types.StringValue(actionData.ReasonDescription),
			EnabledUntil:      types.Int64Value(actionData.EnabledUntil),
		}
		state.Mute.Rule = &muteRuleBlockModel{
			IssueType:  types.StringValue(response.Data.Attributes.Rule.IssueType),
			RuleTypes:  ruleTypes,
			RuleIds:    ruleIds,
			Severities: severities,
			Query:      types.StringValue(response.Data.Attributes.Rule.Query),
		}
	} else if baseAction.actionType == DueDateActionType {
		state.DueDate = &dueDateBlockModel{}
		var actionData apiActionDueDate
		actionBytes, _ := json.Marshal(response.Data.Attributes.Action)
		json.Unmarshal(actionBytes, &actionData)
		var dtps []DueTimePerSeverity
		for _, s := range actionData.DueTimePerSeverity {
			dtps = append(dtps, DueTimePerSeverity{
				Severity: types.StringValue(s.Severity),
				Time:     types.StringValue(s.Time),
			})
		}
		state.DueDate.Action = &dueDateActionBlockModel{
			NotifyBeforeDue:    types.StringValue(actionData.NotifyBeforeDue),
			DueTimePerSeverity: dtps,
		}
		state.DueDate.Rule = &dueDateRuleBlockModel{
			IssueType:  types.StringValue(response.Data.Attributes.Rule.IssueType),
			RuleTypes:  ruleTypes,
			RuleIds:    ruleIds,
			Severities: severities,
			Query:      types.StringValue(response.Data.Attributes.Rule.Query),
		}
	}

	state.ID = types.StringValue(response.Data.ID)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update function
func (r *automationPipelineRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan automationPipelineRuleModel
	var state automationPipelineRuleModel

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

	if err := ValidateTerraform(plan); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	var planAction actionBlock
	var planRule ruleBlock
	var planActionType ActionType
	var planBaseAction BaseAction
	if plan.Inbox != nil {
		planAction = plan.Inbox.Action
		planRule = plan.Inbox.Rule
		planActionType = InboxActionType
		planBaseAction = BaseAction{actionType: InboxActionType}
	} else if plan.Mute != nil {
		planAction = plan.Mute.Action
		planRule = plan.Mute.Rule
		planActionType = MuteActionType
		planBaseAction = BaseAction{actionType: MuteActionType}
	} else if plan.DueDate != nil {
		planAction = plan.DueDate.Action
		planRule = plan.DueDate.Rule
		planActionType = DueDateActionType
		planBaseAction = BaseAction{actionType: DueDateActionType}
	}

	var stateActionType ActionType
	if state.Inbox != nil {
		stateActionType = InboxActionType
	} else if state.Mute != nil {
		stateActionType = MuteActionType
	} else if state.DueDate != nil {
		stateActionType = DueDateActionType
	}

	if stateActionType != planActionType {
		// Action type has changed; delete the old action and create the new one
		var url string

		url = fmt.Sprintf("%s/%s/%s", apiBaseURL, BaseAction{actionType: stateActionType}.GetSlug(), state.ID.ValueString())

		// Delete old action
		err := DeleteHTTP(url)
		if err != nil {
			resp.Diagnostics.AddError("Error deleting old action", err.Error())
			return
		}

		// Create new action
		url = fmt.Sprintf("%s/%s", apiBaseURL, planBaseAction.GetSlug())
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"type": planBaseAction.String(),
				"attributes": map[string]interface{}{
					"name":    plan.Name.ValueString(),
					"rule":    planRule.ToPayload(),
					"action":  planAction.ToPayload(),
					"enabled": plan.Enabled.ValueBool(),
				},
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		response, err := UpsertHTTP(url, http.MethodPost, payloadBytes)
		if err != nil {
			resp.Diagnostics.AddError("Error creating new action", err.Error())
			return
		}
		plan.ID = types.StringValue(response.Data.ID)
	} else {
		url := fmt.Sprintf("%s/%s/%s", apiBaseURL, planBaseAction.GetSlug(), state.ID.ValueString())

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"type": planBaseAction.String(),
				"attributes": map[string]interface{}{
					"name":    plan.Name.ValueString(),
					"rule":    planRule.ToPayload(),
					"action":  planAction.ToPayload(),
					"enabled": plan.Enabled.ValueBool(),
				},
				"id": state.ID.ValueString(),
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		response, err := UpsertHTTP(url, http.MethodPut, payloadBytes)
		if err != nil {
			resp.Diagnostics.AddError("Error updating action", err.Error())
			return
		}
		plan.ID = types.StringValue(response.Data.ID)
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *automationPipelineRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state automationPipelineRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := ""
	switch {
	case state.Inbox != nil:
		slug = BaseAction{actionType: InboxActionType}.GetSlug()
	case state.Mute != nil:
		slug = BaseAction{actionType: MuteActionType}.GetSlug()
	case state.DueDate != nil:
		slug = BaseAction{actionType: DueDateActionType}.GetSlug()
	}

	if slug != "" {
		url := fmt.Sprintf("%s/%s/%s", apiBaseURL, slug, state.ID.ValueString())
		if err := DeleteHTTP(url); err != nil {
			resp.Diagnostics.AddError("Error deleting action", err.Error())
		}
	} else {
		resp.Diagnostics.AddError("No action type found", "")
	}
}

// ImportState function
func (r *automationPipelineRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}
