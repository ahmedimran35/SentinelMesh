package models

import "time"

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type AgentStatusEvent struct {
	Agent    string  `json:"agent"`
	State    string  `json:"state"`
	Progress float64 `json:"progress"`
	Message  string  `json:"message"`
}

type FindingEvent struct {
	Agent    string `json:"agent"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Details  string `json:"details"`
}

type InvestigationStatusEvent struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
}

type CommanderMessageEvent struct {
	Message string `json:"message"`
	Type    string `json:"type"` // "analysis", "correlation", "recommendation"
}

type AlertEvent struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Target   string `json:"target"`
}

type ReportReadyEvent struct {
	ID          string     `json:"id"`
	RiskRating  string     `json:"risk_rating"`
	Summary     string     `json:"summary"`
	CompletedAt time.Time  `json:"completed_at"`
}
