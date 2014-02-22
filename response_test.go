package falcore

import (
	"testing"
	"io"
	"io/ioutil"
	"net/http"
	"bytes"
	"./utils"
)

func BenchmarkPipeResponse(b *testing.B) {
	req, _ := http.NewRequest("GET", "/foo", nil)
	data := make([]byte, 1e7)
	for i := 0; i < b.N; i++ {
		wr, res := PipeResponse(req, 200, nil)
		go func(){
			io.Copy(wr, bytes.NewBuffer(data))
			wr.Close()
		}()
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}
}

func BenchmarkBufferedPipeResponse(b *testing.B) {
	req, _ := http.NewRequest("GET", "/foo", nil)
	data := make([]byte, 1e7)
	for i := 0; i < b.N; i++ {
		rd, wr := utils.NewBufferedPipe(utils.NewRingBuffer(1024))
		res := SimpleResponse(req, 200, nil, -1, rd)
		go func(){
			io.Copy(wr, bytes.NewBuffer(data))
			wr.Close()
		}()
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}
}