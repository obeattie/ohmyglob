# ohmyglob

A minimal glob matching utility for Go.

![ohmyglob!](http://i.imgur.com/7vjO2mF.gif)

[![GoDoc](https://godoc.org/github.com/obeattie/ohmyglob?status.svg)](https://godoc.org/github.com/obeattie/ohmyglob) [![Build Status](https://travis-ci.org/obeattie/ohmyglob.svg?branch=master)](https://travis-ci.org/obeattie/ohmyglob)

Works internally by converting glob expressions into [`Regexp`](http://golang.org/pkg/regexp/#Regexp)s, which are then
used to match strings.

## Features

* Customisable separators
* `*` matches any number of characters, but not the separator
* `?` matches a single character, but not the separator
* `!` at the beginning of a pattern will negate the match
* ["Globstar"](http://www.linuxjournal.com/content/globstar-new-bash-globbing-option) (`**`) matching
* Glob sets allow matching against a set of ordered globs, with precedence to later matches

## Usage

    import glob "github.com/obeattie/ohmyglob"
    
    var g glob.Glob
    var err error
    var doesMatch bool
    
    // Standard, with a wildcard
    g, err = glob.Compile("foo/*/baz", glob.DefaultOptions)
    doesMatch = g.MatchString("foo/bar/baz") // true!
    doesMatch = g.MatchString("nope") // false!
    
    // Globstar
    g, err = glob.Compile("foo/**/baz", glob.DefaultOptions)
    doesMatch = g.MatchString("foo/bar/bar/baz") // true!
    doesMatch = g.MatchString("foo/baz") // true!
