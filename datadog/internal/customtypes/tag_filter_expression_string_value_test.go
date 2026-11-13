package customtypes

import "testing"

func TestTagFilterExpressionsEqual(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"env:production AND !region:us-east-1", "NOT region:us-east-1 AND env:production", true},
		{"env:production AND !region:us-east-1", "env:production AND NOT region:us-east-1", true},
		{"team:backend", "team:backend", true},
		{"env:prod and team:a", "team:a AND env:prod", true},
		{"env:production AND !region:us-east-1", "env:production AND region:us-east-1", false},
		{"env:production AND !region:us-east-1", "NOT env:production AND region:us-east-1", false},
		{"a:1 AND b:2", "a:1 AND b:2 AND c:3", false},
		{"a:1 AND a:1", "a:1", false},
		{"a:1 OR b:2", "a:1 OR b:2", true},
		{"a:1 OR b:2", "b:2 OR a:1", false},
		{"!(a:1 OR b:2)", "NOT ( a:1 OR b:2 )", true},
		{"(a:1 AND b:2) OR c:3", "a:1 AND (b:2 OR c:3)", false},
	}
	for _, c := range cases {
		if got := TagFilterExpressionsEqual(c.a, c.b); got != c.want {
			t.Errorf("TagFilterExpressionsEqual(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
