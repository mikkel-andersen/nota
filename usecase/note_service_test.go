package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/mikkel-andersen/nota/domain"
)

// mockRepo implements Repository for testing use cases in isolation.
type mockRepo struct {
	notes  []domain.Note
	nextID int
	err    error // if set, all methods return this error
}

func (m *mockRepo) Add(body string) (domain.Note, error) {
	if m.err != nil {
		return domain.Note{}, m.err
	}
	m.nextID++
	n := domain.Note{ID: m.nextID, Body: body, CreatedAt: time.Now()}
	m.notes = append(m.notes, n)
	return n, nil
}

func (m *mockRepo) List() ([]domain.Note, error) {
	return m.notes, m.err
}

func (m *mockRepo) Delete(id int) error {
	if m.err != nil {
		return m.err
	}
	notes := m.notes[:0]
	for _, n := range m.notes {
		if n.ID != id {
			notes = append(notes, n)
		}
	}
	m.notes = notes
	return nil
}

func (m *mockRepo) Update(id int, body string) error {
	if m.err != nil {
		return m.err
	}
	for i, n := range m.notes {
		if n.ID == id {
			m.notes[i].Body = body
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockRepo) MarkDone(id int) error {
	if m.err != nil {
		return m.err
	}
	for i, n := range m.notes {
		if n.ID == id {
			m.notes[i].Done = true
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockRepo) Clear() error {
	m.notes = nil
	return m.err
}

// mockEditor records calls and returns configured responses.
type mockEditor struct {
	returnContent string
	returnErr     error
}

func (e *mockEditor) Open(content string) (string, error) {
	return e.returnContent, e.returnErr
}

// mockScanner returns a fixed map.
type mockScanner struct {
	result map[string][]domain.Note
	err    error
}

func (s *mockScanner) AllNotes() (map[string][]domain.Note, error) {
	return s.result, s.err
}

func newService(repo Repository) *NoteService {
	return NewNoteService(repo, &mockScanner{}, &mockEditor{})
}

func Test_service_add(t *testing.T) {
	svc := newService(&mockRepo{})
	n, err := svc.Add("hello")
	if err != nil {
		t.Fatal(err)
	}
	if n.Body != "hello" {
		t.Errorf("expected body %q, got %q", "hello", n.Body)
	}
}

func Test_service_list(t *testing.T) {
	repo := &mockRepo{}
	repo.Add("one")
	repo.Add("two")
	svc := newService(repo)
	notes, err := svc.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
}

func Test_service_delete(t *testing.T) {
	repo := &mockRepo{}
	n, _ := repo.Add("remove me")
	svc := newService(repo)
	if err := svc.Delete(n.ID); err != nil {
		t.Fatal(err)
	}
	notes, _ := svc.List()
	if len(notes) != 0 {
		t.Errorf("expected 0 notes after delete, got %d", len(notes))
	}
}

func Test_service_mark_done(t *testing.T) {
	repo := &mockRepo{}
	n, _ := repo.Add("task")
	svc := newService(repo)
	if err := svc.MarkDone(n.ID); err != nil {
		t.Fatal(err)
	}
	notes, _ := svc.List()
	if !notes[0].Done {
		t.Error("expected note to be done")
	}
}

func Test_service_edit_updates_body(t *testing.T) {
	repo := &mockRepo{}
	n, _ := repo.Add("original")
	svc := NewNoteService(repo, &mockScanner{}, &mockEditor{returnContent: "updated"})
	if err := svc.Edit(n.ID); err != nil {
		t.Fatal(err)
	}
	notes, _ := svc.List()
	if notes[0].Body != "updated" {
		t.Errorf("expected %q, got %q", "updated", notes[0].Body)
	}
}

func Test_service_edit_no_changes(t *testing.T) {
	repo := &mockRepo{}
	n, _ := repo.Add("original")
	svc := NewNoteService(repo, &mockScanner{}, &mockEditor{returnContent: "original"})
	err := svc.Edit(n.ID)
	if !errors.Is(err, ErrNoChanges) {
		t.Errorf("expected ErrNoChanges, got %v", err)
	}
}

func Test_service_edit_not_found(t *testing.T) {
	svc := newService(&mockRepo{})
	err := svc.Edit(999)
	if !errors.Is(err, domain.ErrNoteNotFound) {
		t.Errorf("expected domain.ErrNoteNotFound, got %v", err)
	}
}

func Test_service_all_notes(t *testing.T) {
	scanner := &mockScanner{result: map[string][]domain.Note{
		"/home/user/proj": {{ID: 1, Body: "note"}},
	}}
	svc := NewNoteService(&mockRepo{}, scanner, &mockEditor{})
	groups, err := svc.AllNotes()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
}
