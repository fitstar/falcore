package falcore

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"testing"
)

var keepAliveTestData = []struct {
	name            string
	version         int
	useHeader       bool
	shouldKeepAlive bool
}{
	{"1.0", 0, true, true},
	{"1.0 no KA", 0, false, false},
	{"1.1", 1, true, true},
	{"1.1 no KA", 1, false, true},
}

func TestKeepAlive(t *testing.T) {
	// Startup a basic server and get the port
	pipeline := NewPipeline()
	srv := NewServer(0, pipeline)
	go func() {
		srv.ListenAndServe()
	}()
	<-srv.AcceptReady
	serverPort := srv.Port()

	// Connect and make some requests
	// Not using http.Client because transport does too many magics
	for _, test := range keepAliveTestData {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%v", serverPort))
		bconn := bufio.NewReader(conn)
		if err != nil {
			t.Fatal("Couldn't connect")
		}
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("GET", "/", nil)
			req.ProtoMinor = test.version
			req.Proto = fmt.Sprintf("HTTP/1.%v", test.version)
			if test.useHeader {
				req.Header.Set("Connection", "Keep-Alive")
			}
			if err = req.Write(conn); err != nil {
				if test.shouldKeepAlive {
					t.Error(fmt.Sprintf("[%v:%v] Couldn't write request: %v", test.name, i, err))
				}
				break
			}
			
			res, err := http.ReadResponse(bconn, req)
			if err != nil {
				if test.shouldKeepAlive {
					t.Error(fmt.Sprintf("[%v:%v] Couldn't read response: %v", test.name, i, err))
				}
				break
			} else if !test.shouldKeepAlive && i > 0 {
				t.Error(fmt.Sprintf("[%v:%v] Connection should be closed", test.name, i))
				break
			}
			res.Body.Close()
		}
		conn.Close()
	}

	// Clean up
	srv.StopAccepting()
}
