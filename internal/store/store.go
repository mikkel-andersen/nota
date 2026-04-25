package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Note struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	path string
}

type data struct {
	Notes  []Note `json:"notes"`
	NextID int    `json:"next_id"`
}

func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Store{path: filepath.Join(dir, "notes.json")}, nil
}

func (s *Store) load() (data, error) {
	var d data
	b, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	return d, json.Unmarshal(b, &d)
}

func (s *Store) save(d data) error {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0644)
}

func (s *Store) Add(body string) (Note, error) {
	d, err := s.load()
	if err != nil {
		return Note{}, err
	}
	d.NextID++
	n := Note{ID: d.NextID, Body: body, CreatedAt: time.Now()}
	d.Notes = append(d.Notes, n)
	return n, s.save(d)
}

func (s *Store) List() ([]Note, error) {
	d, err := s.load()
	return d.Notes, err
}

func (s *Store) Delete(id int) error {
	d, err := s.load()
	if err != nil {
		return err
	}
	notes := d.Notes[:0]
	for _, n := range d.Notes {
		if n.ID != id {
			notes = append(notes, n)
		}
	}
	d.Notes = notes
	return s.save(d)
}

func (s *Store) Clear() error {
	return s.save(data{})
}
