package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	clientsrepo "riders.agency/pulse/api/internal/modules/clients/repository"
	"riders.agency/pulse/api/internal/platform/errs"
	"riders.agency/pulse/api/internal/platform/problem"
)

func writeRowsOrError[T any](w http.ResponseWriter, rows []T, err error) {
	if writeClientError(w, err) {
		return
	}
	problem.JSON(w, http.StatusOK, rows)
}

func writeClientError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, errs.ErrForbidden):
		problem.WriteModule(w, http.StatusForbidden, "clients", "forbidden", "Нет доступа к клиентам", "", "Проверьте доступ к компании или проекту.")
	case errors.Is(err, clientsrepo.ErrInvalidIdentifiers):
		problem.WriteModule(w, http.StatusBadRequest, "clients", "identifiers", "Некорректные реквизиты клиента", err.Error(), "Проверьте ИНН, КПП, ОГРН или ОГРНИП.")
	case errors.Is(err, clientsrepo.ErrDaDataNotConfigured):
		problem.WriteModule(w, http.StatusBadRequest, "clients", "dadata_not_configured", "DaData не подключена", err.Error(), "Подключите DaData в компании и повторите поиск.")
	case errors.Is(err, clientsrepo.ErrPaymentClosed):
		problem.WriteModule(w, http.StatusConflict, "clients", "payment_closed", "Оплата уже закрыта", err.Error(), "Обновите карточку оплаты и проверьте текущий статус.")
	case errors.Is(err, clientsrepo.ErrPaymentNoService):
		problem.WriteModule(w, http.StatusBadRequest, "clients", "payment_service_required", "Нужна услуга для внесения оплаты", err.Error(), "Сначала выберите или добавьте услугу клиента.")
	case errors.Is(err, clientsrepo.ErrIdempotencyRequired):
		problem.WriteModule(w, http.StatusBadRequest, "clients", "idempotency_required", "Нужен ключ идемпотентности", err.Error(), "Повторите действие через форму внесения оплаты или передайте Idempotency-Key.")
	case errors.Is(err, pgx.ErrNoRows):
		problem.WriteModule(w, http.StatusNotFound, "clients", "not_found", "Клиентский ресурс не найден", "", "Проверьте выбранного клиента, услугу или оплату.")
	default:
		problem.WriteModule(w, http.StatusBadRequest, "clients", "request_failed", "Операция с клиентами не выполнена", err.Error(), "Проверьте данные и повторите действие.")
	}
	return true
}

func clientContactIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	clientID, err := uuid.Parse(chi.URLParam(r, "client_id"))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "client_id", "")
		return uuid.Nil, uuid.Nil, false
	}
	contactID, err := uuid.Parse(chi.URLParam(r, "client_contact_id"))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "client_contact_id", "")
		return uuid.Nil, uuid.Nil, false
	}
	return clientID, contactID, true
}

func clientProjectIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	clientID, err := uuid.Parse(chi.URLParam(r, "client_id"))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "client_id", "")
		return uuid.Nil, uuid.Nil, false
	}
	projectID, err := uuid.Parse(chi.URLParam(r, "project_id"))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "project_id", "")
		return uuid.Nil, uuid.Nil, false
	}
	return clientID, projectID, true
}

func clientResponsibleIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	clientID, err := uuid.Parse(chi.URLParam(r, "client_id"))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "client_id", "")
		return uuid.Nil, uuid.Nil, false
	}
	responsibleID, err := uuid.Parse(chi.URLParam(r, "client_responsible_id"))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "client_responsible_id", "")
		return uuid.Nil, uuid.Nil, false
	}
	return clientID, responsibleID, true
}
