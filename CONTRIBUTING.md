# Contributing to Go-Postfixadmin

First off, thank you for considering contributing to Go-Postfixadmin! It's people like you that make this a great tool for the community.

## 🚀 How Can I Contribute?

### Reporting Bugs
If you find a bug, please open an issue on GitHub. Include:
- A clear and descriptive title.
- Steps to reproduce the bug.
- Actual vs. expected behavior.
- Screenshots if applicable.

### Suggesting Enhancements
Have an idea to make Go-Postfixadmin better?
- Open an issue with the "enhancement" label.
- Describe the feature and why it would be useful.

### Pull Requests
1. **Fork the repository** and create your branch from `main`.
2. **Setup your environment**:
   - Install Go (v1.26+) and Node.js (v20+).
   - Run `make deps` to install all dependencies.
3. **Make your changes**:
   - Follow the [Clean Code](.agent/skills/clean-code/SKILL.md) principles.
   - If adding a new feature, ensure it's documented.
   - If adding a new language, simply add a `.toml` file to the `locales/` directory.
4. **Test your changes**:
   - Run `make run` to test locally.
   - Ensure the UI looks good across different screen sizes.
5. **Submit your PR**:
   - Provide a concise title and detailed description of your changes.
   - Reference any related issues.

## 🎨 Design Guidelines
We use a **Neo-Brutalism** design aesthetic:
- Thick black borders (`2px` or `4px`).
- High-contrast colors.
- Sharp shadows (`neo-shadow`).
- Use of Lucide icons.

## 🌐 Localization (i18n)
The project uses [gotext](https://github.com/leonelquinteros/gotext) with GNU Gettext `.po` files. To add a new language:
1. Create a new directory in `locales/` (e.g., `locales/fr/`).
2. Copy `locales/en/default.po` to `locales/fr/default.po`.
3. Translate the `msgstr` values (keep `msgid` keys unchanged).
4. Add a link to the new language in the language switcher in the following files:
   - `views/layout.html`
   - `views/login.html`
   - `views/users/layout.html`
   - `views/users/login.html`
5. Add the language code to `SetLanguage` in `internal/handlers/handlers.go` and to `Render` in `internal/server/render.go`.

## 🛠 Useful Commands
- `make build-prod`: Build the production binary.
- `make watch-css`: Watch for CSS changes (Tailwind).
- `make clean`: Clean up generated files.

## 📜 Code of Conduct
Please be respectful and professional in all interactions within the project.

---
Happy coding! 🚀
