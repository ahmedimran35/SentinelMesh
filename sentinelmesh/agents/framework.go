package agents

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"sentinelmesh/models"
)

// Agent is the interface all agents must implement
type Agent interface {
	Name() string
	Description() string
	Investigate(ctx context.Context, target models.Target, findings chan<- models.Finding) error
}

// AgentState tracks the current state of an agent
type AgentState struct {
	Agent     string
	State     string // "idle", "scanning", "analyzing", "complete", "error"
	Progress  float64
	Message   string
	Findings  int
	StartTime time.Time
	Error     error
}

// Registry holds all registered agents
type Registry struct {
	agents map[string]Agent
	mu     sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]Agent),
	}
}

func (r *Registry) Register(a Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[a.Name()] = a
}

func (r *Registry) Get(name string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[name]
	return a, ok
}

func (r *Registry) All() map[string]Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]Agent)
	for k, v := range r.agents {
		result[k] = v
	}
	return result
}

// RunAgents runs all agents in parallel and collects findings
func RunAgents(ctx context.Context, agents []Agent, target models.Target, onUpdate func(string, AgentState)) ([]models.Finding, error) {
	var allFindings []models.Finding
	var mu sync.Mutex
	var wg sync.WaitGroup

	states := make(map[string]*AgentState)
	for _, a := range agents {
		states[a.Name()] = &AgentState{
			Agent:     a.Name(),
			State:     "scanning",
			Progress:  0,
			StartTime: time.Now(),
		}
		onUpdate(a.Name(), *states[a.Name()])
	}

	findingCh := make(chan models.Finding, 100)

	// Start all agents
	for _, a := range agents {
		wg.Add(1)
		go func(ag Agent) {
			defer wg.Done()

			state := states[ag.Name()]
			state.State = "scanning"
			state.Progress = 0.1
			onUpdate(ag.Name(), *state)

			err := ag.Investigate(ctx, target, findingCh)

			if err != nil {
				state.State = "error"
				state.Error = err
				state.Message = fmt.Sprintf("Error: %s failed", ag.Name())
			} else {
				state.State = "complete"
				state.Progress = 1.0
				state.Message = "Complete"
			}
			onUpdate(ag.Name(), *state)
		}(a)
	}

	done := make(chan struct{})
	go func() {
		for f := range findingCh {
			mu.Lock()
			allFindings = append(allFindings, f)
			if state, ok := states[f.Agent]; ok {
				state.Findings++
			}
			mu.Unlock()
		}
		close(done)
	}()

	wg.Wait()
	close(findingCh)
	<-done

	return allFindings, nil
}

var findingCounter uint64

// Helper to create a finding
func NewFinding(agent, findingType string, severity models.RiskRating, title, details string, rawData interface{}) models.Finding {
	id := atomic.AddUint64(&findingCounter, 1)
	return models.Finding{
		ID:        fmt.Sprintf("f-%d", id),
		Agent:     agent,
		Type:      findingType,
		Severity:  severity,
		Title:     title,
		Details:   details,
		RawData:   rawData,
		Timestamp: time.Now(),
	}
}
