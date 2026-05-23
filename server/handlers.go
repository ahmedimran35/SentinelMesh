package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"sentinelmesh/agents"
	"sentinelmesh/config"
	"sentinelmesh/correlation"
	"sentinelmesh/fetchers"
	"sentinelmesh/llm"
	"sentinelmesh/models"
	"sentinelmesh/store"
)

// extractID extracts a resource ID from a path like /api/{resource}/{id}
func extractID(path, prefix string) string {
	return strings.TrimPrefix(path, prefix)
}

// extractSubResource extracts ID and sub-resource from paths like /api/investigations/{id}/findings
func extractSubResource(path, prefix string) (id string, sub string, ok bool) {
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

type Handler struct {
	store       *store.Store
	sse         *SSEManager
	registry    *agents.Registry
	llmProvider llm.Provider
	cfg         *config.Config
	llmMu       sync.RWMutex
	invSem      chan struct{}
}

func NewHandler(s *store.Store, sse *SSEManager, cfg *config.Config) *Handler {
	h := &Handler{
		store:      s,
		sse:        sse,
		registry:   agents.NewRegistry(),
		llmProvider: llm.NewProvider(&cfg.LLM),
		cfg:        cfg,
		invSem:     make(chan struct{}, cfg.Monitor.MaxConcurrent),
	}

	// Register all agents
	h.registry.Register(agents.NewReconAgent(cfg.RateLimit))
	h.registry.Register(agents.NewVulnAgent(cfg.RateLimit, h.llmProvider))
	h.registry.Register(agents.NewMalwareAgent(cfg.RateLimit))
	h.registry.Register(agents.NewThreatIntelAgent(h.llmProvider))
	h.registry.Register(agents.NewNewsIntelAgent(cfg.RateLimit))

	return h
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}

// POST /api/investigate
func (h *Handler) StartInvestigation(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit
	var req struct {
		Target string `json:"target"`
		Type   string `json:"type"` // "domain", "ip", "ip_range"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Target == "" {
		h.writeError(w, http.StatusBadRequest, "target is required")
		return
	}
	if req.Type == "" {
		req.Type = "domain"
	}
	validTypes := map[string]bool{"domain": true, "ip": true, "ip_range": true}
	if !validTypes[req.Type] {
		h.writeError(w, http.StatusBadRequest, "type must be domain, ip, or ip_range")
		return
	}

	inv := &models.Investigation{
		ID:         fmt.Sprintf("inv-%d", time.Now().UnixNano()),
		Target:     req.Target,
		TargetType: req.Type,
		Status:     models.StatusRunning,
		CreatedAt:  time.Now(),
	}

	if err := h.store.CreateInvestigation(inv); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to create investigation")
		return
	}

	select {
	case h.invSem <- struct{}{}:
		go func() {
			defer func() { <-h.invSem }()
			h.runInvestigation(inv)
		}()
	default:
		h.writeError(w, http.StatusTooManyRequests, "too many concurrent investigations")
		return
	}

	h.writeJSON(w, http.StatusAccepted, inv)
}

