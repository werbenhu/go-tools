package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/werbenhu/go-tools/websocket"
)

type DefaultReceiver struct {
}

func (h *DefaultReceiver) Recv(id string, payload []byte) error {
	fmt.Printf("id:%s\n", id)
	fmt.Printf("payload:%s\n", string(payload))

	return websocket.Send(id, []byte("response"))
}

func main() {
	websocket.Run()
	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		id := c.Query("id")
		websocket.Server(id, &DefaultReceiver{}, c.Writer, c.Request)
	})

	r.Run(":8889")
}
