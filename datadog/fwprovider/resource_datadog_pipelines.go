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
	Processors   []*ProcessorsModel   `tfsdk:"processors"`
	Destinations []*destinationsModel `tfsdk:"destinations"`
}
type SourcesModel struct {
	DatadogAgentSource []*datadogAgentSourceModel `tfsdk:"datadog_agent"`
}
type datadogAgentSourceModel struct {
	Id  types.String `tfsdk:"id"`
	Tls *tlsModel    `tfsdk:"tls"`
}

type tlsModel struct {
	CrtFile types.String `tfsdk:"crt_file"`
	CaFile  types.String `tfsdk:"ca_file"`
	KeyFile types.String `tfsdk:"key_file"`
}

type ProcessorsModel struct {
	FilterProcessor    []*filterProcessorModel    `tfsdk:"filter"`
	ParseJsonProcessor []*ParseJsonProcessorModel `tfsdk:"parse_json"`
}
type filterProcessorModel struct {
	Id      types.String `tfsdk:"id"`
	Include types.String `tfsdk:"include"`
	Inputs  types.List   `tfsdk:"inputs"`
}

type ParseJsonProcessorModel struct {
	Id      types.String `tfsdk:"id"`
	Inputs  types.List   `tfsdk:"inputs"`
	Include types.String `tfsdk:"include"`
	Field   types.String `tfsdk:"field"`
}

