package utils

import (
	"context"
	"net/http"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

type sensitiveDataScannerConfigCache struct {
	mu       sync.Mutex
	response *datadogV2.SensitiveDataScannerGetConfigResponse
}

// ListSensitiveDataScannerGroups returns one configuration snapshot per
// ApiInstances. Holding the cache lock while fetching coalesces parallel reads.
func (i *ApiInstances) ListSensitiveDataScannerGroups(ctx context.Context) (datadogV2.SensitiveDataScannerGetConfigResponse, *http.Response, error) {
	cache := &i.sensitiveDataScannerConfigCache
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.response != nil {
		return *cache.response, nil, nil
	}

	response, httpResponse, err := i.GetSensitiveDataScannerApiV2().ListScanningGroups(ctx)
	if err == nil && (httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound) {
		cache.response = &response
	}

	return response, httpResponse, err
}

// InvalidateSensitiveDataScannerConfigCache clears the cached configuration so
// the next read fetches a fresh snapshot.
func (i *ApiInstances) InvalidateSensitiveDataScannerConfigCache() {
	cache := &i.sensitiveDataScannerConfigCache
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.response = nil
}
