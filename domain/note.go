package domain

import (
	"errors"
	"time"
)

var ErrNoteNotFound = errors.New("note not found")

type Note struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Done      bool      `json:"done"`
}
