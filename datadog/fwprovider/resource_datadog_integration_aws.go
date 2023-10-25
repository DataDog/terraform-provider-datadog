package fwprovider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var IntegrationAWSMutex = sync.Mutex{}
var accountAndRoleNameIDRegex = regexp.MustCompile("[\\d]+:.*")

var (
	_ resource.ResourceWithConfigure    = &integrationAWSResource{}
	_ resource.ResourceWithImportState  = &integrationAWSResource{}
	_ resource.ResourceWithUpgradeState = &integrationAWSResource{}
)

type integrationAWSResource struct {
	Api  *datadogV1.AWSIntegrationApi
	Auth context.Context
}

type integrationAWSModel struct {
	ID                            types.String `tfsdk:"id"`
	AccountID                     types.String `tfsdk:"account_id"`
	RoleName                      types.String `tfsdk:"role_name"`
	FilterTags                    types.List   `tfsdk:"filter_tags"`
	HostTags                      types.List   `tfsdk:"host_tags"`
	AccountSpecificNamespaceRules types.Map    `tfsdk:"account_specific_namespace_rules"`
	ExcludedRegions               types.Set    `tfsdk:"excluded_regions"`
	ExternalID                    types.String `tfsdk:"external_id"`
	AccessKeyID                   types.String `tfsdk:"access_key_id"`
	SecretAccessKey               types.String `tfsdk:"secret_access_key"`
	MetricsCollectionEnabled      types.Bool   `tfsdk:"metrics_collection_enabled"`
	ResourceCollectionEnabled     types.Bool   `tfsdk:"resource_collection_enabled"`
	CSPMResourceCollectionEnabled types.Bool   `tfsdk:"cspm_resource_collection_enabled"`
}

func NewIntegrationAWSResource() resource.Resource {
	return &integrationAWSResource{}
}

func (r *integrationAWSResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV1()
	r.Auth = providerData.Auth
}

func (r *integrationAWSResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_aws"
}

func (r *integrationAWSResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog - Amazon Web Services integration resource. This can be used to create and manage Datadog - Amazon Web Services integration.\n\n",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Optional:    true,
				Description: "Your AWS Account ID without dashes.",
				Validators:  []validator.String{stringvalidator.ConflictsWith(path.MatchRoot("access_key_id"), path.MatchRoot("secret_access_key"))},
			},
			"role_name": schema.StringAttribute{
				Optional:    true,
				Description: "Your Datadog role delegation name.",
				Validators:  []validator.String{stringvalidator.ConflictsWith(path.MatchRoot("access_key_id"), path.MatchRoot("secret_access_key"))},
			},
			"filter_tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Array of EC2 tags (in the form `key:value`) defines a filter that Datadog uses when collecting metrics from EC2. Wildcards, such as `?` (for single characters) and `*` (for multiple characters) can also be used. Only hosts that match one of the defined tags will be imported into Datadog. The rest will be ignored. Host matching a given tag can also be excluded by adding `!` before the tag. e.x. `env:production,instance-type:c1.*,!region:us-east-1`.",
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"host_tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Array of tags (in the form `key:value`) to add to all hosts and metrics reporting through this integration.",
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"account_specific_namespace_rules": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enables or disables metric collection for specific AWS namespaces for this AWS account only. A list of namespaces can be found at the [available namespace rules API endpoint](https://docs.datadoghq.com/api/v1/aws-integration/#list-namespace-rules).",
				ElementType: types.BoolType,
				Default:     mapdefault.StaticValue(types.MapValueMust(types.BoolType, map[string]attr.Value{})),
			},
			"excluded_regions": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				Description: "An array of AWS regions to exclude from metrics collection.",
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
			"external_id": schema.StringAttribute{
				Computed:    true,
				Description: "AWS External ID. **NOTE** This provider will not be able to detect changes made to the `external_id` field from outside Terraform.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "Your AWS access key ID. Only required if your AWS account is a GovCloud or China account.",
				Validators:  []validator.String{stringvalidator.ConflictsWith(path.MatchRoot("account_id"), path.MatchRoot("role_name"))},
			},
			"secret_access_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Your AWS secret access key. Only required if your AWS account is a GovCloud or China account.",
				Validators:  []validator.String{stringvalidator.ConflictsWith(path.MatchRoot("account_id"), path.MatchRoot("role_name"))},
			},
			"metrics_collection_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether Datadog collects metrics for this AWS account.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_collection_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether Datadog collects a standard set of resources from your AWS account.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cspm_resource_collection_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether Datadog collects cloud security posture management resources from your AWS account. This includes additional resources not covered under the general resource_collection.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
		Version: 1,
	}
}

func (r *integrationAWSResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("external_id"), os.Getenv("EXTERNAL_ID"))...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
}

