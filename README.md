# Party Race Game — Backend

多人派對賽車遊戲的後端服務，使用 Golang + Gin 框架實作。

## 功能

- **房間管理**：建立、加入、離開遊戲房間（HTTP REST API）
- **WebSocket 信令伺服器**：中繼玩家間的訊號（車輛狀態、輸入、選車確認等）
- **執行緒安全寫入**：每個 WebSocket 連線附帶獨立 mutex，防止 gorilla/websocket 並發寫入 panic
- **遊戲中斷線保護**：遊戲進行中（`InGame = true`）玩家暫時斷線不會觸發房間關閉或離線通知
- **記憶體儲存**：房間資料存於記憶體，無需資料庫

## 技術棧

| 套件 | 版本 | 用途 |
|------|------|------|
| Go | 1.20+ | 程式語言 |
| Gin | v1.9 | HTTP Web 框架 |
| gorilla/websocket | v1.5 | WebSocket 支援 |
| gin-contrib/cors | — | CORS 中介軟體 |

## 專案結構

```
backend/
├── handlers/
│   ├── room_handler.go        # HTTP 房間 API 處理器
│   └── websocket_handler.go   # WebSocket 連線處理器
├── models/
│   ├── room.go                # Room / Player 資料模型
│   └── signal.go              # SignalMessage 資料模型
├── services/
│   ├── room_manager.go        # 房間業務邏輯（建立/加入/離開）
│   └── signaling_server.go    # WebSocket 信令伺服器（含執行緒安全包裝）
├── utils/
│   └── id_generator.go        # 房間 ID / 玩家 ID 產生器
├── main.go                    # 應用程式入口、路由設定
└── go.mod                     # Go 模組定義
```

## 安裝與執行

### 安裝依賴

```bash
cd backend
go mod download
```

### 開發模式執行

```bash
go run main.go
```

伺服器預設在 `http://localhost:8080` 啟動。

### 建置執行檔

```bash
go build -o party-race-game-backend .
./party-race-game-backend
```

> 建置出的執行檔名稱為 `party-race-game-backend`（與 `go.mod` 的 module 名稱一致）。

## 環境變數

| 變數 | 預設值 | 說明 |
|------|--------|------|
| `PORT` | `8080` | 伺服器監聽埠號 |
| `ENV` | —（非 production） | 設為 `production` 時啟用嚴格 CORS，僅允許正式前端網域 |

### CORS 允許來源

| 模式 | 允許來源 |
|------|----------|
| 開發（`ENV` 非 `production`） | `http://localhost:3000`、`http://localhost:3001`、`http://127.0.0.1:3000`，以及正式網域 |
| 正式（`ENV=production`） | `https://party-race-game.vercel.app` |

## API 端點

### 房間管理（HTTP REST）

#### 建立房間
```
POST /api/rooms
Content-Type: application/json

{
  "hostId":   "player-uuid",
  "hostName": "玩家名稱",
  "password": "可選密碼"
}
```

#### 加入房間
```
POST /api/rooms/:roomId/join
Content-Type: application/json

{
  "playerId":   "player-uuid",
  "playerName": "玩家名稱",
  "password":   "可選密碼"
}
```

#### 離開房間
```
DELETE /api/rooms/:roomId/leave
Content-Type: application/json

{
  "playerId": "player-uuid"
}
```

#### 取得房間資訊
```
GET /api/rooms/:roomId
```

### WebSocket 信令連線

```
WS /ws?playerId=<player-uuid>&roomId=<room-id>
```

- `playerId`：必填，玩家唯一識別碼
- `roomId`：必填，連線後自動加入對應的信令房間

連線建立後，伺服器透過此 WebSocket 中繼所有玩家間的訊號訊息。

## 注意事項

- 房間資料儲存於記憶體，伺服器重啟後會遺失
- 當房主在**遊戲尚未開始時**離開，房間會立即關閉並通知所有成員
- 當房主在**遊戲進行中**斷線，房間保持存在，不觸發 `room_closed`
- 最多支援 6 名玩家同時在一個房間
- CORS 限制為 `https://party-race-game.vercel.app`；開發時需設定 `ENV` **不為** `production`（或直接不設定），才能允許 localhost

## 開發工具

```bash
# 執行所有測試
go test ./...

# 程式碼格式化
go fmt ./...

# 靜態分析
go vet ./...
```
