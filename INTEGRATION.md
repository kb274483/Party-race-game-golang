# WebSocket 信令伺服器整合說明

本文件說明 SignalingServer 與 RoomManager 的整合方式，以及 WebSocket 連線生命週期的處理邏輯。

## 架構概覽

```
Client ──── HTTP REST ────► RoomHandler ──► RoomManager
       ──── WebSocket ────► WebSocketHandler ──► SignalingServer ──► RoomManager
```

### SignalingServer（`services/signaling_server.go`）

- 管理所有玩家的 WebSocket 連線（`connections` map）
- 依房間分組連線（`rooms` map）
- 中繼訊號訊息（點對點 or 廣播）
- 處理玩家斷線事件
- **執行緒安全**：每個連線以 `wsConn`（含寫入 mutex）包裝，防止 gorilla/websocket 並發寫入 panic

### RoomManager（`services/room_manager.go`）

- 管理房間狀態（玩家清單、房主、密碼、遊戲進行中旗標）
- 驗證房間規則（最多 6 人、密碼保護）
- 房間空時自動刪除

## 執行緒安全設計

gorilla/websocket 不允許多個 goroutine 同時對同一連線呼叫 `WriteJSON`。
系統以 `wsConn` 結構體包裝每個連線，並附加獨立的寫入 mutex：

```go
type wsConn struct {
    conn *websocket.Conn
    wmu  sync.Mutex
}

func (c *wsConn) WriteJSON(v interface{}) error {
    c.wmu.Lock()
    defer c.wmu.Unlock()
    return c.conn.WriteJSON(v)
}
```

`connections` 與 `rooms` 兩個 map 共用同一個 `*wsConn` 實例，確保 mutex 不會重複建立。

## WebSocket 連線流程

### 1. 玩家連線
```
Client ──GET /ws?playerId=<id>&roomId=<roomId>──► Server
Server: Upgrade → WebSocket
Server: RegisterConnection(playerId, conn)   → connections[playerId] = wsConn
Server: JoinSignalingRoom(roomId, playerId)  → rooms[roomId][playerId] = wsConn（重用同一實例）
```

### 2. 玩家加入房間（HTTP）
```
Client ──POST /api/rooms/:roomId/join──► Server
Server: RoomManager.JoinRoom()
Server: BroadcastToRoom("player_joined", room)
```

### 3. 玩家主動離開（HTTP）
```
Client ──DELETE /api/rooms/:roomId/leave──► Server
Server: RoomManager.LeaveRoom()
Server: BroadcastToRoom("player_left") 或 "room_closed"
Server: LeaveSignalingRoom(roomId, playerId)
```

### 4. 玩家斷線（WebSocket 關閉）
```
WebSocket 關閉
Server: HandlePlayerDisconnect(playerId, roomId)
  ├── 取得房間資訊（含 InGame 狀態）
  ├── 釋放 ss.mu 鎖後再次讀取最新 InGame（防止競爭條件）
  ├── InGame = true  → 跳過房間刪除與 player_disconnect 廣播
  └── InGame = false → 移除玩家、廣播 player_disconnect
                       若為房主 → 廣播 room_closed
Server: UnregisterConnection(playerId)
```

### 5. 遊戲開始（game_started 訊號）
```
Host ──game_started──► Server
Server: RoomManager.SetRoomInGame(roomId)  → room.InGame = true
Server: Relay "game_started" to all room members
```

> **設計要點**：`game_started` 訊號由房主透過 WebSocket 發送，後端收到後立即設定 `InGame = true`。
> 此後任何玩家的 WebSocket 斷線（例如頁面切換過渡）都不會廣播 `player_disconnect`，避免假性離線通知。

## 訊息類型

### `player_joined`
玩家透過 HTTP API 加入房間時廣播。
```json
{
  "type": "player_joined",
  "roomId": "ABC123",
  "senderId": "player1",
  "payload": {
    "playerId": "player1",
    "playerName": "玩家一",
    "room": { }
  }
}
```

### `player_left`
玩家透過 HTTP API 主動離開時廣播。
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

### `player_disconnect`
玩家 WebSocket 非預期關閉，且目前**不在遊戲進行中**時廣播。
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

### `room_closed`
房主離開或斷線（且遊戲尚未開始）時廣播。
```json
{
  "type": "room_closed",
  "roomId": "ABC123",
  "payload": {
    "reason": "host_disconnect"
  }
}
```

### `game_started`
房主點擊開始遊戲時廣播，攜帶完整玩家清單。
```json
{
  "type": "game_started",
  "roomId": "ABC123",
  "senderId": "host-player-id",
  "payload": {
    "players": [
      { "id": "p1", "name": "玩家一", "isHost": true },
      { "id": "p2", "name": "玩家二", "isHost": false }
    ]
  }
}
```

### 遊戲訊號（後端直接中繼，不做業務邏輯處理）

| type | 說明 |
|------|------|
| `car_state` | 車輛位置 / 旋轉 / 速度（20Hz） |
| `player_input` | 玩家按鍵輸入（非物理權威玩家發送） |
| `car_confirm` | 選車確認（選車代表發送） |
| `game_action` | 遊戲結束後的動作（再玩一次 / 回到房間） |

## 競爭條件處理

**情境**：`game_started` 處理 goroutine 與玩家斷線 goroutine 並發執行時，斷線 goroutine 可能讀到舊的 `InGame = false`。

**解法**：`HandlePlayerDisconnect` 在釋放 `ss.mu` 後，立即從 RoomManager 重新讀取最新的房間狀態：

```go
// 釋放鎖後再次確認，消除 game_started 並發競爭條件
if roomID != "" {
    if freshRoom, err := ss.roomManager.GetRoom(roomID); err == nil && freshRoom != nil {
        room = freshRoom
    }
}
inGame := room != nil && room.InGame
```

## 測試

```bash
cd backend
go test ./services -v
```

主要測試案例：
- `TestHandlePlayerDisconnect`：驗證玩家斷線後移除邏輯
- `TestHandlePlayerDisconnectHostLeaves`：驗證房主斷線後房間刪除
- `TestHandleMessage`：驗證訊息處理與房間加入流程
