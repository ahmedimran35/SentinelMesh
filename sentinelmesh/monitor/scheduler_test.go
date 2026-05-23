package monitor

import (
	"testing"
	"time"

	"sentinelmesh/models"
	"sentinelmesh/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestParseInterval(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"24h", 24 * time.Hour},
		{"1h", time.Hour},
		{"30m", 30 * time.Minute},
		{"invalid", 24 * time.Hour},
		{"", 24 * time.Hour},
	}
	for _, tt := range tests {
		got := parseInterval(tt.input)
		if got != tt.want {
			t.Errorf("parseInterval(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSchedulerStartStop(t *testing.T) {
	s := newTestStore(t)
	var scanned []string

	sched := NewScheduler(s, func(inv *models.Investigation) {
		scanned = append(scanned, inv.Target)
	})

	// Add a monitored target
	s.CreateTarget(&models.Target{
		ID: "t1", Value: "test.com", Type: "domain",
		MonitorEnabled: true, ScanInterval: "1h", CreatedAt: time.Now(),
	})

	sched.Start()
	sched.Stop()
	// Should not panic
}

func TestAddAndRemoveTarget(t *testing.T) {
	s := newTestStore(t)
	sched := NewScheduler(s, func(inv *models.Investigation) {})

	tgt := models.Target{
		ID: "t-add", Value: "new.com", Type: "domain",
		ScanInterval: "24h", CreatedAt: time.Now(),
	}
	sched.AddTarget(tgt)

	targets, _ := s.GetMonitoredTargets()
	if len(targets) != 1 {
		t.Errorf("got %d targets after add, want 1", len(targets))
	}

	sched.RemoveTarget("t-add")
	targets, _ = s.GetMonitoredTargets()
	if len(targets) != 0 {
		t.Errorf("got %d targets after remove, want 0", len(targets))
	}
}
