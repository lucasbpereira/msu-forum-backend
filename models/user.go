package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID         int            `json:"id" db:"id"`
	Username   string         `json:"username" db:"username"`
	Email      string         `json:"email" db:"email"`
	Password   string         `json:"password" db:"password"`
	Reputation int32          `json:"reputation" db:"reputation"`
	Role       string         `db:"role" json:"role"`
	Phone      string         `json:"phone" db:"phone"`
	Wallet     string         `json:"wallet" db:"wallet"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
	LastSeen   time.Time      `json:"last_seen" db:"last_seen"`
	IsActive   bool           `json:"is_active" db:"is_active"`
	AvatarURL  sql.NullString `json:"avatar_url" db:"avatar_url"`
}
