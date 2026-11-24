package api

import (
	"net"
	"net/http"
	"strings"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
		paths := []string{"/", "/hot", "/latest", "/rank", "/notice", "/rank?category=ai"}
		var month models.InternetRankMonth
		if db.Order("month desc").Take(&month).RowsAffected == 1 {
			var entry models.InternetRankEntry
			if publicInternetRank(db, month.Month).Take(&entry).RowsAffected == 1 {
				paths = append(paths, internetRankPath)
			}
		}
		var cats []models.Category
		db.Select("slug").Order("sort asc, id asc").Find(&cats)
		for _, cat := range cats {
			if cat.Slug != "" {
				paths = append(paths, "/c/"+cat.Slug)
			}
		}
		var articles []models.Article
		db.Select("id").Where("draft = ?", false).Order("id asc").Find(&articles)
		for _, article := range articles {
			paths = append(paths, "/a/"+itoa(article.ID))
		}
		var links []models.Link
		db.Select("slug").Where("status = ? AND slug <> ?", "online", "").Order("id asc").Find(&links)
		for _, link := range links {
			paths = append(paths, "/site/"+link.Slug)
		}
		sitePageData(c, http.StatusOK, "sitemap", nil, paths)
	}
}
