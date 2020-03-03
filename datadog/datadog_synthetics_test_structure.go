package datadog

import (
	datadog "github.com/zorkian/go-datadog-api"
)

func flattenOptions(options datadog.SyntheticsOptions) []interface{} {
	localOptions := make(map[string]interface{})
	if options.HasFollowRedirects() {
		localOptions["follow_redirects"] = options.GetFollowRedirects()
	}
	if options.HasMinFailureDuration() {
		localOptions["min_failure_duration"] = options.GetMinFailureDuration()
	}
	if options.HasMinLocationFailed() {
		localOptions["min_location_failed"] = options.GetMinLocationFailed()
	}
	if options.HasTickEvery() {
		localOptions["tick_every"] = options.GetTickEvery()
	}
	if options.HasAcceptSelfSigned() {
		localOptions["accept_self_signed"] = options.GetAcceptSelfSigned()
	}
	if options.HasRetry() {
		localOptions["retry"] = flattenRetry(options.GetRetry())
	}
	return []interface{}{localOptions}
}

func flattenRetry(r datadog.Retry) []interface{} {
	m := make(map[string]interface{})
	m["interval"] = r.Interval
	m["count"] = r.Count
	return []interface{}{m}
}

func expandOptions(m map[string]interface{}) *datadog.SyntheticsOptions {
	d := &datadog.SyntheticsOptions{
		TickEvery:          datadog.Int(m["tick_every"].(int)),
		FollowRedirects:    datadog.Bool(m["follow_redirects"].(bool)),
		MinFailureDuration: datadog.Int(m["min_failure_duration"].(int)),
		MinLocationFailed:  datadog.Int(m["min_location_failed"].(int)),
		AcceptSelfSigned:   datadog.Bool(m["accept_self_signed"].(bool)),
	}
	if v, ok := m["retry"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		d.Retry = expandRetry(m["retry"].([]interface{})[0].(map[string]interface{}))
	}
	return d
}

func expandRetry(m map[string]interface{}) *datadog.Retry {
	return &datadog.Retry{
		Count:    datadog.Int(m["count"].(int)),
		Interval: datadog.Int(m["interval"].(int)),
	}
}
