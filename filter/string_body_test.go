package filter

import (
	"bytes"
	"github.com/fitstar/falcore"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestStringBody(t *testing.T) {
	expected := []byte("HOT HOT HOT!!!")
	tmp, _ := http.NewRequest("POST", "/hello", bytes.NewReader(expected))
	tmp.Header.Set("Content-Type", "text/plain")
	tmp.ContentLength = int64(len(expected))

	sbf := NewStringBodyFilter()
	req, _ := falcore.TestWithRequest(tmp, sbf, nil)

	if sb, ok := req.HttpRequest.Body.(*StringBody); ok {
		readin, _ := ioutil.ReadAll(sb)
		sb.Close()
		if bytes.Compare(readin, expected) != 0 {
			t.Errorf("Body string not read %q expected %q", readin, expected)
		}
	} else {
		t.Errorf("Body not replaced with StringBody")
	}

	if req.CurrentStage.Status != 0 {
		t.Errorf("SBF failed to parse POST with status %d", req.CurrentStage.Status)
	}

	var body []byte = make([]byte, 100)
	l, _ := req.HttpRequest.Body.Read(body)
	if bytes.Compare(body[0:l], expected) != 0 {
		t.Errorf("Failed to read the right bytes %q expected %q", body, expected)

	}

	l, _ = req.HttpRequest.Body.Read(body)
	if l != 0 {
		t.Errorf("Should have read zero!")
	}

	// Close resets the buffer
	req.HttpRequest.Body.Close()

	l, _ = req.HttpRequest.Body.Read(body)
	if bytes.Compare(body[0:l], expected) != 0 {
		t.Errorf("Failed to read the right bytes after calling Close %q expected %q", body, expected)

	}

}

func BenchmarkStringBody(b *testing.B) {
	b.StopTimer()
	expected := []byte("test=123456&test2=987654&test3=somedatanstuff&test4=moredataontheend")
	expLen := int64(len(expected))

	sbf := NewStringBodyFilter()

	for i := 0; i < b.N; i++ {
		tmp, _ := http.NewRequest("POST", "/hello", bytes.NewReader(expected))
		tmp.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		tmp.ContentLength = expLen
		b.StartTimer()
		// replace the body
		req, _ := falcore.TestWithRequest(tmp, sbf, nil)
		sbf.ReturnBuffer(req)
		// read the body twice
		/* nah, this isn't so useful
		io.CopyN(ioutil.Discard, req.HttpRequest.Body, req.HttpRequest.ContentLength)
		req.HttpRequest.Body	.Close()
		io.CopyN(ioutil.Discard, req.HttpRequest.Body, req.HttpRequest.ContentLength)
		req.HttpRequest.Body	.Close()
		*/
		b.StopTimer()
	}
}
