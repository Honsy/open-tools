package api

import (
	"bytes"
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"opentools/models"
	"opentools/seed"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//go:embed templates/public.html static/site.css static/site.js
var publicFiles embed.FS

var publicPages = template.Must(template.New("public").Funcs(template.FuncMap{
	"inc":     func(i int) int { return i + 1 },
	"iconURL": iconURL,
}).ParseFS(publicFiles, "templates/public.html"))

type shell struct {
	Title       string
	Description string
	Keywords    string
	Canonical   string
	Robots      string
	OGType      string
	Query       string
	JSONLD      template.JS
	Home        bool
	Sidebar     []sideItem
	Content     template.HTML
}

type linkView struct {
	Href   string
	Name   string
	Desc   string
	Letter string
	Host   string
	JSON   string
	Layout string
}

type postView struct {
	ID      uint
	Title   string
	Summary string
	Date    string
	ISO     string
	Views   int
}

type tabView struct {
	Slug  string
	Name  string
	Kind  string
	On    bool
	Links []linkView
	Posts []postView
}

type sectionView struct {
	Slug      string
	Name      string
	Layout    string
	More      string
	HideTitle bool
	Tabs      []tabView
	Groups    []tagGroup
}

type boardView struct {
	Key   string
	Name  string
	On    bool
	Items []boardItem
}

func mountPages(r *gin.Engine, db *gorm.DB) {
	r.GET("/", homePage(db))
	r.GET("/c/:slug", categoryPage(db))
	r.GET("/a/:id", articlePage(db))
	r.GET("/site/:slug", sitePage(db))
	r.GET("/search", searchPage(db))
	r.GET("/submit", submitPage(db))
	r.POST("/submit", submit(db))
	r.GET("/tools/:slug", toolPage(db))
	r.GET("/random", randomRedirect(db))
	r.GET("/static/site.css", func(c *gin.Context) {
		body, _ := publicFiles.ReadFile("static/site.css")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(http.StatusOK, "text/css; charset=utf-8", body)
	})
	mountIcon(r, db)
	r.GET("/favicon.svg", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=86400")
		c.Data(http.StatusOK, "image/svg+xml", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#1d4f91"/><path fill="#f4f4f4" d="M6 8h20v3H6zm3 3h4v13H9zm10 0h4v13h-4z"/></svg>`))
	})
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/favicon.svg")
	})
	r.GET("/static/site.js", func(c *gin.Context) {
		body, _ := publicFiles.ReadFile("static/site.js")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(http.StatusOK, "application/javascript; charset=utf-8", body)
	})
}

func render(c *gin.Context, db *gorm.DB, status int, s shell, name string, data any) {
	if s.Robots == "" {
		s.Robots = "index,follow"
	}
	if s.OGType == "" {
		s.OGType = "website"
	}
	if s.Canonical == "" || s.Canonical[0] == '/' {
		s.Canonical = publicOrigin(c) + s.Canonical
	}
	s.Sidebar = sidebar(db)
	var buf bytes.Buffer
	if err := publicPages.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
		buf.Reset()
		buf.WriteString("<p>页面没有生成。</p>")
	}
	s.Content = template.HTML(buf.String())
	c.Status(status)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := publicPages.ExecuteTemplate(c.Writer, "layout", s); err != nil {
		log.Printf("layout: %v", err)
	}
}

