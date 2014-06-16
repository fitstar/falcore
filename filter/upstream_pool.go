package filter

import (
	"github.com/fitstar/falcore"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type upstreamEntry struct {
	Upstream *Upstream
	Weight   int64
	Down     bool
}

// An UpstreamPool is a collection of Upstreams which are considered
// functionally equivalent.  The pool will balance traffic across Upstreams
// based on their relative weights.
type UpstreamPool struct {
	Name         string              // for logging, etc
	pool         []*upstreamEntry    // the upstreams
	weightSum    int64               // the sum of all weights (for balancer)
	nextUpstream chan *upstreamEntry // the next upstream to use
	kick         chan int            // let the nextServer selector know the config has changed
	shutdown     chan int            // shut it down
	mutex        *sync.RWMutex       // lock around config changes
	pinger       *time.Ticker        // timer for pinging upstreams
}

// The config consists of a map of the servers in the pool in the format host_or_ip:port
// where port is optional and defaults to 80.  The map value is an int with the weight
// only 0 and 1 are supported weights (0 disables a server and 1 enables it)
func NewUpstreamPool(name string) *UpstreamPool {
	up := new(UpstreamPool)
	up.Name = name
	up.nextUpstream = make(chan *upstreamEntry)
	up.kick = make(chan int)
	up.mutex = new(sync.RWMutex)
	up.shutdown = make(chan int)
	up.pinger = time.NewTicker(3 * time.Second) // 3s
	up.pool = make([]*upstreamEntry, 0, 10)
	up.weightSum = 0

	go up.nextServer()
	go up.pingUpstreams()

	return up
}

// Logs the current status of the pool
func (up *UpstreamPool) LogStatus() {
	up.mutex.RLock()
	for _, ue := range up.pool {
		downStatus := "UP"
		if ue.Down {
			downStatus = "DOWN"
		}

		falcore.Info("Upstream %v: %v:%v\t%v\t%v", up.Name, ue.Upstream.Transport.host, ue.Upstream.Transport.port, ue.Weight, downStatus)
	}
	up.mutex.RUnlock()
}

func (up *UpstreamPool) FilterRequest(req *falcore.Request) (res *http.Response) {
	ue := <-up.nextUpstream
	// If we didn't get an entry, return an error and log
	if ue == nil {
		falcore.Error("%s [%s] Upstream Pool error: No upstreams available", req.ID, up.Name)
		res = falcore.StringResponse(req.HttpRequest, 502, nil, "Bad Gateway\n")
		req.CurrentStage.Status = 2 // Fail
		return
	}

	res = ue.Upstream.FilterRequest(req)
	if req.CurrentStage.Status == 2 {
		// this gets set by the upstream for errors
		// so mark this upstream as down
		up.downupstreamEntry(ue, true)
		up.LogStatus()
	}
	return
}

// Updates balancing properties
// IMPORTANT: expects to have write mutex
func (up *UpstreamPool) rebalance() {
	var sum int64 = 0
	for _, e := range up.pool {
		if !e.Down {
			sum += e.Weight
		}
	}
	up.weightSum = sum
}

// Add an upstream
func (up *UpstreamPool) AddUpstream(u *Upstream, weight int64) {
	up.mutex.Lock()
	newPool := make([]*upstreamEntry, len(up.pool)+1)
	newPool[0] = &upstreamEntry{u, weight, false}
	copy(newPool[1:], up.pool)
	up.pool = newPool
	up.rebalance()
	up.mutex.Unlock()
}

// Drop an upstream
func (up *UpstreamPool) RemoveUpstream(u *Upstream) {
	up.mutex.Lock()
	newPool := make([]*upstreamEntry, 0, len(up.pool)-1)
	for _, e := range up.pool {
		if e.Upstream != u {
			newPool = append(newPool, e)
		}
	}
	up.pool = newPool
	up.rebalance()
	up.mutex.Unlock()
}

// Re-weight an upstream
func (up *UpstreamPool) UpdateUpstream(u *Upstream, weight int64) {
	up.mutex.Lock()
	for _, e := range up.pool {
		if e.Upstream == u {
			e.Weight = weight
			break
		}
	}
	up.rebalance()
	up.mutex.Unlock()
}

// Down (or up) an upstream
func (up *UpstreamPool) DownUpstream(u *Upstream, isDown bool) {
	var ue *upstreamEntry = nil
	up.mutex.RLock()
	for _, e := range up.pool {
		if e.Upstream == u {
			ue = e
			break
		}
	}
	up.mutex.RUnlock()
	if ue != nil {
		up.downupstreamEntry(ue, isDown)
	}
}

// Faster version of DownUpstream if we alread have the entry object
// returns true if the value was changed
func (up *UpstreamPool) downupstreamEntry(ue *upstreamEntry, isDown bool) bool {
	// Don't bother if it's the same value
	up.mutex.RLock()
	change := ue.Down != isDown
	up.mutex.RUnlock()

	if change {
		up.mutex.Lock()
		ue.Down = isDown
		up.rebalance()
		up.mutex.Unlock()
	}
	return change
}

// This should only be called if the upstream pool is no longer active or this may deadlock
// Calling this method twice will result in a panic
func (up *UpstreamPool) Shutdown() {
	// ping and nextServer
	close(up.shutdown)
}

func (up *UpstreamPool) nextServer() {
	defer close(up.nextUpstream)
	for {
		// Generate a random number [0, up.weightSum)
		// Choose an entry that matches that value
		up.mutex.RLock()
		var next *upstreamEntry = nil
		var goal = rand.Int63n(up.weightSum)
		for _, e := range up.pool {
			if !e.Down && e.Weight > goal {
				next = e
				break
			}
			goal -= e.Weight
		}
		up.mutex.RUnlock()

		select {
		// Bail out if we're shutting down
		case <-up.shutdown:
			return
		// Try to return the selected entry
		// next will be nil if none are available
		case up.nextUpstream <- next:
		// Escape and start over if the config changes
		case <-up.kick:
		}
	}
}

func (up *UpstreamPool) pingUpstreams() {
	pingable := true
	for pingable {
		select {
		case <-up.shutdown:
			return
		case <-up.pinger.C:
			gotone := false
			up.mutex.RLock()
			for i, ups := range up.pool {
				if ups.Upstream.PingPath != "" {
					go up.pingUpstream(ups, i)
					gotone = true
				}
			}
			up.mutex.RUnlock()
			if !gotone {
				pingable = false
			}
		}
	}
	falcore.Warn("Stopping ping for %v", up.Name)
}

func (up *UpstreamPool) pingUpstream(ups *upstreamEntry, index int) {
	isUp, ok := ups.Upstream.ping()
	if ok {
		if up.downupstreamEntry(ups, !isUp) {
			up.LogStatus()
		}
	}
}
