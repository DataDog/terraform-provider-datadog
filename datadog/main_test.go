package datadog

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/meta"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

// TestMain starts the tracer.
func TestMain(m *testing.M) {
	acctest.UseBinaryDriver("datadog", Provider)
	if _, ok := os.LookupEnv("DD_AGENT_HOST"); !ok {
		log.Println("DD_AGENT_HOST is not configured. Tests are executed without tracer and profiler.")
		code := m.Run()
		os.Exit(code)
	}

	service, ok := os.LookupEnv("DD_SERVICE")
	if !ok {
		service = "terraform-datadog-provider"
	}
	tracer.Start(
		tracer.WithService(service),
		// tracer.WithServiceVersion(version.ProviderVersion),
	)
	defer tracer.Stop()

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
	code := m.Run()
	os.Exit(code)
}
