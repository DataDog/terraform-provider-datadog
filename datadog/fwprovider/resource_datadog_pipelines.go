package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &pipelinesResource{}
	_ resource.ResourceWithImportState = &pipelinesResource{}
)

type pipelinesResource struct {
	Api  *datadogV2.ObservabilityPipelinesApi
	Auth context.Context
}

type pipelinesModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Config *configModel `tfsdk:"config"`
}

type configModel struct {
	Sources      []*SourcesModel      `tfsdk:"sources"`
	Processors   []*processorsModel   `tfsdk:"processors"`
	Destinations []*destinationsModel `tfsdk:"destinations"`
}
type SourcesModel struct {
	DatadogAgentSource *datadogAgentSourceModel `tfsdk:"datadog_agent_source"`
}
type datadogAgentSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
	Tls  *tlsModel    `tfsdk:"tls"`
}
type tlsModel struct {
	CrtFile types.String `tfsdk:"crt_file"`
	CaFile  types.String `tfsdk:"ca_file"`
	KeyFile types.String `tfsdk:"key_file"`
}

type processorsModel struct {
	FilterProcessor *filterProcessorModel `tfsdk:"filter_processor"`
}
type filterProcessorModel struct {
	Id      types.String `tfsdk:"id"`
	Type    types.String `tfsdk:"type"`
	Include types.String `tfsdk:"include"`
}

type destinationsModel struct {
	DatadogLogsDestination *datadogLogsDestinationModel `tfsdk:"datadog_logs_destination"`
}
type datadogLogsDestinationModel struct {
	Id   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

func NewPipelinesResource() resource.Resource {
	return &pipelinesResource{}
}

func (r *pipelinesResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetObsPipelinesV2()
	r.Auth = providerData.Auth
}

func (r *pipelinesResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "pipelines"
}

func (r *pipelinesResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Pipelines resource. This can be used to create and manage Datadog pipelines.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The pipeline name.",
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"sources": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{},
							Blocks: map[string]schema.Block{
								"datadog_agent_source": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Optional:    true,
											Description: "The unique ID of the source.",
										},
										"type": schema.StringAttribute{
											Optional:    true,
											Description: "The type of source.",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"crt_file": schema.StringAttribute{
													Optional:    true,
													Description: "CRT file",
												},
												"ca_file": schema.StringAttribute{
													Optional:    true,
													Description: "CA file",
												},
												"key_file": schema.StringAttribute{
													Optional:    true,
													Description: "Key file",
												},
											},
										},
									},
								},
							},
						},
					},
					"processors": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{},
							Blocks: map[string]schema.Block{
								"filter_processor": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Optional:    true,
											Description: "The unique ID of the processor.",
										},
										"type": schema.StringAttribute{
											Optional:    true,
											Description: "The type of processor.",
										},
										"include": schema.StringAttribute{
											Optional:    true,
											Description: "Inclusion filter for the processor.",
										},
									},
								},
							},
						},
					},
					"destinations": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{},
							Blocks: map[string]schema.Block{
								"datadog_logs_destination": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Optional:    true,
											Description: "The unique ID of the destination.",
										},
										"type": schema.StringAttribute{
											Optional:    true,
											Description: "The type of destination.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *pipelinesResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *pipelinesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state pipelinesModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetPipeline(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Pipelines"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *pipelinesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state pipelinesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildPipelinesRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreatePipeline(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Pipelines"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *pipelinesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state pipelinesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildPipelinesRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdatePipeline(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Pipelines"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *pipelinesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state pipelinesModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeletePipeline(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting pipelines"))
		return
	}
}

func (r *pipelinesResource) updateState(ctx context.Context, state *pipelinesModel, resp *datadogV2.Pipeline) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.Name = types.StringValue(attributes.GetName())

	config := attributes.GetConfig()

	configTf := configModel{}
	if sources, ok := config.GetSourcesOk(); ok && len(*sources) > 0 {
		configTf.Sources = []*SourcesModel{}
		for _, sourcesDd := range *sources {
			sourcesTfItem := SourcesModel{}

			if sourcesDd.DatadogAgentSource != nil {
				datadogAgentSourceTf := datadogAgentSourceModel{}

				datadogAgentSourceTf.Id = types.StringValue(sourcesDd.DatadogAgentSource.Id)
				datadogAgentSourceTf.Type = types.StringValue(string(sourcesDd.DatadogAgentSource.Type))
				if sourcesDd.DatadogAgentSource.Tls != nil {
					tlsTf := tlsModel{}

					tlsTf.CrtFile = types.StringValue(sourcesDd.DatadogAgentSource.Tls.CrtFile)
					if sourcesDd.DatadogAgentSource.Tls.CaFile != nil {
						tlsTf.CaFile = types.StringValue(*sourcesDd.DatadogAgentSource.Tls.CaFile)
					}
					if sourcesDd.DatadogAgentSource.Tls.KeyFile != nil {
						tlsTf.KeyFile = types.StringValue(*sourcesDd.DatadogAgentSource.Tls.KeyFile)
					}
					datadogAgentSourceTf.Tls = &tlsTf
				}
				sourcesTfItem.DatadogAgentSource = &datadogAgentSourceTf
			}

			configTf.Sources = append(configTf.Sources, &sourcesTfItem)
		}
	}
	if processors, ok := config.GetProcessorsOk(); ok && len(*processors) > 0 {
		configTf.Processors = []*processorsModel{}
		for _, processorsDd := range *processors {
			processorsTfItem := processorsModel{}

			if processorsDd.FilterProcessor != nil {
				filterProcessorTf := filterProcessorModel{}

				filterProcessorTf.Id = types.StringValue(processorsDd.FilterProcessor.Id)
				filterProcessorTf.Type = types.StringValue(string(processorsDd.FilterProcessor.Type))
				filterProcessorTf.Include = types.StringValue(processorsDd.FilterProcessor.Include)
				processorsTfItem.FilterProcessor = &filterProcessorTf
			}

			configTf.Processors = append(configTf.Processors, &processorsTfItem)
		}
	}
	if destinations, ok := config.GetDestinationsOk(); ok && len(*destinations) > 0 {
		configTf.Destinations = []*destinationsModel{}
		for _, destinationsDd := range *destinations {
			destinationsTfItem := destinationsModel{}

			if destinationsDd.DatadogLogsDestination != nil {
				datadogLogsDestinationTf := datadogLogsDestinationModel{}

				datadogLogsDestinationTf.Id = types.StringValue(destinationsDd.DatadogLogsDestination.Id)
				datadogLogsDestinationTf.Type = types.StringValue(string(destinationsDd.DatadogLogsDestination.Type))
				destinationsTfItem.DatadogLogsDestination = &datadogLogsDestinationTf
			}

			configTf.Destinations = append(configTf.Destinations, &destinationsTfItem)
		}
	}

	state.Config = &configTf
}

