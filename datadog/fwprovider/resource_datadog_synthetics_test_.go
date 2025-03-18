package fwprovider

import (
	"bytes"
	"compress/zlib"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &syntheticsTestResource{}
	_ resource.ResourceWithImportState = &syntheticsTestResource{}
)

type syntheticsTestResource struct {
	Api  *datadogV1.SyntheticsApi
	Auth context.Context
}

/*
 * Models
 */

type syntheticsTestModel struct {
	Id                           types.String                                  `tfsdk:"id"`
	Name                         types.String                                  `tfsdk:"name"`
	Type                         types.String                                  `tfsdk:"type"`
	Subtype                      types.String                                  `tfsdk:"subtype"`
	Message                      types.String                                  `tfsdk:"message"`
	MonitorId                    types.Int64                                   `tfsdk:"monitor_id"`
	Status                       types.String                                  `tfsdk:"status"`
	Locations                    types.Set                                     `tfsdk:"locations"`
	Tags                         types.List                                    `tfsdk:"tags"`
	ConfigInitialApplicationArgs types.Map                                     `tfsdk:"config_initial_application_arguments"`
	VariablesFromScript          types.String                                  `tfsdk:"variables_from_script"`
	DeviceIds                    types.List                                    `tfsdk:"device_ids"`
	SetCookie                    types.String                                  `tfsdk:"set_cookie"`
	ForceDeleteDependencies      types.Bool                                    `tfsdk:"force_delete_dependencies"`
	RequestHeaders               types.Map                                     `tfsdk:"request_headers"`
	RequestQuery                 types.Map                                     `tfsdk:"request_query"`
	RequestMetadata              types.Map                                     `tfsdk:"request_metadata"`
	RequestDefinition            []syntheticsTestRequestModel                  `tfsdk:"request_definition"`
	RequestBasicAuth             []syntheticsTestRequestBasicAuthModel         `tfsdk:"request_basicauth"`
	RequestProxy                 []syntheticsTestRequestProxyModel             `tfsdk:"request_proxy"`
	RequestClientCertificate     []syntheticsTestRequestClientCertificateModel `tfsdk:"request_client_certificate"`
	RequestFile                  []syntheticsTestRequestFileModel              `tfsdk:"request_file"`
	Assertion                    []syntheticsTestAssertionModel                `tfsdk:"assertion"`
	BrowserVariable              []syntheticsTestVariableModel                 `tfsdk:"browser_variable"`
	ConfigVariable               []syntheticsTestVariableModel                 `tfsdk:"config_variable"`
	OptionsList                  []syntheticsTestOptionsListModel              `tfsdk:"options_list"`
	MobileOptionsList            []syntheticsTestMobileOptionsListModel        `tfsdk:"mobile_options_list"`
	BrowserStep                  []syntheticsTestBrowserStepModel              `tfsdk:"browser_step"`
	ApiStep                      []syntheticsTestAPIStepModel                  `tfsdk:"api_step"`
	MobileStep                   []syntheticsTestMobileStepModel               `tfsdk:"mobile_step"`
}

type syntheticsTestRequestModel struct {
	Method               types.String `tfsdk:"method"`
	Url                  types.String `tfsdk:"url"`
	Body                 types.String `tfsdk:"body"`
	BodyType             types.String `tfsdk:"body_type"`
	Timeout              types.Int64  `tfsdk:"timeout"`
	Host                 types.String `tfsdk:"host"`
	Port                 types.String `tfsdk:"port"`
	DnsServer            types.String `tfsdk:"dns_server"`
	DnsServerPort        types.String `tfsdk:"dns_server_port"`
	NoSavingResponseBody types.Bool   `tfsdk:"no_saving_response_body"`
	NumberOfPackets      types.Int64  `tfsdk:"number_of_packets"`
	ShouldTrackHops      types.Bool   `tfsdk:"should_track_hops"`
	Servername           types.String `tfsdk:"servername"`
	Message              types.String `tfsdk:"message"`
	CallType             types.String `tfsdk:"call_type"`
	Service              types.String `tfsdk:"service"`
	CertificateDomains   types.List   `tfsdk:"certificate_domains"`
	PersistCookies       types.Bool   `tfsdk:"persist_cookies"`
	ProtoJsonDescriptor  types.String `tfsdk:"proto_json_descriptor"`
	PlainProtoFile       types.String `tfsdk:"plain_proto_file"`
	HttpVersion          types.String `tfsdk:"http_version"`
}

type syntheticsTestRequestBasicAuthModel struct {
	Type                   types.String `tfsdk:"type"`
	Username               types.String `tfsdk:"username"`
	Password               types.String `tfsdk:"password"`
	AccessKey              types.String `tfsdk:"access_key"`
	SecretKey              types.String `tfsdk:"secret_key"`
	Region                 types.String `tfsdk:"region"`
	ServiceName            types.String `tfsdk:"service_name"`
	SessionToken           types.String `tfsdk:"session_token"`
	Domain                 types.String `tfsdk:"domain"`
	Workstation            types.String `tfsdk:"workstation"`
	AccessTokenUrl         types.String `tfsdk:"access_token_url"`
	Audience               types.String `tfsdk:"audience"`
	Resource               types.String `tfsdk:"resource"`
	Scope                  types.String `tfsdk:"scope"`
	TokenApiAuthentication types.String `tfsdk:"token_api_authentication"`
	ClientId               types.String `tfsdk:"client_id"`
	ClientSecret           types.String `tfsdk:"client_secret"`
}

type syntheticsTestRequestProxyModel struct {
	Url     types.String `tfsdk:"url"`
	Headers types.Map    `tfsdk:"headers"`
}

type syntheticsTestRequestClientCertificateModel struct {
	Cert []syntheticsTestRequestClientCertificateItemModel `tfsdk:"cert"`
	Key  []syntheticsTestRequestClientCertificateItemModel `tfsdk:"key"`
}

type syntheticsTestRequestClientCertificateItemModel struct {
	Content  types.String `tfsdk:"content"`
	Filename types.String `tfsdk:"filename"`
}

type syntheticsTestRequestFileModel struct {
	Content          types.String `tfsdk:"content"`
	BucketKey        types.String `tfsdk:"bucket_key"`
	Name             types.String `tfsdk:"name"`
	OriginalFileName types.String `tfsdk:"original_file_name"`
	Size             types.Int64  `tfsdk:"size"`
	Type             types.String `tfsdk:"type"`
}

type syntheticsTestAssertionModel struct {
	Type             types.String                                   `tfsdk:"type"`
	Operator         types.String                                   `tfsdk:"operator"`
	Property         types.String                                   `tfsdk:"property"`
	Target           types.String                                   `tfsdk:"target"`
	Code             types.String                                   `tfsdk:"code"`
	TimingsScope     types.String                                   `tfsdk:"timings_scope"`
	TargetJSONSchema []syntheticsTestAssertionTargetJSONSchemaModel `tfsdk:"targetjsonschema"`
	TargetJSONPath   []syntheticsTestAssertionTargetJSONPathModel   `tfsdk:"targetjsonpath"`
	TargetXPath      []syntheticsTestAssertionTargetXPathModel      `tfsdk:"targetxpath"`
}

type syntheticsTestAssertionTargetJSONSchemaModel struct {
	JSONSchema types.String `tfsdk:"jsonschema"`
	MetaSchema types.String `tfsdk:"metaschema"`
}

type syntheticsTestAssertionTargetJSONPathModel struct {
	ElementsOperator types.String `tfsdk:"elementsoperator"`
	Operator         types.String `tfsdk:"operator"`
	JSONPath         types.String `tfsdk:"jsonpath"`
	TargetValue      types.String `tfsdk:"targetvalue"`
}

type syntheticsTestAssertionTargetXPathModel struct {
	Operator    types.String `tfsdk:"operator"`
	XPath       types.String `tfsdk:"xpath"`
	TargetValue types.String `tfsdk:"targetvalue"`
}

type syntheticsTestVariableModel struct {
	Example types.String `tfsdk:"example"`
	Id      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Pattern types.String `tfsdk:"pattern"`
	Type    types.String `tfsdk:"type"`
	Secure  types.Bool   `tfsdk:"secure"`
}

type syntheticsTestOptionsListModel struct {
	AllowInsecure                types.Bool                              `tfsdk:"allow_insecure"`
	FollowRedirects              types.Bool                              `tfsdk:"follow_redirects"`
	TickEvery                    types.Int64                             `tfsdk:"tick_every"`
	AcceptSelfSigned             types.Bool                              `tfsdk:"accept_self_signed"`
	MinLocationFailed            types.Int64                             `tfsdk:"min_location_failed"`
	MinFailureDuration           types.Int64                             `tfsdk:"min_failure_duration"`
	MonitorName                  types.String                            `tfsdk:"monitor_name"`
	MonitorPriority              types.Int64                             `tfsdk:"monitor_priority"`
	RestrictedRoles              types.Set                               `tfsdk:"restricted_roles"`
	NoScreenshot                 types.Bool                              `tfsdk:"no_screenshot"`
	CheckCertificateRevocation   types.Bool                              `tfsdk:"check_certificate_revocation"`
	IgnoreServerCertificateError types.Bool                              `tfsdk:"ignore_server_certificate_error"`
	DisableCSP                   types.Bool                              `tfsdk:"disable_csp"`
	DisableCORS                  types.Bool                              `tfsdk:"disable_cors"`
	InitialNavigationTimeout     types.Int64                             `tfsdk:"initial_navigation_timeout"`
	HttpVersion                  types.String                            `tfsdk:"http_version"`
	Scheduling                   []syntheticsTestAdvancedSchedulingModel `tfsdk:"scheduling"`
	MonitorOptions               []syntheticsTestMonitorOptionsModel     `tfsdk:"monitor_options"`
	Retry                        []syntheticsTestRetryModel              `tfsdk:"retry"`
	CI                           []syntheticsTestCIModel                 `tfsdk:"ci"`
	RUMSettings                  []syntheticsTestRUMSettingsModel        `tfsdk:"rum_settings"`
}

type syntheticsTestAdvancedSchedulingModel struct {
	Timeframes []syntheticsTestAdvancedSchedulingTimeframesModel `tfsdk:"timeframes"`
	Timezone   types.String                                      `tfsdk:"timezone"`
}

type syntheticsTestAdvancedSchedulingTimeframesModel struct {
	Day  types.Int64  `tfsdk:"day"`
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type syntheticsTestMonitorOptionsModel struct {
	RenotifyInterval    types.Int64 `tfsdk:"renotify_interval"`
	RenotifyOccurrences types.Int64 `tfsdk:"renotify_occurrences"`
}

type syntheticsTestRetryModel struct {
	Count    types.Int64 `tfsdk:"count"`
	Interval types.Int64 `tfsdk:"interval"`
}

type syntheticsTestCIModel struct {
	ExecutionRule types.String `tfsdk:"execution_rule"`
}

type syntheticsTestRUMSettingsModel struct {
	IsEnabled     types.Bool   `tfsdk:"is_enabled"`
	ApplicationId types.String `tfsdk:"application_id"`
	ClientTokenId types.Int64  `tfsdk:"client_token_id"`
}

type syntheticsTestMobileOptionsListModel struct {
	MinFailureDuration     types.Int64                               `tfsdk:"min_failure_duration"`
	TickEvery              types.Int64                               `tfsdk:"tick_every"`
	MonitorName            types.String                              `tfsdk:"monitor_name"`
	MonitorPriority        types.Int64                               `tfsdk:"monitor_priority"`
	RestrictedRoles        types.Set                                 `tfsdk:"restricted_roles"`
	DefaultStepTimeout     types.Int64                               `tfsdk:"default_step_timeout"`
	DeviceIds              types.List                                `tfsdk:"device_ids"`
	NoScreenshot           types.Bool                                `tfsdk:"no_screenshot"`
	Verbosity              types.Int64                               `tfsdk:"verbosity"`
	AllowApplicationCrash  types.Bool                                `tfsdk:"allow_application_crash"`
	DisableAutoAcceptAlert types.Bool                                `tfsdk:"disable_auto_accept_alert"`
	Retry                  []syntheticsTestRetryModel                `tfsdk:"retry"`
	Scheduling             []syntheticsTestAdvancedSchedulingModel   `tfsdk:"scheduling"`
	MonitorOptions         []syntheticsTestMobileMonitorOptionsModel `tfsdk:"monitor_options"`
	Bindings               []syntheticsTestBindingsModel             `tfsdk:"bindings"`
	CI                     []syntheticsTestCIModel                   `tfsdk:"ci"`
	MobileApplication      []syntheticsTestMobileApplicationModel    `tfsdk:"mobile_application"`
}

type syntheticsTestMobileMonitorOptionsModel struct {
	syntheticsTestMonitorOptionsModel
	EscalationMessage      types.String `tfsdk:"escalation_message"`
	NotificationPresetName types.String `tfsdk:"notification_preset_name"`
}

type syntheticsTestBindingsModel struct {
	Principals types.List   `tfsdk:"principals"`
	Relation   types.String `tfsdk:"relation"`
}

type syntheticsTestMobileApplicationModel struct {
	ApplicationId types.String `tfsdk:"application_id"`
	ReferenceId   types.String `tfsdk:"reference_id"`
	ReferenceType types.String `tfsdk:"reference_type"`
}

type syntheticsTestBrowserStepModel struct {
	Name               types.String                           `tfsdk:"name"`
	LocalKey           types.String                           `tfsdk:"local_key"`
	PublicId           types.String                           `tfsdk:"public_id"`
	Type               types.String                           `tfsdk:"type"`
	AllowFailure       types.Bool                             `tfsdk:"allow_failure"`
	AlwaysExecute      types.Bool                             `tfsdk:"always_execute"`
	ExitIfSucceed      types.Bool                             `tfsdk:"exit_if_succeed"`
	IsCritical         types.Bool                             `tfsdk:"is_critical"`
	Timeout            types.Int64                            `tfsdk:"timeout"`
	ForceElementUpdate types.Bool                             `tfsdk:"force_element_update"`
	NoScreenshot       types.Bool                             `tfsdk:"no_screenshot"`
	Params             []syntheticsTestBrowserStepParamsModel `tfsdk:"params"`
}

type syntheticsTestBrowserStepParamsModel struct {
	Attribute          types.String                                             `tfsdk:"attribute"`
	Check              types.String                                             `tfsdk:"check"`
	ClickType          types.String                                             `tfsdk:"click_type"`
	Code               types.String                                             `tfsdk:"code"`
	Delay              types.Int64                                              `tfsdk:"delay"`
	Element            types.String                                             `tfsdk:"element"`
	Email              types.String                                             `tfsdk:"email"`
	File               types.String                                             `tfsdk:"file"`
	Files              types.String                                             `tfsdk:"files"`
	Modifiers          types.List                                               `tfsdk:"modifiers"`
	PlayingTabId       types.String                                             `tfsdk:"playing_tab_id"`
	Request            types.String                                             `tfsdk:"request"`
	SubtestPublicId    types.String                                             `tfsdk:"subtest_public_id"`
	Value              types.String                                             `tfsdk:"value"`
	WithClick          types.Bool                                               `tfsdk:"with_click"`
	X                  types.Int64                                              `tfsdk:"x"`
	Y                  types.Int64                                              `tfsdk:"y"`
	ElementUserLocator []syntheticsTestBrowserStepParamsElementUserLocatorModel `tfsdk:"element_user_locator"`
	Variable           []syntheticsTestBrowserStepParamsVariableModel           `tfsdk:"variable"`
}

type syntheticsTestBrowserStepParamsElementUserLocatorModel struct {
	FailTestOnCannotLocate types.Bool                                                    `tfsdk:"fail_test_on_cannot_locate"`
	Value                  []syntheticsTestBrowserStepParamsElementUserLocatorValueModel `tfsdk:"value"`
}

type syntheticsTestBrowserStepParamsElementUserLocatorValueModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type syntheticsTestBrowserStepParamsVariableModel struct {
	Name    types.String `tfsdk:"name"`
	Example types.String `tfsdk:"example"`
	Secure  types.Bool   `tfsdk:"secure"`
}

type syntheticsTestAPIStepModel struct {
	Name                     types.String                                  `tfsdk:"name"`
	Subtype                  types.String                                  `tfsdk:"subtype"`
	RequestHeaders           types.Map                                     `tfsdk:"request_headers"`
	RequestQuery             types.Map                                     `tfsdk:"request_query"`
	RequestMetadata          types.Map                                     `tfsdk:"request_metadata"`
	ExitIfSucceed            types.Bool                                    `tfsdk:"exit_if_succeed"`
	AllowFailure             types.Bool                                    `tfsdk:"allow_failure"`
	IsCritical               types.Bool                                    `tfsdk:"is_critical"`
	Value                    types.Int64                                   `tfsdk:"value"`
	ExtractedValue           []syntheticsTestAPIStepExtractedValueModel    `tfsdk:"extracted_value"`
	RequestDefinition        []syntheticsTestAPIStepRequestModel           `tfsdk:"request_definition"`
	RequestBasicAuth         []syntheticsTestRequestBasicAuthModel         `tfsdk:"request_basicauth"`
	RequestProxy             []syntheticsTestRequestProxyModel             `tfsdk:"request_proxy"`
	RequestClientCertificate []syntheticsTestRequestClientCertificateModel `tfsdk:"request_client_certificate"`
	RequestFile              []syntheticsTestRequestFileModel              `tfsdk:"request_file"`
	Assertion                []syntheticsTestAssertionModel                `tfsdk:"assertion"`
	Retry                    []syntheticsTestRetryModel                    `tfsdk:"retry"`
}

type syntheticsTestAPIStepExtractedValueModel struct {
	Name   types.String                                     `tfsdk:"name"`
	Type   types.String                                     `tfsdk:"type"`
	Field  types.String                                     `tfsdk:"field"`
	Secure types.Bool                                       `tfsdk:"secure"`
	Parser []syntheticsTestAPIStepExtractedValueParserModel `tfsdk:"parser"`
}

type syntheticsTestAPIStepExtractedValueParserModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type syntheticsTestAPIStepRequestModel struct {
	AllowInsecure   types.Bool `tfsdk:"allow_insecure"`
	FollowRedirects types.Bool `tfsdk:"follow_redirects"`
	syntheticsTestRequestModel
}

type syntheticsTestMobileStepModel struct {
	AllowFailure      types.Bool                            `tfsdk:"allow_failure"`
	HasNewStepElement types.Bool                            `tfsdk:"has_new_step_element"`
	IsCritical        types.Bool                            `tfsdk:"is_critical"`
	Name              types.String                          `tfsdk:"name"`
	NoScreenshot      types.Bool                            `tfsdk:"no_screenshot"`
	PublicId          types.String                          `tfsdk:"public_id"`
	Timeout           types.Int64                           `tfsdk:"timeout"`
	Type              types.String                          `tfsdk:"type"`
	Params            []syntheticsTestMobileStepParamsModel `tfsdk:"params"`
}

type syntheticsTestMobileStepParamsModel struct {
	Value           types.String                                  `tfsdk:"value"`
	Check           types.String                                  `tfsdk:"check"`
	SubtestPublicId types.String                                  `tfsdk:"subtest_public_id"`
	X               types.Float64                                 `tfsdk:"x"`
	Y               types.Float64                                 `tfsdk:"y"`
	Direction       types.String                                  `tfsdk:"direction"`
	MaxScrolls      types.Int64                                   `tfsdk:"max_scrolls"`
	Enable          types.Bool                                    `tfsdk:"enable"`
	Delay           types.Int64                                   `tfsdk:"delay"`
	WithEnter       types.Bool                                    `tfsdk:"with_enter"`
	Element         []syntheticsTestMobileStepParamsElementModel  `tfsdk:"element"`
	Variable        []syntheticsTestMobileStepParamsVariableModel `tfsdk:"variable"`
	Positions       []syntheticsTestMobileStepParamsPositionModel `tfsdk:"positions"`
}

type syntheticsTestMobileStepParamsElementModel struct {
	MultiLocator       types.Map                                        `tfsdk:"multi_locator"`
	Context            types.String                                     `tfsdk:"context"`
	ContextType        types.String                                     `tfsdk:"context_type"`
	ElementDescription types.String                                     `tfsdk:"element_description"`
	TextContent        types.String                                     `tfsdk:"text_content"`
	ViewName           types.String                                     `tfsdk:"view_name"`
	UserLocator        []syntheticsTestMobileStepParamsUserLocatorModel `tfsdk:"user_locator"`
	RelativePosition   []syntheticsTestMobileStepParamsPositionModel    `tfsdk:"relative_position"`
}

type syntheticsTestMobileStepParamsUserLocatorModel struct {
	FailTestOnCannotLocate types.Bool                                            `tfsdk:"fail_test_on_cannot_locate"`
	Values                 []syntheticsTestMobileStepParamsUserLocatorValueModel `tfsdk:"values"`
}

type syntheticsTestMobileStepParamsUserLocatorValueModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type syntheticsTestMobileStepParamsPositionModel struct {
	X types.Float64 `tfsdk:"x"`
	Y types.Float64 `tfsdk:"y"`
}

type syntheticsTestMobileStepParamsVariableModel struct {
	Name    types.String `tfsdk:"name"`
	Example types.String `tfsdk:"example"`
}

/*
 * Resource
 */

func NewSyntheticsTestResource() resource.Resource {
	return &syntheticsTestResource{}
}

func (r *syntheticsTestResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSyntheticsApiV1()
	r.Auth = providerData.Auth
}

func (r *syntheticsTestResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "synthetics_test"
}

/*
 * Schemas
 */

func (r *syntheticsTestResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog synthetics test resource. This can be used to create and manage Datadog synthetics tests.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Synthetics test type.",
				Required:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsTestDetailsTypeFromValue),
				},
			},
			"subtype": schema.StringAttribute{
				Description: "The subtype of the Synthetic API test. Defaults to `http`.",
				Optional:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsTestDetailsSubTypeFromValue),
				},
			},
			"request_headers":  syntheticsTestRequestHeaders(),
			"request_query":    syntheticsTestRequestQuery(),
			"request_metadata": syntheticsTestRequestMetadata(),
			"config_initial_application_arguments": schema.MapAttribute{
				ElementType: types.StringType,
				Description: "Initial application arguments for the mobile test.",
				Optional:    true,
			},
			"variables_from_script": schema.StringAttribute{
				Description: "Variables defined from JavaScript code for API HTTP tests.",
				Optional:    true,
			},
			"device_ids": schema.ListAttribute{
				Description: "Required if `type = \"browser\"`. Array with the different device IDs used to run the test.",
				Optional:    true,
				ElementType: types.StringType,
				// TODO: add fw validator for empty string, if really necessary
			},
			"locations": schema.SetAttribute{
				Description: "Array of locations used to run the test. Refer to [the Datadog Synthetics location data source](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/data-sources/synthetics_locations) to retrieve the list of locations.",
				Required:    true,
				ElementType: types.StringType,
			},
			"name": schema.StringAttribute{
				Description: "Name of Datadog synthetics test.",
				Required:    true,
			},
			"message": schema.StringAttribute{
				Description: "A message to include with notifications for this synthetics test. Email notifications can be sent to specific users by using the same `@username` notation as events.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				Description: "A list of tags to associate with your synthetics test. This can help you categorize and filter tests in the manage synthetics page of the UI. Default is an empty list (`[]`).",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"status": schema.StringAttribute{
				Description: "Define whether you want to start (`live`) or pause (`paused`) a Synthetic test.",
				Required:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsTestPauseStatusFromValue),
				},
			},
			"monitor_id": schema.Int64Attribute{
				Description: "ID of the monitor associated with the Datadog synthetics test.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"set_cookie": schema.StringAttribute{
				Description: "Cookies to be used for a browser test request, using the [Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie) syntax.",
				Optional:    true,
			},
			"force_delete_dependencies": schema.BoolAttribute{
				Description: "A boolean indicating whether this synthetics test can be deleted even if it's referenced by other resources (for example, SLOs and composite monitors).",
				Optional:    true,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"request_definition": schema.ListNestedBlock{
				Description:  "Required if `type = \"api\"`. The synthetics test request.",
				NestedObject: syntheticsTestRequest(),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"request_basicauth":          syntheticsTestRequestBasicAuth(),
			"request_proxy":              syntheticsTestRequestProxy(),
			"request_client_certificate": syntheticsTestRequestClientCertificate(),
			"request_file":               syntheticsTestRequestFile(),
			"assertion":                  syntheticsAPIAssertion(),
			"browser_variable":           syntheticsBrowserVariable(),
			"config_variable":            syntheticsConfigVariable(),
			"options_list":               syntheticsTestOptionsList(),
			"mobile_options_list":        syntheticsMobileTestOptionsList(),
			"browser_step":               syntheticsTestBrowserStep(),
			"api_step":                   syntheticsTestAPIStep(),
			"mobile_step":                syntheticsTestMobileStep(),
		},
	}
}

