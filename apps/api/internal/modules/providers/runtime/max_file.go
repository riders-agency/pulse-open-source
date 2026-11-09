package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

const maxDocumentBytes = 16 << 20

type maxPreparedDocument struct {
	Attachment map[string]any
}

type DocumentBasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func ParseDocumentBasicAuth(raw string) (map[string]DocumentBasicAuth, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var values map[string]DocumentBasicAuth
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, errors.New("PULSE_DOCUMENT_BASIC_AUTH_JSON is invalid")
	}
	result := make(map[string]DocumentBasicAuth, len(values))
	for rawHost, credentials := range values {
		host, err := normalizeDocumentAuthHost(rawHost)
		if err != nil {
			return nil, err
		}
		credentials.Username = strings.TrimSpace(credentials.Username)
		if credentials.Username == "" || strings.TrimSpace(credentials.Password) == "" {
			return nil, fmt.Errorf("document Basic Auth credentials for %s are incomplete", host)
		}
		if _, exists := result[host]; exists {
			return nil, fmt.Errorf("duplicate document Basic Auth host %s", host)
		}
		result[host] = credentials
	}
	return result, nil
}

func normalizeDocumentAuthHost(value string) (string, error) {
	host := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "."))
	parsed, err := url.Parse("https://" + host)
	if err != nil || host == "" || parsed.Host != host || parsed.Hostname() == "" || parsed.Port() != "" || parsed.User != nil || parsed.Path != "" {
		return "", errors.New("document Basic Auth host is invalid")
	}
	return host, nil
}

func (r *Runner) documentBasicAuthForHost(host string) (DocumentBasicAuth, bool) {
	if r == nil || len(r.DocumentBasicAuth) == 0 {
		return DocumentBasicAuth{}, false
	}
	normalized, err := normalizeDocumentAuthHost(host)
	if err != nil {
		return DocumentBasicAuth{}, false
	}
	credentials, ok := r.DocumentBasicAuth[normalized]
	return credentials, ok
}

func (r *Runner) prepareMaxFileAttachment(
	ctx context.Context,
	apiBase string,
	token string,
	sourceURL string,
	metadata map[string]any,
) (*maxPreparedDocument, error) {
	source, err := validateMaxDocumentURL(sourceURL, maxDocumentAllowedHosts(metadata), true)
	if err != nil {
		return nil, err
	}
	body, fileName, err := r.downloadMaxDocument(ctx, source)
	if err != nil {
		return nil, err
	}
	uploadURL, err := r.requestMaxUploadURL(ctx, apiBase, token)
	if err != nil {
		return nil, err
	}
	uploadToken, err := r.uploadMaxDocument(ctx, uploadURL, fileName, body)
	if err != nil {
		return nil, err
	}
	return &maxPreparedDocument{
		Attachment: map[string]any{
			"type":    "file",
			"payload": map[string]any{"token": uploadToken},
		},
	}, nil
}

func (r *Runner) downloadMaxDocument(ctx context.Context, source *url.URL) ([]byte, string, error) {
	client := *r.maxHTTPClient()
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("certificate document has too many redirects")
		}
		request.Header.Del("Authorization")
		_, err := validateMaxDocumentURL(request.URL.String(), []string{source.Hostname()}, true)
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source.String(), nil)
	if err != nil {
		return nil, "", err
	}
	request.Header.Set("Accept", "application/pdf")
	if credentials, ok := r.documentBasicAuthForHost(source.Hostname()); ok {
		request.SetBasicAuth(credentials.Username, credentials.Password)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("download certificate document: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("download certificate document returned HTTP %d", response.StatusCode)
	}
	if response.ContentLength > maxDocumentBytes {
		return nil, "", errors.New("certificate document exceeds 16 MiB")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxDocumentBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(body) > maxDocumentBytes {
		return nil, "", errors.New("certificate document exceeds 16 MiB")
	}
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]))
	if contentType != "application/pdf" && !bytes.HasPrefix(body, []byte("%PDF-")) {
		return nil, "", errors.New("certificate document is not a PDF")
	}
	fileName := maxDocumentFileName(response.Header.Get("Content-Disposition"), source.Path)
	return body, fileName, nil
}

