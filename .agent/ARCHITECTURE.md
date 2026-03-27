# Go-PostfixAdmin Agent Kit

> AI Agent Capability Toolkit — tailored for this project (Go · Echo · GORM · Tailwind CSS · Cobra CLI · Docker)

---

## 🏗️ Directory Structure

```plaintext
.agent/
├── ARCHITECTURE.md          # This file
├── agents/                  # Specialist Agents
├── skills/                  # Domain-specific knowledge modules
├── workflows/               # Slash command procedures
├── rules/                   # Global rules (always loaded)
└── scripts/                 # Master validation scripts
```

---

## 🤖 Agents (16)

| Agent                    | Focus                        |
| ------------------------ | ---------------------------- |
| `orchestrator`           | Multi-agent coordination     |
| `project-planner`        | Discovery, task planning     |
| `frontend-specialist`    | HTML templates, Tailwind CSS |
| `backend-specialist`     | Go handlers, business logic  |
| `database-architect`     | GORM schema, SQL, migrations |
| `devops-engineer`        | Docker, Makefile, CI/CD      |
| `security-auditor`       | Auth, sessions, OWASP        |
| `test-engineer`          | Testing strategies           |
| `qa-automation-engineer` | E2E, CI pipelines            |
| `debugger`               | Root cause analysis          |
| `performance-optimizer`  | Profiling, optimization      |
| `documentation-writer`   | Docs, changelogs             |
| `product-manager`        | Requirements, user stories   |
| `product-owner`          | Strategy, backlog            |
| `code-archaeologist`     | Refactoring, legacy code     |
| `explorer-agent`         | Codebase analysis            |

---

## 🧩 Skills

### Backend & CLI

| Skill          | Description                        |
| -------------- | ---------------------------------- |
| `golang-pro`   | Go idioms, patterns, performance   |
| `api-patterns` | REST, auth, rate-limiting, response patterns |
| `bash-linux`   | Linux commands, shell scripting    |

### Database

| Skill             | Description                          |
| ----------------- | ------------------------------------ |
| `database-design` | Schema design, indexing, GORM, migrations |

### Frontend & UI

| Skill                   | Description                                      |
| ----------------------- | ------------------------------------------------ |
| `tailwind-patterns`     | Tailwind CSS v4 utilities                        |
| `frontend-design`       | UI/UX patterns, color systems, typography        |
| `web-design-guidelines` | Web UI audit — accessibility, UX, performance    |
| `ui-ux-pro-max`         | 50 styles, palettes, fonts (stack: html-tailwind) |

### Infrastructure & Deployment

| Skill                   | Description                     |
| ----------------------- | ------------------------------- |
| `deployment-procedures` | CI/CD, Docker, deploy workflows |
| `server-management`     | Linux server, mail infrastructure |

### Testing & Quality

| Skill                   | Description                    |
| ----------------------- | ------------------------------ |
| `testing-patterns`      | Unit, integration, strategies  |
| `webapp-testing`        | E2E with Playwright            |
| `tdd-workflow`          | Test-driven development        |
| `code-review-checklist` | Code review standards          |
| `lint-and-validate`     | Linting, static analysis       |

### Security

| Skill                   | Description              |
| ----------------------- | ------------------------ |
| `vulnerability-scanner` | Security auditing, OWASP |

### Architecture & Planning

| Skill                     | Description                  |
| ------------------------- | ---------------------------- |
| `architecture`            | System design patterns       |
| `app-builder`             | CLI and API scaffolding      |
| `plan-writing`            | Task planning, breakdown     |
| `brainstorming`           | Socratic questioning         |
| `parallel-agents`         | Multi-agent coordination     |
| `behavioral-modes`        | Agent personas               |
| `intelligent-routing`     | Request routing patterns     |

### Internationalization & Docs

| Skill                     | Description                    |
| ------------------------- | ------------------------------ |
| `i18n-localization`       | GNU Gettext, .po files, gotext |
| `documentation-templates` | Doc formats and templates      |

### Performance & Debugging

| Skill                   | Description                    |
| ----------------------- | ------------------------------ |
| `performance-profiling` | Profiling, optimization        |
| `systematic-debugging`  | Troubleshooting, root cause    |
| `clean-code`            | Coding standards               |

---

## 🔄 Workflows (11)

| Command          | Description              |
| ---------------- | ------------------------ |
| `/brainstorm`    | Socratic discovery       |
| `/create`        | Create new features      |
| `/debug`         | Debug issues             |
| `/deploy`        | Deploy application       |
| `/enhance`       | Improve existing code    |
| `/orchestrate`   | Multi-agent coordination |
| `/plan`          | Task breakdown           |
| `/preview`       | Preview changes          |
| `/status`        | Check project status     |
| `/test`          | Run tests                |
| `/ui-ux-pro-max` | Design with 50 styles    |

---

## 🔗 Quick Reference

| Need       | Agent                 | Skills                                      |
| ---------- | --------------------- | ------------------------------------------- |
| Go feature | `backend-specialist`  | golang-pro, api-patterns, database-design   |
| UI/CSS     | `frontend-specialist` | tailwind-patterns, frontend-design          |
| Database   | `database-architect`  | database-design                             |
| Security   | `security-auditor`    | vulnerability-scanner                       |
| Testing    | `test-engineer`       | testing-patterns, tdd-workflow              |
| Debug      | `debugger`            | systematic-debugging                        |
| Deploy     | `devops-engineer`     | deployment-procedures, server-management    |
| Docs       | `documentation-writer`| documentation-templates                     |
| Plan       | `project-planner`     | brainstorming, plan-writing, architecture   |
| i18n       | `backend-specialist`  | i18n-localization                           |
