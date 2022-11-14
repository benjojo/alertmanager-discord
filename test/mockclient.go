package test

import (
	// "bytes"
	"io"
	"net/http"
)

type MockClientRequest struct {
	Url         string
	ContentType string
	Body        io.Reader
}

type MockClientRecorder struct {
	Requests []MockClientRequest
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewMockClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func (mc *MockClientRecorder) NewMockClientWithResponse(statusCode int) *http.Client {
	return NewMockClient(func(req *http.Request) *http.Response {
		mc.Requests = append(mc.Requests, MockClientRequest{
			Url:         req.URL.String(),
			ContentType: req.Header.Get("content-type"),
			Body:        req.Body,
		})

		return &http.Response{
			StatusCode: statusCode,
			// Body: io.NopCloser(bytes.NewBufferString(responseBody)),
		}
	})
}

// intended to cause errors in the client
// to be used for tests of error handling
func (mc *MockClientRecorder) NewMockClientReturnsNil() *http.Client {
	return NewMockClient(func(req *http.Request) *http.Response {
		mc.Requests = append(mc.Requests, MockClientRequest{
			Url:         req.URL.String(),
			ContentType: req.Header.Get("content-type"),
			Body:        req.Body,
		})

		return &http.Response{}
	})
}