func homePage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		label, today := seed.Today(time.Now())
		var sections []models.Category
		db.Where("kind = ? AND show_on_home = ?", "section", true).Order("sort").Find(&sections)
		views := make([]sectionView, 0, len(sections))
		for _, section := range sections {
			layout := "row"
			if section.Slug == "fun" {
				layout = "tile"
			}
			more := ""
			if section.ContentKind != "tags" {
				more = "/c/" + section.Slug
			}
			views = append(views, makeSection(db, section, layout, more, false))
		}
		pinned := pinnedLinks(db)
		latest := recentLinks(db, 12)
		popular := popularLinks(db, 12)
		boards, liveHot := homeBoards(func() ([]boardItem, []boardItem) {
			return boardFrom(popular), boardFrom(latest)
		})
		origin := publicOrigin(c)
		desc := "开物录收录常用网站，覆盖影视、游戏、工具、查询、学习和 AI。点进名字可以先看介绍，再打开原站。"
		ld, _ := json.Marshal(map[string]any{
			"@context":    "https://schema.org",
			"@type":       "WebSite",
			"name":        "开物录",
			"url":         origin + "/",
			"description": desc,
			"potentialAction": map[string]any{
				"@type": "SearchAction",
				"target": map[string]any{
					"@type":       "EntryPoint",
					"urlTemplate": origin + "/search?q={search_term_string}",
				},
				"query-input": "required name=search_term_string",
			},
		})
		render(c, db, http.StatusOK, shell{
			Title:       "开物录 - 网址导航",
			Description: desc,
			Keywords:    "网址导航,网址大全,开物录,在线工具,小游戏,实用查询,AI工具",
			Canonical:   "/",
			JSONLD:      template.JS(ld),
			Home:        true,
		}, "home", map[string]any{
			"TodayLabel": label,
			"Today":      today,
			"Pinned":     toViews(pinned, "tile"),
			"Latest":     toViews(latest, "mini"),
			"Popular":    toViews(popular, "mini"),
			"Boards":     boards,
			"LiveHot":    liveHot,
			"Sections":   views,
		})
	}
}

func categoryPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cat models.Category
		if err := db.Where("slug = ?", c.Param("slug")).First(&cat).Error; err != nil {
			render(c, db, http.StatusNotFound, shell{
				Title: "没有这个分类 - 开物录", Robots: "noindex", Canonical: "/c/" + c.Param("slug"),
			}, "missing", map[string]string{"Heading": "没有这个分类", "Text": "这个分类不存在，或已经撤下。"})
			return
		}
		var parent models.Category
		if cat.ParentSlug != "" {
			db.Where("slug = ?", cat.ParentSlug).First(&parent)
		}
		var view sectionView
		if cat.Kind == "tab" {
			view = makeSection(db, cat, layoutOf(parent.Slug), "", true)
			view.Slug = cat.Slug
			view.Tabs = []tabView{makeTab(db, cat, layoutOf(parent.Slug), true)}
		} else {
			view = makeSection(db, cat, layoutOf(cat.Slug), "", true)
		}
		intro := describeSection(cat.Name, view)
		crumbs := []map[string]string{{"name": "首页", "url": publicOrigin(c) + "/"}}
		if parent.Slug != "" {
			crumbs = append(crumbs, map[string]string{"name": parent.Name, "url": publicOrigin(c) + "/c/" + parent.Slug})
		}
		crumbs = append(crumbs, map[string]string{"name": cat.Name, "url": publicOrigin(c) + "/c/" + cat.Slug})
		ld, _ := json.Marshal(map[string]any{
			"@context":        "https://schema.org",
			"@type":           "BreadcrumbList",
			"itemListElement": crumbLD(crumbs),
		})
		render(c, db, http.StatusOK, shell{
			Title:       cat.Name + " - 开物录",
			Description: intro,
			Keywords:    joinWords(cat.Name, parent.Name, "网址导航"),
			Canonical:   "/c/" + cat.Slug,
			JSONLD:      template.JS(ld),
		}, "category", map[string]any{
			"Name": cat.Name, "Intro": intro, "Section": view,
			"ParentName": parent.Name, "ParentSlug": parent.Slug,
		})
	}
}

