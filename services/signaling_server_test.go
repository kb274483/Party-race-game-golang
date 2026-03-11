package services

import (
	"encoding/json"
	"party-race-game-backend/models"
	"testing"

	"github.com/gorilla/websocket"
)

// Mock WebSocket connection for testing
type mockConn struct {
	messages []interface{}
}

func (m *mockConn) WriteJSON(v interface{}) error {
	m.messages = append(m.messages, v)
	return nil
}

func TestRegisterAndUnregisterConnection(t *testing.T) {
	rm := NewRoomManager()
	ss := NewSignalingServer(rm)

	conn := &mockConn{}
	playerID := "player1"

	// Register connection
	ss.RegisterConnection(playerID, (*websocket.Conn)(nil))

	// Verify connection is registered
	ss.mu.RLock()
	_, exists := ss.connections[playerID]
	ss.mu.RUnlock()

	if !exists {
		t.Error("Connection should be registered")
	}

	// Unregister connection
	ss.UnregisterConnection(playerID)

	// Verify connection is unregistered
	ss.mu.RLock()
	_, exists = ss.connections[playerID]
	ss.mu.RUnlock()

	if exists {
		t.Error("Connection should be unregistered")
	}

	_ = conn // Use conn to avoid unused variable error
}

func TestJoinAndLeaveSignalingRoom(t *testing.T) {
	rm := NewRoomManager()
	ss := NewSignalingServer(rm)

	roomID := "ROOM123"
	playerID := "player1"
	conn := &mockConn{}

	// Join signaling room
	ss.JoinSignalingRoom(roomID, playerID, (*websocket.Conn)(nil))

	// Verify player is in room
	ss.mu.RLock()
	players, exists := ss.rooms[roomID]
	ss.mu.RUnlock()

	if !exists {
		t.Error("Room should exist in signaling server")
	}

	if _, exists := players[playerID]; !exists {
		t.Error("Player should be in signaling room")
	}

	// Leave signaling room
	ss.LeaveSignalingRoom(roomID, playerID)

	// Verify player is removed
	ss.mu.RLock()
	players, exists = ss.rooms[roomID]
	ss.mu.RUnlock()

	if exists && len(players) > 0 {
		t.Error("Player should be removed from signaling room")
	}

	_ = conn // Use conn to avoid unused variable error
}

func TestHandlePlayerDisconnect(t *testing.T) {
	rm := NewRoomManager()
	ss := NewSignalingServer(rm)

	// Create a room with host and another player
	room, _ := rm.CreateRoom("host123", "Host Player", "")
	rm.JoinRoom(room.ID, "player1", "Player One", "")

	// Register connections
	ss.RegisterConnection("host123", (*websocket.Conn)(nil))
	ss.RegisterConnection("player1", (*websocket.Conn)(nil))

	// Join signaling rooms
	ss.JoinSignalingRoom(room.ID, "host123", (*websocket.Conn)(nil))
	ss.JoinSignalingRoom(room.ID, "player1", (*websocket.Conn)(nil))

	// Player disconnects
	ss.HandlePlayerDisconnect("player1", room.ID)

	// Verify player is removed from room manager
	updatedRoom, _ := rm.GetRoom(room.ID)
	if _, exists := updatedRoom.Players["player1"]; exists {
		t.Error("Player should be removed from room")
	}

	// Verify connection is unregistered
	ss.mu.RLock()
	_, exists := ss.connections["player1"]
	ss.mu.RUnlock()

	if exists {
		t.Error("Connection should be unregistered")
	}
}

func TestHandlePlayerDisconnectHostLeaves(t *testing.T) {
	rm := NewRoomManager()
	ss := NewSignalingServer(rm)

	// Create a room with host and another player
	room, _ := rm.CreateRoom("host123", "Host Player", "")
	rm.JoinRoom(room.ID, "player1", "Player One", "")

	// Register connections
	ss.RegisterConnection("host123", (*websocket.Conn)(nil))
	ss.RegisterConnection("player1", (*websocket.Conn)(nil))

	// Join signaling rooms
	ss.JoinSignalingRoom(room.ID, "host123", (*websocket.Conn)(nil))
	ss.JoinSignalingRoom(room.ID, "player1", (*websocket.Conn)(nil))

	// Host disconnects
	ss.HandlePlayerDisconnect("host123", room.ID)

	// Verify room is deleted
	_, err := rm.GetRoom(room.ID)
	if err == nil {
		t.Error("Room should be deleted when host disconnects")
	}
}

func TestHandleMessage(t *testing.T) {
	rm := NewRoomManager()
	ss := NewSignalingServer(rm)

	playerID := "player1"
	roomID := "ROOM123"

	// Register connection
	ss.RegisterConnection(playerID, (*websocket.Conn)(nil))

	// Create a signal message
	signal := models.SignalMessage{
		Type:     "join_room",
		RoomID:   roomID,
		SenderID: playerID,
		Payload: map[string]interface{}{
			"roomId": roomID,
		},
	}

	messageData, _ := json.Marshal(signal)

	// Handle the message
	err := ss.HandleMessage(playerID, messageData)
	if err != nil {
		t.Fatalf("Failed to handle message: %v", err)
	}

	// Verify player joined signaling room
	ss.mu.RLock()
	players, exists := ss.rooms[roomID]
	ss.mu.RUnlock()

	if !exists {
		t.Error("Room should exist after join_room message")
	}

	if _, exists := players[playerID]; !exists {
		t.Error("Player should be in signaling room after join_room message")
	}
}
