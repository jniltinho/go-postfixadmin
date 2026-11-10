#!/usr/bin/env bash
# Captures the visual matrix (theme x page x resolution) with agent-browser and checks the layout rules of the
# ui-high-resolution-layout spec. Needs a running server with seeded data (see tests/visual/README.md).
# Logs in once per role and navigates through the SPA router: every page load and login counts against the
# login rate limiter (3 burst, 1 per 12s), which must stay enabled.
set -euo pipefail
BASE_URL=${BASE_URL:-http://localhost:8080}
OUT=${OUT:-tests/visual/out}
THEMES=${THEMES:-default}
SIZES=${SIZES:-"1366x768 1920x1080 2560x1440 3840x2160"}
PAGES=${PAGES:-"dashboard domains mailboxes settings"}
CHECK=${CHECK:-1}
ab() { agent-browser "$@" >/dev/null; }
js() { agent-browser eval "$1" | tail -1; }
fail=0
mkdir -p "$OUT"

# like the theme store: data-theme plus the theme's color-scheme (clearing the inline one lets the theme CSS set it)
theme() { js "(() => { localStorage.setItem('theme','$1'); const e = document.documentElement; e.dataset.theme='$1'; e.style.colorScheme=''; localStorage.setItem('theme-scheme', getComputedStyle(e).colorScheme); return 'ok' })()" >/dev/null; ab wait 150; }
go() { js "history.pushState({}, '', '$1'); dispatchEvent(new PopStateEvent('popstate')); 'ok'" >/dev/null; ab wait 900; }

login() { # $1 path, $2 user, $3 pass; retries after the rate limiter's refill (1 per 12s)
  for try in 1 2 3; do
    ab open "$BASE_URL$1"
    ab wait 1500
    ab fill 'form input:not([type=password])' "$2"
    ab fill 'form input[type=password]' "$3"
    ab click 'form button[type=submit]'
    ab wait 2000
    [ "$(js 'location.pathname')" != "\"$1\"" ] && return 0
    sleep 13
  done
  echo "login failed at $1"
  exit 2
}

check() { # $1 name, $2 width
  [ "$CHECK" = 1 ] || return 0
  local r
  r=$(js "(() => { const w = innerWidth, fs = parseFloat(getComputedStyle(document.documentElement).fontSize);
    const errs = [];
    if (document.documentElement.scrollWidth > w + 1) errs.push('horizontal scroll');
    const card = document.querySelector('.login-card-neo, .login-card')?.getBoundingClientRect().width;
    if (card && w >= 2560 && card < w * 0.22) errs.push('login card ' + Math.round(card) + 'px');
    if (w >= 3840 && fs < 18) errs.push('root font ' + fs + 'px');
    else if (w >= 2560 && w < 3840 && fs < 16) errs.push('root font ' + fs + 'px');
    else if (w >= 1920 && w < 2560 && (fs < 14 || fs > 16)) errs.push('root font ' + fs + 'px');
    return errs.join('; ') || 'ok' })()")
  r=${r//\"/}
  if [ "$r" != ok ]; then echo "FAIL $1 @ $2: $r"; fail=1; fi
}

matrix() { # $1 page name: one screenshot per theme and size of the current page
  for t in $THEMES; do
    theme "$t"
    for s in $SIZES; do
      ab set viewport "${s%x*}" "${s#*x}"
      ab wait 300
      agent-browser screenshot "$OUT/$t-$1-$s.png" >/dev/null
      check "$t/$1" "${s%x*}"
    done
  done
}

ab open "$BASE_URL/login"
ab cookies clear
js "localStorage.clear(); 'ok'" >/dev/null
ab open "$BASE_URL/login"
ab wait 1500
matrix login

login /login "${ADMIN_USER:-admin@example.com}" "${ADMIN_PASS:-adminpassword}"
for p in $PAGES; do go "/$p"; matrix "$p"; done
go /mailboxes
ab click '.btn-add-big'
ab wait 400
matrix modal

if [ -n "${MAILBOX_USER:-}" ]; then
  js "localStorage.removeItem('access_token'); localStorage.removeItem('user_info'); 'ok'" >/dev/null
  sleep 13
  login /users/login "$MAILBOX_USER" "$MAILBOX_PASS"
  matrix portal
fi
js "localStorage.clear(); 'ok'" >/dev/null
exit $fail
