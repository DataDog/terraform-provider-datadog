package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDatadogDashboardJson_Basic(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueID := uniqueAWSAccountID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardJsonTimeboard_Basic(uniqueID),
			},
			{
				ResourceName:      "datadog_dashboard_json.timeboard_json",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckDatadogDashboardJsonScreenboard_Basic(uniqueID),
			},
			{
				ResourceName:      "datadog_dashboard_json.screenboard_json",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogDashboardJsonTimeboard_Basic(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_json" "timeboard_json" {
   dashboard_json = <<EOF
{
   "title":"%s",
   "description":"Created using the Datadog provider in Terraform",
   "widgets":[
      {
         "id":873623537884427,
         "definition":{
            "title":"Widget Title",
            "type":"alert_graph",
            "alert_id":"895605",
            "viz_type":"timeseries"
         }
      },
      {
         "id":5600215046192430,
         "definition":{
            "title":"Widget Title",
            "type":"alert_value",
            "alert_id":"895605",
            "unit":"b",
            "text_align":"center",
            "precision":3
         }
      },
      {
         "id":5436370674582587,
         "definition":{
            "title":"Widget Title",
            "type":"alert_value",
            "alert_id":"895605",
            "unit":"b",
            "text_align":"center",
            "precision":3
         }
      },
      {
         "id":3887046970315839,
         "definition":{
            "title":"Widget Title",
            "type":"change",
            "requests":[
               {
                  "q":"avg:system.load.1{env:staging} by {account}",
                  "compare_to":"week_before",
                  "order_by":"name",
                  "order_dir":"desc",
                  "increase_good":true,
                  "change_type":"absolute",
                  "show_present":true
               }
            ]
         }
      },
      {
         "id":1219518175048191,
         "definition":{
            "title":"Widget Title",
            "show_legend":false,
            "type":"distribution",
            "requests":[
               {
                  "q":"avg:system.load.1{env:staging} by {account}",
                  "style":{
                     "palette":"warm"
                  }
               }
            ]
         }
      },
      {
         "id":6039041238503416,
         "definition":{
            "title":"Widget Title",
            "type":"check_status",
            "check":"aws.ecs.agent_connected",
            "grouping":"cluster",
            "group_by":[
               "account",
               "cluster"
            ],
            "tags":[
               "account:demo",
               "cluster:awseb-ruthebdog-env-8-dn3m6u3gvk"
            ]
         }
      },
      {
         "id":5186844025489598,
         "definition":{
            "title":"Widget Title",
            "show_legend":false,
            "type":"heatmap",
            "yaxis":{
               "scale":"sqrt",
               "include_zero":true,
               "min":"1",
               "max":"2"
            },
            "requests":[
               {
                  "q":"avg:system.load.1{env:staging} by {account}",
                  "style":{
                     "palette":"warm"
                  }
               }
            ]
         }
      },
      {
         "id":6742660811820435,
         "definition":{
            "title":"Widget Title",
            "type":"hostmap",
            "requests":{
               "fill":{
                  "q":"avg:system.load.1{*} by {host}"
               },
               "size":{
                  "q":"avg:memcache.uptime{*} by {host}"
               }
            },
            "node_type":"container",
            "no_metric_hosts":true,
            "no_group_hosts":true,
            "group":[
               "host",
               "region"
            ],
            "scope":[
               "region:us-east-1",
               "aws_account:727006795293"
            ],
            "style":{
               "palette":"yellow_to_green",
               "palette_flip":true,
               "fill_min":"10",
               "fill_max":"20"
            }
         }
      },
      {
         "id":1986924343921271,
         "definition":{
            "type":"note",
            "content":"note text",
            "background_color":"pink",
            "font_size":"14",
            "text_align":"center",
            "show_tick":true,
            "tick_pos":"50%%",
            "tick_edge":"left"
         }
      },
      {
         "id":3043237513486645,
         "definition":{
            "title":"Widget Title",
            "type":"query_value",
            "requests":[
               {
                  "q":"avg:system.load.1{env:staging} by {account}",
                  "aggregator":"sum",
                  "conditional_formats":[
                     {
                        "hide_value":false,
                        "comparator":"<",
                        "palette":"white_on_green",
                        "value":2
                     },
                     {
                        "hide_value":false,
                        "comparator":">",
                        "palette":"white_on_red",
                        "value":2.2
                     }
                  ]
               }
            ],
            "autoscale":true,
            "custom_unit":"xx",
            "text_align":"right",
            "precision":4
         }
      },
      {
         "id":8636154599297416,
         "definition":{
            "title":"Widget Title",
            "type":"query_table",
            "requests":[
               {
                  "q":"avg:system.load.1{env:staging} by {account}",
                  "aggregator":"sum",
                  "conditional_formats":[
                     {
                        "hide_value":false,
                        "comparator":"<",
                        "palette":"white_on_green",
                        "value":2
                     },
                     {
                        "hide_value":false,
                        "comparator":">",
                        "palette":"white_on_red",
                        "value":2.2
                     }
                  ],
                  "limit":10
               }
            ]
         }
      },
      {
         "id":518322985317720,
         "definition":{
            "title":"Widget Title",
            "type":"scatterplot",
            "requests":{
               "x":{
                  "q":"avg:system.cpu.user{*} by {service, account}",
                  "aggregator":"max"
               },
               "y":{
                  "q":"avg:system.mem.used{*} by {service, account}",
                  "aggregator":"min"
               }
            },
            "xaxis":{
               "scale":"pow",
               "label":"x",
               "include_zero":true,
               "min":"1",
               "max":"2000"
            },
            "yaxis":{
               "scale":"log",
               "label":"y",
               "include_zero":false,
               "min":"5",
               "max":"2222"
            },
            "color_by_groups":[
               "account",
               "apm-role-group"
            ]
         }
      },
      {
         "id":4913548056140044,
         "definition":{
            "title":"env: prod, datacenter:us1.prod.dog, service: master-db",
            "title_size":"16",
            "title_align":"left",
            "type":"servicemap",
            "service":"master-db",
            "filters":[
               "env:prod",
               "datacenter:us1.prod.dog"
            ]
         }
      },
      {
         "id":215209954480975,
         "definition":{
            "title":"Widget Title",
            "show_legend":true,
            "legend_size":"2",
            "type":"timeseries",
            "requests":[
               {
                  "q":"avg:system.cpu.user{app:general} by {env}",
                  "on_right_yaxis":false,
                  "metadata":[
                     {
                        "expression":"avg:system.cpu.user{app:general} by {env}",
                        "alias_name":"Alpha"
                     }
                  ],
                  "style":{
                     "palette":"warm",
                     "line_type":"dashed",
                     "line_width":"thin"
                  },
                  "display_type":"line"
               },
               {
                  "on_right_yaxis":false,
                  "log_query":{
                     "index":"mcnulty",
                     "search":{
                        "query":"status:info"
                     },
                     "group_by":[
                        {
                           "facet":"host",
                           "sort":{
                              "facet":"@duration",
                              "aggregation":"avg",
                              "order":"desc"
                           },
                           "limit":10
                        }
                     ],
                     "compute":{
                        "facet":"@duration",
                        "interval":5000,
                        "aggregation":"avg"
                     }
                  },
                  "display_type":"area"
               },
               {
                  "on_right_yaxis":false,
                  "apm_query":{
                     "index":"apm-search",
                     "search":{
                        "query":"type:web"
                     },
                     "group_by":[
                        {
                           "facet":"resource_name",
                           "sort":{
                              "facet":"@string_query.interval",
                              "aggregation":"avg",
                              "order":"desc"
                           },
                           "limit":50
                        }
                     ],
                     "compute":{
                        "facet":"@duration",
                        "interval":5000,
                        "aggregation":"avg"
                     }
                  },
                  "display_type":"bars"
               },
               {
                  "on_right_yaxis":false,
                  "process_query":{
                     "search_by":"error",
                     "metric":"process.stat.cpu.total_pct",
                     "limit":50,
                     "filter_by":[
                        "active"
                     ]
                  },
                  "display_type":"area"
               }
            ],
            "yaxis":{
               "scale":"log",
               "include_zero":false,
               "max":"100"
            },
            "events":[
               {
                  "q":"sources:test tags:1"
               },
               {
                  "q":"sources:test tags:2"
               }
            ],
            "markers":[
               {
                  "label":" z=6 ",
                  "value":"y = 4",
                  "display_type":"error dashed"
               },
               {
                  "label":" x=8 ",
                  "value":"10 < y < 999",
                  "display_type":"ok solid"
               }
            ]
         }
      },
      {
         "id":8114292022885770,
         "definition":{
            "title":"Widget Title",
            "type":"toplist",
            "requests":[
               {
                  "q":"avg:system.cpu.user{app:general} by {env}",
                  "conditional_formats":[
                     {
                        "hide_value":false,
                        "comparator":"<",
                        "palette":"white_on_green",
                        "value":2
                     },
                     {
                        "hide_value":false,
                        "comparator":">",
                        "palette":"white_on_red",
                        "value":2.2
                     }
                  ]
               }
            ]
         }
      },
      {
         "id":444605829496771,
         "definition":{
            "title":"Group Widget",
            "type":"group",
            "layout_type":"ordered",
            "widgets":[
               {
                  "id":5895282469348513,
                  "definition":{
                     "type":"note",
                     "content":"cluster note widget",
                     "background_color":"pink",
                     "font_size":"14",
                     "text_align":"center",
                     "show_tick":true,
                     "tick_pos":"50%%",
                     "tick_edge":"left"
                  }
               },
               {
                  "id":8096017487317681,
                  "definition":{
                     "title":"Alert Graph",
                     "type":"alert_graph",
                     "alert_id":"123",
                     "viz_type":"toplist"
                  }
               }
            ]
         }
      },
      {
         "id":7981844470437074,
         "definition":{
            "title":"Widget Title",
            "type":"slo",
            "view_type":"detail",
            "time_windows":[
               "7d",
               "previous_week"
            ],
            "slo_id":"56789",
            "show_error_budget":true,
            "view_mode":"overall",
            "global_time_target":"0"
         }
      }
   ],
   "template_variables":[
      {
         "name":"var_1",
         "default":"aws",
         "prefix":"host"
      },
      {
         "name":"var_2",
         "default":"autoscaling",
         "prefix":"service_name"
      }
   ],
   "layout_type":"ordered",
   "is_read_only":true,
   "notify_list":[
      
   ],
   "template_variable_presets":[
      {
         "name":"preset_1",
         "template_variables":[
            {
               "name":"var_1",
               "value":"host.dc"
            },
            {
               "name":"var_2",
               "value":"my_service"
            }
         ]
      }
   ],
   "id":"5uw-bbj-xec"
}
EOF
}`, uniq)
}

func testAccCheckDatadogDashboardJsonScreenboard_Basic(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_json" "screenboard_json" {
   dashboard_json = <<EOF
{
   "title":"%s",
   "description":"Created using the Datadog provider in Terraform",
   "widgets":[
      {
         "id":5574860246831982,
         "layout":{
            "x":5,
            "y":5,
            "width":32,
            "height":43
         },
         "definition":{
            "title":"Widget Title",
            "title_size":"16",
            "title_align":"left",
            "time":{
               "live_span":"1h"
            },
            "type":"event_stream",
            "query":"*",
            "event_size":"l"
         }
      },
      {
         "id":3310490736393290,
         "layout":{
            "x":42,
            "y":73,
            "width":65,
            "height":9
         },
         "definition":{
            "title":"Widget Title",
            "title_size":"16",
            "title_align":"left",
            "time":{
               "live_span":"1h"
            },
            "type":"event_timeline",
            "query":"*"
         }
      },
      {
         "id":1117617615518455,
         "layout":{
            "x":42,
            "y":5,
            "width":30,
            "height":20
         },
         "definition":{
            "type":"free_text",
            "text":"free text content",
            "color":"#d00",
            "font_size":"88",
            "text_align":"left"
         }
      },
      {
         "id":3098118775539428,
         "layout":{
            "x":111,
            "y":8,
            "width":39,
            "height":46
         },
         "definition":{
            "type":"iframe",
            "url":"http://google.com"
         }
      },
      {
         "id":651713243056399,
         "layout":{
            "x":77,
            "y":7,
            "width":30,
            "height":20
         },
         "definition":{
            "type":"image",
            "url":"https://images.pexels.com/photos/67636/rose-blue-flower-rose-blooms-67636.jpeg?auto=compress&cs=tinysrgb&h=350",
            "sizing":"fit",
            "margin":"small"
         }
      },
      {
         "id":5458329230004343,
         "layout":{
            "x":5,
            "y":51,
            "width":32,
            "height":36
         },
         "definition":{
            "type":"log_stream",
            "logset":"",
            "indexes":[
               "main"
            ],
            "query":"error",
            "sort":{
               "column":"time",
               "order":"desc"
            },
            "columns":[
               "core_host",
               "core_service",
               "tag_source"
            ],
            "show_date_column":true,
            "show_message_column":true,
            "message_display":"expanded-md"
         }
      },
      {
         "id":1112741664700765,
         "layout":{
            "x":112,
            "y":55,
            "width":30,
            "height":40
         },
         "definition":{
            "title":"Widget Title",
            "title_size":"16",
            "title_align":"left",
            "type":"manage_status",
            "summary_type":"monitors",
            "display_format":"countsAndList",
            "color_preference":"text",
            "hide_zero_counts":true,
            "show_last_triggered":false,
            "query":"type:metric",
            "sort":"status,asc",
            "count":50,
            "start":0
         }
      },
      {
         "id":6949442529647217,
         "layout":{
            "x":40,
            "y":28,
            "width":67,
            "height":38
         },
         "definition":{
            "title":"alerting-cassandra #env:datad0g.com",
            "title_size":"13",
            "title_align":"center",
            "time":{
               "live_span":"1h"
            },
            "type":"trace_service",
            "env":"datad0g.com",
            "service":"alerting-cassandra",
            "span_name":"cassandra.query",
            "show_hits":true,
            "show_errors":true,
            "show_latency":false,
            "show_breakdown":true,
            "show_distribution":true,
            "show_resource_list":false,
            "size_format":"large",
            "display_format":"three_column"
         }
      }
   ],
   "template_variables":[
      {
         "name":"var_1",
         "default":"aws",
         "prefix":"host"
      },
      {
         "name":"var_2",
         "default":"autoscaling",
         "prefix":"service_name"
      }
   ],
   "layout_type":"free",
   "is_read_only":false,
   "notify_list":[
      
   ],
   "template_variable_presets":[
      {
         "name":"preset_1",
         "template_variables":[
            {
               "name":"var_1",
               "value":"host.dc"
            },
            {
               "name":"var_2",
               "value":"my_service"
            }
         ]
      }
   ],
   "id":"hjf-2xf-xc8"
}
EOF
}`, uniq)
}
