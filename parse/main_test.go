package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockHttpClient struct {
	responseBody []byte
	mock.Mock
}

func (m *MockHttpClient) PostForm(url string, data url.Values) (*http.Response, error) {
	m.Called(url, data)
	resp := &http.Response{}
	resp.Body = ioutil.NopCloser(bytes.NewReader(m.responseBody))
	return resp, nil
}

func TestHandler(t *testing.T) {
	dat, err := ioutil.ReadFile("sample.html")
	if err != nil {
		t.Fail()
	}

	client := &MockHttpClient{responseBody: dat}
	baseUrl := "https://something.com"
	config = Config{
		BaseUrl:  baseUrl,
		Username: "Someuser",
		Password: "Somepassword",
		Client:   client,
	}

	client.On("PostForm", mock.Anything, mock.Anything)

	Handler()
}
