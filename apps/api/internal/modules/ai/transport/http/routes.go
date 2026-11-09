package http

import (
	"github.com/go-chi/chi/v5"

	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
)

func MountReadinessRoutes(r chi.Router) {
	readiness := RuntimeReadiness()
	r.Get("/ai/runtime/readiness", readiness)
	r.Get("/ai/agent/readiness", readiness)
	r.Get("/ai/runtime/capabilities", readiness)
}

func MountProjectRoutes(r chi.Router, ai *aisvc.Commands) {
	r.Get("/ai/model-providers", ModelProviders())
	r.Get("/ai/employee", GetEmployee(ai))
	r.Get("/ai/employee/capability", EmployeeCapability())
	r.Patch("/ai/employee", PatchEmployee(ai))
	r.Post("/ai/employee/sessions", CreateSession(ai))
	r.Get("/ai/employee/sessions/{session_id}", GetSession(ai))
	r.Post("/ai/employee/sessions/{session_id}/messages", PostMessage(ai))
	r.Post("/ai/employee/sessions/{session_id}/messages/stream", PostMessageStream(ai))
	r.Post("/ai/employee/sessions/{session_id}/cancel", CancelSession(ai))
	r.Patch("/ai/tool-calls/{tool_call_id}", UpdateToolCall(ai))
	r.Post("/ai/tool-calls/{tool_call_id}/approve", ApproveToolCall(ai))
	r.Post("/ai/tool-calls/{tool_call_id}/reject", RejectToolCall(ai))
}

func MountActionDependencyModelRoutes(r chi.Router, ai *aisvc.Commands) {
	r.Get("/action-dependencies/{action_dependency_id}/models", ListModels(ai))
	r.Post("/action-dependencies/{action_dependency_id}/models/refresh", RefreshModels(ai))
	r.Post("/action-dependencies/{action_dependency_id}/models/evaluate", EvaluateModels(ai))
}
