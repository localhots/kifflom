package parser

import (
	"github.com/localhots/punk/lexer"
)

type (
	Parser struct {
		lex      *lexer.Lexer
		selector string
	}
)

func New(lex *lexer.Lexer, selector string) {
	return &Parser{
		lex:      lex,
		selector: selector,
	}
}
