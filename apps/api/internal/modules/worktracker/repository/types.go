package repository

import (
	"errors"
	"time"

	"riders.agency/pulse/api/internal/modules/worktracker/contracts"

	"github.com/google/uuid"
	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

var (
	ErrForbidden        = contracts.ErrForbidden
	ErrNotFound         = errors.New("work tracker not found")
	ErrTaskTimerActive  = errors.New("task has an active timer")
	ErrConflict         = errors.New("work tracker conflict")
	ErrValidation       = errors.New("work tracker validation")
	ErrSecondTimer      = errors.New("active timer already exists")
	ErrTransitionDenied = errors.New("transition denied")
)

type WorkTracker struct {
	P          dbpkg.DB
	TaskEvents TaskEventPublisher
}

type Access struct {
	CompanyID          uuid.UUID           `json:"-"`
	ProjectID          uuid.UUID           `json:"-"`
	UserID             uuid.UUID           `json:"-"`
	Role               string              `json:"role"`
	RoleName           string              `json:"role_name"`
	ClientID           *uuid.UUID          `json:"-"`
	ProjectName        string              `json:"project_name"`
	CompanyName        string              `json:"company_name"`
	CanRead            bool                `json:"can_read"`
	CanWriteTasks      bool                `json:"can_write_tasks"`
	CanManageProcesses bool                `json:"can_manage_processes"`
	CanManageMembers   bool                `json:"can_manage_members"`
	CanApproveTime     bool                `json:"can_approve_time"`
	CanSeeReports      bool                `json:"can_see_reports"`
	IsClientGuest      bool                `json:"is_client_guest"`
	VisibleModules     []string            `json:"visible_modules"`
	WorkflowIDs        []string            `json:"task_workflow_ids"`
	ProcessRoles       []ProcessMembership `json:"process_roles"`
	AllWorkflows       bool                `json:"all_workflows"`
}

type Overview struct {
	TodayTasks         int `json:"today_tasks"`
	OverdueTasks       int `json:"overdue_tasks"`
	UnassignedTasks    int `json:"unassigned_tasks"`
	UnestimatedTasks   int `json:"unestimated_tasks"`
	BlockedTasks       int `json:"blocked_tasks"`
	ReviewTasks        int `json:"review_tasks"`
	ClientAcceptance   int `json:"client_acceptance_tasks"`
	LoggedTodayMinutes int `json:"logged_today_minutes"`
	LoggedWeekMinutes  int `json:"logged_week_minutes"`
}

type Workflow struct {
	TaskWorkflowID string        `json:"task_workflow_id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Scope          string        `json:"scope"`
	IsDefault      bool          `json:"is_default"`
	Statuses       []Status      `json:"statuses,omitempty"`
	Roles          []ProcessRole `json:"roles,omitempty"`
}

type ProcessRole struct {
	TaskProcessRoleID string         `json:"task_process_role_id"`
	TaskWorkflowID    string         `json:"task_workflow_id"`
	RoleKey           string         `json:"role_key"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	Permissions       map[string]any `json:"permissions"`
	Position          int            `json:"position"`
	IsSystem          bool           `json:"is_system"`
}

type ProcessMembership struct {
	TaskProcessMembershipID string         `json:"task_process_membership_id,omitempty"`
	TaskWorkflowID          string         `json:"task_workflow_id"`
	WorkflowName            string         `json:"workflow_name,omitempty"`
	TaskProcessRoleID       string         `json:"task_process_role_id"`
	RoleKey                 string         `json:"role_key"`
	RoleName                string         `json:"role_name"`
	Permissions             map[string]any `json:"permissions"`
	ClientID                string         `json:"client_id,omitempty"`
}

type WorkflowTemplate struct {
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	TaskTypes   []string          `json:"task_types"`
	Statuses    []TemplateStatus  `json:"statuses"`
	Transitions []TemplateMove    `json:"transitions"`
	Boards      []BoardTemplate   `json:"boards"`
	Fields      []CustomFieldSpec `json:"fields"`
}

type TemplateStatus struct {
	Name             string `json:"name"`
	Category         string `json:"category"`
	Color            string `json:"color"`
	WIPLimit         *int   `json:"wip_limit,omitempty"`
	RequiresAssignee bool   `json:"requires_assignee"`
	RequiresEstimate bool   `json:"requires_estimate"`
	AutoStartTimer   bool   `json:"auto_start_timer"`
	AutoStopTimer    bool   `json:"auto_stop_timer"`
}

type TemplateMove struct {
	From           string   `json:"from"`
	To             string   `json:"to"`
	Name           string   `json:"name"`
	AllowedRoles   []string `json:"allowed_roles"`
	RequiredFields []string `json:"required_fields"`
	AutomationJSON string   `json:"automation_json"`
}

type BoardTemplate struct {
	Key           string   `json:"key"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Type          string   `json:"type"`
	GroupBy       string   `json:"group_by"`
	SwimlaneBy    string   `json:"swimlane_by"`
	VisibleFields []string `json:"visible_fields"`
	Density       string   `json:"density"`
}

type CustomFieldSpec struct {
	Name               string   `json:"name"`
	Key                string   `json:"key"`
	Type               string   `json:"type"`
	Required           bool     `json:"required"`
	Options            []string `json:"options,omitempty"`
	VisibleByDefault   bool     `json:"visible_by_default"`
	AppliesToTaskTypes []string `json:"applies_to_task_types,omitempty"`
}

type Status struct {
	TaskWorkflowStatusID string `json:"task_workflow_status_id"`
	Name                 string `json:"name"`
	Category             string `json:"category"`
	Color                string `json:"color"`
	Position             int    `json:"position"`
	IsStart              bool   `json:"is_start"`
	IsDone               bool   `json:"is_done"`
	WIPLimit             *int   `json:"wip_limit,omitempty"`
	RequiresAssignee     bool   `json:"requires_assignee"`
	RequiresEstimate     bool   `json:"requires_estimate"`
}

type Board struct {
	TaskBoardID   string         `json:"task_board_id"`
	Name          string         `json:"name"`
	Description   string         `json:"description,omitempty"`
	Type          string         `json:"type"`
	WorkflowID    string         `json:"task_workflow_id,omitempty"`
	GroupBy       string         `json:"group_by"`
	SwimlaneBy    string         `json:"swimlane_by"`
	VisibleFields []string       `json:"visible_fields"`
	CardLayout    map[string]any `json:"card_layout,omitempty"`
	CanCustomize  bool           `json:"can_customize"`
	Columns       []BoardColumn  `json:"columns,omitempty"`
	Summary       map[string]int `json:"summary,omitempty"`
}

type BoardColumn struct {
	Status Status `json:"status"`
	Count  int    `json:"count"`
	Tasks  []Task `json:"tasks"`
}

type Task struct {
	TaskID              string     `json:"task_id"`
	TaskKey             string     `json:"task_key"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	Type                string     `json:"type"`
	StatusID            string     `json:"task_workflow_status_id"`
	WorkflowID          string     `json:"task_workflow_id"`
	StatusName          string     `json:"status_name"`
	StatusCategory      string     `json:"status_category"`
	Priority            string     `json:"priority"`
	Severity            string     `json:"severity"`
	AssigneeUserID      string     `json:"assignee_user_id,omitempty"`
	AssigneeName        string     `json:"assignee_name,omitempty"`
	ReporterUserID      string     `json:"reporter_user_id,omitempty"`
	ReviewerUserID      string     `json:"reviewer_user_id,omitempty"`
	ClientID            string     `json:"client_id,omitempty"`
	ClientName          string     `json:"client_name,omitempty"`
	ProjectID           string     `json:"project_id,omitempty"`
	ProjectName         string     `json:"project_name,omitempty"`
	StartDate           *time.Time `json:"start_date,omitempty"`
	DueDate             *time.Time `json:"due_date,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	EstimateMinutes     *int       `json:"estimate_minutes,omitempty"`
	SpentMinutesCached  int        `json:"spent_minutes_cached"`
	RemainingMinutes    *int       `json:"remaining_minutes,omitempty"`
	StoryPoints         *float64   `json:"story_points,omitempty"`
	Rank                float64    `json:"rank"`
	IsBlocked           bool       `json:"is_blocked"`
	BlockedReason       string     `json:"blocked_reason,omitempty"`
	ClientVisible       bool       `json:"client_visible"`
	Billable            bool       `json:"billable"`
	Labels              []string   `json:"labels"`
	CommentsCount       int        `json:"comments_count"`
	AttachmentsCount    int        `json:"attachments_count"`
	ChecklistDoneCount  int        `json:"checklist_done_count"`
	ChecklistTotalCount int        `json:"checklist_total_count"`
	Revision            int        `json:"revision"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type TaskDetail struct {
	Task
	Comments     []TaskComment    `json:"comments"`
	Checklist    []ChecklistItem  `json:"checklist"`
	CustomFields []CustomField    `json:"custom_fields"`
	Activity     []ActivityEvent  `json:"activity"`
	Time         []TimeEntry      `json:"time_entries"`
	Attachments  []TaskAttachment `json:"attachments"`
}

type TaskComment struct {
	TaskCommentID string    `json:"task_comment_id"`
	AuthorName    string    `json:"author_name"`
	Body          string    `json:"body"`
	Visibility    string    `json:"visibility"`
	CreatedAt     time.Time `json:"created_at"`
}

type ChecklistItem struct {
	TaskChecklistItemID string     `json:"task_checklist_item_id"`
	Title               string     `json:"title"`
	IsDone              bool       `json:"is_done"`
	Position            int        `json:"position"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
}

type CustomField struct {
	TaskCustomFieldID string   `json:"task_custom_field_id"`
	Name              string   `json:"name"`
	Key               string   `json:"key"`
	Type              string   `json:"type"`
	Required          bool     `json:"required"`
	Options           []string `json:"options"`
	Value             any      `json:"value"`
	VisibleByDefault  bool     `json:"visible_by_default"`
}

type ActivityEvent struct {
	EventType string    `json:"event_type"`
	ActorName string    `json:"actor_name"`
	Payload   []byte    `json:"payload_json"`
	CreatedAt time.Time `json:"created_at"`
}

type TimeEntry struct {
	TimeEntryID     string     `json:"time_entry_id"`
	TaskID          string     `json:"task_id,omitempty"`
	TaskKey         string     `json:"task_key,omitempty"`
	TaskTitle       string     `json:"task_title,omitempty"`
	ProjectID       string     `json:"project_id,omitempty"`
	ProjectName     string     `json:"project_name,omitempty"`
	ClientID        string     `json:"client_id,omitempty"`
	ClientName      string     `json:"client_name,omitempty"`
	UserID          string     `json:"user_id"`
	UserName        string     `json:"user_name"`
	EntryDate       time.Time  `json:"entry_date"`
	DurationMinutes int        `json:"duration_minutes"`
	Description     string     `json:"description"`
	Billable        bool       `json:"billable"`
	Source          string     `json:"source"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
}

type TimerSession struct {
	TimerSessionID string    `json:"timer_session_id"`
	TaskID         string    `json:"task_id"`
	TaskKey        string    `json:"task_key"`
	TaskTitle      string    `json:"task_title"`
	StartedAt      time.Time `json:"started_at"`
	State          string    `json:"state"`
	ElapsedMinutes int       `json:"elapsed_minutes"`
}

type TaskAttachment struct {
	TaskAttachmentID string    `json:"task_attachment_id"`
	TaskID           string    `json:"task_id"`
	FileName         string    `json:"file_name"`
	ContentType      string    `json:"content_type"`
	SizeBytes        int64     `json:"size_bytes"`
	UploadedByName   string    `json:"uploaded_by_name"`
	DownloadURL      string    `json:"download_url"`
	CreatedAt        time.Time `json:"created_at"`
}

type Employee struct {
	EmployeeProfileID   string              `json:"employee_profile_id"`
	UserID              string              `json:"user_id"`
	FullName            string              `json:"full_name"`
	Email               string              `json:"email"`
	Position            string              `json:"position"`
	Department          string              `json:"department"`
	Role                string              `json:"role"`
	EmploymentType      string              `json:"employment_type"`
	Status              string              `json:"status"`
	CapacityHoursWeek   int                 `json:"capacity_hours_per_week"`
	Skills              []string            `json:"skills"`
	TasksInProgress     int                 `json:"tasks_in_progress"`
	OverdueTasks        int                 `json:"overdue_tasks"`
	LoggedWeekMinutes   int                 `json:"logged_week_minutes"`
	BillableWeekMinutes int                 `json:"billable_week_minutes"`
	ProcessAccesses     []ProcessMembership `json:"process_accesses,omitempty"`
}

type Report struct {
	CreatedTasks    int            `json:"created_tasks"`
	CompletedTasks  int            `json:"completed_tasks"`
	OverdueTasks    int            `json:"overdue_tasks"`
	BlockedTasks    int            `json:"blocked_tasks"`
	LoggedMinutes   int            `json:"logged_minutes"`
	BillableMinutes int            `json:"billable_minutes"`
	ByClient        []ReportBucket `json:"by_client"`
	ByProject       []ReportBucket `json:"by_project"`
	ByEmployee      []ReportBucket `json:"by_employee"`
	ByStatus        []ReportBucket `json:"by_status"`
}

type ReportBucket struct {
	ID              string `json:"id,omitempty"`
	Name            string `json:"name"`
	Tasks           int    `json:"tasks"`
	LoggedMinutes   int    `json:"logged_minutes"`
	BillableMinutes int    `json:"billable_minutes"`
}

type TaskListInput struct {
	Archived   bool
	ProjectID  uuid.UUID
	Search     string
	WorkflowID *uuid.UUID
	StatusID   *uuid.UUID
	Assignee   string
	Type       string
	Priority   string
	ClientID   *uuid.UUID
	Due        string
	Limit      int
	Offset     int
}

type TaskList struct {
	Items  []Task `json:"items"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type CreateTaskInput = contracts.CreateTaskInput

type MoveTaskInput = contracts.MoveTaskInput

type PatchTaskInput struct {
	TaskID           uuid.UUID
	Title            *string
	Description      *string
	Type             *string
	Priority         *string
	Severity         *string
	AssigneeUserID   *uuid.UUID
	ReviewerUserID   *uuid.UUID
	ClientID         *uuid.UUID
	ClientIDSet      bool
	StartDate        *time.Time
	DueDate          *time.Time
	EstimateMinutes  *int
	StoryPoints      *float64
	Labels           []string
	LabelsSet        bool
	ClientVisible    *bool
	Billable         *bool
	IsBlocked        *bool
	BlockedReason    *string
	ExpectedRevision *int
}

type AddCommentInput struct {
	TaskID     uuid.UUID
	Body       string
	Visibility string
}

type AddChecklistInput struct {
	TaskID uuid.UUID
	Title  string
}

type PatchChecklistInput struct {
	TaskID uuid.UUID
	ItemID uuid.UUID
	Title  *string
	IsDone *bool
}

type UpsertCustomFieldValueInput struct {
	TaskID  uuid.UUID
	FieldID uuid.UUID
	Value   any
}

type ManualTimeInput struct {
	TaskID          uuid.UUID
	DurationMinutes int
	Description     string
	EntryDate       time.Time
	Billable        bool
}

type InviteMemberInput struct {
	ProjectID   uuid.UUID
	WorkflowID  uuid.UUID
	Email       string
	DisplayName string
	Role        string
	RoleID      uuid.UUID
	ClientID    *uuid.UUID
}

type CreateWorkflowInput = contracts.CreateWorkflowInput

type UpsertProcessRoleInput struct {
	WorkflowID  uuid.UUID
	RoleID      uuid.UUID
	Name        string
	Description string
	Permissions map[string]any
	Position    int
}

type CreateBoardInput struct {
	ProjectID     uuid.UUID
	WorkflowID    uuid.UUID
	Name          string
	Type          string
	GroupBy       string
	SwimlaneBy    string
	VisibleFields []string
	CardLayout    map[string]any
	Filters       map[string]any
	Sort          map[string]any
	TemplateKey   string
	IsDefault     bool
}

func (w *WorkTracker) pool() dbpkg.DB {
	if w == nil {
		return nil
	}
	return w.P
}
