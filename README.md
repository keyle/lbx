# lbx

`lbx` is a read-only terminal client for [lobste.rs](https://lobste.rs), built in
Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea), inspired by
[circumflex](https://github.com/bensadeh/circumflex) (`clx`).

There is no login, voting, or commenting — browsing and reading only.

## Features

- Browse the `hottest`, `newest`, `active`, and per-tag story feeds; rotate
  categories with `Tab` and refresh with `r`.
- Read threaded comments with vim navigation, per-level collapse/expand, and
  jumps between top-level comments.
- Reader mode extracts the linked article and renders it in the terminal, with
  heading jumps and in-article search.
- Link selector opens links in place — lobste.rs discussions as comments,
  everything else as reader content.
- Wide (dual-pane) view on terminals 180 columns or wider.
- Favorites (`f`/`x`, `lbx add <id>`) persisted to `favorites.toml`.
- Read history marks visited stories and highlights new comments; disable with
  `--no-history`, clear with `lbx clear-history`.
- `/` opens a lobste.rs search in your browser.
- In-app help overlay (`i`).

## Install

```sh
go install noben.org/lbx/cmd/lbx@latest
```

Or build from a checkout:

```sh
make build      # builds ./lbx
./lbx
```

Run without building:

```sh
go run ./cmd/lbx
```

## Keys

### Story list

| Key       | Action                      |
| --------- | --------------------------- |
| `j` / `k` | next / previous story       |
| `Ctrl-D` / `Ctrl-U` | half-page down / up |
| `g` / `G` | top / bottom                |
| `Enter`   | open comments               |
| `Space`   | open reader mode            |
| `o` / `c` | open story / comments in browser |
| `Tab`     | next category               |
| `r`       | refresh                     |
| `/`       | search (opens the browser)  |
| `f` / `x` | add / remove favorite       |
| `u`       | toggle read                 |
| `z`       | toggle wide view            |
| `i`       | help                        |
| `q`       | quit                        |

### Comments

| Key       | Action                          |
| --------- | ------------------------------- |
| `j` / `k` | scroll down / up                |
| `Ctrl-D` / `Ctrl-U` | half-page down / up    |
| `g` / `G` | top / bottom                    |
| `n` / `N` | next / previous top-level comment |
| `h` / `l` | collapse / expand thread        |
| `Enter`   | toggle all threads              |
| `a`       | navigate comments               |
| `Tab`     | select link                     |
| `Esc`     | leave navigate mode / back      |
| `q`       | back                            |

### Reader

| Key       | Action                          |
| --------- | ------------------------------- |
| `j` / `k` | scroll down / up                |
| `Ctrl-D` / `Ctrl-U` | half-page down / up    |
| `g` / `G` | top / bottom                    |
| `n` / `N` | next / previous heading or match |
| `/`       | search in article               |
| `Tab`     | select link                     |
| `Enter`   | show images                     |
| `Esc`     | clear search / back             |
| `q`       | back                            |

## Flags

| Flag                    | Meaning                                        |
| ----------------------- | ---------------------------------------------- |
| `--categories <list>`   | comma-separated categories to show, in order   |
| `--tag <name>`          | add a tag feed (repeatable)                    |
| `--wide-view <mode>`    | `auto` (default), `never`, or `always`         |
| `--no-history`          | disable read-history tracking                  |

Flags override values in `config.toml`.

## Commands

| Command               | Action                              |
| --------------------- | ----------------------------------- |
| `lbx add <short-id>`  | add a story to favorites            |
| `lbx clear-history`   | remove all stored read history      |
| `lbx default-config`  | write a default `config.toml`       |
| `lbx default-theme`   | write a default `theme.toml`        |
| `lbx help`            | show usage                          |

## Configuration

Config and theme live under `$XDG_CONFIG_HOME/lbx` (default `~/.config/lbx`);
history and cache under `$XDG_CACHE_HOME/lbx` (default `~/.cache/lbx`). The XDG
variables take precedence on every platform, including macOS.

- `config.toml` — categories, tag feeds, wide-view mode, history toggle.
- `theme.toml` — ANSI 256-color numbers per element; empty values fall back to
  the terminal's default.
- `favorites.toml` — saved stories.
- `history.json` — per-story last-visit time and observed comment count.

## Notes and limitations

- **Search uses the browser.** lobste.rs fronts its `/search` route with an
  [Anubis](https://github.com/TecharoHQ/anubis) proof-of-work challenge, so `lbx`
  opens the search page in your default browser instead of fetching it.
- **Images are best-effort.** Reader mode detects Kitty-graphics terminals and
  renders alt-text placeholders elsewhere; true inline image placement is not
  implemented yet.
- **Read-only client.** It uses only lobste.rs's unauthenticated JSON endpoints
  (`/hottest.json`, `/newest.json`, `/active.json`, `/t/<tag>.json`,
  `/s/<id>.json`) with a descriptive User-Agent, timeouts, and a short feed
  cache.

## Development

```sh
make test    # go test ./...
make lint    # go vet + gofmt check
make build   # build the binary
```

## License

MIT. See [`LICENSE`](LICENSE).