func syntheticsTestRequest() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"method": schema.StringAttribute{
				Description: "Either the HTTP method/verb to use or a gRPC method available on the service set in the `service` field. Required if `subtype` is `HTTP` or if `subtype` is `grpc` and `callType` is `unary`.",
				Optional:    true,
			},
			"url": schema.StringAttribute{
				Description: "The URL to send the request to.",
				Optional:    true,
			},
			"body": schema.StringAttribute{
				Description: "The request body.",
				Optional:    true,
			},
			"body_type": schema.StringAttribute{
				Description: "Type of the request body.",
				Optional:    true,
				// ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestRequestBodyTypeFromValue),
			},
			"timeout": schema.Int64Attribute{
				Description: "Timeout in seconds for the test.",
				Optional:    true,
			},
			"host": schema.StringAttribute{
				Description: "Host name to perform the test with.",
				Optional:    true,
			},
			"port": schema.StringAttribute{
				Description: "Port to use when performing the test.",
				Optional:    true,
			},
			"dns_server": schema.StringAttribute{
				Description: "DNS server to use for DNS tests (`subtype = \"dns\"`).",
				Optional:    true,
			},
			"dns_server_port": schema.StringAttribute{
				Description: "DNS server port to use for DNS tests.",
				Optional:    true,
			},
			"no_saving_response_body": schema.BoolAttribute{
				Description: "Determines whether or not to save the response body.",
				Optional:    true,
			},
			"number_of_packets": schema.Int64Attribute{
				Description: "Number of pings to use per test for ICMP tests (`subtype = \"icmp\"`) between 0 and 10.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 10),
				},
			},
			"should_track_hops": schema.BoolAttribute{
				Description: "This will turn on a traceroute probe to discover all gateways along the path to the host destination. For ICMP tests (`subtype = \"icmp\"`).",
				Optional:    true,
			},
			"servername": schema.StringAttribute{
				Description: "For SSL tests, it specifies on which server you want to initiate the TLS handshake, allowing the server to present one of multiple possible certificates on the same IP address and TCP port number.",
				Optional:    true,
			},
			"message": schema.StringAttribute{
				Description: "For UDP and websocket tests, message to send with the request.",
				Optional:    true,
			},
			"call_type": schema.StringAttribute{
				Description: "The type of gRPC call to perform.",
				Optional:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsTestCallTypeFromValue),
				},
			},
			"service": schema.StringAttribute{
				Description: "The gRPC service on which you want to perform the gRPC call.",
				Optional:    true,
			},
			"certificate_domains": schema.ListAttribute{
				Description: "By default, the client certificate is applied on the domain of the starting URL for browser tests. If you want your client certificate to be applied on other domains instead, add them in `certificate_domains`.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"persist_cookies": schema.BoolAttribute{
				Description: "Persist cookies across redirects.",
				Optional:    true,
			},
			"proto_json_descriptor": schema.StringAttribute{
				Description:        "A protobuf JSON descriptor.",
				DeprecationMessage: "Use `plain_proto_file` instead.",
				Optional:           true,
			},
			"plain_proto_file": schema.StringAttribute{
				Description: "The content of a proto file as a string.",
				Optional:    true,
			},
			"http_version": schema.StringAttribute{
				Description:        "HTTP version to use for an HTTP request in an API test or step.",
				DeprecationMessage: "Use `http_version` in the `options_list` field instead.",
				Optional:           true,
			},
		},
	}
}

func syntheticsTestRequestHeaders() schema.MapAttribute {
	return schema.MapAttribute{
		ElementType: types.StringType,
		Description: "Header name and value map.",
		Optional:    true,
		// TODO: add fw validator for http headers
	}
}

func syntheticsTestRequestQuery() schema.MapAttribute {
	return schema.MapAttribute{
		ElementType: types.StringType,
		Description: "Query arguments name and value map.",
		Optional:    true,
	}
}

func syntheticsTestRequestBasicAuth() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The HTTP basic authentication credentials. Exactly one nested block is allowed with the structure below.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Description: "Type of basic authentication to use when performing the test.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("web"),
					Validators: []validator.String{
						stringvalidator.OneOf("web", "sigv4", "ntlm", "oauth-client", "oauth-rop", "digest"),
					},
				},
				"username": schema.StringAttribute{
					Description: "Username for authentication.",
					Optional:    true,
				},
				"password": schema.StringAttribute{
					Description: "Password for authentication.",
					Optional:    true,
					Sensitive:   true,
				},
				"access_key": schema.StringAttribute{
					Description: "Access key for `SIGV4` authentication.",
					Optional:    true,
					Sensitive:   true,
				},
				"secret_key": schema.StringAttribute{
					Description: "Secret key for `SIGV4` authentication.",
					Optional:    true,
					Sensitive:   true,
				},
				"region": schema.StringAttribute{
					Description: "Region for `SIGV4` authentication.",
					Optional:    true,
				},
				"service_name": schema.StringAttribute{
					Description: "Service name for `SIGV4` authentication.",
					Optional:    true,
				},
				"session_token": schema.StringAttribute{
					Description: "Session token for `SIGV4` authentication.",
					Optional:    true,
				},
				"domain": schema.StringAttribute{
					Description: "Domain for `ntlm` authentication.",
					Optional:    true,
				},
				"workstation": schema.StringAttribute{
					Description: "Workstation for `ntlm` authentication.",
					Optional:    true,
				},
				"access_token_url": schema.StringAttribute{
					Description: "Access token URL for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
				},
				"audience": schema.StringAttribute{
					Description: "Audience for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
				},
				"resource": schema.StringAttribute{
					Description: "Resource for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
				},
				"scope": schema.StringAttribute{
					Description: "Scope for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
				},
				"token_api_authentication": schema.StringAttribute{
					Description: "Token API Authentication for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsBasicAuthOauthTokenApiAuthenticationFromValue),
					}},
				"client_id": schema.StringAttribute{
					Description: "Client ID for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
				},
				"client_secret": schema.StringAttribute{
					Description: "Client secret for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
					Sensitive:   true,
				},
			},
		},
	}
}

func syntheticsTestRequestProxy() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The proxy to perform the test.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"url": schema.StringAttribute{
					Description: "URL of the proxy to perform the test.",
					Required:    true,
				},
				"headers": syntheticsTestRequestHeaders(),
			},
		},
	}
}

func syntheticsTestRequestClientCertificate() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Client certificate to use when performing the test request. Exactly one nested block is allowed with the structure below.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"cert": syntheticsTestRequestClientCertificateItem(),
				"key":  syntheticsTestRequestClientCertificateItem(),
			},
		},
	}
}

func syntheticsTestRequestClientCertificateItem() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"content": schema.StringAttribute{
					Description: "Content of the certificate.",
					Required:    true,
					Sensitive:   true,
					// TODO: reimplement hash
				},
				"filename": schema.StringAttribute{
					Description: "File name for the certificate.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("Provided in Terraform config"),
				},
			},
		},
	}
}

func syntheticsTestRequestMetadata() schema.MapAttribute {
	return schema.MapAttribute{
		ElementType: types.StringType,
		Description: "Metadata to include when performing the gRPC request.",
		Optional:    true,
	}
}

func syntheticsAPIAssertion() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Assertions used for the test. Multiple `assertion` blocks are allowed with the structure below.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Description: "Type of assertion. **Note** Only some combinations of `type` and `operator` are valid (please refer to [Datadog documentation](https://docs.datadoghq.com/api/latest/synthetics/#create-a-test)).",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.Any(
							validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionTypeFromValue),
							validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionBodyHashTypeFromValue),
							validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionJavascriptTypeFromValue),
						),
					},
				},
				"operator": schema.StringAttribute{
					Description: "Assertion operator. **Note** Only some combinations of `type` and `operator` are valid (please refer to [Datadog documentation](https://docs.datadoghq.com/api/latest/synthetics/#create-a-test)).",
					Optional:    true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionOperatorFromValue),
						stringvalidator.Any(
							validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionJSONPathOperatorFromValue),
							validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionJSONSchemaOperatorFromValue),
							validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionXPathOperatorFromValue),
							validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionBodyHashOperatorFromValue),
						),
					},
				},
				"property": schema.StringAttribute{
					Description: "If assertion type is `header`, this is the header name.",
					Optional:    true,
				},
				"target": schema.StringAttribute{
					Description: "Expected value. Depends on the assertion type, refer to [Datadog documentation](https://docs.datadoghq.com/api/latest/synthetics/#create-a-test) for details.",
					Optional:    true,
				},
				"code": schema.StringAttribute{
					Description: "If assertion type is `javascript`, this is the JavaScript code that performs the assertions.",
					Optional:    true,
				},
				"timings_scope": schema.StringAttribute{
					Description: "Timings scope for response time assertions.",
					Optional:    true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionTimingsScopeFromValue),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"targetjsonschema": schema.ListNestedBlock{
					Description: "Expected structure if `operator` is `validatesJSONSchema`. Exactly one nested block is allowed with the structure below.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"jsonschema": schema.StringAttribute{
								Description: "The JSON Schema to validate the body against.",
								Required:    true,
							},
							"metaschema": schema.StringAttribute{
								Description: "The meta schema to use for the JSON Schema.",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("draft-07"),
							},
						},
					},
				},
				"targetjsonpath": schema.ListNestedBlock{
					Description: "Expected structure if `operator` is `validatesJSONPath`. Exactly one nested block is allowed with the structure below.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"elementsoperator": schema.StringAttribute{
								Description: "The element from the list of results to assert on.",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("firstElementMatches"),
							},
							"operator": schema.StringAttribute{
								Description: "The specific operator to use on the path.",
								Required:    true,
							},
							"jsonpath": schema.StringAttribute{
								Description: "The JSON path to assert.",
								Required:    true,
							},
							"targetvalue": schema.StringAttribute{
								Description: "Expected matching value.",
								Optional:    true,
							},
						},
					},
				},
				"targetxpath": schema.ListNestedBlock{
					Description: "Expected structure if `operator` is `validatesXPath`. Exactly one nested block is allowed with the structure below.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"operator": schema.StringAttribute{
								Description: "The specific operator to use on the path.",
								Required:    true,
							},
							"xpath": schema.StringAttribute{
								Description: "The xpath to assert.",
								Required:    true,
							},
							"targetvalue": schema.StringAttribute{
								Description: "Expected matching value.",
								Optional:    true,
							},
						},
					},
				},
			},
		},
	}
}

func syntheticsTestOptionsRetry() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"count": schema.Int64Attribute{
					Description: "Number of retries needed to consider a location as failed before sending a notification alert. Maximum value: `5`.",
					Optional:    true,
					Computed:    true,
					Default:     int64default.StaticInt64(0),
				},
				"interval": schema.Int64Attribute{
					Description: "Interval between a failed test and the next retry in milliseconds. Maximum value: `5000`.",
					Optional:    true,
					Computed:    true,
					Default:     int64default.StaticInt64(300),
				},
			},
		},
	}
}

