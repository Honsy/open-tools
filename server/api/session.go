package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const sessionCookie = "ot_reader"

var sessionKey string

func sessionSecret() string {
	// Keep existing user sessions valid as the same session starts serving the admin UI.
	return sessionKey + ":reader"
}

func currentUser(c *gin.Context, db *gorm.DB) (models.User, bool) {
	token, _ := c.Cookie(sessionCookie)
	if token == "" {
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		}
	}
	idText, ok := parseToken(token, sessionSecret())
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

func setSessionCookie(c *gin.Context, user models.User) {
	c.SetSameSite(http.SameSiteLaxMode)
	secure := c.Request.TLS != nil || headerFirst(c, "X-Forwarded-Proto") == "https"
	c.SetCookie(sessionCookie, makeToken(strconv.FormatUint(uint64(user.ID), 10), sessionSecret()), 7*24*3600, "/", "", secure, true)
}

func clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	secure := c.Request.TLS != nil || headerFirst(c, "X-Forwarded-Proto") == "https"
	c.SetCookie(sessionCookie, "", -1, "/", "", secure, true)
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
