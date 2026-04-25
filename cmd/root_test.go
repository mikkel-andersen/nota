package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mikkel-andersen/nota/domain"
)

func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func Test_render_plain(t *testing.T) {
	notes := []domain.Note{
		{ID: 1, Body: "hello", CreatedAt: time.Now().Add(-2 * time.Minute)},
		{ID: 2, Body: "world", CreatedAt: time.Now().Add(-25 * time.Hour)},
	}
	out := captureStdout(func() { renderPlain(notes) })
	if !strings.Contains(out, "#1  hello") {
		t.Errorf("expected #1 in plain output, got: %s", out)
	}
	if !strings.Contains(out, "#2  world") {
		t.Errorf("expected #2 in plain output, got: %s", out)
	}
	if strings.Contains(out, "\033[") {
		t.Error("plain output should not contain ANSI codes")
	}
}

func Test_render_json(t *testing.T) {
	notes := []domain.Note{
		{ID: 1, Body: "test note", CreatedAt: time.Now()},
	}
	out := captureStdout(func() { renderJSON(notes) })
	if !strings.Contains(out, `"body": "test note"`) {
		t.Errorf("expected body in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("expected id in JSON output, got: %s", out)
	}
}

func Test_format_age(t *testing.T) {
	cases := []struct {
		age  time.Duration
		want string
	}{
		{0, "just now"},
		{5 * time.Second, "just now"},
		{59 * time.Second, "just now"},
		{60 * time.Second, "1m ago"},
		{2 * time.Minute, "2m ago"},
		{time.Hour, "1h ago"},
		{90 * time.Minute, "1h ago"},
		{25 * time.Hour, "1d ago"},
		{48 * time.Hour, "2d ago"},
	}
	for _, c := range cases {
		got := formatAge(time.Now().Add(-c.age))
		if got != c.want {
			t.Errorf("formatAge(-%v) = %q, want %q", c.age, got, c.want)
		}
	}
}
