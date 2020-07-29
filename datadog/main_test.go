package datadog

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// TestMain starts the tracer.
func TestMain(m *testing.M) {
	fmt.Println("Hello world")
	service, ok := os.LookupEnv("DD_SERVICE")
	if !ok {
		service = "terraform-datadog-provider"
	}
	tracer.Start(
		tracer.WithService(service),
		// tracer.WithServiceVersion(version.ProviderVersion),
	)
	code := m.Run()
	tracer.Stop()
	fmt.Println("Bye world")
	os.Exit(code)
}
