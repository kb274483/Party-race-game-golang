package services

import (
	"errors"
	"party-race-game-backend/models"
	"party-race-game-backend/utils"
	"sync"
	"time"
)

// RoomManager manages game rooms in memory
type RoomManager struct {
	rooms map[string]*models.Room
	mu    sync.RWMutex
}

// NewRoomManager creates a new RoomManager instance
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*models.Room),
	}
}

// CreateRoom creates a new room with a unique ID
func (rm *RoomManager) CreateRoom(hostID, hostName, password string) (*models.Room, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	roomID := utils.GenerateRoomID()
	
	room := &models.Room{
		ID:         roomID,
		HostID:     hostID,
		Password:   password,
		Players:    make(map[string]*models.Player),
		MaxPlayers: 6,
		CreatedAt:  time.Now(),
	}

	// Add host as first player
	room.Players[hostID] = &models.Player{
		ID:       hostID,
		Name:     hostName,
		IsHost:   true,
		JoinedAt: time.Now(),
	}

	rm.rooms[roomID] = room
	return room, nil
}

// JoinRoom adds a player to an existing room
func (rm *RoomManager) JoinRoom(roomID, playerID, playerName, password string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return errors.New("room not found")
	}

	// Check password if room is password protected
	if room.Password != "" && room.Password != password {
		return errors.New("incorrect password")
	}

	// Check if room is full
	if len(room.Players) >= room.MaxPlayers {
		return errors.New("room is full")
	}

	// Check if player already in room
	if _, exists := room.Players[playerID]; exists {
		return errors.New("player already in room")
	}

	// Add player to room
	room.Players[playerID] = &models.Player{
		ID:       playerID,
		Name:     playerName,
		IsHost:   false,
		JoinedAt: time.Now(),
	}

	return nil
}

// LeaveRoom removes a player from a room
func (rm *RoomManager) LeaveRoom(roomID, playerID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return errors.New("room not found")
	}

	player, exists := room.Players[playerID]
	if !exists {
		return errors.New("player not in room")
	}

	// Remove player
	delete(room.Players, playerID)

	// If host leaves or room is empty, delete the room
	if player.IsHost || len(room.Players) == 0 {
		delete(rm.rooms, roomID)
	}

	return nil
}

// GetRoom retrieves a room by ID
func (rm *RoomManager) GetRoom(roomID string) (*models.Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, errors.New("room not found")
	}

	return room, nil
}

// SetRoomInGame marks a room as having started the game
func (rm *RoomManager) SetRoomInGame(roomID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return errors.New("room not found")
	}
	room.InGame = true
	return nil
}

// DeleteRoom removes a room from memory
func (rm *RoomManager) DeleteRoom(roomID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.rooms[roomID]; !exists {
		return errors.New("room not found")
	}

	delete(rm.rooms, roomID)
	return nil
}
