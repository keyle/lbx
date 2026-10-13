package categories

import "testing"

func TestLabelAndKey(t *testing.T) {
	tag := Category{Kind: KindTag, Tag: "go"}
	if tag.Label() != "go" {
		t.Errorf("Label = %q, want go", tag.Label())
	}
	if tag.Key() != "t/go" {
		t.Errorf("Key = %q, want t/go", tag.Key())
	}
	if got := (Category{Kind: KindHottest}).Key(); got != "hottest" {
		t.Errorf("hottest Key = %q", got)
	}
}

func TestParse(t *testing.T) {
	cats := Parse([]string{"hottest", "newest", "favorites", "go", " rust "})
	want := []string{"hottest", "newest", "favorites", "go", "rust"}
	if len(cats) != len(want) {
		t.Fatalf("got %d categories, want %d", len(cats), len(want))
	}
	for i, w := range want {
		if cats[i].Label() != w {
			t.Errorf("category %d = %q, want %q", i, cats[i].Label(), w)
		}
	}
	if cats[3].Kind != KindTag {
		t.Errorf("go should be a tag category")
	}
}

func TestParseSkipsEmpty(t *testing.T) {
	if cats := Parse([]string{"", "  ", "active"}); len(cats) != 1 || cats[0].Label() != "active" {
		t.Errorf("Parse empty entries = %+v, want just active", cats)
	}
}

func TestDefault(t *testing.T) {
	cats := Default()
	if len(cats) != 4 || cats[0].Label() != "hottest" || cats[3].Label() != "favorites" {
		t.Errorf("Default() = %+v", cats)
	}
}
