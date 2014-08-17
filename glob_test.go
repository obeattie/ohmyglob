package ohmyglob

import (
	"os"
	"testing"

	log "github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

func init() {
	Logger, _ = log.LoggerFromWriterWithMinLevel(os.Stderr, log.TraceLvl)
}

func TestSimpleGlob(t *testing.T) {
	pattern := "foo/*/b?r"
	glob, err := Compile(pattern, nil)
	assert.NoError(t, err)

	match := "foo/baz/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/bbr"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/bdr"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz∆˙¨®˙¨¥ƒ®†˙ƒ†¨®†√˙/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/bar"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)

	// With a * as a partial component
	pattern = "foo/*/baz/--foo/--*/--baz"
	glob, err = Compile(pattern, nil)
	assert.NoError(t, err)
	match = "foo/bar/baz/--foo/--bar/--baz"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
}

func TestGlobStar(t *testing.T) {
	pattern := "foo/**/bar/**"
	glob, err := Compile(pattern, nil)
	assert.NoError(t, err)

	match := "foo/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/bar/barrrrr"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/boop/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/boop/µ∂^~®˙¨˙çƒ®†¨^/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "thisistotal/garbage/ÔÔÈ´^¨∆~∆≈∆∫"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
	match = "foobar"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
	match = "foobar/bar"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
	match = "foo/barbar"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)

	// Check consecutive globstars work correctly
	pattern = "foo/**/**/bar"
	glob, err = Compile(pattern, nil)
	assert.NoError(t, err)
	match = "foo/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/boop/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/baz/boop/µ∂^~®˙¨˙çƒ®†¨^/bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "thisistotal/garbage/ÔÔÈ´^¨∆~∆≈∆∫"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
	match = "foobar"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)

	// Globstar at the head
	pattern = "**/**/foobar"
	glob, err = Compile(pattern, nil)
	assert.NoError(t, err)
	match = "foobar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/foobar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo/bar/foo/bar/foobar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)

	// Globstar at the tail
	pattern = "foobar/**/**"
	glob, err = Compile(pattern, nil)
	assert.NoError(t, err)
	match = "foobar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foobar/baz/boo"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foobar/baz"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
}

// Check that setting MatchAtStart to false allows any prefix
func TestNoMatchAtStart(t *testing.T) {
	pattern := "foo"
	glob, err := Compile(pattern, &Options{
		Separator:    DefaultOptions.Separator,
		MatchAtStart: false,
		MatchAtEnd:   DefaultOptions.MatchAtEnd,
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
	glob, err := Compile(pattern, &Options{
		Separator:    DefaultOptions.Separator,
		MatchAtStart: DefaultOptions.MatchAtStart,
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
	glob, err := Compile(pattern, nil)
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
	glob, err := Compile(pattern, &Options{
		Separator:    '.',
		MatchAtStart: DefaultOptions.MatchAtStart,
		MatchAtEnd:   DefaultOptions.MatchAtEnd,
	})
	assert.NoError(t, err)

	match := "foo.baz.bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo.baz∆˙¨®˙¨¥ƒ®†˙ƒ†¨®†√˙.bar"
	assert.True(t, glob.MatchString(match), "%s should match %s", pattern, match)
	match = "foo.bar"
	assert.False(t, glob.MatchString(match), "%s should not match %s", pattern, match)
}

// Illegal separators should return an error on construction
func TestIllegalSeparator(t *testing.T) {
	_, err := Compile("foo", &Options{
		Separator: '?',
	})
	assert.Error(t, err, "? should not be allowed as a separator")

	_, err = Compile("foo", &Options{
		Separator: '\\',
	})
	assert.Error(t, err, "\\ should not be allowed as a separator")
}

// Meaningful can be escaped with a backslash
func TestEscaping(t *testing.T) {
	// Maps to a pair of (should, shouldn't) strings
	expectations := map[string][2]string{
		`\/`:                        [2]string{`/`, ``},
		`foo\bar/\*\/baz`:           [2]string{`foobar/*/baz`, `foobar/foo/baz`},
		`foo\?bar`:                  [2]string{`foo?bar`, `foosbar`},
		`\/\/\/\*\*.*`:              [2]string{`///**.foobar`, `///foobar.baz`},
		`\`:                         [2]string{``, `\`},
		`foo\bar/baz`:               [2]string{`foobar/baz`, `foo\bar/baz`},
		`foo\\bar\/\//barbaz\*\?\\`: [2]string{`foo\bar///barbaz*?\\`, `a`},
	}

	for pattern, s := range expectations {
		glob, err := Compile(pattern, DefaultOptions)
		assert.NoError(t, err)

		should, shouldnt := s[0], s[1]
		assert.False(t, glob.MatchString(shouldnt), "Glob `%s` should not match `%s`", pattern, shouldnt)
		assert.True(t, glob.MatchString(should), "Glob `%s` should match `%s`", pattern, should)
	}

	// Custom separator
}

// Benchmark globbing from start to finish; constructing and matching
func BenchmarkGlobbing(b *testing.B) {
	pattern := "foo/**/baz/--fo?/*/--baz"
	b.ResetTimer()
	g, err := Compile(pattern, nil)
	assert.NoError(b, err)
	assert.True(b, g.MatchString("foo/bar/bar/baz/--foo/--bar/--baz"))
}
