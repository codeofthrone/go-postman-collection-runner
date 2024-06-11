package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rbretecher/go-postman-collection"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// HttpClient is an interface that represents an HTTP client.
// It has a single method, Do, which sends an HTTP request and returns an HTTP response.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// PM is a struct that represents a bid.
// It has three fields: collectionFile, variables, and httpClient.
type Postman struct {
	collectionFile string
	collection     *postman.Collection
	variables      map[string]string
	httpClient     HttpClient
}

// NewPostman is a function that creates a new instance of Bid.
// It takes a collection file, a map of variables, and an HttpClient as arguments.
func NewPostman(collectionFile string, variables map[string]string, httpClient HttpClient) *Postman {
	return &Postman{
		collectionFile: collectionFile,
		variables:      variables,
		httpClient:     httpClient,
	}
}

// ParsePostmanCollection is a method that parses a Postman collection.
// It returns a parsed Postman collection and an error.
func (b *Postman) ParsePostmanCollection() error {
	file, err := os.Open(b.collectionFile)
	if err != nil {
		return err
	}
	defer file.Close()

	collection, err := postman.ParseCollection(file)
	if err != nil {
		return err
	}

	b.collection = collection
	return nil
}

// FindRequestByName is a method that finds a request by name in a list of Postman items.
// It returns the found Postman item and an error.
func (b *Postman) FindRequestByName(items []*postman.Items, name string) (*postman.Items, error) {
	for _, item := range items {
		if item.Items != nil {
			if foundItem, err := b.FindRequestByName(item.Items, name); err == nil {
				return foundItem, nil
			}
		} else if item.Name == name {
			return item, nil
		}
	}
	return nil, fmt.Errorf("request with name %s not found", name)
}

// ReplaceVariables is a method that replaces Postman variables in the given text with their actual values.
// It returns the text with the variables replaced.
func (b *Postman) ReplaceVariables(text string) string {
	for key, value := range b.variables {
		variablePlaceholder := fmt.Sprintf("{{%s}}", key)
		text = strings.ReplaceAll(text, variablePlaceholder, value)
	}
	return text
}

// CreateRequest is a method that creates an HTTP request from a Postman item.
// It returns the created HTTP request and an error.
func (b *Postman) CreateRequest(item *postman.Items) (*http.Request, error) {
	request := item.Request
	url := b.ReplaceVariables(request.URL.Raw)
	var body []byte
	if request.Body != nil && request.Body.Raw != "" {
		rawBody := b.ReplaceVariables(request.Body.Raw)
		body = []byte(rawBody)
	}
	req, err := http.NewRequest(string(request.Method), url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	for _, header := range request.Header {
		headerValue := b.ReplaceVariables(fmt.Sprintf("%v", header.Value))
		req.Header.Set(header.Key, headerValue)
	}
	return req, nil
}

// SendRequest is a method that sends an HTTP request.
// It returns the response from the request and an error.
func (b *Postman) SendRequest(req *http.Request) (map[string]interface{}, error) {
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// FindAndSendRequest is a method that finds a request by name in a Postman collection and sends it.
// It returns the response from the request and an error.
func (b *Postman) FindAndSendRequest(name string) (map[string]interface{}, error) {
	item, err := b.FindRequestByName(b.collection.Items, name)
	if err != nil {
		return nil, err
	}
	req, err := b.CreateRequest(item)
	if err != nil {
		return nil, err
	}
	return b.SendRequest(req)
}
