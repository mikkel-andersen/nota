package store

import (
	"os"
	"path/filepath"
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
	if _, err := s.Add("first"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add("second"); err != nil {
		t.Fatal(err)
	}

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
	if _, err := s.Add("keep"); err != nil {
		t.Fatal(err)
	}
	n, err := s.Add("delete me")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add("also keep"); err != nil {
		t.Fatal(err)
	}

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
	// deleting a non-existent ID is a no-op by design
	s := newTestStore(t)
	if _, err := s.Add("note"); err != nil {
		t.Fatal(err)
	}

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
	if _, err := s.Add("one"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add("two"); err != nil {
		t.Fatal(err)
	}

	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}

	notes, _ := s.List()
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes after clear, got %d", len(notes))
	}

	// IDs must not restart after clear — next note continues from the last assigned ID
	n, err := s.Add("after clear")
	if err != nil {
		t.Fatal(err)
	}
	if n.ID != 3 {
		t.Errorf("expected ID 3 after clear, got %d", n.ID)
	}
}

func Test_ids_do_not_reset_after_delete(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Add("first"); err != nil {
		t.Fatal(err)
	}
	n2, err := s.Add("second")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(n2.ID); err != nil {
		t.Fatal(err)
	}
	n3, err := s.Add("third")
	if err != nil {
		t.Fatal(err)
	}

	if n3.ID != 3 {
		t.Errorf("expected ID 3, got %d", n3.ID)
	}
}

func Test_persistence_across_instances(t *testing.T) {
	dir := t.TempDir()
	s1, _ := New(dir)
	if _, err := s1.Add("persisted"); err != nil {
		t.Fatal(err)
	}

	s2, _ := New(dir)
	notes, err := s2.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0].Body != "persisted" {
		t.Errorf("note not persisted across store instances")
	}
}

func Test_next_id_persisted_across_instances(t *testing.T) {
	dir := t.TempDir()
	s1, _ := New(dir)
	n1, err := s1.Add("first")
	if err != nil {
		t.Fatal(err)
	}
	if err := s1.Delete(n1.ID); err != nil {
		t.Fatal(err)
	}

	s2, _ := New(dir)
	n2, err := s2.Add("second")
	if err != nil {
		t.Fatal(err)
	}
	if n2.ID != 2 {
		t.Errorf("expected ID 2 after reload+delete cycle, got %d", n2.ID)
	}
}

func Test_migration_from_legacy_json(t *testing.T) {
	// Simulate a notes.json written before next_id was tracked (NextID missing/zero)
	dir := t.TempDir()
	legacy := `{"notes":[{"id":3,"body":"old note","created_at":"2024-01-01T00:00:00Z"},{"id":5,"body":"another","created_at":"2024-01-01T00:00:00Z"}]}`
	if err := os.WriteFile(filepath.Join(dir, "notes.json"), []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}

	s, _ := New(dir)
	n, err := s.Add("new note")
	if err != nil {
		t.Fatal(err)
	}
	if n.ID != 6 {
		t.Errorf("expected ID 6 after migrating legacy file with max ID 5, got %d", n.ID)
	}
}
