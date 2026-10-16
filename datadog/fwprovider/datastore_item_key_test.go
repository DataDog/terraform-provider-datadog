package fwprovider

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

func datastoreItem(primaryColumn string, value map[string]interface{}) datadogV2.ItemApiPayloadData {
	attributes := datadogV2.NewItemApiPayloadDataAttributesWithDefaults()
	attributes.SetPrimaryColumnName(primaryColumn)
	attributes.SetValue(value)

	item := datadogV2.NewItemApiPayloadDataWithDefaults()
	item.SetAttributes(*attributes)
	return *item
}

func TestDatastoreItemByKey(t *testing.T) {
	items := []datadogV2.ItemApiPayloadData{
		datastoreItem("product", map[string]interface{}{"product": "aaa"}),
		datastoreItem("product", map[string]interface{}{"product": "apm"}),
		datastoreItem("id", map[string]interface{}{"id": float64(42)}),
	}

	for _, tc := range []struct {
		name string
		key  string
		want string
	}{
		{"first item", "aaa", "aaa"},
		{"later item", "apm", "apm"},
		{"non-string key", "42", "42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := datastoreItemByKey(items, tc.key)
			if item == nil {
				t.Fatalf("no item for key %q", tc.key)
			}
			attributes := item.GetAttributes()
			got := attributes.GetValue()[attributes.GetPrimaryColumnName()]
			if got != tc.want && got != float64(42) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}

	if item := datastoreItemByKey(items, "missing"); item != nil {
		t.Fatalf("got an item for a key that is not in the listing")
	}
	if item := datastoreItemByKey(nil, "aaa"); item != nil {
		t.Fatalf("got an item from an empty listing")
	}
}