func (h *Handler) runInvestigation(inv *models.Investigation) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	h.sse.BroadcastInvestigationStatus(inv.ID, "running", 0)
	h.sse.BroadcastCommander(fmt.Sprintf("Starting investigation of %s — dispatching all agents", inv.Target), "analysis")

	target := models.Target{
		Value: inv.Target,
		Type:  inv.TargetType,
	}

	allAgents := h.registry.All()
	agentList := make([]agents.Agent, 0, len(allAgents))
	for _, a := range allAgents {
		agentList = append(agentList, a)
	}

	var findings []models.Finding

	if inv.TargetType == "ip_range" {
		ips, cidrErr := fetchers.ExpandCIDR(inv.Target)
		if cidrErr == nil && len(ips) > 0 {
			h.sse.BroadcastCommander(fmt.Sprintf("CIDR expanded: %d hosts to scan", len(ips)), "analysis")
			var mu sync.Mutex
			var wg sync.WaitGroup
			sem := make(chan struct{}, h.cfg.Monitor.MaxConcurrent)
			for _, ip := range ips {
				wg.Add(1)
				sem <- struct{}{}
				go func(ipAddr string) {
					defer wg.Done()
					defer func() { <-sem }()
					ipTarget := models.Target{Value: ipAddr, Type: "ip"}
					fs, err := agents.RunAgents(ctx, agentList, ipTarget, func(name string, state agents.AgentState) {
						h.sse.BroadcastAgentStatus(name, state.State, state.Message, state.Progress)
					})
					if err == nil {
						mu.Lock()
						findings = append(findings, fs...)
						mu.Unlock()
					}
				}(ip)
			}
			wg.Wait()
		} else {
			// CIDR parse failed — fall through to normal scan
			h.sse.BroadcastCommander(fmt.Sprintf("CIDR parse failed (%v), scanning as single target", cidrErr), "analysis")
			findings, _ = agents.RunAgents(ctx, agentList, target, func(name string, state agents.AgentState) {
				h.sse.BroadcastAgentStatus(name, state.State, state.Message, state.Progress)
			})
		}
	} else {
		findings, _ = agents.RunAgents(ctx, agentList, target, func(name string, state agents.AgentState) {
			h.sse.BroadcastAgentStatus(name, state.State, state.Message, state.Progress)
		})
	}

	if len(findings) == 0 && inv.TargetType != "ip_range" {
		inv.Status = models.StatusError
		h.store.UpdateInvestigation(inv)
		h.sse.BroadcastInvestigationStatus(inv.ID, "error", 0)
		return
	}

	for i := range findings {
		findings[i].InvestigationID = inv.ID
		h.store.CreateFinding(&findings[i])
		h.sse.BroadcastFinding(findings[i].Agent, findings[i].Type, string(findings[i].Severity), findings[i].Title, findings[i].Details)
	}

	h.sse.BroadcastCommander(fmt.Sprintf("All agents complete. %d findings collected. Generating report...", len(findings)), "analysis")

	commander := agents.NewCommanderAgent(h.llmProvider)
	report, err := commander.GenerateReport(ctx, target, findings)
	if err != nil {
		inv.Status = models.StatusError
		h.store.UpdateInvestigation(inv)
		return
	}

	inv.Status = models.StatusComplete
	inv.RiskRating = report.Investigation.RiskRating
	inv.ExecutiveSummary = report.Investigation.ExecutiveSummary
	inv.FullReport = report.Investigation.FullReport
	now := time.Now()
	inv.CompletedAt = &now
	h.store.UpdateInvestigation(inv)

	sigmaRules := commander.GenerateSigmaRules(ctx, findings)
	for i, rule := range sigmaRules {
		h.store.CreateRule(fmt.Sprintf("rule-%s-sigma-%d", inv.ID, i), inv.ID, "sigma", rule)
	}
	yaraRules := commander.GenerateYARARules(ctx, findings)
	for i, rule := range yaraRules {
		h.store.CreateRule(fmt.Sprintf("rule-%s-yara-%d", inv.ID, i), inv.ID, "yara", rule)
	}

	h.sse.BroadcastCommander("Correlating findings across agents...", "correlation")
	corr := correlation.NewCorrelator()
	correlations := corr.Analyze(findings)
	for _, c := range correlations {
		for _, fid := range c.FindingIDs {
			for _, f := range findings {
				if f.ID == fid {
					h.sse.BroadcastAlert(c.Type, string(c.Severity), c.Title, inv.Target)
					break
				}
			}
		}
		h.store.CreateAlert(&models.Alert{
			ID:              fmt.Sprintf("alert-%d", time.Now().UnixNano()),
			Target:          inv.Target,
			AlertType:       c.Type,
			Severity:        c.Severity,
			Message:         c.Title,
			Acknowledged:    false,
			CreatedAt:       time.Now(),
		})
	}
	if len(correlations) > 0 {
		h.sse.BroadcastCommander(fmt.Sprintf("⚠ %d correlations detected across agents!", len(correlations)), "correlation")
	}

	h.sse.BroadcastInvestigationStatus(inv.ID, "complete", 1.0)
	h.sse.BroadcastReportReady(inv.ID, string(inv.RiskRating), inv.ExecutiveSummary)
	h.sse.BroadcastCommander(fmt.Sprintf("Investigation complete. Risk: %s. Report ready.", strings.ToUpper(string(inv.RiskRating))), "recommendation")
}

// GET /api/investigations
func (h *Handler) ListInvestigations(w http.ResponseWriter, r *http.Request) {
	invs, err := h.store.ListInvestigations(50, 0)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to list investigations")
		return
	}
	h.writeJSON(w, http.StatusOK, invs)
}

