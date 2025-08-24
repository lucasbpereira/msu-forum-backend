package models

import "time"

type Vote struct {
	ID        uint64    `json:"id" db:"id"`
	UserID    uint64    `json:"user_id" db:"user_id"`
	PostID    uint64    `json:"post_id" db:"post_id"`
	PostType  string    `json:"post_type" db:"post_type"` // "question" ou "answer"
	Type      int8      `json:"type" db:"type"`           // +1 ou -1
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
