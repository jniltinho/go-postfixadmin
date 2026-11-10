## ADDED Requirements

### Requirement: Role-based theme contract
The SPA SHALL style every component through semantic CSS variables declared in `frontend/src/style.css` and overridden per theme by `[data-theme='<id>']` blocks. Components SHALL NOT contain hard-coded color values or Tailwind palette color classes, except inside theme files.

#### Scenario: No hard-coded colors remain
- **WHEN** the frontend sources outside `frontend/src/themes/` are scanned for hex colors and Tailwind palette color classes
- **THEN** no matches are found

#### Scenario: Default theme preserves the current look
- **WHEN** the role-contract refactor is complete and before any layout sizing change, with the `default` theme active
- **THEN** screenshots of login, dashboard, mailboxes, a modal and the user portal match the baselines captured before the change, apart from antialiasing differences

### Requirement: Theme catalog including XP
The system SHALL ship the themes `default`, `admanager`, `breeze`, `breeze-dark`, `violet`, `violet-dark`, `aurora`, `aurora-dark`, `crystal`, `crystal-dark`, `cone`, `graphite`, `grafana`, `rabbitmq`, `matrix`, `winamp`, `pink-panther`, `high-contrast`, `classic-gray` and `xp`. Each theme SHALL declare its color scheme (light or dark). Every theme registry id SHALL have a matching CSS block, and every CSS block SHALL have a registry entry.

#### Scenario: XP theme applied
- **WHEN** the user selects "Windows XP"
- **THEN** `<html>` has `data-theme="xp"`, the page background is the Luna dialog face color and title bars use the Luna blue gradient

#### Scenario: Registry and CSS stay in sync
- **WHEN** the unit test compares the registry ids with the `data-theme` selectors in `frontend/src/themes/*.css`
- **THEN** both sets are equal

#### Scenario: Text contrast
- **WHEN** the contrast check runs for every theme
- **THEN** body text on page and surface backgrounds has a contrast ratio of at least 4.5:1

### Requirement: Theme selection and persistence
The admin UI SHALL offer the themes in Settings › Appearance as cards with a live preview of each theme, and the mailbox user portal SHALL offer a theme select in its header. The choice SHALL be stored in browser `localStorage` and applied before the first paint on later visits. Precedence SHALL be: stored choice, then server default, then `default`.

#### Scenario: Choice survives reload
- **WHEN** the user selects `matrix` and reloads the page
- **THEN** `data-theme="matrix"` and `color-scheme: dark` are set before the first paint, without first showing another theme or light native controls

#### Scenario: Unknown stored value
- **WHEN** `localStorage` holds a theme id that is not in the registry
- **THEN** the server default theme is applied

#### Scenario: Server default unavailable
- **WHEN** no choice is stored and `GET /api/v1/version` fails
- **THEN** the `default` theme is applied and the UI remains usable

#### Scenario: Login pages are themed
- **WHEN** a theme is stored and the user opens the admin or user login page
- **THEN** the login page uses that theme

### Requirement: Server default theme
The configuration SHALL accept `ui.default_theme` (TOML `[ui] default_theme`, env `UI_DEFAULT_THEME`), defaulting to `default`. The public `GET /api/v1/version` response SHALL include it as `default_theme`. Both shipped TOML templates and `--generate-config` SHALL contain the key.

#### Scenario: Operator sets XP as default
- **WHEN** `ui.default_theme = "xp"` and a browser without a stored choice opens the SPA
- **THEN** the XP theme is applied

#### Scenario: Invalid configured theme
- **WHEN** `ui.default_theme` is not a known theme id
- **THEN** browsers without a stored choice fall back to the `default` theme

### Requirement: New screens in PT, EN and ES
Every text of Settings (all tabs, messages, dialogs, the Settings menu entry and breadcrumb) and of the portal's theme/language selects SHALL exist in English, Brazilian Portuguese and Spanish, with the same placeholders. Theme names are proper names; only their descriptive words ("dark", "Default", "High contrast") are translated.

#### Scenario: Portuguese
- **WHEN** the language is PT
- **THEN** Settings shows "CONFIGURAÇÕES", "APARÊNCIA", "SERVIÇOS", and "Breeze (escuro)"

#### Scenario: Dictionaries complete
- **WHEN** the unit test compares the PT and ES dictionaries with EN
- **THEN** they have the same keys, none empty, with the same placeholders

### Requirement: Language chosen in Settings
The UI language SHALL default to the browser language (PT, ES, else EN). The login pages and the admin header SHALL NOT show language buttons; the language SHALL be chosen in Settings › Appearance › Language (Automatic, English, Português, Español), saved per browser and shared by the login pages, the admin UI and the portal (which keeps a compact select, having no Settings page). Choosing a language also sets the backend `lang` cookie.

#### Scenario: Browser language
- **WHEN** a browser set to pt-BR opens the login page with nothing saved
- **THEN** it is in Portuguese, with no language buttons

#### Scenario: Back to automatic
- **WHEN** the admin picks "Automatic" in Settings
- **THEN** the saved choice is removed and the browser language is used again

### Requirement: Settings at AAA contrast
Every text of Settings (Appearance, Services, Terminal tabs, outside the terminal canvas and the theme previews) SHALL have a contrast of at least 7:1 against its background, or 4.5:1 for large text, in every theme; status SHALL not be told by color alone.

#### Scenario: Contrast check
- **WHEN** the in-page contrast checker runs on each Settings tab in default, xp, classic-gray, breeze-dark, matrix and high-contrast
- **THEN** it reports no text below the AAA ratio
