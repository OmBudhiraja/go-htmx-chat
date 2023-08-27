package ws

import (
	"fmt"
	"io"
	"strconv"

	"golang.org/x/net/websocket"
)

type WsServer struct {
	clients map[*websocket.Conn]int
}

func NewWsWsServer() *WsServer {
	return &WsServer{
		clients: make(map[*websocket.Conn]int), // map ws:roomId
	}
}

func (s *WsServer) HandleWS(ws *websocket.Conn) {
	roomId, err := strconv.Atoi(ws.Request().URL.Query().Get("room"))

	if err != nil {
		fmt.Println("Error parsing room id", err.Error())
		ws.Close()
		return
	}

	fmt.Println("New incoming client", ws.RemoteAddr().String())
	s.clients[ws] = roomId

	s.readLoop(ws)
}

func (s *WsServer) CloseWS(ws *websocket.Conn) {
	fmt.Println("Client disconnected", ws.RemoteAddr().String())
	delete(s.clients, ws)
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

func (s *WsServer) Broadcast(msg string, roomId int) {
	for client := range s.clients {
		if s.clients[client] != roomId {
			continue
		}
		client.Write([]byte(msg))
	}
}
