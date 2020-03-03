package datadog

import "log"

func resourceDatadogSyntheticsTestStateUpgradeV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] rawState before Migration: %#v", rawState)
	options := rawState["options"].(map[string]interface{})
	rawState["options"] = []interface{}{options}
	log.Printf("[DEBUG] rawState after Migration: %#v", rawState)
	return rawState, nil
}
