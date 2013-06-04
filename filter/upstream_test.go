package filter

import (
	"github.com/fitstar/falcore"
	"math"
	"net/http"
	"testing"
	"time"
	// "fmt"
)

func TestUpstreamThrottle(t *testing.T) {
	// Start a test server
	sleepPipe := falcore.NewPipeline()
	sleepPipe.Upstream.PushBack(falcore.NewRequestFilter(func(req *falcore.Request) *http.Response {
		time.Sleep(time.Second)
		return falcore.StringResponse(req.HttpRequest, 200, nil, "OK")
	}))
	sleepSrv := falcore.NewServer(0, sleepPipe)
	go func() {
		sleepSrv.ListenAndServe()
	}()
	<-sleepSrv.AcceptReady

	// Build Upstream
	up := NewUpstream(NewUpstreamTransport("localhost", sleepSrv.Port(), 0, nil))
	// pipe := falcore.NewPipeline()
	// pipe.Upstream.PushBack(up)

	resCh := make(chan *http.Response, 10)
	var i int64 = 1
	for ; i < 12; i++ {
		start := time.Now()
		up.SetMaxConcurrent(i)

		for j := 0; j < 10; j++ {
			go func() {
				req, _ := http.NewRequest("GET", "/", nil)
				_, res := falcore.TestWithRequest(req, up, nil)
				resCh <- res
				// fmt.Println("OK")
			}()
		}
		for j := 0; j < 10; j++ {
			res := <-resCh
			if res.StatusCode != 200 {
				t.Fatalf("Error: %v", res)
			}
		}

		duration := time.Since(start)
		seconds := float64(duration) / float64(time.Second)
		goal := math.Ceil(10.0 / float64(i))
		// fmt.Println(i, "Time:", seconds, "Goal:", goal)
		if seconds < goal {
			t.Errorf("%v: Too short: %v < %v", i, seconds, goal)
		}
	}

}
