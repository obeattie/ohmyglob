# ohmyglob

A minimal glob matching utility for Go.

[![Build Status](https://travis-ci.org/obeattie/ohmyglob.svg?branch=master)](https://travis-ci.org/obeattie/ohmyglob)

Works by converting glob expressions into [`Regexp`](http://golang.org/pkg/regexp/#Regexp) objects, which can then be
used to match strings.

Inspired heavily by [minimatch](https://github.com/isaacs/minimatch).
