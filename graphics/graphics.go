package graphics

import (
	"os"
	"strings"
)

// KittySupported reports whether the current terminal likely supports the Kitty
// graphics protocol, which reader mode uses to attempt inline image display.
func KittySupported() bool {
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}
	switch os.Getenv("TERM_PROGRAM") {
	case "kitty", "WezTerm", "ghostty", "Ghostty":
		return true
	}
	return strings.Contains(os.Getenv("TERM"), "kitty")
}

// Placeholder is the textual stand-in shown for an image on terminals without
// graphics support.
func Placeholder(alt string) string {
	if strings.TrimSpace(alt) == "" {
		alt = "image"
	}
	return "[image: " + alt + "]"
}
