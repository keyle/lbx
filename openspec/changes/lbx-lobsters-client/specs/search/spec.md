# Spec Delta

## Purpose

Launch a lobste.rs search from within the client without scraping the site.

## ADDED Requirements

### Requirement: Open search in the browser
Pressing `/` in the main view SHALL prompt for a search term and open the lobste.rs search results page for that term in the user's default browser.

#### Scenario: Search stories
- **WHEN** the user presses `/`, types a term, and presses `Enter`
- **THEN** the browser opens the lobste.rs search results page for that term

#### Scenario: Cancel search
- **WHEN** the user presses `Esc` while entering a search term
- **THEN** no browser page is opened

### Requirement: No server-side search scraping
The client SHALL NOT attempt to fetch or parse lobste.rs search results itself, because that route is protected by a proof-of-work challenge.

#### Scenario: No search request
- **WHEN** the user searches
- **THEN** the client makes no HTTP request to a lobste.rs search route
