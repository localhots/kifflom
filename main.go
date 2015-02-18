package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/localhots/punk/buffer"
	"github.com/localhots/punk/parser"
)

func main() {
	var (
		sel     string
		verbose bool
	)
	flag.StringVar(&sel, "s", "", "Selector")
	flag.BoolVar(&verbose, "v", false, "Verbose parsing")
	flag.Parse()

	if len(sel) == 0 && !verbose {
		fmt.Println("No selectors given and parser is not verbose")
		os.Exit(1)
	}

	sels := strings.Split(sel, " ")
	if len(sel) == 0 {
		sels = []string{}
	}
	buf := buffer.NewStreamBuffer(os.Stdin)
	pars := parser.New(buf, sels)
	if verbose {
		pars.Debug()
	}
	res := pars.ParseStream()
	for {
		if m, ok := <-res; ok {
			fmt.Println(m.Sel, m.Val)
		} else {
			break
		}
	}
}
