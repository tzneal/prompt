package prompt

import (
	"errors"
	"strings"
)

//go:generate stringer -type=completionType
type completionType byte

const (
	completeNone completionType = iota
	completePartial
	completeExact
)

//go:generate stringer -type=matchType
type matchType byte

const (
	matchNone matchType = iota
	matchWildcard
	matchSubstitution
	matchExact
)

type command struct {
	desc    input
	execute Command
}

func (c *command) isWildcard() bool {
	for _, i := range c.desc.words {
		if i.typ == placeholderType && i.value == "*" {
			return true
		}
	}
	return false
}

func (c *command) isSubstWildcard() bool {
	for _, i := range c.desc.words {
		if i.typ == placeholderType && i.value != "*" {
			return true
		}
	}
	return false
}

func (c *command) execMatch(line input) matchType {
	// input is longer than the command, and the command is not a wildcard
	if len(line.words) > len(c.desc.words) && !c.isWildcard() {
		return matchNone
	}

	// input is too short to match
	if len(line.words) < len(c.desc.words) {
		return matchNone
	}

	mt := matchExact
	for i, w := range line.words {
		// still matching, and we're now at a wildcard
		if i >= len(c.desc.words) && c.isWildcard() {
			return matchWildcard
		}
		switch matchesSegment(w, c.desc.words[i]) {
		case matchNone:
			return matchNone
		case matchWildcard:
			if mt == matchExact || mt == matchSubstitution {
				mt = matchWildcard
			}
		case matchSubstitution:
			if mt == matchExact {
				mt = matchSubstitution
			}
		}
	}
	return mt
}

func (c *command) complete(line []segment, completers map[string]Completer) (completionType, []string) {
	// no user input, so complete with the first word of each command
	if len(line) == 0 && c.desc.words[0].typ == wordType {
		return completeExact, []string{c.desc.words[0].value}
	}

	// user input is longer than the command allows, and it's not
	// a wildcard command
	if len(line) > len(c.desc.words) && !c.isWildcard() {
		return completeNone, nil
	} else if len(line) >= len(c.desc.words) && c.isWildcard() {
		last := c.desc.words[len(c.desc.words)-1]
		// do we have a completer?
		if completer, ok := completers[last.ctype]; ok {
			lastInput := line[len(line)-1]
			c := completer(lastInput.value)
			// only one completion, and it's the text we matched against
			if len(c) == 1 && c[0] == lastInput.value {
				return completeNone, nil
			}
			return completePartial, c
		}
		return completeNone, nil
	}

	// find the last command description segment that matches
	matchSegIdx := 0
	mt := completeNone
	for i, l := range line {
		if mt = completionMatchSegment(l, c.desc.words[i]); mt == completeNone {
			return completeNone, nil
		}
		matchSegIdx = i
	}

	// exact match for all segments we checked
	if mt == completeExact {
		// already fully matches this command and there are no segments left
		if matchSegIdx == len(c.desc.words)-1 {
			return completeNone, nil
		}
		// offer the next segment as a completion
		matchSegIdx++
	}

	matchSeg := c.desc.words[matchSegIdx]
	// are we matching a $n, or $* segment?
	if matchSeg.typ == placeholderType {
		completer, ok := completers[matchSeg.ctype]
		// unknown or no completer for this placeholder
		if !ok {
			return completeNone, nil
		}

		var words []string
		if matchSegIdx >= len(line) {
			words = completer("")
		} else {
			words = completer(line[matchSegIdx].value)
		}
		if len(words) == 0 {
			return completeNone, nil
		}
		return mt, words
	}

	// not a placeholder, so just return the single next segment
	return completePartial, []string{matchSeg.value}
}

func matchesSegment(input, cmd segment) matchType {
	if cmd.typ == placeholderType {
		if cmd.value == "*" {
			return matchWildcard
		}
		return matchSubstitution
	}
	if cmd.value == input.value {
		return matchExact
	}
	return matchNone
}

func completionMatchSegment(input, cmd segment) completionType {
	if cmd.typ == placeholderType {
		return completePartial
	}
	if cmd.value == input.value {
		return completeExact
	}
	if strings.HasPrefix(cmd.value, input.value) {
		return completePartial
	}
	return completeNone
}

func parseCommand(desc string, fn Command) (*command, error) {
	cmd := &command{}
	inp, err := parse(desc)
	if err != nil {
		return nil, err
	}
	if len(inp) != 1 {
		return nil, errors.New("expected a single command")
	}
	cmd.desc = inp[0]
	cmd.execute = fn
	return cmd, nil
}
