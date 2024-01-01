package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"
)

// types
// -----
type Note struct {
	ID      string
	Title   string
	Content string
	Created time.Time
}

// NOTE: we could omit the error return value, but then we would need to handle the errors in the handler function...and I don't like that. the HandleFunc from net/http does not return an error, so we need to wrap it in a function that does return an error! So we are going to make a mapping type:
type ApiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiServer struct {
	listAddr string
}

type ApiError struct {
	Error string
}

// utils
// -----
func WriteJSON(w http.ResponseWriter, status int, value any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	return json.NewEncoder(w).Encode(value)
}

func extractID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return parts[2]
	}
	return ""
}

func makeHTTPHandlerFunc(fn ApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			// NOTE: Now we have central place to handle the error
			http.Error(w, err.Error(), 500)

			// TODO: this should be an HTML response not JSON
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func NewHTMLServer(listAddr string) *ApiServer {
	return &ApiServer{listAddr: listAddr}
}

func (s *ApiServer) Start() {
	// TODO: how to get params from path?
	http.HandleFunc("/notes", makeHTTPHandlerFunc(s.notesHandler))
	http.HandleFunc("/notes/", makeHTTPHandlerFunc(s.noteHandler))

	log.Println("listening on", s.listAddr)
	log.Fatal(http.ListenAndServe(s.listAddr, nil)) // Include log.Fatal for proper error handling
}

// main
// ----
var (
	templates = template.Must(template.ParseFiles("list.html", "edit.html", "view.html"))
	notes     = make(map[string]Note)
	mu        = &sync.Mutex{}
)

func (s *ApiServer) notesHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.listNotes(w, r)
	}

	if r.Method == "POST" {
		return s.createNote(w, r)
	}

	return fmt.Errorf("unsupported method: %s", r.Method)
}

func (s *ApiServer) listNotes(w http.ResponseWriter, r *http.Request) error {
	mu.Lock()
	defer mu.Unlock()

	// TODO: how to return HTML and status code?
	return templates.ExecuteTemplate(w, "list.html", notes)
}

func (s *ApiServer) createNote(w http.ResponseWriter, r *http.Request) error {
	r.ParseForm()
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	notes[id] = Note{
		ID:      id,
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
		Created: time.Now(),
	}
	http.Redirect(w, r, "/", http.StatusFound)

	// TODO: how to return HTML and status code?
	return nil
}

// note handler
// ------------
func (s *ApiServer) noteHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.getNote(w, r)
	}

	if r.Method == "PUT" {
		return s.updateNote(w, r)
	}

	if r.Method == "DELETE" {
		return s.deleteNote(w, r)
	}

	return fmt.Errorf("unsupported method: %s", r.Method)
}

func (s *ApiServer) getNote(w http.ResponseWriter, r *http.Request) error {
	id := extractID(r.URL.Path)
	mu.Lock()
	note, ok := notes[id]
	mu.Unlock()

	if !ok {
		// TODO: how to return HTML and status code?
		return nil
	}

	templates.ExecuteTemplate(w, "view.html", note)

	// TODO: how to return HTML and status code?
	return nil
}

func (s *ApiServer) updateNote(w http.ResponseWriter, r *http.Request) error {
	id := extractID(r.URL.Path)
	if r.Method == "POST" {
		r.ParseForm()
		mu.Lock()
		notes[id] = Note{
			ID:      id,
			Title:   r.FormValue("title"),
			Content: r.FormValue("content"),
			Created: notes[id].Created,
		}
		mu.Unlock()
		http.Redirect(w, r, "/", http.StatusFound)

		// TODO: how to return HTML and status code?
		return nil
	}
	mu.Lock()
	note, ok := notes[id]
	mu.Unlock()
	if !ok {
		http.NotFound(w, r)

		// TODO: how to return HTML and status code?
		return nil
	}
	templates.ExecuteTemplate(w, "edit.html", note)

	// TODO: how to return HTML and status code?
	return nil
}

func (s *ApiServer) deleteNote(w http.ResponseWriter, r *http.Request) error {
	id := extractID(r.URL.Path)
	mu.Lock()
	delete(notes, id)
	mu.Unlock()
	http.Redirect(w, r, "/", http.StatusFound)

	// TODO: how to return HTML and status code?
	return nil
}

func main() {
	fmt.Println("hello creature ...")

	server := NewHTMLServer(":8080")
	server.Start()
}
