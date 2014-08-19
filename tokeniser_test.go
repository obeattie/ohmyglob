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
	Logger.Tracef("[ohmyglob:TestTokeniser_Simple] Testing \"%s\"", input)
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	e := expectations{
		eToken{"abcdef abcdef abcdef", tcLiteral},
	}
	testTokenRun(t, tokeniser, e)
}

func TestTokeniser_Wildcard(t *testing.T) {
	input := `part1*part2`
	Logger.Tracef("[ohmyglob:TestTokeniser_Wildcard] Testing \"%s\"", input)
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
	Logger.Tracef("[ohmyglob:TestTokeniser_Globstar] Testing \"%s\"", input)
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
	Logger.Tracef("[ohmyglob:TestTokeniser_AnyCharacter] Testing \"%s\"", input)
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
	Logger.Tracef("[ohmyglob: TestTokeniser_Separator] Testing \"%s\"", input)
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	e := expectations{
		eToken{"part1", tcLiteral},
		eToken{"/", tcSeparator},
		eToken{"part2", tcLiteral},
	}
	testTokenRun(t, tokeniser, e)
}

func TestTokeniser_CustomSeparator(t *testing.T) {
	input := `part1فpart2`
	Logger.Tracef("[ohmyglob: TestTokeniser_Separator] Testing \"%s\"", input)
	tokeniser := newGlobTokeniser(strings.NewReader(input), &Options{
		Separator:    'ف',
		MatchAtStart: true,
		MatchAtEnd:   true,
	})

	e := expectations{
		eToken{"part1", tcLiteral},
		eToken{"ف", tcSeparator},
		eToken{"part2", tcLiteral},
	}
	testTokenRun(t, tokeniser, e)
}

func TestTokeniser_Escaper(t *testing.T) {
	input := `part1\*part2\\\\\foobar`
	Logger.Tracef("[ohmyglob:TestTokeniser_Escaper] Testing \"%s\"", input)
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

// Test various cominations; we don't just have one giant function because we want to know which individual components
// are broken, if they are
func TestTokeniser_Combinations(t *testing.T) {
	es := map[string]expectations{
		`foobar/*`: expectations{
			eToken{`foobar`, tcLiteral},
			eToken{`/`, tcSeparator},
			eToken{`*`, tcStar},
		},
		`foobar/*/baz`: expectations{
			eToken{`foobar`, tcLiteral},
			eToken{`/`, tcSeparator},
			eToken{`*`, tcStar},
			eToken{`/`, tcSeparator},
			eToken{`baz`, tcLiteral},
		},
		`foobar/**/baz\/boo?foo!`: expectations{
			eToken{`foobar`, tcLiteral},
			eToken{`/`, tcSeparator},
			eToken{`**`, tcGlobStar},
			eToken{`/`, tcSeparator},
			eToken{`baz`, tcLiteral},
			eToken{`/boo`, tcLiteral},
			eToken{`?`, tcAny},
			eToken{`foo!`, tcLiteral},
		},
		`**/foo//bar/ba?**dee\//这是可\怕的/////\///`: expectations{
			eToken{`**`, tcGlobStar},
			eToken{`/`, tcSeparator},
			eToken{`foo`, tcLiteral},
			eToken{`/`, tcSeparator},
			eToken{`/`, tcSeparator},
			eToken{`bar`, tcLiteral},
			eToken{`/`, tcSeparator},
			eToken{`ba`, tcLiteral},
			eToken{`?`, tcAny},
			eToken{`**`, tcGlobStar},
			eToken{`dee`, tcLiteral},
			eToken{`/`, tcLiteral},
			eToken{`/`, tcSeparator},
			eToken{`这是可`, tcLiteral},
			eToken{`怕的`, tcLiteral},
			eToken{`/`, tcSeparator},
			eToken{`/`, tcSeparator},
			eToken{`/`, tcSeparator},
			eToken{`/`, tcSeparator},
			eToken{`/`, tcSeparator},
			eToken{`/`, tcLiteral},
			eToken{`/`, tcSeparator},
			eToken{`/`, tcSeparator},
		},
	}

	for input, e := range es {
		Logger.Tracef("[ohmyglob:TestTokeniser_Combinations] Testing \"%s\"", input)
		tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)
		testTokenRun(t, tokeniser, e)
	}
}

