package socket

import (
	"sync"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Mapping userID to clients (user can have multiple connections)
	userClients map[uuid.UUID]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Messages to specific user
	sendToUser chan *UserMessage

	logger *zap.Logger
	mu     sync.RWMutex
}

// UserMessage represents a message to be sent to a specific user
type UserMessage struct {
	UserID  uuid.UUID
	Message []byte
}

// NewHub creates a new Hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		userClients: make(map[uuid.UUID]map[*Client]bool),
		sendToUser:  make(chan *UserMessage),
		logger:      logger,
	}
}

// Run starts the hub loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			
			// Add to user map
			if client.userID != uuid.Nil {
				if _, ok := h.userClients[client.userID]; !ok {
					h.userClients[client.userID] = make(map[*Client]bool)
				}
				h.userClients[client.userID][client] = true
			}
			h.mu.Unlock()
			
			h.logger.Debug("Client connected", zap.String("user_id", client.userID.String()))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				
				// Remove from user map
				if client.userID != uuid.Nil {
					if userMap, ok := h.userClients[client.userID]; ok {
						delete(userMap, client)
						if len(userMap) == 0 {
							delete(h.userClients, client.userID)
						}
					}
				}
			}
			h.mu.Unlock()
			h.logger.Debug("Client disconnected", zap.String("user_id", client.userID.String()))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()

		case msg := <-h.sendToUser:
			h.mu.RLock()
			if clients, ok := h.userClients[msg.UserID]; ok {
				for client := range clients {
					select {
					case client.send <- msg.Message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast broadcasts a message to all connected clients
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uuid.UUID, message []byte) {
	h.sendToUser <- &UserMessage{
		UserID:  userID,
		Message: message,
	}
}
