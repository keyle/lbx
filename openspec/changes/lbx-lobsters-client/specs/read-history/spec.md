# Spec Delta

## Purpose

Track which stories the user has visited and highlight comments added since their last visit.

## ADDED Requirements

### Requirement: Mark stories read
Opening a story's comment section SHALL mark that story as read in the list.

#### Scenario: Visited story
- **WHEN** the user opens a story's comments and returns to the list
- **THEN** that story is shown as read

### Requirement: Highlight new comments
When a previously visited story is opened again, comments added since the last visit SHALL be visually highlighted.

#### Scenario: New comments highlighted
- **WHEN** the user reopens a story whose comment count has grown since the last visit
- **THEN** the new comments are highlighted

### Requirement: History persistence
Visit data SHALL be stored in `history.json` under the cache directory, `$XDG_CACHE_HOME/lbx` with a `~/.cache/lbx` default, recording each story's last-visit time and observed comment count.

#### Scenario: Persist across runs
- **WHEN** the user visits a story and restarts the client
- **THEN** the story is still marked read

### Requirement: Disable and clear history
The user SHALL be able to disable history with `--no-history` and clear it with `lbx clear-history`.

#### Scenario: Disabled
- **WHEN** the user runs `lbx --no-history`
- **THEN** no read state is recorded or applied

#### Scenario: Clear
- **WHEN** the user runs `lbx clear-history`
- **THEN** all stored history is removed
