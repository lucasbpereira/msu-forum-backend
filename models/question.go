package models

import "time"

type Question struct {
	ID          uint64    `json:"id" db:"id"`
	UserID      uint64    `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Body        string    `json:"body" db:"body"`
	Votes       int32     `json:"votes" db:"votes"`
	ViewCount   uint32    `json:"view_count" db:"view_count"`
	AnswerCount uint32    `json:"answer_count" db:"answer_count"`
	IsSolved    bool      `json:"is_solved" db:"is_solved"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relacionamentos
	User    *User    `json:"user,omitempty"`
	Tags    []Tag    `json:"tags,omitempty"`
	Answers []Answer `json:"answers,omitempty"`
}
