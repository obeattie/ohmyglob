package ohmyglob

import (
	"bytes"
	"io"
)

type tc uint8

const (
	// An unknown component; returned if there is an error scanning or there are no more tokens
	tcUnknown = tc(255)
	// An Escaper
	tcEscaper = tc(0)
	// A globstar component (tc = type component)
	tcGlobStar = tc(1)
	// Any characters, aside from the separator
	tcStar = tc(2)
	// Any single character, aside from the separator
	tcAny = tc(3)
	// A separator
	tcSeparator = tc(4)
	// A string literal
	tcLiteral = tc(5)
)

// Tokenises a glob input; implements an API very similar to that of bufio.Scanner (though is not identical)
type globTokeniser struct {
	input       io.RuneScanner
	globOptions *Options
	token       string
	tokenType   tc
	lastErr     error
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

	runesConsumed := 0
	for ; ; runesConsumed++ {
		var r rune
		r, _, err = g.input.ReadRune()
		if err != nil {
			break
		}

		runeType := tcUnknown
		if lastTokenType == tcEscaper {
			// If the last token was an escaper, this MUST be a literal
			runeType = tcLiteral
		} else {
			switch r {
			case Escaper:
				runeType = tcEscaper
			case '*':
				if tokenType == tcStar {
					runeType = tcGlobStar
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
		}

		if tokenType != tcUnknown && tokenType != runeType {
			// We've stumbled upon the next token; backtrack
			g.input.UnreadRune()
			break
		}

		tokenType = runeType
		tokenBuf.WriteRune(r)

		if tokenType == tcEscaper ||
			tokenType == tcGlobStar ||
			tokenType == tcStar ||
			tokenType == tcAny ||
			tokenType == tcSeparator {
			// Deal with the standalone tokens; these cannot continue consuming
			break
		}
	}

	if err == io.EOF && runesConsumed > 0 {
		// Don't report this at all if we have a token
		err = nil
	}

	if err != nil && err != io.EOF {
		// Ensure this is set back, in case an error occured after these were set in "the loop" (EOF errors don't count)
		g.token = ""
		g.tokenType = tcUnknown
	} else {
		g.token = tokenBuf.String()
		g.tokenType = tokenType

		if tokenType == tcEscaper {
			// Escapers should never be yielded; recurse to find the next token
			return g.parse()
		}
	}

	Logger.Tracef("[Tokeniser] parse(): err %v, token %v, tokenType %v", err, g.token, g.tokenType)

	return err
}

// Scan advances the Scanner to the next token, which will then be available through the Token method. It returns false
// when the scan stops, either by reaching the end of the input or an error. After Scan returns false, the Err method
// will return any error that occurred during scanning, except that if it was io.EOF, Err will return nil.
func (g *globTokeniser) Scan() bool {
	err := g.parse()
	if err == io.EOF {
		// EOF errors aren't stored as lastErr
		g.lastErr = nil
	} else {
		g.lastErr = err
	}
	return err == nil
}

// Err returns the first non-EOF error that was encountered by the tokeniser
func (g *globTokeniser) Err() error {
	return g.lastErr
}

func (g *globTokeniser) Token() (token string, tokenType tc) {
	return g.token, g.tokenType
}
