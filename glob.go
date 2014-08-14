package ohmyglob

import (
	"bytes"
	"regexp"
	"strings"
)

const globStarComponent = ".*"

type parserState struct {
	// Filled during parsing, before the regular expression is compiled
	regexBuffer *bytes.Buffer
	// Set to true if the last parsed component was a globstar
	lastComponentWasGlobStar bool
	options                  *GlobOptions
	// The regex-escaped separator character
	escapedSeparator string
}

// Glob is a glob pattern that has been compiled into a regular expression
type Glob struct {
	*regexp.Regexp
	// The separator from options, escaped for appending to the regexBuffer (only available during parsing)
	// The input pattern
	globPattern string
	// State only available during parsing
	parserState *parserState
}

// GlobOptions
type GlobOptions struct {
	// The character used to split path components
	Separator rune
	// Set to false to allow any prefix before the glob match
	MatchAtStart bool
	// Set to false to allow any suffix after the glob match
	MatchAtEnd bool
}

var DefaultGlobOptions *GlobOptions = &GlobOptions{
	Separator:    '.',
	MatchAtStart: true,
	MatchAtEnd:   true,
}

// String returns the string that was used to create the Glob.
func (g *Glob) String() string {
	return g.globPattern
}

func NewGlob(pattern string, options *GlobOptions) (*Glob, error) {
	pattern = strings.TrimSpace(pattern)
	if options == nil {
		options = DefaultGlobOptions
	}

	glob := &Glob{
		Regexp:      nil,
		globPattern: pattern,
		parserState: &parserState{
			regexBuffer:              new(bytes.Buffer),
			lastComponentWasGlobStar: false,
			options:                  options,
			escapedSeparator:         escapeRegexComponent(string(options.Separator)),
		},
	}

	if options.MatchAtStart {
		glob.parserState.regexBuffer.WriteRune('^')
	}

	// Transform into a regular expression pattern
	// 1. Parse negation prefixes

	// 2. Split into a series of path portion matches
	components := strings.Split(pattern, string(options.Separator))
	for _, component := range components {
		err := parseComponent(component, glob)
		if err != nil {
			return nil, err
		}
	}

	if options.MatchAtEnd {
		glob.parserState.regexBuffer.WriteRune('$')
	}

	re, err := regexp.Compile(glob.parserState.regexBuffer.String())
	if err != nil {
		return nil, err
	}

	glob.parserState = nil
	glob.Regexp = re

	return glob, nil
}

func parseComponent(component string, glob *Glob) error {
	if component == "**" {
		if glob.parserState.lastComponentWasGlobStar {
			return nil
		} else {
			glob.parserState.regexBuffer.WriteString(globStarComponent)
			glob.parserState.lastComponentWasGlobStar = true
			return nil
		}
	} else {
		glob.parserState.regexBuffer.WriteString(escapeRegexComponent(component))
		glob.parserState.lastComponentWasGlobStar = false
		return nil
	}
}
