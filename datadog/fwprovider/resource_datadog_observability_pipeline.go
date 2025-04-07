package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &observabilityPipelineResource{}
	_ resource.ResourceWithImportState = &observabilityPipelineResource{}
)

type observabilityPipelineResource struct {
	Api  *datadogV2.ObservabilityPipelinesApi
	Auth context.Context
}

type observabilityPipelineModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Config configModel  `tfsdk:"config"`
}

type configModel struct {
	Sources      sourcesModel      `tfsdk:"sources"`
	Processors   processorsModel   `tfsdk:"processors"`
	Destinations destinationsModel `tfsdk:"destinations"`
}
type sourcesModel struct {
	DatadogAgentSource []*datadogAgentSourceModel `tfsdk:"datadog_agent"`
	KafkaSource        []*kafkaSourceModel        `tfsdk:"kafka"`
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

type processorsModel struct {
	FilterProcessor    []*filterProcessorModel    `tfsdk:"filter"`
	ParseJsonProcessor []*parseJsonProcessorModel `tfsdk:"parse_json"`
}
type filterProcessorModel struct {
	Id      types.String `tfsdk:"id"`
	Include types.String `tfsdk:"include"`
	Inputs  types.List   `tfsdk:"inputs"`
}

type parseJsonProcessorModel struct {
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

type kafkaSourceModel struct {
	Id                types.String            `tfsdk:"id"`
	GroupId           types.String            `tfsdk:"group_id"`
	Topics            []types.String          `tfsdk:"topics"`
	LibrdkafkaOptions []librdkafkaOptionModel `tfsdk:"librdkafka_option"`
	Sasl              *kafkaSourceSaslModel   `tfsdk:"sasl"`
	Tls               *tlsModel               `tfsdk:"tls"`
}

type librdkafkaOptionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type kafkaSourceSaslModel struct {
	Mechanism types.String `tfsdk:"mechanism"`
}

func NewObservabilitPipelineResource() resource.Resource {
	return &observabilityPipelineResource{}
}

func (r *observabilityPipelineResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetObsPipelinesV2()
	r.Auth = providerData.Auth
}

func (r *observabilityPipelineResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "observability_pipeline"
}

func (r *observabilityPipelineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					"sources": schema.SingleNestedBlock{
						Description: "List of sources.",
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
										"tls": tlsSchema(),
									},
								},
							},
							"kafka": schema.ListNestedBlock{
								Description: "Kafka sources.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Optional:    true,
											Description: "The unique ID of the source.",
										},
										"group_id": schema.StringAttribute{
											Required:    true,
											Description: "The Kafka consumer group ID.",
										},
										"topics": schema.ListAttribute{
											Required:    true,
											Description: "List of Kafka topics to consume.",
											ElementType: types.StringType,
										},
									},
									Blocks: map[string]schema.Block{
										"librdkafka_option": schema.ListNestedBlock{
											Description: "Advanced librdkafka client configuration options.",
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Required:    true,
														Description: "The name of the librdkafka option.",
													},
													"value": schema.StringAttribute{
														Required:    true,
														Description: "The value of the librdkafka option.",
													},
												},
											},
										},
										"sasl": schema.SingleNestedBlock{
											Description: "SASL authentication settings.",
											Attributes: map[string]schema.Attribute{
												"mechanism": schema.StringAttribute{
													Required:    true,
													Description: "SASL mechanism to use (e.g., PLAIN, SCRAM-SHA-256, SCRAM-SHA-512).",
													Validators: []validator.String{
														stringvalidator.OneOf("PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"),
													},
												},
											},
										},
										"tls": tlsSchema(),
									},
								},
							},
						},
					},
					"processors": schema.SingleNestedBlock{
						Description: "List of processors.",
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
					"destinations": schema.SingleNestedBlock{
						Description: "List of destinations.",
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
	}
}

func tlsSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "TLS client configuration.",
		Attributes: map[string]schema.Attribute{
			"crt_file": schema.StringAttribute{
				Required:    true,
				Description: "Path to the TLS certificate file.",
			},
			"ca_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to the Certificate Authority file.",
			},
			"key_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to the private key file.",
			},
		},
	}
}

