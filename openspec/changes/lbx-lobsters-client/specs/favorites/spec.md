# Spec Delta

## Purpose

Save interesting stories locally so the user can find and revisit them later.

## ADDED Requirements

### Requirement: Toggle favorites
The user SHALL toggle the highlighted story in favorites with `f`, adding it when absent and removing it when present, and SHALL remove it with `x`.

#### Scenario: Add
- **WHEN** the user presses `f` on a story that is not a favorite
- **THEN** the story is added to favorites

#### Scenario: Remove by toggling
- **WHEN** the user presses `f` on a story that is already a favorite
- **THEN** the story is removed from favorites

#### Scenario: Remove
- **WHEN** the user presses `x` on a story that is already a favorite
- **THEN** the story is removed from favorites

### Requirement: Favorites category
Favorites SHALL appear as a category in the category rotation, listing saved stories with the same metadata as other feeds.

#### Scenario: View favorites
- **WHEN** the user switches to the favorites category
- **THEN** only saved stories are listed

### Requirement: Persistence
Favorites SHALL be stored in a human-readable `favorites.toml` under the config directory, `$XDG_CONFIG_HOME/lbx` with a `~/.config/lbx` default, and loaded on startup.

#### Scenario: Persist across runs
- **WHEN** the user adds a favorite and restarts the client
- **THEN** the favorite is still present

#### Scenario: Add by ID
- **WHEN** the user runs `lbx add <id>`
- **THEN** the story with that short id is added to favorites without opening the UI

### Requirement: No duplicates
Adding a story that is already a favorite SHALL not create a duplicate entry.

#### Scenario: Re-add
- **WHEN** the user adds an already-favorited story
- **THEN** the favorites list is unchanged
