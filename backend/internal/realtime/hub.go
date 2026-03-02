package realtime

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/Sharvari1892/examshield/internal/domain"
)

type Client struct {
	Conn   *websocket.Conn
	Role   string
	UserID string
}

type Hub struct {
	clients map[*Client]bool
	mu      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(alert domain.IntegrityAlert) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {

		// Admin gets all alerts
		if client.Role == "admin" {
			client.Conn.WriteJSON(alert)
		}

		// Student gets only own session alert
		if client.Role == "student" && client.UserID == alert.SessionID {
			client.Conn.WriteJSON(alert)
		}
	}
}