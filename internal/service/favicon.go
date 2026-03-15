package service

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	linkIconRe = regexp.MustCompile(`(?i)<link[^>]+rel=["'](?:shortcut )?icon["'][^>]*>`)
	hrefRe     = regexp.MustCompile(`(?i)href=["']([^"']+)["']`)
)

// FetchFavicon fetches a favicon for the given URL.
// It tries: 1) parse HTML for <link rel="icon">, 2) /favicon.ico
// Returns a data URI (data:image/...;base64,...) or empty string on failure.
func FetchFavicon(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// 1) Fetch HTML and look for <link rel="icon">
	if iconURL := findIconInHTML(client, u); iconURL != "" {
		if data := fetchImageAsDataURI(client, iconURL); data != "" {
			return data
		}
	}

	// 2) Try /favicon.ico
	faviconURL := fmt.Sprintf("%s://%s/favicon.ico", u.Scheme, u.Host)
	if data := fetchImageAsDataURI(client, faviconURL); data != "" {
		return data
	}

	// 3) Try Google Favicons API (accepts non-200 since it returns image even on 404)
	googleURL := fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=32", u.Hostname())
	if data := fetchImageAsDataURIAnyStatus(client, googleURL); data != "" {
		return data
	}

	return ""
}

func findIconInHTML(client *http.Client, u *url.URL) string {
	pageURL := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; konbu/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// Read limited HTML (first 64KB is enough for <head>)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return ""
	}
	html := string(body)

	// Find <link rel="icon" href="...">
	match := linkIconRe.FindString(html)
	if match == "" {
		return ""
	}
	hrefMatch := hrefRe.FindStringSubmatch(match)
	if len(hrefMatch) < 2 {
		return ""
	}

	href := hrefMatch[1]
	return resolveURL(u, href)
}

func resolveURL(base *url.URL, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "//") {
		return base.Scheme + ":" + href
	}
	if strings.HasPrefix(href, "/") {
		return fmt.Sprintf("%s://%s%s", base.Scheme, base.Host, href)
	}
	return fmt.Sprintf("%s://%s/%s", base.Scheme, base.Host, href)
}

func fetchImageAsDataURIAnyStatus(client *http.Client, imgURL string) string {
	req, err := http.NewRequest("GET", imgURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; konbu/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		return ""
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil || len(data) == 0 {
		return ""
	}

	return fmt.Sprintf("data:%s;base64,%s", ct, base64.StdEncoding.EncodeToString(data))
}

func fetchImageAsDataURI(client *http.Client, imgURL string) string {
	req, err := http.NewRequest("GET", imgURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; konbu/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		// Try to detect from URL path (strip query params)
		path := imgURL
		if i := strings.Index(path, "?"); i != -1 {
			path = path[:i]
		}
		switch {
		case strings.HasSuffix(path, ".svg"):
			ct = "image/svg+xml"
		case strings.HasSuffix(path, ".png"):
			ct = "image/png"
		case strings.HasSuffix(path, ".ico"):
			ct = "image/x-icon"
		default:
			return ""
		}
	}

	// Limit to 512KB
	data, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil || len(data) == 0 {
		return ""
	}

	return fmt.Sprintf("data:%s;base64,%s", ct, base64.StdEncoding.EncodeToString(data))
}
