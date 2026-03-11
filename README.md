# Party Race Game - Backend

多人派對賽車遊戲的後端服務，使用 Golang 和 Gin 框架實作。

## 功能

- **房間管理**: 建立、加入、離開遊戲房間
- **WebSocket 信令伺服器**: 中繼 WebRTC 信令訊息
- **記憶體儲存**: 房間資料儲存在記憶體中，無需資料庫

## 技術棧

- **Golang 1.21+**
- **Gin**: Web 框架
- **Gorilla WebSocket**: WebSocket 支援
- **CORS 中介軟體**: 跨域請求處理

## 專案結構

```
backend/
├── handlers/          # HTTP 和 WebSocket 處理器
│   ├── room_handler.go
│   └── websocket_handler.go
├── models/            # 資料模型
│   ├── room.go
│   └── signal.go
├── services/          # 業務邏輯
│   ├── room_manager.go
│   └── signaling_server.go
├── utils/             # 工具函數
│   └── id_generator.go
├── main.go            # 應用程式入口
└── go.mod             # Go 模組定義
```

## 安裝與執行

### 安裝依賴

```bash
cd backend
go mod download
```

### 執行伺服器

```bash
go run main.go
```

伺服器將在 `http://localhost:8080` 啟動。

### 建置執行檔

```bash
go build -o bin/server main.go
./bin/server
```

## API 端點

### 房間管理

#### 建立房間
```
POST /api/rooms
Content-Type: application/json

{
  "hostId": "player-uuid",
  "password": "optional-password"
}
```

#### 加入房間
```
POST /api/rooms/:roomId/join
Content-Type: application/json

{
  "playerId": "player-uuid",
  "playerName": "Player Name",
  "password": "optional-password"
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

### WebSocket 連線

```
WS /ws?playerId=player-uuid
```

## 開發

### 執行測試

```bash
go test ./...
```

### 程式碼格式化

```bash
go fmt ./...
```

## 環境變數

目前不需要環境變數配置。未來可以新增：

- `PORT`: 伺服器埠號（預設 8080）
- `ALLOWED_ORIGINS`: CORS 允許的來源

## 注意事項

- 房間資料儲存在記憶體中，伺服器重啟後會遺失
- 當房間內所有玩家離開時，房間會自動刪除
- 當房主離開時，房間會立即關閉
- 最多支援 6 名玩家同時在一個房間
