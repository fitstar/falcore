package falcore

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Generate an http.Response using the basic fields
func SimpleResponse(req *http.Request, status int, headers http.Header, contentLength int64, body io.Reader) *http.Response {
	res := new(http.Response)
	res.StatusCode = status
	res.ProtoMajor = 1
	res.ProtoMinor = 1
	res.ContentLength = contentLength
	res.Request = req
	res.Header = make(map[string][]string)
	if body_rdr, ok := body.(io.ReadCloser); ok {
		res.Body = body_rdr
	} else if body != nil {
		res.Body = ioutil.NopCloser(body)
	}
	if headers != nil {
		res.Header = headers
	}
	return res
}

// Like SimpleResponse but uses a []byte for the body.
func ByteResponse(req *http.Request, status int, headers http.Header, body []byte) *http.Response {
	return SimpleResponse(req, status, headers, int64(len(body)), bytes.NewBuffer(body))
}

// Like StringResponse but uses a string for the body.
func StringResponse(req *http.Request, status int, headers http.Header, body string) *http.Response {
	return SimpleResponse(req, status, headers, int64(len(body)), strings.NewReader(body))
}

// A 302 redirect response
func RedirectResponse(req *http.Request, url string) *http.Response {
	h := make(http.Header)
	h.Set("Location", url)
	return SimpleResponse(req, 302, h, 0, nil)
}

// Generate an http.Response by json encoding body using
// the standard library's json.Encoder.  error will be nil
// unless json encoding fails.
func JSONResponse(req *http.Request, status int, headers http.Header, body interface{}) (*http.Response, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, err
	}

	if headers == nil {
		headers = make(http.Header)
	}
	if headers.Get("Content-Type") == "" {
		headers.Set("Content-Type", "application/json")
	}

	return SimpleResponse(req, status, headers, int64(buf.Len()), buf), nil
}
