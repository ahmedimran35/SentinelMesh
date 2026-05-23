package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"sentinelmesh/config"
)

type OllamaProvider struct {
	cfg    *config.OllamaConfig
	client *http.Client
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system,omitempty"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
	Done     bool   `json:"done"`
}

func NewOllamaProvider(cfg *config.OllamaConfig) *OllamaProvider {
	return &OllamaProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (p *OllamaProvider) ChatCompletion(systemPrompt, userPrompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:  p.cfg.Model,
		Prompt: userPrompt,
		System: systemPrompt,
		Stream: false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := p.cfg.URL + "/api/generate"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama error %d", resp.StatusCode)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}

func (p *OllamaProvider) StreamCompletion(systemPrompt, userPrompt string, ch chan<- string) error {
	defer close(ch)

	reqBody := ollamaRequest{
		Model:  p.cfg.Model,
		Prompt: userPrompt,
		System: systemPrompt,
		Stream: true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := p.cfg.URL + "/api/generate"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama error %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var ollamaResp ollamaResponse
		if err := json.Unmarshal(line, &ollamaResp); err != nil {
			continue
		}

		if ollamaResp.Error != "" {
			return fmt.Errorf("ollama error: %s", ollamaResp.Error)
		}

		if ollamaResp.Response != "" {
			ch <- ollamaResp.Response
		}

		if ollamaResp.Done {
			break
		}
	}

	return scanner.Err()
}
