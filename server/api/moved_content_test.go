package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"opentools/config"
	"opentools/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestMigratedPublicRoutes(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "unused@tcp(127.0.0.1:3306)/unused", SkipInitializeWithVersion: true}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	router := Router(db)
	MountLab(router, db, config.Config{})
	for _, test := range []struct {
		path, target string
	}{
		{"/ai/tools", "/c/ai"}, {"/ai/articles", "/c/daily"}, {"/ranking/ai", "/rank?category=ai"},
		{"/ai/lab/mortgage-calc", "/tools/mortgage-calc"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusMovedPermanently || response.Header().Get("Location") != test.target {
			t.Errorf("%s got status %d and location %q", test.path, response.Code, response.Header().Get("Location"))
		}
	}
	for _, test := range []struct {
		path, content string
	}{
		{"/tools/", "AI 图片换背景"}, {"/tools/mortgage-calc", "商业贷款年利率"},
		{"/tools/horoscope", "登录后使用"}, {"/tools/sbti", "/static/lab/sbti-page.js"},
		{"/feedback", "登录后反馈"}, {"/static/lab/tools.js", "retirementPlan"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), test.content) {
			t.Errorf("%s got status %d, expected content %q", test.path, response.Code, test.content)
		}
		if strings.Contains(response.Body.String(), "data-ai-tool=") {
			t.Errorf("%s renders a signed-in AI form to a guest", test.path)
		}
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/tools/poster", nil))
	if response.Code != http.StatusUnauthorized {
		t.Errorf("guest AI API: got status %d, want 401", response.Code)
	}
}

func TestImportedArticleFormattingSanitizesHTML(t *testing.T) {
	item := models.Article{Format: "markdown", Body: "# 标题\n\n[参考](https://example.test)\n\n<img src='https://example.test/a.png' onerror='alert(1)'><script>unexpected()</script>"}
	rendered := string(articleHTML(item))
	for _, want := range []string{"<h1>标题</h1>", "https://example.test", "<img"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("missing %q in %q", want, rendered)
		}
	}
	for _, blocked := range []string{"onerror", "unexpected()", "<script"} {
		if strings.Contains(rendered, blocked) {
			t.Errorf("unsafe %q in %q", blocked, rendered)
		}
	}
}
