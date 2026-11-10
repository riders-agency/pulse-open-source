package uon

import (
	"strings"
	"testing"
)

func TestBuildSource(t *testing.T) {
	got := BuildSource("example.test", "заявка на тур")
	want := "example.test / заявка на тур"
	if got != want {
		t.Fatalf("BuildSource() = %q, want %q", got, want)
	}
}

func TestFlattenLeadPostExtendedFields(t *testing.T) {
	post := flattenLeadPost(map[string]any{
		"source":  "example.test / обратный звонок",
		"note":    "Имя: Ivan",
		"u_name":  "Ivan",
		"u_phone": "+79261234567",
		"extended_fields": map[string]any{
			"193320": "Камчатка",
			"192647": "",
		},
	})
	extended, ok := post["extended_fields"].(map[string]string)
	if !ok || extended["193320"] != "Камчатка" {
		t.Fatalf("unexpected extended_fields: %#v", post["extended_fields"])
	}
	if _, exists := extended["192647"]; exists {
		t.Fatalf("empty extended field must be omitted")
	}
	body := encodeLeadPost(post)
	if !strings.Contains(body, "u_phone=%2B79261234567") || !strings.Contains(body, "extended_fields%5B193320%5D=") {
		t.Fatalf("unexpected encoded body: %q", body)
	}
}

func TestFlattenLeadPostFioFields(t *testing.T) {
	post := flattenLeadPost(map[string]any{
		"source":    "example.test / заявка на тур",
		"note":      "Заявка",
		"u_surname": "Петров",
		"u_name":    "Иван",
		"u_sname":   "Сергеевич",
		"u_phone":   "+79261234567",
	})
	if post["u_surname"] != "Петров" || post["u_name"] != "Иван" || post["u_sname"] != "Сергеевич" {
		t.Fatalf("unexpected fio fields: %#v", post)
	}
	body := encodeLeadPost(post)
	for _, part := range []string{"u_surname=%D0%9F%D0%B5%D1%82%D1%80%D0%BE%D0%B2", "u_name=%D0%98%D0%B2%D0%B0%D0%BD", "u_sname="} {
		if !strings.Contains(body, part) {
			t.Fatalf("missing %q in body: %q", part, body)
		}
	}
}

func TestHasContact(t *testing.T) {
	if !HasContact(map[string]any{"u_phone": "+7"}) {
		t.Fatal("expected phone contact")
	}
	if HasContact(map[string]any{"u_email": "bad"}) {
		t.Fatal("invalid email must not count as contact")
	}
}

func TestManagerDistributionWarnings(t *testing.T) {
	warnings := ManagerDistributionWarnings(map[string]any{
		"comment": "No active (and not busy) managers for lead distribution (key: )",
	})
	if len(warnings) != 1 {
		t.Fatalf("expected manager warning, got %#v", warnings)
	}
}

func TestSanitizeResponseKeepsLeadID(t *testing.T) {
	out := SanitizeResponse(200, []byte(`{"result":200,"id":"13111"}`))
	if out["id"] != "13111" {
		t.Fatalf("expected U-ON id in sanitized response, got %#v", out)
	}
}
