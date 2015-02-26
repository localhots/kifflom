package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/davecheney/profile"
	"github.com/localhots/kifflom/buffer"
	"github.com/localhots/kifflom/parser"
)

func main() {
	var (
		sel     string
		verbose bool
		prof    string
	)

	flag.StringVar(&sel, "s", "", "Selector")
	flag.StringVar(&prof, "prof", "", "Performance profiling output")
	flag.BoolVar(&verbose, "v", false, "Verbose parsing")
	flag.Parse()

	if len(sel) == 0 && !verbose {
		fmt.Println("No selectors given and parser is not verbose")
		os.Exit(1)
	}

	if prof != "" {
		defer profile.Start(&profile.Config{
			CPUProfile:  true,
			ProfilePath: prof,
		}).Stop()
	}

	sels := strings.Split(sel, " ")
	if len(sel) == 0 {
		sels = []string{}
	}

	buf := buffer.NewReaderBuffer(os.Stdin)
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
