package fwprovider

import (
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// datastoreItemByKey returns the item whose primary key column equals key, or
// nil when the listing holds no such item.
//
// The listing is filtered server-side, but nothing guarantees the response is a
// single row, so the key is checked here as well.
func datastoreItemByKey(items []datadogV2.ItemApiPayloadData, key string) *datadogV2.ItemApiPayloadData {
	for i := range items {
		attributes := items[i].GetAttributes()
		column := attributes.GetPrimaryColumnName()
		if column == "" {
			continue
		}
		if value, ok := attributes.GetValue()[column]; ok && fmt.Sprintf("%v", value) == key {
			return &items[i]
		}
	}
	return nil
}

// datastoreItemValueToMap flattens an item's columns to strings, the element
// type of the `value` attribute.
//
// A column of the datastore's JSON type arrives as a Go map or slice and is
// rendered as JSON, so jsondecode() can read it. Scalars keep their fmt
// rendering, and a string column is passed through unquoted.
func datastoreItemValueToMap(value map[string]interface{}) (map[string]string, error) {
	out := make(map[string]string, len(value))
	for column, raw := range value {
		switch typed := raw.(type) {
		case string:
			out[column] = typed
		case map[string]interface{}, []interface{}:
			encoded, err := json.Marshal(typed)
			if err != nil {
				return nil, fmt.Errorf("column %q: %w", column, err)
			}
			out[column] = string(encoded)
		default:
			out[column] = fmt.Sprintf("%v", typed)
		}
	}
	return out, nil
}
