# go-postman-collection-runner

This package provides a quick way to write a Postman collection runner in Go. It is based on the `go-postman-collection` package.

## Features

- Parse Postman collections
- Find requests by name in a Postman collection
- Replace Postman variables in text with their actual values within test script setup

## Usage

Please replace "path/to/collection.json" and "requestName" with your actual collection file path and request name.

First, create a new Postman instance:

```go
variables := map[string]string{"var1": "value1", "var2": "value2"}
httpClient := &http.Client{}
postman := NewPostman("path/to/collection.json", variables, httpClient)
```

## Parser postman collection
```go
err := postman.ParsePostmanCollection()
if err != nil {
    log.Fatal(err)
}
```

## Find request by name
```go
item, err := postman.FindRequestByName(postman.collection.Items, "requestName")
if err != nil {
    log.Fatal(err)
}
```



Dependencies
[github.com/rbretecher/go-postman-collection]github.com/rbretecher/go-postman-collection


