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
)

func TestSensitiveDataScannerConfigCacheCoalescesConcurrentReads(t *testing.T) {
	var requestCount atomic.Int32
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})

	apiInstances, closeServer := newSensitiveDataScannerCacheTestAPIInstances(t, func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		close(requestStarted)
		<-releaseRequest
		writeSensitiveDataScannerConfigResponse(t, w, "snapshot")
	})
	defer closeServer()

	const readers = 20
	var wg sync.WaitGroup
	wg.Add(readers)
	errs := make(chan error, readers)
	for range readers {
		go func() {
			defer wg.Done()
			response, _, err := apiInstances.ListSensitiveDataScannerGroups(context.Background())
			if err != nil {
				errs <- err
				return
			}
			if got := response.Data.GetId(); got != "snapshot" {
				errs <- fmt.Errorf("response ID = %q, want snapshot", got)
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

func TestSensitiveDataScannerConfigCacheHitsAndInvalidation(t *testing.T) {
	var requestCount atomic.Int32
	apiInstances, closeServer := newSensitiveDataScannerCacheTestAPIInstances(t, func(w http.ResponseWriter, _ *http.Request) {
		count := requestCount.Add(1)
		writeSensitiveDataScannerConfigResponse(t, w, fmt.Sprintf("snapshot-%d", count))
	})
	defer closeServer()

	for call := 0; call < 2; call++ {
		response, _, err := apiInstances.ListSensitiveDataScannerGroups(context.Background())
		if err != nil {
			t.Fatalf("cached read %d failed: %v", call+1, err)
		}
		if got := response.Data.GetId(); got != "snapshot-1" {
			t.Fatalf("cached read %d response ID = %q, want snapshot-1", call+1, got)
		}
	}
	if got := requestCount.Load(); got != 1 {
		t.Fatalf("request count before invalidation = %d, want 1", got)
	}

	apiInstances.InvalidateSensitiveDataScannerConfigCache()

	response, _, err := apiInstances.ListSensitiveDataScannerGroups(context.Background())
	if err != nil {
		t.Fatalf("read after invalidation failed: %v", err)
	}
	if got := response.Data.GetId(); got != "snapshot-2" {
		t.Fatalf("response ID after invalidation = %q, want snapshot-2", got)
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("request count after invalidation = %d, want 2", got)
	}
}

func TestSensitiveDataScannerConfigCacheDoesNotCacheFailures(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusNotFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var requestCount atomic.Int32
			apiInstances, closeServer := newSensitiveDataScannerCacheTestAPIInstances(t, func(w http.ResponseWriter, _ *http.Request) {
				if requestCount.Add(1) == 1 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_, _ = w.Write([]byte(`{"errors":["test failure"]}`))
					return
				}
				writeSensitiveDataScannerConfigResponse(t, w, "recovered")
			})
			defer closeServer()

			if _, _, err := apiInstances.ListSensitiveDataScannerGroups(context.Background()); err == nil {
				t.Fatalf("first read with status %d unexpectedly succeeded", status)
			}

			for call := 0; call < 2; call++ {
				response, _, err := apiInstances.ListSensitiveDataScannerGroups(context.Background())
				if err != nil {
					t.Fatalf("successful read %d failed: %v", call+1, err)
				}
				if got := response.Data.GetId(); got != "recovered" {
					t.Fatalf("successful read %d response ID = %q, want recovered", call+1, got)
				}
			}
			if got := requestCount.Load(); got != 2 {
				t.Fatalf("request count = %d, want 2", got)
			}
		})
	}
}

func newSensitiveDataScannerCacheTestAPIInstances(t *testing.T, handler http.HandlerFunc) (*ApiInstances, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Errorf("request method = %s, want GET", request.Method)
		}
		if request.URL.Path != "/api/v2/sensitive-data-scanner/config" {
			t.Errorf("request path = %s, want /api/v2/sensitive-data-scanner/config", request.URL.Path)
		}
		handler(w, request)
	}))

	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()

	return &ApiInstances{HttpClient: datadog.NewAPIClient(config)}, server.Close
}

func writeSensitiveDataScannerConfigResponse(t *testing.T, w http.ResponseWriter, id string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if _, err := fmt.Fprintf(w, `{"data":{"id":%q,"type":"sensitive_data_scanner_configuration"}}`, id); err != nil {
		t.Errorf("writing response: %v", err)
	}
}
