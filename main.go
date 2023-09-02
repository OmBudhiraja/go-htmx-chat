package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/OmBudhiraja/go-htmx-chat/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/net/websocket"
)

type ChatRoom struct {
	Id       int
	Name     string
	Messages []Message
}

type Message struct {
	Sender  string
	Content string
}

var chatRooms = []ChatRoom{
	{
		Id:   0,
		Name: "General",
		Messages: []Message{
			{Sender: "Tim", Content: "Good morning!"},
			{Sender: "Jane", Content: "Hello there!"},
		},
	},
}

func main() {

	wsServer := ws.NewWsWsServer()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("index.html"))

		fmt.Println("Chat rooms", chatRooms)

		tmpl.Execute(w, map[string]interface{}{
			"Rooms":      chatRooms,
			"ActiveRoom": chatRooms[0],
		})
	})

	r.Post("/chat", func(w http.ResponseWriter, r *http.Request) {

		roomId, err := strconv.Atoi(r.FormValue("room"))

		if err != nil {
			fmt.Println("Error parsing room id", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		room := chatRooms[roomId]

		msg := Message{
			Sender:  "Anonymous",
			Content: r.FormValue("content"),
		}

		room.Messages = append(room.Messages, msg)

		chatRooms[roomId] = room

		fmt.Println("Message received", room)

		tmpl := template.Must(template.ParseFiles("partialTemplates/message.html"))
		var messageStr bytes.Buffer
		if err := tmpl.Execute(&messageStr, msg); err != nil {
			fmt.Println("Error executing template", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		wsServer.Broadcast(messageStr.String(), roomId)

		w.WriteHeader(http.StatusOK)
		w.Write(nil)
	})

	r.Post("/create-room", func(w http.ResponseWriter, r *http.Request) {
		newChatRoom := ChatRoom{
			Id:       len(chatRooms),
			Name:     "Untitled",
			Messages: []Message{},
		}
		chatRooms = append(chatRooms, newChatRoom)

		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.ExecuteTemplate(w, "roomBtn", newChatRoom)
	})

	r.Get("/room", func(w http.ResponseWriter, r *http.Request) {
		roomId, err := strconv.Atoi(r.URL.Query().Get("id"))

		if err != nil {
			fmt.Println("Error parsing room id", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		chatRoom := chatRooms[roomId]
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.ExecuteTemplate(w, "ChatSection", map[string]interface{}{
			"ActiveRoom": chatRoom,
		})

	})

	r.Handle("/ws", websocket.Handler(wsServer.HandleWS))

	fmt.Println("Server running on port http://localhost:5000")

	log.Fatal(http.ListenAndServe(":5000", r))
}
