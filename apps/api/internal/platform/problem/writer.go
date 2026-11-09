package problem

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Body struct {
	Type      string         `json:"type"`
	Module    string         `json:"module,omitempty"`
	Code      string         `json:"code,omitempty"`
	Title     string         `json:"title"`
	Status    int            `json:"status"`
	Detail    string         `json:"detail,omitempty"`
	Action    string         `json:"action,omitempty"`
	Validate  map[string]any `json:"validate,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
	Extra     map[string]any `json:"extra,omitempty"`
}

type Spec struct {
	Status   int
	Module   string
	Code     string
	Title    string
	Detail   string
	Action   string
	Validate map[string]any
	Extra    map[string]any
}

func Write(w http.ResponseWriter, status int, title, detail string) {
	spec := Spec{
		Status: status,
		Module: "platform",
		Code:   codeFromTitle(status, title),
		Title:  titleFromStatus(status, title),
		Detail: detail,
		Action: actionFromStatus(status),
	}
	WriteSpec(w, spec)
}

func WriteModule(w http.ResponseWriter, status int, module, code, title, detail, action string) {
	WriteSpec(w, Spec{
		Status: status,
		Module: module,
		Code:   code,
		Title:  title,
		Detail: detail,
		Action: action,
	})
}

func WriteSpec(w http.ResponseWriter, spec Spec) {
	if spec.Status == 0 {
		spec.Status = http.StatusInternalServerError
	}
	if spec.Module == "" {
		spec.Module = "platform"
	}
	if spec.Code == "" {
		spec.Code = codeFromTitle(spec.Status, spec.Title)
	}
	if spec.Title == "" {
		spec.Title = titleFromStatus(spec.Status, spec.Code)
	}
	if spec.Action == "" {
		spec.Action = actionFromStatus(spec.Status)
	}
	validate := spec.Validate
	if validate == nil {
		validate = map[string]any{}
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.Header().Set("X-App-Result", "error")
	w.Header().Set("X-App-Action", spec.Action)
	w.WriteHeader(spec.Status)
	_ = json.NewEncoder(w).Encode(Body{
		Type:      "about:blank",
		Module:    spec.Module,
		Code:      spec.Code,
		Title:     spec.Title,
		Status:    spec.Status,
		Detail:    spec.Detail,
		Action:    spec.Action,
		Validate:  validate,
		RequestID: w.Header().Get("X-Request-Id"),
		Extra:     spec.Extra,
	})
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	if w.Header().Get("X-App-Result") == "" {
		res := "ok"
		if status >= 400 {
			res = "error"
		}
		w.Header().Set("X-App-Result", res)
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func codeFromTitle(status int, title string) string {
	code := strings.TrimSpace(strings.ToLower(title))
	code = strings.ReplaceAll(code, " ", "_")
	code = strings.ReplaceAll(code, "-", "_")
	if code == "" {
		code = strings.ToLower(strings.ReplaceAll(http.StatusText(status), " ", "_"))
	}
	if code == "" {
		code = "http_error"
	}
	return code
}

func titleFromStatus(status int, title string) string {
	trimmed := strings.TrimSpace(title)
	switch status {
	case http.StatusBadRequest:
		if trimmed == "" || isCodeLike(trimmed) {
			return "Запрос не прошел проверку"
		}
	case http.StatusUnauthorized:
		return "Нужна авторизация"
	case http.StatusForbidden:
		return "Нет доступа"
	case http.StatusNotFound:
		return "Ресурс не найден"
	case http.StatusMethodNotAllowed:
		return "Метод не поддерживается"
	case http.StatusConflict:
		if trimmed == "" || isCodeLike(trimmed) {
			return "Конфликт состояния"
		}
	default:
		if status >= 500 {
			return "Ошибка сервера"
		}
	}
	if trimmed != "" {
		return trimmed
	}
	if text := http.StatusText(status); text != "" {
		return text
	}
	return "Ошибка запроса"
}

func actionFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "Проверьте введенные данные и повторите действие."
	case http.StatusUnauthorized:
		return "Войдите снова и повторите действие."
	case http.StatusForbidden:
		return "Проверьте доступ к компании, проекту или ресурсу."
	case http.StatusNotFound:
		return "Проверьте выбранный контекст и обновите страницу."
	case http.StatusMethodNotAllowed:
		return "Используйте поддерживаемое действие для этого раздела."
	case http.StatusConflict:
		return "Обновите данные и повторите действие."
	default:
		if status >= 500 {
			return "Повторите позже или передайте request_id разработчику."
		}
	}
	return "Проверьте запрос и повторите действие."
}

func isCodeLike(value string) bool {
	value = strings.TrimSpace(value)
	return value == strings.ToLower(value) && !strings.Contains(value, " ")
}
