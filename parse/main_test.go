package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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

	_ = os.Setenv("CONFIG", "ewogICJhY2NvdW50cyI6IFsKICAgIHsKICAgICAgIm5hbWUiOiAiYW5uYSIsCiAgICAgICJ1c2VybmFtZSI6ICIzODQ5ODMyODQ5MiIsCiAgICAgICJwYXNzd29yZCI6ICJzdXBlcnNlY3JldCIKICAgIH0sCiAgICB7CiAgICAgICJuYW1lIjogInRvbSIsCiAgICAgICJ1c2VybmFtZSI6ICI3ODUzNDc4NzU4IiwKICAgICAgInBhc3N3b3JkIjogInNlY3JldHN1cGVyIgogICAgfQogIF0KfQo=")
	parseAccountConfig()

	client := &MockHttpClient{responseBody: dat}
	baseUrl := "https://something.com"
	dbClient := &mockDynamoDBClient{}
	config = Config{
		Accounts:      config.Accounts,
		BaseUrl:       baseUrl,
		Client:        client,
		DynamoDBTable: "sometable",
		DBClient:      dbClient,
	}

	client.On("PostForm", mock.Anything, mock.Anything)
	dbClient.On("PutItem", mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)

	Handler()
}
