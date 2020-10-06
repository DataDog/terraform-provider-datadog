package datadog

import (
	"testing"
)

// JSON export used as test scenario
//{
//    "notify_list": [],
//    "description": null,
//    "author_name": "--redacted--",
//    "id": "--redacted--",
//    "url": "--redacted--",
//    "template_variables": [],
//    "is_read_only": false,
//    "title": "TF - Timeseries example",
//    "created_at": "2020-03-12T15:04:00.466540+00:00",
//    "modified_at": "2020-06-09T10:15:58.451756+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "yaxis": {
//                    "max": "599999"
//                },
//                "title_align": "left",
//                "markers": [
//                    {
//                        "display_type": "error dashed",
//                        "value": "y = 500000",
//                        "label": "y = 500000"
//                    }
//                ],
//                "show_legend": true,
//                "requests": [
//                    {
//                        "q": "avg:system.cpu.user{env:prod} by {app}",
//                        "style": {
//                            "line_width": "thin",
//                            "palette": "dog_classic",
//                            "line_type": "solid"
//                        },
//                        "display_type": "line"
//                    },
//                    {
//                        "style": {
//                            "line_width": "normal",
//                            "palette": "cool",
//                            "line_type": "solid"
//                        },
//                        "display_type": "line",
//                        "log_query": {
//                            "index": "*",
//                            "search": {
//                                "query": ""
//                            },
//                            "group_by": [
//                                {
//                                    "facet": "service",
//                                    "sort": {
//                                        "aggregation": "count",
//                                        "order": "desc"
//                                    },
//                                    "limit": 10
//                                }
//                            ],
//                            "compute": {
//                                "aggregation": "count"
//                            }
//                        }
//                    },
//                    {
//                        "style": {
//                            "line_width": "thick",
//                            "palette": "warm",
//                            "line_type": "dashed"
//                        },
//                        "apm_query": {
//                            "index": "trace-search",
//                            "search": {
//                                "query": ""
//                            },
//                            "group_by": [
//                                {
//                                    "facet": "status",
//                                    "sort": {
//                                        "facet": "env",
//                                        "aggregation": "cardinality",
//                                        "order": "desc"
//                                    },
//                                    "limit": 10
//                                }
//                            ],
//                            "compute": {
//                                "facet": "env",
//                                "interval": 1000,
//                                "aggregation": "cardinality"
//                            }
//                        },
//                        "display_type": "line"
//                    },
//                    {
//                        "style": {
//                            "line_width": "normal",
//                            "palette": "purple",
//                            "line_type": "solid"
//                        },
//                        "process_query": {
//                            "search_by": "",
//                            "metric": "process.stat.cpu.total_pct.norm",
//                            "limit": 10,
//                            "filter_by": [
//                                "account:prod"
//                            ]
//                        },
//                        "display_type": "line"
//                    },
//                    {
//                        "style": {
//                            "line_width": "normal",
//                            "palette": "orange",
//                            "line_type": "solid"
//                        },
//                        "display_type": "area",
//                        "network_query": {
//                            "index": "netflow-search",
//                            "search": {
//                                "query": "network.transport:udp network.destination.ip:\"*\""
//                            },
//                            "group_by": [
//                                {
//                                    "facet": "source_region"
//                                },
//                                {
//                                    "facet": "dest_environment"
//                                }
//                            ],
//                            "compute": {
//                                "facet": "network.bytes_read",
//                                "aggregation": "sum"
//                            }
//                        }
//                    },
//                    {
//                        "style": {
//                            "line_width": "normal",
//                            "palette": "grey",
//                            "line_type": "solid"
//                        },
//                        "rum_query": {
//                            "index": "*",
//                            "search": {
//                                "query": ""
//                            },
//                            "group_by": [
//                                {
//                                    "facet": "service",
//                                    "sort": {
//                                        "facet": "@duration",
//                                        "aggregation": "avg",
//                                        "order": "desc"
//                                    },
//                                    "limit": 10
//                                }
//                            ],
//                            "compute": {
//                                "facet": "@duration",
//                                "interval": 10,
//                                "aggregation": "avg"
//                            }
//                        },
//                        "display_type": "area"
//                    },
//                    {
//                        "style": {
//                            "line_width": "normal",
//                            "palette": "red",
//                            "line_type": "solid"
//                        },
//                        "display_type": "line",
//                        "profilemetrics_query": {
//                            "index": "*",
//                            "search": {
//                                "query": ""
//                            },
//                            "group_by": [
//                                {
//                                    "facet": "language",
//                                    "sort": {
//                                        "aggregation": "avg",
//                                        "order": "desc"
//                                    },
//                                    "limit": 10
//                                }
//                            ],
//                            "compute": {
//                                "interval": 300000,
//                                "aggregation": "avg"
//                            }
//                        }
//                    },
//                    {
//                        "style": {
//                            "line_width": "thin",
//                            "palette": "green",
//                            "line_type": "dotted"
//                        },
//                        "security_query": {
//                            "index": "*",
//                            "search": {
//                                "query": ""
//                            },
//                            "group_by": [
//                                {
//                                    "facet": "service",
//                                    "sort": {
//                                        "facet": "status",
//                                        "aggregation": "cardinality",
//                                        "order": "desc"
//                                    },
//                                    "limit": 10
//                                }
//                            ],
//                            "compute": {
//                                "facet": "status",
//                                "aggregation": "cardinality"
//                            }
//                        },
//                        "display_type": "bars"
//                    }
//                ],
//                "time": {
//                    "live_span": "5m"
//                },
//                "title": "system.cpu.user, env, process.stat.cpu.total_pct.norm, network.bytes_read, @d...",
//                "legend_size": "2",
//                "type": "timeseries",
//                "events": [
//                    {
//                        "q": "sources:test tags:1",
//                        "tags_execution": "and"
//                    }
//                ]
//            },
//            "layout": {
//                "y": 2,
//                "x": 1,
//                "height": 15,
//                "width": 47
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardTimeseriesConfig = `
resource "datadog_dashboard" "timeseries_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		timeseries_definition {
			title_size = "16"
			title_align = "left"
			show_legend = "true"
			title = "system.cpu.user, env, process.stat.cpu.total_pct.norm, network.bytes_read, @d..."
			legend_size = "2"
			yaxis {
				label = ""
				min = "0"
				include_zero = "true"
				max = "599999"
				scale = ""
			}
			right_yaxis {
				label = ""
				min = "1"
				include_zero = "false"
				max = "599998"
				scale = ""
			}
			marker {
				display_type = "error dashed"
				value = "y=500000"
				label = "y=500000"
			}
			marker {
				display_type = "warning dashed"
				value = "y=400000"
				label = "y=400000"
			}
			time = {
				live_span = "5m"
			}
			event {
				q = "sources:test tags:1"
				tags_execution = "and"
			}
			request {
				q = "avg:system.cpu.user{env:prod} by {app}"
				style {
					line_width = "thin"
					palette = "dog_classic"
					line_type = "solid"
				}
				display_type = "line"
			}
			request {
				style {
					line_width = "normal"
					palette = "cool"
					line_type = "solid"
				}
				display_type = "line"
				log_query {
					index = "*"
					search = {
						query = ""
					}
					group_by {
						facet = "service"
						sort = {
							aggregation = "count"
							order = "desc"
						}
						limit = "10"
					}
					compute = {
						aggregation = "count"
					}
				}
			}
			request {
				style {
					line_width = "thick"
					palette = "warm"
					line_type = "dashed"
				}
				apm_query {
					index = "trace-search"
					search = {
						query = ""
					}
					group_by {
						facet = "status"
						sort = {
							facet = "env"
							aggregation = "cardinality"
							order = "desc"
						}
						limit = "10"
					}
					compute = {
						facet = "env"
						interval = "1000"
						aggregation = "cardinality"
					}
				}
				display_type = "line"
			}
			request {
				style {
					line_width = "normal"
					palette = "purple"
					line_type = "solid"
				}
				process_query {
					search_by = ""
					metric = "process.stat.cpu.total_pct.norm"
					limit = "10"
					filter_by = ["account:prod"]
				}
				display_type = "line"
			}
			request {
				style {
					line_width = "normal"
					palette = "orange"
					line_type = "solid"
				}
				display_type = "area"
				network_query {
					index = "netflow-search"
					search = {
						query = "network.transport:udp network.destination.ip:\"*\""
					}
					group_by {
						facet = "source_region"
					}
					group_by {
						facet = "dest_environment"
					}
					compute = {
						facet = "network.bytes_read"
						aggregation = "sum"
					}
				}
			}
			request {
				style {
					line_width = "normal"
					palette = "grey"
					line_type = "solid"
				}
				rum_query {
					index = "*"
					search = {
						query = ""
					}
					group_by {
						facet = "service"
						sort = {
							facet = "@duration"
							aggregation = "avg"
							order = "desc"
						}
						limit = "10"
					}
					compute = {
						facet = "@duration"
						interval = "10"
						aggregation = "avg"
					}
				}
				display_type = "area"
			}
			//request {
			//	style {
			//		line_width = "normal"
			//		palette = "red"
			//		line_type = "solid"
			//	}
			//	display_type = "line"
			//	profilemetrics_query {
			//		index = "*"
			//		search {
			//			query = ""
			//		}
			//		group_by {
			//			facet = "language"
			//			sort = {
			//				aggregation = "avg"
			//				order = "desc"
			//			}
			//			limit = "10"
			//		}
			//		compute {
			//			interval = "300000"
			//			aggregation = "avg"
			//		}
			//	}
			//}
			//request {
			//	style {
			//		line_width = "thin"
			//		palette = "green"
			//		line_type = "dotted"
			//	}
			//	security_query {
			//		index = "*"
			//		search {
			//			query = ""
			//		}
			//		group_by {
			//			facet = "service"
			//			sort = {
			//				facet = "status"
			//				aggregation = "cardinality"
			//				order = "desc"
			//			}
			//			limit = "10"
			//		}
			//		compute {
			//			facet = "status"
			//			aggregation = "cardinality"
			//		}
			//	}
			//	display_type = "bars"
			//}
		}
	}
}
`

var datadogDashboardTimeseriesAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.timeseries_definition.0.show_legend = true",
	"widget.0.timeseries_definition.0.yaxis.0.min = 0",
	"widget.0.timeseries_definition.0.yaxis.0.max = 599999",
	"widget.0.timeseries_definition.0.yaxis.0.label =",
	"widget.0.timeseries_definition.0.yaxis.0.include_zero = true",
	"widget.0.timeseries_definition.0.yaxis.0.scale =",
	"widget.0.timeseries_definition.0.right_yaxis.0.min = 1",
	"widget.0.timeseries_definition.0.right_yaxis.0.max = 599998",
	"widget.0.timeseries_definition.0.right_yaxis.0.label =",
	"widget.0.timeseries_definition.0.right_yaxis.0.include_zero = false",
	"widget.0.timeseries_definition.0.right_yaxis.0.scale =",
	"widget.0.timeseries_definition.0.legend_size = 2",
	"widget.0.timeseries_definition.0.time.live_span = 5m",
	"widget.0.timeseries_definition.0.title_align = left",
	"widget.0.timeseries_definition.0.title = system.cpu.user, env, process.stat.cpu.total_pct.norm, network.bytes_read, @d...",
	"widget.0.timeseries_definition.0.title_size = 16",
	"widget.0.timeseries_definition.0.event.0.q = sources:test tags:1",
	"widget.0.timeseries_definition.0.event.0.tags_execution = and",
	"widget.0.timeseries_definition.0.marker.# = 2",
	"widget.0.timeseries_definition.0.marker.0.label = y=500000",
	"widget.0.timeseries_definition.0.marker.0.value = y=500000",
	"widget.0.timeseries_definition.0.marker.0.display_type = error dashed",
	"widget.0.timeseries_definition.0.marker.1.label = y=400000",
	"widget.0.timeseries_definition.0.marker.1.display_type = warning dashed",
	"widget.0.timeseries_definition.0.marker.1.value = y=400000",
	"widget.0.timeseries_definition.0.request.# = 6",
	"widget.0.timeseries_definition.0.request.0.style.0.line_width = thin",
	"widget.0.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.0.process_query.# = 0",
	"widget.0.timeseries_definition.0.request.0.metadata.# = 0",
	"widget.0.timeseries_definition.0.request.0.log_query.# = 0",
	"widget.0.timeseries_definition.0.request.0.display_type = line",
	"widget.0.timeseries_definition.0.request.0.style.# = 1",
	"widget.0.timeseries_definition.0.request.0.apm_query.# = 0",
	"widget.0.timeseries_definition.0.request.0.style.0.palette = dog_classic",
	"widget.0.timeseries_definition.0.request.0.q = avg:system.cpu.user{env:prod} by {app}",
	"widget.0.timeseries_definition.0.request.1.log_query.0.index = *",
	"widget.0.timeseries_definition.0.request.1.style.# = 1",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.sort.aggregation = count",
	"widget.0.timeseries_definition.0.request.1.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.1.log_query.0.search.query =",
	"widget.0.timeseries_definition.0.request.1.style.0.palette = cool",
	"widget.0.timeseries_definition.0.request.1.log_query.0.compute.% = 1",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.facet = service",
	"widget.0.timeseries_definition.0.request.1.log_query.0.compute.aggregation = count",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.sort.order = desc",
	"widget.0.timeseries_definition.0.request.1.metadata.# = 0",
	"widget.0.timeseries_definition.0.request.1.q =",
	"widget.0.timeseries_definition.0.request.1.log_query.0.search.% = 1",
	"widget.0.timeseries_definition.0.request.1.apm_query.# = 0",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.# = 1",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.1.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.sort.% = 2",
	"widget.0.timeseries_definition.0.request.1.process_query.# = 0",
	"widget.0.timeseries_definition.0.request.1.display_type = line",
	"widget.0.timeseries_definition.0.request.1.log_query.# = 1",
	"widget.0.timeseries_definition.0.request.3.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.3.process_query.0.metric = process.stat.cpu.total_pct.norm",
	"widget.0.timeseries_definition.0.request.2.style.0.line_type = dashed",
	"widget.0.timeseries_definition.0.request.2.display_type = line",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.facet = status",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.# = 1",
	"widget.0.timeseries_definition.0.request.2.apm_query.# = 1",
	"widget.0.timeseries_definition.0.request.2.process_query.# = 0",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.order = desc",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.search.query =",
	"widget.0.timeseries_definition.0.request.2.log_query.# = 0",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.compute.interval = 1000",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.compute.% = 3",
	"widget.0.timeseries_definition.0.request.2.metadata.# = 0",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.search.% = 1",
	"widget.0.timeseries_definition.0.request.2.style.0.line_width = thick",
	"widget.0.timeseries_definition.0.request.2.q =",
	"widget.0.timeseries_definition.0.request.2.style.0.palette = warm",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.% = 3",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.compute.facet = env",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.2.style.# = 1",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.aggregation = cardinality",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.compute.aggregation = cardinality",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.facet = env",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.index = trace-search",
	"widget.0.timeseries_definition.0.request.3.log_query.# = 0",
	"widget.0.timeseries_definition.0.request.3.process_query.0.search_by =",
	"widget.0.timeseries_definition.0.request.3.style.# = 1",
	"widget.0.timeseries_definition.0.request.3.metadata.# = 0",
	"widget.0.timeseries_definition.0.request.3.process_query.0.limit = 10",
	"widget.0.timeseries_definition.0.request.3.process_query.# = 1",
	"widget.0.timeseries_definition.0.request.3.process_query.0.filter_by.0 = account:prod",
	"widget.0.timeseries_definition.0.request.3.process_query.0.filter_by.# = 1",
	"widget.0.timeseries_definition.0.request.3.q =",
	"widget.0.timeseries_definition.0.request.3.display_type = line",
	"widget.0.timeseries_definition.0.request.3.apm_query.# = 0",
	"widget.0.timeseries_definition.0.request.3.style.0.palette = purple",
	"widget.0.timeseries_definition.0.request.3.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.sort.% = 3",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.0.facet = source_region",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.1.sort.% = 0",
	"widget.0.timeseries_definition.0.request.4.network_query.0.compute.% = 2",
	"widget.0.timeseries_definition.0.request.4.network_query.0.compute.facet = network.bytes_read",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.search.% = 1",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.4.network_query.0.search.% = 1",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.1.limit = 0",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.compute.facet = @duration",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.1.facet = dest_environment",
	"widget.0.timeseries_definition.0.request.4.network_query.0.search.query = network.transport:udp network.destination.ip:\"*\"",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.0.limit = 0",
	"widget.0.timeseries_definition.0.request.5.display_type = area",
	"widget.0.timeseries_definition.0.request.4.network_query.0.index = netflow-search",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.sort.facet = @duration",
	"widget.0.timeseries_definition.0.request.4.q =",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.compute.% = 3",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.sort.aggregation = avg",
	"widget.0.timeseries_definition.0.request.5.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.facet = service",
	"widget.0.timeseries_definition.0.request.4.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.compute.interval = 10",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.compute.aggregation = avg",
	"widget.0.timeseries_definition.0.request.5.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.4.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.4.style.0.palette = orange",
	"widget.0.timeseries_definition.0.request.4.display_type = area",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.0.sort.% = 0",
	"widget.0.timeseries_definition.0.request.5.style.0.palette = grey",
	"widget.0.timeseries_definition.0.request.4.network_query.0.compute.aggregation = sum",
	"widget.0.timeseries_definition.0.request.5.q =",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.index = *",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.sort.order = desc",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.search.query =",
}

