## ADDED Requirements

### Requirement: Image published to GitHub Packages
A GitHub Actions workflow SHALL build the final Docker stage and push it to `ghcr.io/jniltinho/go-postfixadmin` for every pushed semver tag matching `v*.*.*`, and SHALL support manual runs. It SHALL use only `GITHUB_TOKEN` with `packages: write` and `contents: read`.

#### Scenario: Release tag
- **WHEN** tag `v1.2.3` is pushed
- **THEN** the image is pushed with tags `1.2.3`, `1.2` and `latest`

#### Scenario: Manual run without tag
- **WHEN** the workflow is run manually on a branch
- **THEN** the image is built and pushed with a `sha-<short>` tag only, without moving `latest`

#### Scenario: Pre-release tag
- **WHEN** tag `v1.3.0-rc.1` is pushed
- **THEN** the image is pushed as `1.3.0-rc.1` and `latest` is not moved

### Requirement: Multi-arch and metadata
The image SHALL be published for `linux/amd64` and `linux/arm64` and SHALL carry OCI labels for source, version, revision, created date and license, linking the package to the repository.

#### Scenario: Pull on ARM
- **WHEN** `docker pull ghcr.io/jniltinho/go-postfixadmin:latest` runs on an arm64 host
- **THEN** the arm64 image is pulled and `postfixadmin version` prints the tag version and commit

#### Scenario: Package page
- **WHEN** the package is opened on GitHub
- **THEN** it is linked to the `jniltinho/go-postfixadmin` repository and shows the description set as an `index:org.opencontainers.image.description` annotation on the multi-arch manifest index

#### Scenario: Anonymous pull
- **WHEN** a client without GitHub credentials runs `docker pull ghcr.io/jniltinho/go-postfixadmin:latest`
- **THEN** the pull succeeds because the package is public

### Requirement: Documented install from GHCR
`README.md` and `DOCUMENTS/setup/README.md` SHALL document running the published image and `docker-compose.yml` SHALL reference it as the default image while keeping a build option.

#### Scenario: Quick start
- **WHEN** a user follows the README Docker quick start
- **THEN** the stack starts from the GHCR image without building locally
