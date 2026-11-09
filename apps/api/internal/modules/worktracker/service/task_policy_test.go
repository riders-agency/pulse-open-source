package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

func TestValidateMovePolicyAllowsProcessRole(t *testing.T) {
	t.Parallel()

	workflowID := uuid.NewString()
	estimate := 30
	err := ValidateMovePolicy(
		workrepo.Access{ProcessRoles: []workrepo.ProcessMembership{{
			TaskWorkflowID: workflowID,
			RoleKey:        "executor",
			Permissions:    map[string]any{"move_tasks": true},
		}}},
		workrepo.Task{WorkflowID: workflowID, AssigneeUserID: uuid.NewString(), EstimateMinutes: &estimate},
		TransitionRule{AllowedRoles: []string{"executor"}, RequiredFields: []string{"assignee_user_id", "estimate_minutes"}},
		"",
	)
	if err != nil {
		t.Fatalf("validate move policy: %v", err)
	}
}

func TestValidateMovePolicyDeniesWrongTransitionRole(t *testing.T) {
	t.Parallel()

	workflowID := uuid.NewString()
	err := ValidateMovePolicy(
		workrepo.Access{ProcessRoles: []workrepo.ProcessMembership{{
			TaskWorkflowID: workflowID,
			RoleKey:        "executor",
			Permissions:    map[string]any{"move_tasks": true},
		}}},
		workrepo.Task{WorkflowID: workflowID},
		TransitionRule{AllowedRoles: []string{"reviewer"}},
		"",
	)
	if !errors.Is(err, workrepo.ErrTransitionDenied) {
		t.Fatalf("validate move policy error = %v, want transition denied", err)
	}
}

func TestValidateMoveRequirements(t *testing.T) {
	t.Parallel()

	estimate := 60
	task := workrepo.Task{AssigneeUserID: uuid.NewString(), EstimateMinutes: &estimate}
	if err := ValidateMoveRequirements(task, []string{"assignee_user_id", "estimate_minutes", "comment"}, "ready"); err != nil {
		t.Fatalf("validate complete requirements: %v", err)
	}
	if err := ValidateMoveRequirements(workrepo.Task{}, []string{"assignee_user_id"}, "ready"); !errors.Is(err, workrepo.ErrValidation) {
		t.Fatalf("missing assignee error = %v, want validation", err)
	}
	if err := ValidateMoveRequirements(task, []string{"comment"}, " "); !errors.Is(err, workrepo.ErrValidation) {
		t.Fatalf("missing comment error = %v, want validation", err)
	}
}

func TestTransitionRuleCarriesTimerAutomation(t *testing.T) {
	t.Parallel()

	rule := TransitionRule{Automation: TransitionAutomation{StartTimer: true, StopTimer: true, SetClientVisible: true}}
	if !rule.Automation.StartTimer || !rule.Automation.StopTimer || !rule.Automation.SetClientVisible {
		t.Fatalf("timer automation flags were not preserved: %#v", rule.Automation)
	}
}

func TestExpectedRevision(t *testing.T) {
	t.Parallel()

	if got := expectedRevision(7, nil); got != 7 {
		t.Fatalf("expected revision without client token = %d, want 7", got)
	}
	clientRevision := 3
	if got := expectedRevision(7, &clientRevision); got != 3 {
		t.Fatalf("expected revision with client token = %d, want 3", got)
	}
}
