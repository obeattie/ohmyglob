package ohmyglob

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// eToken is an expected token {token string, token type}
type eToken [2]interface{}

// expectations is an ordered set of eToken token expectations
type expectations []eToken

func testTokenRun(t *testing.T, tokeniser *globTokeniser, e expectations) {
	for i, expectation := range e {
		expectedToken, expectedTokenType := expectation[0], expectation[1]

		assert.True(t, tokeniser.Scan(), "Expected token %d not yielded", i)
		assert.NoError(t, tokeniser.Err(), "Unexpected error retrieving token %d", i)
		token, tokenType := tokeniser.Token()
		assert.Equal(t, expectedTokenType, tokenType, "Token %d type does not match expectation", i)
		assert.Equal(t, expectedToken, token, "Token %d value does not match expectation", i)
	}

	assert.False(t, tokeniser.Scan(), "Should only return %d tokens", len(e))
	assert.NoError(t, tokeniser.Err()) // EOF's aren't reported
}

func TestTokeniser_Simple(t *testing.T) {
	input := `abcdef abcdef abcdef`
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	e := expectations{
		eToken{"abcdef abcdef abcdef", tcLiteral},
	}
	testTokenRun(t, tokeniser, e)
}

func TestTokeniser_Wildcard(t *testing.T) {
	input := `part1*part2`
	Logger.Tracef("Testing \"%s\"", input)
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	e := expectations{
		eToken{"part1", tcLiteral},
		eToken{"*", tcStar},
		eToken{"part2", tcLiteral},
	}
	testTokenRun(t, tokeniser, e)
}

func TestTokeniser_Globstar(t *testing.T) {
	input := `part1**part2****`
	Logger.Tracef("Testing \"%s\"", input)
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	e := expectations{
		eToken{"part1", tcLiteral},
		eToken{"**", tcGlobStar},
		eToken{"part2", tcLiteral},
		eToken{"**", tcGlobStar},
		eToken{"**", tcGlobStar},
	}
	testTokenRun(t, tokeniser, e)
}

func TestTokeniser_AnyCharacter(t *testing.T) {
	input := `part1?part2`
	Logger.Tracef("Testing \"%s\"", input)
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	e := expectations{
		eToken{"part1", tcLiteral},
		eToken{"?", tcAny},
		eToken{"part2", tcLiteral},
	}
	testTokenRun(t, tokeniser, e)
}

func TestTokeniser_Separator(t *testing.T) {
	input := `part1/part2`
	Logger.Tracef("Testing \"%s\"", input)
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	e := expectations{
		eToken{"part1", tcLiteral},
		eToken{"/", tcSeparator},
		eToken{"part2", tcLiteral},
	}
	testTokenRun(t, tokeniser, e)
}

func TestTokeniser_Escaper(t *testing.T) {
	input := `part1\*part2\\\\\foobar`
	Logger.Tracef("Testing \"%s\"", input)
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	e := expectations{
		// The tokeniser will generate two tokens here; split at the location of the escaper. As consumers should be
		// able to handle consecutive literals anyway, this is not a problem
		eToken{`part1`, tcLiteral},
		eToken{`*part2`, tcLiteral},
		eToken{`\`, tcLiteral},
		eToken{`\`, tcLiteral},
		// The "f" here was escaped, and it should have no effect
		eToken{`foobar`, tcLiteral},
	}
	testTokenRun(t, tokeniser, e)
}
