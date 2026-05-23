package agents

import (
	"context"
	"fmt"
	"strings"

	"sentinelmesh/fetchers"
	"sentinelmesh/llm"
	"sentinelmesh/models"
)

type VulnAgent struct {
	nvd *fetchers.NVDFetcher
	llm llm.Provider
}

func NewVulnAgent(rateLimit int, llmProvider llm.Provider) *VulnAgent {
	return &VulnAgent{
		nvd: fetchers.NewNVDFetcher(rateLimit),
		llm: llmProvider,
	}
}

func (a *VulnAgent) Name() string        { return "vuln" }
func (a *VulnAgent) Description() string { return "CVE analysis and vulnerability assessment" }

func (a *VulnAgent) Investigate(ctx context.Context, target models.Target, findings chan<- models.Finding) error {
	searchTerms := []string{target.Value}

	if target.Type == "domain" {
		searchTerms = append(searchTerms,
			target.Value+" apache",
			target.Value+" nginx",
			target.Value+" wordpress",
		)
	}

	for _, term := range searchTerms {
		cves, err := a.nvd.SearchCVEs(ctx, term, 10)
		if err != nil {
			continue
		}

		for _, cve := range cves {
			severity := models.RiskInfo
			switch strings.ToUpper(cve.Severity) {
			case "CRITICAL":
				severity = models.RiskCritical
			case "HIGH":
				severity = models.RiskHigh
			case "MEDIUM":
				severity = models.RiskMedium
			case "LOW":
				severity = models.RiskLow
			}

			details := fmt.Sprintf("CVE: %s\nScore: %.1f (%s)\nPublished: %s\nDescription: %s",
				cve.ID, cve.BaseScore, cve.Severity, cve.Published.Format("2006-01-02"), cve.Description)

			if len(cve.References) > 0 {
				details += "\nReferences:\n"
				for _, ref := range cve.References[:min(len(cve.References), 5)] {
					details += "  - " + ref + "\n"
				}
			}

			findings <- NewFinding("vuln", "cve", severity,
				fmt.Sprintf("%s (%.1f %s)", cve.ID, cve.BaseScore, cve.Severity),
				details, cve)
		}
	}

	return nil
}
