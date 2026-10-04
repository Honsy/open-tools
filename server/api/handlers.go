package api

import (
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"opentools/models"
	"opentools/seed"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Router(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), allowCORS)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/api/nav", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"name": "开物录", "sidebar": sidebar(db)})
	})
	r.GET("/api/home", func(c *gin.Context) { c.JSON(http.StatusOK, home(db)) })
	r.GET("/api/categories/:slug", category(db))
	r.GET("/api/search", search(db))
	r.GET("/api/articles/:id", article(db))
	r.GET("/api/random", randomLink(db))
	r.POST("/api/submit", submit(db))
	r.GET("/go/:id", goOut(db))
	mountPages(r, db)
	mountSEO(r, db)
	return r
}

func allowCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}

type sideItem struct {
	Name     string     `json:"name"`
	Href     string     `json:"href"`
	Icon     string     `json:"icon,omitempty"`
	Tool     bool       `json:"tool,omitempty"`
	Children []sideItem `json:"children,omitempty"`
}

type linkDTO struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Desc         string `json:"desc"`
	Clicks       int    `json:"clicks"`
	CategorySlug string `json:"categorySlug"`
	Slug         string `json:"slug"`
}

type articleCard struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Views     int       `json:"views"`
	CreatedAt time.Time `json:"createdAt"`
}

type tabDTO struct {
	Slug     string        `json:"slug"`
	Name     string        `json:"name"`
	Kind     string        `json:"kind"`
	Links    []linkDTO     `json:"links,omitempty"`
	Articles []articleCard `json:"articles,omitempty"`
}

type tagGroup struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type sectionDTO struct {
	Slug   string     `json:"slug"`
	Name   string     `json:"name"`
	Kind   string     `json:"kind"`
	Tabs   []tabDTO   `json:"tabs,omitempty"`
	Groups []tagGroup `json:"groups,omitempty"`
}

type boardItem struct {
	Title    string `json:"title"`
	Heat     string `json:"heat"`
	Href     string `json:"href"`
	External bool   `json:"external,omitempty"`
}

type board struct {
	Key   string      `json:"key"`
	Name  string      `json:"name"`
	Items []boardItem `json:"items"`
}

func sidebar(db *gorm.DB) []sideItem {
	items := []sideItem{{Name: "自定义", Href: "/"}}
	var sections []models.Category
	db.Where("kind = ? AND show_on_home = ?", "section", true).Order("sort").Find(&sections)
	for _, s := range sections {
		item := sideItem{Name: s.Name, Href: "/c/" + s.Slug}
		if s.ContentKind != "tags" {
			var tabs []models.Category
			db.Where("parent_slug = ?", s.Slug).Order("sort").Find(&tabs)
			for _, t := range tabs {
				item.Children = append(item.Children, sideItem{Name: t.Name, Href: "/c/" + t.Slug})
			}
		}
		items = append(items, item)
	}
	items = append(items,
		sideItem{Name: "最新收录", Href: "/latest"},
		sideItem{Name: "公告", Href: "/notice", Icon: "bell", Tool: true},
		sideItem{Name: "站点排行", Href: "/rank", Icon: "rank", Tool: true},
		sideItem{Name: "网址提交", Href: "/submit", Icon: "plus", Tool: true},
		sideItem{Name: "广告合作", Href: "/cooperate", Icon: "ad", Tool: true},
	)
	return items
}

func home(db *gorm.DB) gin.H {
	label, today := seed.Today(time.Now())
	var sections []models.Category
	db.Where("kind = ? AND show_on_home = ?", "section", true).Order("sort").Find(&sections)
	out := make([]sectionDTO, 0, len(sections))
	for _, s := range sections {
		out = append(out, buildSection(db, s))
	}
	pinned := pinnedLinks(db)
	latest := recentLinks(db, 12)
	popular := popularLinks(db, 12)
	views, live := homeBoards(func() ([]boardItem, []boardItem) {
		return boardFrom(popular), boardFrom(latest)
	})
	boards := make([]board, 0, len(views))
	for _, view := range views {
		boards = append(boards, board{Key: view.Key, Name: view.Name, Items: view.Items})
	}
	return gin.H{
		"todayLabel": label,
		"today":      today,
		"pinned":     toLinks(pinned),
		"sections":   out,
		"latest":     toLinks(latest),
		"popular":    toLinks(popular),
		"boards":     boards,
		"liveHot":    live,
	}
}

