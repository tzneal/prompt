package prompt

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

const eof = -1

//go:generate stringer -type=itemType
type itemType byte

const (
	itemChanClose itemType = iota
	itemError
	itemWord
	itemSemi
	itemQuotedString
	itemLineCont
	itemFilename
	itemPlaceholder
	itemCompletionType
	itemPipe
	itemRAngle
	itemEOF
)

type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	return fmt.Sprintf("<%s \"%s\">", i.typ, i.val)
}

type lexer struct {
	input string    // the string being scanned.
	mode  lexMode   // the lex mode (command description or user input)
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	items chan item // channel of scanned items.
}

type lexStateFn func(*lexer) lexStateFn

func startState(*lexer) lexStateFn {
	return startState
}

type lexMode byte

const (
	cmdDescMode lexMode = iota
	userInputMode
)

func lex(input string, mode lexMode) chan item {
	l := &lexer{
		input: input,
		mode:  mode,
		items: make(chan item),
	}
	go l.run() // Concurrently run state machine.
	return l.items
}

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) run() {
	for state := lexCommand; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) lexStateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// lexQuote returns a function scans a quoted string.
func lexQuote(delimeter rune) func(l *lexer) lexStateFn {

	return func(l *lexer) lexStateFn {
	Loop:
		for {
			switch l.next() {
			case '\\':
				if r := l.next(); r != eof && r != '\n' {
					break
				}
				fallthrough
			case eof, '\n':
				return l.errorf("unterminated quoted string")
			case delimeter:
				break Loop
			}
		}
		l.emit(itemQuotedString)
		return lexCommand
	}
}

func lexWord(l *lexer) lexStateFn {
	for {
		r := l.next()

		if !isWord(r) {
			break
		}
	}
	l.backup()
	l.emit(itemWord)
	return lexCommand
}

func lexFilename(l *lexer) lexStateFn {
	l.skipSpace()
	for {
		switch r := l.next(); {
		case isWord(r):
			continue

		case isSpace(r):
			l.emit(itemFilename)
			return lexCommand

		case r == ';':
			l.backup()
			l.emit(itemFilename)
			return lexCommand

		case r == eof:
			l.emit(itemFilename)
			return nil

		default:
			return l.errorf("unexpected filename character '%c'", r)
		}
	}
}

func lexFilter(l *lexer) lexStateFn {
	l.emit(itemPipe)
	for {
		l.skipSpace()
		switch r := l.next(); {
		case isWord(r):
			return lexWord
		case r == '|':
			return lexFilter
		case r == ';':
			l.emit(itemSemi)
			return lexCommand
		default:
			return l.errorf("unexpected filter character '%c'", r)
		}
	}
}

func (l *lexer) skipSpace() {
	// skip any leading spaces
	for isSpace(l.peek()) {
		l.next()
		l.ignore()
	}
}

func lexCompletionType(l *lexer) lexStateFn {
	l.next()
	l.ignore()
	for r := l.next(); isAlphaNumeric(r); r = l.next() {
	}
	l.backup()
	l.emit(itemCompletionType)
	return lexCommand(l)
}

func lexPlaceholder(l *lexer) lexStateFn {
	l.ignore()
	for {
		switch r := l.next(); {
		case r == '*':
			l.emit(itemPlaceholder)
			if l.peek() == ':' {
				return lexCompletionType
			}
			return lexCommand
		case r == eof:
			return l.errorf("unterminated placeholder")
		case isDigit(r):
			for isDigit(l.peek()) {
				l.next()
			}
			l.emit(itemPlaceholder)
			if l.peek() == ':' {
				return lexCompletionType
			}
			return lexCommand
		default:
			return l.errorf("invalid placeholder character '%c'", r)
		}
	}
}

func lexCommand(l *lexer) lexStateFn {
	for {
		l.skipSpace()
		switch r := l.next(); {
		case r == '`':
			return lexQuote('`')
		case r == '"':
			return lexQuote('"')
		case r == '\'':
			return lexQuote('\'')
		case r == ';':
			l.emit(itemSemi)
		case r == '|':
			return lexFilter
		case r == '>':
			l.emit(itemRAngle)
			return lexFilename
		case l.mode == cmdDescMode && r == '$':
			return lexPlaceholder
		case r == '\\' && isEndOfLine(l.peek()):
			for isEndOfLine(l.peek()) {
				l.next()
				l.ignore()
			}
			l.emit(itemLineCont)
		case isEndOfLine(r):
			l.ignore()
			l.emit(itemSemi)
		case r == eof:
			return nil
		default:
			return lexWord
		}
	}
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func isWord(r rune) bool {
	// being very lenient here for now
	switch r {
	case '$', ' ', '\t', '\n', ';', '|', '>', eof:
		return false
	default:
		return true
	}
}
