package api

import (
	"encoding/json"
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

type shell struct {
	Title       string
	Description string
	Keywords    string
	Canonical   string
	Robots      string
	OGType      string
	Query       string
	JSONLD      json.RawMessage
	Home        bool
	Path        string
	Sidebar     []sideItem
	Utils       []utilMenu
	Launcher    []sideItem
	IconBase    string
	UserName    string
	IsAdmin     bool
	Account     bool
}

type utilLink struct {
	Name string
	Href string
}

type utilMenu struct {
	Name  string
	Slug  string
	Links []utilLink
}

type linkView struct {
	Href   string
	Name   string
	Desc   string
	Letter string
	Host   string
	Icon   string
	JSON   string
	Layout string
	Clicks int
	Star   string
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
	Slug      string
	Name      string
	Kind      string
	On        bool
	Links     []linkView
	Posts     []postView
	Total     int
	Page      int
	PageCount int
	PrevURL   string
	NextURL   string
	Pages     []pageLink
	Topics    []string
	Topic     string
}

type pageLink struct {
	Number  int
	URL     string
	Current bool
}

type searchView struct {
	Query     string
	Links     []linkView
	Total     int
	Page      int
	PageCount int
	PrevURL   string
	NextURL   string
	Pages     []pageLink
}

type sectionView struct {
	Slug      string
	Name      string
	Layout    string
	More      string
	HideTitle bool
	Fold      bool
	Tabs      []tabView
	Groups    []tagGroup
}

type categoryGuide struct {
	Title string
	Intro string
}

var categoryGuides = map[string]categoryGuide{
	"chat": {
		Title: "AI 对话助手",
		Intro: "写提纲、读长文、查资料和解释代码，可以从 AI 对话助手开始。读文档前先确认能上传的格式和大小，查资料时优先选能列出来源的工具，并打开原文核对。",
	},
	"draw": {
		Title: "AI 绘画与视频生成",
		Intro: "做插画、商品图或短视频时，先按文生图、图生图或图生视频挑工具。需要交付成品时，生成前先看导出尺寸、水印和商用条款。",
	},
	"webtool": {
		Title: "在线网页工具",
		Intro: "临时压缩图片、格式化 JSON、换算时间戳或生成二维码，可以在浏览器里完成。处理客户文件、密钥和日志时，先确认工具是否会上传数据。",
	},
	"query": {
		Title: "生活实用查询",
		Intro: "按任务找查询入口，出行可查车次、航班和天气，寄件可查物流，核对企业信息可查公示平台。涉及办事或资质核验时，以对应官方入口为准。",
	},
	"design": {
		Title: "设计素材与在线设计",
		Intro: "找图标、配色、作品参考或在线协作工具，可以按项目需要挑选。下载素材前核对单个作品的授权和署名要求，交付前确认能否导出所需格式。",
	},
	"learn": {
		Title: "学习网站与公开课",
		Intro: "想系统学习可从大学公开课入手，查 Web API 可看开发文档，刚开始编程可跟着短示例练习。选课时先看先修要求、授课语言和课后练习。",
	},
}

type boardView struct {
	Key   string
	Name  string
	On    bool
	Items []boardItem
}

func mountSiteAPI(r *gin.RouterGroup, db *gorm.DB) {
	r.GET("/", homePage(db))
	r.GET("/c/:slug", categoryPage(db))
	r.GET("/latest", latestPage(db))
	r.GET("/hot", hotPage(db))
	r.GET("/notice", noticePage(db))
	r.GET("/rank", rankPage(db))
	r.GET("/cooperate", cooperatePage(db))
	r.GET("/a/:id", articlePage(db))
	mountMovedContent(r, db)
	r.GET("/site/:slug", sitePage(db))
	r.GET("/search", searchPage(db))
	r.GET("/login", loginPage(db))
	r.POST("/login", loginPost(db))
	r.GET("/register", registerPage(db))
	r.POST("/register", registerPost(db))
	r.GET("/forgot", forgotPage(db))
	r.POST("/forgot", forgotPost(db))
	r.GET("/reset", resetPage(db))
	r.POST("/reset", resetPost(db))
	r.POST("/logout", logout)
	r.POST("/mine/email", saveEmail(db))
	r.GET("/mine", minePage(db))
	r.GET("/submit", submitPage(db))
	r.POST("/submit", submit(db))
	r.GET("/tools/", toolIndex(db))
	r.GET("/tools/:slug", toolPage(db))
	r.GET("/random", randomRedirect(db))
	r.GET("/sitemap.xml", sitemap(db))
}

const siteAPIPrefix = "/_site"
const sitePageType = "application/vnd.kaiwulu.page.v1+json"

type siteResponse struct {
	Template string `json:"template"`
	Shell    *shell `json:"shell,omitempty"`
	Data     any    `json:"data"`
}

func sitePageData(c *gin.Context, status int, name string, layout *shell, data any) {
	c.Header("Content-Type", sitePageType)
	c.JSON(status, siteResponse{Template: name, Shell: layout, Data: data})
}

func sendSitePage(c *gin.Context, db *gorm.DB, status int, s shell, name string, data any) {
	if s.Robots == "" {
		s.Robots = "index,follow"
	}
	if s.OGType == "" {
		s.OGType = "website"
	}
	if s.Canonical == "" || s.Canonical[0] == '/' {
		s.Canonical = publicOrigin(c) + s.Canonical
	}
	s.Path = strings.TrimPrefix(c.Request.URL.Path, siteAPIPrefix)
	s.IconBase = iconBase()
	if user, ok := currentUser(c, db); ok {
		s.UserName = user.Username
		s.IsAdmin = user.IsAdmin
		c.Header("Cache-Control", "private, no-store")
	}
	s.Sidebar = sidebar(db)
	s.Utils = utilMenus(db)
	s.Launcher = launcherItems()
	sitePageData(c, status, name, &s, data)
}

func homePage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		label, today := seed.Today(time.Now())
		if len(today) == 1 && today[0].Year == 0 {
			today = nil
		}
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
			view := makeSection(db, section, layout, more, false)
			view.Fold = section.ContentKind == "tags"
			for i := range view.Tabs {
				if view.Tabs[i].On {
					paginatePreview(&view.Tabs[i])
				}
			}
			views = append(views, view)
		}
		pinned := pinnedLinks(db)
		latest := recentLinks(db, 12)
		popular := popularLinks(db, 12)
		boards, liveHot := homeBoards(func() ([]boardItem, []boardItem) {
			return boardFrom(popular), boardFrom(latest)
		})
		latest = firstModelLinks(latest, 12)
		popular = firstModelLinks(popular, 12)
		for i := range boards {
			if len(boards[i].Items) > 8 {
				boards[i].Items = boards[i].Items[:8]
			}
		}
		origin := publicOrigin(c)
		desc := "查找 AI 助手、在线工具、设计素材、学习资源和生活查询入口。按用途浏览网站，先看介绍，再打开原站。"
		var tasks []sideItem
		for _, section := range views {
			for _, tab := range section.Tabs {
				if guide, ok := categoryGuides[tab.Slug]; ok {
					tasks = append(tasks, sideItem{Name: guide.Title, Href: "/c/" + tab.Slug})
				}
			}
		}
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
		sendSitePage(c, db, http.StatusOK, shell{
			Title:       "开物录 - 常用网站与在线工具导航",
			Description: desc,
			Keywords:    "网址导航,网址大全,开物录,在线工具,小游戏,实用查询,AI工具",
			Canonical:   "/",
			JSONLD:      ld,
			Home:        true,
		}, "home", map[string]any{
			"Intro":      desc,
			"Tasks":      tasks,
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

func latestPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		intro := "最新收录按入库时间排列，热门网址按站内打开次数排列。点名字先看介绍，再打开原站。"
		sendSitePage(c, db, http.StatusOK, shell{
			Title:       "最新收录 - 开物录",
			Description: intro,
			Keywords:    "最新收录,热门网址,开物录",
			Canonical:   "/latest",
		}, "latest", map[string]any{
			"Intro":   intro,
			"Latest":  toViews(recentLinks(db, 48), "mini"),
			"Popular": toViews(popularLinks(db, 48), "mini"),
		})
	}
}

func hotPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		boards, live := homeBoards(func() ([]boardItem, []boardItem) {
			return boardFrom(popularLinks(db, 12)), boardFrom(recentLinks(db, 12))
		})
		intro := "今日热榜列出百度、微博、知乎、哔哩哔哩、抖音和站内正在被打开的名字，每条榜单最多 30 条。"
		if !live {
			intro = "实时热搜暂时取不到。这里是站内点击和刚收录的网站。"
		}
		sendSitePage(c, db, http.StatusOK, shell{
			Title:       "今日热榜 - 开物录",
			Description: intro,
			Keywords:    "今日热榜,微博热搜,百度热搜,知乎热榜,开物录",
			Canonical:   "/hot",
		}, "hot", map[string]any{
			"Intro":  intro,
			"Boards": boards,
		})
	}
}

func noticePage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		intro := "开物录的站点说明。收录、审核和页面调整记在这里。"
		sendSitePage(c, db, http.StatusOK, shell{
			Title: "公告 - 开物录", Description: intro, Keywords: "公告,开物录", Canonical: "/notice",
		}, "notice", map[string]string{"Intro": intro})
	}
}

func rankPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		intro := "站点排行按从本站打开的次数排列。次数相同的，浏览多的靠前。点名字先看介绍，再打开原站。"
		title, canonical := "站点排行", "/rank"
		var rows []models.Link
		isAI := c.Query("category") == "ai"
		if isAI {
			title, canonical = "AI 站点排行", "/rank?category=ai"
			categories := db.Model(&models.Category{}).Select("slug").Where("parent_slug = ?", "ai")
			db.Where("status = ? AND category_slug IN (?)", "online", categories).
				Order("clicks desc, views desc, sort asc, id asc").Limit(60).Find(&rows)
		} else {
			rows = popularLinks(db, 60)
		}
		sendSitePage(c, db, http.StatusOK, shell{
			Title: title + " - 开物录", Description: intro, Keywords: "站点排行,热门网站,开物录", Canonical: canonical,
		}, "rank", map[string]any{
			"Intro": intro, "Title": title, "AI": isAI,
			"Links": toViews(rows, "row"),
		})
	}
}

