package test

import (
	"fmt"
	"log"
	"os"
	"testing"

	ddtesting "github.com/DataDog/dd-sdk-go-testing"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

// TestMain starts the tracer.
func TestMain(m *testing.M) {
	// Enable env-gated resources for the test package. These resources are
	// registered conditionally in framework_provider.go and would otherwise
	// be invisible to tests in the suite.
	os.Setenv("DD_TERRAFORM_DATABRICKS_INTEGRATION_ENABLED", "true")

	if _, ok := os.LookupEnv("DD_AGENT_HOST"); !ok {
		log.Println("DD_AGENT_HOST is not configured. Tests are executed without tracer and profiler.")
		code := m.Run()
		os.Exit(code)
	}

	service, ok := os.LookupEnv("DD_SERVICE")
	if !ok {
		service = "terraform-datadog-provider"
	}

	profilerOpts := []profiler.Option{
		profiler.WithService(service),
		profiler.WithTags(
			fmt.Sprintf("terraform.sdk:%s", meta.SDKVersionString()),
			// fmt.Sprintf("terraform.cli:%s", datadogProvider.TerraformVersion),
		),
		profiler.WithProfileTypes(profiler.BlockProfile, profiler.CPUProfile, profiler.GoroutineProfile, profiler.HeapProfile),
	}

	if v := os.Getenv("DD_PROFILER_API_KEY"); v != "" {
		profilerOpts = append(profilerOpts, profiler.WithAPIKey(v))
	}

	err := profiler.Start(profilerOpts...)
	if err != nil {
		log.Fatal(err)
	} else {
		defer profiler.Stop()
	}

	code := ddtesting.Run(m)
	os.Exit(code)
}
