# Spec Delta

## Purpose

Provide the user with ranked lobste.rs story lists they can browse, filter by tag, switch between, and refresh, without leaving the terminal.

## ADDED Requirements

### Requirement: Ranked feed browsing
The client SHALL fetch ranked story feeds from the lobste.rs JSON API and display them as a scrollable list, allowing the user to switch categories in place.

#### Scenario: Initial load
- **WHEN** the client starts and history/favorites are configured
- **THEN** it loads and displays the first configured category

#### Scenario: Switch category
- **WHEN** the user presses `Tab`
- **THEN** the client advances to the next category and loads its feed

### Requirement: Feed sources
The client SHALL support the lobste.rs `hottest`, `newest`, and `active` feeds and per-tag feeds, and SHALL allow disabling or reordering categories via configuration.

#### Scenario: Tag-restricted feed
- **WHEN** the user supplies a tag (for example `lbx --tag go`)
- **THEN** the list shows stories from `https://lobste.rs/t/go.json`

#### Scenario: Custom category list
- **WHEN** the user passes `--categories hottest,newest,favorites`
- **THEN** only those categories appear, in that order

### Requirement: Story entry metadata
Each story row SHALL show the title, source domain, tags, score, comment count, submitter, and relative age.

#### Scenario: Rendered story row
- **WHEN** a feed is displayed
- **THEN** every row includes title, domain, tags, score, flaired comment count, submitter, and age

### Requirement: Refresh
The user SHALL be able to reload the current feed, preserving the highlighted story when it still exists.

#### Scenario: Refresh keeps selection
- **WHEN** the user presses `r` while a story is highlighted
- **THEN** the feed reloads and the same story remains highlighted if present

### Requirement: Story actions from the list
From the list, `Enter` SHALL open the story's comment section, `Space` SHALL open reader mode, `o` SHALL open the story URL in the system browser, and `c` SHALL open the comments URL in the system browser.

#### Scenario: Open comments
- **WHEN** the user presses `Enter` on a highlighted story
- **THEN** the comment section for that story is displayed

#### Scenario: Open article in browser
- **WHEN** the user presses `o` on a highlighted story
- **THEN** the story URL is opened in the system browser
