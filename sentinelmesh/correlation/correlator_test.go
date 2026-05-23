package correlation

import (
	"testing"
	"time"

	"sentinelmesh/models"
)

func makeFinding(id, agent, ftype string, severity models.RiskRating, title string) models.Finding {
	return models.Finding{
		ID:        id,
		Agent:     agent,
		Type:      ftype,
		Severity:  severity,
		Title:     title,
		Details:   "test details",
		Timestamp: time.Now(),
	}
}

func TestCVEExploitCorrelation(t *testing.T) {
	c := NewCorrelator()
	findings := []models.Finding{
		makeFinding("f1", "vuln", "cve", models.RiskCritical, "CVE-2024-1234 (9.8 CRITICAL)"),
		makeFinding("f2", "news_intel", "exploit", models.RiskHigh, "GitHub Exploit PoC: user/CVE-2024-1234"),
	}
	correlations := c.Analyze(findings)
	found := false
	for _, corr := range correlations {
		if corr.Type == "cve_exploit" {
			found = true
			if corr.Severity != models.RiskCritical {
				t.Errorf("expected critical severity, got %s", corr.Severity)
			}
		}
	}
	if !found {
		t.Error("expected cve_exploit correlation")
	}
}

func TestIOCAttackSurfaceCorrelation(t *testing.T) {
	c := NewCorrelator()
	findings := []models.Finding{
		makeFinding("f1", "malware", "ioc", models.RiskHigh, "ThreatFox IOC: malware"),
		makeFinding("f2", "recon", "port", models.RiskInfo, "Open Ports on 1.2.3.4"),
	}
	correlations := c.Analyze(findings)
	found := false
	for _, corr := range correlations {
		if corr.Type == "ioc_attack_surface" {
			found = true
		}
	}
	if !found {
		t.Error("expected ioc_attack_surface correlation")
	}
}

func TestCampaignCorrelation(t *testing.T) {
	c := NewCorrelator()
	findings := []models.Finding{
		makeFinding("f1", "recon", "port", models.RiskHigh, "Dangerous port"),
		makeFinding("f2", "vuln", "cve", models.RiskCritical, "CVE-2024-0001"),
		makeFinding("f3", "malware", "ioc", models.RiskHigh, "IOC match"),
	}
	correlations := c.Analyze(findings)
	found := false
	for _, corr := range correlations {
		if corr.Type == "campaign" {
			found = true
			if len(corr.FindingIDs) < 3 {
				t.Errorf("campaign should have 3+ finding IDs, got %d", len(corr.FindingIDs))
			}
		}
	}
	if !found {
		t.Error("expected campaign correlation")
	}
}

func TestDangerousPortCVECorrelation(t *testing.T) {
	c := NewCorrelator()
	findings := []models.Finding{
		makeFinding("f1", "recon", "port", models.RiskHigh, "Open Ports on 1.2.3.4"),
		makeFinding("f2", "vuln", "cve", models.RiskCritical, "CVE-2024-5678"),
	}
	correlations := c.Analyze(findings)
	found := false
	for _, corr := range correlations {
		if corr.Type == "attack_surface" {
			found = true
		}
	}
	if !found {
		t.Error("expected attack_surface correlation")
	}
}

func TestNoCorrelationWithLowSeverity(t *testing.T) {
	c := NewCorrelator()
	findings := []models.Finding{
		makeFinding("f1", "recon", "dns", models.RiskInfo, "DNS Records"),
		makeFinding("f2", "vuln", "cve", models.RiskLow, "CVE-2024-9999"),
	}
	correlations := c.Analyze(findings)
	if len(correlations) != 0 {
		t.Errorf("expected 0 correlations for low-severity findings, got %d", len(correlations))
	}
}

func TestExtractCVEID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"CVE-2024-1234 (9.8 CRITICAL)", "CVE-2024-1234"},
		{"No CVE here", ""},
		{"Found CVE-2023-5555 in title", "CVE-2023-5555"},
	}
	for _, tt := range tests {
		got := extractCVEID(tt.input)
		if got != tt.want {
			t.Errorf("extractCVEID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
