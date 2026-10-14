package favorites

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pelletier/go-toml/v2"

	"noben.org/lbx/lobsters"
)

// Favorite is a saved story in a compact, human-editable form.
type Favorite struct {
	ShortID       string    `toml:"short_id"`
	Title         string    `toml:"title"`
	URL           string    `toml:"url"`
	Score         int       `toml:"score"`
	CommentCount  int       `toml:"comment_count"`
	SubmitterUser string    `toml:"submitter_user,omitempty"`
	Tags          []string  `toml:"tags,omitempty"`
	CreatedAt     time.Time `toml:"created_at"`
}

// FromStory converts a feed story into a stored favorite.
func FromStory(s lobsters.Story) Favorite {
	return Favorite{
		ShortID:       s.ShortID,
		Title:         s.Title,
		URL:           s.URL,
		Score:         s.Score,
		CommentCount:  s.CommentCount,
		SubmitterUser: s.SubmitterUser,
		Tags:          s.Tags,
		CreatedAt:     s.CreatedAt,
	}
}

// Story converts a favorite back into a story for rendering.
func (f Favorite) Story() lobsters.Story {
	return lobsters.Story{
		ShortID:       f.ShortID,
		Title:         f.Title,
		URL:           f.URL,
		Score:         f.Score,
		CommentCount:  f.CommentCount,
		SubmitterUser: f.SubmitterUser,
		Tags:          f.Tags,
		CreatedAt:     f.CreatedAt,
		ShortIDURL:    lobsters.BaseURL + "/s/" + f.ShortID,
		CommentsURL:   lobsters.BaseURL + "/s/" + f.ShortID,
	}
}

type file struct {
	Stories []Favorite `toml:"stories"`
}

// Store persists favorites to a TOML file. It is safe for concurrent use.
type Store struct {
	mu    sync.Mutex
	path  string
	items []Favorite
}

// Load reads favorites.toml. A missing file yields an empty store.
func Load(path string) (*Store, error) {
	s := &Store{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil
		}
		return nil, err
	}
	var f file
	if err := toml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	s.items = f.Stories
	return s, nil
}

// List returns a copy of the stored favorites in insertion order.
func (s *Store) List() []Favorite {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Favorite(nil), s.items...)
}

// Has reports whether a short id is already saved.
func (s *Store) Has(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.has(id)
}

func (s *Store) has(id string) bool {
	for _, it := range s.items {
		if it.ShortID == id {
			return true
		}
	}
	return false
}

// Add saves a favorite and persists, ignoring duplicates by short id.
func (s *Store) Add(f Favorite) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.has(f.ShortID) {
		return nil
	}
	s.items = append(s.items, f)
	return s.save()
}

// Remove deletes a favorite by short id and persists.
func (s *Store) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Favorite, 0, len(s.items))
	for _, it := range s.items {
		if it.ShortID != id {
			out = append(out, it)
		}
	}
	s.items = out
	return s.save()
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(file{Stories: s.items})
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
