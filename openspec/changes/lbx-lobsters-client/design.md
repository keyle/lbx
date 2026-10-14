# Design

## Context

See `proposal.md` — Why. This is a greenfield Go module (`noben.org/lbx`, currently only `go.mod`). The reference implementation, `circumflex` (`clx`), is a Bubble Tea app whose package layout maps cleanly onto the same problem: `hn` (client), `domain` (models), `view` (navigation), `comment`, `article`, `categories`, `header`, `headline`, `highlight`, `history`, `favorites`, `settings`, `style`, `theme`, `layout`, `frame`, `timeago`, `nerdfonts`, `ansi`, `graphics`. We mirror those boundaries, replacing `hn` with a `lobsters` client.

lobste.rs exposes unauthenticated JSON (verified during discovery):

- Feeds: `/hottest.json`, `/newest.json`, `/active.json`, `/t/<tag>.json` — arrays of story objects with `short_id`, `title`, `url`, `score`, `comment_count`, `flags`, `tags`, `submitter_user`, `created_at`, `short_id_url`, `comments_url`, `description_plain`.
- Story: `/s/<short_id>.json` — the story object plus a flat `comments` array. Each comment has `short_id`, `parent_comment` (null for top level), `depth`, `commenting_user`, `score`, `created_at`, `last_edited_at`, `comment` (HTML), `comment_plain`, `is_deleted`, `is_moderated`.

Constraint discovered: lobste.rs fronts routes (including `/search`) with an **Anubis** proof-of-work challenge, so search is delegated to the browser and the client must be a well-behaved, cached consumer of the JSON endpoints only.

## Goals / Non-Goals

**Goals:**
- A maintainable, keyboard-driven lobste.rs TUI matching `clx` behavior and keymaps where lobste.rs semantics allow.
- One shared HTML→terminal renderer used for both comments and reader-mode articles.
- Deterministic, testable domain layer (JSON decoding, comment tree, history/favorites) independent of the TUI.

**Non-Goals:**
- Authentication, voting, posting, or any write path.
- Scraping lobste.rs HTML or solving the Anubis challenge.
- Windows-specific escape-sequence tuning beyond cross-platform build support.
- Perfect article extraction on every site.

## Decisions

### Package layout mirrors circumflex
Keep the proven separation: `cmd/lbx` (entrypoint/flags), `lobsters` (HTTP client + JSON decoding), `domain` (story/comment models and the comment tree), `view` (screen stack + main model), `article` (readable extraction), `categories`, `header`, `history`, `favorites`, `help`, `keymap`, `style`, `theme`, `settings`, `timeago`, `render` (HTML→ANSI), `version`.
*Alternatives:* a flatter single-package app (simpler to start, but the views/rendering/history concerns quickly tangle) or a framework-heavy MVU split. Reusing clx's boundaries keeps files small and mirrors a battle-tested shape.

### JSON client over typed structs
`lobsters` wraps `net/http` with a descriptive `User-Agent`, request timeout, conditional requests, and typed `Story`/`Comment` structs decoded with `encoding/json`. Feeds are cached in-memory with a short TTL; concurrent fetches are bounded.
*Alternatives:* a generic `map[string]any` decoder (loses type safety, harder to test) or a third-party lobste.rs library (none maintained). Typed structs + table-driven decode tests are the most testable.

### Shared HTML→ANSI renderer
Comments and article bodies are both HTML, so one `render` package parses with `golang.org/x/net/html` and emits styled, wrapped ANSI for paragraphs, headings, emphasis, links, lists, quotes, and code (fenced code highlighted with `alexflint/go-arg`-free `alecthomas/chroma`). Reader mode feeds `go-readability`'s cleaned HTML into the same renderer.
*Alternatives:* `html-to-markdown` → `glamour` (reuses a markdown renderer but adds fidelity losses and two conversion layers), or plain `html-to-text` (loses headings/code/link affordances the spec requires). A single renderer gives control over native colors and code detection and is straightforward to unit-test with golden files.

### Single Bubble Tea program with a view stack
One `tea.Model` (`view`) owns a stack of subviews (list, comments, reader) so `q`/`Esc` pop back and wide view can render the top two. `update` dispatches to the active view; `tea.Cmd`s perform IO (fetch, extract, open browser) and return typed messages.
*Alternatives:* nested `tea.Program`s (harder to share header/keymaps) or per-screen programs (loses back-navigation state). A stack matches clx and the spec's "return to previous view".

### Comment tree built from the flat array
lobste.rs returns comments flat with `parent_comment`/`depth`; `domain` builds an ordered tree once per load so rendering, per-level collapse, and `n`/`N` top-level jumps are pure functions over the tree.
*Alternatives:* rebuild relationships during render (repeated work, harder to test) or trust `depth` alone (breaks on malformed threads). Building once from `parent_comment` is explicit and testable.

### Config and state in XDG dirs
State lives under the config dir, resolved from `$XDG_CONFIG_HOME` (default `~/.config`) + `/lbx`: `config.toml`, `theme.toml`, and `favorites.toml`. Cache lives under the cache dir, resolved from `$XDG_CACHE_HOME` (default `~/.cache`) + `/lbx`: `history.json` (and any on-disk feed cache). Resolution is XDG-style on every platform so macOS uses `~/.config/lbx` and `~/.cache/lbx` too, not `~/Library/...`. Files are created on demand with `MkdirAll(0o755)`. Parsed with `github.com/pelletier/go-toml/v2`. Flags override file values.
*Alternatives:* `os.UserConfigDir()`/`os.UserCacheDir()` (idiomatic but diverge on macOS); a single config file for theme too (harder to hand-edit) or a state DB (heavier than the tiny key/value data warrants).

### Search opens the browser
`/` prompts, then opens `https://lobste.rs/search?q=<term>&what=stories` via `github.com/pkg/browser`. No HTTP request is made to search routes.
*Alternatives:* scraping `/search` (blocked by Anubis and fragile) or shipping no search (fails the parity goal).

### Images as text placeholders
Reader mode renders every image as an `[img (alt text)]` placeholder rather than displaying it inline, keeping the client dependency-free.
*Alternatives:* the Kitty graphics protocol (deferred) or a native image library (a heavy dependency for a read-only text client).

## Risks / Trade-offs

- **Anubis proof-of-work may begin challenging JSON endpoints** → keep to documented JSON routes, send a descriptive User-Agent, cache and rate-limit, and surface a clear error offering to open the URL in the browser.
- **Article extraction is inherently lossy** (paywalls, JS-heavy sites) → detect thin/failed extraction and offer to open in the browser; never crash the view.
- **Large threads / very long articles can lag the TUI** → build trees once, precompute rendered lines, and only re-render on width/data change.
- **HTML→ANSI fidelity and terminal differences** (truecolor vs 256 vs none) → drive colors from the theme with graceful fallback to terminal-native colors.
- **Terminal resize and wide-view edge cases** → treat width as a function of state and recompute layout on `tea.WindowSizeMsg`.

## Migration Plan

Greenfield: no data or users to migrate. Build order: scaffold module and interfaces, land the client + domain + tests, then the TUI views, then persistence, then reader mode/assets. Each milestone is independently runnable. Rollback is `git revert`; nothing external is mutated.

## Open Questions

- Exact set of lobste.rs tags to surface as optional `--tag` feeds, and whether to add a dedicated tag picker view in a later change.
- Whether to support Nerd Font icons on by default or behind a flag.
