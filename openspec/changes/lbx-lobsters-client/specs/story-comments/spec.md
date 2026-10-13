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

### Requirement: Read-mode navigation
In read mode the user SHALL scroll with `j`/`k` and `Ctrl-D`/`Ctrl-U`, move to the start/end with `g`/`G`, and jump between top-level comments with `n`/`N`.

#### Scenario: Jump between top-level comments
- **WHEN** the user presses `n` in read mode
- **THEN** the view moves to the next top-level comment

### Requirement: Expand and collapse
The user SHALL collapse the current thread one level with `h`, expand it with `l`, and toggle all threads with `Enter`.

#### Scenario: Collapse one level
- **WHEN** the user presses `h`
- **THEN** the current comment's replies are hidden

#### Scenario: Toggle all
- **WHEN** the user presses `Enter` in read mode
- **THEN** all threads collapse if any are expanded, otherwise all expand

### Requirement: Navigate mode
Pressing `a` SHALL enter navigate mode, where the user selects individual comments with `j`/`k`, collapses or expands the selected thread with `h`/`l`, and returns to read mode with `Esc`.

#### Scenario: Enter and leave navigate mode
- **WHEN** the user presses `a`
- **THEN** the first comment becomes selected
- **WHEN** the user presses `Esc`
- **THEN** the view returns to read mode

#### Scenario: Collapse selected thread
- **WHEN** the user selects a comment and presses `h`
- **THEN** only that comment's replies are hidden
