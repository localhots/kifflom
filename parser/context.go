package parser

import (
	"strconv"
	"strings"
)

type (
	context struct {
		exps []expectation
	}

	// Context building block
	expectation struct {
		typ    expectationType
		greedy bool
		key    string // Object key
		index  int64  // Array index
	}

	// Type of expectation: object or array
	expectationType byte
)

const (
	unknown expectationType = iota
	object
	array
)

func (c *context) compare(c2 *context) bool {
	if len(c.exps) != len(c2.exps) {
		return false
	}

	for i, exp := range c.exps {
		exp2 := c2.exps[i]
		if exp.typ != exp2.typ {
			return false
		}
		if exp.greedy || exp2.greedy {
			continue
		}
		switch exp.typ {
		case array:
			if exp.index != exp2.index {
				return false
			}
		case object:
			if exp.key != exp2.key {
				return false
			}
		}
	}

	return true
}

func (c *context) push(typ expectationType) {
	c.exps = append(c.exps, expectation{
		typ: typ,
	})
}

func (c *context) pop() {
	if len(c.exps) == 0 {
		return
	}
	c.exps = c.exps[:len(c.exps)-1]
}

func (c *context) setKey(key string) {
	c.exps[len(c.exps)-1].key = key
}

func (c *context) setIndex(i int64) {
	c.exps[len(c.exps)-1].index = i
}

func parseSelectors(sels []string) map[string]*context {
	ctxs := map[string]*context{}
	for _, sel := range sels {
		ctxs[sel] = &context{
			exps: parseSelector(sel),
		}
	}
	return ctxs
}

func parseSelector(sel string) []expectation {
	tmp := strings.Replace(sel, ".", "/.", -1)
	tmp = strings.Replace(tmp, "#", "/#", -1)
	parts := strings.Split(tmp[1:], "/")

	exps := []expectation{}
	for _, part := range parts {
		c := expectation{}

		if len(part) < 2 {
			panic("Invalid selector: " + sel)
		} else if part[:1] == "." {
			c.typ = object
		} else if part[:1] == "#" {
			c.typ = array
		} else {
			panic("Invalid selector: " + sel)
		}

		if part[1:2] == "*" {
			c.greedy = true
			if len(part) > 2 {
				panic("Invalid selector: " + sel)
			}
		}

		if c.greedy {
		} else if c.typ == object {
			c.key = part[1:]
		} else if i, err := strconv.ParseInt(part[1:], 10, 64); err == nil {
			c.index = i
		} else {
			panic("Array index should be numeric: " + part)
		}

		exps = append(exps, c)
	}

	return exps
}

func (e expectationType) String() string {
	switch e {
	case array:
		return "Index"
	case object:
		return "Key"
	default:
		return "Unknown"
	}
}
