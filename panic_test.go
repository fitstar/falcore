package falcore

import (
	"fmt"
	"net"
	"net/http"
	"testing"
)

func TestFilterPanic(t *testing.T) {
	pipeline := NewPipeline()
	pipeline.Upstream.PushBack(NewRequestFilter(func(*Request) *http.Response { panic("this isn't supposed to happen") }))
	srv := NewServer(0, pipeline)
	defer srv.StopAccepting()
	go func() {
		srv.ListenAndServe()
	}()
	<-srv.AcceptReady

	var caught bool
	srv.PanicHandler = func(c net.Conn, err interface{}) {
		if c != nil && err != nil {
			caught = true
		}
	}
	http.Get(fmt.Sprintf("http://localhost:%d", srv.Port()))
	if !caught {
		t.Fatal("panic handler was not called")
	}
}
