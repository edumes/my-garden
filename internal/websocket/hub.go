package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// EventType represents different types of garden events
type EventType string

const (
	EventPlant      EventType = "plant"
	EventHarvest    EventType = "harvest"
	EventWeather    EventType = "weather"
	EventWater      EventType = "water"
	EventProgress   EventType = "progress"
	EventPresence   EventType = "presence"
	EventActionLock EventType = "action_lock"
	EventChat       EventType = "chat"
)

// Event represents a garden event to be broadcasted
type Event struct {
	Type      EventType   `json:"type"`
	GardenID  uuid.UUID   `json:"garden_id"`
	UserID    uuid.UUID   `json:"user_id"`
	Username  string      `json:"username,omitempty"` // Added for chat messages
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// ChatMessage represents a chat message in a garden
type ChatMessage struct {
	Message   string    `json:"message"`
	Username  string    `json:"username"`
	UserID    uuid.UUID `json:"user_id"`
	Timestamp int64     `json:"timestamp"`
}

// Client represents a connected websocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   uuid.UUID
	username string
	gardens  map[uuid.UUID]bool // gardens the user has access to
	mu       sync.RWMutex
	position *int    // Current plot position being interacted with
	action   *string // Current action being performed
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Event
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	locks      map[uuid.UUID]map[int]*ActionLock // garden_id -> position -> lock
	locksMu    sync.RWMutex
}

type ActionLock struct {
	UserID    uuid.UUID
	Username  string
	Action    string
	ExpiresAt time.Time
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Event),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		locks:      make(map[uuid.UUID]map[int]*ActionLock),
	}
}

func (h *Hub) Run() {
	lockCleanupTicker := time.NewTicker(5 * time.Second)
	defer lockCleanupTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.broadcastPresence()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.broadcastPresence()
			h.releaseLocks(client.userID)

		case event := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Check if client has access to this garden
				client.mu.RLock()
				hasAccess := client.gardens[event.GardenID]
				client.mu.RUnlock()

				if hasAccess {
					eventJSON, err := json.Marshal(event)
					if err != nil {
						continue
					}
					select {
					case client.send <- eventJSON:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()

		case <-lockCleanupTicker.C:
			h.cleanupExpiredLocks()
		}
	}
}

func (h *Hub) broadcastPresence() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Group clients by garden
	gardenUsers := make(map[uuid.UUID][]map[string]interface{})
	for client := range h.clients {
		client.mu.RLock()
		for gardenID := range client.gardens {
			userData := map[string]interface{}{
				"id":       client.userID,
				"username": client.username,
			}
			if client.position != nil {
				userData["position"] = *client.position
			}
			if client.action != nil {
				userData["action"] = *client.action
			}
			gardenUsers[gardenID] = append(gardenUsers[gardenID], userData)
		}
		client.mu.RUnlock()
	}

	// Broadcast presence for each garden
	for gardenID, users := range gardenUsers {
		h.broadcast <- &Event{
			Type:      EventPresence,
			GardenID:  gardenID,
			Data:      map[string]interface{}{"users": users},
			Timestamp: time.Now().Unix(),
		}
	}
}

func (h *Hub) acquireLock(gardenID, userID uuid.UUID, username string, position int, action string) bool {
	h.locksMu.Lock()
	defer h.locksMu.Unlock()

	// Initialize garden locks if not exists
	if _, ok := h.locks[gardenID]; !ok {
		h.locks[gardenID] = make(map[int]*ActionLock)
	}

	// Check if position is locked
	if lock, ok := h.locks[gardenID][position]; ok {
		if time.Now().Before(lock.ExpiresAt) {
			return false
		}
	}

	// Acquire lock
	h.locks[gardenID][position] = &ActionLock{
		UserID:    userID,
		Username:  username,
		Action:    action,
		ExpiresAt: time.Now().Add(30 * time.Second),
	}

	// Broadcast lock event
	h.broadcast <- &Event{
		Type:     EventActionLock,
		GardenID: gardenID,
		UserID:   userID,
		Data: map[string]interface{}{
			"position": position,
			"user_id":  userID,
			"username": username,
			"action":   action,
			"locked":   true,
		},
		Timestamp: time.Now().Unix(),
	}

	return true
}

func (h *Hub) releaseLock(gardenID, userID uuid.UUID, position int) {
	h.locksMu.Lock()
	defer h.locksMu.Unlock()

	if gardenLocks, ok := h.locks[gardenID]; ok {
		if lock, ok := gardenLocks[position]; ok && lock.UserID == userID {
			delete(gardenLocks, position)

			// Broadcast lock release
			h.broadcast <- &Event{
				Type:     EventActionLock,
				GardenID: gardenID,
				UserID:   userID,
				Data: map[string]interface{}{
					"position": position,
					"user_id":  userID,
					"locked":   false,
				},
				Timestamp: time.Now().Unix(),
			}
		}
	}
}

func (h *Hub) releaseLocks(userID uuid.UUID) {
	h.locksMu.Lock()
	defer h.locksMu.Unlock()

	for gardenID, gardenLocks := range h.locks {
		for position, lock := range gardenLocks {
			if lock.UserID == userID {
				delete(gardenLocks, position)

				// Broadcast lock release
				h.broadcast <- &Event{
					Type:     EventActionLock,
					GardenID: gardenID,
					UserID:   userID,
					Data: map[string]interface{}{
						"position": position,
						"user_id":  userID,
						"locked":   false,
					},
					Timestamp: time.Now().Unix(),
				}
			}
		}
	}
}

func (h *Hub) cleanupExpiredLocks() {
	h.locksMu.Lock()
	defer h.locksMu.Unlock()

	now := time.Now()
	for gardenID, gardenLocks := range h.locks {
		for position, lock := range gardenLocks {
			if now.After(lock.ExpiresAt) {
				delete(gardenLocks, position)

				// Broadcast lock release
				h.broadcast <- &Event{
					Type:     EventActionLock,
					GardenID: gardenID,
					UserID:   lock.UserID,
					Data: map[string]interface{}{
						"position": position,
						"user_id":  lock.UserID,
						"locked":   false,
					},
					Timestamp: now.Unix(),
				}
			}
		}
	}
}

// AddGardenAccess adds garden access for a client
func (c *Client) AddGardenAccess(gardenID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gardens[gardenID] = true
}

// RemoveGardenAccess removes garden access for a client
func (c *Client) RemoveGardenAccess(gardenID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.gardens, gardenID)
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}
