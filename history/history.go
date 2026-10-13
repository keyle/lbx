package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// MaxEntries is how many stories are remembered. Beyond this, the entries with
// the oldest last visit are dropped.
const MaxEntries = 1000

// Entry records a story's last visit and the comment count seen at that time.
type Entry struct {
	LastVisit    time.Time `json:"last_visit"`
	CommentCount int       `json:"comment_count"`
}

// Store persists visit history to a JSON file.
type Store struct {
	path    string
	entries map[string]Entry
}

// Load reads history.json. A missing file yields an empty store.
func Load(path string) (*Store, error) {
	s := &Store{path: path, entries: map[string]Entry{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, &s.entries); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if s.entries == nil {
		s.entries = map[string]Entry{}
	}
	s.trim()
	return s, nil
}

// trim drops the least recently visited entries beyond MaxEntries.
func (s *Store) trim() {
	excess := len(s.entries) - MaxEntries
	if excess <= 0 {
		return
	}
	ids := make([]string, 0, len(s.entries))
	for id := range s.entries {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := s.entries[ids[i]], s.entries[ids[j]]
		if !a.LastVisit.Equal(b.LastVisit) {
			return a.LastVisit.Before(b.LastVisit)
		}
		return ids[i] < ids[j]
	})
	for _, id := range ids[:excess] {
		delete(s.entries, id)
	}
}

// IsRead reports whether a story has a recorded visit.
func (s *Store) IsRead(id string) bool {
	_, ok := s.entries[id]
	return ok
}

// Entry returns the recorded visit for a story.
func (s *Store) Entry(id string) (Entry, bool) {
	e, ok := s.entries[id]
	return e, ok
}

// Mark records a visit with the observed comment count and persists.
func (s *Store) Mark(id string, commentCount int, at time.Time) error {
	s.entries[id] = Entry{LastVisit: at, CommentCount: commentCount}
	s.trim()
	return s.save()
}

// Toggle flips a story's read state, returning the new state, and persists.
func (s *Store) Toggle(id string, commentCount int, at time.Time) (bool, error) {
	if s.IsRead(id) {
		delete(s.entries, id)
		return false, s.save()
	}
	s.entries[id] = Entry{LastVisit: at, CommentCount: commentCount}
	s.trim()
	return true, s.save()
}

// Clear removes all history and persists.
func (s *Store) Clear() error {
	s.entries = map[string]Entry{}
	return s.save()
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// MarkIfUnread records a visit only when the story has none, so an existing
// entry's last-visit time and comment count stay intact.
func (s *Store) MarkIfUnread(id string, commentCount int, at time.Time) error {
	if s.IsRead(id) {
		return nil
	}
	return s.Mark(id, commentCount, at)
}
