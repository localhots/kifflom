package main

import (
	"io/ioutil"
	"os"

	"github.com/kr/pretty"
	"github.com/localhots/punk/parser"
)

func main() {
	f, _ := os.Open("test.json")
	b, _ := ioutil.ReadAll(f)

	p := parser.New(b, "/bananas/[1]/weight")
	pretty.Println(p)
	p.Parse()
}
