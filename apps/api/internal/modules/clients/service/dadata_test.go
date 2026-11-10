package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	clientsrepo "riders.agency/pulse/api/internal/modules/clients/repository"
)

type fakeDaDataRepo struct {
	cfg       clientsrepo.DaDataLookupConfig
	saved     bool
	savedProj bool
}

func (f fakeDaDataRepo) GetDaDataIntegration(context.Context, uuid.UUID, uuid.UUID) (*clientsrepo.DaDataIntegrationRow, error) {
	return &clientsrepo.DaDataIntegrationRow{Configured: f.saved, Source: "company"}, nil
}

func (f fakeDaDataRepo) GetDaDataProjectIntegration(context.Context, uuid.UUID, uuid.UUID) (*clientsrepo.DaDataIntegrationRow, error) {
	return &clientsrepo.DaDataIntegrationRow{Configured: f.savedProj, Source: "project"}, nil
}

func (f fakeDaDataRepo) DaDataCompanyLookupConfig(context.Context, uuid.UUID, uuid.UUID) (clientsrepo.DaDataLookupConfig, error) {
	return f.cfg, nil
}

func (f fakeDaDataRepo) DaDataProjectLookupConfig(context.Context, uuid.UUID, uuid.UUID) (clientsrepo.DaDataLookupConfig, error) {
	return f.cfg, nil
}

func (f fakeDaDataRepo) SaveDaDataIntegrationTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, clientsrepo.DaDataIntegrationInput) error {
	return nil
}

func (f fakeDaDataRepo) SaveDaDataProjectIntegrationTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, clientsrepo.DaDataIntegrationInput) error {
	return nil
}

type recordingTransactor struct {
	called bool
}

func (t *recordingTransactor) InTx(_ context.Context, fn func(pgx.Tx) error) error {
	t.called = true
	return fn(nil)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestDaDataFindPartyUsesServiceHTTPClient(t *testing.T) {
	var sawPath, sawAuth string
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		sawPath = req.URL.Path
		sawAuth = req.Header.Get("Authorization")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"suggestions":[{"value":"ООО Ромашка","data":{"type":"LEGAL","inn":"7707083893"}}]}`)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})}
	svc := NewDaData(fakeDaDataRepo{cfg: clientsrepo.DaDataLookupConfig{APIKey: "secret", BaseURL: "https://dadata.test/root/"}}, client, nil)
	got, err := svc.FindParty(context.Background(), uuid.New(), uuid.New(), " 7707083893 ")
	if err != nil {
		t.Fatalf("FindParty returned error: %v", err)
	}
	if sawPath != "/root/findById/party" || sawAuth != "Token secret" {
		t.Fatalf("unexpected request path/auth: %q %q", sawPath, sawAuth)
	}
	if len(got["suggestions"].([]map[string]any)) != 1 {
		t.Fatalf("unexpected suggestions: %#v", got["suggestions"])
	}
}

func TestDaDataFindPartyRejectsEmptyQuery(t *testing.T) {
	svc := NewDaData(fakeDaDataRepo{cfg: clientsrepo.DaDataLookupConfig{APIKey: "secret"}}, nil, nil)
	if _, err := svc.FindParty(context.Background(), uuid.New(), uuid.New(), " "); err == nil {
		t.Fatalf("expected empty query error")
	}
}

func TestDaDataSaveCompanyIntegrationUsesTransactionBoundary(t *testing.T) {
	tx := &recordingTransactor{}
	svc := NewDaData(fakeDaDataRepo{saved: true}, nil, tx)
	row, err := svc.SaveCompanyIntegration(context.Background(), uuid.New(), uuid.New(), clientsrepo.DaDataIntegrationInput{APIKey: "secret"})
	if err != nil {
		t.Fatalf("SaveCompanyIntegration returned error: %v", err)
	}
	if !tx.called || row == nil || !row.Configured {
		t.Fatalf("transaction not used or row not returned: called=%v row=%#v", tx.called, row)
	}
}
