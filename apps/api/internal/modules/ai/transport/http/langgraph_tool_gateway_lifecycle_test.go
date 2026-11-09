package http

import "testing"

func TestApprovalReplyUsesExecutionResultInsteadOfPreviousProposal(t *testing.T) {
	for _, result := range []string{"Задача создана: Презентация", "Действие не выполнено: недостаточно прав"} {
		session := map[string]any{"messages": []map[string]any{
			{"role": "assistant", "body": "Подготовил карточку задачи"},
			{"role": "tool", "body": result},
		}}
		if got := approvalCompletionObservation(session)["message"]; got != result {
			t.Fatalf("approval reply = %q, want current execution result %q", got, result)
		}
	}
}

func TestApprovalWithoutToolResultDoesNotReplayPreviousAssistant(t *testing.T) {
	session := map[string]any{"messages": []map[string]any{{"role": "assistant", "body": "Подготовил карточку задачи"}}}
	if got := approvalCompletionObservation(session)["message"]; got != "Действие обработано." {
		t.Fatalf("unexpected reply: %q", got)
	}
}
