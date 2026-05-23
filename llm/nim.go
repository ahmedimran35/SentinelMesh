package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sentinelmesh/config"
)

type NIMProvider struct {
	cfg       *config.NIMConfig
	client    *http.Client
}

type nimRequest struct {
	Model     string        `json:"model"`
	Messages  []nimMessage  `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
	Stream    bool          `json:"stream"`
}

type nimMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type nimResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewNIMProvider(cfg *config.NIMConfig) *NIMProvider {
	return &NIMProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (p *NIMProvider) ChatCompletion(systemPrompt, userPrompt string) (string, error) {
	reqBody := nimRequest{
		Model: p.cfg.Model,
		Messages: []nimMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: p.cfg.MaxTokens,
		Stream:    false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("NIM request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("NIM API error %d", resp.StatusCode)
	}

	var nimResp nimResponse
	if err := json.Unmarshal(respBody, &nimResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if nimResp.Error != nil {
		return "", fmt.Errorf("NIM error: %s", nimResp.Error.Message)
	}

	if len(nimResp.Choices) == 0 {
		return "", fmt.Errorf("NIM returned no choices")
	}

	return nimResp.Choices[0].Message.Content, nil
}

func (p *NIMProvider) StreamCompletion(systemPrompt, userPrompt string, ch chan<- string) error {
	defer close(ch)

	reqBody := nimRequest{
		Model: p.cfg.Model,
		Messages: []nimMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: p.cfg.MaxTokens,
		Stream:    true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("NIM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NIM API error %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var nimResp nimResponse
		if err := json.Unmarshal([]byte(data), &nimResp); err != nil {
			continue
		}
		if len(nimResp.Choices) > 0 && nimResp.Choices[0].Delta.Content != "" {
			ch <- nimResp.Choices[0].Delta.Content
		}
	}

	return scanner.Err()
}
