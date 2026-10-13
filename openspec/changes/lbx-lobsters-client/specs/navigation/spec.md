# Spec Delta

## Purpose

Give every view consistent vim-style keybindings and make them discoverable through an in-app help overlay.

## ADDED Requirements

### Requirement: Vim-style scrolling
Every scrollable view SHALL support scrolling with `j`/`k` and `Ctrl-D`/`Ctrl-U`, and jumping to the top or bottom with `g`/`G`.

#### Scenario: Line scroll
- **WHEN** the user presses `j` in any scrollable view
- **THEN** the view scrolls down by a line

#### Scenario: Half-page scroll
- **WHEN** the user presses `Ctrl-D`
- **THEN** the view scrolls down by half a page

### Requirement: Global keymap
The main view SHALL provide `Enter` (view comments), `Space` (reader mode), `Tab` (next category), `/` (search), `r` (refresh), `o` (open story in browser), `c` (open comments in browser), `f` (add favorite), `x` (remove favorite), `u` (toggle read), `z` (toggle wide view), `i` (help), and `q` (quit).

#### Scenario: Toggle read
- **WHEN** the user presses `u` on a story
- **THEN** its read state is toggled

### Requirement: Help overlay
Pressing `i` SHALL show an overlay listing the keybindings for the current view, including comment and reader modes, and `Esc` or `q` SHALL close it.

#### Scenario: Show and close help
- **WHEN** the user presses `i`
- **THEN** the help overlay lists the current view's keybindings
- **WHEN** the user presses `Esc`
- **THEN** the overlay closes and the previous view is restored

### Requirement: Contextual quit and back
`q` SHALL quit the application from the main view and SHALL return to the previous view from a subview.

#### Scenario: Quit from main view
- **WHEN** the user presses `q` in the main view
- **THEN** the application exits

#### Scenario: Back from subview
- **WHEN** the user presses `q` in the comment section or reader mode
- **THEN** the client returns to the previous view
