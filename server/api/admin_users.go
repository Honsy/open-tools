package api

import (
	"net/http"
	"strings"
	"time"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type adminUserRow struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"isAdmin"`
	CreatedAt time.Time `json:"createdAt"`
}

func adminUsers(db *gorm.DB) gin.HandlerFunc {
	const pageSize = 50
	escapeLike := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
	return func(c *gin.Context) {
		query := db.WithContext(c.Request.Context()).Model(&models.User{})
		if text := strings.TrimSpace(c.Query("q")); text != "" {
			if utf8Count(text) > 254 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "搜索内容不能超过 254 个字符"})
				return
			}
			like := "%" + escapeLike.Replace(text) + "%"
			query = query.Where("username LIKE ? ESCAPE '!' OR email LIKE ? ESCAPE '!'", like, like)
		}
		readError := gin.H{"error": "读取用户列表失败"}
		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, readError)
			return
		}
		pageCount := max(1, int((total+pageSize-1)/pageSize))
		page := min(positivePage(c.Query("page")), pageCount)
		rows := make([]adminUserRow, 0)
		if err := query.Select("id", "username", "email", "is_admin", "created_at").Order("id desc").
			Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, readError)
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": rows, "total": total, "page": page, "pageCount": pageCount})
	}
}
