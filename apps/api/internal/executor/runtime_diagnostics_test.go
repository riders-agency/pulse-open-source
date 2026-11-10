package executor

import "testing"

func TestDiagnosticsDoesNotReportWaitingAsFailure(t *testing.T) {
	for _, status := range []string{"waiting", "pending", "running", "not_started", "skipped", "cancelled"} {
		got := DiagnosticsFromStored("office.approve", []byte(`{}`), []byte(`{"provider":"pulse","ok":true}`), nil, status)
		if got["status"] != status || got["reason_code"] != nil {
			t.Fatalf("%s: unexpected failure diagnostics: %#v", status, got)
		}
	}
	errMsg := "provider unavailable"
	got := DiagnosticsFromStored("office.classify", []byte(`{}`), []byte(`{}`), &errMsg, "failed")
	if got["status"] != "failed" {
		t.Fatalf("real failures must stay visible: %#v", got)
	}
}
