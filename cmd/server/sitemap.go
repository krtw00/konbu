package main

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/krtw00/konbu/internal/service"
)

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

func newSitemapHandler(publishSvc *service.PublishService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := absoluteURL(r, "")

		urls := []sitemapURL{
			{Loc: base + "/", LastMod: time.Now().Format("2006-01-02")},
		}

		slugs, err := publishSvc.ListPublicSlugs(r.Context())
		if err == nil {
			for _, s := range slugs {
				if s.ResourceType == "memo" {
					urls = append(urls, sitemapURL{
						Loc:     base + "/memo/" + s.Slug,
						LastMod: s.UpdatedAt.Format("2006-01-02"),
					})
				}
			}
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Write([]byte(xml.Header))
		xml.NewEncoder(w).Encode(sitemapURLSet{
			XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  urls,
		})
	}
}
