package main

import (
	"log"
	"os"
	"party-race-game-backend/handlers"
	"party-race-game-backend/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize services
	roomManager := services.NewRoomManager()
	signalingServer := services.NewSignalingServer(roomManager)

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS - 允許所有來源（含區域網路 IP，適用本地派對遊戲）
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	router.Use(cors.New(config))

	// Initialize handlers
	roomHandler := handlers.NewRoomHandler(roomManager, signalingServer)
	wsHandler := handlers.NewWebSocketHandler(signalingServer)

	// Setup routes
	api := router.Group("/api")
	{
		// Room management routes
		api.POST("/rooms", roomHandler.CreateRoom)
		api.POST("/rooms/:roomId/join", roomHandler.JoinRoom)
		api.DELETE("/rooms/:roomId/leave", roomHandler.LeaveRoom)
		api.GET("/rooms/:roomId", roomHandler.GetRoom)
	}

	// WebSocket route for signaling
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Starting server on :" + port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
