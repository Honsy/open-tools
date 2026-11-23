package api

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"opentools/models"
	"opentools/storage"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"gorm.io/gorm"
)

func mountMovedContent(router *gin.RouterGroup, db *gorm.DB) {
	for from, to := range map[string]string{
		"/ai": "/c/ai", "/ai/tools": "/c/ai", "/ai/sites": "/c/ai",
		"/ai/articles": "/c/daily", "/ranking/ai": "/rank?category=ai", "/ai/lab": "/tools/",
	} {
		destination := to
		router.GET(from, func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, destination) })
	}
	router.GET("/ai/articles/:legacy", func(c *gin.Context) {
		var row models.Article
		value := c.Param("legacy")
		err := db.Select("id").Where("draft = ? AND (legacy_id = ? OR legacy_slug = ?)", false, value, value).First(&row).Error
		if err != nil {
			sendSitePage(c, db, http.StatusNotFound, shell{Title: "没有这篇文章 - 开物录", Robots: "noindex"}, "missing",
				map[string]string{"Heading": "没有这篇文章", "Text": "文章不存在，或仍为草稿。"})
			return
		}
		c.Redirect(http.StatusMovedPermanently, "/a/"+itoa(row.ID))
	})
	router.GET("/ai/lab/:slug", func(c *gin.Context) {
		for _, tool := range migratedTools {
			if c.Param("slug") == tool.Slug {
				c.Redirect(http.StatusMovedPermanently, "/tools/"+tool.Slug)
				return
			}
		}
		c.Status(http.StatusNotFound)
	})
	router.POST("/a/:id/comments", postArticleComment(db))
	router.GET("/feedback", feedbackPage(db))
	router.POST("/feedback", postFeedback(db))
}

var articleMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithRendererOptions(html.WithUnsafe()))
var articlePolicy = bluemonday.UGCPolicy()

func articleHTML(item models.Article) string {
	switch item.Format {
	case "html":
		return articlePolicy.Sanitize(item.Body)
	case "markdown":
		return markdownHTML(item.Body)
	}
	return ""
}

func markdownHTML(body string) string {
	var out bytes.Buffer
	if err := articleMarkdown.Convert([]byte(body), &out); err == nil {
		return string(articlePolicy.SanitizeBytes(out.Bytes()))
	}
	return ""
}

func publicMedia(raw string) string {
	n, err := url.Parse(raw)
	if err != nil || n.Hostname() == "" || n.User != nil || (n.Scheme != "http" && n.Scheme != "https") {
		return ""
	}
	return n.String()
}

var youtubeID = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

func youtubeEmbed(raw string) string {
	video, err := url.Parse(raw)
	if err != nil || (video.Scheme != "https" && video.Scheme != "http") {
		return ""
	}
	var id string
	switch strings.ToLower(video.Hostname()) {
	case "youtu.be":
		id = strings.TrimPrefix(video.Path, "/")
	case "youtube.com", "www.youtube.com", "m.youtube.com", "www.youtube-nocookie.com":
		switch {
		case video.Path == "/watch":
			id = video.Query().Get("v")
		case strings.HasPrefix(video.Path, "/embed/"):
			id = strings.TrimPrefix(video.Path, "/embed/")
		case strings.HasPrefix(video.Path, "/shorts/"):
			id = strings.TrimPrefix(video.Path, "/shorts/")
		}
	}
	if !youtubeID.MatchString(id) {
		return ""
	}
	return "https://www.youtube-nocookie.com/embed/" + id
}

type articleCommentView struct {
	models.ArticleComment
	CanEdit bool
	Depth   int
}

func articleDiscussion(db *gorm.DB, articleID uint, user models.User, page int) ([]articleCommentView, bool) {
	var roots []models.ArticleComment
	query := db.Where("article_id = ? AND hidden = ?", articleID, false).
		Where("parent_id IS NULL OR parent_id NOT IN (?)", db.Model(&models.ArticleComment{}).
			Select("id").Where("article_id = ? AND hidden = ?", articleID, false))
	const perPage = 30
	if err := query.Order("id desc").Offset((page - 1) * perPage).Limit(perPage + 1).Find(&roots).Error; err != nil {
		return nil, false
	}
	more := len(roots) > perPage
	if more {
		roots = roots[:perPage]
	}
	if len(roots) == 0 {
		return nil, false
	}
	var replies []models.ArticleComment
	db.Where("article_id = ? AND hidden = ? AND parent_id IS NOT NULL", articleID, false).Order("id").Find(&replies)
	children := make(map[uint][]models.ArticleComment)
	for _, row := range replies {
		children[*row.ParentID] = append(children[*row.ParentID], row)
	}
	visited := make(map[uint]bool)
	var out []articleCommentView
	var appendComment func(models.ArticleComment, int)
	appendComment = func(row models.ArticleComment, depth int) {
		if visited[row.ID] {
			return
		}
		visited[row.ID] = true
		out = append(out, articleCommentView{row, user.IsAdmin || user.ID != 0 && user.ID == row.UserID, depth})
		for _, child := range children[row.ID] {
			appendComment(child, depth+1)
		}
	}
	for _, root := range roots {
		appendComment(root, 0)
	}
	return out, more
}

