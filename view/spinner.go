package view

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"

	"noben.org/lbx/style"
)

const (
	cometWidth = 8
	cometFPS   = 70 * time.Millisecond
)

// cometTrail is the glyph drawn at each distance behind the comet's head.
var cometTrail = []string{"█", "▓", "▒", "░"}

// cometFrames builds the frames of a comet that sweeps back and forth across a
// short track, leaving a fading trail behind it.
func cometFrames() []string {
	var frames []string
	frame := func(head, dir int) string {
		var b strings.Builder
		for i := 0; i < cometWidth; i++ {
			dist := (head - i) * dir
			if dist >= 0 && dist < len(cometTrail) {
				b.WriteString(cometTrail[dist])
			} else {
				b.WriteString("·")
			}
		}
		return b.String()
	}
	for h := 0; h < cometWidth; h++ {
		frames = append(frames, frame(h, 1))
	}
	for h := cometWidth - 2; h >= 1; h-- {
		frames = append(frames, frame(h, -1))
	}
	return frames
}

// newSpinner returns the spinner shared by every loading screen.
func newSpinner() spinner.Model {
	return spinner.New(spinner.WithSpinner(spinner.Spinner{Frames: cometFrames(), FPS: cometFPS}))
}

// loadingView renders the animated spinner followed by a muted label.
func loadingView(spin spinner.Model, label string) string {
	return style.CategoryOn.Render(spin.View()) + " " + style.Status.Render(label)
}
