// This lexer is based on ideas and code presented in the talk by Rob Pike
// called "Lexical Scanning in Go". More info could be found in Golang's Blog:
// http://blog.golang.org/two-go-talks-lexical-scanning-in-go-and
package lexer

import (
	"fmt"

	"github.com/localhots/punk/buffer"
)

type (
	// Holds the state of the scanner
	Lexer struct {
		input     buffer.Bufferer
		stack     []rune    // Lexer stack
		pos       int       // Current stack position
		lineNum   int       // Line number
		colNum    int       // Column number
		startLine int       // Start line of this item
		startCol  int       // Start column of this item
		items     chan Item // Channel of scanned items
	}

	// Represents a token returned from the scanner
	Item struct {
		Token  Token  // The type of this item
		Val    string // The value of this item
		Line   int    // Line number
		Column int    // Column number
	}

	// Identifies the type of the item
	Token byte

	// Represents the state of the scanner as a function that returns the next state
	stateFn func(*Lexer) stateFn
)

const (
	// Special
	Error Token = iota
	EOF

	// Symbols
	BraceOpen    // {
	BraceClose   // }
	BracketOpen  // [
	BracketClose // [
	Quote        // "
	Colon        // :
	Comma        // ,

	// Types
	Null
	Bool
	Number
	String
)

// Creates a new scanner for the input buffer
func New(input buffer.Bufferer) *Lexer {
	return &Lexer{
		input:   input,
		items:   make(chan Item),
		lineNum: 1,
		colNum:  0,
	}
}

// Starts the state machine for the lexer
func (l *Lexer) Run() {
	for state := lexInitial; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// Returns the next scanned item and a boolean, which is false on EOF
func (l *Lexer) NextItem() (item Item, ok bool) {
	item, ok = <-l.items
	return
}

// Returns the next rune in the stack
func (l *Lexer) next() rune {
	var r rune
	// Reading next rune from buffer
	if l.pos > len(l.stack)-1 {
		l.stack = append(l.stack, l.input.Next())
	}
	r = l.stack[l.pos]
	l.pos++

	// Counting lines and columns - token coordinates
	if r == '\n' {
		l.lineNum++
		l.colNum = 0
	} else {
		l.colNum++
	}

	return r
}

// Returns the value for the next token
func (l *Lexer) val() string {
	return string(l.stack[:l.pos])
}

// Returns but does not consume the next rune in the stack
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup(1)
	return r
}

// Tells if the following stack matches the given string
func (l *Lexer) acceptString(s string) (ok bool) {
	for i, c := range s {
		if l.next() != c {
			l.backup(i + 1)
			return false
		}
	}
	return true
}

// Steps back one rune
// Backup is never called right after a new line char so we don't care
// about the line number. This is also true for the ignore function
func (l *Lexer) backup(n int) {
	l.pos -= n
	l.colNum -= n
}

// Clears the stack items preceding the current position
func (l *Lexer) ignore() {
	if l.pos < len(l.stack) {
		l.stack = l.stack[l.pos:]
	} else {
		l.stack = []rune{}
	}
	l.pos = 0
	l.startLine = l.lineNum
	l.startCol = l.colNum
}

// Passes an item back to the client
func (l *Lexer) emit(t Token) {
	// Single-character tokens never backup
	if len(l.val()) == 1 {
		l.startCol++
	}
	l.items <- Item{
		Token:  t,
		Val:    l.val(),
		Line:   l.startLine,
		Column: l.startCol,
	}
	l.ignore() // Cleaning up stack
}

// Emits an error token with given string as a value and stops lexing
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{
		Token:  Error,
		Val:    fmt.Sprintf(format, args...),
		Line:   l.startLine,
		Column: l.startCol,
	}
	return nil
}

//
// States
//

func lexInitial(l *Lexer) stateFn {
	for {
		switch r := l.next(); r {
		case ' ', '\t', '\n':
			l.ignore()
		case 'n':
			l.backup(1)
			return lexNull(l)
		case 't', 'f':
			l.backup(1)
			return lexBool(l)
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			l.backup(1)
			return lexNumber(l)
		case '"':
			return lexString(l)
		case '[':
			l.emit(BracketOpen)
		case ']':
			l.emit(BracketClose)
		case '{':
			l.emit(BraceOpen)
		case '}':
			l.emit(BraceClose)
		case ':':
			l.emit(Colon)
		case ',':
			l.emit(Comma)
		case 0:
			l.emit(EOF)
			return nil
		default:
			return l.errorf("Unexpected symbol: %q", r)
		}
	}
}

func lexNull(l *Lexer) stateFn {
	if l.acceptString("null") {
		l.emit(Null)
	} else {
		return l.errorf("Unexpected (null) symbol: %q", l.val())
	}
	return lexInitial
}

func lexBool(l *Lexer) stateFn {
	if l.acceptString("true") || l.acceptString("false") {
		l.emit(Bool)
	} else {
		return l.errorf("Unexpected (bool) symbol: %q", l.val())
	}
	return lexInitial
}

func lexNumber(l *Lexer) stateFn {
	numDots := 0
	for {
		switch r := l.next(); r {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		case '.':
			numDots++
		default:
			l.backup(1)
			if numDots > 1 || r == '.' {
				return l.errorf("Invalid number: %q", l.val())
			}
			l.emit(Number)
			return lexInitial
		}
	}
}

func lexString(l *Lexer) stateFn {
	// Skipping opening quote
	l.ignore()
	escaped := false
	for {
		switch r := l.next(); r {
		case '\\':
			escaped = !escaped
		case '"':
			if escaped {
				escaped = false
			} else {
				// Going before closing quote and emitting
				l.backup(1)
				l.emit(String)
				// Skipping closing quote
				l.next()
				l.ignore()
				return lexInitial
			}
		case 0:
			return l.errorf("Unterminated string")
		default:
			escaped = false
		}
	}
}

//
// Debug
//

func (i Item) String() string {
	var label string
	switch i.Token {
	case BraceOpen, BraceClose, BracketOpen, BracketClose, Quote, Colon, Comma:
		label = i.Val
	case EOF:
		label = "EOF"
	case Error:
		label = fmt.Sprintf("(Error: %s)", i.Val)
	case Null:
		label = fmt.Sprintf("(NULL: %q)", i.Val)
	case Bool:
		label = fmt.Sprintf("(Bool: %q)", i.Val)
	case Number:
		label = fmt.Sprintf("(Number: %q)", i.Val)
	case String:
		label = fmt.Sprintf("(String: %q)", i.Val)
	default:
		panic("Unreachable")
	}

	return fmt.Sprintf("[%.3d:%.3d] %s", i.Line, i.Column, label)
}
