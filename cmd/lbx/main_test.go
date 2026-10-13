package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"noben.org/lbx/favorites"
	"noben.org/lbx/lobsters"
	"noben.org/lbx/settings"
)

func TestAddFavorite(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"short_id":"abc","title":"T","url":"https://e.com","score":1,"comment_count":0}`)
	}))
	defer srv.Close()

	paths := settings.Paths{Config: t.TempDir(), Cache: t.TempDir()}
	client := lobsters.New(lobsters.WithBaseURL(srv.URL))
	if err := addFavorite(client, []string{"abc"}, paths); err != nil {
		t.Fatal(err)
	}

	store, err := favorites.Load(paths.FavoritesFile())
	if err != nil {
		t.Fatal(err)
	}
	if !store.Has("abc") {
		t.Fatal("story was not added to favorites")
	}
}
