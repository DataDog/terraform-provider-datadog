package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogDashboardJSONBasicTimeboard(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	uniqUpdated := fmt.Sprintf("%s-updated", uniq)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		// Import checkDashboardDestroy() from Dashboard resource
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardJSONTimeboardJSON(uniq),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_json", "dashboard", fmt.Sprintf("{\"description\":\"Created using the Datadog provider in Terraform\",\"is_read_only\":true,\"layout_type\":\"ordered\",\"notify_list\":[],\"template_variable_presets\":[{\"name\":\"preset_1\",\"template_variables\":[{\"name\":\"var_1\",\"value\":\"host.dc\"},{\"name\":\"var_2\",\"value\":\"my_service\"}]}],\"template_variables\":[{\"default\":\"aws\",\"name\":\"var_1\",\"prefix\":\"host\"},{\"default\":\"autoscaling\",\"name\":\"var_2\",\"prefix\":\"service_name\"}],\"title\":\"%s\",\"widgets\":[{\"definition\":{\"alert_id\":\"895605\",\"title\":\"Widget Title\",\"type\":\"alert_graph\",\"viz_type\":\"timeseries\"}},{\"definition\":{\"alert_id\":\"895605\",\"precision\":3,\"text_align\":\"center\",\"title\":\"Widget Title\",\"type\":\"alert_value\",\"unit\":\"b\"}},{\"definition\":{\"alert_id\":\"895605\",\"precision\":3,\"text_align\":\"center\",\"title\":\"Widget Title\",\"type\":\"alert_value\",\"unit\":\"b\"}},{\"definition\":{\"requests\":[{\"change_type\":\"absolute\",\"compare_to\":\"week_before\",\"increase_good\":true,\"order_by\":\"name\",\"order_dir\":\"desc\",\"q\":\"avg:system.load.1{env:staging} by {account}\",\"show_present\":true}],\"title\":\"Widget Title\",\"type\":\"change\"}},{\"definition\":{\"requests\":[{\"q\":\"avg:system.load.1{env:staging} by {account}\",\"style\":{\"palette\":\"warm\"}}],\"show_legend\":false,\"title\":\"Widget Title\",\"type\":\"distribution\"}},{\"definition\":{\"check\":\"aws.ecs.agent_connected\",\"group_by\":[\"account\",\"cluster\"],\"grouping\":\"cluster\",\"tags\":[\"account:demo\",\"cluster:awseb-ruthebdog-env-8-dn3m6u3gvk\"],\"title\":\"Widget Title\",\"type\":\"check_status\"}},{\"definition\":{\"requests\":[{\"q\":\"avg:system.load.1{env:staging} by {account}\",\"style\":{\"palette\":\"warm\"}}],\"show_legend\":false,\"title\":\"Widget Title\",\"type\":\"heatmap\",\"yaxis\":{\"include_zero\":true,\"max\":\"2\",\"min\":\"1\",\"scale\":\"sqrt\"}}},{\"definition\":{\"group\":[\"host\",\"region\"],\"no_group_hosts\":true,\"no_metric_hosts\":true,\"node_type\":\"container\",\"requests\":{\"fill\":{\"q\":\"avg:system.load.1{*} by {host}\"},\"size\":{\"q\":\"avg:memcache.uptime{*} by {host}\"}},\"scope\":[\"region:us-east-1\",\"aws_account:727006795293\"],\"style\":{\"fill_max\":\"20\",\"fill_min\":\"10\",\"palette\":\"yellow_to_green\",\"palette_flip\":true},\"title\":\"Widget Title\",\"type\":\"hostmap\"}},{\"definition\":{\"background_color\":\"pink\",\"content\":\"note text\",\"font_size\":\"14\",\"show_tick\":true,\"text_align\":\"center\",\"tick_edge\":\"left\",\"tick_pos\":\"50%%\",\"type\":\"note\"}},{\"definition\":{\"autoscale\":true,\"custom_unit\":\"xx\",\"precision\":4,\"requests\":[{\"aggregator\":\"sum\",\"conditional_formats\":[{\"comparator\":\"\\u003c\",\"hide_value\":false,\"palette\":\"white_on_green\",\"value\":2},{\"comparator\":\"\\u003e\",\"hide_value\":false,\"palette\":\"white_on_red\",\"value\":2.2}],\"q\":\"avg:system.load.1{env:staging} by {account}\"}],\"text_align\":\"right\",\"title\":\"Widget Title\",\"type\":\"query_value\"}},{\"definition\":{\"requests\":[{\"aggregator\":\"sum\",\"conditional_formats\":[{\"comparator\":\"\\u003c\",\"hide_value\":false,\"palette\":\"white_on_green\",\"value\":2},{\"comparator\":\"\\u003e\",\"hide_value\":false,\"palette\":\"white_on_red\",\"value\":2.2}],\"limit\":10,\"q\":\"avg:system.load.1{env:staging} by {account}\"}],\"title\":\"Widget Title\",\"type\":\"query_table\"}},{\"definition\":{\"color_by_groups\":[\"account\",\"apm-role-group\"],\"requests\":{\"x\":{\"aggregator\":\"max\",\"q\":\"avg:system.cpu.user{*} by {service, account}\"},\"y\":{\"aggregator\":\"min\",\"q\":\"avg:system.mem.used{*} by {service, account}\"}},\"title\":\"Widget Title\",\"type\":\"scatterplot\",\"xaxis\":{\"include_zero\":true,\"label\":\"x\",\"max\":\"2000\",\"min\":\"1\",\"scale\":\"pow\"},\"yaxis\":{\"include_zero\":false,\"label\":\"y\",\"max\":\"2222\",\"min\":\"5\",\"scale\":\"log\"}}},{\"definition\":{\"filters\":[\"env:prod\",\"datacenter:dc1\"],\"service\":\"master-db\",\"title\":\"env: prod, datacenter:dc1, service: master-db\",\"title_align\":\"left\",\"title_size\":\"16\",\"type\":\"servicemap\"}},{\"definition\":{\"events\":[{\"q\":\"sources:test tags:1\"},{\"q\":\"sources:test tags:2\"}],\"legend_size\":\"2\",\"markers\":[{\"display_type\":\"error dashed\",\"label\":\" z=6 \",\"value\":\"y = 4\"},{\"display_type\":\"ok solid\",\"label\":\" x=8 \",\"value\":\"10 \\u003c y \\u003c 999\"}],\"requests\":[{\"display_type\":\"line\",\"metadata\":[{\"alias_name\":\"Alpha\",\"expression\":\"avg:system.cpu.user{app:general} by {env}\"}],\"on_right_yaxis\":false,\"q\":\"avg:system.cpu.user{app:general} by {env}\",\"style\":{\"line_type\":\"dashed\",\"line_width\":\"thin\",\"palette\":\"warm\"}},{\"display_type\":\"area\",\"log_query\":{\"compute\":{\"aggregation\":\"avg\",\"facet\":\"@duration\",\"interval\":5000},\"group_by\":[{\"facet\":\"host\",\"limit\":10,\"sort\":{\"aggregation\":\"avg\",\"facet\":\"@duration\",\"order\":\"desc\"}}],\"index\":\"mcnulty\",\"search\":{\"query\":\"status:info\"}},\"on_right_yaxis\":false},{\"apm_query\":{\"compute\":{\"aggregation\":\"avg\",\"facet\":\"@duration\",\"interval\":5000},\"group_by\":[{\"facet\":\"resource_name\",\"limit\":50,\"sort\":{\"aggregation\":\"avg\",\"facet\":\"@string_query.interval\",\"order\":\"desc\"}}],\"index\":\"apm-search\",\"search\":{\"query\":\"type:web\"}},\"display_type\":\"bars\",\"on_right_yaxis\":false},{\"display_type\":\"area\",\"on_right_yaxis\":false,\"process_query\":{\"filter_by\":[\"active\"],\"limit\":50,\"metric\":\"process.stat.cpu.total_pct\",\"search_by\":\"error\"}}],\"show_legend\":true,\"title\":\"Widget Title\",\"type\":\"timeseries\",\"yaxis\":{\"include_zero\":false,\"max\":\"100\",\"scale\":\"log\"}}},{\"definition\":{\"requests\":[{\"conditional_formats\":[{\"comparator\":\"\\u003c\",\"hide_value\":false,\"palette\":\"white_on_green\",\"value\":2},{\"comparator\":\"\\u003e\",\"hide_value\":false,\"palette\":\"white_on_red\",\"value\":2.2}],\"q\":\"avg:system.cpu.user{app:general} by {env}\"}],\"title\":\"Widget Title\",\"type\":\"toplist\"}},{\"definition\":{\"layout_type\":\"ordered\",\"title\":\"Group Widget\",\"type\":\"group\",\"widgets\":[{\"definition\":{\"background_color\":\"pink\",\"content\":\"cluster note widget\",\"font_size\":\"14\",\"show_tick\":true,\"text_align\":\"center\",\"tick_edge\":\"left\",\"tick_pos\":\"50%%\",\"type\":\"note\"}},{\"definition\":{\"alert_id\":\"123\",\"title\":\"Alert Graph\",\"type\":\"alert_graph\",\"viz_type\":\"toplist\"}}]}},{\"definition\":{\"global_time_target\":\"0\",\"show_error_budget\":true,\"slo_id\":\"56789\",\"time_windows\":[\"7d\",\"previous_week\"],\"title\":\"Widget Title\",\"type\":\"slo\",\"view_mode\":\"overall\",\"view_type\":\"detail\"}}]}", uniq)),
				),
			},
			{
				Config: testAccCheckDatadogDashboardJSONTimeboardJSONUpdated(uniqUpdated),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_json", "dashboard", fmt.Sprintf("{\"description\":\"Created using the Datadog provider in Terraform\",\"is_read_only\":true,\"layout_type\":\"ordered\",\"notify_list\":[],\"template_variable_presets\":[{\"name\":\"preset_1\",\"template_variables\":[{\"name\":\"var_1\",\"value\":\"host.dc\"},{\"name\":\"var_2\",\"value\":\"my_service\"}]}],\"template_variables\":[{\"default\":\"aws\",\"name\":\"var_1\",\"prefix\":\"host\"},{\"default\":\"autoscaling\",\"name\":\"var_2\",\"prefix\":\"service_name\"}],\"title\":\"%s\",\"widgets\":[{\"definition\":{\"alert_id\":\"895605\",\"title\":\"Widget Title\",\"type\":\"alert_graph\",\"viz_type\":\"timeseries\"}},{\"definition\":{\"alert_id\":\"895605\",\"precision\":3,\"text_align\":\"center\",\"title\":\"Widget Title\",\"type\":\"alert_value\",\"unit\":\"b\"}},{\"definition\":{\"alert_id\":\"895605\",\"precision\":3,\"text_align\":\"center\",\"title\":\"Widget Title\",\"type\":\"alert_value\",\"unit\":\"b\"}},{\"definition\":{\"requests\":[{\"change_type\":\"absolute\",\"compare_to\":\"week_before\",\"increase_good\":true,\"order_by\":\"name\",\"order_dir\":\"desc\",\"q\":\"avg:system.load.1{env:staging} by {account}\",\"show_present\":true}],\"title\":\"Widget Title\",\"type\":\"change\"}},{\"definition\":{\"requests\":[{\"q\":\"avg:system.load.1{env:staging} by {account}\",\"style\":{\"palette\":\"warm\"}}],\"show_legend\":false,\"title\":\"Widget Title\",\"type\":\"distribution\"}},{\"definition\":{\"check\":\"aws.ecs.agent_connected\",\"group_by\":[\"account\",\"cluster\"],\"grouping\":\"cluster\",\"tags\":[\"account:demo\",\"cluster:awseb-ruthebdog-env-8-dn3m6u3gvk\"],\"title\":\"Widget Title\",\"type\":\"check_status\"}},{\"definition\":{\"requests\":[{\"q\":\"avg:system.load.1{env:staging} by {account}\",\"style\":{\"palette\":\"warm\"}}],\"show_legend\":false,\"title\":\"Widget Title\",\"type\":\"heatmap\",\"yaxis\":{\"include_zero\":true,\"max\":\"2\",\"min\":\"1\",\"scale\":\"sqrt\"}}},{\"definition\":{\"group\":[\"host\",\"region\"],\"no_group_hosts\":true,\"no_metric_hosts\":true,\"node_type\":\"container\",\"requests\":{\"fill\":{\"q\":\"avg:system.load.1{*} by {host}\"},\"size\":{\"q\":\"avg:memcache.uptime{*} by {host}\"}},\"scope\":[\"region:us-east-1\",\"aws_account:727006795293\"],\"style\":{\"fill_max\":\"20\",\"fill_min\":\"10\",\"palette\":\"yellow_to_green\",\"palette_flip\":true},\"title\":\"Widget Title\",\"type\":\"hostmap\"}},{\"definition\":{\"background_color\":\"pink\",\"content\":\"note text\",\"font_size\":\"14\",\"show_tick\":true,\"text_align\":\"center\",\"tick_edge\":\"left\",\"tick_pos\":\"50%%\",\"type\":\"note\"}},{\"definition\":{\"autoscale\":true,\"custom_unit\":\"xx\",\"precision\":4,\"requests\":[{\"aggregator\":\"sum\",\"conditional_formats\":[{\"comparator\":\"\\u003c\",\"hide_value\":false,\"palette\":\"white_on_green\",\"value\":2},{\"comparator\":\"\\u003e\",\"hide_value\":false,\"palette\":\"white_on_red\",\"value\":2.2}],\"q\":\"avg:system.load.1{env:staging} by {account}\"}],\"text_align\":\"right\",\"title\":\"Widget Title\",\"type\":\"query_value\"}},{\"definition\":{\"requests\":[{\"aggregator\":\"sum\",\"conditional_formats\":[{\"comparator\":\"\\u003c\",\"hide_value\":false,\"palette\":\"white_on_green\",\"value\":2},{\"comparator\":\"\\u003e\",\"hide_value\":false,\"palette\":\"white_on_red\",\"value\":2.2}],\"limit\":10,\"q\":\"avg:system.load.1{env:staging} by {account}\"}],\"title\":\"Widget Title\",\"type\":\"query_table\"}},{\"definition\":{\"color_by_groups\":[\"account\",\"apm-role-group\"],\"requests\":{\"x\":{\"aggregator\":\"max\",\"q\":\"avg:system.cpu.user{*} by {service, account}\"},\"y\":{\"aggregator\":\"min\",\"q\":\"avg:system.mem.used{*} by {service, account}\"}},\"title\":\"Widget Title\",\"type\":\"scatterplot\",\"xaxis\":{\"include_zero\":true,\"label\":\"x\",\"max\":\"2000\",\"min\":\"1\",\"scale\":\"pow\"},\"yaxis\":{\"include_zero\":false,\"label\":\"y\",\"max\":\"2222\",\"min\":\"5\",\"scale\":\"log\"}}},{\"definition\":{\"filters\":[\"env:prod\",\"datacenter:dc1\"],\"service\":\"master-db\",\"title\":\"env: prod, datacenter:dc1, service: master-db\",\"title_align\":\"left\",\"title_size\":\"16\",\"type\":\"servicemap\"}},{\"definition\":{\"events\":[{\"q\":\"sources:test tags:1\"},{\"q\":\"sources:test tags:2\"}],\"legend_size\":\"2\",\"markers\":[{\"display_type\":\"error dashed\",\"label\":\" z=6 \",\"value\":\"y = 4\"},{\"display_type\":\"ok solid\",\"label\":\" x=8 \",\"value\":\"10 \\u003c y \\u003c 999\"}],\"requests\":[{\"display_type\":\"line\",\"metadata\":[{\"alias_name\":\"Alpha\",\"expression\":\"avg:system.cpu.user{app:general} by {env}\"}],\"on_right_yaxis\":false,\"q\":\"avg:system.cpu.user{app:general} by {env}\",\"style\":{\"line_type\":\"dashed\",\"line_width\":\"thin\",\"palette\":\"warm\"}},{\"display_type\":\"area\",\"log_query\":{\"compute\":{\"aggregation\":\"avg\",\"facet\":\"@duration\",\"interval\":5000},\"group_by\":[{\"facet\":\"host\",\"limit\":10,\"sort\":{\"aggregation\":\"avg\",\"facet\":\"@duration\",\"order\":\"desc\"}}],\"index\":\"mcnulty\",\"search\":{\"query\":\"status:info\"}},\"on_right_yaxis\":false},{\"apm_query\":{\"compute\":{\"aggregation\":\"avg\",\"facet\":\"@duration\",\"interval\":5000},\"group_by\":[{\"facet\":\"resource_name\",\"limit\":50,\"sort\":{\"aggregation\":\"avg\",\"facet\":\"@string_query.interval\",\"order\":\"desc\"}}],\"index\":\"apm-search\",\"search\":{\"query\":\"type:web\"}},\"display_type\":\"bars\",\"on_right_yaxis\":false},{\"display_type\":\"area\",\"on_right_yaxis\":false,\"process_query\":{\"filter_by\":[\"active\"],\"limit\":50,\"metric\":\"process.stat.cpu.total_pct\",\"search_by\":\"error\"}}],\"show_legend\":true,\"title\":\"Widget Title\",\"type\":\"timeseries\",\"yaxis\":{\"include_zero\":false,\"max\":\"100\",\"scale\":\"log\"}}},{\"definition\":{\"requests\":[{\"conditional_formats\":[{\"comparator\":\"\\u003c\",\"hide_value\":false,\"palette\":\"white_on_green\",\"value\":2},{\"comparator\":\"\\u003e\",\"hide_value\":false,\"palette\":\"white_on_red\",\"value\":2.2}],\"q\":\"avg:system.cpu.user{app:general} by {env}\"}],\"title\":\"Widget Title\",\"type\":\"toplist\"}},{\"definition\":{\"layout_type\":\"ordered\",\"title\":\"Group Widget\",\"type\":\"group\",\"widgets\":[{\"definition\":{\"background_color\":\"pink\",\"content\":\"cluster note widget\",\"font_size\":\"14\",\"show_tick\":true,\"text_align\":\"center\",\"tick_edge\":\"left\",\"tick_pos\":\"50%%\",\"type\":\"note\"}},{\"definition\":{\"alert_id\":\"123\",\"title\":\"Alert Graph\",\"type\":\"alert_graph\",\"viz_type\":\"toplist\"}}]}},{\"definition\":{\"global_time_target\":\"0\",\"show_error_budget\":true,\"slo_id\":\"56789\",\"time_windows\":[\"7d\",\"previous_week\"],\"title\":\"Widget Title\",\"type\":\"slo\",\"view_mode\":\"overall\",\"view_type\":\"detail\"}}]}", uniqUpdated)),
				),
			},
			{
				Config: testAccCheckDatadogDashboardJSONTimeboardYAML(uniq),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_yaml", "dashboard", fmt.Sprintf("{\"description\":\"Created using the Datadog provider in Terraform\",\"is_read_only\":true,\"layout_type\":\"ordered\",\"notify_list\":[],\"template_variable_presets\":[{\"name\":\"preset_1\",\"template_variables\":[{\"name\":\"var_1\",\"value\":\"host.dc\"},{\"name\":\"var_2\",\"value\":\"my_service\"}]}],\"template_variables\":[{\"default\":\"aws\",\"name\":\"var_1\",\"prefix\":\"host\"},{\"default\":\"autoscaling\",\"name\":\"var_2\",\"prefix\":\"service_name\"}],\"title\":\"%s\",\"widgets\":[{\"definition\":{\"alert_id\":\"895605\",\"title\":\"Widget Title\",\"type\":\"alert_graph\",\"viz_type\":\"timeseries\"}},{\"definition\":{\"alert_id\":\"895605\",\"precision\":3,\"text_align\":\"center\",\"title\":\"Widget Title\",\"type\":\"alert_value\",\"unit\":\"b\"}},{\"definition\":{\"alert_id\":\"895605\",\"precision\":3,\"text_align\":\"center\",\"title\":\"Widget Title\",\"type\":\"alert_value\",\"unit\":\"b\"}},{\"definition\":{\"requests\":[{\"change_type\":\"absolute\",\"compare_to\":\"week_before\",\"increase_good\":true,\"order_by\":\"name\",\"order_dir\":\"desc\",\"q\":\"avg:system.load.1{env:staging} by {account}\",\"show_present\":true}],\"title\":\"Widget Title\",\"type\":\"change\"}},{\"definition\":{\"requests\":[{\"q\":\"avg:system.load.1{env:staging} by {account}\",\"style\":{\"palette\":\"warm\"}}],\"show_legend\":false,\"title\":\"Widget Title\",\"type\":\"distribution\"}},{\"definition\":{\"check\":\"aws.ecs.agent_connected\",\"group_by\":[\"account\",\"cluster\"],\"grouping\":\"cluster\",\"tags\":[\"account:demo\",\"cluster:awseb-ruthebdog-env-8-dn3m6u3gvk\"],\"title\":\"Widget Title\",\"type\":\"check_status\"}},{\"definition\":{\"requests\":[{\"q\":\"avg:system.load.1{env:staging} by {account}\",\"style\":{\"palette\":\"warm\"}}],\"show_legend\":false,\"title\":\"Widget Title\",\"type\":\"heatmap\",\"yaxis\":{\"include_zero\":true,\"max\":\"2\",\"min\":\"1\",\"scale\":\"sqrt\"}}},{\"definition\":{\"group\":[\"host\",\"region\"],\"no_group_hosts\":true,\"no_metric_hosts\":true,\"node_type\":\"container\",\"requests\":{\"fill\":{\"q\":\"avg:system.load.1{*} by {host}\"},\"size\":{\"q\":\"avg:memcache.uptime{*} by {host}\"}},\"scope\":[\"region:us-east-1\",\"aws_account:727006795293\"],\"style\":{\"fill_max\":\"20\",\"fill_min\":\"10\",\"palette\":\"yellow_to_green\",\"palette_flip\":true},\"title\":\"Widget Title\",\"type\":\"hostmap\"}},{\"definition\":{\"background_color\":\"pink\",\"content\":\"note text\",\"font_size\":\"14\",\"show_tick\":true,\"text_align\":\"center\",\"tick_edge\":\"left\",\"tick_pos\":\"50%%\",\"type\":\"note\"}},{\"definition\":{\"autoscale\":true,\"custom_unit\":\"xx\",\"precision\":4,\"requests\":[{\"aggregator\":\"sum\",\"conditional_formats\":[{\"comparator\":\"\\u003c\",\"hide_value\":false,\"palette\":\"white_on_green\",\"value\":2},{\"comparator\":\"\\u003e\",\"hide_value\":false,\"palette\":\"white_on_red\",\"value\":2.2}],\"q\":\"avg:system.load.1{env:staging} by {account}\"}],\"text_align\":\"right\",\"title\":\"Widget Title\",\"type\":\"query_value\"}},{\"definition\":{\"requests\":[{\"aggregator\":\"sum\",\"conditional_formats\":[{\"comparator\":\"\\u003c\",\"hide_value\":false,\"palette\":\"white_on_green\",\"value\":2},{\"comparator\":\"\\u003e\",\"hide_value\":false,\"palette\":\"white_on_red\",\"value\":2.2}],\"limit\":10,\"q\":\"avg:system.load.1{env:staging} by {account}\"}],\"title\":\"Widget Title\",\"type\":\"query_table\"}},{\"definition\":{\"color_by_groups\":[\"account\",\"apm-role-group\"],\"requests\":{\"x\":{\"aggregator\":\"max\",\"q\":\"avg:system.cpu.user{*} by {service, account}\"},\"y\":{\"aggregator\":\"min\",\"q\":\"avg:system.mem.used{*} by {service, account}\"}},\"title\":\"Widget Title\",\"type\":\"scatterplot\",\"xaxis\":{\"include_zero\":true,\"label\":\"x\",\"max\":\"2000\",\"min\":\"1\",\"scale\":\"pow\"},\"yaxis\":{\"include_zero\":false,\"label\":\"y\",\"max\":\"2222\",\"min\":\"5\",\"scale\":\"log\"}}},{\"definition\":{\"filters\":[\"env:prod\",\"datacenter:dc1\"],\"service\":\"master-db\",\"title\":\"env: prod, datacenter:dc1, service: master-db\",\"title_align\":\"left\",\"title_size\":\"16\",\"type\":\"servicemap\"}},{\"definition\":{\"events\":[{\"q\":\"sources:test tags:1\"},{\"q\":\"sources:test tags:2\"}],\"legend_size\":\"2\",\"markers\":[{\"display_type\":\"error dashed\",\"label\":\" z=6 \",\"value\":\"y = 4\"},{\"display_type\":\"ok solid\",\"label\":\" x=8 \",\"value\":\"10 \\u003c y \\u003c 999\"}],\"requests\":[{\"display_type\":\"line\",\"metadata\":[{\"alias_name\":\"Alpha\",\"expression\":\"avg:system.cpu.user{app:general} by {env}\"}],\"on_right_yaxis\":false,\"q\":\"avg:system.cpu.user{app:general} by {env}\",\"style\":{\"line_type\":\"dashed\",\"line_width\":\"thin\",\"palette\":\"warm\"}},{\"display_type\":\"area\",\"log_query\":{\"compute\":{\"aggregation\":\"avg\",\"facet\":\"@duration\",\"interval\":5000},\"group_by\":[{\"facet\":\"host\",\"limit\":10,\"sort\":{\"aggregation\":\"avg\",\"facet\":\"@duration\",\"order\":\"desc\"}}],\"index\":\"mcnulty\",\"search\":{\"query\":\"status:info\"}},\"on_right_yaxis\":false},{\"apm_query\":{\"compute\":{\"aggregation\":\"avg\",\"facet\":\"@duration\",\"interval\":5000},\"group_by\":[{\"facet\":\"resource_name\",\"limit\":50,\"sort\":{\"aggregation\":\"avg\",\"facet\":\"@string_query.interval\",\"order\":\"desc\"}}],\"index\":\"apm-search\",\"search\":{\"query\":\"type:web\"}},\"display_type\":\"bars\",\"on_right_yaxis\":false},{\"display_type\":\"area\",\"on_right_yaxis\":false,\"process_query\":{\"filter_by\":[\"active\"],\"limit\":50,\"metric\":\"process.stat.cpu.total_pct\",\"search_by\":\"error\"}}],\"show_legend\":true,\"title\":\"Widget Title\",\"type\":\"timeseries\",\"yaxis\":{\"include_zero\":false,\"max\":\"100\",\"scale\":\"log\"}}},{\"definition\":{\"requests\":[{\"conditional_formats\":[{\"comparator\":\"\\u003c\",\"hide_value\":false,\"palette\":\"white_on_green\",\"value\":2},{\"comparator\":\"\\u003e\",\"hide_value\":false,\"palette\":\"white_on_red\",\"value\":2.2}],\"q\":\"avg:system.cpu.user{app:general} by {env}\"}],\"title\":\"Widget Title\",\"type\":\"toplist\"}},{\"definition\":{\"layout_type\":\"ordered\",\"title\":\"Group Widget\",\"type\":\"group\",\"widgets\":[{\"definition\":{\"background_color\":\"pink\",\"content\":\"cluster note widget\",\"font_size\":\"14\",\"show_tick\":true,\"text_align\":\"center\",\"tick_edge\":\"left\",\"tick_pos\":\"50%%\",\"type\":\"note\"}},{\"definition\":{\"alert_id\":\"123\",\"title\":\"Alert Graph\",\"type\":\"alert_graph\",\"viz_type\":\"toplist\"}}]}},{\"definition\":{\"global_time_target\":\"0\",\"show_error_budget\":true,\"slo_id\":\"56789\",\"time_windows\":[\"7d\",\"previous_week\"],\"title\":\"Widget Title\",\"type\":\"slo\",\"view_mode\":\"overall\",\"view_type\":\"detail\"}}]}", uniq)),
				),
			},
		},
	})
}

