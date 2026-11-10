package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type TaskDraftExtraction struct {
	Title              string
	Description        string
	ClientQuery        string
	ClientID           string
	ClientName         string
	AssigneeHint       string
	DueDate            string
	DueDateLabel       string
	Priority           string
	Billable           bool
	Labels             []string
	Checklist          []string
	NotificationIntent string
	StartTimer         bool
	MissingFields      []map[string]any
	Confidence         map[string]float64
}

func (a *AI) extractTaskDraft(ctx context.Context, projectID, userID uuid.UUID, body string, intent aiIntent, localDate string, timezone string) TaskDraftExtraction {
	lower := strings.ToLower(body)
	clientQuery := bestClientQuery(body, stringEntity(intent, "client_query"))
	title := stringEntity(intent, "title")
	if title == "" {
		title = taskTitleFromText(body)
	}
	title = cleanTaskTitle(title, clientQuery)
	if title == "" {
		if strings.Contains(lower, "оплат") {
			title = "Проверить оплату"
		} else {
			title = "Новая задача"
		}
	}

	dueDate := dueDateFromTextWithBase(body, localDate, timezone)
	dueLabel := dueDateLabelFromText(body)
	matches := a.findClients(ctx, projectID, userID, clientQuery, 5)
	clientID, clientName := firstClientMatch(matches)
	if clientID == "" {
		if fallbackQuery := clientQueryFromText(body); fallbackQuery != "" && !strings.EqualFold(fallbackQuery, clientQuery) {
			fallbackMatches := a.findClients(ctx, projectID, userID, fallbackQuery, 5)
			if fallbackID, fallbackName := firstClientMatch(fallbackMatches); fallbackID != "" {
				clientQuery = fallbackQuery
				clientID = fallbackID
				clientName = fallbackName
			}
		}
	}

	labels := []string{"ai-created"}
	checklist := []string{}
	notification := ""
	description := taskDescriptionFromText(body)
	if containsSEOIntent(lower) {
		labels = appendUnique(labels, "seo")
		title = "Внести правки для SEO-специалистов"
		description = "Сделать все правки для SEO-специалистов и сообщить им о результате."
		checklist = []string{
			"Собрать список правок",
			"Внести правки",
			"Проверить результат",
			"Сообщить SEO-специалистам",
		}
	}
	if strings.Contains(lower, "сообщ") {
		notification = "Сообщить SEO-специалистам после выполнения"
	}
	startTimer := strings.Contains(lower, "таймер") || strings.Contains(lower, "трекер времени") || strings.Contains(lower, "стартани")

	assigneeHint := ""
	if strings.Contains(lower, "мне") || strings.Contains(lower, "на меня") {
		assigneeHint = "me"
	}
	missing := []map[string]any{}
	if clientQuery != "" && clientID == "" {
		missing = append(missing, map[string]any{
			"field": "client_id",
			"label": "Клиент",
			"query": clientQuery,
		})
	}
	return TaskDraftExtraction{
		Title:              title,
		Description:        description,
		ClientQuery:        clientQuery,
		ClientID:           clientID,
		ClientName:         clientName,
		AssigneeHint:       assigneeHint,
		DueDate:            dueDate,
		DueDateLabel:       dueLabel,
		Priority:           "medium",
		Billable:           false,
		Labels:             labels,
		Checklist:          checklist,
		NotificationIntent: notification,
		StartTimer:         startTimer,
		MissingFields:      missing,
		Confidence: map[string]float64{
			"client_query":  confidenceValue(clientQuery != ""),
			"client_match":  confidenceValue(clientID != ""),
			"date_resolved": confidenceValue(dueDate != ""),
			"title":         confidenceValue(title != "" && title != "Новая задача"),
		},
	}
}

func dueDateLabelFromText(text string) string {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "послезавтра"):
		return "послезавтра"
	case strings.Contains(lower, "завтра"):
		return "завтра"
	case strings.Contains(lower, "сегодня"):
		return "сегодня"
	default:
		return ""
	}
}

func containsSEOIntent(lower string) bool {
	return strings.Contains(lower, "сео") ||
		strings.Contains(lower, "seo") ||
		strings.Contains(lower, "сеош") ||
		strings.Contains(lower, "сешо")
}

func appendUnique(values []string, next string) []string {
	for _, value := range values {
		if value == next {
			return values
		}
	}
	return append(values, next)
}

func confidenceValue(ok bool) float64 {
	if ok {
		return 0.95
	}
	return 0.2
}
