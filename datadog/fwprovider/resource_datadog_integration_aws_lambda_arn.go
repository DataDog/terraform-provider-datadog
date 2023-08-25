package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &integrationAWSLambdaARNResource{}
	_ resource.ResourceWithImportState = &integrationAWSLambdaARNResource{}
)

type integrationAWSLambdaARNResource struct {
	Api  *datadogV1.AWSLogsIntegrationApi
	Auth context.Context
}

type integrationAWSLambdaARNModel struct {
	ID        types.String `tfsdk:"id"`
	AccountID types.String `tfsdk:"account_id"`
	LambdaARN types.String `tfsdk:"lambda_arn"`
}

func NewIntegrationAWSLambdaARNResource() resource.Resource {
	return &integrationAWSLambdaARNResource{}
}

func (r *integrationAWSLambdaARNResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSLogsIntegrationApiV1()
	r.Auth = providerData.Auth
}

func (r *integrationAWSLambdaARNResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_aws_lambda_arn"
}

func (r *integrationAWSLambdaARNResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog - Amazon Web Services integration Lambda ARN resource. This can be used to create and manage the log collection Lambdas for an account.\n\nUpdate operations are currently not supported with datadog API so any change forces a new resource.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "Your AWS Account ID without dashes. If your account is a GovCloud or China account, specify the `access_key_id` here.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"lambda_arn": schema.StringAttribute{
				Required:    true,
				Description: "The ARN of the Datadog forwarder Lambda.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationAWSLambdaARNResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationAWSLambdaARNResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationAWSLambdaARNModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	logCollection, logCollectionLambdaArn, diags := r.getAWSLambdaArnAccount(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if logCollection == nil || logCollectionLambdaArn == nil {
		response.State.RemoveResource(ctx)
		return
	}

	r.updateState(ctx, &state, logCollection, logCollectionLambdaArn)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSLambdaARNResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationAWSLambdaARNModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	attachLambdaArnRequest := r.buildDatadogIntegrationAWSLambdaARNStruct(ctx, &state)
	resp, httpresp, err := r.Api.CreateAWSLambdaARN(r.Auth, *attachLambdaArnRequest)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error attaching Lambda ARN to AWS integration account"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	res := resp.(map[string]interface{})
	if status, ok := res["status"]; ok && status == "error" {
		response.Diagnostics.AddError("error attaching Lambda ARN to AWS integration account", fmt.Sprintf("%s", httpresp.Body))
		return
	}

	logCollection, logCollectionLambdaArn, diags := r.getAWSLambdaArnAccount(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if logCollection == nil || logCollectionLambdaArn == nil {
		response.Diagnostics.AddError("error retrieving Lambda ARN", "")
		return
	}

	r.updateState(ctx, &state, logCollection, logCollectionLambdaArn)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSLambdaARNResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("resource does not support update", "aws_lambda_arn resource should never call update.")
}

func (r *integrationAWSLambdaARNResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationAWSLambdaARNModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	accountID, lambdaArn, err := utils.AccountAndLambdaArnFromID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("error extracting account_id and lambda_arn from id", err.Error())
		return
	}

	attachLambdaArnRequest := datadogV1.NewAWSAccountAndLambdaRequest(accountID, lambdaArn)
	_, _, err = r.Api.DeleteAWSLambdaARN(r.Auth, *attachLambdaArnRequest)
	if err != nil {
		response.Diagnostics.AddError("error deleting an AWS integration Lambda ARN", err.Error())
	}

}

func (r *integrationAWSLambdaARNResource) getAWSLambdaArnAccount(ctx context.Context, state *integrationAWSLambdaARNModel) (*datadogV1.AWSLogsListResponse, *datadogV1.AWSLogsLambda, diag.Diagnostics) {
	var diags diag.Diagnostics

	logCollections, _, err := r.Api.ListAWSLogsIntegrations(r.Auth)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error getting aws log integrations for datadog account."))
		return nil, nil, diags
	}

	var collection *datadogV1.AWSLogsListResponse
	var collectionLambdaArn *datadogV1.AWSLogsLambda
	for _, logCollection := range logCollections {
		if logCollection.GetAccountId() == state.AccountID.ValueString() {
			for _, logCollectionLambdaArn := range logCollection.GetLambdas() {
				if state.LambdaARN.ValueString() == logCollectionLambdaArn.GetArn() {
					collection = &logCollection
					collectionLambdaArn = &logCollectionLambdaArn
					if err := utils.CheckForUnparsed(collection); err != nil {
						diags.AddError("response contains unparsedObject", err.Error())
					}

					break
				}
			}
		}
	}

	return collection, collectionLambdaArn, diags
}

func (r *integrationAWSLambdaARNResource) updateState(ctx context.Context, state *integrationAWSLambdaARNModel, collection *datadogV1.AWSLogsListResponse, collectionLambdaARN *datadogV1.AWSLogsLambda) {
	state.ID = types.StringValue(fmt.Sprintf("%s %s", collection.GetAccountId(), collectionLambdaARN.GetArn()))
	state.AccountID = types.StringValue(collection.GetAccountId())
	state.LambdaARN = types.StringValue(collectionLambdaARN.GetArn())
}

func (r *integrationAWSLambdaARNResource) buildDatadogIntegrationAWSLambdaARNStruct(ctx context.Context, state *integrationAWSLambdaARNModel) *datadogV1.AWSAccountAndLambdaRequest {
	attachLambdaArnRequest := datadogV1.NewAWSAccountAndLambdaRequest(state.AccountID.ValueString(), state.LambdaARN.ValueString())
	return attachLambdaArnRequest
}
