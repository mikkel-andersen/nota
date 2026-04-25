package domain

import "time"

type Note struct {
	ID        int
	Body      string
	CreatedAt time.Time
	Done      bool
}
