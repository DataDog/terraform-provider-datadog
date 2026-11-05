package fwprovider

import (
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