func postArticleComment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := "/a/" + c.Param("id")
		user, ok := currentUser(c, db)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login?next="+url.QueryEscape(path))
			return
		}
		if !sameOrigin(c) {
			c.Status(http.StatusForbidden)
			return
		}
		articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		var article models.Article
		if err != nil || db.Select("id").Where("draft = ?", false).First(&article, articleID).Error != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 12<<10)
		action := c.PostForm("action")
		body := strings.TrimSpace(c.PostForm("body"))
		if action != "delete" && (body == "" || utf8.RuneCountInString(body) > 1000) {
			c.Redirect(http.StatusSeeOther, path+"?commentError=1#comments")
			return
		}
		if action == "edit" || action == "delete" {
			id, _ := strconv.ParseUint(c.PostForm("comment"), 10, 64)
			query := db.Model(&models.ArticleComment{}).Where("id = ? AND article_id = ? AND hidden = ?", id, article.ID, false)
			if !user.IsAdmin || action == "edit" {
				query = query.Where("user_id = ?", user.ID)
			}
			var result *gorm.DB
			if action == "delete" {
				result = query.Update("hidden", true)
			} else {
				result = query.Update("body", body)
			}
			if result.Error != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
		} else {
			row := models.ArticleComment{ArticleID: article.ID, UserID: user.ID, Author: user.Username, Body: body}
			if parent := c.PostForm("parent"); parent != "" {
				id, err := strconv.ParseUint(parent, 10, 64)
				var reply models.ArticleComment
				if err != nil || db.Select("id").Where("article_id = ? AND hidden = ?", article.ID, false).First(&reply, id).Error != nil {
					c.Redirect(http.StatusSeeOther, path+"?commentError=1#comments")
					return
				}
				row.ParentID = &reply.ID
			}
			if err := db.Create(&row).Error; err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
		}
		c.Redirect(http.StatusSeeOther, path+"#comments")
	}
}

type feedbackView struct {
	SignedIn bool
	Items    []models.Feedback
	Error    string
	Sent     bool
}

func feedbackPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) { showFeedback(c, db, "", c.Query("sent") == "1") }
}

func showFeedback(c *gin.Context, db *gorm.DB, warning string, sent bool) {
	user, ok := currentUser(c, db)
	view := feedbackView{SignedIn: ok, Error: warning, Sent: sent}
	if ok {
		db.Where("user_id = ?", user.ID).Order("id desc").Limit(100).Find(&view.Items)
	}
	sendSitePage(c, db, http.StatusOK, shell{Title: "站点反馈 - 开物录", Canonical: "/feedback", Robots: "noindex,follow"}, "feedback", view)
}

func postFeedback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c, db)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login?next=/feedback")
			return
		}
		if !sameOrigin(c) {
			c.Status(http.StatusForbidden)
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 12<<10)
		body := strings.TrimSpace(c.PostForm("body"))
		if length := utf8.RuneCountInString(body); length < 1 || length > 1000 {
			showFeedback(c, db, "请填写 1 到 1000 个字", false)
			return
		}
		if err := db.Create(&models.Feedback{UserID: user.ID, Kind: "message", Body: body, Status: "pending"}).Error; err != nil {
			showFeedback(c, db, "保存失败，请稍后试试", false)
			return
		}
		c.Redirect(http.StatusSeeOther, "/feedback?sent=1")
	}
}

func updateFeedback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct{ Status string }
		if err := c.ShouldBindJSON(&data); err != nil || (data.Status != "pending" && data.Status != "done" && data.Status != "rejected") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "状态不正确"})
			return
		}
		var row models.Feedback
		if err := db.First(&row, c.Param("id")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "没有这条反馈"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取失败"})
			return
		}
		if err := db.Model(&row).Update("status", data.Status).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
		c.JSON(http.StatusOK, row)
	}
}

func uploadArticleImage(c *gin.Context) {
	if r2Client == nil || r2Public == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "图片存储未配置"})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, storage.MaxArticleImage+(1<<20))
	if err := c.Request.ParseMultipartForm(storage.MaxArticleImage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": storage.ErrImage.Error()})
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择图片"})
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, storage.MaxArticleImage+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片读取失败"})
		return
	}
	address, err := storage.PutArticleImage(c.Request.Context(), r2Client, r2Bucket, r2Public, data)
	if errors.Is(err, storage.ErrImage) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "图片上传失败"})
	} else {
		c.JSON(http.StatusOK, gin.H{"url": address})
	}
}
