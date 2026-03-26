package main

import (
	"errors"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

var markdownSyntaxStripper = regexp.MustCompile(`[#>*_` + "`" + `~\[\]\(\)!-]+`)

type pageMeta struct {
	Title       string
	Description string
	URL         string
	ImageURL    string
}

func newPublishedMemoPageHandler(publishSvc *service.PublishService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		view, err := publishSvc.GetPublicMemoView(r.Context(), slug)
		if err != nil {
			status := http.StatusInternalServerError
			var appErr *apperror.Error
			if errors.As(err, &appErr) {
				status = appErr.HTTPStatus
			}
			_ = serveSPAWithMetadata(w, status, pageMeta{
				Title:       "Published memo not found",
				Description: "This published memo is unavailable.",
				URL:         absoluteURL(r, r.URL.Path),
				ImageURL:    absoluteURL(r, "/hero.png"),
			})
			return
		}

		_ = serveSPAWithMetadata(w, http.StatusOK, buildPublishedMemoPageMeta(r, view))
	}
}

func buildPublishedMemoPageMeta(r *http.Request, view *model.PublishedMemoView) pageMeta {
	title := strings.TrimSpace(view.Publish.Title)
	if title == "" {
		title = strings.TrimSpace(view.Memo.Title)
	}
	if title == "" {
		title = "konbu"
	}

	description := strings.TrimSpace(view.Publish.Description)
	if description == "" {
		description = summarizeMemoContent(view.Memo.Content)
	}
	if description == "" {
		description = "Published with konbu."
	}

	return pageMeta{
		Title:       title,
		Description: description,
		URL:         absoluteURL(r, r.URL.Path),
		ImageURL:    absoluteURL(r, "/hero.png"),
	}
}

func serveSPAWithMetadata(w http.ResponseWriter, status int, meta pageMeta) error {
	indexHTML, err := os.ReadFile("web/static/index.html")
	if err != nil {
		http.Error(w, "failed to load app", http.StatusInternalServerError)
		return err
	}

	body := injectPageMetadata(string(indexHTML), meta)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write([]byte(body))
	return err
}

func injectPageMetadata(indexHTML string, meta pageMeta) string {
	title := template.HTMLEscapeString(meta.Title)
	description := template.HTMLEscapeString(meta.Description)
	url := template.HTMLEscapeString(meta.URL)
	imageURL := template.HTMLEscapeString(meta.ImageURL)

	html := strings.Replace(indexHTML, "<title>konbu</title>", "<title>"+title+" | konbu</title>", 1)
	headTags := strings.Join([]string{
		`<meta name="description" content="` + description + `" />`,
		`<meta property="og:type" content="article" />`,
		`<meta property="og:site_name" content="konbu" />`,
		`<meta property="og:title" content="` + title + `" />`,
		`<meta property="og:description" content="` + description + `" />`,
		`<meta property="og:url" content="` + url + `" />`,
		`<meta property="og:image" content="` + imageURL + `" />`,
		`<meta name="twitter:card" content="summary_large_image" />`,
		`<meta name="twitter:title" content="` + title + `" />`,
		`<meta name="twitter:description" content="` + description + `" />`,
		`<meta name="twitter:image" content="` + imageURL + `" />`,
		`<link rel="canonical" href="` + url + `" />`,
	}, "\n    ")

	jsonLD := `<script type="application/ld+json">{"@context":"https://schema.org","@type":"Article","headline":"` + title + `","description":"` + description + `","url":"` + url + `","image":"` + imageURL + `","publisher":{"@type":"Organization","name":"konbu"}}</script>`

	return strings.Replace(html, "</head>", "    "+headTags+"\n    "+jsonLD+"\n  </head>", 1)
}

func absoluteURL(r *http.Request, path string) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + r.Host + path
}

func summarizeMemoContent(content *string) string {
	if content == nil {
		return ""
	}

	text := strings.TrimSpace(*content)
	if text == "" {
		return ""
	}

	text = markdownSyntaxStripper.ReplaceAllString(text, " ")
	text = strings.Join(strings.Fields(text), " ")
	if len(text) <= 160 {
		return text
	}
	return strings.TrimSpace(text[:157]) + "..."
}
