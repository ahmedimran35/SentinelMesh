package fetchers

import (
	"context"
	"fmt"
	"net/url"
)

type URLHausFetcher struct {
	*BaseFetcher
}

type URLHausResponse struct {
	QueryStatus string           `json:"query_status"`
	Results     []URLHausResult  `json:"urls,omitempty"`
}

type URLHausResult struct {
	ID          int    `json:"id"`
	URL         string `json:"url"`
	URLStatus   string `json:"url_status"`
	Host        string `json:"host"`
	DateAdded   string `json:"date_added"`
	Threat      string `json:"threat"`
	Tags        []string `json:"tags"`
	Reporter    string `json:"reporter"`
}

func NewURLHausFetcher(rateLimit int) *URLHausFetcher {
	return &URLHausFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

func (f *URLHausFetcher) SearchURL(ctx context.Context, targetURL string) (*URLHausResponse, error) {
	params := url.Values{}
	params.Set("url", targetURL)

	reqURL := "https://urlhaus-api.abuse.ch/v1/url/?" + params.Encode()

	var resp URLHausResponse
	if err := f.GetJSON(ctx, reqURL, nil, &resp); err != nil {
		return nil, fmt.Errorf("urlhaus lookup for %s failed: %w", targetURL, err)
	}

	return &resp, nil
}

func (f *URLHausFetcher) SearchHost(ctx context.Context, host string) (*URLHausResponse, error) {
	params := url.Values{}
	params.Set("host", host)

	reqURL := "https://urlhaus-api.abuse.ch/v1/host/?" + params.Encode()

	var resp URLHausResponse
	if err := f.GetJSON(ctx, reqURL, nil, &resp); err != nil {
		return nil, fmt.Errorf("urlhaus host lookup for %s failed: %w", host, err)
	}

	return &resp, nil
}
