package falcore

import (
	// "bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
)

var contentLengthTestData = []struct {
	path                  string
	body                  []byte
	resContentLength      int64 // what to send in the response
	expectedContentLength int64 // what the client should expect
	chunked               bool
}{
	{"/basic", []byte("ABC"), 3, 3, false},
	{"/chunked", []byte("ABC"), -1, -1, true},
	{"/zero", []byte(""), 0, 0, false},
	{"/unset", []byte("ABC"), 0, -1, true},
	{"/nil_body", nil, 0, 0, false},
}

func TestContentLength(t *testing.T) {
	// Startup a basic server and get the port
	pipeline := NewPipeline()
	srv := NewServer(0, pipeline)
	pipeline.Upstream.PushBack(NewRequestFilter(func(req *Request) *http.Response {
		for _, entry := range contentLengthTestData {
			if entry.path == req.HttpRequest.URL.Path {
				var body io.Reader
				if entry.body != nil {
					body = bytes.NewBuffer(entry.body)
				}
				return SimpleResponse(req.HttpRequest, 200, nil, entry.resContentLength, body)
			}
		}
		panic("Thing not found")
	}))
	go func() {
		srv.ListenAndServe()
	}()
	<-srv.AcceptReady
	serverPort := srv.Port()

	// Connect and make some requests
	c := new(http.Client)
	for _, test := range contentLengthTestData {
		res, err := c.Get(fmt.Sprintf("http://localhost:%v%v", serverPort, test.path))
		if err != nil || res == nil {
			t.Fatal(fmt.Sprintf("%v Couldn't get req: %v", test.path, err))
		}

		var isChunked bool = res.TransferEncoding != nil && len(res.TransferEncoding) > 0 && res.TransferEncoding[0] == "chunked"
		if test.chunked {
			if !isChunked {
				t.Errorf("%v Expected a chunked response.  Didn't get one.  Content-Length: %v", test.path, res.ContentLength)
			}
		} else {
			if isChunked {
				t.Errorf("%v Response is chunked.  Expected a content length", test.path)
			}
			if res.ContentLength != test.expectedContentLength {
				t.Errorf("%v Incorrect content length. Expected: %v Got: %v", test.path, test.expectedContentLength, res.ContentLength)
			}
		}

		bodyBuf := new(bytes.Buffer)
		io.Copy(bodyBuf, res.Body)
		body := bodyBuf.Bytes()
		if !bytes.Equal(body, test.body) {
			t.Errorf("%v Body mismatch.\n\tExpecting:\n\t%v\n\tGot:\n\t%v", test.path, test.body, body)
		}

		res.Body.Close()
	}

	// Clean up
	srv.StopAccepting()
}
