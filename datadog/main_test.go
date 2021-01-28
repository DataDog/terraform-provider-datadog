package datadog

import (
	"log"
	"os"
	"testing"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

// TestMain starts the tracer.
func TestMain(m *testing.M) {
	service, ok := os.LookupEnv("DD_SERVICE")
	if !ok {
		service = "terraform-datadog-provider"
	}
	tracer.Start(
		tracer.WithService(service),
		// tracer.WithServiceVersion(version.ProviderVersion),
	)
	defer tracer.Stop()

	err := profiler.Start(
		profiler.WithService(service),
		profiler.WithProfileTypes(profiler.BlockProfile, profiler.CPUProfile, profiler.GoroutineProfile, profiler.HeapProfile),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer profiler.Stop()

	code := m.Run()
	os.Exit(code)
}
