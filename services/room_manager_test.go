package services

import (
	"testing"
)

func TestCreateRoom(t *testing.T) {
	rm := NewRoomManager()
	
	// Test creating a room
	room, err := rm.CreateRoom("host123", "Host Player", "password123")
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	
	if room.ID == "" {
		t.Error("Room ID should not be empty")
	}
	
	if room.HostID != "host123" {
		t.Errorf("Expected HostID to be 'host123', got '%s'", room.HostID)
	}
	
	if room.Password != "password123" {
		t.Errorf("Expected Password to be 'password123', got '%s'", room.Password)
	}
	
	if len(room.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(room.Players))
	}
	
	host, exists := room.Players["host123"]
	if !exists {
		t.Error("Host should be in the room")
	}
	
	if host.Name != "Host Player" {
		t.Errorf("Expected host name to be 'Host Player', got '%s'", host.Name)
	}
	
	if !host.IsHost {
		t.Error("Host should have IsHost set to true")
	}
}

func TestJoinRoom(t *testing.T) {
	rm := NewRoomManager()
	
	// Create a room
	room, _ := rm.CreateRoom("host123", "Host Player", "password123")
	
	// Test joining with correct password
	err := rm.JoinRoom(room.ID, "player1", "Player One", "password123")
	if err != nil {
		t.Fatalf("Failed to join room: %v", err)
	}
	
	// Verify player was added
	updatedRoom, _ := rm.GetRoom(room.ID)
	if len(updatedRoom.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(updatedRoom.Players))
	}
	
	// Test joining with incorrect password
	err = rm.JoinRoom(room.ID, "player2", "Player Two", "wrongpassword")
	if err == nil {
		t.Error("Should fail with incorrect password")
	}
	
	// Test joining non-existent room
	err = rm.JoinRoom("INVALID", "player3", "Player Three", "password123")
	if err == nil {
		t.Error("Should fail when room doesn't exist")
	}
}

func TestJoinRoomPlayerLimit(t *testing.T) {
	rm := NewRoomManager()
	
	// Create a room
	room, _ := rm.CreateRoom("host123", "Host Player", "")
	
	// Add 5 more players (total 6)
	for i := 1; i <= 5; i++ {
		err := rm.JoinRoom(room.ID, string(rune('a'+i)), "Player", "")
		if err != nil {
			t.Fatalf("Failed to add player %d: %v", i, err)
		}
	}
	
	// Try to add 7th player
	err := rm.JoinRoom(room.ID, "player7", "Player Seven", "")
	if err == nil {
		t.Error("Should fail when room is full")
	}
}

func TestLeaveRoom(t *testing.T) {
	rm := NewRoomManager()
	
	// Create a room and add a player
	room, _ := rm.CreateRoom("host123", "Host Player", "")
	rm.JoinRoom(room.ID, "player1", "Player One", "")
	
	// Test player leaving
	err := rm.LeaveRoom(room.ID, "player1")
	if err != nil {
		t.Fatalf("Failed to leave room: %v", err)
	}
	
	// Verify player was removed
	updatedRoom, _ := rm.GetRoom(room.ID)
	if len(updatedRoom.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(updatedRoom.Players))
	}
}

func TestLeaveRoomHostLeaves(t *testing.T) {
	rm := NewRoomManager()
	
	// Create a room and add a player
	room, _ := rm.CreateRoom("host123", "Host Player", "")
	rm.JoinRoom(room.ID, "player1", "Player One", "")
	
	// Host leaves - room should be deleted
	err := rm.LeaveRoom(room.ID, "host123")
	if err != nil {
		t.Fatalf("Failed to leave room: %v", err)
	}
	
	// Verify room was deleted
	_, err = rm.GetRoom(room.ID)
	if err == nil {
		t.Error("Room should be deleted when host leaves")
	}
}

func TestLeaveRoomAutoDelete(t *testing.T) {
	rm := NewRoomManager()
	
	// Create a room
	room, _ := rm.CreateRoom("host123", "Host Player", "")
	
	// Host leaves - room should be deleted (empty room)
	err := rm.LeaveRoom(room.ID, "host123")
	if err != nil {
		t.Fatalf("Failed to leave room: %v", err)
	}
	
	// Verify room was deleted
	_, err = rm.GetRoom(room.ID)
	if err == nil {
		t.Error("Room should be deleted when all players leave")
	}
}

func TestGetRoom(t *testing.T) {
	rm := NewRoomManager()
	
	// Create a room
	room, _ := rm.CreateRoom("host123", "Host Player", "password123")
	
	// Test getting existing room
	retrievedRoom, err := rm.GetRoom(room.ID)
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}
	
	if retrievedRoom.ID != room.ID {
		t.Errorf("Expected room ID '%s', got '%s'", room.ID, retrievedRoom.ID)
	}
	
	// Test getting non-existent room
	_, err = rm.GetRoom("INVALID")
	if err == nil {
		t.Error("Should fail when room doesn't exist")
	}
}

func TestDeleteRoom(t *testing.T) {
	rm := NewRoomManager()
	
	// Create a room
	room, _ := rm.CreateRoom("host123", "Host Player", "")
	
	// Delete the room
	err := rm.DeleteRoom(room.ID)
	if err != nil {
		t.Fatalf("Failed to delete room: %v", err)
	}
	
	// Verify room was deleted
	_, err = rm.GetRoom(room.ID)
	if err == nil {
		t.Error("Room should be deleted")
	}
	
	// Test deleting non-existent room
	err = rm.DeleteRoom("INVALID")
	if err == nil {
		t.Error("Should fail when room doesn't exist")
	}
}
