package lexer

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/localhots/punk/buffer"
)

func BenchmarkRun(t *testing.B) {
	f, _ := os.Open("big1.json")
	b, _ := ioutil.ReadAll(f)
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		lex(b)
	}
}

func BenchmarkStandardJSON(t *testing.B) {
	f, _ := os.Open("big1.json")
	b, _ := ioutil.ReadAll(f)
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		var res interface{}
		json.Unmarshal(b, &res)
	}
}

func TestEmpty(t *testing.T) {
	compare(t, lex([]byte("")), []Item{
		{EOF, "\x00", 0, 0},
	})
}

func TestNull(t *testing.T) {
	compare(t, lex([]byte("null")), []Item{
		{Null, "null", 0, 0},
		{EOF, "\x00", 0, 0},
	})
}

func TesBool(t *testing.T) {
	compare(t, lex([]byte("true")), []Item{
		{Bool, "true", 0, 0},
		{EOF, "\x00", 0, 0},
	})
	compare(t, lex([]byte("false")), []Item{
		{Bool, "false", 0, 0},
		{EOF, "\x00", 0, 0},
	})
}

func TestString(t *testing.T) {
	compare(t, lex([]byte(`"foo"`)), []Item{
		{String, "foo", 0, 0},
		{EOF, "\x00", 0, 0},
	})
}

func TestNumber(t *testing.T) {
	compare(t, lex([]byte("123")), []Item{
		{Number, "123", 0, 0},
		{EOF, "\x00", 0, 0},
	})
	compare(t, lex([]byte("123.456")), []Item{
		{Number, "123.456", 0, 0},
		{EOF, "\x00", 0, 0},
	})
	compare(t, lex([]byte("123.456.789")), []Item{
		{Error, `Invalid number: "123.456.789"`, 0, 0},
	})
	compare(t, lex([]byte("123.")), []Item{
		{Error, `Invalid number: "123."`, 0, 0},
	})
}

func TestArray(t *testing.T) {
	compare(t, lex([]byte(`[1, "2", 3]`)), []Item{
		{BracketOpen, "[", 0, 0},
		{Number, "1", 0, 0},
		{Comma, ",", 0, 0},
		{String, "2", 0, 0},
		{Comma, ",", 0, 0},
		{Number, "3", 0, 0},
		{BracketClose, "]", 0, 0},
		{EOF, "\x00", 0, 0},
	})
}

func TestObject(t *testing.T) {
	compare(t, lex([]byte(`{"a": 1, "b": 2}`)), []Item{
		{BraceOpen, "{", 0, 0},
		{String, "a", 0, 0},
		{Colon, ":", 0, 0},
		{Number, "1", 0, 0},
		{Comma, ",", 0, 0},
		{String, "b", 0, 0},
		{Colon, ":", 0, 0},
		{Number, "2", 0, 0},
		{BraceClose, "}", 0, 0},
		{EOF, "\x00", 0, 0},
	})
}

// Yay!
func TestEverything(t *testing.T) {
	input := []byte(`{
	    "foo": true,
	    "bar": false,
	    "zilch": null,
	    "numbers": [1, 23, 4.56, 7.89],
	    "bullshit": {
	        "nothing": "anything"
	    }!
	}`)

	compare(t, lex(input), []Item{
		{BraceOpen, "{", 0, 0},
		{String, "foo", 0, 0},
		{Colon, ":", 0, 0},
		{Bool, "true", 0, 0},
		{Comma, ",", 0, 0},
		{String, "bar", 0, 0},
		{Colon, ":", 0, 0},
		{Bool, "false", 0, 0},
		{Comma, ",", 0, 0},
		{String, "zilch", 0, 0},
		{Colon, ":", 0, 0},
		{Null, "null", 0, 0},
		{Comma, ",", 0, 0},
		{String, "numbers", 0, 0},
		{Colon, ":", 0, 0},
		{BracketOpen, "[", 0, 0},
		{Number, "1", 0, 0},
		{Comma, ",", 0, 0},
		{Number, "23", 0, 0},
		{Comma, ",", 0, 0},
		{Number, "4.56", 0, 0},
		{Comma, ",", 0, 0},
		{Number, "7.89", 0, 0},
		{BracketClose, "]", 0, 0},
		{Comma, ",", 0, 0},
		{String, "bullshit", 0, 0},
		{Colon, ":", 0, 0},
		{BraceOpen, "{", 0, 0},
		{String, "nothing", 0, 0},
		{Colon, ":", 0, 0},
		{String, "anything", 0, 0},
		{BraceClose, "}", 0, 0},
		{Error, "Unexpected symbol: '!'", 0, 0},
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

func lex(b []byte) []Item {
	buf := buffer.NewBytesBuffer(b)
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
