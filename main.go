package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/OmBudhiraja/go-htmx-chat/internal/auth"
	"github.com/OmBudhiraja/go-htmx-chat/internal/db"
	"github.com/OmBudhiraja/go-htmx-chat/internal/scrapper"
	"github.com/OmBudhiraja/go-htmx-chat/internal/utils"
	"github.com/OmBudhiraja/go-htmx-chat/internal/ws"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"golang.org/x/net/websocket"
)

type ChatRoom struct {
	Id       int
	Name     string
	Messages []Message
}

type Message struct {
	Sender  string
	Content template.HTML
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

	db.InitDB()

	r := chi.NewRouter()
	// r.Use(middleware.Logger)

	fs := http.FileServer(http.Dir("./public"))
	r.Handle("/public/*", http.StripPrefix("/public/", fs))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		tmpl := template.Must(template.ParseFiles("views/index.html"))

		tmpl.Execute(w, map[string]interface{}{
			"Rooms":      chatRooms,
			"ActiveRoom": chatRooms[0],
		})
	})

	r.With(auth.AuthMiddleWare).Post("/chat", func(w http.ResponseWriter, r *http.Request) {

		user := r.Context().Value(auth.UserContextKey).(db.User)

		fmt.Println("User", user)

		roomId, err := strconv.Atoi(r.FormValue("room"))

		if err != nil {
			fmt.Println("Error parsing room id", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		room := chatRooms[roomId]

		msg := Message{
			Sender:  "Anonymous",
			Content: utils.RenderMessageWithLinks(r.FormValue("content")),
		}

		room.Messages = append(room.Messages, msg)

		chatRooms[roomId] = room

		fmt.Println("Message received", room)

		tmpl := template.Must(template.ParseFiles("views/fragments/message.html"))
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

		tmpl := template.Must(template.ParseFiles("views/index.html"))
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
		tmpl := template.Must(template.ParseFiles("views/index.html"))
		tmpl.ExecuteTemplate(w, "ChatSection", map[string]interface{}{
			"ActiveRoom": chatRoom,
		})

	})

	r.Post("/link-preview", func(w http.ResponseWriter, r *http.Request) {
		messageInput := r.FormValue("content")
		var url string

		for _, word := range strings.Split(messageInput, " ") {
			if utils.IsValidURL(word) {
				url = word
				break
			}
		}

		if url == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tmpl := template.Must(template.ParseFiles("views/fragments/linkPreviewSkeleton.html"))
		tmpl.Execute(w, map[string]interface{}{
			"Url": url,
		})
	})

	r.Get("/preview-details", func(w http.ResponseWriter, r *http.Request) {

		metadata, err := scrapper.GetMetadata(r.URL.Query().Get("url"))

		if err != nil {
			fmt.Println("Error getting metadata", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tmpl := template.Must(template.ParseFiles("views/fragments/linkPreview.html"))
		tmpl.Execute(w, &metadata)
	})

	r.Handle("/ws", websocket.Handler(wsServer.HandleWS))

	r.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	})

	auth.InitOauth(r)

	fmt.Println("Server running on port http://localhost:5000")

	log.Fatal(http.ListenAndServe(":5000", r))
}
