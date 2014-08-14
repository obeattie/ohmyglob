package ohmyglob

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleGlob(t *testing.T) {
	pattern := "foo/*/bar"
	glob, err := CompileGlob(pattern, nil)
	assert.NoError(t, err)

	match := "foo/baz/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz∆˙¨®˙¨¥ƒ®†˙ƒ†¨®†√˙/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/bar"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}

func TestGlobStar(t *testing.T) {
	pattern := "foo/**/bar"
	glob, err := CompileGlob(pattern, nil)
	assert.NoError(t, err)

	match := "foo/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/boop/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/boop/µ∂^~®˙¨˙çƒ®†¨^/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "thisistotal/garbage/ÔÔÈ´^¨∆~∆≈∆∫"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}

// Check that setting MatchAtStart to false allows any prefix
func TestNoMatchAtStart(t *testing.T) {
	pattern := "foo"
	glob, err := CompileGlob(pattern, &GlobOptions{
		Separator:    DefaultGlobOptions.Separator,
		MatchAtStart: false,
		MatchAtEnd:   DefaultGlobOptions.MatchAtEnd,
	})
	assert.NoError(t, err)

	// Test without a prefix still works
	match := "foo"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Test with a prefix still works
	match = "bar/baz/foo"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Test that a prefix and a suffix does not
	match = "bar/baz/foo/boop"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}

// Check that setting MatchAtStart to false allows any suffix
func TestNoMatchAtEnd(t *testing.T) {
	pattern := "foo"
	glob, err := CompileGlob(pattern, &GlobOptions{
		Separator:    DefaultGlobOptions.Separator,
		MatchAtStart: DefaultGlobOptions.MatchAtStart,
		MatchAtEnd:   false,
	})
	assert.NoError(t, err)

	// Test without a suffix still works
	match := "foo"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Test with a suffix still works
	match = "foo/bar/baz"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Test that a prefix and a suffix does not
	match = "bar/baz/foo/boop"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}

func TestNegation(t *testing.T) {
	pattern := "!foo"
	glob, err := CompileGlob(pattern, nil)
	assert.NoError(t, err)
	assert.True(t, glob.IsNegative(), "Glob should be negative")

	// Test it negates the exact string (it should report a match, though it is negative)
	match := "foo"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Should not match other strings
	match = "foo.bar.baz"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
	match = "h4ughfrfg4598yf5uh"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}

func TestCustomSeparator(t *testing.T) {
	pattern := "foo.*.bar"
	glob, err := CompileGlob(pattern, &GlobOptions{
		Separator:    '.',
		MatchAtStart: DefaultGlobOptions.MatchAtStart,
		MatchAtEnd:   DefaultGlobOptions.MatchAtEnd,
	})
	assert.NoError(t, err)

	match := "foo.baz.bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo.baz∆˙¨®˙¨¥ƒ®†˙ƒ†¨®†√˙.bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo.bar"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}
