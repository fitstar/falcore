package filter

import (
	"github.com/fitstar/falcore"
	"net/http"
	"time"
)

// If Date header is non-existent in the response, this filter
// will automatically set it to the current date
type DateFilter struct {
}

func (f *DateFilter) FilterResponse(request *falcore.Request, res *http.Response) {
	request.CurrentStage.Status = 1 // Skipped (default)
	if res.Header.Get("Date") == "" {
		request.CurrentStage.Status = 0 // Success
		res.Header.Set("Date", time.Now().Format(time.RFC1123))
	}
}
