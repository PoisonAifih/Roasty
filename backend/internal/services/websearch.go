package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/poisonaifih/roasty/backend/internal/models"
)

var (
	ddgResultRe  = regexp.MustCompile(`(?is)<a[^>]*class="[^"]*result__a[^"]*"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	ddgLiteRe    = regexp.MustCompile(`(?is)<a[^>]*rel="nofollow"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	ddgSnippetRe = regexp.MustCompile(`(?is)<(?:a|td)[^>]*class="[^"]*result__snippet[^"]*"[^>]*>(.*?)</(?:a|td)>`)
	tagRe        = regexp.MustCompile(`(?is)<[^>]+>`)
	spaceRe      = regexp.MustCompile(`\s+`)
)

type shopSearchClient struct {
	http *http.Client
}

func newShopSearchClient() *shopSearchClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	return &shopSearchClient{
		http: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}
}

func (s *ScoutService) FindShops(ctx context.Context, origin, variety string) ([]models.ScoutShop, error) {
	origin = strings.TrimSpace(origin)
	variety = strings.TrimSpace(variety)
	if origin == "" {
		return nil, fmt.Errorf("origin required")
	}

	q := strings.TrimSpace(origin + " " + variety)
	onlineQ := fmt.Sprintf("%s green coffee beans buy", q)
	offlineQ := fmt.Sprintf("%s coffee supplier warehouse OR depot OR farmer", q)

	client := newShopSearchClient()
	onlineHits := client.searchAll(ctx, onlineQ+" site:tokopedia.com OR site:shopee.co.id OR site:bukalapak.com", 5)
	if len(onlineHits) < 2 {
		onlineHits = append(onlineHits, client.searchAll(ctx, onlineQ+" Indonesia marketplace", 5)...)
	}
	offlineHits := client.searchAll(ctx, offlineQ+" Indonesia", 5)

	var out []models.ScoutShop
	seen := map[string]bool{}
	add := func(shop models.ScoutShop) {
		key := strings.ToLower(shop.URL)
		if shop.URL == "" || seen[key] {
			return
		}
		seen[key] = true
		out = append(out, shop)
	}

	for _, h := range onlineHits {
		if isNoiseURL(h.url) {
			continue
		}
		add(models.ScoutShop{
			Name:     h.title,
			Presence: "online",
			URL:      h.url,
			Snippet:  firstNonEmpty(h.snippet, hostLabel(h.url)),
		})
	}
	for _, h := range offlineHits {
		if isNoiseURL(h.url) {
			continue
		}
		add(models.ScoutShop{
			Name:     h.title,
			Presence: "offline",
			URL:      h.url,
			Snippet:  firstNonEmpty(h.snippet, hostLabel(h.url)),
		})
	}

	// Direct marketplace / maps links so the UI always has usable shop destinations.
	for _, shop := range marketplaceLinks(origin, variety) {
		add(shop)
	}

	if len(out) == 0 {
		add(models.ScoutShop{
			Name:     fmt.Sprintf("Google · shops for %s", origin),
			Presence: "online",
			URL:      googleSearchURL(onlineQ + " Indonesia"),
			Snippet:  "Open Google results for sellers.",
		})
	}

	note := s.shopNote(ctx, origin, variety, out)
	for i := range out {
		out[i].Note = note
	}
	return out, nil
}

func marketplaceLinks(origin, variety string) []models.ScoutShop {
	q := strings.TrimSpace(origin + " " + variety + " green bean")
	mapsQ := strings.TrimSpace(origin + " coffee supplier")
	return []models.ScoutShop{
		{
			Name:     "Tokopedia · " + origin,
			Presence: "online",
			URL:      "https://www.tokopedia.com/search?st=product&q=" + url.QueryEscape(q),
			Snippet:  "Search green beans on Tokopedia.",
		},
		{
			Name:     "Shopee · " + origin,
			Presence: "online",
			URL:      "https://shopee.co.id/search?keyword=" + url.QueryEscape(q),
			Snippet:  "Search green beans on Shopee.",
		},
		{
			Name:     "Bukalapak · " + origin,
			Presence: "online",
			URL:      "https://www.bukalapak.com/products?search%5Bkeywords%5D=" + url.QueryEscape(q),
			Snippet:  "Search green beans on Bukalapak.",
		},
		{
			Name:     "Google Maps · suppliers near " + origin,
			Presence: "offline",
			URL:      "https://www.google.com/maps/search/" + url.PathEscape(mapsQ),
			Snippet:  "Find offline suppliers and depots on Maps.",
		},
	}
}

