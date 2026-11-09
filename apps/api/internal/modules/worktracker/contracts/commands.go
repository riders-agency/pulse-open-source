package contracts

import (
	"time"

	"github.com/google/uuid"
)

type CreateTaskInput struct {
	ProjectID       uuid.UUID
	ClientID        *uuid.UUID
	StatusID        uuid.UUID
	Title           string
	Description     string
	Type            string
	Priority        string
	Severity        string
	AssigneeUserID  *uuid.UUID
	ReviewerUserID  *uuid.UUID
	StartDate       *time.Time
	DueDate         *time.Time
	EstimateMinutes *int
	StoryPoints     *float64
	Labels          []string
	ClientVisible   bool
	Billable        bool
}

type MoveTaskInput struct {
	TaskID           uuid.UUID
	StatusID         uuid.UUID
	Rank             float64
	Comment          string
	ExpectedRevision *int
}

type CreateWorkflowInput struct {
	ProjectID           uuid.UUID
	Name                string
	Description         string
	StatusNames         []string
	TemplateKey         string
	CreateDefaultBoards bool
}
