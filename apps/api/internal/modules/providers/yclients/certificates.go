package yclients

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrCertificateIntegrationDisabled = errors.New("YCLIENTS certificate integration is disabled")
	ErrCertificateWriteDisabled       = errors.New("YCLIENTS certificate writes are disabled")
	ErrCertificateMapping             = errors.New("YCLIENTS certificate mapping is invalid")
)

type CertificateIssuer interface {
	IssueCertificate(context.Context, CertificateIssueInput) (CertificateIssueResult, error)
}

type CertificateConfig struct {
	Mode            string
	BaseURL         string
	PartnerToken    string
	UserToken       string
	Login           string
	Password        string
	ChainID         int64
	LocationID      int64
	StorageID       int64
	SalonIDs        []int64
	AllowTypeCreate bool
	Timeout         time.Duration
	ResolutionTTL   time.Duration
	HTTPClient      *http.Client
	Now             func() time.Time
}

type CertificateIssueInput struct {
	OperationKey string
	ProductKind  string
	ProductTitle string
	AmountRub    int
	ServiceIDs   []int64
}

type CertificateIssueResult struct {
	Code              string
	CertificateTypeID int64
	ProductID         int64
	LocationID        int64
	DocumentID        int64
	TransactionID     int64
}

type CertificateAPIError struct {
	Code           string
	StatusCode     int
	PossibleEffect bool
	Retryable      bool
	message        string
}

func (e *CertificateAPIError) Error() string {
	if e == nil {
		return "YCLIENTS certificate API error"
	}
	message := strings.TrimSpace(e.message)
	if message == "" {
		message = "request failed"
	}
	return fmt.Sprintf("YCLIENTS certificate API %s: %s", e.Code, message)
}

func CertificateErrorCode(err error) string {
	var apiErr *CertificateAPIError
	if errors.As(err, &apiErr) && apiErr.Code != "" {
		return apiErr.Code
	}
	switch {
	case errors.Is(err, ErrCertificateIntegrationDisabled):
		return "integration_disabled"
	case errors.Is(err, ErrCertificateWriteDisabled):
		return "write_disabled"
	case errors.Is(err, ErrCertificateMapping):
		return "mapping_invalid"
	default:
		return "yclients_error"
	}
}

func CertificatePossibleEffect(err error) bool {
	var apiErr *CertificateAPIError
	return errors.As(err, &apiErr) && apiErr.PossibleEffect
}

func CertificateRetryable(err error) bool {
	var apiErr *CertificateAPIError
	return errors.As(err, &apiErr) && apiErr.Retryable && !apiErr.PossibleEffect
}

type CertificateClient struct {
	mode            string
	baseURL         string
	authorization   string
	partnerToken    string
	login           string
	password        string
	chainID         int64
	locationID      int64
	storageID       int64
	salonIDs        []int64
	allowTypeCreate bool
	client          *http.Client
	now             func() time.Time
	resolutionTTL   time.Duration
	mu              sync.Mutex
	authMu          sync.Mutex
	authSource      string
	resolved        map[string]cachedCertificateProduct
}

type cachedCertificateProduct struct {
	value     resolvedCertificateProduct
	expiresAt time.Time
}

type resolvedCertificateProduct struct {
	CertificateTypeID int64
	ProductID         int64
}

func NewCertificateClient(config CertificateConfig) (*CertificateClient, error) {
	mode := strings.ToLower(strings.TrimSpace(config.Mode))
	if mode == "" {
		mode = "disabled"
	}
	if mode != "disabled" && mode != "verify" && mode != "write" {
		return nil, fmt.Errorf("%w: unsupported mode", ErrCertificateMapping)
	}
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.yclients.com/api/v1"
	}
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	ttl := config.ResolutionTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	result := &CertificateClient{
		mode: mode, baseURL: baseURL,
		authorization: certificateAuthorization(config.PartnerToken, config.UserToken),
		partnerToken:  strings.TrimSpace(strings.TrimPrefix(config.PartnerToken, "Bearer ")),
		login:         strings.TrimSpace(config.Login),
		password:      config.Password,
		chainID:       config.ChainID, locationID: config.LocationID, storageID: config.StorageID,
		salonIDs: sortedUniqueInt64(config.SalonIDs), allowTypeCreate: config.AllowTypeCreate,
		client: client, now: now, resolutionTTL: ttl,
		resolved: make(map[string]cachedCertificateProduct),
	}
	if result.authorization != "" {
		result.authSource = "user_token"
	}
	if mode != "disabled" {
		hasCredentialLogin := result.partnerToken != "" && result.login != "" && result.password != ""
		if (result.authorization == "" && !hasCredentialLogin) || result.chainID <= 0 || result.locationID <= 0 || result.storageID <= 0 || len(result.salonIDs) == 0 {
			return nil, fmt.Errorf("%w: credentials, chain, location, storage and salons are required", ErrCertificateMapping)
		}
	}
	return result, nil
}

