package buffer

import (
	"bytes"
	"unicode/utf8"
)

type (
	DataBuffer struct {
		input []byte
		size  uint64
		pos   uint64
	}
)

func NewDataBuffer(input []byte) *DataBuffer {
	return &DataBuffer{
		input: input,
		size:  uint64(len(input)),
	}
}

func (b *DataBuffer) Next() rune {
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
