package timeago

import (
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		t    time.Time
		want string
	}{
		{"future", now.Add(time.Hour), "now"},
		{"just now", now.Add(-30 * time.Second), "now"},
		{"minutes", now.Add(-5 * time.Minute), "5m"},
		{"hours", now.Add(-3 * time.Hour), "3h"},
		{"days", now.Add(-2 * 24 * time.Hour), "2d"},
		{"weeks", now.Add(-3 * 7 * 24 * time.Hour), "3w"},
		{"months", now.Add(-4 * 30 * 24 * time.Hour), "4mo"},
		{"years", now.Add(-3 * 365 * 24 * time.Hour), "3y"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Format(tc.t, now); got != tc.want {
				t.Errorf("Format = %q, want %q", got, tc.want)
			}
		})
	}
}
