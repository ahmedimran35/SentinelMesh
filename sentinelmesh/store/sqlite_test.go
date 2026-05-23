package store

import (
	"fmt"
	"testing"
	"time"

	"sentinelmesh/models"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateAndGetInvestigation(t *testing.T) {
	s := newTestStore(t)
	inv := &models.Investigation{
		ID:         "inv-test-1",
		Target:     "example.com",
		TargetType: "domain",
		Status:     models.StatusRunning,
		CreatedAt:  time.Now(),
	}
	if err := s.CreateInvestigation(inv); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := s.GetInvestigation("inv-test-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Target != "example.com" {
		t.Errorf("target = %q, want %q", got.Target, "example.com")
	}
	if got.Status != models.StatusRunning {
		t.Errorf("status = %q, want %q", got.Status, models.StatusRunning)
	}
}

func TestUpdateInvestigation(t *testing.T) {
	s := newTestStore(t)
	inv := &models.Investigation{
		ID:         "inv-test-2",
		Target:     "example.com",
		TargetType: "domain",
		Status:     models.StatusRunning,
		CreatedAt:  time.Now(),
	}
	s.CreateInvestigation(inv)
	now := time.Now()
	inv.Status = models.StatusComplete
	inv.RiskRating = models.RiskHigh
	inv.ExecutiveSummary = "test summary"
	inv.CompletedAt = &now
	if err := s.UpdateInvestigation(inv); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := s.GetInvestigation("inv-test-2")
	if got.Status != models.StatusComplete {
		t.Errorf("status = %q, want %q", got.Status, models.StatusComplete)
	}
	if got.RiskRating != models.RiskHigh {
		t.Errorf("risk = %q, want %q", got.RiskRating, models.RiskHigh)
	}
}

func TestListInvestigations(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 5; i++ {
		s.CreateInvestigation(&models.Investigation{
			ID:        fmt.Sprintf("inv-%d", i),
			Target:    "test.com",
			Status:    models.StatusComplete,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		})
	}
	invs, err := s.ListInvestigations(3, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(invs) != 3 {
		t.Errorf("got %d investigations, want 3", len(invs))
	}
}

func TestDeleteInvestigation(t *testing.T) {
	s := newTestStore(t)
	s.CreateInvestigation(&models.Investigation{
		ID: "inv-del", Target: "x.com", Status: models.StatusComplete, CreatedAt: time.Now(),
	})
	if err := s.DeleteInvestigation("inv-del"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := s.GetInvestigation("inv-del")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestCreateAndGetFindings(t *testing.T) {
	s := newTestStore(t)
	s.CreateInvestigation(&models.Investigation{
		ID: "inv-f", Target: "x.com", Status: models.StatusRunning, CreatedAt: time.Now(),
	})
	f := &models.Finding{
		ID:              "f-1",
		InvestigationID: "inv-f",
		Agent:           "recon",
		Type:            "port",
		Severity:        models.RiskHigh,
		Title:           "Open port 22",
		Details:         "SSH open",
		Timestamp:       time.Now(),
	}
	if err := s.CreateFinding(f); err != nil {
		t.Fatalf("create finding: %v", err)
	}
	findings, err := s.GetFindings("inv-f")
	if err != nil {
		t.Fatalf("get findings: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}
	if findings[0].Title != "Open port 22" {
		t.Errorf("title = %q", findings[0].Title)
	}
}

func TestSearchFindings(t *testing.T) {
	s := newTestStore(t)
	s.CreateInvestigation(&models.Investigation{
		ID: "inv-s", Target: "x.com", Status: models.StatusComplete, CreatedAt: time.Now(),
	})
	s.CreateFinding(&models.Finding{ID: "f-a", InvestigationID: "inv-s", Agent: "recon", Type: "port", Severity: models.RiskHigh, Title: "SSH open", Details: "port 22", Timestamp: time.Now()})
	s.CreateFinding(&models.Finding{ID: "f-b", InvestigationID: "inv-s", Agent: "vuln", Type: "cve", Severity: models.RiskCritical, Title: "CVE-2024-1234", Details: "rce", Timestamp: time.Now()})

	findings, err := s.SearchFindings("SSH", "", "", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(findings) != 1 {
		t.Errorf("got %d, want 1", len(findings))
	}

	findings, _ = s.SearchFindings("", "critical", "", 10)
	if len(findings) != 1 {
		t.Errorf("severity filter: got %d, want 1", len(findings))
	}

	findings, _ = s.SearchFindings("", "", "recon", 10)
	if len(findings) != 1 {
		t.Errorf("agent filter: got %d, want 1", len(findings))
	}
}

func TestIOC(t *testing.T) {
	s := newTestStore(t)
	ioc := &models.IOC{
		ID: "ioc-1", Type: "ip", Value: "1.2.3.4", Source: "threatfox",
		Confidence: 0.9, FirstSeen: time.Now(), LastSeen: time.Now(),
	}
	if err := s.UpsertIOC(ioc); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	// Upsert again (should not fail)
	s.UpsertIOC(ioc)
	iocs, err := s.GetIOCsByValue("1.2.3.4")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(iocs) != 1 {
		t.Errorf("got %d iocs, want 1", len(iocs))
	}
}

func TestAlerts(t *testing.T) {
	s := newTestStore(t)
	s.CreateAlert(&models.Alert{
		ID: "alert-1", Target: "x.com", AlertType: "ioc_match",
		Severity: models.RiskCritical, Message: "test alert", CreatedAt: time.Now(),
	})
	alerts, err := s.GetAlerts(false)
	if err != nil {
		t.Fatalf("get alerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("got %d alerts, want 1", len(alerts))
	}
	if alerts[0].Acknowledged {
		t.Error("should not be acknowledged")
	}
	s.AcknowledgeAlert("alert-1")
	alerts, _ = s.GetAlerts(true) // unacknowledged only
	if len(alerts) != 0 {
		t.Errorf("after ack, got %d unack, want 0", len(alerts))
	}
}

func TestRules(t *testing.T) {
	s := newTestStore(t)
	s.CreateInvestigation(&models.Investigation{
		ID: "inv-r", Target: "x.com", Status: models.StatusComplete, CreatedAt: time.Now(),
	})
	s.CreateRule("rule-1", "inv-r", "sigma", "title: test\ndetection:\n  selection:\n    title|contains: 'test'")
	s.CreateRule("rule-2", "inv-r", "yara", "rule test { condition: true }")
	rules, err := s.GetRules("inv-r")
	if err != nil {
		t.Fatalf("get rules: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("got %d rules, want 2", len(rules))
	}
	if rules[0]["type"] != "sigma" {
		t.Errorf("first rule type = %q", rules[0]["type"])
	}
}

func TestSettings(t *testing.T) {
	s := newTestStore(t)
	if err := s.SetSetting("test_key", "test_value"); err != nil {
		t.Fatalf("set: %v", err)
	}
	val, err := s.GetSetting("test_key")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if val != "test_value" {
		t.Errorf("got %q, want %q", val, "test_value")
	}
	// Update
	s.SetSetting("test_key", "new_value")
	val, _ = s.GetSetting("test_key")
	if val != "new_value" {
		t.Errorf("after update got %q", val)
	}
	// Get all
	all, _ := s.GetAllSettings()
	if _, ok := all["test_key"]; !ok {
		t.Error("test_key not in all settings")
	}
	// Delete
	s.DeleteSetting("test_key")
	val, _ = s.GetSetting("test_key")
	if val != "" {
		t.Errorf("after delete got %q", val)
	}
}

func TestStats(t *testing.T) {
	s := newTestStore(t)
	s.CreateInvestigation(&models.Investigation{ID: "s1", Target: "x.com", Status: models.StatusComplete, CreatedAt: time.Now()})
	s.CreateInvestigation(&models.Investigation{ID: "s2", Target: "y.com", Status: models.StatusRunning, CreatedAt: time.Now()})
	s.CreateFinding(&models.Finding{ID: "sf1", InvestigationID: "s1", Agent: "recon", Type: "port", Severity: models.RiskCritical, Title: "t", Timestamp: time.Now()})
	s.CreateAlert(&models.Alert{ID: "sa1", Target: "x.com", AlertType: "test", Severity: models.RiskHigh, Message: "m", CreatedAt: time.Now()})

	stats, err := s.GetStats()
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats["total_scans"] != 2 {
		t.Errorf("total_scans = %v", stats["total_scans"])
	}
	if stats["active_scans"] != 1 {
		t.Errorf("active_scans = %v", stats["active_scans"])
	}
	if stats["total_findings"] != 1 {
		t.Errorf("total_findings = %v", stats["total_findings"])
	}
}

func TestMonitoredTargets(t *testing.T) {
	s := newTestStore(t)
	tgt := &models.Target{
		ID: "t1", Value: "example.com", Type: "domain",
		MonitorEnabled: true, ScanInterval: "24h", CreatedAt: time.Now(),
	}
	s.CreateTarget(tgt)
	targets, err := s.GetMonitoredTargets()
	if err != nil {
		t.Fatalf("get monitored: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(targets))
	}
	s.RemoveMonitor("t1")
	targets, _ = s.GetMonitoredTargets()
	if len(targets) != 0 {
		t.Errorf("after remove, got %d", len(targets))
	}
}