func articlePage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var item models.Article
		if err := db.First(&item, c.Param("id")).Error; err != nil {
			render(c, db, http.StatusNotFound, shell{
				Title: "没有这篇文章 - 开物录", Robots: "noindex", Canonical: "/a/" + c.Param("id"),
			}, "missing", map[string]string{"Heading": "没有这篇文章", "Text": "这篇说明不存在。"})
			return
		}
		db.Model(&item).UpdateColumn("views", gorm.Expr("views + ?", 1))
		item.Views++
		origin := publicOrigin(c)
		canonical := origin + "/a/" + itoa(item.ID)
		ld, _ := json.Marshal(map[string]any{
			"@context":         "https://schema.org",
			"@type":            "Article",
			"headline":         item.Title,
			"description":      item.Summary,
			"datePublished":    item.CreatedAt.Format(time.RFC3339),
			"mainEntityOfPage": canonical,
		})
		render(c, db, http.StatusOK, shell{
			Title:       item.Title + " - 开物录",
			Description: clip(item.Summary, 120),
			Canonical:   "/a/" + itoa(item.ID),
			OGType:      "article",
			JSONLD:      template.JS(ld),
		}, "article", map[string]any{
			"Title": item.Title, "Summary": item.Summary, "Body": item.Body, "Views": item.Views,
			"Date": item.CreatedAt.Format("2006-01-02"), "ISO": item.CreatedAt.Format(time.RFC3339),
		})
	}
}

func sitePage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var link models.Link
		if err := db.Where("slug = ? AND status = ?", c.Param("slug"), "online").First(&link).Error; err != nil {
			render(c, db, http.StatusNotFound, shell{
				Title: "没有这个网站 - 开物录", Robots: "noindex", Canonical: "/site/" + c.Param("slug"),
			}, "missing", map[string]string{"Heading": "没有这个网站", "Text": "这个网址不存在，或还没有上架。"})
			return
		}
		db.Model(&models.Link{}).Where("id = ?", link.ID).UpdateColumn("views", gorm.Expr("views + ?", 1))
		link.Views++
		var cat models.Category
		db.Where("slug = ?", link.CategorySlug).First(&cat)
		var parent models.Category
		if cat.ParentSlug != "" {
			db.Where("slug = ?", cat.ParentSlug).First(&parent)
		}
		var related []models.Link
		db.Where("category_slug = ? AND status = ? AND id <> ?", link.CategorySlug, "online", link.ID).
			Order("sort asc, id asc").Limit(12).Find(&related)
		origin := publicOrigin(c)
		canonical := origin + "/site/" + link.Slug
		desc := strings.TrimSpace(link.Desc)
		if desc == "" {
			desc = link.Name + "，收录在开物录导航。"
		}
		intro := link.Name + "收录在开物录"
		if parent.Name != "" {
			intro += "的" + parent.Name + " / " + cat.Name
		} else if cat.Name != "" {
			intro += "的" + cat.Name
		}
		intro += "。" + desc
		title := link.Name
		if link.Desc != "" {
			title = link.Name + "-" + link.Desc
		}
		tags := seed.SplitTags(link.Tags)
		summary := strings.TrimSpace(link.Body)
		if summary == "" {
			summary = intro
		} else if i := strings.Index(summary, "\n"); i > 0 {
			summary = summary[:i]
		}
		host := link.URL
		iconHost := ""
		letter := "站"
		if rs := []rune(link.Name); len(rs) > 0 {
			letter = string(rs[0])
		}
		if parsed, err := url.Parse(link.URL); err == nil && parsed.Hostname() != "" {
			host = parsed.Hostname()
			if showIcon(host) {
				iconHost = host
			}
		}
		crumbs := []map[string]string{{"name": "首页", "url": origin + "/"}}
		if parent.Slug != "" {
			crumbs = append(crumbs, map[string]string{"name": parent.Name, "url": origin + "/c/" + parent.Slug})
		}
		if cat.Slug != "" {
			crumbs = append(crumbs, map[string]string{"name": cat.Name, "url": origin + "/c/" + cat.Slug})
		}
		crumbs = append(crumbs, map[string]string{"name": link.Name, "url": canonical})
		ld, _ := json.Marshal(map[string]any{
			"@context":    "https://schema.org",
			"@type":       "WebPage",
			"name":        title + " - 开物录",
			"description": summary,
			"url":         canonical,
			"breadcrumb": map[string]any{
				"@type":           "BreadcrumbList",
				"itemListElement": crumbLD(crumbs),
			},
		})
		external := strings.HasPrefix(link.URL, "http://") || strings.HasPrefix(link.URL, "https://")
		scheme := "站内"
		if strings.HasPrefix(link.URL, "https://") {
			scheme = "HTTPS"
		} else if strings.HasPrefix(link.URL, "http://") {
			scheme = "HTTP"
		}
		render(c, db, http.StatusOK, shell{
			Title:       title + " - 开物录",
			Description: clip(summary, 140),
			Keywords:    joinWords(append([]string{link.Name}, append(tags, cat.Name, parent.Name, "网址导航")...)...),
			Canonical:   "/site/" + link.Slug,
			OGType:      "article",
			JSONLD:      template.JS(ld),
		}, "site", map[string]any{
			"Name": link.Name, "Desc": desc, "Body": link.Body, "Views": link.Views, "Clicks": link.Clicks, "ID": link.ID,
			"ViewsText": formatCount(link.Views), "ClicksText": formatCount(link.Clicks),
			"External": external, "URL": link.URL, "Host": host, "IconHost": iconHost, "Letter": letter,
			"Tags": tags, "TagText": strings.Join(tags, "、"),
			"Alias": link.Alias, "Lang": link.Lang, "Region": link.Region,
			"Spare": link.Spare, "ICP": link.ICP, "Scheme": scheme,
			"Date": link.CreatedAt.Format("2006-01-02"), "ISO": link.CreatedAt.Format(time.RFC3339),
			"ParentName": parent.Name, "ParentSlug": parent.Slug,
			"CatName": cat.Name, "CatSlug": cat.Slug,
			"Related": toViews(related, "row"),
		})
	}
}

func searchPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		title := "搜索 - 开物录"
		desc := "在开物录导航里搜索已收录的网站。"
		var links []models.Link
		if q != "" && utf8Count(q) <= 40 {
			safe := strings.NewReplacer("%", "", "_", "").Replace(q)
			like := "%" + safe + "%"
			db.Where("status = ? AND (name LIKE ? OR `desc` LIKE ? OR tags LIKE ? OR alias LIKE ?)", "online", like, like, like, like).
				Order("clicks desc").Limit(50).Find(&links)
			title = q + " - 搜索 - 开物录"
			desc = "开物录导航里和「" + q + "」有关的网站。"
		}
		render(c, db, http.StatusOK, shell{
			Title: title, Description: desc, Canonical: "/search", Robots: "noindex,follow", Query: q,
		}, "search", map[string]any{"Query": q, "Links": toViews(links, "row")})
	}
}

func submitPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		showSubmit(c, db, http.StatusOK, submitView{OK: c.Query("ok") == "1"})
	}
}

type submitView struct {
	Name  string
	URL   string
	Desc  string
	Error string
	OK    bool
}

func showSubmit(c *gin.Context, db *gorm.DB, status int, view submitView) {
	render(c, db, status, shell{
		Title: "提交网站 - 开物录", Description: "把网站提交到开物录导航，审核后才会出现在首页。",
		Canonical: "/submit", Robots: "noindex,follow",
	}, "submit", view)
}

func toolPage(db *gorm.DB) gin.HandlerFunc {
	pages := map[string]struct{ Name, Desc string }{
		"prettier":   {"格式化", "把 JSON、代码整理成可读的样子。"},
		"crypto":     {"加密解密", "摘要、编码和简单加解密。"},
		"hexconvert": {"进制转换", "二、八、十、十六进制互转。"},
		"moment":     {"时间戳", "时间和 Unix 时间戳互转。"},
		"rgb":        {"颜色", "颜色值换算。"},
		"calculator": {"计算器", "四则运算。"},
		"protobuf":   {"Protobuf", "查看 Protobuf 数据。"},
		"rmbconvert": {"人民币大写", "金额转中文大写。"},
	}
	return func(c *gin.Context) {
		item, ok := pages[c.Param("slug")]
		if !ok {
			render(c, db, http.StatusNotFound, shell{
				Title: "没有这个工具 - 开物录", Robots: "noindex", Canonical: "/tools/" + c.Param("slug"),
			}, "missing", map[string]string{"Heading": "没有这个工具", "Text": "这个工具不在列表里。"})
			return
		}
		render(c, db, http.StatusOK, shell{
			Title: item.Name + " - 开物录", Description: item.Desc, Canonical: "/tools/" + c.Param("slug"), Robots: "noindex,follow",
		}, "tool", item)
	}
}