func buildSection(db *gorm.DB, s models.Category) sectionDTO {
	dto := sectionDTO{Slug: s.Slug, Name: s.Name, Kind: s.ContentKind}
	if s.ContentKind == "tags" {
		dto.Groups = tagGroups(db)
		return dto
	}
	var tabs []models.Category
	db.Where("parent_slug = ?", s.Slug).Order("sort").Find(&tabs)
	dto.Tabs = make([]tabDTO, 0, len(tabs))
	for _, t := range tabs {
		tab := tabDTO{Slug: t.Slug, Name: t.Name, Kind: t.ContentKind}
		if t.ContentKind == "articles" {
			tab.Articles = articleCards(db)
		} else {
			tab.Links = toLinks(onlineLinks(db, t.Slug))
		}
		dto.Tabs = append(dto.Tabs, tab)
	}
	return dto
}

func category(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cat models.Category
		if err := db.Where("slug = ?", c.Param("slug")).First(&cat).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "没有这个分类"})
			return
		}
		if cat.Kind == "tab" {
			tab := tabDTO{Slug: cat.Slug, Name: cat.Name, Kind: cat.ContentKind}
			if cat.ContentKind == "articles" {
				tab.Articles = articleCards(db)
			} else {
				tab.Links = toLinks(onlineLinks(db, cat.Slug))
			}
			c.JSON(http.StatusOK, sectionDTO{
				Slug: cat.Slug, Name: cat.Name, Kind: cat.ContentKind, Tabs: []tabDTO{tab},
			})
			return
		}
		c.JSON(http.StatusOK, buildSection(db, cat))
	}
}

func search(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		if utf8.RuneCountInString(q) > 40 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "关键词太长"})
			return
		}
		q = strings.NewReplacer("%", "", "_", "").Replace(q)
		if q == "" {
			c.JSON(http.StatusOK, gin.H{"q": "", "links": []linkDTO{}})
			return
		}
		like := "%" + q + "%"
		var links []models.Link
		db.Where("status = ? AND (name LIKE ? OR `desc` LIKE ? OR tags LIKE ? OR alias LIKE ?)", "online", like, like, like, like).
			Order("clicks desc").Limit(50).Find(&links)
		c.JSON(http.StatusOK, gin.H{"q": q, "links": toLinks(links)})
	}
}

func article(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var item models.Article
		if err := db.First(&item, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "没有这篇文章"})
			return
		}
		db.Model(&item).UpdateColumn("views", gorm.Expr("views + ?", 1))
		item.Views++
		c.JSON(http.StatusOK, gin.H{
			"id": item.ID, "title": item.Title, "summary": item.Summary,
			"body": item.Body, "views": item.Views, "createdAt": item.CreatedAt,
		})
	}
}

func randomLink(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var link models.Link
		err := db.Where("status = ? AND (url LIKE ? OR url LIKE ?)", "online", "https://%", "http://%").
			Order("RAND()").First(&link).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "没有可打开的网址"})
			return
		}
		c.JSON(http.StatusOK, toLink(link))
	}
}

func submit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name"`
			URL  string `json:"url"`
			Desc string `json:"desc"`
		}
		form := c.ContentType() == "application/x-www-form-urlencoded" || c.ContentType() == "multipart/form-data"
		if form {
			req.Name = c.PostForm("name")
			req.URL = c.PostForm("url")
			req.Desc = c.PostForm("desc")
		} else if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提交格式不对"})
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.URL = strings.TrimSpace(req.URL)
		req.Desc = strings.TrimSpace(req.Desc)
		fail := func(status int, msg string) {
			if form {
				showSubmit(c, db, status, submitView{Name: req.Name, URL: req.URL, Desc: req.Desc, Error: msg})
				return
			}
			c.JSON(status, gin.H{"error": msg})
		}
		if req.Name == "" || utf8.RuneCountInString(req.Name) > 40 {
			fail(http.StatusBadRequest, "填一个 40 字以内的名字")
			return
		}
		if utf8.RuneCountInString(req.Desc) > 120 {
			fail(http.StatusBadRequest, "简介太长")
			return
		}
		u, err := url.Parse(req.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			fail(http.StatusBadRequest, "网址需要以 http 或 https 开头")
			return
		}
		row := models.Link{
			Name: req.Name, URL: req.URL, Desc: req.Desc,
			Status: "pending", CategorySlug: "inbox", CreatedAt: time.Now(),
		}
		seed.FillSlug(db, &row)
		if err := db.Create(&row).Error; err != nil {
			fail(http.StatusInternalServerError, "没有存上")
			return
		}
		if form {
			c.Redirect(http.StatusSeeOther, "/submit?ok=1")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已收下，审核后才会出现在首页"})
	}
}