func TestAccDatadogDashboardJSONBasicScreenboard(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		// Import checkDashboardDestroy() from Dashboard resource
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardJSONScreenboardJSON(uniq),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.screenboard_json", "dashboard", fmt.Sprintf("{\"description\":\"Created using the Datadog provider in Terraform\",\"is_read_only\":false,\"layout_type\":\"free\",\"notify_list\":[],\"template_variable_presets\":[{\"name\":\"preset_1\",\"template_variables\":[{\"name\":\"var_1\",\"value\":\"host.dc\"},{\"name\":\"var_2\",\"value\":\"my_service\"}]}],\"template_variables\":[{\"default\":\"aws\",\"name\":\"var_1\",\"prefix\":\"host\"},{\"default\":\"autoscaling\",\"name\":\"var_2\",\"prefix\":\"service_name\"}],\"title\":\"%s\",\"widgets\":[{\"definition\":{\"event_size\":\"l\",\"query\":\"*\",\"time\":{\"live_span\":\"1h\"},\"title\":\"Widget Title\",\"title_align\":\"left\",\"title_size\":\"16\",\"type\":\"event_stream\"},\"layout\":{\"height\":43,\"width\":32,\"x\":5,\"y\":5}},{\"definition\":{\"query\":\"*\",\"time\":{\"live_span\":\"1h\"},\"title\":\"Widget Title\",\"title_align\":\"left\",\"title_size\":\"16\",\"type\":\"event_timeline\"},\"layout\":{\"height\":9,\"width\":65,\"x\":42,\"y\":73}},{\"definition\":{\"color\":\"#d00\",\"font_size\":\"88\",\"text\":\"free text content\",\"text_align\":\"left\",\"type\":\"free_text\"},\"layout\":{\"height\":20,\"width\":30,\"x\":42,\"y\":5}},{\"definition\":{\"type\":\"iframe\",\"url\":\"http://google.com\"},\"layout\":{\"height\":46,\"width\":39,\"x\":111,\"y\":8}},{\"definition\":{\"margin\":\"small\",\"sizing\":\"fit\",\"type\":\"image\",\"url\":\"https://images.pexels.com/photos/67636/rose-blue-flower-rose-blooms-67636.jpeg?auto=compress\\u0026cs=tinysrgb\\u0026h=350\"},\"layout\":{\"height\":20,\"width\":30,\"x\":77,\"y\":7}},{\"definition\":{\"columns\":[\"core_host\",\"core_service\",\"tag_source\"],\"indexes\":[\"main\"],\"logset\":\"\",\"message_display\":\"expanded-md\",\"query\":\"error\",\"show_date_column\":true,\"show_message_column\":true,\"sort\":{\"column\":\"time\",\"order\":\"desc\"},\"type\":\"log_stream\"},\"layout\":{\"height\":36,\"width\":32,\"x\":5,\"y\":51}},{\"definition\":{\"color_preference\":\"text\",\"count\":50,\"display_format\":\"countsAndList\",\"hide_zero_counts\":true,\"query\":\"type:metric\",\"show_last_triggered\":false,\"sort\":\"status,asc\",\"start\":0,\"summary_type\":\"monitors\",\"title\":\"Widget Title\",\"title_align\":\"left\",\"title_size\":\"16\",\"type\":\"manage_status\"},\"layout\":{\"height\":40,\"width\":30,\"x\":112,\"y\":55}},{\"definition\":{\"display_format\":\"three_column\",\"env\":\"datadog.com\",\"service\":\"alerting-cassandra\",\"show_breakdown\":true,\"show_distribution\":true,\"show_errors\":true,\"show_hits\":true,\"show_latency\":false,\"show_resource_list\":false,\"size_format\":\"large\",\"span_name\":\"cassandra.query\",\"time\":{\"live_span\":\"1h\"},\"title\":\"alerting-cassandra #env:datadog.com\",\"title_align\":\"center\",\"title_size\":\"13\",\"type\":\"trace_service\"},\"layout\":{\"height\":38,\"width\":67,\"x\":40,\"y\":28}}]}", uniq)),
				),
			},
			{
				Config: testAccCheckDatadogDashboardJSONScreenboardYAML(uniq),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.screenboard_yaml", "dashboard", fmt.Sprintf("{\"description\":\"Created using the Datadog provider in Terraform\",\"is_read_only\":false,\"layout_type\":\"free\",\"notify_list\":[],\"template_variable_presets\":[{\"name\":\"preset_1\",\"template_variables\":[{\"name\":\"var_1\",\"value\":\"host.dc\"},{\"name\":\"var_2\",\"value\":\"my_service\"}]}],\"template_variables\":[{\"default\":\"aws\",\"name\":\"var_1\",\"prefix\":\"host\"},{\"default\":\"autoscaling\",\"name\":\"var_2\",\"prefix\":\"service_name\"}],\"title\":\"%s\",\"widgets\":[{\"definition\":{\"event_size\":\"l\",\"query\":\"*\",\"time\":{\"live_span\":\"1h\"},\"title\":\"Widget Title\",\"title_align\":\"left\",\"title_size\":\"16\",\"type\":\"event_stream\"},\"layout\":{\"height\":43,\"width\":32,\"x\":5,\"y\":5}},{\"definition\":{\"query\":\"*\",\"time\":{\"live_span\":\"1h\"},\"title\":\"Widget Title\",\"title_align\":\"left\",\"title_size\":\"16\",\"type\":\"event_timeline\"},\"layout\":{\"height\":9,\"width\":65,\"x\":42,\"y\":73}},{\"definition\":{\"color\":\"#d00\",\"font_size\":\"88\",\"text\":\"free text content\",\"text_align\":\"left\",\"type\":\"free_text\"},\"layout\":{\"height\":20,\"width\":30,\"x\":42,\"y\":5}},{\"definition\":{\"type\":\"iframe\",\"url\":\"http://google.com\"},\"layout\":{\"height\":46,\"width\":39,\"x\":111,\"y\":8}},{\"definition\":{\"margin\":\"small\",\"sizing\":\"fit\",\"type\":\"image\",\"url\":\"https://images.pexels.com/photos/67636/rose-blue-flower-rose-blooms-67636.jpeg?auto=compress\\u0026cs=tinysrgb\\u0026h=350\"},\"layout\":{\"height\":20,\"width\":30,\"x\":77,\"y\":7}},{\"definition\":{\"columns\":[\"core_host\",\"core_service\",\"tag_source\"],\"indexes\":[\"main\"],\"logset\":\"\",\"message_display\":\"expanded-md\",\"query\":\"error\",\"show_date_column\":true,\"show_message_column\":true,\"sort\":{\"column\":\"time\",\"order\":\"desc\"},\"type\":\"log_stream\"},\"layout\":{\"height\":36,\"width\":32,\"x\":5,\"y\":51}},{\"definition\":{\"color_preference\":\"text\",\"count\":50,\"display_format\":\"countsAndList\",\"hide_zero_counts\":true,\"query\":\"type:metric\",\"show_last_triggered\":false,\"sort\":\"status,asc\",\"start\":0,\"summary_type\":\"monitors\",\"title\":\"Widget Title\",\"title_align\":\"left\",\"title_size\":\"16\",\"type\":\"manage_status\"},\"layout\":{\"height\":40,\"width\":30,\"x\":112,\"y\":55}},{\"definition\":{\"display_format\":\"three_column\",\"env\":\"datadog.com\",\"service\":\"alerting-cassandra\",\"show_breakdown\":true,\"show_distribution\":true,\"show_errors\":true,\"show_hits\":true,\"show_latency\":false,\"show_resource_list\":false,\"size_format\":\"large\",\"span_name\":\"cassandra.query\",\"time\":{\"live_span\":\"1h\"},\"title\":\"alerting-cassandra #env:datadog.com\",\"title_align\":\"center\",\"title_size\":\"13\",\"type\":\"trace_service\"},\"layout\":{\"height\":38,\"width\":67,\"x\":40,\"y\":28}}]}", uniq)),
				),
			},
		},
	})
}

