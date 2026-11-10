package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/modules/clients/invoicing"
	"riders.agency/pulse/api/internal/modules/clients/service"
	"riders.agency/pulse/api/internal/platform/authz"
)

type forbiddenDraftStore struct{ calls int }

func (s *forbiddenDraftStore) ReadBankInvoice(context.Context, uuid.UUID, uuid.UUID) (*invoicing.Workspace, error) {
	s.calls++
	return nil, nil
}
func (s *forbiddenDraftStore) SaveBankInvoice(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, invoicing.Input) (*invoicing.Draft, error) {
	s.calls++
	return nil, nil
}

func TestBankInvoiceRejectsUntrustedFieldsBeforePersistence(t *testing.T) {
	store := &forbiddenDraftStore{}
	router := chi.NewRouter()
	MountBankInvoiceRoutes(router, service.NewBankInvoices(store, invoicing.NewRegistry()))
	valid := `{"revision":0,"kind":"bank_invoice","payer_name":"Demo","purpose":"Service","tax_id":"","due_date":""}`
	for _, body := range []string{strings.TrimSuffix(valid, "}") + `,"amount_minor":"1"}`, strings.TrimSuffix(valid, "}") + `,"status":"paid"}`, valid + ` {}`, strings.Repeat(" ", 17000) + valid} {
		request := httptest.NewRequest(http.MethodPut, "/payment-actions/"+uuid.NewString()+"/bank-invoice", strings.NewReader(body))
		request.Header.Set("Idempotency-Key", uuid.NewString())
		request = request.WithContext(authz.WithSessionClaims(request.Context(), authz.SessionClaims{UserID: uuid.New()}))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != 400 {
			t.Fatalf("untrusted payload status: %d", response.Code)
		}
	}
	if store.calls != 0 {
		t.Fatalf("unexpected persistence calls: %d", store.calls)
	}
}
