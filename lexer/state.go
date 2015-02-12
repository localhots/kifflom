package lexer

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
