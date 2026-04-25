package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mikkel-andersen/nota/domain"
)

func notFound(id int) error {
	return fmt.Errorf("%w: #%d", domain.ErrNoteNotFound, id)
}

type jsonNote struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Done      bool      `json:"done"`
}

type jsonData struct {
	Notes  []jsonNote `json:"notes"`
	NextID int        `json:"next_id"`
}

type JSONRepo struct {
	path string
}

func NewJSONRepo(dir string) (*JSONRepo, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &JSONRepo{path: filepath.Join(dir, "notes.json")}, nil
}

func (r *JSONRepo) load() (jsonData, error) {
	var d jsonData
	b, err := os.ReadFile(r.path)
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

func (r *JSONRepo) save(d jsonData) error {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, b, 0644)
}

func toDomain(n jsonNote) domain.Note {
	return domain.Note{ID: n.ID, Body: n.Body, CreatedAt: n.CreatedAt, Done: n.Done}
}

func (r *JSONRepo) Add(body string) (domain.Note, error) {
	d, err := r.load()
	if err != nil {
		return domain.Note{}, err
	}
	d.NextID++
	n := jsonNote{ID: d.NextID, Body: body, CreatedAt: time.Now()}
	d.Notes = append(d.Notes, n)
	return toDomain(n), r.save(d)
}

func (r *JSONRepo) List() ([]domain.Note, error) {
	d, err := r.load()
	if err != nil {
		return nil, err
	}
	notes := make([]domain.Note, len(d.Notes))
	for i, n := range d.Notes {
		notes[i] = toDomain(n)
	}
	return notes, nil
}

func (r *JSONRepo) Delete(id int) error {
	d, err := r.load()
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
	return r.save(d)
}

func (r *JSONRepo) Update(id int, body string) error {
	d, err := r.load()
	if err != nil {
		return err
	}
	for i, n := range d.Notes {
		if n.ID == id {
			d.Notes[i].Body = body
			return r.save(d)
		}
	}
	return notFound(id)
}

func (r *JSONRepo) MarkDone(id int) error {
	d, err := r.load()
	if err != nil {
		return err
	}
	for i, n := range d.Notes {
		if n.ID == id {
			d.Notes[i].Done = true
			return r.save(d)
		}
	}
	return notFound(id)
}

func (r *JSONRepo) Clear() error {
	d, err := r.load()
	if err != nil {
		return err
	}
	return r.save(jsonData{NextID: d.NextID})
}
