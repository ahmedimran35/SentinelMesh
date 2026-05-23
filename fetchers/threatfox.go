package fetchers

import (
	"context"
	"fmt"
)

type ThreatFoxFetcher struct {
	*BaseFetcher
}

type threatFoxRequest struct {
	Query      string `json:"query"`
	SearchTerm string `json:"search_term,omitempty"`
}

type ThreatFoxResponse struct {
	QueryStatus string         `json:"query_status"`
	Data        []ThreatFoxIOC `json:"data,omitempty"`
}

type ThreatFoxIOC struct {
	ID           string   `json:"id"`
	IOC          string   `json:"ioc"`
	ThreatType   string   `json:"threat_type"`
	Malware      string   `json:"malware"`
	MalwareAlias string   `json:"malware_alias"`
	MalwarePrint string   `json:"malware_print"`
	FirstSeen    string   `json:"first_seen"`
	LastSeen     string   `json:"last_seen"`
	Reporter     string   `json:"reporter"`
	Confidence   int      `json:"confidence_level"`
	Reference    string   `json:"reference"`
	Tags         []string `json:"tags,omitempty"`
}

func NewThreatFoxFetcher(rateLimit int) *ThreatFoxFetcher {
	return &ThreatFoxFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

func (f *ThreatFoxFetcher) post(ctx context.Context, reqBody threatFoxRequest) (*ThreatFoxResponse, error) {
	var tfResp ThreatFoxResponse
	if err := f.PostJSON(ctx, "https://threatfox-api.abuse.ch/api/v1/", reqBody, &tfResp); err != nil {
		return nil, fmt.Errorf("threatfox request failed: %w", err)
	}
	return &tfResp, nil
}

func (f *ThreatFoxFetcher) SearchIOC(ctx context.Context, searchTerm string) (*ThreatFoxResponse, error) {
	return f.post(ctx, threatFoxRequest{Query: "search_ioc", SearchTerm: searchTerm})
}

func (f *ThreatFoxFetcher) GetRecentIOCs(ctx context.Context) (*ThreatFoxResponse, error) {
	return f.post(ctx, threatFoxRequest{Query: "get_iocs"})
}
