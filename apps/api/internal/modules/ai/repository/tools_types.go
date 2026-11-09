package repository

import aiproviders "riders.agency/pulse/api/internal/modules/ai/providers"

type aiDependency = aiproviders.ModelDependency

type aiIntent struct {
	Key              string         `json:"key"`
	Confidence       float64        `json:"confidence"`
	Entities         map[string]any `json:"entities"`
	RequiresApproval bool           `json:"requires_approval"`
	CandidateTools   []string       `json:"candidate_tools"`
}

type aiTurnUsage = aiproviders.ModelUsage

type apinetChatMessage = aiproviders.ChatMessage
