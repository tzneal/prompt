Prompt
======

[![Build Status](https://travis-ci.org/tzneal/prompt.svg?branch=master)](https://travis-ci.org/tzneal/prompt)
[![GoDoc](https://godoc.org/github.com/tzneal/prompt?status.svg)](https://godoc.org/github.com/tzneal/prompt)
[![Coverage Status](https://coveralls.io/repos/github/tzneal/prompt/badge.svg?branch=master)](https://coveralls.io/github/tzneal/prompt?branch=master)

Prompt is a library for adding shell like interfaces to command line
applications.  It's intended to be similar to the interface used by some
networking equipment.

[![asciicast](https://asciinema.org/a/bz1e2gczb14gqhdgci51ku30x.png)](https://asciinema.org/a/bz1e2gczb14gqhdgci51ku30x)

Features
--------

Prompt supports:
* built-in command completion
* history
* context sensitive completion
* command sets
* command output to a file
* command output filtering (e.g. 'grep')


License
-------
MIT

Warning
-------
This library should be considered unstable for now.  Interfaces may change,
and your code may break as I add features and re-think existing features.
