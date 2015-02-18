package lexer

import (
	"runtime"
	"testing"

	"github.com/localhots/punk/buffer"
)

func TestEmpty(t *testing.T) {
	compare(t, lex(""), []Item{
		Item{EOF, "\x00", 0, 0},
	})
}

func TestNull(t *testing.T) {
	compare(t, lex("null"), []Item{
		Item{Null, "null", 0, 0},
		Item{EOF, "\x00", 0, 0},
	})
}

func TesBool(t *testing.T) {
	compare(t, lex("true"), []Item{
		Item{Bool, "true", 0, 0},
		Item{EOF, "\x00", 0, 0},
	})
	compare(t, lex("false"), []Item{
		Item{Bool, "false", 0, 0},
		Item{EOF, "\x00", 0, 0},
	})
}

func TestString(t *testing.T) {
	compare(t, lex(`"foo"`), []Item{
		Item{String, "foo", 0, 0},
		Item{EOF, "\x00", 0, 0},
	})
}

func TestNumber(t *testing.T) {
	compare(t, lex("123"), []Item{
		Item{Number, "123", 0, 0},
		Item{EOF, "\x00", 0, 0},
	})
	compare(t, lex("123.456"), []Item{
		Item{Number, "123.456", 0, 0},
		Item{EOF, "\x00", 0, 0},
	})
	compare(t, lex("123.456.789"), []Item{
		Item{Error, `Invalid number: "123.456.789"`, 0, 0},
	})
	compare(t, lex("123."), []Item{
		Item{Error, `Invalid number: "123."`, 0, 0},
	})
}

func TestArray(t *testing.T) {
	compare(t, lex(`[1, "2", 3]`), []Item{
		Item{BracketOpen, "[", 0, 0},
		Item{Number, "1", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{String, "2", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{Number, "3", 0, 0},
		Item{BracketClose, "]", 0, 0},
		Item{EOF, "\x00", 0, 0},
	})
}

func TestObject(t *testing.T) {
	compare(t, lex(`{"a": 1, "b": 2}`), []Item{
		Item{BraceOpen, "{", 0, 0},
		Item{String, "a", 0, 0},
		Item{Colon, ":", 0, 0},
		Item{Number, "1", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{String, "b", 0, 0},
		Item{Colon, ":", 0, 0},
		Item{Number, "2", 0, 0},
		Item{BraceClose, "}", 0, 0},
		Item{EOF, "\x00", 0, 0},
	})
}

// Yay!
func TestEverything(t *testing.T) {
	input := `
{
    "foo": true,
    "bar": false,
    "zilch": null,
    "numbers": [1, 23, 4.56, 7.89],
    "bullshit": {
        "nothing": "anything"
    }!
}
`
	compare(t, lex(input), []Item{
		Item{BraceOpen, "{", 0, 0},
		Item{String, "foo", 0, 0},
		Item{Colon, ":", 0, 0},
		Item{Bool, "true", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{String, "bar", 0, 0},
		Item{Colon, ":", 0, 0},
		Item{Bool, "false", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{String, "zilch", 0, 0},
		Item{Colon, ":", 0, 0},
		Item{Null, "null", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{String, "numbers", 0, 0},
		Item{Colon, ":", 0, 0},
		Item{BracketOpen, "[", 0, 0},
		Item{Number, "1", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{Number, "23", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{Number, "4.56", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{Number, "7.89", 0, 0},
		Item{BracketClose, "]", 0, 0},
		Item{Comma, ",", 0, 0},
		Item{String, "bullshit", 0, 0},
		Item{Colon, ":", 0, 0},
		Item{BraceOpen, "{", 0, 0},
		Item{String, "nothing", 0, 0},
		Item{Colon, ":", 0, 0},
		Item{String, "anything", 0, 0},
		Item{BraceClose, "}", 0, 0},
		Item{Error, "Unexpected symbol: '!'", 0, 0},
	})
}

func compare(t *testing.T, reality, expectations []Item) {
	if len(reality) != len(expectations) {
		t.Errorf("Expected %d tokens, got %d", len(reality), len(expectations))
		t.Error(runtime.Caller(1))
		return
	}
	for i, exp := range expectations {
		if exp.Token != reality[i].Token {
			t.Errorf("Expected an %s token, got %s", exp, reality[i])
			t.Error(runtime.Caller(1))
			continue
		}
		if exp.Val != reality[i].Val {
			t.Errorf("Expected an %s token to hold value of %q, got %q", exp, exp.Val, reality[i].Val)
			t.Error(runtime.Caller(1))
		}
	}
}

func lex(json string) []Item {
	buf := buffer.NewBytesBuffer([]byte(json))
	lex := New(buf)
	go lex.Run()

	items := []Item{}
	for {
		if item, ok := lex.NextItem(); ok {
			items = append(items, item)
		} else {
			break
		}
	}

	return items
}
