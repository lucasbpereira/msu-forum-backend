package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID         int            `json:"id" db:"id"`
	Username   sql.NullString `json:"username" db:"username"`
	Email      sql.NullString `json:"email" db:"email"`
	Password   sql.NullString `json:"password" db:"password"`
	Phone      sql.NullString `json:"phone" db:"phone"`
	Reputation int32          `json:"reputation" db:"reputation"`
	Role       string         `json:"role" db:"role"`
	Wallet     string         `json:"wallet" db:"wallet"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
	LastSeen   time.Time      `json:"last_seen" db:"last_seen"`
	IsActive   bool           `json:"is_active" db:"is_active"`
	AvatarURL  sql.NullString `json:"avatar_url" db:"avatar_url"`
}
