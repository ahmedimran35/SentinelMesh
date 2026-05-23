package fetchers

import (
	"context"
	"fmt"
	"net"
)

type InternetDBFetcher struct {
	*BaseFetcher
}

type InternetDBResult struct {
	Ports        []int    `json:"ports"`
	Vulns        []string `json:"vulns"`
	Hostnames    []string `json:"hostnames"`
	CPES         []string `json:"cpes"`
	Tags         []string `json:"tags"`
}

func NewInternetDBFetcher(rateLimit int) *InternetDBFetcher {
	return &InternetDBFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

func (f *InternetDBFetcher) Lookup(ctx context.Context, ip string) (*InternetDBResult, error) {
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}
	lookupURL := fmt.Sprintf("https://internetdb.shodan.io/%s", ip)

	var result InternetDBResult
	if err := f.GetJSON(ctx, lookupURL, nil, &result); err != nil {
		return nil, fmt.Errorf("internetdb lookup for %s failed: %w", ip, err)
	}

	return &result, nil
}
