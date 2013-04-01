package filter

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fitstar/falcore"
	"io"
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
					return falcore.StringResponse(req.HttpRequest, data.status, header, string(data.body))
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
	path   string
	status int
	etag   string
	body   []byte
}{
	{
		"/hello",
		200,
		"abc123",
		[]byte("hello world"),
	},
	{
		"/pre",
		304,
		"abc123",
		[]byte{},
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
}

func eget(p string, etag string) (r *http.Response, err error) {
	var conn net.Conn
	if conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%v", eport())); err == nil {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%v", path.Join(fmt.Sprintf("localhost:%v/", eport()), p)), nil)
		req.Header.Set("If-None-Match", etag)
		req.Write(conn)
		buf := bufio.NewReader(conn)
		r, err = http.ReadResponse(buf, req)
	}
	return
}

func TestEtagFilter(t *testing.T) {
	// select{}
	for _, test := range etestData {
		if res, err := eget(test.path, test.etag); err == nil {
			bodyBuf := new(bytes.Buffer)
			io.Copy(bodyBuf, res.Body)
			body := bodyBuf.Bytes()
			if st := res.StatusCode; st != test.status {
				t.Errorf("%v StatusCode mismatch. Expecting: %v Got: %v", test.name, test.status, st)
			}
			if !bytes.Equal(body, test.body) {
				t.Errorf("%v Body mismatch.\n\tExpecting:\n\t%v\n\tGot:\n\t%v", test.name, test.body, body)
			}
		} else {
			t.Errorf("%v HTTP Error %v", test.name, err)
		}
	}
}