func (c *CertificateClient) AuthSource() string {
	if c == nil {
		return ""
	}
	c.authMu.Lock()
	defer c.authMu.Unlock()
	return c.authSource
}

func (c *CertificateClient) IssueCertificate(ctx context.Context, input CertificateIssueInput) (CertificateIssueResult, error) {
	if c == nil || c.mode == "disabled" {
		return CertificateIssueResult{}, ErrCertificateIntegrationDisabled
	}
	input.OperationKey = strings.TrimSpace(input.OperationKey)
	input.ProductKind = strings.ToLower(strings.TrimSpace(input.ProductKind))
	input.ProductTitle = strings.TrimSpace(input.ProductTitle)
	input.ServiceIDs = sortedUniqueInt64(input.ServiceIDs)
	if input.OperationKey == "" || input.AmountRub <= 0 || (input.ProductKind != "nominal" && input.ProductKind != "service") {
		return CertificateIssueResult{}, fmt.Errorf("%w: operation, product kind and amount are required", ErrCertificateMapping)
	}
	if input.ProductKind == "service" && len(input.ServiceIDs) == 0 {
		return CertificateIssueResult{}, fmt.Errorf("%w: service mappings are required", ErrCertificateMapping)
	}
	product, err := c.resolveCertificateProduct(ctx, input)
	if err != nil {
		return CertificateIssueResult{}, err
	}
	if c.mode != "write" {
		return CertificateIssueResult{}, ErrCertificateWriteDisabled
	}
	code, err := c.generateCode(ctx, product.ProductID)
	if err != nil {
		return CertificateIssueResult{}, err
	}
	result, err := c.createSale(ctx, input, product, code)
	if err != nil {
		return CertificateIssueResult{}, err
	}
	result.CertificateTypeID = product.CertificateTypeID
	result.ProductID = product.ProductID
	result.LocationID = c.locationID
	result.Code = code
	return result, nil
}

func (c *CertificateClient) resolveCertificateProduct(ctx context.Context, input CertificateIssueInput) (resolvedCertificateProduct, error) {
	cacheKey := certificateResolutionKey(input)
	now := c.now()
	c.mu.Lock()
	if cached, ok := c.resolved[cacheKey]; ok && now.Before(cached.expiresAt) {
		c.mu.Unlock()
		return cached.value, nil
	}
	delete(c.resolved, cacheKey)
	c.mu.Unlock()

	types, err := c.listCertificateTypes(ctx)
	if err != nil {
		return resolvedCertificateProduct{}, err
	}
	typeRecord, found, err := selectCertificateType(types, input, c.salonIDs)
	if err != nil {
		return resolvedCertificateProduct{}, err
	}
	if !found {
		if !c.allowTypeCreate {
			return resolvedCertificateProduct{}, fmt.Errorf("%w: certificate type is absent", ErrCertificateMapping)
		}
		typeRecord, err = c.createCertificateType(ctx, types, input)
		if err != nil {
			return resolvedCertificateProduct{}, err
		}
	}

	var product certificateProduct
	for attempt := 0; attempt < 5; attempt++ {
		products, listErr := c.listProducts(ctx)
		if listErr != nil {
			return resolvedCertificateProduct{}, listErr
		}
		product, found, err = selectCertificateProduct(products, typeRecord.ID)
		if err != nil {
			return resolvedCertificateProduct{}, err
		}
		if found {
			break
		}
		if attempt < 4 {
			select {
			case <-ctx.Done():
				return resolvedCertificateProduct{}, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 250 * time.Millisecond):
			}
		}
	}
	if !found {
		return resolvedCertificateProduct{}, fmt.Errorf("%w: inventory product for certificate type is absent", ErrCertificateMapping)
	}
	resolved := resolvedCertificateProduct{CertificateTypeID: typeRecord.ID, ProductID: int64(product.ID)}
	c.mu.Lock()
	c.resolved[cacheKey] = cachedCertificateProduct{value: resolved, expiresAt: now.Add(c.resolutionTTL)}
	c.mu.Unlock()
	return resolved, nil
}

