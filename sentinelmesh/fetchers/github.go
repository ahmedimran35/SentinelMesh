package fetchers

import (
	"context"
	"fmt"
	"net/url"
)

type GitHubFetcher struct {
	*BaseFetcher
}

type GitHubSearchResponse struct {
	TotalCount int            `json:"total_count"`
	Items      []GitHubRepo   `json:"items"`
}

type GitHubRepo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Language    string `json:"language"`
	UpdatedAt   string `json:"updated_at"`
}

func NewGitHubFetcher(rateLimit int) *GitHubFetcher {
	return &GitHubFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

// SearchRepos searches GitHub repos by query
func (f *GitHubFetcher) SearchRepos(ctx context.Context, query string, maxResults int) ([]GitHubRepo, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("sort", "updated")
	params.Set("order", "desc")
	params.Set("per_page", fmt.Sprintf("%d", maxResults))

	reqURL := "https://api.github.com/search/repositories?" + params.Encode()
	headers := map[string]string{
		"Accept": "application/vnd.github.v3+json",
	}

	var resp GitHubSearchResponse
	if err := f.GetJSON(ctx, reqURL, headers, &resp); err != nil {
		return nil, fmt.Errorf("github search for '%s' failed: %w", query, err)
	}

	return resp.Items, nil
}

// SearchExploits searches for exploit PoCs on GitHub
func (f *GitHubFetcher) SearchExploits(ctx context.Context, cveID string) ([]GitHubRepo, error) {
	query := fmt.Sprintf("%s exploit PoC", cveID)
	return f.SearchRepos(ctx, query, 5)
}
