package falcore

// Interface for defining Routers. Routers may be added to the Pipeline.Upstream,
// This interface may be used to choose between many mutually exclusive Filters.
// The Router's SelectPipeline method will be called and if it returns a Filter,
// the Filter's FilterRequest method will be called.  If SelectPipeline returns
// nil, the next stage in the Pipeline will be executed.
//
// In addition, a Pipeline is itself a RequestFilter so SelectPipeline may
// return a Pipeline as well.  This allows branching of the Pipeline flow.
type Router interface {
	// Returns a Filter to be executed or nil if one can't be found.
	SelectPipeline(req *Request) (pipe RequestFilter)
}

// Generate a new Router instance using f for SelectPipeline
func NewRouter(f func(req *Request) (pipe RequestFilter)) Router {
	return genericRouter(f)
}

type genericRouter func(req *Request) (pipe RequestFilter)

func (f genericRouter) SelectPipeline(req *Request) (pipe RequestFilter) {
	return f(req)
}