func randomRedirect(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var link models.Link
		err := db.Where("status = ? AND slug <> ?", "online", "").Order("RAND()").First(&link).Error
		c.Header("X-Robots-Tag", "noindex, nofollow")
		if err != nil || link.Slug == "" {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.Redirect(http.StatusFound, "/site/"+link.Slug)
	}
}

func makeSection(db *gorm.DB, cat models.Category, layout, more string, hideTitle bool) sectionView {
	view := sectionView{Slug: cat.Slug, Name: cat.Name, Layout: layout, More: more, HideTitle: hideTitle}
	if cat.Kind == "tab" {
		view.Tabs = []tabView{makeTab(db, cat, layout, true)}
		return view
	}
	if cat.ContentKind == "tags" {
		view.Groups = tagGroups(db)
		return view
	}
	var tabs []models.Category
	db.Where("parent_slug = ?", cat.Slug).Order("sort").Find(&tabs)
	view.Tabs = make([]tabView, 0, len(tabs))
	for i, tab := range tabs {
		view.Tabs = append(view.Tabs, makeTab(db, tab, layout, i == 0))
	}
	return view
}

func makeTab(db *gorm.DB, tab models.Category, layout string, on bool) tabView {
	view := tabView{Slug: tab.Slug, Name: tab.Name, Kind: tab.ContentKind, On: on}
	if tab.ContentKind == "articles" {
		for _, row := range articleCards(db) {
			view.Posts = append(view.Posts, postView{
				ID: row.ID, Title: row.Title, Summary: row.Summary, Views: row.Views,
				Date: row.CreatedAt.Format("2006-01-02"), ISO: row.CreatedAt.Format(time.RFC3339),
			})
		}
		return view
	}
	view.Links = toViews(onlineLinks(db, tab.Slug), layout)
	return view
}

func toViews(links []models.Link, layout string) []linkView {
	out := make([]linkView, 0, len(links))
	for _, link := range links {
		raw, _ := json.Marshal(map[string]any{
			"id": link.ID, "name": link.Name, "url": link.URL, "desc": link.Desc,
			"slug": link.Slug, "categorySlug": link.CategorySlug, "clicks": link.Clicks,
		})
		host := ""
		if parsed, err := url.Parse(link.URL); err == nil {
			host = parsed.Hostname()
		}
		if !showIcon(host) {
			host = ""
		}
		href := link.URL
		if link.Slug != "" {
			href = "/site/" + link.Slug
		}
		letter := "站"
		if rs := []rune(link.Name); len(rs) > 0 {
			letter = string(rs[0])
		}
		out = append(out, linkView{
			Href: href, Name: link.Name, Desc: link.Desc, Letter: letter, Host: host, JSON: string(raw), Layout: layout,
		})
	}
	return out
}

func layoutOf(slug string) string {
	if slug == "fun" {
		return "tile"
	}
	return "row"
}

func describeSection(name string, view sectionView) string {
	names := make([]string, 0, 8)
	for _, tab := range view.Tabs {
		for _, link := range tab.Links {
			if len(names) == 8 {
				break
			}
			names = append(names, link.Name)
		}
	}
	if len(names) == 0 {
		parts := make([]string, 0, len(view.Tabs))
		for _, tab := range view.Tabs {
			parts = append(parts, tab.Name)
		}
		if len(parts) == 0 {
			return name + "，收录在开物录导航。"
		}
		return name + "包括" + strings.Join(parts, "、") + "。"
	}
	return name + "收录了" + strings.Join(names, "、") + "。"
}

func clip(s string, n int) string {
	rs := []rune(strings.TrimSpace(s))
	if len(rs) <= n {
		return string(rs)
	}
	return string(rs[:n-1]) + "…"
}

func utf8Count(s string) int {
	return len([]rune(s))
}

func formatCount(n int) string {
	if n < 0 {
		n = 0
	}
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	lead := len(s) % 3
	if lead == 0 {
		lead = 3
	}
	b.WriteString(s[:lead])
	for i := lead; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
