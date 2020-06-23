package datadog

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/jonboulle/clockwork"
	"github.com/terraform-providers/terraform-provider-datadog/version"
	datadogCommunity "github.com/zorkian/go-datadog-api"
)

func isRecording() bool {
	return os.Getenv("RECORD") == "true"
}

func isReplaying() bool {
	return os.Getenv("RECORD") == "false"
}

func isDebug() bool {
	return os.Getenv("DEBUG") == "true"
}

func isAPIKeySet() bool {
	if os.Getenv("DATADOG_API_KEY") != "" {
		return true
	}
	if os.Getenv("DD_API_KEY") != "" {
		return true
	}
	return false
}

func isAPPKeySet() bool {
	if os.Getenv("DATADOG_APP_KEY") != "" {
		return true
	}
	if os.Getenv("DD_APP_KEY") != "" {
		return true
	}
	return false
}

func setClock(t *testing.T) clockwork.FakeClock {
	os.MkdirAll("cassettes", 0755)
	f, err := os.Create(fmt.Sprintf("cassettes/%s.freeze", t.Name()))
	if err != nil {
		t.Fatalf("Could not set clock: %v", err)
	}
	defer f.Close()
	now := clockwork.NewRealClock().Now()
	f.WriteString(now.Format(time.RFC3339Nano))
	return clockwork.NewFakeClockAt(now)
}

func restoreClock(t *testing.T) clockwork.FakeClock {
	data, err := ioutil.ReadFile(fmt.Sprintf("cassettes/%s.freeze", t.Name()))
	if err != nil {
		t.Logf("Could not load clock: %v", err)
		return setClock(t)
	}
	now, err := time.Parse(time.RFC3339Nano, string(data))
	if err != nil {
		t.Fatalf("Could not parse clock date: %v", err)
	}
	return clockwork.NewFakeClockAt(now)
}

func testClock(t *testing.T) clockwork.FakeClock {
	if isRecording() {
		return setClock(t)
	} else if isReplaying() {
		return restoreClock(t)
	}
	// do not set or restore frozen time
	return clockwork.NewFakeClockAt(clockwork.NewRealClock().Now())
}

func removeURLSecrets(u *url.URL) *url.URL {
	query := u.Query()
	query.Del("api_key")
	query.Del("application_key")
	u.RawQuery = query.Encode()
	return u
}

func initRecorder(t *testing.T) *recorder.Recorder {
	var mode recorder.Mode
	if isRecording() {
		mode = recorder.ModeRecording
	} else if isReplaying() {
		mode = recorder.ModeReplaying
	} else {
		mode = recorder.ModeDisabled
	}

	rec, err := recorder.NewAsMode(fmt.Sprintf("cassettes/%s", t.Name()), mode, nil)
	if err != nil {
		log.Fatal(err)
	}

	rec.SetMatcher(func(r *http.Request, i cassette.Request) bool {
		return r.Method == i.Method && removeURLSecrets(r.URL).String() == i.URL
	})

	rec.AddFilter(func(i *cassette.Interaction) error {
		u, err := url.Parse(i.URL)
		if err != nil {
			return err
		}
		i.URL = removeURLSecrets(u).String()
		i.Request.Headers.Del("Dd-Api-Key")
		i.Request.Headers.Del("Dd-Application-Key")
		return nil
	})
	return rec
}

func initAccProvider(t *testing.T, httpClient *http.Client) *schema.Provider {

	p := Provider().(*schema.Provider)
	p.ConfigureFunc = testProviderConfigure(httpClient)

	return p
}

func buildAuthV1(apiKey string, appKey string, apiURL string) (context.Context, error) {
	authV1 := context.WithValue(
		context.Background(),
		datadogV1.ContextAPIKeys,
		map[string]datadogV1.APIKey{
			"apiKeyAuth": datadogV1.APIKey{
				Key: apiKey,
			},
			"appKeyAuth": datadogV1.APIKey{
				Key: appKey,
			},
		},
	)
	if apiURL != "" {
		parsedApiUrl, parseErr := url.Parse(apiURL)
		if parseErr != nil {
			return nil, fmt.Errorf(`invalid API Url : %v`, parseErr)
		}
		if parsedApiUrl.Host == "" || parsedApiUrl.Scheme == "" {
			return nil, fmt.Errorf(`missing protocol or host : %v`, apiURL)
		}
		// If api url is passed, set and use the api name and protocol on ServerIndex{1}
		authV1 = context.WithValue(authV1, datadogV1.ContextServerIndex, 1)
		authV1 = context.WithValue(authV1, datadogV1.ContextServerVariables, map[string]string{
			"name":     parsedApiUrl.Host,
			"protocol": parsedApiUrl.Scheme,
		})
	}
	return authV1, nil
}

func buildDatadogClientV1(httpClient *http.Client) *datadogV1.APIClient {
	//Datadog V1 API config.HTTPClient
	configV1 := datadogV1.NewConfiguration()
	configV1.Debug = isDebug()
	configV1.HTTPClient = httpClient
	return datadogV1.NewAPIClient(configV1)
}

