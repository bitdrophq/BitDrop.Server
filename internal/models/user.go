package models

import "time"

type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	AvatarURL  string    `json:"avatar_url"`
	Bio        string    `json:"bio"`
	Tier       string    `json:"tier"`
	BoostsLeft int       `json:"boosts_left"`
	IsAdmin    bool      `json:"is_admin"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
}