func (r *observabilityPipelineResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *observabilityPipelineResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state observabilityPipelineModel
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

func (r *observabilityPipelineResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state observabilityPipelineModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildPipelinesRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	createRequest := datadogV2.NewObservabilityPipelineCreateRequestWithDefaults()
	createRequest.Data = *datadogV2.NewObservabilityPipelineCreateRequestDataWithDefaults()
	createRequest.Data.Attributes = body.Data.Attributes
	resp, _, err := r.Api.CreatePipeline(r.Auth, *createRequest)
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

func (r *observabilityPipelineResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state observabilityPipelineModel
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

func (r *observabilityPipelineResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state observabilityPipelineModel
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

func (r *observabilityPipelineResource) updateState(ctx context.Context, state *observabilityPipelineModel, resp *datadogV2.ObservabilityPipeline) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.Name = types.StringValue(attributes.GetName())

	config := attributes.GetConfig()
	stateConfig := configModel{}

	if sources, ok := config.GetSourcesOk(); ok {
		for _, src := range *sources {

			if src.ObservabilityPipelineDatadogAgentSource != nil {
				datadogAgentSourceTf := datadogAgentSourceModel{}

				datadogAgentSourceTf.Id = types.StringValue(src.ObservabilityPipelineDatadogAgentSource.Id)
				if src.ObservabilityPipelineDatadogAgentSource != nil {
					tlsTf := tlsModel{}

					tlsTf.CrtFile = types.StringValue(src.ObservabilityPipelineDatadogAgentSource.Tls.CrtFile)
					if src.ObservabilityPipelineDatadogAgentSource.Tls.CaFile != nil {
						caFile := types.StringValue(*src.ObservabilityPipelineDatadogAgentSource.Tls.CaFile)
						tlsTf.CaFile = caFile
					}
					if src.ObservabilityPipelineDatadogAgentSource.Tls.KeyFile != nil {
						keyFile := types.StringValue(*src.ObservabilityPipelineDatadogAgentSource.Tls.KeyFile)
						tlsTf.KeyFile = keyFile
					}
					datadogAgentSourceTf.Tls = &tlsTf
				}
				stateConfig.Sources.DatadogAgentSource = append(stateConfig.Sources.DatadogAgentSource, &datadogAgentSourceTf)
			}

			if src.ObservabilityPipelineKafkaSource != nil {
				srcKafka := src.ObservabilityPipelineKafkaSource
				kafka := &kafkaSourceModel{
					Id:      types.StringValue(srcKafka.GetId()),
					GroupId: types.StringValue(srcKafka.GetGroupId()),
				}

				topics := srcKafka.GetTopics()
				for _, t := range topics {
					kafka.Topics = append(kafka.Topics, types.StringValue(t))
				}

				if tls, ok := srcKafka.GetTlsOk(); ok {
					tlsModel := &tlsModel{
						CrtFile: types.StringValue(tls.GetCrtFile()),
					}
					if tls.CaFile != nil {
						val := types.StringValue(*tls.CaFile)
						tlsModel.CaFile = val
					}
					if tls.KeyFile != nil {
						val := types.StringValue(*tls.KeyFile)
						tlsModel.KeyFile = val
					}
					kafka.Tls = tlsModel
				}

				if sasl, ok := srcKafka.GetSaslOk(); ok {
					kafka.Sasl = &kafkaSourceSaslModel{
						Mechanism: types.StringValue(string(sasl.GetMechanism())),
					}
				}

				for _, opt := range srcKafka.GetLibrdkafkaOptions() {
					kafka.LibrdkafkaOptions = append(kafka.LibrdkafkaOptions, librdkafkaOptionModel{
						Name:  types.StringValue(opt.Name),
						Value: types.StringValue(opt.Value),
					})
				}

				stateConfig.Sources.KafkaSource = append(stateConfig.Sources.KafkaSource, kafka)
			}
		}
	}
	if processors, ok := config.GetProcessorsOk(); ok {
		for _, processorsDd := range *processors {

			if processorsDd.ObservabilityPipelineFilterProcessor != nil {
				filterProcessorTf := filterProcessorModel{}

				filterProcessorTf.Id = types.StringValue(processorsDd.ObservabilityPipelineFilterProcessor.Id)
				filterProcessorTf.Include = types.StringValue(processorsDd.ObservabilityPipelineFilterProcessor.Include)
				filterProcessorTf.Inputs, _ = types.ListValueFrom(ctx, types.StringType, processorsDd.ObservabilityPipelineFilterProcessor.Inputs)

				stateConfig.Processors.FilterProcessor = append(stateConfig.Processors.FilterProcessor, &filterProcessorTf)
			}

			parseJSONProcessor := processorsDd.ObservabilityPipelineParseJSONProcessor
			if parseJSONProcessor != nil {
				parseJsonProcessorTf := parseJsonProcessorModel{}

				parseJsonProcessorTf.Id = types.StringValue(parseJSONProcessor.Id)
				parseJsonProcessorTf.Include = types.StringValue(parseJSONProcessor.Include)
				parseJsonProcessorTf.Inputs, _ = types.ListValueFrom(ctx, types.StringType, parseJSONProcessor.Inputs)
				parseJsonProcessorTf.Field = types.StringValue(parseJSONProcessor.Field)

				stateConfig.Processors.ParseJsonProcessor = append(stateConfig.Processors.ParseJsonProcessor, &parseJsonProcessorTf)
			}
		}
	}
	if destinations, ok := config.GetDestinationsOk(); ok {
		for _, destinationsDd := range *destinations {

			if destinationsDd.ObservabilityPipelineDatadogLogsDestination != nil {
				datadogLogsDestinationTf := datadogLogsDestinationModel{}

				datadogLogsDestinationTf.Id = types.StringValue(destinationsDd.ObservabilityPipelineDatadogLogsDestination.Id)
				datadogLogsDestinationTf.Inputs, _ = types.ListValueFrom(ctx, types.StringType, destinationsDd.ObservabilityPipelineDatadogLogsDestination.Inputs)
				stateConfig.Destinations.DatadogLogsDestination = append(stateConfig.Destinations.DatadogLogsDestination, &datadogLogsDestinationTf)
			}
		}
	}

	state.Config = stateConfig
}

func (r *observabilityPipelineResource) buildPipelinesRequestBody(ctx context.Context, state *observabilityPipelineModel) (*datadogV2.ObservabilityPipeline, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.ObservabilityPipeline{}
	attributes := datadogV2.NewObservabilityPipelineDataAttributesWithDefaults()

	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}

	var config datadogV2.ObservabilityPipelineConfig

	var sources []datadogV2.ObservabilityPipelineConfigSourceItem
	sourcesTFItem := state.Config.Sources
	sourcesDDItem := datadogV2.ObservabilityPipelineConfigSourceItem{}

	for _, ddSource := range sourcesTFItem.DatadogAgentSource {
		datadogAgentSource := datadogV2.NewObservabilityPipelineDatadogAgentSourceWithDefaults()
		datadogAgentSource.SetId(ddSource.Id.ValueString())

		if ddSource.Tls != nil {
			var tls datadogV2.ObservabilityPipelineTls

			tls.SetCrtFile(ddSource.Tls.CrtFile.ValueString())
			if !ddSource.Tls.CaFile.IsNull() {
				tls.SetCaFile(ddSource.Tls.CaFile.ValueString())
			}
			if !ddSource.Tls.KeyFile.IsNull() {
				tls.SetKeyFile(ddSource.Tls.KeyFile.ValueString())
			}
			datadogAgentSource.Tls = &tls
		}

		sourcesDDItem.ObservabilityPipelineDatadogAgentSource = datadogAgentSource
		sources = append(sources, sourcesDDItem)
	}

	for _, kafka := range sourcesTFItem.KafkaSource {
		kafkaSource := datadogV2.NewObservabilityPipelineKafkaSourceWithDefaults()
		kafkaSource.SetId(kafka.Id.ValueString())
		kafkaSource.SetGroupId(kafka.GroupId.ValueString())

		topics := []string{}
		for _, t := range kafka.Topics {
			topics = append(topics, t.ValueString())
		}
		kafkaSource.SetTopics(topics)

		if kafka.Tls != nil {
			tls := datadogV2.ObservabilityPipelineTls{}
			tls.SetCrtFile(kafka.Tls.CrtFile.ValueString())
			if !kafka.Tls.CaFile.IsNull() {
				tls.SetCaFile(kafka.Tls.CaFile.ValueString())
			}
			if !kafka.Tls.KeyFile.IsNull() {
				tls.SetKeyFile(kafka.Tls.KeyFile.ValueString())
			}
			kafkaSource.SetTls(tls)
		}

		if kafka.Sasl != nil {
			mechanism, _ := datadogV2.NewObservabilityPipelinePipelineKafkaSourceSaslMechanismFromValue(kafka.Sasl.Mechanism.ValueString())
			if mechanism == nil {
				diags.AddError("InvalidSaslMechanism", "Invalid Kafka SASL mechanism provided")
				return nil, diags
			}
			sasl := datadogV2.ObservabilityPipelineKafkaSourceSasl{}
			sasl.SetMechanism(*mechanism)
			kafkaSource.SetSasl(sasl)
		}

		if len(kafka.LibrdkafkaOptions) > 0 {
			opts := []datadogV2.ObservabilityPipelineKafkaSourceLibrdkafkaOption{}
			for _, opt := range kafka.LibrdkafkaOptions {
				opts = append(opts, datadogV2.ObservabilityPipelineKafkaSourceLibrdkafkaOption{
					Name:  opt.Name.ValueString(),
					Value: opt.Value.ValueString(),
				})
			}
			kafkaSource.SetLibrdkafkaOptions(opts)
		}

		sources = append(sources, datadogV2.ObservabilityPipelineConfigSourceItem{
			ObservabilityPipelineKafkaSource: kafkaSource,
		})

	}
	config.SetSources(sources)

	var processors []datadogV2.ObservabilityPipelineConfigProcessorItem
	processorsTFItem := state.Config.Processors
	for _, filterProcessorTF := range processorsTFItem.FilterProcessor {
		processorsDDItem := datadogV2.ObservabilityPipelineConfigProcessorItem{}
		if filterProcessorTF != nil {
			filterProcessor := datadogV2.NewObservabilityPipelineFilterProcessorWithDefaults()
			filterProcessor.SetId(filterProcessorTF.Id.ValueString())
			filterProcessor.SetInclude(filterProcessorTF.Include.ValueString())
			var inputs []string
			filterProcessorTF.Inputs.ElementsAs(ctx, &inputs, false)
			filterProcessor.SetInputs(inputs)
			processorsDDItem.ObservabilityPipelineFilterProcessor = filterProcessor
		}
		processors = append(processors, processorsDDItem)
	}

	for _, parseJsonProcessorTF := range processorsTFItem.ParseJsonProcessor {

		processorsDDItem := datadogV2.ObservabilityPipelineConfigProcessorItem{}
		parseJsonProcessor := datadogV2.NewObservabilityPipelineParseJSONProcessorWithDefaults()

		parseJsonProcessor.SetId(parseJsonProcessorTF.Id.ValueString())
		parseJsonProcessor.SetInclude(parseJsonProcessorTF.Include.ValueString())
		var inputs []string
		parseJsonProcessorTF.Inputs.ElementsAs(ctx, &inputs, false)
		parseJsonProcessor.SetInputs(inputs)
		parseJsonProcessor.SetField(parseJsonProcessorTF.Field.ValueString())
		processorsDDItem.ObservabilityPipelineParseJSONProcessor = parseJsonProcessor
		processors = append(processors, processorsDDItem)

	}
	config.SetProcessors(processors)

	var destinations []datadogV2.ObservabilityPipelineConfigDestinationItem
	destinationsTFItem := state.Config.Destinations
	destinationsDDItem := datadogV2.ObservabilityPipelineConfigDestinationItem{}

	for _, destination := range destinationsTFItem.DatadogLogsDestination {
		datadogLogsDestination := datadogV2.NewObservabilityPipelineDatadogLogsDestinationWithDefaults()
		datadogLogsDestination.SetId(destination.Id.ValueString())
		var inputs []string
		destination.Inputs.ElementsAs(ctx, &inputs, false)
		datadogLogsDestination.SetInputs(inputs)
		destinationsDDItem.ObservabilityPipelineDatadogLogsDestination = datadogLogsDestination

		destinations = append(destinations, destinationsDDItem)
	}

	config.SetDestinations(destinations)

	attributes.SetConfig(config)

	pipelineData := datadogV2.NewObservabilityPipelineDataWithDefaults()
	pipelineData.SetAttributes(*attributes)
	req.SetData(*pipelineData)

	return req, diags
}
