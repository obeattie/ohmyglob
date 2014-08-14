package ohmyglob

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleGlob(t *testing.T) {
	pattern := "foo.*.bar"
	glob, err := NewGlob(pattern, nil)
	assert.NoError(t, err)

	match := "foo.baz.boo/zoo.bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
}

func TestGlobStar(t *testing.T) {
	pattern := "foo.**.bar"
	glob, err := NewGlob(pattern, nil)
	assert.NoError(t, err)

	match := "foo.bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo.baz.bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo.baz.boop.bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo.baz.boop.µ∂^~®˙¨˙çƒ®†¨^.bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "thisistotal.garbage.ÔÔÈ´^¨∆~∆≈∆∫"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}

// Check that setting MatchAtStart to false allows any prefix
func TestNoMatchAtStart(t *testing.T) {
	pattern := "foo"
	glob, err := NewGlob(pattern, &GlobOptions{
		Separator:    DefaultGlobOptions.Separator,
		MatchAtStart: false,
		MatchAtEnd:   DefaultGlobOptions.MatchAtEnd,
	})
	assert.NoError(t, err)

	// Test without a prefix still works
	match := "foo"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Test with a prefix still works
	match = "bar.baz.foo"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Test that a prefix and a suffix does not
	match = "bar.baz.foo.boop"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}

// Check that setting MatchAtStart to false allows any suffix
func TestNoMatchAtEnd(t *testing.T) {
	pattern := "foo"
	glob, err := NewGlob(pattern, &GlobOptions{
		Separator:    DefaultGlobOptions.Separator,
		MatchAtStart: DefaultGlobOptions.MatchAtStart,
		MatchAtEnd:   false,
	})
	assert.NoError(t, err)

	// Test without a prefix still works
	match := "foo"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Test with a prefix still works
	match = "foo.bar.baz"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Test that a prefix and a suffix does not
	match = "bar.baz.foo.boop"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}
