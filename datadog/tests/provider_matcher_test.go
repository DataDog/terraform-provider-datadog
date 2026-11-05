package test

import "testing"

func TestJSONEquivalentWidgetIDs(t *testing.T) {
	cases := []struct {
		name     string
		request  string
		cassette string
		want     bool
	}{
		{
			name:     "widget id not recorded by v1 cassette",
			request:  `{"widgets":[{"definition":{"type":"note"},"id":123}]}`,
			cassette: `{"widgets":[{"definition":{"type":"note"}}]}`,
			want:     true,
		},
		{
			name:     "group child widget id not recorded",
			request:  `{"widgets":[{"definition":{"type":"group","widgets":[{"definition":{"type":"note"},"id":456}]},"id":123}]}`,
			cassette: `{"widgets":[{"definition":{"type":"group","widgets":[{"definition":{"type":"note"}}]}}]}`,
			want:     true,
		},
		{
			name:     "recorded widget id must match",
			request:  `{"widgets":[{"definition":{"type":"note"},"id":123}]}`,
			cassette: `{"widgets":[{"definition":{"type":"note"},"id":999}]}`,
			want:     false,
		},
		{
			name:     "non-widget id is still compared",
			request:  `{"id":123,"widgets":[]}`,
			cassette: `{"widgets":[]}`,
			want:     false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := jsonEquivalent(tc.request, tc.cassette); got != tc.want {
				t.Errorf("jsonEquivalent() = %v, want %v", got, tc.want)
			}
		})
	}
}
