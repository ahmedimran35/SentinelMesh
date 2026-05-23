package correlation

import (
	"strings"

	"sentinelmesh/models"
)

// Correlator cross-references findings across agents to detect patterns
type Correlator struct{}

func NewCorrelator() *Correlator {
	return &Correlator{}
}

// Correlation represents a detected pattern across findings
type Correlation struct {
	Type        string   `json:"type"` // "cve_exploit", "campaign", "attack_surface", "ioc_match"
	Severity    models.RiskRating `json:"severity"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	FindingIDs  []string `json:"finding_ids"`
}

// Analyze finds correlations between findings from different agents
func (c *Correlator) Analyze(findings []models.Finding) []Correlation {
	var correlations []Correlation

	// Group by agent
	byAgent := make(map[string][]models.Finding)
	for _, f := range findings {
		byAgent[f.Agent] = append(byAgent[f.Agent], f)
	}

	// Correlation 1: CVE + Exploit PoC = Critical
	reconFindings := byAgent["recon"]
	vulnFindings := byAgent["vuln"]
	newsFindings := byAgent["news_intel"]

	for _, vf := range vulnFindings {
		if vf.Severity == models.RiskCritical || vf.Severity == models.RiskHigh {
			// Check if there's an exploit PoC for this CVE
			for _, nf := range newsFindings {
				if nf.Type == "exploit" && strings.Contains(nf.Title, extractCVEID(vf.Title)) {
					correlations = append(correlations, Correlation{
						Type:        "cve_exploit",
						Severity:    models.RiskCritical,
						Title:       "Active Exploit Available: " + vf.Title,
						Description: "CVE has known exploit PoC on GitHub — immediate patching required",
						FindingIDs:  []string{vf.ID, nf.ID},
					})
				}
			}
		}
	}

	// Correlation 2: IOC match + Open ports = Attack surface
	malwareFindings := byAgent["malware"]
	for _, mf := range malwareFindings {
		if mf.Type == "ioc" {
			for _, rf := range reconFindings {
				if rf.Type == "port" {
					correlations = append(correlations, Correlation{
						Type:        "ioc_attack_surface",
						Severity:    models.RiskCritical,
						Title:       "IOC Match on Exposed Infrastructure",
						Description: "Known malware IOC found on target with open ports — active compromise likely",
						FindingIDs:  []string{mf.ID, rf.ID},
					})
				}
			}
		}
	}

	// Correlation 3: Multiple high-severity findings = Campaign
	highCount := 0
	for _, f := range findings {
		if f.Severity == models.RiskCritical || f.Severity == models.RiskHigh {
			highCount++
		}
	}
	if highCount >= 3 {
		ids := make([]string, 0)
		for _, f := range findings {
			if f.Severity == models.RiskCritical || f.Severity == models.RiskHigh {
				ids = append(ids, f.ID)
			}
		}
		correlations = append(correlations, Correlation{
			Type:        "campaign",
			Severity:    models.RiskCritical,
			Title:       "Multiple Critical Findings — Possible Campaign",
			Description: "3+ high/critical findings detected — may indicate coordinated attack infrastructure",
			FindingIDs:  ids,
		})
	}

	// Correlation 4: Dangerous ports + CVEs
	for _, rf := range reconFindings {
		if rf.Type == "port" && rf.Severity == models.RiskHigh {
			for _, vf := range vulnFindings {
				if vf.Severity == models.RiskCritical {
					correlations = append(correlations, Correlation{
						Type:        "attack_surface",
						Severity:    models.RiskCritical,
						Title:       "Critical Vuln on Exposed Service",
						Description: "Critical CVE found on host with dangerous open ports",
						FindingIDs:  []string{rf.ID, vf.ID},
					})
				}
			}
		}
	}

	return correlations
}

func extractCVEID(title string) string {
	// Extract CVE-YYYY-NNNNN from title
	parts := strings.Fields(title)
	for _, p := range parts {
		if strings.HasPrefix(p, "CVE-") {
			return strings.Trim(p, "()[]{}.,;:")
		}
	}
	return ""
}
