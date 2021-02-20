package ws

import (
	"encoding/json"
	"log"
	"sync"
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
	// Registered Clients.
	Clients map[*Client]struct{}

	// Broadcast messages to all Clients.
	Broadcast chan []byte

	// Register a new client to the hub.
	Register chan *Client

	// Unregister a client from the hub.
	Unregister chan *Client

	// lock to prevent write to closed channel
	sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]struct{}),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// NoClients returns the number of Clients registered
func (h *Hub) NoClients() int {
	h.RLock()
	defer h.RUnlock()
	return len(h.Clients)
}

// Run is the main hub event loop handling register, unregister and broadcast events.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Lock()
			h.Clients[client] = struct{}{}
			h.Unlock()
			client.Done()
		case client := <-h.Unregister:
			h.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				go h.SendInfo(h.GetInfo()) // this way the number of Clients does not change between calling the goroutine and executing it
			}
			h.Unlock()
		case message := <-h.Broadcast:
			h.RLock()
			for client := range h.Clients {
				client.Send <- message
			}
			h.RUnlock()
		}
	}
}

func (h *Hub) GetInfo() Info {
	return Info{
		NoConnections: h.NoClients(),
	}
}

// SendInfo broadcasts hub statistics to all Clients.
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
