package ws

import (
	"fmt"
	"io"

	"golang.org/x/net/websocket"
)

type WsServer struct {
	clients map[*websocket.Conn]bool
}

func NewWsWsServer() *WsServer {
	return &WsServer{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (s *WsServer) HandleWS(ws *websocket.Conn) {
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

func (s *WsServer) Broadcast(msg string) {
	for client := range s.clients {
		client.Write([]byte(msg))
	}
}
