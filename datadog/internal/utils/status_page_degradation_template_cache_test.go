package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/google/uuid"
)

var (
	degradationTemplateTestPageA = uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	degradationTemplateTestPageB = uuid.MustParse("00000000-0000-0000-0000-0000000000b2")
)

func TestStatusPageDegradationTemplateCacheCoalescesConcurrentReads(t *testing.T) {
	var requestCount atomic.Int32
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})

	apiInstances, closeServer := newStatusPageDegradationTemplateCacheTestAPIInstances(t, func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		close(requestStarted)
		<-releaseRequest
		writeDegradationTemplateListResponse(t, w, "snapshot")
	})
	defer closeServer()

	const readers = 20
	var wg sync.WaitGroup
	wg.Add(readers)
	errs := make(chan error, readers)
	for range readers {
		go func() {
			defer wg.Done()
			response, _, err := apiInstances.ListStatusPageDegradationTemplates(context.Background(), degradationTemplateTestPageA)
			if err != nil {
				errs <- err
				return
			}
			if got := response.GetData()[0].GetId(); got != "snapshot" {
				errs <- fmt.Errorf("template ID = %q, want snapshot", got)
			}
		}()
	}

	<-requestStarted
	close(releaseRequest)
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
	if got := requestCount.Load(); got != 1 {
		t.Fatalf("request count = %d, want 1", got)
	}
}

func TestStatusPageDegradationTemplateCacheHitsAndInvalidation(t *testing.T) {
	var requestCount atomic.Int32
	apiInstances, closeServer := newStatusPageDegradationTemplateCacheTestAPIInstances(t, func(w http.ResponseWriter, _ *http.Request) {
		count := requestCount.Add(1)
		writeDegradationTemplateListResponse(t, w, fmt.Sprintf("snapshot-%d", count))
	})
	defer closeServer()

	for call := 0; call < 2; call++ {
		response, _, err := apiInstances.ListStatusPageDegradationTemplates(context.Background(), degradationTemplateTestPageA)
		if err != nil {
			t.Fatalf("cached read %d failed: %v", call+1, err)
		}
		if got := response.GetData()[0].GetId(); got != "snapshot-1" {
			t.Fatalf("cached read %d template ID = %q, want snapshot-1", call+1, got)
		}
	}
	if got := requestCount.Load(); got != 1 {
		t.Fatalf("request count before invalidation = %d, want 1", got)
	}

	apiInstances.InvalidateStatusPageDegradationTemplateCache(degradationTemplateTestPageA)

	response, _, err := apiInstances.ListStatusPageDegradationTemplates(context.Background(), degradationTemplateTestPageA)
	if err != nil {
		t.Fatalf("read after invalidation failed: %v", err)
	}
	if got := response.GetData()[0].GetId(); got != "snapshot-2" {
		t.Fatalf("template ID after invalidation = %q, want snapshot-2", got)
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("request count after invalidation = %d, want 2", got)
	}
}

// Each status page is cached independently: reading one page must not serve the
// other's listing, and invalidating one must not evict the other.
func TestStatusPageDegradationTemplateCacheIsPerPage(t *testing.T) {
	var requestCount atomic.Int32
	apiInstances, closeServer := newStatusPageDegradationTemplateCacheTestAPIInstances(t, func(w http.ResponseWriter, request *http.Request) {
		requestCount.Add(1)
		writeDegradationTemplateListResponse(t, w, request.URL.Path)
	})
	defer closeServer()

	pageAPath := fmt.Sprintf("/api/v2/statuspages/%s/degradation_templates", degradationTemplateTestPageA)
	pageBPath := fmt.Sprintf("/api/v2/statuspages/%s/degradation_templates", degradationTemplateTestPageB)

	for _, tc := range []struct {
		pageID uuid.UUID
		want   string
	}{
		{degradationTemplateTestPageA, pageAPath},
		{degradationTemplateTestPageB, pageBPath},
		{degradationTemplateTestPageA, pageAPath},
	} {
		response, _, err := apiInstances.ListStatusPageDegradationTemplates(context.Background(), tc.pageID)
		if err != nil {
			t.Fatalf("read of page %s failed: %v", tc.pageID, err)
		}
		if got := response.GetData()[0].GetId(); got != tc.want {
			t.Fatalf("template ID = %q, want %q", got, tc.want)
		}
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("request count = %d, want 2", got)
	}

	apiInstances.InvalidateStatusPageDegradationTemplateCache(degradationTemplateTestPageA)

	if _, _, err := apiInstances.ListStatusPageDegradationTemplates(context.Background(), degradationTemplateTestPageB); err != nil {
		t.Fatalf("read of page B after invalidating page A failed: %v", err)
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("request count after invalidating page A = %d, want 2", got)
	}
}

func TestStatusPageDegradationTemplateCacheDoesNotCacheFailures(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusTooManyRequests, http.StatusNotFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var requestCount atomic.Int32
			apiInstances, closeServer := newStatusPageDegradationTemplateCacheTestAPIInstances(t, func(w http.ResponseWriter, _ *http.Request) {
				if requestCount.Add(1) == 1 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_, _ = w.Write([]byte(`{"errors":["test failure"]}`))
					return
				}
				writeDegradationTemplateListResponse(t, w, "recovered")
			})
			defer closeServer()

			if _, _, err := apiInstances.ListStatusPageDegradationTemplates(context.Background(), degradationTemplateTestPageA); err == nil {
				t.Fatalf("first read with status %d unexpectedly succeeded", status)
			}

			for call := 0; call < 2; call++ {
				response, _, err := apiInstances.ListStatusPageDegradationTemplates(context.Background(), degradationTemplateTestPageA)
				if err != nil {
					t.Fatalf("successful read %d failed: %v", call+1, err)
				}
				if got := response.GetData()[0].GetId(); got != "recovered" {
					t.Fatalf("successful read %d template ID = %q, want recovered", call+1, got)
				}
			}
			if got := requestCount.Load(); got != 2 {
				t.Fatalf("request count = %d, want 2", got)
			}
		})
	}
}

func newStatusPageDegradationTemplateCacheTestAPIInstances(t *testing.T, handler http.HandlerFunc) (*ApiInstances, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Errorf("request method = %s, want GET", request.Method)
		}
		handler(w, request)
	}))

	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()

	return &ApiInstances{HttpClient: datadog.NewAPIClient(config)}, server.Close
}

func writeDegradationTemplateListResponse(t *testing.T, w http.ResponseWriter, id string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if _, err := fmt.Fprintf(w, `{"data":[{"id":%q,"type":"degradation_templates"}]}`, id); err != nil {
		t.Errorf("writing response: %v", err)
	}
}
