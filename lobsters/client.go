package lobsters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// UserAgent identifies the client to lobste.rs. Keep it descriptive so traffic
// is easy to attribute and rate-limit politely.
const UserAgent = "lbx/0.1 (+https://github.com/noben/lbx)"

const (
	defaultTimeout       = 15 * time.Second
	defaultFeedTTL       = 60 * time.Second
	defaultMaxConcurrent = 4
)

// Client is a read-only lobste.rs JSON client. Feed responses are cached
// in-memory with a short TTL and refreshed with conditional requests.
type Client struct {
	http *http.Client
	base string
	ua   string
	ttl  time.Duration
	sem  chan struct{}

	mu    sync.Mutex
	feeds map[string]*feedEntry
}

type feedEntry struct {
	stories []Story
	etag    string
	at      time.Time
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the lobste.rs origin (used in tests).
func WithBaseURL(base string) Option {
	return func(c *Client) { c.base = base }
}

// WithHTTPClient overrides the underlying HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.http = hc }
}

// WithFeedTTL sets how long a cached feed is served without revalidation.
func WithFeedTTL(ttl time.Duration) Option {
	return func(c *Client) { c.ttl = ttl }
}

// New creates a Client with a descriptive User-Agent, request timeout, and a
// bound on concurrent requests.
func New(opts ...Option) *Client {
	c := &Client{
		http:  &http.Client{Timeout: defaultTimeout},
		base:  BaseURL,
		ua:    UserAgent,
		ttl:   defaultFeedTTL,
		sem:   make(chan struct{}, defaultMaxConcurrent),
		feeds: make(map[string]*feedEntry),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Hottest returns the hottest feed.
func (c *Client) Hottest(ctx context.Context) ([]Story, error) { return c.feed(ctx, "hottest") }

// Newest returns the newest feed.
func (c *Client) Newest(ctx context.Context) ([]Story, error) { return c.feed(ctx, "newest") }

// Active returns the active feed.
func (c *Client) Active(ctx context.Context) ([]Story, error) { return c.feed(ctx, "active") }

// Tag returns the feed for a single tag.
func (c *Client) Tag(ctx context.Context, tag string) ([]Story, error) {
	return c.feed(ctx, "t/"+url.PathEscape(tag))
}

func (c *Client) feed(ctx context.Context, key string) ([]Story, error) {
	c.mu.Lock()
	entry := c.feeds[key]
	if entry != nil && time.Since(entry.at) < c.ttl {
		stories := entry.stories
		c.mu.Unlock()
		return stories, nil
	}
	var prevETag string
	if entry != nil {
		prevETag = entry.etag
	}
	c.mu.Unlock()

	var stories []Story
	etag, notModified, err := c.getJSON(ctx, "/"+key+".json", &stories, prevETag)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if notModified {
		if entry == nil {
			return nil, nil
		}
		entry.at = time.Now()
		return entry.stories, nil
	}
	c.feeds[key] = &feedEntry{stories: stories, etag: etag, at: time.Now()}
	return stories, nil
}

// Story returns a single story with its comments.
func (c *Client) Story(ctx context.Context, id string) (*StoryWithComments, error) {
	var story StoryWithComments
	if _, _, err := c.getJSON(ctx, "/s/"+url.PathEscape(id)+".json", &story, ""); err != nil {
		return nil, err
	}
	return &story, nil
}

// getJSON performs a GET, decoding a JSON body into v. It returns the response
// ETag and whether the server answered 304 Not Modified.
func (c *Client) getJSON(ctx context.Context, path string, v any, etag string) (string, bool, error) {
	c.sem <- struct{}{}
	defer func() { <-c.sem }()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/json")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("lobste.rs: %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return etag, true, nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("lobste.rs: %s: %s", path, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return "", false, fmt.Errorf("lobste.rs: decode %s: %w", path, err)
	}
	return resp.Header.Get("ETag"), false, nil
}
