package api

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const readerCookie = "ot_reader"

func readerSecret() string {
	return readerKey + ":reader"
}

func currentReader(c *gin.Context, db *gorm.DB) (models.User, bool) {
	token, _ := c.Cookie(readerCookie)
	if token == "" {
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		}
	}
	idText, ok := parseToken(token, readerSecret())
	if !ok {
		return models.User{}, false
	}
	id, err := strconv.ParseUint(idText, 10, 64)
	if err != nil || id == 0 {
		return models.User{}, false
	}
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		return models.User{}, false
	}
	return user, true
}

func setReaderCookie(c *gin.Context, user models.User) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(readerCookie, makeToken(strconv.FormatUint(uint64(user.ID), 10), readerSecret()), 7*24*3600, "/", "", false, true)
}

func clearReaderCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(readerCookie, "", -1, "/", "", false, true)
}

func safeNext(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || strings.Contains(raw, "://") {
		return "/"
	}
	return raw
}

func validUsername(name string) bool {
	n := utf8.RuneCountInString(name)
	if n < 2 || n > 20 {
		return false
	}
	for _, r := range name {
		if r <= ' ' || r == '/' || r == '\\' {
			return false
		}
	}
	return true
}

type authView struct {
	Title    string
	Action   string
	Switch   string
	SwitchTo string
	User     string
	Error    string
	Next     string
}

func loginPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		next := safeNext(c.Query("next"))
		if _, ok := currentReader(c, db); ok {
			c.Redirect(http.StatusFound, next)
			return
		}
		showAuth(c, db, http.StatusOK, authView{
			Title: "登录", Action: "/login", Switch: "没有账号？去注册", SwitchTo: "/register", Next: next,
		})
	}
}

func registerPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		next := safeNext(c.Query("next"))
		if _, ok := currentReader(c, db); ok {
			c.Redirect(http.StatusFound, next)
			return
		}
		showAuth(c, db, http.StatusOK, authView{
			Title: "注册", Action: "/register", Switch: "已有账号？去登录", SwitchTo: "/login", Next: next,
		})
	}
}

func authModal(c *gin.Context) bool {
	return c.GetHeader("X-Auth-Modal") == "1"
}

func showAuth(c *gin.Context, db *gorm.DB, status int, view authView) {
	if authModal(c) {
		c.JSON(status, gin.H{"error": view.Error})
		return
	}
	if view.Next != "" && view.Next != "/" {
		view.SwitchTo += "?next=" + url.QueryEscape(view.Next)
	}
	c.Status(status)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := publicPages.ExecuteTemplate(c.Writer, "authpage", view); err != nil {
		log.Printf("auth: %v", err)
	}
}

func loginPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		next := safeNext(c.PostForm("next"))
		name := strings.TrimSpace(c.PostForm("username"))
		password := c.PostForm("password")
		fail := func(msg string) {
			showAuth(c, db, http.StatusUnauthorized, authView{
				Title: "登录", Action: "/login", Switch: "没有账号？去注册", SwitchTo: "/register",
				User: name, Error: msg, Next: next,
			})
		}
		var user models.User
		if err := db.Where("username = ?", name).First(&user).Error; err != nil ||
			bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
			fail("账号或密码不对")
			return
		}
		setReaderCookie(c, user)
		if authModal(c) {
			c.JSON(http.StatusOK, gin.H{"ok": true, "next": next})
			return
		}
		c.Redirect(http.StatusSeeOther, next)
	}
}

func registerPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		next := safeNext(c.PostForm("next"))
		name := strings.TrimSpace(c.PostForm("username"))
		password := c.PostForm("password")
		fail := func(status int, msg string) {
			showAuth(c, db, status, authView{
				Title: "注册", Action: "/register", Switch: "已有账号？去登录", SwitchTo: "/login",
				User: name, Error: msg, Next: next,
			})
		}
		if !validUsername(name) {
			fail(http.StatusBadRequest, "用户名用 2 到 20 个字，不要空格")
			return
		}
		if len(password) < 8 || len(password) > 72 {
			fail(http.StatusBadRequest, "密码至少 8 位")
			return
		}
		var n int64
		db.Model(&models.User{}).Where("username = ?", name).Count(&n)
		if n > 0 {
			fail(http.StatusConflict, "这个用户名已经有人用了")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fail(http.StatusInternalServerError, "没有注册上")
			return
		}
		user := models.User{Username: name, PasswordHash: string(hash)}
		if err := db.Create(&user).Error; err != nil {
			fail(http.StatusInternalServerError, "没有注册上")
			return
		}
		setReaderCookie(c, user)
		if authModal(c) {
			c.JSON(http.StatusOK, gin.H{"ok": true, "next": next})
			return
		}
		c.Redirect(http.StatusSeeOther, next)
	}
}

func logout(c *gin.Context) {
	clearReaderCookie(c)
	c.Redirect(http.StatusSeeOther, "/")
}

type mineItem struct {
	Name   string
	URL    string
	Status string
	Date   string
}

func requireReader(c *gin.Context, db *gorm.DB) (models.User, bool) {
	user, ok := currentReader(c, db)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return models.User{}, false
	}
	return user, true
}

