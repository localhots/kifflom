package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type (
	// lexer holds the state of the scanner.
	Lexer struct {
		input      string    // the string being scanned
		state      stateFn   // the next lexing function to enter
		lineNum    int       // Line number
		pos        int       // current position in the input
		start      int       // start position of this item
		width      int       // width of last rune read from input
		lastPos    int       // position of most recent item returned by nextItem
		items      chan Item // channel of scanned items
		parenDepth int       // nesting depth of ( ) exprs
	}

	// stateFn represents the state of the scanner as a function that returns the next state.
	stateFn func(*Lexer) stateFn

	// item represents a token or text string returned from the scanner.
	Item struct {
		typ itemType // The type of this item.
		pos int      // The starting position, in bytes, of this item in the input string.
		val string   // The value of this item.
	}

	// itemType identifies the type of lex items.
	itemType int
)

const (
	// Special
	itemError itemType = iota // error occurred; value is text of error
	itemEOF

	// Symbols
	itemBraceOpen    // {
	itemBraceClose   // }
	itemBracketOpen  // [
	itemBracketClose // [
	itemQuote        // "
	itemColon        // :
	itemComma        // ,

	// Types
	itemNull   // null
	itemBool   // true, false
	itemNumber // 0, 2.5
	itemString // "foo"
)

const (
	EOF = -1
)

// lex creates a new scanner for the input string.
func New(name, input string) *Lexer {
	l := &Lexer{
		input: input,
		items: make(chan Item),
	}
	return l
}

// run runs the state machine for the lexer.
func (l *Lexer) Run() {
	for l.state = lexInitial; l.state != nil; {
		l.state = l.state(l)
	}
}

func (l *Lexer) NextItem() Item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

//
// Lexer stuff
//

func (i Item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return "Error: " + i.val
	case itemBraceOpen:
		return "{"
	case itemBraceClose:
		return "}"
	case itemBracketOpen:
		return "["
	case itemBracketClose:
		return "]"
	case itemQuote:
		return "\""
	case itemColon:
		return ":"
	case itemComma:
		return ","
	case itemNull:
		return "NULL"
	case itemBool:
		return "Bool: " + i.val
	case itemNumber:
		return "Number: " + i.val
	case itemString:
		return "String: " + i.val
	default:
		panic("Unreachable")
	}
}

// next returns the next rune in the input.
func (l *Lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return EOF
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *Lexer) emit(t itemType) {
	l.items <- Item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
}

func (l *Lexer) acceptString(s string) (ok bool) {
	if strings.HasPrefix(l.input[l.pos:], s) {
		l.pos += len(s)
		return true
	}
	return false
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}