func (r *integrationAWSResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationAWSModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccount, diags := r.getAWSAccount(ctx, state.ID.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	if awsAccount == nil {
		response.State.RemoveResource(ctx)
		return
	}

	r.updateState(ctx, &state, awsAccount)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationAWSModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	iaws := r.buildDatadogIntegrationAWSStruct(ctx, &state)

	resp, _, err := r.Api.CreateAWSAccount(r.Auth, *iaws)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating AWS integration"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	state.ExternalID = types.StringValue(resp.GetExternalId())

	var ID string
	if !state.AccessKeyID.IsNull() {
		ID = state.AccessKeyID.ValueString()
	} else {
		ID = fmt.Sprintf("%s:%s", state.AccountID.ValueString(), state.RoleName.ValueString())
	}
	state.ID = types.StringValue(ID)

	awsAccount, diags := r.getAWSAccount(ctx, ID)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if awsAccount == nil {
		response.Diagnostics.AddError("error retrieving AWS account", "")
		return
	}

	r.updateState(ctx, &state, awsAccount)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationAWSModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	iaws := r.buildDatadogIntegrationAWSStruct(ctx, &state)

	accountID, roleName, accessKeyID, _ := r.getAWSAccountIdRoleNameAccessKeyIDFromID(state.ID.ValueString())
	if accessKeyID != "" {
		_, _, err := r.Api.UpdateAWSAccount(r.Auth, *iaws,
			*datadogV1.NewUpdateAWSAccountOptionalParameters().
				WithAccessKeyId(accessKeyID),
		)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating AWS integration"))
			return
		}
	} else {
		_, _, err := r.Api.UpdateAWSAccount(r.Auth, *iaws,
			*datadogV1.NewUpdateAWSAccountOptionalParameters().
				WithAccountId(accountID).
				WithRoleName(roleName),
		)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating AWS integration"))
			return
		}
	}

	awsAccount, diags := r.getAWSAccount(ctx, state.ID.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if awsAccount == nil {
		response.Diagnostics.AddError("error retrieving AWS account", "")
		return
	}

	r.updateState(ctx, &state, awsAccount)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationAWSModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	iaws := datadogV1.NewAWSAccountDeleteRequest()
	if !state.AccountID.IsNull() && state.AccountID.ValueString() != "" {
		iaws.SetAccountId(state.AccountID.ValueString())
	}
	if !state.RoleName.IsNull() && state.RoleName.ValueString() != "" {
		iaws.SetRoleName(state.RoleName.ValueString())
	}
	if !state.AccessKeyID.IsNull() && state.AccessKeyID.ValueString() != "" {
		iaws.SetAccessKeyId(state.AccessKeyID.ValueString())
	}

	_, _, err := r.Api.DeleteAWSAccount(r.Auth, *iaws)
	if err != nil {
		response.Diagnostics.AddError("error deleting AWS integration", err.Error())
	}
}

func (r *integrationAWSResource) getAWSAccountIdRoleNameAccessKeyIDFromID(ID string) (string, string, string, diag.Diagnostics) {
	var accountID, roleName, accessKeyID string
	var err error
	var diags diag.Diagnostics

	if accountAndRoleNameIDRegex.MatchString(ID) {
		accountID, roleName, err = utils.AccountAndRoleFromID(ID)
		if err != nil {
			diags.Append(utils.FrameworkErrorDiag(err, ""))
			return "", "", "", diags
		}
	} else {
		accessKeyID = ID
	}

	return accountID, roleName, accessKeyID, diags
}

func (r *integrationAWSResource) getAWSAccount(ctx context.Context, ID string) (*datadogV1.AWSAccount, diag.Diagnostics) {
	accountID, roleName, accessKeyID, diags := r.getAWSAccountIdRoleNameAccessKeyIDFromID(ID)
	if diags.HasError() {
		return nil, diags
	}

	integrations, httpResp, err := r.Api.ListAWSAccounts(r.Auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 400 {
			return nil, diags
		}
		diags.Append(utils.FrameworkErrorDiag(err, "error getting AWS integration"))
		return nil, diags
	}

	var account *datadogV1.AWSAccount
	for _, integration := range integrations.GetAccounts() {
		if (accountID != "" && integration.GetAccountId() == accountID && integration.GetRoleName() == roleName) ||
			(accessKeyID != "" && integration.GetAccessKeyId() == accessKeyID) {
			account = &integration

			if err := utils.CheckForUnparsed(account); err != nil {
				diags.AddError("response contains unparsedObject", err.Error())
			}
			break
		}
	}

	return account, diags
}

