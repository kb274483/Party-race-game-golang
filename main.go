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

	// Configure CORS - 限制允許的來源
	allowedOrigins := []string{
		"https://party-race-game.vercel.app",
	}
	// 開發環境額外允許 localhost
	if os.Getenv("ENV") != "production" {
		allowedOrigins = append(allowedOrigins,
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
		)
	}
	config := cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}
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
