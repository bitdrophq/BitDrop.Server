package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/richiethie/BitDrop.Server/internal/db"
	"github.com/richiethie/BitDrop.Server/internal/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create or open the log file in append mode
	logFile, err := os.OpenFile("gin.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open log file: %v", err)
	}
	defer logFile.Close()

	// Write logs to both terminal and file
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)

	// Initialize Gin engine with logger and recovery
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Public route
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	routes.RegisterRoutes(r)

	// Run server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s\n", port)
	r.Run("0.0.0.0:" + port)
}
