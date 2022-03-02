package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

var m *Manager

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Run() {
	m = NewManager()
	go m.Run()
}

func Server(userId string, receiver Receiver, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		id:      xid.New().String(),
		userId:  userId,
		manager: m,
		conn:    conn,
		send:    make(chan []byte, 1000),
		// send:     make(chan []byte),
		receiver: receiver,
	}
	m.register <- client

	go client.writeLoop()
	go client.readLoop()
}

func Send(id string, payload []byte) error {
	return m.Send(id, payload)
}
