package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestClientQueryFromTaskText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "payment task keeps only client name",
			text: "Создай задачу проверить оплату Орион завтра",
			want: "Орион",
		},
		{
			name: "client typo still extracts entity",
			text: "создай задачу для клиеета Орион закрытыть правки сешоникрв",
			want: "Орион",
		},
		{
			name: "correction keeps client only",
			text: "нет, клиент Орион",
			want: "Орион",
		},
		{
			name: "plain for client phrase",
			text: "создай задача сделать правки сеошников для Орион и запусти таймер по ней",
			want: "Орион",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clientQueryFromText(tt.text); got != tt.want {
				t.Fatalf("clientQueryFromText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTaskTitleFromTextRemovesClientAndDate(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "payment title",
			text: "Создай задачу проверить оплату Орион завтра",
			want: "Проверить оплату",
		},
		{
			name: "seo typo title",
			text: "создай задачу для клиеета Орион закрытыть правки сешоникрв",
			want: "Внести правки для SEO-специалистов",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := taskTitleFromText(tt.text); got != tt.want {
				t.Fatalf("taskTitleFromText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExplicitNewCommandBypassesPendingProposal(t *testing.T) {
	if !isExplicitNewCommand("создай задачу для клиента орион") {
		t.Fatal("expected new task command to bypass pending proposal")
	}
	if isExplicitNewCommand("давай сегодня") {
		t.Fatal("date follow-up must keep pending proposal")
	}
}

func TestTaskCreateFollowUpPriorityRussian(t *testing.T) {
	input := map[string]any{"title": "Проверить оплату", "priority": "medium"}
	updated, ok, reply := (&AI{}).applyTaskCreateFollowUp(
		context.Background(),
		uuid.New(),
		uuid.New(),
		input,
		"сделай приоритет высокий",
		MessageInput{},
	)
	if !ok {
		t.Fatal("expected priority follow-up to update proposal")
	}
	if updated["priority"] != "high" {
		t.Fatalf("priority = %#v, want high", updated["priority"])
	}
	if !strings.Contains(reply, "высокий") {
		t.Fatalf("reply = %q, want high priority label", reply)
	}

	updated, ok, reply = (&AI{}).applyTaskCreateFollowUp(
		context.Background(),
		uuid.New(),
		uuid.New(),
		input,
		"приоритет низкий",
		MessageInput{},
	)
	if !ok {
		t.Fatal("expected low priority follow-up to update proposal")
	}
	if updated["priority"] != "low" {
		t.Fatalf("priority = %#v, want low", updated["priority"])
	}
	if !strings.Contains(reply, "низкий") {
		t.Fatalf("reply = %q, want low priority label", reply)
	}
}

func TestResolveProposalEditPriorityRussian(t *testing.T) {
	toolCallID := uuid.New()
	edit, ok := (&AI{}).ResolveProposalEdit(
		context.Background(),
		uuid.New(),
		uuid.New(),
		MessageInput{Body: "сделай приоритет высокий"},
		SessionState{
			ActiveToolCall: &ToolCallState{
				ID:      toolCallID,
				ToolKey: "task.create",
				Status:  "pending_approval",
				Input:   map[string]any{"title": "Проверить оплату"},
			},
		},
	)
	if !ok {
		t.Fatal("expected proposal edit")
	}
	if edit.ToolCallID != toolCallID {
		t.Fatalf("ToolCallID = %s, want %s", edit.ToolCallID, toolCallID)
	}
	if edit.Patch["priority"] != "high" {
		t.Fatalf("priority patch = %#v, want high", edit.Patch["priority"])
	}
	if !strings.Contains(edit.HumanSummary, "высокий") {
		t.Fatalf("HumanSummary = %q, want high label", edit.HumanSummary)
	}
}

func TestProposalPayloadIncludesEntitySearchSnippet(t *testing.T) {
	fields := map[string]any{
		"title":           "Проверить оплату",
		"client_query":    "Орион",
		"client_id":       "client-1",
		"client_name":     "Орион Codex",
		"client_resolved": true,
		"client_search_snippet": clientSearchSnippet("Орион", []aiClientMatch{
			{ID: "client-1", Name: "Орион Codex"},
		}),
	}
	payload := proposalPayload("tool-call-1", "task.create", "task.create", "high", fields)
	snippets, ok := payload["entity_snippets_json"].([]map[string]any)
	if !ok || len(snippets) != 1 {
		t.Fatalf("entity_snippets_json = %#v, want one snippet", payload["entity_snippets_json"])
	}
	if snippets[0]["tool_key"] != "client.search" {
		t.Fatalf("tool_key = %#v, want client.search", snippets[0]["tool_key"])
	}
	if snippets[0]["selected_name"] != "Орион Codex" {
		t.Fatalf("selected_name = %#v, want Орион Codex", snippets[0]["selected_name"])
	}
}

func TestExtractTaskDraftSEOScenario(t *testing.T) {
	draft := (&AI{}).extractTaskDraft(
		context.Background(),
		uuid.New(),
		uuid.New(),
		"создай задачу мне для клиента Орион завтра сделать все правка для сеошников и сообщить им тоже",
		aiIntent{Key: "task.create", Entities: map[string]any{}},
		"2026-06-15",
		"Europe/Moscow",
	)
	if draft.ClientQuery != "Орион" {
		t.Fatalf("ClientQuery = %q, want Орион", draft.ClientQuery)
	}
	if draft.Title != "Внести правки для SEO-специалистов" {
		t.Fatalf("Title = %q", draft.Title)
	}
	if draft.Description != "Сделать все правки для SEO-специалистов и сообщить им о результате." {
		t.Fatalf("Description = %q", draft.Description)
	}
	if draft.DueDate != "2026-06-16" || draft.DueDateLabel != "завтра" {
		t.Fatalf("due = %q/%q, want 2026-06-16/завтра", draft.DueDate, draft.DueDateLabel)
	}
	if draft.AssigneeHint != "me" {
		t.Fatalf("AssigneeHint = %q, want me", draft.AssigneeHint)
	}
	if !containsString(draft.Labels, "seo") || !containsString(draft.Labels, "ai-created") {
		t.Fatalf("Labels = %#v, want seo and ai-created", draft.Labels)
	}
	if len(draft.Checklist) != 4 {
		t.Fatalf("Checklist = %#v, want 4 items", draft.Checklist)
	}
	if len(draft.MissingFields) != 1 || draft.MissingFields[0]["query"] != "Орион" {
		t.Fatalf("MissingFields = %#v, want missing client Орион", draft.MissingFields)
	}
}

func TestExtractTaskDraftPlainClientAndTimer(t *testing.T) {
	draft := (&AI{}).extractTaskDraft(
		context.Background(),
		uuid.New(),
		uuid.New(),
		"создай задача сделать правки сеошников для Орион и запусти таймер по ней",
		aiIntent{Key: "task.create", Entities: map[string]any{}},
		"2026-06-15",
		"Europe/Moscow",
	)
	if draft.ClientQuery != "Орион" {
		t.Fatalf("ClientQuery = %q, want Орион", draft.ClientQuery)
	}
	if !draft.StartTimer {
		t.Fatal("StartTimer = false, want true")
	}
	if draft.Title != "Внести правки для SEO-специалистов" {
		t.Fatalf("Title = %q", draft.Title)
	}
}

func TestClientSearchSnippetSkipsEmptyQuery(t *testing.T) {
	if got := clientSearchSnippet("", []aiClientMatch{{ID: "client-1", Name: "Орион"}}); got != nil {
		t.Fatalf("clientSearchSnippet(empty) = %#v, want nil", got)
	}
}

func TestPageContextAnswerLabelsSiteChat(t *testing.T) {
	answer := pageContextAnswer(
		[]byte(`{"projectId":"project-1","pageKind":"site_chat","pathname":"/site-chat"}`),
		"Где я сейчас?",
	)
	if !strings.Contains(answer, "Чат сайта / Live Assist") {
		t.Fatalf("answer = %q, want site chat label", answer)
	}
}

func TestDueDateFromTextWithBase(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{text: "сегодня", want: "2026-06-15"},
		{text: "завтра", want: "2026-06-16"},
		{text: "послезавтра", want: "2026-06-17"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := dueDateFromTextWithBase(tt.text, "2026-06-15", "Europe/Moscow"); got != tt.want {
				t.Fatalf("dueDateFromTextWithBase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveProposalEditPatchesToday(t *testing.T) {
	toolCallID := uuid.New()
	edit, ok := (&AI{}).ResolveProposalEdit(context.Background(), uuid.New(), uuid.New(), MessageInput{
		Body:      "давай сегодня",
		LocalDate: "2026-06-15",
		Timezone:  "Europe/Moscow",
	}, SessionState{
		ActiveToolCall: &ToolCallState{
			ID:      toolCallID,
			ToolKey: "task.create",
			Status:  "pending_approval",
		},
	})
	if !ok {
		t.Fatal("expected proposal edit")
	}
	if edit.ToolCallID != toolCallID {
		t.Fatalf("ToolCallID = %s, want %s", edit.ToolCallID, toolCallID)
	}
	if edit.Patch["due_date"] != "2026-06-15" || edit.Patch["due_date_label"] != "сегодня" {
		t.Fatalf("Patch = %#v, want due_date today", edit.Patch)
	}
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
