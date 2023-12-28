package models

import "time"

type Paste struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}