func syntheticsTestAdvancedSchedulingTimeframes() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Description: "Array containing objects describing the scheduling pattern to apply to each day.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"day": schema.Int64Attribute{
					Description: "Number representing the day of the week",
					Required:    true,
					Validators: []validator.Int64{
						int64validator.Between(1, 7),
					},
				},
				"from": schema.StringAttribute{
					Description: "The hour of the day on which scheduling starts.",
					Required:    true,
				},
				"to": schema.StringAttribute{
					Description: "The hour of the day on which scheduling ends.",
					Required:    true,
				},
			},
		},
		Validators: []validator.Set{
			setvalidator.IsRequired(),
		},
	}
}

func syntheticsTestAdvancedScheduling() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Object containing timeframes and timezone used for advanced scheduling.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"timeframes": syntheticsTestAdvancedSchedulingTimeframes(),
			},
			Attributes: map[string]schema.Attribute{
				"timezone": schema.StringAttribute{
					Description: "Timezone in which the timeframe is based.",
					Required:    true,
				},
			},
		},
	}
}

func syntheticsTestOptionsList() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Options for Synthetic tests.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_insecure":   syntheticsAllowInsecureOption(),
				"follow_redirects": syntheticsFollowRedirectsOption(),
				"tick_every": schema.Int64Attribute{
					Description: "How often the test should run (in seconds).",
					Required:    true,
					Validators: []validator.Int64{
						int64validator.Between(30, 604800),
					},
				},
				"accept_self_signed": schema.BoolAttribute{
					Description: "For SSL test, whether or not the test should allow self-signed certificates.",
					Optional:    true,
				},
				"min_location_failed": schema.Int64Attribute{
					Description: "Minimum number of locations in failure required to trigger an alert.",
					Optional:    true,
					Computed:    true,
					Default:     int64default.StaticInt64(1),
				},
				"min_failure_duration": schema.Int64Attribute{
					Description: "Minimum amount of time in failure required to trigger an alert (in seconds).",
					Optional:    true,
				},
				"monitor_name": schema.StringAttribute{
					Description: "The monitor name is used for the alert title as well as for all monitor dashboard widgets and SLOs.",
					Optional:    true,
				},
				"monitor_priority": schema.Int64Attribute{
					Optional: true,
					Validators: []validator.Int64{
						int64validator.Between(1, 5),
					},
				},
				"restricted_roles": schema.SetAttribute{
					Description:        "A list of role identifiers pulled from the Roles API to restrict read and write access.",
					DeprecationMessage: "This field is no longer supported by the Datadog API. Please use `datadog_restriction_policy` instead.",
					Optional:           true,
					ElementType:        types.StringType,
				},
				"no_screenshot": schema.BoolAttribute{
					Description: "Prevents saving screenshots of the steps.",
					Optional:    true,
				},
				"check_certificate_revocation": schema.BoolAttribute{
					Description: "For SSL test, whether or not the test should fail on revoked certificate in stapled OCSP.",
					Optional:    true,
				},
				"ignore_server_certificate_error": schema.BoolAttribute{
					Description: "Ignore server certificate error for browser tests.",
					Optional:    true,
				},
				"disable_csp": schema.BoolAttribute{
					Description: "Disable Content Security Policy for browser tests.",
					Optional:    true,
				},
				"disable_cors": schema.BoolAttribute{
					Description: "Disable Cross-Origin Resource Sharing for browser tests.",
					Optional:    true,
				},
				"initial_navigation_timeout": schema.Int64Attribute{
					Description: "Timeout before declaring the initial step as failed (in seconds) for browser tests.",
					Optional:    true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"http_version": syntheticsHttpVersionOption(),
			},
			Blocks: map[string]schema.Block{
				"scheduling": syntheticsTestAdvancedScheduling(),
				"monitor_options": schema.ListNestedBlock{
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"renotify_interval": schema.Int64Attribute{
								Description: "Specify a renotification frequency in minutes. Values available by default are `0`, `10`, `20`, `30`, `40`, `50`, `60`, `90`, `120`, `180`, `240`, `300`, `360`, `720`, `1440`.",
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(0),
							},
							"renotify_occurrences": schema.Int64Attribute{
								Description: "The number of times a monitor renotifies. It can only be set if `renotify_interval` is set.",
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(0),
							},
						},
					},
				},
				"retry": syntheticsTestOptionsRetry(),
				"ci": schema.ListNestedBlock{
					Description: "CI/CD options for a Synthetic test.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"execution_rule": schema.StringAttribute{
								Description: "Execution rule for a Synthetics test.",
								Optional:    true,
								Validators: []validator.String{
									validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsTestExecutionRuleFromValue),
								},
							},
						},
					},
				},
				"rum_settings": schema.ListNestedBlock{
					Description: "The RUM data collection settings for the Synthetic browser test.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					// TODO : iumplement the diffsuppress function
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"is_enabled": schema.BoolAttribute{
								Description: "Determines whether RUM data is collected during test runs.",
								Required:    true,
							},
							"application_id": schema.StringAttribute{
								Description: "RUM application ID used to collect RUM data for the browser test.",
								Optional:    true,
							},
							"client_token_id": schema.Int64Attribute{
								Description: "RUM application API key ID used to collect RUM data for the browser test.",
								Optional:    true,
								Sensitive:   true,
							},
						},
					},
				},
			},
		},
	}
}

func syntheticsMobileTestOptionsList() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Options for Synthetic mobile tests.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"min_failure_duration": schema.Int64Attribute{
					Description: "Minimum amount of time in failure required to trigger an alert (in seconds).",
					Optional:    true,
					Computed:    true,
					Default:     int64default.StaticInt64(0),
				},
				"tick_every": schema.Int64Attribute{
					Description: "How often the test should run (in seconds).",
					Required:    true,
					Validators: []validator.Int64{
						int64validator.Between(300, 604800),
					},
				},
				"monitor_name": schema.StringAttribute{
					Description: "The monitor name is used for the alert title as well as for all monitor dashboard widgets and SLOs.",
					Optional:    true,
				},
				"monitor_priority": schema.Int64Attribute{
					Optional: true,
					Validators: []validator.Int64{
						int64validator.Between(1, 5),
					},
				},
				"restricted_roles": schema.SetAttribute{
					Description:        "A list of role identifiers pulled from the Roles API to restrict read and write access.",
					DeprecationMessage: "This field is no longer supported by the Datadog API. Please use `datadog_restriction_policy` instead.",
					Optional:           true,
					ElementType:        types.StringType,
				},
				"default_step_timeout": schema.Int64Attribute{
					Optional: true,
					Validators: []validator.Int64{
						int64validator.Between(1, 300),
					},
				},
				"device_ids": schema.ListAttribute{
					Required:    true,
					ElementType: types.StringType,
					// TODO: implement validator
					// Validators: []validator.String{
					//     validators.ValidateNonEmptyStrings,
					// },
				},
				"no_screenshot": schema.BoolAttribute{
					Description: "Prevents saving screenshots of the steps.",
					Optional:    true,
				},
				"verbosity": schema.Int64Attribute{
					Optional: true,
					Validators: []validator.Int64{
						int64validator.Between(0, 5),
					},
				},
				"allow_application_crash": schema.BoolAttribute{
					Optional: true,
				},
				"disable_auto_accept_alert": schema.BoolAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"retry":      syntheticsTestOptionsRetry(),
				"scheduling": syntheticsTestAdvancedScheduling(),
				"monitor_options": schema.ListNestedBlock{
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"renotify_interval": schema.Int64Attribute{
								Description: "Specify a renotification frequency in minutes. Values available by default are `0`, `10`, `20`, `30`, `40`, `50`, `60`, `90`, `120`, `180`, `240`, `300`, `360`, `720`, `1440`.",
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(0),
							},
							"escalation_message": schema.StringAttribute{
								Optional: true,
							},
							"renotify_occurrences": schema.Int64Attribute{
								Description: "The number of times a monitor renotifies. It can only be set if `renotify_interval` is set.",
								Optional:    true,
							},
							"notification_preset_name": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsTestOptionsMonitorOptionsNotificationPresetNameFromValue),
								},
							},
						},
					},
				},
				"bindings": schema.ListNestedBlock{
					Description: "Restriction policy bindings for the Synthetic mobile test. Should not be used in parallel with a `datadog_restriction_policy` resource",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"principals": schema.ListAttribute{
								Optional:    true,
								ElementType: types.StringType,
							},
							"relation": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsTestRestrictionPolicyBindingRelationFromValue),
								},
							},
						},
					},
				},
				"ci": schema.ListNestedBlock{
					Description: "CI/CD options for a Synthetic test.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"execution_rule": schema.StringAttribute{
								Description: "Execution rule for a Synthetics test.",
								Required:    true,
								Validators: []validator.String{
									validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsTestExecutionRuleFromValue),
								},
							},
						},
					},
				},
				"mobile_application": schema.ListNestedBlock{
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"application_id": schema.StringAttribute{
								Required: true,
							},
							"reference_id": schema.StringAttribute{
								Required: true,
							},
							"reference_type": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsMobileTestsMobileApplicationReferenceTypeFromValue),
								},
							},
						},
					},
				},
			},
		},
	}
}

func syntheticsTestAPIStep() schema.ListNestedBlock {
	requestElemSchema := syntheticsTestRequest()
	// In test `options_list` for single API tests, but in `api_step.request_definition` for API steps.
	requestElemSchema.Attributes["allow_insecure"] = syntheticsAllowInsecureOption()
	requestElemSchema.Attributes["follow_redirects"] = syntheticsFollowRedirectsOption()
	requestElemSchema.Attributes["http_version"] = syntheticsHttpVersionOption()

	return schema.ListNestedBlock{
		Description: "Steps for multi-step API tests.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "The name of the step.",
					Required:    true,
				},
				"subtype": schema.StringAttribute{
					Description: "The subtype of the Synthetic multi-step API test step.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("http"),
					Validators:  []validator.String{
						// TODO: Fix this
						// validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAPITestStepSubtypeFromValue, datadogV1.NewSyntheticsAPIWaitStepSubtypeFromValue),
					},
				},
				"request_headers":  syntheticsTestRequestHeaders(),
				"request_query":    syntheticsTestRequestQuery(),
				"request_metadata": syntheticsTestRequestMetadata(),
				"exit_if_succeed": schema.BoolAttribute{
					Description: "Determines whether or not to exit the test if the step succeeds.",
					Optional:    true,
				},
				"allow_failure": schema.BoolAttribute{
					Description: "Determines whether or not to continue with test if this step fails.",
					Optional:    true,
				},
				"is_critical": schema.BoolAttribute{
					Description: "Determines whether or not to consider the entire test as failed if this step fails. Can be used only if `allow_failure` is `true`.",
					Optional:    true,
				},
				"value": schema.Int64Attribute{
					Description: "The time to wait in seconds. Minimum value: 0. Maximum value: 180.",
					Optional:    true,
				},
			},
			Blocks: map[string]schema.Block{
				"extracted_value": schema.ListNestedBlock{
					Description: "Values to parse and save as variables from the response.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Required: true,
							},
							"type": schema.StringAttribute{
								Description: "Property of the Synthetics Test Response to use for the variable.",
								Required:    true,
								Validators: []validator.String{
									validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsLocalVariableParsingOptionsTypeFromValue),
								},
							},
							"field": schema.StringAttribute{
								Description: "When type is `http_header` or `grpc_metadata`, name of the header or metadatum to extract.",
								Optional:    true,
							},
							"secure": schema.BoolAttribute{
								Description: "Determines whether or not the extracted value will be obfuscated.",
								Optional:    true,
							},
						},
						Blocks: map[string]schema.Block{
							"parser": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
									listvalidator.IsRequired(),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Description: "Type of parser for a Synthetics global variable from a synthetics test.",
											Required:    true,
											Validators: []validator.String{
												validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsGlobalVariableParserTypeFromValue),
											},
										},
										"value": schema.StringAttribute{
											Description: "Regex or JSON path used for the parser. Not used with type `raw`.",
											Optional:    true,
										},
									},
								},
							},
						},
					},
				},
				"request_definition": schema.ListNestedBlock{
					Description: "The request for the API step.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: requestElemSchema,
				},
				"request_basicauth":          syntheticsTestRequestBasicAuth(),
				"request_proxy":              syntheticsTestRequestProxy(),
				"request_client_certificate": syntheticsTestRequestClientCertificate(),
				"request_file":               syntheticsTestRequestFile(),
				"assertion":                  syntheticsAPIAssertion(),
				"retry":                      syntheticsTestOptionsRetry(),
			},
		},
	}
}

func syntheticsTestRequestFile() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Files to be used as part of the request in the test.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"content": schema.StringAttribute{
					Description: "Content of the file.",
					Optional:    true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 3145728),
					},
				},
				"bucket_key": schema.StringAttribute{
					Description: "Bucket key of the file.",
					Computed:    true,
				},
				"name": schema.StringAttribute{
					Description: "Name of the file.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 1500),
					},
				},
				"original_file_name": schema.StringAttribute{
					Description: "Original name of the file.",
					Optional:    true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 1500),
					},
				},
				"size": schema.Int64Attribute{
					Description: "Size of the file.",
					Required:    true,
					Validators: []validator.Int64{
						int64validator.Between(1, 3145728),
					},
				},
				"type": schema.StringAttribute{
					Description: "Type of the file.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 1500),
					},
				},
			},
		},
	}
}

func syntheticsTestBrowserStep() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Steps for browser tests.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "Name of the step.",
					Required:    true,
				},
				"local_key": schema.StringAttribute{
					Description: "A unique identifier used to track steps after reordering.",
					Optional:    true,
				},
				"public_id": schema.StringAttribute{
					Description: "The identifier of the step on the backend.",
					Computed:    true,
				},
				"type": schema.StringAttribute{
					Description: "Type of the step.",
					Required:    true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsStepTypeFromValue),
					},
				},
				"allow_failure": schema.BoolAttribute{
					Description: "Determines if the step should be allowed to fail.",
					Optional:    true,
				},
				"always_execute": schema.BoolAttribute{
					Description: "Determines whether or not to always execute this step even if the previous step failed or was skipped.",
					Optional:    true,
				},
				"exit_if_succeed": schema.BoolAttribute{
					Description: "Determines whether or not to exit the test if the step succeeds.",
					Optional:    true,
				},
				"is_critical": schema.BoolAttribute{
					Description: "Determines whether or not to consider the entire test as failed if this step fails. Can be used only if `allow_failure` is `true`.",
					Optional:    true,
				},
				"timeout": schema.Int64Attribute{
					Description: "Used to override the default timeout of a step.",
					Optional:    true,
				},
				"force_element_update": schema.BoolAttribute{
					Description: "Force update of the \"element\" parameter for the step.",
					Optional:    true,
				},
				"no_screenshot": schema.BoolAttribute{
					Description: "Prevents saving screenshots of the step.",
					Optional:    true,
				},
			},
			Blocks: map[string]schema.Block{
				"params": syntheticsBrowserStepParams(),
			},
		},
	}
}

func syntheticsBrowserStepParams() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Parameters for the step.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.IsRequired(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"attribute": schema.StringAttribute{
					Description: "Name of the attribute to use for an \"assert attribute\" step.",
					Optional:    true,
				},
				"check": schema.StringAttribute{
					Description: "Check type to use for an assertion step.",
					Optional:    true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsCheckTypeFromValue),
					},
				},
				"click_type": schema.StringAttribute{
					Description: "Type of click to use for a \"click\" step.",
					Optional:    true,
					Validators: []validator.String{
						stringvalidator.OneOf("contextual", "double", "primary"),
					},
				},
				"code": schema.StringAttribute{
					Description: "Javascript code to use for the step.",
					Optional:    true,
				},
				"delay": schema.Int64Attribute{
					Description: "Delay between each key stroke for a \"type test\" step.",
					Optional:    true,
				},
				"element": schema.StringAttribute{
					Description: "Element to use for the step, JSON encoded string.",
					Optional:    true,
					// TODO: convert diffsuppress
				},
				"email": schema.StringAttribute{
					Description: "Details of the email for an \"assert email\" step, JSON encoded string.",
					Optional:    true,
				},
				"file": schema.StringAttribute{
					Description: "JSON encoded string used for an \"assert download\" step. Refer to the examples for a usage example showing the schema.",
					Optional:    true,
					// TODO: convert diffsuppress
				},
				"files": schema.StringAttribute{
					Description: "Details of the files for an \"upload files\" step, JSON encoded string.",
					Optional:    true,
				},
				"modifiers": schema.ListAttribute{
					Description: "Modifier to use for a \"press key\" step.",
					Optional:    true,
					ElementType: types.StringType,
					Validators: []validator.List{
						listvalidator.ValueStringsAre(stringvalidator.OneOf("Alt", "Control", "meta", "Shift")),
					},
				},
				"playing_tab_id": schema.StringAttribute{
					Description: "ID of the tab to play the subtest.",
					Optional:    true,
				},
				"request": schema.StringAttribute{
					Description: "Request for an API step.",
					Optional:    true,
				},
				"subtest_public_id": schema.StringAttribute{
					Description: "ID of the Synthetics test to use as subtest.",
					Optional:    true,
				},
				"value": schema.StringAttribute{
					Description: "Value of the step.",
					Optional:    true,
				},
				"with_click": schema.BoolAttribute{
					Description: "For \"file upload\" steps.",
					Optional:    true,
				},
				"x": schema.Int64Attribute{
					Description: "X coordinates for a \"scroll step\".",
					Optional:    true,
				},
				"y": schema.Int64Attribute{
					Description: "Y coordinates for a \"scroll step\".",
					Optional:    true,
				},
			},
			Blocks: map[string]schema.Block{
				"element_user_locator": schema.ListNestedBlock{
					Description: "Custom user selector to use for the step.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fail_test_on_cannot_locate": schema.BoolAttribute{
								Description: "Whether to fail the test if the locator cannot find the element.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
						},
						Blocks: map[string]schema.Block{
							"value": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Optional: true,
											Computed: true,
											Default:  stringdefault.StaticString("css"),
											Validators: []validator.String{
												stringvalidator.OneOf("css", "xpath"),
											},
										},
										"value": schema.StringAttribute{
											Required: true,
										},
									},
								},
							},
						},
					},
				},
				"variable": schema.ListNestedBlock{
					Description: "Details of the variable to extract.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Description: "Name of the extracted variable.",
								Optional:    true,
							},
							"example": schema.StringAttribute{
								Description: "Example of the extracted variable.",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString(""),
							},
							"secure": schema.BoolAttribute{
								Description: "Whether the value of this variable will be obfuscated in test results.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
						},
					},
				},
			},
		},
	}
}

