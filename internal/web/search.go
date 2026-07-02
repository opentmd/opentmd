package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const searchTimeout = 20 * time.Second

var (
	ddgLinkRe    = regexp.MustCompile(`class="result__a"[^>]*href="([^"]+)"[^>]*>([\s\S]*?)</a>`)
	ddgSnippetRe = regexp.MustCompile(`class="result__snippet"[^>]*>([\s\S]*?)</a>`)
	tagRe        = regexp.MustCompile(`<[^>]+>`)
)

type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

func Search(ctx context.Context, query string, maxResults int) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("empty search query")
	}
	if maxResults <= 0 {
		maxResults = 8
	}
	if maxResults > 20 {
		maxResults = 20
	}

	ctx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()

	form := url.Values{"q": {query}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://html.duckduckgo.com/html/", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; OpenTMD/1.0)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", fmt.Errorf("read search response: %w", err)
	}
	html := string(body)
	if len(html) == 0 {
		return "", fmt.Errorf("empty response for %q", query)
	}

	results := parseDDGResults(html, maxResults)
	if len(results) == 0 {
		return "", fmt.Errorf("no results for %q", query)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Search results for %q:\n\n", query)
	for i, r := range results {
		fmt.Fprintf(&sb, "%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, r.Snippet)
	}
	return strings.TrimSpace(sb.String()), nil
}

func parseDDGResults(html string, max int) []SearchResult {
	matches := ddgLinkRe.FindAllStringSubmatch(html, max)
	if len(matches) == 0 {
		return nil
	}
	snippets := ddgSnippetRe.FindAllStringSubmatch(html, max)
	var out []SearchResult
	for i, m := range matches {
		if len(m) < 3 {
			continue
		}
		title := stripTags(m[2])
		link := decodeDDGURL(m[1])
		snippet := ""
		if i < len(snippets) && len(snippets[i]) > 1 {
			snippet = stripTags(snippets[i][1])
		}
		out = append(out, SearchResult{Title: title, URL: link, Snippet: snippet})
	}
	return out
}

func decodeDDGURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	if strings.Contains(raw, "uddg=") {
		if u, err := url.Parse(raw); err == nil {
			if v := u.Query().Get("uddg"); v != "" {
				if decoded, err := url.QueryUnescape(v); err == nil {
					return decoded
				}
			}
		}
	}
	return raw
}

func stripTags(s string) string {
	s = tagRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	return strings.Join(strings.Fields(s), " ")
}
