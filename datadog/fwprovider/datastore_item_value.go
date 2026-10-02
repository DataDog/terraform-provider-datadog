package fwprovider

import (
	"encoding/json"
	"fmt"
	"log"
)

// datastoreItemValueToMap flattens an item's columns to strings, the element
// type of the `value` attribute.
//
// A column of the datastore's JSON type arrives as a Go map or slice and is
// rendered as JSON, so jsondecode() can read it. Scalars keep their fmt
// rendering, and a string column is passed through unquoted.
func datastoreItemValueToMap(value map[string]interface{}) map[string]string {
	out := make(map[string]string, len(value))
	for column, raw := range value {
		switch typed := raw.(type) {
		case string:
			out[column] = typed
		case map[string]interface{}, []interface{}:
			// The value was decoded from the API's JSON, so it encodes back.
			encoded, err := json.Marshal(typed)
			if err != nil {
				log.Printf("[WARN] datadog_datastore_item: column %q is not encodable to JSON (%v); returning its Go rendering", column, err)
				out[column] = fmt.Sprintf("%v", typed)
				break
			}
			out[column] = string(encoded)
		default:
			out[column] = fmt.Sprintf("%v", typed)
		}
	}
	return out
}
