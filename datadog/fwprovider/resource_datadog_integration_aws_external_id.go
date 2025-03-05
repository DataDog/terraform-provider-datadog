package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const EXPIRY_WARNING_MESSAGE = "A new external ID must be used to create an AWS account integration in Datadog within 48 hours of creation or it will expire."
const DESTROY_WARNING_MESSAGE = "Running `terraform destroy` only removes the resource from Terraform state and does not deactivate anything in Datadog or AWS."

var (
	_ resource.ResourceWithConfigure   = &integrationAwsExternalIDResource{}
	_ resource.ResourceWithImportState = &integrationAwsExternalIDResource{}
)

type integrationAwsExternalIDResource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

type integrationAwsExternalIDModel struct {
	ID types.String `tfsdk:"id"`
}

func NewIntegrationAwsExternalIDResource() resource.Resource {
	return &integrationAwsExternalIDResource{}
}

func (r *integrationAwsExternalIDResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationAwsExternalIDResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_aws_external_id"
}

func (r *integrationAwsExternalIDResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: fmt.Sprintf("!>%s\n\n!>%s\n\n"+
			"Provides a Datadog-Amazon Web Services external ID resource. This can be used to create Datadog-Amazon Web Services external IDs\n\n"+
			"This resource can be used in conjunction with the `datadog_integration_aws_account` resource to manage AWS integrations. The external ID can be referenced as shown in this example:\n\n"+
			"```hcl\n"+
			"resource \"datadog_integration_aws_external_id\" \"foo\" {}\n\n"+
			"resource \"datadog_integration_aws_account\" \"foo-defaults\" {\n"+
			"  aws_account_id = \"123456789019\"\n"+
			"  aws_partition  = \"aws\"\n\n"+
			"  auth_config {\n"+
			"    aws_auth_config_role {\n"+
			"      role_name   = \"DatadogIntegrationRole\"\n"+
			"      external_id = datadog_integration_aws_external_id.foo.id\n"+
			"    }\n"+
			"  }\n"+
			"}\n"+
			"```\n\n"+
			"To force a new external ID value to regenerate, you can use the `-replace` flag:\n\n"+
			"```shell\n"+
			"terraform apply -replace=\"datadog_integration_aws_external_id.foo\"\n"+
			"```", EXPIRY_WARNING_MESSAGE, DESTROY_WARNING_MESSAGE),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The external ID.",
			},
		},
	}
}

func (r *integrationAwsExternalIDResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationAwsExternalIDResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	return
}

func (r *integrationAwsExternalIDResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationAwsExternalIDModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	utils.IntegrationAwsMutex.Lock()
	defer utils.IntegrationAwsMutex.Unlock()

	resp, _, err := r.Api.CreateNewAWSExternalID(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error creating AWS Integration external ID"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	state.ID = types.StringValue(resp.Data.Attributes.ExternalId)

	response.Diagnostics.AddWarning("External ID must be used within 48 hours", EXPIRY_WARNING_MESSAGE)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAwsExternalIDResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported", "AWS Integration external IDs cannot be updated")
}

func (r *integrationAwsExternalIDResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	response.Diagnostics.AddWarning("Destroy does not deactivate an external ID", DESTROY_WARNING_MESSAGE)
	return
}
