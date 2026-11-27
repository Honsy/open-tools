package api

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type reviewTestPool struct{ gorm.ConnPool }
type reviewTestTx struct{ gorm.ConnPool }

func (p reviewTestPool) BeginTx(context.Context, *sql.TxOptions) (gorm.ConnPool, error) {
	return &reviewTestTx{ConnPool: p.ConnPool}, nil
}

func (*reviewTestTx) Commit() error   { return nil }
func (*reviewTestTx) Rollback() error { return nil }

func TestApprovePendingSubmissionPublishesOnlyAfterSelectingCategory(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "unused@tcp(127.0.0.1:3306)/unused", SkipInitializeWithVersion: true}), &gorm.Config{
		DryRun: true, SkipDefaultTransaction: true, DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	db.Statement.ConnPool = reviewTestPool{ConnPool: db.Statement.ConnPool}
	if err := db.Callback().Query().After("gorm:query").Register("review:fixture", func(q *gorm.DB) {
		switch dest := q.Statement.Dest.(type) {
		case *models.Link:
			*dest = models.Link{ID: 23, Status: "pending", CategorySlug: "inbox", URL: "https://example.org/"}
			q.RowsAffected = 1
		case *models.Category:
			if q.Statement.Vars[0] != "learn" {
				q.AddError(gorm.ErrRecordNotFound)
				return
			}
			*dest = models.Category{Slug: "learn", Kind: "tab", ContentKind: "links"}
			q.RowsAffected = 1
		}
	}); err != nil {
		t.Fatal(err)
	}
	var update string
	if err := db.Callback().Update().After("gorm:update").Register("review:capture", func(q *gorm.DB) {
		q.RowsAffected = 1
		update = q.Dialector.Explain(q.Statement.SQL.String(), q.Statement.Vars...)
	}); err != nil {
		t.Fatal(err)
	}
	var recorded *models.SubmissionReview
	if err := db.Callback().Create().After("gorm:create").Register("review:history", func(q *gorm.DB) {
		recorded, _ = q.Statement.Dest.(*models.SubmissionReview)
	}); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.PATCH("/submissions/:id", func(c *gin.Context) { c.Set("adminUser", "editor") }, reviewSubmission(db))
	review := func(category string) int {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/submissions/23", strings.NewReader(`{"decision":"approve","categorySlug":"`+category+`","note":"已核对"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, req)
		return recorder.Code
	}
	if code := review("inbox"); code != http.StatusBadRequest || update != "" {
		t.Fatalf("approving inbox: %d, update %q", code, update)
	}
	if code := review("learn"); code != http.StatusNoContent {
		t.Fatalf("approving into learn: %d", code)
	}
	if recorded == nil || recorded.LinkID != 23 || recorded.Reviewer != "editor" || recorded.Note != "已核对" || recorded.CategorySlug != "learn" {
		t.Fatalf("unexpected review history: %+v", recorded)
	}
	for _, expected := range []string{"`category_slug`='learn'", "`status`='online'", "WHERE id = 23 AND status = 'pending'"} {
		if !strings.Contains(update, expected) {
			t.Errorf("missing %q in %s", expected, update)
		}
	}
	for _, protected := range []string{"`user_id`=", "`desc`=", "`clicks`=", "`views`="} {
		if strings.Contains(update, protected) {
			t.Errorf("approval changed unrelated field %q: %s", protected, update)
		}
	}
}
