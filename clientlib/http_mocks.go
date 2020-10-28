package clientlib

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type mockHTTPClient struct {
	Status  int
	Payload []byte
	Error   error
}

func (mc *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: mc.Status,
		Body:       ioutil.NopCloser(bytes.NewReader(mc.Payload)),
	}, mc.Error
}

// NewClientMock returns a mock implementation of the HTTPClient interface.
// status will be the StatusCode of the returned http.Response.
// payload will be the Body of the returned http.Response.
// err will be the error returned.
func NewClientMock(status int, payload []byte, err error) HTTPClient {
	return &mockHTTPClient{
		Status:  status,
		Payload: payload,
		Error:   err,
	}
}
