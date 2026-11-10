package max

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/platform/httptransport"
)

var defaultCAFiles = []string{
	"/etc/riders-pulse/max-mincifry-ca.pem",
	"/etc/riders-pulse/russian-trusted-ca.pem",
	"/usr/local/share/ca-certificates/russian_trusted_root_ca.crt",
	"/usr/local/share/ca-certificates/russian_trusted_sub_ca.crt",
	"/usr/local/share/ca-certificates/russian-trusted-root-ca.crt",
	"/usr/local/share/ca-certificates/russian-trusted-sub-ca.crt",
}

type TransportConfig struct {
	Timeout time.Duration
	CAFiles []string
	CADirs  []string
}

func NewHTTPClientFromEnv(timeout time.Duration) (*http.Client, error) {
	cfg := TransportConfig{
		Timeout: timeout,
		CAFiles: caFilesFromEnv(),
		CADirs:  httptransport.SplitPathList(os.Getenv("RIDERS_PULSE_MAX_CA_CERT_DIR")),
	}
	return NewHTTPClient(cfg)
}

func NewHTTPClient(cfg TransportConfig) (*http.Client, error) {
	files := append([]string{}, cfg.CAFiles...)
	for _, dir := range cfg.CADirs {
		matches, err := caFilesInDir(dir)
		if err != nil {
			continue
		}
		files = append(files, matches...)
	}

	pool := httptransport.SystemCertPool()
	for _, path := range files {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read MAX CA bundle %s: %w", path, err)
		}
		if err := httptransport.AppendPEMCerts(pool, raw); err != nil {
			return nil, fmt.Errorf("MAX CA bundle %s: %w", path, err)
		}
	}
	return httptransport.NewClientWithRootCAs(cfg.Timeout, pool), nil
}

func caFilesFromEnv() []string {
	files := append([]string{}, defaultCAFiles...)
	for _, key := range []string{"RIDERS_PULSE_MAX_CA_CERT_FILE", "RIDERS_PULSE_MAX_CA_CERT_BUNDLE"} {
		files = append(files, httptransport.SplitPathList(os.Getenv(key))...)
	}
	return files
}

func caFilesInDir(dir string) ([]string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".crt") || strings.HasSuffix(name, ".cer") {
			out = append(out, filepath.Join(dir, entry.Name()))
		}
	}
	return out, nil
}
