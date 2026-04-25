package store

import (
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func Test_list_empty(t *testing.T) {
	s := newTestStore(t)
	notes, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes, got %d", len(notes))
	}
}

func Test_add(t *testing.T) {
	s := newTestStore(t)
	n, err := s.Add("hello world")
	if err != nil {
		t.Fatal(err)
	}
	if n.ID != 1 {
		t.Errorf("expected ID 1, got %d", n.ID)
	}
	if n.Body != "hello world" {
		t.Errorf("unexpected body: %q", n.Body)
	}
	if time.Since(n.CreatedAt) > 5*time.Second {
		t.Error("CreatedAt is too old")
	}
}

func Test_add_increments_id(t *testing.T) {
	s := newTestStore(t)
	for i := 1; i <= 3; i++ {
		n, err := s.Add("note")
		if err != nil {
			t.Fatal(err)
		}
		if n.ID != i {
			t.Errorf("expected ID %d, got %d", i, n.ID)
		}
	}
}

func Test_list_returns_persisted(t *testing.T) {
	s := newTestStore(t)
	s.Add("first")
	s.Add("second")

	notes, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
	if notes[0].Body != "first" || notes[1].Body != "second" {
		t.Errorf("unexpected bodies: %q, %q", notes[0].Body, notes[1].Body)
	}
}

func Test_delete(t *testing.T) {
	s := newTestStore(t)
	s.Add("keep")
	n, _ := s.Add("delete me")
	s.Add("also keep")

	if err := s.Delete(n.ID); err != nil {
		t.Fatal(err)
	}

	notes, _ := s.List()
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes after delete, got %d", len(notes))
	}
	for _, note := range notes {
		if note.ID == n.ID {
			t.Error("deleted note still present")
		}
	}
}

func Test_delete_non_existent(t *testing.T) {
	s := newTestStore(t)
	s.Add("note")

	if err := s.Delete(999); err != nil {
		t.Fatal(err)
	}
	notes, _ := s.List()
	if len(notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(notes))
	}
}

func Test_clear(t *testing.T) {
	s := newTestStore(t)
	s.Add("one")
	s.Add("two")

	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}

	notes, _ := s.List()
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes after clear, got %d", len(notes))
	}
}

func Test_ids_do_not_reset_after_delete(t *testing.T) {
	s := newTestStore(t)
	s.Add("first")
	n2, _ := s.Add("second")
	s.Delete(n2.ID)
	n3, _ := s.Add("third")

	if n3.ID != 3 {
		t.Errorf("expected ID 3, got %d", n3.ID)
	}
}

func Test_persistence_across_instances(t *testing.T) {
	dir := t.TempDir()
	s1, _ := New(dir)
	s1.Add("persisted")

	s2, _ := New(dir)
	notes, err := s2.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0].Body != "persisted" {
		t.Errorf("note not persisted across store instances")
	}
}
