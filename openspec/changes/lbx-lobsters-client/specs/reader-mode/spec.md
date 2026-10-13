# Spec Delta

## Purpose

Read a story's linked article inside the terminal by extracting its main content and rendering it, so the user does not have to leave the client.

## ADDED Requirements

### Requirement: Article extraction
Pressing `Space` on a story SHALL fetch the linked URL and extract its main readable content, discarding navigation, scripts, and ads.

#### Scenario: Readable extraction
- **WHEN** the user presses `Space` on a story with a readable article
- **THEN** the reader shows the article's title and body text rather than raw HTML

#### Scenario: Non-article target
- **WHEN** the linked URL is not an HTML article (for example a PDF or video)
- **THEN** the client reports that the article cannot be shown and offers to open it in the browser

### Requirement: Content rendering
The reader SHALL render headings, paragraphs, lists, blockquotes, code blocks, and inline links with terminal formatting using the active theme.

#### Scenario: Structured content
- **WHEN** the extracted article contains headings, lists, and code
- **THEN** each is rendered as a distinguishable element in the terminal

### Requirement: Heading navigation
In reader mode the user SHALL scroll with vim bindings and jump between headings with `n`/`N`.

#### Scenario: Next heading
- **WHEN** the user presses `n`
- **THEN** the view moves to the next heading in the article

### Requirement: In-content search
The user SHALL be able to search within the loaded content with `/` and step through matches with `n`/`N`.

#### Scenario: Search and step
- **WHEN** the user presses `/`, enters a term, and presses `Enter`
- **THEN** matches are highlighted and `n` moves to the next match

### Requirement: Image display on supported terminals
On terminals with Kitty graphics support, pressing `Enter` in reader mode SHALL attempt to display the article's images; on unsupported terminals images SHALL be shown as their alt text or a placeholder.

#### Scenario: Unsupported terminal
- **WHEN** the terminal lacks Kitty graphics support and the user presses `Enter`
- **THEN** images are represented by alt text or a placeholder without error
