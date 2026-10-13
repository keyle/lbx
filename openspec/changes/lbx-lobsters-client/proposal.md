# Proposal

## Why

Hacker News and lobste.rs are both high-signal link aggregators, but the polished terminal clients — notably `circumflex` (`clx`) — target only Hacker News. lobste.rs exposes a clean, unauthenticated JSON API (feeds, stories with threaded comments), so the same keyboard-driven reading experience can be offered for lobste.rs. This is a greenfield project (`noben.org/lbx`, currently empty), so we build a clean Go/Bubble Tea codebase inspired by circumflex rather than porting it.

## What Changes

- New Go CLI `lbx` built with Bubble Tea / Lip Gloss that browses lobste.rs, matching `clx` keymaps where lobste.rs semantics allow.
- Feed browsing: `hottest`, `newest`, `active`, and per-tag feeds; categories switched with `Tab`; refresh with `r`.
- Comment section: `read` and `navigate` modes, vim scrolling, jump between top-level comments (`n`/`N`), expand/collapse by level (`h`/`l`) and all at once (`Enter`), individual selection (`a`).
- Reader mode: extract the linked article with a readability pass and render it to the terminal; jump between headings (`n`/`N`); in-content search (`/`).
- Link selector: `Tab` to select links, `j`/`k` to move, `Enter` to open in place, `q` to return.
- Wide view: dual-pane list + content on wide terminals (≥180 cols), toggle with `z`.
- Favorites: add/remove (`f`/`x`), `lbx add <id>`, persisted to `favorites.toml`.
- Read history: visited stories marked read, new comments highlighted, `history.json`, `lbx clear-history`, disable with `--no-history`.
- Search: `/` opens the lobste.rs search results page in the user's default browser (the on-site search is behind an Anubis proof-of-work challenge).
- Configuration: `config.toml` and `theme.toml`, plus CLI flags; `lbx default-config` / `lbx default-theme`.
- In-app help overlay listing all keymaps (`i`).
- Read-only: no login, voting, or commenting.

## Capabilities

### New Capabilities

- `story-feeds`: Load and display ranked lobste.rs story lists (hottest, newest, active, per-tag), switch categories, refresh, and show per-story metadata.
- `story-comments`: Read and navigate a story's threaded comment section with vim navigation and expand/collapse.
- `reader-mode`: Fetch and render a linked article's extracted text in the terminal with heading navigation and in-content search.
- `link-selector`: Select and open links found in comments or reader mode without leaving the app.
- `wide-view`: Present the story list and content side by side on wide terminals, with a session toggle.
- `favorites`: Save stories locally, revisit them, and remove them.
- `read-history`: Track visited stories, mark them read, and highlight comments added since the last visit.
- `search`: Launch a lobste.rs search in the user's default browser.
- `navigation`: Vim-style keybindings across views and an in-app help overlay that lists them.
- `configuration`: Customize appearance and behavior through config/theme files and CLI flags.

### Modified Capabilities

- None — greenfield project.

## Impact

- New Go module `noben.org/lbx` with new packages for the lobste.rs client, domain models, views, comments, article rendering, favorites, history, config, and CLI.
- External dependencies: `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `charmbracelet/bubbles`; `go-shiori/go-readability` for article extraction; a terminal renderer for extracted HTML (headings, links, code, lists).
- External systems: lobste.rs public JSON API (`/hottest.json`, `/newest.json`, `/active.json`, `/t/<tag>.json`, `/s/<id>.json`). No authentication.
- Constraint: lobste.rs routes some requests (including `/search`) through an Anubis proof-of-work challenge, so on-site search is delegated to the browser.
- No breaking changes; no existing code or users.
