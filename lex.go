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

	lex := lexer.New("foo", string(b))
	go lex.Run()
	for {
		i := lex.NextItem()
		fmt.Println(i)
		if i.String() == "EOF" {
			break
		}
	}
}
