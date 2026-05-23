package agents

import (
	"context"
	"fmt"

	"sentinelmesh/fetchers"
	"sentinelmesh/models"
)

type NewsIntelAgent struct {
	hn       *fetchers.HackerNewsFetcher
	github   *fetchers.GitHubFetcher
	rss      *fetchers.RSSFetcher
	reddit   *fetchers.RedditFetcher
}

func NewNewsIntelAgent(rateLimit int) *NewsIntelAgent {
	return &NewsIntelAgent{
		hn:       fetchers.NewHackerNewsFetcher(rateLimit),
		github:   fetchers.NewGitHubFetcher(rateLimit),
		rss:      fetchers.NewRSSFetcher(rateLimit),
		reddit:   fetchers.NewRedditFetcher(rateLimit),
	}
}

func (a *NewsIntelAgent) Name() string        { return "news_intel" }
func (a *NewsIntelAgent) Description() string { return "Security news monitoring and exploit tracking" }

func (a *NewsIntelAgent) Investigate(ctx context.Context, target models.Target, findings chan<- models.Finding) error {
	// Security keywords related to the target
	keywords := []string{
		"security", "vulnerability", "exploit", "CVE", "breach",
		"malware", "ransomware", "zero-day", "0day", "patch",
		target.Value,
	}

	hnItems, err := a.hn.SearchSecurityNews(ctx, keywords, 10)
	if err == nil {
		for _, item := range hnItems {
			details := fmt.Sprintf("Title: %s\nURL: %s\nScore: %d\nComments: %d\nBy: %s",
				item.Title, item.URL, item.Score, item.Descendants, item.By)

			severity := models.RiskInfo
			if item.Score > 100 {
				severity = models.RiskMedium
			}

			findings <- NewFinding("news_intel", "news", severity,
				fmt.Sprintf("HN: %s (↑%d)", item.Title, item.Score),
				details, item)
		}
	}

	repos, err := a.github.SearchExploits(ctx, target.Value)
	if err == nil && len(repos) > 0 {
		for _, repo := range repos {
			details := fmt.Sprintf("Repo: %s\nURL: %s\nStars: %d\nLanguage: %s\nDescription: %s",
				repo.FullName, repo.HTMLURL, repo.Stars, repo.Language, repo.Description)

			severity := models.RiskMedium
			if repo.Stars > 50 {
				severity = models.RiskHigh
			}

			findings <- NewFinding("news_intel", "exploit", severity,
				fmt.Sprintf("GitHub Exploit PoC: %s (★%d)", repo.FullName, repo.Stars),
				details, repo)
		}
	}

	// Also search for general security tools related to target
	if target.Type == "domain" || target.Type == "ip" {
		secRepos, err := a.github.SearchRepos(ctx, target.Value+" security scanner", 5)
		if err == nil {
			for _, repo := range secRepos {
				findings <- NewFinding("news_intel", "tool", models.RiskInfo,
					fmt.Sprintf("Related Security Tool: %s (★%d)", repo.FullName, repo.Stars),
					fmt.Sprintf("URL: %s\nDescription: %s", repo.HTMLURL, repo.Description),
					repo)
			}
		}
	}

	rssItems := a.rss.SearchFeeds(ctx, target.Value)
	for _, item := range rssItems {
		details := fmt.Sprintf("Source: %s\nURL: %s\nPublished: %s\n%s",
			item.Source, item.URL, item.Published, item.Description)

		findings <- NewFinding("news_intel", "news", models.RiskInfo,
			fmt.Sprintf("[%s] %s", item.Source, item.Title),
			details, item)
	}

	redditPosts := a.reddit.SearchSecurityPosts(ctx, target.Value)
	for _, post := range redditPosts {
		details := fmt.Sprintf("Subreddit: r/%s\nAuthor: %s\nScore: %d\nComments: %d\nURL: https://reddit.com%s",
			post.Subreddit, post.Author, post.Score, post.NumComments, post.Permalink)

		severity := models.RiskInfo
		if post.Score > 100 {
			severity = models.RiskMedium
		}

		findings <- NewFinding("news_intel", "discussion", severity,
			fmt.Sprintf("r/%s: %s (↑%d)", post.Subreddit, post.Title, post.Score),
			details, post)
	}

	return nil
}
