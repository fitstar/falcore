package routers

import (
	"github.com/fitstar/falcore"
	"container/list"
	"regexp"
)

// Interface for defining individual routes
type Route interface {
	// Returns the route's filter if there's a match.  nil if there isn't
	MatchString(str string) falcore.RequestFilter
}

// Will match any request.  Useful for fallthrough filters.
type MatchAnyRoute struct {
	Filter falcore.RequestFilter
}

func (r *MatchAnyRoute) MatchString(str string) falcore.RequestFilter {
	return r.Filter
}

// Will match based on a regular expression
type RegexpRoute struct {
	Match  *regexp.Regexp
	Filter falcore.RequestFilter
}

func (r *RegexpRoute) MatchString(str string) falcore.RequestFilter {
	if r.Match.MatchString(str) {
		return r.Filter
	}
	return nil
}

// Route requsts based on hostname
type HostRouter struct {
	hosts map[string]falcore.RequestFilter
}

// Generate a new HostRouter instance
func NewHostRouter() *HostRouter {
	r := new(HostRouter)
	r.hosts = make(map[string]falcore.RequestFilter)
	return r
}

// TODO: support for non-exact matches
func (r *HostRouter) AddMatch(host string, pipe falcore.RequestFilter) {
	r.hosts[host] = pipe
}

func (r *HostRouter) SelectPipeline(req *falcore.Request) (pipe falcore.RequestFilter) {
	return r.hosts[req.HttpRequest.Host]
}

// Route requests based on path
type PathRouter struct {
	Routes *list.List
}

// Generate a new instance of PathRouter
func NewPathRouter() *PathRouter {
	r := new(PathRouter)
	r.Routes = list.New()
	return r
}

func (r *PathRouter) AddRoute(route Route) {
	r.Routes.PushBack(route)
}

// convenience method for adding RegexpRoutes
func (r *PathRouter) AddMatch(match string, filter falcore.RequestFilter) (err error) {
	route := &RegexpRoute{Filter: filter}
	if route.Match, err = regexp.Compile(match); err == nil {
		r.Routes.PushBack(route)
	}
	return
}

// Will panic if r.Routes contains an object that isn't a Route
func (r *PathRouter) SelectPipeline(req *falcore.Request) (pipe falcore.RequestFilter) {
	var route Route
	for r := r.Routes.Front(); r != nil; r = r.Next() {
		route = r.Value.(Route)
		if f := route.MatchString(req.HttpRequest.URL.Path); f != nil {
			return f
		}
	}
	return nil
}
