package fetchers

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type HackerNewsFetcher struct {
	*BaseFetcher
}

type HNItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Score       int    `json:"score"`
	By          string `json:"by"`
	Time        int    `json:"time"`
	Descendants int    `json:"descendants"`
	Type        string `json:"type"`
}

func NewHackerNewsFetcher(rateLimit int) *HackerNewsFetcher {
	return &HackerNewsFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

// SearchSecurityNews gets top stories and filters for security-related ones
func (f *HackerNewsFetcher) SearchSecurityNews(ctx context.Context, keywords []string, maxItems int) ([]HNItem, error) {
	// Get top story IDs
	var storyIDs []int
	if err := f.GetJSON(ctx, "https://hacker-news.firebaseio.com/v0/topstories.json", nil, &storyIDs); err != nil {
		return nil, fmt.Errorf("failed to get HN top stories: %w", err)
	}

	if len(storyIDs) > 50 {
		storyIDs = storyIDs[:50]
	}

	// Fetch stories concurrently
	var mu sync.Mutex
	var items []HNItem
	var wg sync.WaitGroup

	sem := make(chan struct{}, 5) // limit concurrency

	for _, id := range storyIDs {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
			var item HNItem
			if err := f.GetJSON(ctx, url, nil, &item); err != nil {
				return
			}

			// Filter for security-related content
			titleLower := strings.ToLower(item.Title)
			for _, kw := range keywords {
				if strings.Contains(titleLower, strings.ToLower(kw)) {
					mu.Lock()
					items = append(items, item)
					mu.Unlock()
					return
				}
			}
		}(id)
	}

	wg.Wait()

	if len(items) > maxItems {
		items = items[:maxItems]
	}

	return items, nil
}
