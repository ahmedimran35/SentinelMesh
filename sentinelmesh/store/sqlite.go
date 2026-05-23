package store

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"sentinelmesh/models"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	dsn := dbPath
	if !strings.Contains(dsn, "?") {
		dsn += "?"
	} else {
		dsn += "&"
	}
	dsn += "_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS investigations (
			id TEXT PRIMARY KEY,
			target TEXT NOT NULL,
			target_type TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			completed_at DATETIME,
			risk_rating TEXT,
			executive_summary TEXT,
			full_report TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS findings (
			id TEXT PRIMARY KEY,
			investigation_id TEXT NOT NULL,
			agent TEXT NOT NULL,
			type TEXT NOT NULL,
			severity TEXT NOT NULL,
			title TEXT NOT NULL,
			details TEXT,
			raw_data TEXT,
			created_at DATETIME NOT NULL,
			FOREIGN KEY(investigation_id) REFERENCES investigations(id)
		)`,
		`CREATE TABLE IF NOT EXISTS iocs (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			value TEXT NOT NULL,
			source TEXT NOT NULL,
			confidence REAL,
			first_seen DATETIME NOT NULL,
			last_seen DATETIME NOT NULL,
			malware_family TEXT,
			UNIQUE(type, value, source)
		)`,
		`CREATE TABLE IF NOT EXISTS targets (
			id TEXT PRIMARY KEY,
			value TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL,
			monitor_enabled BOOLEAN DEFAULT 0,
			scan_interval TEXT,
			last_scanned DATETIME,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			investigation_id TEXT,
			target TEXT NOT NULL,
			alert_type TEXT NOT NULL,
			severity TEXT NOT NULL,
			message TEXT NOT NULL,
			acknowledged BOOLEAN DEFAULT 0,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS agent_runs (
			id TEXT PRIMARY KEY,
			investigation_id TEXT NOT NULL,
			agent_name TEXT NOT NULL,
			status TEXT NOT NULL,
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			findings_count INTEGER DEFAULT 0,
			error TEXT,
			duration_ms INTEGER
		)`,
		`CREATE INDEX IF NOT EXISTS idx_findings_investigation ON findings(investigation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_iocs_value ON iocs(value)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_target ON alerts(target)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_acknowledged ON alerts(acknowledged)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS rules (
			id TEXT PRIMARY KEY,
			investigation_id TEXT NOT NULL,
			rule_type TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY(investigation_id) REFERENCES investigations(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rules_investigation ON rules(investigation_id)`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// Investigation CRUD
func (s *Store) CreateInvestigation(inv *models.Investigation) error {
	_, err := s.db.Exec(
		`INSERT INTO investigations (id, target, target_type, status, created_at, risk_rating, executive_summary)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		inv.ID, inv.Target, inv.TargetType, inv.Status, inv.CreatedAt, inv.RiskRating, inv.ExecutiveSummary,
	)
	return err
}

func (s *Store) UpdateInvestigation(inv *models.Investigation) error {
	_, err := s.db.Exec(
		`UPDATE investigations SET status=?, completed_at=?, risk_rating=?, executive_summary=?, full_report=?
		 WHERE id=?`,
		inv.Status, inv.CompletedAt, inv.RiskRating, inv.ExecutiveSummary, inv.FullReport, inv.ID,
	)
	return err
}

func (s *Store) GetInvestigation(id string) (*models.Investigation, error) {
	inv := &models.Investigation{}
	var fullReport, execSummary sql.NullString
	var riskRating sql.NullString
	var completedAt sql.NullTime
	err := s.db.QueryRow(
		`SELECT id, target, target_type, status, created_at, completed_at, risk_rating, executive_summary, full_report
		 FROM investigations WHERE id=?`, id,
	).Scan(&inv.ID, &inv.Target, &inv.TargetType, &inv.Status, &inv.CreatedAt, &completedAt, &riskRating, &execSummary, &fullReport)
	if err != nil {
		return nil, err
	}
	if fullReport.Valid {
		inv.FullReport = fullReport.String
	}
	if execSummary.Valid {
		inv.ExecutiveSummary = execSummary.String
	}
	if riskRating.Valid {
		inv.RiskRating = models.RiskRating(riskRating.String)
	}
	if completedAt.Valid {
		inv.CompletedAt = &completedAt.Time
	}
	return inv, nil
}

func (s *Store) ListInvestigations(limit, offset int) ([]models.Investigation, error) {
	rows, err := s.db.Query(
		`SELECT id, target, target_type, status, created_at, completed_at, risk_rating, executive_summary
		 FROM investigations ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invs []models.Investigation
	for rows.Next() {
		var inv models.Investigation
		if err := rows.Scan(&inv.ID, &inv.Target, &inv.TargetType, &inv.Status, &inv.CreatedAt, &inv.CompletedAt, &inv.RiskRating, &inv.ExecutiveSummary); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	return invs, nil
}

func (s *Store) DeleteInvestigation(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM findings WHERE investigation_id=?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM agent_runs WHERE investigation_id=?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM investigations WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

// Findings CRUD
func (s *Store) CreateFinding(f *models.Finding) error {
	rawDataJSON, _ := json.Marshal(f.RawData)
	_, err := s.db.Exec(
		`INSERT INTO findings (id, investigation_id, agent, type, severity, title, details, raw_data, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.ID, f.InvestigationID, f.Agent, f.Type, f.Severity, f.Title, f.Details, string(rawDataJSON), f.Timestamp,
	)
	return err
}

func (s *Store) GetFindings(investigationID string) ([]models.Finding, error) {
	rows, err := s.db.Query(
		`SELECT id, investigation_id, agent, type, severity, title, details, raw_data, created_at
		 FROM findings WHERE investigation_id=? ORDER BY created_at`, investigationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []models.Finding
	for rows.Next() {
		var f models.Finding
		var rawData sql.NullString
		if err := rows.Scan(&f.ID, &f.InvestigationID, &f.Agent, &f.Type, &f.Severity, &f.Title, &f.Details, &rawData, &f.Timestamp); err != nil {
			return nil, err
		}
		if rawData.Valid {
			json.Unmarshal([]byte(rawData.String), &f.RawData)
		}
		findings = append(findings, f)
	}
	return findings, nil
}

func (s *Store) SearchFindings(query, severity, agent string, limit int) ([]models.Finding, error) {
	q := `SELECT id, investigation_id, agent, type, severity, title, details, raw_data, created_at
	      FROM findings WHERE 1=1`
	args := []interface{}{}
	if query != "" {
		q += ` AND (title LIKE ? OR details LIKE ?)`
		args = append(args, "%"+query+"%", "%"+query+"%")
	}
	if severity != "" {
		q += ` AND severity=?`
		args = append(args, severity)
	}
	if agent != "" {
		q += ` AND agent=?`
		args = append(args, agent)
	}
	q += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []models.Finding
	for rows.Next() {
		var f models.Finding
		var rawData sql.NullString
		if err := rows.Scan(&f.ID, &f.InvestigationID, &f.Agent, &f.Type, &f.Severity, &f.Title, &f.Details, &rawData, &f.Timestamp); err != nil {
			return nil, err
		}
		if rawData.Valid {
			json.Unmarshal([]byte(rawData.String), &f.RawData)
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// IOC CRUD
func (s *Store) UpsertIOC(oc *models.IOC) error {
	_, err := s.db.Exec(
		`INSERT INTO iocs (id, type, value, source, confidence, first_seen, last_seen, malware_family)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(type, value, source) DO UPDATE SET last_seen=?, confidence=?, malware_family=?`,
		oc.ID, oc.Type, oc.Value, oc.Source, oc.Confidence, oc.FirstSeen, oc.LastSeen, oc.MalwareFamily,
		oc.LastSeen, oc.Confidence, oc.MalwareFamily,
	)
	return err
}

func (s *Store) GetIOCsByValue(value string) ([]models.IOC, error) {
	rows, err := s.db.Query(
		`SELECT id, type, value, source, confidence, first_seen, last_seen, malware_family
		 FROM iocs WHERE value=?`, value,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var iocs []models.IOC
	for rows.Next() {
		var oc models.IOC
		if err := rows.Scan(&oc.ID, &oc.Type, &oc.Value, &oc.Source, &oc.Confidence, &oc.FirstSeen, &oc.LastSeen, &oc.MalwareFamily); err != nil {
			return nil, err
		}
		iocs = append(iocs, oc)
	}
	return iocs, nil
}

// Alert CRUD
func (s *Store) CreateAlert(a *models.Alert) error {
	_, err := s.db.Exec(
		`INSERT INTO alerts (id, investigation_id, target, alert_type, severity, message, acknowledged, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.InvestigationID, a.Target, a.AlertType, a.Severity, a.Message, a.Acknowledged, a.CreatedAt,
	)
	return err
}

func (s *Store) GetAlerts(unacknowledgedOnly bool) ([]models.Alert, error) {
	q := `SELECT id, investigation_id, target, alert_type, severity, message, acknowledged, created_at FROM alerts`
	if unacknowledgedOnly {
		q += ` WHERE acknowledged=0`
	}
	q += ` ORDER BY created_at DESC LIMIT 100`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var a models.Alert
		if err := rows.Scan(&a.ID, &a.InvestigationID, &a.Target, &a.AlertType, &a.Severity, &a.Message, &a.Acknowledged, &a.CreatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

func (s *Store) AcknowledgeAlert(id string) error {
	_, err := s.db.Exec(`UPDATE alerts SET acknowledged=1 WHERE id=?`, id)
	return err
}

// Agent runs
func (s *Store) CreateAgentRun(ar *models.AgentResult) error {
	_, err := s.db.Exec(
		`INSERT INTO agent_runs (id, investigation_id, agent_name, status, started_at, completed_at, findings_count, error, duration_ms)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ar.ID, ar.InvestigationID, ar.AgentName, ar.Status, ar.StartedAt, ar.CompletedAt, ar.FindingsCount, ar.Error, ar.DurationMs,
	)
	return err
}

func (s *Store) UpdateAgentRun(ar *models.AgentResult) error {
	_, err := s.db.Exec(
		`UPDATE agent_runs SET status=?, completed_at=?, findings_count=?, error=?, duration_ms=? WHERE id=?`,
		ar.Status, ar.CompletedAt, ar.FindingsCount, ar.Error, ar.DurationMs, ar.ID,
	)
	return err
}

// Stats
func (s *Store) GetStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	var totalScans, activeScans, totalFindings, unackAlerts int
	var critFindings, highFindings, medFindings int
	err := s.db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM investigations),
			(SELECT COUNT(*) FROM investigations WHERE status='running'),
			(SELECT COUNT(*) FROM findings),
			(SELECT COUNT(*) FROM alerts WHERE acknowledged=0),
			(SELECT COUNT(*) FROM findings WHERE severity='critical'),
			(SELECT COUNT(*) FROM findings WHERE severity='high'),
			(SELECT COUNT(*) FROM findings WHERE severity='medium')
	`).Scan(&totalScans, &activeScans, &totalFindings, &unackAlerts, &critFindings, &highFindings, &medFindings)
	if err != nil {
		return nil, err
	}

	stats["total_scans"] = totalScans
	stats["active_scans"] = activeScans
	stats["total_findings"] = totalFindings
	stats["unacknowledged_alerts"] = unackAlerts
	stats["critical_findings"] = critFindings
	stats["high_findings"] = highFindings
	stats["medium_findings"] = medFindings

	return stats, nil
}

// Monitored targets
func (s *Store) CreateTarget(t *models.Target) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO targets (id, value, type, monitor_enabled, scan_interval, last_scanned, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Value, t.Type, t.MonitorEnabled, t.ScanInterval, t.LastScanned, t.CreatedAt,
	)
	return err
}

func (s *Store) GetMonitoredTargets() ([]models.Target, error) {
	rows, err := s.db.Query(
		`SELECT id, value, type, monitor_enabled, scan_interval, last_scanned, created_at
		 FROM targets WHERE monitor_enabled=1 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []models.Target
	for rows.Next() {
		var t models.Target
		if err := rows.Scan(&t.ID, &t.Value, &t.Type, &t.MonitorEnabled, &t.ScanInterval, &t.LastScanned, &t.CreatedAt); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func (s *Store) RemoveMonitor(id string) error {
	_, err := s.db.Exec(`UPDATE targets SET monitor_enabled=0 WHERE id=?`, id)
	return err
}

func (s *Store) UpdateTargetLastScanned(id string) error {
	_, err := s.db.Exec(`UPDATE targets SET last_scanned=? WHERE id=?`, time.Now(), id)
	return err
}

// Rule CRUD
func (s *Store) CreateRule(id, investigationID, ruleType, content string) error {
	_, err := s.db.Exec(
		`INSERT INTO rules (id, investigation_id, rule_type, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, investigationID, ruleType, content, time.Now(),
	)
	return err
}

func (s *Store) GetRules(investigationID string) ([]map[string]string, error) {
	rows, err := s.db.Query(
		`SELECT id, rule_type, content FROM rules WHERE investigation_id=? ORDER BY created_at`, investigationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []map[string]string
	for rows.Next() {
		var id, ruleType, content string
		if err := rows.Scan(&id, &ruleType, &content); err != nil {
			return nil, err
		}
		rules = append(rules, map[string]string{
			"id":      id,
			"type":    ruleType,
			"content": content,
		})
	}
	return rules, nil
}

// Settings CRUD
func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value=?, updated_at=?`,
		key, value, time.Now(), value, time.Now(),
	)
	return err
}

func (s *Store) GetSetting(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (s *Store) GetAllSettings() (map[string]string, error) {
	rows, err := s.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}
	return settings, nil
}

func (s *Store) DeleteSetting(key string) error {
	_, err := s.db.Exec(`DELETE FROM settings WHERE key=?`, key)
	return err
}