func cooperatePage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		intro := "开物录是个人整理的网址录，不售广告位，也不做付费收录。"
		sendSitePage(c, db, http.StatusOK, shell{
			Title: "广告合作 - 开物录", Description: intro, Keywords: "开物录", Canonical: "/cooperate",
		}, "cooperate", map[string]string{"Intro": intro})
	}
}

func utilMenus(db *gorm.DB) []utilMenu {
	specs := []struct{ Name, Slug string }{
		{"邮箱", "mail"},
		{"网盘", "disk"},
		{"翻译", "translate"},
		{"地图", "map"},
	}
	out := make([]utilMenu, 0, len(specs))
	for _, spec := range specs {
		links := onlineLinks(db, spec.Slug)
		items := make([]utilLink, 0, len(links))
		for _, link := range links {
			if !strings.HasPrefix(link.URL, "http") {
				continue
			}
			items = append(items, utilLink{Name: link.Name, Href: link.URL})
		}
		out = append(out, utilMenu{Name: spec.Name, Slug: spec.Slug, Links: items})
	}
	return out
}

func launcherItems() []sideItem {
	return []sideItem{
		{Name: "首页", Href: "/", Icon: "home"},
		{Name: "简约模式", Href: "#simple", Icon: "simple"},
		{Name: "在线影视", Href: "/c/video", Icon: "video"},
		{Name: "直播电视", Href: "/c/live", Icon: "live"},
		{Name: "次元动漫", Href: "/c/anime", Icon: "anime"},
		{Name: "在线游戏", Href: "/c/game", Icon: "game"},
		{Name: "音乐网站", Href: "/c/music", Icon: "music"},
		{Name: "免费漫画", Href: "/c/comic", Icon: "comic"},
		{Name: "看小说", Href: "/c/novel", Icon: "novel"},
		{Name: "图片壁纸", Href: "/c/wallpaper", Icon: "wall"},
		{Name: "绿色软件", Href: "/c/software", Icon: "soft"},
		{Name: "资源搜索", Href: "/c/res", Icon: "res"},
		{Name: "网页工具", Href: "/c/webtool", Icon: "tool"},
		{Name: "实用查询", Href: "/c/query", Icon: "query"},
		{Name: "学习教程", Href: "/c/learn", Icon: "learn"},
		{Name: "素材创意", Href: "/c/design", Icon: "design"},
		{Name: "趣味酷站", Href: "/c/cool", Icon: "cool"},
		{Name: "网址集", Href: "/c/collection", Icon: "links"},
	}
}

func categoryPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cat models.Category
		if err := db.Where("slug = ?", c.Param("slug")).First(&cat).Error; err != nil {
			sendSitePage(c, db, http.StatusNotFound, shell{
				Title: "没有这个分类 - 开物录", Robots: "noindex", Canonical: "/c/" + c.Param("slug"),
			}, "missing", map[string]string{"Heading": "没有这个分类", "Text": "这个分类不存在，或已经撤下。"})
			return
		}
		var parent models.Category
		if cat.ParentSlug != "" {
			db.Where("slug = ?", cat.ParentSlug).First(&parent)
		}
		var view sectionView
		categoryPageNum := 1
		if cat.Kind == "tab" && cat.ContentKind == "articles" {
			tab := articleCategoryTab(db, cat, positivePage(c.Query("page")), clip(strings.TrimSpace(c.Query("topic")), 64))
			categoryPageNum = tab.Page
			view = sectionView{Slug: cat.Slug, Name: cat.Name, HideTitle: true, Tabs: []tabView{tab}}
		} else if cat.Kind == "tab" {
			view = sectionView{Slug: cat.Slug, Name: cat.Name, Layout: "row", HideTitle: true}
			page := positivePage(c.Query("page"))
			const pageSize = 24
			var total int64
			db.Model(&models.Link{}).Where("category_slug = ? AND status = ?", cat.Slug, "online").Count(&total)
			pageCount := int((total + pageSize - 1) / pageSize)
			if pageCount < 1 {
				pageCount = 1
			}
			if page > pageCount {
				page = pageCount
			}
			categoryPageNum = page
			var rows []models.Link
			db.Where("category_slug = ? AND status = ?", cat.Slug, "online").
				Order("sort asc, id asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows)
			tab := tabView{Slug: cat.Slug, Name: cat.Name, On: true, Links: toViews(rows, "row"), Total: int(total), Page: page, PageCount: pageCount}
			for n := 1; n <= pageCount; n++ {
				pageURL := "/c/" + cat.Slug
				if n > 1 {
					pageURL += "?page=" + strconv.Itoa(n)
				}
				tab.Pages = append(tab.Pages, pageLink{Number: n, URL: pageURL, Current: n == page})
			}
			if page > 1 {
				tab.PrevURL = "/c/" + cat.Slug
				if page > 2 {
					tab.PrevURL += "?page=" + strconv.Itoa(page-1)
				}
			}
			if page < pageCount {
				tab.NextURL = "/c/" + cat.Slug + "?page=" + strconv.Itoa(page+1)
			}
			view.Tabs = []tabView{tab}
		} else {
			view = makeSection(db, cat, layoutOf(cat.Slug), "", true)
			if len(view.Tabs) > 0 && view.Tabs[0].On {
				paginatePreview(&view.Tabs[0])
				view.More = "/c/" + view.Tabs[0].Slug
			}
		}
		title, intro := cat.Name, ""
		if guide, ok := categoryGuides[cat.Slug]; ok {
			title, intro = guide.Title, guide.Intro
		}
		description := intro
		if description == "" {
			description = describeSection(cat.Name, view)
		}
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
		canonical := "/c/" + cat.Slug
		params := url.Values{}
		if categoryPageNum > 1 {
			params.Set("page", strconv.Itoa(categoryPageNum))
		}
		if cat.ContentKind == "articles" && len(view.Tabs) > 0 && view.Tabs[0].Topic != "" {
			params.Set("topic", view.Tabs[0].Topic)
		}
		if len(params) > 0 {
			canonical += "?" + params.Encode()
		}
		sendSitePage(c, db, http.StatusOK, shell{
			Title:       title + " - 开物录",
			Description: description,
			Keywords:    joinWords(cat.Name, parent.Name, "网址导航"),
			Canonical:   canonical,
			JSONLD:      ld,
		}, "category", map[string]any{
			"Name": cat.Name, "Heading": title, "Intro": intro, "Section": view,
			"ParentName": parent.Name, "ParentSlug": parent.Slug,
		})
	}
}

