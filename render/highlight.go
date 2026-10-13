package render

import (
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"golang.org/x/net/html"
)

// highlightCode returns syntax-highlighted lines for code in the given
// language, or nil when the language is unknown or formatting fails.
func highlightCode(code, lang string) []string {
	if lang == "" {
		return nil
	}
	lexer := lexers.Get(lang)
	if lexer == nil {
		return nil
	}
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return nil
	}
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil
	}
	var b strings.Builder
	if err := formatter.Format(&b, style, iterator); err != nil {
		return nil
	}
	return strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
}

func codeLanguage(n *html.Node) string {
	if lang := attrLanguage(n); lang != "" {
		return lang
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "code" {
			if lang := attrLanguage(child); lang != "" {
				return lang
			}
		}
	}
	return ""
}

func attrLanguage(n *html.Node) string {
	for _, attr := range n.Attr {
		switch attr.Key {
		case "class":
			if lang := langFromClass(attr.Val); lang != "" {
				return lang
			}
		case "data-lang":
			if attr.Val != "" {
				return attr.Val
			}
		}
	}
	return ""
}

func langFromClass(class string) string {
	for _, field := range strings.Fields(class) {
		if lang, ok := strings.CutPrefix(field, "language-"); ok {
			return lang
		}
		if lang, ok := strings.CutPrefix(field, "lang-"); ok {
			return lang
		}
	}
	return ""
}
