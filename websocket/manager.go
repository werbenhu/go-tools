package websocket

type Manager struct {
	clients    map[string]map[string]*Client
	message    chan *Message
	register   chan *Client
	unregister chan *Client
}

func NewManager() *Manager {
	return &Manager{
		message:    make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[string]*Client),
	}
}

func (m *Manager) Send(userId string, paylaod []byte) error {
	m.message <- &Message{
		UserId: userId,
		Body:   paylaod,
	}
	return nil
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			if m.clients[client.userId] == nil {
				m.clients[client.userId] = make(map[string]*Client)
			}
			m.clients[client.userId][client.id] = client

		case client := <-m.unregister:
			if _, ok := m.clients[client.userId]; ok {
				delete(m.clients[client.userId], client.id)
				close(client.send)
				if len(m.clients[client.userId]) == 0 {
					delete(m.clients, client.userId)
				}
			}

		case message := <-m.message:
			if clients, ok := m.clients[message.UserId]; ok {
				for _, client := range clients {
					// fmt.Printf("websocket client:%s, send body:%s\n", client.userId, string(message.Body))
					client.send <- message.Body
				}
			}
		}
	}
}
