package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/localhots/punk/lexer"
)

type (
	ExpectationType int
	Expectation     struct {
		Type  ExpectationType
		Key   string
		Index int64
	}
	Parser struct {
		exps []Expectation
	}
)

const (
	Object ExpectationType = iota
	Array
)

func New(selector string) *Parser {
	return &Parser{
		exps: parseSelector(selector),
	}
}

func (p *Parser) Parse(b []byte) {
	lex := lexer.New(string(b))
	go lex.Run()

	for {
		if item, ok := lex.NextItem(); ok {
			fmt.Println(item)
		} else {
			break
		}
	}
}

func parseSelector(sel string) []Expectation {
	exps := []Expectation{}
	parts := strings.Split(sel[1:], "/")
	for _, part := range parts {
		if len(part) > 2 && part[:1] == "[" && part[len(part)-1:] == "]" {
			index, _ := strconv.ParseInt(part[1:len(part)-1], 10, 64)
			exps = append(exps, Expectation{
				Type:  Array,
				Index: index,
			})
		} else {
			exps = append(exps, Expectation{
				Type: Object,
				Key:  part,
			})
		}
	}

	return exps
}
