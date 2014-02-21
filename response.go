package falcore

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Generate an http.Response using the basic fields
func SimpleResponse(req *http.Request, status int, headers http.Header, contentLength int64, body io.Reader) *http.Response {
	res := new(http.Response)
	res.StatusCode = status
	res.ProtoMajor = 1
	res.ProtoMinor = 1
	res.ContentLength = contentLength
	res.Request = req
	res.Header = make(map[string][]string)
	if body_rdr, ok := body.(io.ReadCloser); ok {
		res.Body = body_rdr
	} else if body != nil {
		res.Body = ioutil.NopCloser(body)
	}
	if headers != nil {
		res.Header = headers
	}
	return res
}

// Like SimpleResponse but uses a []byte for the body.
func ByteResponse(req *http.Request, status int, headers http.Header, body []byte) *http.Response {
	return SimpleResponse(req, status, headers, int64(len(body)), bytes.NewBuffer(body))
}

// Like StringResponse but uses a string for the body.
func StringResponse(req *http.Request, status int, headers http.Header, body string) *http.Response {
	return SimpleResponse(req, status, headers, int64(len(body)), strings.NewReader(body))
}

// Returns the write half of an io.Pipe.  The read half will be the Body of the response.
// Use this to stream a generated body without buffering first.  Don't forget to close the writer when finished.
// Writes are blocking until something Reads.  Best to use a separate goroutine for writing.
// Response will be Transfer-Encoding: chunked.
func PipeResponse(req *http.Request, status int, headers http.Header) (io.WriteCloser, *http.Response) {
	pR, pW := io.Pipe()
	return pW, SimpleResponse(req, status, headers, -1, pR)
}
