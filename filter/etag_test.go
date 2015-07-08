package filter

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fitstar/falcore"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"testing"
	"time"
)

var esrv *falcore.Server

func init() {
	go func() {
		// falcore setup
		pipeline := falcore.NewPipeline()
		pipeline.Upstream.PushBack(falcore.NewRequestFilter(func(req *falcore.Request) *http.Response {
			for _, data := range eserverData {
				if data.path == req.HttpRequest.URL.Path {
					header := make(http.Header)
					header.Set("Etag", data.etag)
					if data.chunked {
						buf := new(bytes.Buffer)
						buf.Write(data.body)
						res := falcore.SimpleResponse(req.HttpRequest, data.status, header, -1, ioutil.NopCloser(buf))
						res.TransferEncoding = []string{"chunked"}
						return res
					} else {
						return falcore.StringResponse(req.HttpRequest, data.status, header, string(data.body))
					}
				}
			}
			return falcore.StringResponse(req.HttpRequest, 404, nil, "Not Found")
		}))

		pipeline.Downstream.PushBack(new(EtagFilter))

		esrv = falcore.NewServer(0, pipeline)
		if err := esrv.ListenAndServe(); err != nil {
			panic("Could not start falcore")
		}
	}()
}

func eport() int {
	for esrv.Port() == 0 {
		time.Sleep(1e7)
	}
	return esrv.Port()
}

var eserverData = []struct {
	path    string
	status  int
	etag    string
	body    []byte
	chunked bool
}{
	{
		"/hello",
		200,
		"abc123",
		[]byte("hello world"),
		false,
	},
	{
		"/pre",
		304,
		"abc123",
		[]byte{},
		false,
	},
	{
		"/chunked",
		200,
		"abc123",
		[]byte{},
		true,
	},
}

var etestData = []struct {
	name string
	// input
	path string
	etag string
	// output
	status int
	body   []byte
}{
	{
		"no etag",
		"/hello",
		"",
		200,
		[]byte("hello world"),
	},
	{
		"match",
		"/hello",
		"abc123",
		304,
		[]byte{},
	},
	{
		"pre-filtered",
		"/pre",
		"abc123",
		304,
		[]byte{},
	},
	{
		"chunked",
		"/chunked",
		"abc123",
		304,
		[]byte{},
	},
}

func eget(name, p, etag string, t *testing.T) (r *http.Response, body []byte, err error) {
	var conn net.Conn
	if conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%v", eport())); err == nil {
		// Make request to test server
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%v", path.Join(fmt.Sprintf("localhost:%v/", eport()), p)), nil)
		req.Header.Set("If-None-Match", etag)
		req.Write(conn)

		// Read request back out
		debugBuf := new(bytes.Buffer)
		tee := io.TeeReader(conn, debugBuf)
		buf := bufio.NewReader(tee)

		r, err = http.ReadResponse(buf, req)

		// Read out body
		bodyBuf := new(bytes.Buffer)
		io.Copy(bodyBuf, r.Body)
		body = bodyBuf.Bytes()

		// Check for remaining crap
		t.Errorf("RESPONSE: %v", string(debugBuf.Bytes()))
		if l := buf.Buffered(); l > 0 {
			d, _ := buf.Peek(l)
			t.Errorf("%v Unexpected extra data (%v bytes) in buffer: %v", name, l, string(d))
		}

		// bodyBuf := new(bytes.Buffer)
		// io.Copy(bodyBuf, r.Body)
		// body = bodyBuf.Bytes()
	}
	return
}

func TestEtagFilter(t *testing.T) {
	// select{}
	for _, test := range etestData {
		if res, body, err := eget(test.name, test.path, test.etag, t); err == nil {
			if st := res.StatusCode; st != test.status {
				t.Errorf("%v StatusCode mismatch. Expecting: %v Got: %v", test.name, test.status, st)
			}
			if !bytes.Equal(body, test.body) {
				t.Errorf("%v Body mismatch.\n\tExpecting:\n\t%v\n\tGot:\n\t%v", test.name, test.body, body)
			}
			if test.status == 304 && res.TransferEncoding != nil {
				t.Errorf("%v Transfer encoding mismatch.\n\tExpecting:\n\t%v\n\tGot:\n\t%v", test.name, nil, res.TransferEncoding)
			}
		} else {
			t.Errorf("%v HTTP Error %v", test.name, err)
		}
	}
}
