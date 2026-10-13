package lobsters

import "time"

// BaseURL is the canonical lobste.rs origin.
const BaseURL = "https://lobste.rs"

// Story is a single lobste.rs submission as returned by the feed endpoints.
type Story struct {
	ShortID          string    `json:"short_id"`
	CreatedAt        time.Time `json:"created_at"`
	Title            string    `json:"title"`
	URL              string    `json:"url"`
	Score            int       `json:"score"`
	Flags            int       `json:"flags"`
	CommentCount     int       `json:"comment_count"`
	Description      string    `json:"description"`
	DescriptionPlain string    `json:"description_plain"`
	SubmitterUser    string    `json:"submitter_user"`
	UserIsAuthor     bool      `json:"user_is_author"`
	Tags             []string  `json:"tags"`
	ShortIDURL       string    `json:"short_id_url"`
	CommentsURL      string    `json:"comments_url"`
}

// Comment is a single comment in a story's thread.
type Comment struct {
	ShortID        string    `json:"short_id"`
	Comment        string    `json:"comment"`
	CommentPlain   string    `json:"comment_plain"`
	CommentingUser string    `json:"commenting_user"`
	CreatedAt      time.Time `json:"created_at"`
	LastEditedAt   time.Time `json:"last_edited_at"`
	Depth          int       `json:"depth"`
	Flags          int       `json:"flags"`
	IsDeleted      bool      `json:"is_deleted"`
	IsModerated    bool      `json:"is_moderated"`
	ParentComment  *string   `json:"parent_comment"`
	Score          int       `json:"score"`
	ShortIDURL     string    `json:"short_id_url"`
	URL            string    `json:"url"`
}

// StoryWithComments is the response of the single-story endpoint: the story
// plus its flat comment array.
type StoryWithComments struct {
	Story
	Comments []Comment `json:"comments"`
}
