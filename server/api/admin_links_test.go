package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestAdminLinksPaginatesFilteredResults(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "unused@tcp(127.0.0.1:3306)/unused", SkipInitializeWithVersion: true}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var queries []string
	if err := db.Callback().Query().After("gorm:query").Register("admin:links:test", func(q *gorm.DB) {
		queries = append(queries, q.Dialector.Explain(q.Statement.SQL.String(), q.Statement.Vars...))
		switch rows := q.Statement.Dest.(type) {
		case *int64:
			*rows = 76
		case *[]models.Link:
			*rows = []models.Link{{ID: 42, Name: "示例"}}
		}
		q.RowsAffected = 1
	}); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/links", adminLinks(db))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/links?page=2&category=learn&status=online&q=foo%25", nil))
	var response struct {
		Items          []models.Link
		Page, PageSize int
		Total          int
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || recorder.Code != http.StatusOK ||
		response.Total != 76 || response.Page != 2 || response.PageSize != 25 || len(response.Items) != 1 {
		t.Fatalf("unexpected paginated response %d: %s", recorder.Code, recorder.Body.String())
	}
	if len(queries) != 2 || !strings.Contains(queries[0], "count(*)") || !strings.Contains(queries[1], "LIMIT 25 OFFSET 25") {
		t.Fatalf("unexpected queries: %v", queries)
	}
	for _, query := range queries {
		if !strings.Contains(query, "status = 'online' AND category_slug = 'learn' AND (name ILIKE '%foo!%%'") {
			t.Fatalf("expected both queries to apply the same filters: %s", query)
		}
	}
}
