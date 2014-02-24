package utils

import (
	"errors"
	"io"
	"sync"
)

type bufferedPipe struct {
	l     *sync.Mutex // Gates access to buffer
	buff  *RingBuffer // the buffer
	rwait *sync.Cond  // waiting reader
	wwait *sync.Cond  // waiting writer
	cwait *sync.Cond  // wait on close (for buffer reuse)
	rerr  error       // if reader is closed, the error
	werr  error       // if writer is closed, the error
}

type BufferedPipeReader struct {
	pipe *bufferedPipe
}

type BufferedPipeWriter struct {
	pipe *bufferedPipe
}

// ErrClosedPipe is the error used for read or write operations on a closed pipe.
var ErrClosedPipe = errors.New("io: read/write on closed pipe")

// This is largely borrowed from io.Pipe, but it uses
// an internal buffer to be faster and have fewer hits to the
// scheduler
func NewBufferedPipe(buff *RingBuffer) (*BufferedPipeReader, *BufferedPipeWriter) {
	l := new(sync.Mutex)
	pipe := &bufferedPipe{
		l,
		buff,
		sync.NewCond(l),
		sync.NewCond(l),
		sync.NewCond(l),
		nil,
		nil,
	}

	return &BufferedPipeReader{pipe}, &BufferedPipeWriter{pipe}
}

func (p *bufferedPipe) read(b []byte) (int, error) {
	p.l.Lock()
	defer p.l.Unlock()

	for {
		if p.rerr != nil {
			return 0, ErrClosedPipe
		}
		if !p.buff.Empty() {
			break
		}
		if p.werr != nil {
			return 0, p.werr
		}
		p.rwait.Wait()
	}
	n, err := p.buff.Read(b)

	if !p.buff.Full() {
		p.wwait.Broadcast()
	}
	return n, err
}

func (p *bufferedPipe) write(b []byte) (int, error) {
	p.l.Lock()
	defer p.l.Unlock()

	if p.werr != nil {
		return 0, ErrClosedPipe
	}

	for {
		if p.rerr != nil {
			return 0, p.rerr
		}
		if p.werr != nil {
			return 0, ErrClosedPipe
		}
		if !p.buff.Full() {
			break
		}
		p.wwait.Wait()
	}
	n, err := p.buff.Write(b)

	if !p.buff.Empty() {
		p.rwait.Broadcast()
	}
	return n, err
}

func (p *bufferedPipe) wclose(err error) {
	if err == nil {
		err = io.EOF
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.werr = err
	p.rwait.Broadcast()
	p.wwait.Broadcast()
	p.cwait.Broadcast()
}

func (p *bufferedPipe) rclose(err error) {
	if err == nil {
		err = ErrClosedPipe
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.rerr = err
	p.rwait.Broadcast()
	p.wwait.Broadcast()
	p.cwait.Broadcast()
}

func (p *bufferedPipe) closeWait() {
	p.l.Lock()
	defer p.l.Unlock()
	for {
		if p.werr != nil && p.rerr != nil {
			return
		}
		p.cwait.Wait()
	}
}

func (r *BufferedPipeReader) Read(p []byte) (int, error) {
	return r.pipe.read(p)
}

func (r *BufferedPipeReader) Close() error {
	return r.CloseWithError(nil)
}

func (r *BufferedPipeReader) CloseWithError(err error) error {
	r.pipe.rclose(err)
	return nil
}

// Wait for both ends of the buffer to close
func (r *BufferedPipeReader) CloseWait() {
	r.pipe.closeWait()
}

func (w *BufferedPipeWriter) Write(p []byte) (int, error) {
	return w.pipe.write(p)
}

func (w *BufferedPipeWriter) Close() error {
	return w.CloseWithError(nil)
}

func (w *BufferedPipeWriter) CloseWithError(err error) error {
	w.pipe.wclose(err)
	return nil
}

// Wait for both ends of the buffer to close
func (w *BufferedPipeWriter) CloseWait() {
	w.pipe.closeWait()
}
