package service

import (
	"strings"

	workrepo "riders.agency/pulse/api/internal/modules/worktracker/repository"
)

type TransitionAutomation struct {
	SetClientVisible bool `json:"set_client_visible"`
	StartTimer       bool `json:"start_timer"`
	StopTimer        bool `json:"stop_timer"`
}

type TransitionRule struct {
	AllowedRoles   []string
	RequiredFields []string
	Automation     TransitionAutomation
}

func ValidateMovePolicy(access workrepo.Access, task workrepo.Task, rule TransitionRule, comment string) error {
	if !access.CanMoveTaskInWorkflow(task.WorkflowID) {
		return workrepo.ErrForbidden
	}
	if !access.HasTransitionRole(task.WorkflowID, rule.AllowedRoles) {
		return workrepo.ErrTransitionDenied
	}
	return ValidateMoveRequirements(task, rule.RequiredFields, comment)
}

func ValidateMoveRequirements(task workrepo.Task, fields []string, comment string) error {
	for _, field := range fields {
		switch field {
		case "assignee_user_id":
			if task.AssigneeUserID == "" {
				return workrepo.ErrValidation
			}
		case "estimate_minutes":
			if task.EstimateMinutes == nil {
				return workrepo.ErrValidation
			}
		case "comment":
			if strings.TrimSpace(comment) == "" {
				return workrepo.ErrValidation
			}
		}
	}
	return nil
}
