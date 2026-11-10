package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	clientsrepo "riders.agency/pulse/api/internal/modules/clients/repository"
	"riders.agency/pulse/api/internal/platform/httptransport"
)

var ErrTransactionBoundaryRequired = errors.New("transaction boundary is required")

type DaDataRepository interface {
	GetDaDataIntegration(context.Context, uuid.UUID, uuid.UUID) (*clientsrepo.DaDataIntegrationRow, error)
	GetDaDataProjectIntegration(context.Context, uuid.UUID, uuid.UUID) (*clientsrepo.DaDataIntegrationRow, error)
	DaDataCompanyLookupConfig(context.Context, uuid.UUID, uuid.UUID) (clientsrepo.DaDataLookupConfig, error)
	DaDataProjectLookupConfig(context.Context, uuid.UUID, uuid.UUID) (clientsrepo.DaDataLookupConfig, error)
	SaveDaDataIntegrationTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, clientsrepo.DaDataIntegrationInput) error
	SaveDaDataProjectIntegrationTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, clientsrepo.DaDataIntegrationInput) error
}

type Transactor interface {
	InTx(context.Context, func(pgx.Tx) error) error
}

type DaData struct {
	repo       DaDataRepository
	httpClient *http.Client
	tx         Transactor
}

func NewDaData(repo DaDataRepository, httpClient *http.Client, tx Transactor) *DaData {
	return &DaData{repo: repo, httpClient: httpClient, tx: tx}
}

func (d *DaData) SaveCompanyIntegration(ctx context.Context, companyID, userID uuid.UUID, in clientsrepo.DaDataIntegrationInput) (*clientsrepo.DaDataIntegrationRow, error) {
	if d.tx == nil {
		return nil, ErrTransactionBoundaryRequired
	}
	if err := d.tx.InTx(ctx, func(tx pgx.Tx) error {
		return d.repo.SaveDaDataIntegrationTx(ctx, tx, companyID, userID, in)
	}); err != nil {
		return nil, err
	}
	return d.repo.GetDaDataIntegration(ctx, companyID, userID)
}

func (d *DaData) SaveProjectIntegration(ctx context.Context, projectID, userID uuid.UUID, in clientsrepo.DaDataIntegrationInput) (*clientsrepo.DaDataIntegrationRow, error) {
	if d.tx == nil {
		return nil, ErrTransactionBoundaryRequired
	}
	if err := d.tx.InTx(ctx, func(tx pgx.Tx) error {
		return d.repo.SaveDaDataProjectIntegrationTx(ctx, tx, projectID, userID, in)
	}); err != nil {
		return nil, err
	}
	return d.repo.GetDaDataProjectIntegration(ctx, projectID, userID)
}

func (d *DaData) FindParty(ctx context.Context, companyID, userID uuid.UUID, query string) (map[string]any, error) {
	cfg, err := d.repo.DaDataCompanyLookupConfig(ctx, companyID, userID)
	if err != nil {
		return nil, err
	}
	return d.findParty(ctx, cfg, query)
}

func (d *DaData) FindPartyByProject(ctx context.Context, projectID, userID uuid.UUID, query string) (map[string]any, error) {
	cfg, err := d.repo.DaDataProjectLookupConfig(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	return d.findParty(ctx, cfg, query)
}

func (d *DaData) findParty(ctx context.Context, cfg clientsrepo.DaDataLookupConfig, query string) (map[string]any, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://suggestions.dadata.ru/suggestions/api/4_1/rs"
	}
	body, _ := json.Marshal(map[string]string{"query": query})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/findById/party", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token "+strings.TrimSpace(cfg.APIKey))
	res, err := d.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var raw map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("dadata status %d", res.StatusCode)
	}
	return map[string]any{"suggestions": clientsrepo.MapDaDataSuggestions(raw), "raw": raw}, nil
}

func (d *DaData) client() *http.Client {
	if d.httpClient != nil {
		return d.httpClient
	}
	return httptransport.NewDefaultClient(30 * time.Second)
}
