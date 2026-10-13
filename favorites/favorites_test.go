package favorites

import (
	"path/filepath"
	"testing"
	"time"

	"noben.org/lbx/lobsters"
)

func TestAddRemovePersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.toml")
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Add(Favorite{ShortID: "a", Title: "A"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Add(Favorite{ShortID: "b", Title: "B"}); err != nil {
		t.Fatal(err)
	}
	if !s.Has("a") || !s.Has("b") {
		t.Fatal("expected both favorites")
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.List(); len(got) != 2 || got[0].ShortID != "a" || got[1].ShortID != "b" {
		t.Fatalf("reloaded = %+v", got)
	}

	if err := reloaded.Remove("a"); err != nil {
		t.Fatal(err)
	}
	again, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if again.Has("a") || !again.Has("b") {
		t.Fatalf("after remove = %+v", again.List())
	}
}

func TestAddDuplicate(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "favorites.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Add(Favorite{ShortID: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Add(Favorite{ShortID: "a"}); err != nil {
		t.Fatal(err)
	}
	if len(s.List()) != 1 {
		t.Fatalf("duplicate add produced %d entries", len(s.List()))
	}
}

func TestFromStoryRoundTrip(t *testing.T) {
	created := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	src := lobsters.Story{ShortID: "x1", Title: "T", URL: "https://e.com", Score: 3, CommentCount: 7, SubmitterUser: "u", Tags: []string{"go"}, CreatedAt: created}
	got := FromStory(src).Story()
	if got.ShortID != src.ShortID || got.Title != src.Title || got.Score != src.Score || got.CommentCount != src.CommentCount {
		t.Fatalf("round trip = %+v", got)
	}
	if got.CommentsURL != lobsters.BaseURL+"/s/x1" {
		t.Errorf("CommentsURL = %q", got.CommentsURL)
	}
}
