---
trigger: always_on
---

# Golang Build Environment Rule

## Context
This rule applies whenever the agent is tasked with building, running, or testing the Golang application.

### Required Command Patterns

* **For Running (Development):**
  `make run`

* **For Building:**
  `make build-prod`

* **For Executing the Binary:**
  `./postfixadmin server`

---