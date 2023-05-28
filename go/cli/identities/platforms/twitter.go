package platforms

import (
	"context"
	"fmt"
	twitterscraper "github.com/n0madic/twitter-scraper"
	"regexp"
)

var _ Platform = &twitterPlatform{}

type twitterPlatform struct {
}

func newTwitterPlatform() *twitterPlatform {
	return &twitterPlatform{}
}

func (tp *twitterPlatform) FetchPublicProof(ctx context.Context, url string) (*PublicProof, error) {
	tweetID, err := getTweetIDFromURL(url)
	if err != nil {
		return nil, err
	}

	scraper := twitterscraper.New()
	tweet, err := scraper.GetTweet(tweetID)
	if err != nil {
		return nil, fmt.Errorf("Error fetching public proof tweet: %s", err)
	}

	// TODO: convert tweet content from html text to text
	// TODO: convert short urls inside the tweet content to full urls

	return &PublicProof{
		Author:  tweet.Username,
		Content: tweet.Text,
	}, nil
}

func getTweetIDFromURL(url string) (string, error) {
	pattern := `^https?://(?:www\.)?twitter\.com/(?:#!/)?[^/]+/status/(\d+).*`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	matches := regex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("Invalid public proof tweet URL")
	}

	tweetID := matches[1]
	return tweetID, nil
}
