package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"opentools/config"
	"opentools/models"
	"opentools/seed"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var slugRe = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

func MountAdmin(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	r.POST("/api/admin/login", adminLogin(db, cfg))
	g := r.Group("/api/admin", requireAdmin(cfg.AdminSecret))
	g.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"username": c.GetString("adminUser")})
	})
	g.GET("/stats", adminStats(db))
	g.GET("/links", adminLinks(db))
	g.POST("/links", adminSaveLink(db, false))
	g.PUT("/links/:id", adminSaveLink(db, true))
	g.DELETE("/links/:id", adminDelete(db, func() any { return &models.Link{} }))
	g.GET("/categories", adminCategories(db))
	g.POST("/categories", adminSaveCategory(db, false))
	g.PUT("/categories/:id", adminSaveCategory(db, true))
	g.DELETE("/categories/:id", adminDeleteCategory(db))
	g.GET("/articles", adminArticles(db))
	g.POST("/articles", adminSaveArticle(db, false))
	g.PUT("/articles/:id", adminSaveArticle(db, true))
	g.DELETE("/articles/:id", adminDelete(db, func() any { return &models.Article{} }))
	g.GET("/tags", adminTags(db))
	g.POST("/tags", adminSaveTag(db, false))
	g.PUT("/tags/:id", adminSaveTag(db, true))
	g.DELETE("/tags/:id", adminDelete(db, func() any { return &models.Tag{} }))
}

func adminLogin(db *gorm.DB, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提交格式不对"})
			return
		}
		var admin models.Admin
		if err := db.Where("username = ?", strings.TrimSpace(req.Username)).First(&admin).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码不对"})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码不对"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"username": admin.Username,
			"token":    makeToken(admin.Username, cfg.AdminSecret),
		})
	}
}

func requireAdmin(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := ""
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			raw = strings.TrimPrefix(h, "Bearer ")
		}
		user, ok := parseToken(raw, secret)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		c.Set("adminUser", user)
		c.Next()
	}
}

func makeToken(user, secret string) string {
	exp := time.Now().Add(7 * 24 * time.Hour).Unix()
	payload := user + "|" + strconv.FormatInt(exp, 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig
}

func parseToken(token, secret string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	if !hmac.Equal([]byte(parts[1]), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		return "", false
	}
	bits := strings.Split(string(raw), "|")
	if len(bits) != 2 || bits[0] == "" {
		return "", false
	}
	exp, err := strconv.ParseInt(bits[1], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return "", false
	}
	return bits[0], true
}

func adminStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		count := func(q *gorm.DB) int64 {
			var n int64
			q.Count(&n)
			return n
		}
		c.JSON(http.StatusOK, gin.H{
			"online":   count(db.Model(&models.Link{}).Where("status = ?", "online")),
			"pending":  count(db.Model(&models.Link{}).Where("status = ?", "pending")),
			"off":      count(db.Model(&models.Link{}).Where("status = ?", "off")),
			"articles": count(db.Model(&models.Article{})),
			"tags":     count(db.Model(&models.Tag{})),
		})
	}
}

