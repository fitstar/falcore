package filter

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/fitstar/falcore"
)

var DefaultTypes = []string{"text/plain", "text/html", "application/json", "text/xml"}

type CompressionFilter struct {
	types []string
}

func NewCompressionFilter(types []string) *CompressionFilter {
	f := new(CompressionFilter)
	if types != nil {
		f.types = types
	} else {
		f.types = DefaultTypes
	}
	return f
}

func (c *CompressionFilter) FilterResponse(request *falcore.Request, res *http.Response) {
	req := request.HttpRequest
	if accept := req.Header.Get("Accept-Encoding"); accept != "" {

		// Is content an acceptable type for encoding?
		var compress = false
		var content_type = res.Header.Get("Content-Type")
		for _, t := range c.types {
			if content_type == t {
				compress = true
				break
			}
		}

		// Is the content already compressed
		if res.Header.Get("Content-Encoding") != "" {
			compress = false
		}

		if !compress {
			request.CurrentStage.Status = 1 // Skip
			return
		}

		// Figure out which encoding to use
		options := strings.Split(accept, ",")
		var mode string
		for _, opt := range options {
			if m := strings.TrimSpace(opt); m == "gzip" || m == "deflate" {
				mode = m
				break
			}
		}

		var compressor io.WriteCloser
		pReader, pWriter := io.Pipe()
		switch mode {
		case "gzip":
			compressor = gzip.NewWriter(pWriter)
		case "deflate":
			comp, err := flate.NewWriter(pWriter, -1)
			if err != nil {
				falcore.Error("Compression Error: %v", err)
				request.CurrentStage.Status = 1 // Skip
				return
			}
			compressor = comp
		default:
			request.CurrentStage.Status = 1 // Skip
			return
		}

		// Perform compression
		var rdr = res.Body
		go func() {
			_, err := io.Copy(compressor, rdr)
			compressor.Close()
			pWriter.Close()
			rdr.Close()
			if err != nil {
				falcore.Error("Error compressing body: %v", err)
			}
		}()

		res.ContentLength = -1
		res.Body = pReader
		res.Header.Set("Content-Encoding", mode)
	} else {
		request.CurrentStage.Status = 1 // Skip
	}
}
