package indexer

import (
	"regexp"
	"strings"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func Tokenize(text string) []string {
	lower := strings.ToLower(text)
	alphanumeric := nonAlphanumericRegex.ReplaceAllString(lower, "")
	tokens := strings.Fields(alphanumeric)

	if tokens == nil {
		return []string{}
	}
	return tokens
}
