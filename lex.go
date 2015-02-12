package main

import (
	"io/ioutil"
	"os"

	"github.com/localhots/punk/lexer"
)

func main() {
	f, _ := os.Open("test.json")
	b, _ := ioutil.ReadAll(f)

	lex := lexer.New("foo", string(b))
	lex.Run()
}
