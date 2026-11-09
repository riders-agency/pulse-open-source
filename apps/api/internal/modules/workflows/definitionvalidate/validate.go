package definitionvalidate

import (
	"encoding/json"
	"fmt"
	"strings"
)

type nodeDef struct {
	ActionKey     string         `json:"action_key"`
	ActionTypeKey string         `json:"action_type_key"`
	Config        map[string]any `json:"config"`
}

type workflowDefinition struct {
	Nodes []nodeDef `json:"nodes"`
}

func stringFromConfig(cfg map[string]any, key string) string {
	if cfg == nil {
		return ""
	}
	v, ok := cfg[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		// JSON numbers (e.g. mis-typed ids)
		return strings.TrimSpace(fmt.Sprintf("%.0f", t))
	case bool:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

// Validate checks workflow draft/published definition_json for actions that must be fully configured.
func Validate(defJSON []byte) error {
	if len(strings.TrimSpace(string(defJSON))) == 0 {
		return fmt.Errorf("описание сценария пустое")
	}
	var def workflowDefinition
	if err := json.Unmarshal(defJSON, &def); err != nil {
		return fmt.Errorf("некорректный JSON сценария: %w", err)
	}
	for _, n := range def.Nodes {
		key := strings.TrimSpace(n.ActionKey)
		if key == "" {
			key = "(без ключа)"
		}
		atk := strings.TrimSpace(n.ActionTypeKey)
		if value, exists := n.Config["disabled"]; exists {
			disabled, ok := value.(bool)
			if !ok {
				return fmt.Errorf("действие %q: disabled должен быть логическим значением", key)
			}
			if disabled {
				if strings.HasSuffix(atk, ".trigger") || atk == "task.event_trigger" {
					return fmt.Errorf("действие %q: отключайте запуск через настройки триггера или всего сценария", key)
				}
				continue
			}
		}
		switch atk {
		case "telegram.relay.trigger":
			if stringFromConfig(n.Config, "action_dependency_id") == "" {
				return fmt.Errorf("действие %q (Telegram relay): выберите интеграцию telegram.relay", key)
			}
		case "schedule.trigger":
			if stringFromConfig(n.Config, "schedule_key") == "" {
				return fmt.Errorf("действие %q (периодический запуск): укажите ключ расписания", key)
			}
		case "http.request":
			if stringFromConfig(n.Config, "url") == "" {
				return fmt.Errorf("действие %q (HTTP запрос): укажите URL", key)
			}
		case "email.send":
			if stringFromConfig(n.Config, "smtp_action_dependency_id") == "" {
				return fmt.Errorf("действие %q (отправить email): выберите интеграцию SMTP", key)
			}
			if stringFromConfig(n.Config, "to") == "" {
				return fmt.Errorf("действие %q (отправить email): укажите получателя (to)", key)
			}
			if stringFromConfig(n.Config, "subject") == "" {
				return fmt.Errorf("действие %q (отправить email): укажите тему письма (subject)", key)
			}
		case "bot.set_state":
			if stringFromConfig(n.Config, "state_key") == "" {
				return fmt.Errorf("действие %q (состояние бота): укажите state_key", key)
			}
		}
	}
	return nil
}
