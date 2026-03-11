package models

import "time"

// Room represents a game room
type Room struct {
	ID         string             `json:"id"`
	HostID     string             `json:"hostId"`
	Password   string             `json:"password,omitempty"`
	Players    map[string]*Player `json:"players"`
	MaxPlayers int                `json:"maxPlayers"`
	InGame     bool               `json:"inGame"`
	CreatedAt  time.Time          `json:"createdAt"`
}

// Player represents a player in a room
type Player struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	IsHost   bool      `json:"isHost"`
	JoinedAt time.Time `json:"joinedAt"`
}

// CreateRoomRequest represents the request to create a room
type CreateRoomRequest struct {
	HostID   string `json:"hostId" binding:"required"`
	HostName string `json:"hostName" binding:"required"`
	Password string `json:"password"`
}

// JoinRoomRequest represents the request to join a room
type JoinRoomRequest struct {
	PlayerID   string `json:"playerId" binding:"required"`
	PlayerName string `json:"playerName" binding:"required"`
	Password   string `json:"password"`
}

// LeaveRoomRequest represents the request to leave a room
type LeaveRoomRequest struct {
	PlayerID string `json:"playerId" binding:"required"`
}
