package fwprovider

import (
	"context"
	"fmt"
	"regexp"

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
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
	Cert []syntheticsTestClientCert `tfsdk:"cert"`
	Key  []syntheticsTestClientCert `tfsdk:"key"`
}

type syntheticsTestClientCert struct {
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
	syntheticsTestRequestModel
	AllowInsecure   types.Bool   `tfsdk:"allow_insecure"`
	FollowRedirects types.Bool   `tfsdk:"follow_redirects"`
	HttpVersion     types.String `tfsdk:"http_version"`
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
		Description: "Provides a Datadog SyntheticsTest resource. This can be used to create and manage Datadog synthetics_test.",
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
		Description: "Header name and value map.",
		Optional:    true,
		// TODO: add fw validator for http headers
	}
}

func syntheticsTestRequestQuery() schema.MapAttribute {
	return schema.MapAttribute{
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
					Validators:  []validator.String{
						// TODO: fix this
						// validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsAssertionTypeFromValue, datadogV1.NewSyntheticsAssertionBodyHashTypeFromValue, datadogV1.NewSyntheticsAssertionJavascriptTypeFromValue),
					},
				},
				"operator": schema.StringAttribute{
					Description: "Assertion operator. **Note** Only some combinations of `type` and `operator` are valid (please refer to [Datadog documentation](https://docs.datadoghq.com/api/latest/synthetics/#create-a-test)).",
					Optional:    true,
					Validators:  []validator.String{
						// TODO: reimplement validateSyntheticsAssertionOperator
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
					Computed:    true,
					Default:     int64default.StaticInt64(0),
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
								Optional: true,
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
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsTestResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state syntheticsTestModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var testId string
	testType := state.Type.ValueString()

	switch testType {
	case string(datadogV1.SYNTHETICSTESTDETAILSTYPE_API):
		apiTestBody, diags := r.buildSyntheticsAPITestRequestBody(ctx, &state)
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

	case string(datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER):
		browserTestBody, diags := r.buildSyntheticsBrowserTestRequestBody(ctx, &state)
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

	case string(datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE):
		mobileTestBody, diags := r.buildSyntheticsMobileTestRequestBody(ctx, &state)
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
	testType := state.Type.ValueString()

	switch testType {
	case string(datadogV1.SYNTHETICSTESTDETAILSTYPE_API):
		apiTestBody, diags := r.buildSyntheticsAPITestRequestBody(ctx, &state)
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
	case string(datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER):
		browserTestBody, diags := r.buildSyntheticsBrowserTestRequestBody(ctx, &state)
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
	case string(datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE):
		mobileTestBody, diags := r.buildSyntheticsMobileTestRequestBody(ctx, &state)
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
}

func (r *syntheticsTestResource) updateSyntheticsBrowserTestLocalState(ctx context.Context, state *syntheticsTestModel, resp *datadogV1.SyntheticsBrowserTest) {
}

func (r *syntheticsTestResource) updateSyntheticsMobileTestLocalState(ctx context.Context, state *syntheticsTestModel, resp *datadogV1.SyntheticsMobileTest) {
}

/*
 * Transformer functions between datadog and terraform
 */

func (r *syntheticsTestResource) buildSyntheticsAPITestRequestBody(ctx context.Context, state *syntheticsTestModel) (*datadogV1.SyntheticsAPITest, diag.Diagnostics) {
	return nil, nil
}

func (r *syntheticsTestResource) buildSyntheticsBrowserTestRequestBody(ctx context.Context, state *syntheticsTestModel) (*datadogV1.SyntheticsBrowserTest, diag.Diagnostics) {
	return nil, nil
}

func (r *syntheticsTestResource) buildSyntheticsMobileTestRequestBody(ctx context.Context, state *syntheticsTestModel) (*datadogV1.SyntheticsMobileTest, diag.Diagnostics) {
	return nil, nil
}