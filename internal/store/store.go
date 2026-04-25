package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Note struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Done      bool      `json:"done"`
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
	if err := json.Unmarshal(b, &d); err != nil {
		return d, err
	}
	// migrate: derive NextID from max existing ID for files written before next_id was tracked
	if d.NextID == 0 && len(d.Notes) > 0 {
		for _, n := range d.Notes {
			if n.ID > d.NextID {
				d.NextID = n.ID
			}
		}
	}
	return d, nil
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

func (s *Store) MarkDone(id int) error {
	d, err := s.load()
	if err != nil {
		return err
	}
	for i, n := range d.Notes {
		if n.ID == id {
			d.Notes[i].Done = true
			return s.save(d)
		}
	}
	return fmt.Errorf("note #%d not found", id)
}

func (s *Store) Update(id int, body string) error {
	d, err := s.load()
	if err != nil {
		return err
	}
	for i, n := range d.Notes {
		if n.ID == id {
			d.Notes[i].Body = body
			return s.save(d)
		}
	}
	return fmt.Errorf("note #%d not found", id)
}

func (s *Store) Clear() error {
	d, err := s.load()
	if err != nil {
		return err
	}
	return s.save(data{NextID: d.NextID})
}
