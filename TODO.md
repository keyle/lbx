# lbx — TODO

`lbx` is a terminal client for [lobste.rs](https://lobste.rs), built in Go with
[Bubble Tea](https://github.com/charmbracelet/bubbletea), inspired by
[circumflex](https://github.com/bensadeh/circumflex) (`clx`). Read-only; no
login, voting, or commenting.

This file tracks the project at a high level. Detailed, tracked work for the
first milestone lives in the OpenSpec change
[`openspec/changes/lbx-lobsters-client/`](openspec/changes/lbx-lobsters-client/)
(proposal → specs → design → tasks).

## Milestone 0 — Foundations

- [x] Confirm binary name `lbx` and module `noben.org/lbx`
- [ ] `git init` and initial commit
- [ ] Scaffold package layout: `cmd/lbx`, `lobsters`, `domain`, `view`, `render`, `style`, `theme`, `settings`, `timeago`, `history`, `favorites`, `help`, `categories`, `header`, `highlight`
- [ ] Add dependencies: bubbletea, lipgloss, bubbles, go-readability, chroma, go-toml/v2, pkg/browser, x/net/html
- [ ] Add Makefile (`build`, `test`, `lint`)
- [ ] Add LICENSE and README skeleton

## Milestone 1 — Data layer

- [ ] lobste.rs JSON client (typed `Story`/`Comment`, feeds, single story)
- [ ] Comment tree built from the flat comment array
- [ ] Fixture-based decode tests

## Milestone 2 — Feed browsing

- [ ] Root model + view stack
- [ ] Story list rendering with metadata and vim list navigation
- [ ] Categories (`hottest`, `newest`, `active`, `favorites`, per-tag) and refresh

## Milestone 3 — Comments

- [ ] Shared HTML→ANSI renderer
- [ ] Threaded comment rendering
- [x] Always-highlighted comment (`j`/`k`), `l`/`h` expand/collapse, `L` expand sub-thread, `a`/`c` expand/collapse all
- [x] Header: `name [+] N pts age`
- [x] Blank line between stories; `?`/`i` help

## Milestone 4 — Reader mode

- [ ] Article extraction via go-readability
- [ ] Rendered reading with heading jumps and in-content search
- [ ] Best-effort image display on Kitty-graphics terminals

## Milestone 5 — Navigation and layout

- [ ] Link selector with in-place open and back
- [ ] Wide (dual-pane) view with `z` toggle and `--wide-view`
- [ ] Centralized keymaps and `i` help overlay

## Milestone 6 — Persistence and config

- [ ] Path resolution: config/state in `$XDG_CONFIG_HOME/lbx` (default `~/.config/lbx`), cache in `$XDG_CACHE_HOME/lbx` (default `~/.cache/lbx`)
- [ ] `favorites.toml` (`f`/`x`, `lbx add <id>`) under the config dir
- [ ] `history.json` read-marking, new-comment highlighting, `--no-history`, `lbx clear-history` under the cache dir
- [ ] `config.toml` / `theme.toml`, `default-config`, `default-theme`, flags

## Milestone 7 — Polish and release

- [ ] End-to-end smoke pass against live lobste.rs (watch for Anubis challenges)
- [ ] README: install, features, keymaps, flags
- [ ] Cross-platform build check (darwin, linux, windows)
- [ ] Archiving the OpenSpec change and updating `openspec/specs/`

## Notes / constraints

- Search (`/`) opens the lobste.rs search page in the browser; the on-site
  search route is behind an Anubis proof-of-work challenge we do not attempt to solve.
- Use only lobste.rs JSON endpoints and stay a polite, cached client (descriptive
  User-Agent, timeouts, rate limiting).
