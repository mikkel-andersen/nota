package usecase

import "github.com/mikkel-andersen/nota/domain"

// Repository is the port for persisting notes in a single directory.
type Repository interface {
	Add(body string) (domain.Note, error)
	List() ([]domain.Note, error)
	Delete(id int) error
	Update(id int, body string) error
	MarkDone(id int) error
	Clear() error
}

// Scanner is the port for reading notes across all directories.
type Scanner interface {
	AllNotes() (map[string][]domain.Note, error)
}

// Editor is the port for opening a text editor and returning the edited content.
type Editor interface {
	Open(content string) (string, error)
}
