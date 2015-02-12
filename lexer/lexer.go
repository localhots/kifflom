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
		pos        Pos       // current position in the input
		start      Pos       // start position of this item
		width      Pos       // width of last rune read from input
		lastPos    Pos       // position of most recent item returned by nextItem
		items      chan item // channel of scanned items
		parenDepth int       // nesting depth of ( ) exprs
	}
	Pos int

	// stateFn represents the state of the scanner as a function that returns the next state.
	stateFn func(*Lexer) stateFn

	// item represents a token or text string returned from the scanner.
	item struct {
		typ itemType // The type of this item.
		pos Pos      // The starting position, in bytes, of this item in the input string.
		val string   // The value of this item.
	}

	// itemType identifies the type of lex items.
	itemType int
)

const (
	// Special
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemSpace

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
	itemArray  // [1, 2, 3]
	itemObject // {"a": 1, "b": 2}
)

const (
	EOF = -1
)

var (
	itemMap = map[string]itemType{
		"null":  itemNull,
		"true":  itemBool,
		"false": itemBool,
	}
)

// lex creates a new scanner for the input string.
func New(name, input string) *Lexer {
	l := &Lexer{
		input: input,
		items: make(chan item),
	}
	return l
}

// run runs the state machine for the lexer.
func (l *Lexer) Run() {
	for l.state = lexInitial; l.state != nil; {
		l.state = l.state(l)
	}
}

//
// States
//

func lexInitial(l *Lexer) stateFn {
	for {
		switch l.next() {
		case EOF:
			break
		default:
			panic("Unexpected symbol!")
		}
	}

	// Correctly reached EOF.
	l.emit(itemEOF)

	return nil
}

func lexNumber(l *Lexer) stateFn {
	return lexInitial
}

func lexString(l *Lexer) stateFn {
	return lexInitial
}

func lexArray(l *Lexer) stateFn {
	return lexInitial
}

func lexObject(l *Lexer) stateFn {
	return lexInitial
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *Lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(itemSpace)
	return lexInitial
}

//
// Lexer stuff
//

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// next returns the next rune in the input.
func (l *Lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return EOF
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
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
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *Lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *Lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

//
// Helpers
//

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}
