package scraper

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// FetchFeed mengambil RSS/Atom feed dari feedURL dan mengembalikan daftar item.
func FetchFeed(ctx context.Context, feedURL string) ([]FeedItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned status %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}

	items := make([]FeedItem, 0, len(feed.Items))
	for _, it := range feed.Items {
		if it.Link == "" {
			continue
		}
		fi := FeedItem{
			Link:        it.Link,
			Title:       strings.TrimSpace(it.Title),
			Description: cleanHTML(it.Description),
		}
		if it.PublishedParsed != nil {
			fi.PubDate = it.PublishedParsed
		} else if it.UpdatedParsed != nil {
			fi.PubDate = it.UpdatedParsed
		}
		if len(it.Enclosures) > 0 && it.Enclosures[0].URL != "" {
			fi.Thumbnail = it.Enclosures[0].URL
		} else if it.Image != nil {
			fi.Thumbnail = it.Image.URL
		}
		items = append(items, fi)
	}
	return items, nil
}

// BuildFeedURL menggabungkan base + suffix, atau memakai override jika diisi.
func BuildFeedURL(base, suffix, override string) string {
	if override != "" {
		return override
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(suffix, "/")
}

// cleanHTML membuang tag HTML dari string, menyisakan teks bersih.
func cleanHTML(s string) string {
	if s == "" {
		return ""
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(doc.Text())
}
