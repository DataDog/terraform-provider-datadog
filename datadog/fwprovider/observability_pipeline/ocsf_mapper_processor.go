package observability_pipeline

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OcsfMapperProcessorSchema returns the schema for the OcsfMapper processor, including
// library mappings and custom OCSF mappings.
func OcsfMapperProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `ocsf_mapper` processor transforms logs into the OCSF schema using predefined library mappings or custom mapping configuration.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{},
			Blocks: map[string]schema.Block{
				"mapping": schema.ListNestedBlock{
					Description: "List of OCSF mapping entries. Each entry uses either a library mapping or a custom mapping.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"include": schema.StringAttribute{
								Required:    true,
								Description: "Search query for selecting which logs the mapping applies to.",
							},
							"library_mapping": schema.StringAttribute{
								Optional:    true,
								Description: "Predefined library mapping for log transformation. Use this or custom_mapping, not both.",
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("custom_mapping")),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"custom_mapping": schema.ListNestedBlock{
								Description: "Custom OCSF mapping configuration for transforming logs.",
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
									listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("library_mapping")),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"version": schema.Int64Attribute{
											Required:    true,
											Description: "The version of the custom mapping configuration.",
										},
									},
									Blocks: map[string]schema.Block{
										"metadata": schema.ListNestedBlock{
											Description: "Metadata for the custom OCSF mapping.",
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
												listvalidator.IsRequired(),
											},
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"class": schema.StringAttribute{
														Required:    true,
														Description: "The OCSF event class name.",
													},
													"version": schema.StringAttribute{
														Required:    true,
														Description: "The OCSF schema version.",
													},
													"profiles": schema.ListAttribute{
														ElementType: types.StringType,
														Optional:    true,
														Description: "A list of OCSF profiles to apply.",
													},
												},
											},
										},
										"mapping": schema.ListNestedBlock{
											Description: "A list of field mapping rules for transforming log fields to OCSF schema fields.",
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"dest": schema.StringAttribute{
														Required:    true,
														Description: "The destination OCSF field path.",
													},
													"source": schema.StringAttribute{
														Optional:    true,
														Description: "The source field path from the log event.",
													},
													"sources": schema.ListAttribute{
														ElementType: types.StringType,
														Optional:    true,
														Description: "Multiple source field paths for combined mapping.",
													},
													"value": schema.StringAttribute{
														Optional:    true,
														Description: "A static value to use for the destination field.",
													},
													"default": schema.StringAttribute{
														Optional:    true,
														Description: "The default value to use if the source field is missing or empty.",
													},
												},
												Blocks: map[string]schema.Block{
													"lookup": schema.ListNestedBlock{
														Description: "Lookup table configuration for mapping source values to destination values.",
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														NestedObject: schema.NestedBlockObject{
															Attributes: map[string]schema.Attribute{
																"default": schema.StringAttribute{
																	Optional:    true,
																	Description: "The default value to use if no lookup match is found.",
																},
															},
															Blocks: map[string]schema.Block{
																"table": schema.ListNestedBlock{
																	Description: "A list of lookup table entries for value transformation.",
																	NestedObject: schema.NestedBlockObject{
																		Attributes: map[string]schema.Attribute{
																			"contains": schema.StringAttribute{
																				Optional:    true,
																				Description: "The substring to match in the source value.",
																			},
																			"equals": schema.StringAttribute{
																				Optional:    true,
																				Description: "The exact value to match in the source.",
																			},
																			"equals_source": schema.StringAttribute{
																				Optional:    true,
																				Description: "The source field to match against.",
																			},
																			"matches": schema.StringAttribute{
																				Optional:    true,
																				Description: "A regex pattern to match in the source value.",
																			},
																			"not_matches": schema.StringAttribute{
																				Optional:    true,
																				Description: "A regex pattern that must not match the source value.",
																			},
																			"value": schema.StringAttribute{
																				Optional:    true,
																				Description: "The value to use when a match is found.",
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
