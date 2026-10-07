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
	componentTestPageA     = uuid.MustParse("00000000-0000-0000-0000-0000000000c1")
	componentTestPageB     = uuid.MustParse("00000000-0000-0000-0000-0000000000c2")
	componentTestIDFirst   = uuid.MustParse("00000000-0000-0000-0000-00000000f001")
	componentTestIDSecond  = uuid.MustParse("00000000-0000-0000-0000-00000000f002")
	componentTestIDRecover = uuid.MustParse("00000000-0000-0000-0000-00000000f003")
)

func TestStatusPageComponentCacheCoalescesConcurrentReads(t *testing.T) {
	var requestCount atomic.Int32
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})

	apiInstances, closeServer := newStatusPageComponentCacheTestAPIInstances(t, func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		close(requestStarted)
		<-releaseRequest
		writeComponentListResponse(t, w, componentTestIDFirst)
	})
	defer closeServer()

	const readers = 20
	var wg sync.WaitGroup
	wg.Add(readers)
	errs := make(chan error, readers)
	for range readers {
		go func() {
			defer wg.Done()
			response, _, err := apiInstances.ListStatusPageComponents(context.Background(), componentTestPageA)
			if err != nil {
				errs <- err
				return
			}
			if got := response.GetData()[0].GetId(); got != componentTestIDFirst {
				errs <- fmt.Errorf("component ID = %s, want %s", got, componentTestIDFirst)
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

func TestStatusPageComponentCacheHitsAndInvalidation(t *testing.T) {
	var requestCount atomic.Int32
	apiInstances, closeServer := newStatusPageComponentCacheTestAPIInstances(t, func(w http.ResponseWriter, _ *http.Request) {
		if requestCount.Add(1) == 1 {
			writeComponentListResponse(t, w, componentTestIDFirst)
			return
		}
		writeComponentListResponse(t, w, componentTestIDSecond)
	})
	defer closeServer()

	for call := 0; call < 2; call++ {
		response, _, err := apiInstances.ListStatusPageComponents(context.Background(), componentTestPageA)
		if err != nil {
			t.Fatalf("cached read %d failed: %v", call+1, err)
		}
		if got := response.GetData()[0].GetId(); got != componentTestIDFirst {
			t.Fatalf("cached read %d component ID = %s, want %s", call+1, got, componentTestIDFirst)
		}
	}
	if got := requestCount.Load(); got != 1 {
		t.Fatalf("request count before invalidation = %d, want 1", got)
	}

	apiInstances.InvalidateStatusPageComponentCache(componentTestPageA)

	response, _, err := apiInstances.ListStatusPageComponents(context.Background(), componentTestPageA)
	if err != nil {
		t.Fatalf("read after invalidation failed: %v", err)
	}
	if got := response.GetData()[0].GetId(); got != componentTestIDSecond {
		t.Fatalf("component ID after invalidation = %s, want %s", got, componentTestIDSecond)
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("request count after invalidation = %d, want 2", got)
	}
}

// Each status page is cached independently: reading one page must not serve the
// other's listing, and invalidating one must not evict the other.
func TestStatusPageComponentCacheIsPerPage(t *testing.T) {
	var requestCount atomic.Int32
	apiInstances, closeServer := newStatusPageComponentCacheTestAPIInstances(t, func(w http.ResponseWriter, request *http.Request) {
		requestCount.Add(1)
		id := componentTestIDFirst
		if request.URL.Path == fmt.Sprintf("/api/v2/statuspages/%s/components", componentTestPageB) {
			id = componentTestIDSecond
		}
		writeComponentListResponse(t, w, id)
	})
	defer closeServer()

	for _, tc := range []struct {
		pageID uuid.UUID
		want   uuid.UUID
	}{
		{componentTestPageA, componentTestIDFirst},
		{componentTestPageB, componentTestIDSecond},
		{componentTestPageA, componentTestIDFirst},
	} {
		response, _, err := apiInstances.ListStatusPageComponents(context.Background(), tc.pageID)
		if err != nil {
			t.Fatalf("read of page %s failed: %v", tc.pageID, err)
		}
		if got := response.GetData()[0].GetId(); got != tc.want {
			t.Fatalf("component ID = %s, want %s", got, tc.want)
		}
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("request count = %d, want 2", got)
	}

	apiInstances.InvalidateStatusPageComponentCache(componentTestPageA)

	if _, _, err := apiInstances.ListStatusPageComponents(context.Background(), componentTestPageB); err != nil {
		t.Fatalf("read of page B after invalidating page A failed: %v", err)
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("request count after invalidating page A = %d, want 2", got)
	}
}

func TestStatusPageComponentCacheDoesNotCacheFailures(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusTooManyRequests, http.StatusNotFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var requestCount atomic.Int32
			apiInstances, closeServer := newStatusPageComponentCacheTestAPIInstances(t, func(w http.ResponseWriter, _ *http.Request) {
				if requestCount.Add(1) == 1 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_, _ = w.Write([]byte(`{"errors":["test failure"]}`))
					return
				}
				writeComponentListResponse(t, w, componentTestIDRecover)
			})
			defer closeServer()

			if _, _, err := apiInstances.ListStatusPageComponents(context.Background(), componentTestPageA); err == nil {
				t.Fatalf("first read with status %d unexpectedly succeeded", status)
			}

			for call := 0; call < 2; call++ {
				response, _, err := apiInstances.ListStatusPageComponents(context.Background(), componentTestPageA)
				if err != nil {
					t.Fatalf("successful read %d failed: %v", call+1, err)
				}
				if got := response.GetData()[0].GetId(); got != componentTestIDRecover {
					t.Fatalf("successful read %d component ID = %s, want %s", call+1, got, componentTestIDRecover)
				}
			}
			if got := requestCount.Load(); got != 2 {
				t.Fatalf("request count = %d, want 2", got)
			}
		})
	}
}

func newStatusPageComponentCacheTestAPIInstances(t *testing.T, handler http.HandlerFunc) (*ApiInstances, func()) {
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

func writeComponentListResponse(t *testing.T, w http.ResponseWriter, id uuid.UUID) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if _, err := fmt.Fprintf(w, `{"data":[{"id":%q,"type":"components"}]}`, id); err != nil {
		t.Errorf("writing response: %v", err)
	}
}
