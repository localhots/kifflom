package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type (
	// lexer holds the state of the scanner.
	Lexer struct {
		input   string    // the string being scanned
		lineNum int       // Line number
		pos     int       // current position in the input
		start   int       // start position of this item
		width   int       // width of last rune read from input
		items   chan Item // channel of scanned items
	}

	// Item represents a token or text string returned from the scanner.
	Item struct {
		Token Token  // The type of this item.
		Pos   int    // The starting position, in bytes, of this item in the input string.
		Val   string // The value of this item.
	}

	// Token identifies the type of lex items.
	Token int

	// stateFn represents the state of the scanner as a function that returns the next state.
	stateFn func(*Lexer) stateFn
)

const (
	// Special
	Error Token = iota // error occurred; value is text of error
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
	Null   // null
	Bool   // true, false
	Number // 0, 2.5
	String // "foo"
)

// lex creates a new scanner for the input string.
func New(input string) *Lexer {
	l := &Lexer{
		input: input,
		items: make(chan Item),
	}
	return l
}

// run runs the state machine for the lexer.
func (l *Lexer) Run() {
	for state := lexInitial; state != nil; {
		state = state(l)
	}
}

func (l *Lexer) NextItem() (item Item, ok bool) {
	item, ok = <-l.items
	return
}

//
// Lexer stuff
//

func (i Item) String() string {
	switch i.Token {
	case EOF:
		return "EOF"
	case Error:
		return "Error: " + i.Val
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
		return "NULL"
	case Bool:
		return "Bool: " + i.Val
	case Number:
		return "Number: " + i.Val
	case String:
		return "String: " + i.Val
	default:
		panic("Unreachable")
	}
}

// next returns the next rune in the input.
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
func (l *Lexer) emit(t Token) {
	l.items <- Item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
	if t == EOF {
		close(l.items)
	}
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

func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{Error, l.start, fmt.Sprintf(format, args...)}
	return nil // Stop lexing
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
			panic("Unexpected symbol: " + string(r))
		}
	}
}

// Skip all spaces
func lexSpace(l *Lexer) stateFn {
	for {
		if r := l.peek(); r == ' ' || r == '\t' {
			l.next()
		} else {
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
		return l.errorf("Unexpected token")
	}
	return lexInitial
}

func lexBool(l *Lexer) stateFn {
	if l.acceptString("true") || l.acceptString("false") {
		l.emit(Bool)
	}
	return lexInitial
}

func lexNumber(l *Lexer) stateFn {
	hasDot := false
	for {
		switch r := l.peek(); r {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			l.next()
		case '.':
			if hasDot {
				return l.errorf("Invalid number")
			} else {
				hasDot = true
				l.next()
			}
		default:
			l.emit(Number)
			return lexInitial
		}
	}
}

func lexString(l *Lexer) stateFn {
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
				l.backup() // Going before closing quote
				l.emit(String)
				l.next() // Skipping closing quote
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
