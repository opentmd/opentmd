package web

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	fetchTimeout     = 20 * time.Second
	fetchConnect     = 5 * time.Second
	maxResponseBytes = 2 << 20
	maxRedirects     = 5
	defaultMaxChars  = 20000
)

var (
	blockTagRe = regexp.MustCompile(`(?is)<script[^>]*>[\s\S]*?</script>`)
	styleTagRe = regexp.MustCompile(`(?is)<style[^>]*>[\s\S]*?</style>`)
	tagStripRe = regexp.MustCompile(`<[^>]+>`)
)

func Fetch(ctx context.Context, rawURL string, maxChars int) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("empty url")
	}
	if maxChars <= 0 {
		maxChars = defaultMaxChars
	}
	if maxChars > 50000 {
		maxChars = 50000
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("scheme %q not allowed — only http(s)", u.Scheme)
	}

	client := &http.Client{
		Timeout: fetchTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	var hops int
	for {
		if err := validateHost(ctx, u); err != nil {
			return "", err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; OpenTMD/web_fetch)")

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("fetch %s: %w", u, err)
		}

		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			loc := resp.Header.Get("Location")
			_ = resp.Body.Close()
			if loc == "" || hops >= maxRedirects {
				return "", fmt.Errorf("redirect loop or missing Location from %s", u)
			}
			next, err := u.Parse(loc)
			if err != nil {
				return "", fmt.Errorf("bad redirect %q: %w", loc, err)
			}
			u = next
			hops++
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return "", fmt.Errorf("HTTP %d from %s", resp.StatusCode, u)
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
		_ = resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("read body: %w", err)
		}
		if len(body) == 0 {
			return "", fmt.Errorf("empty response from %s", u)
		}

		text := string(body)
		ct := strings.ToLower(resp.Header.Get("Content-Type"))
		if strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml") ||
			(!strings.Contains(ct, "/") && strings.HasPrefix(strings.TrimSpace(text), "<")) {
			text = htmlToText(text)
		}

		if len(text) > maxChars {
			text = truncateRunes(text, maxChars) + "\n… (truncated)"
		}
		return fmt.Sprintf("Fetched %s (%d chars):\n\n%s", u, len(text), text), nil
	}
}

func validateHost(ctx context.Context, u *url.URL) error {
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing host")
	}
	if ip := net.ParseIP(host); ip != nil {
		return checkIP(ip)
	}

	resolver := net.DefaultResolver
	ctx, cancel := context.WithTimeout(ctx, fetchConnect)
	defer cancel()
	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("dns lookup %s: %w", host, err)
	}
	for _, addr := range ips {
		if err := checkIP(addr.IP); err != nil {
			return err
		}
	}
	return nil
}

func checkIP(ip net.IP) error {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return fmt.Errorf("refusing to connect to %s (SSRF protection)", ip)
	}
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 0 || v4[0] >= 240 {
			return fmt.Errorf("refusing to connect to %s (reserved range)", ip)
		}
		if v4[0] == 100 && (v4[1]&0xc0) == 64 {
			return fmt.Errorf("refusing to connect to %s (CGNAT)", ip)
		}
	}
	return nil
}

func htmlToText(html string) string {
	html = blockTagRe.ReplaceAllString(html, "")
	html = styleTagRe.ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "<br>", "\n")
	html = strings.ReplaceAll(html, "<br/>", "\n")
	html = strings.ReplaceAll(html, "<br />", "\n")
	html = strings.ReplaceAll(html, "</p>", "\n\n")
	html = strings.ReplaceAll(html, "</div>", "\n")
	html = tagStripRe.ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	lines := strings.Split(html, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}