func goOut(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var link models.Link
		if err := db.First(&link, c.Param("id")).Error; err != nil || link.Status != "online" {
			c.String(http.StatusNotFound, "没有这个网址")
			return
		}
		u, err := url.Parse(link.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			c.String(http.StatusBadRequest, "这个地址不能跳转")
			return
		}
		db.Model(&models.Link{}).Where("id = ?", link.ID).UpdateColumn("clicks", gorm.Expr("clicks + ?", 1))
		q := u.Query()
		if q.Get("from") == "" {
			if from := requestHost(c); from != "" {
				q.Set("from", from)
			}
		}
		u.RawQuery = q.Encode()
		c.Header("X-Robots-Tag", "noindex, nofollow")
		c.Redirect(http.StatusFound, u.String())
	}
}

func onlineLinks(db *gorm.DB, slug string) []models.Link {
	var links []models.Link
	db.Where("category_slug = ? AND status = ?", slug, "online").Order("sort asc, id asc").Find(&links)
	return links
}

func pinnedLinks(db *gorm.DB) []models.Link {
	var links []models.Link
	db.Where("pinned = ? AND status = ?", true, "online").Order("pin_sort asc").Find(&links)
	return links
}

func recentLinks(db *gorm.DB, n int) []models.Link {
	var links []models.Link
	db.Where("status = ?", "online").Order("created_at desc").Limit(n).Find(&links)
	return links
}

func popularLinks(db *gorm.DB, n int) []models.Link {
	var links []models.Link
	db.Where("status = ? AND url LIKE ?", "online", "https://%").Order("clicks desc, views desc, id asc").Limit(n).Find(&links)
	return links
}

func articleCards(db *gorm.DB) []articleCard {
	var rows []models.Article
	db.Order("created_at desc").Find(&rows)
	out := make([]articleCard, 0, len(rows))
	for _, row := range rows {
		out = append(out, articleCard{
			ID: row.ID, Title: row.Title, Summary: row.Summary, Views: row.Views, CreatedAt: row.CreatedAt,
		})
	}
	return out
}

func tagGroups(db *gorm.DB) []tagGroup {
	var rows []models.Tag
	db.Order("sort asc").Find(&rows)
	order := make([]string, 0)
	bucket := map[string][]string{}
	for _, row := range rows {
		if _, ok := bucket[row.Group]; !ok {
			order = append(order, row.Group)
			bucket[row.Group] = []string{}
		}
		bucket[row.Group] = append(bucket[row.Group], row.Name)
	}
	out := make([]tagGroup, 0, len(order))
	for _, name := range order {
		out = append(out, tagGroup{Name: name, Tags: bucket[name]})
	}
	return out
}

func toLinks(in []models.Link) []linkDTO {
	out := make([]linkDTO, 0, len(in))
	for _, link := range in {
		out = append(out, toLink(link))
	}
	return out
}

func toLink(link models.Link) linkDTO {
	return linkDTO{
		ID: link.ID, Name: link.Name, URL: link.URL, Desc: link.Desc,
		Clicks: link.Clicks, CategorySlug: link.CategorySlug, Slug: link.Slug,
	}
}

func boardFrom(links []models.Link) []boardItem {
	items := make([]boardItem, 0, len(links))
	for _, link := range links {
		if len(items) == 20 {
			break
		}
		href := link.URL
		heat := link.Clicks
		if link.Slug != "" {
			href = "/site/" + link.Slug
		} else if strings.HasPrefix(link.URL, "http") {
			href = "/go/" + itoa(link.ID)
		}
		items = append(items, boardItem{Title: link.Name, Heat: formatCount(heat), Href: href})
	}
	return items
}

func itoa(n uint) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
