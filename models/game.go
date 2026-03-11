package models

// Vector3 represents a 3D vector
type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Quaternion represents a rotation quaternion
type Quaternion struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	W float64 `json:"w"`
}

// BoundingBox represents an axis-aligned bounding box
type BoundingBox struct {
	Min Vector3 `json:"min"`
	Max Vector3 `json:"max"`
}

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseWaiting   GamePhase = "waiting"
	GamePhaseCountdown GamePhase = "countdown"
	GamePhaseRacing    GamePhase = "racing"
	GamePhaseFinished  GamePhase = "finished"
)

// ControlType represents a control input type
type ControlType string

const (
	ControlTypeAccelerate ControlType = "accelerate"
	ControlTypeBrake      ControlType = "brake"
	ControlTypeTurnLeft   ControlType = "turn_left"
	ControlTypeTurnRight  ControlType = "turn_right"
)

// RaceCar represents a race car in the game
type RaceCar struct {
	ID            string     `json:"id"`
	Position      Vector3    `json:"position"`
	Rotation      Quaternion `json:"rotation"`
	Velocity      Vector3    `json:"velocity"`
	Speed         float64    `json:"speed"`
	MaxSpeed      float64    `json:"maxSpeed"`
	Acceleration  float64    `json:"acceleration"`
	TurnSpeed     float64    `json:"turnSpeed"`
	HasSpeedBoost bool       `json:"hasSpeedBoost"`
	BoostEndTime  int64      `json:"boostEndTime"`
}

// Obstacle represents an obstacle on the track
type Obstacle struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"` // "wall" or "mine"
	Position    Vector3     `json:"position"`
	Size        Vector3     `json:"size"`
	BoundingBox BoundingBox `json:"boundingBox"`
}

// SpeedBoost represents a speed boost on the track
type SpeedBoost struct {
	ID       string  `json:"id"`
	Position Vector3 `json:"position"`
	Radius   float64 `json:"radius"`
	Active   bool    `json:"active"`
}

// RaceTrack represents the race track
type RaceTrack struct {
	ID            string        `json:"id"`
	StartPosition Vector3       `json:"startPosition"`
	Bounds        BoundingBox   `json:"bounds"`
	Obstacles     []Obstacle    `json:"obstacles"`
	SpeedBoosts   []SpeedBoost  `json:"speedBoosts"`
}

// Team represents a team in the game
type Team struct {
	ID        int      `json:"id"`
	PlayerIDs []string `json:"playerIds"`
	CarID     string   `json:"carId"`
	Score     float64  `json:"score"`
}

// ControlAssignment represents control assignment for a player
type ControlAssignment struct {
	PlayerID string        `json:"playerId"`
	TeamID   int           `json:"teamId"`
	Controls []ControlType `json:"controls"`
}

// GameState represents the complete game state
type GameState struct {
	GameID             string                       `json:"gameId"`
	RoomID             string                       `json:"roomId"`
	Timestamp          int64                        `json:"timestamp"`
	SequenceNumber     int                          `json:"sequenceNumber"`
	Phase              GamePhase                    `json:"phase"`
	CountdownTime      float64                      `json:"countdownTime"`
	RaceTime           float64                      `json:"raceTime"`
	Track              RaceTrack                    `json:"track"`
	Cars               map[string]*RaceCar          `json:"cars"`
	Teams              map[int]*Team                `json:"teams"`
	ControlAssignments map[string]*ControlAssignment `json:"controlAssignments"`
	Scores             map[int]float64              `json:"scores"`
}

// InputState represents player input state
type InputState struct {
	Accelerate     bool `json:"accelerate"`
	Brake          bool `json:"brake"`
	TurnLeft       bool `json:"turnLeft"`
	TurnRight      bool `json:"turnRight"`
	SequenceNumber int  `json:"sequenceNumber"`
}

// InputMessage represents an input message from player to host
type InputMessage struct {
	Type      string     `json:"type"`
	PlayerID  string     `json:"playerId"`
	Input     InputState `json:"input"`
	Timestamp int64      `json:"timestamp"`
}

// StateUpdateMessage represents a state update from host to players
type StateUpdateMessage struct {
	Type      string    `json:"type"`
	GameState GameState `json:"gameState"`
	Timestamp int64     `json:"timestamp"`
}

// GameStartMessage represents a game start message
type GameStartMessage struct {
	Type               string              `json:"type"`
	Track              RaceTrack           `json:"track"`
	ControlAssignments []ControlAssignment `json:"controlAssignments"`
	Timestamp          int64               `json:"timestamp"`
}

// WinnerInfo represents winner information
type WinnerInfo struct {
	WinnerTeamID   *int            `json:"winnerTeamId"`
	Scores         map[int]float64 `json:"scores"`
	IsSinglePlayer bool            `json:"isSinglePlayer"`
}

// GameEndMessage represents a game end message
type GameEndMessage struct {
	Type        string          `json:"type"`
	Winner      WinnerInfo      `json:"winner"`
	FinalScores map[int]float64 `json:"finalScores"`
	Timestamp   int64           `json:"timestamp"`
}

// PlayerDisconnectMessage represents a player disconnect message
type PlayerDisconnectMessage struct {
	Type       string `json:"type"`
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
	Timestamp  int64  `json:"timestamp"`
}
