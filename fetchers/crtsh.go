package fetchers

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type CrtShFetcher struct {
	*BaseFetcher
}

type CrtShEntry struct {
	ID          int       `json:"id"`
	IssuerCAID  int       `json:"issuer_ca_id"`
	IssuerName  string    `json:"issuer_name"`
	NameValue   string    `json:"name_value"`
	EntryTimestamp time.Time `json:"entry_timestamp"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	SerialNumber string   `json:"serial_number"`
}

type CrtShResult struct {
	Subdomains []string
	Issuers    []string
	Entries    []CrtShEntry
}

func NewCrtShFetcher(rateLimit int) *CrtShFetcher {
	return &CrtShFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

func (f *CrtShFetcher) Lookup(ctx context.Context, domain string) (*CrtShResult, error) {
	lookupURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", url.QueryEscape(domain))

	var entries []CrtShEntry
	if err := f.GetJSON(ctx, lookupURL, nil, &entries); err != nil {
		return nil, fmt.Errorf("crt.sh lookup failed: %w", err)
	}

	result := &CrtShResult{
		Entries: entries,
	}

	// Deduplicate subdomains
	seen := make(map[string]bool)
	issuerSeen := make(map[string]bool)

	for _, e := range entries {
		// NameValue can contain multiple domains separated by newlines
		for _, name := range splitNames(e.NameValue) {
			if !seen[name] {
				seen[name] = true
				result.Subdomains = append(result.Subdomains, name)
			}
		}
		if !issuerSeen[e.IssuerName] {
			issuerSeen[e.IssuerName] = true
			result.Issuers = append(result.Issuers, e.IssuerName)
		}
	}

	return result, nil
}

func splitNames(s string) []string {
	var names []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			names = append(names, trimmed)
		}
	}
	return names
}
