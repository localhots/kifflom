package buffer

import (
	"bytes"
	"io"
	"unicode/utf8"
)

type (
	StreamBuffer struct {
		input io.Reader
	}
)

func NewStreamBuffer(input io.Reader) *StreamBuffer {
	return &StreamBuffer{
		input: input,
	}
}

func (b *StreamBuffer) Next() rune {
	var (
		buf bytes.Buffer
	)
	for {
		rbuf := make([]byte, 1)
		if n, err := b.input.Read(rbuf); n != 1 || err != nil {
			return 0
		}
		buf.Write(rbuf)
		if ok := utf8.FullRune(buf.Bytes()); ok {
			r, _ := utf8.DecodeRune(buf.Bytes())
			return r
		}
	}
	return 0
}
