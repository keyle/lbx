package categories

import "strings"

// Kind identifies a feed source.
type Kind int

const (
	KindHottest Kind = iota
	KindNewest
	KindActive
	KindTag
	KindFavorites
)

// Category is a selectable feed.
type Category struct {
	Kind Kind
	Tag  string
}

// Label is the user-facing name.
func (c Category) Label() string {
	switch c.Kind {
	case KindHottest:
		return "hottest"
	case KindNewest:
		return "newest"
	case KindActive:
		return "active"
	case KindFavorites:
		return "favorites"
	case KindTag:
		return c.Tag
	default:
		return "unknown"
	}
}

// Key is a stable identifier used for cache and selection lookups.
func (c Category) Key() string {
	if c.Kind == KindTag {
		return "t/" + c.Tag
	}
	return c.Label()
}

// Default returns the built-in category order.
func Default() []Category {
	return []Category{{Kind: KindHottest}, {Kind: KindNewest}, {Kind: KindActive}, {Kind: KindFavorites}}
}

// Parse turns names (for example from --categories) into categories. Known
// feed names map to their kinds; any other name becomes a tag feed. Empty
// names are ignored.
func Parse(names []string) []Category {
	cats := make([]Category, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		switch name {
		case "hottest":
			cats = append(cats, Category{Kind: KindHottest})
		case "newest":
			cats = append(cats, Category{Kind: KindNewest})
		case "active":
			cats = append(cats, Category{Kind: KindActive})
		case "favorites":
			cats = append(cats, Category{Kind: KindFavorites})
		default:
			cats = append(cats, Category{Kind: KindTag, Tag: name})
		}
	}
	return cats
}

// Tag returns tag-restricted categories.
func Tag(tags ...string) []Category {
	cats := make([]Category, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		cats = append(cats, Category{Kind: KindTag, Tag: tag})
	}
	return cats
}
