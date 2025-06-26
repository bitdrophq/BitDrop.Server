// handlers/auth_handlers.go
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/richiethie/BitDrop.Server/internal/db"
	"github.com/richiethie/BitDrop.Server/internal/models"
)

func SignUp(c *gin.Context) {
	fmt.Println("📥 SignUp handler hit")

	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("❌ Error parsing JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	fmt.Println("📝 Parsed signup request:", req)

	if db.DB == nil {
		fmt.Println("❌ DB pool is nil!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	var existing string
	err := db.DB.QueryRow(context.Background(), `
		SELECT email FROM users WHERE email = $1 OR username = $2
	`, req.Email, req.Username).Scan(&existing)

	if err != pgx.ErrNoRows && err != nil {
		fmt.Printf("❌ Error checking existing user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email or username already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ Error hashing password: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	fmt.Printf("🚀 Inserting user: %s %s\n", req.Username, req.Email)

	_, err = db.DB.Exec(context.Background(), `
		INSERT INTO users (id, username, email, password, avatar_url, bio, tier, boosts_left, is_admin, created_at, last_active)
		VALUES (gen_random_uuid(), $1, $2, $3, '', '', 'free', 0, false, NOW(), NOW())
	`, req.Username, strings.ToLower(req.Email), string(hashedPassword))

	if err != nil {
		fmt.Printf("❌ Error creating user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	fmt.Println("✅ User created successfully!")
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var storedHashedPassword string
	var user models.User

	err := db.DB.QueryRow(context.Background(), `
		SELECT id, email, username, password, avatar_url, bio, tier, boosts_left, is_admin, created_at, last_active
		FROM users
		WHERE email = $1
	`, strings.ToLower(req.Email)).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&storedHashedPassword,
		&user.AvatarURL,
		&user.Bio,
		&user.Tier,
		&user.BoostsLeft,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.LastActive,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// TODO: Generate and return access token if implementing JWT
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
	})
}

func Logout(c *gin.Context) {
	// Optional: log logout events, audit logs, etc.
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
