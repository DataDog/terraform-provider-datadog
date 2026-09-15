package dashboardmapping

import (
	"reflect"
	"testing"
)

func TestEmbeddedAppWidgetRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"embedded_app_definition": []interface{}{map[string]interface{}{
			"app_id": "7e7745f9-4343-4927-b038-80934a355915",
			"title":  "Environment manager",
			"input": []interface{}{
				map[string]interface{}{"name": "environment", "value": `"production"`},
				map[string]interface{}{"name": "replicas", "value": `3`},
				map[string]interface{}{"name": "dry_run", "value": `false`},
				map[string]interface{}{"name": "filters", "value": `{"service":"web","regions":["us-east-1","eu-west-1"]}`},
			},
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	expectedDefinition := map[string]interface{}{
		"type":   "embedded_app",
		"app_id": "7e7745f9-4343-4927-b038-80934a355915",
		"title":  "Environment manager",
		"inputs": []interface{}{
			map[string]interface{}{"name": "environment", "value": "production"},
			map[string]interface{}{"name": "replicas", "value": float64(3)},
			map[string]interface{}{"name": "dry_run", "value": false},
			map[string]interface{}{"name": "filters", "value": map[string]interface{}{
				"service": "web",
				"regions": []interface{}{"us-east-1", "eu-west-1"},
			}},
		},
	}
	if !reflect.DeepEqual(built["definition"], expectedDefinition) {
		t.Fatalf("embedded app widget was not serialized correctly:\n got: %#v\nwant: %#v", built["definition"], expectedDefinition)
	}

	flattened, warnings := FlattenWidgetEngineJSON(built)
	if len(warnings) != 0 {
		t.Fatalf("unexpected flatten warnings: %v", warnings)
	}
	definition := flattened["embedded_app_definition"].([]interface{})[0].(map[string]interface{})
	if definition["app_id"] != "7e7745f9-4343-4927-b038-80934a355915" {
		t.Fatalf("app_id was not restored: %#v", definition)
	}
	inputs := definition["input"].([]interface{})
	wantValues := []string{`"production"`, `3`, `false`, `{"regions":["us-east-1","eu-west-1"],"service":"web"}`}
	for i, want := range wantValues {
		got := inputs[i].(map[string]interface{})["value"]
		if got != want {
			t.Fatalf("input %d value was not restored: got %q, want %q", i, got, want)
		}
	}
}

func TestEmbeddedAppTemplateWidgetRoundTrip(t *testing.T) {
	widget := map[string]interface{}{
		"embedded_app_definition": []interface{}{map[string]interface{}{
			"template_id": "ec2_manager",
		}},
	}

	built := BuildWidgetEngineJSONFromMap(widget)
	definition := built["definition"].(map[string]interface{})
	if definition["type"] != "embedded_app" || definition["template_id"] != "ec2_manager" {
		t.Fatalf("embedded app template widget was not serialized correctly: %#v", built)
	}

	flattened, warnings := FlattenWidgetEngineJSON(built)
	if len(warnings) != 0 {
		t.Fatalf("unexpected flatten warnings: %v", warnings)
	}
	flattenedDefinition := flattened["embedded_app_definition"].([]interface{})[0].(map[string]interface{})
	if flattenedDefinition["template_id"] != "ec2_manager" {
		t.Fatalf("template_id was not restored: %#v", flattenedDefinition)
	}
}
