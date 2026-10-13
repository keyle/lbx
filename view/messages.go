package view

import (
	"noben.org/lbx/article"
	"noben.org/lbx/lobsters"
)

// feedLoadedMsg carries the result of loading a feed.
type feedLoadedMsg struct {
	key     string
	stories []lobsters.Story
	err     error
}

// storyLoadedMsg carries the result of loading a single story with comments.
type storyLoadedMsg struct {
	id    string
	story *lobsters.StoryWithComments
	err   error
}

// articleLoadedMsg carries the result of fetching and extracting an article.
type articleLoadedMsg struct {
	art *article.Article
	err error
}

// openCommentsMsg asks the root model to push the comment section for a story.
type openCommentsMsg struct {
	story lobsters.Story
}

// openReaderMsg asks the root model to push reader mode for a story.
type openReaderMsg struct {
	story lobsters.Story
}

// categoryChangedMsg tells the root model which category is now selected.
type categoryChangedMsg struct {
	index int
}

// popMsg asks the root model to pop the active screen.
type popMsg struct{}

// quitMsg asks the root model to exit.
type quitMsg struct{}