func testProviderConfigure(httpClient *http.Client) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		communityClient := datadogCommunity.NewClient(d.Get("api_key").(string), d.Get("app_key").(string))
		if apiURL := d.Get("api_url").(string); apiURL != "" {
			communityClient.SetBaseUrl(apiURL)
		}

		c := httpClient
		communityClient.HttpClient = c
		communityClient.ExtraHeader["User-Agent"] = fmt.Sprintf("Datadog/%s/terraform (%s)", version.ProviderVersion, runtime.Version())

		authV1, err := buildAuthV1(d.Get("api_key").(string), d.Get("app_key").(string), d.Get("api_url").(string))
		if err != nil {
			return nil, err
		}
		datadogClientV1 := buildDatadogClientV1(httpClient)

		// Initialize the official datadog v2 API client
		authV2 := context.WithValue(
			context.Background(),
			datadogV2.ContextAPIKeys,
			map[string]datadogV2.APIKey{
				"apiKeyAuth": datadogV2.APIKey{
					Key: d.Get("api_key").(string),
				},
				"appKeyAuth": datadogV2.APIKey{
					Key: d.Get("app_key").(string),
				},
			},
		)
		//Datadog V2 API config.HTTPClient
		configV2 := datadogV2.NewConfiguration()
		configV2.Debug = isDebug()
		configV2.HTTPClient = c
		if apiURL := d.Get("api_url").(string); apiURL != "" {
			parsedApiUrl, parseErr := url.Parse(apiURL)
			if parseErr != nil {
				return nil, fmt.Errorf(`invalid API Url : %v`, parseErr)
			}
			if parsedApiUrl.Host == "" || parsedApiUrl.Scheme == "" {
				return nil, fmt.Errorf(`missing protocol or host : %v`, apiURL)
			}
			// If api url is passed, set and use the api name and protocol on ServerIndex{1}
			authV2 = context.WithValue(authV2, datadogV2.ContextServerIndex, 1)
			authV2 = context.WithValue(authV2, datadogV2.ContextServerVariables, map[string]string{
				"name":     parsedApiUrl.Host,
				"protocol": parsedApiUrl.Scheme,
			})
		}
		datadogClientV2 := datadogV2.NewAPIClient(configV2)

		return &ProviderConfiguration{
			CommunityClient: communityClient,
			DatadogClientV1: datadogClientV1,
			DatadogClientV2: datadogClientV2,
			AuthV1:          authV1,
			AuthV2:          authV2,
		}, nil
	}
}

func testAccProvidersWithHttpClient(t *testing.T, httpClient *http.Client) map[string]terraform.ResourceProvider {
	provider := initAccProvider(t, httpClient)
	return map[string]terraform.ResourceProvider{
		"datadog": provider,
	}
}

func testAccProviders(t *testing.T, rec *recorder.Recorder) (map[string]terraform.ResourceProvider, func(t *testing.T)) {
	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewTransport("Datadog", rec)
	return testAccProvidersWithHttpClient(t, c), func(t *testing.T) {
		rec.Stop()
	}
}

func testAccProvider(t *testing.T, accProviders map[string]terraform.ResourceProvider) *schema.Provider {
	accProvider, ok := accProviders["datadog"]
	if !ok {
		t.Fatal("could not find datadog provider")
	}
	return accProvider.(*schema.Provider)
}

func TestProvider(t *testing.T) {
	rec := initRecorder(t)
	defer rec.Stop()
	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewTransport("Datadog", rec)
	accProvider := initAccProvider(t, c)

	if err := accProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if isReplaying() {
		return
	}
	if !isAPIKeySet() {
		t.Fatal("DD_API_KEY must be set for acceptance tests")
	}
	if !isAPPKeySet() {
		t.Fatal("DD_APP_KEY must be set for acceptance tests")
	}
}

func testCheckResourceAttrs(name string, checkExists resource.TestCheckFunc, assertions []string) []resource.TestCheckFunc {
	funcs := []resource.TestCheckFunc{}
	funcs = append(funcs, checkExists)
	for _, assertion := range assertions {
		assertionPair := strings.Split(assertion, " = ")
		if len(assertionPair) == 1 {
			assertionPair = strings.Split(assertion, " =")
		}
		key := assertionPair[0]
		value := ""
		if len(assertionPair) > 1 {
			value = assertionPair[1]
		}
		funcs = append(funcs, resource.TestCheckResourceAttr(name, key, value))
		// Use utility method below, instead of the above one, to print out all state keys/values during test debugging
		//funcs = append(funcs, CheckResourceAttr(name, key, value))
	}
	return funcs
}

/* Utility method for Debugging purpose. This method helps list assertions as well
It is a duplication of `resource.TestCheckResourceAttr` into which we added print statements.
*/
func CheckResourceAttr(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return nil
		}

		is := rs.Primary
		if is == nil {
			return nil
		}

		for k, val := range is.Attributes {
			fmt.Println(fmt.Sprintf("%v = %v", k, val))
		}

		// Empty containers may be elided from the state.
		// If the intent here is to check for an empty container, allow the key to
		// also be non-existent.
		emptyCheck := value == "0" && (strings.HasSuffix(key, ".#") || strings.HasSuffix(key, ".%"))

		if v, ok := is.Attributes[key]; !ok || v != value {

			if emptyCheck && !ok {
				return nil
			}

			if !ok {
				return fmt.Errorf("%s: Attribute '%s' not found", name, key)
			}

			return fmt.Errorf(
				"%s: Attribute '%s' expected %#v, got %#v",
				name,
				key,
				value,
				v)
		}
		return nil
	}
}