func syntheticsTestMobileStep() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Steps for mobile tests",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_failure": schema.BoolAttribute{
					Description: "A boolean set to allow this step to fail.",
					Optional:    true,
				},
				"has_new_step_element": schema.BoolAttribute{
					Description: "A boolean set to determine if the step has a new step element.",
					Optional:    true,
				},
				"is_critical": schema.BoolAttribute{
					Description: "A boolean to use in addition to `allowFailure` to determine if the test should be marked as failed when the step fails.",
					Optional:    true,
				},
				"name": schema.StringAttribute{
					Description: "The name of the step.",
					Required:    true,
				},
				"no_screenshot": schema.BoolAttribute{
					Description: "A boolean set to not take a screenshot for the step.",
					Optional:    true,
				},
				"public_id": schema.StringAttribute{
					Description: "The public ID of the step.",
					Optional:    true,
				},
				"timeout": schema.Int64Attribute{
					Description: "The time before declaring a step failed.",
					Optional:    true,
				},
				"type": schema.StringAttribute{
					Description: "The type of the step.",
					Required:    true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsMobileStepTypeFromValue),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"params": syntheticsMobileStepParams(),
			},
		},
	}
}

func syntheticsMobileStepParams() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Parameters for the step.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.IsRequired(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"value": schema.StringAttribute{
					Description: "Value of the step.",
					Optional:    true,
				},
				"check": schema.StringAttribute{
					Description: "Check type to use for an assertion step.",
					Optional:    true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsCheckTypeFromValue),
					},
				},
				"subtest_public_id": schema.StringAttribute{
					Description: "ID of the Synthetics test to use as subtest.",
					Optional:    true,
				},
				"x": schema.Float64Attribute{
					Description: "X coordinates for a scroll step.",
					Optional:    true,
				},
				"y": schema.Float64Attribute{
					Description: "Y coordinates for a scroll step.",
					Optional:    true,
				},
				"direction": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsMobileStepParamsDirectionFromValue),
					},
				},
				"max_scrolls": schema.Int64Attribute{
					Optional: true,
				},
				"enable": schema.BoolAttribute{
					Optional: true,
				},
				"delay": schema.Int64Attribute{
					Description: "Delay between each key stroke for a type test step.",
					Optional:    true,
				},
				"with_enter": schema.BoolAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"element": schema.ListNestedBlock{
					Description: "Element to use for the step, JSON encoded string.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"multi_locator": schema.MapAttribute{
								ElementType: types.StringType,
								Optional:    true,
							},
							"context": schema.StringAttribute{
								Optional: true,
							},
							"context_type": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsMobileStepParamsElementContextTypeFromValue),
								},
							},
							"element_description": schema.StringAttribute{
								Optional: true,
							},
							"text_content": schema.StringAttribute{
								Optional: true,
							},
							"view_name": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"user_locator": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"fail_test_on_cannot_locate": schema.BoolAttribute{
											Optional: true,
										},
									},
									Blocks: map[string]schema.Block{
										"values": schema.ListNestedBlock{
											Validators: []validator.List{
												listvalidator.SizeAtMost(5),
												listvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"type": schema.StringAttribute{
														Optional: true,
														Validators: []validator.String{
															validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsMobileStepParamsElementUserLocatorValuesItemsTypeFromValue),
														},
													},
													"value": schema.StringAttribute{
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
							"relative_position": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"x": schema.Float64Attribute{
											Optional: true,
										},
										"y": schema.Float64Attribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
				"variable": schema.ListNestedBlock{
					Description: "Details of the variable to extract.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Description: "Name of the extracted variable.",
								Required:    true,
							},
							"example": schema.StringAttribute{
								Description: "Example of the extracted variable.",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString(""),
							},
						},
					},
				},
				"positions": schema.ListNestedBlock{
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"x": schema.Float64Attribute{
								Optional: true,
							},
							"y": schema.Float64Attribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func syntheticsBrowserVariable() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Variables used for a browser test step. Multiple `variable` blocks are allowed with the structure below.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"example": schema.StringAttribute{
					Description: "Example for the variable.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
				},
				"id": schema.StringAttribute{
					Description: "ID of the global variable to use. This is actually only used (and required) in the case of using a variable of type `global`.",
					Optional:    true,
				},
				"name": schema.StringAttribute{
					Description: "Name of the variable.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
					},
				},
				"pattern": schema.StringAttribute{
					Description: "Pattern of the variable.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
				},
				"type": schema.StringAttribute{
					Description: "Type of browser test variable.",
					Required:    true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsBrowserVariableTypeFromValue),
					},
				},
				"secure": schema.BoolAttribute{
					Description: "Determines whether or not the browser test variable is obfuscated. Can only be used with a browser variable of type `text`.",
					Optional:    true,
				},
			},
		},
	}
}

func syntheticsConfigVariable() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Variables used for the test configuration. Multiple `config_variable` blocks are allowed with the structure below.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"example": schema.StringAttribute{
					Description: "Example for the variable. This value is not returned by the API when `secure = true`. Avoid drift by only making updates to this value from within Terraform.",
					Optional:    true,
				},
				"name": schema.StringAttribute{
					Description: "Name of the variable.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
					},
				},
				"pattern": schema.StringAttribute{
					Description: "Pattern of the variable. This value is not returned by the API when `secure = true`. Avoid drift by only making updates to this value from within Terraform.",
					Optional:    true,
				},
				"type": schema.StringAttribute{
					Description: "Type of test configuration variable.",
					Required:    true,
					Validators: []validator.String{
						validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsConfigVariableTypeFromValue),
					},
				},
				"id": schema.StringAttribute{
					Description: "When type = `global`, ID of the global variable to use.",
					Optional:    true,
				},
				"secure": schema.BoolAttribute{
					Description: "Whether the value of this variable will be obfuscated in test results.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
				},
			},
		},
	}
}

func syntheticsAllowInsecureOption() schema.BoolAttribute {
	return schema.BoolAttribute{
		Description: "Allows loading insecure content for a request in an API test or in a multistep API test step.",
		Optional:    true,
	}
}

func syntheticsFollowRedirectsOption() schema.BoolAttribute {
	return schema.BoolAttribute{
		Description: "Determines whether or not the API HTTP test should follow redirects.",
		Optional:    true,
	}
}

func syntheticsHttpVersionOption() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "HTTP version to use for an HTTP request in an API test or step.",
		Optional:    true,
		Computed:    true,
		Default:     stringdefault.StaticString(string(datadogV1.SYNTHETICSTESTOPTIONSHTTPVERSION_ANY)),
		Validators: []validator.String{
			validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsTestOptionsHTTPVersionFromValue),
		},
	}
}

/*
 * CRUD functions
 */

func (r *syntheticsTestResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *syntheticsTestResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state syntheticsTestModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	// Get the generic test to detect if it's an API, browser, or mobile test
	syntheticsTest, httpResp, err := r.Api.GetTest(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SyntheticsTest"))
		return
	}
	if err := utils.CheckForUnparsed(syntheticsTest); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	testType := syntheticsTest.GetType()

	switch testType {
	case datadogV1.SYNTHETICSTESTDETAILSTYPE_API:
		var syntheticsAPITest datadogV1.SyntheticsAPITest
		syntheticsAPITest, _, err = r.Api.GetAPITest(r.Auth, id)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving API SyntheticsTest"))
			return
		}
		if err := utils.CheckForUnparsed(syntheticsAPITest); err != nil {
			response.Diagnostics.AddError("API response contains unparsedObject", err.Error())
			return
		}
		r.updateSyntheticsAPITestLocalState(ctx, &state, &syntheticsAPITest)
	case datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER:
		syntheticsBrowserTest, _, err := r.Api.GetBrowserTest(r.Auth, id)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Browser SyntheticsTest"))
			return
		}
		if err := utils.CheckForUnparsed(syntheticsBrowserTest); err != nil {
			response.Diagnostics.AddError("Browser response contains unparsedObject", err.Error())
			return
		}
		r.updateSyntheticsBrowserTestLocalState(ctx, &state, &syntheticsBrowserTest)
	case datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE:
		var syntheticsMobileTest datadogV1.SyntheticsMobileTest
		syntheticsMobileTest, _, err = r.Api.GetMobileTest(r.Auth, id)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Mobile SyntheticsTest"))
			return
		}
		if err := utils.CheckForUnparsed(syntheticsMobileTest); err != nil {
			response.Diagnostics.AddError("Mobile response contains unparsedObject", err.Error())
			return
		}
		r.updateSyntheticsMobileTestLocalState(ctx, &state, &syntheticsMobileTest)
	default:
		response.Diagnostics.AddError("Invalid Synthetics Test Type", fmt.Sprintf("Unsupported synthetics test type: %s", testType))
	}

	// Save data into Terraform state
	fmt.Printf("Final state monitor_id: %v\n", state.MonitorId.ValueInt64())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsTestResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state syntheticsTestModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var testId string
	testType := getSyntheticsTestType(state)

	switch testType {
	case datadogV1.SYNTHETICSTESTDETAILSTYPE_API:
		apiTestBody, diags := r.buildDatadogSyntheticsAPITest(ctx, &state)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		createdTest, _, err := r.Api.CreateSyntheticsAPITest(r.Auth, *apiTestBody)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating API SyntheticsTest"))
			return
		}
		if err := utils.CheckForUnparsed(createdTest); err != nil {
			response.Diagnostics.AddError("API response contains unparsedObject", err.Error())
			return
		}
		testId = createdTest.GetPublicId()
		r.updateSyntheticsAPITestLocalState(ctx, &state, &createdTest)

	case datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER:
		browserTestBody, diags := r.buildDatadogSyntheticsBrowserTest(ctx, &state)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		createdTest, _, err := r.Api.CreateSyntheticsBrowserTest(r.Auth, *browserTestBody)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating Browser SyntheticsTest"))
			return
		}
		if err := utils.CheckForUnparsed(createdTest); err != nil {
			response.Diagnostics.AddError("Browser response contains unparsedObject", err.Error())
			return
		}
		testId = createdTest.GetPublicId()
		r.updateSyntheticsBrowserTestLocalState(ctx, &state, &createdTest)

	case datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE:
		mobileTestBody, diags := r.buildDatadogSyntheticsMobileTest(ctx, &state)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		createdTest, _, err := r.Api.CreateSyntheticsMobileTest(r.Auth, *mobileTestBody)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating Mobile SyntheticsTest"))
			return
		}
		if err := utils.CheckForUnparsed(createdTest); err != nil {
			response.Diagnostics.AddError("Mobile response contains unparsedObject", err.Error())
			return
		}
		testId = createdTest.GetPublicId()
		r.updateSyntheticsMobileTestLocalState(ctx, &state, &createdTest)

	default:
		response.Diagnostics.AddError("Invalid Synthetics Test Type", fmt.Sprintf("Unsupported synthetics test type: %s", testType))
		return
	}

	state.Id = types.StringValue(testId)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsTestResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state syntheticsTestModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	testType := getSyntheticsTestType(state)

	switch testType {
	case datadogV1.SYNTHETICSTESTDETAILSTYPE_API:
		apiTestBody, diags := r.buildDatadogSyntheticsAPITest(ctx, &state)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		updatedTest, _, err := r.Api.UpdateAPITest(r.Auth, id, *apiTestBody)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating API SyntheticsTest"))
			return
		}
		if err := utils.CheckForUnparsed(updatedTest); err != nil {
			response.Diagnostics.AddError("API response contains unparsedObject", err.Error())
			return
		}
		r.updateSyntheticsAPITestLocalState(ctx, &state, &updatedTest)
	case datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER:
		browserTestBody, diags := r.buildDatadogSyntheticsBrowserTest(ctx, &state)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		updatedTest, _, err := r.Api.UpdateBrowserTest(r.Auth, id, *browserTestBody)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating Browser SyntheticsTest"))
			return
		}
		if err := utils.CheckForUnparsed(updatedTest); err != nil {
			response.Diagnostics.AddError("Browser response contains unparsedObject", err.Error())
			return
		}
		r.updateSyntheticsBrowserTestLocalState(ctx, &state, &updatedTest)
	case datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE:
		mobileTestBody, diags := r.buildDatadogSyntheticsMobileTest(ctx, &state)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		updatedTest, _, err := r.Api.UpdateMobileTest(r.Auth, id, *mobileTestBody)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating Mobile SyntheticsTest"))
			return
		}
		if err := utils.CheckForUnparsed(updatedTest); err != nil {
			response.Diagnostics.AddError("Mobile response contains unparsedObject", err.Error())
			return
		}
		r.updateSyntheticsMobileTestLocalState(ctx, &state, &updatedTest)
	default:
		response.Diagnostics.AddError("Invalid Synthetics Test Type", fmt.Sprintf("Unsupported synthetics test type: %s", testType))
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsTestResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state syntheticsTestModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	syntheticsDeleteTestsPayload := datadogV1.SyntheticsDeleteTestsPayload{PublicIds: []string{id}}

	if state.ForceDeleteDependencies.ValueBool() {
		syntheticsDeleteTestsPayload.SetForceDeleteDependencies(true)
	}

	_, httpResp, err := r.Api.DeleteTests(r.Auth, syntheticsDeleteTestsPayload)
	if err != nil {
		if httpResp == nil || httpResp.StatusCode != 404 {
			// The resource is assumed to still exist, and all prior state is preserved.
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting SyntheticsTest"))
			return
		}
	}

	// The resource is assumed to be destroyed, so we remove it from the state
	response.State.RemoveResource(ctx)
}

/*
 * State update functions
 */

func (r *syntheticsTestResource) updateSyntheticsAPITestLocalState(ctx context.Context, state *syntheticsTestModel, resp *datadogV1.SyntheticsAPITest) {
	state.Type = types.StringValue(string(resp.GetType()))
	if resp.HasSubtype() {
		state.Subtype = types.StringValue(string(resp.GetSubtype()))
	}

	fmt.Println("hey! inside of updateSyntheticsAPITestLocalState")

	config := resp.GetConfig()
	actualRequest := config.GetRequest()
	localRequest := buildTerraformTestRequest(ctx, actualRequest)

	if resp.GetSubtype() != "multi" {
		state.RequestDefinition = []syntheticsTestRequestModel{localRequest}
	}
	if headers := actualRequest.GetHeaders(); headers != nil {
		state.RequestHeaders, _ = types.MapValueFrom(ctx, types.StringType, headers)
	}
	if query := actualRequest.GetQuery(); query != nil {
		state.RequestQuery, _ = types.MapValueFrom(ctx, types.StringType, query)
	}
	if metadata := actualRequest.GetMetadata(); metadata != nil {
		state.RequestMetadata, _ = types.MapValueFrom(ctx, types.StringType, metadata)
	}

	if basicAuth, ok := actualRequest.GetBasicAuthOk(); ok {
		state.RequestBasicAuth = []syntheticsTestRequestBasicAuthModel{buildTerraformBasicAuth(basicAuth)}
	}

	if clientCertificate, ok := actualRequest.GetCertificateOk(); ok {
		oldCertificate := state.RequestClientCertificate
		state.RequestClientCertificate = []syntheticsTestRequestClientCertificateModel{buildTerraformRequestCertificates(*clientCertificate, oldCertificate)}
	}

	if proxy, ok := actualRequest.GetProxyOk(); ok {
		state.RequestProxy = []syntheticsTestRequestProxyModel{buildTerraformTestRequestProxy(*proxy)}
	}

	if files, ok := actualRequest.GetFilesOk(); ok && files != nil && len(*files) > 0 {
		state.RequestFile = buildTerraformBodyFiles(files, state.RequestFile)
	}

	assertions := config.GetAssertions()
	localAssertions := buildTerraformAssertions(assertions)
	state.Assertion = localAssertions

	configVariables := config.GetConfigVariables()
	state.ConfigVariable = buildTerraformConfigVariables(configVariables, state.ConfigVariable)

	if variablesFromScript, ok := config.GetVariablesFromScriptOk(); ok {
		state.VariablesFromScript = types.StringValue(*variablesFromScript)
	}

	if steps, ok := config.GetStepsOk(); ok {
		localSteps := make([]syntheticsTestAPIStepModel, len(*steps))
		for i, step := range *steps {
			localSteps[i] = buildTerraformAPITestStep(ctx, step)
		}
		state.ApiStep = localSteps
	}

	state.DeviceIds, _ = types.ListValueFrom(ctx, types.StringType, resp.GetOptions().DeviceIds)
	state.Locations, _ = types.SetValueFrom(ctx, types.StringType, resp.Locations)
	state.OptionsList = buildTerraformTestOptions(ctx, resp.GetOptions())
	state.Name = types.StringValue(resp.GetName())
	state.Message = types.StringValue(resp.GetMessage())
	state.Status = types.StringValue(string(resp.GetStatus()))
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, resp.Tags)

	if monitorId, ok := resp.GetMonitorIdOk(); ok {
		fmt.Println("HEYYYY monitorId", *monitorId)
		state.MonitorId = types.Int64Value(*monitorId)
	}

	// show option list certificate domains
	// fmt.Println("show option list certificate domains")
	// fmt.Println(state.OptionsList[0].)
}

func (r *syntheticsTestResource) updateSyntheticsBrowserTestLocalState(ctx context.Context, state *syntheticsTestModel, resp *datadogV1.SyntheticsBrowserTest) {
}

func (r *syntheticsTestResource) updateSyntheticsMobileTestLocalState(ctx context.Context, state *syntheticsTestModel, resp *datadogV1.SyntheticsMobileTest) {
}

