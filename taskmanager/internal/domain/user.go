package domain

import "time"

type User struct{
	ID           int       `json:"id"`
	Email        int       `json:"email"`
	PasswordHash int       `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
