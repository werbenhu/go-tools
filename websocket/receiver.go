package websocket

type Receiver interface {
	Recv(id string, payload []byte) error
}
