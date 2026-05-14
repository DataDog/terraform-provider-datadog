package utils

import (
	"context"
	"net/http"
	"strings"
)

type contextKeyTerraformResource struct{}

// WithTerraformResource returns a copy of ctx annotated with the Terraform
// resource name.
func WithTerraformResource(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextKeyTerraformResource{}, name)
}

// GetTerraformResource extracts the resource name previously stored by
// WithTerraformResource.
func GetTerraformResource(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextKeyTerraformResource{}).(string)
	return v, ok && v != ""
}

// resourceUserAgentTransport is an http.RoundTripper that appends a
// "terraform_resource" comment to the User-Agent header when the request
// context carries a resource name.
type resourceUserAgentTransport struct {
	base http.RoundTripper
}

func (t *resourceUserAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if name, ok := GetTerraformResource(req.Context()); ok {
		req = req.Clone(req.Context())
		ua := req.Header.Get("User-Agent")
		if i := strings.Index(ua, ")"); i >= 0 {
			ua = ua[:i] + "; terraform_resource " + name + ua[i:]
			req.Header.Set("User-Agent", ua)
		}
	}
	return t.base.RoundTrip(req)
}

// WrapTransportWithResourceUserAgent wraps base so that every request whose
// context carries a Terraform resource name gets the resource type appended
// to the User-Agent comment section.
func WrapTransportWithResourceUserAgent(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &resourceUserAgentTransport{base: base}
}