func positivePage(raw string) int {
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}
	return page
}

// Homepage and parent-category previews link to the same paginated category URLs.
func paginatePreview(tab *tabView) {
	if tab.Kind == "articles" {
		return
	}
	tab.Total = len(tab.Links)
	tab.Page = 1
	tab.PageCount = (tab.Total + 23) / 24
	tab.Links = firstLinks(tab.Links, 24)
	for n := 1; n <= tab.PageCount; n++ {
		url := "/c/" + tab.Slug
		if n > 1 {
			url += "?page=" + strconv.Itoa(n)
		}
		tab.Pages = append(tab.Pages, pageLink{Number: n, URL: url, Current: n == 1})
	}
	if tab.PageCount > 1 {
		tab.NextURL = "/c/" + tab.Slug + "?page=2"
	}
}

func firstLinks(links []linkView, limit int) []linkView {
	if len(links) > limit {
		return links[:limit]
	}
	return links
}

func firstModelLinks(links []models.Link, limit int) []models.Link {
	if len(links) > limit {
		return links[:limit]
	}
	return links
}

func articlePage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var item models.Article
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || db.Where("draft = ?", false).First(&item, id).Error != nil {
			sendSitePage(c, db, http.StatusNotFound, shell{
				Title: "没有这篇文章 - 开物录", Robots: "noindex", Canonical: "/a/" + c.Param("id"),
			}, "missing", map[string]string{"Heading": "没有这篇文章", "Text": "这篇说明不存在。"})
			return
		}
		db.Model(&item).UpdateColumn("views", gorm.Expr("views + ?", 1))
		item.Views++
		origin := publicOrigin(c)
		canonical := origin + "/a/" + itoa(item.ID)
		user, signedIn := currentUser(c, db)
		commentPage := min(positivePage(c.Query("comments")), 1000000)
		comments, moreComments := articleDiscussion(db, item.ID, user, commentPage)
		ld, _ := json.Marshal(map[string]any{
			"@context":         "https://schema.org",
			"@type":            "Article",
			"headline":         item.Title,
			"description":      item.Summary,
			"datePublished":    item.CreatedAt.Format(time.RFC3339),
			"mainEntityOfPage": canonical,
		})
		sendSitePage(c, db, http.StatusOK, shell{
			Title:       item.Title + " - 开物录",
			Description: clip(item.Summary, 120),
			Canonical:   "/a/" + itoa(item.ID),
			OGType:      "article",
			JSONLD:      ld,
		}, "article", map[string]any{
			"Title": item.Title, "Summary": item.Summary, "Body": item.Body, "Views": item.Views,
			"Date": item.CreatedAt.Format("2006-01-02"), "ISO": item.CreatedAt.Format(time.RFC3339),
			"Category": item.Category, "HTML": articleHTML(item), "Rich": item.Format == "html" || item.Format == "markdown",
			"Cover": publicMedia(item.CoverURL), "Video": publicMedia(item.VideoURL), "Embed": youtubeEmbed(item.VideoURL), "ID": item.ID,
			"Comments": comments, "SignedIn": signedIn, "UserID": user.ID, "CommentError": c.Query("commentError") == "1",
			"PreviousComments": commentPage - 1, "NextComments": commentPage + 1, "MoreComments": moreComments,
		})
	}
}

