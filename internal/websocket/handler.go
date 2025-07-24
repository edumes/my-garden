package websocket

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, configure this based on your CORS policy
	},
}

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

type ClientMessage struct {
	Type     string      `json:"type"`
	GardenID uuid.UUID   `json:"garden_id"`
	Data     interface{} `json:"data"`
}

type ActionData struct {
	Position int    `json:"position"`
	Action   string `json:"action"`
}

// HandleWebSocket upgrades the HTTP connection to WebSocket
func (h *Handler) HandleWebSocket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade connection"})
		return
	}

	client := &Client{
		hub:      h.hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID.(uuid.UUID),
		username: username.(string),
		gardens:  make(map[uuid.UUID]bool),
	}

	// Register client
	h.hub.register <- client

	// Start client write pump
	go client.WritePump()

	// Start client read pump
	go h.readPump(client)
}

func (h *Handler) readPump(client *Client) {
	defer func() {
		h.hub.unregister <- client
		client.conn.Close()
	}()

	for {
		var message ClientMessage
		err := client.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log error
			}
			break
		}

		switch message.Type {
		case "start_action":
			if data, ok := message.Data.(map[string]interface{}); ok {
				if position, ok := data["position"].(float64); ok {
					if action, ok := data["action"].(string); ok {
						success := h.hub.acquireLock(message.GardenID, client.userID, client.username, int(position), action)
						if success {
							client.position = new(int)
							*client.position = int(position)
							client.action = new(string)
							*client.action = action
							h.hub.broadcastPresence()
						}
					}
				}
			}

		case "end_action":
			if data, ok := message.Data.(map[string]interface{}); ok {
				if position, ok := data["position"].(float64); ok {
					h.hub.releaseLock(message.GardenID, client.userID, int(position))
					client.position = nil
					client.action = nil
					h.hub.broadcastPresence()
				}
			}

		case "chat":
			if data, ok := message.Data.(map[string]interface{}); ok {
				if messageText, ok := data["message"].(string); ok {
					chatMessage := &ChatMessage{
						Message:   messageText,
						Username:  client.username,
						UserID:    client.userID,
						Timestamp: time.Now().Unix(),
					}

					// Broadcast chat message to all clients in the garden
					h.hub.broadcast <- &Event{
						Type:      EventChat,
						GardenID:  message.GardenID,
						UserID:    client.userID,
						Username:  client.username,
						Data:      chatMessage,
						Timestamp: time.Now().Unix(),
					}
				}
			}
		}
	}
}

// BroadcastEvent sends an event to all connected clients with access to the garden
func (h *Handler) BroadcastEvent(eventType EventType, gardenID, userID uuid.UUID, data interface{}) {
	event := &Event{
		Type:      eventType,
		GardenID:  gardenID,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	h.hub.broadcast <- event
}
