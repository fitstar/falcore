package falcore

import (
	"bufio"
	"io"
)

// A leaky bucket buffer pool for bufio.Writers
// Dramatically reduces garbage when you have lots of short lived
// http connections.
type WriteBufferPool struct {
	// size of buffer when creating new ones
	bufSize int
	// the actual pool of buffers ready for reuse
	pool chan *WriteBufferPoolEntry
}

// This is what's stored in the buffer.  It allows
// for the underlying io.Writer to be changed out
// inside a bufio.Writer.  This is required for reuse.
type WriteBufferPoolEntry struct {
	Br     *bufio.Writer
	source io.Writer
}

// make bufferPoolEntry a passthrough io.Writer
func (bpe *WriteBufferPoolEntry) Write(p []byte) (n int, err error) {
	return bpe.source.Write(p)
}

func NewWriteBufferPool(poolSize, bufferSize int) *WriteBufferPool {
	return &WriteBufferPool{
		bufSize: bufferSize,
		pool:    make(chan *WriteBufferPoolEntry, poolSize),
	}
}

// Take a buffer from the pool and set
// it up to read from r
func (p *WriteBufferPool) Take(r io.Writer) (bpe *WriteBufferPoolEntry) {
	select {
	case bpe = <-p.pool:
		bpe.source = r
	default:
		// none available.  create a new one
		bpe = &WriteBufferPoolEntry{nil, r}
		bpe.Br = bufio.NewWriterSize(bpe, p.bufSize)
	}
	return
}

// Return a buffer to the pool
func (p *WriteBufferPool) Give(bpe *WriteBufferPoolEntry) {
	if bpe.Br.Buffered() > 0 {
		return
	}
	if err := bpe.Br.Flush(); err != nil {
		return
	}
	select {
	case p.pool <- bpe: // return to pool
	default: // discard
	}
}
