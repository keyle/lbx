package render

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"noben.org/lbx/lobsters"
)

func TestParagraphs(t *testing.T) {
	got := HTML("<p>First</p><p>Second</p>", 40)
	if got != "First\n\nSecond" {
		t.Errorf("got %q", got)
	}
}

func TestWrapping(t *testing.T) {
	got := HTML("<p>alpha beta gamma delta epsilon zeta</p>", 12)
	for _, line := range strings.Split(got, "\n") {
		if w := lipgloss.Width(line); w > 12 {
			t.Errorf("line %q width %d > 12", line, w)
		}
	}
	if !strings.Contains(got, "epsilon") {
		t.Errorf("words lost in wrapping: %q", got)
	}
}

func TestCodeBlock(t *testing.T) {
	got := HTML("<pre><code>package main\n\nfunc main() {}</code></pre>", 40)
	if !strings.Contains(got, "package main") || !strings.Contains(got, "func main() {}") {
		t.Errorf("code block not preserved:\n%s", got)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 || strings.TrimSpace(lines[1]) != "" {
		t.Errorf("blank line inside code block lost:\n%s", got)
	}
}

func TestCodeBlockHighlighted(t *testing.T) {
	got := HTML(`<pre><code class="language-go">package main

func main() {}</code></pre>`, 60)
	if !strings.Contains(got, "package") || !strings.Contains(got, "main") {
		t.Errorf("highlighted code lost content:\n%s", got)
	}
}

func TestInlineCodeAndEmphasis(t *testing.T) {
	got := HTML("<p>run <code>go test</code> and be <em>happy</em></p>", 60)
	if !strings.Contains(got, "go test") || !strings.Contains(got, "happy") {
		t.Errorf("inline content lost: %q", got)
	}
}

func TestLink(t *testing.T) {
	got := HTML(`<p>see <a href="https://example.com/x">the docs</a></p>`, 60)
	if !strings.Contains(got, "the docs") {
		t.Errorf("link text lost: %q", got)
	}
	if !strings.Contains(got, "https://example.com/x") {
		t.Errorf("link target lost: %q", got)
	}
}

func TestLinkWithSameTextNoDuplicate(t *testing.T) {
	got := HTML(`<p><a href="https://example.com">https://example.com</a></p>`, 60)
	if strings.Count(got, "https://example.com") != 1 {
		t.Errorf("expected the URL once, got: %q", got)
	}
}

func TestList(t *testing.T) {
	got := HTML("<ul><li>one</li><li>two</li></ul>", 40)
	if !strings.Contains(got, "one") || !strings.Contains(got, "two") {
		t.Errorf("list items lost:\n%s", got)
	}
	if !strings.Contains(got, "•") {
		t.Errorf("list bullet missing:\n%s", got)
	}
}

func TestBlockquote(t *testing.T) {
	got := HTML("<blockquote><p>quoted text</p></blockquote>", 40)
	if !strings.Contains(got, ">") || !strings.Contains(got, "quoted text") {
		t.Errorf("blockquote missing marker:\n%s", got)
	}
}

func TestHeading(t *testing.T) {
	got := HTML("<h2>Section</h2><p>body</p>", 40)
	if !strings.Contains(got, "Section") || !strings.Contains(got, "body") {
		t.Errorf("heading/body lost:\n%s", got)
	}
}

func TestStripsScriptAndStyle(t *testing.T) {
	got := HTML("<p>keep</p><script>evil()</script><style>.x{}</style>", 40)
	if strings.Contains(got, "evil") || strings.Contains(got, ".x") {
		t.Errorf("script/style leaked:\n%s", got)
	}
}

func TestNoRawTags(t *testing.T) {
	got := HTML("<p>hello <b>world</b></p>", 40)
	if strings.Contains(got, "<") || strings.Contains(got, ">") {
		t.Errorf("raw tags leaked: %q", got)
	}
}

func TestRenderFixtureComment(t *testing.T) {
	data, err := os.ReadFile("../lobsters/testdata/story.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var story lobsters.StoryWithComments
	if err := json.Unmarshal(data, &story); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	got := HTML(story.Comments[0].Comment, 80)
	if got == "" {
		t.Fatal("rendered empty comment")
	}
	if strings.Contains(got, "<p>") || strings.Contains(got, "</p>") {
		t.Errorf("comment retained paragraph tags:\n%s", got)
	}
	for _, line := range strings.Split(got, "\n") {
		if w := lipgloss.Width(line); w > 80 {
			t.Errorf("comment line width %d > 80: %q", w, line)
		}
	}
}

func TestLinesHeadingMetadata(t *testing.T) {
	lines := New(40).Lines("<h2>One</h2><p>body</p><h3>Two</h3>")
	var headings []string
	for _, l := range lines {
		if l.Heading {
			headings = append(headings, l.Text)
		}
	}
	if len(headings) != 2 || !strings.Contains(headings[0], "One") || !strings.Contains(headings[1], "Two") {
		t.Errorf("headings = %v", headings)
	}
}

func TestImagePlaceholder(t *testing.T) {
	got := HTML(`<img src="x.png" alt="a cat">`, 40)
	if !strings.Contains(got, "a cat") {
		t.Errorf("image alt text lost: %q", got)
	}
}

func TestLinks(t *testing.T) {
	got := Links(`<p><a href="https://a.example/x">one</a> <a href="https://a.example/x">dup</a> <a href="/s/abc">story</a></p>`)
	if len(got) != 2 {
		t.Fatalf("Links = %+v, want 2 unique", got)
	}
	if got[0].Text != "one" || got[0].Href != "https://a.example/x" {
		t.Errorf("first link = %+v", got[0])
	}
	if got[1].Href != "/s/abc" {
		t.Errorf("second link = %+v", got[1])
	}
}

func TestLinksSkipsFragments(t *testing.T) {
	if got := Links(`<a href="#top">top</a>`); len(got) != 0 {
		t.Errorf("fragment link kept: %+v", got)
	}
}

func TestEmptyInput(t *testing.T) {
	if got := HTML("", 40); got != "" {
		t.Errorf("empty input = %q, want empty", got)
	}
}
