package responder

import (
	"github.com/fitstar/falcore"
	"github.com/fitstar/falcore/utils"
	"io"
	"net/http"
)

// Returns the write half of an io.Pipe.  The read half will be the Body of the response.
// Use this to stream a generated body without buffering first.  Don't forget to close the writer when finished.
// Writes are blocking until something Reads.  Best to use a separate goroutine for writing.
// Response will be Transfer-Encoding: chunked.
func PipeResponse(req *http.Request, status int, headers http.Header) (io.WriteCloser, *http.Response) {
	pR, pW := io.Pipe()
	return pW, falcore.SimpleResponse(req, status, headers, -1, pR)
}

// Returns the write half of an buffered pipe.  The read half will be the Body of the response.
// The buffered version is way faster if you're going to use lots of small writes.
// Use this to stream a generated body without buffering first.  Don't forget to close the writer when finished.
// Writes are blocking until something Reads if the buffer is full.  Best to use a separate goroutine for writing.
// Response will be Transfer-Encoding: chunked.
func BufferedPipeResponse(req *http.Request, status int, headers http.Header) (io.WriteCloser, *http.Response) {
	// get a buffer from the pool
	var buff *utils.RingBuffer
	select {
	case buff = <-pipeBufferPool:
		buff.Reset()
	default:
		buff = utils.NewRingBuffer(1024)
	}

	pR, pW := utils.NewBufferedPipe(buff)

	// return the buffer to the pool, leaky bucket style
	go func() {
		pR.CloseWait()
		select {
		case pipeBufferPool <- buff:
		default:
		}
	}()

	return pW, falcore.SimpleResponse(req, status, headers, -1, pR)
}

// 1024 buffers x 1024 bytes = 1MB
var pipeBufferPool = make(chan *utils.RingBuffer, 1024)
