package datadog

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/jonboulle/clockwork"
	"github.com/terraform-providers/terraform-provider-datadog/version"
	datadog "github.com/zorkian/go-datadog-api"
)

func isRecording() bool {
	return os.Getenv("RECORD") == "true"
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
	}
	return restoreClock(t)
}

func removeURLSecrets(u *url.URL) *url.URL {
	query := u.Query()
	query.Del("api_key")
	query.Del("application_key")
	u.RawQuery = query.Encode()
	return u
}

func initAccProvider(t *testing.T) (*schema.Provider, func(t *testing.T)) {
	var mode recorder.Mode
	if isRecording() {
		mode = recorder.ModeRecording
	} else {
		mode = recorder.ModeReplaying
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

	p := Provider().(*schema.Provider)
	p.ConfigureFunc = testProviderConfigure(rec)

	cleanup := func(t *testing.T) {
		rec.Stop()
	}
	return p, cleanup
}

func testProviderConfigure(r *recorder.Recorder) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {

		client := datadog.NewClient(d.Get("api_key").(string), d.Get("app_key").(string))
		if apiURL := d.Get("api_url").(string); apiURL != "" {
			client.SetBaseUrl(apiURL)
		}

		c := cleanhttp.DefaultClient()
		c.Transport = logging.NewTransport("Datadog", r)
		client.HttpClient = c
		client.ExtraHeader["User-Agent"] = fmt.Sprintf("Datadog/%s/terraform (%s)", version.ProviderVersion, runtime.Version())

		if os.Getenv("RECORD") != "true" {
			return client, nil
		}

		// Do not validate API and APP keys when we are not recording
		log.Println("[INFO] Datadog client successfully initialized, now validating...")
		ok, err := client.Validate()
		if err != nil {
			log.Printf("[ERROR] Datadog Client validation error: %v", err)
			return client, err
		} else if !ok {
			err := errors.New(`Invalid or missing credentials provided to the Datadog Provider. Please confirm your API and APP keys are valid and see https://terraform.io/docs/providers/datadog/index.html for more information on providing credentials for the Datadog Provider`)
			log.Printf("[ERROR] Datadog Client validation error: %v", err)
			return client, err
		}
		log.Printf("[INFO] Datadog Client successfully validated.")

		return client, nil
	}
}

func testAccProviders(t *testing.T) (map[string]terraform.ResourceProvider, func(t *testing.T)) {
	provider, cleanup := initAccProvider(t)
	return map[string]terraform.ResourceProvider{
		"datadog": provider,
	}, cleanup
}

func testAccProvider(t *testing.T, accProviders map[string]terraform.ResourceProvider) *schema.Provider {
	accProvider, ok := accProviders["datadog"]
	if !ok {
		t.Fatal("could not find datadog provider")
	}
	return accProvider.(*schema.Provider)
}

func TestProvider(t *testing.T) {
	accProvider, cleanup := initAccProvider(t)
	defer cleanup(t)

	if err := accProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DATADOG_API_KEY"); v == "" {
		t.Fatal("DATADOG_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("DATADOG_APP_KEY"); v == "" {
		t.Fatal("DATADOG_APP_KEY must be set for acceptance tests")
	}
}
