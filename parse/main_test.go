package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

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

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	mock.Mock
}

func (m *mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func TestHandler(t *testing.T) {
	dat, err := ioutil.ReadFile("sample.html")
	if err != nil {
		t.Fail()
	}

	client := &MockHttpClient{responseBody: dat}
	baseUrl := "https://something.com"
	dbClient := &mockDynamoDBClient{}
	config = Config{
		BaseUrl:       baseUrl,
		Username:      "Someuser",
		Password:      "Somepassword",
		Client:        client,
		DynamoDBTable: "sometable",
		DBClient:      dbClient,
	}

	client.On("PostForm", mock.Anything, mock.Anything)
	dbClient.On("PutItem", mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)

	Handler()
}
