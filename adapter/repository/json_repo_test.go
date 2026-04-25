package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestRepo(t *testing.T) *JSONRepo {
	t.Helper()
	r, err := NewJSONRepo(t.TempDir())
	if err != nil {
		t.Fatalf("NewJSONRepo: %v", err)
	}
	return r
}

func Test_list_empty(t *testing.T) {
	r := newTestRepo(t)
	notes, err := r.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes, got %d", len(notes))
	}
}

func Test_add(t *testing.T) {
	r := newTestRepo(t)
	n, err := r.Add("hello world")
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
	r := newTestRepo(t)
	for i := 1; i <= 3; i++ {
		n, err := r.Add("note")
		if err != nil {
			t.Fatal(err)
		}
		if n.ID != i {
			t.Errorf("expected ID %d, got %d", i, n.ID)
		}
	}
}

func Test_list_returns_persisted(t *testing.T) {
	r := newTestRepo(t)
	if _, err := r.Add("first"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Add("second"); err != nil {
		t.Fatal(err)
	}
	notes, err := r.List()
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
	r := newTestRepo(t)
	if _, err := r.Add("keep"); err != nil {
		t.Fatal(err)
	}
	n, err := r.Add("delete me")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Add("also keep"); err != nil {
		t.Fatal(err)
	}
	if err := r.Delete(n.ID); err != nil {
		t.Fatal(err)
	}
	notes, _ := r.List()
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
	r := newTestRepo(t)
	if _, err := r.Add("note"); err != nil {
		t.Fatal(err)
	}
	if err := r.Delete(999); err != nil {
		t.Fatal(err)
	}
	notes, _ := r.List()
	if len(notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(notes))
	}
}

func Test_clear(t *testing.T) {
	r := newTestRepo(t)
	if _, err := r.Add("one"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Add("two"); err != nil {
		t.Fatal(err)
	}
	if err := r.Clear(); err != nil {
		t.Fatal(err)
	}
	notes, _ := r.List()
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes after clear, got %d", len(notes))
	}
	n, err := r.Add("after clear")
	if err != nil {
		t.Fatal(err)
	}
	if n.ID != 3 {
		t.Errorf("expected ID 3 after clear, got %d", n.ID)
	}
}

func Test_ids_do_not_reset_after_delete(t *testing.T) {
	r := newTestRepo(t)
	if _, err := r.Add("first"); err != nil {
		t.Fatal(err)
	}
	n2, err := r.Add("second")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Delete(n2.ID); err != nil {
		t.Fatal(err)
	}
	n3, err := r.Add("third")
	if err != nil {
		t.Fatal(err)
	}
	if n3.ID != 3 {
		t.Errorf("expected ID 3, got %d", n3.ID)
	}
}

func Test_persistence_across_instances(t *testing.T) {
	dir := t.TempDir()
	r1, _ := NewJSONRepo(dir)
	if _, err := r1.Add("persisted"); err != nil {
		t.Fatal(err)
	}
	r2, _ := NewJSONRepo(dir)
	notes, err := r2.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0].Body != "persisted" {
		t.Errorf("note not persisted across repo instances")
	}
}

func Test_next_id_persisted_across_instances(t *testing.T) {
	dir := t.TempDir()
	r1, _ := NewJSONRepo(dir)
	n1, err := r1.Add("first")
	if err != nil {
		t.Fatal(err)
	}
	if err := r1.Delete(n1.ID); err != nil {
		t.Fatal(err)
	}
	r2, _ := NewJSONRepo(dir)
	n2, err := r2.Add("second")
	if err != nil {
		t.Fatal(err)
	}
	if n2.ID != 2 {
		t.Errorf("expected ID 2 after reload+delete cycle, got %d", n2.ID)
	}
}

func Test_mark_done(t *testing.T) {
	r := newTestRepo(t)
	n, err := r.Add("finish the feature")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.MarkDone(n.ID); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
	notes, _ := r.List()
	if !notes[0].Done {
		t.Error("expected note to be marked done")
	}
}

func Test_mark_done_unknown_id(t *testing.T) {
	r := newTestRepo(t)
	if _, err := r.Add("a note"); err != nil {
		t.Fatal(err)
	}
	if err := r.MarkDone(999); err == nil {
		t.Fatal("expected error for unknown ID, got nil")
	}
}

func Test_update(t *testing.T) {
	r := newTestRepo(t)
	n, err := r.Add("original")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Update(n.ID, "updated"); err != nil {
		t.Fatal(err)
	}
	notes, _ := r.List()
	if notes[0].Body != "updated" {
		t.Errorf("expected %q, got %q", "updated", notes[0].Body)
	}
}

func Test_update_not_found(t *testing.T) {
	r := newTestRepo(t)
	if err := r.Update(999, "body"); err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
}

func Test_update_persists(t *testing.T) {
	dir := t.TempDir()
	r1, _ := NewJSONRepo(dir)
	n, _ := r1.Add("before")
	r1.Update(n.ID, "after")
	r2, _ := NewJSONRepo(dir)
	notes, _ := r2.List()
	if notes[0].Body != "after" {
		t.Errorf("update not persisted: got %q", notes[0].Body)
	}
}

func Test_migration_from_legacy_json(t *testing.T) {
	dir := t.TempDir()
	legacy := `{"notes":[{"id":3,"body":"old note","created_at":"2024-01-01T00:00:00Z"},{"id":5,"body":"another","created_at":"2024-01-01T00:00:00Z"}]}`
	if err := os.WriteFile(filepath.Join(dir, "notes.json"), []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}
	r, _ := NewJSONRepo(dir)
	n, err := r.Add("new note")
	if err != nil {
		t.Fatal(err)
	}
	if n.ID != 6 {
		t.Errorf("expected ID 6 after migrating legacy file with max ID 5, got %d", n.ID)
	}
}
