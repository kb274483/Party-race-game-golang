package handlers

import (
	"net/http"
	"party-race-game-backend/models"
	"party-race-game-backend/services"

	"github.com/gin-gonic/gin"
)

// RoomHandler handles room-related HTTP requests
type RoomHandler struct {
	roomManager     *services.RoomManager
	signalingServer *services.SignalingServer
}

// NewRoomHandler creates a new RoomHandler instance
func NewRoomHandler(roomManager *services.RoomManager, signalingServer *services.SignalingServer) *RoomHandler {
	return &RoomHandler{
		roomManager:     roomManager,
		signalingServer: signalingServer,
	}
}

// CreateRoom handles room creation requests
func (rh *RoomHandler) CreateRoom(c *gin.Context) {
	var req models.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := rh.roomManager.CreateRoom(req.HostID, req.HostName, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, room)
}

// JoinRoom handles room join requests
func (rh *RoomHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	
	var req models.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := rh.roomManager.JoinRoom(roomID, req.PlayerID, req.PlayerName, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, _ := rh.roomManager.GetRoom(roomID)

	// Notify other players in the room that a new player joined
	joinMsg := models.SignalMessage{
		Type:     "player_joined",
		RoomID:   roomID,
		SenderID: req.PlayerID,
		Payload: map[string]interface{}{
			"playerId":   req.PlayerID,
			"playerName": req.PlayerName,
			"room":       room,
		},
	}
	rh.signalingServer.BroadcastToRoom(roomID, joinMsg)

	c.JSON(http.StatusOK, room)
}

// LeaveRoom handles room leave requests
func (rh *RoomHandler) LeaveRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	
	var req models.LeaveRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get player info before leaving
	room, err := rh.roomManager.GetRoom(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	player, exists := room.Players[req.PlayerID]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player not in room"})
		return
	}

	isHost := player.IsHost

	// Remove player from room
	err = rh.roomManager.LeaveRoom(roomID, req.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Notify other players
	if isHost {
		// Host left, room is closing
		closeMsg := models.SignalMessage{
			Type:   "room_closed",
			RoomID: roomID,
			Payload: map[string]interface{}{
				"reason": "host_left",
			},
		}
		rh.signalingServer.BroadcastToRoom(roomID, closeMsg)
	} else {
		// Regular player left
		leaveMsg := models.SignalMessage{
			Type:     "player_left",
			RoomID:   roomID,
			SenderID: req.PlayerID,
			Payload: map[string]interface{}{
				"playerId": req.PlayerID,
			},
		}
		rh.signalingServer.BroadcastToRoom(roomID, leaveMsg)
	}

	// Remove from signaling room
	rh.signalingServer.LeaveSignalingRoom(roomID, req.PlayerID)

	c.JSON(http.StatusOK, gin.H{"message": "Left room successfully"})
}

// GetRoom handles room info requests
func (rh *RoomHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	
	room, err := rh.roomManager.GetRoom(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, room)
}
