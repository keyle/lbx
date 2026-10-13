package history

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestMarkAndPersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	at := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)

	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.IsRead("a") {
		t.Fatal("empty store should not report read")
	}
	if err := s.Mark("a", 5, at); err != nil {
		t.Fatal(err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reloaded.IsRead("a") {
		t.Fatal("reloaded store should report read")
	}
	e, ok := reloaded.Entry("a")
	if !ok || e.CommentCount != 5 || !e.LastVisit.Equal(at) {
		t.Fatalf("entry = %+v", e)
	}
}

func TestToggle(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "history.json"))
	if err != nil {
		t.Fatal(err)
	}
	at := time.Now()
	read, err := s.Toggle("a", 3, at)
	if err != nil {
		t.Fatal(err)
	}
	if !read || !s.IsRead("a") {
		t.Fatal("toggle should mark read")
	}
	read, err = s.Toggle("a", 3, at)
	if err != nil {
		t.Fatal(err)
	}
	if read || s.IsRead("a") {
		t.Fatal("toggle should clear read")
	}
}

func TestClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Mark("a", 1, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	reloaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.IsRead("a") {
		t.Fatal("clear should remove all history")
	}
}

func TestTrimKeepsNewest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < MaxEntries+5; i++ {
		s.entries[fmt.Sprintf("id%04d", i)] = Entry{LastVisit: base.Add(time.Duration(i) * time.Minute)}
	}
	if err := s.Mark("fresh", 1, base.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	reloaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(reloaded.entries); got != MaxEntries {
		t.Fatalf("entries = %d, want %d", got, MaxEntries)
	}
	if !reloaded.IsRead("fresh") || !reloaded.IsRead("id1004") {
		t.Fatal("newest entries should be kept")
	}
	if reloaded.IsRead("id0000") || reloaded.IsRead("id0005") {
		t.Fatal("oldest entries should be dropped")
	}
}
