package ohmyglob

import (
	"io"
	"strings"

	log "github.com/cihub/seelog"
)

// GlobSet represents an ordered set of Globs, and has the same matching capabilities as a Glob. Globbing is done
// in-order, with later globs taking precedence over earlier globs in the set. A GlobSet is immutable.
type GlobSet interface {
	GlobMatcher
	// String returns the patterns used to create the GlobSet
	String() string
	// MatchingGlob returns the Glob that matches the specified pattern (or does not match, in the case of a negative
	// glob)
	MatchingGlob(b []byte) Glob
}

type globSetImpl []Glob

func (g globSetImpl) String() string {
	strs := make([]string, len(g))
	for i, glob := range g {
		strs[i] = glob.String()
	}
	return strings.Join(strs, ", ")
}

func (g globSetImpl) MatchingGlob(b []byte) Glob {
	for i := len(g) - 1; i >= 0; i-- {
		glob := g[i]
		matches := glob.Match(b)
		if matches {
			log.Tracef("[GlobSet] %s matched to %s", string(b), glob.String())
			return glob
		}
	}

	return nil
}

func (g globSetImpl) Match(b []byte) bool {
	glob := g.MatchingGlob(b)
	return glob != nil && !glob.IsNegative()
}

func (g globSetImpl) MatchReader(r io.RuneReader) bool {
	// Drain the reader and convert to a byte array
	b := make([]byte, 0, 10)
	for {
		rn, _, err := r.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return false
		}
		b = append(b, byte(rn))
	}

	return g.Match(b)
}

func (g globSetImpl) MatchString(s string) bool {
	return g.Match([]byte(s))
}

// NewGlobSet constructs a GlobSet from a slice of Globs
func NewGlobSet(globs []Glob) (GlobSet, error) {
	set := make(globSetImpl, len(globs))
	for i, glob := range globs {
		set[i] = glob
	}
	return set, nil
}

// CompileGlobSet constructs a GlobSet from a slice of strings, which will be compiled individually to Globs
func CompileGlobSet(patterns []string, options *GlobOptions) (GlobSet, error) {
	globs := make(globSetImpl, len(patterns))
	for i, pattern := range patterns {
		glob, err := CompileGlob(pattern, options)
		if err != nil {
			return nil, err
		}
		globs[i] = glob
	}

	return globs, nil
}
