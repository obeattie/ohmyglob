package ohmyglob

import (
	"bytes"
	"io"
)

type tc uint8

const (
	// An unknown component; returned if there is an error scanning or there are no more tokens
	tcUnknown = tc(0x0)
	// A string literal
	tcLiteral = tc(0x1)
	// An Escaper
	tcEscaper = tc(0x2)
	// Any characters, aside from the separator
	tcStar = tc(0x3)
	// A globstar component (tc = type component)
	tcGlobStar = tc(0x4)
	// Any single character, aside from the separator
	tcAny = tc(0x5)
	// A separator
	tcSeparator = tc(0x6)
)

// Tokenises a glob input; implements an API very similar to that of bufio.Scanner (though is not identical)
type globTokeniser struct {
	input         io.RuneScanner
	globOptions   *Options
	token         string
	tokenType     tc
	err           error
	hasPeek       bool
	peekToken     string
	peekTokenType tc
	peekErr       error
}

func newGlobTokeniser(input io.RuneScanner, globOptions *Options) *globTokeniser {
	return &globTokeniser{
		input:       input,
		globOptions: globOptions,
	}
}

// Advances by a single token
func (g *globTokeniser) parse(lastTokenType tc) (string, tc, error) {
	var err error

	tokenBuf := new(bytes.Buffer)
	tokenType := tcUnknown
	escaped := lastTokenType == tcEscaper

	for {
		var r rune
		r, _, err = g.input.ReadRune()
		if err != nil {
			break
		}

		runeType := tcUnknown
		switch r {
		case Escaper:
			runeType = tcEscaper
		case '*':
			if tokenType == tcStar {
				runeType = tcGlobStar
				tokenType = tcGlobStar
			} else {
				runeType = tcStar
			}
		case '?':
			runeType = tcAny
		case g.globOptions.Separator:
			runeType = tcSeparator
		default:
			runeType = tcLiteral
		}

		if escaped {
			// If the last token was an Escaper, this MUST be a literal
			runeType = tcLiteral
			escaped = false
		}

		if (tokenType != tcUnknown) && (tokenType != runeType) {
			// We've stumbled into the next token; backtrack
			g.input.UnreadRune()
			break
		}

		tokenType = runeType
		tokenBuf.WriteRune(r)

		if tokenType == tcEscaper ||
			tokenType == tcGlobStar ||
			tokenType == tcAny ||
			tokenType == tcSeparator {
			// These tokens are standalone; continued consumption must be a separate token
			break
		}
	}

	if err == io.EOF && tokenType != tcUnknown {
		// If we have a token, we can't have an EOF: we want the EOF on the next pass
		err = nil
	}

	if err != nil {
		return "", tcUnknown, err
	}

	if tokenType == tcEscaper {
		// Escapers should never be yielded; recurse to find the next token
		return g.parse(tokenType)
	}

	return tokenBuf.String(), tokenType, err
}

// Scan advances the tokeniser to the next token, which will then be available through the Token method. It returns
// false when the tokenisation stops, either by reaching the end of the input or an error. After Scan returns false,
// the Err method will return any error that occurred during scanning, except that if it was io.EOF, Err will return
// nil.
func (g *globTokeniser) Scan() bool {
	if g.hasPeek {
		g.token, g.tokenType, g.err = g.peekToken, g.peekTokenType, g.peekErr
	} else {
		g.token, g.tokenType, g.err = g.parse(g.tokenType)
	}

	g.peekErr = nil
	g.peekToken = ""
	g.peekTokenType = tcUnknown
	g.hasPeek = false
	return g.err == nil
}

// Peek peeks to the next token, making it available as PeekToken(). Next time Scan() is called it will advance the
// tokeniser to the peeked token. If there is already a peaked token, it will not advance.
func (g *globTokeniser) Peek() bool {
	if !g.hasPeek {
		g.peekToken, g.peekTokenType, g.peekErr = g.parse(g.tokenType)
		g.hasPeek = true
	}

	return g.peekErr == nil
}

// Err returns the first non-EOF error that was encountered by the tokeniser
func (g *globTokeniser) Err() error {
	if g.err == io.EOF {
		return nil
	}

	return g.err
}

func (g *globTokeniser) Token() (token string, tokenType tc) {
	return g.token, g.tokenType
}

// PeekToken returns the peeked token
func (g *globTokeniser) PeekToken() (token string, tokenType tc) {
	return g.peekToken, g.peekTokenType
}

// PeekErr returns the error that will be returned by Err() next time Scan() is called. Peek() must be called first.
func (g *globTokeniser) PeekErr() error {
	if g.peekErr == io.EOF {
		return nil
	}

	return g.peekErr
}
