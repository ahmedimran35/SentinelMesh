package fetchers

import (
	"context"
	"fmt"
	"net/url"
)

type DNSFetcher struct {
	*BaseFetcher
}

type DNSResponse struct {
	Answer []DNSAnswer `json:"Answer"`
}

type DNSAnswer struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

type DNSResult struct {
	A     []string
	AAAA  []string
	MX    []string
	NS    []string
	TXT   []string
	CNAME []string
}

func NewDNSFetcher(rateLimit int) *DNSFetcher {
	return &DNSFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

func (f *DNSFetcher) Lookup(ctx context.Context, domain string) (*DNSResult, error) {
	result := &DNSResult{}
	recordTypes := map[string]int{
		"A":     1,
		"AAAA":  28,
		"MX":    15,
		"NS":    2,
		"TXT":   16,
		"CNAME": 5,
	}

	for typeName, typeNum := range recordTypes {
		records, err := f.queryType(ctx, domain, typeNum)
		if err != nil {
			continue // some types may not exist
		}
		switch typeName {
		case "A":
			result.A = records
		case "AAAA":
			result.AAAA = records
		case "MX":
			result.MX = records
		case "NS":
			result.NS = records
		case "TXT":
			result.TXT = records
		case "CNAME":
			result.CNAME = records
		}
	}

	return result, nil
}

func (f *DNSFetcher) queryType(ctx context.Context, domain string, qtype int) ([]string, error) {
	params := url.Values{}
	params.Set("name", domain)
	params.Set("type", fmt.Sprintf("%d", qtype))
	params.Set("ct", "application/dns-json")

	reqURL := "https://cloudflare-dns.com/dns-query?" + params.Encode()

	var resp DNSResponse
	if err := f.GetJSON(ctx, reqURL, map[string]string{"Accept": "application/dns-json"}, &resp); err != nil {
		return nil, err
	}

	var records []string
	for _, a := range resp.Answer {
		if a.Data != "" {
			records = append(records, a.Data)
		}
	}
	return records, nil
}

// GetIPs returns all IP addresses (A + AAAA)
func (r *DNSResult) GetIPs() []string {
	var ips []string
	ips = append(ips, r.A...)
	ips = append(ips, r.AAAA...)
	return ips
}
