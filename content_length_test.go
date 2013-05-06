package falcore

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
)

var contentLengthTestData = []struct {
	method                string
	path                  string
	body                  []byte
	resContentLength      int64 // what to send in the response
	expectedContentLength int64 // what the client should expect
	chunked               bool
}{
	{"GET", "/basic", []byte("ABC"), 3, 3, false},
	{"GET", "/chunked", []byte("ABC"), -1, -1, true},
	{"GET", "/zero", []byte(""), 0, 0, false},
	{"GET", "/unset", []byte("ABC"), 0, -1, true},
	{"GET", "/nil_body", nil, 0, 0, false},
	{"HEAD", "/basic", []byte("ABC"), 3, 3, false},
	{"HEAD", "/chunked", []byte("ABC"), -1, -1, false},
	{"HEAD", "/unset", []byte("ABC"), 0, -1, false},
	{"HEAD", "/zero", []byte(""), 0, 0, false},
	{"HEAD", "/nil_body", nil, 0, 0, false},
	{"HEAD", "/unset_nil", nil, -1, -1, false},
	{"HEAD", "/length_nil", nil, 10, 10, false},
}

func TestContentLength(t *testing.T) {
	// Startup a basic server and get the port
	pipeline := NewPipeline()
	srv := NewServer(0, pipeline)
	pipeline.Upstream.PushBack(NewRequestFilter(func(req *Request) *http.Response {
		for _, entry := range contentLengthTestData {
			if entry.method == req.HttpRequest.Method && entry.path == req.HttpRequest.URL.Path {
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
		var res *http.Response
		var err error
		if test.method == "GET" {
			res, err = c.Get(fmt.Sprintf("http://localhost:%v%v", serverPort, test.path))
		} else {
			res, err = c.Head(fmt.Sprintf("http://localhost:%v%v", serverPort, test.path))
		}
		if err != nil || res == nil {
			t.Fatal(fmt.Sprintf("Couldn't get req: %v", err))
		}

		var isChunked bool = res.TransferEncoding != nil && len(res.TransferEncoding) > 0 && res.TransferEncoding[0] == "chunked"
		if test.chunked {
			if !isChunked {
				t.Errorf("%s %s Expected a chunked response.  Didn't get one.  Content-Length: %v", test.method, test.path, res.ContentLength)
			}
		} else {
			if isChunked {
				t.Errorf("%s %s Response is chunked.  Expected a content length", test.method, test.path)
			}
			if res.ContentLength != test.expectedContentLength {
				t.Errorf("%s %s Incorrect content length. Expected: %v Got: %v", test.method, test.path, test.expectedContentLength, res.ContentLength)
			}
		}

		if test.method == "GET" {
			bodyBuf := new(bytes.Buffer)
			io.Copy(bodyBuf, res.Body)
			body := bodyBuf.Bytes()
			if !bytes.Equal(body, test.body) {
				t.Errorf("%v Body mismatch.\n\tExpecting:\n\t%v\n\tGot:\n\t%v", test.path, test.body, body)
			}
		}

		res.Body.Close()
	}

	// Clean up
	srv.StopAccepting()
}