type destinationsModel struct {
	DatadogLogsDestination []*datadogLogsDestinationModel `tfsdk:"datadog_logs"`
}
type datadogLogsDestinationModel struct {
	Id     types.String `tfsdk:"id"`
	Inputs types.List   `tfsdk:"inputs"`
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

func (r *pipelinesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog Pipelines resource. This can be used to create and manage Datadog pipelines.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The pipeline name.",
			},
		},
		Blocks: map[string]schema.Block{
			"config": schema.SingleNestedBlock{
				Description: "Configuration for the pipeline.",
				Blocks: map[string]schema.Block{
					"sources": schema.ListNestedBlock{
						Description: "List of sources.",
						NestedObject: schema.NestedBlockObject{
							Blocks: map[string]schema.Block{
								"datadog_agent": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												Optional:    true,
												Description: "The unique ID of the source.",
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
					},
					"processors": schema.ListNestedBlock{
						Description: "List of processors.",
						NestedObject: schema.NestedBlockObject{
							Blocks: map[string]schema.Block{
								"filter": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												Optional:    true,
												Description: "The unique ID of the processor.",
											},
											"include": schema.StringAttribute{
												Optional:    true,
												Description: "Inclusion filter for the processor.",
											},
											"inputs": schema.ListAttribute{
												Description: "The inputs for the processor.",
												ElementType: types.StringType,
												Required:    true,
											},
										},
									},
								},
								"parse_json": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												Optional:    true,
												Description: "The unique ID of the processor.",
											},
											"include": schema.StringAttribute{
												Optional:    true,
												Description: "Inclusion filter for the processor.",
											},
											"inputs": schema.ListAttribute{
												Required:    true,
												Description: "The inputs for the processor.",
												ElementType: types.StringType,
											},
											"field": schema.StringAttribute{
												Required:    true,
												Description: "The field to parse.",
											},
										},
									},
								},
							},
						},
					},
					"destinations": schema.ListNestedBlock{
						Description: "List of destinations.",
						NestedObject: schema.NestedBlockObject{
							Blocks: map[string]schema.Block{
								"datadog_logs": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												Required:    true,
												Description: "The unique ID of the source.",
											},
											"inputs": schema.ListAttribute{
												Description: "The inputs for the processor.",
												ElementType: types.StringType,
												Required:    true,
											},
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
				sourcesTfItem.DatadogAgentSource[0] = &datadogAgentSourceTf
			}

			configTf.Sources = append(configTf.Sources, &sourcesTfItem)
		}
	}
	if processors, ok := config.GetProcessorsOk(); ok && len(*processors) > 0 {
		configTf.Processors = []*ProcessorsModel{}
		for _, processorsDd := range *processors {
			processorsTfItem := ProcessorsModel{}

			if processorsDd.FilterProcessor != nil {
				processorsTfItem.FilterProcessor = []*filterProcessorModel{}
				filterProcessorTf := filterProcessorModel{}

				filterProcessorTf.Id = types.StringValue(processorsDd.FilterProcessor.Id)
				filterProcessorTf.Include = types.StringValue(processorsDd.FilterProcessor.Include)
				filterProcessorTf.Inputs, _ = types.ListValueFrom(ctx, types.StringType, processorsDd.FilterProcessor.Inputs)
				processorsTfItem.FilterProcessor = append(processorsTfItem.FilterProcessor, &filterProcessorTf)
			}

			parseJSONProcessor := processorsDd.ParseJSONProcessor
			if parseJSONProcessor != nil {
				processorsTfItem.ParseJsonProcessor = []*ParseJsonProcessorModel{}
				parseJsonProcessorTf := ParseJsonProcessorModel{}

				parseJsonProcessorTf.Id = types.StringValue(parseJSONProcessor.Id)
				parseJsonProcessorTf.Include = types.StringValue(*parseJSONProcessor.Include)
				parseJsonProcessorTf.Inputs, _ = types.ListValueFrom(ctx, types.StringType, parseJSONProcessor.Inputs)
				parseJsonProcessorTf.Field = types.StringValue(parseJSONProcessor.Field)
				processorsTfItem.ParseJsonProcessor = append(processorsTfItem.ParseJsonProcessor, &parseJsonProcessorTf)
			}

			configTf.Processors = append(configTf.Processors, &processorsTfItem)

		}
	}
	if destinations, ok := config.GetDestinationsOk(); ok && len(*destinations) > 0 {
		configTf.Destinations = []*destinationsModel{}
		for _, destinationsDd := range *destinations {
			destinationsTfItem := destinationsModel{}
			destinationsTfItem.DatadogLogsDestination = []*datadogLogsDestinationModel{}

			if destinationsDd.DatadogLogsDestination != nil {
				datadogLogsDestinationTf := datadogLogsDestinationModel{}

				datadogLogsDestinationTf.Id = types.StringValue(destinationsDd.DatadogLogsDestination.Id)
				datadogLogsDestinationTf.Inputs, _ = types.ListValueFrom(ctx, types.StringType, destinationsDd.DatadogLogsDestination.Inputs)
				destinationsTfItem.DatadogLogsDestination = append(destinationsTfItem.DatadogLogsDestination, &datadogLogsDestinationTf)
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

				for _, ddSource := range sourcesTFItem.DatadogAgentSource {
					if ddSource != nil {
						var datadogAgentSource datadogV2.DatadogAgentSource

						if !ddSource.Id.IsNull() {
							datadogAgentSource.SetId(ddSource.Id.ValueString())
						}
						datadogAgentSource.SetType("datadog_agent")

						if ddSource.Tls != nil {
							var tls datadogV2.Tls

							if !ddSource.Tls.CrtFile.IsNull() {
								tls.SetCrtFile(ddSource.Tls.CrtFile.ValueString())
							}
							if !ddSource.Tls.CaFile.IsNull() {
								tls.SetCaFile(ddSource.Tls.CaFile.ValueString())
							}
							if !ddSource.Tls.KeyFile.IsNull() {
								tls.SetKeyFile(ddSource.Tls.KeyFile.ValueString())
							}
							datadogAgentSource.Tls = &tls
						}

						sourcesDDItem.DatadogAgentSource = &datadogAgentSource
					}
					sources = append(sources, sourcesDDItem)
				}
			}
			config.SetSources(sources)
			attributes.SetConfig(config)
		}

		if state.Config.Processors != nil {
			var processors []datadogV2.PipelineDataAttributesConfigProcessorsItem
			for _, processorsTFItem := range state.Config.Processors {
				for _, filterProcessorTF := range processorsTFItem.FilterProcessor {
					processorsDDItem := datadogV2.PipelineDataAttributesConfigProcessorsItem{}
					if filterProcessorTF != nil {
						var filterProcessor datadogV2.FilterProcessor
						if !filterProcessorTF.Id.IsNull() {
							filterProcessor.SetId(filterProcessorTF.Id.ValueString())
						}
						filterProcessor.SetType("filter")
						if !filterProcessorTF.Include.IsNull() {
							filterProcessor.SetInclude(filterProcessorTF.Include.ValueString())
						}
						if !filterProcessorTF.Inputs.IsNull() {
							var inputs []string
							filterProcessorTF.Inputs.ElementsAs(ctx, &inputs, false)
							filterProcessor.SetInputs(inputs)
						}
						processorsDDItem.FilterProcessor = &filterProcessor
					}
					processors = append(processors, processorsDDItem)
				}

				for _, parseJsonProcessorTF := range processorsTFItem.ParseJsonProcessor {
					processorsDDItem := datadogV2.PipelineDataAttributesConfigProcessorsItem{}
					if parseJsonProcessorTF != nil {
						var parseJsonProcessor datadogV2.ParseJSONProcessor
						if !parseJsonProcessorTF.Id.IsNull() {
							parseJsonProcessor.SetId(parseJsonProcessorTF.Id.ValueString())
						}
						parseJsonProcessor.SetType("parse_json")
						if !parseJsonProcessorTF.Include.IsNull() {
							parseJsonProcessor.SetInclude(parseJsonProcessorTF.Include.ValueString())
						}
						if !parseJsonProcessorTF.Inputs.IsNull() {
							var inputs []string
							parseJsonProcessorTF.Inputs.ElementsAs(ctx, &inputs, false)
							parseJsonProcessor.SetInputs(inputs)
						}
						if !parseJsonProcessorTF.Field.IsNull() {
							parseJsonProcessor.SetField(parseJsonProcessorTF.Field.ValueString())
						}
						processorsDDItem.ParseJSONProcessor = &parseJsonProcessor
					}
					processors = append(processors, processorsDDItem)
				}
			}
			config.SetProcessors(processors)
		}

		if state.Config.Destinations != nil {
			var destinations []datadogV2.PipelineDataAttributesConfigDestinationsItem
			for _, destinationsTFItem := range state.Config.Destinations {
				destinationsDDItem := datadogV2.PipelineDataAttributesConfigDestinationsItem{}
				for _, destination := range destinationsTFItem.DatadogLogsDestination {
					if destination != nil {
						var datadogLogsDestination datadogV2.DatadogLogsDestination
						if !destination.Id.IsNull() {
							datadogLogsDestination.SetId(destination.Id.ValueString())
						}
						datadogLogsDestination.SetType("datadog_logs")
						if !destination.Inputs.IsNull() {
							var inputs []string
							destination.Inputs.ElementsAs(ctx, &inputs, false)
							datadogLogsDestination.SetInputs(inputs)
						}
						destinationsDDItem.DatadogLogsDestination = &datadogLogsDestination
					}
					destinations = append(destinations, destinationsDDItem)
				}
			}
			config.SetDestinations(destinations)
		}
		attributes.SetConfig(config)
	}

	pipelineData := datadogV2.NewPipelineDataWithDefaults()
	pipelineData.SetAttributes(*attributes)
	req.SetData(*pipelineData)

	return req, diags
}
