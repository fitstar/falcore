package falcore

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	RESPONSE_BODY        = "Hello World!"
	RESPONSE_WRITE_ERROR = "Error writing"
	RESPONSE_FLUSH_ERROR = "Error flushing"
)

var testHandlerWriteResponseData = []struct {
	Name          string
	ErrorReturned string
}{
	{
		"Success",
		RESPONSE_BODY,
	},
	{
		"Write Error",
		RESPONSE_WRITE_ERROR,
	},
	{
		"Flush Error",
		RESPONSE_FLUSH_ERROR,
	},
}

// This implements the net.Conn interface to simulate a connection.
type TestDevnullConn struct {
}

func (c *TestDevnullConn) Read(_ []byte) (n int, err error) {
	return 0, nil
}

func (c *TestDevnullConn) Write(_ []byte) (n int, err error) {
	return 0, nil
}

func (c *TestDevnullConn) Close() error {
	return nil
}

func (c *TestDevnullConn) LocalAddr() net.Addr {
	return nil
}

func (c *TestDevnullConn) RemoteAddr() net.Addr {
	return nil
}

func (c *TestDevnullConn) SetDeadline(_ time.Time) error {
	return nil
}

func (c *TestDevnullConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (c *TestDevnullConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

// This simulates the Response.Body as a io.ReaderCloser.  It returns an
// error when needed and records whether Close() was called.
type TestResponseBody struct {
	strings.Reader
	ReturnError  bool
	ClosedCalled bool
}

func (r *TestResponseBody) Read(p []byte) (n int, err error) {
	// This will simulate a write error for Response
	if r.ReturnError {
		return 0, errors.New(RESPONSE_WRITE_ERROR)
	}
	return r.Reader.Read(p)
}

func (r *TestResponseBody) Close() error {
	// Record if Close() was called
	r.ClosedCalled = true
	return nil
}

func NewTestResponseBody(data string, retError bool) *TestResponseBody {
	return &TestResponseBody{ReturnError: retError, Reader: *strings.NewReader(data), ClosedCalled: false}
}

// Simulates a connection io.Writter.  It supports returning an error so that
// bw.Flush() fails when trying to write to the client's connection.
type ResponseConnectionTest struct {
	bytes.Buffer
	ReturnError bool
}

func (c *ResponseConnectionTest) Write(p []byte) (int, error) {
	// This should trigger a Flush() error
	if c.ReturnError {
		return 0, errors.New(RESPONSE_FLUSH_ERROR)
	}
	return c.Buffer.Write(p)
}

// Setup the correct data for the given test
func setupTestData(errorReturned string) (*Server, *Request, *http.Response, net.Conn, *ResponseConnectionTest, *WriteBufferPoolEntry) {
	pipeline := NewPipeline()
	srv := NewServer(0, pipeline)
	connectionResponse := ResponseConnectionTest{}
	bw := srv.writeBufferPool.Take(&connectionResponse)
	httpReq, _ := http.NewRequest("GET", "/", nil)
	request := newRequest(httpReq, nil, time.Unix(0, 0))
	returnBodyError := false
	switch errorReturned {
	case RESPONSE_WRITE_ERROR:
		returnBodyError = true
		break
	case RESPONSE_FLUSH_ERROR:
		connectionResponse.ReturnError = true
		break
	}
	responseBody := NewTestResponseBody(RESPONSE_BODY, returnBodyError)
	res := SimpleResponse(httpReq, 200, nil, int64(responseBody.Len()), responseBody)
	c := new(TestDevnullConn)
	return srv, request, res, c, &connectionResponse, bw
}

func TestHandlerWriteResponse(t *testing.T) {
	for _, test := range testHandlerWriteResponseData {
		srv, request, res, c, connectionResponse, bw := setupTestData(test.ErrorReturned)
		err := srv.handlerWriteResponse(request, res, c, bw.Br)

		// Close() should always be called on response.Body
		if bodyReader, ok := res.Body.(*TestResponseBody); !ok || !bodyReader.ClosedCalled {
			t.Errorf("%v Close() was not called on response body", test.Name)
		}

		// When everything is successful expect the correct body
		if err == nil {
			// We were expecting an error
			if test.ErrorReturned != RESPONSE_BODY {
				t.Errorf("%v Expected error instead of successful response", test.Name)
			}
			// Make sure the correct content was written to response
			if !strings.Contains(connectionResponse.String(), RESPONSE_BODY) {
				t.Errorf("%v Wrong response written", test.Name)
			}
		} else {
			// Check for the correct error
			if err.Error() != test.ErrorReturned {
				t.Errorf("%v The correct error was not returned, expecting %v instead of %v", test.Name, test.ErrorReturned, err.Error())
			}
			// The body shouldn't have been written to the response
			if connectionResponse.String() != "" {
				t.Errorf("%v Body written to response when it wasn't supposed to: ", test.Name, connectionResponse.String())
			}
		}
	}
}