func adminLinks(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := db.Model(&models.Link{}).Order("id desc")
		if status := c.Query("status"); status != "" {
			q = q.Where("status = ?", status)
		}
		if text := strings.TrimSpace(c.Query("q")); text != "" {
			like := "%" + strings.NewReplacer("%", "", "_", "").Replace(text) + "%"
			q = q.Where("name LIKE ? OR url LIKE ?", like, like)
		}
		var rows []models.Link
		q.Limit(200).Find(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

func adminSaveLink(db *gorm.DB, update bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CategorySlug string `json:"categorySlug"`
			Name         string `json:"name"`
			URL          string `json:"url"`
			Desc         string `json:"desc"`
			Tags         string `json:"tags"`
			Alias        string `json:"alias"`
			Lang         string `json:"lang"`
			Region       string `json:"region"`
			Spare        string `json:"spare"`
			ICP          string `json:"icp"`
			Body         string `json:"body"`
			Status       string `json:"status"`
			Sort         int    `json:"sort"`
			PinSort      int    `json:"pinSort"`
			Pinned       bool   `json:"pinned"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提交格式不对"})
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.URL = strings.TrimSpace(req.URL)
		req.Desc = strings.TrimSpace(req.Desc)
		req.Tags = strings.TrimSpace(req.Tags)
		req.Alias = strings.TrimSpace(req.Alias)
		req.Lang = strings.TrimSpace(req.Lang)
		req.Region = strings.TrimSpace(req.Region)
		req.Spare = strings.TrimSpace(req.Spare)
		req.ICP = strings.TrimSpace(req.ICP)
		req.Body = strings.TrimSpace(req.Body)
		req.CategorySlug = strings.TrimSpace(req.CategorySlug)
		if req.Name == "" || utf8.RuneCountInString(req.Name) > 40 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "名字要在 40 字以内"})
			return
		}
		if utf8.RuneCountInString(req.Desc) > 80 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "简介留在 80 字以内，长文写在网站介绍"})
			return
		}
		if utf8.RuneCountInString(req.Tags) > 40 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "标签留在 40 字以内"})
			return
		}
		if utf8.RuneCountInString(req.Alias) > 20 || utf8.RuneCountInString(req.Lang) > 20 || utf8.RuneCountInString(req.Region) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "别名、语言、地区各留在 20 字以内"})
			return
		}
		if utf8.RuneCountInString(req.ICP) > 40 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "备案号太长"})
			return
		}
		if req.Spare != "" && !strings.HasPrefix(req.Spare, "http://") && !strings.HasPrefix(req.Spare, "https://") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "备用地址要以 http 或 https 开头"})
			return
		}
		if utf8.RuneCountInString(req.Body) > 4000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "网站介绍太长"})
			return
		}
		if req.Status != "online" && req.Status != "pending" && req.Status != "off" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "状态只能是上架、待审或下架"})
			return
		}
		if !strings.HasPrefix(req.URL, "/") && !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "网址要以 http、https 或 / 开头"})
			return
		}
		var cat models.Category
		if err := db.Where("slug = ? AND kind = ?", req.CategorySlug, "tab").First(&cat).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选一个子分类"})
			return
		}
		row := models.Link{}
		if update {
			if err := db.First(&row, c.Param("id")).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "没有这条网址"})
				return
			}
		} else {
			row.CreatedAt = time.Now()
		}
		row.CategorySlug = req.CategorySlug
		row.Name = req.Name
		row.URL = req.URL
		row.Desc = req.Desc
		row.Tags = req.Tags
		row.Alias = req.Alias
		row.Lang = req.Lang
		row.Region = req.Region
		row.Spare = req.Spare
		row.ICP = req.ICP
		row.Body = req.Body
		row.Status = req.Status
		row.Sort = req.Sort
		row.Pinned = req.Pinned
		row.PinSort = req.PinSort
		seed.FillSlug(db, &row)
		var err error
		if update {
			err = db.Save(&row).Error
		} else {
			err = db.Create(&row).Error
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "没有存上"})
			return
		}
		c.JSON(http.StatusOK, row)
	}
}

func adminCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows []models.Category
		db.Order("parent_slug, sort, id").Find(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

func adminSaveCategory(db *gorm.DB, update bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Slug        string `json:"slug"`
			Name        string `json:"name"`
			ParentSlug  string `json:"parentSlug"`
			Kind        string `json:"kind"`
			ContentKind string `json:"contentKind"`
			Sort        int    `json:"sort"`
			ShowOnHome  bool   `json:"showOnHome"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提交格式不对"})
			return
		}
		req.Slug = strings.TrimSpace(req.Slug)
		req.Name = strings.TrimSpace(req.Name)
		req.ParentSlug = strings.TrimSpace(req.ParentSlug)
		if !slugRe.MatchString(req.Slug) || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "标识用小写字母、数字和横线，名字不能空"})
			return
		}
		if req.Kind != "section" && req.Kind != "tab" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "类型只能是分区或子类"})
			return
		}
		if req.ContentKind != "links" && req.ContentKind != "articles" && req.ContentKind != "tags" && req.ContentKind != "mixed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "内容类型不对"})
			return
		}
		row := models.Category{}
		if update {
			if err := db.First(&row, c.Param("id")).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "没有这个分类"})
				return
			}
			if row.Slug != req.Slug {
				c.JSON(http.StatusBadRequest, gin.H{"error": "标识建好后不能改"})
				return
			}
		}
		row.Slug = req.Slug
		row.Name = req.Name
		row.ParentSlug = req.ParentSlug
		row.Kind = req.Kind
		row.ContentKind = req.ContentKind
		row.Sort = req.Sort
		row.ShowOnHome = req.ShowOnHome
		var err error
		if update {
			err = db.Save(&row).Error
		} else {
			err = db.Create(&row).Error
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "没有存上，标识可能重复了"})
			return
		}
		c.JSON(http.StatusOK, row)
	}
}

func adminDeleteCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var row models.Category
		if err := db.First(&row, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "没有这个分类"})
			return
		}
		var n int64
		db.Model(&models.Category{}).Where("parent_slug = ?", row.Slug).Count(&n)
		if n > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "下面还有子类，先删子类"})
			return
		}
		db.Model(&models.Link{}).Where("category_slug = ?", row.Slug).Count(&n)
		if n > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "下面还有网址，先移走或删掉"})
			return
		}
		db.Delete(&row)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func adminArticles(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows []models.Article
		db.Order("id desc").Find(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

func adminSaveArticle(db *gorm.DB, update bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title   string `json:"title"`
			Summary string `json:"summary"`
			Body    string `json:"body"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提交格式不对"})
			return
		}
		req.Title = strings.TrimSpace(req.Title)
		req.Summary = strings.TrimSpace(req.Summary)
		req.Body = strings.TrimSpace(req.Body)
		if req.Title == "" || req.Body == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "标题和正文不能空"})
			return
		}
		row := models.Article{}
		if update {
			if err := db.First(&row, c.Param("id")).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "没有这篇文章"})
				return
			}
		} else {
			row.CreatedAt = time.Now()
		}
		row.Title = req.Title
		row.Summary = req.Summary
		row.Body = req.Body
		var err error
		if update {
			err = db.Save(&row).Error
		} else {
			err = db.Create(&row).Error
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "没有存上"})
			return
		}
		c.JSON(http.StatusOK, row)
	}
}

func adminTags(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows []models.Tag
		db.Order("sort, id").Find(&rows)
		c.JSON(http.StatusOK, rows)
	}
}

func adminSaveTag(db *gorm.DB, update bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Group string `json:"group"`
			Name  string `json:"name"`
			Sort  int    `json:"sort"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提交格式不对"})
			return
		}
		req.Group = strings.TrimSpace(req.Group)
		req.Name = strings.TrimSpace(req.Name)
		if req.Group == "" || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "分组和名字不能空"})
			return
		}
		row := models.Tag{}
		if update {
			if err := db.First(&row, c.Param("id")).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "没有这个标签"})
				return
			}
		}
		row.Group = req.Group
		row.Name = req.Name
		row.Sort = req.Sort
		var err error
		if update {
			err = db.Save(&row).Error
		} else {
			err = db.Create(&row).Error
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "没有存上"})
			return
		}
		c.JSON(http.StatusOK, row)
	}
}

func adminDelete(db *gorm.DB, model func() any) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Delete(model(), c.Param("id")).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "没有删掉"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
