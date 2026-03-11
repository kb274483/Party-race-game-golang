# WebSocket and Room Manager Integration

This document describes how the signaling server integrates with the room manager to handle player connections and disconnections.

## Overview

The integration ensures that:
1. WebSocket connections are properly associated with rooms
2. Player disconnections automatically update room state
3. Other players are notified when someone joins, leaves, or disconnects
4. Rooms are automatically cleaned up when the host leaves or disconnects

## Components

### SignalingServer
- Manages WebSocket connections
- Relays WebRTC signaling messages
- Broadcasts messages to rooms
- Handles player disconnections

### RoomManager
- Manages room state (players, host, passwords)
- Enforces room rules (max players, password protection)
- Automatically deletes rooms when host leaves or room is empty

## WebSocket Connection Flow

### 1. Player Connects
```
Client -> GET /ws?playerId=<id>&roomId=<roomId>
Server -> Upgrade to WebSocket
Server -> Register connection in SignalingServer
Server -> Join signaling room (if roomId provided)
```

### 2. Player Joins Room (via HTTP)
```
Client -> POST /api/rooms/:roomId/join
Server -> Add player to RoomManager
Server -> Broadcast "player_joined" message to room via WebSocket
```

### 3. Player Leaves Room (via HTTP)
```
Client -> DELETE /api/rooms/:roomId/leave
Server -> Get player info from RoomManager
Server -> Remove player from RoomManager
Server -> Broadcast "player_left" or "room_closed" message
Server -> Remove from signaling room
```

### 4. Player Disconnects (WebSocket closes)
```
WebSocket closes
Server -> HandlePlayerDisconnect()
Server -> Find player's room
Server -> Remove player from RoomManager
Server -> Broadcast "player_disconnect" message
Server -> If host disconnected, broadcast "room_closed"
Server -> Unregister connection
```

## Message Types

### player_joined
Sent when a player joins a room via HTTP API.
```json
{
  "type": "player_joined",
  "roomId": "ABC123",
  "senderId": "player1",
  "payload": {
    "playerId": "player1",
    "playerName": "Player One",
    "room": { /* full room object */ }
  }
}
```

### player_left
Sent when a non-host player leaves a room via HTTP API.
```json
{
  "type": "player_left",
  "roomId": "ABC123",
  "senderId": "player1",
  "payload": {
    "playerId": "player1"
  }
}
```

### player_disconnect
Sent when a player's WebSocket connection closes unexpectedly.
```json
{
  "type": "player_disconnect",
  "roomId": "ABC123",
  "senderId": "player1",
  "payload": {
    "playerId": "player1",
    "isHost": false
  }
}
```

### room_closed
Sent when the host leaves or disconnects, causing the room to close.
```json
{
  "type": "room_closed",
  "roomId": "ABC123",
  "payload": {
    "reason": "host_disconnect" | "host_left"
  }
}
```

## Requirements Satisfied

This integration satisfies the following requirements:

- **1.1**: Room creation and management
- **1.4**: Room joining with validation
- **10.3**: Host leaving closes the room
- **10.5**: Players are notified when someone disconnects

## Testing

Run the integration tests:
```bash
cd backend
go test ./services -v
```

Key test cases:
- `TestHandlePlayerDisconnect`: Verifies player removal on disconnect
- `TestHandlePlayerDisconnectHostLeaves`: Verifies room deletion when host disconnects
- `TestHandleMessage`: Verifies message handling and room joining
