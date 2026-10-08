package api

import (
	"html/template"
	"net"
	"net/http"
	"strings"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func mountSEO(r *gin.Engine, db *gorm.DB) {
	r.GET("/sitemap.xml", sitemap(db))
	r.GET("/robots.txt", robots())
}

func crumbLD(items []map[string]string) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for i, item := range items {
		out = append(out, map[string]any{
			"@type":    "ListItem",
			"position": i + 1,
			"name":     item["name"],
			"item":     item["url"],
		})
	}
	return out
}

func joinWords(parts ...string) string {
	seen := map[string]bool{}
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	return strings.Join(out, ",")
}

func requestHost(c *gin.Context) string {
	host := headerFirst(c, "X-Forwarded-Host")
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}
	if name, _, err := net.SplitHostPort(host); err == nil {
		return name
	}
	return host
}

func publicOrigin(c *gin.Context) string {
	proto := headerFirst(c, "X-Forwarded-Proto")
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := headerFirst(c, "X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return proto + "://" + host
}

func headerFirst(c *gin.Context, name string) string {
	value := strings.TrimSpace(c.GetHeader(name))
	if i := strings.Index(value, ","); i >= 0 {
		value = strings.TrimSpace(value[:i])
	}
	return value
}

func sitemap(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := publicOrigin(c)
		var b strings.Builder
		b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
		b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
		writeURL(&b, origin+"/")
		writeURL(&b, origin+"/hot")
		writeURL(&b, origin+"/latest")
		writeURL(&b, origin+"/rank")
		writeURL(&b, origin+"/notice")
		var cats []models.Category
		db.Order("sort asc, id asc").Find(&cats)
		for _, cat := range cats {
			if cat.Slug == "" {
				continue
			}
			writeURL(&b, origin+"/c/"+template.HTMLEscapeString(cat.Slug))
		}
		var articles []models.Article
		db.Order("id asc").Find(&articles)
		for _, article := range articles {
			writeURL(&b, origin+"/a/"+itoa(article.ID))
		}
		var links []models.Link
		db.Where("status = ? AND slug <> ?", "online", "").Order("id asc").Find(&links)
		for _, link := range links {
			writeURL(&b, origin+"/site/"+template.HTMLEscapeString(link.Slug))
		}
		b.WriteString(`</urlset>`)
		c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(b.String()))
	}
}

func writeURL(b *strings.Builder, loc string) {
	b.WriteString("<url><loc>")
	b.WriteString(loc)
	b.WriteString("</loc></url>")
}

func robots() gin.HandlerFunc {
	return func(c *gin.Context) {
		body := "User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /api/\nDisallow: /go/\nDisallow: /search\nDisallow: /submit\nDisallow: /login\nDisallow: /register\nDisallow: /mine\nDisallow: /random\nDisallow: /tools/\nSitemap: " + publicOrigin(c) + "/sitemap.xml\n"
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(body))
	}
}
