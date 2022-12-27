package utils

import (
	"unicode"
	"unicode/utf8"
)

// NormalizeTag takes a string and parses it in accordance to USM Tagging conventions
func NormalizeTag(tag string) string {
	// Fast path: Check if the tag is valid and only contains ASCII characters,
	// if yes return it as-is right away.
	MaxTagLength := 200

	if isNormalizedASCIITag(tag, MaxTagLength) {
		return tag
	}

	bufSize := len(tag)
	if bufSize > MaxTagLength {
		bufSize = MaxTagLength // Limit size of allocation
	}
	buf := make([]byte, 0, bufSize+utf8.UTFMax-1)

	lastWasUnderscore := false
	for _, c := range tag {
		if len(buf) >= MaxTagLength {
			break
		}

		switch {
		// fast path for ascii alphabet
		case c >= 'a' && c <= 'z':
			buf = append(buf, byte(c))
			lastWasUnderscore = false
		case c >= 'A' && c <= 'Z':
			buf = append(buf, byte(c+('a'-'A')))
			lastWasUnderscore = false
		// ':' may be the first character of a tag
		case c == ':':
			buf = append(buf, byte(c))
			lastWasUnderscore = false
		case unicode.IsLetter(c):
			buf = utf8.AppendRune(buf, unicode.ToLower(c))
			lastWasUnderscore = false
		// skip any code points that can't start the string
		// (we only reach this case if we've not seen a valid starting code point yet)
		case len(buf) == 0:
		// handle valid code points that can't start the string.
		case c == '.' || c == '/' || c == '-':
			buf = append(buf, byte(c))
			lastWasUnderscore = false
		case unicode.IsDigit(c):
			buf = utf8.AppendRune(buf, c)
			lastWasUnderscore = false
		// convert anything else to underscores (including underscores), but only allow one in a row.
		case !lastWasUnderscore:
			buf = append(buf, '_')
			lastWasUnderscore = true
		}
	}

	if lastWasUnderscore {
		buf = buf[:len(buf)-1]
	}
	return string(buf)
}

func isNormalizedASCIITag(tag string, maxTagLength int) bool {
	if len(tag) == 0 {
		return true
	}
	if len(tag) > maxTagLength {
		return false
	}
	if !isValidASCIIStartChar(tag[0]) {
		return false
	}
	for i := 1; i < len(tag); i++ {
		b := tag[i]
		// TODO: Attempt to optimize this check using SIMD/vectorization.
		if isValidASCIITagChar(b) {
			// okay!
		} else if b == '_' {
			// an underscore is only okay if followed by a valid non-underscore character
			i++
			if i == len(tag) || !isValidASCIITagChar(tag[i]) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func isValidASCIIStartChar(c byte) bool {
	return ('a' <= c && c <= 'z') || c == ':'
}

func isValidASCIITagChar(c byte) bool {
	return isValidASCIIStartChar(c) || ('0' <= c && c <= '9') || c == '.' || c == '/' || c == '-'
}
