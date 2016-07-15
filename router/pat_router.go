package router

import (
	"github.com/fitstar/falcore"
)

type PatternRouter struct {
	contextKey string
	patterns   []*Pattern
	routes     []falcore.RequestFilter
}

func NewPatternRouter(contextKey string) *PatternRouter {
	return &PatternRouter{
		contextKey: contextKey,
		patterns:   make([]*Pattern, 0),
		routes:     make([]falcore.RequestFilter, 0),
	}
}

func (r *PatternRouter) AddRoute(pattern string, route falcore.RequestFilter) error {
	if pat, err := ParsePattern(pattern); err == nil {
		r.patterns = append(r.patterns, pat)
		r.routes = append(r.routes, route)
		return nil
	} else {
		return err
	}
}

func (r *PatternRouter) SelectPipeline(req *falcore.Request) falcore.RequestFilter {
	for i, pat := range r.patterns {
		if ok, captures := pat.match(req.HttpRequest.URL.Path); ok {
			req.Context[r.contextKey] = captures
			return r.routes[i]
		}
	}
	return nil
}
