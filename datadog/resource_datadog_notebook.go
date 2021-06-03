package datadog

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatadogNotebook() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource for interacting with the notebook API",
		CreateContext: resourceDatadogNotebookCreate,
		ReadContext:   resourceDatadogNotebookRead,
		UpdateContext: resourceDatadogNotebookUpdate,
		DeleteContext: resourceDatadogNotebookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{

			"author": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						"created_at": {
							Type:         schema.TypeString,
							ValidateFunc: validation.IsRFC3339Time,
							Optional:     true,
							Description:  "Creation time of the user.",
						},

						"disabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether the user is disabled.",
						},

						"email": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Email of the user.",
						},

						"handle": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Handle of the user.",
						},

						"icon": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "URL of the user's icon.",
						},

						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name of the user.",
						},

						"status": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Status of the user.",
						},

						"title": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Title of the user.",
						},

						"verified": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether the user is verified.",
						},
					},
				},
			},

			"cell": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of cells to display in the notebook.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						"attributes": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{

									"notebook_distribution_cell_attributes": {
										Type:        schema.TypeList,
										MaxItems:    1,
										Optional:    true,
										Description: "The attributes of a notebook &#x60;distribution&#x60; cell.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{

												"definition": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"legend_size": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "(Deprecated) The widget legend was replaced by a tooltip and sidebar.",
															},

															"marker": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "List of markers.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"display_type": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Combination of:   - A severity error, warning, ok, or info   - A line type: dashed, solid, or bold In this case of a Distribution widget, this can be set to be `x_axis_percentile`. ",
																		},

																		"label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Label to display over the marker.",
																		},

																		"time": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Timestamp for the widget.",
																		},

																		"value": {
																			Type:        schema.TypeString,
																			Required:    true,
																			Description: "Value to apply. Can be a single value y = 15 or a range of values 0 < y < 10.",
																		},
																	},
																},
															},

															"request": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "Array of one request object to display in the widget.  See the dedicated [Request JSON schema documentation](https://docs.datadoghq.com/dashboards/graphing_json/request_json)  to learn how to build the `REQUEST_SCHEMA`.",
																MinItems:    1,

																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"apm_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"event_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"log_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"network_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"process_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"filter_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of processes.",
																						Elem: &schema.Schema{
																							Type: schema.TypeString,
																						},
																					},

																					"limit": {
																						Type:        schema.TypeInt,
																						Optional:    true,
																						Description: "Max number of items in the filter list.",
																					},

																					"metric": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "Your chosen metric.",
																					},

																					"search_by": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Your chosen search term.",
																					},
																				},
																			},
																		},

																		"profile_metrics_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"q": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Widget query.",
																		},

																		"rum_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"security_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"style": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"palette": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Color palette to apply to the widget.",
																					},
																				},
																			},
																		},
																	},
																},
															},

															"show_legend": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "(Deprecated) The widget legend was replaced by a tooltip and sidebar.",
															},

															"time": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Optional:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
																		},
																	},
																},
															},

															"title": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Title of the widget.",
															},

															"title_align": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
																Optional:         true,
																Description:      "How to align the text on the widget.",
															},

															"title_size": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Size of the title.",
															},

															"type": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewDistributionWidgetDefinitionTypeFromValue),
																Required:         true,
																Description:      "Type of the distribution widget.",
															},

															"xaxis": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"include_zero": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "True includes zero.",
																		},

																		"max": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies maximum value to show on the x-axis. It takes a number, percentile (p90 === 90th percentile), or auto for default behavior.",
																			Default:     "auto",
																		},

																		"min": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies minimum value to show on the x-axis. It takes a number, percentile (p90 === 90th percentile), or auto for default behavior.",
																			Default:     "auto",
																		},

																		"scale": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies the scale type. Possible values are `linear`.",
																			Default:     "linear",
																		},
																	},
																},
															},

															"yaxis": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"include_zero": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "True includes zero.",
																		},

																		"label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label of the axis to display on the graph.",
																		},

																		"max": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies the maximum value to show on the y-axis. It takes a number, or auto for default behavior.",
																			Default:     "auto",
																		},

																		"min": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies minimum value to show on the y-axis. It takes a number, or auto for default behavior.",
																			Default:     "auto",
																		},

																		"scale": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies the scale type. Possible values are `linear` or `log`.",
																			Default:     "linear",
																		},
																	},
																},
															},
														},
													},
												},

												"graph_size": {
													Type:             schema.TypeString,
													ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookGraphSizeFromValue),
													Optional:         true,
													Description:      "The size of the graph.",
												},

												"split_by": {
													Type:        schema.TypeList,
													Optional:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"keys": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "Keys to split on.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},

															"tags": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "Tags to split on.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
														},
													},
												},

												"time": {
													Type:        schema.TypeList,
													Optional:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"notebook_absolute_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Absolute timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"end": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The end time.",
																		},

																		"live": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "Indicates whether the timeframe should be shifted to end at the current time.",
																		},

																		"start": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The start time.",
																		},
																	},
																},
															},
															"notebook_relative_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Relative timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Required:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
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
									"notebook_heat_map_cell_attributes": {
										Type:        schema.TypeList,
										MaxItems:    1,
										Optional:    true,
										Description: "The attributes of a notebook &#x60;heatmap&#x60; cell.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{

												"definition": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"custom_link": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "List of custom links.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"is_hidden": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "The flag for toggling context menu link visibility.",
																		},

																		"label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label for the custom link URL. Keep the label short and descriptive. Use metrics and tags as variables.",
																		},

																		"link": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The URL of the custom link. URL must include `http` or `https`. A relative URL must start with `/`.",
																		},

																		"override_label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label ID that refers to a context menu link. Can be `logs`, `hosts`, `traces`, `profiles`, `processes`, `containers`, or `rum`.",
																		},
																	},
																},
															},

															"event": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "List of widget events.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"q": {
																			Type:        schema.TypeString,
																			Required:    true,
																			Description: "Query definition.",
																		},

																		"tags_execution": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The execution method for multi-value filters.",
																		},
																	},
																},
															},

															"legend_size": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Available legend sizes for a widget. Should be one of \"0\", \"2\", \"4\", \"8\", \"16\", or \"auto\".",
															},

															"request": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "List of widget types.",
																MinItems:    1,

																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"apm_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"event_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"search": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "The query being made on the event.",
																					},

																					"tags_execution": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "The execution method for multi-value filters. Can be either and or or.",
																					},
																				},
																			},
																		},

																		"log_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"network_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"process_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"filter_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of processes.",
																						Elem: &schema.Schema{
																							Type: schema.TypeString,
																						},
																					},

																					"limit": {
																						Type:        schema.TypeInt,
																						Optional:    true,
																						Description: "Max number of items in the filter list.",
																					},

																					"metric": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "Your chosen metric.",
																					},

																					"search_by": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Your chosen search term.",
																					},
																				},
																			},
																		},

																		"profile_metrics_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"q": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Widget query.",
																		},

																		"rum_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"security_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"style": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"palette": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Color palette to apply to the widget.",
																					},
																				},
																			},
																		},
																	},
																},
															},

															"show_legend": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Whether or not to display the legend on this widget.",
															},

															"time": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Optional:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
																		},
																	},
																},
															},

															"title": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Title of the widget.",
															},

															"title_align": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
																Optional:         true,
																Description:      "How to align the text on the widget.",
															},

															"title_size": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Size of the title.",
															},

															"type": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewHeatMapWidgetDefinitionTypeFromValue),
																Required:         true,
																Description:      "Type of the heat map widget.",
															},

															"yaxis": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"include_zero": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "True includes zero.",
																		},

																		"label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label of the axis to display on the graph.",
																		},

																		"max": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies the maximum value to show on the y-axis. It takes a number, or auto for default behavior.",
																			Default:     "auto",
																		},

																		"min": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies minimum value to show on the y-axis. It takes a number, or auto for default behavior.",
																			Default:     "auto",
																		},

																		"scale": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies the scale type. Possible values are `linear`, `log`, `sqrt`, `pow##` (e.g. `pow2`, `pow0.5` etc.).",
																			Default:     "linear",
																		},
																	},
																},
															},
														},
													},
												},

												"graph_size": {
													Type:             schema.TypeString,
													ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookGraphSizeFromValue),
													Optional:         true,
													Description:      "The size of the graph.",
												},

												"split_by": {
													Type:        schema.TypeList,
													Optional:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"keys": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "Keys to split on.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},

															"tags": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "Tags to split on.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
														},
													},
												},

												"time": {
													Type:        schema.TypeList,
													Optional:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"notebook_absolute_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Absolute timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"end": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The end time.",
																		},

																		"live": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "Indicates whether the timeframe should be shifted to end at the current time.",
																		},

																		"start": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The start time.",
																		},
																	},
																},
															},
															"notebook_relative_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Relative timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Required:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
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
									"notebook_log_stream_cell_attributes": {
										Type:        schema.TypeList,
										MaxItems:    1,
										Optional:    true,
										Description: "The attributes of a notebook &#x60;log_stream&#x60; cell.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{

												"definition": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"columns": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "Which columns to display on the widget.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},

															"indexes": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "An array of index names to query in the stream. Use [] to query all indexes at once.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},

															"logset": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "ID of the log set to use.",
															},

															"message_display": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetMessageDisplayFromValue),
																Optional:         true,
																Description:      "Amount of log lines to display",
															},

															"query": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Query to filter the log stream with.",
															},

															"show_date_column": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Whether to show the date column or not",
															},

															"show_message_column": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Whether to show the message column or not",
															},

															"sort": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"column": {
																			Type:        schema.TypeString,
																			Required:    true,
																			Description: "Facet path for the column",
																		},

																		"order": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																			Required:         true,
																			Description:      "Widget sorting methods.",
																		},
																	},
																},
															},

															"time": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Optional:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
																		},
																	},
																},
															},

															"title": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Title of the widget.",
															},

															"title_align": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
																Optional:         true,
																Description:      "How to align the text on the widget.",
															},

															"title_size": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Size of the title.",
															},

															"type": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewLogStreamWidgetDefinitionTypeFromValue),
																Required:         true,
																Description:      "Type of the log stream widget.",
															},
														},
													},
												},

												"graph_size": {
													Type:             schema.TypeString,
													ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookGraphSizeFromValue),
													Optional:         true,
													Description:      "The size of the graph.",
												},

												"time": {
													Type:        schema.TypeList,
													Optional:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"notebook_absolute_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Absolute timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"end": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The end time.",
																		},

																		"live": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "Indicates whether the timeframe should be shifted to end at the current time.",
																		},

																		"start": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The start time.",
																		},
																	},
																},
															},
															"notebook_relative_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Relative timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Required:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
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
									"notebook_markdown_cell_attributes": {
										Type:        schema.TypeList,
										MaxItems:    1,
										Optional:    true,
										Description: "The attributes of a notebook &#x60;markdown&#x60; cell.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{

												"definition": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"text": {
																Type:        schema.TypeString,
																Required:    true,
																Description: "The markdown content.",
															},

															"type": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookMarkdownCellDefinitionTypeFromValue),
																Required:         true,
																Description:      "Type of the markdown cell.",
															},
														},
													},
												},
											},
										},
									},
									"notebook_timeseries_cell_attributes": {
										Type:        schema.TypeList,
										MaxItems:    1,
										Optional:    true,
										Description: "The attributes of a notebook &#x60;timeseries&#x60; cell.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{

												"definition": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"custom_link": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "List of custom links.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"is_hidden": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "The flag for toggling context menu link visibility.",
																		},

																		"label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label for the custom link URL. Keep the label short and descriptive. Use metrics and tags as variables.",
																		},

																		"link": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The URL of the custom link. URL must include `http` or `https`. A relative URL must start with `/`.",
																		},

																		"override_label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label ID that refers to a context menu link. Can be `logs`, `hosts`, `traces`, `profiles`, `processes`, `containers`, or `rum`.",
																		},
																	},
																},
															},

															"event": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "List of widget events.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"q": {
																			Type:        schema.TypeString,
																			Required:    true,
																			Description: "Query definition.",
																		},

																		"tags_execution": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The execution method for multi-value filters.",
																		},
																	},
																},
															},

															"legend_columns": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "Columns displayed in the legend.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},

															"legend_layout": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTimeseriesWidgetLegendLayoutFromValue),
																Optional:         true,
																Description:      "Layout of the legend.",
															},

															"legend_size": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Available legend sizes for a widget. Should be one of \"0\", \"2\", \"4\", \"8\", \"16\", or \"auto\".",
															},

															"marker": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "List of markers.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"display_type": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Combination of:   - A severity error, warning, ok, or info   - A line type: dashed, solid, or bold In this case of a Distribution widget, this can be set to be `x_axis_percentile`. ",
																		},

																		"label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Label to display over the marker.",
																		},

																		"time": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Timestamp for the widget.",
																		},

																		"value": {
																			Type:        schema.TypeString,
																			Required:    true,
																			Description: "Value to apply. Can be a single value y = 15 or a range of values 0 < y < 10.",
																		},
																	},
																},
															},

															"request": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "List of timeseries widget requests.",
																MinItems:    1,

																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"apm_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"display_type": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetDisplayTypeFromValue),
																			Optional:         true,
																			Description:      "Type of display to use for the request.",
																		},

																		"event_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"formula": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "List of formulas that operate on queries. **This feature is currently in beta.**",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"alias": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Expression alias.",
																					},

																					"formula": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "String expression built from queries, formulas, and functions.",
																					},

																					"limit": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"count": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Number of results to return.",
																								},

																								"order": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
																									Optional:         true,
																									Description:      "Direction of sort.",
																									Default:          "desc",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"log_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"metadata": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "Used to define expression aliases.",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"alias_name": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Expression alias.",
																					},

																					"expression": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "Expression name.",
																					},
																				},
																			},
																		},

																		"network_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"on_right_yaxis": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "Whether or not to display a second y-axis on the right.",
																		},

																		"process_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"filter_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of processes.",
																						Elem: &schema.Schema{
																							Type: schema.TypeString,
																						},
																					},

																					"limit": {
																						Type:        schema.TypeInt,
																						Optional:    true,
																						Description: "Max number of items in the filter list.",
																					},

																					"metric": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "Your chosen metric.",
																					},

																					"search_by": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Your chosen search term.",
																					},
																				},
																			},
																		},

																		"profile_metrics_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"q": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Widget query.",
																		},

																		"query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "List of queries that can be returned directly or used in formulas. **This feature is currently in beta.**",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"formula_and_function_event_query_definition": {
																						Type:        schema.TypeList,
																						MaxItems:    1,
																						Optional:    true,
																						Description: "A formula and functions events query.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"compute": {
																									Type:        schema.TypeList,
																									Required:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionEventAggregationFromValue),
																												Required:         true,
																												Description:      "Aggregation methods for event platform queries.",
																											},

																											"interval": {
																												Type:        schema.TypeInt,
																												Optional:    true,
																												Description: "A time interval in milliseconds.",
																											},

																											"metric": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Measurable attribute to compute.",
																											},
																										},
																									},
																								},

																								"data_source": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionEventsDataSourceFromValue),
																									Required:         true,
																									Description:      "Data source for event platform-based queries.",
																								},

																								"group_by": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "Group by options.",
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"facet": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "Event facet.",
																											},

																											"limit": {
																												Type:        schema.TypeInt,
																												Optional:    true,
																												Description: "Number of groups to return.",
																											},

																											"sort": {
																												Type:        schema.TypeList,
																												Optional:    true,
																												Description: "",
																												MaxItems:    1,
																												Elem: &schema.Resource{
																													Schema: map[string]*schema.Schema{

																														"aggregation": {
																															Type:             schema.TypeString,
																															ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionEventAggregationFromValue),
																															Required:         true,
																															Description:      "Aggregation methods for event platform queries.",
																														},

																														"metric": {
																															Type:        schema.TypeString,
																															Optional:    true,
																															Description: "Metric used for sorting group by results.",
																														},

																														"order": {
																															Type:             schema.TypeString,
																															ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
																															Optional:         true,
																															Description:      "Direction of sort.",
																															Default:          "desc",
																														},
																													},
																												},
																											},
																										},
																									},
																								},

																								"indexes": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "An array of index names to query in the stream. Omit or use `[]` to query all indexes at once.",
																									Elem: &schema.Schema{
																										Type: schema.TypeString,
																									},
																								},

																								"name": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Name of the query for use in formulas.",
																								},

																								"search": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"query": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "Events search string.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},
																					"formula_and_function_metric_query_definition": {
																						Type:        schema.TypeList,
																						MaxItems:    1,
																						Optional:    true,
																						Description: "A formula and functions metrics query.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregator": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionMetricAggregationFromValue),
																									Optional:         true,
																									Description:      "The aggregation methods available for metrics queries.",
																								},

																								"data_source": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionMetricDataSourceFromValue),
																									Required:         true,
																									Description:      "Data source for metrics queries.",
																								},

																								"name": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Name of the query for use in formulas.",
																								},

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Metrics query definition.",
																								},
																							},
																						},
																					},
																					"formula_and_function_process_query_definition": {
																						Type:        schema.TypeList,
																						MaxItems:    1,
																						Optional:    true,
																						Description: "Process query using formulas and functions.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregator": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionMetricAggregationFromValue),
																									Optional:         true,
																									Description:      "The aggregation methods available for metrics queries.",
																								},

																								"data_source": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionProcessQueryDataSourceFromValue),
																									Required:         true,
																									Description:      "Data sources that rely on the process backend.",
																								},

																								"is_normalized_cpu": {
																									Type:        schema.TypeBool,
																									Optional:    true,
																									Description: "Whether to normalize the CPU percentages.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Number of hits to return.",
																								},

																								"metric": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Process metric name.",
																								},

																								"name": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Name of query for use in formulas.",
																								},

																								"sort": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
																									Optional:         true,
																									Description:      "Direction of sort.",
																									Default:          "desc",
																								},

																								"tag_filters": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "An array of tags to filter by.",
																									Elem: &schema.Schema{
																										Type: schema.TypeString,
																									},
																								},

																								"text_filter": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Text to use as filter.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"response_format": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionResponseFormatFromValue),
																			Optional:         true,
																			Description:      "Timeseries or Scalar response. **This feature is currently in beta.**",
																		},

																		"rum_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"security_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"style": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"line_type": {
																						Type:             schema.TypeString,
																						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLineTypeFromValue),
																						Optional:         true,
																						Description:      "Type of lines displayed.",
																					},

																					"line_width": {
																						Type:             schema.TypeString,
																						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLineWidthFromValue),
																						Optional:         true,
																						Description:      "Width of line displayed.",
																					},

																					"palette": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Color palette to apply to the widget.",
																					},
																				},
																			},
																		},
																	},
																},
															},

															"right_yaxis": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"include_zero": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "True includes zero.",
																		},

																		"label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label of the axis to display on the graph.",
																		},

																		"max": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies the maximum value to show on the y-axis. It takes a number, or auto for default behavior.",
																			Default:     "auto",
																		},

																		"min": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies minimum value to show on the y-axis. It takes a number, or auto for default behavior.",
																			Default:     "auto",
																		},

																		"scale": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies the scale type. Possible values are `linear`, `log`, `sqrt`, `pow##` (e.g. `pow2`, `pow0.5` etc.).",
																			Default:     "linear",
																		},
																	},
																},
															},

															"show_legend": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "(screenboard only) Show the legend for this widget.",
															},

															"time": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Optional:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
																		},
																	},
																},
															},

															"title": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Title of your widget.",
															},

															"title_align": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
																Optional:         true,
																Description:      "How to align the text on the widget.",
															},

															"title_size": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Size of the title.",
															},

															"type": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTimeseriesWidgetDefinitionTypeFromValue),
																Required:         true,
																Description:      "Type of the timeseries widget.",
															},

															"yaxis": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"include_zero": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "True includes zero.",
																		},

																		"label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label of the axis to display on the graph.",
																		},

																		"max": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies the maximum value to show on the y-axis. It takes a number, or auto for default behavior.",
																			Default:     "auto",
																		},

																		"min": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies minimum value to show on the y-axis. It takes a number, or auto for default behavior.",
																			Default:     "auto",
																		},

																		"scale": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Specifies the scale type. Possible values are `linear`, `log`, `sqrt`, `pow##` (e.g. `pow2`, `pow0.5` etc.).",
																			Default:     "linear",
																		},
																	},
																},
															},
														},
													},
												},

												"graph_size": {
													Type:             schema.TypeString,
													ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookGraphSizeFromValue),
													Optional:         true,
													Description:      "The size of the graph.",
												},

												"split_by": {
													Type:        schema.TypeList,
													Optional:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"keys": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "Keys to split on.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},

															"tags": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "Tags to split on.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
														},
													},
												},

												"time": {
													Type:        schema.TypeList,
													Optional:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"notebook_absolute_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Absolute timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"end": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The end time.",
																		},

																		"live": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "Indicates whether the timeframe should be shifted to end at the current time.",
																		},

																		"start": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The start time.",
																		},
																	},
																},
															},
															"notebook_relative_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Relative timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Required:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
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
									"notebook_toplist_cell_attributes": {
										Type:        schema.TypeList,
										MaxItems:    1,
										Optional:    true,
										Description: "The attributes of a notebook &#x60;toplist&#x60; cell.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{

												"definition": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"custom_link": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "List of custom links.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"is_hidden": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "The flag for toggling context menu link visibility.",
																		},

																		"label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label for the custom link URL. Keep the label short and descriptive. Use metrics and tags as variables.",
																		},

																		"link": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The URL of the custom link. URL must include `http` or `https`. A relative URL must start with `/`.",
																		},

																		"override_label": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "The label ID that refers to a context menu link. Can be `logs`, `hosts`, `traces`, `profiles`, `processes`, `containers`, or `rum`.",
																		},
																	},
																},
															},

															"request": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "List of top list widget requests.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"apm_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"conditional_format": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "List of conditional formats.",
																			MinItems:    1,

																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"comparator": {
																						Type:             schema.TypeString,
																						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetComparatorFromValue),
																						Required:         true,
																						Description:      "Comparator to apply.",
																					},

																					"custom_bg_color": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Color palette to apply to the background, same values available as palette.",
																					},

																					"custom_fg_color": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Color palette to apply to the foreground, same values available as palette.",
																					},

																					"hide_value": {
																						Type:        schema.TypeBool,
																						Optional:    true,
																						Description: "True hides values.",
																					},

																					"image_url": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Displays an image as the background.",
																					},

																					"metric": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Metric from the request to correlate this conditional format with.",
																					},

																					"palette": {
																						Type:             schema.TypeString,
																						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetPaletteFromValue),
																						Required:         true,
																						Description:      "Color palette to apply.",
																					},

																					"timeframe": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Defines the displayed timeframe.",
																					},

																					"value": {
																						Type:        schema.TypeFloat,
																						Required:    true,
																						Description: "Value for the comparator.",
																					},
																				},
																			},
																		},

																		"event_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"formula": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "List of formulas that operate on queries. **This feature is currently in beta.**",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"alias": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Expression alias.",
																					},

																					"formula": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "String expression built from queries, formulas, and functions.",
																					},

																					"limit": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"count": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Number of results to return.",
																								},

																								"order": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
																									Optional:         true,
																									Description:      "Direction of sort.",
																									Default:          "desc",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"log_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"network_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"process_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"filter_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of processes.",
																						Elem: &schema.Schema{
																							Type: schema.TypeString,
																						},
																					},

																					"limit": {
																						Type:        schema.TypeInt,
																						Optional:    true,
																						Description: "Max number of items in the filter list.",
																					},

																					"metric": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "Your chosen metric.",
																					},

																					"search_by": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Your chosen search term.",
																					},
																				},
																			},
																		},

																		"profile_metrics_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"q": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Widget query.",
																		},

																		"query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "List of queries that can be returned directly or used in formulas. **This feature is currently in beta.**",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"formula_and_function_event_query_definition": {
																						Type:        schema.TypeList,
																						MaxItems:    1,
																						Optional:    true,
																						Description: "A formula and functions events query.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"compute": {
																									Type:        schema.TypeList,
																									Required:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionEventAggregationFromValue),
																												Required:         true,
																												Description:      "Aggregation methods for event platform queries.",
																											},

																											"interval": {
																												Type:        schema.TypeInt,
																												Optional:    true,
																												Description: "A time interval in milliseconds.",
																											},

																											"metric": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Measurable attribute to compute.",
																											},
																										},
																									},
																								},

																								"data_source": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionEventsDataSourceFromValue),
																									Required:         true,
																									Description:      "Data source for event platform-based queries.",
																								},

																								"group_by": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "Group by options.",
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"facet": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "Event facet.",
																											},

																											"limit": {
																												Type:        schema.TypeInt,
																												Optional:    true,
																												Description: "Number of groups to return.",
																											},

																											"sort": {
																												Type:        schema.TypeList,
																												Optional:    true,
																												Description: "",
																												MaxItems:    1,
																												Elem: &schema.Resource{
																													Schema: map[string]*schema.Schema{

																														"aggregation": {
																															Type:             schema.TypeString,
																															ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionEventAggregationFromValue),
																															Required:         true,
																															Description:      "Aggregation methods for event platform queries.",
																														},

																														"metric": {
																															Type:        schema.TypeString,
																															Optional:    true,
																															Description: "Metric used for sorting group by results.",
																														},

																														"order": {
																															Type:             schema.TypeString,
																															ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
																															Optional:         true,
																															Description:      "Direction of sort.",
																															Default:          "desc",
																														},
																													},
																												},
																											},
																										},
																									},
																								},

																								"indexes": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "An array of index names to query in the stream. Omit or use `[]` to query all indexes at once.",
																									Elem: &schema.Schema{
																										Type: schema.TypeString,
																									},
																								},

																								"name": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Name of the query for use in formulas.",
																								},

																								"search": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"query": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "Events search string.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},
																					"formula_and_function_metric_query_definition": {
																						Type:        schema.TypeList,
																						MaxItems:    1,
																						Optional:    true,
																						Description: "A formula and functions metrics query.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregator": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionMetricAggregationFromValue),
																									Optional:         true,
																									Description:      "The aggregation methods available for metrics queries.",
																								},

																								"data_source": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionMetricDataSourceFromValue),
																									Required:         true,
																									Description:      "Data source for metrics queries.",
																								},

																								"name": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Name of the query for use in formulas.",
																								},

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Metrics query definition.",
																								},
																							},
																						},
																					},
																					"formula_and_function_process_query_definition": {
																						Type:        schema.TypeList,
																						MaxItems:    1,
																						Optional:    true,
																						Description: "Process query using formulas and functions.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregator": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionMetricAggregationFromValue),
																									Optional:         true,
																									Description:      "The aggregation methods available for metrics queries.",
																								},

																								"data_source": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionProcessQueryDataSourceFromValue),
																									Required:         true,
																									Description:      "Data sources that rely on the process backend.",
																								},

																								"is_normalized_cpu": {
																									Type:        schema.TypeBool,
																									Optional:    true,
																									Description: "Whether to normalize the CPU percentages.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Number of hits to return.",
																								},

																								"metric": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Process metric name.",
																								},

																								"name": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Name of query for use in formulas.",
																								},

																								"sort": {
																									Type:             schema.TypeString,
																									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
																									Optional:         true,
																									Description:      "Direction of sort.",
																									Default:          "desc",
																								},

																								"tag_filters": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "An array of tags to filter by.",
																									Elem: &schema.Schema{
																										Type: schema.TypeString,
																									},
																								},

																								"text_filter": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Text to use as filter.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"response_format": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionResponseFormatFromValue),
																			Optional:         true,
																			Description:      "Timeseries or Scalar response. **This feature is currently in beta.**",
																		},

																		"rum_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"security_query": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"group_by": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "List of tag prefixes to group by in the case of a cluster check.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"facet": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Facet name.",
																								},

																								"limit": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Maximum number of items in the group.",
																								},

																								"sort": {
																									Type:        schema.TypeList,
																									Optional:    true,
																									Description: "",
																									MaxItems:    1,
																									Elem: &schema.Resource{
																										Schema: map[string]*schema.Schema{

																											"aggregation": {
																												Type:        schema.TypeString,
																												Required:    true,
																												Description: "The aggregation method.",
																											},

																											"facet": {
																												Type:        schema.TypeString,
																												Optional:    true,
																												Description: "Facet name.",
																											},

																											"order": {
																												Type:             schema.TypeString,
																												ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
																												Required:         true,
																												Description:      "Widget sorting methods.",
																											},
																										},
																									},
																								},
																							},
																						},
																					},

																					"index": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "A coma separated-list of index names. Use \"*\" query all indexes at once. [Multiple Indexes](https://docs.datadoghq.com/logs/indexes/#multiple-indexes)",
																					},

																					"multi_compute": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "This field is mutually exclusive with `compute`.",
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"aggregation": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "The aggregation method.",
																								},

																								"facet": {
																									Type:        schema.TypeString,
																									Optional:    true,
																									Description: "Facet name.",
																								},

																								"interval": {
																									Type:        schema.TypeInt,
																									Optional:    true,
																									Description: "Define a time interval in seconds.",
																								},
																							},
																						},
																					},

																					"search": {
																						Type:        schema.TypeList,
																						Optional:    true,
																						Description: "",
																						MaxItems:    1,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{

																								"query": {
																									Type:        schema.TypeString,
																									Required:    true,
																									Description: "Search value to apply.",
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"style": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			Description: "",
																			MaxItems:    1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{

																					"line_type": {
																						Type:             schema.TypeString,
																						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLineTypeFromValue),
																						Optional:         true,
																						Description:      "Type of lines displayed.",
																					},

																					"line_width": {
																						Type:             schema.TypeString,
																						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLineWidthFromValue),
																						Optional:         true,
																						Description:      "Width of line displayed.",
																					},

																					"palette": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Color palette to apply to the widget.",
																					},
																				},
																			},
																		},
																	},
																},
															},

															"time": {
																Type:        schema.TypeList,
																Optional:    true,
																Description: "",
																MaxItems:    1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Optional:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
																		},
																	},
																},
															},

															"title": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Title of your widget.",
															},

															"title_align": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
																Optional:         true,
																Description:      "How to align the text on the widget.",
															},

															"title_size": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Size of the title.",
															},

															"type": {
																Type:             schema.TypeString,
																ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewToplistWidgetDefinitionTypeFromValue),
																Required:         true,
																Description:      "Type of the top list widget.",
															},
														},
													},
												},

												"graph_size": {
													Type:             schema.TypeString,
													ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookGraphSizeFromValue),
													Optional:         true,
													Description:      "The size of the graph.",
												},

												"split_by": {
													Type:        schema.TypeList,
													Optional:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"keys": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "Keys to split on.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},

															"tags": {
																Type:        schema.TypeList,
																Required:    true,
																Description: "Tags to split on.",
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
														},
													},
												},

												"time": {
													Type:        schema.TypeList,
													Optional:    true,
													Description: "",
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{

															"notebook_absolute_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Absolute timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"end": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The end time.",
																		},

																		"live": {
																			Type:        schema.TypeBool,
																			Optional:    true,
																			Description: "Indicates whether the timeframe should be shifted to end at the current time.",
																		},

																		"start": {
																			Type:         schema.TypeString,
																			ValidateFunc: validation.IsRFC3339Time,
																			Required:     true,
																			Description:  "The start time.",
																		},
																	},
																},
															},
															"notebook_relative_time": {
																Type:        schema.TypeList,
																MaxItems:    1,
																Optional:    true,
																Description: "Relative timeframe.",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{

																		"live_span": {
																			Type:             schema.TypeString,
																			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
																			Required:         true,
																			Description:      "The available timeframes depend on the widget you are using.",
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

						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Notebook cell ID.",
						},

						"type": {
							Type:             schema.TypeString,
							ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookCellResourceTypeFromValue),
							Required:         true,
							Description:      "Type of the Notebook Cell resource.",
						},
					},
				},
			},

			"created": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsRFC3339Time,
				Optional:     true,
				Computed:     true,
				Description:  "UTC time stamp for when the notebook was created.",
			},

			"modified": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsRFC3339Time,
				Optional:     true,
				Computed:     true,
				Description:  "UTC time stamp for when the notebook was last modified.",
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the notebook.",
			},

			"status": {
				Type:             schema.TypeString,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookStatusFromValue),
				Optional:         true,
				Description:      "Publication status of the notebook. For now, always \"published\".",
				Default:          "published",
			},

			"time": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						"notebook_absolute_time": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Absolute timeframe.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{

									"end": {
										Type:         schema.TypeString,
										ValidateFunc: validation.IsRFC3339Time,
										Required:     true,
										Description:  "The end time.",
									},

									"live": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Indicates whether the timeframe should be shifted to end at the current time.",
									},

									"start": {
										Type:         schema.TypeString,
										ValidateFunc: validation.IsRFC3339Time,
										Required:     true,
										Description:  "The start time.",
									},
								},
							},
						},
						"notebook_relative_time": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Relative timeframe.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{

									"live_span": {
										Type:             schema.TypeString,
										ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
										Required:         true,
										Description:      "The available timeframes depend on the widget you are using.",
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

func buildDatadogNotebook(d *schema.ResourceData) (*datadogV1.NotebookCreateDataAttributes, error) {
	k := utils.NewResourceDataKey(d, "")
	result := datadogV1.NewNotebookCreateDataAttributesWithDefaults()
	k.Add("cell")
	if cellsArray, ok := k.GetOk(); ok {
		cellsDDArray := make([]datadogV1.NotebookCellCreateRequest, 0)
		for i := range cellsArray.([]interface{}) {
			k.Add(i)

			cellsDDArrayItem := datadogV1.NewNotebookCellCreateRequestWithDefaults()

			// handle attributes, which is a nested model
			k.Add("attributes.0")

			cellsDDArrayItemAttributes := &datadogV1.NotebookCellCreateRequestAttributes{}
			// handle notebook_cell_create_request_attributes, which is a oneOf model
			k.Add("notebook_markdown_cell_attributes.0")
			if _, ok := k.GetOk(); ok {

				ddNotebookMarkdownCellAttributes := datadogV1.NewNotebookMarkdownCellAttributesWithDefaults()

				// handle definition, which is a nested model
				k.Add("definition.0")

				ddNotebookMarkdownCellAttributesDefinition := datadogV1.NewNotebookMarkdownCellDefinitionWithDefaults()

				if v, ok := k.GetOkWith("text"); ok {
					ddNotebookMarkdownCellAttributesDefinition.SetText(v.(string))
				}

				if v, ok := k.GetOkWith("type"); ok {
					ddNotebookMarkdownCellAttributesDefinition.SetType(datadogV1.NotebookMarkdownCellDefinitionType(v.(string)))
				}
				k.Remove("definition.0")
				ddNotebookMarkdownCellAttributes.SetDefinition(*ddNotebookMarkdownCellAttributesDefinition)
				cellsDDArrayItemAttributes.NotebookMarkdownCellAttributes = ddNotebookMarkdownCellAttributes
			}
			k.Remove("notebook_markdown_cell_attributes.0")
			k.Add("notebook_timeseries_cell_attributes.0")
			if _, ok := k.GetOk(); ok {

				ddNotebookTimeseriesCellAttributes := datadogV1.NewNotebookTimeseriesCellAttributesWithDefaults()

				// handle definition, which is a nested model
				k.Add("definition.0")

				ddNotebookTimeseriesCellAttributesDefinition := datadogV1.NewTimeseriesWidgetDefinitionWithDefaults()
				k.Add("custom_link")
				if customLinksArray, ok := k.GetOk(); ok {
					customLinksDDArray := make([]datadogV1.WidgetCustomLink, 0)
					for i := range customLinksArray.([]interface{}) {
						k.Add(i)

						customLinksDDArrayItem := datadogV1.NewWidgetCustomLinkWithDefaults()

						if v, ok := k.GetOkWith("is_hidden"); ok {
							customLinksDDArrayItem.SetIsHidden(v.(bool))
						}

						if v, ok := k.GetOkWith("label"); ok {
							customLinksDDArrayItem.SetLabel(v.(string))
						}

						if v, ok := k.GetOkWith("link"); ok {
							customLinksDDArrayItem.SetLink(v.(string))
						}

						if v, ok := k.GetOkWith("override_label"); ok {
							customLinksDDArrayItem.SetOverrideLabel(v.(string))
						}
						customLinksDDArray = append(customLinksDDArray, *customLinksDDArrayItem)
						k.Remove(i)
					}
					ddNotebookTimeseriesCellAttributesDefinition.SetCustomLinks(customLinksDDArray)
				}
				k.Remove("custom_link")
				k.Add("event")
				if eventsArray, ok := k.GetOk(); ok {
					eventsDDArray := make([]datadogV1.WidgetEvent, 0)
					for i := range eventsArray.([]interface{}) {
						k.Add(i)

						eventsDDArrayItem := datadogV1.NewWidgetEventWithDefaults()

						if v, ok := k.GetOkWith("q"); ok {
							eventsDDArrayItem.SetQ(v.(string))
						}

						if v, ok := k.GetOkWith("tags_execution"); ok {
							eventsDDArrayItem.SetTagsExecution(v.(string))
						}
						eventsDDArray = append(eventsDDArray, *eventsDDArrayItem)
						k.Remove(i)
					}
					ddNotebookTimeseriesCellAttributesDefinition.SetEvents(eventsDDArray)
				}
				k.Remove("event")
				k.Add("legend_columns")
				if legendColumnsArray, ok := k.GetOk(); ok {
					legendColumnsDDArray := make([]datadogV1.TimeseriesWidgetLegendColumn, 0)
					for i := range legendColumnsArray.([]interface{}) {
						legendColumnsArrayItem := k.GetWith(i)
						legendColumnsDDArray = append(legendColumnsDDArray, datadogV1.TimeseriesWidgetLegendColumn(legendColumnsArrayItem.(string)))
					}
					ddNotebookTimeseriesCellAttributesDefinition.SetLegendColumns(legendColumnsDDArray)
				}
				k.Remove("legend_columns")

				if v, ok := k.GetOkWith("legend_layout"); ok {
					ddNotebookTimeseriesCellAttributesDefinition.SetLegendLayout(datadogV1.TimeseriesWidgetLegendLayout(v.(string)))
				}

				if v, ok := k.GetOkWith("legend_size"); ok {
					ddNotebookTimeseriesCellAttributesDefinition.SetLegendSize(v.(string))
				}
				k.Add("marker")
				if markersArray, ok := k.GetOk(); ok {
					markersDDArray := make([]datadogV1.WidgetMarker, 0)
					for i := range markersArray.([]interface{}) {
						k.Add(i)

						markersDDArrayItem := datadogV1.NewWidgetMarkerWithDefaults()

						if v, ok := k.GetOkWith("display_type"); ok {
							markersDDArrayItem.SetDisplayType(v.(string))
						}

						if v, ok := k.GetOkWith("label"); ok {
							markersDDArrayItem.SetLabel(v.(string))
						}

						if v, ok := k.GetOkWith("time"); ok {
							markersDDArrayItem.SetTime(v.(string))
						}

						if v, ok := k.GetOkWith("value"); ok {
							markersDDArrayItem.SetValue(v.(string))
						}
						markersDDArray = append(markersDDArray, *markersDDArrayItem)
						k.Remove(i)
					}
					ddNotebookTimeseriesCellAttributesDefinition.SetMarkers(markersDDArray)
				}
				k.Remove("marker")
				k.Add("request")
				if requestsArray, ok := k.GetOk(); ok {
					requestsDDArray := make([]datadogV1.TimeseriesWidgetRequest, 0)
					for i := range requestsArray.([]interface{}) {
						k.Add(i)

						requestsDDArrayItem := datadogV1.NewTimeseriesWidgetRequestWithDefaults()

						// handle apm_query, which is a nested model
						k.Add("apm_query.0")

						requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemApmQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
						k.Remove("apm_query.0")
						requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)

						if v, ok := k.GetOkWith("display_type"); ok {
							requestsDDArrayItem.SetDisplayType(datadogV1.WidgetDisplayType(v.(string)))
						}

						// handle event_query, which is a nested model
						k.Add("event_query.0")

						requestsDDArrayItemEventQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemEventQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemEventQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemEventQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemEventQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemEventQuery.SetCompute(*requestsDDArrayItemEventQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemEventQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemEventQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemEventQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemEventQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemEventQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemEventQuery.SetSearch(*requestsDDArrayItemEventQuerySearch)
						k.Remove("event_query.0")
						requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)
						k.Add("formula")
						if formulasArray, ok := k.GetOk(); ok {
							formulasDDArray := make([]datadogV1.WidgetFormula, 0)
							for i := range formulasArray.([]interface{}) {
								k.Add(i)

								formulasDDArrayItem := datadogV1.NewWidgetFormulaWithDefaults()

								if v, ok := k.GetOkWith("alias"); ok {
									formulasDDArrayItem.SetAlias(v.(string))
								}

								if v, ok := k.GetOkWith("formula"); ok {
									formulasDDArrayItem.SetFormula(v.(string))
								}

								// handle limit, which is a nested model
								k.Add("limit.0")

								formulasDDArrayItemLimit := datadogV1.NewWidgetFormulaLimitWithDefaults()

								if v, ok := k.GetOkWith("count"); ok {
									formulasDDArrayItemLimit.SetCount(int64(v.(int)))
								}

								if v, ok := k.GetOkWith("order"); ok {
									formulasDDArrayItemLimit.SetOrder(datadogV1.QuerySortOrder(v.(string)))
								}
								k.Remove("limit.0")
								formulasDDArrayItem.SetLimit(*formulasDDArrayItemLimit)
								formulasDDArray = append(formulasDDArray, *formulasDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItem.SetFormulas(formulasDDArray)
						}
						k.Remove("formula")

						// handle log_query, which is a nested model
						k.Add("log_query.0")

						requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemLogQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
						k.Remove("log_query.0")
						requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)
						k.Add("metadata")
						if metadataArray, ok := k.GetOk(); ok {
							metadataDDArray := make([]datadogV1.TimeseriesWidgetExpressionAlias, 0)
							for i := range metadataArray.([]interface{}) {
								k.Add(i)

								metadataDDArrayItem := datadogV1.NewTimeseriesWidgetExpressionAliasWithDefaults()

								if v, ok := k.GetOkWith("alias_name"); ok {
									metadataDDArrayItem.SetAliasName(v.(string))
								}

								if v, ok := k.GetOkWith("expression"); ok {
									metadataDDArrayItem.SetExpression(v.(string))
								}
								metadataDDArray = append(metadataDDArray, *metadataDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItem.SetMetadata(metadataDDArray)
						}
						k.Remove("metadata")

						// handle network_query, which is a nested model
						k.Add("network_query.0")

						requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
						k.Remove("network_query.0")
						requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

						if v, ok := k.GetOkWith("on_right_yaxis"); ok {
							requestsDDArrayItem.SetOnRightYaxis(v.(bool))
						}

						// handle process_query, which is a nested model
						k.Add("process_query.0")

						requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
						k.Add("filter_by")
						if filterByArray, ok := k.GetOk(); ok {
							filterByDDArray := make([]string, 0)
							for i := range filterByArray.([]interface{}) {
								filterByArrayItem := k.GetWith(i)
								filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
							}
							requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
						}
						k.Remove("filter_by")

						if v, ok := k.GetOkWith("limit"); ok {
							requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
						}

						if v, ok := k.GetOkWith("metric"); ok {
							requestsDDArrayItemProcessQuery.SetMetric(v.(string))
						}

						if v, ok := k.GetOkWith("search_by"); ok {
							requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
						}
						k.Remove("process_query.0")
						requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

						// handle profile_metrics_query, which is a nested model
						k.Add("profile_metrics_query.0")

						requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
						k.Remove("profile_metrics_query.0")
						requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

						if v, ok := k.GetOkWith("q"); ok {
							requestsDDArrayItem.SetQ(v.(string))
						}
						k.Add("query")
						if queriesArray, ok := k.GetOk(); ok {
							queriesDDArray := make([]datadogV1.FormulaAndFunctionQueryDefinition, 0)
							for i := range queriesArray.([]interface{}) {
								k.Add(i)

								queriesDDArrayItem := &datadogV1.FormulaAndFunctionQueryDefinition{}
								// handle formula_and_function_query_definition, which is a oneOf model
								k.Add("formula_and_function_metric_query_definition.0")
								if _, ok := k.GetOk(); ok {

									ddFormulaAndFunctionMetricQueryDefinition := datadogV1.NewFormulaAndFunctionMetricQueryDefinitionWithDefaults()

									if v, ok := k.GetOkWith("aggregator"); ok {
										ddFormulaAndFunctionMetricQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
									}

									if v, ok := k.GetOkWith("data_source"); ok {
										ddFormulaAndFunctionMetricQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionMetricDataSource(v.(string)))
									}

									if v, ok := k.GetOkWith("name"); ok {
										ddFormulaAndFunctionMetricQueryDefinition.SetName(v.(string))
									}

									if v, ok := k.GetOkWith("query"); ok {
										ddFormulaAndFunctionMetricQueryDefinition.SetQuery(v.(string))
									}
									queriesDDArrayItem.FormulaAndFunctionMetricQueryDefinition = ddFormulaAndFunctionMetricQueryDefinition
								}
								k.Remove("formula_and_function_metric_query_definition.0")
								k.Add("formula_and_function_event_query_definition.0")
								if _, ok := k.GetOk(); ok {

									ddFormulaAndFunctionEventQueryDefinition := datadogV1.NewFormulaAndFunctionEventQueryDefinitionWithDefaults()

									// handle compute, which is a nested model
									k.Add("compute.0")

									ddFormulaAndFunctionEventQueryDefinitionCompute := datadogV1.NewFormulaAndFunctionEventQueryDefinitionComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										ddFormulaAndFunctionEventQueryDefinitionCompute.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										ddFormulaAndFunctionEventQueryDefinitionCompute.SetInterval(int64(v.(int)))
									}

									if v, ok := k.GetOkWith("metric"); ok {
										ddFormulaAndFunctionEventQueryDefinitionCompute.SetMetric(v.(string))
									}
									k.Remove("compute.0")
									ddFormulaAndFunctionEventQueryDefinition.SetCompute(*ddFormulaAndFunctionEventQueryDefinitionCompute)

									if v, ok := k.GetOkWith("data_source"); ok {
										ddFormulaAndFunctionEventQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionEventsDataSource(v.(string)))
									}
									k.Add("group_by")
									if groupByArray, ok := k.GetOk(); ok {
										groupByDDArray := make([]datadogV1.FormulaAndFunctionEventQueryGroupBy, 0)
										for i := range groupByArray.([]interface{}) {
											k.Add(i)

											groupByDDArrayItem := datadogV1.NewFormulaAndFunctionEventQueryGroupByWithDefaults()

											if v, ok := k.GetOkWith("facet"); ok {
												groupByDDArrayItem.SetFacet(v.(string))
											}

											if v, ok := k.GetOkWith("limit"); ok {
												groupByDDArrayItem.SetLimit(int64(v.(int)))
											}

											// handle sort, which is a nested model
											k.Add("sort.0")

											groupByDDArrayItemSort := datadogV1.NewFormulaAndFunctionEventQueryGroupBySortWithDefaults()

											if v, ok := k.GetOkWith("aggregation"); ok {
												groupByDDArrayItemSort.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
											}

											if v, ok := k.GetOkWith("metric"); ok {
												groupByDDArrayItemSort.SetMetric(v.(string))
											}

											if v, ok := k.GetOkWith("order"); ok {
												groupByDDArrayItemSort.SetOrder(datadogV1.QuerySortOrder(v.(string)))
											}
											k.Remove("sort.0")
											groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
											groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
											k.Remove(i)
										}
										ddFormulaAndFunctionEventQueryDefinition.SetGroupBy(groupByDDArray)
									}
									k.Remove("group_by")
									k.Add("indexes")
									if indexesArray, ok := k.GetOk(); ok {
										indexesDDArray := make([]string, 0)
										for i := range indexesArray.([]interface{}) {
											indexesArrayItem := k.GetWith(i)
											indexesDDArray = append(indexesDDArray, indexesArrayItem.(string))
										}
										ddFormulaAndFunctionEventQueryDefinition.SetIndexes(indexesDDArray)
									}
									k.Remove("indexes")

									if v, ok := k.GetOkWith("name"); ok {
										ddFormulaAndFunctionEventQueryDefinition.SetName(v.(string))
									}

									// handle search, which is a nested model
									k.Add("search.0")

									ddFormulaAndFunctionEventQueryDefinitionSearch := datadogV1.NewFormulaAndFunctionEventQueryDefinitionSearchWithDefaults()

									if v, ok := k.GetOkWith("query"); ok {
										ddFormulaAndFunctionEventQueryDefinitionSearch.SetQuery(v.(string))
									}
									k.Remove("search.0")
									ddFormulaAndFunctionEventQueryDefinition.SetSearch(*ddFormulaAndFunctionEventQueryDefinitionSearch)
									queriesDDArrayItem.FormulaAndFunctionEventQueryDefinition = ddFormulaAndFunctionEventQueryDefinition
								}
								k.Remove("formula_and_function_event_query_definition.0")
								k.Add("formula_and_function_process_query_definition.0")
								if _, ok := k.GetOk(); ok {

									ddFormulaAndFunctionProcessQueryDefinition := datadogV1.NewFormulaAndFunctionProcessQueryDefinitionWithDefaults()

									if v, ok := k.GetOkWith("aggregator"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
									}

									if v, ok := k.GetOkWith("data_source"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionProcessQueryDataSource(v.(string)))
									}

									if v, ok := k.GetOkWith("is_normalized_cpu"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetIsNormalizedCpu(v.(bool))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetLimit(int64(v.(int)))
									}

									if v, ok := k.GetOkWith("metric"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetMetric(v.(string))
									}

									if v, ok := k.GetOkWith("name"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetName(v.(string))
									}

									if v, ok := k.GetOkWith("sort"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetSort(datadogV1.QuerySortOrder(v.(string)))
									}
									k.Add("tag_filters")
									if tagFiltersArray, ok := k.GetOk(); ok {
										tagFiltersDDArray := make([]string, 0)
										for i := range tagFiltersArray.([]interface{}) {
											tagFiltersArrayItem := k.GetWith(i)
											tagFiltersDDArray = append(tagFiltersDDArray, tagFiltersArrayItem.(string))
										}
										ddFormulaAndFunctionProcessQueryDefinition.SetTagFilters(tagFiltersDDArray)
									}
									k.Remove("tag_filters")

									if v, ok := k.GetOkWith("text_filter"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetTextFilter(v.(string))
									}
									queriesDDArrayItem.FormulaAndFunctionProcessQueryDefinition = ddFormulaAndFunctionProcessQueryDefinition
								}
								k.Remove("formula_and_function_process_query_definition.0")

								if queriesDDArrayItem.GetActualInstance() == nil {
									return nil, fmt.Errorf("failed to find valid definition in formula_and_function_query_definition configuration")
								}
								queriesDDArray = append(queriesDDArray, *queriesDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItem.SetQueries(queriesDDArray)
						}
						k.Remove("query")

						if v, ok := k.GetOkWith("response_format"); ok {
							requestsDDArrayItem.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat(v.(string)))
						}

						// handle rum_query, which is a nested model
						k.Add("rum_query.0")

						requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemRumQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
						k.Remove("rum_query.0")
						requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

						// handle security_query, which is a nested model
						k.Add("security_query.0")

						requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
						k.Remove("security_query.0")
						requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

						// handle style, which is a nested model
						k.Add("style.0")

						requestsDDArrayItemStyle := datadogV1.NewWidgetRequestStyleWithDefaults()

						if v, ok := k.GetOkWith("line_type"); ok {
							requestsDDArrayItemStyle.SetLineType(datadogV1.WidgetLineType(v.(string)))
						}

						if v, ok := k.GetOkWith("line_width"); ok {
							requestsDDArrayItemStyle.SetLineWidth(datadogV1.WidgetLineWidth(v.(string)))
						}

						if v, ok := k.GetOkWith("palette"); ok {
							requestsDDArrayItemStyle.SetPalette(v.(string))
						}
						k.Remove("style.0")
						requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
						requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
						k.Remove(i)
					}
					ddNotebookTimeseriesCellAttributesDefinition.SetRequests(requestsDDArray)
				}
				k.Remove("request")

				// handle right_yaxis, which is a nested model
				k.Add("right_yaxis.0")

				ddNotebookTimeseriesCellAttributesDefinitionRightYaxis := datadogV1.NewWidgetAxisWithDefaults()

				if v, ok := k.GetOkWith("include_zero"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetIncludeZero(v.(bool))
				}

				if v, ok := k.GetOkWith("label"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetLabel(v.(string))
				}

				if v, ok := k.GetOkWith("max"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetMax(v.(string))
				}

				if v, ok := k.GetOkWith("min"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetMin(v.(string))
				}

				if v, ok := k.GetOkWith("scale"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetScale(v.(string))
				}
				k.Remove("right_yaxis.0")
				ddNotebookTimeseriesCellAttributesDefinition.SetRightYaxis(*ddNotebookTimeseriesCellAttributesDefinitionRightYaxis)

				if v, ok := k.GetOkWith("show_legend"); ok {
					ddNotebookTimeseriesCellAttributesDefinition.SetShowLegend(v.(bool))
				}

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookTimeseriesCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

				if v, ok := k.GetOkWith("live_span"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
				}
				k.Remove("time.0")
				ddNotebookTimeseriesCellAttributesDefinition.SetTime(*ddNotebookTimeseriesCellAttributesDefinitionTime)

				if v, ok := k.GetOkWith("title"); ok {
					ddNotebookTimeseriesCellAttributesDefinition.SetTitle(v.(string))
				}

				if v, ok := k.GetOkWith("title_align"); ok {
					ddNotebookTimeseriesCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
				}

				if v, ok := k.GetOkWith("title_size"); ok {
					ddNotebookTimeseriesCellAttributesDefinition.SetTitleSize(v.(string))
				}

				if v, ok := k.GetOkWith("type"); ok {
					ddNotebookTimeseriesCellAttributesDefinition.SetType(datadogV1.TimeseriesWidgetDefinitionType(v.(string)))
				}

				// handle yaxis, which is a nested model
				k.Add("yaxis.0")

				ddNotebookTimeseriesCellAttributesDefinitionYaxis := datadogV1.NewWidgetAxisWithDefaults()

				if v, ok := k.GetOkWith("include_zero"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetIncludeZero(v.(bool))
				}

				if v, ok := k.GetOkWith("label"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetLabel(v.(string))
				}

				if v, ok := k.GetOkWith("max"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetMax(v.(string))
				}

				if v, ok := k.GetOkWith("min"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetMin(v.(string))
				}

				if v, ok := k.GetOkWith("scale"); ok {
					ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetScale(v.(string))
				}
				k.Remove("yaxis.0")
				ddNotebookTimeseriesCellAttributesDefinition.SetYaxis(*ddNotebookTimeseriesCellAttributesDefinitionYaxis)
				k.Remove("definition.0")
				ddNotebookTimeseriesCellAttributes.SetDefinition(*ddNotebookTimeseriesCellAttributesDefinition)

				if v, ok := k.GetOkWith("graph_size"); ok {
					ddNotebookTimeseriesCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
				}

				// handle split_by, which is a nested model
				k.Add("split_by.0")

				ddNotebookTimeseriesCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
				k.Add("keys")
				if keysArray, ok := k.GetOk(); ok {
					keysDDArray := make([]string, 0)
					for i := range keysArray.([]interface{}) {
						keysArrayItem := k.GetWith(i)
						keysDDArray = append(keysDDArray, keysArrayItem.(string))
					}
					ddNotebookTimeseriesCellAttributesSplitBy.SetKeys(keysDDArray)
				}
				k.Remove("keys")
				k.Add("tags")
				if tagsArray, ok := k.GetOk(); ok {
					tagsDDArray := make([]string, 0)
					for i := range tagsArray.([]interface{}) {
						tagsArrayItem := k.GetWith(i)
						tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
					}
					ddNotebookTimeseriesCellAttributesSplitBy.SetTags(tagsDDArray)
				}
				k.Remove("tags")
				k.Remove("split_by.0")
				ddNotebookTimeseriesCellAttributes.SetSplitBy(*ddNotebookTimeseriesCellAttributesSplitBy)

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookTimeseriesCellAttributesTime := &datadogV1.NotebookCellTime{}
				// handle notebook_cell_time, which is a oneOf model
				k.Add("notebook_relative_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					ddNotebookTimeseriesCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
				}
				k.Remove("notebook_relative_time.0")
				k.Add("notebook_absolute_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
					// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

					if v, ok := k.GetOkWith("live"); ok {
						ddNotebookAbsoluteTime.SetLive(v.(bool))
					}
					// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
					ddNotebookTimeseriesCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
				}
				k.Remove("notebook_absolute_time.0")

				if ddNotebookTimeseriesCellAttributesTime.GetActualInstance() == nil {
					return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
				}
				k.Remove("time.0")
				ddNotebookTimeseriesCellAttributes.SetTime(*ddNotebookTimeseriesCellAttributesTime)
				cellsDDArrayItemAttributes.NotebookTimeseriesCellAttributes = ddNotebookTimeseriesCellAttributes
			}
			k.Remove("notebook_timeseries_cell_attributes.0")
			k.Add("notebook_toplist_cell_attributes.0")
			if _, ok := k.GetOk(); ok {

				ddNotebookToplistCellAttributes := datadogV1.NewNotebookToplistCellAttributesWithDefaults()

				// handle definition, which is a nested model
				k.Add("definition.0")

				ddNotebookToplistCellAttributesDefinition := datadogV1.NewToplistWidgetDefinitionWithDefaults()
				k.Add("custom_link")
				if customLinksArray, ok := k.GetOk(); ok {
					customLinksDDArray := make([]datadogV1.WidgetCustomLink, 0)
					for i := range customLinksArray.([]interface{}) {
						k.Add(i)

						customLinksDDArrayItem := datadogV1.NewWidgetCustomLinkWithDefaults()

						if v, ok := k.GetOkWith("is_hidden"); ok {
							customLinksDDArrayItem.SetIsHidden(v.(bool))
						}

						if v, ok := k.GetOkWith("label"); ok {
							customLinksDDArrayItem.SetLabel(v.(string))
						}

						if v, ok := k.GetOkWith("link"); ok {
							customLinksDDArrayItem.SetLink(v.(string))
						}

						if v, ok := k.GetOkWith("override_label"); ok {
							customLinksDDArrayItem.SetOverrideLabel(v.(string))
						}
						customLinksDDArray = append(customLinksDDArray, *customLinksDDArrayItem)
						k.Remove(i)
					}
					ddNotebookToplistCellAttributesDefinition.SetCustomLinks(customLinksDDArray)
				}
				k.Remove("custom_link")
				k.Add("request")
				if requestsArray, ok := k.GetOk(); ok {
					requestsDDArray := make([]datadogV1.ToplistWidgetRequest, 0)
					for i := range requestsArray.([]interface{}) {
						k.Add(i)

						requestsDDArrayItem := datadogV1.NewToplistWidgetRequestWithDefaults()

						// handle apm_query, which is a nested model
						k.Add("apm_query.0")

						requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemApmQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
						k.Remove("apm_query.0")
						requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)
						k.Add("conditional_format")
						if conditionalFormatsArray, ok := k.GetOk(); ok {
							conditionalFormatsDDArray := make([]datadogV1.WidgetConditionalFormat, 0)
							for i := range conditionalFormatsArray.([]interface{}) {
								k.Add(i)

								conditionalFormatsDDArrayItem := datadogV1.NewWidgetConditionalFormatWithDefaults()

								if v, ok := k.GetOkWith("comparator"); ok {
									conditionalFormatsDDArrayItem.SetComparator(datadogV1.WidgetComparator(v.(string)))
								}

								if v, ok := k.GetOkWith("custom_bg_color"); ok {
									conditionalFormatsDDArrayItem.SetCustomBgColor(v.(string))
								}

								if v, ok := k.GetOkWith("custom_fg_color"); ok {
									conditionalFormatsDDArrayItem.SetCustomFgColor(v.(string))
								}

								if v, ok := k.GetOkWith("hide_value"); ok {
									conditionalFormatsDDArrayItem.SetHideValue(v.(bool))
								}

								if v, ok := k.GetOkWith("image_url"); ok {
									conditionalFormatsDDArrayItem.SetImageUrl(v.(string))
								}

								if v, ok := k.GetOkWith("metric"); ok {
									conditionalFormatsDDArrayItem.SetMetric(v.(string))
								}

								if v, ok := k.GetOkWith("palette"); ok {
									conditionalFormatsDDArrayItem.SetPalette(datadogV1.WidgetPalette(v.(string)))
								}

								if v, ok := k.GetOkWith("timeframe"); ok {
									conditionalFormatsDDArrayItem.SetTimeframe(v.(string))
								}

								if v, ok := k.GetOkWith("value"); ok {
									conditionalFormatsDDArrayItem.SetValue(v.(float64))
								}
								conditionalFormatsDDArray = append(conditionalFormatsDDArray, *conditionalFormatsDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItem.SetConditionalFormats(conditionalFormatsDDArray)
						}
						k.Remove("conditional_format")

						// handle event_query, which is a nested model
						k.Add("event_query.0")

						requestsDDArrayItemEventQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemEventQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemEventQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemEventQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemEventQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemEventQuery.SetCompute(*requestsDDArrayItemEventQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemEventQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemEventQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemEventQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemEventQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemEventQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemEventQuery.SetSearch(*requestsDDArrayItemEventQuerySearch)
						k.Remove("event_query.0")
						requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)
						k.Add("formula")
						if formulasArray, ok := k.GetOk(); ok {
							formulasDDArray := make([]datadogV1.WidgetFormula, 0)
							for i := range formulasArray.([]interface{}) {
								k.Add(i)

								formulasDDArrayItem := datadogV1.NewWidgetFormulaWithDefaults()

								if v, ok := k.GetOkWith("alias"); ok {
									formulasDDArrayItem.SetAlias(v.(string))
								}

								if v, ok := k.GetOkWith("formula"); ok {
									formulasDDArrayItem.SetFormula(v.(string))
								}

								// handle limit, which is a nested model
								k.Add("limit.0")

								formulasDDArrayItemLimit := datadogV1.NewWidgetFormulaLimitWithDefaults()

								if v, ok := k.GetOkWith("count"); ok {
									formulasDDArrayItemLimit.SetCount(int64(v.(int)))
								}

								if v, ok := k.GetOkWith("order"); ok {
									formulasDDArrayItemLimit.SetOrder(datadogV1.QuerySortOrder(v.(string)))
								}
								k.Remove("limit.0")
								formulasDDArrayItem.SetLimit(*formulasDDArrayItemLimit)
								formulasDDArray = append(formulasDDArray, *formulasDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItem.SetFormulas(formulasDDArray)
						}
						k.Remove("formula")

						// handle log_query, which is a nested model
						k.Add("log_query.0")

						requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemLogQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
						k.Remove("log_query.0")
						requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)

						// handle network_query, which is a nested model
						k.Add("network_query.0")

						requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
						k.Remove("network_query.0")
						requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

						// handle process_query, which is a nested model
						k.Add("process_query.0")

						requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
						k.Add("filter_by")
						if filterByArray, ok := k.GetOk(); ok {
							filterByDDArray := make([]string, 0)
							for i := range filterByArray.([]interface{}) {
								filterByArrayItem := k.GetWith(i)
								filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
							}
							requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
						}
						k.Remove("filter_by")

						if v, ok := k.GetOkWith("limit"); ok {
							requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
						}

						if v, ok := k.GetOkWith("metric"); ok {
							requestsDDArrayItemProcessQuery.SetMetric(v.(string))
						}

						if v, ok := k.GetOkWith("search_by"); ok {
							requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
						}
						k.Remove("process_query.0")
						requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

						// handle profile_metrics_query, which is a nested model
						k.Add("profile_metrics_query.0")

						requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
						k.Remove("profile_metrics_query.0")
						requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

						if v, ok := k.GetOkWith("q"); ok {
							requestsDDArrayItem.SetQ(v.(string))
						}
						k.Add("query")
						if queriesArray, ok := k.GetOk(); ok {
							queriesDDArray := make([]datadogV1.FormulaAndFunctionQueryDefinition, 0)
							for i := range queriesArray.([]interface{}) {
								k.Add(i)

								queriesDDArrayItem := &datadogV1.FormulaAndFunctionQueryDefinition{}
								// handle formula_and_function_query_definition, which is a oneOf model
								k.Add("formula_and_function_metric_query_definition.0")
								if _, ok := k.GetOk(); ok {

									ddFormulaAndFunctionMetricQueryDefinition := datadogV1.NewFormulaAndFunctionMetricQueryDefinitionWithDefaults()

									if v, ok := k.GetOkWith("aggregator"); ok {
										ddFormulaAndFunctionMetricQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
									}

									if v, ok := k.GetOkWith("data_source"); ok {
										ddFormulaAndFunctionMetricQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionMetricDataSource(v.(string)))
									}

									if v, ok := k.GetOkWith("name"); ok {
										ddFormulaAndFunctionMetricQueryDefinition.SetName(v.(string))
									}

									if v, ok := k.GetOkWith("query"); ok {
										ddFormulaAndFunctionMetricQueryDefinition.SetQuery(v.(string))
									}
									queriesDDArrayItem.FormulaAndFunctionMetricQueryDefinition = ddFormulaAndFunctionMetricQueryDefinition
								}
								k.Remove("formula_and_function_metric_query_definition.0")
								k.Add("formula_and_function_event_query_definition.0")
								if _, ok := k.GetOk(); ok {

									ddFormulaAndFunctionEventQueryDefinition := datadogV1.NewFormulaAndFunctionEventQueryDefinitionWithDefaults()

									// handle compute, which is a nested model
									k.Add("compute.0")

									ddFormulaAndFunctionEventQueryDefinitionCompute := datadogV1.NewFormulaAndFunctionEventQueryDefinitionComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										ddFormulaAndFunctionEventQueryDefinitionCompute.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										ddFormulaAndFunctionEventQueryDefinitionCompute.SetInterval(int64(v.(int)))
									}

									if v, ok := k.GetOkWith("metric"); ok {
										ddFormulaAndFunctionEventQueryDefinitionCompute.SetMetric(v.(string))
									}
									k.Remove("compute.0")
									ddFormulaAndFunctionEventQueryDefinition.SetCompute(*ddFormulaAndFunctionEventQueryDefinitionCompute)

									if v, ok := k.GetOkWith("data_source"); ok {
										ddFormulaAndFunctionEventQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionEventsDataSource(v.(string)))
									}
									k.Add("group_by")
									if groupByArray, ok := k.GetOk(); ok {
										groupByDDArray := make([]datadogV1.FormulaAndFunctionEventQueryGroupBy, 0)
										for i := range groupByArray.([]interface{}) {
											k.Add(i)

											groupByDDArrayItem := datadogV1.NewFormulaAndFunctionEventQueryGroupByWithDefaults()

											if v, ok := k.GetOkWith("facet"); ok {
												groupByDDArrayItem.SetFacet(v.(string))
											}

											if v, ok := k.GetOkWith("limit"); ok {
												groupByDDArrayItem.SetLimit(int64(v.(int)))
											}

											// handle sort, which is a nested model
											k.Add("sort.0")

											groupByDDArrayItemSort := datadogV1.NewFormulaAndFunctionEventQueryGroupBySortWithDefaults()

											if v, ok := k.GetOkWith("aggregation"); ok {
												groupByDDArrayItemSort.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
											}

											if v, ok := k.GetOkWith("metric"); ok {
												groupByDDArrayItemSort.SetMetric(v.(string))
											}

											if v, ok := k.GetOkWith("order"); ok {
												groupByDDArrayItemSort.SetOrder(datadogV1.QuerySortOrder(v.(string)))
											}
											k.Remove("sort.0")
											groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
											groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
											k.Remove(i)
										}
										ddFormulaAndFunctionEventQueryDefinition.SetGroupBy(groupByDDArray)
									}
									k.Remove("group_by")
									k.Add("indexes")
									if indexesArray, ok := k.GetOk(); ok {
										indexesDDArray := make([]string, 0)
										for i := range indexesArray.([]interface{}) {
											indexesArrayItem := k.GetWith(i)
											indexesDDArray = append(indexesDDArray, indexesArrayItem.(string))
										}
										ddFormulaAndFunctionEventQueryDefinition.SetIndexes(indexesDDArray)
									}
									k.Remove("indexes")

									if v, ok := k.GetOkWith("name"); ok {
										ddFormulaAndFunctionEventQueryDefinition.SetName(v.(string))
									}

									// handle search, which is a nested model
									k.Add("search.0")

									ddFormulaAndFunctionEventQueryDefinitionSearch := datadogV1.NewFormulaAndFunctionEventQueryDefinitionSearchWithDefaults()

									if v, ok := k.GetOkWith("query"); ok {
										ddFormulaAndFunctionEventQueryDefinitionSearch.SetQuery(v.(string))
									}
									k.Remove("search.0")
									ddFormulaAndFunctionEventQueryDefinition.SetSearch(*ddFormulaAndFunctionEventQueryDefinitionSearch)
									queriesDDArrayItem.FormulaAndFunctionEventQueryDefinition = ddFormulaAndFunctionEventQueryDefinition
								}
								k.Remove("formula_and_function_event_query_definition.0")
								k.Add("formula_and_function_process_query_definition.0")
								if _, ok := k.GetOk(); ok {

									ddFormulaAndFunctionProcessQueryDefinition := datadogV1.NewFormulaAndFunctionProcessQueryDefinitionWithDefaults()

									if v, ok := k.GetOkWith("aggregator"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
									}

									if v, ok := k.GetOkWith("data_source"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionProcessQueryDataSource(v.(string)))
									}

									if v, ok := k.GetOkWith("is_normalized_cpu"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetIsNormalizedCpu(v.(bool))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetLimit(int64(v.(int)))
									}

									if v, ok := k.GetOkWith("metric"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetMetric(v.(string))
									}

									if v, ok := k.GetOkWith("name"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetName(v.(string))
									}

									if v, ok := k.GetOkWith("sort"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetSort(datadogV1.QuerySortOrder(v.(string)))
									}
									k.Add("tag_filters")
									if tagFiltersArray, ok := k.GetOk(); ok {
										tagFiltersDDArray := make([]string, 0)
										for i := range tagFiltersArray.([]interface{}) {
											tagFiltersArrayItem := k.GetWith(i)
											tagFiltersDDArray = append(tagFiltersDDArray, tagFiltersArrayItem.(string))
										}
										ddFormulaAndFunctionProcessQueryDefinition.SetTagFilters(tagFiltersDDArray)
									}
									k.Remove("tag_filters")

									if v, ok := k.GetOkWith("text_filter"); ok {
										ddFormulaAndFunctionProcessQueryDefinition.SetTextFilter(v.(string))
									}
									queriesDDArrayItem.FormulaAndFunctionProcessQueryDefinition = ddFormulaAndFunctionProcessQueryDefinition
								}
								k.Remove("formula_and_function_process_query_definition.0")

								if queriesDDArrayItem.GetActualInstance() == nil {
									return nil, fmt.Errorf("failed to find valid definition in formula_and_function_query_definition configuration")
								}
								queriesDDArray = append(queriesDDArray, *queriesDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItem.SetQueries(queriesDDArray)
						}
						k.Remove("query")

						if v, ok := k.GetOkWith("response_format"); ok {
							requestsDDArrayItem.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat(v.(string)))
						}

						// handle rum_query, which is a nested model
						k.Add("rum_query.0")

						requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemRumQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
						k.Remove("rum_query.0")
						requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

						// handle security_query, which is a nested model
						k.Add("security_query.0")

						requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
						k.Remove("security_query.0")
						requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

						// handle style, which is a nested model
						k.Add("style.0")

						requestsDDArrayItemStyle := datadogV1.NewWidgetRequestStyleWithDefaults()

						if v, ok := k.GetOkWith("line_type"); ok {
							requestsDDArrayItemStyle.SetLineType(datadogV1.WidgetLineType(v.(string)))
						}

						if v, ok := k.GetOkWith("line_width"); ok {
							requestsDDArrayItemStyle.SetLineWidth(datadogV1.WidgetLineWidth(v.(string)))
						}

						if v, ok := k.GetOkWith("palette"); ok {
							requestsDDArrayItemStyle.SetPalette(v.(string))
						}
						k.Remove("style.0")
						requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
						requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
						k.Remove(i)
					}
					ddNotebookToplistCellAttributesDefinition.SetRequests(requestsDDArray)
				}
				k.Remove("request")

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookToplistCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

				if v, ok := k.GetOkWith("live_span"); ok {
					ddNotebookToplistCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
				}
				k.Remove("time.0")
				ddNotebookToplistCellAttributesDefinition.SetTime(*ddNotebookToplistCellAttributesDefinitionTime)

				if v, ok := k.GetOkWith("title"); ok {
					ddNotebookToplistCellAttributesDefinition.SetTitle(v.(string))
				}

				if v, ok := k.GetOkWith("title_align"); ok {
					ddNotebookToplistCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
				}

				if v, ok := k.GetOkWith("title_size"); ok {
					ddNotebookToplistCellAttributesDefinition.SetTitleSize(v.(string))
				}

				if v, ok := k.GetOkWith("type"); ok {
					ddNotebookToplistCellAttributesDefinition.SetType(datadogV1.ToplistWidgetDefinitionType(v.(string)))
				}
				k.Remove("definition.0")
				ddNotebookToplistCellAttributes.SetDefinition(*ddNotebookToplistCellAttributesDefinition)

				if v, ok := k.GetOkWith("graph_size"); ok {
					ddNotebookToplistCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
				}

				// handle split_by, which is a nested model
				k.Add("split_by.0")

				ddNotebookToplistCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
				k.Add("keys")
				if keysArray, ok := k.GetOk(); ok {
					keysDDArray := make([]string, 0)
					for i := range keysArray.([]interface{}) {
						keysArrayItem := k.GetWith(i)
						keysDDArray = append(keysDDArray, keysArrayItem.(string))
					}
					ddNotebookToplistCellAttributesSplitBy.SetKeys(keysDDArray)
				}
				k.Remove("keys")
				k.Add("tags")
				if tagsArray, ok := k.GetOk(); ok {
					tagsDDArray := make([]string, 0)
					for i := range tagsArray.([]interface{}) {
						tagsArrayItem := k.GetWith(i)
						tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
					}
					ddNotebookToplistCellAttributesSplitBy.SetTags(tagsDDArray)
				}
				k.Remove("tags")
				k.Remove("split_by.0")
				ddNotebookToplistCellAttributes.SetSplitBy(*ddNotebookToplistCellAttributesSplitBy)

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookToplistCellAttributesTime := &datadogV1.NotebookCellTime{}
				// handle notebook_cell_time, which is a oneOf model
				k.Add("notebook_relative_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					ddNotebookToplistCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
				}
				k.Remove("notebook_relative_time.0")
				k.Add("notebook_absolute_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
					// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

					if v, ok := k.GetOkWith("live"); ok {
						ddNotebookAbsoluteTime.SetLive(v.(bool))
					}
					// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
					ddNotebookToplistCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
				}
				k.Remove("notebook_absolute_time.0")

				if ddNotebookToplistCellAttributesTime.GetActualInstance() == nil {
					return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
				}
				k.Remove("time.0")
				ddNotebookToplistCellAttributes.SetTime(*ddNotebookToplistCellAttributesTime)
				cellsDDArrayItemAttributes.NotebookToplistCellAttributes = ddNotebookToplistCellAttributes
			}
			k.Remove("notebook_toplist_cell_attributes.0")
			k.Add("notebook_heat_map_cell_attributes.0")
			if _, ok := k.GetOk(); ok {

				ddNotebookHeatMapCellAttributes := datadogV1.NewNotebookHeatMapCellAttributesWithDefaults()

				// handle definition, which is a nested model
				k.Add("definition.0")

				ddNotebookHeatMapCellAttributesDefinition := datadogV1.NewHeatMapWidgetDefinitionWithDefaults()
				k.Add("custom_link")
				if customLinksArray, ok := k.GetOk(); ok {
					customLinksDDArray := make([]datadogV1.WidgetCustomLink, 0)
					for i := range customLinksArray.([]interface{}) {
						k.Add(i)

						customLinksDDArrayItem := datadogV1.NewWidgetCustomLinkWithDefaults()

						if v, ok := k.GetOkWith("is_hidden"); ok {
							customLinksDDArrayItem.SetIsHidden(v.(bool))
						}

						if v, ok := k.GetOkWith("label"); ok {
							customLinksDDArrayItem.SetLabel(v.(string))
						}

						if v, ok := k.GetOkWith("link"); ok {
							customLinksDDArrayItem.SetLink(v.(string))
						}

						if v, ok := k.GetOkWith("override_label"); ok {
							customLinksDDArrayItem.SetOverrideLabel(v.(string))
						}
						customLinksDDArray = append(customLinksDDArray, *customLinksDDArrayItem)
						k.Remove(i)
					}
					ddNotebookHeatMapCellAttributesDefinition.SetCustomLinks(customLinksDDArray)
				}
				k.Remove("custom_link")
				k.Add("event")
				if eventsArray, ok := k.GetOk(); ok {
					eventsDDArray := make([]datadogV1.WidgetEvent, 0)
					for i := range eventsArray.([]interface{}) {
						k.Add(i)

						eventsDDArrayItem := datadogV1.NewWidgetEventWithDefaults()

						if v, ok := k.GetOkWith("q"); ok {
							eventsDDArrayItem.SetQ(v.(string))
						}

						if v, ok := k.GetOkWith("tags_execution"); ok {
							eventsDDArrayItem.SetTagsExecution(v.(string))
						}
						eventsDDArray = append(eventsDDArray, *eventsDDArrayItem)
						k.Remove(i)
					}
					ddNotebookHeatMapCellAttributesDefinition.SetEvents(eventsDDArray)
				}
				k.Remove("event")

				if v, ok := k.GetOkWith("legend_size"); ok {
					ddNotebookHeatMapCellAttributesDefinition.SetLegendSize(v.(string))
				}
				k.Add("request")
				if requestsArray, ok := k.GetOk(); ok {
					requestsDDArray := make([]datadogV1.HeatMapWidgetRequest, 0)
					for i := range requestsArray.([]interface{}) {
						k.Add(i)

						requestsDDArrayItem := datadogV1.NewHeatMapWidgetRequestWithDefaults()

						// handle apm_query, which is a nested model
						k.Add("apm_query.0")

						requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemApmQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
						k.Remove("apm_query.0")
						requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)

						// handle event_query, which is a nested model
						k.Add("event_query.0")

						requestsDDArrayItemEventQuery := datadogV1.NewEventQueryDefinitionWithDefaults()

						if v, ok := k.GetOkWith("search"); ok {
							requestsDDArrayItemEventQuery.SetSearch(v.(string))
						}

						if v, ok := k.GetOkWith("tags_execution"); ok {
							requestsDDArrayItemEventQuery.SetTagsExecution(v.(string))
						}
						k.Remove("event_query.0")
						requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)

						// handle log_query, which is a nested model
						k.Add("log_query.0")

						requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemLogQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
						k.Remove("log_query.0")
						requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)

						// handle network_query, which is a nested model
						k.Add("network_query.0")

						requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
						k.Remove("network_query.0")
						requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

						// handle process_query, which is a nested model
						k.Add("process_query.0")

						requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
						k.Add("filter_by")
						if filterByArray, ok := k.GetOk(); ok {
							filterByDDArray := make([]string, 0)
							for i := range filterByArray.([]interface{}) {
								filterByArrayItem := k.GetWith(i)
								filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
							}
							requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
						}
						k.Remove("filter_by")

						if v, ok := k.GetOkWith("limit"); ok {
							requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
						}

						if v, ok := k.GetOkWith("metric"); ok {
							requestsDDArrayItemProcessQuery.SetMetric(v.(string))
						}

						if v, ok := k.GetOkWith("search_by"); ok {
							requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
						}
						k.Remove("process_query.0")
						requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

						// handle profile_metrics_query, which is a nested model
						k.Add("profile_metrics_query.0")

						requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
						k.Remove("profile_metrics_query.0")
						requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

						if v, ok := k.GetOkWith("q"); ok {
							requestsDDArrayItem.SetQ(v.(string))
						}

						// handle rum_query, which is a nested model
						k.Add("rum_query.0")

						requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemRumQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
						k.Remove("rum_query.0")
						requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

						// handle security_query, which is a nested model
						k.Add("security_query.0")

						requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
						k.Remove("security_query.0")
						requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

						// handle style, which is a nested model
						k.Add("style.0")

						requestsDDArrayItemStyle := datadogV1.NewWidgetStyleWithDefaults()

						if v, ok := k.GetOkWith("palette"); ok {
							requestsDDArrayItemStyle.SetPalette(v.(string))
						}
						k.Remove("style.0")
						requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
						requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
						k.Remove(i)
					}
					ddNotebookHeatMapCellAttributesDefinition.SetRequests(requestsDDArray)
				}
				k.Remove("request")

				if v, ok := k.GetOkWith("show_legend"); ok {
					ddNotebookHeatMapCellAttributesDefinition.SetShowLegend(v.(bool))
				}

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookHeatMapCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

				if v, ok := k.GetOkWith("live_span"); ok {
					ddNotebookHeatMapCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
				}
				k.Remove("time.0")
				ddNotebookHeatMapCellAttributesDefinition.SetTime(*ddNotebookHeatMapCellAttributesDefinitionTime)

				if v, ok := k.GetOkWith("title"); ok {
					ddNotebookHeatMapCellAttributesDefinition.SetTitle(v.(string))
				}

				if v, ok := k.GetOkWith("title_align"); ok {
					ddNotebookHeatMapCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
				}

				if v, ok := k.GetOkWith("title_size"); ok {
					ddNotebookHeatMapCellAttributesDefinition.SetTitleSize(v.(string))
				}

				if v, ok := k.GetOkWith("type"); ok {
					ddNotebookHeatMapCellAttributesDefinition.SetType(datadogV1.HeatMapWidgetDefinitionType(v.(string)))
				}

				// handle yaxis, which is a nested model
				k.Add("yaxis.0")

				ddNotebookHeatMapCellAttributesDefinitionYaxis := datadogV1.NewWidgetAxisWithDefaults()

				if v, ok := k.GetOkWith("include_zero"); ok {
					ddNotebookHeatMapCellAttributesDefinitionYaxis.SetIncludeZero(v.(bool))
				}

				if v, ok := k.GetOkWith("label"); ok {
					ddNotebookHeatMapCellAttributesDefinitionYaxis.SetLabel(v.(string))
				}

				if v, ok := k.GetOkWith("max"); ok {
					ddNotebookHeatMapCellAttributesDefinitionYaxis.SetMax(v.(string))
				}

				if v, ok := k.GetOkWith("min"); ok {
					ddNotebookHeatMapCellAttributesDefinitionYaxis.SetMin(v.(string))
				}

				if v, ok := k.GetOkWith("scale"); ok {
					ddNotebookHeatMapCellAttributesDefinitionYaxis.SetScale(v.(string))
				}
				k.Remove("yaxis.0")
				ddNotebookHeatMapCellAttributesDefinition.SetYaxis(*ddNotebookHeatMapCellAttributesDefinitionYaxis)
				k.Remove("definition.0")
				ddNotebookHeatMapCellAttributes.SetDefinition(*ddNotebookHeatMapCellAttributesDefinition)

				if v, ok := k.GetOkWith("graph_size"); ok {
					ddNotebookHeatMapCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
				}

				// handle split_by, which is a nested model
				k.Add("split_by.0")

				ddNotebookHeatMapCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
				k.Add("keys")
				if keysArray, ok := k.GetOk(); ok {
					keysDDArray := make([]string, 0)
					for i := range keysArray.([]interface{}) {
						keysArrayItem := k.GetWith(i)
						keysDDArray = append(keysDDArray, keysArrayItem.(string))
					}
					ddNotebookHeatMapCellAttributesSplitBy.SetKeys(keysDDArray)
				}
				k.Remove("keys")
				k.Add("tags")
				if tagsArray, ok := k.GetOk(); ok {
					tagsDDArray := make([]string, 0)
					for i := range tagsArray.([]interface{}) {
						tagsArrayItem := k.GetWith(i)
						tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
					}
					ddNotebookHeatMapCellAttributesSplitBy.SetTags(tagsDDArray)
				}
				k.Remove("tags")
				k.Remove("split_by.0")
				ddNotebookHeatMapCellAttributes.SetSplitBy(*ddNotebookHeatMapCellAttributesSplitBy)

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookHeatMapCellAttributesTime := &datadogV1.NotebookCellTime{}
				// handle notebook_cell_time, which is a oneOf model
				k.Add("notebook_relative_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					ddNotebookHeatMapCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
				}
				k.Remove("notebook_relative_time.0")
				k.Add("notebook_absolute_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
					// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

					if v, ok := k.GetOkWith("live"); ok {
						ddNotebookAbsoluteTime.SetLive(v.(bool))
					}
					// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
					ddNotebookHeatMapCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
				}
				k.Remove("notebook_absolute_time.0")

				if ddNotebookHeatMapCellAttributesTime.GetActualInstance() == nil {
					return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
				}
				k.Remove("time.0")
				ddNotebookHeatMapCellAttributes.SetTime(*ddNotebookHeatMapCellAttributesTime)
				cellsDDArrayItemAttributes.NotebookHeatMapCellAttributes = ddNotebookHeatMapCellAttributes
			}
			k.Remove("notebook_heat_map_cell_attributes.0")
			k.Add("notebook_distribution_cell_attributes.0")
			if _, ok := k.GetOk(); ok {

				ddNotebookDistributionCellAttributes := datadogV1.NewNotebookDistributionCellAttributesWithDefaults()

				// handle definition, which is a nested model
				k.Add("definition.0")

				ddNotebookDistributionCellAttributesDefinition := datadogV1.NewDistributionWidgetDefinitionWithDefaults()

				if v, ok := k.GetOkWith("legend_size"); ok {
					ddNotebookDistributionCellAttributesDefinition.SetLegendSize(v.(string))
				}
				k.Add("marker")
				if markersArray, ok := k.GetOk(); ok {
					markersDDArray := make([]datadogV1.WidgetMarker, 0)
					for i := range markersArray.([]interface{}) {
						k.Add(i)

						markersDDArrayItem := datadogV1.NewWidgetMarkerWithDefaults()

						if v, ok := k.GetOkWith("display_type"); ok {
							markersDDArrayItem.SetDisplayType(v.(string))
						}

						if v, ok := k.GetOkWith("label"); ok {
							markersDDArrayItem.SetLabel(v.(string))
						}

						if v, ok := k.GetOkWith("time"); ok {
							markersDDArrayItem.SetTime(v.(string))
						}

						if v, ok := k.GetOkWith("value"); ok {
							markersDDArrayItem.SetValue(v.(string))
						}
						markersDDArray = append(markersDDArray, *markersDDArrayItem)
						k.Remove(i)
					}
					ddNotebookDistributionCellAttributesDefinition.SetMarkers(markersDDArray)
				}
				k.Remove("marker")
				k.Add("request")
				if requestsArray, ok := k.GetOk(); ok {
					requestsDDArray := make([]datadogV1.DistributionWidgetRequest, 0)
					for i := range requestsArray.([]interface{}) {
						k.Add(i)

						requestsDDArrayItem := datadogV1.NewDistributionWidgetRequestWithDefaults()

						// handle apm_query, which is a nested model
						k.Add("apm_query.0")

						requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemApmQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
						k.Remove("apm_query.0")
						requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)

						// handle event_query, which is a nested model
						k.Add("event_query.0")

						requestsDDArrayItemEventQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemEventQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemEventQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemEventQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemEventQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemEventQuery.SetCompute(*requestsDDArrayItemEventQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemEventQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemEventQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemEventQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemEventQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemEventQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemEventQuery.SetSearch(*requestsDDArrayItemEventQuerySearch)
						k.Remove("event_query.0")
						requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)

						// handle log_query, which is a nested model
						k.Add("log_query.0")

						requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemLogQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
						k.Remove("log_query.0")
						requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)

						// handle network_query, which is a nested model
						k.Add("network_query.0")

						requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
						k.Remove("network_query.0")
						requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

						// handle process_query, which is a nested model
						k.Add("process_query.0")

						requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
						k.Add("filter_by")
						if filterByArray, ok := k.GetOk(); ok {
							filterByDDArray := make([]string, 0)
							for i := range filterByArray.([]interface{}) {
								filterByArrayItem := k.GetWith(i)
								filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
							}
							requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
						}
						k.Remove("filter_by")

						if v, ok := k.GetOkWith("limit"); ok {
							requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
						}

						if v, ok := k.GetOkWith("metric"); ok {
							requestsDDArrayItemProcessQuery.SetMetric(v.(string))
						}

						if v, ok := k.GetOkWith("search_by"); ok {
							requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
						}
						k.Remove("process_query.0")
						requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

						// handle profile_metrics_query, which is a nested model
						k.Add("profile_metrics_query.0")

						requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
						k.Remove("profile_metrics_query.0")
						requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

						if v, ok := k.GetOkWith("q"); ok {
							requestsDDArrayItem.SetQ(v.(string))
						}

						// handle rum_query, which is a nested model
						k.Add("rum_query.0")

						requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemRumQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
						k.Remove("rum_query.0")
						requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

						// handle security_query, which is a nested model
						k.Add("security_query.0")

						requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

						// handle compute, which is a nested model
						k.Add("compute.0")

						requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

						if v, ok := k.GetOkWith("aggregation"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
						}

						if v, ok := k.GetOkWith("facet"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
						}

						if v, ok := k.GetOkWith("interval"); ok {
							requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
						}
						k.Remove("compute.0")
						requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
						k.Add("group_by")
						if groupByArray, ok := k.GetOk(); ok {
							groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
							for i := range groupByArray.([]interface{}) {
								k.Add(i)

								groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("limit"); ok {
									groupByDDArrayItem.SetLimit(int64(v.(int)))
								}

								// handle sort, which is a nested model
								k.Add("sort.0")

								groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									groupByDDArrayItemSort.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									groupByDDArrayItemSort.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("order"); ok {
									groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
								}
								k.Remove("sort.0")
								groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
								groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
						}
						k.Remove("group_by")

						if v, ok := k.GetOkWith("index"); ok {
							requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
						}
						k.Add("multi_compute")
						if multiComputeArray, ok := k.GetOk(); ok {
							multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
							for i := range multiComputeArray.([]interface{}) {
								k.Add(i)

								multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

								if v, ok := k.GetOkWith("aggregation"); ok {
									multiComputeDDArrayItem.SetAggregation(v.(string))
								}

								if v, ok := k.GetOkWith("facet"); ok {
									multiComputeDDArrayItem.SetFacet(v.(string))
								}

								if v, ok := k.GetOkWith("interval"); ok {
									multiComputeDDArrayItem.SetInterval(int64(v.(int)))
								}
								multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
								k.Remove(i)
							}
							requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
						}
						k.Remove("multi_compute")

						// handle search, which is a nested model
						k.Add("search.0")

						requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

						if v, ok := k.GetOkWith("query"); ok {
							requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
						}
						k.Remove("search.0")
						requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
						k.Remove("security_query.0")
						requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

						// handle style, which is a nested model
						k.Add("style.0")

						requestsDDArrayItemStyle := datadogV1.NewWidgetStyleWithDefaults()

						if v, ok := k.GetOkWith("palette"); ok {
							requestsDDArrayItemStyle.SetPalette(v.(string))
						}
						k.Remove("style.0")
						requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
						requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
						k.Remove(i)
					}
					ddNotebookDistributionCellAttributesDefinition.SetRequests(requestsDDArray)
				}
				k.Remove("request")

				if v, ok := k.GetOkWith("show_legend"); ok {
					ddNotebookDistributionCellAttributesDefinition.SetShowLegend(v.(bool))
				}

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookDistributionCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

				if v, ok := k.GetOkWith("live_span"); ok {
					ddNotebookDistributionCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
				}
				k.Remove("time.0")
				ddNotebookDistributionCellAttributesDefinition.SetTime(*ddNotebookDistributionCellAttributesDefinitionTime)

				if v, ok := k.GetOkWith("title"); ok {
					ddNotebookDistributionCellAttributesDefinition.SetTitle(v.(string))
				}

				if v, ok := k.GetOkWith("title_align"); ok {
					ddNotebookDistributionCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
				}

				if v, ok := k.GetOkWith("title_size"); ok {
					ddNotebookDistributionCellAttributesDefinition.SetTitleSize(v.(string))
				}

				if v, ok := k.GetOkWith("type"); ok {
					ddNotebookDistributionCellAttributesDefinition.SetType(datadogV1.DistributionWidgetDefinitionType(v.(string)))
				}

				// handle xaxis, which is a nested model
				k.Add("xaxis.0")

				ddNotebookDistributionCellAttributesDefinitionXaxis := datadogV1.NewDistributionWidgetXAxisWithDefaults()

				if v, ok := k.GetOkWith("include_zero"); ok {
					ddNotebookDistributionCellAttributesDefinitionXaxis.SetIncludeZero(v.(bool))
				}

				if v, ok := k.GetOkWith("max"); ok {
					ddNotebookDistributionCellAttributesDefinitionXaxis.SetMax(v.(string))
				}

				if v, ok := k.GetOkWith("min"); ok {
					ddNotebookDistributionCellAttributesDefinitionXaxis.SetMin(v.(string))
				}

				if v, ok := k.GetOkWith("scale"); ok {
					ddNotebookDistributionCellAttributesDefinitionXaxis.SetScale(v.(string))
				}
				k.Remove("xaxis.0")
				ddNotebookDistributionCellAttributesDefinition.SetXaxis(*ddNotebookDistributionCellAttributesDefinitionXaxis)

				// handle yaxis, which is a nested model
				k.Add("yaxis.0")

				ddNotebookDistributionCellAttributesDefinitionYaxis := datadogV1.NewDistributionWidgetYAxisWithDefaults()

				if v, ok := k.GetOkWith("include_zero"); ok {
					ddNotebookDistributionCellAttributesDefinitionYaxis.SetIncludeZero(v.(bool))
				}

				if v, ok := k.GetOkWith("label"); ok {
					ddNotebookDistributionCellAttributesDefinitionYaxis.SetLabel(v.(string))
				}

				if v, ok := k.GetOkWith("max"); ok {
					ddNotebookDistributionCellAttributesDefinitionYaxis.SetMax(v.(string))
				}

				if v, ok := k.GetOkWith("min"); ok {
					ddNotebookDistributionCellAttributesDefinitionYaxis.SetMin(v.(string))
				}

				if v, ok := k.GetOkWith("scale"); ok {
					ddNotebookDistributionCellAttributesDefinitionYaxis.SetScale(v.(string))
				}
				k.Remove("yaxis.0")
				ddNotebookDistributionCellAttributesDefinition.SetYaxis(*ddNotebookDistributionCellAttributesDefinitionYaxis)
				k.Remove("definition.0")
				ddNotebookDistributionCellAttributes.SetDefinition(*ddNotebookDistributionCellAttributesDefinition)

				if v, ok := k.GetOkWith("graph_size"); ok {
					ddNotebookDistributionCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
				}

				// handle split_by, which is a nested model
				k.Add("split_by.0")

				ddNotebookDistributionCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
				k.Add("keys")
				if keysArray, ok := k.GetOk(); ok {
					keysDDArray := make([]string, 0)
					for i := range keysArray.([]interface{}) {
						keysArrayItem := k.GetWith(i)
						keysDDArray = append(keysDDArray, keysArrayItem.(string))
					}
					ddNotebookDistributionCellAttributesSplitBy.SetKeys(keysDDArray)
				}
				k.Remove("keys")
				k.Add("tags")
				if tagsArray, ok := k.GetOk(); ok {
					tagsDDArray := make([]string, 0)
					for i := range tagsArray.([]interface{}) {
						tagsArrayItem := k.GetWith(i)
						tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
					}
					ddNotebookDistributionCellAttributesSplitBy.SetTags(tagsDDArray)
				}
				k.Remove("tags")
				k.Remove("split_by.0")
				ddNotebookDistributionCellAttributes.SetSplitBy(*ddNotebookDistributionCellAttributesSplitBy)

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookDistributionCellAttributesTime := &datadogV1.NotebookCellTime{}
				// handle notebook_cell_time, which is a oneOf model
				k.Add("notebook_relative_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					ddNotebookDistributionCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
				}
				k.Remove("notebook_relative_time.0")
				k.Add("notebook_absolute_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
					// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

					if v, ok := k.GetOkWith("live"); ok {
						ddNotebookAbsoluteTime.SetLive(v.(bool))
					}
					// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
					ddNotebookDistributionCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
				}
				k.Remove("notebook_absolute_time.0")

				if ddNotebookDistributionCellAttributesTime.GetActualInstance() == nil {
					return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
				}
				k.Remove("time.0")
				ddNotebookDistributionCellAttributes.SetTime(*ddNotebookDistributionCellAttributesTime)
				cellsDDArrayItemAttributes.NotebookDistributionCellAttributes = ddNotebookDistributionCellAttributes
			}
			k.Remove("notebook_distribution_cell_attributes.0")
			k.Add("notebook_log_stream_cell_attributes.0")
			if _, ok := k.GetOk(); ok {

				ddNotebookLogStreamCellAttributes := datadogV1.NewNotebookLogStreamCellAttributesWithDefaults()

				// handle definition, which is a nested model
				k.Add("definition.0")

				ddNotebookLogStreamCellAttributesDefinition := datadogV1.NewLogStreamWidgetDefinitionWithDefaults()
				k.Add("columns")
				if columnsArray, ok := k.GetOk(); ok {
					columnsDDArray := make([]string, 0)
					for i := range columnsArray.([]interface{}) {
						columnsArrayItem := k.GetWith(i)
						columnsDDArray = append(columnsDDArray, columnsArrayItem.(string))
					}
					ddNotebookLogStreamCellAttributesDefinition.SetColumns(columnsDDArray)
				}
				k.Remove("columns")
				k.Add("indexes")
				if indexesArray, ok := k.GetOk(); ok {
					indexesDDArray := make([]string, 0)
					for i := range indexesArray.([]interface{}) {
						indexesArrayItem := k.GetWith(i)
						indexesDDArray = append(indexesDDArray, indexesArrayItem.(string))
					}
					ddNotebookLogStreamCellAttributesDefinition.SetIndexes(indexesDDArray)
				}
				k.Remove("indexes")

				if v, ok := k.GetOkWith("logset"); ok {
					ddNotebookLogStreamCellAttributesDefinition.SetLogset(v.(string))
				}

				if v, ok := k.GetOkWith("message_display"); ok {
					ddNotebookLogStreamCellAttributesDefinition.SetMessageDisplay(datadogV1.WidgetMessageDisplay(v.(string)))
				}

				if v, ok := k.GetOkWith("query"); ok {
					ddNotebookLogStreamCellAttributesDefinition.SetQuery(v.(string))
				}

				if v, ok := k.GetOkWith("show_date_column"); ok {
					ddNotebookLogStreamCellAttributesDefinition.SetShowDateColumn(v.(bool))
				}

				if v, ok := k.GetOkWith("show_message_column"); ok {
					ddNotebookLogStreamCellAttributesDefinition.SetShowMessageColumn(v.(bool))
				}

				// handle sort, which is a nested model
				k.Add("sort.0")

				ddNotebookLogStreamCellAttributesDefinitionSort := datadogV1.NewWidgetFieldSortWithDefaults()

				if v, ok := k.GetOkWith("column"); ok {
					ddNotebookLogStreamCellAttributesDefinitionSort.SetColumn(v.(string))
				}

				if v, ok := k.GetOkWith("order"); ok {
					ddNotebookLogStreamCellAttributesDefinitionSort.SetOrder(datadogV1.WidgetSort(v.(string)))
				}
				k.Remove("sort.0")
				ddNotebookLogStreamCellAttributesDefinition.SetSort(*ddNotebookLogStreamCellAttributesDefinitionSort)

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookLogStreamCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

				if v, ok := k.GetOkWith("live_span"); ok {
					ddNotebookLogStreamCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
				}
				k.Remove("time.0")
				ddNotebookLogStreamCellAttributesDefinition.SetTime(*ddNotebookLogStreamCellAttributesDefinitionTime)

				if v, ok := k.GetOkWith("title"); ok {
					ddNotebookLogStreamCellAttributesDefinition.SetTitle(v.(string))
				}

				if v, ok := k.GetOkWith("title_align"); ok {
					ddNotebookLogStreamCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
				}

				if v, ok := k.GetOkWith("title_size"); ok {
					ddNotebookLogStreamCellAttributesDefinition.SetTitleSize(v.(string))
				}

				if v, ok := k.GetOkWith("type"); ok {
					ddNotebookLogStreamCellAttributesDefinition.SetType(datadogV1.LogStreamWidgetDefinitionType(v.(string)))
				}
				k.Remove("definition.0")
				ddNotebookLogStreamCellAttributes.SetDefinition(*ddNotebookLogStreamCellAttributesDefinition)

				if v, ok := k.GetOkWith("graph_size"); ok {
					ddNotebookLogStreamCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
				}

				// handle time, which is a nested model
				k.Add("time.0")

				ddNotebookLogStreamCellAttributesTime := &datadogV1.NotebookCellTime{}
				// handle notebook_cell_time, which is a oneOf model
				k.Add("notebook_relative_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					ddNotebookLogStreamCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
				}
				k.Remove("notebook_relative_time.0")
				k.Add("notebook_absolute_time.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
					// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

					if v, ok := k.GetOkWith("live"); ok {
						ddNotebookAbsoluteTime.SetLive(v.(bool))
					}
					// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
					ddNotebookLogStreamCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
				}
				k.Remove("notebook_absolute_time.0")

				if ddNotebookLogStreamCellAttributesTime.GetActualInstance() == nil {
					return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
				}
				k.Remove("time.0")
				ddNotebookLogStreamCellAttributes.SetTime(*ddNotebookLogStreamCellAttributesTime)
				cellsDDArrayItemAttributes.NotebookLogStreamCellAttributes = ddNotebookLogStreamCellAttributes
			}
			k.Remove("notebook_log_stream_cell_attributes.0")

			if cellsDDArrayItemAttributes.GetActualInstance() == nil {
				return nil, fmt.Errorf("failed to find valid definition in notebook_cell_create_request_attributes configuration")
			}
			k.Remove("attributes.0")
			cellsDDArrayItem.SetAttributes(*cellsDDArrayItemAttributes)

			if v, ok := k.GetOkWith("type"); ok {
				cellsDDArrayItem.SetType(datadogV1.NotebookCellResourceType(v.(string)))
			}
			cellsDDArray = append(cellsDDArray, *cellsDDArrayItem)
			k.Remove(i)
		}
		result.SetCells(cellsDDArray)
	}
	k.Remove("cell")

	if v, ok := k.GetOkWith("name"); ok {
		result.SetName(v.(string))
	}

	if v, ok := k.GetOkWith("status"); ok {
		result.SetStatus(datadogV1.NotebookStatus(v.(string)))
	}

	// handle time, which is a nested model
	k.Add("time.0")

	resultTime := &datadogV1.NotebookGlobalTime{}
	// handle notebook_global_time, which is a oneOf model
	k.Add("notebook_relative_time.0")
	if _, ok := k.GetOk(); ok {

		ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

		if v, ok := k.GetOkWith("live_span"); ok {
			ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
		}
		resultTime.NotebookRelativeTime = ddNotebookRelativeTime
	}
	k.Remove("notebook_relative_time.0")
	k.Add("notebook_absolute_time.0")
	if _, ok := k.GetOk(); ok {

		ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
		// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

		if v, ok := k.GetOkWith("live"); ok {
			ddNotebookAbsoluteTime.SetLive(v.(bool))
		}
		// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
		resultTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
	}
	k.Remove("notebook_absolute_time.0")

	if resultTime.GetActualInstance() == nil {
		return nil, fmt.Errorf("failed to find valid definition in notebook_global_time configuration")
	}
	k.Remove("time.0")
	result.SetTime(*resultTime)
	return result, nil
}

func resourceDatadogNotebookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	resultNotebookCreateDataAttributes, err := buildDatadogNotebook(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error building Notebook object: %s", err))
	}

	resultNotebookCreateData := datadogV1.NewNotebookCreateDataWithDefaults()
	resultNotebookCreateData.SetAttributes(*resultNotebookCreateDataAttributes)

	ddObject := datadogV1.NewNotebookCreateRequestWithDefaults()
	ddObject.SetData(*resultNotebookCreateData)

	resourceNotebookResponse, _, err := datadogClient.NotebooksApi.CreateNotebook(auth, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, "error creating Notebook")
	}
	// FIXME: no property found that looks like an Id for model Notebook
	// you need to manually add code that would call `d.SetId(<the-actual-id>)` to store
	// the Id in the state properly

	resourceNotebookResponseData := resourceNotebookResponse.GetData()

	resource := resourceNotebookResponseData.GetAttributes()

	return updateNotebookTerraformState(d, resource)
}

func updateNotebookTerraformState(d *schema.ResourceData, resource datadogV1.NotebookResponseDataAttributes) diag.Diagnostics {
	var err error

	if ddAuthor, ok := resource.GetAuthorOk(); ok {
		mapAuthor := map[string]interface{}{}
		if v, ok := ddAuthor.GetCreatedAtOk(); ok {
			mapAuthor["created_at"] = func(t *time.Time) *string {
				if t != nil {
					r := t.Format("2006-01-02T15:04:05.000000-0700")
					return &r
				}
				return nil
			}(v)
		}
		if v, ok := ddAuthor.GetDisabledOk(); ok {
			mapAuthor["disabled"] = *v
		}
		if v, ok := ddAuthor.GetEmailOk(); ok {
			mapAuthor["email"] = *v
		}
		if v, ok := ddAuthor.GetHandleOk(); ok {
			mapAuthor["handle"] = *v
		}
		if v, ok := ddAuthor.GetIconOk(); ok {
			mapAuthor["icon"] = *v
		}
		if v, ok := ddAuthor.GetNameOk(); ok {
			mapAuthor["name"] = *v
		}
		if v, ok := ddAuthor.GetStatusOk(); ok {
			mapAuthor["status"] = *v
		}
		if v, ok := ddAuthor.GetTitleOk(); ok {
			mapAuthor["title"] = *v
		}
		if v, ok := ddAuthor.GetVerifiedOk(); ok {
			mapAuthor["verified"] = *v
		}

		arrayAuthor := []interface{}{mapAuthor}
		err = d.Set("author", arrayAuthor)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if array, ok := resource.GetCellsOk(); ok {
		cellsTFArray := make([]map[string]interface{}, 0)

		for _, arrayItem := range *array {
			cellsTFArrayIntf := map[string]interface{}{}
			if attributesDDModel, ok := arrayItem.GetAttributesOk(); ok {
				attributesMap := map[string]interface{}{}
				if ddNotebookMarkdownCellAttributes := attributesDDModel.NotebookMarkdownCellAttributes; ddNotebookMarkdownCellAttributes != nil {
					mapNotebookMarkdownCellAttributes := map[string]interface{}{}
					if definitionDDModel, ok := ddNotebookMarkdownCellAttributes.GetDefinitionOk(); ok {
						definitionMap := map[string]interface{}{}
						if v, ok := definitionDDModel.GetTextOk(); ok {
							definitionMap["text"] = *v
						}
						if v, ok := definitionDDModel.GetTypeOk(); ok {
							definitionMap["type"] = *v
						}

						mapNotebookMarkdownCellAttributes["definition"] = []map[string]interface{}{definitionMap}
					}

					arrayNotebookMarkdownCellAttributes := []interface{}{mapNotebookMarkdownCellAttributes}
					attributesMap["notebook_markdown_cell_attributes"] = arrayNotebookMarkdownCellAttributes
				}
				if ddNotebookTimeseriesCellAttributes := attributesDDModel.NotebookTimeseriesCellAttributes; ddNotebookTimeseriesCellAttributes != nil {
					mapNotebookTimeseriesCellAttributes := map[string]interface{}{}
					if definitionDDModel, ok := ddNotebookTimeseriesCellAttributes.GetDefinitionOk(); ok {
						definitionMap := map[string]interface{}{}
						if customLinksArray, ok := definitionDDModel.GetCustomLinksOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, customLinksArrayItem := range *customLinksArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if v, ok := customLinksArrayItem.GetIsHiddenOk(); ok {
									definitionMapArrayIntf["is_hidden"] = *v
								}
								if v, ok := customLinksArrayItem.GetLabelOk(); ok {
									definitionMapArrayIntf["label"] = *v
								}
								if v, ok := customLinksArrayItem.GetLinkOk(); ok {
									definitionMapArrayIntf["link"] = *v
								}
								if v, ok := customLinksArrayItem.GetOverrideLabelOk(); ok {
									definitionMapArrayIntf["override_label"] = *v
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["custom_link"] = definitionMapArray
						}
						if eventsArray, ok := definitionDDModel.GetEventsOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, eventsArrayItem := range *eventsArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if v, ok := eventsArrayItem.GetQOk(); ok {
									definitionMapArrayIntf["q"] = *v
								}
								if v, ok := eventsArrayItem.GetTagsExecutionOk(); ok {
									definitionMapArrayIntf["tags_execution"] = *v
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["event"] = definitionMapArray
						}
						if legendColumnsArray, ok := definitionDDModel.GetLegendColumnsOk(); ok {
							definitionMapArray := make([]datadogV1.TimeseriesWidgetLegendColumn, len(*legendColumnsArray))
							for i, item := range *legendColumnsArray {
								definitionMapArray[i] = item
							}

							definitionMap["legend_columns"] = definitionMapArray
						}
						if v, ok := definitionDDModel.GetLegendLayoutOk(); ok {
							definitionMap["legend_layout"] = *v
						}
						if v, ok := definitionDDModel.GetLegendSizeOk(); ok {
							definitionMap["legend_size"] = *v
						}
						if markersArray, ok := definitionDDModel.GetMarkersOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, markersArrayItem := range *markersArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if v, ok := markersArrayItem.GetDisplayTypeOk(); ok {
									definitionMapArrayIntf["display_type"] = *v
								}
								if v, ok := markersArrayItem.GetLabelOk(); ok {
									definitionMapArrayIntf["label"] = *v
								}
								if v, ok := markersArrayItem.GetTimeOk(); ok {
									definitionMapArrayIntf["time"] = *v
								}
								if v, ok := markersArrayItem.GetValueOk(); ok {
									definitionMapArrayIntf["value"] = *v
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["marker"] = definitionMapArray
						}
						if requestsArray, ok := definitionDDModel.GetRequestsOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, requestsArrayItem := range *requestsArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if apmQueryDDModel, ok := requestsArrayItem.GetApmQueryOk(); ok {
									apmQueryMap := map[string]interface{}{}
									if computeDDModel, ok := apmQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										apmQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := apmQueryDDModel.GetGroupByOk(); ok {
										apmQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											apmQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												apmQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												apmQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												apmQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											apmQueryMapArray = append(apmQueryMapArray, apmQueryMapArrayIntf)
										}

										apmQueryMap["group_by"] = apmQueryMapArray
									}
									if v, ok := apmQueryDDModel.GetIndexOk(); ok {
										apmQueryMap["index"] = *v
									}
									if multiComputeArray, ok := apmQueryDDModel.GetMultiComputeOk(); ok {
										apmQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											apmQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												apmQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												apmQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												apmQueryMapArrayIntf["interval"] = *v
											}

											apmQueryMapArray = append(apmQueryMapArray, apmQueryMapArrayIntf)
										}

										apmQueryMap["multi_compute"] = apmQueryMapArray
									}
									if searchDDModel, ok := apmQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										apmQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["apm_query"] = []map[string]interface{}{apmQueryMap}
								}
								if v, ok := requestsArrayItem.GetDisplayTypeOk(); ok {
									definitionMapArrayIntf["display_type"] = *v
								}
								if eventQueryDDModel, ok := requestsArrayItem.GetEventQueryOk(); ok {
									eventQueryMap := map[string]interface{}{}
									if computeDDModel, ok := eventQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										eventQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := eventQueryDDModel.GetGroupByOk(); ok {
										eventQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											eventQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												eventQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												eventQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												eventQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											eventQueryMapArray = append(eventQueryMapArray, eventQueryMapArrayIntf)
										}

										eventQueryMap["group_by"] = eventQueryMapArray
									}
									if v, ok := eventQueryDDModel.GetIndexOk(); ok {
										eventQueryMap["index"] = *v
									}
									if multiComputeArray, ok := eventQueryDDModel.GetMultiComputeOk(); ok {
										eventQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											eventQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												eventQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												eventQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												eventQueryMapArrayIntf["interval"] = *v
											}

											eventQueryMapArray = append(eventQueryMapArray, eventQueryMapArrayIntf)
										}

										eventQueryMap["multi_compute"] = eventQueryMapArray
									}
									if searchDDModel, ok := eventQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										eventQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["event_query"] = []map[string]interface{}{eventQueryMap}
								}
								if formulasArray, ok := requestsArrayItem.GetFormulasOk(); ok {
									definitionMapArrayIntfArray := make([]map[string]interface{}, 0)

									for _, formulasArrayItem := range *formulasArray {
										definitionMapArrayIntfArrayIntf := map[string]interface{}{}
										if v, ok := formulasArrayItem.GetAliasOk(); ok {
											definitionMapArrayIntfArrayIntf["alias"] = *v
										}
										if v, ok := formulasArrayItem.GetFormulaOk(); ok {
											definitionMapArrayIntfArrayIntf["formula"] = *v
										}
										if limitDDModel, ok := formulasArrayItem.GetLimitOk(); ok {
											limitMap := map[string]interface{}{}
											if v, ok := limitDDModel.GetCountOk(); ok {
												limitMap["count"] = *v
											}
											if v, ok := limitDDModel.GetOrderOk(); ok {
												limitMap["order"] = *v
											}

											definitionMapArrayIntfArrayIntf["limit"] = []map[string]interface{}{limitMap}
										}

										definitionMapArrayIntfArray = append(definitionMapArrayIntfArray, definitionMapArrayIntfArrayIntf)
									}

									definitionMapArrayIntf["formula"] = definitionMapArrayIntfArray
								}
								if logQueryDDModel, ok := requestsArrayItem.GetLogQueryOk(); ok {
									logQueryMap := map[string]interface{}{}
									if computeDDModel, ok := logQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										logQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := logQueryDDModel.GetGroupByOk(); ok {
										logQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											logQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												logQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												logQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												logQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											logQueryMapArray = append(logQueryMapArray, logQueryMapArrayIntf)
										}

										logQueryMap["group_by"] = logQueryMapArray
									}
									if v, ok := logQueryDDModel.GetIndexOk(); ok {
										logQueryMap["index"] = *v
									}
									if multiComputeArray, ok := logQueryDDModel.GetMultiComputeOk(); ok {
										logQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											logQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												logQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												logQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												logQueryMapArrayIntf["interval"] = *v
											}

											logQueryMapArray = append(logQueryMapArray, logQueryMapArrayIntf)
										}

										logQueryMap["multi_compute"] = logQueryMapArray
									}
									if searchDDModel, ok := logQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										logQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["log_query"] = []map[string]interface{}{logQueryMap}
								}
								if metadataArray, ok := requestsArrayItem.GetMetadataOk(); ok {
									definitionMapArrayIntfArray := make([]map[string]interface{}, 0)

									for _, metadataArrayItem := range *metadataArray {
										definitionMapArrayIntfArrayIntf := map[string]interface{}{}
										if v, ok := metadataArrayItem.GetAliasNameOk(); ok {
											definitionMapArrayIntfArrayIntf["alias_name"] = *v
										}
										if v, ok := metadataArrayItem.GetExpressionOk(); ok {
											definitionMapArrayIntfArrayIntf["expression"] = *v
										}

										definitionMapArrayIntfArray = append(definitionMapArrayIntfArray, definitionMapArrayIntfArrayIntf)
									}

									definitionMapArrayIntf["metadata"] = definitionMapArrayIntfArray
								}
								if networkQueryDDModel, ok := requestsArrayItem.GetNetworkQueryOk(); ok {
									networkQueryMap := map[string]interface{}{}
									if computeDDModel, ok := networkQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										networkQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := networkQueryDDModel.GetGroupByOk(); ok {
										networkQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											networkQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												networkQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												networkQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												networkQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											networkQueryMapArray = append(networkQueryMapArray, networkQueryMapArrayIntf)
										}

										networkQueryMap["group_by"] = networkQueryMapArray
									}
									if v, ok := networkQueryDDModel.GetIndexOk(); ok {
										networkQueryMap["index"] = *v
									}
									if multiComputeArray, ok := networkQueryDDModel.GetMultiComputeOk(); ok {
										networkQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											networkQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												networkQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												networkQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												networkQueryMapArrayIntf["interval"] = *v
											}

											networkQueryMapArray = append(networkQueryMapArray, networkQueryMapArrayIntf)
										}

										networkQueryMap["multi_compute"] = networkQueryMapArray
									}
									if searchDDModel, ok := networkQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										networkQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["network_query"] = []map[string]interface{}{networkQueryMap}
								}
								if v, ok := requestsArrayItem.GetOnRightYaxisOk(); ok {
									definitionMapArrayIntf["on_right_yaxis"] = *v
								}
								if processQueryDDModel, ok := requestsArrayItem.GetProcessQueryOk(); ok {
									processQueryMap := map[string]interface{}{}
									if filterByArray, ok := processQueryDDModel.GetFilterByOk(); ok {
										processQueryMapArray := make([]string, len(*filterByArray))
										for i, item := range *filterByArray {
											processQueryMapArray[i] = item
										}

										processQueryMap["filter_by"] = processQueryMapArray
									}
									if v, ok := processQueryDDModel.GetLimitOk(); ok {
										processQueryMap["limit"] = *v
									}
									if v, ok := processQueryDDModel.GetMetricOk(); ok {
										processQueryMap["metric"] = *v
									}
									if v, ok := processQueryDDModel.GetSearchByOk(); ok {
										processQueryMap["search_by"] = *v
									}

									definitionMapArrayIntf["process_query"] = []map[string]interface{}{processQueryMap}
								}
								if profileMetricsQueryDDModel, ok := requestsArrayItem.GetProfileMetricsQueryOk(); ok {
									profileMetricsQueryMap := map[string]interface{}{}
									if computeDDModel, ok := profileMetricsQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										profileMetricsQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := profileMetricsQueryDDModel.GetGroupByOk(); ok {
										profileMetricsQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											profileMetricsQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												profileMetricsQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												profileMetricsQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												profileMetricsQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											profileMetricsQueryMapArray = append(profileMetricsQueryMapArray, profileMetricsQueryMapArrayIntf)
										}

										profileMetricsQueryMap["group_by"] = profileMetricsQueryMapArray
									}
									if v, ok := profileMetricsQueryDDModel.GetIndexOk(); ok {
										profileMetricsQueryMap["index"] = *v
									}
									if multiComputeArray, ok := profileMetricsQueryDDModel.GetMultiComputeOk(); ok {
										profileMetricsQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											profileMetricsQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												profileMetricsQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												profileMetricsQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												profileMetricsQueryMapArrayIntf["interval"] = *v
											}

											profileMetricsQueryMapArray = append(profileMetricsQueryMapArray, profileMetricsQueryMapArrayIntf)
										}

										profileMetricsQueryMap["multi_compute"] = profileMetricsQueryMapArray
									}
									if searchDDModel, ok := profileMetricsQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										profileMetricsQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["profile_metrics_query"] = []map[string]interface{}{profileMetricsQueryMap}
								}
								if v, ok := requestsArrayItem.GetQOk(); ok {
									definitionMapArrayIntf["q"] = *v
								}
								if queriesArray, ok := requestsArrayItem.GetQueriesOk(); ok {
									definitionMapArrayIntfArray := make([]map[string]interface{}, 0)

									for _, queriesArrayItem := range *queriesArray {
										definitionMapArrayIntfArrayIntf := map[string]interface{}{}
										if ddFormulaAndFunctionMetricQueryDefinition := queriesArrayItem.FormulaAndFunctionMetricQueryDefinition; ddFormulaAndFunctionMetricQueryDefinition != nil {
											mapFormulaAndFunctionMetricQueryDefinition := map[string]interface{}{}
											if v, ok := ddFormulaAndFunctionMetricQueryDefinition.GetAggregatorOk(); ok {
												mapFormulaAndFunctionMetricQueryDefinition["aggregator"] = *v
											}
											if v, ok := ddFormulaAndFunctionMetricQueryDefinition.GetDataSourceOk(); ok {
												mapFormulaAndFunctionMetricQueryDefinition["data_source"] = *v
											}
											if v, ok := ddFormulaAndFunctionMetricQueryDefinition.GetNameOk(); ok {
												mapFormulaAndFunctionMetricQueryDefinition["name"] = *v
											}
											if v, ok := ddFormulaAndFunctionMetricQueryDefinition.GetQueryOk(); ok {
												mapFormulaAndFunctionMetricQueryDefinition["query"] = *v
											}

											arrayFormulaAndFunctionMetricQueryDefinition := []interface{}{mapFormulaAndFunctionMetricQueryDefinition}
											definitionMapArrayIntfArrayIntf["formula_and_function_metric_query_definition"] = arrayFormulaAndFunctionMetricQueryDefinition
										}
										if ddFormulaAndFunctionEventQueryDefinition := queriesArrayItem.FormulaAndFunctionEventQueryDefinition; ddFormulaAndFunctionEventQueryDefinition != nil {
											mapFormulaAndFunctionEventQueryDefinition := map[string]interface{}{}
											if computeDDModel, ok := ddFormulaAndFunctionEventQueryDefinition.GetComputeOk(); ok {
												computeMap := map[string]interface{}{}
												if v, ok := computeDDModel.GetAggregationOk(); ok {
													computeMap["aggregation"] = *v
												}
												if v, ok := computeDDModel.GetIntervalOk(); ok {
													computeMap["interval"] = *v
												}
												if v, ok := computeDDModel.GetMetricOk(); ok {
													computeMap["metric"] = *v
												}

												mapFormulaAndFunctionEventQueryDefinition["compute"] = []map[string]interface{}{computeMap}
											}
											if v, ok := ddFormulaAndFunctionEventQueryDefinition.GetDataSourceOk(); ok {
												mapFormulaAndFunctionEventQueryDefinition["data_source"] = *v
											}
											if groupByArray, ok := ddFormulaAndFunctionEventQueryDefinition.GetGroupByOk(); ok {
												mapFormulaAndFunctionEventQueryDefinitionArray := make([]map[string]interface{}, 0)

												for _, groupByArrayItem := range *groupByArray {
													mapFormulaAndFunctionEventQueryDefinitionArrayIntf := map[string]interface{}{}
													if v, ok := groupByArrayItem.GetFacetOk(); ok {
														mapFormulaAndFunctionEventQueryDefinitionArrayIntf["facet"] = *v
													}
													if v, ok := groupByArrayItem.GetLimitOk(); ok {
														mapFormulaAndFunctionEventQueryDefinitionArrayIntf["limit"] = *v
													}
													if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
														sortMap := map[string]interface{}{}
														if v, ok := sortDDModel.GetAggregationOk(); ok {
															sortMap["aggregation"] = *v
														}
														if v, ok := sortDDModel.GetMetricOk(); ok {
															sortMap["metric"] = *v
														}
														if v, ok := sortDDModel.GetOrderOk(); ok {
															sortMap["order"] = *v
														}

														mapFormulaAndFunctionEventQueryDefinitionArrayIntf["sort"] = []map[string]interface{}{sortMap}
													}

													mapFormulaAndFunctionEventQueryDefinitionArray = append(mapFormulaAndFunctionEventQueryDefinitionArray, mapFormulaAndFunctionEventQueryDefinitionArrayIntf)
												}

												mapFormulaAndFunctionEventQueryDefinition["group_by"] = mapFormulaAndFunctionEventQueryDefinitionArray
											}
											if indexesArray, ok := ddFormulaAndFunctionEventQueryDefinition.GetIndexesOk(); ok {
												mapFormulaAndFunctionEventQueryDefinitionArray := make([]string, len(*indexesArray))
												for i, item := range *indexesArray {
													mapFormulaAndFunctionEventQueryDefinitionArray[i] = item
												}

												mapFormulaAndFunctionEventQueryDefinition["indexes"] = mapFormulaAndFunctionEventQueryDefinitionArray
											}
											if v, ok := ddFormulaAndFunctionEventQueryDefinition.GetNameOk(); ok {
												mapFormulaAndFunctionEventQueryDefinition["name"] = *v
											}
											if searchDDModel, ok := ddFormulaAndFunctionEventQueryDefinition.GetSearchOk(); ok {
												searchMap := map[string]interface{}{}
												if v, ok := searchDDModel.GetQueryOk(); ok {
													searchMap["query"] = *v
												}

												mapFormulaAndFunctionEventQueryDefinition["search"] = []map[string]interface{}{searchMap}
											}

											arrayFormulaAndFunctionEventQueryDefinition := []interface{}{mapFormulaAndFunctionEventQueryDefinition}
											definitionMapArrayIntfArrayIntf["formula_and_function_event_query_definition"] = arrayFormulaAndFunctionEventQueryDefinition
										}
										if ddFormulaAndFunctionProcessQueryDefinition := queriesArrayItem.FormulaAndFunctionProcessQueryDefinition; ddFormulaAndFunctionProcessQueryDefinition != nil {
											mapFormulaAndFunctionProcessQueryDefinition := map[string]interface{}{}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetAggregatorOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["aggregator"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetDataSourceOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["data_source"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetIsNormalizedCpuOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["is_normalized_cpu"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetLimitOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["limit"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetMetricOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["metric"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetNameOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["name"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetSortOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["sort"] = *v
											}
											if tagFiltersArray, ok := ddFormulaAndFunctionProcessQueryDefinition.GetTagFiltersOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinitionArray := make([]string, len(*tagFiltersArray))
												for i, item := range *tagFiltersArray {
													mapFormulaAndFunctionProcessQueryDefinitionArray[i] = item
												}

												mapFormulaAndFunctionProcessQueryDefinition["tag_filters"] = mapFormulaAndFunctionProcessQueryDefinitionArray
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetTextFilterOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["text_filter"] = *v
											}

											arrayFormulaAndFunctionProcessQueryDefinition := []interface{}{mapFormulaAndFunctionProcessQueryDefinition}
											definitionMapArrayIntfArrayIntf["formula_and_function_process_query_definition"] = arrayFormulaAndFunctionProcessQueryDefinition
										}

										definitionMapArrayIntfArray = append(definitionMapArrayIntfArray, definitionMapArrayIntfArrayIntf)
									}

									definitionMapArrayIntf["query"] = definitionMapArrayIntfArray
								}
								if v, ok := requestsArrayItem.GetResponseFormatOk(); ok {
									definitionMapArrayIntf["response_format"] = *v
								}
								if rumQueryDDModel, ok := requestsArrayItem.GetRumQueryOk(); ok {
									rumQueryMap := map[string]interface{}{}
									if computeDDModel, ok := rumQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										rumQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := rumQueryDDModel.GetGroupByOk(); ok {
										rumQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											rumQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												rumQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												rumQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												rumQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											rumQueryMapArray = append(rumQueryMapArray, rumQueryMapArrayIntf)
										}

										rumQueryMap["group_by"] = rumQueryMapArray
									}
									if v, ok := rumQueryDDModel.GetIndexOk(); ok {
										rumQueryMap["index"] = *v
									}
									if multiComputeArray, ok := rumQueryDDModel.GetMultiComputeOk(); ok {
										rumQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											rumQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												rumQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												rumQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												rumQueryMapArrayIntf["interval"] = *v
											}

											rumQueryMapArray = append(rumQueryMapArray, rumQueryMapArrayIntf)
										}

										rumQueryMap["multi_compute"] = rumQueryMapArray
									}
									if searchDDModel, ok := rumQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										rumQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["rum_query"] = []map[string]interface{}{rumQueryMap}
								}
								if securityQueryDDModel, ok := requestsArrayItem.GetSecurityQueryOk(); ok {
									securityQueryMap := map[string]interface{}{}
									if computeDDModel, ok := securityQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										securityQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := securityQueryDDModel.GetGroupByOk(); ok {
										securityQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											securityQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												securityQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												securityQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												securityQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											securityQueryMapArray = append(securityQueryMapArray, securityQueryMapArrayIntf)
										}

										securityQueryMap["group_by"] = securityQueryMapArray
									}
									if v, ok := securityQueryDDModel.GetIndexOk(); ok {
										securityQueryMap["index"] = *v
									}
									if multiComputeArray, ok := securityQueryDDModel.GetMultiComputeOk(); ok {
										securityQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											securityQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												securityQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												securityQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												securityQueryMapArrayIntf["interval"] = *v
											}

											securityQueryMapArray = append(securityQueryMapArray, securityQueryMapArrayIntf)
										}

										securityQueryMap["multi_compute"] = securityQueryMapArray
									}
									if searchDDModel, ok := securityQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										securityQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["security_query"] = []map[string]interface{}{securityQueryMap}
								}
								if styleDDModel, ok := requestsArrayItem.GetStyleOk(); ok {
									styleMap := map[string]interface{}{}
									if v, ok := styleDDModel.GetLineTypeOk(); ok {
										styleMap["line_type"] = *v
									}
									if v, ok := styleDDModel.GetLineWidthOk(); ok {
										styleMap["line_width"] = *v
									}
									if v, ok := styleDDModel.GetPaletteOk(); ok {
										styleMap["palette"] = *v
									}

									definitionMapArrayIntf["style"] = []map[string]interface{}{styleMap}
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["request"] = definitionMapArray
						}
						if rightYaxisDDModel, ok := definitionDDModel.GetRightYaxisOk(); ok {
							rightYaxisMap := map[string]interface{}{}
							if v, ok := rightYaxisDDModel.GetIncludeZeroOk(); ok {
								rightYaxisMap["include_zero"] = *v
							}
							if v, ok := rightYaxisDDModel.GetLabelOk(); ok {
								rightYaxisMap["label"] = *v
							}
							if v, ok := rightYaxisDDModel.GetMaxOk(); ok {
								rightYaxisMap["max"] = *v
							}
							if v, ok := rightYaxisDDModel.GetMinOk(); ok {
								rightYaxisMap["min"] = *v
							}
							if v, ok := rightYaxisDDModel.GetScaleOk(); ok {
								rightYaxisMap["scale"] = *v
							}

							definitionMap["right_yaxis"] = []map[string]interface{}{rightYaxisMap}
						}
						if v, ok := definitionDDModel.GetShowLegendOk(); ok {
							definitionMap["show_legend"] = *v
						}
						if timeDDModel, ok := definitionDDModel.GetTimeOk(); ok {
							timeMap := map[string]interface{}{}
							if v, ok := timeDDModel.GetLiveSpanOk(); ok {
								timeMap["live_span"] = *v
							}

							definitionMap["time"] = []map[string]interface{}{timeMap}
						}
						if v, ok := definitionDDModel.GetTitleOk(); ok {
							definitionMap["title"] = *v
						}
						if v, ok := definitionDDModel.GetTitleAlignOk(); ok {
							definitionMap["title_align"] = *v
						}
						if v, ok := definitionDDModel.GetTitleSizeOk(); ok {
							definitionMap["title_size"] = *v
						}
						if v, ok := definitionDDModel.GetTypeOk(); ok {
							definitionMap["type"] = *v
						}
						if yaxisDDModel, ok := definitionDDModel.GetYaxisOk(); ok {
							yaxisMap := map[string]interface{}{}
							if v, ok := yaxisDDModel.GetIncludeZeroOk(); ok {
								yaxisMap["include_zero"] = *v
							}
							if v, ok := yaxisDDModel.GetLabelOk(); ok {
								yaxisMap["label"] = *v
							}
							if v, ok := yaxisDDModel.GetMaxOk(); ok {
								yaxisMap["max"] = *v
							}
							if v, ok := yaxisDDModel.GetMinOk(); ok {
								yaxisMap["min"] = *v
							}
							if v, ok := yaxisDDModel.GetScaleOk(); ok {
								yaxisMap["scale"] = *v
							}

							definitionMap["yaxis"] = []map[string]interface{}{yaxisMap}
						}

						mapNotebookTimeseriesCellAttributes["definition"] = []map[string]interface{}{definitionMap}
					}
					if v, ok := ddNotebookTimeseriesCellAttributes.GetGraphSizeOk(); ok {
						mapNotebookTimeseriesCellAttributes["graph_size"] = *v
					}
					if splitByDDModel, ok := ddNotebookTimeseriesCellAttributes.GetSplitByOk(); ok {
						splitByMap := map[string]interface{}{}
						if keysArray, ok := splitByDDModel.GetKeysOk(); ok {
							splitByMapArray := make([]string, len(*keysArray))
							for i, item := range *keysArray {
								splitByMapArray[i] = item
							}

							splitByMap["keys"] = splitByMapArray
						}
						if tagsArray, ok := splitByDDModel.GetTagsOk(); ok {
							splitByMapArray := make([]string, len(*tagsArray))
							for i, item := range *tagsArray {
								splitByMapArray[i] = item
							}

							splitByMap["tags"] = splitByMapArray
						}

						mapNotebookTimeseriesCellAttributes["split_by"] = []map[string]interface{}{splitByMap}
					}
					if timeDDModel, ok := ddNotebookTimeseriesCellAttributes.GetTimeOk(); ok {
						timeMap := map[string]interface{}{}
						if ddNotebookRelativeTime := timeDDModel.NotebookRelativeTime; ddNotebookRelativeTime != nil {
							mapNotebookRelativeTime := map[string]interface{}{}
							if v, ok := ddNotebookRelativeTime.GetLiveSpanOk(); ok {
								mapNotebookRelativeTime["live_span"] = *v
							}

							arrayNotebookRelativeTime := []interface{}{mapNotebookRelativeTime}
							timeMap["notebook_relative_time"] = arrayNotebookRelativeTime
						}
						if ddNotebookAbsoluteTime := timeDDModel.NotebookAbsoluteTime; ddNotebookAbsoluteTime != nil {
							mapNotebookAbsoluteTime := map[string]interface{}{}
							if v, ok := ddNotebookAbsoluteTime.GetEndOk(); ok {
								mapNotebookAbsoluteTime["end"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}
							if v, ok := ddNotebookAbsoluteTime.GetLiveOk(); ok {
								mapNotebookAbsoluteTime["live"] = *v
							}
							if v, ok := ddNotebookAbsoluteTime.GetStartOk(); ok {
								mapNotebookAbsoluteTime["start"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}

							arrayNotebookAbsoluteTime := []interface{}{mapNotebookAbsoluteTime}
							timeMap["notebook_absolute_time"] = arrayNotebookAbsoluteTime
						}

						mapNotebookTimeseriesCellAttributes["time"] = []map[string]interface{}{timeMap}
					}

					arrayNotebookTimeseriesCellAttributes := []interface{}{mapNotebookTimeseriesCellAttributes}
					attributesMap["notebook_timeseries_cell_attributes"] = arrayNotebookTimeseriesCellAttributes
				}
				if ddNotebookToplistCellAttributes := attributesDDModel.NotebookToplistCellAttributes; ddNotebookToplistCellAttributes != nil {
					mapNotebookToplistCellAttributes := map[string]interface{}{}
					if definitionDDModel, ok := ddNotebookToplistCellAttributes.GetDefinitionOk(); ok {
						definitionMap := map[string]interface{}{}
						if customLinksArray, ok := definitionDDModel.GetCustomLinksOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, customLinksArrayItem := range *customLinksArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if v, ok := customLinksArrayItem.GetIsHiddenOk(); ok {
									definitionMapArrayIntf["is_hidden"] = *v
								}
								if v, ok := customLinksArrayItem.GetLabelOk(); ok {
									definitionMapArrayIntf["label"] = *v
								}
								if v, ok := customLinksArrayItem.GetLinkOk(); ok {
									definitionMapArrayIntf["link"] = *v
								}
								if v, ok := customLinksArrayItem.GetOverrideLabelOk(); ok {
									definitionMapArrayIntf["override_label"] = *v
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["custom_link"] = definitionMapArray
						}
						if requestsArray, ok := definitionDDModel.GetRequestsOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, requestsArrayItem := range *requestsArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if apmQueryDDModel, ok := requestsArrayItem.GetApmQueryOk(); ok {
									apmQueryMap := map[string]interface{}{}
									if computeDDModel, ok := apmQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										apmQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := apmQueryDDModel.GetGroupByOk(); ok {
										apmQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											apmQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												apmQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												apmQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												apmQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											apmQueryMapArray = append(apmQueryMapArray, apmQueryMapArrayIntf)
										}

										apmQueryMap["group_by"] = apmQueryMapArray
									}
									if v, ok := apmQueryDDModel.GetIndexOk(); ok {
										apmQueryMap["index"] = *v
									}
									if multiComputeArray, ok := apmQueryDDModel.GetMultiComputeOk(); ok {
										apmQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											apmQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												apmQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												apmQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												apmQueryMapArrayIntf["interval"] = *v
											}

											apmQueryMapArray = append(apmQueryMapArray, apmQueryMapArrayIntf)
										}

										apmQueryMap["multi_compute"] = apmQueryMapArray
									}
									if searchDDModel, ok := apmQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										apmQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["apm_query"] = []map[string]interface{}{apmQueryMap}
								}
								if conditionalFormatsArray, ok := requestsArrayItem.GetConditionalFormatsOk(); ok {
									definitionMapArrayIntfArray := make([]map[string]interface{}, 0)

									for _, conditionalFormatsArrayItem := range *conditionalFormatsArray {
										definitionMapArrayIntfArrayIntf := map[string]interface{}{}
										if v, ok := conditionalFormatsArrayItem.GetComparatorOk(); ok {
											definitionMapArrayIntfArrayIntf["comparator"] = *v
										}
										if v, ok := conditionalFormatsArrayItem.GetCustomBgColorOk(); ok {
											definitionMapArrayIntfArrayIntf["custom_bg_color"] = *v
										}
										if v, ok := conditionalFormatsArrayItem.GetCustomFgColorOk(); ok {
											definitionMapArrayIntfArrayIntf["custom_fg_color"] = *v
										}
										if v, ok := conditionalFormatsArrayItem.GetHideValueOk(); ok {
											definitionMapArrayIntfArrayIntf["hide_value"] = *v
										}
										if v, ok := conditionalFormatsArrayItem.GetImageUrlOk(); ok {
											definitionMapArrayIntfArrayIntf["image_url"] = *v
										}
										if v, ok := conditionalFormatsArrayItem.GetMetricOk(); ok {
											definitionMapArrayIntfArrayIntf["metric"] = *v
										}
										if v, ok := conditionalFormatsArrayItem.GetPaletteOk(); ok {
											definitionMapArrayIntfArrayIntf["palette"] = *v
										}
										if v, ok := conditionalFormatsArrayItem.GetTimeframeOk(); ok {
											definitionMapArrayIntfArrayIntf["timeframe"] = *v
										}
										if v, ok := conditionalFormatsArrayItem.GetValueOk(); ok {
											definitionMapArrayIntfArrayIntf["value"] = *v
										}

										definitionMapArrayIntfArray = append(definitionMapArrayIntfArray, definitionMapArrayIntfArrayIntf)
									}

									definitionMapArrayIntf["conditional_format"] = definitionMapArrayIntfArray
								}
								if eventQueryDDModel, ok := requestsArrayItem.GetEventQueryOk(); ok {
									eventQueryMap := map[string]interface{}{}
									if computeDDModel, ok := eventQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										eventQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := eventQueryDDModel.GetGroupByOk(); ok {
										eventQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											eventQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												eventQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												eventQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												eventQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											eventQueryMapArray = append(eventQueryMapArray, eventQueryMapArrayIntf)
										}

										eventQueryMap["group_by"] = eventQueryMapArray
									}
									if v, ok := eventQueryDDModel.GetIndexOk(); ok {
										eventQueryMap["index"] = *v
									}
									if multiComputeArray, ok := eventQueryDDModel.GetMultiComputeOk(); ok {
										eventQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											eventQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												eventQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												eventQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												eventQueryMapArrayIntf["interval"] = *v
											}

											eventQueryMapArray = append(eventQueryMapArray, eventQueryMapArrayIntf)
										}

										eventQueryMap["multi_compute"] = eventQueryMapArray
									}
									if searchDDModel, ok := eventQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										eventQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["event_query"] = []map[string]interface{}{eventQueryMap}
								}
								if formulasArray, ok := requestsArrayItem.GetFormulasOk(); ok {
									definitionMapArrayIntfArray := make([]map[string]interface{}, 0)

									for _, formulasArrayItem := range *formulasArray {
										definitionMapArrayIntfArrayIntf := map[string]interface{}{}
										if v, ok := formulasArrayItem.GetAliasOk(); ok {
											definitionMapArrayIntfArrayIntf["alias"] = *v
										}
										if v, ok := formulasArrayItem.GetFormulaOk(); ok {
											definitionMapArrayIntfArrayIntf["formula"] = *v
										}
										if limitDDModel, ok := formulasArrayItem.GetLimitOk(); ok {
											limitMap := map[string]interface{}{}
											if v, ok := limitDDModel.GetCountOk(); ok {
												limitMap["count"] = *v
											}
											if v, ok := limitDDModel.GetOrderOk(); ok {
												limitMap["order"] = *v
											}

											definitionMapArrayIntfArrayIntf["limit"] = []map[string]interface{}{limitMap}
										}

										definitionMapArrayIntfArray = append(definitionMapArrayIntfArray, definitionMapArrayIntfArrayIntf)
									}

									definitionMapArrayIntf["formula"] = definitionMapArrayIntfArray
								}
								if logQueryDDModel, ok := requestsArrayItem.GetLogQueryOk(); ok {
									logQueryMap := map[string]interface{}{}
									if computeDDModel, ok := logQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										logQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := logQueryDDModel.GetGroupByOk(); ok {
										logQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											logQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												logQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												logQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												logQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											logQueryMapArray = append(logQueryMapArray, logQueryMapArrayIntf)
										}

										logQueryMap["group_by"] = logQueryMapArray
									}
									if v, ok := logQueryDDModel.GetIndexOk(); ok {
										logQueryMap["index"] = *v
									}
									if multiComputeArray, ok := logQueryDDModel.GetMultiComputeOk(); ok {
										logQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											logQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												logQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												logQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												logQueryMapArrayIntf["interval"] = *v
											}

											logQueryMapArray = append(logQueryMapArray, logQueryMapArrayIntf)
										}

										logQueryMap["multi_compute"] = logQueryMapArray
									}
									if searchDDModel, ok := logQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										logQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["log_query"] = []map[string]interface{}{logQueryMap}
								}
								if networkQueryDDModel, ok := requestsArrayItem.GetNetworkQueryOk(); ok {
									networkQueryMap := map[string]interface{}{}
									if computeDDModel, ok := networkQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										networkQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := networkQueryDDModel.GetGroupByOk(); ok {
										networkQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											networkQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												networkQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												networkQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												networkQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											networkQueryMapArray = append(networkQueryMapArray, networkQueryMapArrayIntf)
										}

										networkQueryMap["group_by"] = networkQueryMapArray
									}
									if v, ok := networkQueryDDModel.GetIndexOk(); ok {
										networkQueryMap["index"] = *v
									}
									if multiComputeArray, ok := networkQueryDDModel.GetMultiComputeOk(); ok {
										networkQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											networkQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												networkQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												networkQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												networkQueryMapArrayIntf["interval"] = *v
											}

											networkQueryMapArray = append(networkQueryMapArray, networkQueryMapArrayIntf)
										}

										networkQueryMap["multi_compute"] = networkQueryMapArray
									}
									if searchDDModel, ok := networkQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										networkQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["network_query"] = []map[string]interface{}{networkQueryMap}
								}
								if processQueryDDModel, ok := requestsArrayItem.GetProcessQueryOk(); ok {
									processQueryMap := map[string]interface{}{}
									if filterByArray, ok := processQueryDDModel.GetFilterByOk(); ok {
										processQueryMapArray := make([]string, len(*filterByArray))
										for i, item := range *filterByArray {
											processQueryMapArray[i] = item
										}

										processQueryMap["filter_by"] = processQueryMapArray
									}
									if v, ok := processQueryDDModel.GetLimitOk(); ok {
										processQueryMap["limit"] = *v
									}
									if v, ok := processQueryDDModel.GetMetricOk(); ok {
										processQueryMap["metric"] = *v
									}
									if v, ok := processQueryDDModel.GetSearchByOk(); ok {
										processQueryMap["search_by"] = *v
									}

									definitionMapArrayIntf["process_query"] = []map[string]interface{}{processQueryMap}
								}
								if profileMetricsQueryDDModel, ok := requestsArrayItem.GetProfileMetricsQueryOk(); ok {
									profileMetricsQueryMap := map[string]interface{}{}
									if computeDDModel, ok := profileMetricsQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										profileMetricsQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := profileMetricsQueryDDModel.GetGroupByOk(); ok {
										profileMetricsQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											profileMetricsQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												profileMetricsQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												profileMetricsQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												profileMetricsQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											profileMetricsQueryMapArray = append(profileMetricsQueryMapArray, profileMetricsQueryMapArrayIntf)
										}

										profileMetricsQueryMap["group_by"] = profileMetricsQueryMapArray
									}
									if v, ok := profileMetricsQueryDDModel.GetIndexOk(); ok {
										profileMetricsQueryMap["index"] = *v
									}
									if multiComputeArray, ok := profileMetricsQueryDDModel.GetMultiComputeOk(); ok {
										profileMetricsQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											profileMetricsQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												profileMetricsQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												profileMetricsQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												profileMetricsQueryMapArrayIntf["interval"] = *v
											}

											profileMetricsQueryMapArray = append(profileMetricsQueryMapArray, profileMetricsQueryMapArrayIntf)
										}

										profileMetricsQueryMap["multi_compute"] = profileMetricsQueryMapArray
									}
									if searchDDModel, ok := profileMetricsQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										profileMetricsQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["profile_metrics_query"] = []map[string]interface{}{profileMetricsQueryMap}
								}
								if v, ok := requestsArrayItem.GetQOk(); ok {
									definitionMapArrayIntf["q"] = *v
								}
								if queriesArray, ok := requestsArrayItem.GetQueriesOk(); ok {
									definitionMapArrayIntfArray := make([]map[string]interface{}, 0)

									for _, queriesArrayItem := range *queriesArray {
										definitionMapArrayIntfArrayIntf := map[string]interface{}{}
										if ddFormulaAndFunctionMetricQueryDefinition := queriesArrayItem.FormulaAndFunctionMetricQueryDefinition; ddFormulaAndFunctionMetricQueryDefinition != nil {
											mapFormulaAndFunctionMetricQueryDefinition := map[string]interface{}{}
											if v, ok := ddFormulaAndFunctionMetricQueryDefinition.GetAggregatorOk(); ok {
												mapFormulaAndFunctionMetricQueryDefinition["aggregator"] = *v
											}
											if v, ok := ddFormulaAndFunctionMetricQueryDefinition.GetDataSourceOk(); ok {
												mapFormulaAndFunctionMetricQueryDefinition["data_source"] = *v
											}
											if v, ok := ddFormulaAndFunctionMetricQueryDefinition.GetNameOk(); ok {
												mapFormulaAndFunctionMetricQueryDefinition["name"] = *v
											}
											if v, ok := ddFormulaAndFunctionMetricQueryDefinition.GetQueryOk(); ok {
												mapFormulaAndFunctionMetricQueryDefinition["query"] = *v
											}

											arrayFormulaAndFunctionMetricQueryDefinition := []interface{}{mapFormulaAndFunctionMetricQueryDefinition}
											definitionMapArrayIntfArrayIntf["formula_and_function_metric_query_definition"] = arrayFormulaAndFunctionMetricQueryDefinition
										}
										if ddFormulaAndFunctionEventQueryDefinition := queriesArrayItem.FormulaAndFunctionEventQueryDefinition; ddFormulaAndFunctionEventQueryDefinition != nil {
											mapFormulaAndFunctionEventQueryDefinition := map[string]interface{}{}
											if computeDDModel, ok := ddFormulaAndFunctionEventQueryDefinition.GetComputeOk(); ok {
												computeMap := map[string]interface{}{}
												if v, ok := computeDDModel.GetAggregationOk(); ok {
													computeMap["aggregation"] = *v
												}
												if v, ok := computeDDModel.GetIntervalOk(); ok {
													computeMap["interval"] = *v
												}
												if v, ok := computeDDModel.GetMetricOk(); ok {
													computeMap["metric"] = *v
												}

												mapFormulaAndFunctionEventQueryDefinition["compute"] = []map[string]interface{}{computeMap}
											}
											if v, ok := ddFormulaAndFunctionEventQueryDefinition.GetDataSourceOk(); ok {
												mapFormulaAndFunctionEventQueryDefinition["data_source"] = *v
											}
											if groupByArray, ok := ddFormulaAndFunctionEventQueryDefinition.GetGroupByOk(); ok {
												mapFormulaAndFunctionEventQueryDefinitionArray := make([]map[string]interface{}, 0)

												for _, groupByArrayItem := range *groupByArray {
													mapFormulaAndFunctionEventQueryDefinitionArrayIntf := map[string]interface{}{}
													if v, ok := groupByArrayItem.GetFacetOk(); ok {
														mapFormulaAndFunctionEventQueryDefinitionArrayIntf["facet"] = *v
													}
													if v, ok := groupByArrayItem.GetLimitOk(); ok {
														mapFormulaAndFunctionEventQueryDefinitionArrayIntf["limit"] = *v
													}
													if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
														sortMap := map[string]interface{}{}
														if v, ok := sortDDModel.GetAggregationOk(); ok {
															sortMap["aggregation"] = *v
														}
														if v, ok := sortDDModel.GetMetricOk(); ok {
															sortMap["metric"] = *v
														}
														if v, ok := sortDDModel.GetOrderOk(); ok {
															sortMap["order"] = *v
														}

														mapFormulaAndFunctionEventQueryDefinitionArrayIntf["sort"] = []map[string]interface{}{sortMap}
													}

													mapFormulaAndFunctionEventQueryDefinitionArray = append(mapFormulaAndFunctionEventQueryDefinitionArray, mapFormulaAndFunctionEventQueryDefinitionArrayIntf)
												}

												mapFormulaAndFunctionEventQueryDefinition["group_by"] = mapFormulaAndFunctionEventQueryDefinitionArray
											}
											if indexesArray, ok := ddFormulaAndFunctionEventQueryDefinition.GetIndexesOk(); ok {
												mapFormulaAndFunctionEventQueryDefinitionArray := make([]string, len(*indexesArray))
												for i, item := range *indexesArray {
													mapFormulaAndFunctionEventQueryDefinitionArray[i] = item
												}

												mapFormulaAndFunctionEventQueryDefinition["indexes"] = mapFormulaAndFunctionEventQueryDefinitionArray
											}
											if v, ok := ddFormulaAndFunctionEventQueryDefinition.GetNameOk(); ok {
												mapFormulaAndFunctionEventQueryDefinition["name"] = *v
											}
											if searchDDModel, ok := ddFormulaAndFunctionEventQueryDefinition.GetSearchOk(); ok {
												searchMap := map[string]interface{}{}
												if v, ok := searchDDModel.GetQueryOk(); ok {
													searchMap["query"] = *v
												}

												mapFormulaAndFunctionEventQueryDefinition["search"] = []map[string]interface{}{searchMap}
											}

											arrayFormulaAndFunctionEventQueryDefinition := []interface{}{mapFormulaAndFunctionEventQueryDefinition}
											definitionMapArrayIntfArrayIntf["formula_and_function_event_query_definition"] = arrayFormulaAndFunctionEventQueryDefinition
										}
										if ddFormulaAndFunctionProcessQueryDefinition := queriesArrayItem.FormulaAndFunctionProcessQueryDefinition; ddFormulaAndFunctionProcessQueryDefinition != nil {
											mapFormulaAndFunctionProcessQueryDefinition := map[string]interface{}{}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetAggregatorOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["aggregator"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetDataSourceOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["data_source"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetIsNormalizedCpuOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["is_normalized_cpu"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetLimitOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["limit"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetMetricOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["metric"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetNameOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["name"] = *v
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetSortOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["sort"] = *v
											}
											if tagFiltersArray, ok := ddFormulaAndFunctionProcessQueryDefinition.GetTagFiltersOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinitionArray := make([]string, len(*tagFiltersArray))
												for i, item := range *tagFiltersArray {
													mapFormulaAndFunctionProcessQueryDefinitionArray[i] = item
												}

												mapFormulaAndFunctionProcessQueryDefinition["tag_filters"] = mapFormulaAndFunctionProcessQueryDefinitionArray
											}
											if v, ok := ddFormulaAndFunctionProcessQueryDefinition.GetTextFilterOk(); ok {
												mapFormulaAndFunctionProcessQueryDefinition["text_filter"] = *v
											}

											arrayFormulaAndFunctionProcessQueryDefinition := []interface{}{mapFormulaAndFunctionProcessQueryDefinition}
											definitionMapArrayIntfArrayIntf["formula_and_function_process_query_definition"] = arrayFormulaAndFunctionProcessQueryDefinition
										}

										definitionMapArrayIntfArray = append(definitionMapArrayIntfArray, definitionMapArrayIntfArrayIntf)
									}

									definitionMapArrayIntf["query"] = definitionMapArrayIntfArray
								}
								if v, ok := requestsArrayItem.GetResponseFormatOk(); ok {
									definitionMapArrayIntf["response_format"] = *v
								}
								if rumQueryDDModel, ok := requestsArrayItem.GetRumQueryOk(); ok {
									rumQueryMap := map[string]interface{}{}
									if computeDDModel, ok := rumQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										rumQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := rumQueryDDModel.GetGroupByOk(); ok {
										rumQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											rumQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												rumQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												rumQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												rumQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											rumQueryMapArray = append(rumQueryMapArray, rumQueryMapArrayIntf)
										}

										rumQueryMap["group_by"] = rumQueryMapArray
									}
									if v, ok := rumQueryDDModel.GetIndexOk(); ok {
										rumQueryMap["index"] = *v
									}
									if multiComputeArray, ok := rumQueryDDModel.GetMultiComputeOk(); ok {
										rumQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											rumQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												rumQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												rumQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												rumQueryMapArrayIntf["interval"] = *v
											}

											rumQueryMapArray = append(rumQueryMapArray, rumQueryMapArrayIntf)
										}

										rumQueryMap["multi_compute"] = rumQueryMapArray
									}
									if searchDDModel, ok := rumQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										rumQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["rum_query"] = []map[string]interface{}{rumQueryMap}
								}
								if securityQueryDDModel, ok := requestsArrayItem.GetSecurityQueryOk(); ok {
									securityQueryMap := map[string]interface{}{}
									if computeDDModel, ok := securityQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										securityQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := securityQueryDDModel.GetGroupByOk(); ok {
										securityQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											securityQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												securityQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												securityQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												securityQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											securityQueryMapArray = append(securityQueryMapArray, securityQueryMapArrayIntf)
										}

										securityQueryMap["group_by"] = securityQueryMapArray
									}
									if v, ok := securityQueryDDModel.GetIndexOk(); ok {
										securityQueryMap["index"] = *v
									}
									if multiComputeArray, ok := securityQueryDDModel.GetMultiComputeOk(); ok {
										securityQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											securityQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												securityQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												securityQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												securityQueryMapArrayIntf["interval"] = *v
											}

											securityQueryMapArray = append(securityQueryMapArray, securityQueryMapArrayIntf)
										}

										securityQueryMap["multi_compute"] = securityQueryMapArray
									}
									if searchDDModel, ok := securityQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										securityQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["security_query"] = []map[string]interface{}{securityQueryMap}
								}
								if styleDDModel, ok := requestsArrayItem.GetStyleOk(); ok {
									styleMap := map[string]interface{}{}
									if v, ok := styleDDModel.GetLineTypeOk(); ok {
										styleMap["line_type"] = *v
									}
									if v, ok := styleDDModel.GetLineWidthOk(); ok {
										styleMap["line_width"] = *v
									}
									if v, ok := styleDDModel.GetPaletteOk(); ok {
										styleMap["palette"] = *v
									}

									definitionMapArrayIntf["style"] = []map[string]interface{}{styleMap}
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["request"] = definitionMapArray
						}
						if timeDDModel, ok := definitionDDModel.GetTimeOk(); ok {
							timeMap := map[string]interface{}{}
							if v, ok := timeDDModel.GetLiveSpanOk(); ok {
								timeMap["live_span"] = *v
							}

							definitionMap["time"] = []map[string]interface{}{timeMap}
						}
						if v, ok := definitionDDModel.GetTitleOk(); ok {
							definitionMap["title"] = *v
						}
						if v, ok := definitionDDModel.GetTitleAlignOk(); ok {
							definitionMap["title_align"] = *v
						}
						if v, ok := definitionDDModel.GetTitleSizeOk(); ok {
							definitionMap["title_size"] = *v
						}
						if v, ok := definitionDDModel.GetTypeOk(); ok {
							definitionMap["type"] = *v
						}

						mapNotebookToplistCellAttributes["definition"] = []map[string]interface{}{definitionMap}
					}
					if v, ok := ddNotebookToplistCellAttributes.GetGraphSizeOk(); ok {
						mapNotebookToplistCellAttributes["graph_size"] = *v
					}
					if splitByDDModel, ok := ddNotebookToplistCellAttributes.GetSplitByOk(); ok {
						splitByMap := map[string]interface{}{}
						if keysArray, ok := splitByDDModel.GetKeysOk(); ok {
							splitByMapArray := make([]string, len(*keysArray))
							for i, item := range *keysArray {
								splitByMapArray[i] = item
							}

							splitByMap["keys"] = splitByMapArray
						}
						if tagsArray, ok := splitByDDModel.GetTagsOk(); ok {
							splitByMapArray := make([]string, len(*tagsArray))
							for i, item := range *tagsArray {
								splitByMapArray[i] = item
							}

							splitByMap["tags"] = splitByMapArray
						}

						mapNotebookToplistCellAttributes["split_by"] = []map[string]interface{}{splitByMap}
					}
					if timeDDModel, ok := ddNotebookToplistCellAttributes.GetTimeOk(); ok {
						timeMap := map[string]interface{}{}
						if ddNotebookRelativeTime := timeDDModel.NotebookRelativeTime; ddNotebookRelativeTime != nil {
							mapNotebookRelativeTime := map[string]interface{}{}
							if v, ok := ddNotebookRelativeTime.GetLiveSpanOk(); ok {
								mapNotebookRelativeTime["live_span"] = *v
							}

							arrayNotebookRelativeTime := []interface{}{mapNotebookRelativeTime}
							timeMap["notebook_relative_time"] = arrayNotebookRelativeTime
						}
						if ddNotebookAbsoluteTime := timeDDModel.NotebookAbsoluteTime; ddNotebookAbsoluteTime != nil {
							mapNotebookAbsoluteTime := map[string]interface{}{}
							if v, ok := ddNotebookAbsoluteTime.GetEndOk(); ok {
								mapNotebookAbsoluteTime["end"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}
							if v, ok := ddNotebookAbsoluteTime.GetLiveOk(); ok {
								mapNotebookAbsoluteTime["live"] = *v
							}
							if v, ok := ddNotebookAbsoluteTime.GetStartOk(); ok {
								mapNotebookAbsoluteTime["start"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}

							arrayNotebookAbsoluteTime := []interface{}{mapNotebookAbsoluteTime}
							timeMap["notebook_absolute_time"] = arrayNotebookAbsoluteTime
						}

						mapNotebookToplistCellAttributes["time"] = []map[string]interface{}{timeMap}
					}

					arrayNotebookToplistCellAttributes := []interface{}{mapNotebookToplistCellAttributes}
					attributesMap["notebook_toplist_cell_attributes"] = arrayNotebookToplistCellAttributes
				}
				if ddNotebookHeatMapCellAttributes := attributesDDModel.NotebookHeatMapCellAttributes; ddNotebookHeatMapCellAttributes != nil {
					mapNotebookHeatMapCellAttributes := map[string]interface{}{}
					if definitionDDModel, ok := ddNotebookHeatMapCellAttributes.GetDefinitionOk(); ok {
						definitionMap := map[string]interface{}{}
						if customLinksArray, ok := definitionDDModel.GetCustomLinksOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, customLinksArrayItem := range *customLinksArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if v, ok := customLinksArrayItem.GetIsHiddenOk(); ok {
									definitionMapArrayIntf["is_hidden"] = *v
								}
								if v, ok := customLinksArrayItem.GetLabelOk(); ok {
									definitionMapArrayIntf["label"] = *v
								}
								if v, ok := customLinksArrayItem.GetLinkOk(); ok {
									definitionMapArrayIntf["link"] = *v
								}
								if v, ok := customLinksArrayItem.GetOverrideLabelOk(); ok {
									definitionMapArrayIntf["override_label"] = *v
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["custom_link"] = definitionMapArray
						}
						if eventsArray, ok := definitionDDModel.GetEventsOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, eventsArrayItem := range *eventsArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if v, ok := eventsArrayItem.GetQOk(); ok {
									definitionMapArrayIntf["q"] = *v
								}
								if v, ok := eventsArrayItem.GetTagsExecutionOk(); ok {
									definitionMapArrayIntf["tags_execution"] = *v
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["event"] = definitionMapArray
						}
						if v, ok := definitionDDModel.GetLegendSizeOk(); ok {
							definitionMap["legend_size"] = *v
						}
						if requestsArray, ok := definitionDDModel.GetRequestsOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, requestsArrayItem := range *requestsArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if apmQueryDDModel, ok := requestsArrayItem.GetApmQueryOk(); ok {
									apmQueryMap := map[string]interface{}{}
									if computeDDModel, ok := apmQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										apmQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := apmQueryDDModel.GetGroupByOk(); ok {
										apmQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											apmQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												apmQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												apmQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												apmQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											apmQueryMapArray = append(apmQueryMapArray, apmQueryMapArrayIntf)
										}

										apmQueryMap["group_by"] = apmQueryMapArray
									}
									if v, ok := apmQueryDDModel.GetIndexOk(); ok {
										apmQueryMap["index"] = *v
									}
									if multiComputeArray, ok := apmQueryDDModel.GetMultiComputeOk(); ok {
										apmQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											apmQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												apmQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												apmQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												apmQueryMapArrayIntf["interval"] = *v
											}

											apmQueryMapArray = append(apmQueryMapArray, apmQueryMapArrayIntf)
										}

										apmQueryMap["multi_compute"] = apmQueryMapArray
									}
									if searchDDModel, ok := apmQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										apmQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["apm_query"] = []map[string]interface{}{apmQueryMap}
								}
								if eventQueryDDModel, ok := requestsArrayItem.GetEventQueryOk(); ok {
									eventQueryMap := map[string]interface{}{}
									if v, ok := eventQueryDDModel.GetSearchOk(); ok {
										eventQueryMap["search"] = *v
									}
									if v, ok := eventQueryDDModel.GetTagsExecutionOk(); ok {
										eventQueryMap["tags_execution"] = *v
									}

									definitionMapArrayIntf["event_query"] = []map[string]interface{}{eventQueryMap}
								}
								if logQueryDDModel, ok := requestsArrayItem.GetLogQueryOk(); ok {
									logQueryMap := map[string]interface{}{}
									if computeDDModel, ok := logQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										logQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := logQueryDDModel.GetGroupByOk(); ok {
										logQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											logQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												logQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												logQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												logQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											logQueryMapArray = append(logQueryMapArray, logQueryMapArrayIntf)
										}

										logQueryMap["group_by"] = logQueryMapArray
									}
									if v, ok := logQueryDDModel.GetIndexOk(); ok {
										logQueryMap["index"] = *v
									}
									if multiComputeArray, ok := logQueryDDModel.GetMultiComputeOk(); ok {
										logQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											logQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												logQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												logQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												logQueryMapArrayIntf["interval"] = *v
											}

											logQueryMapArray = append(logQueryMapArray, logQueryMapArrayIntf)
										}

										logQueryMap["multi_compute"] = logQueryMapArray
									}
									if searchDDModel, ok := logQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										logQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["log_query"] = []map[string]interface{}{logQueryMap}
								}
								if networkQueryDDModel, ok := requestsArrayItem.GetNetworkQueryOk(); ok {
									networkQueryMap := map[string]interface{}{}
									if computeDDModel, ok := networkQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										networkQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := networkQueryDDModel.GetGroupByOk(); ok {
										networkQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											networkQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												networkQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												networkQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												networkQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											networkQueryMapArray = append(networkQueryMapArray, networkQueryMapArrayIntf)
										}

										networkQueryMap["group_by"] = networkQueryMapArray
									}
									if v, ok := networkQueryDDModel.GetIndexOk(); ok {
										networkQueryMap["index"] = *v
									}
									if multiComputeArray, ok := networkQueryDDModel.GetMultiComputeOk(); ok {
										networkQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											networkQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												networkQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												networkQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												networkQueryMapArrayIntf["interval"] = *v
											}

											networkQueryMapArray = append(networkQueryMapArray, networkQueryMapArrayIntf)
										}

										networkQueryMap["multi_compute"] = networkQueryMapArray
									}
									if searchDDModel, ok := networkQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										networkQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["network_query"] = []map[string]interface{}{networkQueryMap}
								}
								if processQueryDDModel, ok := requestsArrayItem.GetProcessQueryOk(); ok {
									processQueryMap := map[string]interface{}{}
									if filterByArray, ok := processQueryDDModel.GetFilterByOk(); ok {
										processQueryMapArray := make([]string, len(*filterByArray))
										for i, item := range *filterByArray {
											processQueryMapArray[i] = item
										}

										processQueryMap["filter_by"] = processQueryMapArray
									}
									if v, ok := processQueryDDModel.GetLimitOk(); ok {
										processQueryMap["limit"] = *v
									}
									if v, ok := processQueryDDModel.GetMetricOk(); ok {
										processQueryMap["metric"] = *v
									}
									if v, ok := processQueryDDModel.GetSearchByOk(); ok {
										processQueryMap["search_by"] = *v
									}

									definitionMapArrayIntf["process_query"] = []map[string]interface{}{processQueryMap}
								}
								if profileMetricsQueryDDModel, ok := requestsArrayItem.GetProfileMetricsQueryOk(); ok {
									profileMetricsQueryMap := map[string]interface{}{}
									if computeDDModel, ok := profileMetricsQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										profileMetricsQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := profileMetricsQueryDDModel.GetGroupByOk(); ok {
										profileMetricsQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											profileMetricsQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												profileMetricsQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												profileMetricsQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												profileMetricsQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											profileMetricsQueryMapArray = append(profileMetricsQueryMapArray, profileMetricsQueryMapArrayIntf)
										}

										profileMetricsQueryMap["group_by"] = profileMetricsQueryMapArray
									}
									if v, ok := profileMetricsQueryDDModel.GetIndexOk(); ok {
										profileMetricsQueryMap["index"] = *v
									}
									if multiComputeArray, ok := profileMetricsQueryDDModel.GetMultiComputeOk(); ok {
										profileMetricsQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											profileMetricsQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												profileMetricsQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												profileMetricsQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												profileMetricsQueryMapArrayIntf["interval"] = *v
											}

											profileMetricsQueryMapArray = append(profileMetricsQueryMapArray, profileMetricsQueryMapArrayIntf)
										}

										profileMetricsQueryMap["multi_compute"] = profileMetricsQueryMapArray
									}
									if searchDDModel, ok := profileMetricsQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										profileMetricsQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["profile_metrics_query"] = []map[string]interface{}{profileMetricsQueryMap}
								}
								if v, ok := requestsArrayItem.GetQOk(); ok {
									definitionMapArrayIntf["q"] = *v
								}
								if rumQueryDDModel, ok := requestsArrayItem.GetRumQueryOk(); ok {
									rumQueryMap := map[string]interface{}{}
									if computeDDModel, ok := rumQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										rumQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := rumQueryDDModel.GetGroupByOk(); ok {
										rumQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											rumQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												rumQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												rumQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												rumQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											rumQueryMapArray = append(rumQueryMapArray, rumQueryMapArrayIntf)
										}

										rumQueryMap["group_by"] = rumQueryMapArray
									}
									if v, ok := rumQueryDDModel.GetIndexOk(); ok {
										rumQueryMap["index"] = *v
									}
									if multiComputeArray, ok := rumQueryDDModel.GetMultiComputeOk(); ok {
										rumQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											rumQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												rumQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												rumQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												rumQueryMapArrayIntf["interval"] = *v
											}

											rumQueryMapArray = append(rumQueryMapArray, rumQueryMapArrayIntf)
										}

										rumQueryMap["multi_compute"] = rumQueryMapArray
									}
									if searchDDModel, ok := rumQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										rumQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["rum_query"] = []map[string]interface{}{rumQueryMap}
								}
								if securityQueryDDModel, ok := requestsArrayItem.GetSecurityQueryOk(); ok {
									securityQueryMap := map[string]interface{}{}
									if computeDDModel, ok := securityQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										securityQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := securityQueryDDModel.GetGroupByOk(); ok {
										securityQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											securityQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												securityQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												securityQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												securityQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											securityQueryMapArray = append(securityQueryMapArray, securityQueryMapArrayIntf)
										}

										securityQueryMap["group_by"] = securityQueryMapArray
									}
									if v, ok := securityQueryDDModel.GetIndexOk(); ok {
										securityQueryMap["index"] = *v
									}
									if multiComputeArray, ok := securityQueryDDModel.GetMultiComputeOk(); ok {
										securityQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											securityQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												securityQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												securityQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												securityQueryMapArrayIntf["interval"] = *v
											}

											securityQueryMapArray = append(securityQueryMapArray, securityQueryMapArrayIntf)
										}

										securityQueryMap["multi_compute"] = securityQueryMapArray
									}
									if searchDDModel, ok := securityQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										securityQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["security_query"] = []map[string]interface{}{securityQueryMap}
								}
								if styleDDModel, ok := requestsArrayItem.GetStyleOk(); ok {
									styleMap := map[string]interface{}{}
									if v, ok := styleDDModel.GetPaletteOk(); ok {
										styleMap["palette"] = *v
									}

									definitionMapArrayIntf["style"] = []map[string]interface{}{styleMap}
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["request"] = definitionMapArray
						}
						if v, ok := definitionDDModel.GetShowLegendOk(); ok {
							definitionMap["show_legend"] = *v
						}
						if timeDDModel, ok := definitionDDModel.GetTimeOk(); ok {
							timeMap := map[string]interface{}{}
							if v, ok := timeDDModel.GetLiveSpanOk(); ok {
								timeMap["live_span"] = *v
							}

							definitionMap["time"] = []map[string]interface{}{timeMap}
						}
						if v, ok := definitionDDModel.GetTitleOk(); ok {
							definitionMap["title"] = *v
						}
						if v, ok := definitionDDModel.GetTitleAlignOk(); ok {
							definitionMap["title_align"] = *v
						}
						if v, ok := definitionDDModel.GetTitleSizeOk(); ok {
							definitionMap["title_size"] = *v
						}
						if v, ok := definitionDDModel.GetTypeOk(); ok {
							definitionMap["type"] = *v
						}
						if yaxisDDModel, ok := definitionDDModel.GetYaxisOk(); ok {
							yaxisMap := map[string]interface{}{}
							if v, ok := yaxisDDModel.GetIncludeZeroOk(); ok {
								yaxisMap["include_zero"] = *v
							}
							if v, ok := yaxisDDModel.GetLabelOk(); ok {
								yaxisMap["label"] = *v
							}
							if v, ok := yaxisDDModel.GetMaxOk(); ok {
								yaxisMap["max"] = *v
							}
							if v, ok := yaxisDDModel.GetMinOk(); ok {
								yaxisMap["min"] = *v
							}
							if v, ok := yaxisDDModel.GetScaleOk(); ok {
								yaxisMap["scale"] = *v
							}

							definitionMap["yaxis"] = []map[string]interface{}{yaxisMap}
						}

						mapNotebookHeatMapCellAttributes["definition"] = []map[string]interface{}{definitionMap}
					}
					if v, ok := ddNotebookHeatMapCellAttributes.GetGraphSizeOk(); ok {
						mapNotebookHeatMapCellAttributes["graph_size"] = *v
					}
					if splitByDDModel, ok := ddNotebookHeatMapCellAttributes.GetSplitByOk(); ok {
						splitByMap := map[string]interface{}{}
						if keysArray, ok := splitByDDModel.GetKeysOk(); ok {
							splitByMapArray := make([]string, len(*keysArray))
							for i, item := range *keysArray {
								splitByMapArray[i] = item
							}

							splitByMap["keys"] = splitByMapArray
						}
						if tagsArray, ok := splitByDDModel.GetTagsOk(); ok {
							splitByMapArray := make([]string, len(*tagsArray))
							for i, item := range *tagsArray {
								splitByMapArray[i] = item
							}

							splitByMap["tags"] = splitByMapArray
						}

						mapNotebookHeatMapCellAttributes["split_by"] = []map[string]interface{}{splitByMap}
					}
					if timeDDModel, ok := ddNotebookHeatMapCellAttributes.GetTimeOk(); ok {
						timeMap := map[string]interface{}{}
						if ddNotebookRelativeTime := timeDDModel.NotebookRelativeTime; ddNotebookRelativeTime != nil {
							mapNotebookRelativeTime := map[string]interface{}{}
							if v, ok := ddNotebookRelativeTime.GetLiveSpanOk(); ok {
								mapNotebookRelativeTime["live_span"] = *v
							}

							arrayNotebookRelativeTime := []interface{}{mapNotebookRelativeTime}
							timeMap["notebook_relative_time"] = arrayNotebookRelativeTime
						}
						if ddNotebookAbsoluteTime := timeDDModel.NotebookAbsoluteTime; ddNotebookAbsoluteTime != nil {
							mapNotebookAbsoluteTime := map[string]interface{}{}
							if v, ok := ddNotebookAbsoluteTime.GetEndOk(); ok {
								mapNotebookAbsoluteTime["end"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}
							if v, ok := ddNotebookAbsoluteTime.GetLiveOk(); ok {
								mapNotebookAbsoluteTime["live"] = *v
							}
							if v, ok := ddNotebookAbsoluteTime.GetStartOk(); ok {
								mapNotebookAbsoluteTime["start"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}

							arrayNotebookAbsoluteTime := []interface{}{mapNotebookAbsoluteTime}
							timeMap["notebook_absolute_time"] = arrayNotebookAbsoluteTime
						}

						mapNotebookHeatMapCellAttributes["time"] = []map[string]interface{}{timeMap}
					}

					arrayNotebookHeatMapCellAttributes := []interface{}{mapNotebookHeatMapCellAttributes}
					attributesMap["notebook_heat_map_cell_attributes"] = arrayNotebookHeatMapCellAttributes
				}
				if ddNotebookDistributionCellAttributes := attributesDDModel.NotebookDistributionCellAttributes; ddNotebookDistributionCellAttributes != nil {
					mapNotebookDistributionCellAttributes := map[string]interface{}{}
					if definitionDDModel, ok := ddNotebookDistributionCellAttributes.GetDefinitionOk(); ok {
						definitionMap := map[string]interface{}{}
						if v, ok := definitionDDModel.GetLegendSizeOk(); ok {
							definitionMap["legend_size"] = *v
						}
						if markersArray, ok := definitionDDModel.GetMarkersOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, markersArrayItem := range *markersArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if v, ok := markersArrayItem.GetDisplayTypeOk(); ok {
									definitionMapArrayIntf["display_type"] = *v
								}
								if v, ok := markersArrayItem.GetLabelOk(); ok {
									definitionMapArrayIntf["label"] = *v
								}
								if v, ok := markersArrayItem.GetTimeOk(); ok {
									definitionMapArrayIntf["time"] = *v
								}
								if v, ok := markersArrayItem.GetValueOk(); ok {
									definitionMapArrayIntf["value"] = *v
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["marker"] = definitionMapArray
						}
						if requestsArray, ok := definitionDDModel.GetRequestsOk(); ok {
							definitionMapArray := make([]map[string]interface{}, 0)

							for _, requestsArrayItem := range *requestsArray {
								definitionMapArrayIntf := map[string]interface{}{}
								if apmQueryDDModel, ok := requestsArrayItem.GetApmQueryOk(); ok {
									apmQueryMap := map[string]interface{}{}
									if computeDDModel, ok := apmQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										apmQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := apmQueryDDModel.GetGroupByOk(); ok {
										apmQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											apmQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												apmQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												apmQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												apmQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											apmQueryMapArray = append(apmQueryMapArray, apmQueryMapArrayIntf)
										}

										apmQueryMap["group_by"] = apmQueryMapArray
									}
									if v, ok := apmQueryDDModel.GetIndexOk(); ok {
										apmQueryMap["index"] = *v
									}
									if multiComputeArray, ok := apmQueryDDModel.GetMultiComputeOk(); ok {
										apmQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											apmQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												apmQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												apmQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												apmQueryMapArrayIntf["interval"] = *v
											}

											apmQueryMapArray = append(apmQueryMapArray, apmQueryMapArrayIntf)
										}

										apmQueryMap["multi_compute"] = apmQueryMapArray
									}
									if searchDDModel, ok := apmQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										apmQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["apm_query"] = []map[string]interface{}{apmQueryMap}
								}
								if eventQueryDDModel, ok := requestsArrayItem.GetEventQueryOk(); ok {
									eventQueryMap := map[string]interface{}{}
									if computeDDModel, ok := eventQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										eventQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := eventQueryDDModel.GetGroupByOk(); ok {
										eventQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											eventQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												eventQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												eventQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												eventQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											eventQueryMapArray = append(eventQueryMapArray, eventQueryMapArrayIntf)
										}

										eventQueryMap["group_by"] = eventQueryMapArray
									}
									if v, ok := eventQueryDDModel.GetIndexOk(); ok {
										eventQueryMap["index"] = *v
									}
									if multiComputeArray, ok := eventQueryDDModel.GetMultiComputeOk(); ok {
										eventQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											eventQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												eventQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												eventQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												eventQueryMapArrayIntf["interval"] = *v
											}

											eventQueryMapArray = append(eventQueryMapArray, eventQueryMapArrayIntf)
										}

										eventQueryMap["multi_compute"] = eventQueryMapArray
									}
									if searchDDModel, ok := eventQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										eventQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["event_query"] = []map[string]interface{}{eventQueryMap}
								}
								if logQueryDDModel, ok := requestsArrayItem.GetLogQueryOk(); ok {
									logQueryMap := map[string]interface{}{}
									if computeDDModel, ok := logQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										logQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := logQueryDDModel.GetGroupByOk(); ok {
										logQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											logQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												logQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												logQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												logQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											logQueryMapArray = append(logQueryMapArray, logQueryMapArrayIntf)
										}

										logQueryMap["group_by"] = logQueryMapArray
									}
									if v, ok := logQueryDDModel.GetIndexOk(); ok {
										logQueryMap["index"] = *v
									}
									if multiComputeArray, ok := logQueryDDModel.GetMultiComputeOk(); ok {
										logQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											logQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												logQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												logQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												logQueryMapArrayIntf["interval"] = *v
											}

											logQueryMapArray = append(logQueryMapArray, logQueryMapArrayIntf)
										}

										logQueryMap["multi_compute"] = logQueryMapArray
									}
									if searchDDModel, ok := logQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										logQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["log_query"] = []map[string]interface{}{logQueryMap}
								}
								if networkQueryDDModel, ok := requestsArrayItem.GetNetworkQueryOk(); ok {
									networkQueryMap := map[string]interface{}{}
									if computeDDModel, ok := networkQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										networkQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := networkQueryDDModel.GetGroupByOk(); ok {
										networkQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											networkQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												networkQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												networkQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												networkQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											networkQueryMapArray = append(networkQueryMapArray, networkQueryMapArrayIntf)
										}

										networkQueryMap["group_by"] = networkQueryMapArray
									}
									if v, ok := networkQueryDDModel.GetIndexOk(); ok {
										networkQueryMap["index"] = *v
									}
									if multiComputeArray, ok := networkQueryDDModel.GetMultiComputeOk(); ok {
										networkQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											networkQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												networkQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												networkQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												networkQueryMapArrayIntf["interval"] = *v
											}

											networkQueryMapArray = append(networkQueryMapArray, networkQueryMapArrayIntf)
										}

										networkQueryMap["multi_compute"] = networkQueryMapArray
									}
									if searchDDModel, ok := networkQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										networkQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["network_query"] = []map[string]interface{}{networkQueryMap}
								}
								if processQueryDDModel, ok := requestsArrayItem.GetProcessQueryOk(); ok {
									processQueryMap := map[string]interface{}{}
									if filterByArray, ok := processQueryDDModel.GetFilterByOk(); ok {
										processQueryMapArray := make([]string, len(*filterByArray))
										for i, item := range *filterByArray {
											processQueryMapArray[i] = item
										}

										processQueryMap["filter_by"] = processQueryMapArray
									}
									if v, ok := processQueryDDModel.GetLimitOk(); ok {
										processQueryMap["limit"] = *v
									}
									if v, ok := processQueryDDModel.GetMetricOk(); ok {
										processQueryMap["metric"] = *v
									}
									if v, ok := processQueryDDModel.GetSearchByOk(); ok {
										processQueryMap["search_by"] = *v
									}

									definitionMapArrayIntf["process_query"] = []map[string]interface{}{processQueryMap}
								}
								if profileMetricsQueryDDModel, ok := requestsArrayItem.GetProfileMetricsQueryOk(); ok {
									profileMetricsQueryMap := map[string]interface{}{}
									if computeDDModel, ok := profileMetricsQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										profileMetricsQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := profileMetricsQueryDDModel.GetGroupByOk(); ok {
										profileMetricsQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											profileMetricsQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												profileMetricsQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												profileMetricsQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												profileMetricsQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											profileMetricsQueryMapArray = append(profileMetricsQueryMapArray, profileMetricsQueryMapArrayIntf)
										}

										profileMetricsQueryMap["group_by"] = profileMetricsQueryMapArray
									}
									if v, ok := profileMetricsQueryDDModel.GetIndexOk(); ok {
										profileMetricsQueryMap["index"] = *v
									}
									if multiComputeArray, ok := profileMetricsQueryDDModel.GetMultiComputeOk(); ok {
										profileMetricsQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											profileMetricsQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												profileMetricsQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												profileMetricsQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												profileMetricsQueryMapArrayIntf["interval"] = *v
											}

											profileMetricsQueryMapArray = append(profileMetricsQueryMapArray, profileMetricsQueryMapArrayIntf)
										}

										profileMetricsQueryMap["multi_compute"] = profileMetricsQueryMapArray
									}
									if searchDDModel, ok := profileMetricsQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										profileMetricsQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["profile_metrics_query"] = []map[string]interface{}{profileMetricsQueryMap}
								}
								if v, ok := requestsArrayItem.GetQOk(); ok {
									definitionMapArrayIntf["q"] = *v
								}
								if rumQueryDDModel, ok := requestsArrayItem.GetRumQueryOk(); ok {
									rumQueryMap := map[string]interface{}{}
									if computeDDModel, ok := rumQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										rumQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := rumQueryDDModel.GetGroupByOk(); ok {
										rumQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											rumQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												rumQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												rumQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												rumQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											rumQueryMapArray = append(rumQueryMapArray, rumQueryMapArrayIntf)
										}

										rumQueryMap["group_by"] = rumQueryMapArray
									}
									if v, ok := rumQueryDDModel.GetIndexOk(); ok {
										rumQueryMap["index"] = *v
									}
									if multiComputeArray, ok := rumQueryDDModel.GetMultiComputeOk(); ok {
										rumQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											rumQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												rumQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												rumQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												rumQueryMapArrayIntf["interval"] = *v
											}

											rumQueryMapArray = append(rumQueryMapArray, rumQueryMapArrayIntf)
										}

										rumQueryMap["multi_compute"] = rumQueryMapArray
									}
									if searchDDModel, ok := rumQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										rumQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["rum_query"] = []map[string]interface{}{rumQueryMap}
								}
								if securityQueryDDModel, ok := requestsArrayItem.GetSecurityQueryOk(); ok {
									securityQueryMap := map[string]interface{}{}
									if computeDDModel, ok := securityQueryDDModel.GetComputeOk(); ok {
										computeMap := map[string]interface{}{}
										if v, ok := computeDDModel.GetAggregationOk(); ok {
											computeMap["aggregation"] = *v
										}
										if v, ok := computeDDModel.GetFacetOk(); ok {
											computeMap["facet"] = *v
										}
										if v, ok := computeDDModel.GetIntervalOk(); ok {
											computeMap["interval"] = *v
										}

										securityQueryMap["compute"] = []map[string]interface{}{computeMap}
									}
									if groupByArray, ok := securityQueryDDModel.GetGroupByOk(); ok {
										securityQueryMapArray := make([]map[string]interface{}, 0)

										for _, groupByArrayItem := range *groupByArray {
											securityQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := groupByArrayItem.GetFacetOk(); ok {
												securityQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := groupByArrayItem.GetLimitOk(); ok {
												securityQueryMapArrayIntf["limit"] = *v
											}
											if sortDDModel, ok := groupByArrayItem.GetSortOk(); ok {
												sortMap := map[string]interface{}{}
												if v, ok := sortDDModel.GetAggregationOk(); ok {
													sortMap["aggregation"] = *v
												}
												if v, ok := sortDDModel.GetFacetOk(); ok {
													sortMap["facet"] = *v
												}
												if v, ok := sortDDModel.GetOrderOk(); ok {
													sortMap["order"] = *v
												}

												securityQueryMapArrayIntf["sort"] = []map[string]interface{}{sortMap}
											}

											securityQueryMapArray = append(securityQueryMapArray, securityQueryMapArrayIntf)
										}

										securityQueryMap["group_by"] = securityQueryMapArray
									}
									if v, ok := securityQueryDDModel.GetIndexOk(); ok {
										securityQueryMap["index"] = *v
									}
									if multiComputeArray, ok := securityQueryDDModel.GetMultiComputeOk(); ok {
										securityQueryMapArray := make([]map[string]interface{}, 0)

										for _, multiComputeArrayItem := range *multiComputeArray {
											securityQueryMapArrayIntf := map[string]interface{}{}
											if v, ok := multiComputeArrayItem.GetAggregationOk(); ok {
												securityQueryMapArrayIntf["aggregation"] = *v
											}
											if v, ok := multiComputeArrayItem.GetFacetOk(); ok {
												securityQueryMapArrayIntf["facet"] = *v
											}
											if v, ok := multiComputeArrayItem.GetIntervalOk(); ok {
												securityQueryMapArrayIntf["interval"] = *v
											}

											securityQueryMapArray = append(securityQueryMapArray, securityQueryMapArrayIntf)
										}

										securityQueryMap["multi_compute"] = securityQueryMapArray
									}
									if searchDDModel, ok := securityQueryDDModel.GetSearchOk(); ok {
										searchMap := map[string]interface{}{}
										if v, ok := searchDDModel.GetQueryOk(); ok {
											searchMap["query"] = *v
										}

										securityQueryMap["search"] = []map[string]interface{}{searchMap}
									}

									definitionMapArrayIntf["security_query"] = []map[string]interface{}{securityQueryMap}
								}
								if styleDDModel, ok := requestsArrayItem.GetStyleOk(); ok {
									styleMap := map[string]interface{}{}
									if v, ok := styleDDModel.GetPaletteOk(); ok {
										styleMap["palette"] = *v
									}

									definitionMapArrayIntf["style"] = []map[string]interface{}{styleMap}
								}

								definitionMapArray = append(definitionMapArray, definitionMapArrayIntf)
							}

							definitionMap["request"] = definitionMapArray
						}
						if v, ok := definitionDDModel.GetShowLegendOk(); ok {
							definitionMap["show_legend"] = *v
						}
						if timeDDModel, ok := definitionDDModel.GetTimeOk(); ok {
							timeMap := map[string]interface{}{}
							if v, ok := timeDDModel.GetLiveSpanOk(); ok {
								timeMap["live_span"] = *v
							}

							definitionMap["time"] = []map[string]interface{}{timeMap}
						}
						if v, ok := definitionDDModel.GetTitleOk(); ok {
							definitionMap["title"] = *v
						}
						if v, ok := definitionDDModel.GetTitleAlignOk(); ok {
							definitionMap["title_align"] = *v
						}
						if v, ok := definitionDDModel.GetTitleSizeOk(); ok {
							definitionMap["title_size"] = *v
						}
						if v, ok := definitionDDModel.GetTypeOk(); ok {
							definitionMap["type"] = *v
						}
						if xaxisDDModel, ok := definitionDDModel.GetXaxisOk(); ok {
							xaxisMap := map[string]interface{}{}
							if v, ok := xaxisDDModel.GetIncludeZeroOk(); ok {
								xaxisMap["include_zero"] = *v
							}
							if v, ok := xaxisDDModel.GetMaxOk(); ok {
								xaxisMap["max"] = *v
							}
							if v, ok := xaxisDDModel.GetMinOk(); ok {
								xaxisMap["min"] = *v
							}
							if v, ok := xaxisDDModel.GetScaleOk(); ok {
								xaxisMap["scale"] = *v
							}

							definitionMap["xaxis"] = []map[string]interface{}{xaxisMap}
						}
						if yaxisDDModel, ok := definitionDDModel.GetYaxisOk(); ok {
							yaxisMap := map[string]interface{}{}
							if v, ok := yaxisDDModel.GetIncludeZeroOk(); ok {
								yaxisMap["include_zero"] = *v
							}
							if v, ok := yaxisDDModel.GetLabelOk(); ok {
								yaxisMap["label"] = *v
							}
							if v, ok := yaxisDDModel.GetMaxOk(); ok {
								yaxisMap["max"] = *v
							}
							if v, ok := yaxisDDModel.GetMinOk(); ok {
								yaxisMap["min"] = *v
							}
							if v, ok := yaxisDDModel.GetScaleOk(); ok {
								yaxisMap["scale"] = *v
							}

							definitionMap["yaxis"] = []map[string]interface{}{yaxisMap}
						}

						mapNotebookDistributionCellAttributes["definition"] = []map[string]interface{}{definitionMap}
					}
					if v, ok := ddNotebookDistributionCellAttributes.GetGraphSizeOk(); ok {
						mapNotebookDistributionCellAttributes["graph_size"] = *v
					}
					if splitByDDModel, ok := ddNotebookDistributionCellAttributes.GetSplitByOk(); ok {
						splitByMap := map[string]interface{}{}
						if keysArray, ok := splitByDDModel.GetKeysOk(); ok {
							splitByMapArray := make([]string, len(*keysArray))
							for i, item := range *keysArray {
								splitByMapArray[i] = item
							}

							splitByMap["keys"] = splitByMapArray
						}
						if tagsArray, ok := splitByDDModel.GetTagsOk(); ok {
							splitByMapArray := make([]string, len(*tagsArray))
							for i, item := range *tagsArray {
								splitByMapArray[i] = item
							}

							splitByMap["tags"] = splitByMapArray
						}

						mapNotebookDistributionCellAttributes["split_by"] = []map[string]interface{}{splitByMap}
					}
					if timeDDModel, ok := ddNotebookDistributionCellAttributes.GetTimeOk(); ok {
						timeMap := map[string]interface{}{}
						if ddNotebookRelativeTime := timeDDModel.NotebookRelativeTime; ddNotebookRelativeTime != nil {
							mapNotebookRelativeTime := map[string]interface{}{}
							if v, ok := ddNotebookRelativeTime.GetLiveSpanOk(); ok {
								mapNotebookRelativeTime["live_span"] = *v
							}

							arrayNotebookRelativeTime := []interface{}{mapNotebookRelativeTime}
							timeMap["notebook_relative_time"] = arrayNotebookRelativeTime
						}
						if ddNotebookAbsoluteTime := timeDDModel.NotebookAbsoluteTime; ddNotebookAbsoluteTime != nil {
							mapNotebookAbsoluteTime := map[string]interface{}{}
							if v, ok := ddNotebookAbsoluteTime.GetEndOk(); ok {
								mapNotebookAbsoluteTime["end"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}
							if v, ok := ddNotebookAbsoluteTime.GetLiveOk(); ok {
								mapNotebookAbsoluteTime["live"] = *v
							}
							if v, ok := ddNotebookAbsoluteTime.GetStartOk(); ok {
								mapNotebookAbsoluteTime["start"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}

							arrayNotebookAbsoluteTime := []interface{}{mapNotebookAbsoluteTime}
							timeMap["notebook_absolute_time"] = arrayNotebookAbsoluteTime
						}

						mapNotebookDistributionCellAttributes["time"] = []map[string]interface{}{timeMap}
					}

					arrayNotebookDistributionCellAttributes := []interface{}{mapNotebookDistributionCellAttributes}
					attributesMap["notebook_distribution_cell_attributes"] = arrayNotebookDistributionCellAttributes
				}
				if ddNotebookLogStreamCellAttributes := attributesDDModel.NotebookLogStreamCellAttributes; ddNotebookLogStreamCellAttributes != nil {
					mapNotebookLogStreamCellAttributes := map[string]interface{}{}
					if definitionDDModel, ok := ddNotebookLogStreamCellAttributes.GetDefinitionOk(); ok {
						definitionMap := map[string]interface{}{}
						if columnsArray, ok := definitionDDModel.GetColumnsOk(); ok {
							definitionMapArray := make([]string, len(*columnsArray))
							for i, item := range *columnsArray {
								definitionMapArray[i] = item
							}

							definitionMap["columns"] = definitionMapArray
						}
						if indexesArray, ok := definitionDDModel.GetIndexesOk(); ok {
							definitionMapArray := make([]string, len(*indexesArray))
							for i, item := range *indexesArray {
								definitionMapArray[i] = item
							}

							definitionMap["indexes"] = definitionMapArray
						}
						if v, ok := definitionDDModel.GetLogsetOk(); ok {
							definitionMap["logset"] = *v
						}
						if v, ok := definitionDDModel.GetMessageDisplayOk(); ok {
							definitionMap["message_display"] = *v
						}
						if v, ok := definitionDDModel.GetQueryOk(); ok {
							definitionMap["query"] = *v
						}
						if v, ok := definitionDDModel.GetShowDateColumnOk(); ok {
							definitionMap["show_date_column"] = *v
						}
						if v, ok := definitionDDModel.GetShowMessageColumnOk(); ok {
							definitionMap["show_message_column"] = *v
						}
						if sortDDModel, ok := definitionDDModel.GetSortOk(); ok {
							sortMap := map[string]interface{}{}
							if v, ok := sortDDModel.GetColumnOk(); ok {
								sortMap["column"] = *v
							}
							if v, ok := sortDDModel.GetOrderOk(); ok {
								sortMap["order"] = *v
							}

							definitionMap["sort"] = []map[string]interface{}{sortMap}
						}
						if timeDDModel, ok := definitionDDModel.GetTimeOk(); ok {
							timeMap := map[string]interface{}{}
							if v, ok := timeDDModel.GetLiveSpanOk(); ok {
								timeMap["live_span"] = *v
							}

							definitionMap["time"] = []map[string]interface{}{timeMap}
						}
						if v, ok := definitionDDModel.GetTitleOk(); ok {
							definitionMap["title"] = *v
						}
						if v, ok := definitionDDModel.GetTitleAlignOk(); ok {
							definitionMap["title_align"] = *v
						}
						if v, ok := definitionDDModel.GetTitleSizeOk(); ok {
							definitionMap["title_size"] = *v
						}
						if v, ok := definitionDDModel.GetTypeOk(); ok {
							definitionMap["type"] = *v
						}

						mapNotebookLogStreamCellAttributes["definition"] = []map[string]interface{}{definitionMap}
					}
					if v, ok := ddNotebookLogStreamCellAttributes.GetGraphSizeOk(); ok {
						mapNotebookLogStreamCellAttributes["graph_size"] = *v
					}
					if timeDDModel, ok := ddNotebookLogStreamCellAttributes.GetTimeOk(); ok {
						timeMap := map[string]interface{}{}
						if ddNotebookRelativeTime := timeDDModel.NotebookRelativeTime; ddNotebookRelativeTime != nil {
							mapNotebookRelativeTime := map[string]interface{}{}
							if v, ok := ddNotebookRelativeTime.GetLiveSpanOk(); ok {
								mapNotebookRelativeTime["live_span"] = *v
							}

							arrayNotebookRelativeTime := []interface{}{mapNotebookRelativeTime}
							timeMap["notebook_relative_time"] = arrayNotebookRelativeTime
						}
						if ddNotebookAbsoluteTime := timeDDModel.NotebookAbsoluteTime; ddNotebookAbsoluteTime != nil {
							mapNotebookAbsoluteTime := map[string]interface{}{}
							if v, ok := ddNotebookAbsoluteTime.GetEndOk(); ok {
								mapNotebookAbsoluteTime["end"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}
							if v, ok := ddNotebookAbsoluteTime.GetLiveOk(); ok {
								mapNotebookAbsoluteTime["live"] = *v
							}
							if v, ok := ddNotebookAbsoluteTime.GetStartOk(); ok {
								mapNotebookAbsoluteTime["start"] = func(t *time.Time) *string {
									if t != nil {
										r := t.Format("2006-01-02T15:04:05.000000-0700")
										return &r
									}
									return nil
								}(v)
							}

							arrayNotebookAbsoluteTime := []interface{}{mapNotebookAbsoluteTime}
							timeMap["notebook_absolute_time"] = arrayNotebookAbsoluteTime
						}

						mapNotebookLogStreamCellAttributes["time"] = []map[string]interface{}{timeMap}
					}

					arrayNotebookLogStreamCellAttributes := []interface{}{mapNotebookLogStreamCellAttributes}
					attributesMap["notebook_log_stream_cell_attributes"] = arrayNotebookLogStreamCellAttributes
				}

				cellsTFArrayIntf["attributes"] = []map[string]interface{}{attributesMap}
			}
			if v, ok := arrayItem.GetIdOk(); ok {
				cellsTFArrayIntf["id"] = *v
			}
			if v, ok := arrayItem.GetTypeOk(); ok {
				cellsTFArrayIntf["type"] = *v
			}

			cellsTFArray = append(cellsTFArray, cellsTFArrayIntf)
		}

		err = d.Set("cell", cellsTFArray)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := resource.GetCreatedOk(); ok {
		err = d.Set("created", func(t *time.Time) *string {
			if t != nil {
				r := t.Format("2006-01-02T15:04:05.000000-0700")
				return &r
			}
			return nil
		}(v))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := resource.GetModifiedOk(); ok {
		err = d.Set("modified", func(t *time.Time) *string {
			if t != nil {
				r := t.Format("2006-01-02T15:04:05.000000-0700")
				return &r
			}
			return nil
		}(v))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := resource.GetNameOk(); ok {
		err = d.Set("name", *v)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := resource.GetStatusOk(); ok {
		err = d.Set("status", *v)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if ddTime, ok := resource.GetTimeOk(); ok {
		mapTime := map[string]interface{}{}
		if ddNotebookRelativeTime := ddTime.NotebookRelativeTime; ddNotebookRelativeTime != nil {
			mapNotebookRelativeTime := map[string]interface{}{}
			if v, ok := ddNotebookRelativeTime.GetLiveSpanOk(); ok {
				mapNotebookRelativeTime["live_span"] = *v
			}

			arrayNotebookRelativeTime := []interface{}{mapNotebookRelativeTime}
			mapTime["notebook_relative_time"] = arrayNotebookRelativeTime
		}
		if ddNotebookAbsoluteTime := ddTime.NotebookAbsoluteTime; ddNotebookAbsoluteTime != nil {
			mapNotebookAbsoluteTime := map[string]interface{}{}
			if v, ok := ddNotebookAbsoluteTime.GetEndOk(); ok {
				mapNotebookAbsoluteTime["end"] = func(t *time.Time) *string {
					if t != nil {
						r := t.Format("2006-01-02T15:04:05.000000-0700")
						return &r
					}
					return nil
				}(v)
			}
			if v, ok := ddNotebookAbsoluteTime.GetLiveOk(); ok {
				mapNotebookAbsoluteTime["live"] = *v
			}
			if v, ok := ddNotebookAbsoluteTime.GetStartOk(); ok {
				mapNotebookAbsoluteTime["start"] = func(t *time.Time) *string {
					if t != nil {
						r := t.Format("2006-01-02T15:04:05.000000-0700")
						return &r
					}
					return nil
				}(v)
			}

			arrayNotebookAbsoluteTime := []interface{}{mapNotebookAbsoluteTime}
			mapTime["notebook_absolute_time"] = arrayNotebookAbsoluteTime
		}

		arrayTime := []interface{}{mapTime}
		err = d.Set("time", arrayTime)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceDatadogNotebookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1
	var err error

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceNotebookResponse, httpResp, err := datadogClient.NotebooksApi.GetNotebook(auth, id)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// this condition takes on the job of the deprecated Exists handlers
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, "error reading Notebook")
	}

	resourceNotebookResponseData := resourceNotebookResponse.GetData()

	resource := resourceNotebookResponseData.GetAttributes()

	return updateNotebookTerraformState(d, resource)
}

func buildDatadogNotebookUpdate(d *schema.ResourceData) (*datadogV1.NotebookUpdateDataAttributes, error) {
	k := utils.NewResourceDataKey(d, "")
	result := datadogV1.NewNotebookUpdateDataAttributesWithDefaults()
	k.Add("cell")
	if cellsArray, ok := k.GetOk(); ok {
		cellsDDArray := make([]datadogV1.NotebookUpdateCell, 0)
		for i := range cellsArray.([]interface{}) {
			k.Add(i)

			cellsDDArrayItem := &datadogV1.NotebookUpdateCell{}
			// handle notebook_update_cell, which is a oneOf model
			k.Add("notebook_cell_create_request.0")
			if _, ok := k.GetOk(); ok {

				ddNotebookCellCreateRequest := datadogV1.NewNotebookCellCreateRequestWithDefaults()

				// handle attributes, which is a nested model
				k.Add("attributes.0")

				ddNotebookCellCreateRequestAttributes := &datadogV1.NotebookCellCreateRequestAttributes{}
				// handle notebook_cell_create_request_attributes, which is a oneOf model
				k.Add("notebook_markdown_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookMarkdownCellAttributes := datadogV1.NewNotebookMarkdownCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookMarkdownCellAttributesDefinition := datadogV1.NewNotebookMarkdownCellDefinitionWithDefaults()

					if v, ok := k.GetOkWith("text"); ok {
						ddNotebookMarkdownCellAttributesDefinition.SetText(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookMarkdownCellAttributesDefinition.SetType(datadogV1.NotebookMarkdownCellDefinitionType(v.(string)))
					}
					k.Remove("definition.0")
					ddNotebookMarkdownCellAttributes.SetDefinition(*ddNotebookMarkdownCellAttributesDefinition)
					ddNotebookCellCreateRequestAttributes.NotebookMarkdownCellAttributes = ddNotebookMarkdownCellAttributes
				}
				k.Remove("notebook_markdown_cell_attributes.0")
				k.Add("notebook_timeseries_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookTimeseriesCellAttributes := datadogV1.NewNotebookTimeseriesCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookTimeseriesCellAttributesDefinition := datadogV1.NewTimeseriesWidgetDefinitionWithDefaults()
					k.Add("custom_link")
					if customLinksArray, ok := k.GetOk(); ok {
						customLinksDDArray := make([]datadogV1.WidgetCustomLink, 0)
						for i := range customLinksArray.([]interface{}) {
							k.Add(i)

							customLinksDDArrayItem := datadogV1.NewWidgetCustomLinkWithDefaults()

							if v, ok := k.GetOkWith("is_hidden"); ok {
								customLinksDDArrayItem.SetIsHidden(v.(bool))
							}

							if v, ok := k.GetOkWith("label"); ok {
								customLinksDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("link"); ok {
								customLinksDDArrayItem.SetLink(v.(string))
							}

							if v, ok := k.GetOkWith("override_label"); ok {
								customLinksDDArrayItem.SetOverrideLabel(v.(string))
							}
							customLinksDDArray = append(customLinksDDArray, *customLinksDDArrayItem)
							k.Remove(i)
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetCustomLinks(customLinksDDArray)
					}
					k.Remove("custom_link")
					k.Add("event")
					if eventsArray, ok := k.GetOk(); ok {
						eventsDDArray := make([]datadogV1.WidgetEvent, 0)
						for i := range eventsArray.([]interface{}) {
							k.Add(i)

							eventsDDArrayItem := datadogV1.NewWidgetEventWithDefaults()

							if v, ok := k.GetOkWith("q"); ok {
								eventsDDArrayItem.SetQ(v.(string))
							}

							if v, ok := k.GetOkWith("tags_execution"); ok {
								eventsDDArrayItem.SetTagsExecution(v.(string))
							}
							eventsDDArray = append(eventsDDArray, *eventsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetEvents(eventsDDArray)
					}
					k.Remove("event")
					k.Add("legend_columns")
					if legendColumnsArray, ok := k.GetOk(); ok {
						legendColumnsDDArray := make([]datadogV1.TimeseriesWidgetLegendColumn, 0)
						for i := range legendColumnsArray.([]interface{}) {
							legendColumnsArrayItem := k.GetWith(i)
							legendColumnsDDArray = append(legendColumnsDDArray, datadogV1.TimeseriesWidgetLegendColumn(legendColumnsArrayItem.(string)))
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetLegendColumns(legendColumnsDDArray)
					}
					k.Remove("legend_columns")

					if v, ok := k.GetOkWith("legend_layout"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetLegendLayout(datadogV1.TimeseriesWidgetLegendLayout(v.(string)))
					}

					if v, ok := k.GetOkWith("legend_size"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetLegendSize(v.(string))
					}
					k.Add("marker")
					if markersArray, ok := k.GetOk(); ok {
						markersDDArray := make([]datadogV1.WidgetMarker, 0)
						for i := range markersArray.([]interface{}) {
							k.Add(i)

							markersDDArrayItem := datadogV1.NewWidgetMarkerWithDefaults()

							if v, ok := k.GetOkWith("display_type"); ok {
								markersDDArrayItem.SetDisplayType(v.(string))
							}

							if v, ok := k.GetOkWith("label"); ok {
								markersDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("time"); ok {
								markersDDArrayItem.SetTime(v.(string))
							}

							if v, ok := k.GetOkWith("value"); ok {
								markersDDArrayItem.SetValue(v.(string))
							}
							markersDDArray = append(markersDDArray, *markersDDArrayItem)
							k.Remove(i)
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetMarkers(markersDDArray)
					}
					k.Remove("marker")
					k.Add("request")
					if requestsArray, ok := k.GetOk(); ok {
						requestsDDArray := make([]datadogV1.TimeseriesWidgetRequest, 0)
						for i := range requestsArray.([]interface{}) {
							k.Add(i)

							requestsDDArrayItem := datadogV1.NewTimeseriesWidgetRequestWithDefaults()

							// handle apm_query, which is a nested model
							k.Add("apm_query.0")

							requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemApmQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
							k.Remove("apm_query.0")
							requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)

							if v, ok := k.GetOkWith("display_type"); ok {
								requestsDDArrayItem.SetDisplayType(datadogV1.WidgetDisplayType(v.(string)))
							}

							// handle event_query, which is a nested model
							k.Add("event_query.0")

							requestsDDArrayItemEventQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemEventQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemEventQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemEventQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemEventQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemEventQuery.SetCompute(*requestsDDArrayItemEventQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemEventQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemEventQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemEventQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemEventQuery.SetSearch(*requestsDDArrayItemEventQuerySearch)
							k.Remove("event_query.0")
							requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)
							k.Add("formula")
							if formulasArray, ok := k.GetOk(); ok {
								formulasDDArray := make([]datadogV1.WidgetFormula, 0)
								for i := range formulasArray.([]interface{}) {
									k.Add(i)

									formulasDDArrayItem := datadogV1.NewWidgetFormulaWithDefaults()

									if v, ok := k.GetOkWith("alias"); ok {
										formulasDDArrayItem.SetAlias(v.(string))
									}

									if v, ok := k.GetOkWith("formula"); ok {
										formulasDDArrayItem.SetFormula(v.(string))
									}

									// handle limit, which is a nested model
									k.Add("limit.0")

									formulasDDArrayItemLimit := datadogV1.NewWidgetFormulaLimitWithDefaults()

									if v, ok := k.GetOkWith("count"); ok {
										formulasDDArrayItemLimit.SetCount(int64(v.(int)))
									}

									if v, ok := k.GetOkWith("order"); ok {
										formulasDDArrayItemLimit.SetOrder(datadogV1.QuerySortOrder(v.(string)))
									}
									k.Remove("limit.0")
									formulasDDArrayItem.SetLimit(*formulasDDArrayItemLimit)
									formulasDDArray = append(formulasDDArray, *formulasDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetFormulas(formulasDDArray)
							}
							k.Remove("formula")

							// handle log_query, which is a nested model
							k.Add("log_query.0")

							requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemLogQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
							k.Remove("log_query.0")
							requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)
							k.Add("metadata")
							if metadataArray, ok := k.GetOk(); ok {
								metadataDDArray := make([]datadogV1.TimeseriesWidgetExpressionAlias, 0)
								for i := range metadataArray.([]interface{}) {
									k.Add(i)

									metadataDDArrayItem := datadogV1.NewTimeseriesWidgetExpressionAliasWithDefaults()

									if v, ok := k.GetOkWith("alias_name"); ok {
										metadataDDArrayItem.SetAliasName(v.(string))
									}

									if v, ok := k.GetOkWith("expression"); ok {
										metadataDDArrayItem.SetExpression(v.(string))
									}
									metadataDDArray = append(metadataDDArray, *metadataDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetMetadata(metadataDDArray)
							}
							k.Remove("metadata")

							// handle network_query, which is a nested model
							k.Add("network_query.0")

							requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
							k.Remove("network_query.0")
							requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

							if v, ok := k.GetOkWith("on_right_yaxis"); ok {
								requestsDDArrayItem.SetOnRightYaxis(v.(bool))
							}

							// handle process_query, which is a nested model
							k.Add("process_query.0")

							requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
							k.Add("filter_by")
							if filterByArray, ok := k.GetOk(); ok {
								filterByDDArray := make([]string, 0)
								for i := range filterByArray.([]interface{}) {
									filterByArrayItem := k.GetWith(i)
									filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
								}
								requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
							}
							k.Remove("filter_by")

							if v, ok := k.GetOkWith("limit"); ok {
								requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
							}

							if v, ok := k.GetOkWith("metric"); ok {
								requestsDDArrayItemProcessQuery.SetMetric(v.(string))
							}

							if v, ok := k.GetOkWith("search_by"); ok {
								requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
							}
							k.Remove("process_query.0")
							requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

							// handle profile_metrics_query, which is a nested model
							k.Add("profile_metrics_query.0")

							requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
							k.Remove("profile_metrics_query.0")
							requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

							if v, ok := k.GetOkWith("q"); ok {
								requestsDDArrayItem.SetQ(v.(string))
							}
							k.Add("query")
							if queriesArray, ok := k.GetOk(); ok {
								queriesDDArray := make([]datadogV1.FormulaAndFunctionQueryDefinition, 0)
								for i := range queriesArray.([]interface{}) {
									k.Add(i)

									queriesDDArrayItem := &datadogV1.FormulaAndFunctionQueryDefinition{}
									// handle formula_and_function_query_definition, which is a oneOf model
									k.Add("formula_and_function_metric_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionMetricQueryDefinition := datadogV1.NewFormulaAndFunctionMetricQueryDefinitionWithDefaults()

										if v, ok := k.GetOkWith("aggregator"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionMetricDataSource(v.(string)))
										}

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetName(v.(string))
										}

										if v, ok := k.GetOkWith("query"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetQuery(v.(string))
										}
										queriesDDArrayItem.FormulaAndFunctionMetricQueryDefinition = ddFormulaAndFunctionMetricQueryDefinition
									}
									k.Remove("formula_and_function_metric_query_definition.0")
									k.Add("formula_and_function_event_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionEventQueryDefinition := datadogV1.NewFormulaAndFunctionEventQueryDefinitionWithDefaults()

										// handle compute, which is a nested model
										k.Add("compute.0")

										ddFormulaAndFunctionEventQueryDefinitionCompute := datadogV1.NewFormulaAndFunctionEventQueryDefinitionComputeWithDefaults()

										if v, ok := k.GetOkWith("aggregation"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("interval"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetInterval(int64(v.(int)))
										}

										if v, ok := k.GetOkWith("metric"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetMetric(v.(string))
										}
										k.Remove("compute.0")
										ddFormulaAndFunctionEventQueryDefinition.SetCompute(*ddFormulaAndFunctionEventQueryDefinitionCompute)

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionEventQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionEventsDataSource(v.(string)))
										}
										k.Add("group_by")
										if groupByArray, ok := k.GetOk(); ok {
											groupByDDArray := make([]datadogV1.FormulaAndFunctionEventQueryGroupBy, 0)
											for i := range groupByArray.([]interface{}) {
												k.Add(i)

												groupByDDArrayItem := datadogV1.NewFormulaAndFunctionEventQueryGroupByWithDefaults()

												if v, ok := k.GetOkWith("facet"); ok {
													groupByDDArrayItem.SetFacet(v.(string))
												}

												if v, ok := k.GetOkWith("limit"); ok {
													groupByDDArrayItem.SetLimit(int64(v.(int)))
												}

												// handle sort, which is a nested model
												k.Add("sort.0")

												groupByDDArrayItemSort := datadogV1.NewFormulaAndFunctionEventQueryGroupBySortWithDefaults()

												if v, ok := k.GetOkWith("aggregation"); ok {
													groupByDDArrayItemSort.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
												}

												if v, ok := k.GetOkWith("metric"); ok {
													groupByDDArrayItemSort.SetMetric(v.(string))
												}

												if v, ok := k.GetOkWith("order"); ok {
													groupByDDArrayItemSort.SetOrder(datadogV1.QuerySortOrder(v.(string)))
												}
												k.Remove("sort.0")
												groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
												groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
												k.Remove(i)
											}
											ddFormulaAndFunctionEventQueryDefinition.SetGroupBy(groupByDDArray)
										}
										k.Remove("group_by")
										k.Add("indexes")
										if indexesArray, ok := k.GetOk(); ok {
											indexesDDArray := make([]string, 0)
											for i := range indexesArray.([]interface{}) {
												indexesArrayItem := k.GetWith(i)
												indexesDDArray = append(indexesDDArray, indexesArrayItem.(string))
											}
											ddFormulaAndFunctionEventQueryDefinition.SetIndexes(indexesDDArray)
										}
										k.Remove("indexes")

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionEventQueryDefinition.SetName(v.(string))
										}

										// handle search, which is a nested model
										k.Add("search.0")

										ddFormulaAndFunctionEventQueryDefinitionSearch := datadogV1.NewFormulaAndFunctionEventQueryDefinitionSearchWithDefaults()

										if v, ok := k.GetOkWith("query"); ok {
											ddFormulaAndFunctionEventQueryDefinitionSearch.SetQuery(v.(string))
										}
										k.Remove("search.0")
										ddFormulaAndFunctionEventQueryDefinition.SetSearch(*ddFormulaAndFunctionEventQueryDefinitionSearch)
										queriesDDArrayItem.FormulaAndFunctionEventQueryDefinition = ddFormulaAndFunctionEventQueryDefinition
									}
									k.Remove("formula_and_function_event_query_definition.0")
									k.Add("formula_and_function_process_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionProcessQueryDefinition := datadogV1.NewFormulaAndFunctionProcessQueryDefinitionWithDefaults()

										if v, ok := k.GetOkWith("aggregator"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionProcessQueryDataSource(v.(string)))
										}

										if v, ok := k.GetOkWith("is_normalized_cpu"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetIsNormalizedCpu(v.(bool))
										}

										if v, ok := k.GetOkWith("limit"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetLimit(int64(v.(int)))
										}

										if v, ok := k.GetOkWith("metric"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetMetric(v.(string))
										}

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetName(v.(string))
										}

										if v, ok := k.GetOkWith("sort"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetSort(datadogV1.QuerySortOrder(v.(string)))
										}
										k.Add("tag_filters")
										if tagFiltersArray, ok := k.GetOk(); ok {
											tagFiltersDDArray := make([]string, 0)
											for i := range tagFiltersArray.([]interface{}) {
												tagFiltersArrayItem := k.GetWith(i)
												tagFiltersDDArray = append(tagFiltersDDArray, tagFiltersArrayItem.(string))
											}
											ddFormulaAndFunctionProcessQueryDefinition.SetTagFilters(tagFiltersDDArray)
										}
										k.Remove("tag_filters")

										if v, ok := k.GetOkWith("text_filter"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetTextFilter(v.(string))
										}
										queriesDDArrayItem.FormulaAndFunctionProcessQueryDefinition = ddFormulaAndFunctionProcessQueryDefinition
									}
									k.Remove("formula_and_function_process_query_definition.0")

									if queriesDDArrayItem.GetActualInstance() == nil {
										return nil, fmt.Errorf("failed to find valid definition in formula_and_function_query_definition configuration")
									}
									queriesDDArray = append(queriesDDArray, *queriesDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetQueries(queriesDDArray)
							}
							k.Remove("query")

							if v, ok := k.GetOkWith("response_format"); ok {
								requestsDDArrayItem.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat(v.(string)))
							}

							// handle rum_query, which is a nested model
							k.Add("rum_query.0")

							requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemRumQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
							k.Remove("rum_query.0")
							requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

							// handle security_query, which is a nested model
							k.Add("security_query.0")

							requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
							k.Remove("security_query.0")
							requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

							// handle style, which is a nested model
							k.Add("style.0")

							requestsDDArrayItemStyle := datadogV1.NewWidgetRequestStyleWithDefaults()

							if v, ok := k.GetOkWith("line_type"); ok {
								requestsDDArrayItemStyle.SetLineType(datadogV1.WidgetLineType(v.(string)))
							}

							if v, ok := k.GetOkWith("line_width"); ok {
								requestsDDArrayItemStyle.SetLineWidth(datadogV1.WidgetLineWidth(v.(string)))
							}

							if v, ok := k.GetOkWith("palette"); ok {
								requestsDDArrayItemStyle.SetPalette(v.(string))
							}
							k.Remove("style.0")
							requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
							requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetRequests(requestsDDArray)
					}
					k.Remove("request")

					// handle right_yaxis, which is a nested model
					k.Add("right_yaxis.0")

					ddNotebookTimeseriesCellAttributesDefinitionRightYaxis := datadogV1.NewWidgetAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("label"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetLabel(v.(string))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetScale(v.(string))
					}
					k.Remove("right_yaxis.0")
					ddNotebookTimeseriesCellAttributesDefinition.SetRightYaxis(*ddNotebookTimeseriesCellAttributesDefinitionRightYaxis)

					if v, ok := k.GetOkWith("show_legend"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetShowLegend(v.(bool))
					}

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookTimeseriesCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookTimeseriesCellAttributesDefinition.SetTime(*ddNotebookTimeseriesCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetType(datadogV1.TimeseriesWidgetDefinitionType(v.(string)))
					}

					// handle yaxis, which is a nested model
					k.Add("yaxis.0")

					ddNotebookTimeseriesCellAttributesDefinitionYaxis := datadogV1.NewWidgetAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("label"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetLabel(v.(string))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetScale(v.(string))
					}
					k.Remove("yaxis.0")
					ddNotebookTimeseriesCellAttributesDefinition.SetYaxis(*ddNotebookTimeseriesCellAttributesDefinitionYaxis)
					k.Remove("definition.0")
					ddNotebookTimeseriesCellAttributes.SetDefinition(*ddNotebookTimeseriesCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookTimeseriesCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle split_by, which is a nested model
					k.Add("split_by.0")

					ddNotebookTimeseriesCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
					k.Add("keys")
					if keysArray, ok := k.GetOk(); ok {
						keysDDArray := make([]string, 0)
						for i := range keysArray.([]interface{}) {
							keysArrayItem := k.GetWith(i)
							keysDDArray = append(keysDDArray, keysArrayItem.(string))
						}
						ddNotebookTimeseriesCellAttributesSplitBy.SetKeys(keysDDArray)
					}
					k.Remove("keys")
					k.Add("tags")
					if tagsArray, ok := k.GetOk(); ok {
						tagsDDArray := make([]string, 0)
						for i := range tagsArray.([]interface{}) {
							tagsArrayItem := k.GetWith(i)
							tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
						}
						ddNotebookTimeseriesCellAttributesSplitBy.SetTags(tagsDDArray)
					}
					k.Remove("tags")
					k.Remove("split_by.0")
					ddNotebookTimeseriesCellAttributes.SetSplitBy(*ddNotebookTimeseriesCellAttributesSplitBy)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookTimeseriesCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookTimeseriesCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookTimeseriesCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookTimeseriesCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookTimeseriesCellAttributes.SetTime(*ddNotebookTimeseriesCellAttributesTime)
					ddNotebookCellCreateRequestAttributes.NotebookTimeseriesCellAttributes = ddNotebookTimeseriesCellAttributes
				}
				k.Remove("notebook_timeseries_cell_attributes.0")
				k.Add("notebook_toplist_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookToplistCellAttributes := datadogV1.NewNotebookToplistCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookToplistCellAttributesDefinition := datadogV1.NewToplistWidgetDefinitionWithDefaults()
					k.Add("custom_link")
					if customLinksArray, ok := k.GetOk(); ok {
						customLinksDDArray := make([]datadogV1.WidgetCustomLink, 0)
						for i := range customLinksArray.([]interface{}) {
							k.Add(i)

							customLinksDDArrayItem := datadogV1.NewWidgetCustomLinkWithDefaults()

							if v, ok := k.GetOkWith("is_hidden"); ok {
								customLinksDDArrayItem.SetIsHidden(v.(bool))
							}

							if v, ok := k.GetOkWith("label"); ok {
								customLinksDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("link"); ok {
								customLinksDDArrayItem.SetLink(v.(string))
							}

							if v, ok := k.GetOkWith("override_label"); ok {
								customLinksDDArrayItem.SetOverrideLabel(v.(string))
							}
							customLinksDDArray = append(customLinksDDArray, *customLinksDDArrayItem)
							k.Remove(i)
						}
						ddNotebookToplistCellAttributesDefinition.SetCustomLinks(customLinksDDArray)
					}
					k.Remove("custom_link")
					k.Add("request")
					if requestsArray, ok := k.GetOk(); ok {
						requestsDDArray := make([]datadogV1.ToplistWidgetRequest, 0)
						for i := range requestsArray.([]interface{}) {
							k.Add(i)

							requestsDDArrayItem := datadogV1.NewToplistWidgetRequestWithDefaults()

							// handle apm_query, which is a nested model
							k.Add("apm_query.0")

							requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemApmQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
							k.Remove("apm_query.0")
							requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)
							k.Add("conditional_format")
							if conditionalFormatsArray, ok := k.GetOk(); ok {
								conditionalFormatsDDArray := make([]datadogV1.WidgetConditionalFormat, 0)
								for i := range conditionalFormatsArray.([]interface{}) {
									k.Add(i)

									conditionalFormatsDDArrayItem := datadogV1.NewWidgetConditionalFormatWithDefaults()

									if v, ok := k.GetOkWith("comparator"); ok {
										conditionalFormatsDDArrayItem.SetComparator(datadogV1.WidgetComparator(v.(string)))
									}

									if v, ok := k.GetOkWith("custom_bg_color"); ok {
										conditionalFormatsDDArrayItem.SetCustomBgColor(v.(string))
									}

									if v, ok := k.GetOkWith("custom_fg_color"); ok {
										conditionalFormatsDDArrayItem.SetCustomFgColor(v.(string))
									}

									if v, ok := k.GetOkWith("hide_value"); ok {
										conditionalFormatsDDArrayItem.SetHideValue(v.(bool))
									}

									if v, ok := k.GetOkWith("image_url"); ok {
										conditionalFormatsDDArrayItem.SetImageUrl(v.(string))
									}

									if v, ok := k.GetOkWith("metric"); ok {
										conditionalFormatsDDArrayItem.SetMetric(v.(string))
									}

									if v, ok := k.GetOkWith("palette"); ok {
										conditionalFormatsDDArrayItem.SetPalette(datadogV1.WidgetPalette(v.(string)))
									}

									if v, ok := k.GetOkWith("timeframe"); ok {
										conditionalFormatsDDArrayItem.SetTimeframe(v.(string))
									}

									if v, ok := k.GetOkWith("value"); ok {
										conditionalFormatsDDArrayItem.SetValue(v.(float64))
									}
									conditionalFormatsDDArray = append(conditionalFormatsDDArray, *conditionalFormatsDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetConditionalFormats(conditionalFormatsDDArray)
							}
							k.Remove("conditional_format")

							// handle event_query, which is a nested model
							k.Add("event_query.0")

							requestsDDArrayItemEventQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemEventQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemEventQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemEventQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemEventQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemEventQuery.SetCompute(*requestsDDArrayItemEventQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemEventQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemEventQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemEventQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemEventQuery.SetSearch(*requestsDDArrayItemEventQuerySearch)
							k.Remove("event_query.0")
							requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)
							k.Add("formula")
							if formulasArray, ok := k.GetOk(); ok {
								formulasDDArray := make([]datadogV1.WidgetFormula, 0)
								for i := range formulasArray.([]interface{}) {
									k.Add(i)

									formulasDDArrayItem := datadogV1.NewWidgetFormulaWithDefaults()

									if v, ok := k.GetOkWith("alias"); ok {
										formulasDDArrayItem.SetAlias(v.(string))
									}

									if v, ok := k.GetOkWith("formula"); ok {
										formulasDDArrayItem.SetFormula(v.(string))
									}

									// handle limit, which is a nested model
									k.Add("limit.0")

									formulasDDArrayItemLimit := datadogV1.NewWidgetFormulaLimitWithDefaults()

									if v, ok := k.GetOkWith("count"); ok {
										formulasDDArrayItemLimit.SetCount(int64(v.(int)))
									}

									if v, ok := k.GetOkWith("order"); ok {
										formulasDDArrayItemLimit.SetOrder(datadogV1.QuerySortOrder(v.(string)))
									}
									k.Remove("limit.0")
									formulasDDArrayItem.SetLimit(*formulasDDArrayItemLimit)
									formulasDDArray = append(formulasDDArray, *formulasDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetFormulas(formulasDDArray)
							}
							k.Remove("formula")

							// handle log_query, which is a nested model
							k.Add("log_query.0")

							requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemLogQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
							k.Remove("log_query.0")
							requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)

							// handle network_query, which is a nested model
							k.Add("network_query.0")

							requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
							k.Remove("network_query.0")
							requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

							// handle process_query, which is a nested model
							k.Add("process_query.0")

							requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
							k.Add("filter_by")
							if filterByArray, ok := k.GetOk(); ok {
								filterByDDArray := make([]string, 0)
								for i := range filterByArray.([]interface{}) {
									filterByArrayItem := k.GetWith(i)
									filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
								}
								requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
							}
							k.Remove("filter_by")

							if v, ok := k.GetOkWith("limit"); ok {
								requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
							}

							if v, ok := k.GetOkWith("metric"); ok {
								requestsDDArrayItemProcessQuery.SetMetric(v.(string))
							}

							if v, ok := k.GetOkWith("search_by"); ok {
								requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
							}
							k.Remove("process_query.0")
							requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

							// handle profile_metrics_query, which is a nested model
							k.Add("profile_metrics_query.0")

							requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
							k.Remove("profile_metrics_query.0")
							requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

							if v, ok := k.GetOkWith("q"); ok {
								requestsDDArrayItem.SetQ(v.(string))
							}
							k.Add("query")
							if queriesArray, ok := k.GetOk(); ok {
								queriesDDArray := make([]datadogV1.FormulaAndFunctionQueryDefinition, 0)
								for i := range queriesArray.([]interface{}) {
									k.Add(i)

									queriesDDArrayItem := &datadogV1.FormulaAndFunctionQueryDefinition{}
									// handle formula_and_function_query_definition, which is a oneOf model
									k.Add("formula_and_function_metric_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionMetricQueryDefinition := datadogV1.NewFormulaAndFunctionMetricQueryDefinitionWithDefaults()

										if v, ok := k.GetOkWith("aggregator"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionMetricDataSource(v.(string)))
										}

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetName(v.(string))
										}

										if v, ok := k.GetOkWith("query"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetQuery(v.(string))
										}
										queriesDDArrayItem.FormulaAndFunctionMetricQueryDefinition = ddFormulaAndFunctionMetricQueryDefinition
									}
									k.Remove("formula_and_function_metric_query_definition.0")
									k.Add("formula_and_function_event_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionEventQueryDefinition := datadogV1.NewFormulaAndFunctionEventQueryDefinitionWithDefaults()

										// handle compute, which is a nested model
										k.Add("compute.0")

										ddFormulaAndFunctionEventQueryDefinitionCompute := datadogV1.NewFormulaAndFunctionEventQueryDefinitionComputeWithDefaults()

										if v, ok := k.GetOkWith("aggregation"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("interval"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetInterval(int64(v.(int)))
										}

										if v, ok := k.GetOkWith("metric"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetMetric(v.(string))
										}
										k.Remove("compute.0")
										ddFormulaAndFunctionEventQueryDefinition.SetCompute(*ddFormulaAndFunctionEventQueryDefinitionCompute)

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionEventQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionEventsDataSource(v.(string)))
										}
										k.Add("group_by")
										if groupByArray, ok := k.GetOk(); ok {
											groupByDDArray := make([]datadogV1.FormulaAndFunctionEventQueryGroupBy, 0)
											for i := range groupByArray.([]interface{}) {
												k.Add(i)

												groupByDDArrayItem := datadogV1.NewFormulaAndFunctionEventQueryGroupByWithDefaults()

												if v, ok := k.GetOkWith("facet"); ok {
													groupByDDArrayItem.SetFacet(v.(string))
												}

												if v, ok := k.GetOkWith("limit"); ok {
													groupByDDArrayItem.SetLimit(int64(v.(int)))
												}

												// handle sort, which is a nested model
												k.Add("sort.0")

												groupByDDArrayItemSort := datadogV1.NewFormulaAndFunctionEventQueryGroupBySortWithDefaults()

												if v, ok := k.GetOkWith("aggregation"); ok {
													groupByDDArrayItemSort.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
												}

												if v, ok := k.GetOkWith("metric"); ok {
													groupByDDArrayItemSort.SetMetric(v.(string))
												}

												if v, ok := k.GetOkWith("order"); ok {
													groupByDDArrayItemSort.SetOrder(datadogV1.QuerySortOrder(v.(string)))
												}
												k.Remove("sort.0")
												groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
												groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
												k.Remove(i)
											}
											ddFormulaAndFunctionEventQueryDefinition.SetGroupBy(groupByDDArray)
										}
										k.Remove("group_by")
										k.Add("indexes")
										if indexesArray, ok := k.GetOk(); ok {
											indexesDDArray := make([]string, 0)
											for i := range indexesArray.([]interface{}) {
												indexesArrayItem := k.GetWith(i)
												indexesDDArray = append(indexesDDArray, indexesArrayItem.(string))
											}
											ddFormulaAndFunctionEventQueryDefinition.SetIndexes(indexesDDArray)
										}
										k.Remove("indexes")

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionEventQueryDefinition.SetName(v.(string))
										}

										// handle search, which is a nested model
										k.Add("search.0")

										ddFormulaAndFunctionEventQueryDefinitionSearch := datadogV1.NewFormulaAndFunctionEventQueryDefinitionSearchWithDefaults()

										if v, ok := k.GetOkWith("query"); ok {
											ddFormulaAndFunctionEventQueryDefinitionSearch.SetQuery(v.(string))
										}
										k.Remove("search.0")
										ddFormulaAndFunctionEventQueryDefinition.SetSearch(*ddFormulaAndFunctionEventQueryDefinitionSearch)
										queriesDDArrayItem.FormulaAndFunctionEventQueryDefinition = ddFormulaAndFunctionEventQueryDefinition
									}
									k.Remove("formula_and_function_event_query_definition.0")
									k.Add("formula_and_function_process_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionProcessQueryDefinition := datadogV1.NewFormulaAndFunctionProcessQueryDefinitionWithDefaults()

										if v, ok := k.GetOkWith("aggregator"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionProcessQueryDataSource(v.(string)))
										}

										if v, ok := k.GetOkWith("is_normalized_cpu"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetIsNormalizedCpu(v.(bool))
										}

										if v, ok := k.GetOkWith("limit"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetLimit(int64(v.(int)))
										}

										if v, ok := k.GetOkWith("metric"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetMetric(v.(string))
										}

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetName(v.(string))
										}

										if v, ok := k.GetOkWith("sort"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetSort(datadogV1.QuerySortOrder(v.(string)))
										}
										k.Add("tag_filters")
										if tagFiltersArray, ok := k.GetOk(); ok {
											tagFiltersDDArray := make([]string, 0)
											for i := range tagFiltersArray.([]interface{}) {
												tagFiltersArrayItem := k.GetWith(i)
												tagFiltersDDArray = append(tagFiltersDDArray, tagFiltersArrayItem.(string))
											}
											ddFormulaAndFunctionProcessQueryDefinition.SetTagFilters(tagFiltersDDArray)
										}
										k.Remove("tag_filters")

										if v, ok := k.GetOkWith("text_filter"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetTextFilter(v.(string))
										}
										queriesDDArrayItem.FormulaAndFunctionProcessQueryDefinition = ddFormulaAndFunctionProcessQueryDefinition
									}
									k.Remove("formula_and_function_process_query_definition.0")

									if queriesDDArrayItem.GetActualInstance() == nil {
										return nil, fmt.Errorf("failed to find valid definition in formula_and_function_query_definition configuration")
									}
									queriesDDArray = append(queriesDDArray, *queriesDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetQueries(queriesDDArray)
							}
							k.Remove("query")

							if v, ok := k.GetOkWith("response_format"); ok {
								requestsDDArrayItem.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat(v.(string)))
							}

							// handle rum_query, which is a nested model
							k.Add("rum_query.0")

							requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemRumQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
							k.Remove("rum_query.0")
							requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

							// handle security_query, which is a nested model
							k.Add("security_query.0")

							requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
							k.Remove("security_query.0")
							requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

							// handle style, which is a nested model
							k.Add("style.0")

							requestsDDArrayItemStyle := datadogV1.NewWidgetRequestStyleWithDefaults()

							if v, ok := k.GetOkWith("line_type"); ok {
								requestsDDArrayItemStyle.SetLineType(datadogV1.WidgetLineType(v.(string)))
							}

							if v, ok := k.GetOkWith("line_width"); ok {
								requestsDDArrayItemStyle.SetLineWidth(datadogV1.WidgetLineWidth(v.(string)))
							}

							if v, ok := k.GetOkWith("palette"); ok {
								requestsDDArrayItemStyle.SetPalette(v.(string))
							}
							k.Remove("style.0")
							requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
							requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookToplistCellAttributesDefinition.SetRequests(requestsDDArray)
					}
					k.Remove("request")

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookToplistCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookToplistCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookToplistCellAttributesDefinition.SetTime(*ddNotebookToplistCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookToplistCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookToplistCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookToplistCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookToplistCellAttributesDefinition.SetType(datadogV1.ToplistWidgetDefinitionType(v.(string)))
					}
					k.Remove("definition.0")
					ddNotebookToplistCellAttributes.SetDefinition(*ddNotebookToplistCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookToplistCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle split_by, which is a nested model
					k.Add("split_by.0")

					ddNotebookToplistCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
					k.Add("keys")
					if keysArray, ok := k.GetOk(); ok {
						keysDDArray := make([]string, 0)
						for i := range keysArray.([]interface{}) {
							keysArrayItem := k.GetWith(i)
							keysDDArray = append(keysDDArray, keysArrayItem.(string))
						}
						ddNotebookToplistCellAttributesSplitBy.SetKeys(keysDDArray)
					}
					k.Remove("keys")
					k.Add("tags")
					if tagsArray, ok := k.GetOk(); ok {
						tagsDDArray := make([]string, 0)
						for i := range tagsArray.([]interface{}) {
							tagsArrayItem := k.GetWith(i)
							tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
						}
						ddNotebookToplistCellAttributesSplitBy.SetTags(tagsDDArray)
					}
					k.Remove("tags")
					k.Remove("split_by.0")
					ddNotebookToplistCellAttributes.SetSplitBy(*ddNotebookToplistCellAttributesSplitBy)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookToplistCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookToplistCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookToplistCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookToplistCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookToplistCellAttributes.SetTime(*ddNotebookToplistCellAttributesTime)
					ddNotebookCellCreateRequestAttributes.NotebookToplistCellAttributes = ddNotebookToplistCellAttributes
				}
				k.Remove("notebook_toplist_cell_attributes.0")
				k.Add("notebook_heat_map_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookHeatMapCellAttributes := datadogV1.NewNotebookHeatMapCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookHeatMapCellAttributesDefinition := datadogV1.NewHeatMapWidgetDefinitionWithDefaults()
					k.Add("custom_link")
					if customLinksArray, ok := k.GetOk(); ok {
						customLinksDDArray := make([]datadogV1.WidgetCustomLink, 0)
						for i := range customLinksArray.([]interface{}) {
							k.Add(i)

							customLinksDDArrayItem := datadogV1.NewWidgetCustomLinkWithDefaults()

							if v, ok := k.GetOkWith("is_hidden"); ok {
								customLinksDDArrayItem.SetIsHidden(v.(bool))
							}

							if v, ok := k.GetOkWith("label"); ok {
								customLinksDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("link"); ok {
								customLinksDDArrayItem.SetLink(v.(string))
							}

							if v, ok := k.GetOkWith("override_label"); ok {
								customLinksDDArrayItem.SetOverrideLabel(v.(string))
							}
							customLinksDDArray = append(customLinksDDArray, *customLinksDDArrayItem)
							k.Remove(i)
						}
						ddNotebookHeatMapCellAttributesDefinition.SetCustomLinks(customLinksDDArray)
					}
					k.Remove("custom_link")
					k.Add("event")
					if eventsArray, ok := k.GetOk(); ok {
						eventsDDArray := make([]datadogV1.WidgetEvent, 0)
						for i := range eventsArray.([]interface{}) {
							k.Add(i)

							eventsDDArrayItem := datadogV1.NewWidgetEventWithDefaults()

							if v, ok := k.GetOkWith("q"); ok {
								eventsDDArrayItem.SetQ(v.(string))
							}

							if v, ok := k.GetOkWith("tags_execution"); ok {
								eventsDDArrayItem.SetTagsExecution(v.(string))
							}
							eventsDDArray = append(eventsDDArray, *eventsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookHeatMapCellAttributesDefinition.SetEvents(eventsDDArray)
					}
					k.Remove("event")

					if v, ok := k.GetOkWith("legend_size"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetLegendSize(v.(string))
					}
					k.Add("request")
					if requestsArray, ok := k.GetOk(); ok {
						requestsDDArray := make([]datadogV1.HeatMapWidgetRequest, 0)
						for i := range requestsArray.([]interface{}) {
							k.Add(i)

							requestsDDArrayItem := datadogV1.NewHeatMapWidgetRequestWithDefaults()

							// handle apm_query, which is a nested model
							k.Add("apm_query.0")

							requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemApmQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
							k.Remove("apm_query.0")
							requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)

							// handle event_query, which is a nested model
							k.Add("event_query.0")

							requestsDDArrayItemEventQuery := datadogV1.NewEventQueryDefinitionWithDefaults()

							if v, ok := k.GetOkWith("search"); ok {
								requestsDDArrayItemEventQuery.SetSearch(v.(string))
							}

							if v, ok := k.GetOkWith("tags_execution"); ok {
								requestsDDArrayItemEventQuery.SetTagsExecution(v.(string))
							}
							k.Remove("event_query.0")
							requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)

							// handle log_query, which is a nested model
							k.Add("log_query.0")

							requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemLogQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
							k.Remove("log_query.0")
							requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)

							// handle network_query, which is a nested model
							k.Add("network_query.0")

							requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
							k.Remove("network_query.0")
							requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

							// handle process_query, which is a nested model
							k.Add("process_query.0")

							requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
							k.Add("filter_by")
							if filterByArray, ok := k.GetOk(); ok {
								filterByDDArray := make([]string, 0)
								for i := range filterByArray.([]interface{}) {
									filterByArrayItem := k.GetWith(i)
									filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
								}
								requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
							}
							k.Remove("filter_by")

							if v, ok := k.GetOkWith("limit"); ok {
								requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
							}

							if v, ok := k.GetOkWith("metric"); ok {
								requestsDDArrayItemProcessQuery.SetMetric(v.(string))
							}

							if v, ok := k.GetOkWith("search_by"); ok {
								requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
							}
							k.Remove("process_query.0")
							requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

							// handle profile_metrics_query, which is a nested model
							k.Add("profile_metrics_query.0")

							requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
							k.Remove("profile_metrics_query.0")
							requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

							if v, ok := k.GetOkWith("q"); ok {
								requestsDDArrayItem.SetQ(v.(string))
							}

							// handle rum_query, which is a nested model
							k.Add("rum_query.0")

							requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemRumQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
							k.Remove("rum_query.0")
							requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

							// handle security_query, which is a nested model
							k.Add("security_query.0")

							requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
							k.Remove("security_query.0")
							requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

							// handle style, which is a nested model
							k.Add("style.0")

							requestsDDArrayItemStyle := datadogV1.NewWidgetStyleWithDefaults()

							if v, ok := k.GetOkWith("palette"); ok {
								requestsDDArrayItemStyle.SetPalette(v.(string))
							}
							k.Remove("style.0")
							requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
							requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookHeatMapCellAttributesDefinition.SetRequests(requestsDDArray)
					}
					k.Remove("request")

					if v, ok := k.GetOkWith("show_legend"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetShowLegend(v.(bool))
					}

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookHeatMapCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookHeatMapCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookHeatMapCellAttributesDefinition.SetTime(*ddNotebookHeatMapCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetType(datadogV1.HeatMapWidgetDefinitionType(v.(string)))
					}

					// handle yaxis, which is a nested model
					k.Add("yaxis.0")

					ddNotebookHeatMapCellAttributesDefinitionYaxis := datadogV1.NewWidgetAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("label"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetLabel(v.(string))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetScale(v.(string))
					}
					k.Remove("yaxis.0")
					ddNotebookHeatMapCellAttributesDefinition.SetYaxis(*ddNotebookHeatMapCellAttributesDefinitionYaxis)
					k.Remove("definition.0")
					ddNotebookHeatMapCellAttributes.SetDefinition(*ddNotebookHeatMapCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookHeatMapCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle split_by, which is a nested model
					k.Add("split_by.0")

					ddNotebookHeatMapCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
					k.Add("keys")
					if keysArray, ok := k.GetOk(); ok {
						keysDDArray := make([]string, 0)
						for i := range keysArray.([]interface{}) {
							keysArrayItem := k.GetWith(i)
							keysDDArray = append(keysDDArray, keysArrayItem.(string))
						}
						ddNotebookHeatMapCellAttributesSplitBy.SetKeys(keysDDArray)
					}
					k.Remove("keys")
					k.Add("tags")
					if tagsArray, ok := k.GetOk(); ok {
						tagsDDArray := make([]string, 0)
						for i := range tagsArray.([]interface{}) {
							tagsArrayItem := k.GetWith(i)
							tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
						}
						ddNotebookHeatMapCellAttributesSplitBy.SetTags(tagsDDArray)
					}
					k.Remove("tags")
					k.Remove("split_by.0")
					ddNotebookHeatMapCellAttributes.SetSplitBy(*ddNotebookHeatMapCellAttributesSplitBy)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookHeatMapCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookHeatMapCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookHeatMapCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookHeatMapCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookHeatMapCellAttributes.SetTime(*ddNotebookHeatMapCellAttributesTime)
					ddNotebookCellCreateRequestAttributes.NotebookHeatMapCellAttributes = ddNotebookHeatMapCellAttributes
				}
				k.Remove("notebook_heat_map_cell_attributes.0")
				k.Add("notebook_distribution_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookDistributionCellAttributes := datadogV1.NewNotebookDistributionCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookDistributionCellAttributesDefinition := datadogV1.NewDistributionWidgetDefinitionWithDefaults()

					if v, ok := k.GetOkWith("legend_size"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetLegendSize(v.(string))
					}
					k.Add("marker")
					if markersArray, ok := k.GetOk(); ok {
						markersDDArray := make([]datadogV1.WidgetMarker, 0)
						for i := range markersArray.([]interface{}) {
							k.Add(i)

							markersDDArrayItem := datadogV1.NewWidgetMarkerWithDefaults()

							if v, ok := k.GetOkWith("display_type"); ok {
								markersDDArrayItem.SetDisplayType(v.(string))
							}

							if v, ok := k.GetOkWith("label"); ok {
								markersDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("time"); ok {
								markersDDArrayItem.SetTime(v.(string))
							}

							if v, ok := k.GetOkWith("value"); ok {
								markersDDArrayItem.SetValue(v.(string))
							}
							markersDDArray = append(markersDDArray, *markersDDArrayItem)
							k.Remove(i)
						}
						ddNotebookDistributionCellAttributesDefinition.SetMarkers(markersDDArray)
					}
					k.Remove("marker")
					k.Add("request")
					if requestsArray, ok := k.GetOk(); ok {
						requestsDDArray := make([]datadogV1.DistributionWidgetRequest, 0)
						for i := range requestsArray.([]interface{}) {
							k.Add(i)

							requestsDDArrayItem := datadogV1.NewDistributionWidgetRequestWithDefaults()

							// handle apm_query, which is a nested model
							k.Add("apm_query.0")

							requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemApmQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
							k.Remove("apm_query.0")
							requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)

							// handle event_query, which is a nested model
							k.Add("event_query.0")

							requestsDDArrayItemEventQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemEventQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemEventQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemEventQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemEventQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemEventQuery.SetCompute(*requestsDDArrayItemEventQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemEventQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemEventQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemEventQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemEventQuery.SetSearch(*requestsDDArrayItemEventQuerySearch)
							k.Remove("event_query.0")
							requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)

							// handle log_query, which is a nested model
							k.Add("log_query.0")

							requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemLogQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
							k.Remove("log_query.0")
							requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)

							// handle network_query, which is a nested model
							k.Add("network_query.0")

							requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
							k.Remove("network_query.0")
							requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

							// handle process_query, which is a nested model
							k.Add("process_query.0")

							requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
							k.Add("filter_by")
							if filterByArray, ok := k.GetOk(); ok {
								filterByDDArray := make([]string, 0)
								for i := range filterByArray.([]interface{}) {
									filterByArrayItem := k.GetWith(i)
									filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
								}
								requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
							}
							k.Remove("filter_by")

							if v, ok := k.GetOkWith("limit"); ok {
								requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
							}

							if v, ok := k.GetOkWith("metric"); ok {
								requestsDDArrayItemProcessQuery.SetMetric(v.(string))
							}

							if v, ok := k.GetOkWith("search_by"); ok {
								requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
							}
							k.Remove("process_query.0")
							requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

							// handle profile_metrics_query, which is a nested model
							k.Add("profile_metrics_query.0")

							requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
							k.Remove("profile_metrics_query.0")
							requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

							if v, ok := k.GetOkWith("q"); ok {
								requestsDDArrayItem.SetQ(v.(string))
							}

							// handle rum_query, which is a nested model
							k.Add("rum_query.0")

							requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemRumQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
							k.Remove("rum_query.0")
							requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

							// handle security_query, which is a nested model
							k.Add("security_query.0")

							requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
							k.Remove("security_query.0")
							requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

							// handle style, which is a nested model
							k.Add("style.0")

							requestsDDArrayItemStyle := datadogV1.NewWidgetStyleWithDefaults()

							if v, ok := k.GetOkWith("palette"); ok {
								requestsDDArrayItemStyle.SetPalette(v.(string))
							}
							k.Remove("style.0")
							requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
							requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookDistributionCellAttributesDefinition.SetRequests(requestsDDArray)
					}
					k.Remove("request")

					if v, ok := k.GetOkWith("show_legend"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetShowLegend(v.(bool))
					}

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookDistributionCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookDistributionCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookDistributionCellAttributesDefinition.SetTime(*ddNotebookDistributionCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetType(datadogV1.DistributionWidgetDefinitionType(v.(string)))
					}

					// handle xaxis, which is a nested model
					k.Add("xaxis.0")

					ddNotebookDistributionCellAttributesDefinitionXaxis := datadogV1.NewDistributionWidgetXAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookDistributionCellAttributesDefinitionXaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookDistributionCellAttributesDefinitionXaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookDistributionCellAttributesDefinitionXaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookDistributionCellAttributesDefinitionXaxis.SetScale(v.(string))
					}
					k.Remove("xaxis.0")
					ddNotebookDistributionCellAttributesDefinition.SetXaxis(*ddNotebookDistributionCellAttributesDefinitionXaxis)

					// handle yaxis, which is a nested model
					k.Add("yaxis.0")

					ddNotebookDistributionCellAttributesDefinitionYaxis := datadogV1.NewDistributionWidgetYAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("label"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetLabel(v.(string))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetScale(v.(string))
					}
					k.Remove("yaxis.0")
					ddNotebookDistributionCellAttributesDefinition.SetYaxis(*ddNotebookDistributionCellAttributesDefinitionYaxis)
					k.Remove("definition.0")
					ddNotebookDistributionCellAttributes.SetDefinition(*ddNotebookDistributionCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookDistributionCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle split_by, which is a nested model
					k.Add("split_by.0")

					ddNotebookDistributionCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
					k.Add("keys")
					if keysArray, ok := k.GetOk(); ok {
						keysDDArray := make([]string, 0)
						for i := range keysArray.([]interface{}) {
							keysArrayItem := k.GetWith(i)
							keysDDArray = append(keysDDArray, keysArrayItem.(string))
						}
						ddNotebookDistributionCellAttributesSplitBy.SetKeys(keysDDArray)
					}
					k.Remove("keys")
					k.Add("tags")
					if tagsArray, ok := k.GetOk(); ok {
						tagsDDArray := make([]string, 0)
						for i := range tagsArray.([]interface{}) {
							tagsArrayItem := k.GetWith(i)
							tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
						}
						ddNotebookDistributionCellAttributesSplitBy.SetTags(tagsDDArray)
					}
					k.Remove("tags")
					k.Remove("split_by.0")
					ddNotebookDistributionCellAttributes.SetSplitBy(*ddNotebookDistributionCellAttributesSplitBy)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookDistributionCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookDistributionCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookDistributionCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookDistributionCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookDistributionCellAttributes.SetTime(*ddNotebookDistributionCellAttributesTime)
					ddNotebookCellCreateRequestAttributes.NotebookDistributionCellAttributes = ddNotebookDistributionCellAttributes
				}
				k.Remove("notebook_distribution_cell_attributes.0")
				k.Add("notebook_log_stream_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookLogStreamCellAttributes := datadogV1.NewNotebookLogStreamCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookLogStreamCellAttributesDefinition := datadogV1.NewLogStreamWidgetDefinitionWithDefaults()
					k.Add("columns")
					if columnsArray, ok := k.GetOk(); ok {
						columnsDDArray := make([]string, 0)
						for i := range columnsArray.([]interface{}) {
							columnsArrayItem := k.GetWith(i)
							columnsDDArray = append(columnsDDArray, columnsArrayItem.(string))
						}
						ddNotebookLogStreamCellAttributesDefinition.SetColumns(columnsDDArray)
					}
					k.Remove("columns")
					k.Add("indexes")
					if indexesArray, ok := k.GetOk(); ok {
						indexesDDArray := make([]string, 0)
						for i := range indexesArray.([]interface{}) {
							indexesArrayItem := k.GetWith(i)
							indexesDDArray = append(indexesDDArray, indexesArrayItem.(string))
						}
						ddNotebookLogStreamCellAttributesDefinition.SetIndexes(indexesDDArray)
					}
					k.Remove("indexes")

					if v, ok := k.GetOkWith("logset"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetLogset(v.(string))
					}

					if v, ok := k.GetOkWith("message_display"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetMessageDisplay(datadogV1.WidgetMessageDisplay(v.(string)))
					}

					if v, ok := k.GetOkWith("query"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetQuery(v.(string))
					}

					if v, ok := k.GetOkWith("show_date_column"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetShowDateColumn(v.(bool))
					}

					if v, ok := k.GetOkWith("show_message_column"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetShowMessageColumn(v.(bool))
					}

					// handle sort, which is a nested model
					k.Add("sort.0")

					ddNotebookLogStreamCellAttributesDefinitionSort := datadogV1.NewWidgetFieldSortWithDefaults()

					if v, ok := k.GetOkWith("column"); ok {
						ddNotebookLogStreamCellAttributesDefinitionSort.SetColumn(v.(string))
					}

					if v, ok := k.GetOkWith("order"); ok {
						ddNotebookLogStreamCellAttributesDefinitionSort.SetOrder(datadogV1.WidgetSort(v.(string)))
					}
					k.Remove("sort.0")
					ddNotebookLogStreamCellAttributesDefinition.SetSort(*ddNotebookLogStreamCellAttributesDefinitionSort)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookLogStreamCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookLogStreamCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookLogStreamCellAttributesDefinition.SetTime(*ddNotebookLogStreamCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetType(datadogV1.LogStreamWidgetDefinitionType(v.(string)))
					}
					k.Remove("definition.0")
					ddNotebookLogStreamCellAttributes.SetDefinition(*ddNotebookLogStreamCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookLogStreamCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookLogStreamCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookLogStreamCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookLogStreamCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookLogStreamCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookLogStreamCellAttributes.SetTime(*ddNotebookLogStreamCellAttributesTime)
					ddNotebookCellCreateRequestAttributes.NotebookLogStreamCellAttributes = ddNotebookLogStreamCellAttributes
				}
				k.Remove("notebook_log_stream_cell_attributes.0")

				if ddNotebookCellCreateRequestAttributes.GetActualInstance() == nil {
					return nil, fmt.Errorf("failed to find valid definition in notebook_cell_create_request_attributes configuration")
				}
				k.Remove("attributes.0")
				ddNotebookCellCreateRequest.SetAttributes(*ddNotebookCellCreateRequestAttributes)

				if v, ok := k.GetOkWith("type"); ok {
					ddNotebookCellCreateRequest.SetType(datadogV1.NotebookCellResourceType(v.(string)))
				}
				cellsDDArrayItem.NotebookCellCreateRequest = ddNotebookCellCreateRequest
			}
			k.Remove("notebook_cell_create_request.0")
			k.Add("notebook_cell_update_request.0")
			if _, ok := k.GetOk(); ok {

				ddNotebookCellUpdateRequest := datadogV1.NewNotebookCellUpdateRequestWithDefaults()

				// handle attributes, which is a nested model
				k.Add("attributes.0")

				ddNotebookCellUpdateRequestAttributes := &datadogV1.NotebookCellUpdateRequestAttributes{}
				// handle notebook_cell_update_request_attributes, which is a oneOf model
				k.Add("notebook_markdown_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookMarkdownCellAttributes := datadogV1.NewNotebookMarkdownCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookMarkdownCellAttributesDefinition := datadogV1.NewNotebookMarkdownCellDefinitionWithDefaults()

					if v, ok := k.GetOkWith("text"); ok {
						ddNotebookMarkdownCellAttributesDefinition.SetText(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookMarkdownCellAttributesDefinition.SetType(datadogV1.NotebookMarkdownCellDefinitionType(v.(string)))
					}
					k.Remove("definition.0")
					ddNotebookMarkdownCellAttributes.SetDefinition(*ddNotebookMarkdownCellAttributesDefinition)
					ddNotebookCellUpdateRequestAttributes.NotebookMarkdownCellAttributes = ddNotebookMarkdownCellAttributes
				}
				k.Remove("notebook_markdown_cell_attributes.0")
				k.Add("notebook_timeseries_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookTimeseriesCellAttributes := datadogV1.NewNotebookTimeseriesCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookTimeseriesCellAttributesDefinition := datadogV1.NewTimeseriesWidgetDefinitionWithDefaults()
					k.Add("custom_link")
					if customLinksArray, ok := k.GetOk(); ok {
						customLinksDDArray := make([]datadogV1.WidgetCustomLink, 0)
						for i := range customLinksArray.([]interface{}) {
							k.Add(i)

							customLinksDDArrayItem := datadogV1.NewWidgetCustomLinkWithDefaults()

							if v, ok := k.GetOkWith("is_hidden"); ok {
								customLinksDDArrayItem.SetIsHidden(v.(bool))
							}

							if v, ok := k.GetOkWith("label"); ok {
								customLinksDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("link"); ok {
								customLinksDDArrayItem.SetLink(v.(string))
							}

							if v, ok := k.GetOkWith("override_label"); ok {
								customLinksDDArrayItem.SetOverrideLabel(v.(string))
							}
							customLinksDDArray = append(customLinksDDArray, *customLinksDDArrayItem)
							k.Remove(i)
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetCustomLinks(customLinksDDArray)
					}
					k.Remove("custom_link")
					k.Add("event")
					if eventsArray, ok := k.GetOk(); ok {
						eventsDDArray := make([]datadogV1.WidgetEvent, 0)
						for i := range eventsArray.([]interface{}) {
							k.Add(i)

							eventsDDArrayItem := datadogV1.NewWidgetEventWithDefaults()

							if v, ok := k.GetOkWith("q"); ok {
								eventsDDArrayItem.SetQ(v.(string))
							}

							if v, ok := k.GetOkWith("tags_execution"); ok {
								eventsDDArrayItem.SetTagsExecution(v.(string))
							}
							eventsDDArray = append(eventsDDArray, *eventsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetEvents(eventsDDArray)
					}
					k.Remove("event")
					k.Add("legend_columns")
					if legendColumnsArray, ok := k.GetOk(); ok {
						legendColumnsDDArray := make([]datadogV1.TimeseriesWidgetLegendColumn, 0)
						for i := range legendColumnsArray.([]interface{}) {
							legendColumnsArrayItem := k.GetWith(i)
							legendColumnsDDArray = append(legendColumnsDDArray, datadogV1.TimeseriesWidgetLegendColumn(legendColumnsArrayItem.(string)))
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetLegendColumns(legendColumnsDDArray)
					}
					k.Remove("legend_columns")

					if v, ok := k.GetOkWith("legend_layout"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetLegendLayout(datadogV1.TimeseriesWidgetLegendLayout(v.(string)))
					}

					if v, ok := k.GetOkWith("legend_size"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetLegendSize(v.(string))
					}
					k.Add("marker")
					if markersArray, ok := k.GetOk(); ok {
						markersDDArray := make([]datadogV1.WidgetMarker, 0)
						for i := range markersArray.([]interface{}) {
							k.Add(i)

							markersDDArrayItem := datadogV1.NewWidgetMarkerWithDefaults()

							if v, ok := k.GetOkWith("display_type"); ok {
								markersDDArrayItem.SetDisplayType(v.(string))
							}

							if v, ok := k.GetOkWith("label"); ok {
								markersDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("time"); ok {
								markersDDArrayItem.SetTime(v.(string))
							}

							if v, ok := k.GetOkWith("value"); ok {
								markersDDArrayItem.SetValue(v.(string))
							}
							markersDDArray = append(markersDDArray, *markersDDArrayItem)
							k.Remove(i)
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetMarkers(markersDDArray)
					}
					k.Remove("marker")
					k.Add("request")
					if requestsArray, ok := k.GetOk(); ok {
						requestsDDArray := make([]datadogV1.TimeseriesWidgetRequest, 0)
						for i := range requestsArray.([]interface{}) {
							k.Add(i)

							requestsDDArrayItem := datadogV1.NewTimeseriesWidgetRequestWithDefaults()

							// handle apm_query, which is a nested model
							k.Add("apm_query.0")

							requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemApmQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
							k.Remove("apm_query.0")
							requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)

							if v, ok := k.GetOkWith("display_type"); ok {
								requestsDDArrayItem.SetDisplayType(datadogV1.WidgetDisplayType(v.(string)))
							}

							// handle event_query, which is a nested model
							k.Add("event_query.0")

							requestsDDArrayItemEventQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemEventQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemEventQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemEventQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemEventQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemEventQuery.SetCompute(*requestsDDArrayItemEventQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemEventQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemEventQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemEventQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemEventQuery.SetSearch(*requestsDDArrayItemEventQuerySearch)
							k.Remove("event_query.0")
							requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)
							k.Add("formula")
							if formulasArray, ok := k.GetOk(); ok {
								formulasDDArray := make([]datadogV1.WidgetFormula, 0)
								for i := range formulasArray.([]interface{}) {
									k.Add(i)

									formulasDDArrayItem := datadogV1.NewWidgetFormulaWithDefaults()

									if v, ok := k.GetOkWith("alias"); ok {
										formulasDDArrayItem.SetAlias(v.(string))
									}

									if v, ok := k.GetOkWith("formula"); ok {
										formulasDDArrayItem.SetFormula(v.(string))
									}

									// handle limit, which is a nested model
									k.Add("limit.0")

									formulasDDArrayItemLimit := datadogV1.NewWidgetFormulaLimitWithDefaults()

									if v, ok := k.GetOkWith("count"); ok {
										formulasDDArrayItemLimit.SetCount(int64(v.(int)))
									}

									if v, ok := k.GetOkWith("order"); ok {
										formulasDDArrayItemLimit.SetOrder(datadogV1.QuerySortOrder(v.(string)))
									}
									k.Remove("limit.0")
									formulasDDArrayItem.SetLimit(*formulasDDArrayItemLimit)
									formulasDDArray = append(formulasDDArray, *formulasDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetFormulas(formulasDDArray)
							}
							k.Remove("formula")

							// handle log_query, which is a nested model
							k.Add("log_query.0")

							requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemLogQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
							k.Remove("log_query.0")
							requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)
							k.Add("metadata")
							if metadataArray, ok := k.GetOk(); ok {
								metadataDDArray := make([]datadogV1.TimeseriesWidgetExpressionAlias, 0)
								for i := range metadataArray.([]interface{}) {
									k.Add(i)

									metadataDDArrayItem := datadogV1.NewTimeseriesWidgetExpressionAliasWithDefaults()

									if v, ok := k.GetOkWith("alias_name"); ok {
										metadataDDArrayItem.SetAliasName(v.(string))
									}

									if v, ok := k.GetOkWith("expression"); ok {
										metadataDDArrayItem.SetExpression(v.(string))
									}
									metadataDDArray = append(metadataDDArray, *metadataDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetMetadata(metadataDDArray)
							}
							k.Remove("metadata")

							// handle network_query, which is a nested model
							k.Add("network_query.0")

							requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
							k.Remove("network_query.0")
							requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

							if v, ok := k.GetOkWith("on_right_yaxis"); ok {
								requestsDDArrayItem.SetOnRightYaxis(v.(bool))
							}

							// handle process_query, which is a nested model
							k.Add("process_query.0")

							requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
							k.Add("filter_by")
							if filterByArray, ok := k.GetOk(); ok {
								filterByDDArray := make([]string, 0)
								for i := range filterByArray.([]interface{}) {
									filterByArrayItem := k.GetWith(i)
									filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
								}
								requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
							}
							k.Remove("filter_by")

							if v, ok := k.GetOkWith("limit"); ok {
								requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
							}

							if v, ok := k.GetOkWith("metric"); ok {
								requestsDDArrayItemProcessQuery.SetMetric(v.(string))
							}

							if v, ok := k.GetOkWith("search_by"); ok {
								requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
							}
							k.Remove("process_query.0")
							requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

							// handle profile_metrics_query, which is a nested model
							k.Add("profile_metrics_query.0")

							requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
							k.Remove("profile_metrics_query.0")
							requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

							if v, ok := k.GetOkWith("q"); ok {
								requestsDDArrayItem.SetQ(v.(string))
							}
							k.Add("query")
							if queriesArray, ok := k.GetOk(); ok {
								queriesDDArray := make([]datadogV1.FormulaAndFunctionQueryDefinition, 0)
								for i := range queriesArray.([]interface{}) {
									k.Add(i)

									queriesDDArrayItem := &datadogV1.FormulaAndFunctionQueryDefinition{}
									// handle formula_and_function_query_definition, which is a oneOf model
									k.Add("formula_and_function_metric_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionMetricQueryDefinition := datadogV1.NewFormulaAndFunctionMetricQueryDefinitionWithDefaults()

										if v, ok := k.GetOkWith("aggregator"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionMetricDataSource(v.(string)))
										}

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetName(v.(string))
										}

										if v, ok := k.GetOkWith("query"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetQuery(v.(string))
										}
										queriesDDArrayItem.FormulaAndFunctionMetricQueryDefinition = ddFormulaAndFunctionMetricQueryDefinition
									}
									k.Remove("formula_and_function_metric_query_definition.0")
									k.Add("formula_and_function_event_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionEventQueryDefinition := datadogV1.NewFormulaAndFunctionEventQueryDefinitionWithDefaults()

										// handle compute, which is a nested model
										k.Add("compute.0")

										ddFormulaAndFunctionEventQueryDefinitionCompute := datadogV1.NewFormulaAndFunctionEventQueryDefinitionComputeWithDefaults()

										if v, ok := k.GetOkWith("aggregation"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("interval"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetInterval(int64(v.(int)))
										}

										if v, ok := k.GetOkWith("metric"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetMetric(v.(string))
										}
										k.Remove("compute.0")
										ddFormulaAndFunctionEventQueryDefinition.SetCompute(*ddFormulaAndFunctionEventQueryDefinitionCompute)

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionEventQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionEventsDataSource(v.(string)))
										}
										k.Add("group_by")
										if groupByArray, ok := k.GetOk(); ok {
											groupByDDArray := make([]datadogV1.FormulaAndFunctionEventQueryGroupBy, 0)
											for i := range groupByArray.([]interface{}) {
												k.Add(i)

												groupByDDArrayItem := datadogV1.NewFormulaAndFunctionEventQueryGroupByWithDefaults()

												if v, ok := k.GetOkWith("facet"); ok {
													groupByDDArrayItem.SetFacet(v.(string))
												}

												if v, ok := k.GetOkWith("limit"); ok {
													groupByDDArrayItem.SetLimit(int64(v.(int)))
												}

												// handle sort, which is a nested model
												k.Add("sort.0")

												groupByDDArrayItemSort := datadogV1.NewFormulaAndFunctionEventQueryGroupBySortWithDefaults()

												if v, ok := k.GetOkWith("aggregation"); ok {
													groupByDDArrayItemSort.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
												}

												if v, ok := k.GetOkWith("metric"); ok {
													groupByDDArrayItemSort.SetMetric(v.(string))
												}

												if v, ok := k.GetOkWith("order"); ok {
													groupByDDArrayItemSort.SetOrder(datadogV1.QuerySortOrder(v.(string)))
												}
												k.Remove("sort.0")
												groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
												groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
												k.Remove(i)
											}
											ddFormulaAndFunctionEventQueryDefinition.SetGroupBy(groupByDDArray)
										}
										k.Remove("group_by")
										k.Add("indexes")
										if indexesArray, ok := k.GetOk(); ok {
											indexesDDArray := make([]string, 0)
											for i := range indexesArray.([]interface{}) {
												indexesArrayItem := k.GetWith(i)
												indexesDDArray = append(indexesDDArray, indexesArrayItem.(string))
											}
											ddFormulaAndFunctionEventQueryDefinition.SetIndexes(indexesDDArray)
										}
										k.Remove("indexes")

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionEventQueryDefinition.SetName(v.(string))
										}

										// handle search, which is a nested model
										k.Add("search.0")

										ddFormulaAndFunctionEventQueryDefinitionSearch := datadogV1.NewFormulaAndFunctionEventQueryDefinitionSearchWithDefaults()

										if v, ok := k.GetOkWith("query"); ok {
											ddFormulaAndFunctionEventQueryDefinitionSearch.SetQuery(v.(string))
										}
										k.Remove("search.0")
										ddFormulaAndFunctionEventQueryDefinition.SetSearch(*ddFormulaAndFunctionEventQueryDefinitionSearch)
										queriesDDArrayItem.FormulaAndFunctionEventQueryDefinition = ddFormulaAndFunctionEventQueryDefinition
									}
									k.Remove("formula_and_function_event_query_definition.0")
									k.Add("formula_and_function_process_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionProcessQueryDefinition := datadogV1.NewFormulaAndFunctionProcessQueryDefinitionWithDefaults()

										if v, ok := k.GetOkWith("aggregator"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionProcessQueryDataSource(v.(string)))
										}

										if v, ok := k.GetOkWith("is_normalized_cpu"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetIsNormalizedCpu(v.(bool))
										}

										if v, ok := k.GetOkWith("limit"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetLimit(int64(v.(int)))
										}

										if v, ok := k.GetOkWith("metric"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetMetric(v.(string))
										}

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetName(v.(string))
										}

										if v, ok := k.GetOkWith("sort"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetSort(datadogV1.QuerySortOrder(v.(string)))
										}
										k.Add("tag_filters")
										if tagFiltersArray, ok := k.GetOk(); ok {
											tagFiltersDDArray := make([]string, 0)
											for i := range tagFiltersArray.([]interface{}) {
												tagFiltersArrayItem := k.GetWith(i)
												tagFiltersDDArray = append(tagFiltersDDArray, tagFiltersArrayItem.(string))
											}
											ddFormulaAndFunctionProcessQueryDefinition.SetTagFilters(tagFiltersDDArray)
										}
										k.Remove("tag_filters")

										if v, ok := k.GetOkWith("text_filter"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetTextFilter(v.(string))
										}
										queriesDDArrayItem.FormulaAndFunctionProcessQueryDefinition = ddFormulaAndFunctionProcessQueryDefinition
									}
									k.Remove("formula_and_function_process_query_definition.0")

									if queriesDDArrayItem.GetActualInstance() == nil {
										return nil, fmt.Errorf("failed to find valid definition in formula_and_function_query_definition configuration")
									}
									queriesDDArray = append(queriesDDArray, *queriesDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetQueries(queriesDDArray)
							}
							k.Remove("query")

							if v, ok := k.GetOkWith("response_format"); ok {
								requestsDDArrayItem.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat(v.(string)))
							}

							// handle rum_query, which is a nested model
							k.Add("rum_query.0")

							requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemRumQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
							k.Remove("rum_query.0")
							requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

							// handle security_query, which is a nested model
							k.Add("security_query.0")

							requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
							k.Remove("security_query.0")
							requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

							// handle style, which is a nested model
							k.Add("style.0")

							requestsDDArrayItemStyle := datadogV1.NewWidgetRequestStyleWithDefaults()

							if v, ok := k.GetOkWith("line_type"); ok {
								requestsDDArrayItemStyle.SetLineType(datadogV1.WidgetLineType(v.(string)))
							}

							if v, ok := k.GetOkWith("line_width"); ok {
								requestsDDArrayItemStyle.SetLineWidth(datadogV1.WidgetLineWidth(v.(string)))
							}

							if v, ok := k.GetOkWith("palette"); ok {
								requestsDDArrayItemStyle.SetPalette(v.(string))
							}
							k.Remove("style.0")
							requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
							requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookTimeseriesCellAttributesDefinition.SetRequests(requestsDDArray)
					}
					k.Remove("request")

					// handle right_yaxis, which is a nested model
					k.Add("right_yaxis.0")

					ddNotebookTimeseriesCellAttributesDefinitionRightYaxis := datadogV1.NewWidgetAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("label"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetLabel(v.(string))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionRightYaxis.SetScale(v.(string))
					}
					k.Remove("right_yaxis.0")
					ddNotebookTimeseriesCellAttributesDefinition.SetRightYaxis(*ddNotebookTimeseriesCellAttributesDefinitionRightYaxis)

					if v, ok := k.GetOkWith("show_legend"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetShowLegend(v.(bool))
					}

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookTimeseriesCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookTimeseriesCellAttributesDefinition.SetTime(*ddNotebookTimeseriesCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookTimeseriesCellAttributesDefinition.SetType(datadogV1.TimeseriesWidgetDefinitionType(v.(string)))
					}

					// handle yaxis, which is a nested model
					k.Add("yaxis.0")

					ddNotebookTimeseriesCellAttributesDefinitionYaxis := datadogV1.NewWidgetAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("label"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetLabel(v.(string))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookTimeseriesCellAttributesDefinitionYaxis.SetScale(v.(string))
					}
					k.Remove("yaxis.0")
					ddNotebookTimeseriesCellAttributesDefinition.SetYaxis(*ddNotebookTimeseriesCellAttributesDefinitionYaxis)
					k.Remove("definition.0")
					ddNotebookTimeseriesCellAttributes.SetDefinition(*ddNotebookTimeseriesCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookTimeseriesCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle split_by, which is a nested model
					k.Add("split_by.0")

					ddNotebookTimeseriesCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
					k.Add("keys")
					if keysArray, ok := k.GetOk(); ok {
						keysDDArray := make([]string, 0)
						for i := range keysArray.([]interface{}) {
							keysArrayItem := k.GetWith(i)
							keysDDArray = append(keysDDArray, keysArrayItem.(string))
						}
						ddNotebookTimeseriesCellAttributesSplitBy.SetKeys(keysDDArray)
					}
					k.Remove("keys")
					k.Add("tags")
					if tagsArray, ok := k.GetOk(); ok {
						tagsDDArray := make([]string, 0)
						for i := range tagsArray.([]interface{}) {
							tagsArrayItem := k.GetWith(i)
							tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
						}
						ddNotebookTimeseriesCellAttributesSplitBy.SetTags(tagsDDArray)
					}
					k.Remove("tags")
					k.Remove("split_by.0")
					ddNotebookTimeseriesCellAttributes.SetSplitBy(*ddNotebookTimeseriesCellAttributesSplitBy)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookTimeseriesCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookTimeseriesCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookTimeseriesCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookTimeseriesCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookTimeseriesCellAttributes.SetTime(*ddNotebookTimeseriesCellAttributesTime)
					ddNotebookCellUpdateRequestAttributes.NotebookTimeseriesCellAttributes = ddNotebookTimeseriesCellAttributes
				}
				k.Remove("notebook_timeseries_cell_attributes.0")
				k.Add("notebook_toplist_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookToplistCellAttributes := datadogV1.NewNotebookToplistCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookToplistCellAttributesDefinition := datadogV1.NewToplistWidgetDefinitionWithDefaults()
					k.Add("custom_link")
					if customLinksArray, ok := k.GetOk(); ok {
						customLinksDDArray := make([]datadogV1.WidgetCustomLink, 0)
						for i := range customLinksArray.([]interface{}) {
							k.Add(i)

							customLinksDDArrayItem := datadogV1.NewWidgetCustomLinkWithDefaults()

							if v, ok := k.GetOkWith("is_hidden"); ok {
								customLinksDDArrayItem.SetIsHidden(v.(bool))
							}

							if v, ok := k.GetOkWith("label"); ok {
								customLinksDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("link"); ok {
								customLinksDDArrayItem.SetLink(v.(string))
							}

							if v, ok := k.GetOkWith("override_label"); ok {
								customLinksDDArrayItem.SetOverrideLabel(v.(string))
							}
							customLinksDDArray = append(customLinksDDArray, *customLinksDDArrayItem)
							k.Remove(i)
						}
						ddNotebookToplistCellAttributesDefinition.SetCustomLinks(customLinksDDArray)
					}
					k.Remove("custom_link")
					k.Add("request")
					if requestsArray, ok := k.GetOk(); ok {
						requestsDDArray := make([]datadogV1.ToplistWidgetRequest, 0)
						for i := range requestsArray.([]interface{}) {
							k.Add(i)

							requestsDDArrayItem := datadogV1.NewToplistWidgetRequestWithDefaults()

							// handle apm_query, which is a nested model
							k.Add("apm_query.0")

							requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemApmQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
							k.Remove("apm_query.0")
							requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)
							k.Add("conditional_format")
							if conditionalFormatsArray, ok := k.GetOk(); ok {
								conditionalFormatsDDArray := make([]datadogV1.WidgetConditionalFormat, 0)
								for i := range conditionalFormatsArray.([]interface{}) {
									k.Add(i)

									conditionalFormatsDDArrayItem := datadogV1.NewWidgetConditionalFormatWithDefaults()

									if v, ok := k.GetOkWith("comparator"); ok {
										conditionalFormatsDDArrayItem.SetComparator(datadogV1.WidgetComparator(v.(string)))
									}

									if v, ok := k.GetOkWith("custom_bg_color"); ok {
										conditionalFormatsDDArrayItem.SetCustomBgColor(v.(string))
									}

									if v, ok := k.GetOkWith("custom_fg_color"); ok {
										conditionalFormatsDDArrayItem.SetCustomFgColor(v.(string))
									}

									if v, ok := k.GetOkWith("hide_value"); ok {
										conditionalFormatsDDArrayItem.SetHideValue(v.(bool))
									}

									if v, ok := k.GetOkWith("image_url"); ok {
										conditionalFormatsDDArrayItem.SetImageUrl(v.(string))
									}

									if v, ok := k.GetOkWith("metric"); ok {
										conditionalFormatsDDArrayItem.SetMetric(v.(string))
									}

									if v, ok := k.GetOkWith("palette"); ok {
										conditionalFormatsDDArrayItem.SetPalette(datadogV1.WidgetPalette(v.(string)))
									}

									if v, ok := k.GetOkWith("timeframe"); ok {
										conditionalFormatsDDArrayItem.SetTimeframe(v.(string))
									}

									if v, ok := k.GetOkWith("value"); ok {
										conditionalFormatsDDArrayItem.SetValue(v.(float64))
									}
									conditionalFormatsDDArray = append(conditionalFormatsDDArray, *conditionalFormatsDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetConditionalFormats(conditionalFormatsDDArray)
							}
							k.Remove("conditional_format")

							// handle event_query, which is a nested model
							k.Add("event_query.0")

							requestsDDArrayItemEventQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemEventQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemEventQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemEventQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemEventQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemEventQuery.SetCompute(*requestsDDArrayItemEventQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemEventQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemEventQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemEventQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemEventQuery.SetSearch(*requestsDDArrayItemEventQuerySearch)
							k.Remove("event_query.0")
							requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)
							k.Add("formula")
							if formulasArray, ok := k.GetOk(); ok {
								formulasDDArray := make([]datadogV1.WidgetFormula, 0)
								for i := range formulasArray.([]interface{}) {
									k.Add(i)

									formulasDDArrayItem := datadogV1.NewWidgetFormulaWithDefaults()

									if v, ok := k.GetOkWith("alias"); ok {
										formulasDDArrayItem.SetAlias(v.(string))
									}

									if v, ok := k.GetOkWith("formula"); ok {
										formulasDDArrayItem.SetFormula(v.(string))
									}

									// handle limit, which is a nested model
									k.Add("limit.0")

									formulasDDArrayItemLimit := datadogV1.NewWidgetFormulaLimitWithDefaults()

									if v, ok := k.GetOkWith("count"); ok {
										formulasDDArrayItemLimit.SetCount(int64(v.(int)))
									}

									if v, ok := k.GetOkWith("order"); ok {
										formulasDDArrayItemLimit.SetOrder(datadogV1.QuerySortOrder(v.(string)))
									}
									k.Remove("limit.0")
									formulasDDArrayItem.SetLimit(*formulasDDArrayItemLimit)
									formulasDDArray = append(formulasDDArray, *formulasDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetFormulas(formulasDDArray)
							}
							k.Remove("formula")

							// handle log_query, which is a nested model
							k.Add("log_query.0")

							requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemLogQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
							k.Remove("log_query.0")
							requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)

							// handle network_query, which is a nested model
							k.Add("network_query.0")

							requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
							k.Remove("network_query.0")
							requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

							// handle process_query, which is a nested model
							k.Add("process_query.0")

							requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
							k.Add("filter_by")
							if filterByArray, ok := k.GetOk(); ok {
								filterByDDArray := make([]string, 0)
								for i := range filterByArray.([]interface{}) {
									filterByArrayItem := k.GetWith(i)
									filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
								}
								requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
							}
							k.Remove("filter_by")

							if v, ok := k.GetOkWith("limit"); ok {
								requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
							}

							if v, ok := k.GetOkWith("metric"); ok {
								requestsDDArrayItemProcessQuery.SetMetric(v.(string))
							}

							if v, ok := k.GetOkWith("search_by"); ok {
								requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
							}
							k.Remove("process_query.0")
							requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

							// handle profile_metrics_query, which is a nested model
							k.Add("profile_metrics_query.0")

							requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
							k.Remove("profile_metrics_query.0")
							requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

							if v, ok := k.GetOkWith("q"); ok {
								requestsDDArrayItem.SetQ(v.(string))
							}
							k.Add("query")
							if queriesArray, ok := k.GetOk(); ok {
								queriesDDArray := make([]datadogV1.FormulaAndFunctionQueryDefinition, 0)
								for i := range queriesArray.([]interface{}) {
									k.Add(i)

									queriesDDArrayItem := &datadogV1.FormulaAndFunctionQueryDefinition{}
									// handle formula_and_function_query_definition, which is a oneOf model
									k.Add("formula_and_function_metric_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionMetricQueryDefinition := datadogV1.NewFormulaAndFunctionMetricQueryDefinitionWithDefaults()

										if v, ok := k.GetOkWith("aggregator"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionMetricDataSource(v.(string)))
										}

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetName(v.(string))
										}

										if v, ok := k.GetOkWith("query"); ok {
											ddFormulaAndFunctionMetricQueryDefinition.SetQuery(v.(string))
										}
										queriesDDArrayItem.FormulaAndFunctionMetricQueryDefinition = ddFormulaAndFunctionMetricQueryDefinition
									}
									k.Remove("formula_and_function_metric_query_definition.0")
									k.Add("formula_and_function_event_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionEventQueryDefinition := datadogV1.NewFormulaAndFunctionEventQueryDefinitionWithDefaults()

										// handle compute, which is a nested model
										k.Add("compute.0")

										ddFormulaAndFunctionEventQueryDefinitionCompute := datadogV1.NewFormulaAndFunctionEventQueryDefinitionComputeWithDefaults()

										if v, ok := k.GetOkWith("aggregation"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("interval"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetInterval(int64(v.(int)))
										}

										if v, ok := k.GetOkWith("metric"); ok {
											ddFormulaAndFunctionEventQueryDefinitionCompute.SetMetric(v.(string))
										}
										k.Remove("compute.0")
										ddFormulaAndFunctionEventQueryDefinition.SetCompute(*ddFormulaAndFunctionEventQueryDefinitionCompute)

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionEventQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionEventsDataSource(v.(string)))
										}
										k.Add("group_by")
										if groupByArray, ok := k.GetOk(); ok {
											groupByDDArray := make([]datadogV1.FormulaAndFunctionEventQueryGroupBy, 0)
											for i := range groupByArray.([]interface{}) {
												k.Add(i)

												groupByDDArrayItem := datadogV1.NewFormulaAndFunctionEventQueryGroupByWithDefaults()

												if v, ok := k.GetOkWith("facet"); ok {
													groupByDDArrayItem.SetFacet(v.(string))
												}

												if v, ok := k.GetOkWith("limit"); ok {
													groupByDDArrayItem.SetLimit(int64(v.(int)))
												}

												// handle sort, which is a nested model
												k.Add("sort.0")

												groupByDDArrayItemSort := datadogV1.NewFormulaAndFunctionEventQueryGroupBySortWithDefaults()

												if v, ok := k.GetOkWith("aggregation"); ok {
													groupByDDArrayItemSort.SetAggregation(datadogV1.FormulaAndFunctionEventAggregation(v.(string)))
												}

												if v, ok := k.GetOkWith("metric"); ok {
													groupByDDArrayItemSort.SetMetric(v.(string))
												}

												if v, ok := k.GetOkWith("order"); ok {
													groupByDDArrayItemSort.SetOrder(datadogV1.QuerySortOrder(v.(string)))
												}
												k.Remove("sort.0")
												groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
												groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
												k.Remove(i)
											}
											ddFormulaAndFunctionEventQueryDefinition.SetGroupBy(groupByDDArray)
										}
										k.Remove("group_by")
										k.Add("indexes")
										if indexesArray, ok := k.GetOk(); ok {
											indexesDDArray := make([]string, 0)
											for i := range indexesArray.([]interface{}) {
												indexesArrayItem := k.GetWith(i)
												indexesDDArray = append(indexesDDArray, indexesArrayItem.(string))
											}
											ddFormulaAndFunctionEventQueryDefinition.SetIndexes(indexesDDArray)
										}
										k.Remove("indexes")

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionEventQueryDefinition.SetName(v.(string))
										}

										// handle search, which is a nested model
										k.Add("search.0")

										ddFormulaAndFunctionEventQueryDefinitionSearch := datadogV1.NewFormulaAndFunctionEventQueryDefinitionSearchWithDefaults()

										if v, ok := k.GetOkWith("query"); ok {
											ddFormulaAndFunctionEventQueryDefinitionSearch.SetQuery(v.(string))
										}
										k.Remove("search.0")
										ddFormulaAndFunctionEventQueryDefinition.SetSearch(*ddFormulaAndFunctionEventQueryDefinitionSearch)
										queriesDDArrayItem.FormulaAndFunctionEventQueryDefinition = ddFormulaAndFunctionEventQueryDefinition
									}
									k.Remove("formula_and_function_event_query_definition.0")
									k.Add("formula_and_function_process_query_definition.0")
									if _, ok := k.GetOk(); ok {

										ddFormulaAndFunctionProcessQueryDefinition := datadogV1.NewFormulaAndFunctionProcessQueryDefinitionWithDefaults()

										if v, ok := k.GetOkWith("aggregator"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetAggregator(datadogV1.FormulaAndFunctionMetricAggregation(v.(string)))
										}

										if v, ok := k.GetOkWith("data_source"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetDataSource(datadogV1.FormulaAndFunctionProcessQueryDataSource(v.(string)))
										}

										if v, ok := k.GetOkWith("is_normalized_cpu"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetIsNormalizedCpu(v.(bool))
										}

										if v, ok := k.GetOkWith("limit"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetLimit(int64(v.(int)))
										}

										if v, ok := k.GetOkWith("metric"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetMetric(v.(string))
										}

										if v, ok := k.GetOkWith("name"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetName(v.(string))
										}

										if v, ok := k.GetOkWith("sort"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetSort(datadogV1.QuerySortOrder(v.(string)))
										}
										k.Add("tag_filters")
										if tagFiltersArray, ok := k.GetOk(); ok {
											tagFiltersDDArray := make([]string, 0)
											for i := range tagFiltersArray.([]interface{}) {
												tagFiltersArrayItem := k.GetWith(i)
												tagFiltersDDArray = append(tagFiltersDDArray, tagFiltersArrayItem.(string))
											}
											ddFormulaAndFunctionProcessQueryDefinition.SetTagFilters(tagFiltersDDArray)
										}
										k.Remove("tag_filters")

										if v, ok := k.GetOkWith("text_filter"); ok {
											ddFormulaAndFunctionProcessQueryDefinition.SetTextFilter(v.(string))
										}
										queriesDDArrayItem.FormulaAndFunctionProcessQueryDefinition = ddFormulaAndFunctionProcessQueryDefinition
									}
									k.Remove("formula_and_function_process_query_definition.0")

									if queriesDDArrayItem.GetActualInstance() == nil {
										return nil, fmt.Errorf("failed to find valid definition in formula_and_function_query_definition configuration")
									}
									queriesDDArray = append(queriesDDArray, *queriesDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItem.SetQueries(queriesDDArray)
							}
							k.Remove("query")

							if v, ok := k.GetOkWith("response_format"); ok {
								requestsDDArrayItem.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat(v.(string)))
							}

							// handle rum_query, which is a nested model
							k.Add("rum_query.0")

							requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemRumQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
							k.Remove("rum_query.0")
							requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

							// handle security_query, which is a nested model
							k.Add("security_query.0")

							requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
							k.Remove("security_query.0")
							requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

							// handle style, which is a nested model
							k.Add("style.0")

							requestsDDArrayItemStyle := datadogV1.NewWidgetRequestStyleWithDefaults()

							if v, ok := k.GetOkWith("line_type"); ok {
								requestsDDArrayItemStyle.SetLineType(datadogV1.WidgetLineType(v.(string)))
							}

							if v, ok := k.GetOkWith("line_width"); ok {
								requestsDDArrayItemStyle.SetLineWidth(datadogV1.WidgetLineWidth(v.(string)))
							}

							if v, ok := k.GetOkWith("palette"); ok {
								requestsDDArrayItemStyle.SetPalette(v.(string))
							}
							k.Remove("style.0")
							requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
							requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookToplistCellAttributesDefinition.SetRequests(requestsDDArray)
					}
					k.Remove("request")

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookToplistCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookToplistCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookToplistCellAttributesDefinition.SetTime(*ddNotebookToplistCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookToplistCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookToplistCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookToplistCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookToplistCellAttributesDefinition.SetType(datadogV1.ToplistWidgetDefinitionType(v.(string)))
					}
					k.Remove("definition.0")
					ddNotebookToplistCellAttributes.SetDefinition(*ddNotebookToplistCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookToplistCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle split_by, which is a nested model
					k.Add("split_by.0")

					ddNotebookToplistCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
					k.Add("keys")
					if keysArray, ok := k.GetOk(); ok {
						keysDDArray := make([]string, 0)
						for i := range keysArray.([]interface{}) {
							keysArrayItem := k.GetWith(i)
							keysDDArray = append(keysDDArray, keysArrayItem.(string))
						}
						ddNotebookToplistCellAttributesSplitBy.SetKeys(keysDDArray)
					}
					k.Remove("keys")
					k.Add("tags")
					if tagsArray, ok := k.GetOk(); ok {
						tagsDDArray := make([]string, 0)
						for i := range tagsArray.([]interface{}) {
							tagsArrayItem := k.GetWith(i)
							tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
						}
						ddNotebookToplistCellAttributesSplitBy.SetTags(tagsDDArray)
					}
					k.Remove("tags")
					k.Remove("split_by.0")
					ddNotebookToplistCellAttributes.SetSplitBy(*ddNotebookToplistCellAttributesSplitBy)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookToplistCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookToplistCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookToplistCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookToplistCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookToplistCellAttributes.SetTime(*ddNotebookToplistCellAttributesTime)
					ddNotebookCellUpdateRequestAttributes.NotebookToplistCellAttributes = ddNotebookToplistCellAttributes
				}
				k.Remove("notebook_toplist_cell_attributes.0")
				k.Add("notebook_heat_map_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookHeatMapCellAttributes := datadogV1.NewNotebookHeatMapCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookHeatMapCellAttributesDefinition := datadogV1.NewHeatMapWidgetDefinitionWithDefaults()
					k.Add("custom_link")
					if customLinksArray, ok := k.GetOk(); ok {
						customLinksDDArray := make([]datadogV1.WidgetCustomLink, 0)
						for i := range customLinksArray.([]interface{}) {
							k.Add(i)

							customLinksDDArrayItem := datadogV1.NewWidgetCustomLinkWithDefaults()

							if v, ok := k.GetOkWith("is_hidden"); ok {
								customLinksDDArrayItem.SetIsHidden(v.(bool))
							}

							if v, ok := k.GetOkWith("label"); ok {
								customLinksDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("link"); ok {
								customLinksDDArrayItem.SetLink(v.(string))
							}

							if v, ok := k.GetOkWith("override_label"); ok {
								customLinksDDArrayItem.SetOverrideLabel(v.(string))
							}
							customLinksDDArray = append(customLinksDDArray, *customLinksDDArrayItem)
							k.Remove(i)
						}
						ddNotebookHeatMapCellAttributesDefinition.SetCustomLinks(customLinksDDArray)
					}
					k.Remove("custom_link")
					k.Add("event")
					if eventsArray, ok := k.GetOk(); ok {
						eventsDDArray := make([]datadogV1.WidgetEvent, 0)
						for i := range eventsArray.([]interface{}) {
							k.Add(i)

							eventsDDArrayItem := datadogV1.NewWidgetEventWithDefaults()

							if v, ok := k.GetOkWith("q"); ok {
								eventsDDArrayItem.SetQ(v.(string))
							}

							if v, ok := k.GetOkWith("tags_execution"); ok {
								eventsDDArrayItem.SetTagsExecution(v.(string))
							}
							eventsDDArray = append(eventsDDArray, *eventsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookHeatMapCellAttributesDefinition.SetEvents(eventsDDArray)
					}
					k.Remove("event")

					if v, ok := k.GetOkWith("legend_size"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetLegendSize(v.(string))
					}
					k.Add("request")
					if requestsArray, ok := k.GetOk(); ok {
						requestsDDArray := make([]datadogV1.HeatMapWidgetRequest, 0)
						for i := range requestsArray.([]interface{}) {
							k.Add(i)

							requestsDDArrayItem := datadogV1.NewHeatMapWidgetRequestWithDefaults()

							// handle apm_query, which is a nested model
							k.Add("apm_query.0")

							requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemApmQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
							k.Remove("apm_query.0")
							requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)

							// handle event_query, which is a nested model
							k.Add("event_query.0")

							requestsDDArrayItemEventQuery := datadogV1.NewEventQueryDefinitionWithDefaults()

							if v, ok := k.GetOkWith("search"); ok {
								requestsDDArrayItemEventQuery.SetSearch(v.(string))
							}

							if v, ok := k.GetOkWith("tags_execution"); ok {
								requestsDDArrayItemEventQuery.SetTagsExecution(v.(string))
							}
							k.Remove("event_query.0")
							requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)

							// handle log_query, which is a nested model
							k.Add("log_query.0")

							requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemLogQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
							k.Remove("log_query.0")
							requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)

							// handle network_query, which is a nested model
							k.Add("network_query.0")

							requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
							k.Remove("network_query.0")
							requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

							// handle process_query, which is a nested model
							k.Add("process_query.0")

							requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
							k.Add("filter_by")
							if filterByArray, ok := k.GetOk(); ok {
								filterByDDArray := make([]string, 0)
								for i := range filterByArray.([]interface{}) {
									filterByArrayItem := k.GetWith(i)
									filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
								}
								requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
							}
							k.Remove("filter_by")

							if v, ok := k.GetOkWith("limit"); ok {
								requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
							}

							if v, ok := k.GetOkWith("metric"); ok {
								requestsDDArrayItemProcessQuery.SetMetric(v.(string))
							}

							if v, ok := k.GetOkWith("search_by"); ok {
								requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
							}
							k.Remove("process_query.0")
							requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

							// handle profile_metrics_query, which is a nested model
							k.Add("profile_metrics_query.0")

							requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
							k.Remove("profile_metrics_query.0")
							requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

							if v, ok := k.GetOkWith("q"); ok {
								requestsDDArrayItem.SetQ(v.(string))
							}

							// handle rum_query, which is a nested model
							k.Add("rum_query.0")

							requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemRumQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
							k.Remove("rum_query.0")
							requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

							// handle security_query, which is a nested model
							k.Add("security_query.0")

							requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
							k.Remove("security_query.0")
							requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

							// handle style, which is a nested model
							k.Add("style.0")

							requestsDDArrayItemStyle := datadogV1.NewWidgetStyleWithDefaults()

							if v, ok := k.GetOkWith("palette"); ok {
								requestsDDArrayItemStyle.SetPalette(v.(string))
							}
							k.Remove("style.0")
							requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
							requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookHeatMapCellAttributesDefinition.SetRequests(requestsDDArray)
					}
					k.Remove("request")

					if v, ok := k.GetOkWith("show_legend"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetShowLegend(v.(bool))
					}

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookHeatMapCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookHeatMapCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookHeatMapCellAttributesDefinition.SetTime(*ddNotebookHeatMapCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookHeatMapCellAttributesDefinition.SetType(datadogV1.HeatMapWidgetDefinitionType(v.(string)))
					}

					// handle yaxis, which is a nested model
					k.Add("yaxis.0")

					ddNotebookHeatMapCellAttributesDefinitionYaxis := datadogV1.NewWidgetAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("label"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetLabel(v.(string))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookHeatMapCellAttributesDefinitionYaxis.SetScale(v.(string))
					}
					k.Remove("yaxis.0")
					ddNotebookHeatMapCellAttributesDefinition.SetYaxis(*ddNotebookHeatMapCellAttributesDefinitionYaxis)
					k.Remove("definition.0")
					ddNotebookHeatMapCellAttributes.SetDefinition(*ddNotebookHeatMapCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookHeatMapCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle split_by, which is a nested model
					k.Add("split_by.0")

					ddNotebookHeatMapCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
					k.Add("keys")
					if keysArray, ok := k.GetOk(); ok {
						keysDDArray := make([]string, 0)
						for i := range keysArray.([]interface{}) {
							keysArrayItem := k.GetWith(i)
							keysDDArray = append(keysDDArray, keysArrayItem.(string))
						}
						ddNotebookHeatMapCellAttributesSplitBy.SetKeys(keysDDArray)
					}
					k.Remove("keys")
					k.Add("tags")
					if tagsArray, ok := k.GetOk(); ok {
						tagsDDArray := make([]string, 0)
						for i := range tagsArray.([]interface{}) {
							tagsArrayItem := k.GetWith(i)
							tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
						}
						ddNotebookHeatMapCellAttributesSplitBy.SetTags(tagsDDArray)
					}
					k.Remove("tags")
					k.Remove("split_by.0")
					ddNotebookHeatMapCellAttributes.SetSplitBy(*ddNotebookHeatMapCellAttributesSplitBy)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookHeatMapCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookHeatMapCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookHeatMapCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookHeatMapCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookHeatMapCellAttributes.SetTime(*ddNotebookHeatMapCellAttributesTime)
					ddNotebookCellUpdateRequestAttributes.NotebookHeatMapCellAttributes = ddNotebookHeatMapCellAttributes
				}
				k.Remove("notebook_heat_map_cell_attributes.0")
				k.Add("notebook_distribution_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookDistributionCellAttributes := datadogV1.NewNotebookDistributionCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookDistributionCellAttributesDefinition := datadogV1.NewDistributionWidgetDefinitionWithDefaults()

					if v, ok := k.GetOkWith("legend_size"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetLegendSize(v.(string))
					}
					k.Add("marker")
					if markersArray, ok := k.GetOk(); ok {
						markersDDArray := make([]datadogV1.WidgetMarker, 0)
						for i := range markersArray.([]interface{}) {
							k.Add(i)

							markersDDArrayItem := datadogV1.NewWidgetMarkerWithDefaults()

							if v, ok := k.GetOkWith("display_type"); ok {
								markersDDArrayItem.SetDisplayType(v.(string))
							}

							if v, ok := k.GetOkWith("label"); ok {
								markersDDArrayItem.SetLabel(v.(string))
							}

							if v, ok := k.GetOkWith("time"); ok {
								markersDDArrayItem.SetTime(v.(string))
							}

							if v, ok := k.GetOkWith("value"); ok {
								markersDDArrayItem.SetValue(v.(string))
							}
							markersDDArray = append(markersDDArray, *markersDDArrayItem)
							k.Remove(i)
						}
						ddNotebookDistributionCellAttributesDefinition.SetMarkers(markersDDArray)
					}
					k.Remove("marker")
					k.Add("request")
					if requestsArray, ok := k.GetOk(); ok {
						requestsDDArray := make([]datadogV1.DistributionWidgetRequest, 0)
						for i := range requestsArray.([]interface{}) {
							k.Add(i)

							requestsDDArrayItem := datadogV1.NewDistributionWidgetRequestWithDefaults()

							// handle apm_query, which is a nested model
							k.Add("apm_query.0")

							requestsDDArrayItemApmQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemApmQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemApmQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemApmQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemApmQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemApmQuery.SetCompute(*requestsDDArrayItemApmQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemApmQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemApmQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemApmQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemApmQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemApmQuery.SetSearch(*requestsDDArrayItemApmQuerySearch)
							k.Remove("apm_query.0")
							requestsDDArrayItem.SetApmQuery(*requestsDDArrayItemApmQuery)

							// handle event_query, which is a nested model
							k.Add("event_query.0")

							requestsDDArrayItemEventQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemEventQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemEventQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemEventQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemEventQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemEventQuery.SetCompute(*requestsDDArrayItemEventQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemEventQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemEventQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemEventQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemEventQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemEventQuery.SetSearch(*requestsDDArrayItemEventQuerySearch)
							k.Remove("event_query.0")
							requestsDDArrayItem.SetEventQuery(*requestsDDArrayItemEventQuery)

							// handle log_query, which is a nested model
							k.Add("log_query.0")

							requestsDDArrayItemLogQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemLogQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemLogQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemLogQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemLogQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemLogQuery.SetCompute(*requestsDDArrayItemLogQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemLogQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemLogQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemLogQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemLogQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemLogQuery.SetSearch(*requestsDDArrayItemLogQuerySearch)
							k.Remove("log_query.0")
							requestsDDArrayItem.SetLogQuery(*requestsDDArrayItemLogQuery)

							// handle network_query, which is a nested model
							k.Add("network_query.0")

							requestsDDArrayItemNetworkQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemNetworkQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemNetworkQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemNetworkQuery.SetCompute(*requestsDDArrayItemNetworkQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemNetworkQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemNetworkQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemNetworkQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemNetworkQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemNetworkQuery.SetSearch(*requestsDDArrayItemNetworkQuerySearch)
							k.Remove("network_query.0")
							requestsDDArrayItem.SetNetworkQuery(*requestsDDArrayItemNetworkQuery)

							// handle process_query, which is a nested model
							k.Add("process_query.0")

							requestsDDArrayItemProcessQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
							k.Add("filter_by")
							if filterByArray, ok := k.GetOk(); ok {
								filterByDDArray := make([]string, 0)
								for i := range filterByArray.([]interface{}) {
									filterByArrayItem := k.GetWith(i)
									filterByDDArray = append(filterByDDArray, filterByArrayItem.(string))
								}
								requestsDDArrayItemProcessQuery.SetFilterBy(filterByDDArray)
							}
							k.Remove("filter_by")

							if v, ok := k.GetOkWith("limit"); ok {
								requestsDDArrayItemProcessQuery.SetLimit(int64(v.(int)))
							}

							if v, ok := k.GetOkWith("metric"); ok {
								requestsDDArrayItemProcessQuery.SetMetric(v.(string))
							}

							if v, ok := k.GetOkWith("search_by"); ok {
								requestsDDArrayItemProcessQuery.SetSearchBy(v.(string))
							}
							k.Remove("process_query.0")
							requestsDDArrayItem.SetProcessQuery(*requestsDDArrayItemProcessQuery)

							// handle profile_metrics_query, which is a nested model
							k.Add("profile_metrics_query.0")

							requestsDDArrayItemProfileMetricsQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemProfileMetricsQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemProfileMetricsQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemProfileMetricsQuery.SetCompute(*requestsDDArrayItemProfileMetricsQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemProfileMetricsQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemProfileMetricsQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemProfileMetricsQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemProfileMetricsQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemProfileMetricsQuery.SetSearch(*requestsDDArrayItemProfileMetricsQuerySearch)
							k.Remove("profile_metrics_query.0")
							requestsDDArrayItem.SetProfileMetricsQuery(*requestsDDArrayItemProfileMetricsQuery)

							if v, ok := k.GetOkWith("q"); ok {
								requestsDDArrayItem.SetQ(v.(string))
							}

							// handle rum_query, which is a nested model
							k.Add("rum_query.0")

							requestsDDArrayItemRumQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemRumQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemRumQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemRumQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemRumQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemRumQuery.SetCompute(*requestsDDArrayItemRumQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemRumQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemRumQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemRumQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemRumQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemRumQuery.SetSearch(*requestsDDArrayItemRumQuerySearch)
							k.Remove("rum_query.0")
							requestsDDArrayItem.SetRumQuery(*requestsDDArrayItemRumQuery)

							// handle security_query, which is a nested model
							k.Add("security_query.0")

							requestsDDArrayItemSecurityQuery := datadogV1.NewLogQueryDefinitionWithDefaults()

							// handle compute, which is a nested model
							k.Add("compute.0")

							requestsDDArrayItemSecurityQueryCompute := datadogV1.NewLogsQueryComputeWithDefaults()

							if v, ok := k.GetOkWith("aggregation"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetAggregation(v.(string))
							}

							if v, ok := k.GetOkWith("facet"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetFacet(v.(string))
							}

							if v, ok := k.GetOkWith("interval"); ok {
								requestsDDArrayItemSecurityQueryCompute.SetInterval(int64(v.(int)))
							}
							k.Remove("compute.0")
							requestsDDArrayItemSecurityQuery.SetCompute(*requestsDDArrayItemSecurityQueryCompute)
							k.Add("group_by")
							if groupByArray, ok := k.GetOk(); ok {
								groupByDDArray := make([]datadogV1.LogQueryDefinitionGroupBy, 0)
								for i := range groupByArray.([]interface{}) {
									k.Add(i)

									groupByDDArrayItem := datadogV1.NewLogQueryDefinitionGroupByWithDefaults()

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("limit"); ok {
										groupByDDArrayItem.SetLimit(int64(v.(int)))
									}

									// handle sort, which is a nested model
									k.Add("sort.0")

									groupByDDArrayItemSort := datadogV1.NewLogQueryDefinitionGroupBySortWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										groupByDDArrayItemSort.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										groupByDDArrayItemSort.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("order"); ok {
										groupByDDArrayItemSort.SetOrder(datadogV1.WidgetSort(v.(string)))
									}
									k.Remove("sort.0")
									groupByDDArrayItem.SetSort(*groupByDDArrayItemSort)
									groupByDDArray = append(groupByDDArray, *groupByDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetGroupBy(groupByDDArray)
							}
							k.Remove("group_by")

							if v, ok := k.GetOkWith("index"); ok {
								requestsDDArrayItemSecurityQuery.SetIndex(v.(string))
							}
							k.Add("multi_compute")
							if multiComputeArray, ok := k.GetOk(); ok {
								multiComputeDDArray := make([]datadogV1.LogsQueryCompute, 0)
								for i := range multiComputeArray.([]interface{}) {
									k.Add(i)

									multiComputeDDArrayItem := datadogV1.NewLogsQueryComputeWithDefaults()

									if v, ok := k.GetOkWith("aggregation"); ok {
										multiComputeDDArrayItem.SetAggregation(v.(string))
									}

									if v, ok := k.GetOkWith("facet"); ok {
										multiComputeDDArrayItem.SetFacet(v.(string))
									}

									if v, ok := k.GetOkWith("interval"); ok {
										multiComputeDDArrayItem.SetInterval(int64(v.(int)))
									}
									multiComputeDDArray = append(multiComputeDDArray, *multiComputeDDArrayItem)
									k.Remove(i)
								}
								requestsDDArrayItemSecurityQuery.SetMultiCompute(multiComputeDDArray)
							}
							k.Remove("multi_compute")

							// handle search, which is a nested model
							k.Add("search.0")

							requestsDDArrayItemSecurityQuerySearch := datadogV1.NewLogQueryDefinitionSearchWithDefaults()

							if v, ok := k.GetOkWith("query"); ok {
								requestsDDArrayItemSecurityQuerySearch.SetQuery(v.(string))
							}
							k.Remove("search.0")
							requestsDDArrayItemSecurityQuery.SetSearch(*requestsDDArrayItemSecurityQuerySearch)
							k.Remove("security_query.0")
							requestsDDArrayItem.SetSecurityQuery(*requestsDDArrayItemSecurityQuery)

							// handle style, which is a nested model
							k.Add("style.0")

							requestsDDArrayItemStyle := datadogV1.NewWidgetStyleWithDefaults()

							if v, ok := k.GetOkWith("palette"); ok {
								requestsDDArrayItemStyle.SetPalette(v.(string))
							}
							k.Remove("style.0")
							requestsDDArrayItem.SetStyle(*requestsDDArrayItemStyle)
							requestsDDArray = append(requestsDDArray, *requestsDDArrayItem)
							k.Remove(i)
						}
						ddNotebookDistributionCellAttributesDefinition.SetRequests(requestsDDArray)
					}
					k.Remove("request")

					if v, ok := k.GetOkWith("show_legend"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetShowLegend(v.(bool))
					}

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookDistributionCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookDistributionCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookDistributionCellAttributesDefinition.SetTime(*ddNotebookDistributionCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookDistributionCellAttributesDefinition.SetType(datadogV1.DistributionWidgetDefinitionType(v.(string)))
					}

					// handle xaxis, which is a nested model
					k.Add("xaxis.0")

					ddNotebookDistributionCellAttributesDefinitionXaxis := datadogV1.NewDistributionWidgetXAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookDistributionCellAttributesDefinitionXaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookDistributionCellAttributesDefinitionXaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookDistributionCellAttributesDefinitionXaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookDistributionCellAttributesDefinitionXaxis.SetScale(v.(string))
					}
					k.Remove("xaxis.0")
					ddNotebookDistributionCellAttributesDefinition.SetXaxis(*ddNotebookDistributionCellAttributesDefinitionXaxis)

					// handle yaxis, which is a nested model
					k.Add("yaxis.0")

					ddNotebookDistributionCellAttributesDefinitionYaxis := datadogV1.NewDistributionWidgetYAxisWithDefaults()

					if v, ok := k.GetOkWith("include_zero"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetIncludeZero(v.(bool))
					}

					if v, ok := k.GetOkWith("label"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetLabel(v.(string))
					}

					if v, ok := k.GetOkWith("max"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetMax(v.(string))
					}

					if v, ok := k.GetOkWith("min"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetMin(v.(string))
					}

					if v, ok := k.GetOkWith("scale"); ok {
						ddNotebookDistributionCellAttributesDefinitionYaxis.SetScale(v.(string))
					}
					k.Remove("yaxis.0")
					ddNotebookDistributionCellAttributesDefinition.SetYaxis(*ddNotebookDistributionCellAttributesDefinitionYaxis)
					k.Remove("definition.0")
					ddNotebookDistributionCellAttributes.SetDefinition(*ddNotebookDistributionCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookDistributionCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle split_by, which is a nested model
					k.Add("split_by.0")

					ddNotebookDistributionCellAttributesSplitBy := datadogV1.NewNotebookSplitByWithDefaults()
					k.Add("keys")
					if keysArray, ok := k.GetOk(); ok {
						keysDDArray := make([]string, 0)
						for i := range keysArray.([]interface{}) {
							keysArrayItem := k.GetWith(i)
							keysDDArray = append(keysDDArray, keysArrayItem.(string))
						}
						ddNotebookDistributionCellAttributesSplitBy.SetKeys(keysDDArray)
					}
					k.Remove("keys")
					k.Add("tags")
					if tagsArray, ok := k.GetOk(); ok {
						tagsDDArray := make([]string, 0)
						for i := range tagsArray.([]interface{}) {
							tagsArrayItem := k.GetWith(i)
							tagsDDArray = append(tagsDDArray, tagsArrayItem.(string))
						}
						ddNotebookDistributionCellAttributesSplitBy.SetTags(tagsDDArray)
					}
					k.Remove("tags")
					k.Remove("split_by.0")
					ddNotebookDistributionCellAttributes.SetSplitBy(*ddNotebookDistributionCellAttributesSplitBy)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookDistributionCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookDistributionCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookDistributionCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookDistributionCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookDistributionCellAttributes.SetTime(*ddNotebookDistributionCellAttributesTime)
					ddNotebookCellUpdateRequestAttributes.NotebookDistributionCellAttributes = ddNotebookDistributionCellAttributes
				}
				k.Remove("notebook_distribution_cell_attributes.0")
				k.Add("notebook_log_stream_cell_attributes.0")
				if _, ok := k.GetOk(); ok {

					ddNotebookLogStreamCellAttributes := datadogV1.NewNotebookLogStreamCellAttributesWithDefaults()

					// handle definition, which is a nested model
					k.Add("definition.0")

					ddNotebookLogStreamCellAttributesDefinition := datadogV1.NewLogStreamWidgetDefinitionWithDefaults()
					k.Add("columns")
					if columnsArray, ok := k.GetOk(); ok {
						columnsDDArray := make([]string, 0)
						for i := range columnsArray.([]interface{}) {
							columnsArrayItem := k.GetWith(i)
							columnsDDArray = append(columnsDDArray, columnsArrayItem.(string))
						}
						ddNotebookLogStreamCellAttributesDefinition.SetColumns(columnsDDArray)
					}
					k.Remove("columns")
					k.Add("indexes")
					if indexesArray, ok := k.GetOk(); ok {
						indexesDDArray := make([]string, 0)
						for i := range indexesArray.([]interface{}) {
							indexesArrayItem := k.GetWith(i)
							indexesDDArray = append(indexesDDArray, indexesArrayItem.(string))
						}
						ddNotebookLogStreamCellAttributesDefinition.SetIndexes(indexesDDArray)
					}
					k.Remove("indexes")

					if v, ok := k.GetOkWith("logset"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetLogset(v.(string))
					}

					if v, ok := k.GetOkWith("message_display"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetMessageDisplay(datadogV1.WidgetMessageDisplay(v.(string)))
					}

					if v, ok := k.GetOkWith("query"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetQuery(v.(string))
					}

					if v, ok := k.GetOkWith("show_date_column"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetShowDateColumn(v.(bool))
					}

					if v, ok := k.GetOkWith("show_message_column"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetShowMessageColumn(v.(bool))
					}

					// handle sort, which is a nested model
					k.Add("sort.0")

					ddNotebookLogStreamCellAttributesDefinitionSort := datadogV1.NewWidgetFieldSortWithDefaults()

					if v, ok := k.GetOkWith("column"); ok {
						ddNotebookLogStreamCellAttributesDefinitionSort.SetColumn(v.(string))
					}

					if v, ok := k.GetOkWith("order"); ok {
						ddNotebookLogStreamCellAttributesDefinitionSort.SetOrder(datadogV1.WidgetSort(v.(string)))
					}
					k.Remove("sort.0")
					ddNotebookLogStreamCellAttributesDefinition.SetSort(*ddNotebookLogStreamCellAttributesDefinitionSort)

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookLogStreamCellAttributesDefinitionTime := datadogV1.NewWidgetTimeWithDefaults()

					if v, ok := k.GetOkWith("live_span"); ok {
						ddNotebookLogStreamCellAttributesDefinitionTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
					}
					k.Remove("time.0")
					ddNotebookLogStreamCellAttributesDefinition.SetTime(*ddNotebookLogStreamCellAttributesDefinitionTime)

					if v, ok := k.GetOkWith("title"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetTitle(v.(string))
					}

					if v, ok := k.GetOkWith("title_align"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v.(string)))
					}

					if v, ok := k.GetOkWith("title_size"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetTitleSize(v.(string))
					}

					if v, ok := k.GetOkWith("type"); ok {
						ddNotebookLogStreamCellAttributesDefinition.SetType(datadogV1.LogStreamWidgetDefinitionType(v.(string)))
					}
					k.Remove("definition.0")
					ddNotebookLogStreamCellAttributes.SetDefinition(*ddNotebookLogStreamCellAttributesDefinition)

					if v, ok := k.GetOkWith("graph_size"); ok {
						ddNotebookLogStreamCellAttributes.SetGraphSize(datadogV1.NotebookGraphSize(v.(string)))
					}

					// handle time, which is a nested model
					k.Add("time.0")

					ddNotebookLogStreamCellAttributesTime := &datadogV1.NotebookCellTime{}
					// handle notebook_cell_time, which is a oneOf model
					k.Add("notebook_relative_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

						if v, ok := k.GetOkWith("live_span"); ok {
							ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
						}
						ddNotebookLogStreamCellAttributesTime.NotebookRelativeTime = ddNotebookRelativeTime
					}
					k.Remove("notebook_relative_time.0")
					k.Add("notebook_absolute_time.0")
					if _, ok := k.GetOk(); ok {

						ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
						// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

						if v, ok := k.GetOkWith("live"); ok {
							ddNotebookAbsoluteTime.SetLive(v.(bool))
						}
						// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
						ddNotebookLogStreamCellAttributesTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
					}
					k.Remove("notebook_absolute_time.0")

					if ddNotebookLogStreamCellAttributesTime.GetActualInstance() == nil {
						return nil, fmt.Errorf("failed to find valid definition in notebook_cell_time configuration")
					}
					k.Remove("time.0")
					ddNotebookLogStreamCellAttributes.SetTime(*ddNotebookLogStreamCellAttributesTime)
					ddNotebookCellUpdateRequestAttributes.NotebookLogStreamCellAttributes = ddNotebookLogStreamCellAttributes
				}
				k.Remove("notebook_log_stream_cell_attributes.0")

				if ddNotebookCellUpdateRequestAttributes.GetActualInstance() == nil {
					return nil, fmt.Errorf("failed to find valid definition in notebook_cell_update_request_attributes configuration")
				}
				k.Remove("attributes.0")
				ddNotebookCellUpdateRequest.SetAttributes(*ddNotebookCellUpdateRequestAttributes)

				if v, ok := k.GetOkWith("id"); ok {
					ddNotebookCellUpdateRequest.SetId(v.(string))
				}

				if v, ok := k.GetOkWith("type"); ok {
					ddNotebookCellUpdateRequest.SetType(datadogV1.NotebookCellResourceType(v.(string)))
				}
				cellsDDArrayItem.NotebookCellUpdateRequest = ddNotebookCellUpdateRequest
			}
			k.Remove("notebook_cell_update_request.0")

			if cellsDDArrayItem.GetActualInstance() == nil {
				return nil, fmt.Errorf("failed to find valid definition in notebook_update_cell configuration")
			}
			cellsDDArray = append(cellsDDArray, *cellsDDArrayItem)
			k.Remove(i)
		}
		result.SetCells(cellsDDArray)
	}
	k.Remove("cell")

	if v, ok := k.GetOkWith("name"); ok {
		result.SetName(v.(string))
	}

	if v, ok := k.GetOkWith("status"); ok {
		result.SetStatus(datadogV1.NotebookStatus(v.(string)))
	}

	// handle time, which is a nested model
	k.Add("time.0")

	resultTime := &datadogV1.NotebookGlobalTime{}
	// handle notebook_global_time, which is a oneOf model
	k.Add("notebook_relative_time.0")
	if _, ok := k.GetOk(); ok {

		ddNotebookRelativeTime := datadogV1.NewNotebookRelativeTimeWithDefaults()

		if v, ok := k.GetOkWith("live_span"); ok {
			ddNotebookRelativeTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v.(string)))
		}
		resultTime.NotebookRelativeTime = ddNotebookRelativeTime
	}
	k.Remove("notebook_relative_time.0")
	k.Add("notebook_absolute_time.0")
	if _, ok := k.GetOk(); ok {

		ddNotebookAbsoluteTime := datadogV1.NewNotebookAbsoluteTimeWithDefaults()
		// FIXME: date-time value handling not implemented yet; please implement handling "end" manually

		if v, ok := k.GetOkWith("live"); ok {
			ddNotebookAbsoluteTime.SetLive(v.(bool))
		}
		// FIXME: date-time value handling not implemented yet; please implement handling "start" manually
		resultTime.NotebookAbsoluteTime = ddNotebookAbsoluteTime
	}
	k.Remove("notebook_absolute_time.0")

	if resultTime.GetActualInstance() == nil {
		return nil, fmt.Errorf("failed to find valid definition in notebook_global_time configuration")
	}
	k.Remove("time.0")
	result.SetTime(*resultTime)
	return result, nil
}

func resourceDatadogNotebookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	resultNotebookUpdateDataAttributes, err := buildDatadogNotebookUpdate(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error building Notebook object: %s", err))
	}

	resultNotebookUpdateData := datadogV1.NewNotebookUpdateDataWithDefaults()
	resultNotebookUpdateData.SetAttributes(*resultNotebookUpdateDataAttributes)

	ddObject := datadogV1.NewNotebookUpdateRequestWithDefaults()
	ddObject.SetData(*resultNotebookUpdateData)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceNotebookResponse, _, err := datadogClient.NotebooksApi.UpdateNotebook(auth, id, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, "error creating Notebook")
	}

	resourceNotebookResponseData := resourceNotebookResponse.GetData()

	resource := resourceNotebookResponseData.GetAttributes()

	return updateNotebookTerraformState(d, resource)
}

func resourceDatadogNotebookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1
	var err error

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = datadogClient.NotebooksApi.DeleteNotebook(auth, id)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, "error deleting Notebook")
	}

	return nil
}
