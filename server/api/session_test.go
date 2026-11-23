package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSharedLoginChecksAdminRoleOnEveryRequest(t *testing.T) {
	oldSessionKey := sessionKey
	sessionKey = "test-session-key"
	t.Cleanup(func() { sessionKey = oldSessionKey })
	password := "test-password"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := models.User{ID: 7, Username: "account", PasswordHash: string(hash)}
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "unused@tcp(127.0.0.1:3306)/unused", SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Query().Replace("gorm:query", func(q *gorm.DB) {
		*q.Statement.Dest.(*models.User) = user
		q.RowsAffected = 1
	}); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET(siteAPIPrefix+"/login", loginPage(db))
	router.POST(siteAPIPrefix+"/login", loginPost(db))
	router.POST(siteAPIPrefix+"/logout", logout)
	MountAdmin(router, db)

	adminRequest := func(method, path string, cookie *http.Cookie, site string) int {
		t.Helper()
		request := httptest.NewRequest(method, "https://kaiwulu.example"+path, nil)
		if cookie != nil {
			request.AddCookie(cookie)
		}
		if site != "" {
			request.Header.Set("Sec-Fetch-Site", site)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response.Code
	}
	if got := adminRequest(http.MethodGet, "/api/admin/me", nil, ""); got != http.StatusUnauthorized {
		t.Fatalf("admin without cookie = %d, want 401", got)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, siteAPIPrefix+"/login?next=/admin/links", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != sitePageType ||
		!strings.Contains(response.Body.String(), `"Next":"/admin/links"`) {
		t.Fatalf("login page = %d, missing the admin destination", response.Code)
	}

	form := url.Values{"username": {user.Username}, "password": {password}, "next": {"/admin/links"}}
	request := httptest.NewRequest(http.MethodPost, "https://kaiwulu.example"+siteAPIPrefix+"/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/admin/links" {
		t.Fatalf("login = %d, location %q", response.Code, response.Header().Get("Location"))
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookie || !cookies[0].Secure || !cookies[0].HttpOnly {
		t.Fatal("login did not set the secure shared session cookie")
	}
	cookie := cookies[0]
	if got := adminRequest(http.MethodGet, "/api/admin/me", cookie, ""); got != http.StatusForbidden {
		t.Fatalf("ordinary user in admin = %d, want 403", got)
	}
	user.IsAdmin = true
	if got := adminRequest(http.MethodGet, "/api/admin/me", cookie, ""); got != http.StatusOK {
		t.Fatalf("admin with the shared session = %d, want 200", got)
	}
	if got := adminRequest(http.MethodPost, "/api/admin/tags", cookie, "cross-site"); got != http.StatusForbidden {
		t.Fatalf("cross-site admin write = %d, want 403", got)
	}
	user.IsAdmin = false
	if got := adminRequest(http.MethodGet, "/api/admin/me", cookie, ""); got != http.StatusForbidden {
		t.Fatalf("admin after role removal = %d, want 403", got)
	}

	request = httptest.NewRequest(http.MethodPost, "https://kaiwulu.example"+siteAPIPrefix+"/logout", nil)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	cookies = response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookie || cookies[0].MaxAge >= 0 {
		t.Fatal("logout did not clear the shared session cookie")
	}
}

func TestSafeNext(t *testing.T) {
	for _, path := range []string{"//example.com", `/\example.com`, "/\t/example.com", "https://example.com"} {
		if got := safeNext(path); got != "/" {
			t.Errorf("safeNext(%q) = %q, want /", path, got)
		}
	}
	for _, path := range []string{"/admin/links?status=pending", "/login?next=/admin/", "/admin/#review"} {
		if got := safeNext(path); got != path {
			t.Errorf("safeNext(%q) = %q", path, got)
		}
	}
}
