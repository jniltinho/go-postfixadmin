---
trigger: always_on
---

# Golang Build Environment Rule

## Context
This rule applies whenever the agent is tasked with building, running, or testing the Golang application.

## Instructions
Before executing any Go-related build or execution commands (such as `go build`, `go run`, or `go test`), you **must** ensure the `DB_URL` environment variable is exported in the terminal session.


### Required Command Patterns

* **For Running (Development):**
  `go run main.go server`

* **For Building:**
  `go build -o postfixadmin`

* **For Executing the Binary:**
  `./postfixadmin server`

---