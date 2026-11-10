package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrSMTPMetadata = errors.New("smtp metadata invalid")

func validateSMTPMetadata(meta json.RawMessage) error {
	if len(meta) == 0 {
		return fmt.Errorf("%w: задайте host, port и from_email в metadata_json", ErrSMTPMetadata)
	}
	var m map[string]any
	if err := json.Unmarshal(meta, &m); err != nil {
		return fmt.Errorf("%w: %v", ErrSMTPMetadata, err)
	}
	host := strings.TrimSpace(fmt.Sprint(m["host"]))
	if host == "" {
		return fmt.Errorf("%w: поле host обязательно", ErrSMTPMetadata)
	}
	from := strings.TrimSpace(fmt.Sprint(m["from_email"]))
	if from == "" {
		return fmt.Errorf("%w: поле from_email обязательно", ErrSMTPMetadata)
	}
	port := 0
	switch v := m["port"].(type) {
	case float64:
		port = int(v)
	case int:
		port = v
	case string:
		p, _ := strconv.Atoi(strings.TrimSpace(v))
		port = p
	default:
		if v != nil {
			p, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprint(v)))
			port = p
		}
	}
	if port <= 0 || port > 65535 {
		return fmt.Errorf("%w: port должен быть от 1 до 65535", ErrSMTPMetadata)
	}
	return nil
}
