package email

import (
	"net/smtp"
	"testing"
)

func TestSMTPAuthNegotiatesAdvertisedMechanism(t *testing.T) {
	t.Parallel()

	login := NewSMTPAuth("sender@example.test", "secret", "smtp.example.test")
	mechanism, initial, err := login.Start(&smtp.ServerInfo{
		Name: "smtp.example.test", TLS: true, Auth: []string{"PLAIN", "LOGIN"},
	})
	if err != nil || mechanism != "LOGIN" || initial != nil {
		t.Fatalf("LOGIN preference failed: mechanism=%q initial=%q err=%v", mechanism, initial, err)
	}

	mechanism, initial, err = login.Start(&smtp.ServerInfo{
		Name: "smtp.example.test", TLS: true, Auth: []string{"LOGIN"},
	})
	if err != nil || mechanism != "LOGIN" || initial != nil {
		t.Fatalf("LOGIN negotiation failed: mechanism=%q initial=%q err=%v", mechanism, initial, err)
	}
	username, err := login.Next([]byte("Username:"), true)
	if err != nil || string(username) != "sender@example.test" {
		t.Fatalf("LOGIN username response failed: response=%q err=%v", username, err)
	}
	password, err := login.Next([]byte("Password:"), true)
	if err != nil || string(password) != "secret" {
		t.Fatalf("LOGIN password response failed: response=%q err=%v", password, err)
	}

	plain := NewSMTPAuth("sender@example.test", "secret", "smtp.example.test")
	mechanism, initial, err = plain.Start(&smtp.ServerInfo{
		Name: "smtp.example.test", TLS: true, Auth: []string{"PLAIN"},
	})
	if err != nil || mechanism != "PLAIN" || string(initial) != "\x00sender@example.test\x00secret" {
		t.Fatalf("PLAIN fallback failed: mechanism=%q initial=%q err=%v", mechanism, initial, err)
	}
}

func TestSMTPAuthRejectsUnsafeOrUnsupportedSessions(t *testing.T) {
	t.Parallel()
	auth := NewSMTPAuth("sender@example.test", "secret", "smtp.example.test")

	for name, server := range map[string]*smtp.ServerInfo{
		"without TLS":      {Name: "smtp.example.test", TLS: false, Auth: []string{"LOGIN"}},
		"different host":   {Name: "attacker.example.test", TLS: true, Auth: []string{"LOGIN"}},
		"unsupported auth": {Name: "smtp.example.test", TLS: true, Auth: []string{"CRAM-MD5"}},
	} {
		name, server := name, server
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, _, err := auth.Start(server); err == nil {
				t.Fatal("unsafe SMTP session unexpectedly accepted")
			}
		})
	}
	if _, err := auth.Next([]byte("OTP:"), true); err == nil {
		t.Fatal("unknown SMTP challenge unexpectedly accepted")
	}
}
