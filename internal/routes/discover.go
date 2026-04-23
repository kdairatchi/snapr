package routes

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Route struct {
	URL  string
	Name string
}

// FromSitemap fetches /sitemap.xml and returns all URLs.
func FromSitemap(base string) ([]Route, error) {
	target := strings.TrimRight(base, "/") + "/sitemap.xml" //nolint:noctx
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(target)
	if err != nil {
		return nil, fmt.Errorf("fetch sitemap: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sitemap %d", resp.StatusCode)
	}
	return parseSitemap(resp.Body, base)
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}
type sitemap struct {
	URLs []sitemapURL `xml:"url"`
}

func parseSitemap(r io.Reader, base string) ([]Route, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse base: %w", err)
	}
	var sm sitemap
	if err := xml.NewDecoder(r).Decode(&sm); err != nil {
		return nil, fmt.Errorf("decode sitemap: %w", err)
	}
	var routes []Route
	for _, u := range sm.URLs {
		parsed, err := url.Parse(u.Loc)
		if err != nil || parsed.Host != baseURL.Host {
			continue // skip off-host URLs to prevent SSRF via malicious sitemaps
		}
		routes = append(routes, Route{
			URL:  u.Loc,
			Name: urlToName(u.Loc, base),
		})
	}
	return routes, nil
}

// Crawl visits a URL via chromedp and collects same-origin links.
func Crawl(ctx context.Context, base string, maxPages int) ([]Route, error) {
	// NoSandbox required for containerized/CI environments where kernel user namespaces are unavailable.
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.Flag("no-zygote", true),
	)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	taskCtx, cancelTimeout := context.WithTimeout(taskCtx, 5*time.Minute)
	defer cancelTimeout()

	seen := map[string]bool{base: true}
	queue := []string{base}
	var found []Route

	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	for len(queue) > 0 && len(found) < maxPages {
		current := queue[0]
		queue = queue[1:]
		found = append(found, Route{URL: current, Name: urlToName(current, base)})

		var hrefs []string
		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(current),
			chromedp.WaitReady("body", chromedp.ByQuery),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a=>a.href)`, &hrefs),
		); err != nil {
			continue
		}

		for _, href := range hrefs {
			parsed, err := url.Parse(href)
			if err != nil || parsed.Host != baseURL.Host {
				continue
			}
			clean := parsed.Scheme + "://" + parsed.Host + parsed.Path
			if !seen[clean] {
				seen[clean] = true
				queue = append(queue, clean)
			}
		}
	}
	return found, nil
}

func urlToName(rawURL, base string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "page"
	}
	path := parsed.Path
	if path == "" || path == "/" {
		return "index"
	}
	name := strings.Trim(path, "/")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}
