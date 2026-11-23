package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const resetTTL = 30 * time.Minute

func validEmail(raw string) bool {
	email := strings.TrimSpace(raw)
	if len(email) < 6 || len(email) > 254 || strings.ContainsAny(email, " \t\r\n") {
		return false
	}
	at := strings.LastIndex(email, "@")
	if at < 1 || at >= len(email)-3 {
		return false
	}
	local, domain := email[:at], email[at+1:]
	if strings.Contains(local, "@") || !strings.Contains(domain, ".") {
		return false
	}
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	return utf8.ValidString(email)
}

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newResetToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func forgotPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		showAuth(c, db, http.StatusOK, authView{
			Kind: "forgot", Title: "重置密码", Action: "/forgot",
		})
	}
}

func forgotPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := normalizeEmail(c.PostForm("email"))
		view := authView{Kind: "forgot", Title: "重置密码", Action: "/forgot", Email: email}
		if !validEmail(email) {
			view.Error = "邮箱格式不对"
			showAuth(c, db, http.StatusBadRequest, view)
			return
		}
		if !mailConfigured() {
			view.Error = "邮件暂时发不出去，稍后再试"
			showAuth(c, db, http.StatusServiceUnavailable, view)
			return
		}
		sent := "如果这个邮箱注册过，重置邮件已经发出。去收件箱看，也包括垃圾箱。链接 30 分钟内有效。"
		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			view.Notice = sent
			showAuth(c, db, http.StatusOK, view)
			return
		}
		var latest models.PasswordReset
		err := db.Where("user_id = ?", user.ID).Order("id desc").First(&latest).Error
		if err == nil && time.Since(latest.CreatedAt) < time.Minute && latest.ExpiresAt.After(time.Now()) {
			view.Notice = sent
			showAuth(c, db, http.StatusOK, view)
			return
		}
		token, err := newResetToken()
		if err != nil {
			view.Error = "邮件没有发出去，稍后再试"
			showAuth(c, db, http.StatusInternalServerError, view)
			return
		}
		db.Where("user_id = ? OR expires_at < ?", user.ID, time.Now()).Delete(&models.PasswordReset{})
		row := models.PasswordReset{
			UserID: user.ID, TokenHash: hashResetToken(token),
			ExpiresAt: time.Now().Add(resetTTL), CreatedAt: time.Now(),
		}
		if err := db.Create(&row).Error; err != nil {
			view.Error = "邮件没有发出去，稍后再试"
			showAuth(c, db, http.StatusInternalServerError, view)
			return
		}
		link := strings.TrimRight(publicOrigin(c), "/") + "/reset?token=" + token
		if err := sendMail(email, "重置开物录密码", resetMailBody(link)); err != nil {
			db.Delete(&row)
			log.Printf("reset mail: %v", err)
			view.Error = "邮件没有发出去，稍后再试"
			showAuth(c, db, http.StatusBadGateway, view)
			return
		}
		view.Notice = sent
		showAuth(c, db, http.StatusOK, view)
	}
}

func resetPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Query("token"))
		if _, ok := resetUser(db, token); !ok {
			showAuth(c, db, http.StatusBadRequest, authView{
				Kind: "forgot", Title: "重置密码", Action: "/forgot",
				Error: "这个链接用不了，或已经过期。再填一次邮箱。",
			})
			return
		}
		showAuth(c, db, http.StatusOK, authView{
			Kind: "reset", Title: "设置新密码", Action: "/reset", Token: token,
		})
	}
}

func resetPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.PostForm("token"))
		password := c.PostForm("password")
		again := c.PostForm("password2")
		view := authView{Kind: "reset", Title: "设置新密码", Action: "/reset", Token: token}
		user, ok := resetUser(db, token)
		if !ok {
			showAuth(c, db, http.StatusBadRequest, authView{
				Kind: "forgot", Title: "重置密码", Action: "/forgot",
				Error: "这个链接用不了，或已经过期。再填一次邮箱。",
			})
			return
		}
		if len(password) < 8 || len(password) > 72 {
			view.Error = "密码至少 8 位"
			showAuth(c, db, http.StatusBadRequest, view)
			return
		}
		if password != again {
			view.Error = "两次密码不一致"
			showAuth(c, db, http.StatusBadRequest, view)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			view.Error = "没有改成"
			showAuth(c, db, http.StatusInternalServerError, view)
			return
		}
		if err := db.Model(&user).Update("password_hash", string(hash)).Error; err != nil {
			view.Error = "没有改成"
			showAuth(c, db, http.StatusInternalServerError, view)
			return
		}
		db.Where("user_id = ?", user.ID).Delete(&models.PasswordReset{})
		setSessionCookie(c, user)
		c.Redirect(http.StatusSeeOther, "/mine")
	}
}

func resetUser(db *gorm.DB, token string) (models.User, bool) {
	if len(token) < 32 {
		return models.User{}, false
	}
	var row models.PasswordReset
	err := db.Where("token_hash = ? AND expires_at > ?", hashResetToken(token), time.Now()).First(&row).Error
	if err != nil {
		return models.User{}, false
	}
	var user models.User
	if err := db.First(&user, row.UserID).Error; err != nil {
		return models.User{}, false
	}
	return user, true
}

func emailTaken(db *gorm.DB, email string, exceptID uint) bool {
	var n int64
	q := db.Model(&models.User{}).Where("email = ?", email)
	if exceptID != 0 {
		q = q.Where("id <> ?", exceptID)
	}
	q.Count(&n)
	return n > 0
}

func saveEmail(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c, db)
		if !ok {
			c.Redirect(http.StatusFound, "/login?next=/mine")
			return
		}
		email := normalizeEmail(c.PostForm("email"))
		password := c.PostForm("password")
		fail := func(msg string) {
			renderMine(c, db, user, "", msg, email)
		}
		if !validEmail(email) {
			fail("邮箱格式不对")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
			fail("当前密码不对")
			return
		}
		if emailTaken(db, email, user.ID) {
			fail("这个邮箱已经注册过")
			return
		}
		if err := db.Model(&user).Update("email", email).Error; err != nil {
			fail("没有保存")
			return
		}
		c.Redirect(http.StatusSeeOther, "/mine?mail=1")
	}
}
