package utils

import (
	"context"
	"net/http"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
)

type statusPageDegradationTemplateCache struct {
	mu        sync.Mutex
	responses map[uuid.UUID]*datadogV2.DegradationTemplateArray
}

// ListStatusPageDegradationTemplates returns one listing per status page per
// ApiInstances. Reading each degradation template individually exhausts the
// 60 req/min per-user limit on
// GET /statuspages/{page_id}/degradation_templates/{template_id} once a page
// holds more than a few dozen templates, so callers resolve a single template
// out of the page listing instead. Holding the cache lock while fetching
// coalesces parallel reads.
func (i *ApiInstances) ListStatusPageDegradationTemplates(ctx context.Context, pageID uuid.UUID) (datadogV2.DegradationTemplateArray, *http.Response, error) {
	cache := &i.statusPageDegradationTemplateCache
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if response, ok := cache.responses[pageID]; ok {
		return *response, nil, nil
	}

	response, httpResponse, err := i.GetStatusPagesApiV2().ListDegradationTemplates(ctx, pageID)
	if err == nil && (httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound) {
		if cache.responses == nil {
			cache.responses = make(map[uuid.UUID]*datadogV2.DegradationTemplateArray)
		}
		cache.responses[pageID] = &response
	}

	return response, httpResponse, err
}

// InvalidateStatusPageDegradationTemplateCache clears the cached listing for a
// status page so the next read fetches a fresh snapshot. Callers must
// invalidate after mutating a template.
func (i *ApiInstances) InvalidateStatusPageDegradationTemplateCache(pageID uuid.UUID) {
	cache := &i.statusPageDegradationTemplateCache
	cache.mu.Lock()
	defer cache.mu.Unlock()

	delete(cache.responses, pageID)
}
