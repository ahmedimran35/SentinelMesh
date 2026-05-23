package fetchers

import (
	"context"
	"fmt"
	"strings"
)

type RedditFetcher struct {
	*BaseFetcher
}

type RedditPost struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Permalink   string `json:"permalink"`
	Score       int    `json:"score"`
	NumComments int    `json:"num_comments"`
	Subreddit   string `json:"subreddit"`
	Author      string `json:"author"`
	CreatedUTC  float64 `json:"created_utc"`
	SelfText    string `json:"selftext"`
	Domain      string `json:"domain"`
}

type redditResponse struct {
	Data struct {
		Children []struct {
			Data RedditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

var securitySubreddits = []string{
	"netsec",
	"malware",
	"cybersecurity",
	"blueteamsec",
	"threatintel",
	"infosec",
	"security",
}

func NewRedditFetcher(rateLimit int) *RedditFetcher {
	return &RedditFetcher{BaseFetcher: NewBaseFetcher(rateLimit)}
}

func (f *RedditFetcher) getSubredditPosts(ctx context.Context, subreddit string) ([]RedditPost, error) {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/hot.json?limit=15", subreddit)

	var resp redditResponse
	headers := map[string]string{
		"User-Agent": "SentinelMesh/1.0 (security research tool; contact: local)",
	}

	if err := f.GetJSON(ctx, url, headers, &resp); err != nil {
		return nil, fmt.Errorf("reddit lookup for r/%s failed: %w", subreddit, err)
	}

	var posts []RedditPost
	for _, child := range resp.Data.Children {
		posts = append(posts, child.Data)
	}

	return posts, nil
}

func (f *RedditFetcher) FetchAllSecurityPosts(ctx context.Context) []RedditPost {
	type result struct {
		posts []RedditPost
		err   error
	}
	ch := make(chan result, len(securitySubreddits))
	for _, sub := range securitySubreddits {
		go func(s string) {
			posts, err := f.getSubredditPosts(ctx, s)
			ch <- result{posts, err}
		}(sub)
	}
	var allPosts []RedditPost
	for range securitySubreddits {
		r := <-ch
		if r.err == nil {
			allPosts = append(allPosts, r.posts...)
		}
	}
	return allPosts
}

func (f *RedditFetcher) SearchSecurityPosts(ctx context.Context, keyword string) []RedditPost {
	var matched []RedditPost
	allPosts := f.FetchAllSecurityPosts(ctx)

	keywordLower := strings.ToLower(keyword)
	for _, post := range allPosts {
		if strings.Contains(strings.ToLower(post.Title), keywordLower) ||
			strings.Contains(strings.ToLower(post.SelfText), keywordLower) {
			matched = append(matched, post)
		}
	}
	return matched
}