func (s *ScoutService) shopNote(ctx context.Context, origin, variety string, shops []models.ScoutShop) string {
	fallback := fmt.Sprintf(
		"Use the linked marketplaces and maps results to compare online sellers and offline suppliers for %s %s.",
		origin, variety,
	)
	names := make([]string, 0, 4)
	for _, sh := range shops {
		names = append(names, sh.Name)
		if len(names) >= 4 {
			break
		}
	}
	features := fmt.Sprintf(
		"Write in English. origin=%s variety=%s shops=%s",
		origin, variety, strings.Join(names, "; "),
	)
	res, err := s.ai.Infer(ctx, "SHOP", features)
	if err != nil || strings.TrimSpace(res.Text) == "" {
		return fallback
	}
	return strings.TrimSpace(res.Text)
}

type searchHit struct {
	title   string
	url     string
	snippet string
}

func (c *shopSearchClient) searchAll(ctx context.Context, query string, limit int) []searchHit {
	hits := c.searchDDG(ctx, "https://html.duckduckgo.com/html/", query, limit, true)
	if len(hits) == 0 {
		hits = c.searchDDG(ctx, "https://lite.duckduckgo.com/lite/", query, limit, false)
	}
	return dedupeHits(hits, limit)
}

func (c *shopSearchClient) searchDDG(ctx context.Context, endpoint, query string, limit int, post bool) []searchHit {
	build := func() (*http.Request, error) {
		if post {
			form := url.Values{}
			form.Set("q", query)
			form.Set("kl", "id-id")
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return req, nil
		}
		u, _ := url.Parse(endpoint)
		q := u.Query()
		q.Set("q", query)
		u.RawQuery = q.Encode()
		return http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	}

	req, err := build()
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")

	body, err := c.doRead(req)
	if err != nil || body == "" {
		// searchAll already falls back to the lite endpoint, so a failure here
		// simply means this source returned no hits.
		return nil
	}

	return parseDDGBody(body, limit)
}

func (c *shopSearchClient) doRead(req *http.Request) (string, error) {
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("status %d", res.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func parseDDGBody(body string, limit int) []searchHit {
	matches := ddgResultRe.FindAllStringSubmatch(body, limit*3)
	if len(matches) == 0 {
		matches = ddgLiteRe.FindAllStringSubmatch(body, limit*3)
	}
	snippets := ddgSnippetRe.FindAllStringSubmatch(body, limit*3)

	var hits []searchHit
	seen := map[string]bool{}
	for i, m := range matches {
		if len(m) < 3 {
			continue
		}
		link := unwrapDDG(m[1])
		title := cleanHTMLText(m[2])
		if title == "" || link == "" || seen[link] || !strings.HasPrefix(link, "http") {
			continue
		}
		if isNoiseURL(link) {
			continue
		}
		seen[link] = true
		snip := ""
		if i < len(snippets) && len(snippets[i]) > 1 {
			snip = cleanHTMLText(snippets[i][1])
		}
		hits = append(hits, searchHit{title: title, url: link, snippet: snip})
		if len(hits) >= limit {
			break
		}
	}
	return hits
}

func dedupeHits(hits []searchHit, limit int) []searchHit {
	seen := map[string]bool{}
	var out []searchHit
	for _, h := range hits {
		if seen[h.url] {
			continue
		}
		seen[h.url] = true
		out = append(out, h)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func unwrapDDG(raw string) string {
	raw = strings.TrimSpace(htmlUnescape(raw))
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if uddg := u.Query().Get("uddg"); uddg != "" {
		if decoded, err := url.QueryUnescape(uddg); err == nil {
			return decoded
		}
		return uddg
	}
	return raw
}

func isNoiseURL(link string) bool {
	l := strings.ToLower(link)
	noise := []string{
		"duckduckgo.com",
		"google.com/search",
		"javascript:",
		"facebook.com",
		"twitter.com",
		"x.com",
		"youtube.com",
	}
	for _, n := range noise {
		if strings.Contains(l, n) {
			return true
		}
	}
	return false
}

func hostLabel(link string) string {
	u, err := url.Parse(link)
	if err != nil || u.Host == "" {
		return ""
	}
	return "Source: " + strings.TrimPrefix(u.Host, "www.")
}

func googleSearchURL(q string) string {
	return "https://www.google.com/search?q=" + url.QueryEscape(q)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func cleanHTMLText(s string) string {
	s = tagRe.ReplaceAllString(s, "")
	s = htmlUnescape(s)
	s = spaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func htmlUnescape(s string) string {
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&nbsp;", " ",
	)
	return replacer.Replace(s)
}
