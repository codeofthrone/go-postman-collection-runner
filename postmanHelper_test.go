package postmanHelper

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/rbretecher/go-postman-collection"
)

type MockHttpClient struct{}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{}, nil
}

func TestNewPostman(t *testing.T) {
	_, err := NewPostman("collection.json", map[string]string{}, &MockHttpClient{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestFindRequestByName(t *testing.T) {
	p, _ := NewPostman("collection.json", map[string]string{}, &MockHttpClient{})
	_, err := p.FindRequestByName(p.collection.Items, "create_user")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestReplaceVariables(t *testing.T) {
	p, _ := NewPostman("collection.json", map[string]string{"var1": "value1"}, &MockHttpClient{})
	result := p.ReplaceVariables("{{var1}}")
	if result != "value1" {
		t.Errorf("Expected 'value1', got %v", result)
	}
}

func TestCreateRequest(t *testing.T) {
	p, _ := NewPostman("collection.json", map[string]string{}, &MockHttpClient{})
	item, _ := p.FindRequestByName(p.collection.Items, "create_user")
	_, err := p.CreateRequest(item)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendRequest(t *testing.T) {
	mockClient := &MockHttpClient{}

	p, _ := NewPostman("collection.json", map[string]string{}, &MockHttpClient{})
	req, err := http.NewRequest("POST", "http://example.com", bytes.NewBuffer([]byte{}))
	_, err = p.SendRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestFindAndSendRequest(t *testing.T) {
	p, _ := NewPostman("collection.json", map[string]string{}, &MockHttpClient{})
	_, err := p.FindAndSendRequest("create_user")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetDataFromResponse(t *testing.T) {
	p, _ := NewPostman("collection.json", map[string]string{}, &MockHttpClient{})
	data := p.GetDataFromResponse(map[string]interface{}{"key": "value"}, []string{"key"})
	if data != "value" {
		t.Errorf("Expected 'value', got %v", data)
	}
}

func TestReplaceVariablesInScript(t *testing.T) {
	p, _ := NewPostman("collection.json", map[string]string{}, &MockHttpClient{})
	p.ReplaceVariablesInScript([]*postman.Event{{Listen: "test", Script: &postman.Script{Exec: []string{"pm.environment.set(\"var1\", \"value1\");"}}}}, map[string]interface{}{})
	if p.Variables["var1"] != "value1" {
		t.Errorf("Expected 'value1', got %v", p.Variables["var1"])
	}
}
