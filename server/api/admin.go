package api

import (
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"opentools/models"
	"opentools/seed"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var slugRe = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

func MountAdmin(r *gin.Engine, db *gorm.DB) {
	g := r.Group("/api/admin", requireAdmin(db))
	g.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"username": c.GetString("adminUser")})
	})
	g.GET("/stats", adminStats(db))
	g.GET("/submissions", adminSubmissions(db))
	g.GET("/submission-reviews", adminSubmissionReviews(db))
	g.PATCH("/submissions/:id", reviewSubmission(db))
	mountAdminInternetRank(g, db)
	g.GET("/users", adminUsers(db))
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
	g.POST("/articles/images", uploadArticleImage)
	g.PUT("/articles/:id", adminSaveArticle(db, true))
	g.DELETE("/articles/:id", adminDelete(db, func() any { return &models.Article{} }))
	g.GET("/feedback", func(c *gin.Context) {
		rows := make([]models.Feedback, 0)
		if err := db.Order("id desc").Find(&rows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "反馈读取失败"})
			return
		}
		c.JSON(http.StatusOK, rows)
	})
	g.PUT("/feedback/:id", updateFeedback(db))
	g.GET("/tags", adminTags(db))
	g.POST("/tags", adminSaveTag(db, false))
	g.PUT("/tags/:id", adminSaveTag(db, true))
	g.DELETE("/tags/:id", adminDelete(db, func() any { return &models.Tag{} }))
}

func requireAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		user, ok := currentUser(c, db)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		if !user.IsAdmin || (c.Request.Method != http.MethodGet && !sameOrigin(c)) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "没有后台操作权限"})
			return
		}
		c.Set("adminUser", user.Username)
		c.Next()
	}
}

func sameOrigin(c *gin.Context) bool {
	if site := c.GetHeader("Sec-Fetch-Site"); site != "" {
		return site == "same-origin"
	}
	origin := c.GetHeader("Origin")
	return origin == "" || strings.EqualFold(origin, publicOrigin(c))
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
	escapeLike := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
	return func(c *gin.Context) {
		q := db.WithContext(c.Request.Context()).Model(&models.Link{})
		if status := c.Query("status"); status != "" {
			q = q.Where("status = ?", status)
		}
		if category := c.Query("category"); category != "" {
			q = q.Where("category_slug = ?", category)
		}
		if text := strings.TrimSpace(c.Query("q")); text != "" {
			like := "%" + escapeLike.Replace(clip(text, 254)) + "%"
			q = q.Where("name LIKE ? ESCAPE '!' OR url LIKE ? ESCAPE '!'", like, like)
		}
		rows := make([]models.Link, 0)
		if _, paginated := c.GetQuery("page"); !paginated {
			if err := q.Order("id desc").Limit(200).Find(&rows).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "读取网址列表失败"})
				return
			}
			c.JSON(http.StatusOK, rows)
			return
		}
		loadAdminPage(c, q.Order("id desc"), &rows, "读取网址列表失败")
	}
}

func loadAdminPage(c *gin.Context, query *gorm.DB, rows any, errorMessage string) {
	const pageSize = 25
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMessage})
		return
	}
	page := min(positivePage(c.Query("page")), max(1, int((total+pageSize-1)/pageSize)))
	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Find(rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMessage})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows, "page": page, "pageSize": pageSize, "total": total})
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
		oldURL := row.URL
		oldStatus := row.Status
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
		if update && (oldURL != req.URL || (req.Status == "online" && oldStatus != "online")) {
			row.CheckFails = 0
			row.CheckedAt = nil
		}
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
			Title    string  `json:"title"`
			Summary  string  `json:"summary"`
			Body     string  `json:"body"`
			Format   *string `json:"format"`
			CoverURL *string `json:"coverUrl"`
			VideoURL *string `json:"videoUrl"`
			Category *string `json:"category"`
			Featured *bool   `json:"featured"`
			Sort     *int    `json:"sort"`
			Draft    *bool   `json:"draft"`
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
		if req.Format != nil {
			if *req.Format != "" && *req.Format != "html" && *req.Format != "markdown" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "正文格式不正确"})
				return
			}
			row.Format = *req.Format
		}
		if req.CoverURL != nil {
			row.CoverURL = publicMedia(strings.TrimSpace(*req.CoverURL))
		}
		if req.VideoURL != nil {
			row.VideoURL = publicMedia(strings.TrimSpace(*req.VideoURL))
		}
		if req.Category != nil {
			row.Category = clip(strings.TrimSpace(*req.Category), 64)
		}
		if req.Featured != nil {
			row.Featured = *req.Featured
		}
		if req.Sort != nil {
			row.Sort = *req.Sort
		}
		if req.Draft != nil {
			row.Draft = *req.Draft
		}
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
