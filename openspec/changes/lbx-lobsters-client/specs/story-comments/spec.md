# Spec Delta

## Purpose

Read and navigate a lobste.rs story's threaded comment section in the terminal, with vim navigation and per-thread expand/collapse.

## ADDED Requirements

### Requirement: Comment thread rendering
The client SHALL display a story's comments as a nested thread, indented by reply depth, and SHALL render each comment's HTML with syntax-aware formatting for code, emphasis, and links.

#### Scenario: Open comments
- **WHEN** the user selects a story and presses `Enter`
- **THEN** its comments are displayed in lobste.rs order, indented by depth

#### Scenario: Comment metadata
- **WHEN** a comment is displayed
- **THEN** it shows the author, relative age, score, and its body

#### Scenario: Code formatting
- **WHEN** a comment body contains a code block
- **THEN** the code is rendered as a distinct monospaced block

### Requirement: Comment navigation
The highlighted comment SHALL move with `j`/`k`, scroll by half a page with `Ctrl-D`/`Ctrl-U`, and jump to the start/end with `g`/`G`.

#### Scenario: Move the highlight
- **WHEN** the user presses `j`
- **THEN** the highlight moves to the next visible comment

### Requirement: Expand and collapse
The user SHALL collapse the highlighted thread with `h`, expand it with `l`, expand the whole sub-thread with `L`, and expand or collapse all threads with `a`/`c`.

#### Scenario: Collapse the highlighted thread
- **WHEN** the user presses `h`
- **THEN** only that comment's replies are hidden

#### Scenario: Toggle all threads
- **WHEN** the user presses `c`
- **THEN** every thread is collapsed
- **WHEN** the user presses `a`
- **THEN** every thread is expanded
