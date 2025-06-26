package services

import (
	"context"
	"log"
	"time"

	"github.com/richiethie/BitDrop.Server/internal/db"
	"github.com/richiethie/BitDrop.Server/internal/models"
)

// GetOrCreateUser tries to fetch the user by ID.
// If not found, it inserts a new user and returns the created record.
func GetOrCreateUser(id string, email string) (*models.User, error) {
	var user models.User

	// 1. Try to fetch existing user
	err := db.DB.QueryRow(context.Background(), `
		SELECT id, email, username, avatar_url, bio, tier, boosts_left, is_admin, created_at, last_active
		FROM users WHERE id = $1
	`, id).Scan(
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

	if err == nil {
		return &user, nil // user found, return
	}

	// 2. Insert new user
	defaultUsername := "newuser" // or you can generate from email, e.g., email prefix

	_, err = db.DB.Exec(context.Background(), `
		INSERT INTO users (
			id, email, username, avatar_url, bio, tier, boosts_left, is_admin, created_at, last_active
		)
		VALUES ($1, $2, $3, '', '', 'free', 0, false, NOW(), NOW())
	`, id, email, defaultUsername)

	if err != nil {
		log.Printf("❌ Error inserting user: %v", err)
		return nil, err
	}

	// 3. Return the inserted user data (mirrors default values)
	now := time.Now()
	user = models.User{
		ID:         id,
		Email:      email,
		Username:   defaultUsername,
		AvatarURL:  "",
		Bio:        "",
		Tier:       "free",
		BoostsLeft: 0,
		IsAdmin:    false,
		CreatedAt:  now,
		LastActive: now,
	}

	return &user, nil
}
