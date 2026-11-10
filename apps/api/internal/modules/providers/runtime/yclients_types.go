package runtime

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	yclientsprovider "riders.agency/pulse/api/internal/modules/providers/yclients"
)

type yclientsCredentials struct {
	PartnerToken      string `json:"partner_token"`
	UserToken         string `json:"user_token"`
	Login             string `json:"login"`
	Password          string `json:"password"`
	ReceiptHMACSecret string `json:"receipt_hmac_secret"`
	ReceiptBasicUser  string `json:"receipt_basic_username"`
	ReceiptBasicPass  string `json:"receipt_basic_password"`
}

type yclientsCommand struct {
	OperationKey      string
	OrderID           string
	OrderItemID       string
	InstanceID        string
	CertificateNumber string
	ProductKind       string
	ProductTitle      string
	AmountRub         int
	ServiceIDs        []int64
}

type yclientsReceipt struct {
	WorkflowRunID     string `json:"workflowRunId"`
	OperationID       string `json:"operationId"`
	OperationKey      string `json:"operationKey"`
	OrderID           string `json:"orderId"`
	OrderItemID       string `json:"orderItemId"`
	InstanceID        string `json:"instanceId"`
	CertificateNumber string `json:"certificateNumber"`
	Outcome           string `json:"outcome"`
	CertificateCode   string `json:"certificateCode,omitempty"`
	CertificateTypeID int64  `json:"certificateTypeId,omitempty"`
	ProductID         int64  `json:"productId,omitempty"`
	LocationID        int64  `json:"locationId,omitempty"`
	DocumentID        int64  `json:"documentId,omitempty"`
	TransactionID     int64  `json:"transactionId,omitempty"`
	ErrorCode         string `json:"errorCode,omitempty"`
	PossibleEffect    bool   `json:"possibleEffect"`
	Retryable         bool   `json:"retryable"`
}

func parseYClientsCredentials(raw string) (yclientsCredentials, error) {
	var credentials yclientsCredentials
	if json.Unmarshal([]byte(strings.TrimSpace(raw)), &credentials) != nil {
		return credentials, fmt.Errorf("YClients encrypted credential profile is invalid")
	}
	credentials.PartnerToken = strings.TrimSpace(credentials.PartnerToken)
	credentials.UserToken = strings.TrimSpace(credentials.UserToken)
	credentials.Login = strings.TrimSpace(credentials.Login)
	credentials.ReceiptBasicUser = strings.TrimSpace(credentials.ReceiptBasicUser)
	if credentials.PartnerToken == "" || credentials.ReceiptHMACSecret == "" {
		return credentials, fmt.Errorf("YClients encrypted credential profile is incomplete")
	}
	if credentials.UserToken == "" && (credentials.Login == "" || credentials.Password == "") {
		return credentials, fmt.Errorf("YClients user token or login fallback is required")
	}
	if (credentials.ReceiptBasicUser == "") != (credentials.ReceiptBasicPass == "") {
		return credentials, fmt.Errorf("receipt Basic Auth username and password must be configured together")
	}
	return credentials, nil
}

func yclientsClientConfig(dep *dependencyRecord, credentials yclientsCredentials, clientTimeout time.Duration) yclientsprovider.CertificateConfig {
	return yclientsprovider.CertificateConfig{
		Mode: stringMetadata(dep.Metadata, "mode"), BaseURL: stringMetadata(dep.Metadata, "api_base"),
		PartnerToken: credentials.PartnerToken, UserToken: credentials.UserToken,
		Login: credentials.Login, Password: credentials.Password,
		ChainID: metadataInt64(dep.Metadata, "chain_id"), LocationID: metadataInt64(dep.Metadata, "location_id"),
		StorageID: metadataInt64(dep.Metadata, "storage_id"), SalonIDs: metadataInt64Slice(dep.Metadata, "salon_ids"),
		AllowTypeCreate: boolMetadata(dep.Metadata, "allow_type_create"), Timeout: clientTimeout,
		ResolutionTTL: time.Duration(metadataInt64(dep.Metadata, "resolution_ttl_seconds")) * time.Second,
	}
}

func parseServiceIDs(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" || raw == "[]" {
		return nil, nil
	}
	var values []int64
	if err := json.Unmarshal([]byte(raw), &values); err == nil {
		return values, nil
	}
	for _, part := range strings.Split(raw, ",") {
		value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || value <= 0 {
			return nil, fmt.Errorf("service_ids_json must contain positive YClients IDs")
		}
		values = append(values, value)
	}
	return values, nil
}

func metadataInt64(metadata map[string]any, key string) int64 {
	raw := metadata[key]
	if value, ok := raw.(float64); ok {
		if value <= 0 || value != math.Trunc(value) || value > 1<<53 {
			return 0
		}
		return int64(value)
	}
	if value, ok := raw.(json.Number); ok {
		parsed, err := value.Int64()
		if err != nil || parsed <= 0 {
			return 0
		}
		return parsed
	}
	value := strings.TrimSpace(fmt.Sprint(raw))
	if value == "" || value == "<nil>" {
		return 0
	}
	parsed, _ := strconv.ParseInt(strings.TrimSuffix(value, ".0"), 10, 64)
	if parsed <= 0 {
		return 0
	}
	return parsed
}

func metadataInt64Slice(metadata map[string]any, key string) []int64 {
	raw, _ := json.Marshal(metadata[key])
	var values []int64
	if json.Unmarshal(raw, &values) == nil {
		return values
	}
	return nil
}

func metadataStringSlice(metadata map[string]any, key string) []string {
	raw, _ := json.Marshal(metadata[key])
	var values []string
	_ = json.Unmarshal(raw, &values)
	return values
}
