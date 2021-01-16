package ws

import (
	"encoding/json"
	"log"
	"time"
)

const (
	maxMessageSize = 4096
	pongWait       = 2 * time.Minute
	pingPeriod     = time.Minute
	writeWait      = 10 * time.Second
)

type Info struct {
	NoConnections int `json:"no_connections"`
}

type Hub struct {
	// Registered clients.
	clients map[*Client]struct{}

	// Broadcast messages to all clients.
	Broadcast chan []byte

	// Register a new client to the hub.
	Register chan *Client

	// Unregister a client from the hub.
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// NoClients returns the number of clients registered
func (h *Hub) NoClients() int {
	return len(h.clients)
}

// Run is the main hub event loop handling register, unregister and broadcast events.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = struct{}{}
		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				go h.SendInfo(h.GetInfo()) // this way the number of clients does not change between calling the goroutine and executing it
			}
		case message := <-h.Broadcast:
			for client := range h.clients {
				client.Send <- message
			}
		}
	}
}

func (h *Hub) GetInfo() Info {
	return Info{
		NoConnections: h.NoClients(),
	}
}

// SendInfo broadcasts hub statistics to all clients.
func (h *Hub) SendInfo(info Info) {
	i, err := json.Marshal(info)
	if err != nil {
		log.Printf("could not marshal ws info: %s", err)
	}
	if msg, err := json.Marshal(WebsocketMessage{
		Event: MessageTypeInfo,
		Data:  string(i),
	}); err == nil {
		h.Broadcast <- msg
	} else {
		log.Printf("could not marshal ws message: %s", err)
	}
}