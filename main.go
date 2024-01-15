package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/a-h/templ"
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
type TemplComponentFunc func(name string) templ.Component

// utils
// -----
func WriteHTML(w http.ResponseWriter, status int, tmpl *template.Template, tmplName string, data any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/html")

	return tmpl.ExecuteTemplate(w, tmplName, data)
}

func WriteHTML2(r *http.Request, w http.ResponseWriter, status int, component templ.Component) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/html")

	return component.Render(r.Context(), w)
}

func extractID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return parts[2]
	}
	return ""
}

func makeHTMLHandlerFunc(fn ApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			// Use WriteHTML to send an HTML response
			// this is the last fall back case if the handler fails and returned an error
			// handlers should never do this and should always write their own success or error responses
			// this is here as a last resort
			// this way, when you want to throw a 500 error, you can just return an error from the handler
			//  another idea is to have the handler return a status code with the error, but that is not as clean and I THINK that its better to just let the handler function return its own error and success responses
			err = WriteHTML(w, http.StatusInternalServerError, templates, "error.html", ApiError{Error: err.Error()})

			// if WriteHtml fails, fall back to plain text
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}
	}
}

func NewHTMLServer(listAddr string) *ApiServer {
	return &ApiServer{listAddr: listAddr}
}

func (s *ApiServer) Start() {
	http.HandleFunc("/", makeHTMLHandlerFunc(s.indexHandler)) // Use makeHTMLHandlerFunc to wrap the notesHandler functio
	http.HandleFunc("/notes", makeHTMLHandlerFunc(s.notesHandler))
	http.HandleFunc("/notes/", makeHTMLHandlerFunc(s.noteHandler))

	log.Println("listening on", s.listAddr)
	log.Fatal(http.ListenAndServe(s.listAddr, nil)) // Include log.Fatal for proper error handling
}

// main
// ----
var (
	templates = template.Must(template.ParseFiles("index.html", "list.html", "edit.html", "error.html", "view.html"))
	notes     = make(map[string]Note)
	mu        = &sync.Mutex{}
)

func (s *ApiServer) indexHandler(w http.ResponseWriter, r *http.Request) error {
	return WriteHTML(w, http.StatusOK, templates, "index.html", nil)
}

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

	return WriteHTML2(r, w, http.StatusOK, hello("world"))
}

func (s *ApiServer) createNote(w http.ResponseWriter, r *http.Request) error {
	r.ParseForm()
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	note := Note{
		ID:      id,
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
		Created: time.Now(),
	}

	notes[id] = note
	http.Redirect(w, r, "/", http.StatusFound)

	return WriteHTML(w, http.StatusOK, templates, "view.html", notes)
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
		return WriteHTML(w, http.StatusNotFound, templates, "error.html", "Note not found")
	}

	return WriteHTML(w, http.StatusOK, templates, "view.html", note)
}

func (s *ApiServer) updateNote(w http.ResponseWriter, r *http.Request) error {
	id := extractID(r.URL.Path)

	// Lock the notes map for safe concurrent access
	mu.Lock()
	defer mu.Unlock()

	// Check if the note exists
	note, exists := notes[id]
	if !exists {
		return WriteHTML(w, http.StatusNotFound, templates, "error.html", ApiError{Error: "Note not found"})
	}

	// Parse the form data
	if err := r.ParseForm(); err != nil {
		return WriteHTML(w, http.StatusInternalServerError, templates, "error.html", ApiError{Error: "Error parsing form"})
	}

	// Update the note with new values
	notes[id] = Note{
		ID:      id,
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
		Created: note.Created,
	}

	// Redirect to the updated note's view
	http.Redirect(w, r, "/notes/"+id, http.StatusFound)

	return nil
}

func (s *ApiServer) deleteNote(w http.ResponseWriter, r *http.Request) error {
	id := extractID(r.URL.Path)

	mu.Lock()
	// Check if the note exists before deleting
	if _, exists := notes[id]; !exists {
		mu.Unlock() // Unlock before returning
		return WriteHTML(w, http.StatusNotFound, templates, "error.html", ApiError{Error: "Note not found"})
	}

	delete(notes, id)
	mu.Unlock()

	// Redirect to the main notes listing page after deletion
	http.Redirect(w, r, "/notes", http.StatusFound)

	return nil
}

func main() {
	fmt.Println("hello creature ...")

	server := NewHTMLServer(":8080")
	server.Start()
}
