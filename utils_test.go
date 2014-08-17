package ohmyglob

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestRegexComponentEscaping(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	charsTested := 0
	sampled := make([]rune, 10)

	for _, table := range unicode.PrintRanges {
		for _, charRange := range table.R16 {
			for c := charRange.Lo; c <= charRange.Hi; c += charRange.Stride {
				charsTested++
				if rand.Float32() <= 0.05 {
					sampled = append(sampled, rune(c))
				}
				component := escapeRegexComponent(string(c))
				re := regexp.MustCompile(fmt.Sprintf("^%s$", component))
				if !assert.True(t, re.MatchString(string(c))) {
					return
				}
			}
		}
	}

	t.Logf("Tested %d runes individually", charsTested)

	// Test permutations of the sampled runes
	t.Logf("Sampled %d runes", len(sampled))
	permsTested := 0
	for i := 0; i <= len(sampled)/2; i++ {
		permsTested++

		seq := rand.Perm(rand.Intn(len(sampled)))
		buf := new(bytes.Buffer)
		for _, idx := range seq {
			buf.WriteRune(sampled[idx])
		}

		charSeq := buf.String()
		re := regexp.MustCompile(fmt.Sprintf("^%s$", escapeRegexComponent(charSeq)))
		if !assert.True(t, re.MatchString(charSeq)) {
			return
		}
	}
	t.Logf("Tested %d permutations", permsTested)
}

func TestEscapeGlobComponent(t *testing.T) {
	expectations := map[string]string{
		`foobar`:                 `foobar`,
		`foo/bar`:                `foo\/bar`,
		`foobarbaz*/foobar`:      `foobarbaz\*\/foobar`,
		`.*////.**/foobar///baz`: `.\*\/\/\/\/.\*\*\/foobar\/\/\/baz`,
		`\foobar`:                `\\foobar`,
		`/∆≈¨´∂#ª˙ƒ¨∞˙**®´∆¢#º....///..∂˚ø´∂˚®≥...`: `\/∆≈¨´∂#ª˙ƒ¨∞˙\*\*®´∆¢#º....\/\/\/..∂˚ø´∂˚®≥...`,
	}

	for src, result := range expectations {
		assert.Equal(t, result, EscapeGlobComponent(src, DefaultOptions))
	}
}
