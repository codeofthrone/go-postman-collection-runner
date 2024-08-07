package postmanHelper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/rbretecher/go-postman-collection"
)

// HttpClient is an interface that represents an HTTP client.
// It has a single method, Do, which sends an HTTP request and returns an HTTP response.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// PM is a struct that represents a bid.
// It has three fields: collectionFile, collection, Variables, and httpClient.
type Postman struct {
	collectionFile string
	collection     *postman.Collection
	Variables      map[string]interface{}
	httpClient     HttpClient
}

// NewPostman is a function that creates a new instance of Bid.
// It takes a collection file, a map of variables, and an HttpClient as arguments.
func NewPostman(collectionFile string, variables map[string]interface{}, httpClient HttpClient) (*Postman, error) {
	file, err := os.Open(collectionFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	collection, err := postman.ParseCollection(file)
	if err != nil {
		return nil, err
	}

	return &Postman{
		collectionFile: collectionFile,
		collection:     collection,
		Variables:      variables,
		httpClient:     httpClient,
	}, nil
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
	re := regexp.MustCompile(`{{(.*?)}}`)
	matches := re.FindAllStringSubmatch(text, -1)
	for key, value := range b.Variables {
		for _, match := range matches {
			if key == match[1] {
				variablePlaceholder := fmt.Sprintf("{{%s}}", key)
				switch v := value.(type) {
				case int:
					text = strings.ReplaceAll(text, variablePlaceholder, strconv.Itoa(v))
				case string:
					text = strings.ReplaceAll(text, variablePlaceholder, v)
				case []string:
					// Replace the variable with a comma-separated string representation of the []string.
					quoted := make([]string, len(v))
					for i, s := range v {
						quoted[i] = fmt.Sprintf(`"%s"`, s) // Enclose each string in double quotes
					}
					result := fmt.Sprintf(`[%s]`, strings.Join(quoted, ",")) // Join the quoted strings
					text = strings.ReplaceAll(text, variablePlaceholder, result)
				case nil:
					text = strings.ReplaceAll(text, variablePlaceholder, "null")
				default:
					// Handle all other types of variables.
					text = strings.ReplaceAll(text, variablePlaceholder, fmt.Sprintf("%v", v))
				}
			}
		}
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
		// Remove the extra double quotes around the image IDs in the request body.
		rawBody = strings.ReplaceAll(rawBody, `"[`, `[`)
		rawBody = strings.ReplaceAll(rawBody, `]"`, `]`)
		rawBody = strings.ReplaceAll(rawBody, `<nil>`, "null")
		body = []byte(rawBody)
	}
	req, err := http.NewRequest(string(request.Method), url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if request.Auth != nil {
		switch request.Auth.Type {
		case "bearer":
			authValue := b.ReplaceVariables(fmt.Sprintf("%v", request.Auth.Bearer[0].Value))
			req.Header.Set("Authorization", "Bearer "+authValue)
		case "basic":
			authValue := b.ReplaceVariables(fmt.Sprintf("%v", request.Auth.Basic[0].Value))
			req.Header.Set("Authorization", "Basic "+authValue)
		default:
			log.Println("Unknown auth type", request.Auth.Type)
		}
	}

	for _, header := range request.Header {
		headerValue := b.ReplaceVariables(fmt.Sprintf("%v", header.Value))
		if header.Key == "Authorization" {
			headerValue = "Bearer " + b.ReplaceVariables(headerValue)
		}
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
	result, err := b.SendRequest(req)
	// parser b.collection.events
	// if event is test, execute the script
	events := item.Events
	if result != nil {
		b.ReplaceVariablesInScript(events, result)
	}
	return result, err
}

// GetDataFromResponse retrieves data from a response based on the given query.
// It takes a response map and a query as input and returns the corresponding data.
// The query is a list of keys that represent the path to the desired data in the response map.
// If the data is found, it is returned. Otherwise, nil is returned.
func (b *Postman) GetDataFromResponse(response map[string]interface{}, query []string) interface{} {
	insertRes := response
	for i, s := range query {
		if i == len(query)-1 {
			return insertRes[s]
		}
		if s != "responseData" {
			if response[s] != nil {
				insertRes = response[s].(map[string]interface{})
				return b.GetDataFromResponse(insertRes, query[i+1:])
			}
		}
	}
	return nil
}

// ReplaceVariablesInScript replaces variables in the script based on the provided events and result.
// Based on the events provided which was saved in script, the function checks if the event is a test event.
// This Function current only supports pm.response.json() and pm.environment.set() functions.
// It then retrieves the data from the response based on the extracted information and replaces the variables accordingly.
// The replaced variables are stored in the `Variables` map of the `Postman` struct.
//
// Parameters:
// - events: A slice of `Event` pointers representing the events to process.
// - result: A map[string]interface{} representing the result data.
func (b *Postman) ReplaceVariablesInScript(events []*postman.Event, result map[string]interface{}) {
	for _, event := range events {
		if event.Listen == "test" {
			script := event.Script.Exec
			var source string
			responeseFlag := false
			for _, value := range script {
				if strings.Contains(value, "pm.response.json()") {
					responeseFlag = true
					parts := strings.Split(value, "=")
					if len(parts) < 2 {
						continue
					}
					source = strings.TrimSpace(parts[0])
					source = strings.ReplaceAll(source, "var", "")
					source = strings.ReplaceAll(source, "let", "")
					source = strings.TrimSpace(source)
				}
				if strings.Contains(value, "pm.environment.set") && responeseFlag {
					pattern := `pm\.environment\.set\(\"(.*)\",\s*(.*)\)\;`
					re := regexp.MustCompile(pattern)
					match := re.FindStringSubmatch(value)
					if len(match) < 3 {
						continue
					}
					queryJson := strings.TrimSpace(match[2])
					if strings.Contains(queryJson, source) {
						query := strings.Split(queryJson, ".")
						replaceVariable := b.GetDataFromResponse(result, query)
						if replaceVariable != nil {
							switch v := replaceVariable.(type) {
							case string:
								b.Variables[match[1]] = v
							case []interface{}:
								var strSlice []string
								for _, val := range v {
									strSlice = append(strSlice, val.(string))
								}
								b.Variables[match[1]] = "\"" + strings.Join(strSlice, ",") + "\""
							case nil:
								b.Variables[match[1]] = nil
							default:
								b.Variables[match[1]] = v
							}
						}
					} else {
						b.Variables[match[1]] = match[2]
					}
				}
			}
		}
	}
}
