package fetchers

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type RSSFetcher struct {
	*BaseFetcher
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Creator     string `xml:"dc:creator"`
	Category    string `xml:"category"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSFeed struct {
	Channel RSSChannel `xml:"channel"`
}

type SecurityNewsItem struct {
	Source      string `json:"source"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Published   string `json:"published"`
	Creator     string `json:"creator,omitempty"`
	Category    string `json:"category,omitempty"`
}

var securityFeeds = []struct {
	Name string
	URL  string
}{
	{Name: "KrebsOnSecurity", URL: "https://krebsonsecurity.com/feed/"},
	{Name: "Schneier", URL: "https://www.schneier.com/feed/atom/"},
	{Name: "TheHackerNews", URL: "https://thehackernews.com/feeds/posts/default"},
	{Name: "BleepingComputer", URL: "https://www.bleepingcomputer.com/feed/"},
}

func NewRSSFetcher(rateLimit int) *RSSFetcher {
	return &RSSFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

func (f *RSSFetcher) FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	f.WaitForRateLimit()

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "SentinelMesh/1.0 (security research)")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rss fetch failed for %s: %w", feedURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("rss parse failed for %s: %w", feedURL, err)
	}

	return &feed, nil
}

func (f *RSSFetcher) FetchAllFeeds(ctx context.Context) []SecurityNewsItem {
	var items []SecurityNewsItem
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, feed := range securityFeeds {
		wg.Add(1)
		go func(name, url string) {
			defer wg.Done()

			feedData, err := f.FetchFeed(ctx, url)
			if err != nil {
				return
			}

			mu.Lock()
			for i, item := range feedData.Channel.Items {
				if i >= 10 {
					break
				}
				items = append(items, SecurityNewsItem{
					Source:      name,
					Title:       item.Title,
					URL:         item.Link,
					Description: stripHTML(item.Description),
					Published:   item.PubDate,
					Creator:     item.Creator,
				})
			}
			mu.Unlock()
		}(feed.Name, feed.URL)
	}

	wg.Wait()
	return items
}

func (f *RSSFetcher) SearchFeeds(ctx context.Context, keyword string) []SecurityNewsItem {
	allItems := f.FetchAllFeeds(ctx)
	var matched []SecurityNewsItem

	keywordLower := strings.ToLower(keyword)
	for _, item := range allItems {
		if strings.Contains(strings.ToLower(item.Title), keywordLower) ||
			strings.Contains(strings.ToLower(item.Description), keywordLower) {
			matched = append(matched, item)
		}
	}
	return matched
}

func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, c := range s {
		if c == '<' {
			inTag = true
			continue
		}
		if c == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(c)
		}
	}
	return strings.TrimSpace(result.String())
}