package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Message struct {
	Sender  string
	Content string
}

var messages = []Message{
	{Sender: "Tim", Content: "Good morning!"},
	{Sender: "Jane", Content: "Hello there!"},
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, messages)
	})

	r.Post("/chat", func(w http.ResponseWriter, r *http.Request) {
		msg := Message{
			Sender:  "Anonymous",
			Content: r.FormValue("content"),
		}

		messages = append(messages, msg)

		tmpl := template.Must(template.ParseFiles("message.html"))
		tmpl.Execute(w, msg)
	})

	fmt.Println("Server running on port http://localhost:5000")

	log.Fatal(http.ListenAndServe(":5000", r))
}
