package bufpool

import (
	"bytes"
	"sync"
)

var bufferPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func PutBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}

type Buffer struct {
	*bytes.Buffer
}

func Get() Buffer {
	return Buffer{
		bufferPool.Get().(*bytes.Buffer),
	}
}

func (c Buffer) Close() error {
	bufferPool.Put(c.Buffer)
	return nil
}
