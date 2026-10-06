package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	api "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/terraformauth"
)

// The SDK calls Done when it starts waiting for an exchange. Observing that
// point lets the test hold the exchange open until all followers have joined.
type wifWaitingContext struct {
	context.Context
	once    sync.Once
	waiting chan<- struct{}
}

func (c *wifWaitingContext) Done() <-chan struct{} {
	c.once.Do(func() { c.waiting <- struct{}{} })
	return c.Context.Done()
}

func awaitWIFRequest[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for parallel WIF requests")
		var zero T
		return zero
	}
}

func TestWIFParallelRequests(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		for _, identity := range []string{"terraform", "aws"} {
			t.Run(implementation+"/"+identity, func(t *testing.T) {
				clearWorkloadIdentityTestEnv(t)
				token := workloadIdentityTestToken(t, "https://app.terraform.io")
				if identity == "terraform" {
					t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, token)
				}
				type exchangePhase struct {
					entered chan struct{}
					release chan struct{}
					reject  bool
				}
				var phase atomic.Pointer[exchangePhase]
				var exchanges, requests atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					switch r.URL.Path {
					case "/api/v2/delegated-token":
						number := exchanges.Add(1)
						assert.Equal(t, http.MethodPost, r.Method)
						if identity == "terraform" {
							assert.Equal(t, "Delegated "+token, r.Header.Get("Authorization"))
						} else {
							assert.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Delegated "))
							assert.Len(t, strings.Split(strings.TrimPrefix(r.Header.Get("Authorization"), "Delegated "), "|"), 4)
						}
						p := phase.Load()
						p.entered <- struct{}{}
						<-p.release
						if p.reject {
							w.WriteHeader(http.StatusUnauthorized)
							_, _ = w.Write([]byte(`{"errors":["exchange rejected"]}`))
							return
						}
						_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"attributes": map[string]interface{}{
							"access_token": fmt.Sprintf("delegated-%d", number),
							"expires":      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
						}}})
					case "/api/v2/roles":
						requests.Add(1)
						assert.Equal(t, fmt.Sprintf("Bearer delegated-%d", exchanges.Load()), r.Header.Get("Authorization"))
						assert.Empty(t, r.Header.Get("DD-API-KEY"))
						assert.Empty(t, r.Header.Get("DD-APPLICATION-KEY"))
						_, _ = w.Write([]byte(`{"data":[]}`))
					default:
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer server.Close()
				client, auth, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
					"validate": "false", "api_url": server.URL,
					"cloud_provider_type": "aws", "cloud_provider_region": "us-east-1",
					"org_uuid":          workloadIdentityTestOrg,
					"aws_access_key_id": "test-aws-id", "aws_secret_access_key": "test-aws-secret",
					"aws_session_token": "test-aws-session",
					"bearer_token":      "test-bearer", "api_key": "test-api-key", "app_key": "test-app-key",
				})
				require.NoError(t, err)
				require.Equal(t, identity, client.Cfg.DelegatedTokenConfig.Provider)
				roles := datadogV2.NewRolesApi(client)
				shared := auth.Value(api.ContextDelegatedToken).(*api.DelegatedTokenCredentials)
				const parallel = 16
				for _, stage := range []string{"first use", "cached", "refresh", "failure", "recovery"} {
					t.Run(stage, func(t *testing.T) {
						if stage == "refresh" || stage == "failure" {
							// The previous batch is finished; no concurrent field access.
							shared.Expiration = time.Now().Add(-time.Hour)
						}
						p := &exchangePhase{make(chan struct{}, parallel), make(chan struct{}), stage == "failure"}
						phase.Store(p)
						unblock := sync.OnceFunc(func() { close(p.release) })
						defer unblock()
						beforeExchanges, beforeRequests := exchanges.Load(), requests.Load()
						results := make(chan error, parallel)
						request := func(ctx context.Context) {
							_, _, err := roles.ListRoles(ctx)
							results <- err
						}
						go request(auth)
						waiting := make(chan struct{}, parallel)
						if stage != "cached" {
							awaitWIFRequest(t, p.entered)
						}
						for i := 1; i < parallel; i++ {
							go request(&wifWaitingContext{Context: auth, waiting: waiting})
						}
						if stage != "cached" {
							for i := 1; i < parallel; i++ {
								awaitWIFRequest(t, waiting)
							}
						}
						unblock()
						for i := 0; i < parallel; i++ {
							err := awaitWIFRequest(t, results)
							if p.reject {
								require.ErrorContains(t, err, "401")
							} else {
								require.NoError(t, err)
							}
						}
						if stage == "cached" {
							require.Equal(t, beforeExchanges, exchanges.Load())
						} else {
							require.Equal(t, beforeExchanges+1, exchanges.Load(), "parallel requests must share one exchange")
						}
						if p.reject {
							require.Equal(t, beforeRequests, requests.Load(), "failed exchange must not send API requests or fall back to another identity")
						} else {
							require.Equal(t, beforeRequests+parallel, requests.Load())
						}
					})
				}
			})
		}
	}
}
