package falcore

import (
	"container/list"
	"log"
	"net/http"
	"reflect"
)

// Pipelines have an Upstream and Downstream list of filters.
// FilterRequest is called with the Request for all the items in
// the Upstream list in order UNTIL a Response is returned.
// Once a Response is returned iteration through the Upstream list
// is ended, and FilterResponse is called for ALL ResponseFilters
// in the Downstream list, in order.
//
// If no Response is returned from any of the Upstream filters,
// then execution of the Downstream is skipped and the server
// will return a default 404 response.
//
// The Upstream list may also contain instances of Router.
type Pipeline struct {
	Upstream   *list.List
	Downstream *list.List
}

func NewPipeline() (l *Pipeline) {
	l = new(Pipeline)
	l.Upstream = list.New()
	l.Downstream = list.New()
	return
}

// Pipelines are also RequestFilters... wacky eh?
func (p *Pipeline) FilterRequest(req *Request) *http.Response {
	return p.execute(req)
}

func (p *Pipeline) execute(req *Request) (res *http.Response) {
	for e := p.Upstream.Front(); e != nil && res == nil; e = e.Next() {
		switch filter := e.Value.(type) {
		case Router:
			t := reflect.TypeOf(filter)
			req.startPipelineStage(t.String())
			req.CurrentStage.Type = PipelineStageTypeRouter
			pipe := filter.SelectPipeline(req)
			req.finishPipelineStage()
			if pipe != nil {
				res = p.execFilter(req, pipe)
				if res != nil {
					break
				}
			}
		case RequestFilter:
			res = p.execFilter(req, filter)
			if res != nil {
				break
			}
		default:
			log.Printf("%v (%T) is not a RequestFilter\n", e.Value, e.Value)
			break
		}
	}

	if res != nil {
		p.down(req, res)
	}

	return
}

func (p *Pipeline) execFilter(req *Request, filter RequestFilter) *http.Response {
	if _, skipTracking := filter.(*Pipeline); !skipTracking {
		t := reflect.TypeOf(filter)
		req.startPipelineStage(t.String())
		req.CurrentStage.Type = PipelineStageTypeUpstream
		defer req.finishPipelineStage()
	}
	return filter.FilterRequest(req)
}

func (p *Pipeline) down(req *Request, res *http.Response) {
	for e := p.Downstream.Front(); e != nil; e = e.Next() {
		if filter, ok := e.Value.(ResponseFilter); ok {
			t := reflect.TypeOf(filter)
			req.startPipelineStage(t.String())
			req.CurrentStage.Type = PipelineStageTypeDownstream
			filter.FilterResponse(req, res)
			req.finishPipelineStage()
		} else {
			// TODO
			break
		}
	}
}
