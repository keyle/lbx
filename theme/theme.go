package theme

import (
	"errors"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Theme holds terminal color values, one per styled element. An empty value
// means "use the terminal's default" for that element.
type Theme struct {
	CategoryOn  string `toml:"category_on"`
	CategoryOff string `toml:"category_off"`
	Title       string `toml:"title"`
	TitleRead   string `toml:"title_read"`
	Meta        string `toml:"meta"`
	Domain      string `toml:"domain"`
	Tag         string `toml:"tag"`
	Score       string `toml:"score"`
	Comments    string `toml:"comments"`
	Submitter   string `toml:"submitter"`
	Selection   string `toml:"selection"`
	Error       string `toml:"error"`
}

// Default returns the built-in palette.
func Default() Theme {
	return Theme{
		CategoryOn: "205",
		Domain:     "244",
		Tag:        "39",
		Score:      "214",
		Comments:   "111",
		Submitter:  "141",
		Error:      "196",
	}
}

// Load reads theme.toml, layering file values over the defaults. A missing
// file yields the defaults, not an error.
func Load(path string) (Theme, error) {
	th := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return th, nil
		}
		return th, err
	}
	if err := toml.Unmarshal(data, &th); err != nil {
		return Default(), fmt.Errorf("parse %s: %w", path, err)
	}
	return th, nil
}

// defaultThemeTOML is written verbatim by default-theme so the file keeps its
// explanatory comments. Empty values fall back to the terminal defaults.
const defaultThemeTOML = `# lbx theme — colors as ANSI 256-color numbers.
# Leave a value empty ("") to use the terminal's default.

category_on  = "205"
category_off = ""
title        = ""
title_read   = ""
meta         = ""
domain       = "244"
tag          = "39"
score        = "214"
comments     = "111"
submitter    = "141"
selection    = ""
error        = "196"
`

// DefaultTOML is the text of the default theme.toml.
func DefaultTOML() string { return defaultThemeTOML }

// WriteDefault writes the default theme to path.
func WriteDefault(path string) error {
	return os.WriteFile(path, []byte(defaultThemeTOML), 0o644)
}
