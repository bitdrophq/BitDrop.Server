// handlers/availability.go
package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/richiethie/BitDrop.Server/internal/db"
)

func CheckAvailability(c *gin.Context) {
	username := c.Query("username")
	email := c.Query("email")

	var exists bool

	// Check username
	err := db.DB.QueryRow(context.Background(), `
		SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)
	`, username).Scan(&exists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking username"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	}

	// Check email
	err = db.DB.QueryRow(context.Background(), `
		SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)
	`, email).Scan(&exists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking email"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Available"})
}
