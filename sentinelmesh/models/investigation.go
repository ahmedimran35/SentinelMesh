package models

import "time"

type InvestigationStatus string

const (
	StatusPending  InvestigationStatus = "pending"
	StatusRunning  InvestigationStatus = "running"
	StatusComplete InvestigationStatus = "complete"
	StatusError    InvestigationStatus = "error"
)

type RiskRating string

const (
	RiskCritical RiskRating = "critical"
	RiskHigh     RiskRating = "high"
	RiskMedium   RiskRating = "medium"
	RiskLow      RiskRating = "low"
	RiskInfo     RiskRating = "info"
)

type Investigation struct {
	ID               string              `json:"id"`
	Target           string              `json:"target"`
	TargetType       string              `json:"target_type"` // "domain", "ip", "ip_range"
	Status           InvestigationStatus `json:"status"`
	CreatedAt        time.Time           `json:"created_at"`
	CompletedAt      *time.Time          `json:"completed_at,omitempty"`
	RiskRating       RiskRating          `json:"risk_rating"`
	ExecutiveSummary string              `json:"executive_summary"`
	FullReport       string              `json:"full_report,omitempty"`
	Findings         []Finding           `json:"findings,omitempty"`
	AgentResults     []AgentResult       `json:"agent_results,omitempty"`
}

type Finding struct {
	ID              string      `json:"id"`
	InvestigationID string      `json:"investigation_id"`
	Agent           string      `json:"agent"`
	Type            string      `json:"type"` // "port", "cve", "ioc", "subdomain", "geo", "campaign"
	Severity        RiskRating  `json:"severity"`
	Title           string      `json:"title"`
	Details         string      `json:"details"`
	RawData         interface{} `json:"raw_data,omitempty"`
	Timestamp       time.Time   `json:"timestamp"`
}

type AgentResult struct {
	ID              string     `json:"id"`
	InvestigationID string     `json:"investigation_id"`
	AgentName       string     `json:"agent_name"`
	Status          string     `json:"status"` // "idle", "scanning", "analyzing", "complete", "error"
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	FindingsCount   int        `json:"findings_count"`
	Error           string     `json:"error,omitempty"`
	DurationMs      int64      `json:"duration_ms"`
}

type IOC struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // "ip", "domain", "hash", "url"
	Value        string    `json:"value"`
	Source       string    `json:"source"` // "threatfox", "urlhaus", "malwarebazaar"
	Confidence   float64   `json:"confidence"`
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	MalwareFamily string   `json:"malware_family,omitempty"`
}

type Target struct {
	ID             string     `json:"id"`
	Value          string     `json:"value"`
	Type           string     `json:"type"`
	MonitorEnabled bool       `json:"monitor_enabled"`
	ScanInterval   string     `json:"scan_interval,omitempty"`
	LastScanned    *time.Time `json:"last_scanned,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Alert struct {
	ID              string    `json:"id"`
	InvestigationID string    `json:"investigation_id,omitempty"`
	Target          string    `json:"target"`
	AlertType       string    `json:"alert_type"` // "new_port", "new_cve", "ioc_match", "campaign", "cert_change", "dns_change"
	Severity        RiskRating `json:"severity"`
	Message         string    `json:"message"`
	Acknowledged    bool      `json:"acknowledged"`
	CreatedAt       time.Time `json:"created_at"`
}

type Report struct {
	Investigation      Investigation `json:"investigation"`
	Findings           []Finding     `json:"findings"`
	IOCs               []IOC         `json:"iocs"`
	SigmaRules         []string      `json:"sigma_rules,omitempty"`
	YARARules          []string      `json:"yara_rules,omitempty"`
	MITRETechniques    []string      `json:"mitre_techniques,omitempty"`
	RecommendedActions []string      `json:"recommended_actions,omitempty"`
}

type NetworkGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"` // "domain", "ip", "asn", "org", "malware", "cve"
	Size  int    `json:"size"`
	Color string `json:"color"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}