type certificateType struct {
	ID                      int64   `json:"id"`
	Title                   string  `json:"title"`
	Balance                 float64 `json:"balance"`
	IsMulti                 bool    `json:"is_multi"`
	ItemTypeID              int     `json:"item_type_id"`
	IsAllowEmptyCode        bool    `json:"is_allow_empty_code"`
	BalanceEditTypeID       int     `json:"balance_edit_type_id"`
	ServiceIDs              []int64 `json:"service_ids"`
	SalonIDs                []int64 `json:"salon_ids"`
	ExpirationTypeID        int     `json:"expiration_type_id"`
	ExpirationDate          string  `json:"expiration_date"`
	ExpirationTimeout       int     `json:"expiration_timeout"`
	ExpirationTimeoutUnitID int     `json:"expiration_timeout_unit_id"`
	IsArchived              bool    `json:"is_archived"`
}

type certificateProduct struct {
	ID                       int64Value `json:"id"`
	LoyaltyCertificateTypeID int64Value `json:"loyalty_certificate_type_id"`
}

func (c *CertificateClient) listCertificateTypes(ctx context.Context) ([]certificateType, error) {
	var raw json.RawMessage
	if err := c.requestJSON(ctx, http.MethodGet, fmt.Sprintf("/chain/%d/loyalty/certificate_types", c.chainID), nil, &raw, false); err != nil {
		return nil, err
	}
	var result []certificateType
	if err := decodeEnvelopeData(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *CertificateClient) listProducts(ctx context.Context) ([]certificateProduct, error) {
	var raw json.RawMessage
	path := fmt.Sprintf("/goods/%d?page=1&count=1000", c.locationID)
	if err := c.requestJSON(ctx, http.MethodGet, path, nil, &raw, false); err != nil {
		return nil, err
	}
	var result []certificateProduct
	if err := decodeEnvelopeData(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *CertificateClient) createCertificateType(ctx context.Context, types []certificateType, input CertificateIssueInput) (certificateType, error) {
	template, err := selectCertificateTypeTemplate(types, input.ProductKind, c.salonIDs)
	if err != nil {
		return certificateType{}, err
	}
	itemTypeID := 1
	serviceIDs := []int64{}
	if input.ProductKind == "service" {
		itemTypeID = 3
		serviceIDs = input.ServiceIDs
	}
	payload := map[string]any{
		"title":                certificateTypeTitle(input),
		"balance":              input.AmountRub,
		"is_multi":             template.IsMulti,
		"item_type_id":         itemTypeID,
		"is_allow_empty_code":  false,
		"category_id":          nil,
		"balance_edit_type_id": maxInt(template.BalanceEditTypeID, 1),
		"service_ids":          serviceIDs,
		"salon_ids":            c.salonIDs,
		"expiration": map[string]any{
			"type_id":         template.ExpirationTypeID,
			"date":            emptyToNil(template.ExpirationDate),
			"timeout":         template.ExpirationTimeout,
			"timeout_unit_id": template.ExpirationTimeoutUnitID,
		},
		"online_sale": map[string]any{"is_enabled": false},
	}
	var raw json.RawMessage
	if err := c.requestJSON(ctx, http.MethodPost, fmt.Sprintf("/chain/%d/loyalty/certificate_types", c.chainID), payload, &raw, true); err != nil {
		return certificateType{}, err
	}
	var created certificateType
	if err := decodeEnvelopeData(raw, &created); err != nil {
		return certificateType{}, err
	}
	if created.ID <= 0 {
		return certificateType{}, &CertificateAPIError{Code: "type_response_invalid", PossibleEffect: true, message: "created type id is absent"}
	}
	return created, nil
}

func (c *CertificateClient) generateCode(ctx context.Context, productID int64) (string, error) {
	var raw json.RawMessage
	if err := c.requestJSON(ctx, http.MethodGet, fmt.Sprintf("/loyalty/generate_code/%d/%d", c.locationID, productID), nil, &raw, false); err != nil {
		return "", err
	}
	var data struct {
		Code stringValue `json:"code"`
	}
	if err := decodeEnvelopeData(raw, &data); err != nil {
		return "", err
	}
	code := strings.TrimSpace(string(data.Code))
	if code == "" {
		return "", &CertificateAPIError{Code: "generated_code_invalid", message: "generated code is empty"}
	}
	return code, nil
}

type certificateSaleResponse struct {
	Document struct {
		ID int64Value `json:"id"`
	} `json:"document"`
	Transactions []certificateTransaction `json:"transactions"`
}

type certificateTransaction struct {
	ID                int64Value  `json:"id"`
	DocumentID        int64Value  `json:"document_id"`
	GoodID            int64Value  `json:"good_id"`
	Amount            float64     `json:"amount"`
	CostPerUnit       float64     `json:"cost_per_unit"`
	Cost              float64     `json:"cost"`
	Deleted           bool        `json:"deleted"`
	GoodSpecialNumber stringValue `json:"good_special_number"`
}

func (c *CertificateClient) createSale(ctx context.Context, input CertificateIssueInput, product resolvedCertificateProduct, code string) (CertificateIssueResult, error) {
	payload := map[string]any{
		"type_id":     1,
		"comment":     input.OperationKey,
		"create_date": c.now().Format(time.RFC3339),
		"storage_id":  c.storageID,
		"goods_transactions": []map[string]any{{
			"good_id":             product.ProductID,
			"amount":              1,
			"cost_per_unit":       input.AmountRub,
			"discount":            0,
			"cost":                input.AmountRub,
			"operation_unit_type": 1,
			"good_special_number": code,
			"comment":             input.OperationKey,
		}},
	}
	var raw json.RawMessage
	if err := c.requestJSON(ctx, http.MethodPost, fmt.Sprintf("/storage_operations/operation/%d", c.locationID), payload, &raw, true); err != nil {
		return CertificateIssueResult{}, err
	}
	var created certificateSaleResponse
	if err := decodeEnvelopeData(raw, &created); err != nil {
		return CertificateIssueResult{}, &CertificateAPIError{Code: "sale_response_invalid", PossibleEffect: true, message: err.Error()}
	}
	if int64(created.Document.ID) <= 0 || len(created.Transactions) != 1 || int64(created.Transactions[0].ID) <= 0 {
		return CertificateIssueResult{}, &CertificateAPIError{Code: "sale_response_invalid", PossibleEffect: true, message: "document or transaction is absent"}
	}
	transactionID := int64(created.Transactions[0].ID)
	transaction, err := c.getTransaction(ctx, transactionID)
	if err != nil {
		return CertificateIssueResult{}, &CertificateAPIError{Code: CertificateErrorCode(err), PossibleEffect: true, Retryable: false, message: "sale was created but read-back failed"}
	}
	if err := verifyCertificateTransaction(transaction, int64(created.Document.ID), product.ProductID, input.AmountRub, code); err != nil {
		return CertificateIssueResult{}, &CertificateAPIError{Code: "sale_verification_mismatch", PossibleEffect: true, message: err.Error()}
	}
	return CertificateIssueResult{DocumentID: int64(created.Document.ID), TransactionID: transactionID}, nil
}

func (c *CertificateClient) getTransaction(ctx context.Context, transactionID int64) (certificateTransaction, error) {
	var raw json.RawMessage
	if err := c.requestJSON(ctx, http.MethodGet, fmt.Sprintf("/storage_operations/goods_transactions/%d/%d", c.locationID, transactionID), nil, &raw, false); err != nil {
		return certificateTransaction{}, err
	}
	var result certificateTransaction
	if err := decodeEnvelopeData(raw, &result); err != nil {
		return certificateTransaction{}, err
	}
	return result, nil
}

func verifyCertificateTransaction(transaction certificateTransaction, documentID, productID int64, amountRub int, code string) error {
	if int64(transaction.ID) <= 0 || int64(transaction.DocumentID) != documentID || int64(transaction.GoodID) != productID || transaction.Deleted {
		return errors.New("transaction identity mismatch")
	}
	if math.Abs(math.Abs(transaction.Amount)-1) > 0.0001 {
		return errors.New("transaction quantity mismatch")
	}
	if math.Abs(transaction.CostPerUnit-float64(amountRub)) > 0.01 || math.Abs(transaction.Cost-float64(amountRub)) > 0.01 {
		return errors.New("transaction amount mismatch")
	}
	if strings.TrimSpace(string(transaction.GoodSpecialNumber)) != strings.TrimSpace(code) {
		return errors.New("certificate code mismatch")
	}
	return nil
}

func selectCertificateType(types []certificateType, input CertificateIssueInput, salonIDs []int64) (certificateType, bool, error) {
	title := certificateTypeTitle(input)
	matches := make([]certificateType, 0, 1)
	itemTypeID := 1
	serviceIDs := []int64{}
	if input.ProductKind == "service" {
		itemTypeID = 3
		serviceIDs = input.ServiceIDs
	}
	for _, candidate := range types {
		if candidate.IsArchived || candidate.IsAllowEmptyCode || candidate.ItemTypeID != itemTypeID {
			continue
		}
		if input.ProductKind != "service" && candidate.Title != title {
			continue
		}
		if math.Abs(candidate.Balance-float64(input.AmountRub)) > 0.01 || !sameInt64Set(candidate.ServiceIDs, serviceIDs) || !sameInt64Set(candidate.SalonIDs, salonIDs) {
			continue
		}
		matches = append(matches, candidate)
	}
	if len(matches) > 1 {
		return certificateType{}, false, fmt.Errorf("%w: duplicate exact certificate types", ErrCertificateMapping)
	}
	if len(matches) == 0 {
		return certificateType{}, false, nil
	}
	return matches[0], true, nil
}

func selectCertificateTypeTemplate(types []certificateType, productKind string, salonIDs []int64) (certificateType, error) {
	itemTypeID := 1
	titlePrefix := "Сайт произвольный номинал - "
	if productKind == "service" {
		itemTypeID = 3
		titlePrefix = "Сайт "
	}
	candidates := make([]certificateType, 0)
	for _, candidate := range types {
		if candidate.IsArchived || candidate.IsAllowEmptyCode || candidate.ItemTypeID != itemTypeID || !strings.HasPrefix(candidate.Title, titlePrefix) || !sameInt64Set(candidate.SalonIDs, salonIDs) {
			continue
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 {
		return certificateType{}, fmt.Errorf("%w: approved certificate type template is absent", ErrCertificateMapping)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	baseline := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.IsMulti != baseline.IsMulti || candidate.BalanceEditTypeID != baseline.BalanceEditTypeID ||
			candidate.ExpirationTypeID != baseline.ExpirationTypeID || candidate.ExpirationTimeout != baseline.ExpirationTimeout ||
			candidate.ExpirationTimeoutUnitID != baseline.ExpirationTimeoutUnitID {
			return certificateType{}, fmt.Errorf("%w: certificate templates have conflicting policies", ErrCertificateMapping)
		}
	}
	return baseline, nil
}

func selectCertificateProduct(products []certificateProduct, typeID int64) (certificateProduct, bool, error) {
	matches := make([]certificateProduct, 0, 1)
	for _, product := range products {
		if int64(product.LoyaltyCertificateTypeID) == typeID && int64(product.ID) > 0 {
			matches = append(matches, product)
		}
	}
	if len(matches) > 1 {
		return certificateProduct{}, false, fmt.Errorf("%w: duplicate inventory products for certificate type", ErrCertificateMapping)
	}
	if len(matches) == 0 {
		return certificateProduct{}, false, nil
	}
	return matches[0], true, nil
}

func certificateTypeTitle(input CertificateIssueInput) string {
	if input.ProductKind == "nominal" {
		return fmt.Sprintf("Сайт произвольный номинал - %d", input.AmountRub)
	}
	title := strings.TrimSpace(input.ProductTitle)
	title = strings.TrimPrefix(title, "Сайт ")
	suffix := fmt.Sprintf(" - %d", input.AmountRub)
	title = strings.TrimSuffix(title, suffix)
	return "Сайт " + certificateServiceTypeTitle(title) + suffix
}

func certificateServiceTypeTitle(title string) string {
	title = strings.Join(strings.Fields(strings.TrimSpace(title)), " ")
	separator := strings.LastIndex(title, " - ")
	if separator <= 0 {
		return title
	}
	duration := strings.TrimSpace(title[separator+3:])
	parts := strings.Fields(duration)
	if len(parts) != 2 || !strings.HasPrefix(parts[1], "минут") {
		return title
	}
	minutes, err := strconv.Atoi(parts[0])
	if err != nil || minutes <= 0 {
		return title
	}
	return strings.TrimSpace(title[:separator]) + " " + duration
}

func certificateResolutionKey(input CertificateIssueInput) string {
	payload, _ := json.Marshal(map[string]any{
		"kind": input.ProductKind, "title": certificateTypeTitle(input), "amount": input.AmountRub,
		"serviceIds": sortedUniqueInt64(input.ServiceIDs),
	})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func (c *CertificateClient) requestJSON(ctx context.Context, method, path string, payload any, target any, possibleEffect bool) error {
	var rawPayload []byte
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		rawPayload = raw
	}
	response, raw, err := c.authorizedRequest(ctx, method, path, rawPayload, possibleEffect)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &CertificateAPIError{
			Code: "http_" + strconv.Itoa(response.StatusCode), StatusCode: response.StatusCode,
			PossibleEffect: possibleEffect && response.StatusCode >= 500,
			Retryable:      response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500,
			message:        certificateAPIMessage(raw),
		}
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return &CertificateAPIError{Code: "response_decode_error", StatusCode: response.StatusCode, PossibleEffect: possibleEffect, message: "invalid JSON response"}
	}
	return nil
}

func (c *CertificateClient) authorizedRequest(ctx context.Context, method, path string, payload []byte, possibleEffect bool) (*http.Response, []byte, error) {
	response, raw, err := c.doAuthorizedRequest(ctx, method, path, payload, possibleEffect)
	if err != nil {
		return nil, nil, err
	}
	if (response.StatusCode != http.StatusUnauthorized && response.StatusCode != http.StatusForbidden) || !c.hasLoginFallback() || c.AuthSource() == "login_fallback" {
		return response, raw, nil
	}
	response.Body.Close()
	c.authMu.Lock()
	c.authorization = ""
	c.authSource = ""
	c.authMu.Unlock()
	return c.doAuthorizedRequest(ctx, method, path, payload, possibleEffect)
}

func (c *CertificateClient) doAuthorizedRequest(ctx context.Context, method, path string, payload []byte, possibleEffect bool) (*http.Response, []byte, error) {
	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, nil, err
	}
	request.Header.Set("Accept", "application/vnd.api.v2+json")
	request.Header.Set("Content-Type", "application/json")
	authorization, err := c.certificateAuthorizationHeader(ctx)
	if err != nil {
		return nil, nil, err
	}
	request.Header.Set("Authorization", authorization)
	response, err := c.client.Do(request)
	if err != nil {
		return nil, nil, &CertificateAPIError{Code: "transport_error", PossibleEffect: possibleEffect, Retryable: !possibleEffect, message: "transport failed"}
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	response.Body.Close()
	if err != nil {
		return nil, nil, &CertificateAPIError{Code: "response_read_error", StatusCode: response.StatusCode, PossibleEffect: possibleEffect, Retryable: !possibleEffect, message: "response read failed"}
	}
	return response, raw, nil
}

func (c *CertificateClient) hasLoginFallback() bool {
	return c != nil && c.partnerToken != "" && c.login != "" && c.password != ""
}

func decodeEnvelopeData(raw json.RawMessage, target any) error {
	var envelope struct {
		Success *bool           `json:"success"`
		Data    json.RawMessage `json:"data"`
		Meta    json.RawMessage `json:"meta"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Success != nil {
		if !*envelope.Success {
			return &CertificateAPIError{Code: "provider_rejected", message: certificateAPIMessage(raw)}
		}
		if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
			return &CertificateAPIError{Code: "response_data_absent", message: "response data is absent"}
		}
		return json.Unmarshal(envelope.Data, target)
	}
	return json.Unmarshal(raw, target)
}

func (c *CertificateClient) certificateAuthorizationHeader(ctx context.Context) (string, error) {
	c.authMu.Lock()
	defer c.authMu.Unlock()
	if c.authorization != "" {
		return c.authorization, nil
	}
	if c.partnerToken == "" || c.login == "" || c.password == "" {
		return "", fmt.Errorf("%w: YClients authorization is incomplete", ErrCertificateMapping)
	}
	payload, err := json.Marshal(map[string]string{"login": c.login, "password": c.password})
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/vnd.api.v2+json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.partnerToken)
	response, err := c.client.Do(request)
	if err != nil {
		return "", &CertificateAPIError{Code: "auth_transport_error", Retryable: true, message: "authentication transport failed"}
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", &CertificateAPIError{Code: "auth_response_read_error", StatusCode: response.StatusCode, Retryable: true, message: "authentication response read failed"}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", &CertificateAPIError{
			Code: "auth_http_" + strconv.Itoa(response.StatusCode), StatusCode: response.StatusCode,
			Retryable: response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500,
			message:   "authentication rejected",
		}
	}
	var envelope struct {
		Data struct {
			UserToken string `json:"user_token"`
		} `json:"data"`
		UserToken string `json:"user_token"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return "", &CertificateAPIError{Code: "auth_invalid_json", StatusCode: response.StatusCode, message: "authentication response is invalid"}
	}
	userToken := strings.TrimSpace(envelope.Data.UserToken)
	if userToken == "" {
		userToken = strings.TrimSpace(envelope.UserToken)
	}
	c.authorization = certificateAuthorization(c.partnerToken, userToken)
	if c.authorization == "" {
		return "", &CertificateAPIError{Code: "auth_user_token_missing", StatusCode: response.StatusCode, message: "authentication token is absent"}
	}
	c.authSource = "login_fallback"
	return c.authorization, nil
}

func certificateAPIMessage(raw []byte) string {
	var envelope struct {
		Meta struct {
			Message string `json:"message"`
		} `json:"meta"`
	}
	if json.Unmarshal(raw, &envelope) != nil {
		return "provider rejected the request"
	}
	message := strings.Join(strings.Fields(envelope.Meta.Message), " ")
	if message == "" {
		return "provider rejected the request"
	}
	if len(message) > 180 {
		message = message[:180]
	}
	return message
}

func certificateAuthorization(partner, user string) string {
	partner = strings.TrimSpace(partner)
	user = strings.TrimSpace(user)
	partner = strings.TrimSpace(strings.TrimPrefix(partner, "Bearer "))
	user = strings.TrimSpace(strings.TrimPrefix(user, "User "))
	if partner == "" || user == "" || strings.ContainsAny(partner+user, "\r\n") {
		return ""
	}
	return "Bearer " + partner + ", User " + user
}

func sortedUniqueInt64(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func sameInt64Set(left, right []int64) bool {
	left = sortedUniqueInt64(left)
	right = sortedUniqueInt64(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func emptyToNil(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func maxInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

type int64Value int64

func (value *int64Value) UnmarshalJSON(raw []byte) error {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) || bytes.Equal(raw, []byte(`""`)) {
		*value = 0
		return nil
	}
	var number json.Number
	if raw[0] == '"' {
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return err
		}
		number = json.Number(text)
	} else {
		number = json.Number(string(raw))
	}
	parsed, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil {
		floatValue, floatErr := strconv.ParseFloat(number.String(), 64)
		if floatErr != nil {
			return err
		}
		parsed = int64(floatValue)
	}
	*value = int64Value(parsed)
	return nil
}

type stringValue string

func (value *stringValue) UnmarshalJSON(raw []byte) error {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		*value = ""
		return nil
	}
	if raw[0] == '"' {
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return err
		}
		*value = stringValue(text)
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var number json.Number
	if err := decoder.Decode(&number); err != nil {
		return err
	}
	*value = stringValue(number.String())
	return nil
}
