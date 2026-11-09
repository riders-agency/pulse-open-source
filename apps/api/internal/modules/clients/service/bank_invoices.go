package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"riders.agency/pulse/api/internal/modules/clients/invoicing"
)

type BankInvoiceStore interface {
	ReadBankInvoice(context.Context, uuid.UUID, uuid.UUID) (*invoicing.Workspace, error)
	SaveBankInvoice(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, invoicing.Input) (*invoicing.Draft, error)
}

type BankInvoices struct {
	store     BankInvoiceStore
	providers invoicing.Registry
}

func NewBankInvoices(store BankInvoiceStore, providers invoicing.Registry) *BankInvoices {
	return &BankInvoices{store: store, providers: providers}
}

func (s *BankInvoices) Read(ctx context.Context, actionID, userID uuid.UUID) (*invoicing.Workspace, error) {
	result, err := s.store.ReadBankInvoice(ctx, actionID, userID)
	if err != nil {
		return nil, err
	}
	result.Providers = s.providers.Capabilities()
	return result, nil
}

func (s *BankInvoices) Save(ctx context.Context, actionID, userID, key uuid.UUID, in invoicing.Input) (*invoicing.Draft, error) {
	if key == uuid.Nil {
		return nil, invoicing.ErrInvalid
	}
	in, err := invoicing.Normalize(in)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	return s.store.SaveBankInvoice(ctx, actionID, userID, key, hex.EncodeToString(hash[:]), in)
}
