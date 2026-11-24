package api

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
		}
		if c.ShouldBindJSON(&req) != nil || (req.Decision != "approve" && req.Decision != "reject" && req.Decision != "block") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "审核操作无效"})
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
				return blockRankDomain(tx, blocked)
			}
			return nil
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
