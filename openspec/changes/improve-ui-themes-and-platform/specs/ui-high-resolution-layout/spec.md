## ADDED Requirements

### Requirement: Proportional scaling on large displays
The SPA layout SHALL scale its root font size, header height, sidebar width and content width with the viewport so that FullHD (1920x1080), QHD (2560x1440) and 4K (3840x2160) at device pixel ratio 1 show proportionally sized text and controls, without CSS `zoom` or `transform: scale`.

#### Scenario: 4K login page
- **WHEN** the login page is opened at 3840x2160
- **THEN** the login card is at least 22% of the viewport width and the base text is at least 18 px

#### Scenario: QHD dashboard
- **WHEN** the dashboard is opened at 2560x1440
- **THEN** statistic cards use at least 3 columns and the base text is at least 16 px

#### Scenario: FullHD unchanged density
- **WHEN** any admin page is opened at 1920x1080
- **THEN** the base text is between 14 px and 16 px and there is no horizontal scrollbar

### Requirement: Use of available width
Tables, dialogs and forms SHALL widen with the content area up to a maximum content width, and the content area SHALL NOT leave more than 25% of the viewport empty on each side at QHD and 4K.

#### Scenario: Mailbox table at 4K
- **WHEN** the mailboxes page is opened at 3840x2160
- **THEN** the table fills the content area up to the maximum content width and no column text is truncated that was not truncated at 1920x1080

### Requirement: Small screens keep working
The layout SHALL remain usable at 1366x768 and SHALL NOT introduce horizontal scrolling on any page at that size.

#### Scenario: Laptop viewport
- **WHEN** any page is opened at 1366x768
- **THEN** the sidebar, header and content are fully visible without horizontal scrolling

### Requirement: Visual regression matrix
The repository SHALL contain an `agent-browser` script that captures login, dashboard, mailboxes, one modal and the user portal for every theme at 1920x1080, 2560x1440 and 3840x2160 and checks the scenarios of this capability.

#### Scenario: Matrix run
- **WHEN** the script runs against the docker-compose stack
- **THEN** it writes one screenshot per theme, page and resolution and exits non-zero if any check fails
