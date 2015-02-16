package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/localhots/punk/lexer"
)

type (
	ContextType int
	Context     struct {
		Type ContextType
		Key  string
	}
	Parser struct {
		exps    []Context
		context []Context
		lex     *lexer.Lexer
	}
)

const (
	Unknown ContextType = iota
	Object
	Array
)

func New(b []byte, sel string) *Parser {
	p := &Parser{
		exps:    parseSelector(sel),
		context: []Context{},
		lex:     lexer.New(string(b)),
	}

	return p
}

func (p *Parser) Parse() {
	go p.lex.Run()
	p.parseValue(p.next())
}

func (p *Parser) parseValue(item lexer.Item) {
	if item.Token == lexer.BraceOpen {
		p.enterContext(Object)
		p.parseObject()
		p.leaveContext()
		return
	}
	if item.Token == lexer.BracketOpen {
		p.enterContext(Array)
		p.parseArray(0)
		p.leaveContext()
		return
	}

	isMatch := p.checkContext()
	switch item.Token {
	case lexer.Null, lexer.Bool, lexer.Number, lexer.String:
		if isMatch {
			fmt.Printf("\n\nFOUND MATCH!\nVALUE: %s\n\n", item.String())
			panic("Match found")
		}
	default:
		if isMatch {
			panic("Cannot print your match, sorry :(")
		} else {
			unexpected(item)
		}
	}
}

func (p *Parser) parseArray(i int) {
	p.context[len(p.context)-1].Key = strconv.Itoa(i)
	item := p.next()
	if item.Token == lexer.BracketClose {
		return
	}

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
		return
	case lexer.String:
		p.context[len(p.context)-1].Key = item.Val
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

func (p *Parser) checkContext() bool {
	depth := len(p.context)
	if depth != len(p.exps) {
		return false
	}

	fmt.Println("Checking...")
	fmt.Println(p.exps)
	fmt.Println(p.context)

	for i, exp := range p.exps {
		ctx := p.context[i]
		if exp.Type != ctx.Type || exp.Key != ctx.Key {
			return false
		}
	}

	return true
}

func (p *Parser) next() lexer.Item {
	if item, ok := p.lex.NextItem(); ok {
		fmt.Println(item)
		return item
	} else {
		panic("EOF reached")
	}
}

func (p *Parser) enterContext(typ ContextType) {
	p.context = append(p.context, Context{
		Type: typ,
	})
}

func (p *Parser) leaveContext() {
	p.context = p.context[:len(p.context)-1]
}

func unexpected(item lexer.Item) {
	panic(fmt.Errorf("Unexpected token: %s", item.String()))
}

func parseSelector(sel string) []Context {
	exps := []Context{}
	parts := strings.Split(sel[1:], "/")
	for _, part := range parts {
		typ := Object
		if len(part) > 2 && part[:1] == "[" && part[len(part)-1:] == "]" {
			part = part[1 : len(part)-1]
			typ = Array
		}
		exps = append(exps, Context{
			Type: typ,
			Key:  part,
		})
	}

	return exps
}

func (e ContextType) String() string {
	switch e {
	case Array:
		return "Index"
	case Object:
		return "Key"
	default:
		return "Unknown"
	}
}
