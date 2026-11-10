package tools

import "testing"

func TestEmployeePolicyOnlyNarrowsCapabilities(t *testing.T) {
	for _, raw := range []string{`{"write_requires_approval":false}`, `{"critical_requires_elevated_approval":false}`, `{"allowed_tools":"*"}`} {
		if _, err := ParseEmployeePolicy([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	p, err := ParseEmployeePolicy([]byte(`{"allowed_tools":["task.create","client.read"],"denied_tools":["task.create"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if p.Allows("task.create") || p.Allows("shell.exec") || !p.Allows("client.read") {
		t.Fatal("policy widened access")
	}
	p, err = ParseEmployeePolicy([]byte(`{"allowed_tools":[]}`))
	if err != nil || p.Allows("client.read") {
		t.Fatal("empty allowlist must deny all")
	}
}