func (r *pipelinesResource) buildPipelinesRequestBody(ctx context.Context, state *pipelinesModel) (*datadogV2.Pipeline, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.Pipeline{}
	attributes := datadogV2.NewPipelineDataAttributesWithDefaults()

	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}

	if state.Config != nil {
		var config datadogV2.PipelineDataAttributesConfig

		if state.Config.Sources != nil {
			var sources []datadogV2.PipelineDataAttributesConfigSourcesItem
			for _, sourcesTFItem := range state.Config.Sources {
				sourcesDDItem := datadogV2.PipelineDataAttributesConfigSourcesItem{}

				if sourcesTFItem.DatadogAgentSource != nil {
					var datadogAgentSource datadogV2.DatadogAgentSource

					if !sourcesTFItem.DatadogAgentSource.Id.IsNull() {
						datadogAgentSource.SetId(sourcesTFItem.DatadogAgentSource.Id.ValueString())
					}
					if !sourcesTFItem.DatadogAgentSource.Type.IsNull() {
						datadogAgentSource.SetType(datadogV2.DatadogAgentSourceType(sourcesTFItem.DatadogAgentSource.Type.ValueString()))
					}

					if sourcesTFItem.DatadogAgentSource.Tls != nil {
						var tls datadogV2.Tls

						if !sourcesTFItem.DatadogAgentSource.Tls.CrtFile.IsNull() {
							tls.SetCrtFile(sourcesTFItem.DatadogAgentSource.Tls.CrtFile.ValueString())
						}
						if !sourcesTFItem.DatadogAgentSource.Tls.CaFile.IsNull() {
							tls.SetCaFile(sourcesTFItem.DatadogAgentSource.Tls.CaFile.ValueString())
						}
						if !sourcesTFItem.DatadogAgentSource.Tls.KeyFile.IsNull() {
							tls.SetKeyFile(sourcesTFItem.DatadogAgentSource.Tls.KeyFile.ValueString())
						}
						datadogAgentSource.Tls = &tls
					}

					sourcesDDItem.DatadogAgentSource = &datadogAgentSource
				}
			}
			config.SetSources(sources)
			attributes.SetConfig(config)
		}

		if state.Config.Processors != nil {
			var processors []datadogV2.PipelineDataAttributesConfigProcessorsItem
			for _, processorsTFItem := range state.Config.Processors {
				processorsDDItem := datadogV2.PipelineDataAttributesConfigProcessorsItem{}
				if processorsTFItem.FilterProcessor != nil {
					var filterProcessor datadogV2.FilterProcessor
					if !processorsTFItem.FilterProcessor.Id.IsNull() {
						filterProcessor.SetId(processorsTFItem.FilterProcessor.Id.ValueString())
					}
					if !processorsTFItem.FilterProcessor.Type.IsNull() {
						filterProcessor.SetType(datadogV2.FilterProcessorType(processorsTFItem.FilterProcessor.Type.ValueString()))
					}
					if !processorsTFItem.FilterProcessor.Include.IsNull() {
						filterProcessor.SetInclude(processorsTFItem.FilterProcessor.Include.ValueString())
					}
					processorsDDItem.FilterProcessor = &filterProcessor
				}
			}
			config.SetProcessors(processors)
		}

		if state.Config.Destinations != nil {
			var destinations []datadogV2.PipelineDataAttributesConfigDestinationsItem
			for _, destinationsTFItem := range state.Config.Destinations {
				destinationsDDItem := datadogV2.PipelineDataAttributesConfigDestinationsItem{}
				if destinationsTFItem.DatadogLogsDestination != nil {
					var datadogLogsDestination datadogV2.DatadogLogsDestination
					if !destinationsTFItem.DatadogLogsDestination.Id.IsNull() {
						datadogLogsDestination.SetId(destinationsTFItem.DatadogLogsDestination.Id.ValueString())
					}
					if !destinationsTFItem.DatadogLogsDestination.Type.IsNull() {
						datadogLogsDestination.SetType(datadogV2.DatadogLogsDestinationType(destinationsTFItem.DatadogLogsDestination.Type.ValueString()))
					}
					destinationsDDItem.DatadogLogsDestination = &datadogLogsDestination
				}
			}
			config.SetDestinations(destinations)
		}
		attributes.SetConfig(config)
	}

	pipelineData := datadogV2.NewPipelineDataWithDefaults()
	req.SetData(*pipelineData)

	return req, diags
}
