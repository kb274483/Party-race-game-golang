package services

import (
	"encoding/json"
	"log"
	"party-race-game-backend/models"
	"sync"

	"github.com/gorilla/websocket"
)

// SignalingServer manages WebSocket connections and relays signaling messages
type SignalingServer struct {
	roomManager *RoomManager
	connections map[string]*websocket.Conn
	rooms       map[string]map[string]*websocket.Conn // roomID -> playerID -> connection
	mu          sync.RWMutex
}

// NewSignalingServer creates a new SignalingServer instance
func NewSignalingServer(roomManager *RoomManager) *SignalingServer {
	return &SignalingServer{
		roomManager: roomManager,
		connections: make(map[string]*websocket.Conn),
		rooms:       make(map[string]map[string]*websocket.Conn),
	}
}

// RegisterConnection registers a WebSocket connection for a player
func (ss *SignalingServer) RegisterConnection(playerID string, conn *websocket.Conn) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.connections[playerID] = conn
}

// UnregisterConnection removes a WebSocket connection
func (ss *SignalingServer) UnregisterConnection(playerID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	delete(ss.connections, playerID)
	
	// Remove from all rooms
	for roomID, players := range ss.rooms {
		delete(players, playerID)
		if len(players) == 0 {
			delete(ss.rooms, roomID)
		}
	}
}

// JoinSignalingRoom adds a player's connection to a room
func (ss *SignalingServer) JoinSignalingRoom(roomID, playerID string, conn *websocket.Conn) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.rooms[roomID] == nil {
		ss.rooms[roomID] = make(map[string]*websocket.Conn)
	}
	ss.rooms[roomID][playerID] = conn
}

// LeaveSignalingRoom removes a player's connection from a room
func (ss *SignalingServer) LeaveSignalingRoom(roomID, playerID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if players, exists := ss.rooms[roomID]; exists {
		delete(players, playerID)
		if len(players) == 0 {
			delete(ss.rooms, roomID)
		}
	}
}

// RelaySignal relays a signaling message to the target player or broadcasts to room
func (ss *SignalingServer) RelaySignal(signal models.SignalMessage) error {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	// If targetID is specified, send to specific player
	if signal.TargetID != "" {
		conn, exists := ss.connections[signal.TargetID]
		if !exists {
			log.Printf("Target player %s not found", signal.TargetID)
			return nil
		}
		// Skip nil connections
		if conn == nil {
			return nil
		}
		return conn.WriteJSON(signal)
	}

	// Otherwise, broadcast to all players in the room except sender
	players, exists := ss.rooms[signal.RoomID]
	if !exists {
		log.Printf("Room %s not found", signal.RoomID)
		return nil
	}

	for playerID, conn := range players {
		if playerID != signal.SenderID {
			// Skip nil connections
			if conn == nil {
				continue
			}
			if err := conn.WriteJSON(signal); err != nil {
				log.Printf("Error sending to player %s: %v", playerID, err)
			}
		}
	}

	return nil
}

// BroadcastToRoom broadcasts a message to all players in a room
func (ss *SignalingServer) BroadcastToRoom(roomID string, message interface{}) error {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	players, exists := ss.rooms[roomID]
	if !exists {
		log.Printf("Room %s not found", roomID)
		return nil
	}

	for playerID, conn := range players {
		// Skip nil connections (can happen in tests or during cleanup)
		if conn == nil {
			continue
		}
		
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error broadcasting to player %s: %v", playerID, err)
		}
	}

	return nil
}

// HandleMessage processes incoming WebSocket messages
func (ss *SignalingServer) HandleMessage(playerID string, messageData []byte) error {
	var signal models.SignalMessage
	if err := json.Unmarshal(messageData, &signal); err != nil {
		log.Printf("Error unmarshaling signal: %v", err)
		return err
	}

	signal.SenderID = playerID

	// Handle special message types
	switch signal.Type {
	case "join_room":
		// Player is joining a room via WebSocket
		if roomID, ok := signal.Payload.(map[string]interface{})["roomId"].(string); ok {
			ss.JoinSignalingRoom(roomID, playerID, ss.connections[playerID])
		}
	case "leave_room":
		// Player is leaving a room via WebSocket
		if roomID, ok := signal.Payload.(map[string]interface{})["roomId"].(string); ok {
			ss.LeaveSignalingRoom(roomID, playerID)
		}
	case "game_started":
		// Mark room as in-game so host disconnect won't trigger room_closed
		if signal.RoomID != "" {
			if err := ss.roomManager.SetRoomInGame(signal.RoomID); err != nil {
				log.Printf("Error marking room %s as in-game: %v", signal.RoomID, err)
			} else {
				log.Printf("Room %s marked as in-game", signal.RoomID)
			}
		}
	}

	return ss.RelaySignal(signal)
}

// HandlePlayerDisconnect handles player disconnection events
func (ss *SignalingServer) HandlePlayerDisconnect(playerID, roomID string) {
	ss.mu.Lock()
	
	// Find which room the player is in if roomID not provided
	if roomID == "" {
		for rID, players := range ss.rooms {
			if _, exists := players[playerID]; exists {
				roomID = rID
				break
			}
		}
	}
	
	// Get room info before removing player
	var room *models.Room
	var isHost bool
	if roomID != "" {
		room, _ = ss.roomManager.GetRoom(roomID)
		if room != nil {
			if player, exists := room.Players[playerID]; exists {
				isHost = player.IsHost
			}
		}
	}
	
	ss.mu.Unlock()

	// If game is in progress, skip room deletion and room_closed broadcast
	inGame := room != nil && room.InGame

	// Remove player from room manager (skip if game is in progress to keep room alive)
	if roomID != "" && !inGame {
		if err := ss.roomManager.LeaveRoom(roomID, playerID); err != nil {
			log.Printf("Error removing player %s from room %s: %v", playerID, roomID, err)
		}
	}

	// Unregister connection
	ss.UnregisterConnection(playerID)

	// Notify other players in the room
	if roomID != "" {
		disconnectMsg := models.SignalMessage{
			Type:     "player_disconnect",
			RoomID:   roomID,
			SenderID: playerID,
			Payload: map[string]interface{}{
				"playerId": playerID,
				"isHost":   isHost,
			},
		}

		if err := ss.BroadcastToRoom(roomID, disconnectMsg); err != nil {
			log.Printf("Error broadcasting disconnect message: %v", err)
		}

		// If host disconnected and game has NOT started, notify room is closing
		if isHost && !inGame {
			closeMsg := models.SignalMessage{
				Type:   "room_closed",
				RoomID: roomID,
				Payload: map[string]interface{}{
					"reason": "host_disconnect",
				},
			}
			ss.BroadcastToRoom(roomID, closeMsg)
		}
	}

	log.Printf("Player %s disconnected from room %s", playerID, roomID)
}
