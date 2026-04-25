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
		{5 * time.Second, "just now"},
		{59 * time.Second, "just now"},
		{2 * time.Minute, "2m ago"},
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
