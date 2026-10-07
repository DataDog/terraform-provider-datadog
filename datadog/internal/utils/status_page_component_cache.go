package utils

import (
	"context"
	"net/http"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
)

type statusPageComponentCache struct {
	mu        sync.Mutex
	responses map[uuid.UUID]*datadogV2.StatusPagesComponentArray
}

// ListStatusPageComponents returns one component listing per status page per
// ApiInstances. Reading each component individually exhausts the 60 req/min
// per-user limit on GET /statuspages/{page_id}/components/{component_id} once a
// configuration spans more than a few dozen components, so callers resolve a
// single component out of the page listing instead. Holding the cache lock
// while fetching coalesces parallel reads.
//
// The listing only contains top-level components; components nested inside a
// group are returned under the group's attributes in a reduced form. Callers
// must fall back to GetComponent when an ID is absent from the listing rather
// than treating it as deleted.
func (i *ApiInstances) ListStatusPageComponents(ctx context.Context, pageID uuid.UUID) (datadogV2.StatusPagesComponentArray, *http.Response, error) {
	cache := &i.statusPageComponentCache
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if response, ok := cache.responses[pageID]; ok {
		return *response, nil, nil
	}

	response, httpResponse, err := i.GetStatusPagesApiV2().ListComponents(ctx, pageID)
	if err == nil && (httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound) {
		if cache.responses == nil {
			cache.responses = make(map[uuid.UUID]*datadogV2.StatusPagesComponentArray)
		}
		cache.responses[pageID] = &response
	}

	return response, httpResponse, err
}

// InvalidateStatusPageComponentCache clears the cached listing for a status
// page so the next read fetches a fresh snapshot. Callers must invalidate after
// mutating a component.
func (i *ApiInstances) InvalidateStatusPageComponentCache(pageID uuid.UUID) {
	cache := &i.statusPageComponentCache
	cache.mu.Lock()
	defer cache.mu.Unlock()

	delete(cache.responses, pageID)
}
