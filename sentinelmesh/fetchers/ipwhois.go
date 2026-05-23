package fetchers

import (
	"context"
	"fmt"
	"net"
)

type IPWhoisFetcher struct {
	*BaseFetcher
}

type IPWhoisResult struct {
	IP           string  `json:"ip"`
	Success      bool    `json:"success"`
	Type         string  `json:"type"`
	Continent    string  `json:"continent"`
	ContinentCode string `json:"continent_code"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	Region       string  `json:"region"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ISP          string  `json:"isp"`
	Org          string  `json:"org"`
	ASN          string  `json:"asn"`
	ASNOrg       string  `json:"asn_org"`
	Timezone     string  `json:"timezone"`
	CountryFlag  string  `json:"country_flag"`
	Currency     string  `json:"currency"`
	IsProxy      bool    `json:"proxy"`
	IsHosting    bool    `json:"hosting"`
}

func NewIPWhoisFetcher(rateLimit int) *IPWhoisFetcher {
	return &IPWhoisFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

func (f *IPWhoisFetcher) Lookup(ctx context.Context, ip string) (*IPWhoisResult, error) {
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}
	lookupURL := fmt.Sprintf("https://ipwhois.io/json/%s", ip)

	var result IPWhoisResult
	if err := f.GetJSON(ctx, lookupURL, nil, &result); err != nil {
		return nil, fmt.Errorf("ipwhois lookup for %s failed: %w", ip, err)
	}

	if !result.Success {
		return nil, fmt.Errorf("ipwhois returned unsuccessful for %s", ip)
	}

	return &result, nil
}