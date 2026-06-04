# Releasing a Go Project

This skill guides the complete process for creating version tags and GitHub Releases in any Go project.
It applies to this project (`go-postfixadmin`) and can be adapted to any Go repo.

---

## Before You Start

Gather project facts automatically:

```bash
# Project name and remote
APP=$(basename $(git rev-parse --show-toplevel))
REMOTE=$(git remote get-url origin)

# Latest tag and next tag candidate
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
echo "Last tag: $LAST_TAG | Remote: $REMOTE"
```

Check prerequisites:

```bash
gh auth status          # must have repo scope
git status              # must be clean
git fetch origin
git checkout main
git pull --ff-only
```

---

## Versioning Scheme

This project uses `v1.0.x` (semantic-like patch increments):

- `v1.0.83`, `v1.0.84`, `v1.0.85`, …
- Tags are **always** prefixed with `v`.
- GitHub Release titles must also use the `v` prefix.

To determine the next version automatically:

```bash
LAST=$(git describe --tags --abbrev=0)
NEXT="v1.0.$((${LAST##*.} + 1))"
echo "Next version: $NEXT"
```

---

## Release Approaches

| Approach             | Best For                     | Notes Quality                       | Recommended |
|----------------------|------------------------------|-------------------------------------|-------------|
| Tag + push only      | Hotfixes, CI-only bumps      | Minimal (auto-generated link)       | No          |
| Tag + polished notes | Regular user-facing releases | Structured sections with emojis     | **Yes**     |

---

## Recommended Process

### 1. Review changes since last release

```bash
git log $LAST_TAG..HEAD --oneline
git log $LAST_TAG..HEAD --stat   # more detail
```

### 2. Categorize commits

Use these sections in this exact order. Omit any section that has no items.

| Section | Prefix(es) | Description |
|---------|-----------|-------------|
| **✨ New Features** | `feat:` | New user-visible functionality |
| **🔧 Improvements** | `fix:`, `perf:`, `refactor:`, `ci:` | Fixes, perf, refactors, CI |
| **🧹 Cleanup** | `chore:`, `cleanup:` | Dead code removal, dep cleanup |
| **📚 Documentation** | `docs:`, `doc:` | README, guides, comments |

### 3. Create an annotated tag

```bash
git tag -a $NEXT -m "Release $NEXT"
```

### 4. Push the tag (triggers CI)

```bash
git push origin $NEXT
```

The `release.yml` workflow will:

1. Build the frontend (`make frontend`)
2. Compile and compress the Go binary (`make build-prod` → UPX)
3. Package `postfixadmin_X.Y.Z_linux_amd64.tar.gz`
4. Build the Debian package (`make deb`)
5. Build the RPM package (`make rpm`)
6. Create the GitHub Release and attach all three artifacts

### 5. Monitor CI

```bash
gh run list --limit 5
gh release view $NEXT   # wait until assets appear
```

### 6. Edit release notes

Prepare the notes file:

```bash
cat > /tmp/release-notes-$NEXT.md << 'EOF'
# Release vX.Y.Z

## ✨ New Features

- ...

## 🔧 Improvements

- ...

## 🧹 Cleanup

- ...

## 📚 Documentation

- ...

**Full Changelog**: https://github.com/jniltinho/go-postfixadmin/compare/vPREV...vNEXT
EOF
```

Then publish:

```bash
gh release edit $NEXT \
  --title "$NEXT" \
  --notes-file /tmp/release-notes-$NEXT.md
```

### 7. Verify

```bash
gh release view $NEXT
```

Check:
- Title matches `$NEXT`
- Structured sections are present
- Three artifacts attached: `.tar.gz`, `.deb`, `.rpm`
- Full Changelog link points to the correct range

---

## Quick Reference

```bash
# Determine next version
LAST=$(git describe --tags --abbrev=0)
NEXT="v1.0.$((${LAST##*.} + 1))"

# Tag and push
git tag -a $NEXT -m "Release $NEXT"
git push origin $NEXT

# Monitor CI
gh run list --limit 3
gh release view $NEXT

# Edit notes after CI finishes
gh release edit $NEXT --title "$NEXT" --notes-file /tmp/release-notes-$NEXT.md

# Final check
gh release view $NEXT
```

---

## Workflow Capabilities (`release.yml`)

| Artifact              | Status  |
|-----------------------|---------|
| `linux/amd64` binary  | ✅ Built with UPX compression |
| `.tar.gz` tarball     | ✅ |
| `.deb` package        | ✅ Includes systemd services + config |
| `.rpm` package        | ✅ Includes systemd services + config |
| Multi-arch builds     | ❌ Not yet implemented |
| Auto release notes    | ❌ Notes are added manually (this skill) |

---

## Adapting to Another Go Project

Replace these project-specific values when reusing this skill:

| Item | This project | Your project |
|------|-------------|--------------|
| Repo URL | `github.com/jniltinho/go-postfixadmin` | your repo |
| Version scheme | `v1.0.x` | `v0.x.y` or semver |
| Build command | `make build-prod` | your build target |
| Artifacts | `.tar.gz` + `.deb` + `.rpm` | whatever CI produces |
| Branch | `main` | `main` or `master` |
