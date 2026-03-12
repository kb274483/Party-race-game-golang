package handlers

import (
	"log"
	"net/http"
	"party-race-game-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	signalingServer *services.SignalingServer
}

// NewWebSocketHandler creates a new WebSocketHandler instance
func NewWebSocketHandler(signalingServer *services.SignalingServer) *WebSocketHandler {
	return &WebSocketHandler{
		signalingServer: signalingServer,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket and handles messages
func (wsh *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	playerID := c.Query("playerId")
	roomID := c.Query("roomId")
	
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "playerId is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Register connection
	wsh.signalingServer.RegisterConnection(playerID, conn)
	defer func() {
		// Handle disconnection
		wsh.signalingServer.HandlePlayerDisconnect(playerID, roomID)
	}()

	// If roomID is provided, join the signaling room
	if roomID != "" {
		wsh.signalingServer.JoinSignalingRoom(roomID, playerID)
	}

	log.Printf("Player %s connected (room: %s)", playerID, roomID)

	// Listen for messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			log.Printf("Player %s disconnected", playerID)
			break
		}

		// Handle the message
		if err := wsh.signalingServer.HandleMessage(playerID, message); err != nil {
			log.Printf("Error handling message from player %s: %v", playerID, err)
		}
	}
}
