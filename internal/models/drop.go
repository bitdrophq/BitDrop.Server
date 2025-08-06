package models

import (
	"time"

	"github.com/google/uuid"
)

type Drop struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	GroupID   *uuid.UUID `json:"group_id,omitempty" db:"group_id"` // optional: for group drops
	VideoURL  string     `json:"video_url" db:"video_url"`
	Thumbnail string     `json:"thumbnail" db:"thumbnail"` // store preview image
	Caption   string     `json:"caption,omitempty" db:"caption"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	Votes     int        `json:"votes" db:"votes"`
	Visibility string    `json:"visibility" db:"visibility"` // "private", "public", or "shared"
}
