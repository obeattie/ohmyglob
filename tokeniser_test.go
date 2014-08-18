package ohmyglob

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokeniser_Simple(t *testing.T) {
	input := "abcdef abcdef abcdef"
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType := tokeniser.Token()
	assert.Equal(t, tcLiteral, tokenType)
	assert.Equal(t, "abcdef abcdef abcdef", token)

	assert.False(t, tokeniser.Scan(), "Should only return one token")
	assert.NoError(t, tokeniser.Err()) // EOF's aren't reported
}

func TestTokeniser_Separator(t *testing.T) {
	input := "part1/part2"
	tokeniser := newGlobTokeniser(strings.NewReader(input), DefaultOptions)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType := tokeniser.Token()
	assert.Equal(t, tcLiteral, tokenType)
	assert.Equal(t, "part1", token)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, tcSeparator, tokenType)
	assert.Equal(t, string(DefaultOptions.Separator), token)

	assert.True(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err())
	token, tokenType = tokeniser.Token()
	assert.Equal(t, tcLiteral, tokenType)
	assert.Equal(t, "part2", token)

	assert.False(t, tokeniser.Scan())
	assert.NoError(t, tokeniser.Err()) // EOF's aren't reported
}
