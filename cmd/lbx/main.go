// Command lbx is a read-only terminal client for lobste.rs.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"noben.org/lbx/categories"
	"noben.org/lbx/favorites"
	"noben.org/lbx/history"
	"noben.org/lbx/lobsters"
	"noben.org/lbx/settings"
	"noben.org/lbx/style"
	"noben.org/lbx/theme"
	"noben.org/lbx/view"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "lbx:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "add":
			return runAdd(args[1:])
		case "clear-history":
			return runClearHistory()
		case "default-config":
			return runDefaultConfig()
		case "default-theme":
			return runDefaultTheme()
		case "help":
			usage()
			return nil
		default:
			return fmt.Errorf("unknown command %q", args[0])
		}
	}
	return runUI(args)
}

func runUI(args []string) error {
	fs := flag.NewFlagSet("lbx", flag.ContinueOnError)
	fs.Usage = usage
	catsFlag := fs.String("categories", "", "comma-separated categories")
	wideFlag := fs.String("wide-view", "", "wide layout: auto, never, or always")
	noHistFlag := fs.Bool("no-history", false, "disable read-history tracking")
	var tags tagList
	fs.Var(&tags, "tag", "add a tag feed (repeatable)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	paths := settings.Resolve()
	cfg, err := settings.Load(paths.ConfigFile())
	if err != nil {
		return err
	}
	th, err := theme.Load(paths.ThemeFile())
	if err != nil {
		return err
	}
	style.Apply(th)

	ov := settings.Overrides{}
	if isSet(fs, "categories") {
		ov.Categories = strings.Split(*catsFlag, ",")
	}
	if isSet(fs, "tag") {
		ov.Tags = tags
	}
	if isSet(fs, "wide-view") {
		ov.WideView = wideFlag
	}
	if isSet(fs, "no-history") {
		ov.NoHistory = noHistFlag
	}
	cfg = ov.Apply(cfg)

	cats := categories.Parse(cfg.Categories)
	if len(cfg.Tags) > 0 {
		cats = append(cats, categories.Tag(cfg.Tags...)...)
	}

	fav, err := favorites.Load(paths.FavoritesFile())
	if err != nil {
		return err
	}
	var hist *history.Store
	if !cfg.NoHistory {
		hist, err = history.Load(paths.HistoryFile())
		if err != nil {
			return err
		}
	}

	model := view.New(view.Options{
		Client:     lobsters.New(),
		Categories: cats,
		WideView:   cfg.WideView,
		Favorites:  fav,
		History:    hist,
		NoHistory:  cfg.NoHistory,
	})
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		return err
	}
	return nil
}

func runAdd(args []string) error {
	return addFavorite(lobsters.New(), args, settings.Resolve())
}

func addFavorite(client *lobsters.Client, args []string, paths settings.Paths) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return errors.New("add: missing story id")
	}
	id := strings.TrimSpace(args[0])
	store, err := favorites.Load(paths.FavoritesFile())
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	story, err := client.Story(ctx, id)
	if err != nil {
		return err
	}
	if err := store.Add(favorites.FromStory(story.Story)); err != nil {
		return err
	}
	fmt.Println("added", id)
	return nil
}

func runClearHistory() error {
	paths := settings.Resolve()
	store, err := history.Load(paths.HistoryFile())
	if err != nil {
		return err
	}
	if err := store.Clear(); err != nil {
		return err
	}
	fmt.Println("cleared", paths.HistoryFile())
	return nil
}

func runDefaultConfig() error {
	path, err := settings.Resolve().WriteDefaultConfig()
	if err != nil {
		return err
	}
	fmt.Println("wrote", path)
	return nil
}

func runDefaultTheme() error {
	paths := settings.Resolve()
	if err := paths.EnsureConfigDir(); err != nil {
		return err
	}
	if err := theme.WriteDefault(paths.ThemeFile()); err != nil {
		return err
	}
	fmt.Println("wrote", paths.ThemeFile())
	return nil
}

func isSet(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

type tagList []string

func (t *tagList) String() string { return strings.Join(*t, ",") }

func (t *tagList) Set(v string) error {
	if v = strings.TrimSpace(v); v != "" {
		*t = append(*t, v)
	}
	return nil
}

func usage() {
	fmt.Fprint(os.Stderr, `lbx — read-only terminal client for lobste.rs

usage:
  lbx [flags]            browse stories
  lbx add <short-id>     add a story to favorites
  lbx clear-history      remove all read history
  lbx default-config     write a default config.toml
  lbx default-theme      write a default theme.toml

flags:
  --categories <list>    comma-separated categories
  --tag <name>           add a tag feed (repeatable)
  --wide-view <mode>     auto, never, or always
  --no-history           disable read-history tracking
`)
}
