package buffer

import (
	"bytes"
	"unicode/utf8"
)

type (
	BytesBuffer struct {
		input []byte
		size  uint64
		pos   uint64
	}
)

func NewBytesBuffer(input []byte) *BytesBuffer {
	return &BytesBuffer{
		input: input,
		size:  uint64(len(input)),
	}
}

func (b *BytesBuffer) Next() rune {
	var buf bytes.Buffer
	for b.pos < b.size-1 {
		buf.WriteByte(b.input[b.pos])
		b.pos++
		if ok := utf8.FullRune(buf.Bytes()); ok {
			r, _ := utf8.DecodeRune(buf.Bytes())
			return r
		}
	}
	return 0
}
