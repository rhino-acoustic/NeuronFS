package main

import (
	"testing"
)

// Test policy YAML (inline for test isolation)
const testPolicyYAML = `
version: "1.0"
roles:
  bot1:
    regions:
      brainstem:
        actions: ["read"]
      limbic:
        actions: ["read"]
      hippocampus:
        actions: ["read", "write"]
      sensors:
        actions: ["read"]
      cortex:
        actions: ["read", "write"]
      ego:
        actions: ["read"]
      prefrontal:
        actions: ["deny"]
  enfp:
    regions:
      brainstem:
        actions: ["deny"]
      cortex:
        actions: ["read", "write"]
      prefrontal:
        actions: ["read", "write"]
  entp:
    regions:
      brainstem:
        actions: ["read"]
      cortex:
        actions: ["read", "write"]
      prefrontal:
        actions: ["read", "write"]
  admin:
    regions:
      brainstem:
        actions: ["read", "write"]
      limbic:
        actions: ["read", "write"]
      hippocampus:
        actions: ["read", "write"]
      sensors:
        actions: ["read", "write"]
      cortex:
        actions: ["read", "write"]
      ego:
        actions: ["read", "write"]
      prefrontal:
        actions: ["read", "write"]
`

func setupTestPolicy(t *testing.T) {
	t.Helper()
	if err := LoadAccessPolicyFromBytes([]byte(testPolicyYAML)); err != nil {
		t.Fatalf("failed to load test policy: %v", err)
	}
}

func TestCanAccess_Bot1_PrefrontalDeny(t *testing.T) {
	setupTestPolicy(t)

	allowed, err := CanAccess("bot1", "prefrontal", "write")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("bot1 should NOT have write access to prefrontal")
	}
	t.Log("OK: bot1 write→prefrontal correctly DENIED")
}

func TestCanAccess_Bot1_CortexWrite(t *testing.T) {
	setupTestPolicy(t)

	allowed, err := CanAccess("bot1", "cortex", "write")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("bot1 should have write access to cortex")
	}
	t.Log("OK: bot1 write→cortex correctly ALLOWED")
}

func TestCanAccess_Bot1_BrainstemRead(t *testing.T) {
	setupTestPolicy(t)

	allowed, err := CanAccess("bot1", "brainstem", "read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("bot1 should have read access to brainstem")
	}
	t.Log("OK: bot1 read→brainstem correctly ALLOWED")
}

func TestCanAccess_Bot1_BrainstemWriteDenied(t *testing.T) {
	setupTestPolicy(t)

	allowed, err := CanAccess("bot1", "brainstem", "write")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("bot1 should NOT have write access to brainstem (not in actions list)")
	}
	t.Log("OK: bot1 write→brainstem correctly DENIED (implicit)")
}

func TestCanAccess_ENFP_BrainstemExplicitDeny(t *testing.T) {
	setupTestPolicy(t)

	allowed, err := CanAccess("enfp", "brainstem", "read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("enfp should NOT have read access to brainstem (explicit deny)")
	}
	t.Log("OK: enfp read→brainstem correctly DENIED (explicit)")
}

func TestCanAccess_Admin_FullAccess(t *testing.T) {
	setupTestPolicy(t)

	regions := []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"}
	for _, region := range regions {
		for _, action := range []string{"read", "write"} {
			allowed, err := CanAccess("admin", region, action)
			if err != nil {
				t.Fatalf("admin %s→%s error: %v", action, region, err)
			}
			if !allowed {
				t.Fatalf("admin should have %s access to %s", action, region)
			}
		}
	}
	t.Logf("OK: admin has full r/w access to all %d regions", len(regions))
}

func TestCanAccess_UnlistedRegion_DefaultDeny(t *testing.T) {
	setupTestPolicy(t)

	// enfp doesn't have limbic listed → default deny
	allowed, err := CanAccess("enfp", "limbic", "read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("unlisted region should default to DENY")
	}
	t.Log("OK: unlisted region correctly defaults to DENY (Zero Trust)")
}

func TestCanAccess_UnknownRole(t *testing.T) {
	setupTestPolicy(t)

	_, err := CanAccess("unknown_agent", "cortex", "read")
	if err != errRoleNotFound {
		t.Fatalf("expected errRoleNotFound, got: %v", err)
	}
	t.Log("OK: unknown role correctly rejected")
}

func TestCanAccess_InvalidAction(t *testing.T) {
	setupTestPolicy(t)

	_, err := CanAccess("bot1", "cortex", "execute")
	if err != errInvalidAction {
		t.Fatalf("expected errInvalidAction, got: %v", err)
	}
	t.Log("OK: invalid action correctly rejected")
}

func TestCanAccess_PolicyNotLoaded(t *testing.T) {
	savedPolicy := globalPolicy
	globalPolicy = nil
	defer func() { globalPolicy = savedPolicy }()

	_, err := CanAccess("bot1", "cortex", "read")
	if err != errPolicyNotLoaded {
		t.Fatalf("expected errPolicyNotLoaded, got: %v", err)
	}
	t.Log("OK: nil policy correctly returns error")
}

func TestCanAccess_CaseInsensitive(t *testing.T) {
	setupTestPolicy(t)

	allowed, err := CanAccess("BOT1", "CORTEX", "WRITE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("case-insensitive matching should work")
	}
	t.Log("OK: case-insensitive role/region/action matching verified")
}

func TestListRolePermissions(t *testing.T) {
	setupTestPolicy(t)

	summary, err := ListRolePermissions("bot1")
	if err != nil {
		t.Fatalf("ListRolePermissions: %v", err)
	}
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
	t.Logf("OK: bot1 permissions summary:\n%s", summary)
}
