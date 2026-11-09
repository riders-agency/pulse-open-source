package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/modules/clients/invoicing"
	"riders.agency/pulse/api/internal/modules/clients/service"
	"riders.agency/pulse/api/internal/platform/errs"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

func MountBankInvoiceRoutes(r chi.Router, invoices *service.BankInvoices) {
	r.Get("/payment-actions/{payment_action_id}/bank-invoice", readBankInvoice(invoices))
	r.Put("/payment-actions/{payment_action_id}/bank-invoice", saveBankInvoice(invoices))
}

func readBankInvoice(invoices *service.BankInvoices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionID, userID, ok := bankInvoiceIDs(w, r)
		if !ok {
			return
		}
		result, err := invoices.Read(r.Context(), actionID, userID)
		if bankInvoiceError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, result)
	}
}

func saveBankInvoice(invoices *service.BankInvoices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionID, userID, ok := bankInvoiceIDs(w, r)
		if !ok {
			return
		}
		key, err := uuid.Parse(r.Header.Get("Idempotency-Key"))
		var input invoicing.Input
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
		decoder.DisallowUnknownFields()
		if err != nil || decoder.Decode(&input) != nil || decoder.Decode(&struct{}{}) != io.EOF {
			bankInvoiceError(w, invoicing.ErrInvalid)
			return
		}
		result, err := invoices.Save(r.Context(), actionID, userID, key, input)
		if bankInvoiceError(w, err) {
			return
		}
		problem.JSON(w, http.StatusOK, result)
	}
}

func bankInvoiceIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	claims, ok := mware.SessionClaimsFrom(r.Context())
	if !ok {
		problem.Write(w, http.StatusUnauthorized, "Требуется вход", "")
		return uuid.Nil, uuid.Nil, false
	}
	id, err := uuid.Parse(chi.URLParam(r, "payment_action_id"))
	if err != nil {
		bankInvoiceError(w, invoicing.ErrInvalid)
		return uuid.Nil, uuid.Nil, false
	}
	return id, claims.UserID, true
}

func bankInvoiceError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, invoicing.ErrInvalid):
		problem.Write(w, http.StatusBadRequest, "Проверьте данные черновика", "Нужны плательщик, назначение и корректный срок оплаты.")
	case errors.Is(err, invoicing.ErrConflict):
		problem.Write(w, http.StatusConflict, "Черновик изменён или оплата закрыта", "Обновите карточку перед сохранением.")
	case errors.Is(err, errs.ErrForbidden):
		problem.Write(w, http.StatusForbidden, "Нет доступа к оплате", "")
	default:
		problem.Write(w, http.StatusInternalServerError, "Не удалось сохранить или загрузить черновик", "Повторите попытку позже.")
	}
	return true
}
