package filter

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
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

var ccsrv *falcore.Server

func init() {
	go func() {
		// falcore setup
		pipeline := falcore.NewPipeline()
		pipeline.Upstream.PushBack(falcore.NewRequestFilter(func(req *falcore.Request) *http.Response {
			for _, data := range ccserverData {
				if data.path == req.HttpRequest.URL.Path {
					header := make(http.Header)
					header.Set("Content-Type", data.mime)
					header.Set("Content-Encoding", data.encoding)
					return falcore.StringResponse(req.HttpRequest, 200, header, string(data.body))
				}
			}
			return falcore.StringResponse(req.HttpRequest, 404, nil, "Not Found")
		}))

		pipeline.Downstream.PushBack(NewCompressionFilter(nil))

		ccsrv = falcore.NewServer(0, pipeline)
		if err := ccsrv.ListenAndServe(); err != nil {
			panic("Could not start falcore")
		}
	}()
}

func ccport() int {
	for ccsrv.Port() == 0 {
		time.Sleep(1e7)
	}
	return ccsrv.Port()
}

var ccserverData = []struct {
	path     string
	mime     string
	encoding string
	body     []byte
}{
	{
		"/hello",
		"text/plain",
		"",
		[]byte("hello world"),
	},
	{
		"/hello.gz",
		"text/plain",
		"gzip",
		compress_gzip([]byte("hello world")),
	},
	{
		"/images/face.png",
		"image/png",
		"",
		readfile("../test/images/face.png"),
	},
}

var testData = []struct {
	name string
	// input
	path   string
	accept string
	// output
	encoding     string
	encoded_body []byte
}{
	{
		"no compression",
		"/hello",
		"",
		"",
		[]byte("hello world"),
	},
	{
		"gzip",
		"/hello",
		"gzip",
		"gzip",
		compress_gzip([]byte("hello world")),
	},
	{
		"deflate",
		"/hello",
		"deflate",
		"deflate",
		compress_deflate([]byte("hello world")),
	},
	{
		"preference",
		"/hello",
		"gzip, deflate",
		"gzip",
		compress_gzip([]byte("hello world")),
	},
	{
		"precompressed",
		"/hello.gz",
		"gzip",
		"gzip",
		compress_gzip([]byte("hello world")),
	},
	{
		"image",
		"/images/face.png",
		"gzip",
		"",
		readfile("../test/images/face.png"),
	},
}

func compress_gzip(body []byte) []byte {
	buf := new(bytes.Buffer)
	comp := gzip.NewWriter(buf)
	comp.Write(body)
	comp.Close()
	b := buf.Bytes()
	// fmt.Println(b)
	return b
}

func compress_deflate(body []byte) []byte {
	buf := new(bytes.Buffer)
	comp, err := flate.NewWriter(buf, -1)
	if err != nil {
		panic(fmt.Sprintf("Error using compress/flate.NewWriter() %v", err))
	}
	comp.Write(body)
	comp.Close()
	b := buf.Bytes()
	// fmt.Println(b)
	return b
}

func readfile(path string) []byte {
	if data, err := ioutil.ReadFile(path); err == nil {
		return data
	} else {
		panic(fmt.Sprintf("Error reading file %v: %v", path, err))
	}
	return nil
}

func ctget(p string, accept string) (r *http.Response, err error) {
	var conn net.Conn
	if conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%v", ccport())); err == nil {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%v", path.Join(fmt.Sprintf("localhost:%v/", ccport()), p)), nil)
		req.Header.Set("Accept-Encoding", accept)
		req.Write(conn)
		buf := bufio.NewReader(conn)
		r, err = http.ReadResponse(buf, req)
	}
	return
}

func TestCompressionFilter(t *testing.T) {
	// select{}
	for _, test := range testData {
		if res, err := ctget(test.path, test.accept); err == nil {
			bodyBuf := new(bytes.Buffer)
			io.Copy(bodyBuf, res.Body)
			body := bodyBuf.Bytes()
			var isChunked bool = res.TransferEncoding != nil && len(res.TransferEncoding) > 0 && res.TransferEncoding[0] == "chunked"
			if !isChunked && res.ContentLength != int64(len(body)) {
				t.Errorf("%v Invalid content length.  %v != %v", test.name, res.ContentLength, len(body))
			}
			if enc := res.Header.Get("Content-Encoding"); enc != test.encoding {
				t.Errorf("%v Header mismatch. Expecting: %v Got: %v", test.name, test.encoding, enc)
			}
			if !bytes.Equal(body, test.encoded_body) {
				t.Errorf("%v Body mismatch.\n\tExpecting:\n\t%v\n\tGot:\n\t%v", test.name, test.encoded_body, body)
			}
		} else {
			t.Errorf("%v HTTP Error %v", test.name, err)
		}
	}
}
