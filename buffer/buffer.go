package buffer

import (
	"bufio"
	"bytes"
	"io"
)

type (
	Buffer struct {
		input *bufio.Reader
		ready chan rune
	}
)

func NewBytesBuffer(b []byte) *Buffer {
	return NewReaderBuffer(bytes.NewReader(b))
}

func NewReaderBuffer(rd io.Reader) *Buffer {
	return New(bufio.NewReader(rd))
}

func New(input *bufio.Reader) *Buffer {
	b := &Buffer{
		input: input,
		ready: make(chan rune, 100),
	}
	go b.stream()
	return b
}

func (b *Buffer) Next() rune {
	if next, ok := <-b.ready; ok {
		return next
	} else {
		return 0
	}
}

func (b *Buffer) stream() {
	defer close(b.ready)
	for {
		if r, _, err := b.input.ReadRune(); err == nil {
			b.ready <- r
		} else {
			return
		}
	}
}
