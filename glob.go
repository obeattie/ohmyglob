package ohmyglob

import (
	"bytes"
	"io"
	"regexp"
	"strings"

	log "github.com/cihub/seelog"
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

type Glob interface {
	// String returns the pattern that was used to create the Glob
	String() string
	// Match reports whether the Regexp matches the byte slice b
	Match(b []byte) bool
	// MatchReader reports whether the Regexp matches the text read by the RuneReader
	MatchReader(r io.RuneReader) bool
	// MatchString reports whether the Regexp matches the string s
	MatchString(s string) bool
}

// Glob is a glob pattern that has been compiled into a regular expression
type globImpl struct {
	*regexp.Regexp
	// The separator from options, escaped for appending to the regexBuffer (only available during parsing)
	// The input pattern
	globPattern string
	// State only available during parsing
	parserState *parserState
	// Set to true if the pattern was negated
	negated bool
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
	Separator:    '/',
	MatchAtStart: true,
	MatchAtEnd:   true,
}

func (g *globImpl) String() string {
	return g.globPattern
}

func (g *globImpl) Match(b []byte) bool {
	if g.negated {
		return !g.Regexp.Match(b)
	} else {
		return g.Regexp.Match(b)
	}
}

func (g *globImpl) MatchReader(r io.RuneReader) bool {
	if g.negated {
		return !g.Regexp.MatchReader(r)
	} else {
		return g.Regexp.MatchReader(r)
	}
}

func (g *globImpl) MatchString(s string) bool {
	if g.negated {
		return !g.Regexp.MatchString(s)
	} else {
		return g.Regexp.MatchString(s)
	}
}

func NewGlob(pattern string, options *GlobOptions) (Glob, error) {
	pattern = strings.TrimSpace(pattern)
	if options == nil {
		options = DefaultGlobOptions
	}

	glob := &globImpl{
		Regexp:      nil,
		globPattern: pattern,
		negated:     false,
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

	var err error
	// Transform into a regular expression pattern
	// 1. Parse negation prefixes
	pattern, err = parseNegation(pattern, glob)
	if err != nil {
		return nil, err
	}

	// 2. Split into a series of path portion matches
	components := strings.Split(pattern, string(options.Separator))
	for idx, component := range components {
		err = parseComponent(component, idx, glob)
		if err != nil {
			return nil, err
		}
	}

	if options.MatchAtEnd {
		glob.parserState.regexBuffer.WriteRune('$')
	}

	regexString := glob.parserState.regexBuffer.String()
	log.Debugf("Compiled \"%s\" to regex `%s` (negated: %v)", pattern, regexString, glob.negated)
	re, err := regexp.Compile(regexString)
	if err != nil {
		return nil, err
	}

	glob.parserState = nil
	glob.Regexp = re

	return glob, nil
}

func parseNegation(pattern string, glob *globImpl) (string, error) {
	if pattern == "" {
		return pattern, nil
	}

	negations := 0
	for _, char := range pattern {
		if char == '!' {
			glob.negated = !glob.negated
			negations++
		}
	}

	return pattern[negations:], nil
}

func parseComponent(component string, idx int, glob *globImpl) error {
	isGlobStar := false
	buf := glob.parserState.regexBuffer

	if component == "**" {
		isGlobStar = true
		// Only add another globstar if the last component wasn't a globstar
		if !glob.parserState.lastComponentWasGlobStar {
			buf.WriteString(globStarComponent)
		}
	} else if component == "*" {
		if idx != 0 {
			buf.WriteString(glob.parserState.escapedSeparator)
		}
		buf.WriteString("[^")
		buf.WriteString(glob.parserState.escapedSeparator)
		buf.WriteString("]*")
	} else {
		if idx != 0 {
			buf.WriteString(glob.parserState.escapedSeparator)
		}
		buf.WriteString(escapeRegexComponent(component))
	}

	glob.parserState.lastComponentWasGlobStar = isGlobStar
	return nil
}