// GET /api/investigations/:id
func (h *Handler) GetInvestigation(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/investigations/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	inv, err := h.store.GetInvestigation(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "investigation not found")
		return
	}
	h.writeJSON(w, http.StatusOK, inv)
}

// DELETE /api/investigations/:id
func (h *Handler) DeleteInvestigation(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/investigations/")
	if err := h.store.DeleteInvestigation(id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to delete investigation")
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GET /api/investigations/:id/findings
func (h *Handler) GetFindings(w http.ResponseWriter, r *http.Request) {
	id, _, ok := extractSubResource(r.URL.Path, "/api/investigations/")
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	findings, err := h.store.GetFindings(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to get findings")
		return
	}
	h.writeJSON(w, http.StatusOK, findings)
}

// GET /api/findings/search
func (h *Handler) SearchFindings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	severity := r.URL.Query().Get("severity")
	agent := r.URL.Query().Get("agent")

	findings, err := h.store.SearchFindings(query, severity, agent, 100)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to search findings")
		return
	}
	h.writeJSON(w, http.StatusOK, findings)
}

// GET /api/alerts
func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	unackOnly := r.URL.Query().Get("unacknowledged") == "true"
	alerts, err := h.store.GetAlerts(unackOnly)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to get alerts")
		return
	}
	h.writeJSON(w, http.StatusOK, alerts)
}

// PUT /api/alerts/:id/ack
func (h *Handler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/alerts/")
	id = strings.TrimSuffix(id, "/ack")
	if err := h.store.AcknowledgeAlert(id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to acknowledge alert")
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

// GET /api/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetStats()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	h.writeJSON(w, http.StatusOK, stats)
}

// GET /api/health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// GET /api/settings
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.store.GetAllSettings()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to get settings")
		return
	}

	// Mask API key for security
	if _, ok := settings["nim_api_key"]; ok {
		settings["nim_api_key_masked"] = "***"
		delete(settings, "nim_api_key")
	}

	h.writeJSON(w, http.StatusOK, settings)
}

// POST /api/settings
func (h *Handler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	allowedKeys := map[string]bool{
		"nim_api_key": true, "nim_endpoint": true, "nim_model": true,
		"llm_provider": true, "ollama_url": true, "ollama_model": true,
	}
	for key, value := range settings {
		if !allowedKeys[key] {
			continue
		}
		if err := h.store.SetSetting(key, value); err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save setting: %s", key))
			return
		}
	}

	// Reload LLM provider with new settings
	h.reloadLLMProvider()

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (h *Handler) reloadLLMProvider() {
	apiKey, _ := h.store.GetSetting("nim_api_key")
	endpoint, _ := h.store.GetSetting("nim_endpoint")
	model, _ := h.store.GetSetting("nim_model")
	provider, _ := h.store.GetSetting("llm_provider")

	if provider == "" {
		provider = "nim"
	}

	h.llmMu.Lock()
	defer h.llmMu.Unlock()

	if apiKey != "" && provider == "nim" {
		h.cfg.LLM.Provider = "nim"
		h.cfg.LLM.NIM.APIKey = apiKey
		if endpoint != "" {
			h.cfg.LLM.NIM.Endpoint = endpoint
		}
		if model != "" {
			h.cfg.LLM.NIM.Model = model
		}
	} else {
		h.cfg.LLM.Provider = "ollama"
	}

	h.llmProvider = llm.NewProvider(&h.cfg.LLM)
}

// allowedNIMDomains are the only domains the NIM API will call (SSRF protection)
var allowedNIMDomains = []string{
	"integrate.api.nvidia.com",
	"api.nvcf.nvidia.com",
	"ai.api.nvidia.com",
}

