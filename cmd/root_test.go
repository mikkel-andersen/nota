package cmd

import (
	"testing"
	"time"
)

func Test_format_age(t *testing.T) {
	cases := []struct {
		age  time.Duration
		want string
	}{
		{0, "just now"},
		{5 * time.Second, "just now"},
		{59 * time.Second, "just now"},
		{60 * time.Second, "1m ago"},  // boundary: exactly 1 minute
		{2 * time.Minute, "2m ago"},
		{time.Hour, "1h ago"},          // boundary: exactly 1 hour
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
