package server

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SSEManager manages Server-Sent Events connections
type SSEManager struct {
	clients    map[chan string]bool
	mu         sync.RWMutex
}

func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[chan string]bool),
	}
}

// Subscribe creates a new SSE channel for a client
func (m *SSEManager) Subscribe() chan string {
	ch := make(chan string, 100)
	m.mu.Lock()
	m.clients[ch] = true
	m.mu.Unlock()
	return ch
}

// Unsubscribe removes a client's SSE channel
func (m *SSEManager) Unsubscribe(ch chan string) {
	m.mu.Lock()
	delete(m.clients, ch)
	m.mu.Unlock()
	close(ch)
}

// Broadcast sends an event to all connected clients
func (m *SSEManager) Broadcast(eventType string, data interface{}) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return
	}

	event := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(dataJSON))

	m.mu.RLock()
	defer m.mu.RUnlock()

	for ch := range m.clients {
		select {
		case ch <- event:
		default:
			// Client too slow, skip
		}
	}
}

// BroadcastFinding sends a finding event
func (m *SSEManager) BroadcastFinding(agent, findingType, severity, title, details string) {
	m.Broadcast("finding", map[string]string{
		"agent":    agent,
		"type":     findingType,
		"severity": severity,
		"title":    title,
		"details":  details,
		"time":     time.Now().Format(time.RFC3339),
	})
}

// BroadcastAgentStatus sends agent status update
func (m *SSEManager) BroadcastAgentStatus(agent, state, message string, progress float64) {
	m.Broadcast("agent-status", map[string]interface{}{
		"agent":    agent,
		"state":    state,
		"progress": progress,
		"message":  message,
		"time":     time.Now().Format(time.RFC3339),
	})
}

// BroadcastCommander sends commander message
func (m *SSEManager) BroadcastCommander(message, msgType string) {
	m.Broadcast("commander-message", map[string]string{
		"message": message,
		"type":    msgType,
		"time":    time.Now().Format(time.RFC3339),
	})
}

// BroadcastAlert sends an alert event
func (m *SSEManager) BroadcastAlert(alertType, severity, message, target string) {
	m.Broadcast("alert", map[string]string{
		"type":     alertType,
		"severity": severity,
		"message":  message,
		"target":   target,
		"time":     time.Now().Format(time.RFC3339),
	})
}

// BroadcastInvestigationStatus sends investigation status
func (m *SSEManager) BroadcastInvestigationStatus(id, status string, progress float64) {
	m.Broadcast("investigation-status", map[string]interface{}{
		"id":       id,
		"status":   status,
		"progress": progress,
		"time":     time.Now().Format(time.RFC3339),
	})
}

// BroadcastReportReady sends report ready event
func (m *SSEManager) BroadcastReportReady(id, riskRating, summary string) {
	m.Broadcast("report-ready", map[string]string{
		"id":         id,
		"risk_rating": riskRating,
		"summary":    summary,
		"time":       time.Now().Format(time.RFC3339),
	})
}
