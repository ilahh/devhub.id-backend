package utils

import (
	"regexp"
	"strings"
)

var nonAlnumRegex = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlnumRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "post"
	}
	if len(s) > 80 {
		s = strings.Trim(s[:80], "-")
	}
	return s
}
