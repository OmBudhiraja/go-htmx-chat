package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/net/websocket"
)

type Message struct {
	Sender  string
	Content string
}

type WsServer struct {
	clients map[*websocket.Conn]bool
}

func NewWsWsServer() *WsServer {
	return &WsServer{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (s *WsServer) handleWS(ws *websocket.Conn) {
	fmt.Println("New incoming client", ws.RemoteAddr().String())
	s.clients[ws] = true

	s.readLoop(ws)
}

func (s *WsServer) readLoop(ws *websocket.Conn) {
	var buff []byte
	for {
		n, err := ws.Read(buff)

		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading message", err.Error())
			continue
		}
		msg := buff[:n]
		fmt.Println("Message received", string(msg))

	}
}

func (s *WsServer) broadcast(msg string) {
	for client := range s.clients {
		client.Write([]byte(msg))
	}
}

var messages = []Message{
	{Sender: "Tim", Content: "Good morning!"},
	{Sender: "Jane", Content: "Hello there!"},
}

func main() {

	wsServer := NewWsWsServer()

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

		tmpl := template.Must(template.ParseFiles("index.html"))
		var tpl bytes.Buffer
		if err := tmpl.ExecuteTemplate(&tpl, "message", msg); err != nil {
			fmt.Println("Error executing template", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		messageStr := `<div id="messages" hx-swap-oob="beforeend">` + tpl.String() + `</div>`
		wsServer.broadcast(messageStr)

		w.WriteHeader(http.StatusOK)
		w.Write(nil)
	})

	r.Handle("/ws", websocket.Handler(wsServer.handleWS))

	fmt.Println("Server running on port http://localhost:5000")

	log.Fatal(http.ListenAndServe(":5000", r))
}
