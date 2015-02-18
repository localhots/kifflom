package main

import (
	"os"

	"github.com/kr/pretty"
	"github.com/localhots/punk/buffer"
	"github.com/localhots/punk/parser"
)

func main() {
	f, _ := os.Open("test.json")
	// b, _ := ioutil.ReadAll(f)
	buf := buffer.NewStreamBuffer(f)

	p := parser.New(buf, []string{
		"/prices/*",
		"/bananas/[*]/weight",
	})
	res := p.Parse()
	pretty.Println(res)
}
