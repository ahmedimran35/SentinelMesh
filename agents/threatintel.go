package agents

import (
	"context"
	"fmt"
	"strings"

	"sentinelmesh/llm"
	"sentinelmesh/models"
)

type ThreatIntelAgent struct {
	llm llm.Provider
}

func NewThreatIntelAgent(llmProvider llm.Provider) *ThreatIntelAgent {
	return &ThreatIntelAgent{llm: llmProvider}
}

func (a *ThreatIntelAgent) Name() string        { return "threat_intel" }
func (a *ThreatIntelAgent) Description() string { return "Threat intelligence correlation and campaign detection" }

func (a *ThreatIntelAgent) Investigate(ctx context.Context, target models.Target, findings chan<- models.Finding) error {
	// This agent analyzes the target using LLM for threat assessment
	// It doesn't fetch external data directly — it reasons about what other agents found

	systemPrompt := `You are a senior threat intelligence analyst. Analyze the given target and provide:
1. Threat assessment (what kind of target is this likely to be)
2. Common attack vectors for this type of target
3. MITRE ATT&CK techniques likely applicable
4. Recommended monitoring focus areas

Be concise and technical. Output in plain text, not JSON.`

	safeTarget := sanitizeForPrompt(target.Value)
	safeType := sanitizeForPrompt(target.Type)

	userPrompt := fmt.Sprintf(`Analyze this target for threat intelligence:
Target: %s
Type: %s

What are the likely threat vectors? What MITRE ATT&CK techniques should we look for?
Provide a brief threat assessment.`, safeTarget, safeType)

	response, err := a.llm.ChatCompletion(systemPrompt, userPrompt)
	if err != nil {
		// Don't fail the whole investigation if LLM is unavailable
		findings <- NewFinding("threat_intel", "analysis", models.RiskInfo,
			"Threat Intelligence Analysis",
			"LLM analysis unavailable — manual review recommended",
			nil)
		return nil
	}

	// Parse MITRE techniques from response
	techniques := extractMITRETechniques(response)

	details := response
	if len(techniques) > 0 {
		details += "\n\nDetected MITRE Techniques:\n"
		for _, t := range techniques {
			details += "  - " + t + "\n"
		}
	}

	findings <- NewFinding("threat_intel", "analysis", models.RiskMedium,
		fmt.Sprintf("Threat Assessment for %s", target.Value),
		details, map[string]interface{}{
			"analysis":    response,
			"techniques":  techniques,
		})

	return nil
}

func extractMITRETechniques(text string) []string {
	var techniques []string
	// Look for T1234 patterns
	parts := strings.Fields(text)
	for _, p := range parts {
		p = strings.Trim(p, ".,;:()[]{}\"'")
		if strings.HasPrefix(p, "T") && len(p) >= 5 {
			// Check if it looks like a MITRE technique ID
			valid := true
			for _, c := range p[1:] {
				if c < '0' || c > '9' {
					valid = false
					break
				}
			}
			if valid {
				techniques = append(techniques, p)
			}
		}
	}
	return techniques
}