func TestTokeniser_Peeking(t *testing.T) {
	input := `foo/*/bar/**/baz?`
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType := tokeniser.Token()
	assert.Equal(t, `foo`, token)
	assert.Equal(t, tcLiteral, tokenType)

	assert.True(t, tokeniser.Peek(), "Expected peek not possible")
	assert.NoError(t, tokeniser.Err(), "Peek caused unexpected error")
	assert.NoError(t, tokeniser.PeekErr(), "Unexpected peek error")
	newToken, newTokenType := tokeniser.Token()
	assert.Equal(t, token, newToken, "Peek altered token returned by Token()")
	assert.Equal(t, tokenType, newTokenType, "Peek altered tokenType returned by Token()")
	peekedToken, peekedTokenType := tokeniser.PeekToken()
	assert.Equal(t, `/`, peekedToken)
	assert.Equal(t, tcSeparator, peekedTokenType)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, peekedToken, token, "Scan()'d token does not match prior Peek()")
	assert.Equal(t, peekedTokenType, tokenType, "Scan()'d tokenType does not match prior Peek()")

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, `*`, token)
	assert.Equal(t, tcStar, tokenType)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, `/`, token)
	assert.Equal(t, tcSeparator, tokenType)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, `bar`, token)
	assert.Equal(t, tcLiteral, tokenType)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, `/`, token)
	assert.Equal(t, tcSeparator, tokenType)

	assert.True(t, tokeniser.Peek(), "Expected peek not possible")
	assert.NoError(t, tokeniser.Err(), "Peek caused unexpected error")
	assert.NoError(t, tokeniser.PeekErr(), "Unexpected peek error")
	newToken, newTokenType = tokeniser.Token()
	assert.Equal(t, token, newToken, "Peek altered token returned by Token()")
	assert.Equal(t, tokenType, newTokenType, "Peek altered tokenType returned by Token()")
	peekedToken, peekedTokenType = tokeniser.PeekToken()
	assert.Equal(t, `**`, peekedToken)
	assert.Equal(t, tcGlobStar, peekedTokenType)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, peekedToken, token, "Scan()'d token does not match prior Peek()")
	assert.Equal(t, peekedTokenType, tokenType, "Scan()'d tokenType does not match prior Peek()")

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, `/`, token)
	assert.Equal(t, tcSeparator, tokenType)

	assert.True(t, tokeniser.Peek(), "Expected peek not possible")
	assert.NoError(t, tokeniser.Err(), "Peek caused unexpected error")
	assert.NoError(t, tokeniser.PeekErr(), "Unexpected peek error")
	newToken, newTokenType = tokeniser.Token()
	assert.Equal(t, token, newToken, "Peek altered token returned by Token()")
	assert.Equal(t, tokenType, newTokenType, "Peek altered tokenType returned by Token()")
	peekedToken, peekedTokenType = tokeniser.PeekToken()
	assert.Equal(t, `baz`, peekedToken)
	assert.Equal(t, tcLiteral, peekedTokenType)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, peekedToken, token, "Scan()'d token does not match prior Peek()")
	assert.Equal(t, peekedTokenType, tokenType, "Scan()'d tokenType does not match prior Peek()")

	assert.True(t, tokeniser.Peek(), "Expected peek not possible")
	assert.NoError(t, tokeniser.Err(), "Peek caused unexpected error")
	assert.NoError(t, tokeniser.PeekErr(), "Unexpected peek error")
	newToken, newTokenType = tokeniser.Token()
	assert.Equal(t, token, newToken, "Peek altered token returned by Token()")
	assert.Equal(t, tokenType, newTokenType, "Peek altered tokenType returned by Token()")
	peekedToken, peekedTokenType = tokeniser.PeekToken()
	assert.Equal(t, `?`, peekedToken)
	assert.Equal(t, tcAny, peekedTokenType)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, peekedToken, token, "Scan()'d token does not match prior Peek()")
	assert.Equal(t, peekedTokenType, tokenType, "Scan()'d tokenType does not match prior Peek()")

	assert.False(t, tokeniser.Peek(), "Peeked past end of input")
	assert.NoError(t, tokeniser.PeekErr()) // EOF's aren't reported
	assert.NoError(t, tokeniser.Err())     // EOF's aren't reported

	assert.False(t, tokeniser.Scan(), "Scanned past end of input")
	assert.NoError(t, tokeniser.Err())     // EOF's aren't reported
	assert.NoError(t, tokeniser.PeekErr()) // EOF's aren't reported
}
