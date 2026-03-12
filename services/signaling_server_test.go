package services

import (
	"encoding/json"
	"party-race-game-backend/models"
	"testing"
)

func TestRegisterAndUnregisterConnection(t *testing.T) {
	rm := NewRoomManager()
	ss := NewSignalingServer(rm)

	playerID := "player1"

	// Register connection（使用 nil conn，測試 map 操作即可）
	ss.RegisterConnection(playerID, nil)

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
}

func TestJoinAndLeaveSignalingRoom(t *testing.T) {
	rm := NewRoomManager()
	ss := NewSignalingServer(rm)

	roomID := "ROOM123"
	playerID := "player1"

	// 先 Register，再 Join（JoinSignalingRoom 從 connections 重用 wsConn）
	ss.RegisterConnection(playerID, nil)
	ss.JoinSignalingRoom(roomID, playerID)

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
}

func TestHandlePlayerDisconnect(t *testing.T) {
	rm := NewRoomManager()
	ss := NewSignalingServer(rm)

	// Create a room with host and another player
	room, _ := rm.CreateRoom("host123", "Host Player", "")
	rm.JoinRoom(room.ID, "player1", "Player One", "")

	// Register connections
	ss.RegisterConnection("host123", nil)
	ss.RegisterConnection("player1", nil)

	// Join signaling rooms
	ss.JoinSignalingRoom(room.ID, "host123")
	ss.JoinSignalingRoom(room.ID, "player1")

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
	ss.RegisterConnection("host123", nil)
	ss.RegisterConnection("player1", nil)

	// Join signaling rooms
	ss.JoinSignalingRoom(room.ID, "host123")
	ss.JoinSignalingRoom(room.ID, "player1")

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
	ss.RegisterConnection(playerID, nil)

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
