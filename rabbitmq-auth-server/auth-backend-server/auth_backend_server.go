package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	// response values
	allowRequest = "allow"
	denyRequest  = "deny"
)

// handler is a http handler wrapper for generic request operations
func handler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logTag := fmt.Sprintf("%s, %v", r.RequestURI, r.Method)
			// log incoming request
			fmt.Printf("-> %s\n", logTag)
			// if the request is a POST, the content should be urlencoded
			if r.Method == http.MethodPost && r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				http.Error(w, "Expected POST body as application/x-www-form-urlencoded", http.StatusBadRequest)
				return
			}
			// parse the urlencoded request body
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Unable to parse request", http.StatusBadRequest)
				return
			}
			// call original
			h.ServeHTTP(w, r)
			fmt.Printf("<- %s\n", logTag)
		})
}

func writeResult(w http.ResponseWriter, ok bool) {
	var body string
	if ok {
		body = allowRequest
	} else {
		body = denyRequest
	}
	if _, err := fmt.Fprint(w, body); err != nil {
		println("error writing response body", body)
	}
	fmt.Println(body)
}

// userPath handles a user authentication request
func userPath(w http.ResponseWriter, r *http.Request) {
	params := &userAuthN{}
	if err := params.Parse(r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(params)
	writeResult(w, false)
	return
}

// vhostPath handles a virtual host authorisation request
func vhostPath(w http.ResponseWriter, r *http.Request) {
	params := &vHostAuthZ{}
	if err := params.Parse(r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(params)
	writeResult(w, true)
	return
}

// resourcePath handles a resource authorisation request
func resourcePath(w http.ResponseWriter, r *http.Request) {
	params := &resourceAuthZ{}
	if err := params.Parse(r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(params)
	writeResult(w, true)
	return
}

// topicPath handles a topic authorisation request
func topicPath(w http.ResponseWriter, r *http.Request) {
	params := &topicAuthZ{}
	if err := params.Parse(r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(params)
	writeResult(w, true)
	return
}

func main() {
	// create router
	router := mux.NewRouter()
	router.HandleFunc("/auth/user", userPath)
	router.HandleFunc("/auth/vhost", vhostPath)
	router.HandleFunc("/auth/resource", resourcePath)
	router.HandleFunc("/auth/topic", topicPath)
	//log.Fatal(http.ListenAndServeTLS(":8008", "cert/auth_server.crt", "cert/auth_server.key", handler(router)))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", "8008"), handler(router)))
}
