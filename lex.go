package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/localhots/punk/lexer"
)

func main() {
	f, _ := os.Open("test.json")
	b, _ := ioutil.ReadAll(f)

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
