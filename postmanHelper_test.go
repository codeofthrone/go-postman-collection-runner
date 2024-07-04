package postmanHelper

import (
	"net/http"
	"testing"

	"github.com/rbretecher/go-postman-collection"
)

type MockHttpClient struct{}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{}, nil
}

func TestFindAndSendRequest(t *testing.T) {
	mockClient := &MockHttpClient{}
	p, _ := NewPostman("collection.json", map[string]interface{}{}, mockClient)
	// Create a mock request item
	mockItem := &postman.Items{
		Name: "create_user",
		Request: &postman.Request{
			Method: "POST",
			URL: &postman.URL{
				Raw: "http://example.com",
			},
		},
	}
	// Add the mock item to the collection
	p.collection.Items = append(p.collection.Items, mockItem)

	// Call the FindAndSendRequest method
	resp, err := p.FindAndSendRequest("create_user")

	// Check if the request was sent successfully
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check if the response is not nil
	if resp == nil {
		t.Errorf("Expected response to be not nil")
	}

	// Add more assertions if needed
}
