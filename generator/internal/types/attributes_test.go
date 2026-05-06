package types

import (
	"testing"
)

func TestResolveAttributeMode(t *testing.T) {
	tests := []struct {
		name         string
		isPathParam  bool
		isQueryParam bool
		want         AttributeMode
	}{
		{
			name:        "path param -> Required",
			isPathParam: true,
			want:        Required,
		},
		{
			name:         "query param -> Optional",
			isQueryParam: true,
			want:         Optional,
		},
		{
			name: "response field -> Computed",
			want: Computed,
		},
		{
			name:         "path param takes precedence over query",
			isPathParam:  true,
			isQueryParam: true,
			want:         Required,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveAttributeMode(tt.isPathParam, tt.isQueryParam)
			if got != tt.want {
				t.Errorf("ResolveAttributeMode() = %d, want %d", got, tt.want)
			}
		})
	}
}