func desk(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := requireReader(c, db)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"stars":  userMarks(db, user.ID, "star", 16),
			"recent": userMarks(db, user.ID, "recent", 12),
		})
	}
}

func toggleStar(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := requireReader(c, db)
		if !ok {
			return
		}
		link, ok := markLink(c, db)
		if !ok {
			return
		}
		var row models.UserLink
		err := db.Where("user_id = ? AND link_id = ? AND kind = ?", user.ID, link.ID, "star").First(&row).Error
		if err == nil {
			db.Delete(&row)
		} else {
			db.Create(&models.UserLink{UserID: user.ID, LinkID: link.ID, Kind: "star", UpdatedAt: time.Now()})
			trimMarks(db, user.ID, "star", 16)
		}
		c.JSON(http.StatusOK, gin.H{"stars": userMarks(db, user.ID, "star", 16)})
	}
}

func touchRecent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := requireReader(c, db)
		if !ok {
			return
		}
		link, ok := markLink(c, db)
		if !ok {
			return
		}
		var row models.UserLink
		err := db.Where("user_id = ? AND link_id = ? AND kind = ?", user.ID, link.ID, "recent").First(&row).Error
		now := time.Now()
		if err == nil {
			db.Model(&row).Update("updated_at", now)
		} else {
			db.Create(&models.UserLink{UserID: user.ID, LinkID: link.ID, Kind: "recent", UpdatedAt: now})
		}
		trimMarks(db, user.ID, "recent", 12)
		c.JSON(http.StatusOK, gin.H{"recent": userMarks(db, user.ID, "recent", 12)})
	}
}

func markLink(c *gin.Context, db *gorm.DB) (models.Link, bool) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有这个网址"})
		return models.Link{}, false
	}
	var link models.Link
	if err := db.First(&link, req.ID).Error; err != nil || link.Status != "online" {
		c.JSON(http.StatusNotFound, gin.H{"error": "没有这个网址"})
		return models.Link{}, false
	}
	return link, true
}

func markedLinks(db *gorm.DB, userID uint, kind string, limit int) []models.Link {
	var rows []models.UserLink
	q := db.Where("user_id = ? AND kind = ?", userID, kind)
	if kind == "recent" {
		q = q.Order("updated_at desc")
	} else {
		q = q.Order("updated_at asc")
	}
	q.Limit(limit).Find(&rows)
	if len(rows) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.LinkID)
	}
	var links []models.Link
	db.Where("id IN ? AND status = ?", ids, "online").Find(&links)
	byID := map[uint]models.Link{}
	for _, link := range links {
		byID[link.ID] = link
	}
	out := make([]models.Link, 0, len(rows))
	for _, row := range rows {
		link, ok := byID[row.LinkID]
		if ok {
			out = append(out, link)
		}
	}
	return out
}

func userMarks(db *gorm.DB, userID uint, kind string, limit int) []map[string]any {
	links := markedLinks(db, userID, kind, limit)
	out := make([]map[string]any, 0, len(links))
	for _, link := range links {
		out = append(out, map[string]any{
			"id": link.ID, "name": link.Name, "url": link.URL, "desc": link.Desc,
			"slug": link.Slug, "categorySlug": link.CategorySlug, "clicks": link.Clicks,
		})
	}
	return out
}

func trimMarks(db *gorm.DB, userID uint, kind string, keep int) {
	var rows []models.UserLink
	db.Where("user_id = ? AND kind = ?", userID, kind).Order("updated_at desc").Find(&rows)
	if len(rows) <= keep {
		return
	}
	drop := make([]uint, 0, len(rows)-keep)
	for _, row := range rows[keep:] {
		drop = append(drop, row.ID)
	}
	db.Delete(&models.UserLink{}, drop)
}

func minePage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentReader(c, db)
		if !ok {
			c.Redirect(http.StatusFound, "/login?next=/mine")
			return
		}
		var rows []models.Link
		db.Where("user_id = ?", user.ID).Order("id desc").Limit(50).Find(&rows)
		items := make([]mineItem, 0, len(rows))
		for _, row := range rows {
			label := "未通过"
			if row.Status == "pending" {
				label = "待审"
			} else if row.Status == "online" {
				label = "已收录"
			}
			items = append(items, mineItem{
				Name: row.Name, URL: row.URL, Status: label, Date: row.CreatedAt.Format("2006-01-02"),
			})
		}
		stars := toViews(markedLinks(db, user.ID, "star", 16), "tile")
		for i := range stars {
			stars[i].Star = " on"
		}
		render(c, db, http.StatusOK, shell{
			Title: user.Username + " - 开物录", Description: "这个账号在开物录的收藏、最近使用和提交。",
			Canonical: "/mine", Robots: "noindex,follow", Account: true,
		}, "mine", map[string]any{
			"User": user.Username, "Date": user.CreatedAt.Format("2006-01-02"),
			"Stars": stars, "Recent": toViews(markedLinks(db, user.ID, "recent", 12), "tile"), "Items": items,
		})
	}
}
