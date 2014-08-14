package ohmyglob

import (
	"regexp"
)

var escapedComponentRegex = regexp.MustCompile(`[-\/\\^$*+?.()|[\]{}]`)

// Escapes any characters that would have special meaning in a regular expression, returning the escaped string
func escapeRegexComponent(str string) string {
	return escapedComponentRegex.ReplaceAllString(str, "\\$0")
}
