package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/localhots/punk/buffer"
	"github.com/localhots/punk/lexer"
)

type (
	// Holds the state of parser
	Parser struct {
		lex  *lexer.Lexer
		ctx  *context
		sels map[string]*context
		res  chan Match
	}
	Match struct {
		Sel string
		Val interface{}
	}
)

// Creates a new parser
func New(buf buffer.Bufferer, sels []string) *Parser {
	return &Parser{
		lex: lexer.New(buf),
		ctx: &context{
			exps: []expectation{},
		},
		sels: parseSelectors(sels),
		res:  make(chan Match),
	}
}

// Parse all and return matches
func (p *Parser) Parse() map[string][]interface{} {
	p.ParseStream()
	out := map[string][]interface{}{}
	for {
		if m, ok := <-p.res; ok {
			out[m.Sel] = append(out[m.Sel], m.Val)
		} else {
			break
		}
	}
	return out
}

// Starts parsing
func (p *Parser) ParseStream() <-chan Match {
	go p.lex.Run()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("\nParse error! Yay!")
				fmt.Println(err)
			}
			close(p.res)
		}()
		for {
			if item := p.next(); item.Token != lexer.EOF {
				p.parseValue(item)
			} else {
				break
			}
		}
	}()
	return p.res
}

func (p *Parser) parseValue(item lexer.Item) {
	switch item.Token {
	case lexer.Null, lexer.Bool, lexer.Number, lexer.String:
		p.pushValue(item)
	case lexer.BraceOpen:
		p.ctx.push(object)
		p.parseObject()
		p.ctx.pop()
	case lexer.BracketOpen:
		p.ctx.push(array)
		p.parseArray(0)
		p.ctx.pop()
	default:
		unexpected(item)
	}
}

// Parses array recursively part by part
// Is called after '[' and ',' tokens
// Expects a value followed by ']' or ',' tokens
func (p *Parser) parseArray(i int64) {
	item := p.next()
	if item.Token == lexer.BracketClose {
		// Neither a bug nor a feature
		// This allows an array to have a trailing comma
		// [1, 2, 3, ]
		return
	}

	p.ctx.setIndex(i)
	p.parseValue(item)

	switch item := p.next(); item.Token {
	case lexer.BracketClose:
		return
	case lexer.Comma:
		p.parseArray(i + 1)
	}
}

func (p *Parser) parseObject() {
	item := p.next()
	switch item.Token {
	case lexer.BraceClose:
		// Neither a bug nor a feature
		// This allows an object to have a trailing comma
		// {"foo": 1, "bar": 2, }
		return
	case lexer.String:
		p.ctx.setKey(item.Val)
	default:
		unexpected(item)
	}

	if item := p.next(); item.Token != lexer.Colon {
		unexpected(item)
	}

	p.parseValue(p.next())

	switch item := p.next(); item.Token {
	case lexer.BraceClose:
		return
	case lexer.Comma:
		p.parseObject()
	default:
		unexpected(item)
	}
}

func (p *Parser) pushValue(item lexer.Item) {
	for sel, exp := range p.sels {
		if ok := exp.compare(p.ctx); ok {
			if val, err := castValue(item); err == nil {
				p.res <- Match{
					Sel: sel,
					Val: val,
				}
			} else {
				p.res <- Match{
					Sel: sel,
					Val: err,
				}
			}
			return
		}
	}
}

func (p *Parser) next() lexer.Item {
	if item, ok := p.lex.NextItem(); ok {
		if item.Token == lexer.Error {
			panic(item)
		}

		fmt.Println(item)
		return item
	} else {
		panic("EOF reached")
	}
}

func castValue(item lexer.Item) (val interface{}, err error) {
	switch item.Token {
	case lexer.Null:
		val = nil
	case lexer.Bool:
		val = (item.Val == "true")
	case lexer.String:
		val = item.Val
	case lexer.Number:
		if strings.Index(item.Val, ".") > -1 {
			val, err = strconv.ParseFloat(item.Val, 64)
		} else {
			val, err = strconv.ParseInt(item.Val, 10, 64)
		}
	}
	return
}

func unexpected(item lexer.Item) {
	panic(fmt.Errorf("Unexpected token: %s", item.String()))
}