func isAllowedNIMEndpoint(endpoint string) bool {
	u, err := url.Parse(endpoint)
	if err != nil {
		return false
	}
	host := u.Hostname()
	for _, d := range allowedNIMDomains {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

// GET /api/nim/models - Fetch available models from NVIDIA NIM
func (h *Handler) ListNIMModels(w http.ResponseWriter, r *http.Request) {
	var apiKey, endpoint string
	if r.Method == http.MethodPost {
		var req struct {
			APIKey   string `json:"api_key"`
			Endpoint string `json:"endpoint"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		apiKey = req.APIKey
		endpoint = req.Endpoint
	}

	if apiKey == "" {
		apiKey, _ = h.store.GetSetting("nim_api_key")
	}
	if endpoint == "" {
		endpoint, _ = h.store.GetSetting("nim_endpoint")
	}
	if endpoint == "" {
		endpoint = "https://integrate.api.nvidia.com/v1"
	}

	if !isAllowedNIMEndpoint(endpoint) {
		h.writeError(w, http.StatusBadRequest, "invalid NIM endpoint domain")
		return
	}

	if apiKey == "" {
		h.writeError(w, http.StatusBadRequest, "NIM API key required")
		return
	}

	// Fetch models from NIM API
	modelsEndpoint := strings.Replace(endpoint, "/chat/completions", "/models", 1)
	if !strings.HasSuffix(modelsEndpoint, "/models") {
		modelsEndpoint = strings.TrimRight(modelsEndpoint, "/") + "/models"
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", modelsEndpoint, nil) // #nosec endpoint validated by isAllowedNIMEndpoint above
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to create request")
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req) // #nosec endpoint validated by isAllowedNIMEndpoint above
	if err != nil {
		h.writeError(w, http.StatusBadGateway, "failed to connect to NIM API")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyLen := len(body)
		if bodyLen > 200 {
			bodyLen = 200
		}
		h.writeError(w, resp.StatusCode, fmt.Sprintf("NIM API returned %d: %s", resp.StatusCode, string(body[:bodyLen])))
		return
	}

	var nimResp struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int    `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&nimResp); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to parse NIM response")
		return
	}

	type ModelInfo struct {
		ID      string `json:"id"`
		OwnedBy string `json:"owned_by"`
	}

	var models []ModelInfo
	for _, m := range nimResp.Data {
		models = append(models, ModelInfo{
			ID:      m.ID,
			OwnedBy: m.OwnedBy,
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"models": models,
		"count":  len(models),
	})
}

// POST /api/nim/test - Test NIM connection
func (h *Handler) TestNIMConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey   string `json:"api_key"`
		Endpoint string `json:"endpoint"`
		Model    string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.APIKey == "" {
		req.APIKey, _ = h.store.GetSetting("nim_api_key")
	}
	if req.Endpoint == "" {
		req.Endpoint, _ = h.store.GetSetting("nim_endpoint")
	}
	if req.Endpoint == "" {
		req.Endpoint = "https://integrate.api.nvidia.com/v1/chat/completions"
	}

	if !isAllowedNIMEndpoint(req.Endpoint) {
		h.writeError(w, http.StatusBadRequest, "invalid NIM endpoint domain")
		return
	}

	if req.Model == "" {
		req.Model, _ = h.store.GetSetting("nim_model")
	}
	if req.Model == "" {
		req.Model = "meta/llama-3.1-8b-instruct"
	}

	// Send a simple test request
	testBody := map[string]interface{}{
		"model": req.Model,
		"messages": []map[string]string{
			{"role": "user", "content": "Say 'SentinelMesh connected' in 5 words or less."},
		},
		"max_tokens": 50,
	}

	body, _ := json.Marshal(testBody)
	client := &http.Client{Timeout: 30 * time.Second}
	httpReq, err := http.NewRequest("POST", req.Endpoint, strings.NewReader(string(body)))
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to create request")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	start := time.Now()
	resp, err := client.Do(httpReq) // #nosec endpoint validated by isAllowedNIMEndpoint above
	latency := time.Since(start).Milliseconds()

	if err != nil {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "error",
			"message": "Connection failed: unable to reach endpoint",
		})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":     "error",
			"message":    fmt.Sprintf("API returned %d", resp.StatusCode),
			"response":   string(respBody[:min(len(respBody), 300)]),
			"latency_ms": latency,
		})
		return
	}

	var nimResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &nimResp); err != nil {
		_ = err
	}

	response := ""
	if len(nimResp.Choices) > 0 {
		response = nimResp.Choices[0].Message.Content
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "ok",
		"message":    "Connection successful",
		"model":      req.Model,
		"response":   response,
		"latency_ms": latency,
	})
}

// GET /events (SSE)
func (h *Handler) SSEHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	ch := h.sse.Subscribe()
	defer h.sse.Unsubscribe(ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprint(w, event)
			flusher.Flush()
		}
	}
}