func TestAccDatadogDashboardJSONImport(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueID := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		// Use checkDashboardDestroy() from Dashboard resource
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardJSONTimeboardJSON(uniqueID),
			},
			{
				ResourceName:      "datadog_dashboard_json.timeboard_json",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckDatadogDashboardJSONTimeboardYAML(uniqueID),
			},
			{
				ResourceName:      "datadog_dashboard_json.timeboard_yaml",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckDatadogDashboardJSONScreenboardJSON(uniqueID),
			},
			{
				ResourceName:      "datadog_dashboard_json.screenboard_json",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckDatadogDashboardJSONScreenboardYAML(uniqueID),
			},
			{
				ResourceName:      "datadog_dashboard_json.screenboard_yaml",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestDatadogDashListInDashboardJSON(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDashListDestroyWithFw(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashListConfigInDashboardJSON(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_json", "dashboard_lists.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_json", "dashboard_lists_removed.#", "0"),
				),
				// The plan is non empty, because in this case the list is the same file
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCheckDatadogDashListConfigRemoveFromDashboardJSON(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_json", "dashboard_lists.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_json", "dashboard_lists_removed.#", "1"),
				),
			},
		},
	})
}

func TestAccDatadogDashboardJSONRbacDiff(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDashListDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardJSONRbacDiff(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_json", "dashboard", fmt.Sprintf("{\"description\":\"Created using the Datadog provider in Terraform\",\"layout_type\":\"ordered\",\"notify_list\":[],\"restricted_roles\":[],\"template_variables\":[],\"title\":\"%s\",\"widgets\":[{\"definition\":{\"alert_id\":\"895605\",\"precision\":3,\"text_align\":\"center\",\"title\":\"Widget Title\",\"type\":\"alert_value\",\"unit\":\"b\"}}]}", uniqueName)),
				),
			},
		},
	})
}