/*
 * Transformer functions between datadog and terraform
 */

func (r *syntheticsTestResource) buildDatadogSyntheticsAPITest(ctx context.Context, state *syntheticsTestModel) (*datadogV1.SyntheticsAPITest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	syntheticsTest := datadogV1.NewSyntheticsAPITestWithDefaults()

	syntheticsTest.SetName(state.Name.ValueString())

	if !state.Subtype.IsNull() {
		syntheticsTest.SetSubtype(datadogV1.SyntheticsTestDetailsSubType(state.Subtype.ValueString()))
	} else {
		syntheticsTest.SetSubtype(datadogV1.SYNTHETICSTESTDETAILSSUBTYPE_HTTP)
	}

	request := datadogV1.SyntheticsTestRequest{}
	requestDefinition := state.RequestDefinition[0]
	if !requestDefinition.Method.IsNull() {
		request.SetMethod(requestDefinition.Method.ValueString())
	}
	if !requestDefinition.Url.IsNull() {
		request.SetUrl(requestDefinition.Url.ValueString())
	}
	// Only set the body if the request method allows it
	body := requestDefinition.Body
	if !body.IsNull() && body.ValueString() != "" {
		method := requestDefinition.Method
		httpVersion := requestDefinition.HttpVersion
		if !method.IsNull() && (method.ValueString() == "GET" || method.ValueString() == "HEAD" || method.ValueString() == "DELETE") && (!httpVersion.IsNull() && httpVersion.ValueString() != "http1") {
			diags.AddError("body", fmt.Sprintf("[WARN] body is not valid for %s requests. It'll be ignored.", method.ValueString()))
		} else {
			request.SetBody(body.ValueString())
		}
	}
	if !requestDefinition.BodyType.IsNull() {
		request.SetBodyType(datadogV1.SyntheticsTestRequestBodyType(requestDefinition.BodyType.ValueString()))
	}
	if len(state.RequestFile) > 0 {
		request.SetFiles(buildDatadogBodyFiles(state.RequestFile))
	}
	if !requestDefinition.Timeout.IsNull() {
		request.SetTimeout(float64(requestDefinition.Timeout.ValueInt64()))
	}
	if !requestDefinition.Host.IsNull() {
		request.SetHost(requestDefinition.Host.ValueString())
	}
	if !requestDefinition.Port.IsNull() {
		port := requestDefinition.Port.ValueString()
		request.SetPort(datadogV1.SyntheticsTestRequestPort{
			SyntheticsTestRequestVariablePort: &port,
		})
	}
	if !requestDefinition.DnsServer.IsNull() {
		request.SetDnsServer(requestDefinition.DnsServer.ValueString())
	}
	if !requestDefinition.DnsServerPort.IsNull() {
		request.SetDnsServerPort(requestDefinition.DnsServerPort.ValueString())
	}
	if !requestDefinition.NoSavingResponseBody.IsNull() {
		request.SetNoSavingResponseBody(requestDefinition.NoSavingResponseBody.ValueBool())
	}
	if !requestDefinition.NumberOfPackets.IsNull() {
		request.SetNumberOfPackets(int32(requestDefinition.NumberOfPackets.ValueInt64()))
	}
	if !requestDefinition.ShouldTrackHops.IsNull() {
		request.SetShouldTrackHops(requestDefinition.ShouldTrackHops.ValueBool())
	}
	if !requestDefinition.Servername.IsNull() {
		request.SetServername(requestDefinition.Servername.ValueString())
	}
	if !requestDefinition.Message.IsNull() {
		request.SetMessage(requestDefinition.Message.ValueString())
	}
	if !requestDefinition.CallType.IsNull() {
		request.SetCallType(datadogV1.SyntheticsTestCallType(requestDefinition.CallType.ValueString()))
	}
	if syntheticsTest.GetSubtype() == "grpc" {
		request.SetService(requestDefinition.Service.ValueString())
	}
	if !requestDefinition.PersistCookies.IsNull() {
		request.SetPersistCookies(requestDefinition.PersistCookies.ValueBool())
	}
	if !requestDefinition.ProtoJsonDescriptor.IsNull() {
		request.SetCompressedJsonDescriptor(compressAndEncodeValue(requestDefinition.ProtoJsonDescriptor.ValueString()))
	}
	if !requestDefinition.PlainProtoFile.IsNull() {
		request.SetCompressedProtoFile(compressAndEncodeValue(requestDefinition.PlainProtoFile.ValueString()))
	}

	if len(state.RequestClientCertificate) > 0 {
		request.SetCertificate(buildDatadogRequestCertificates(state.RequestClientCertificate[0]))
	}

	if !state.RequestHeaders.IsNull() {
		request.SetHeaders(terraformMapToStringMap(ctx, state.RequestHeaders))
	}
	if !state.RequestQuery.IsNull() {
		request.SetQuery(terraformMapToStringMap(ctx, state.RequestQuery))
	}
	if len(state.RequestBasicAuth) > 0 {
		basicAuth, basicAuthDiags := buildDatadogBasicAuth(state.RequestBasicAuth[0])
		diags.Append(basicAuthDiags...)
		request.SetBasicAuth(basicAuth)
	}
	if len(state.RequestProxy) > 0 {
		request.SetProxy(buildDatadogTestRequestProxy(ctx, state.RequestProxy[0]))
	}
	if !state.RequestMetadata.IsNull() {
		request.SetMetadata(terraformMapToStringMap(ctx, state.RequestMetadata))
	}

	config := datadogV1.NewSyntheticsAPITestConfigWithDefaults()
	if syntheticsTest.GetSubtype() != "multi" {
		config.SetRequest(request)
	}

	config.Assertions = []datadogV1.SyntheticsAssertion{}
	if len(state.Assertion) > 0 {
		assertions, assertionDiags := buildDatadogAssertions(state.Assertion)
		diags.Append(assertionDiags...)
		config.SetAssertions(assertions)
	}

	config.SetConfigVariables(buildDatadogConfigVariables(state.ConfigVariable))

	if !state.VariablesFromScript.IsNull() {
		config.SetVariablesFromScript(state.VariablesFromScript.ValueString())
	}

	if len(state.ApiStep) > 0 && syntheticsTest.GetSubtype() == "multi" {
		steps := make([]datadogV1.SyntheticsAPIStep, len(state.ApiStep))
		for _, stateStep := range state.ApiStep {
			step := datadogV1.SyntheticsAPIStep{}

			stepSubtype := stateStep.Subtype.ValueString()

			if stepSubtype == "" || stepSubtype == "http" || stepSubtype == "grpc" {
				step.SyntheticsAPITestStep = datadogV1.NewSyntheticsAPITestStepWithDefaults()
				step.SyntheticsAPITestStep.SetName(stateStep.Name.ValueString())
				step.SyntheticsAPITestStep.SetSubtype(datadogV1.SyntheticsAPITestStepSubtype(stepSubtype))

				extractedValues := buildDatadogExtractedValues(stateStep.ExtractedValue)
				step.SyntheticsAPITestStep.SetExtractedValues(extractedValues)

				assertions := stateStep.Assertion
				stepAssertions, assertionDiags := buildDatadogAssertions(assertions)
				diags.Append(assertionDiags...)
				step.SyntheticsAPITestStep.SetAssertions(stepAssertions)

				request := datadogV1.SyntheticsTestRequest{}
				if len(stateStep.RequestDefinition) > 0 {
					requestDefinition := stateStep.RequestDefinition[0]
					method := requestDefinition.Method.ValueString()
					request.SetMethod(method)
					request.SetTimeout(float64(requestDefinition.Timeout.ValueInt64()))
					request.SetAllowInsecure(requestDefinition.AllowInsecure.ValueBool())
					if stepSubtype == "grpc" {
						request.SetHost(requestDefinition.Host.ValueString())
						port := requestDefinition.Port.ValueString()
						request.SetPort(datadogV1.SyntheticsTestRequestPort{
							SyntheticsTestRequestVariablePort: &port,
						})
						request.SetService(requestDefinition.Service.ValueString())
						request.SetMessage(requestDefinition.Message.ValueString())
						if requestDefinition.CallType.ValueString() != "" {
							request.SetCallType(datadogV1.SyntheticsTestCallType(requestDefinition.CallType.ValueString()))
						}
						if requestDefinition.PlainProtoFile.ValueString() != "" {
							request.SetCompressedProtoFile(compressAndEncodeValue(requestDefinition.PlainProtoFile.ValueString()))
						}
					} else if stepSubtype == "http" {
						request.SetUrl(requestDefinition.Url.ValueString())
						httpVersion := requestDefinition.HttpVersion.ValueString()
						if httpVersion != "" {
							request.SetHttpVersion(datadogV1.SyntheticsTestOptionsHTTPVersion(httpVersion))
						}
						body := requestDefinition.Body.ValueString()
						if body != "" {
							if (method == "GET" || method == "HEAD" || method == "DELETE") && httpVersion != "http1" {
								diags.AddWarning("Invalid body", fmt.Sprintf("body is not valid for %s requests. It'll be ignored.", method))
							} else {
								request.SetBody(body)
							}
						}
						request.SetFollowRedirects(requestDefinition.FollowRedirects.ValueBool())
						request.SetPersistCookies(requestDefinition.PersistCookies.ValueBool())
						request.SetNoSavingResponseBody(requestDefinition.NoSavingResponseBody.ValueBool())
						if requestDefinition.BodyType.ValueString() != "" {
							request.SetBodyType(datadogV1.SyntheticsTestRequestBodyType(requestDefinition.BodyType.ValueString()))
						}
						if len(stateStep.RequestFile) > 0 {
							request.SetFiles(buildDatadogBodyFiles(stateStep.RequestFile))
						}
					}
				}

				// TODO: find a good new implementation for this. We should be able to do something else than getting the raw config.
				// Override the request client certificate with the one from the config
				// configCertContent, configKeyContent := getConfigCertAndKeyContent(d, i)
				if len(stateStep.RequestClientCertificate) > 0 {
					request.SetCertificate(buildDatadogRequestCertificates(stateStep.RequestClientCertificate[0]))
				}

				if !stateStep.RequestHeaders.IsNull() {
					request.SetHeaders(terraformMapToStringMap(ctx, stateStep.RequestHeaders))
				}
				if !stateStep.RequestQuery.IsNull() {
					request.SetQuery(terraformMapToStringMap(ctx, stateStep.RequestQuery))
				}
				if len(stateStep.RequestBasicAuth) > 0 {
					basicAuth, basicAuthDiags := buildDatadogBasicAuth(stateStep.RequestBasicAuth[0])
					diags.Append(basicAuthDiags...)
					request.SetBasicAuth(basicAuth)
				}
				if len(stateStep.RequestProxy) > 0 {
					request.SetProxy(buildDatadogTestRequestProxy(ctx, stateStep.RequestProxy[0]))
				}
				if !stateStep.RequestMetadata.IsNull() {
					request.SetMetadata(terraformMapToStringMap(ctx, stateStep.RequestMetadata))
				}

				step.SyntheticsAPITestStep.SetRequest(request)

				step.SyntheticsAPITestStep.SetAllowFailure(stateStep.AllowFailure.ValueBool())
				step.SyntheticsAPITestStep.SetExitIfSucceed(stateStep.ExitIfSucceed.ValueBool())
				step.SyntheticsAPITestStep.SetIsCritical(stateStep.IsCritical.ValueBool())

				optionsRetry := datadogV1.SyntheticsTestOptionsRetry{}
				if len(stateStep.Retry) > 0 {
					retry := stateStep.Retry[0]

					if !retry.Count.IsNull() {
						optionsRetry.SetCount(retry.Count.ValueInt64())
					}
					if !retry.Interval.IsNull() {
						optionsRetry.SetInterval(float64(retry.Interval.ValueInt64()))
					}
					step.SyntheticsAPITestStep.SetRetry(optionsRetry)
				}
			} else if stepSubtype == "wait" {
				step.SyntheticsAPIWaitStep = datadogV1.NewSyntheticsAPIWaitStepWithDefaults()
				step.SyntheticsAPIWaitStep.SetName(stateStep.Name.ValueString())
				step.SyntheticsAPIWaitStep.SetSubtype(datadogV1.SyntheticsAPIWaitStepSubtype(stepSubtype))
				step.SyntheticsAPIWaitStep.SetValue(int32(stateStep.Value.ValueInt64()))
			}

			steps = append(steps, step)
		}

		config.SetSteps(steps)
	}

	options := buildDatadogTestOptions(ctx, *state)
	syntheticsTest.SetConfig(*config)
	syntheticsTest.SetOptions(*options)
	syntheticsTest.SetMessage(state.Message.ValueString())
	syntheticsTest.SetStatus(datadogV1.SyntheticsTestPauseStatus(state.Status.ValueString()))

	if len(state.Locations.Elements()) > 0 {
		syntheticsTest.SetLocations(terraformSetToStringArray(ctx, state.Locations))
	}

	if len(state.Tags.Elements()) > 0 {
		syntheticsTest.SetTags(terraformListToStringArray(ctx, state.Tags))
	}

	return syntheticsTest, diags
}

func (r *syntheticsTestResource) buildDatadogSyntheticsBrowserTest(ctx context.Context, state *syntheticsTestModel) (*datadogV1.SyntheticsBrowserTest, diag.Diagnostics) {
	return nil, nil
}

func (r *syntheticsTestResource) buildDatadogSyntheticsMobileTest(ctx context.Context, state *syntheticsTestModel) (*datadogV1.SyntheticsMobileTest, diag.Diagnostics) {
	return nil, nil
}

