package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	ctyjson "github.com/hashicorp/go-cty/cty/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// powerpackTestResourceData builds ResourceData whose raw config matches cfg,
// the way Terraform passes it to Create and Update.
func powerpackTestResourceData(t *testing.T, r *schema.Resource, cfg map[string]interface{}) *schema.ResourceData {
	t.Helper()
	d := schema.TestResourceDataRaw(t, r.SchemaMap(), cfg)
	d.SetId("powerpack-id")
	state := d.State()

	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := ctyjson.Unmarshal(b, r.CoreConfigSchema().ImpliedType())
	if err != nil {
		t.Fatal(err)
	}
	state.RawConfig = raw
	return r.Data(state)
}

func TestPowerpackShowTitlePayload(t *testing.T) {
	falseVal, trueVal := false, true
	builders := map[string]struct {
		resource *schema.Resource
		build    func(d *schema.ResourceData) (*datadogV2.Powerpack, error)
	}{
		"datadog_powerpack": {
			resource: resourceDatadogPowerpack(),
			build: func(d *schema.ResourceData) (*datadogV2.Powerpack, error) {
				p, diags := buildDatadogPowerpack(context.Background(), d)
				if diags.HasError() {
					return nil, fmt.Errorf("%v", diags)
				}
				return p, nil
			},
		},
		"datadog_powerpack_v2": {
			resource: resourceDatadogPowerpackV2(),
			build:    buildPowerpackPayload,
		},
	}
	cases := []struct {
		name string
		cfg  map[string]interface{}
		want *bool
	}{
		{name: "explicit false is sent", cfg: map[string]interface{}{"name": "pp", "show_title": false}, want: &falseVal},
		{name: "explicit true is sent", cfg: map[string]interface{}{"name": "pp", "show_title": true}, want: &trueVal},
		{name: "unset is omitted", cfg: map[string]interface{}{"name": "pp"}, want: nil},
	}

	for resourceName, b := range builders {
		for _, c := range cases {
			t.Run(resourceName+"/"+c.name, func(t *testing.T) {
				d := powerpackTestResourceData(t, b.resource, c.cfg)
				payload, err := b.build(d)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				got := payload.Data.Attributes.GroupWidget.Definition.ShowTitle
				switch {
				case c.want == nil && got != nil:
					t.Fatalf("expected show_title to be omitted, got %v", *got)
				case c.want != nil && got == nil:
					t.Fatalf("expected show_title = %v, got it omitted", *c.want)
				case c.want != nil && *got != *c.want:
					t.Fatalf("expected show_title = %v, got %v", *c.want, *got)
				}
			})
		}
	}
}
