## ADDED Requirements

### Requirement: Go 1.27
The module SHALL declare Go 1.27 and every build path (Makefile, Dockerfile, GitHub Actions) SHALL use Go 1.27.

#### Scenario: Consistent Go version
- **WHEN** `go.mod`, `Dockerfile` and `.github/workflows/*.yml` are inspected
- **THEN** all of them reference Go 1.27

#### Scenario: Tests pass
- **WHEN** `go vet ./...` and `go test ./...` run with Go 1.27
- **THEN** both succeed

### Requirement: Current Echo v5
The server SHALL depend on the latest Echo v5.4.x release and on `github.com/labstack/echo-contrib/v5` (the Echo v5 series), with no Echo v4 module in the build graph.

#### Scenario: No retracted modules
- **WHEN** `go list -m -u -retracted all` runs
- **THEN** no direct dependency is reported as retracted

#### Scenario: Routes behave as before
- **WHEN** HTTP regression tests for login, token refresh, an RBAC-guarded route (with and without the permission) and a public route, written and passing before the upgrade, run after the upgrade
- **THEN** they return the same status codes as before

### Requirement: Vite+ frontend toolchain
The frontend SHALL build with Vite+ (`vite-plus`, CLI `vp`) at a pinned version. `npm run dev`, `npm run build`, `npm run check` and `npm run test` SHALL call `vp`; type checking with `vue-tsc` SHALL be kept; the build output SHALL remain `web/dist`.

#### Scenario: Full build
- **WHEN** `make build` runs on a clean checkout
- **THEN** the SPA is built by `vp build` into `web/dist` and embedded in `bin/postfixadmin`

#### Scenario: TypeScript config is the one used
- **WHEN** the frontend directory is listed
- **THEN** only `vite.config.ts` exists (no generated `vite.config.js` / `.d.ts`) and the generated files are git-ignored

#### Scenario: Dev server proxy
- **WHEN** `npm run dev` runs and the Go server listens on port 8080
- **THEN** requests to `/api` and `/lang` from the dev server reach the Go server
