package filter

import (
	"testing"
	"io/ioutil"
	"github.com/fitstar/falcore"
	"net/http"
)

var upstreamPoolTestData = []struct{
	name string
	weights []int64
	ratio float64 // % of As
}{
	{"simple", []int64{1,1}, 0.5},
	{"double", []int64{2,1}, 0.66},
	{"triple", []int64{3,1}, 0.75},
	{"big", []int64{200,100}, 0.66},
	{"single", []int64{1,0}, 1.0},
	{"single reverse", []int64{0,1}, 0.0},
}

func TestUpstreamPool(t *testing.T) {
	serverA, upstreamA := upstreamPoolTestServer("A")
	serverB, upstreamB := upstreamPoolTestServer("B")
	defer serverA.StopAccepting()
	defer serverB.StopAccepting()

	iterations := 1000
	for _, test := range upstreamPoolTestData {
		pool := NewUpstreamPool("TESTPOOL")
		pool.AddUpstream(upstreamA, test.weights[0])
		pool.AddUpstream(upstreamB, test.weights[1])
		aCount := 0
		for i := 0; i < iterations; i++ {
			req, _ := http.NewRequest("GET", "http://localhost/test", nil)
			_, res := falcore.TestWithRequest(req, pool, nil)
			rbody, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if rbody[0] == 'A' {
				aCount++
			}
		}
		percent := float64(aCount) / float64(iterations)
		if percent < test.ratio * 0.9 || percent > test.ratio * 1.1 {
			t.Errorf("[%v] Result distribution %0.4f is out of range of goal %0.4f", test.name, percent, test.ratio)
		}
	}
}

func upstreamPoolTestServer(res string)(*falcore.Server, *Upstream) {
	// Start a test server
	pipe := falcore.NewPipeline()
	pipe.Upstream.PushBack(falcore.NewRequestFilter(func(req *falcore.Request) *http.Response {
		return falcore.StringResponse(req.HttpRequest, 200, nil, res)
	}))
	srv := falcore.NewServer(0, pipe)
	go func() {
		srv.ListenAndServe()
	}()
	<-srv.AcceptReady
	return srv, NewUpstream(NewUpstreamTransport("localhost", srv.Port(), 0, nil))
}