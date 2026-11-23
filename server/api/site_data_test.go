package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"opentools/models"
	"opentools/seed"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSiteDetailUsesEditorialIntroAndSuppressesPlaceholder(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "unused@tcp(127.0.0.1:3306)/unused", SkipInitializeWithVersion: true}), &gorm.Config{
		DryRun: true, SkipDefaultTransaction: true, DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	link := models.Link{ID: 1, Slug: "poe", Name: "Poe", Desc: "多模型 AI 聊天", Status: "online"}
	if err := db.Callback().Query().Replace("gorm:query", func(q *gorm.DB) {
		if dest, ok := q.Statement.Dest.(*models.Link); ok {
			*dest = link
			q.RowsAffected = 1
		}
	}); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET(siteAPIPrefix+"/site/:slug", sitePage(db))
	for _, example := range []struct{ body, summary, heading string }{
		{seed.LegacyPlaceholderIntro("Poe", "多个模型对话"), "Poe｜多模型 AI 聊天", ""},
		{"Poe 可以切换模型。\n\n### Poe 怎么选模型？\n\n从目录搜索模型。", "Poe 可以切换模型。", "<h3>Poe 怎么选模型？</h3>"},
	} {
		link.Body = example.body
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "https://kaiwulu.example/_site/site/poe", nil))
		var view struct {
			Shell shell
			Data  struct{ BodyHTML string }
		}
		if err := json.Unmarshal(response.Body.Bytes(), &view); err != nil || response.Code != http.StatusOK {
			t.Fatalf("site response = %d: %s", response.Code, response.Body.String())
		}
		if view.Shell.Description != example.summary || view.Shell.Canonical != "https://kaiwulu.example/site/poe" {
			t.Fatalf("unexpected metadata: %+v", view.Shell)
		}
		if example.heading == "" && view.Data.BodyHTML != "" || !strings.Contains(view.Data.BodyHTML, example.heading) {
			t.Fatalf("unexpected site introduction: %q", view.Data.BodyHTML)
		}
	}
}

func TestPaginatePreviewReadsOnlyFirstPage(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "unused@tcp(127.0.0.1:3306)/unused", SkipInitializeWithVersion: true}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	var queries []string
	if err := db.Callback().Query().After("gorm:query").Register("preview:test", func(q *gorm.DB) {
		queries = append(queries, q.Dialector.Explain(q.Statement.SQL.String(), q.Statement.Vars...))
		switch dest := q.Statement.Dest.(type) {
		case *[]models.Link:
			*dest = make([]models.Link, 24)
			q.RowsAffected = 24
		case *int64:
			*dest = 53
			q.RowsAffected = 1
		}
	}); err != nil {
		t.Fatal(err)
	}
	tab := tabView{Slug: "video", On: true}
	paginatePreview(db, &tab, "row")
	if len(queries) != 2 || !strings.Contains(queries[0], "ORDER BY sort asc, id asc LIMIT 24") ||
		!strings.Contains(queries[1], "count(*)") {
		t.Fatalf("expected a limited query and a count, got %v", queries)
	}
	for _, query := range queries {
		if !strings.Contains(query, "category_slug = 'video' AND status = 'online'") {
			t.Fatalf("preview query did not filter active links in its category: %q", query)
		}
	}
	if len(tab.Links) != 24 || tab.Total != 53 || tab.Page != 1 || tab.PageCount != 3 || tab.NextURL != "/c/video?page=2" {
		t.Fatalf("unexpected preview pagination: %d links, total %d, page %d/%d, next %q",
			len(tab.Links), tab.Total, tab.Page, tab.PageCount, tab.NextURL)
	}
}
