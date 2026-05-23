package fetchers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Base fetcher with HTTP client, retry, and rate limiting
type BaseFetcher struct {
	Client    *http.Client
	RateLimit time.Duration
	mu        sync.Mutex
	lastCall  time.Time
}

func NewBaseFetcher(rateLimit int) *BaseFetcher {
	rps := time.Second / time.Duration(rateLimit)
	if rps < 100*time.Millisecond {
		rps = 100 * time.Millisecond
	}
	return &BaseFetcher{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		RateLimit: rps,
	}
}

func (f *BaseFetcher) WaitForRateLimit() {
	f.mu.Lock()
	defer f.mu.Unlock()
	elapsed := time.Since(f.lastCall)
	if elapsed < f.RateLimit {
		time.Sleep(f.RateLimit - elapsed)
	}
	f.lastCall = time.Now()
}

func (f *BaseFetcher) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	f.WaitForRateLimit()

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := f.Client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10MB cap
		resp.Body.Close()
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		if resp.StatusCode == 429 {
			// rate limited, wait longer
			time.Sleep(time.Duration(attempt+2) * time.Second)
			lastErr = fmt.Errorf("rate limited")
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
		}

		return body, nil
	}
	return nil, fmt.Errorf("request failed after 3 attempts: %w", lastErr)
}

func (f *BaseFetcher) GetJSON(ctx context.Context, url string, headers map[string]string, target interface{}) error {
	body, err := f.Get(ctx, url, headers)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func (f *BaseFetcher) PostJSON(ctx context.Context, url string, reqBody interface{}, target interface{}) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	f.WaitForRateLimit()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10MB cap
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 200)]))
	}

	return json.Unmarshal(respBody, target)
}

// min uses Go builtin (1.22+)
