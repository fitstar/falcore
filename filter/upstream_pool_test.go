package filter

import (
	"github.com/fitstar/falcore"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

var upstreamPoolTestData = []struct {
	name    string
	weights []int64
	ratio   float64 // % of As
	action  []string
}{
	{"simple", []int64{1, 1}, 0.5, nil},
	{"three", []int64{1, 1, 1}, 0.33, nil},
	{"double", []int64{2, 1}, 0.66, nil},
	{"triple", []int64{3, 1}, 0.75, nil},
	{"big", []int64{200, 100}, 0.66, nil},
	{"single", []int64{1, 0}, 1.0, nil},
	{"revers", []int64{0, 1}, 0.0, nil},
	{"downB", []int64{1, 1}, 1.0, []string{"", "D"}},
	{"downA", []int64{1, 1}, 0.001, []string{"D", ""}},
	{"remB", []int64{1, 1}, 1.0, []string{"", "R"}},
	{"remA", []int64{1, 1}, 0.0, []string{"R", ""}},
}

func TestUpstreamPool(t *testing.T) {
	iterations := 1000
	for _, test := range upstreamPoolTestData {
		// Setup test environment
		resChar := 'A'
		servers := make([]*falcore.Server, len(test.weights))
		upstreams := make([]*Upstream, len(test.weights))
		pool := NewUpstreamPool("TESTPOOL")
		for i, w := range test.weights {
			servers[i], upstreams[i] = upstreamPoolTestServer(string(resChar))
			upstreams[i].PingPath = "/"
			pool.AddUpstream(upstreams[i], w)
			resChar++
			if test.action != nil {
				switch test.action[i] {
				case "D":
					servers[i].StopAccepting()
				case "R":
					pool.RemoveUpstream(upstreams[i])
				}
			}
		}

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
		if percent < test.ratio*0.9 || percent > test.ratio*1.1 {
			t.Errorf("[%v] Result distribution %0.4f is out of range of goal %0.4f", test.name, percent, test.ratio)
		}

		// shutdown test servers (they might already have been shutdown)
		for _, s := range servers {
			go func() {
				defer func() {
					recover()
				}()
				s.StopAccepting()
			}()
		}
	}
}

func upstreamPoolTestServer(res string) (*falcore.Server, *Upstream) {
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
	return srv, NewUpstream(NewUpstreamTransport("localhost", srv.Port(), 100*time.Millisecond, nil))
}
