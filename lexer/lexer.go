// This lexer is based on ideas and code presented in the talk by Rob Pike
// called "Lexical Scanning in Go". More info could be found in Golang's Blog:
// http://blog.golang.org/two-go-talks-lexical-scanning-in-go-and
package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type (
	// Holds the state of the scanner
	Lexer struct {
		input   string    // The string being scanned
		lineNum int       // Line number
		pos     int       // Current position in the input
		start   int       // Start position of this item
		width   int       // Width of last rune read from input
		items   chan Item // Channel of scanned items
	}

	// Represents a token returned from the scanner
	Item struct {
		Token Token  // The type of this item
		Val   string // The value of this item
		Pos   int    // The starting position, in bytes, of this item in the input string
	}

	// Identifies the type of the item
	Token int

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

// Creates a new scanner for the input string
func New(input string) *Lexer {
	return &Lexer{
		input: input,
		items: make(chan Item),
	}
}

// Starts the state machine for the lexer
func (l *Lexer) Run() {
	for state := lexInitial; state != nil; {
		state = state(l)
	}
}

// Returns the next scanned item and a boolean, which is false on EOF
func (l *Lexer) NextItem() (item Item, ok bool) {
	item, ok = <-l.items
	return
}

// Returns the next rune in the input
func (l *Lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// Returns the value for the next token
func (l *Lexer) val() string {
	return l.input[l.start:l.pos]
}

// Returns but does not consume the next rune in the input
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// Tells if the following input matches the given string
func (l *Lexer) acceptString(s string) (ok bool) {
	if strings.HasPrefix(l.input[l.pos:], s) {
		l.pos += len(s)
		return true
	}
	return false
}

// Steps back one rune
func (l *Lexer) backup() {
	l.pos -= l.width
}

// Skips over the pending input before this point
func (l *Lexer) ignore() {
	l.start = l.pos
}

// Passes an item back to the client
func (l *Lexer) emit(t Token) {
	l.items <- Item{
		Token: t,
		Val:   l.input[l.start:l.pos],
		Pos:   l.start,
	}
	l.start = l.pos
	if t == EOF {
		close(l.items)
	}
}

// Emits an error token with given string as a value and stops lexing
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{
		Token: Error,
		Val:   fmt.Sprintf(format, args...),
		Pos:   l.start,
	}
	close(l.items)
	return nil
}

//
// States
//

func lexInitial(l *Lexer) stateFn {
	for {
		switch r := l.next(); r {
		case ' ', '\t':
			return lexSpace(l)
		case '\n':
			l.lineNum++
			l.ignore()
		case 'n':
			l.backup()
			return lexNull(l)
		case 't', 'f':
			l.backup()
			return lexBool(l)
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			l.backup()
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
			return l.errorf("Unexpected symbol: %c", r)
		}
	}
}

// Skips all spaces in the input until a visible character is found
func lexSpace(l *Lexer) stateFn {
	for {
		if r := l.next(); r != ' ' && r != '\t' {
			l.backup()
			break
		}
	}
	l.ignore()
	return lexInitial
}

func lexNull(l *Lexer) stateFn {
	if l.acceptString("null") {
		l.emit(Null)
	} else {
		return l.errorf("Unexpected (null) token: %q", l.val())
	}
	return lexInitial
}

func lexBool(l *Lexer) stateFn {
	if l.acceptString("true") || l.acceptString("false") {
		l.emit(Bool)
	} else {
		return l.errorf("Unexpected (bool) token: %q", l.val())
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
			l.backup()
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
				l.backup()
				l.emit(String)
				// Skipping closing quote
				l.next()
				l.ignore()
				return lexInitial
			}
		case '\n':
			l.lineNum++
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
	switch i.Token {
	case EOF:
		return "EOF"
	case Error:
		return fmt.Sprintf("(Error: %q)", i.Val)
	case BraceOpen:
		return "{"
	case BraceClose:
		return "}"
	case BracketOpen:
		return "["
	case BracketClose:
		return "]"
	case Quote:
		return "\""
	case Colon:
		return ":"
	case Comma:
		return ","
	case Null:
		return fmt.Sprintf("(NULL: %q)", i.Val)
	case Bool:
		return fmt.Sprintf("(Bool: %q)", i.Val)
	case Number:
		return fmt.Sprintf("(Number: %q)", i.Val)
	case String:
		return fmt.Sprintf("(String: %q)", i.Val)
	default:
		panic("Unreachable")
	}
}