func buildDatadogAssertions(assertions []syntheticsTestAssertionModel) ([]datadogV1.SyntheticsAssertion, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	requestAssertions := make([]datadogV1.SyntheticsAssertion, len(assertions))

	for i, assertion := range assertions {
		if assertion.Type.ValueString() == string(datadogV1.SYNTHETICSASSERTIONJAVASCRIPTTYPE_JAVASCRIPT) {
			// Handling the case for javascript assertion that does not contains any `operator`
			assertionJavascript := datadogV1.NewSyntheticsAssertionJavascriptWithDefaults()
			assertionJavascript.SetType(datadogV1.SYNTHETICSASSERTIONJAVASCRIPTTYPE_JAVASCRIPT)
			if !assertion.Code.IsNull() {
				assertionJavascript.SetCode(assertion.Code.ValueString())
			}
			requestAssertions[i] = datadogV1.SyntheticsAssertionJavascriptAsSyntheticsAssertion(assertionJavascript)
		} else if assertion.Operator.ValueString() == string(datadogV1.SYNTHETICSASSERTIONJSONSCHEMAOPERATOR_VALIDATES_JSON_SCHEMA) {
			assertionJSONSchemaTarget := datadogV1.NewSyntheticsAssertionJSONSchemaTargetWithDefaults()
			assertionJSONSchemaTarget.SetOperator(datadogV1.SyntheticsAssertionJSONSchemaOperator(assertion.Operator.ValueString()))
			assertionJSONSchemaTarget.SetType(datadogV1.SyntheticsAssertionType(assertion.Type.ValueString()))

			if len(assertion.TargetJSONSchema) > 0 {
				subTarget := datadogV1.NewSyntheticsAssertionJSONSchemaTargetTarget()
				jsonSchema := assertion.TargetJSONSchema[0]
				if jsonSchema.JSONSchema.ValueString() != "" {
					subTarget.SetJsonSchema(jsonSchema.JSONSchema.ValueString())
				}
				if jsonSchema.MetaSchema.ValueString() != "" {
					if metaSchema, err := datadogV1.NewSyntheticsAssertionJSONSchemaMetaSchemaFromValue(jsonSchema.MetaSchema.ValueString()); err == nil {
						subTarget.SetMetaSchema(*metaSchema)
					} else {
						diags.AddError("Invalid meta schema", fmt.Sprintf("Error converting json schema meta schema: %v", err))
					}
				}
				assertionJSONSchemaTarget.SetTarget(*subTarget)
			}
			// TODO: move this to a proper validator
			if !assertion.Target.IsNull() {
				diags.AddWarning("Invalid target", "Target shouldn't be specified for validateJSONSchema operator, only targetJSONSchema")
			}
			requestAssertions[i] = datadogV1.SyntheticsAssertionJSONSchemaTargetAsSyntheticsAssertion(assertionJSONSchemaTarget)
		} else if assertion.Operator.ValueString() == string(datadogV1.SYNTHETICSASSERTIONJSONPATHOPERATOR_VALIDATES_JSON_PATH) {
			assertionJSONPathTarget := datadogV1.NewSyntheticsAssertionJSONPathTargetWithDefaults()
			assertionJSONPathTarget.SetOperator(datadogV1.SyntheticsAssertionJSONPathOperator(assertion.Operator.ValueString()))
			assertionJSONPathTarget.SetType(datadogV1.SyntheticsAssertionType(assertion.Type.ValueString()))
			if assertion.Property.ValueString() != "" {
				assertionJSONPathTarget.SetProperty(assertion.Property.ValueString())
			}
			if len(assertion.TargetJSONPath) > 0 {
				subTarget := datadogV1.NewSyntheticsAssertionJSONPathTargetTarget()
				targetJsonPath := assertion.TargetJSONPath[0]
				if !targetJsonPath.JSONPath.IsNull() {
					subTarget.SetJsonPath(targetJsonPath.JSONPath.ValueString())
				}
				if !targetJsonPath.Operator.IsNull() {
					subTarget.SetOperator(targetJsonPath.Operator.ValueString())
				}
				if !targetJsonPath.TargetValue.IsNull() {
					targetValue := targetJsonPath.TargetValue.ValueString()
					operator := datadogV1.SyntheticsAssertionOperator(targetJsonPath.Operator.ValueString())
					switch operator {
					case datadogV1.SYNTHETICSASSERTIONOPERATOR_IS_UNDEFINED:
						// no target value must be set for isUndefined operator
					case datadogV1.SYNTHETICSASSERTIONOPERATOR_LESS_THAN,
						datadogV1.SYNTHETICSASSERTIONOPERATOR_MORE_THAN:
						if match, _ := regexp.MatchString("{{\\s*([^{}]*?)\\s*}}", targetValue); match {
							subTarget.SetTargetValue(targetValue)
						} else {
							if floatValue, err := strconv.ParseFloat(targetValue, 64); err == nil {
								subTarget.SetTargetValue(floatValue)
							}
						}
					default:
						subTarget.SetTargetValue(targetValue)
					}
				}
				if !targetJsonPath.ElementsOperator.IsNull() {
					subTarget.SetElementsOperator(targetJsonPath.ElementsOperator.ValueString())
				}
				assertionJSONPathTarget.SetTarget(*subTarget)
			}
			// TODO: move this to a proper validator
			if !assertion.Target.IsNull() {
				diags.AddWarning("Invalid target", "Target shouldn't be specified for validatesJSONPath operator, only targetJSONPath")
			}
			requestAssertions[i] = datadogV1.SyntheticsAssertionJSONPathTargetAsSyntheticsAssertion(assertionJSONPathTarget)
		} else if assertion.Operator.ValueString() == string(datadogV1.SYNTHETICSASSERTIONXPATHOPERATOR_VALIDATES_X_PATH) {
			assertionXPathTarget := datadogV1.NewSyntheticsAssertionXPathTargetWithDefaults()
			assertionXPathTarget.SetOperator(datadogV1.SyntheticsAssertionXPathOperator(assertion.Operator.ValueString()))
			assertionXPathTarget.SetType(datadogV1.SyntheticsAssertionType(assertion.Type.ValueString()))
			if assertion.Property.ValueString() != "" {
				assertionXPathTarget.SetProperty(assertion.Property.ValueString())
			}
			if len(assertion.TargetXPath) > 0 {
				subTarget := datadogV1.NewSyntheticsAssertionXPathTargetTarget()
				xPath := assertion.TargetXPath[0]
				if !xPath.XPath.IsNull() {
					subTarget.SetXPath(xPath.XPath.ValueString())
				}
				if !xPath.Operator.IsNull() {
					subTarget.SetOperator(xPath.Operator.ValueString())
				}
				if !xPath.TargetValue.IsNull() {
					targetValue := xPath.TargetValue.ValueString()
					operator := datadogV1.SyntheticsAssertionOperator(xPath.Operator.ValueString())
					switch operator {
					case datadogV1.SYNTHETICSASSERTIONOPERATOR_LESS_THAN,
						datadogV1.SYNTHETICSASSERTIONOPERATOR_MORE_THAN:
						if match, _ := regexp.MatchString("{{\\s*([^{}]*?)\\s*}}", targetValue); match {
							subTarget.SetTargetValue(targetValue)
						} else {
							if floatValue, err := strconv.ParseFloat(targetValue, 64); err == nil {
								subTarget.SetTargetValue(floatValue)
							}
						}
					default:
						subTarget.SetTargetValue(targetValue)
					}
				}

				assertionXPathTarget.SetTarget(*subTarget)
			}
			// TODO: move this to a proper validator
			if !assertion.Target.IsNull() {
				diags.AddWarning("Invalid target", "Target shouldn't be specified for validatesXPath operator, only targetXPath")
			}
			requestAssertions[i] = datadogV1.SyntheticsAssertionXPathTargetAsSyntheticsAssertion(assertionXPathTarget)
		} else {
			assertionTarget := datadogV1.NewSyntheticsAssertionTargetWithDefaults()
			assertionTarget.SetOperator(datadogV1.SyntheticsAssertionOperator(assertion.Operator.ValueString()))
			assertionTarget.SetType(datadogV1.SyntheticsAssertionType(assertion.Type.ValueString()))
			if assertion.Property.ValueString() != "" {
				assertionTarget.SetProperty(assertion.Property.ValueString())
			}
			if !assertion.Target.IsNull() {
				if isTargetOfTypeInt(assertionTarget.GetType(), assertionTarget.GetOperator()) {
					targetInt, _ := strconv.Atoi(assertion.Target.ValueString())
					assertionTarget.SetTarget(targetInt)
				} else if assertionTarget.GetType() == datadogV1.SYNTHETICSASSERTIONTYPE_PACKET_LOSS_PERCENTAGE {
					targetFloat, _ := strconv.ParseFloat(assertion.Target.ValueString(), 64)
					assertionTarget.SetTarget(targetFloat)
				} else {
					assertionTarget.SetTarget(assertion.Target.ValueString())
				}
			}
			if assertion.TimingsScope.ValueString() != "" {
				assertionTarget.SetTimingsScope(datadogV1.SyntheticsAssertionTimingsScope(assertion.TimingsScope.ValueString()))
			}
			if len(assertion.TargetJSONSchema) > 0 {
				diags.AddWarning("Invalid target", "targetjsonschema shouldn't be specified for non-validatesJSONSchema operator, only target")
			}
			if len(assertion.TargetJSONPath) > 0 {
				diags.AddWarning("Invalid target", "targetjsonpath shouldn't be specified for non-validatesJSONPath operator, only target")
			}
			if len(assertion.TargetXPath) > 0 {
				diags.AddWarning("Invalid target", "targetxpath shouldn't be specified for non-validatesXPath operator, only target")
			}
			requestAssertions[i] = datadogV1.SyntheticsAssertionTargetAsSyntheticsAssertion(assertionTarget)
		}
	}

	return requestAssertions, diags
}

func buildTerraformAssertions(actualAssertions []datadogV1.SyntheticsAssertion) []syntheticsTestAssertionModel {
	localAssertions := make([]syntheticsTestAssertionModel, len(actualAssertions))
	for i, assertion := range actualAssertions {
		localAssertion := syntheticsTestAssertionModel{}
		if assertion.SyntheticsAssertionTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion.Operator = types.StringValue(string(*v))
			}
			if assertionTarget.HasProperty() {
				localAssertion.Property = types.StringValue(assertionTarget.GetProperty())
			}
			if target := assertionTarget.GetTarget(); target != nil {
				localAssertion.Target = types.StringValue(convertToString(target))
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion.Type = types.StringValue(string(*v))
			}
			if assertionTarget.HasTimingsScope() {
				localAssertion.TimingsScope = types.StringValue(string(assertionTarget.GetTimingsScope()))
			}
		} else if assertion.SyntheticsAssertionJSONSchemaTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionJSONSchemaTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion.Operator = types.StringValue(string(*v))
			}
			if target, ok := assertionTarget.GetTargetOk(); ok {
				localTarget := syntheticsTestAssertionTargetJSONSchemaModel{}
				if v, ok := target.GetJsonSchemaOk(); ok {
					localTarget.JSONSchema = types.StringValue(string(*v))
				}
				if v, ok := target.GetMetaSchemaOk(); ok {
					localTarget.MetaSchema = types.StringValue(string(*v))
				}
				localAssertion.TargetJSONSchema = []syntheticsTestAssertionTargetJSONSchemaModel{localTarget}
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion.Type = types.StringValue(string(*v))
			}
		} else if assertion.SyntheticsAssertionJSONPathTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionJSONPathTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion.Operator = types.StringValue(string(*v))
			}
			if assertionTarget.HasProperty() {
				localAssertion.Property = types.StringValue(assertionTarget.GetProperty())
			}
			if target, ok := assertionTarget.GetTargetOk(); ok {
				localTarget := syntheticsTestAssertionTargetJSONPathModel{}
				if v, ok := target.GetJsonPathOk(); ok {
					localTarget.JSONPath = types.StringValue(string(*v))
				}
				if v, ok := target.GetOperatorOk(); ok {
					localTarget.Operator = types.StringValue(string(*v))
				}
				if v, ok := target.GetTargetValueOk(); ok {
					localTarget.TargetValue = types.StringValue(fmt.Sprintf("%v", *v))
				}
				if v, ok := target.GetElementsOperatorOk(); ok {
					localTarget.ElementsOperator = types.StringValue(string(*v))
				}
				localAssertion.TargetJSONPath = []syntheticsTestAssertionTargetJSONPathModel{localTarget}
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion.Type = types.StringValue(string(*v))
			}
		} else if assertion.SyntheticsAssertionXPathTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionXPathTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion.Operator = types.StringValue(string(*v))
			}
			if assertionTarget.HasProperty() {
				localAssertion.Property = types.StringValue(assertionTarget.GetProperty())
			}
			if target, ok := assertionTarget.GetTargetOk(); ok {
				localTarget := syntheticsTestAssertionTargetXPathModel{}
				if v, ok := target.GetXPathOk(); ok {
					localTarget.XPath = types.StringValue(string(*v))
				}
				if v, ok := target.GetOperatorOk(); ok {
					localTarget.Operator = types.StringValue(string(*v))
				}
				if v, ok := target.GetTargetValueOk(); ok {
					localTarget.TargetValue = types.StringValue(fmt.Sprintf("%v", *v))
				}
				localAssertion.TargetXPath = []syntheticsTestAssertionTargetXPathModel{localTarget}
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion.Type = types.StringValue(string(*v))
			}
		} else if assertion.SyntheticsAssertionBodyHashTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionBodyHashTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion.Operator = types.StringValue(string(*v))
			}
			if target := assertionTarget.GetTarget(); target != nil {
				localAssertion.Target = types.StringValue(fmt.Sprintf("%v", target))
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion.Type = types.StringValue(string(*v))
			}
		} else if assertion.SyntheticsAssertionJavascript != nil {
			assertionTarget := assertion.SyntheticsAssertionJavascript
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion.Type = types.StringValue(string(*v))
			}
			if v, ok := assertionTarget.GetCodeOk(); ok {
				localAssertion.Code = types.StringValue(*v)
			}
		}
		localAssertions[i] = localAssertion
	}
	return localAssertions
}

func buildDatadogBasicAuth(basicAuth syntheticsTestRequestBasicAuthModel) (datadogV1.SyntheticsBasicAuth, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	result := datadogV1.SyntheticsBasicAuth{}

	switch basicAuth.Type.ValueString() {
	case "web":
		if basicAuth.Username.ValueString() != "" {
			webAuth := datadogV1.NewSyntheticsBasicAuthWebWithDefaults()
			webAuth.SetUsername(basicAuth.Username.ValueString())
			webAuth.SetPassword(basicAuth.Password.ValueString())
			result.SyntheticsBasicAuthWeb = webAuth
			return result, diags
		}
	case "sigv4":
		if basicAuth.AccessKey.ValueString() != "" {
			sigv4Auth := datadogV1.NewSyntheticsBasicAuthSigv4WithDefaults()
			sigv4Auth.SetAccessKey(basicAuth.AccessKey.ValueString())
			sigv4Auth.SetSecretKey(basicAuth.SecretKey.ValueString())
			if basicAuth.Region.ValueString() != "" {
				sigv4Auth.SetRegion(basicAuth.Region.ValueString())
			}
			if basicAuth.SessionToken.ValueString() != "" {
				sigv4Auth.SetSessionToken(basicAuth.SessionToken.ValueString())
			}
			if basicAuth.ServiceName.ValueString() != "" {
				sigv4Auth.SetServiceName(basicAuth.ServiceName.ValueString())
			}
			result.SyntheticsBasicAuthSigv4 = sigv4Auth
			return result, diags
		}
	case "ntlm":
		ntlmAuth := datadogV1.NewSyntheticsBasicAuthNTLMWithDefaults()
		if basicAuth.Username.ValueString() != "" {
			ntlmAuth.SetUsername(basicAuth.Username.ValueString())
		}
		if basicAuth.Password.ValueString() != "" {
			ntlmAuth.SetPassword(basicAuth.Password.ValueString())
		}
		if basicAuth.Domain.ValueString() != "" {
			ntlmAuth.SetDomain(basicAuth.Domain.ValueString())
		}
		if basicAuth.Workstation.ValueString() != "" {
			ntlmAuth.SetWorkstation(basicAuth.Workstation.ValueString())
		}
		result.SyntheticsBasicAuthNTLM = ntlmAuth
		return result, diags
	case "oauth-client":
		if basicAuth.AccessTokenUrl.ValueString() != "" &&
			basicAuth.ClientId.ValueString() != "" &&
			basicAuth.ClientSecret.ValueString() != "" {
			oauthClientAuth := datadogV1.NewSyntheticsBasicAuthOauthClientWithDefaults()
			oauthClientAuth.SetAccessTokenUrl(basicAuth.AccessTokenUrl.ValueString())
			oauthClientAuth.SetClientId(basicAuth.ClientId.ValueString())
			oauthClientAuth.SetClientSecret(basicAuth.ClientSecret.ValueString())
			oauthClientAuth.SetTokenApiAuthentication(datadogV1.SyntheticsBasicAuthOauthTokenApiAuthentication(basicAuth.TokenApiAuthentication.ValueString()))
			if basicAuth.Audience.ValueString() != "" {
				oauthClientAuth.SetAudience(basicAuth.Audience.ValueString())
			}
			if basicAuth.Scope.ValueString() != "" {
				oauthClientAuth.SetScope(basicAuth.Scope.ValueString())
			}
			if basicAuth.Resource.ValueString() != "" {
				oauthClientAuth.SetResource(basicAuth.Resource.ValueString())
			}
			result.SyntheticsBasicAuthOauthClient = oauthClientAuth
			return result, diags
		}
	case "oauth-rop":
		if basicAuth.AccessTokenUrl.ValueString() != "" &&
			basicAuth.Username.ValueString() != "" &&
			basicAuth.Password.ValueString() != "" {
			oauthRopAuth := datadogV1.NewSyntheticsBasicAuthOauthROPWithDefaults()
			oauthRopAuth.SetAccessTokenUrl(basicAuth.AccessTokenUrl.ValueString())
			if basicAuth.ClientId.ValueString() != "" {
				oauthRopAuth.SetClientId(basicAuth.ClientId.ValueString())
			}
			if basicAuth.ClientSecret.ValueString() != "" {
				oauthRopAuth.SetClientSecret(basicAuth.ClientSecret.ValueString())
			}
			oauthRopAuth.SetTokenApiAuthentication(datadogV1.SyntheticsBasicAuthOauthTokenApiAuthentication(basicAuth.TokenApiAuthentication.ValueString()))
			if basicAuth.Audience.ValueString() != "" {
				oauthRopAuth.SetAudience(basicAuth.Audience.ValueString())
			}
			if basicAuth.Scope.ValueString() != "" {
				oauthRopAuth.SetScope(basicAuth.Scope.ValueString())
			}
			if basicAuth.Resource.ValueString() != "" {
				oauthRopAuth.SetResource(basicAuth.Resource.ValueString())
			}
			oauthRopAuth.SetUsername(basicAuth.Username.ValueString())
			oauthRopAuth.SetPassword(basicAuth.Password.ValueString())
			result.SyntheticsBasicAuthOauthROP = oauthRopAuth
			return result, diags
		}
	case "digest":
		if basicAuth.Username.ValueString() != "" {
			digestAuth := datadogV1.NewSyntheticsBasicAuthDigestWithDefaults()
			digestAuth.SetUsername(basicAuth.Username.ValueString())
			digestAuth.SetPassword(basicAuth.Password.ValueString())
			result.SyntheticsBasicAuthDigest = digestAuth
			return result, diags
		}
	}

	diags.AddWarning("unrecognized basic auth type", fmt.Sprintf("unrecognized basic auth type %s", basicAuth.Type.ValueString()))
	return result, diags
}

func buildTerraformBasicAuth(basicAuth *datadogV1.SyntheticsBasicAuth) syntheticsTestRequestBasicAuthModel {
	localAuth := syntheticsTestRequestBasicAuthModel{}

	if basicAuth.SyntheticsBasicAuthWeb != nil {
		basicAuthWeb := basicAuth.SyntheticsBasicAuthWeb
		localAuth.Username = types.StringValue(basicAuthWeb.Username)
		localAuth.Password = types.StringValue(basicAuthWeb.Password)
		localAuth.Type = types.StringValue("web")
	}

	if basicAuth.SyntheticsBasicAuthSigv4 != nil {
		basicAuthSigv4 := basicAuth.SyntheticsBasicAuthSigv4
		localAuth.AccessKey = types.StringValue(basicAuthSigv4.AccessKey)
		localAuth.SecretKey = types.StringValue(basicAuthSigv4.SecretKey)
		if v, ok := basicAuthSigv4.GetRegionOk(); ok {
			localAuth.Region = types.StringValue(*v)
		}
		if v, ok := basicAuthSigv4.GetSessionTokenOk(); ok {
			localAuth.SessionToken = types.StringValue(*v)
		}
		if v, ok := basicAuthSigv4.GetServiceNameOk(); ok {
			localAuth.ServiceName = types.StringValue(*v)
		}
		localAuth.Type = types.StringValue("sigv4")
	}

	if basicAuth.SyntheticsBasicAuthNTLM != nil {
		basicAuthNtlm := basicAuth.SyntheticsBasicAuthNTLM
		if v, ok := basicAuthNtlm.GetUsernameOk(); ok {
			localAuth.Username = types.StringValue(*v)
		}
		if v, ok := basicAuthNtlm.GetPasswordOk(); ok {
			localAuth.Password = types.StringValue(*v)
		}
		if v, ok := basicAuthNtlm.GetDomainOk(); ok {
			localAuth.Domain = types.StringValue(*v)
		}
		if v, ok := basicAuthNtlm.GetWorkstationOk(); ok {
			localAuth.Workstation = types.StringValue(*v)
		}
		localAuth.Type = types.StringValue("ntlm")
	}

	if basicAuth.SyntheticsBasicAuthOauthClient != nil {
		basicAuthOauthClient := basicAuth.SyntheticsBasicAuthOauthClient
		localAuth.AccessTokenUrl = types.StringValue(basicAuthOauthClient.AccessTokenUrl)
		localAuth.ClientId = types.StringValue(basicAuthOauthClient.ClientId)
		localAuth.ClientSecret = types.StringValue(basicAuthOauthClient.ClientSecret)
		localAuth.TokenApiAuthentication = types.StringValue(string(basicAuthOauthClient.TokenApiAuthentication))
		if v, ok := basicAuthOauthClient.GetAudienceOk(); ok {
			localAuth.Audience = types.StringValue(*v)
		}
		if v, ok := basicAuthOauthClient.GetScopeOk(); ok {
			localAuth.Scope = types.StringValue(*v)
		}
		if v, ok := basicAuthOauthClient.GetResourceOk(); ok {
			localAuth.Resource = types.StringValue(*v)
		}
		localAuth.Type = types.StringValue("oauth-client")
	}
	if basicAuth.SyntheticsBasicAuthOauthROP != nil {
		basicAuthOauthROP := basicAuth.SyntheticsBasicAuthOauthROP
		localAuth.AccessTokenUrl = types.StringValue(basicAuthOauthROP.AccessTokenUrl)
		if v, ok := basicAuthOauthROP.GetClientIdOk(); ok {
			localAuth.ClientId = types.StringValue(*v)
		}
		if v, ok := basicAuthOauthROP.GetClientSecretOk(); ok {
			localAuth.ClientSecret = types.StringValue(*v)
		}
		localAuth.TokenApiAuthentication = types.StringValue(string(basicAuthOauthROP.TokenApiAuthentication))
		if v, ok := basicAuthOauthROP.GetAudienceOk(); ok {
			localAuth.Audience = types.StringValue(*v)
		}
		if v, ok := basicAuthOauthROP.GetScopeOk(); ok {
			localAuth.Scope = types.StringValue(*v)
		}
		if v, ok := basicAuthOauthROP.GetResourceOk(); ok {
			localAuth.Resource = types.StringValue(*v)
		}
		localAuth.Username = types.StringValue(basicAuthOauthROP.Username)
		localAuth.Password = types.StringValue(basicAuthOauthROP.Password)
		localAuth.Type = types.StringValue("oauth-rop")
	}

	if basicAuth.SyntheticsBasicAuthDigest != nil {
		basicAuthDigest := basicAuth.SyntheticsBasicAuthDigest
		localAuth.Username = types.StringValue(basicAuthDigest.Username)
		localAuth.Password = types.StringValue(basicAuthDigest.Password)
		localAuth.Type = types.StringValue("digest")
	}

	return localAuth
}

