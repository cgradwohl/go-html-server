package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// utils
// -----
func WriteJSON(w http.ResponseWriter, status int, value any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	return json.NewEncoder(w).Encode(value)
}

// HTTP Types and Utils
// --------------------------------
// NOTE: we could omit the error return value, but then we would need to handle the errors in the handler function...and I don't like that. the HandleFunc from net/http does not return an error, so we need to wrap it in a function that does return an error! So we are going to make a mapping type:
type ApiFunc func(w http.ResponseWriter, r *http.Request) error

func makeHTTPHandlerFunc(fn ApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			// Now we have central place to handle the error
			http.Error(w, err.Error(), 500)

			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

type ApiError struct {
	Error string
}

// /account Handler
// ----------------
// net/http cannot handle GET, PUT, POST, DELETE, etc. in the same handler function, so we need to take care of that routing ourselves here in the main /account Handler
func (s *ApiServer) accountHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.getAccount(w, r)
	}

	if r.Method == "POST" {
		return s.createAccount(w, r)
	}

	if r.Method == "DELETE" {
		return s.deleteAccount(w, r)
	}

	return fmt.Errorf("unsupported method: %s", r.Method)
}

func (s *ApiServer) getAccount(w http.ResponseWriter, r *http.Request) error {
	return WriteJSON(w, http.StatusOK, "hello creature")
}

func (s *ApiServer) createAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ApiServer) deleteAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// /transfer Handler
// ------------------
func (s *ApiServer) transferHandler(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Server
// ------
type ApiServer struct {
	listAddr string
}

func NewApiServer(listAddr string) *ApiServer {
	return &ApiServer{listAddr: listAddr}
}

func (s *ApiServer) Run() {
	// TODO: how to get params from path?
	http.HandleFunc("/account", makeHTTPHandlerFunc(s.accountHandler))
	http.HandleFunc("/transfer", makeHTTPHandlerFunc(s.transferHandler))

	log.Println("listening on", s.listAddr)
	log.Fatal(http.ListenAndServe(s.listAddr, nil)) // Include log.Fatal for proper error handling
}

func main() {
	fmt.Println("hello creature ...")

	server := NewApiServer(":8080")
	server.Run()
}