func (r *integrationAWSResource) updateState(ctx context.Context, state *integrationAWSModel, resp *datadogV1.AWSAccount) {
	if !accountAndRoleNameIDRegex.MatchString(state.ID.ValueString()) {
		state.ID = types.StringValue(resp.GetAccessKeyId())
	} else {
		state.ID = types.StringValue(fmt.Sprintf("%s:%s", resp.GetAccountId(), resp.GetRoleName()))
	}

	state.MetricsCollectionEnabled = types.BoolValue(*resp.MetricsCollectionEnabled)
	state.ResourceCollectionEnabled = types.BoolValue(*resp.ResourceCollectionEnabled)
	state.CSPMResourceCollectionEnabled = types.BoolValue(*resp.CspmResourceCollectionEnabled)

	if v, ok := resp.GetAccountIdOk(); ok && v != nil {
		state.AccountID = types.StringValue(*v)
	}

	if v, ok := resp.GetRoleNameOk(); ok && v != nil {
		state.RoleName = types.StringValue(*v)
	}

	if v, ok := resp.GetAccessKeyIdOk(); ok && v != nil {
		state.AccessKeyID = types.StringValue(*v)
	}

	if v, ok := resp.GetFilterTagsOk(); ok {
		state.FilterTags, _ = types.ListValueFrom(ctx, types.StringType, *v)
	}

	if v, ok := resp.GetHostTagsOk(); ok {
		state.HostTags, _ = types.ListValueFrom(ctx, types.StringType, *v)
	}

	if v, ok := resp.GetAccountSpecificNamespaceRulesOk(); ok {
		state.AccountSpecificNamespaceRules, _ = types.MapValueFrom(ctx, types.BoolType, *v)
	}

	if v, ok := resp.GetExcludedRegionsOk(); ok {
		state.ExcludedRegions, _ = types.SetValueFrom(ctx, types.StringType, *v)
	}
}

func (r *integrationAWSResource) buildDatadogIntegrationAWSStruct(ctx context.Context, state *integrationAWSModel) *datadogV1.AWSAccount {
	iaws := datadogV1.NewAWSAccount()

	if !state.AccountID.IsNull() {
		iaws.SetAccountId(state.AccountID.ValueString())
	}

	if !state.RoleName.IsNull() {
		iaws.SetRoleName(state.RoleName.ValueString())
	}

	if !state.AccessKeyID.IsNull() {
		iaws.SetAccessKeyId(state.AccessKeyID.ValueString())
	}

	if !state.SecretAccessKey.IsNull() {
		iaws.SetSecretAccessKey(state.SecretAccessKey.ValueString())
	}

	if !state.FilterTags.IsNull() {
		filterTags := make([]string, 0)
		state.FilterTags.ElementsAs(ctx, &filterTags, false)
		iaws.SetFilterTags(filterTags)
	}

	if !state.HostTags.IsNull() {
		hostTags := make([]string, 0)
		state.HostTags.ElementsAs(ctx, &hostTags, false)
		iaws.SetHostTags(hostTags)
	}

	if !state.AccountSpecificNamespaceRules.IsNull() {
		accountSpecificNamespaceRules := make(map[string]bool)
		state.AccountSpecificNamespaceRules.ElementsAs(ctx, &accountSpecificNamespaceRules, false)
		iaws.SetAccountSpecificNamespaceRules(accountSpecificNamespaceRules)
	}

	if !state.ExcludedRegions.IsNull() {
		excludedRegions := make([]string, 0)
		state.ExcludedRegions.ElementsAs(ctx, &excludedRegions, false)
		iaws.SetExcludedRegions(excludedRegions)
	}

	if !state.MetricsCollectionEnabled.IsUnknown() {
		iaws.SetMetricsCollectionEnabled(state.MetricsCollectionEnabled.ValueBool())
	}

	if !state.ResourceCollectionEnabled.IsUnknown() {
		iaws.SetResourceCollectionEnabled(state.ResourceCollectionEnabled.ValueBool())
	}

	if !state.CSPMResourceCollectionEnabled.IsUnknown() {
		iaws.SetCspmResourceCollectionEnabled(state.CSPMResourceCollectionEnabled.ValueBool())
	}

	return iaws
}

func (r *integrationAWSResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":                               schema.StringAttribute{},
					"account_id":                       schema.StringAttribute{},
					"role_name":                        schema.StringAttribute{},
					"filter_tags":                      schema.ListAttribute{ElementType: types.StringType},
					"host_tags":                        schema.ListAttribute{ElementType: types.StringType},
					"account_specific_namespace_rules": schema.MapAttribute{ElementType: types.BoolType},
					"excluded_regions":                 schema.SetAttribute{ElementType: types.StringType},
					"external_id":                      schema.StringAttribute{},
					"access_key_id":                    schema.StringAttribute{},
					"secret_access_key":                schema.StringAttribute{},
					"metrics_collection_enabled":       schema.BoolAttribute{},
					"resource_collection_enabled":      schema.BoolAttribute{},
					"cspm_resource_collection_enabled": schema.BoolAttribute{},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var state integrationAWSModel
				resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
				if resp.Diagnostics.HasError() {
					return
				}

				if !state.AccountID.IsNull() && state.AccountID.ValueString() == "" {
					state.AccountID = types.StringNull()
				}

				if !state.RoleName.IsNull() && state.RoleName.ValueString() == "" {
					state.RoleName = types.StringNull()
				}

				if !state.AccessKeyID.IsNull() && state.AccessKeyID.ValueString() == "" {
					state.AccessKeyID = types.StringNull()
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			},
		},
	}
}
