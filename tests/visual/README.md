# Visual checks

`capture.sh` drives [agent-browser](https://github.com/vercel-labs/agent-browser) through the SPA and writes one
screenshot per theme, page and resolution to `$OUT` (default `tests/visual/out`, git-ignored). With `CHECK=1` (default)
it also fails on:

- horizontal scrolling at any size;
- a login card narrower than 22% of the screen at QHD/4K;
- a root font outside 14–16px at FullHD, below 16px at QHD or below 18px at 4K.

It logs in once per role and moves through the SPA router, because every page load and login counts against the login
rate limiter (3 burst, 1 per 12 s), which stays on.

```bash
# a server with seeded data (see DEVELOPMENT.md), then:
BASE_URL=http://localhost:8080 \
ADMIN_USER=admin@example.com ADMIN_PASS=adminpassword \
MAILBOX_USER=ana@example.com MAILBOX_PASS='...' \
THEMES="default xp matrix" SIZES="1366x768 1920x1080 2560x1440 3840x2160" \
tests/visual/capture.sh
```

`PAGES` (default `dashboard domains mailboxes settings`) picks the admin pages; the login page, the add-mailbox modal
and (with `MAILBOX_USER`) the user portal are always captured. All themes: `THEMES="$(grep -o "id: '[a-z-]*'" frontend/src/themes/index.ts | cut -d"'" -f2 | tr '\n' ' ')"`.
