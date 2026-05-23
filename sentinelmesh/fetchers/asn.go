package fetchers

import (
	"context"
	"fmt"
	"net"
	"strings"
)

type ASNFetcher struct {
	*BaseFetcher
}

type BGPViewResult struct {
	Data struct {
		IP               string `json:"ip"`
		Prefix           string `json:"prefix"`
		ASN              int    `json:"asn"`
		ASNName          string `json:"asn_name"`
		ASNCountry       string `json:"asn_country"`
		AllocationRIR    string `json:"rir_allocation"`
		AllocationDate   string `json:"allocation_date"`
		Description      string `json:"description"`
	} `json:"data"`
}

type RIRWhoisResult struct {
	ASN           string `json:"asn"`
	ASNName       string `json:"asn_name"`
	ASNCountry    string `json:"asn_country"`
	ASNRange      string `json:"asn_range"`
	Description   string `json:"description"`
	AbuseContact  string `json:"abuse_contact"`
	LookingGlass  string `json:"looking_glass"`
}

func NewASNFetcher(rateLimit int) *ASNFetcher {
	return &ASNFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

func (f *ASNFetcher) LookupIP(ctx context.Context, ip string) (*BGPViewResult, error) {
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}
	lookupURL := fmt.Sprintf("https://api.bgpview.io/ip/%s", ip)

	var raw struct {
		Status string        `json:"status"`
		Data   struct {
			IP               string `json:"ip"`
			Prefix           string `json:"prefix"`
			ASN              int    `json:"asn"`
			ASNName          string `json:"asn_name"`
			ASNCountry       string `json:"asn_country"`
			AllocationRIR    string `json:"rir_allocation"`
			AllocationDate   string `json:"allocation_date"`
			Description      string `json:"description"`
		} `json:"data"`
	}

	headers := map[string]string{
		"Accept": "application/json",
	}

	if err := f.GetJSON(ctx, lookupURL, headers, &raw); err != nil {
		return nil, fmt.Errorf("bgpview lookup for %s failed: %w", ip, err)
	}

	result := &BGPViewResult{}
	result.Data = raw.Data
	return result, nil
}

func (f *ASNFetcher) GetASNInfo(ctx context.Context, asn int) (*RIRWhoisResult, error) {
	url := fmt.Sprintf("https://api.bgpview.io/asn/%d", asn)
	headers := map[string]string{
		"Accept": "application/json",
	}

	var resp struct {
		Status string `json:"status"`
		Data   struct {
			ASN           int    `json:"asn"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			CountryCode   string `json:"country_code"`
			RirAllocation string `json:"rir_allocation"`
		} `json:"data"`
	}

	if err := f.GetJSON(ctx, url, headers, &resp); err != nil {
		return nil, fmt.Errorf("bgpview ASN lookup for %d failed: %w", asn, err)
	}

	result := &RIRWhoisResult{
		ASN:          fmt.Sprintf("AS%d", resp.Data.ASN),
		ASNName:      resp.Data.Name,
		ASNCountry:   resp.Data.CountryCode,
		Description:  resp.Data.Description,
		ASNRange:     string(resp.Data.RirAllocation),
	}

	return result, nil
}

func (f *ASNFetcher) GetPrefixes(ctx context.Context, asn int) ([]string, error) {
	url := fmt.Sprintf("https://api.bgpview.io/asn/%d/prefixes", asn)
	headers := map[string]string{
		"Accept": "application/json",
	}

	var resp struct {
		Status string `json:"status"`
		Data   struct {
			IPv4Prefixes []struct {
				Prefix string `json:"prefix"`
			} `json:"ipv4_prefixes"`
			IPv6Prefixes []struct {
				Prefix string `json:"prefix"`
			} `json:"ipv6_prefixes"`
		} `json:"data"`
	}

	if err := f.GetJSON(ctx, url, headers, &resp); err != nil {
		return nil, fmt.Errorf("bgpview prefixes for AS%d failed: %w", asn, err)
	}

	var prefixes []string
	for _, p := range resp.Data.IPv4Prefixes {
		prefixes = append(prefixes, p.Prefix)
	}
	for _, p := range resp.Data.IPv6Prefixes {
		prefixes = append(prefixes, p.Prefix)
	}

	return prefixes, nil
}

func ParseASN(asnStr string) int {
	var asn int
	fmt.Sscanf(strings.TrimPrefix(asnStr, "AS"), "%d", &asn)
	return asn
}