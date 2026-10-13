# Spec Delta

## Purpose

Let the user customize the client's appearance and behavior through config and theme files and command-line flags.

## ADDED Requirements

### Requirement: Config file
The client SHALL read settings from `config.toml` in the config directory, with command-line flags overriding file values.

#### Scenario: Write defaults
- **WHEN** the user runs `lbx default-config`
- **THEN** a `config.toml` containing the default settings is written to the config directory

#### Scenario: Flag overrides file
- **WHEN** a setting appears in both `config.toml` and a flag
- **THEN** the flag value is used

### Requirement: Directory resolution
The config directory SHALL resolve from `$XDG_CONFIG_HOME` (default `~/.config`) + `/lbx`, and the cache directory from `$XDG_CACHE_HOME` (default `~/.cache`) + `/lbx`, on every platform including macOS.

#### Scenario: XDG override
- **WHEN** `XDG_CONFIG_HOME` is set
- **THEN** config, theme, and favorites files are read and written under `$XDG_CONFIG_HOME/lbx`

#### Scenario: Default location
- **WHEN** `XDG_CONFIG_HOME` and `XDG_CACHE_HOME` are unset
- **THEN** state is written to `~/.config/lbx` and cache to `~/.cache/lbx`

### Requirement: Theme file
The client SHALL read colors from `theme.toml` and default to the terminal's native color scheme when no theme is configured.

#### Scenario: Write default theme
- **WHEN** the user runs `lbx default-theme`
- **THEN** a `theme.toml` containing the default colors is written to the config directory

### Requirement: Category and feed flags
The client SHALL support `--categories` to select and order the categories shown, and `--tag` to add a tag-restricted feed.

#### Scenario: Configure categories
- **WHEN** the user runs `lbx --categories hottest,newest,favorites`
- **THEN** those categories appear in that order

### Requirement: Behavior flags
The client SHALL support `--wide-view` to control the dual-pane layout and `--no-history` to disable visit tracking.

#### Scenario: Disable history
- **WHEN** the user runs `lbx --no-history`
- **THEN** no read state is recorded or shown