func TestAccDatadogDashboardJSONNoDiff(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDashListDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardJSONNoDiff(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_json", "dashboard", fmt.Sprintf("{\"description\":\"\",\"is_read_only\":false,\"layout_type\":\"ordered\",\"notify_list\":[],\"reflow_type\":\"fixed\",\"template_variables\":[],\"title\":\"%s\",\"widgets\":[]}", uniqueName)),
				),
			},
		},
	})
}

func TestAccDatadogDashboardJSONNotifyListDiff(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDashListDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardJSONNotifyListDiff(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_dashboard_json.timeboard_json", "dashboard", fmt.Sprintf("{\"description\":\"Created using the Datadog provider in Terraform\",\"layout_type\":\"ordered\",\"notify_list\":[\"a-user@example.com\",\"k-user@example.com\",\"z-user1@example.com\"],\"restricted_roles\":[],\"template_variables\":[],\"title\":\"%s\",\"widgets\":[]}", uniqueName)),
				),
			},
		},
	})
}

func testAccCheckDatadogDashboardJSONTimeboardJSON(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_json" "timeboard_json" {
   dashboard = <<EOF
{
   "author_handle":"removed_handle",
   "title":"%s",
   "description":"Created using the Datadog provider in Terraform",
   "widgets":[
      {
         "definition":{
            "title":"Widget Title",
            "type":"alert_graph",
            "alert_id":"895605",
            "viz_type":"timeseries"
         }
      },
      {
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
            "title":"env: prod, datacenter:dc1, service: master-db",
            "title_size":"16",
            "title_align":"left",
            "type":"servicemap",
            "service":"master-db",
            "filters":[
               "env:prod",
               "datacenter:dc1"
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

func testAccCheckDatadogDashboardJSONTimeboardJSONUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_json" "timeboard_json" {
   dashboard = <<EOF
{
   "author_handle":"removed_handle",
   "title":"%s",
   "description":"Created using the Datadog provider in Terraform",
   "widgets":[
      {
         "definition":{
            "title":"Widget Title",
            "type":"alert_graph",
            "alert_id":"895605",
            "viz_type":"timeseries"
         }
      },
      {
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
            "title":"env: prod, datacenter:dc1, service: master-db",
            "title_size":"16",
            "title_align":"left",
            "type":"servicemap",
            "service":"master-db",
            "filters":[
               "env:prod",
               "datacenter:dc1"
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

func testAccCheckDatadogDashboardJSONTimeboardYAML(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_json" "timeboard_yaml" {
   dashboard = jsonencode(yamldecode(
<<EOF
author_handle: removed_handle
title: '%s'
description: Created using the Datadog provider in Terraform
widgets:
 - definition:
     title: Widget Title
     type: alert_graph
     alert_id: '895605'
     viz_type: timeseries
 - id: 5600215046192430
   definition:
     title: Widget Title
     type: alert_value
     alert_id: '895605'
     unit: b
     text_align: center
     precision: 3
 - definition:
     title: Widget Title
     type: alert_value
     alert_id: '895605'
     unit: b
     text_align: center
     precision: 3
 - id: 3887046970315839
   definition:
     title: Widget Title
     type: change
     requests:
       - q: 'avg:system.load.1{env:staging} by {account}'
         compare_to: week_before
         order_by: name
         order_dir: desc
         increase_good: true
         change_type: absolute
         show_present: true
 - id: 1219518175048191
   definition:
     title: Widget Title
     show_legend: false
     type: distribution
     requests:
       - q: 'avg:system.load.1{env:staging} by {account}'
         style:
           palette: warm
 - id: 6039041238503416
   definition:
     title: Widget Title
     type: check_status
     check: aws.ecs.agent_connected
     grouping: cluster
     group_by:
       - account
       - cluster
     tags:
       - 'account:demo'
       - 'cluster:awseb-ruthebdog-env-8-dn3m6u3gvk'
 - id: 5186844025489598
   definition:
     title: Widget Title
     show_legend: false
     type: heatmap
     yaxis:
       scale: sqrt
       include_zero: true
       min: '1'
       max: '2'
     requests:
       - q: 'avg:system.load.1{env:staging} by {account}'
         style:
           palette: warm
 - id: 6742660811820435
   definition:
     title: Widget Title
     type: hostmap
     requests:
       fill:
         q: 'avg:system.load.1{*} by {host}'
       size:
         q: 'avg:memcache.uptime{*} by {host}'
     node_type: container
     no_metric_hosts: true
     no_group_hosts: true
     group:
       - host
       - region
     scope:
       - 'region:us-east-1'
       - 'aws_account:727006795293'
     style:
       palette: yellow_to_green
       palette_flip: true
       fill_min: '10'
       fill_max: '20'
 - id: 1986924343921271
   definition:
     type: note
     content: note text
     background_color: pink
     font_size: '14'
     text_align: center
     show_tick: true
     tick_pos: 50%%
     tick_edge: left
 - id: 3043237513486645
   definition:
     title: Widget Title
     type: query_value
     requests:
       - q: 'avg:system.load.1{env:staging} by {account}'
         aggregator: sum
         conditional_formats:
           - hide_value: false
             comparator: <
             palette: white_on_green
             value: 2
           - hide_value: false
             comparator: '>'
             palette: white_on_red
             value: 2.2
     autoscale: true
     custom_unit: xx
     text_align: right
     precision: 4
 - id: 8636154599297416
   definition:
     title: Widget Title
     type: query_table
     requests:
       - q: 'avg:system.load.1{env:staging} by {account}'
         aggregator: sum
         conditional_formats:
           - hide_value: false
             comparator: <
             palette: white_on_green
             value: 2
           - hide_value: false
             comparator: '>'
             palette: white_on_red
             value: 2.2
         limit: 10
 - id: 518322985317720
   definition:
     title: Widget Title
     type: scatterplot
     requests:
       x:
         q: 'avg:system.cpu.user{*} by {service, account}'
         aggregator: max
       'y':
         q: 'avg:system.mem.used{*} by {service, account}'
         aggregator: min
     xaxis:
       scale: pow
       label: x
       include_zero: true
       min: '1'
       max: '2000'
     yaxis:
       scale: log
       label: 'y'
       include_zero: false
       min: '5'
       max: '2222'
     color_by_groups:
       - account
       - apm-role-group
 - id: 4913548056140044
   definition:
     title: 'env: prod, datacenter:dc1, service: master-db'
     title_size: '16'
     title_align: left
     type: servicemap
     service: master-db
     filters:
       - 'env:prod'
       - 'datacenter:dc1'
 - id: 215209954480975
   definition:
     title: Widget Title
     show_legend: true
     legend_size: '2'
     type: timeseries
     requests:
       - q: 'avg:system.cpu.user{app:general} by {env}'
         on_right_yaxis: false
         metadata:
           - expression: 'avg:system.cpu.user{app:general} by {env}'
             alias_name: Alpha
         style:
           palette: warm
           line_type: dashed
           line_width: thin
         display_type: line
       - on_right_yaxis: false
         log_query:
           index: mcnulty
           search:
             query: 'status:info'
           group_by:
             - facet: host
               sort:
                 facet: '@duration'
                 aggregation: avg
                 order: desc
               limit: 10
           compute:
             facet: '@duration'
             interval: 5000
             aggregation: avg
         display_type: area
       - on_right_yaxis: false
         apm_query:
           index: apm-search
           search:
             query: 'type:web'
           group_by:
             - facet: resource_name
               sort:
                 facet: '@string_query.interval'
                 aggregation: avg
                 order: desc
               limit: 50
           compute:
             facet: '@duration'
             interval: 5000
             aggregation: avg
         display_type: bars
       - on_right_yaxis: false
         process_query:
           search_by: error
           metric: process.stat.cpu.total_pct
           limit: 50
           filter_by:
             - active
         display_type: area
     yaxis:
       scale: log
       include_zero: false
       max: '100'
     events:
       - q: 'sources:test tags:1'
       - q: 'sources:test tags:2'
     markers:
       - label: ' z=6 '
         value: y = 4
         display_type: error dashed
       - label: ' x=8 '
         value: 10 < y < 999
         display_type: ok solid
 - id: 8114292022885770
   definition:
     title: Widget Title
     type: toplist
     requests:
       - q: 'avg:system.cpu.user{app:general} by {env}'
         conditional_formats:
           - hide_value: false
             comparator: <
             palette: white_on_green
             value: 2
           - hide_value: false
             comparator: '>'
             palette: white_on_red
             value: 2.2
 - definition:
     title: Group Widget
     type: group
     layout_type: ordered
     widgets:
       - definition:
           type: note
           content: cluster note widget
           background_color: pink
           font_size: '14'
           text_align: center
           show_tick: true
           tick_pos: 50%%
           tick_edge: left
       - id: 8096017487317681
         definition:
           title: Alert Graph
           type: alert_graph
           alert_id: '123'
           viz_type: toplist
 - id: 7981844470437074
   definition:
     title: Widget Title
     type: slo
     view_type: detail
     time_windows:
       - 7d
       - previous_week
     slo_id: '56789'
     show_error_budget: true
     view_mode: overall
     global_time_target: '0'
template_variables:
 - name: var_1
   default: aws
   prefix: host
 - name: var_2
   default: autoscaling
   prefix: service_name
layout_type: ordered
is_read_only: true
notify_list: []
template_variable_presets:
 - name: preset_1
   template_variables:
     - name: var_1
       value: host.dc
     - name: var_2
       value: my_service
id: 5uw-bbj-xec
EOF
))
}`, uniq)
}

func testAccCheckDatadogDashboardJSONScreenboardJSON(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_json" "screenboard_json" {
   dashboard = <<EOF
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
         "layout":{
            "x":40,
            "y":28,
            "width":67,
            "height":38
         },
         "definition":{
            "title":"alerting-cassandra #env:datadog.com",
            "title_size":"13",
            "title_align":"center",
            "time":{
               "live_span":"1h"
            },
            "type":"trace_service",
            "env":"datadog.com",
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

func testAccCheckDatadogDashboardJSONScreenboardYAML(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_json" "screenboard_yaml" {
   dashboard = jsonencode(yamldecode(
<<EOF
title: '%s'
description: Created using the Datadog provider in Terraform
widgets:
 - layout:
     x: 5
     'y': 5
     width: 32
     height: 43
   definition:
     title: Widget Title
     title_size: '16'
     title_align: left
     time:
       live_span: 1h
     type: event_stream
     query: '*'
     event_size: l
 - layout:
     x: 42
     'y': 73
     width: 65
     height: 9
   definition:
     title: Widget Title
     title_size: '16'
     title_align: left
     time:
       live_span: 1h
     type: event_timeline
     query: '*'
 - id: 1117617615518455
   layout:
     x: 42
     'y': 5
     width: 30
     height: 20
   definition:
     type: free_text
     text: free text content
     color: '#d00'
     font_size: '88'
     text_align: left
 - id: 3098118775539428
   layout:
     x: 111
     'y': 8
     width: 39
     height: 46
   definition:
     type: iframe
     url: 'http://google.com'
 - id: 651713243056399
   layout:
     x: 77
     'y': 7
     width: 30
     height: 20
   definition:
     type: image
     url: >-
       https://images.pexels.com/photos/67636/rose-blue-flower-rose-blooms-67636.jpeg?auto=compress&cs=tinysrgb&h=350
     sizing: fit
     margin: small
 - id: 5458329230004343
   layout:
     x: 5
     'y': 51
     width: 32
     height: 36
   definition:
     type: log_stream
     logset: ''
     indexes:
       - main
     query: error
     sort:
       column: time
       order: desc
     columns:
       - core_host
       - core_service
       - tag_source
     show_date_column: true
     show_message_column: true
     message_display: expanded-md
 - id: 1112741664700765
   layout:
     x: 112
     'y': 55
     width: 30
     height: 40
   definition:
     title: Widget Title
     title_size: '16'
     title_align: left
     type: manage_status
     summary_type: monitors
     display_format: countsAndList
     color_preference: text
     hide_zero_counts: true
     show_last_triggered: false
     query: 'type:metric'
     sort: 'status,asc'
     count: 50
     start: 0
 - id: 6949442529647217
   layout:
     x: 40
     'y': 28
     width: 67
     height: 38
   definition:
     title: 'alerting-cassandra #env:datadog.com'
     title_size: '13'
     title_align: center
     time:
       live_span: 1h
     type: trace_service
     env: datadog.com
     service: alerting-cassandra
     span_name: cassandra.query
     show_hits: true
     show_errors: true
     show_latency: false
     show_breakdown: true
     show_distribution: true
     show_resource_list: false
     size_format: large
     display_format: three_column
template_variables:
 - name: var_1
   default: aws
   prefix: host
 - name: var_2
   default: autoscaling
   prefix: service_name
layout_type: free
is_read_only: false
notify_list: []
template_variable_presets:
 - name: preset_1
   template_variables:
     - name: var_1
       value: host.dc
     - name: var_2
       value: my_service
id: hjf-2xf-xc8
EOF
))
}`, uniq)
}

func testAccCheckDatadogDashListConfigInDashboardJSON(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_list" "new_list" {
	name = "%s"
}

resource "datadog_dashboard_json" "timeboard_json" {
   dashboard_lists = ["${datadog_dashboard_list.new_list.id}"]
   dashboard = <<EOF
{
   "author_handle":"removed_handle",
   "title":"%s",
   "description":"Created using the Datadog provider in Terraform",
   "widgets":[
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
      }
   ],
   "template_variables":[
      
   ],
   "layout_type":"ordered",
   "is_read_only":true,
   "notify_list":[
      
   ],
   "id":"5uw-bbj-xec"
}
EOF
}`, uniq, uniq)
}

func testAccCheckDatadogDashListConfigRemoveFromDashboardJSON(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_list" "new_list" {
	name = "%s"
}

resource "datadog_dashboard_json" "timeboard_json" {
   dashboard = <<EOF
{
   "author_handle":"removed_handle",
   "title":"%s",
   "description":"Created using the Datadog provider in Terraform",
   "widgets":[
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
      }
   ],
   "template_variables":[
      
   ],
   "layout_type":"ordered",
   "is_read_only":true,
   "notify_list":[
      
   ],
   "id":"5uw-bbj-xec"
}
EOF
}`, uniq, uniq)
}

func testAccCheckDatadogDashboardJSONRbacDiff(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_json" "timeboard_json" {
   dashboard = <<EOF
{
   "author_handle":"removed_handle",
   "title":"%s",
   "description":"Created using the Datadog provider in Terraform",
   "widgets":[
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
      }
   ],
   "template_variables":[
      
   ],
   "layout_type":"ordered",
   "is_read_only":true,
   "restricted_roles":[],
   "notify_list":[
      
   ],
   "id":"5uw-bbj-xec"
}
EOF
}`, uniq)
}

func testAccCheckDatadogDashboardJSONNoDiff(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_json" "timeboard_json" {
   dashboard = <<EOF
{
   "title": "%s",
   "description": "",
   "widgets": [],
   "template_variables": [],
   "layout_type": "ordered",
   "notify_list": [],
   "reflow_type": "fixed",
   "id": "3fa-nkp-wty"
}
EOF
}`, uniq)
}

func testAccCheckDatadogDashboardJSONNotifyListDiff(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "one" {
  email     = "z-user1@example.com"
  name      = "Test User"
}
resource "datadog_user" "two" {
  email     = "a-user@example.com"
  name      = "Test User"
}
resource "datadog_user" "three" {
  email     = "k-user@example.com"
  name      = "Test User"
}

resource "datadog_dashboard_json" "timeboard_json" {
   dashboard = <<EOF
{
   "author_handle":"removed_handle",
   "title":"%s",
   "description":"Created using the Datadog provider in Terraform",
   "widgets":[],
   "template_variables":[
      
   ],
   "layout_type":"ordered",
   "is_read_only":true,
   "restricted_roles":[],
   "notify_list":["${datadog_user.one.email}","${datadog_user.two.email}","${datadog_user.three.email}"],
   "id":"5uw-bbj-xec"
}
EOF
   depends_on = [
      datadog_user.one,
      datadog_user.two,
      datadog_user.three,
   ]
}`, uniq)
}