// RunInvestigationPublic is called by the monitor scheduler
func (h *Handler) RunInvestigationPublic(inv *models.Investigation) {
	h.runInvestigation(inv)
}

// POST /api/monitors — add target to monitoring
func (h *Handler) AddMonitor(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target       string `json:"target"`
		Type         string `json:"type"`
		ScanInterval string `json:"scan_interval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Target == "" {
		h.writeError(w, http.StatusBadRequest, "target is required")
		return
	}
	if req.Type == "" {
		req.Type = "domain"
	}
	if req.ScanInterval == "" {
		req.ScanInterval = "24h"
	}

	t := models.Target{
		ID:             fmt.Sprintf("tgt-%d", time.Now().UnixNano()),
		Value:          req.Target,
		Type:           req.Type,
		MonitorEnabled: true,
		ScanInterval:   req.ScanInterval,
		CreatedAt:      time.Now(),
	}
	if err := h.store.CreateTarget(&t); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to create monitor")
		return
	}
	h.writeJSON(w, http.StatusCreated, t)
}

// GET /api/monitors — list monitored targets
func (h *Handler) ListMonitors(w http.ResponseWriter, r *http.Request) {
	targets, err := h.store.GetMonitoredTargets()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to list monitors")
		return
	}
	h.writeJSON(w, http.StatusOK, targets)
}

// DELETE /api/monitors/:id
func (h *Handler) RemoveMonitor(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/monitors/")
	if err := h.store.RemoveMonitor(id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to remove monitor")
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// GET /api/investigations/:id/rules
func (h *Handler) GetRules(w http.ResponseWriter, r *http.Request) {
	id, _, ok := extractSubResource(r.URL.Path, "/api/investigations/")
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	rules, err := h.store.GetRules(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to get rules")
		return
	}
	h.writeJSON(w, http.StatusOK, rules)
}

// GET /api/investigations/:id/export?format=json|csv|stix
func (h *Handler) ExportReport(w http.ResponseWriter, r *http.Request) {
	id, _, ok := extractSubResource(r.URL.Path, "/api/investigations/")
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	inv, err := h.store.GetInvestigation(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "investigation not found")
		return
	}

	findings, _ := h.store.GetFindings(id)

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sentinelmesh-%s.csv\"", id))
		fmt.Fprintln(w, "target,agent,type,severity,title,details,timestamp")
		for _, f := range findings {
			fmt.Fprintf(w, "%q,%q,%q,%q,%q,%q,%s\n",
				inv.Target, f.Agent, f.Type, string(f.Severity),
				f.Title, f.Details, f.Timestamp.Format(time.RFC3339))
		}
	case "stix":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sentinelmesh-%s.json\"", id))
		h.writeJSON(w, http.StatusOK, generateSTIX(inv, findings))
	default: // json
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sentinelmesh-%s.json\"", id))
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"investigation": inv,
			"findings":      findings,
		})
	}
}

func generateSTIX(inv *models.Investigation, findings []models.Finding) map[string]interface{} {
	objects := []map[string]interface{}{
		{
			"type":        "identity",
			"spec_version": "2.1",
			"id":          "identity--" + inv.ID,
			"name":        "SentinelMesh",
			"identity_class": "tool",
		},
	}

	for _, f := range findings {
		indicator := map[string]interface{}{
			"type":         "indicator",
			"spec_version": "2.1",
			"id":           "indicator--" + f.ID,
			"created":      f.Timestamp.Format(time.RFC3339),
			"modified":     f.Timestamp.Format(time.RFC3339),
			"name":         f.Title,
			"description":  f.Details,
			"pattern":      fmt.Sprintf("[x-custom:agent = '%s' AND x-custom:type = '%s']", strings.ReplaceAll(f.Agent, "'", "\\'"), strings.ReplaceAll(f.Type, "'", "\\'")),
			"pattern_type": "stix",
			"valid_from":   f.Timestamp.Format(time.RFC3339),
			"labels":       []string{string(f.Severity)},
			"created_by_ref": "identity--" + inv.ID,
		}
		objects = append(objects, indicator)
	}

	return map[string]interface{}{
		"type":    "bundle",
		"id":      "bundle--" + inv.ID,
		"objects": objects,
	}
}