func articleCategoryTab(db *gorm.DB, category models.Category, page int, topic string) tabView {
	const perPage = 24
	var topics []string
	db.Model(&models.Article{}).Where("draft = ? AND category <> ''", false).Distinct("category").Order("category").Pluck("category", &topics)
	query := func() *gorm.DB {
		rows := db.Model(&models.Article{}).Where("draft = ?", false)
		if topic != "" {
			rows = rows.Where("category = ?", topic)
		}
		return rows
	}
	var total int64
	query().Count(&total)
	pageCount := max(1, int((total+perPage-1)/perPage))
	page = min(page, pageCount)
	var rows []models.Article
	query().Select("id", "title", "summary", "views", "created_at").
		Order("featured desc, sort asc, created_at desc, id desc").Offset((page - 1) * perPage).Limit(perPage).Find(&rows)
	tab := tabView{Slug: category.Slug, Name: category.Name, Kind: "articles", On: true, Total: int(total), Page: page, PageCount: pageCount, Topics: topics, Topic: topic}
	for _, row := range rows {
		tab.Posts = append(tab.Posts, postView{ID: row.ID, Title: row.Title, Summary: clip(row.Summary, 160), Views: row.Views,
			Date: row.CreatedAt.Format("2006-01-02"), ISO: row.CreatedAt.Format(time.RFC3339)})
	}
	pageURL := func(n int) string {
		values := url.Values{"page": {strconv.Itoa(n)}}
		if topic != "" {
			values.Set("topic", topic)
		}
		return "/c/" + category.Slug + "?" + values.Encode()
	}
	if page > 1 {
		tab.PrevURL = pageURL(page - 1)
	}
	if page < pageCount {
		tab.NextURL = pageURL(page + 1)
	}
	return tab
}

func sitePage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var link models.Link
		if err := db.Where("slug = ? AND status = ?", c.Param("slug"), "online").First(&link).Error; err != nil {
			sendSitePage(c, db, http.StatusNotFound, shell{
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
		bodyParagraphs := splitParagraphs(link.Body)
		bodyLead := ""
		var bodyTips []string
		if len(bodyParagraphs) > 0 {
			bodyLead = bodyParagraphs[0]
			bodyTips = bodyParagraphs[1:]
		}
		favoriteJSON := "{}"
		if views := toViews([]models.Link{link}, "row"); len(views) > 0 {
			favoriteJSON = views[0].JSON
		}
		sendSitePage(c, db, http.StatusOK, shell{
			Title:       title + " - 开物录",
			Description: clip(summary, 140),
			Keywords:    joinWords(append([]string{link.Name}, append(tags, cat.Name, parent.Name, "网址导航")...)...),
			Canonical:   "/site/" + link.Slug,
			OGType:      "article",
			JSONLD:      ld,
		}, "site", map[string]any{
			"Name": link.Name, "Desc": desc, "BodyLead": bodyLead, "BodyTips": bodyTips, "ID": link.ID,
			"ViewsText": formatCount(link.Views), "ClicksText": formatCount(link.Clicks),
			"External": external, "URL": link.URL, "Host": host, "IconHost": iconHost, "Icon": iconURL(iconHost), "Letter": letter,
			"FavoriteJSON": favoriteJSON,
			"Tags":         tags,
			"Alias":        link.Alias, "Lang": link.Lang, "Region": link.Region,
			"Spare": link.Spare, "ICP": link.ICP,
			"Date": link.CreatedAt.Format("2006-01-02"), "ISO": link.CreatedAt.Format(time.RFC3339),
			"ParentName": parent.Name, "ParentSlug": parent.Slug,
			"CatName": cat.Name, "CatSlug": cat.Slug,
			"Related": toViews(related, "row"),
		})
	}
}

func splitParagraphs(body string) []string {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	parts := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n\n")
	paragraphs := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			paragraphs = append(paragraphs, part)
		}
	}
	return paragraphs
}

func searchPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		title := "搜索 - 开物录"
		desc := "在开物录导航里搜索已收录的网站。"
		view := searchView{Query: q, Page: 1, PageCount: 1}
		if q != "" && utf8Count(q) <= 40 {
			safe := strings.NewReplacer("%", "", "_", "").Replace(q)
			like := "%" + safe + "%"
			query := db.Model(&models.Link{}).Where("status = ? AND (name LIKE ? OR `desc` LIKE ? OR tags LIKE ? OR alias LIKE ?)", "online", like, like, like, like)
			var total int64
			query.Count(&total)
			page := positivePage(c.Query("page"))
			const pageSize = 24
			pageCount := int((total + pageSize - 1) / pageSize)
			if pageCount < 1 {
				pageCount = 1
			}
			if page > pageCount {
				page = pageCount
			}
			var links []models.Link
			query.Order("clicks desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&links)
			view.Total, view.Page, view.PageCount = int(total), page, pageCount
			view.Links = toViews(links, "row")
			for n := 1; n <= pageCount; n++ {
				pageURL := "/search?q=" + url.QueryEscape(q)
				if n > 1 {
					pageURL += "&page=" + strconv.Itoa(n)
				}
				view.Pages = append(view.Pages, pageLink{Number: n, URL: pageURL, Current: n == page})
			}
			if page > 1 {
				view.PrevURL = "/search?q=" + url.QueryEscape(q)
				if page > 2 {
					view.PrevURL += "&page=" + strconv.Itoa(page-1)
				}
			}
			if page < pageCount {
				view.NextURL = "/search?q=" + url.QueryEscape(q) + "&page=" + strconv.Itoa(page+1)
			}
			title = q + " - 搜索 - 开物录"
			desc = "开物录导航里和「" + q + "」有关的网站。"
		}
		sendSitePage(c, db, http.StatusOK, shell{
			Title: title, Description: desc, Canonical: "/search", Robots: "noindex,follow", Query: q,
		}, "search", view)
	}
}

func submitPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c, db)
		if !ok {
			c.Redirect(http.StatusFound, "/login?next=/submit")
			return
		}
		showSubmit(c, db, http.StatusOK, submitView{OK: c.Query("ok") == "1", User: user.Username})
	}
}

type submitView struct {
	Name  string
	URL   string
	Desc  string
	Error string
	OK    bool
	User  string
}

func showSubmit(c *gin.Context, db *gorm.DB, status int, view submitView) {
	sendSitePage(c, db, status, shell{
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
		for _, tool := range migratedTools {
			if c.Param("slug") == tool.Slug {
				showLabTool(c, db, tool)
				return
			}
		}
		item, ok := pages[c.Param("slug")]
		if !ok {
			sendSitePage(c, db, http.StatusNotFound, shell{
				Title: "没有这个工具 - 开物录", Robots: "noindex", Canonical: "/tools/" + c.Param("slug"),
			}, "missing", map[string]string{"Heading": "没有这个工具", "Text": "这个工具不在列表里。"})
			return
		}
		sendSitePage(c, db, http.StatusOK, shell{
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
	if !on {
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
			Href: href, Name: link.Name, Desc: link.Desc, Letter: letter, Host: host, Icon: iconURL(host), JSON: string(raw), Layout: layout, Clicks: link.Clicks,
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
