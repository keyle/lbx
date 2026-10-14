# Spec Delta

## Purpose

Present the story list and the selected content side by side on wide terminals so the user can browse and read at the same time.

## ADDED Requirements

### Requirement: Automatic wide layout
On terminals at least 180 columns wide, the client SHALL open in a dual-pane layout showing the story list beside the selected content.

#### Scenario: Wide terminal
- **WHEN** the terminal is at least 180 columns wide
- **THEN** the client starts in dual-pane mode

#### Scenario: Narrow terminal
- **WHEN** the terminal is narrower than 180 columns
- **THEN** the client starts in single-pane mode

### Requirement: Session toggle
The user SHALL be able to toggle wide view for the current session with `z`.

#### Scenario: Toggle on
- **WHEN** the user presses `z` in single-pane mode on a wide enough terminal
- **THEN** the layout switches to dual-pane
