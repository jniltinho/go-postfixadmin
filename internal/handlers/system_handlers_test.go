package handlers

import "testing"

func TestParseUnitShow(t *testing.T) {
	got := parseUnitShow("postfix.service", "LoadState=loaded\nActiveState=active\nSubState=exited\nUnitFileState=enabled\nActiveEnterTimestamp=Sat 2026-11-07 20:15:02 -03\n")
	if !got.Found || got.Active != "active" || got.Sub != "exited" || got.Enabled != "enabled" || got.Since != "Sat 2026-11-07 20:15:02 -03" {
		t.Fatalf("loaded unit: %+v", got)
	}
	if got := parseUnitShow("dovecot.service", "LoadState=not-found\nActiveState=inactive\n"); got.Found {
		t.Fatalf("missing unit must not be found: %+v", got)
	}
	if got := parseUnitShow("x.service", ""); got.Found {
		t.Fatalf("empty output must not be found: %+v", got)
	}
}
