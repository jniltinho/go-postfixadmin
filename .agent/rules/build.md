---
trigger: always_on
---

# Golang Build Environment Rule

## Context
This rule applies whenever the agent is tasked with building, running, or testing the Golang application.

## Instructions
Before executing any Go-related build or execution commands (such as `go build`, `go run`, or `go test`), you **must** ensure the `DATABASE_URL` environment variable is exported in the terminal session.


### Required Command Patterns

* **For Running (Development):**
  `source .env && go run main.go`

* **For Building:**
  `source .env && go build -o postfixadmin`

* **For Executing the Binary:**
  `source .env && ./postfixadmin server`

---

### Pro-Tip: Persistent Local Environment
If a `.env` file exists, use `source .env` or prefix commands with `env $(cat .env | xargs)` before execution to ensure all variables are loaded correctly.

## Exceptions
If the `DATABASE_URL` is already defined in a local `.env` file and the project uses a loader (like `godotenv`), verify if the export is still necessary. Otherwise, prioritize the manual export as defined above.