func (r *Runner) requestMaxUploadURL(ctx context.Context, apiBase string, token string) (*url.URL, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(apiBase, "/")+"/uploads?type=file", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", maxAuthorizationHeader(token, nil))
	response, err := r.maxHTTPClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("request MAX upload URL: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("MAX upload negotiation returned HTTP %d", response.StatusCode)
	}
	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(raw, &result); err != nil || strings.TrimSpace(result.URL) == "" {
		return nil, errors.New("MAX upload negotiation returned no URL")
	}
	uploadURL, err := url.Parse(result.URL)
	if err != nil {
		return nil, errors.New("MAX upload URL is invalid")
	}
	if !maxUploadHostAllowed(uploadURL) {
		return nil, errors.New("MAX upload URL host is not allowed")
	}
	return uploadURL, nil
}

func (r *Runner) uploadMaxDocument(ctx context.Context, uploadURL *url.URL, fileName string, body []byte) (string, error) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("data", fileName)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(body); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL.String(), &requestBody)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := r.maxHTTPClient().Do(request)
	if err != nil {
		return "", fmt.Errorf("upload certificate document to MAX: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("MAX file upload returned HTTP %d", response.StatusCode)
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(raw, &result); err != nil || strings.TrimSpace(result.Token) == "" {
		return "", errors.New("MAX file upload returned no token")
	}
	return result.Token, nil
}

func validateMaxDocumentURL(raw string, allowedHosts []string, allowLoopback bool) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Hostname() == "" || parsed.User != nil {
		return nil, errors.New("certificate document URL is invalid")
	}
	host := strings.ToLower(parsed.Hostname())
	isLoopback := net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
	if parsed.Scheme != "https" && !(allowLoopback && parsed.Scheme == "http" && isLoopback) {
		return nil, errors.New("certificate document URL must use HTTPS")
	}
	allowed := false
	for _, candidate := range allowedHosts {
		if host == strings.ToLower(strings.TrimSpace(candidate)) {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, errors.New("certificate document URL host is not allowed")
	}
	return parsed, nil
}

func maxDocumentAllowedHosts(metadata map[string]any) []string {
	hosts := []string{"shop.example.test"}
	switch value := metadata["document_allowed_hosts"].(type) {
	case string:
		for _, host := range strings.Split(value, ",") {
			if strings.TrimSpace(host) != "" {
				hosts = append(hosts, strings.TrimSpace(host))
			}
		}
	case []any:
		for _, raw := range value {
			if host := strings.TrimSpace(fmt.Sprint(raw)); host != "" {
				hosts = append(hosts, host)
			}
		}
	case []string:
		hosts = append(hosts, value...)
	}
	return hosts
}

func maxDocumentMetadata(metadata map[string]any, configuredHosts []string) map[string]any {
	if len(configuredHosts) == 0 {
		return metadata
	}
	result := make(map[string]any, len(metadata)+1)
	for key, value := range metadata {
		result[key] = value
	}
	result["document_allowed_hosts"] = configuredHosts
	return result
}

func maxUploadHostAllowed(target *url.URL) bool {
	host := strings.ToLower(target.Hostname())
	if target.Scheme == "https" && host == "fu.oneme.ru" {
		return true
	}
	ip := net.ParseIP(host)
	return target.Scheme == "http" && ip != nil && ip.IsLoopback()
}

func maxDocumentFileName(contentDisposition string, sourcePath string) string {
	if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
		if name := filepath.Base(strings.TrimSpace(params["filename"])); strings.HasSuffix(strings.ToLower(name), ".pdf") {
			return name
		}
	}
	name := filepath.Base(sourcePath)
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		return "demo-shop-certificate.pdf"
	}
	return name
}

func maxAttachmentNotReady(provider map[string]any) bool {
	if strings.EqualFold(fmt.Sprint(provider["code"]), "attachment.not.ready") {
		return true
	}
	root, _ := provider["raw_response_json"].(map[string]any)
	return strings.EqualFold(fmt.Sprint(root["code"]), "attachment.not.ready")
}