func TestAccDatadogDashboardTimeseries(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTimeseriesConfig, "datadog_dashboard.timeseries_dashboard", datadogDashboardTimeseriesAsserts)
}

func TestAccDatadogDashboardTimeseries_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardTimeseriesConfig, "datadog_dashboard.timeseries_dashboard")
}

const datadogDashboardTimeseriesMultiComputeConfig = `
resource "datadog_dashboard" "timeseries_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		timeseries_definition {
			title_size = "16"
			title_align = "left"
			show_legend = "true"
			title = "system.cpu.user, env, process.stat.cpu.total_pct.norm, network.bytes_read, @d..."
			legend_size = "2"
			yaxis {
				label = ""
				min = "0"
				include_zero = "true"
				max = "599999"
				scale = ""
			}
			right_yaxis {
				label = ""
				min = "1"
				include_zero = "false"
				max = "599998"
				scale = ""
			}
			marker {
				display_type = "error dashed"
				value = "y=500000"
				label = "y=500000"
			}
			marker {
				display_type = "warning dashed"
				value = "y=400000"
				label = "y=400000"
			}
			time = {
				live_span = "5m"
			}
			event {
				q = "sources:test tags:1"
				tags_execution = "and"
			}
			request {
				style {
					line_width = "normal"
					palette = "cool"
					line_type = "solid"
				}
				display_type = "line"
				log_query {
					index = "*"
					search = {
						query = ""
					}
					group_by {
						facet = "service"
						sort = {
							aggregation = "count"
							order = "desc"
						}
						limit = "10"
					}
					multi_compute {
						aggregation = "count"
					}
					multi_compute {
						facet = "env"
						interval = "1000"
						aggregation = "cardinality"
					}
				}
			}
		}
	}
}
`

var datadogDashboardTimeseriesMultiComputeAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.timeseries_definition.0.event.0.q = sources:test tags:1",
	"widget.0.timeseries_definition.0.event.0.tags_execution = and",
	"widget.0.timeseries_definition.0.legend_size = 2",
	"widget.0.timeseries_definition.0.marker.# = 2",
	"widget.0.timeseries_definition.0.marker.0.display_type = error dashed",
	"widget.0.timeseries_definition.0.marker.0.label = y=500000",
	"widget.0.timeseries_definition.0.marker.0.value = y=500000",
	"widget.0.timeseries_definition.0.marker.1.display_type = warning dashed",
	"widget.0.timeseries_definition.0.marker.1.label = y=400000",
	"widget.0.timeseries_definition.0.marker.1.value = y=400000",
	"widget.0.timeseries_definition.0.request.# = 1",
	"widget.0.timeseries_definition.0.request.0.display_type = line",
	"widget.0.timeseries_definition.0.request.0.log_query.# = 1",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.# = 2",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.0.aggregation = count",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.1.aggregation = cardinality",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.1.facet = env",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.1.interval = 1000",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.# = 1",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.0.facet = service",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.0.sort.aggregation = count",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.0.sort.order = desc",
	"widget.0.timeseries_definition.0.request.0.log_query.0.index = *",
	"widget.0.timeseries_definition.0.request.0.style.# = 1",
	"widget.0.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.0.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.0.style.0.palette = cool",
	"widget.0.timeseries_definition.0.show_legend = true",
	"widget.0.timeseries_definition.0.time.live_span = 5m",
	"widget.0.timeseries_definition.0.title = system.cpu.user, env, process.stat.cpu.total_pct.norm, network.bytes_read, @d...",
	"widget.0.timeseries_definition.0.title_align = left",
	"widget.0.timeseries_definition.0.title_size = 16",
	"widget.0.timeseries_definition.0.yaxis.# = 1",
	"widget.0.timeseries_definition.0.yaxis.0.include_zero = true",
	"widget.0.timeseries_definition.0.yaxis.0.max = 599999",
	"widget.0.timeseries_definition.0.yaxis.0.min = 0",
	"widget.0.timeseries_definition.0.right_yaxis.# = 1",
	"widget.0.timeseries_definition.0.right_yaxis.0.include_zero = false",
	"widget.0.timeseries_definition.0.right_yaxis.0.max = 599998",
	"widget.0.timeseries_definition.0.right_yaxis.0.min = 1",
}

func TestAccDatadogDashboardTimeseriesMultiCompute(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTimeseriesMultiComputeConfig, "datadog_dashboard.timeseries_dashboard", datadogDashboardTimeseriesMultiComputeAsserts)
}

func TestAccDatadogDashboardTimeseriesMultiCompute_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardTimeseriesMultiComputeConfig, "datadog_dashboard.timeseries_dashboard")
}
