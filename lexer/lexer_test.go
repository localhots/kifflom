package lexer

import "testing"

func TestEmpty(t *testing.T) {
	compare(t, lex(""), []Item{
		Item{EOF, "", 0},
	})
}

func TestNull(t *testing.T) {
	compare(t, lex("null"), []Item{
		Item{Null, "null", 0},
		Item{EOF, "", 0},
	})
}

func TesBool(t *testing.T) {
	compare(t, lex("true"), []Item{
		Item{Bool, "true", 0},
		Item{EOF, "", 0},
	})
	compare(t, lex("false"), []Item{
		Item{Bool, "false", 0},
		Item{EOF, "", 0},
	})
}

func TestString(t *testing.T) {
	compare(t, lex("\"foo\""), []Item{
		Item{String, "foo", 0},
		Item{EOF, "", 0},
	})
}

func TestNumber(t *testing.T) {
	compare(t, lex("123"), []Item{
		Item{Number, "123", 0},
		Item{EOF, "", 0},
	})
	compare(t, lex("123.456"), []Item{
		Item{Number, "123.456", 0},
		Item{EOF, "", 0},
	})
	compare(t, lex("123.456.789"), []Item{
		Item{Error, "Invalid number", 0},
	})
	compare(t, lex("123."), []Item{
		Item{Error, "Invalid number", 0},
	})
}

func TestArray(t *testing.T) {
	compare(t, lex("[1, \"2\", 3]"), []Item{
		Item{BracketOpen, "[", 0},
		Item{Number, "1", 0},
		Item{Comma, ",", 0},
		Item{String, "2", 0},
		Item{Comma, ",", 0},
		Item{Number, "3", 0},
		Item{BracketClose, "]", 0},
		Item{EOF, "", 0},
	})
}

func TestObject(t *testing.T) {
	compare(t, lex("{\"a\": 1, \"b\": 2}"), []Item{
		Item{BraceOpen, "{", 0},
		Item{String, "a", 0},
		Item{Colon, ":", 0},
		Item{Number, "1", 0},
		Item{Comma, ",", 0},
		Item{String, "b", 0},
		Item{Colon, ":", 0},
		Item{Number, "2", 0},
		Item{BraceClose, "}", 0},
		Item{EOF, "", 0},
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
		Item{BraceOpen, "{", 0},
		Item{String, "foo", 0},
		Item{Colon, ":", 0},
		Item{Bool, "true", 0},
		Item{Comma, ",", 0},
		Item{String, "bar", 0},
		Item{Colon, ":", 0},
		Item{Bool, "false", 0},
		Item{Comma, ",", 0},
		Item{String, "zilch", 0},
		Item{Colon, ":", 0},
		Item{Null, "null", 0},
		Item{Comma, ",", 0},
		Item{String, "numbers", 0},
		Item{Colon, ":", 0},
		Item{BracketOpen, "[", 0},
		Item{Number, "1", 0},
		Item{Comma, ",", 0},
		Item{Number, "23", 0},
		Item{Comma, ",", 0},
		Item{Number, "4.56", 0},
		Item{Comma, ",", 0},
		Item{Number, "7.89", 0},
		Item{BracketClose, "]", 0},
		Item{Comma, ",", 0},
		Item{String, "bullshit", 0},
		Item{Colon, ":", 0},
		Item{BraceOpen, "{", 0},
		Item{String, "nothing", 0},
		Item{Colon, ":", 0},
		Item{String, "anything", 0},
		Item{BraceClose, "}", 0},
		Item{Error, "Unexpected symbol: !", 0},
	})
}

func compare(t *testing.T, reality, expectations []Item) {
	if len(reality) != len(expectations) {
		t.Errorf("Expected %d tokens, got %d", len(reality), len(expectations))
		return
	}
	for i, exp := range expectations {
		if exp.Token != reality[i].Token {
			t.Errorf("Expected an %s token, got %s", exp, reality[i])
			continue
		}
		if exp.Val != reality[i].Val {
			t.Errorf("Expected an %s token to hold value of %q, got %q", exp, exp.Val, reality[i].Val)
		}
	}
}

func lex(json string) []Item {
	l := New(json)
	go l.Run()

	items := []Item{}
	for {
		if item, ok := l.NextItem(); ok {
			items = append(items, item)
		} else {
			break
		}
	}

	return items
}
