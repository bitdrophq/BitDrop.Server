package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/richiethie/BitDrop.Server/internal/db"
	"github.com/richiethie/BitDrop.Server/internal/models"
)

func GetProfile(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userId, ok := userIdVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid userId format"})
		return
	}

	var user models.User

	err := db.DB.QueryRow(context.Background(), `
		SELECT id, email, username, avatar_url, bio, tier, boosts_left, is_admin, created_at, last_active
		FROM users
		WHERE id = $1
	`, userId).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.AvatarURL,
		&user.Bio,
		&user.Tier,
		&user.BoostsLeft,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.LastActive,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
