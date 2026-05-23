package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"sentinelmesh/models"
	"sentinelmesh/store"
)

type ScanFunc func(inv *models.Investigation)

type Scheduler struct {
	store    *store.Store
	scanFunc ScanFunc
	timers   map[string]*time.Timer
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewScheduler(s *store.Store, scanFunc ScanFunc) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		store:    s,
		scanFunc: scanFunc,
		timers:   make(map[string]*time.Timer),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start loads monitored targets and begins scheduling
func (s *Scheduler) Start() {
	targets, err := s.store.GetMonitoredTargets()
	if err != nil {
		log.Printf("Monitor: failed to load targets: %v", err)
		return
	}
	for _, t := range targets {
		s.scheduleTarget(t)
	}
	log.Printf("Monitor: started with %d monitored targets", len(targets))
}

// Stop cancels all scheduled scans
func (s *Scheduler) Stop() {
	s.cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, t := range s.timers {
		t.Stop()
		delete(s.timers, id)
	}
}

// AddTarget adds a target to monitoring
func (s *Scheduler) AddTarget(t models.Target) {
	t.MonitorEnabled = true
	s.store.CreateTarget(&t)
	s.scheduleTarget(t)
}

// RemoveTarget removes a target from monitoring
func (s *Scheduler) RemoveTarget(id string) {
	s.mu.Lock()
	if timer, ok := s.timers[id]; ok {
		timer.Stop()
		delete(s.timers, id)
	}
	s.mu.Unlock()
	s.store.RemoveMonitor(id)
}

func (s *Scheduler) scheduleTarget(t models.Target) {
	interval := parseInterval(t.ScanInterval)
	if interval < time.Minute {
		interval = 24 * time.Hour // default
	}

	// If never scanned, scan now; otherwise wait for next interval
	delay := interval
	if t.LastScanned == nil {
		delay = 5 * time.Second // small delay to avoid startup burst
	} else {
		elapsed := time.Since(*t.LastScanned)
		if elapsed < interval {
			delay = interval - elapsed
		}
	}

	s.mu.Lock()
	// Cancel existing timer if any
	if old, ok := s.timers[t.ID]; ok {
		old.Stop()
	}
	s.mu.Unlock()

	timer := time.AfterFunc(delay, func() {
		s.runScan(t)
	})

	s.mu.Lock()
	s.timers[t.ID] = timer
	s.mu.Unlock()
}

func (s *Scheduler) runScan(t models.Target) {
	select {
	case <-s.ctx.Done():
		return
	default:
	}

	log.Printf("Monitor: scanning %s (%s)", t.Value, t.Type)

	inv := &models.Investigation{
		ID:         fmt.Sprintf("mon-%d", time.Now().UnixNano()),
		Target:     t.Value,
		TargetType: t.Type,
		Status:     models.StatusRunning,
		CreatedAt:  time.Now(),
	}

	if err := s.store.CreateInvestigation(inv); err != nil {
		log.Printf("Monitor: failed to create investigation: %v", err)
		return
	}

	s.scanFunc(inv)
	s.store.UpdateTargetLastScanned(t.ID)

	// Schedule next scan
	s.scheduleTarget(t)
}

func parseInterval(s string) time.Duration {
	if s == "" {
		return 24 * time.Hour
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
