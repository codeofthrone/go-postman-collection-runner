package main

import (
	"fmt"
	postmanHelper "github.com/codeofthrone/go-postman-collection-runner"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	// Define the variables
	variables := map[string]string{
		"host_url":   "http://127.0.0.1:8888",
		"stage_info": ".one",
	}

	pm, err := postmanHelper.NewPostman("postman_collection.json",
		variables, http.DefaultClient)
	if err != nil {
		panic(err)
	}
	response, err := pm.FindAndSendRequest("create_user")
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println(response)
	response, err = pm.FindAndSendRequest("user_upgrade_oa")
	if err != nil {
		fmt.Println(err)
	}
	log.Println(response)
}
