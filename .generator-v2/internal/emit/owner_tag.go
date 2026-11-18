package emit

import (
	"regexp"
	"strconv"
)

// ownerTagRe matches the owner-tag marker the data-source templates write
// below the generated-code marker.
var ownerTagRe = regexp.MustCompile(`(?m)^// tfgen:owner-tag ("(?:[^"\\]|\\.)*")$`)

// GeneratedOwnerTag returns the x-datadog-tf-generator tag recorded in a
// generated data source's source, and whether one was recorded.
func GeneratedOwnerTag(src []byte) (string, bool) {
	m := ownerTagRe.FindSubmatch(src)
	if m == nil {
		return "", false
	}
	tag, err := strconv.Unquote(string(m[1]))
	if err != nil {
		return "", false
	}
	return tag, true
}
