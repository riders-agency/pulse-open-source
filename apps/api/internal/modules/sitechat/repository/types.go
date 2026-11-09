package repository

import (
	"time"
)

type WidgetRow struct {
	ChatWidgetID         string         `json:"chat_widget_id"`
	CompanyID            string         `json:"company_id"`
	ProjectID            string         `json:"project_id"`
	PublicKey            string         `json:"public_key"`
	Name                 string         `json:"name"`
	Status               string         `json:"status"`
	Title                string         `json:"title"`
	Subtitle             string         `json:"subtitle"`
	AccentColor          string         `json:"accent_color"`
	Position             string         `json:"position"`
	HeavyTrackingDefault bool           `json:"heavy_tracking_default"`
	SamplingRate         float64        `json:"sampling_rate"`
	Settings             map[string]any `json:"settings_json"`
	EmbedScript          string         `json:"embed_script,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type WidgetInput struct {
	Name                 string         `json:"name"`
	Title                string         `json:"title"`
	Subtitle             string         `json:"subtitle"`
	AccentColor          string         `json:"accent_color"`
	Position             string         `json:"position"`
	HeavyTrackingDefault bool           `json:"heavy_tracking_default"`
	SamplingRate         float64        `json:"sampling_rate"`
	Settings             map[string]any `json:"settings_json"`
}

type WidgetPatch struct {
	Name                 *string         `json:"name"`
	Status               *string         `json:"status"`
	Title                *string         `json:"title"`
	Subtitle             *string         `json:"subtitle"`
	AccentColor          *string         `json:"accent_color"`
	Position             *string         `json:"position"`
	HeavyTrackingDefault *bool           `json:"heavy_tracking_default"`
	SamplingRate         *float64        `json:"sampling_rate"`
	Settings             *map[string]any `json:"settings_json"`
}

type PublicWidgetConfig struct {
	ChatWidgetID         string         `json:"chat_widget_id"`
	ProjectID            string         `json:"project_id"`
	PublicKey            string         `json:"public_key"`
	Title                string         `json:"title"`
	Subtitle             string         `json:"subtitle"`
	AccentColor          string         `json:"accent_color"`
	Position             string         `json:"position"`
	HeavyTrackingDefault bool           `json:"heavy_tracking_default"`
	SamplingRate         float64        `json:"sampling_rate"`
	EndpointBase         string         `json:"endpoint_base"`
	Settings             map[string]any `json:"settings_json"`
	Voice                map[string]any `json:"voice,omitempty"`
}

type PresenceInput struct {
	PublicKey        string         `json:"public_key"`
	AnonymousID      string         `json:"anonymous_id"`
	VisitSessionID   string         `json:"visit_session_id"`
	URL              string         `json:"url"`
	Title            string         `json:"title"`
	Referrer         string         `json:"referrer"`
	UTM              map[string]any `json:"utm"`
	ViewportWidth    int            `json:"viewport_width"`
	ViewportHeight   int            `json:"viewport_height"`
	ScrollX          int            `json:"scroll_x"`
	ScrollY          int            `json:"scroll_y"`
	DurationSeconds  int            `json:"duration_seconds"`
	Browser          string         `json:"browser"`
	BrowserVersion   string         `json:"browser_version"`
	OS               string         `json:"os"`
	DeviceType       string         `json:"device_type"`
	ScreenWidth      int            `json:"screen_width"`
	ScreenHeight     int            `json:"screen_height"`
	Language         string         `json:"language"`
	Timezone         string         `json:"timezone"`
	MetrikaClientID  string         `json:"metrika_client_id"`
	MetrikaUserID    string         `json:"metrika_user_id"`
	MetrikaSessionID string         `json:"metrika_session_id"`
	UserAgent        string         `json:"user_agent"`
	ViewportContext  map[string]any `json:"viewport_context"`
	IP               string         `json:"-"`
	Metadata         map[string]any `json:"metadata"`
}

type PresenceResponse struct {
	VisitorID            string         `json:"visitor_id"`
	VisitSessionID       string         `json:"visit_session_id"`
	ChatWidgetID         string         `json:"chat_widget_id"`
	ProjectID            string         `json:"project_id"`
	ConversationID       string         `json:"conversation_id"`
	ConversationURL      string         `json:"conversation_url"`
	PanelURL             string         `json:"panel_url"`
	CallURL              string         `json:"call_url"`
	IsFirstVisit         bool           `json:"is_first_visit"`
	MonitoringStatus     string         `json:"monitoring_status"`
	Recorder             map[string]any `json:"recorder,omitempty"`
	Voice                map[string]any `json:"voice,omitempty"`
	OperatorInvited      bool           `json:"operator_invited"`
	VisitorStateSnapshot map[string]any `json:"visitor_state_snapshot"`
}

type EventInput struct {
	EventType  string         `json:"event_type"`
	OccurredAt string         `json:"occurred_at"`
	Payload    map[string]any `json:"payload"`
	Priority   string         `json:"priority"`
}

type EventBatchInput struct {
	PublicKey      string       `json:"public_key"`
	VisitorID      string       `json:"visitor_id"`
	VisitSessionID string       `json:"visit_session_id"`
	Events         []EventInput `json:"events"`
}

type VisitorSessionRow struct {
	VisitorID            string         `json:"visitor_id"`
	VisitSessionID       string         `json:"visit_session_id"`
	ChatWidgetID         string         `json:"chat_widget_id"`
	ProjectID            string         `json:"project_id"`
	ConversationID       string         `json:"conversation_id"`
	VisitorName          string         `json:"visitor_name"`
	VisitorContact       string         `json:"visitor_contact"`
	ConversationURL      string         `json:"conversation_url"`
	PanelURL             string         `json:"panel_url"`
	CallURL              string         `json:"call_url"`
	CurrentURL           string         `json:"current_url"`
	CurrentTitle         string         `json:"current_title"`
	Status               string         `json:"status"`
	StartedAt            time.Time      `json:"started_at"`
	EndedAt              *time.Time     `json:"ended_at,omitempty"`
	LastActivityAt       time.Time      `json:"last_activity_at"`
	DurationSeconds      int            `json:"duration_seconds"`
	VisitorMessageCount  int            `json:"visitor_message_count"`
	MonitoringStatus     string         `json:"monitoring_status"`
	OperatorInvited      bool           `json:"operator_invited"`
	ChatOpen             bool           `json:"chat_open"`
	VisitorStateSnapshot map[string]any `json:"visitor_state_snapshot"`
}

type ConversationRow struct {
	ConversationID string    `json:"conversation_id"`
	ProjectID      string    `json:"project_id"`
	ChatWidgetID   string    `json:"chat_widget_id"`
	VisitorID      string    `json:"visitor_id"`
	VisitSessionID string    `json:"visit_session_id,omitempty"`
	Status         string    `json:"status"`
	Title          string    `json:"title"`
	VisitorName    string    `json:"visitor_name"`
	VisitorContact string    `json:"visitor_contact"`
	OperatorURL    string    `json:"operator_url"`
	CallURL        string    `json:"call_url"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MessageRow struct {
	MessageID      string          `json:"message_id"`
	ConversationID string          `json:"conversation_id"`
	AuthorKind     string          `json:"author_kind"`
	AuthorUserID   string          `json:"author_user_id,omitempty"`
	AuthorName     string          `json:"author_name"`
	Body           string          `json:"body"`
	Attachments    []AttachmentRow `json:"attachments,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type MessageInput struct {
	PublicKey      string   `json:"public_key"`
	ConversationID string   `json:"conversation_id"`
	VisitSessionID string   `json:"visit_session_id"`
	VisitorID      string   `json:"visitor_id"`
	AuthorName     string   `json:"author_name"`
	VisitorContact string   `json:"visitor_contact"`
	Body           string   `json:"body"`
	AttachmentIDs  []string `json:"attachment_ids"`
}
