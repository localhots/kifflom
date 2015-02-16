package lexer

import "strings"

func lexInitial(l *Lexer) stateFn {
loop:
	for {
		switch r := l.next(); r {
		case ' ', '\t':
			return lexSpace(l)
		case '\n':
			l.lineNum++
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
			l.emit(itemBracketOpen)
		case ']':
			l.emit(itemBracketClose)
		case '{':
			l.emit(itemBraceOpen)
		case '}':
			l.emit(itemBraceClose)
		case ':':
			l.emit(itemColon)
		case ',':
			l.emit(itemComma)
		case EOF:
			break loop
		default:
			panic("Unexpected symbol: " + string(r))
		}
	}

	// Correctly reached EOF.
	l.emit(itemEOF)

	return nil
}

// Skip all spaces
// One space has already been seen
func lexSpace(l *Lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.ignore()
	return lexInitial
}

func lexNull(l *Lexer) stateFn {
	if l.acceptString("null") {
		l.emit(itemNull)
	} else {
		return l.errorf("Unexpected token")
	}
	return lexInitial
}

func lexBool(l *Lexer) stateFn {
	if l.acceptString("true") || l.acceptString("false") {
		l.emit(itemBool)
	}
	return lexInitial
}

func lexNumber(l *Lexer) stateFn {
	hasDot := false
	for {
		if r := l.peek(); isDigit(r) {
			l.next()
		} else if r == '.' {
			if hasDot {
				return l.errorf("Invalid number")
			} else {
				hasDot = true
				l.next()
			}
		} else {
			break
		}
	}

	l.emit(itemNumber)
	return lexInitial
}

func lexString(l *Lexer) stateFn {
	escaped := false
loop:
	for {
		switch r := l.next(); r {
		case '\\':
			escaped = true
		case '"':
			if escaped {
				escaped = false
			} else {
				l.emit(itemString)
				break loop
			}
		case EOF:
			return l.errorf("String hits EOF")
		default:
			escaped = false
		}
	}

	return lexInitial
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isDigit(r rune) bool {
	return strings.IndexRune("1234567890", r) > -1
}
