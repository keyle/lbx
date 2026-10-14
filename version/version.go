// Package version holds the build version of lbx. The Makefile overrides
// Version at link time via -ldflags "-X noben.org/lbx/version.Version=...".
package version

// Version is the build version; "dev" for a plain checkout.
var Version = "dev"

// repository is the project's canonical source location.
const repository = "https://github.com/noben/lbx"

// UserAgent is the HTTP User-Agent sent on every outbound request.
func UserAgent() string {
	return "lbx/" + Version + " (+" + repository + ")"
}
