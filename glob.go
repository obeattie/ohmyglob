// Package ohmyglob provides a minimal glob matching utility.
package ohmyglob

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	log "github.com/cihub/seelog"
)

const globStarComponent = ".+"

var (
	// Logger is used to log trace-level info; logging is completely disabled by default but can be changed by replacing
	// this with a configured logger
	Logger    log.LoggerInterface
	expanders = []rune{'?', '*'}
)

func init() {
	if Logger == nil {
		var err error
		Logger, err = log.LoggerFromWriterWithMinLevel(os.Stderr, log.CriticalLvl) // seelog bug means we can't use log.Off
		if err != nil {
			panic(err)
		}
	}
}

type parserState struct {
	// Filled during parsing, before the regular expression is compiled
	regexBuffer *bytes.Buffer
	// Set to true if the last parsed component was a globstar
	lastComponentWasGlobStar bool
	options                  *Options
	// The regex-escaped separator character
	escapedSeparator string
}

// GlobMatcher is the basic interface provided by a Glob or GlobSet, which provides a Regexp-style interface for
// checking matches.
type GlobMatcher interface {
	// Match reports whether the Glob matches the byte slice b
	Match(b []byte) bool
	// MatchReader reports whether the Glob matches the text read by the RuneReader
	MatchReader(r io.RuneReader) bool
	// MatchString reports whether the Glob matches the string s
	MatchString(s string) bool
}

// Glob is a single glob pattern; implements GlobMatcher. A Glob is immutable.
type Glob interface {
	GlobMatcher
	// String returns the pattern that was used to create the Glob
	String() string
	// IsNegative returns whether the pattern was negated (prefixed with !)
	IsNegative() bool
}

// Glob is a glob pattern that has been compiled into a regular expression.
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

// Options modify the behaviour of Glob parsing
type Options struct {
	// The character used to split path components
	Separator rune
	// Set to false to allow any prefix before the glob match
	MatchAtStart bool
	// Set to false to allow any suffix after the glob match
	MatchAtEnd bool
}

// DefaultOptions are a default set of Options that uses a forward slash as a separator, and require a full match
var DefaultOptions = &Options{
	Separator:    '/',
	MatchAtStart: true,
	MatchAtEnd:   true,
}

func (g *globImpl) String() string {
	return g.globPattern
}

func (g *globImpl) IsNegative() bool {
	return g.negated
}

// Compile parses the given glob pattern and convertes it to a Glob. If no options are given, the DefaultOptions are
// used.
func Compile(pattern string, options *Options) (Glob, error) {
	pattern = strings.TrimSpace(pattern)
	if options == nil {
		options = DefaultOptions
	} else {
		// Check that the separator is not an expander
		for _, expander := range expanders {
			if options.Separator == expander {
				return nil, fmt.Errorf("'%s' is not allowed as a separator", string(options.Separator))
			}
		}
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
	scanner := bufio.NewScanner(strings.NewReader(pattern))
	scanner.Split(separatorsScanner([]rune{options.Separator}))
	for i := 0; scanner.Scan(); i++ {
		component := scanner.Text()

		// If the component is just a separator, discard it
		if component == string(options.Separator) {
			continue
		}

		err = parseComponent(scanner.Text(), i, glob)
		if err != nil {
			return nil, err
		}
	}

	if options.MatchAtEnd {
		glob.parserState.regexBuffer.WriteRune('$')
	}

	regexString := glob.parserState.regexBuffer.String()
	Logger.Debugf("[ohmyglob:Glob] Compiled \"%s\" to regex `%s` (negated: %v)", pattern, regexString, glob.negated)
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
	reBuf := glob.parserState.regexBuffer

	if component == "**" {
		isGlobStar = true
		// Only add another globstar if the last component wasn't a globstar
		if !glob.parserState.lastComponentWasGlobStar {
			// Enclose in a optional non-captured group so we can enforce the need for a separator, even if there is no
			// intervening path component
			reBuf.WriteString("(?:")
			if idx != 0 {
				reBuf.WriteString(glob.parserState.escapedSeparator)
			}
			reBuf.WriteString(globStarComponent)
			reBuf.WriteString(")?")
		}
	} else if component == "*" {
		if idx != 0 {
			reBuf.WriteString(glob.parserState.escapedSeparator)
		}
		reBuf.WriteString("[^")
		reBuf.WriteString(glob.parserState.escapedSeparator)
		reBuf.WriteString("]+")
	} else {
		if idx != 0 {
			reBuf.WriteString(glob.parserState.escapedSeparator)
		}

		// Scan through, expanding ? and *'s
		scan := bufio.NewScanner(strings.NewReader(component))
		scan.Split(separatorsScanner(expanders))

		for scan.Scan() {
			part := scan.Text()
			switch part {
			case "?":
				reBuf.WriteString("[^")
				reBuf.WriteString(glob.parserState.escapedSeparator)
				reBuf.WriteString("]")
			case "*":
				reBuf.WriteString("[^")
				reBuf.WriteString(glob.parserState.escapedSeparator)
				reBuf.WriteString("]*")
			default:
				reBuf.WriteString(escapeRegexComponent(part))
			}
		}
	}

	glob.parserState.lastComponentWasGlobStar = isGlobStar
	return nil
}
