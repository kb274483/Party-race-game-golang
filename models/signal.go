package models

// SignalMessage represents a WebRTC signaling message
type SignalMessage struct {
	Type     string      `json:"type"`
	RoomID   string      `json:"roomId"`
	SenderID string      `json:"senderId"`
	TargetID string      `json:"targetId,omitempty"`
	Payload  interface{} `json:"payload"`
}

// WebSocketMessage represents a generic WebSocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
