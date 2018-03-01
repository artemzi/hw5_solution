package main

import (
	"net/http"
	"encoding/json"
	"fmt"
)

func (api *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data := r.URL.Query()
		fmt.Printf("MyApi GET %v\n", data)
	case "POST":
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		data := r.Form
		fmt.Printf("MyApi POST %v\n", data)
	}
}

func (api *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data := r.URL.Query()
		fmt.Printf("MyApi GET %v\n", data)
	case "POST":
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		data := r.Form
		fmt.Printf("MyApi POST %v\n", data)
	}
}

// ServeHTTP comment was here
func (api *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})
	switch r.URL.Path {
	
	case "/user/profile":
		api.handlerProfile(w, r)

	case "/user/create":
		api.handlerCreate(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
		resp["error"] = "unknown method"
		body, _ := json.Marshal(resp)
		w.Write(body)
	}
}

func (api *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data := r.URL.Query()
		fmt.Printf("OtherApi GET %v\n", data)
	case "POST":
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		data := r.Form
		fmt.Printf("OtherApi POST %v\n", data)
	}
}

// ServeHTTP comment was here
func (api *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})
	switch r.URL.Path {
	
	case "/user/create":
		api.handlerCreate(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
		resp["error"] = "unknown method"
		body, _ := json.Marshal(resp)
		w.Write(body)
	}
}
