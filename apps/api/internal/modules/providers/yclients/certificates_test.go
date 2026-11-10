package yclients

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCertificateClientIssuesAndReadsBackOneInstance(t *testing.T) {
	var mu sync.Mutex
	requests := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer partner-token, User user-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/vnd.api.v2+json" {
			t.Errorf("Accept = %q", got)
		}
		mu.Lock()
		requests[r.URL.Path]++
		mu.Unlock()
		switch r.URL.Path {
		case "/chain/1972020/loyalty/certificate_types":
			writeJSON(t, w, map[string]any{"success": true, "data": []map[string]any{{
				"id": 858192, "title": "Сайт произвольный номинал - 12000", "balance": 12000,
				"is_multi": true, "item_type_id": 1, "is_allow_empty_code": false,
				"balance_edit_type_id": 1, "service_ids": []int64{}, "salon_ids": []int64{1943196, 1943547},
				"expiration_type_id": 2, "expiration_timeout": 12, "expiration_timeout_unit_id": 3,
			}}})
		case "/goods/2099931":
			writeJSON(t, w, map[string]any{"success": true, "data": []map[string]any{{
				"id": "70001", "loyalty_certificate_type_id": "858192",
			}}})
		case "/loyalty/generate_code/2099931/70001":
			writeJSON(t, w, map[string]any{"success": true, "data": map[string]any{"code": 99112233}})
		case "/storage_operations/operation/2099931":
			if r.Method != http.MethodPost {
				t.Errorf("sale method = %s", r.Method)
			}
			var payload struct {
				TypeID            int `json:"type_id"`
				StorageID         int `json:"storage_id"`
				GoodsTransactions []struct {
					GoodID            int64  `json:"good_id"`
					Amount            int    `json:"amount"`
					Cost              int    `json:"cost"`
					GoodSpecialNumber string `json:"good_special_number"`
				} `json:"goods_transactions"`
			}
			decodeJSON(t, r.Body, &payload)
			if payload.TypeID != 1 || payload.StorageID != 300 || len(payload.GoodsTransactions) != 1 {
				t.Fatalf("unexpected sale payload: %+v", payload)
			}
			line := payload.GoodsTransactions[0]
			if line.GoodID != 70001 || line.Amount != 1 || line.Cost != 12000 || line.GoodSpecialNumber != "99112233" {
				t.Fatalf("unexpected sale line: %+v", line)
			}
			writeJSON(t, w, map[string]any{"success": true, "data": map[string]any{
				"document":     map[string]any{"id": 88001},
				"transactions": []map[string]any{{"id": 99001, "document_id": 88001, "good_id": 70001}},
			}})
		case "/storage_operations/goods_transactions/2099931/99001":
			writeJSON(t, w, map[string]any{"success": true, "data": map[string]any{
				"id": 99001, "document_id": 88001, "good_id": 70001, "amount": -1,
				"cost_per_unit": 12000, "cost": 12000, "deleted": false, "good_special_number": "99112233",
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestCertificateClient(t, server.URL, "write", false)
	input := CertificateIssueInput{OperationKey: "yclients:issue:instance-1", ProductKind: "nominal", AmountRub: 12000}
	result, err := client.IssueCertificate(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Code != "99112233" || result.CertificateTypeID != 858192 || result.ProductID != 70001 || result.DocumentID != 88001 || result.TransactionID != 99001 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if _, err := client.IssueCertificate(context.Background(), CertificateIssueInput{OperationKey: "yclients:issue:instance-2", ProductKind: "nominal", AmountRub: 12000}); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	typeCalls := requests["/chain/1972020/loyalty/certificate_types"]
	goodsCalls := requests["/goods/2099931"]
	mu.Unlock()
	if typeCalls != 1 || goodsCalls != 1 {
		t.Fatalf("resolution cache was not used: type=%d goods=%d", typeCalls, goodsCalls)
	}
}

func TestSelectCertificateTypeUsesStructuredServiceFingerprint(t *testing.T) {
	input := CertificateIssueInput{
		ProductKind: "service", ProductTitle: "Тайский массаж в пижаме - 60 минут",
		AmountRub: 4500, ServiceIDs: []int64{30, 20},
	}
	types := []certificateType{
		{ID: 1, Title: "Сайт Тайский массаж в пижаме 60 минут - 4500", Balance: 4500, ItemTypeID: 3, ServiceIDs: []int64{20}, SalonIDs: []int64{1, 2}},
		{ID: 2, Title: "Сайт Прежнее название услуги 60 минут - 4500", Balance: 4500, ItemTypeID: 3, ServiceIDs: []int64{20, 30}, SalonIDs: []int64{2, 1}},
		{ID: 3, Title: "Сайт Тайский массаж в пижаме 60 минут - 4500", Balance: 4500, ItemTypeID: 3, IsAllowEmptyCode: true, ServiceIDs: []int64{20, 30}, SalonIDs: []int64{1, 2}},
	}
	got, found, err := selectCertificateType(types, input, []int64{1, 2})
	if err != nil || !found || got.ID != 2 {
		t.Fatalf("selectCertificateType() = %+v, %v, %v", got, found, err)
	}
}

func TestSelectCertificateTypeRejectsDuplicateServiceMappingsAcrossRenames(t *testing.T) {
	input := CertificateIssueInput{
		ProductKind: "service", ProductTitle: "Новое название - 60 минут",
		AmountRub: 4500, ServiceIDs: []int64{20},
	}
	types := []certificateType{
		{ID: 1, Title: "Сайт Старое название 60 минут - 4500", Balance: 4500, ItemTypeID: 3, ServiceIDs: []int64{20}, SalonIDs: []int64{1}},
		{ID: 2, Title: "Сайт Новое название 60 минут - 4500", Balance: 4500, ItemTypeID: 3, ServiceIDs: []int64{20}, SalonIDs: []int64{1}},
	}
	if _, _, err := selectCertificateType(types, input, []int64{1}); !errors.Is(err, ErrCertificateMapping) {
		t.Fatalf("selectCertificateType() error = %v, want ErrCertificateMapping", err)
	}
}

func TestSelectCertificateTypeRejectsDuplicateExactMappings(t *testing.T) {
	input := CertificateIssueInput{ProductKind: "nominal", AmountRub: 5000}
	typeRecord := certificateType{
		Title: "Сайт произвольный номинал - 5000", Balance: 5000, ItemTypeID: 1, SalonIDs: []int64{1},
	}
	one, two := typeRecord, typeRecord
	one.ID, two.ID = 1, 2
	_, _, err := selectCertificateType([]certificateType{one, two}, input, []int64{1})
	if !errors.Is(err, ErrCertificateMapping) {
		t.Fatalf("error = %v", err)
	}
}

func TestCertificateClientCreatesMissingNominalTypeFromApprovedPolicy(t *testing.T) {
	created := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chain/1972020/loyalty/certificate_types":
			if r.Method == http.MethodPost {
				var payload map[string]any
				decodeJSON(t, r.Body, &payload)
				if payload["title"] != "Сайт произвольный номинал - 12500" || payload["balance"] != float64(12500) {
					t.Fatalf("unexpected type payload: %+v", payload)
				}
				created = true
				writeJSON(t, w, map[string]any{"success": true, "data": map[string]any{
					"id": 501, "title": payload["title"], "balance": 12500, "is_multi": true,
					"item_type_id": 1, "is_allow_empty_code": false, "balance_edit_type_id": 2,
					"service_ids": []int64{}, "salon_ids": []int64{1943196, 1943547},
				}})
				return
			}
			writeJSON(t, w, map[string]any{"success": true, "data": []map[string]any{{
				"id": 500, "title": "Сайт произвольный номинал - 5000", "balance": 5000,
				"is_multi": true, "item_type_id": 1, "is_allow_empty_code": false, "balance_edit_type_id": 2,
				"service_ids": []int64{}, "salon_ids": []int64{1943196, 1943547},
				"expiration_type_id": 2, "expiration_timeout": 1, "expiration_timeout_unit_id": 4,
			}}})
		case "/goods/2099931":
			writeJSON(t, w, map[string]any{"success": true, "data": []map[string]any{{"id": 701, "loyalty_certificate_type_id": 501}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestCertificateClient(t, server.URL, "verify", true)
	_, err := client.IssueCertificate(context.Background(), CertificateIssueInput{
		OperationKey: "yclients:issue:instance", ProductKind: "nominal", AmountRub: 12500,
	})
	if !errors.Is(err, ErrCertificateWriteDisabled) || !created {
		t.Fatalf("error = %v, created = %v", err, created)
	}
}

func TestCertificateSaleServerErrorIsPossibleEffect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chain/1972020/loyalty/certificate_types":
			writeJSON(t, w, map[string]any{"success": true, "data": []map[string]any{{
				"id": 1, "title": "Сайт произвольный номинал - 12000", "balance": 12000,
				"is_multi": true, "item_type_id": 1, "is_allow_empty_code": false,
				"service_ids": []int64{}, "salon_ids": []int64{1943196, 1943547},
			}}})
		case "/goods/2099931":
			writeJSON(t, w, map[string]any{"success": true, "data": []map[string]any{{"id": 2, "loyalty_certificate_type_id": 1}}})
		case "/loyalty/generate_code/2099931/2":
			writeJSON(t, w, map[string]any{"success": true, "data": map[string]any{"code": "code"}})
		case "/storage_operations/operation/2099931":
			http.Error(w, `{"success":false,"meta":{"message":"temporary"}}`, http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestCertificateClient(t, server.URL, "write", false)
	_, err := client.IssueCertificate(context.Background(), CertificateIssueInput{
		OperationKey: "yclients:issue:instance", ProductKind: "nominal", AmountRub: 12000,
	})
	if err == nil || !CertificatePossibleEffect(err) || CertificateErrorCode(err) != "http_500" {
		t.Fatalf("error = %v, possible = %v, code = %s", err, CertificatePossibleEffect(err), CertificateErrorCode(err))
	}
}

func TestCertificateReadBackMismatchIsPossibleEffect(t *testing.T) {
	transaction := certificateTransaction{
		ID: 1, DocumentID: 2, GoodID: 3, Amount: -1, CostPerUnit: 5000, Cost: 5000,
		GoodSpecialNumber: "different",
	}
	if err := verifyCertificateTransaction(transaction, 2, 3, 5000, "expected"); err == nil {
		t.Fatal("code mismatch must fail")
	}
}

func TestCertificateTypeTitleIsCanonical(t *testing.T) {
	tests := []struct {
		input CertificateIssueInput
		want  string
	}{
		{input: CertificateIssueInput{ProductKind: "nominal", AmountRub: 5000}, want: "Сайт произвольный номинал - 5000"},
		{input: CertificateIssueInput{ProductKind: "service", ProductTitle: "Тайский массаж - 60 минут", AmountRub: 4500}, want: "Сайт Тайский массаж 60 минут - 4500"},
		{input: CertificateIssueInput{ProductKind: "service", ProductTitle: "Сайт Тайский массаж 60 минут - 4500", AmountRub: 4500}, want: "Сайт Тайский массаж 60 минут - 4500"},
		{input: CertificateIssueInput{ProductKind: "service", ProductTitle: "SPA-ритуал - восстановление - 90 минут", AmountRub: 7200}, want: "Сайт SPA-ритуал - восстановление 90 минут - 7200"},
	}
	for _, test := range tests {
		if got := certificateTypeTitle(test.input); got != test.want {
			t.Fatalf("certificateTypeTitle() = %q, want %q", got, test.want)
		}
	}
}

func newTestCertificateClient(t *testing.T, baseURL, mode string, allowCreate bool) *CertificateClient {
	t.Helper()
	client, err := NewCertificateClient(CertificateConfig{
		Mode: mode, BaseURL: baseURL, PartnerToken: "partner-token", UserToken: "user-token",
		ChainID: 1972020, LocationID: 2099931, StorageID: 300,
		SalonIDs: []int64{1943196, 1943547}, AllowTypeCreate: allowCreate,
		Timeout: 2 * time.Second, ResolutionTTL: time.Minute,
		Now: func() time.Time { return time.Date(2026, 8, 5, 12, 0, 0, 0, time.FixedZone("MSK", 3*60*60)) },
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}

func decodeJSON(t *testing.T, body io.Reader, target any) {
	t.Helper()
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(target); err != nil {
		t.Fatal(err)
	}
}

func TestCertificateAuthorizationRejectsHeaderInjection(t *testing.T) {
	if got := certificateAuthorization("partner\nInjected: yes", "user"); got != "" {
		t.Fatalf("authorization = %q", got)
	}
	if got := certificateAuthorization("Bearer partner", "User user"); got != "Bearer partner, User user" {
		t.Fatalf("authorization = %q", got)
	}
}

func TestCertificateClientAuthenticatesFromRuntimeCredentialsOnce(t *testing.T) {
	var authCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth" {
			authCalls++
			if got := r.Header.Get("Authorization"); got != "Bearer partner-token" {
				t.Fatalf("auth Authorization = %q", got)
			}
			var payload map[string]string
			decodeJSON(t, r.Body, &payload)
			if payload["login"] != "operator@example.test" || payload["password"] != "runtime-password" {
				t.Fatalf("unexpected auth payload")
			}
			w.WriteHeader(http.StatusCreated)
			writeJSON(t, w, map[string]any{"success": true, "data": map[string]any{"user_token": "fresh-user-token"}})
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer partner-token, User fresh-user-token" {
			t.Fatalf("business Authorization = %q", got)
		}
		writeJSON(t, w, []any{})
	}))
	defer server.Close()

	client, err := NewCertificateClient(CertificateConfig{
		Mode: "verify", BaseURL: server.URL, PartnerToken: "partner-token",
		Login: "operator@example.test", Password: "runtime-password",
		ChainID: 1972020, LocationID: 2099931, StorageID: 300,
		SalonIDs: []int64{1943196}, HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 2; index++ {
		var target []any
		if err := client.requestJSON(context.Background(), http.MethodGet, "/probe", nil, &target, false); err != nil {
			t.Fatal(err)
		}
	}
	if authCalls != 1 {
		t.Fatalf("auth calls = %d, want 1", authCalls)
	}
}

func TestCertificateAPIMessageIsBounded(t *testing.T) {
	raw := []byte(`{"meta":{"message":"` + strings.Repeat("x", 300) + `"}}`)
	if got := certificateAPIMessage(raw); len(got) != 180 {
		t.Fatalf("message length = %d", len(got))
	}
}

func TestCertificateClientFallsBackWhenConfiguredUserTokenIsRejected(t *testing.T) {
	var authCalls int
	var probeCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			authCalls++
			if got := r.Header.Get("Authorization"); got != "Bearer partner-token" {
				t.Fatalf("auth authorization = %q", got)
			}
			writeJSON(t, w, map[string]any{"success": true, "data": map[string]any{"user_token": "fallback-token"}})
		case "/probe":
			probeCalls++
			if r.Header.Get("Authorization") == "Bearer partner-token, User stale-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer partner-token, User fallback-token" {
				t.Fatalf("fallback authorization = %q", got)
			}
			writeJSON(t, w, []any{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewCertificateClient(CertificateConfig{
		Mode: "write", BaseURL: server.URL, PartnerToken: "partner-token", UserToken: "stale-token",
		Login: "operator@example.test", Password: "fallback-password",
		ChainID: 1972020, LocationID: 1, StorageID: 2, SalonIDs: []int64{3}, HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	var target []any
	if err := client.requestJSON(context.Background(), http.MethodGet, "/probe", nil, &target, false); err != nil {
		t.Fatal(err)
	}
	if authCalls != 1 || probeCalls != 2 || client.AuthSource() != "login_fallback" {
		t.Fatalf("auth=%d probe=%d source=%q", authCalls, probeCalls, client.AuthSource())
	}
}
