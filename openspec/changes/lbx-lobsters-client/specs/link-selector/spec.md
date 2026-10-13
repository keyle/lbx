# Spec Delta

## Purpose

Let the user select and open links found in comments or reader mode without leaving the client.

## ADDED Requirements

### Requirement: Link selection
In the comment section and in reader mode, `Tab` SHALL enter link selection, `j`/`k` SHALL move between the links in the current view, and `Enter` SHALL open the selected link.

#### Scenario: Select and move
- **WHEN** the user presses `Tab` in a view that contains links
- **THEN** the first link is highlighted and `j`/`k` move the highlight

### Requirement: Open in place
Opening a link SHALL load its target within the client: an article SHALL open in reader mode, and a link to another lobste.rs discussion SHALL load that discussion's comments.

#### Scenario: Open linked discussion
- **WHEN** the user opens a link to another lobste.rs story
- **THEN** that story's comment section is displayed in place

#### Scenario: Open linked article
- **WHEN** the user opens a link to an article
- **THEN** the article opens in reader mode

### Requirement: Return to previous view
Pressing `q` after following a link SHALL return the user to where they were.

#### Scenario: Go back
- **WHEN** the user presses `q` after following a link
- **THEN** the previous view and its position are restored

### Requirement: External links
A selected link that is neither a readable article nor a lobste.rs discussion SHALL open in the system browser.

#### Scenario: Link opens externally
- **WHEN** the user opens an external link that cannot be shown in the client
- **THEN** it is opened in the system browser
