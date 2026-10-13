# Tasks

## 1. Project scaffold

- [ ] 1.1 Create the package layout (`cmd/lbx`, `lobsters`, `domain`, `view`, `style`, `theme`, `settings`, `timeago`, `render`, `history`, `favorites`, `help`, `categories`, `header`, `highlight`) with placeholder files, and verify `go build ./...` succeeds
- [ ] 1.2 Add dependencies (`bubbletea`, `lipgloss`, `bubbles`, `go-readability`, `chroma`, `go-toml/v2`, `pkg/browser`, `golang.org/x/net/html`) and verify `go mod tidy` and `go build ./...` succeed
- [ ] 1.3 Add a Makefile (or task list) with `build`, `test`, and `lint` targets and verify `make build` produces the `lbx` binary

## 2. lobste.rs client and domain

- [ ] 2.1 Define `Story` and `Comment` structs matching the lobste.rs JSON, with table-driven decode tests over captured fixtures (`hottest.json`, `s/<id>.json`) and verify `go test ./lobsters/...` passes
- [ ] 2.2 Implement the HTTP client with descriptive User-Agent, timeout, and methods for hottest/newest/active/tag feeds and single-story fetch, and verify with `httptest` server tests
- [ ] 2.3 Build the comment tree from the flat `comments` array (by `parent_comment`/`depth`) and verify ordering, nesting, and malformed-input unit tests pass

## 3. Feed browsing

- [ ] 3.1 Implement the root Bubble Tea model and view stack (list, comments, reader) and verify `go run ./cmd/lbx` renders the default feed
- [ ] 3.2 Implement story row rendering (title, domain, tags, score, comment count, submitter, age) and list navigation with `j`/`k`, `Ctrl-D`/`Ctrl-U`, `g`/`G`, verified by golden render tests and a manual run
- [ ] 3.3 Implement category rotation (`Tab`) and refresh (`r`) preserving selection, and verify state-transition unit tests pass

## 4. Comment section

- [ ] 4.1 Implement the shared HTML→ANSI renderer (paragraphs, emphasis, links, lists, quotes, code via chroma) and verify golden-file tests pass
- [ ] 4.2 Render the comment tree with depth, author, age, score, and hidden-reply markers, verified by golden tests
- [ ] 4.3 Implement read/navigate modes with `h`/`l` per-level collapse, `Enter` toggle-all, `a` selection, and `n`/`N` top-level jumps, verified by state-transition tests

## 5. Reader mode

- [ ] 5.1 Implement article fetch and `go-readability` extraction over local HTML fixtures, and verify extraction tests pass
- [ ] 5.2 Render extracted content through the shared renderer with `j`/`k` scrolling, `n`/`N` heading jumps, and `/` in-content search, verified by tests plus a manual run
- [ ] 5.3 Add best-effort image display behind a Kitty-graphics capability check with alt-text fallback, verified manually on a supported and an unsupported terminal

## 6. Link selector and wide view

- [ ] 6.1 Implement link extraction and selection (`Tab`, `j`/`k`, `Enter`) with in-place open (article → reader, lobste.rs discussion → comments) and `q` to return, verified by tests on link parsing and manual runs
- [ ] 6.2 Implement dual-pane wide view (auto at ≥180 columns, `z` toggle, `--wide-view never`), verified by layout-selection tests across widths

## 7. Persistence

- [ ] 7.1 Implement `favorites.toml` load/save with add, remove, and de-duplication, verified by round-trip tests
- [ ] 7.2 Implement `lbx add <id>` to add a favorite without opening the UI, verified by a test using an `httptest` server
- [ ] 7.3 Implement `history.json` read-marking, new-comment detection, `u` toggle, `--no-history`, and `lbx clear-history`, verified by round-trip tests

## 8. Navigation and help

- [ ] 8.1 Centralize keymaps for all views and implement contextual `q`/`Esc` (quit vs. back), verified by key-dispatch tests
- [ ] 8.2 Implement the `i` help overlay listing per-view keybindings, verified by a golden test and a manual run

## 9. Configuration and CLI

- [ ] 9.1 Implement XDG path resolution (`$XDG_CONFIG_HOME/lbx`, `$XDG_CACHE_HOME/lbx`), `config.toml` and `theme.toml` loading, the `default-config`/`default-theme` writers, and flag-over-file precedence, verified by tests
- [ ] 9.2 Implement CLI flags (`--categories`, `--tag`, `--wide-view`, `--no-history`) and subcommands (`add`, `clear-history`, `default-config`, `default-theme`), verified by parsing tests and `lbx help`

## 10. Integration and polish

- [ ] 10.1 Run an end-to-end smoke pass against the live lobste.rs JSON API (feeds, comments, reader, favorites, history) and confirm Anubis does not block the JSON routes; record findings
- [ ] 10.2 Write the README (install, features, keymaps, flags) and verify the documented build and run commands work as written
- [ ] 10.3 Verify `GOOS=darwin GOOS=linux GOOS=windows go build ./cmd/lbx` succeeds for all three targets

## Workflow follow-up

- Archive the change with `openspec archive lbx-lobsters-client` once implementation is complete and reviewed.
- Verify the archived specs under `openspec/specs/` match the shipped behavior.
