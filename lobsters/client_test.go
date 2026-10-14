package lobsters

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"noben.org/lbx/version"
)

func TestFeedRequests(t *testing.T) {
	var lastUA, lastPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastUA = r.Header.Get("User-Agent")
		lastPath = r.URL.Path
		w.Header().Set("ETag", `"v1"`)
		io.WriteString(w, `[{"short_id":"x","title":"t","created_at":"2026-10-11T13:09:49.352-05:00"}]`)
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL), WithFeedTTL(0))
	stories, err := c.Hottest(context.Background())
	if err != nil {
		t.Fatalf("Hottest: %v", err)
	}
	if len(stories) != 1 || stories[0].ShortID != "x" {
		t.Fatalf("unexpected stories: %+v", stories)
	}
	if lastPath != "/hottest.json" {
		t.Errorf("path = %q, want /hottest.json", lastPath)
	}
	if lastUA != version.UserAgent() {
		t.Errorf("User-Agent = %q, want %q", lastUA, version.UserAgent())
	}

	if _, err := c.Tag(context.Background(), "go"); err != nil {
		t.Fatalf("Tag: %v", err)
	}
	if lastPath != "/t/go.json" {
		t.Errorf("tag path = %q, want /t/go.json", lastPath)
	}
}

func TestFeedRevalidation(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Header.Get("If-None-Match") == `"v1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		io.WriteString(w, `[{"short_id":"x","title":"t","created_at":"2026-10-11T13:09:49.352-05:00"}]`)
	}))
	defer srv.Close()

	cached := New(WithBaseURL(srv.URL), WithFeedTTL(time.Minute))
	for i := 0; i < 3; i++ {
		if _, err := cached.Hottest(context.Background()); err != nil {
			t.Fatalf("cached Hottest: %v", err)
		}
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("cache hits = %d requests, want 1", got)
	}

	revalidating := New(WithBaseURL(srv.URL), WithFeedTTL(0))
	if _, err := revalidating.Hottest(context.Background()); err != nil {
		t.Fatalf("revalidate Hottest: %v", err)
	}
	stories, err := revalidating.Hottest(context.Background())
	if err != nil {
		t.Fatalf("revalidate Hottest (304): %v", err)
	}
	if len(stories) != 1 || stories[0].ShortID != "x" {
		t.Errorf("304 returned %+v, want cached story x", stories)
	}
}

func TestStory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/s/abc.json" {
			t.Errorf("path = %q, want /s/abc.json", r.URL.Path)
		}
		io.WriteString(w, `{"short_id":"abc","title":"t","comments":[{"short_id":"c1","depth":0}]}`)
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL))
	story, err := c.Story(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Story: %v", err)
	}
	if story.ShortID != "abc" {
		t.Errorf("ShortID = %q, want abc", story.ShortID)
	}
	if len(story.Comments) != 1 || story.Comments[0].ShortID != "c1" {
		t.Errorf("comments = %+v, want c1", story.Comments)
	}
}

func TestErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL))
	if _, err := c.Hottest(context.Background()); err == nil {
		t.Fatal("Hottest: want error on 403, got nil")
	}
}
