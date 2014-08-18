package ohmyglob

import (
	"bytes"
	"io"
)

type tc uint8

const (
	// An unknown component; returned if there is an error scanning or there are no more tokens
	tcUnknown = tc(0)
	// A string literal
	tcLiteral = tc(1)
	// An Escaper
	tcEscaper = tc(2)
	// Any characters, aside from the separator
	tcStar = tc(3)
	// A globstar component (tc = type component)
	tcGlobStar = tc(4)
	// Any single character, aside from the separator
	tcAny = tc(5)
	// A separator
	tcSeparator = tc(6)
)

// Tokenises a glob input; implements an API very similar to that of bufio.Scanner (though is not identical)
type globTokeniser struct {
	input       io.RuneScanner
	globOptions *Options
	token       string
	tokenType   tc
	err         error
}

func newGlobTokeniser(input io.RuneScanner, globOptions *Options) *globTokeniser {
	return &globTokeniser{
		input:       input,
		globOptions: globOptions,
	}
}

// Advances by a single token
func (g *globTokeniser) parse() error {
	var err error
	lastTokenType := g.tokenType

	tokenBuf := new(bytes.Buffer)
	tokenType := tcUnknown
	escaped := lastTokenType == tcEscaper

consumer:
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
			break consumer
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
		// Ensure this is set back, in case an error occured after these were set in "the loop" (EOF errors don't count)
		g.token = ""
		g.tokenType = tcUnknown
		return err
	}

	g.token = tokenBuf.String()
	g.tokenType = tokenType

	if tokenType == tcEscaper {
		// Escapers should never be yielded; recurse to find the next token
		Logger.Tracef("[Tokeniser] parse() got tcEscaper; recursing")
		err = g.parse()
		Logger.Tracef("[Tokeniser] parse() recursed")
	} else {
		Logger.Tracef("[Tokeniser] parse() result: err %v, token %v, tokenType %v", err, g.token, g.tokenType)
	}

	return err
}

// Scan advances the Scanner to the next token, which will then be available through the Token method. It returns false
// when the scan stops, either by reaching the end of the input or an error. After Scan returns false, the Err method
// will return any error that occurred during scanning, except that if it was io.EOF, Err will return nil.
func (g *globTokeniser) Scan() bool {
	err := g.parse()
	if err == io.EOF {
		// EOF errors aren't stored
		g.err = nil
	} else {
		g.err = err
	}
	return err == nil
}

// Err returns the first non-EOF error that was encountered by the tokeniser
func (g *globTokeniser) Err() error {
	return g.err
}

func (g *globTokeniser) Token() (token string, tokenType tc) {
	return g.token, g.tokenType
}