func buildDatadogBodyFiles(stateFiles []syntheticsTestRequestFileModel) []datadogV1.SyntheticsTestRequestBodyFile {
	requestFiles := make([]datadogV1.SyntheticsTestRequestBodyFile, len(stateFiles))
	for i, stateFile := range stateFiles {
		requestFile := datadogV1.SyntheticsTestRequestBodyFile{}

		requestFile.SetName(stateFile.Name.ValueString())
		requestFile.SetOriginalFileName(stateFile.OriginalFileName.ValueString())
		requestFile.SetType(stateFile.Type.ValueString())
		requestFile.SetSize(stateFile.Size.ValueInt64())

		if !stateFile.Content.IsNull() {
			requestFile.SetContent(stateFile.Content.ValueString())
		}

		// We aren't sure yet how to let the provider check if the file content was updated to upload it again.
		// Hence, the provider is uploading the file every time the resource is modified.
		// Always adding the bucket key to the request would prevent updating the file content.
		// Always omitting the existing bucket key from the request update the file every time the resource is updated.
		// We purposely choose the latter.
		// if bucketKey, ok := fileMap["bucket_key"]; ok && bucketKey != "" {
		// 	file.SetBucketKey(bucketKey.(string))
		// }

		requestFiles[i] = requestFile
	}
	return requestFiles
}

func buildTerraformBodyFiles(actualBodyFiles *[]datadogV1.SyntheticsTestRequestBodyFile, oldLocalBodyFiles []syntheticsTestRequestFileModel) []syntheticsTestRequestFileModel {
	localBodyFiles := make([]syntheticsTestRequestFileModel, len(*actualBodyFiles))
	for i, file := range *actualBodyFiles {
		localFile := syntheticsTestRequestFileModel{}
		if i < len(oldLocalBodyFiles) {
			// The file content is kept from the existing localFile from the state,
			// as the response from the backend contains the bucket key rather than the content.
			localFile = oldLocalBodyFiles[i]
		}
		localFile.Name = types.StringValue(file.GetName())
		localFile.OriginalFileName = types.StringValue(file.GetOriginalFileName())
		localFile.Type = types.StringValue(file.GetType())
		localFile.Size = types.Int64Value(file.GetSize())

		if bucketKey, ok := file.GetBucketKeyOk(); ok {
			localFile.BucketKey = types.StringValue(*bucketKey)
		}
		localBodyFiles[i] = localFile
	}
	return localBodyFiles
}

func buildDatadogConfigVariables(stateConfigVariables []syntheticsTestVariableModel) []datadogV1.SyntheticsConfigVariable {
	configVariables := make([]datadogV1.SyntheticsConfigVariable, len(stateConfigVariables))
	for _, configVariable := range stateConfigVariables {
		variable := datadogV1.SyntheticsConfigVariable{}

		variable.SetType(datadogV1.SyntheticsConfigVariableType(configVariable.Type.ValueString()))
		variable.SetName(configVariable.Name.ValueString())

		if variable.GetType() != "global" {
			variable.SetPattern(configVariable.Pattern.ValueString())
			variable.SetExample(configVariable.Example.ValueString())
			variable.SetSecure(configVariable.Secure.ValueBool())
		}

		if configVariable.Id.ValueString() != "" {
			variable.SetId(configVariable.Id.ValueString())
		}
		configVariables = append(configVariables, variable)
	}
	return configVariables
}

func buildTerraformConfigVariables(configVariables []datadogV1.SyntheticsConfigVariable, oldConfigVariables []syntheticsTestVariableModel) []syntheticsTestVariableModel {
	localConfigVariables := make([]syntheticsTestVariableModel, len(configVariables))
	for i, configVariable := range configVariables {
		localVariable := syntheticsTestVariableModel{}
		if v, ok := configVariable.GetTypeOk(); ok {
			localVariable.Type = types.StringValue(string(*v))
		}
		if v, ok := configVariable.GetNameOk(); ok {
			localVariable.Name = types.StringValue(*v)
		}
		if v, ok := configVariable.GetSecureOk(); ok {
			localVariable.Secure = types.BoolValue(*v)
		}

		if configVariable.GetType() != "global" {
			// If the variable is secure, the example and pattern are not returned by the API,
			// so we need to keep the values from the Terraform state.
			if localVariable.Secure.ValueBool() {
				if i < len(oldConfigVariables) {
					localVariable.Example = oldConfigVariables[i].Example
					localVariable.Pattern = oldConfigVariables[i].Pattern
				}
			} else {
				if v, ok := configVariable.GetExampleOk(); ok {
					localVariable.Example = types.StringValue(*v)
				}
				if v, ok := configVariable.GetPatternOk(); ok {
					localVariable.Pattern = types.StringValue(*v)
				}
			}
		}
		if v, ok := configVariable.GetIdOk(); ok {
			localVariable.Id = types.StringValue(*v)
		}
		localConfigVariables[i] = localVariable
	}
	return localConfigVariables
}

func buildDatadogExtractedValues(stateExtractedValues []syntheticsTestAPIStepExtractedValueModel) []datadogV1.SyntheticsParsingOptions {
	extractedValues := make([]datadogV1.SyntheticsParsingOptions, len(stateExtractedValues))
	for i, stateExtractedValue := range stateExtractedValues {
		extractedValue := datadogV1.SyntheticsParsingOptions{}

		extractedValue.SetName(stateExtractedValue.Name.ValueString())
		extractedValue.SetType(datadogV1.SyntheticsLocalVariableParsingOptionsType(stateExtractedValue.Type.ValueString()))
		if stateExtractedValue.Field.ValueString() != "" {
			extractedValue.SetField(stateExtractedValue.Field.ValueString())
		}

		parser := datadogV1.SyntheticsVariableParser{}
		parser.SetType(datadogV1.SyntheticsGlobalVariableParserType(stateExtractedValue.Parser[0].Type.ValueString()))
		if stateExtractedValue.Parser[0].Value.ValueString() != "" {
			parser.SetValue(stateExtractedValue.Parser[0].Value.ValueString())
		}

		extractedValue.SetParser(parser)
		extractedValue.SetSecure(stateExtractedValue.Secure.ValueBool())

		extractedValues[i] = extractedValue
	}
	return extractedValues
}

func buildTerraformExtractedValues(extractedValues []datadogV1.SyntheticsParsingOptions) []syntheticsTestAPIStepExtractedValueModel {
	localExtractedValues := make([]syntheticsTestAPIStepExtractedValueModel, len(extractedValues))
	for i, extractedValue := range extractedValues {
		localExtractedValue := syntheticsTestAPIStepExtractedValueModel{
			Name:   types.StringValue(extractedValue.GetName()),
			Type:   types.StringValue(string(extractedValue.GetType())),
			Field:  types.StringValue(extractedValue.GetField()),
			Secure: types.BoolValue(extractedValue.GetSecure()),
		}

		parser := extractedValue.GetParser()
		localParser := syntheticsTestAPIStepExtractedValueParserModel{
			Type:  types.StringValue(string(parser.GetType())),
			Value: types.StringValue(parser.GetValue()),
		}
		localExtractedValue.Parser = []syntheticsTestAPIStepExtractedValueParserModel{localParser}

		localExtractedValues[i] = localExtractedValue
	}
	return localExtractedValues
}

func buildDatadogRequestCertificates(certModel syntheticsTestRequestClientCertificateModel) datadogV1.SyntheticsTestRequestCertificate {
	cert := datadogV1.SyntheticsTestRequestCertificateItem{}
	key := datadogV1.SyntheticsTestRequestCertificateItem{}

	if len(certModel.Cert) > 0 {
		clientCert := certModel.Cert[0]
		if !clientCert.Content.IsNull() {
			// only set the certificate content if it is not an already hashed string
			// this is needed for the update function that receives the data from the state
			// and not from the config. So we get a hash of the certificate and not it's real
			// value.
			if isHash := isCertHash(clientCert.Content.ValueString()); !isHash {
				cert.SetContent(clientCert.Content.ValueString())
			}
		}
		if !clientCert.Filename.IsNull() {
			cert.SetFilename(clientCert.Filename.ValueString())
		}
	}

	if len(certModel.Key) > 0 {
		clientKey := certModel.Key[0]
		if !clientKey.Content.IsNull() {
			// only set the key content if it is not an already hashed string
			if isHash := isCertHash(clientKey.Content.ValueString()); !isHash {
				key.SetContent(clientKey.Content.ValueString())
			}
		}
		if !clientKey.Filename.IsNull() {
			key.SetFilename(clientKey.Filename.ValueString())
		}
	}

	return datadogV1.SyntheticsTestRequestCertificate{
		Cert: &cert,
		Key:  &key,
	}
}

func buildTerraformRequestCertificates(clientCertificate datadogV1.SyntheticsTestRequestCertificate, oldClientCertificates []syntheticsTestRequestClientCertificateModel) syntheticsTestRequestClientCertificateModel {
	localCertificate := syntheticsTestRequestClientCertificateModel{
		Cert: []syntheticsTestRequestClientCertificateItemModel{{}},
		Key:  []syntheticsTestRequestClientCertificateItemModel{{}},
	}

	cert := clientCertificate.GetCert()
	localCertificate.Cert[0].Filename = types.StringValue(cert.GetFilename())

	key := clientCertificate.GetKey()
	localCertificate.Key[0].Filename = types.StringValue(key.GetFilename())

	// The content of the client certificate is write-only, so it is not returned by the API.
	// To prevent unnecessary diffs and avoid storing the value in clear in the state,
	// we store a hash of the value.
	if len(oldClientCertificates) > 0 {
		if len(oldClientCertificates[0].Cert) > 0 {
			localCertificate.Cert[0].Content = oldClientCertificates[0].Cert[0].Content
		}
		if len(oldClientCertificates[0].Key) > 0 {
			localCertificate.Key[0].Content = oldClientCertificates[0].Key[0].Content
		}
	}

	return localCertificate
}

func buildDatadogTestOptions(ctx context.Context, state syntheticsTestModel) *datadogV1.SyntheticsTestOptions {
	options := datadogV1.NewSyntheticsTestOptions()

	if len(state.OptionsList) > 0 {
		// common browser and API tests options
		optionsList := state.OptionsList[0]
		if !optionsList.TickEvery.IsNull() {
			options.SetTickEvery(int64(optionsList.TickEvery.ValueInt64()))
		}
		if !optionsList.HttpVersion.IsNull() {
			options.SetHttpVersion(datadogV1.SyntheticsTestOptionsHTTPVersion(optionsList.HttpVersion.ValueString()))
		}
		if !optionsList.AcceptSelfSigned.IsNull() {
			options.SetAcceptSelfSigned(optionsList.AcceptSelfSigned.ValueBool())
		}
		if !optionsList.CheckCertificateRevocation.IsNull() {
			options.SetCheckCertificateRevocation(optionsList.CheckCertificateRevocation.ValueBool())
		}
		if !optionsList.MinLocationFailed.IsNull() {
			options.SetMinLocationFailed(int64(optionsList.MinLocationFailed.ValueInt64()))
		}
		if !optionsList.MinFailureDuration.IsNull() {
			options.SetMinFailureDuration(int64(optionsList.MinFailureDuration.ValueInt64()))
		}
		if !optionsList.FollowRedirects.IsNull() {
			options.SetFollowRedirects(optionsList.FollowRedirects.ValueBool())
		}
		if !optionsList.AllowInsecure.IsNull() {
			options.SetAllowInsecure(optionsList.AllowInsecure.ValueBool())
		}

		if len(optionsList.Scheduling) > 0 {
			optionsScheduling := datadogV1.SyntheticsTestOptionsScheduling{}
			scheduling := optionsList.Scheduling[0]
			if len(scheduling.Timeframes) > 0 {
				var timeframes []datadogV1.SyntheticsTestOptionsSchedulingTimeframe
				for _, tf := range scheduling.Timeframes {
					timeframe := datadogV1.SyntheticsTestOptionsSchedulingTimeframe{
						Day:  int32(tf.Day.ValueInt64()),
						From: tf.From.ValueString(),
						To:   tf.To.ValueString(),
					}
					timeframes = append(timeframes, timeframe)
				}
				optionsScheduling.SetTimeframes(timeframes)
			}
			if !scheduling.Timezone.IsNull() {
				optionsScheduling.SetTimezone(scheduling.Timezone.ValueString())
			}
			options.SetScheduling(optionsScheduling)
		}

		if len(optionsList.Retry) > 0 {
			optionsRetry := datadogV1.SyntheticsTestOptionsRetry{}
			retry := optionsList.Retry[0]
			if !retry.Count.IsNull() {
				optionsRetry.SetCount(retry.Count.ValueInt64())
			}
			if !retry.Interval.IsNull() {
				optionsRetry.SetInterval(float64(retry.Interval.ValueInt64()))
			}
			options.SetRetry(optionsRetry)
		}

		if len(optionsList.MonitorOptions) > 0 {
			monitorOptions := optionsList.MonitorOptions[0]
			optionsMonitorOptions := datadogV1.SyntheticsTestOptionsMonitorOptions{}

			if !monitorOptions.RenotifyInterval.IsNull() {
				renotifyInterval := monitorOptions.RenotifyInterval.ValueInt64()
				optionsMonitorOptions.SetRenotifyInterval(renotifyInterval)
				if !monitorOptions.RenotifyOccurrences.IsNull() && renotifyInterval != 0 {
					optionsMonitorOptions.SetRenotifyOccurrences(monitorOptions.RenotifyOccurrences.ValueInt64())
				}
			}
			options.SetMonitorOptions(optionsMonitorOptions)
		}

		if !optionsList.MonitorName.IsNull() {
			options.SetMonitorName(optionsList.MonitorName.ValueString())
		}

		if !optionsList.MonitorPriority.IsNull() {
			options.SetMonitorPriority(int32(optionsList.MonitorPriority.ValueInt64()))
		}

		if len(optionsList.RestrictedRoles.Elements()) > 0 {
			options.SetRestrictedRoles(terraformSetToStringArray(ctx, optionsList.RestrictedRoles))
		}

		if len(optionsList.CI) > 0 {
			ci := optionsList.CI[0]
			ciOptions := datadogV1.SyntheticsTestCiOptions{}
			ciOptions.SetExecutionRule(datadogV1.SyntheticsTestExecutionRule(ci.ExecutionRule.ValueString()))
			options.SetCi(ciOptions)
		}

		if !optionsList.IgnoreServerCertificateError.IsNull() {
			options.SetIgnoreServerCertificateError(optionsList.IgnoreServerCertificateError.ValueBool())
		}

		// browser tests specific options
		if !optionsList.NoScreenshot.IsNull() {
			options.SetNoScreenshot(optionsList.NoScreenshot.ValueBool())
		}

		if len(optionsList.RUMSettings) > 0 {
			settings := optionsList.RUMSettings[0]
			isEnabled := settings.IsEnabled.ValueBool()

			rumSettings := datadogV1.SyntheticsBrowserTestRumSettings{}

			if isEnabled {
				rumSettings.SetIsEnabled(true)

				if settings.ApplicationId.ValueString() != "" {
					rumSettings.SetApplicationId(settings.ApplicationId.ValueString())
				}

				if settings.ClientTokenId.ValueInt64() != 0 {
					rumSettings.SetClientTokenId(settings.ClientTokenId.ValueInt64())
				}
			} else {
				rumSettings.SetIsEnabled(false)
			}

			options.SetRumSettings(rumSettings)
		}

		if !optionsList.DisableCSP.IsNull() {
			options.SetDisableCsp(optionsList.DisableCSP.ValueBool())
		}

		if !optionsList.DisableCORS.IsNull() {
			options.SetDisableCors(optionsList.DisableCORS.ValueBool())
		}

		if !optionsList.InitialNavigationTimeout.IsNull() {
			options.SetInitialNavigationTimeout(int64(optionsList.InitialNavigationTimeout.ValueInt64()))
		}

		if len(state.DeviceIds.Elements()) > 0 {
			options.SetDeviceIds(terraformListToStringArray(ctx, state.DeviceIds))
		}
	}

	return options
}

