package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mikkel-andersen/nota/domain"
)

type jsonNote struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Done      bool      `json:"done"`
}

type jsonData struct {
	Notes []jsonNote `json:"notes"`
}

type JSONScanner struct {
	baseDir string
}

func NewJSONScanner(baseDir string) *JSONScanner {
	return &JSONScanner{baseDir: baseDir}
}

func (s *JSONScanner) AllNotes() (map[string][]domain.Note, error) {
	result := make(map[string][]domain.Note)
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	replacer := strings.NewReplacer("_", "/")
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.baseDir, e.Name(), "notes.json"))
		if err != nil {
			continue
		}
		var d jsonData
		if err := json.Unmarshal(b, &d); err != nil || len(d.Notes) == 0 {
			continue
		}
		notes := make([]domain.Note, len(d.Notes))
		for i, n := range d.Notes {
			notes[i] = domain.Note{ID: n.ID, Body: n.Body, CreatedAt: n.CreatedAt, Done: n.Done}
		}
		result[replacer.Replace(e.Name())] = notes
	}
	return result, nil
}
