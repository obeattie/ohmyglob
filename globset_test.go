package ohmyglob

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobSet_MatchString(t *testing.T) {
	patterns := []string{
		"foo/*/baz",
		"!foo/notme/baz",
		"!foo/butyesme/baz",
		"foo/butyesme/baz",
		"!foo/notme/baz",
		"!foo/*/baz/foo/**",
	}
	set, err := CompileGlobSet(patterns, DefaultOptions)
	assert.NoError(t, err)

	match := "foo/baz/baz"
	assert.True(t, set.MatchString(match), "(%s) should match %s", set.String(), match)
	match = "foo/notme/baz"
	assert.False(t, set.MatchString(match), "(%s) should not match %s", set.String(), match)
	match = "foo/butyesme/baz"
	assert.True(t, set.MatchString(match), "(%s) should match %s", set.String(), match)
	match = "foo/boop/baz/foo/bar/baz"
	assert.False(t, set.MatchString(match), "(%s) should not match %s", set.String(), match)
}

func TestGlobSet_Match(t *testing.T) {
	patterns := []string{
		"foo/*/baz",
		"!foo/notme/baz",
		"!foo/butyesme/baz",
		"foo/butyesme/baz",
		"!foo/notme/baz",
		"!foo/*/baz/foo/**",
	}
	set, err := CompileGlobSet(patterns, DefaultOptions)
	assert.NoError(t, err)

	match := []byte("foo/baz/baz")
	assert.True(t, set.Match(match), "(%s) should match %s", set.String(), match)
	match = []byte("foo/notme/baz")
	assert.False(t, set.Match(match), "(%s) should not match %s", set.String(), match)
	match = []byte("foo/butyesme/baz")
	assert.True(t, set.Match(match), "(%s) should match %s", set.String(), match)
	match = []byte("foo/boop/baz/foo/bar/baz")
	assert.False(t, set.Match(match), "(%s) should not match %s", set.String(), match)
}

func TestGlobSet_MatchReader(t *testing.T) {
	patterns := []string{
		"foo/*/baz",
		"!foo/notme/baz",
		"!foo/butyesme/baz",
		"foo/butyesme/baz",
		"!foo/notme/baz",
		"!foo/*/baz/foo/**",
	}
	set, err := CompileGlobSet(patterns, DefaultOptions)
	assert.NoError(t, err)

	match := "foo/baz/baz"
	matchReader := strings.NewReader(match)
	assert.True(t, set.MatchReader(matchReader), "(%s) should match %s", set.String(), match)
	match = "foo/notme/baz"
	matchReader = strings.NewReader(match)
	assert.False(t, set.MatchReader(matchReader), "(%s) should not match %s", set.String(), match)
	match = "foo/butyesme/baz"
	matchReader = strings.NewReader(match)
	assert.True(t, set.MatchReader(matchReader), "(%s) should match %s", set.String(), match)
	match = "foo/boop/baz/foo/bar/baz"
	matchReader = strings.NewReader(match)
	assert.False(t, set.MatchReader(matchReader), "(%s) should not match %s", set.String(), match)
}

func TestGlobSet_AllMatchingGlobs(t *testing.T) {
	patterns := []string{
		"foo/**/baz",
		"!foo/notme/baz",
		"!foo/butyesme/baz",
		"foo/butyesme/baz",
		"!foo/notme/baz",
		"!foo/*/baz/foo/**",
	}
	set, err := CompileGlobSet(patterns, DefaultOptions)
	assert.NoError(t, err)

	match := []byte("foo/notme/baz")
	assert.Len(t, set.AllMatchingGlobs(match), 3)
	match = []byte("foo/butyesme/baz")
	assert.Len(t, set.AllMatchingGlobs(match), 3)
	match = []byte("foo/yes/baz/foo/123")
	assert.Len(t, set.AllMatchingGlobs(match), 1)
	match = []byte("foo/yes/baz/foo/123/baz")
	assert.Len(t, set.AllMatchingGlobs(match), 2)
}
