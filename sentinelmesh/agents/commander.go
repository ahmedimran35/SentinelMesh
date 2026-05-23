package agents

import (
	"context"
	"fmt"
	"strings"

	"sentinelmesh/llm"
	"sentinelmesh/models"
)

type CommanderAgent struct {
	llm llm.Provider
}

func NewCommanderAgent(llmProvider llm.Provider) *CommanderAgent {
	return &CommanderAgent{llm: llmProvider}
}

func (a *CommanderAgent) Name() string        { return "commander" }
func (a *CommanderAgent) Description() string { return "Orchestrates agents and generates intelligence reports" }

// GenerateReport takes all findings and generates a comprehensive intelligence report
func (a *CommanderAgent) GenerateReport(ctx context.Context, target models.Target, findings []models.Finding) (*models.Report, error) {
	findingsByAgent := make(map[string][]models.Finding)
	for _, f := range findings {
		findingsByAgent[f.Agent] = append(findingsByAgent[f.Agent], f)
	}

	systemPrompt := `You are a senior cybersecurity analyst and the commander of a security operations center.
Your job is to analyze findings from multiple specialized agents and produce a comprehensive intelligence report.

Your report must include:
1. EXECUTIVE SUMMARY (2-3 sentences, C-suite readable)
2. RISK RATING (Critical/High/Medium/Low with reasoning)
3. KEY FINDINGS (prioritized by severity, with technical details)
4. ATTACK SURFACE ANALYSIS
5. RECOMMENDED ACTIONS (prioritized, actionable)
6. MITRE ATT&CK MAPPING
7. INDICATORS OF COMPROMISE

Be precise, technical, and actionable. Write like a human analyst, not an AI.`

	var findingsText strings.Builder
	findingsText.WriteString(fmt.Sprintf("TARGET: %s (%s)\n\n", target.Value, target.Type))

	for agent, agentFindings := range findingsByAgent {
		findingsText.WriteString(fmt.Sprintf("=== %s Agent Findings (%d) ===\n", strings.ToUpper(agent), len(agentFindings)))
		for _, f := range agentFindings {
			findingsText.WriteString(fmt.Sprintf("\n[%s] %s\n%s\n", strings.ToUpper(string(f.Severity)), f.Title, f.Details))
		}
		findingsText.WriteString("\n")
	}

	userPrompt := fmt.Sprintf(`Analyze these findings and produce a complete intelligence report:

%s

Write the full report now.`, findingsText.String())

	response, err := a.llm.ChatCompletion(systemPrompt, userPrompt)
	if err != nil {
		// Generate a basic report without LLM
		return a.generateBasicReport(target, findings), nil
	}

	riskRating := determineRiskRating(findings)
	execSummary := extractExecutiveSummary(response)

	report := &models.Report{
		Investigation: models.Investigation{
			Target:           target.Value,
			TargetType:       target.Type,
			Status:           models.StatusComplete,
			RiskRating:       riskRating,
			ExecutiveSummary: execSummary,
			FullReport:       response,
		},
		Findings: findings,
	}

	techniques := extractMITRETechniques(response)
	report.MITRETechniques = techniques

	return report, nil
}

// GenerateSigmaRules generates Sigma detection rules for critical findings
func (a *CommanderAgent) GenerateSigmaRules(ctx context.Context, findings []models.Finding) []string {
	var rules []string

	for _, f := range findings {
		if f.Severity == models.RiskCritical || f.Severity == models.RiskHigh {
			if f.Type == "cve" || f.Type == "port" {
				rule := fmt.Sprintf(`title: %s
id: %s
status: experimental
description: Auto-generated detection rule for %s
logsource:
  category: application
detection:
  selection:
    title|contains: '%s'
  condition: selection
level: %s
tags:
  - attack.initial_access
  - auto-generated`,
					f.Title, f.ID, f.Title,
					strings.ReplaceAll(f.Title, "'", ""),
					string(f.Severity))
				rules = append(rules, rule)
			}
		}
	}

	return rules
}

// GenerateYARARules generates YARA rules for malware IOCs
func (a *CommanderAgent) GenerateYARARules(ctx context.Context, findings []models.Finding) []string {
	var rules []string

	for _, f := range findings {
		if f.Type == "ioc" || f.Type == "sample" {
			rule := fmt.Sprintf(`rule SentinelMesh_%s {
    meta:
        description = "Auto-generated rule for %s"
        author = "SentinelMesh"
        severity = "%s"
    strings:
        $s1 = "%s"
    condition:
        any of them
}`,
				sanitizeName(f.ID),
				strings.ReplaceAll(f.Title, "\"", ""),
				string(f.Severity),
				strings.ReplaceAll(f.Title, "\"", ""))
			rules = append(rules, rule)
		}
	}

	return rules
}

func (a *CommanderAgent) generateBasicReport(target models.Target, findings []models.Finding) *models.Report {
	riskRating := determineRiskRating(findings)

	var report strings.Builder
	report.WriteString(fmt.Sprintf("INTELLIGENCE REPORT\nTarget: %s\nRisk Rating: %s\n\n", target.Value, strings.ToUpper(string(riskRating))))
	report.WriteString("KEY FINDINGS:\n\n")

	for i, f := range findings {
		report.WriteString(fmt.Sprintf("%d. [%s] %s\n   %s\n\n", i+1, strings.ToUpper(string(f.Severity)), f.Title, f.Details))
	}

	return &models.Report{
		Investigation: models.Investigation{
			Target:           target.Value,
			TargetType:       target.Type,
			Status:           models.StatusComplete,
			RiskRating:       riskRating,
			ExecutiveSummary: fmt.Sprintf("Investigation of %s completed with %d findings. Risk level: %s.", target.Value, len(findings), string(riskRating)),
			FullReport:       report.String(),
		},
		Findings: findings,
	}
}

func determineRiskRating(findings []models.Finding) models.RiskRating {
	for _, f := range findings {
		if f.Severity == models.RiskCritical {
			return models.RiskCritical
		}
	}
	for _, f := range findings {
		if f.Severity == models.RiskHigh {
			return models.RiskHigh
		}
	}
	for _, f := range findings {
		if f.Severity == models.RiskMedium {
			return models.RiskMedium
		}
	}
	if len(findings) > 0 {
		return models.RiskLow
	}
	return models.RiskInfo
}

func extractExecutiveSummary(report string) string {
	// Try to extract the first meaningful paragraph
	lines := strings.Split(report, "\n")
	var summary []string
	inSummary := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "executive summary") {
			inSummary = true
			continue
		}
		if inSummary && trimmed == "" && len(summary) > 0 {
			break
		}
		if inSummary && trimmed != "" {
			summary = append(summary, trimmed)
		}
	}
	if len(summary) > 0 {
		return strings.Join(summary, " ")
	}
	// Fallback: first 300 chars
	if len(report) > 300 {
		return report[:300] + "..."
	}
	return report
}

func sanitizeName(s string) string {
	var result strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			result.WriteRune(c)
		} else {
			result.WriteRune('_')
		}
	}
	return result.String()
}
