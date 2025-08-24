package models

import "time"

type Answer struct {
	ID         uint64    `json:"id" db:"id"`
	QuestionID uint64    `json:"question_id" db:"question_id"`
	UserID     uint64    `json:"user_id" db:"user_id"`
	Body       string    `json:"body" db:"body"`
	Votes      int32     `json:"votes" db:"votes"`
	IsAccepted bool      `json:"is_accepted" db:"is_accepted"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// Otimização: cache do usuário
	user *User
}
