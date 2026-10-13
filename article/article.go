package article

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	readability "github.com/go-shiori/go-readability"
)

// UserAgent identifies the client when fetching article pages.
const UserAgent = "lbx/0.1 (+https://github.com/noben/lbx)"

// maxBytes caps how much of a page is read before extraction.
const maxBytes = 8 << 20

// ErrNotReadable indicates the target had no extractable article body.
var ErrNotReadable = errors.New("article: no readable content")

// Article is the extracted, cleaned-up content of a page.
type Article struct {
	Title   string
	Byline  string
	Site    string
	Content string
}

// Extract runs readability over an HTML stream. pageURL is used to resolve
// relative links.
func Extract(r io.Reader, pageURL string) (*Article, error) {
	u, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("article: parse url: %w", err)
	}
	a, err := readability.FromReader(r, u)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotReadable, err)
	}
	if strings.TrimSpace(a.Content) == "" {
		return nil, ErrNotReadable
	}
	return &Article{Title: a.Title, Byline: a.Byline, Site: a.SiteName, Content: a.Content}, nil
}

// Fetch downloads and extracts an article. Non-HTML responses are rejected.
func Fetch(ctx context.Context, client *http.Client, rawURL string) (*Article, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("article: %s: %s", rawURL, resp.Status)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" && !strings.Contains(ct, "html") {
		return nil, fmt.Errorf("%w: content-type %q", ErrNotReadable, ct)
	}
	return Extract(io.LimitReader(resp.Body, maxBytes), rawURL)
}
