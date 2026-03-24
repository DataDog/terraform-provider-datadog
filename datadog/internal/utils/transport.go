package utils

import (
	"context"
	"net/http"
)

type contextKeyTerraformResource struct{}

// WithTerraformResource returns a copy of ctx annotated with the Terraform
// resource name. The returned context is intended to be passed as the "auth"
// context through the Datadog SDK, which attaches it to every outgoing
// http.Request. The resourceHeaderTransport installed at provider init reads
// the value back and sets the DD-Terraform-Resource header.
func WithTerraformResource(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextKeyTerraformResource{}, name)
}

// GetTerraformResource extracts the resource name previously stored by
// WithTerraformResource.
func GetTerraformResource(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextKeyTerraformResource{}).(string)
	return v, ok && v != ""
}

// resourceHeaderTransport is an http.RoundTripper that injects the
// DD-Terraform-Resource header when the request context carries a resource
// name.
type resourceHeaderTransport struct {
	base http.RoundTripper
}

func (t *resourceHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if name, ok := GetTerraformResource(req.Context()); ok {
		req = req.Clone(req.Context())
		req.Header.Set("DD-Terraform-Resource", name)
	}
	return t.base.RoundTrip(req)
}

// WrapTransportWithResourceHeader wraps base so that every request whose
// context carries a Terraform resource name (set via WithTerraformResource)
// gets a DD-Terraform-Resource header.
func WrapTransportWithResourceHeader(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &resourceHeaderTransport{base: base}
}