func buildTerraformTestOptions(ctx context.Context, actualOptions datadogV1.SyntheticsTestOptions) []syntheticsTestOptionsListModel {
	localOptions := syntheticsTestOptionsListModel{}

	if actualOptions.HasFollowRedirects() {
		localOptions.FollowRedirects = types.BoolValue(actualOptions.GetFollowRedirects())
	}
	if actualOptions.HasMinFailureDuration() {
		localOptions.MinFailureDuration = types.Int64Value(actualOptions.GetMinFailureDuration())
	}
	if actualOptions.HasMinLocationFailed() {
		localOptions.MinLocationFailed = types.Int64Value(actualOptions.GetMinLocationFailed())
	}
	if actualOptions.HasTickEvery() {
		localOptions.TickEvery = types.Int64Value(actualOptions.GetTickEvery())
	}
	if actualOptions.HasHttpVersion() {
		localOptions.HttpVersion = types.StringValue(string(actualOptions.GetHttpVersion()))
	}
	if actualOptions.HasAcceptSelfSigned() {
		localOptions.AcceptSelfSigned = types.BoolValue(actualOptions.GetAcceptSelfSigned())
	}
	if actualOptions.HasCheckCertificateRevocation() {
		localOptions.CheckCertificateRevocation = types.BoolValue(actualOptions.GetCheckCertificateRevocation())
	}
	if actualOptions.HasAllowInsecure() {
		localOptions.AllowInsecure = types.BoolValue(actualOptions.GetAllowInsecure())
	}

	if actualOptions.HasScheduling() {
		scheduling := actualOptions.GetScheduling()
		localScheduling := syntheticsTestAdvancedSchedulingModel{}
		timeFrames := scheduling.GetTimeframes()
		for _, tf := range timeFrames {
			timeframe := syntheticsTestAdvancedSchedulingTimeframesModel{
				From: types.StringValue(tf.GetFrom()),
				Day:  types.Int64Value(int64(tf.GetDay())),
				To:   types.StringValue(tf.GetTo()),
			}
			localScheduling.Timeframes = append(localScheduling.Timeframes, timeframe)
		}
		localScheduling.Timezone = types.StringValue(scheduling.GetTimezone())
		localOptions.Scheduling = []syntheticsTestAdvancedSchedulingModel{localScheduling}
	}

	if actualOptions.HasRetry() {
		retry := actualOptions.GetRetry()
		localRetry := syntheticsTestRetryModel{}
		if count, ok := retry.GetCountOk(); ok {
			localRetry.Count = types.Int64Value(*count)
		}
		if interval, ok := retry.GetIntervalOk(); ok {
			localRetry.Interval = types.Int64Value(int64(*interval))
		}
		localOptions.Retry = []syntheticsTestRetryModel{localRetry}
	}

	if actualOptions.HasMonitorOptions() {
		actualMonitorOptions := actualOptions.GetMonitorOptions()
		localMonitorOptions := syntheticsTestMonitorOptionsModel{}
		shouldUpdate := false

		if actualMonitorOptions.HasRenotifyInterval() {
			localMonitorOptions.RenotifyInterval = types.Int64Value(actualMonitorOptions.GetRenotifyInterval())
			shouldUpdate = true
		}
		if actualMonitorOptions.HasRenotifyOccurrences() {
			localMonitorOptions.RenotifyOccurrences = types.Int64Value(actualMonitorOptions.GetRenotifyOccurrences())
			shouldUpdate = true
		}
		if shouldUpdate {
			localOptions.MonitorOptions = []syntheticsTestMonitorOptionsModel{localMonitorOptions}
		}
	}

	if actualOptions.HasNoScreenshot() {
		localOptions.NoScreenshot = types.BoolValue(actualOptions.GetNoScreenshot())
	}
	if actualOptions.HasMonitorName() {
		localOptions.MonitorName = types.StringValue(actualOptions.GetMonitorName())
	}
	if actualOptions.HasMonitorPriority() {
		localOptions.MonitorPriority = types.Int64Value(int64(actualOptions.GetMonitorPriority()))
	}
	localOptions.RestrictedRoles, _ = types.SetValueFrom(ctx, types.StringType, actualOptions.GetRestrictedRoles())
	if actualOptions.HasCi() {
		actualCi := actualOptions.GetCi()
		localOptions.CI = []syntheticsTestCIModel{
			{ExecutionRule: types.StringValue(string(actualCi.GetExecutionRule()))},
		}
	}
	if rumSettings, ok := actualOptions.GetRumSettingsOk(); ok {
		localRumSettings := syntheticsTestRUMSettingsModel{}
		localRumSettings.IsEnabled = types.BoolValue(rumSettings.GetIsEnabled())
		if rumSettings.HasApplicationId() {
			localRumSettings.ApplicationId = types.StringValue(rumSettings.GetApplicationId())
		}
		if rumSettings.HasClientTokenId() {
			localRumSettings.ClientTokenId = types.Int64Value(rumSettings.GetClientTokenId())
		}
		localOptions.RUMSettings = []syntheticsTestRUMSettingsModel{localRumSettings}
	}
	if actualOptions.HasIgnoreServerCertificateError() {
		localOptions.IgnoreServerCertificateError = types.BoolValue(actualOptions.GetIgnoreServerCertificateError())
	}
	if actualOptions.HasDisableCsp() {
		localOptions.DisableCSP = types.BoolValue(actualOptions.GetDisableCsp())
	}
	if actualOptions.HasDisableCors() {
		localOptions.DisableCORS = types.BoolValue(actualOptions.GetDisableCors())
	}
	if actualOptions.HasInitialNavigationTimeout() {
		localOptions.InitialNavigationTimeout = types.Int64Value(actualOptions.GetInitialNavigationTimeout())
	}

	return []syntheticsTestOptionsListModel{localOptions}
}

func buildDatadogMobileTestOptions() {}

func buildTerraformMobileTestOptions() {}

func buildTerraformMobileTestSteps() {}

func completeSyntheticsTestRequest() {}

func buildTerraformTestRequest(ctx context.Context, request datadogV1.SyntheticsTestRequest) syntheticsTestRequestModel {
	localRequest := syntheticsTestRequestModel{}
	if request.HasBody() {
		localRequest.Body = types.StringValue(request.GetBody())
	}
	if request.HasBodyType() {
		localRequest.BodyType = types.StringValue(string(request.GetBodyType()))
	}
	if request.HasMethod() {
		localRequest.Method = types.StringValue(request.GetMethod())
	}
	if request.HasTimeout() {
		localRequest.Timeout = types.Int64Value(int64(request.GetTimeout()))
	}
	if request.HasUrl() {
		localRequest.Url = types.StringValue(request.GetUrl())
	}
	if request.HasHost() {
		localRequest.Host = types.StringValue(request.GetHost())
	}
	if request.HasPort() {
		var port = request.GetPort()
		if port.SyntheticsTestRequestNumericalPort != nil {
			localRequest.Port = types.StringValue(strconv.FormatInt(*port.SyntheticsTestRequestNumericalPort, 10))
		} else if port.SyntheticsTestRequestVariablePort != nil {
			localRequest.Port = types.StringValue(*port.SyntheticsTestRequestVariablePort)
		}
	}
	if request.HasDnsServer() {
		localRequest.DnsServer = types.StringValue(string(request.GetDnsServer()))
	}
	if request.HasDnsServerPort() {
		localRequest.DnsServerPort = types.StringValue(request.GetDnsServerPort())
	}
	if request.HasNoSavingResponseBody() {
		localRequest.NoSavingResponseBody = types.BoolValue(request.GetNoSavingResponseBody())
	}
	if request.HasNumberOfPackets() {
		localRequest.NumberOfPackets = types.Int64Value(int64(request.GetNumberOfPackets()))
	}
	if request.HasShouldTrackHops() {
		localRequest.ShouldTrackHops = types.BoolValue(request.GetShouldTrackHops())
	}
	if request.HasServername() {
		localRequest.Servername = types.StringValue(request.GetServername())
	}
	if request.HasMessage() {
		localRequest.Message = types.StringValue(request.GetMessage())
	}
	if request.HasCallType() {
		localRequest.CallType = types.StringValue(string(request.GetCallType()))
	}
	if request.HasService() {
		localRequest.Service = types.StringValue(request.GetService())
	}
	localRequest.CertificateDomains, _ = types.ListValueFrom(ctx, types.StringType, request.GetCertificateDomains())
	if request.HasPersistCookies() {
		localRequest.PersistCookies = types.BoolValue(request.GetPersistCookies())
	}
	if request.HasHttpVersion() {
		localRequest.HttpVersion = types.StringValue(string(request.GetHttpVersion()))
	}
	if request.HasCompressedJsonDescriptor() {
		localRequest.ProtoJsonDescriptor = types.StringValue(decompressAndDecodeValue(request.GetCompressedJsonDescriptor()))
	}
	if request.HasCompressedProtoFile() {
		localRequest.PlainProtoFile = types.StringValue(decompressAndDecodeValue(request.GetCompressedProtoFile()))
	}
	return localRequest
}

func buildDatadogTestRequestProxy(ctx context.Context, requestProxy syntheticsTestRequestProxyModel) datadogV1.SyntheticsTestRequestProxy {
	testRequestProxy := datadogV1.SyntheticsTestRequestProxy{}
	testRequestProxy.SetUrl(requestProxy.Url.ValueString())
	testRequestProxy.SetHeaders(terraformMapToStringMap(ctx, requestProxy.Headers))
	return testRequestProxy
}

func buildTerraformTestRequestProxy(proxy datadogV1.SyntheticsTestRequestProxy) syntheticsTestRequestProxyModel {
	return syntheticsTestRequestProxyModel{
		Url:     types.StringValue(proxy.GetUrl()),
		Headers: stringMapToTerraformMap(proxy.GetHeaders()),
	}
}

func buildTerraformAPITestStep(ctx context.Context, step datadogV1.SyntheticsAPIStep) syntheticsTestAPIStepModel {
	localStep := syntheticsTestAPIStepModel{}

	if step.SyntheticsAPITestStep != nil {
		localStep.Name = types.StringValue(step.SyntheticsAPITestStep.GetName())
		localStep.Subtype = types.StringValue(string(step.SyntheticsAPITestStep.GetSubtype()))

		apiTestStep := step.SyntheticsAPITestStep

		localStep.Assertion = buildTerraformAssertions(apiTestStep.GetAssertions())
		localStep.ExtractedValue = buildTerraformExtractedValues(apiTestStep.GetExtractedValues())

		stepRequest := apiTestStep.GetRequest()

		localRequest := syntheticsTestAPIStepRequestModel{
			syntheticsTestRequestModel: buildTerraformTestRequest(ctx, stepRequest),
		}
		localRequest.AllowInsecure = types.BoolValue(stepRequest.GetAllowInsecure())
		localRequest.FollowRedirects = types.BoolValue(stepRequest.GetFollowRedirects())
		if apiTestStep.GetSubtype() == "grpc" {
			// the schema defines a default value of `http_version` for any kind of step,
			// but it's not supported for `grpc` - so we save `any` in the local state to avoid diffs
			localRequest.HttpVersion = types.StringValue(string(datadogV1.SYNTHETICSTESTOPTIONSHTTPVERSION_ANY))
		}
		localStep.RequestDefinition = []syntheticsTestAPIStepRequestModel{localRequest}
		localStep.RequestHeaders, _ = types.MapValueFrom(ctx, types.StringType, stepRequest.GetHeaders())
		localStep.RequestQuery, _ = types.MapValueFrom(ctx, types.StringType, stepRequest.GetQuery())
		localStep.RequestMetadata, _ = types.MapValueFrom(ctx, types.StringType, stepRequest.GetMetadata())

		if basicAuth, ok := stepRequest.GetBasicAuthOk(); ok {
			localStep.RequestBasicAuth = []syntheticsTestRequestBasicAuthModel{buildTerraformBasicAuth(basicAuth)}
		}

		if clientCertificate, ok := stepRequest.GetCertificateOk(); ok {
			localStep.RequestClientCertificate = []syntheticsTestRequestClientCertificateModel{buildTerraformRequestCertificates(*clientCertificate, localStep.RequestClientCertificate)}
		}

		if proxy, ok := stepRequest.GetProxyOk(); ok {
			localStep.RequestProxy = []syntheticsTestRequestProxyModel{buildTerraformTestRequestProxy(*proxy)}
		}

		if files, ok := stepRequest.GetFilesOk(); ok && files != nil && len(*files) > 0 {
			localStep.RequestFile = buildTerraformBodyFiles(files, localStep.RequestFile)
		}

		localStep.AllowFailure = types.BoolValue(apiTestStep.GetAllowFailure())
		localStep.ExitIfSucceed = types.BoolValue(apiTestStep.GetExitIfSucceed())
		localStep.IsCritical = types.BoolValue(apiTestStep.GetIsCritical())

		if retry, ok := apiTestStep.GetRetryOk(); ok {
			localRetry := syntheticsTestRetryModel{}
			if count, ok := retry.GetCountOk(); ok {
				localRetry.Count = types.Int64Value(*count)
			}
			if interval, ok := retry.GetIntervalOk(); ok {
				localRetry.Interval = types.Int64Value(int64(*interval))
			}
			localStep.Retry = []syntheticsTestRetryModel{localRetry}
		}
	} else if step.SyntheticsAPIWaitStep != nil {
		localStep.Name = types.StringValue(step.SyntheticsAPIWaitStep.GetName())
		localStep.Subtype = types.StringValue(string(step.SyntheticsAPIWaitStep.GetSubtype()))
		localStep.Value = types.Int64Value(int64(step.SyntheticsAPIWaitStep.GetValue()))
	}

	return localStep
}

/*
 * Utils
 */

func compressAndEncodeValue(value string) string {
	var compressedValue bytes.Buffer
	zl := zlib.NewWriter(&compressedValue)
	zl.Write([]byte(value))
	zl.Close()
	encodedCompressedValue := b64.StdEncoding.EncodeToString(compressedValue.Bytes())
	return encodedCompressedValue
}

func decompressAndDecodeValue(value string) string {
	decodedValue, _ := b64.StdEncoding.DecodeString(value)
	decodedBytes := bytes.NewReader(decodedValue)
	zl, _ := zlib.NewReader(decodedBytes)
	defer zl.Close()
	compressedProtoFile, _ := io.ReadAll(zl)
	return string(compressedProtoFile)
}

func convertStepParamsValueForConfig() {}

func convertStepParamsValueForState() {}

func convertStepParamsKey() {}

func convertToString(i interface{}) string {
	switch v := i.(type) {
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	default:
		// TODO: manage target for JSON body assertions
		valStrr, err := json.Marshal(v)
		if err == nil {
			return string(valStrr)
		}
		return ""
	}
}

// get the sha256 of a client certificate content
// in some case where Terraform compares the state value
// we already get the hashed value so we don't need to
// hash it again
func getCertificateStateValue(content string) string {
	if isHash := isCertHash(content); isHash {
		return content
	}

	return utils.ConvertToSha256(content)
}

func getStepParams() {}

func getParamsKeysForStepType() {}

func getParamsKeysForMobileStepType() {}

func buildDatadogParamsForMobileStep() {}

func buildDatadogParamsElementForMobileStep() {}

func getSyntheticsTestType(state syntheticsTestModel) datadogV1.SyntheticsTestDetailsType {
	v := datadogV1.SyntheticsTestDetailsType(state.Type.ValueString())
	return v
}

func isCertHash(content string) bool {
	// a sha256 hash consists of 64 hexadecimal characters
	isHash, _ := regexp.MatchString("^[A-Fa-f0-9]{64}$", content)

	return isHash
}

func isTargetOfTypeInt(assertionType datadogV1.SyntheticsAssertionType, assertionOperator datadogV1.SyntheticsAssertionOperator) bool {
	for _, intTargetAssertionType := range []datadogV1.SyntheticsAssertionType{
		datadogV1.SYNTHETICSASSERTIONTYPE_RESPONSE_TIME,
		datadogV1.SYNTHETICSASSERTIONTYPE_CERTIFICATE,
		datadogV1.SYNTHETICSASSERTIONTYPE_LATENCY,
		datadogV1.SYNTHETICSASSERTIONTYPE_PACKETS_RECEIVED,
		datadogV1.SYNTHETICSASSERTIONTYPE_NETWORK_HOP,
		datadogV1.SYNTHETICSASSERTIONTYPE_GRPC_HEALTHCHECK_STATUS,
	} {
		if assertionType == intTargetAssertionType {
			return true
		}
	}
	if assertionType == datadogV1.SYNTHETICSASSERTIONTYPE_STATUS_CODE &&
		(assertionOperator == datadogV1.SYNTHETICSASSERTIONOPERATOR_IS || assertionOperator == datadogV1.SYNTHETICSASSERTIONOPERATOR_IS_NOT) {
		return true
	}
	return false
}

func getConfigCertAndKeyContent() {}

func getCertAndKeyFromMap() {}

func stringMapToTerraformMap(input map[string]string) types.Map {
	attrMap := make(map[string]attr.Value, len(input))
	for k, v := range input {
		attrMap[k] = types.StringValue(v)
	}
	result, _ := types.MapValue(types.StringType, attrMap)
	return result
}

func terraformMapToStringMap(ctx context.Context, input types.Map) map[string]string {
	output := make(map[string]string, len(input.Elements()))
	input.ElementsAs(ctx, &output, false)
	return output
}

func terraformSetToStringArray(ctx context.Context, input types.Set) []string {
	output := make([]string, len(input.Elements()))
	input.ElementsAs(ctx, &output, false)
	return output
}

func terraformListToStringArray(ctx context.Context, input types.List) []string {
	output := make([]string, len(input.Elements()))
	input.ElementsAs(ctx, &output, false)
	return output
}
