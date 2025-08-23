package models

import "time"

type User struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password,omitempty"` // não exibir no JSON
	Role      string    `db:"role" json:"role"`
	Phone     string    `db:"phone" json:"phone"`
	Wallet    string    `db:"wallet" json:"wallet"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
