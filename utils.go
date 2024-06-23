package prue

import (
	"strings"
	"unicode"
)

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func normalizePath(pattern string) string {
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	pattern = strings.ReplaceAll(pattern, "//", "/")
	if len(pattern) > 1 && strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/")
	}
	return pattern
}
