package render

import (
	"strings"

	"golang.org/x/net/html"
)

// Link is an anchor found in HTML content.
type Link struct {
	Text string
	Href string
}

// Links returns the unique anchors in an HTML fragment, skipping fragments and
// empty targets.
func Links(input string) []Link {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return nil
	}
	var links []Link
	seen := map[string]bool{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := ""
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
				}
			}
			text := strings.TrimSpace(nodeText(n))
			if href != "" && !strings.HasPrefix(href, "#") && !seen[href] {
				seen[href] = true
				links = append(links, Link{Text: text, Href: href})
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return links
}
