package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"
)

const maxPages = 10

// FetchDetail mengambil & mengekstrak konten lengkap satu artikel dari URL detail.
// Konten utama diekstrak via Readability; selector dipakai untuk author/tags
// dan sebagai fallback. Otomatis mengikuti halaman multipage jika ada.
func FetchDetail(ctx context.Context, articleURL string, sel Selectors) ArticleDetail {
	var (
		allContent strings.Builder
		allHTML    strings.Builder
		author     string
		tags       []string
		excerpt    string
		visited    = map[string]bool{}
		current    = articleURL
		pageCount  = 0
	)

	for current != "" && !visited[current] && pageCount < maxPages {
		visited[current] = true
		pageCount++

		doc, rawHTML, err := httpGetDoc(ctx, current)
		if err != nil {
			if pageCount == 1 {
				return ArticleDetail{Err: err}
			}
			break // halaman lanjutan gagal — gunakan yang sudah terkumpul
		}

		// Readability ekstrak konten utama
		parsedURL, _ := url.Parse(current)
		article, rerr := readability.FromReader(strings.NewReader(rawHTML), parsedURL)
		if rerr == nil {
			if txt := strings.TrimSpace(article.TextContent); txt != "" {
				if allContent.Len() > 0 {
					allContent.WriteString("\n\n")
				}
				allContent.WriteString(txt)
			}
			if article.Content != "" {
				if allHTML.Len() > 0 {
					allHTML.WriteString(`<hr class="page-divider">`)
				}
				allHTML.WriteString(article.Content)
			}
		}

		// Author, tags, excerpt hanya dari halaman pertama
		if pageCount == 1 {
			author = extractAuthor(doc, sel.Author, &article)
			tags = extractTags(doc, sel.Tags)
			if rerr == nil {
				excerpt = strings.TrimSpace(article.Excerpt)
			}
		}

		current = detectNextPage(doc, current, pageCount)
	}

	if allContent.Len() == 0 {
		return ArticleDetail{Err: fmt.Errorf("no content extracted from %s", articleURL)}
	}

	return ArticleDetail{
		Content:     strings.TrimSpace(allContent.String()),
		ContentHTML: strings.TrimSpace(allHTML.String()),
		Excerpt:     excerpt,
		Author:      author,
		Tags:        tags,
	}
}

// httpGetDoc melakukan GET dan mengembalikan goquery.Document + raw HTML.
func httpGetDoc(ctx context.Context, target string) (*goquery.Document, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://www.google.com/")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", err
	}
	html, err := doc.Html()
	if err != nil {
		return nil, "", err
	}
	return doc, html, nil
}

// extractAuthor mengambil nama penulis dari byline Readability atau selector.
func extractAuthor(doc *goquery.Document, selector string, article *readability.Article) string {
	if article != nil && article.Byline != "" {
		return normalizeSpace(article.Byline)
	}
	if selector == "" {
		selector = ".author, .writer, .byline, .reporter, .penulis, .author-name"
	}
	return normalizeSpace(doc.Find(selector).First().Text())
}

// extractTags mengambil tag artikel via selector.
func extractTags(doc *goquery.Document, selector string) []string {
	if selector == "" {
		selector = ".tags a, .tag-links a, .detail-tag a, .detail_tag a, [href*='/tag/']"
	}
	var tags []string
	seen := map[string]bool{}
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if t != "" && !seen[t] {
			seen[t] = true
			tags = append(tags, t)
		}
	})
	return tags
}

// detectNextPage mendeteksi URL halaman berikutnya untuk artikel multipage.
// Mengembalikan "" jika tidak ada halaman lanjutan.
func detectNextPage(doc *goquery.Document, current string, pageCount int) string {
	// Special case: Tribun menyediakan ?page=all
	if strings.Contains(current, "tribunnews.com") && !strings.Contains(current, "page=all") {
		if strings.Contains(current, "?") {
			return current + "&page=all"
		}
		return current + "?page=all"
	}

	var next string
	doc.Find("a").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := strings.ToLower(strings.TrimSpace(s.Text()))
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return true
		}
		if text == "selanjutnya" || text == "next" || text == fmt.Sprint(pageCount+1) {
			if u, err := resolveURL(current, href); err == nil {
				next = u
				return false
			}
		}
		return true
	})
	return next
}

func resolveURL(base, href string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	r, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	return b.ResolveReference(r).String(), nil
}

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
