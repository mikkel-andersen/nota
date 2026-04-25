package usecase

import (
	"errors"
	"fmt"

	"github.com/mikkel-andersen/nota/domain"
)

var ErrNoChanges = errors.New("no changes")
var ErrNoteNotFound = errors.New("note not found")

type NoteService struct {
	repo    Repository
	scanner Scanner
	editor  Editor
}

func NewNoteService(repo Repository, scanner Scanner, editor Editor) *NoteService {
	return &NoteService{repo: repo, scanner: scanner, editor: editor}
}

func (s *NoteService) Add(body string) (domain.Note, error) {
	return s.repo.Add(body)
}

func (s *NoteService) List() ([]domain.Note, error) {
	return s.repo.List()
}

func (s *NoteService) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *NoteService) MarkDone(id int) error {
	return s.repo.MarkDone(id)
}

func (s *NoteService) Clear() error {
	return s.repo.Clear()
}

func (s *NoteService) AllNotes() (map[string][]domain.Note, error) {
	return s.scanner.AllNotes()
}

func (s *NoteService) Edit(id int) error {
	notes, err := s.repo.List()
	if err != nil {
		return err
	}
	var current string
	for _, n := range notes {
		if n.ID == id {
			current = n.Body
			break
		}
	}
	if current == "" {
		return fmt.Errorf("%w: #%d", ErrNoteNotFound, id)
	}
	updated, err := s.editor.Open(current)
	if err != nil {
		return err
	}
	if updated == "" || updated == current {
		return ErrNoChanges
	}
	return s.repo.Update(id, updated)
}
