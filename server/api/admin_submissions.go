package api

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type adminSubmission struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Desc         string    `json:"desc"`
	CategorySlug string    `json:"categorySlug"`
	UserID       uint      `json:"userId"`
	Submitter    string    `json:"submitter"`
	CreatedAt    time.Time `json:"createdAt"`
	Duplicate    bool      `json:"duplicate"`
}

func adminSubmissions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows := make([]adminSubmission, 0)
		query := db.WithContext(c.Request.Context()).Model(&models.Link{}).
			Select("links.id, links.name, links.url, links.`desc`, links.category_slug, links.user_id, links.created_at, COALESCE(u.username, '') AS submitter, "+
				"EXISTS (SELECT 1 FROM links AS existing WHERE existing.status = 'online' AND existing.url = links.url) AS `duplicate`").
			Joins("LEFT JOIN users AS u ON u.id = links.user_id").Where("links.status = ?", "pending").Order("links.id desc")
		loadAdminPage(c, query, &rows, "读取待审网址失败")
	}
}

func adminSubmissionReviews(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows := make([]models.SubmissionReview, 0)
		query := db.WithContext(c.Request.Context()).Model(&models.SubmissionReview{}).Order("id desc")
		loadAdminPage(c, query, &rows, "读取审核记录失败")
	}
}

func reviewSubmission(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "网址编号无效"})
			return
		}
		var req struct {
			Decision     string `json:"decision"`
			CategorySlug string `json:"categorySlug"`
			Note         string `json:"note"`
		}
		if c.ShouldBindJSON(&req) != nil || (req.Decision != "approve" && req.Decision != "reject" && req.Decision != "block") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "审核操作无效"})
			return
		}
		req.Note = strings.TrimSpace(req.Note)
		if utf8Count(req.Note) > 300 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "审核备注不能超过 300 字"})
			return
		}
		query := db.WithContext(c.Request.Context())
		var link models.Link
		if err := query.First(&link, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "没有这条网址"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "读取网址失败"})
			}
			return
		}
		if link.Status != "pending" {
			c.JSON(http.StatusConflict, gin.H{"error": "这条网址已完成审核，请刷新列表"})
			return
		}
		updates := map[string]any{"status": "off"}
		var blocked string
		if req.Decision != "reject" {
			target, err := url.Parse(link.URL)
			if err != nil || target.User != nil || (target.Scheme != "http" && target.Scheme != "https") || rankDomain(target.Hostname()) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "这条提交没有可用的外部网址，可以选择拒绝"})
				return
			}
			if req.Decision == "block" {
				blocked = rankDomain(target.Hostname())
			} else {
				var category models.Category
				err := query.Where("slug = ? AND kind = ? AND content_kind = ?", req.CategorySlug, "tab", "links").First(&category).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "请选择一个网址子分类"})
					return
				}
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "读取分类失败"})
					return
				}
				updates["status"], updates["category_slug"] = "online", category.Slug
				updates["check_fails"], updates["checked_at"] = 0, nil
			}
		}
		err = query.Transaction(func(tx *gorm.DB) error {
			result := tx.Model(&models.Link{}).Where("id = ? AND status = ? AND url = ?", link.ID, "pending", link.URL).Updates(updates)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
			if blocked != "" {
				if err := blockRankDomain(tx, blocked); err != nil {
					return err
				}
			}
			category, _ := updates["category_slug"].(string)
			return tx.Create(&models.SubmissionReview{
				LinkID: link.ID, Name: link.Name, URL: link.URL, Decision: req.Decision,
				CategorySlug: category, Reviewer: c.GetString("adminUser"), Note: req.Note,
			}).Error
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusConflict, gin.H{"error": "这条网址已变更，请刷新列表"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存审核结果失败"})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
