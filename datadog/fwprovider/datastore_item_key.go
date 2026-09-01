package fwprovider

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// datastoreItemByKey returns the item whose primary key column equals key, or
// nil when the listing holds no such item.
//
// The server-side filter is a glob, so the listing can hold several items or
// none of them the wanted one.
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
