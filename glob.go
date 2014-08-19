// Package ohmyglob provides a minimal glob matching utility.
package ohmyglob

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	log "github.com/cihub/seelog"
)

var (
	// Logger is used to log trace-level info; logging is completely disabled by default but can be changed by replacing
	// this with a configured logger
	Logger log.LoggerInterface
	// Runes that, in addition to the separator, mean something when they appear in the glob (includes Escaper)
	expanders = []rune{'?', '*', Escaper}
	// Map built from expanders
	expandersMap map[rune]bool
	// Character used to escape a meaningful character
	Escaper = '\\'
)

func init() {
	expandersMap = make(map[rune]bool, len(expanders))
	for _, r := range expanders {
		expandersMap[r] = true
	}

	if Logger == nil {
		var err error
		Logger, err = log.LoggerFromWriterWithMinLevel(os.Stderr, log.CriticalLvl) // seelog bug means we can't use log.Off
		if err != nil {
			panic(err)
		}
	}
}

type processedToken struct {
	contents  *bytes.Buffer
	tokenType tc
}

type parserState struct {
	options *Options
	// The regex-escaped separator character
	escapedSeparator string
	processedTokens  []processedToken
}

// GlobMatcher is the basic interface of a Glob or GlobSet. It provides a Regexp-style interface for checking matches.
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

func popLastToken(state *parserState) *processedToken {
	state.processedTokens = state.processedTokens[:len(state.processedTokens)-1]
	if len(state.processedTokens) > 0 {
		return &state.processedTokens[len(state.processedTokens)-1]
	}
	return &processedToken{}
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

	state := &parserState{
		options:          options,
		escapedSeparator: escapeRegexComponent(string(options.Separator)),
		processedTokens:  make([]processedToken, 0, 10),
	}
	glob := &globImpl{
		Regexp:      nil,
		globPattern: pattern,
		negated:     false,
		parserState: state,
	}

	regexBuf := new(bytes.Buffer)
	if options.MatchAtStart {
		regexBuf.WriteRune('^')
	}

	var err error
	// Transform into a regular expression pattern
	// 1. Parse negation prefixes
	pattern, err = parseNegation(pattern, glob)
	if err != nil {
		return nil, err
	}

	// 2. Tokenise and convert!
	tokeniser := newGlobTokeniser(strings.NewReader(pattern), options)
	lastProcessedToken := &processedToken{}
	for tokeniser.Scan() {
		if err = tokeniser.Err(); err != nil {
			return nil, err
		}

		token, tokenType := tokeniser.Token()
		t := processedToken{
			contents:  nil,
			tokenType: tokenType,
		}

		// Special cases
		if tokenType == tcGlobStar && tokeniser.Peek() {
			// If the next token is a separator, consume it
			if err = tokeniser.PeekErr(); err != nil {
				return nil, err
			}
			_, peekedType := tokeniser.PeekToken()
			if peekedType == tcSeparator {
				tokeniser.Scan()
			}
		}
		if tokenType == tcGlobStar && lastProcessedToken.tokenType == tcGlobStar {
			// If the last token was a globstar and this is too, remove the last
			lastProcessedToken = popLastToken(state)
		}
		if tokenType == tcGlobStar && lastProcessedToken.tokenType == tcSeparator && !tokeniser.Peek() {
			// If this is the last token, and it's a globstar, remove a preceeding separator
			lastProcessedToken = popLastToken(state)
		}

		t.contents, err = processToken(token, tokenType, glob, tokeniser)
		if err != nil {
			return nil, err
		}

		lastProcessedToken = &t
		state.processedTokens = append(state.processedTokens, t)
	}

	fmt.Print(lastProcessedToken)

	for _, t := range state.processedTokens {
		t.contents.WriteTo(regexBuf)
	}

	if options.MatchAtEnd {
		regexBuf.WriteRune('$')
	}

	regexString := regexBuf.String()
	Logger.Infof("[ohmyglob:Glob] Compiled \"%s\" to regex `%s` (negated: %v)", pattern, regexString, glob.negated)
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

func processToken(token string, tokenType tc, glob *globImpl, tokeniser *globTokeniser) (*bytes.Buffer, error) {
	state := glob.parserState
	buf := new(bytes.Buffer)

	switch tokenType {
	case tcGlobStar:
		// Globstars also take care of surrounding separators; separator components before and after a globstar are
		// suppressed
		isLast := !tokeniser.Peek()
		buf.WriteString("(?:")
		if isLast {
			buf.WriteString(state.escapedSeparator)
		}
		buf.WriteString(".+")
		if !isLast {
			buf.WriteString(state.escapedSeparator)
		}
		buf.WriteString(")?")
	case tcStar:
		buf.WriteString("[^")
		buf.WriteString(state.escapedSeparator)
		buf.WriteString("]*")
	case tcAny:
		buf.WriteString("[^")
		buf.WriteString(state.escapedSeparator)
		buf.WriteString("]")
	case tcSeparator:
		buf.WriteString(state.escapedSeparator)
	case tcLiteral:
		buf.WriteString(escapeRegexComponent(token))
	}

	return buf, nil
}
