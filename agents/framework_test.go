package agents

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sentinelmesh/models"
)

type mockAgent struct {
	name       string
	findings   int
	shouldFail bool
}

func (m *mockAgent) Name() string        { return m.name }
func (m *mockAgent) Description() string { return "mock" }
func (m *mockAgent) Investigate(ctx context.Context, target models.Target, findings chan<- models.Finding) error {
	if m.shouldFail {
		return fmt.Errorf("mock error")
	}
	for i := 0; i < m.findings; i++ {
		findings <- models.Finding{
			ID:        fmt.Sprintf("%s-f-%d", m.name, i),
			Agent:     m.name,
			Type:      "test",
			Severity:  models.RiskInfo,
			Title:     fmt.Sprintf("Finding %d", i),
			Timestamp: time.Now(),
		}
	}
	return nil
}

func TestRunAgentsCollectsFindings(t *testing.T) {
	agents := []Agent{
		&mockAgent{name: "a1", findings: 2},
		&mockAgent{name: "a2", findings: 3},
	}
	target := models.Target{Value: "test.com", Type: "domain"}
	var states []AgentState

	findings, err := RunAgents(context.Background(), agents, target, func(name string, state AgentState) {
		states = append(states, state)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 5 {
		t.Errorf("got %d findings, want 5", len(findings))
	}
}

func TestRunAgentsHandlesErrors(t *testing.T) {
	agents := []Agent{
		&mockAgent{name: "ok", findings: 1},
		&mockAgent{name: "fail", shouldFail: true},
	}
	target := models.Target{Value: "test.com", Type: "domain"}
	var errorStates []string

	findings, err := RunAgents(context.Background(), agents, target, func(name string, state AgentState) {
		if state.State == "error" {
			errorStates = append(errorStates, name)
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still get findings from successful agent
	if len(findings) < 1 {
		t.Error("expected at least 1 finding from successful agent")
	}
	if len(errorStates) != 1 || errorStates[0] != "fail" {
		t.Errorf("expected error state for 'fail', got %v", errorStates)
	}
}

func TestNewFinding(t *testing.T) {
	f := NewFinding("test", "port", models.RiskHigh, "Test Title", "details", nil)
	if f.Agent != "test" {
		t.Errorf("agent = %q", f.Agent)
	}
	if f.Severity != models.RiskHigh {
		t.Errorf("severity = %q", f.Severity)
	}
	if f.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	a := &mockAgent{name: "test-agent"}
	r.Register(a)

	got, ok := r.Get("test-agent")
	if !ok {
		t.Error("expected to find test-agent")
	}
	if got.Name() != "test-agent" {
		t.Errorf("name = %q", got.Name())
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("should not find nonexistent agent")
	}

	all := r.All()
	if len(all) != 1 {
		t.Errorf("got %d agents, want 1", len(all))
	}
}
