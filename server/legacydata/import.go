package legacydata

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"opentools/models"
	"opentools/seed"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Snapshot struct {
	Sites    []legacySite
	Articles []legacyArticle
	Comments []legacyComment
	Messages []legacyMessage
}

type legacySite struct {
	ID                     uint
	Name, URL, Description string
	Category               string
	Tags                   []byte
	Sort, Status           int
	CreatedAt              time.Time
}

type legacyArticle struct {
	ID, Slug, Title, Summary string
	Content, CoverImage      string
	VideoURL, Category       string
	ViewCount, SortOrder     int
	Featured                 bool
	Status                   int
	CreatedAt                time.Time
}

type legacyComment struct {
	ID        uint
	ParentID  *uint
	TargetID  string
	Author    string
	Content   string
	Status    int
	CreatedAt time.Time
}

type legacyMessage struct {
	ID        uint
	Type      int
	Content   string
	Status    int
	CreatedAt time.Time
}

func ReadSnapshot(db *gorm.DB) (Snapshot, error) {
	var data Snapshot
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("ai_sites").Select("id", "name", "url", "description", "category", "tags", "sort", "status", "created_at").Order("status desc, id asc").Find(&data.Sites).Error; err != nil {
			return fmt.Errorf("read AI sites: %w", err)
		}
		if err := tx.Table("ai_articles").Order("id").Find(&data.Articles).Error; err != nil {
			return fmt.Errorf("read articles: %w", err)
		}
		if err := tx.Table("comments").Select("comments.id, comments.parent_id, comments.target_id, COALESCE(users.username, '') AS author, comments.content, comments.status, comments.created_at").
			Joins("LEFT JOIN users ON users.id = comments.user_id").Where("comments.target_type = ?", "article").Order("comments.id").Find(&data.Comments).Error; err != nil {
			return fmt.Errorf("read article comments: %w", err)
		}
		if err := tx.Table("memos").Select("id", "type", "content", "status", "created_at").Where("type = ? OR (type = ? AND relation_type = ?)", 3, 2, "ai_product").Order("id").Find(&data.Messages).Error; err != nil {
			return fmt.Errorf("read site messages: %w", err)
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return data, err
}

type Result struct {
	Sites, Articles, Comments, Messages         int
	ExistingSites, InvalidSites, OrphanComments int
}

func Import(db *gorm.DB, data Snapshot) (Result, error) {
	var result Result
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := importSites(tx, data.Sites, &result); err != nil {
			return err
		}
		articleIDs, err := importArticles(tx, data.Articles, &result)
		if err != nil {
			return err
		}
		if err := importComments(tx, data.Comments, articleIDs, &result); err != nil {
			return err
		}
		for _, old := range data.Messages {
			status := "pending"
			switch old.Status {
			case 1:
				status = "done"
			case 2:
				status = "rejected"
			}
			kind := "message"
			if old.Type == 2 {
				kind = "ai-submission"
			}
			entry := models.Feedback{LegacyID: &old.ID, Kind: kind, Body: old.Content, Status: status, CreatedAt: old.CreatedAt}
			insert := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&entry)
			if insert.Error != nil {
				return insert.Error
			}
			result.Messages += int(insert.RowsAffected)
		}
		return nil
	})
	return result, err
}

func importSites(tx *gorm.DB, sites []legacySite, result *Result) error {
	var existing []models.Link
	if err := tx.Select("id", "url", "imported_from").Find(&existing).Error; err != nil {
		return err
	}
	urls := make(map[string]bool, len(existing))
	imported := make(map[string]bool)
	for _, link := range existing {
		urls[siteIdentity(link.URL)] = true
		if link.ImportedFrom != nil {
			imported[*link.ImportedFrom] = true
		}
	}
	var ai models.Category
	if len(sites) > 0 {
		if err := tx.Where("slug = ?", "ai").First(&ai).Error; err != nil {
			return fmt.Errorf("AI 分类未初始化: %w", err)
		}
	}
	categorySlugs := make(map[string]string)
	for _, old := range sites {
		address := strings.TrimSpace(old.URL)
		key := siteIdentity(address)
		if key == "" || len([]rune(address)) > 512 {
			result.InvalidSites++
			continue
		}
		sourceID := "laoyou8:ai-site:" + strconv.FormatUint(uint64(old.ID), 10)
		if urls[key] || imported[sourceID] {
			result.ExistingSites++
			continue
		}
		category := strings.TrimSpace(old.Category)
		if category == "" {
			category = "其他 AI 工具"
		}
		categorySlug := categorySlugs[category]
		if categorySlug == "" {
			hash := sha256.Sum256([]byte(category))
			categorySlug = fmt.Sprintf("ai-%x", hash[:6])
			newCategory := models.Category{Slug: categorySlug, Name: take(category, 64), ParentSlug: ai.Slug, Kind: "tab", ContentKind: "links", Sort: 20 + len(categorySlugs)}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&newCategory).Error; err != nil {
				return err
			}
			categorySlugs[category] = categorySlug
		}
		status := "off"
		if old.Status == 1 {
			status = "online"
		}
		var tags []string
		if len(old.Tags) != 0 {
			if err := json.Unmarshal(old.Tags, &tags); err != nil {
				return fmt.Errorf("AI 站点 %d 标签格式无效: %w", old.ID, err)
			}
		}
		row := models.Link{
			ImportedFrom: &sourceID, CategorySlug: categorySlug, URL: address, Name: take(old.Name, 128),
			Desc: take(old.Description, 256), Body: old.Description, Tags: take(strings.Join(tags, ","), 160),
			Sort: old.Sort, Status: status, CreatedAt: old.CreatedAt,
		}
		seed.FillSlug(tx, &row)
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		urls[key] = true
		result.Sites++
	}
	return nil
}

func siteIdentity(raw string) string {
	value, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || value.Hostname() == "" || value.User != nil || (value.Scheme != "http" && value.Scheme != "https") {
		return ""
	}
	host := strings.ToLower(value.Hostname())
	if port := value.Port(); port != "" && !(value.Scheme == "https" && port == "443") && !(value.Scheme == "http" && port == "80") {
		host += ":" + port
	}
	query := value.Query()
	for key := range query {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || lower == "invitecode" || lower == "fbclid" || lower == "gclid" {
			query.Del(key)
		}
	}
	return host + strings.TrimRight(value.EscapedPath(), "/") + "?" + query.Encode()
}

func importArticles(tx *gorm.DB, rows []legacyArticle, result *Result) (map[string]uint, error) {
	ids := make(map[string]uint, len(rows))
	for _, old := range rows {
		var entry models.Article
		err := tx.Select("id").Where("legacy_id = ?", old.ID).First(&entry).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			body := replaceMediaHost(old.Content)
			// Attach the old URL to an exact existing copy without changing editable fields.
			var matches []models.Article
			if err := tx.Select("id", "body").Where("legacy_id IS NULL AND title = ? AND draft = ?", old.Title, old.Status != 1).Order("id").Find(&matches).Error; err != nil {
				return nil, err
			}
			for _, match := range matches {
				if match.Body == body {
					entry = match
					break
				}
			}
			if entry.ID != 0 {
				err = tx.Model(&entry).Updates(map[string]any{"legacy_id": old.ID, "legacy_slug": old.Slug}).Error
			} else {
				entry = models.Article{
					LegacyID: &old.ID, LegacySlug: old.Slug, Title: old.Title, Summary: old.Summary, Body: body,
					Format: articleFormat(body), CoverURL: replaceMediaHost(old.CoverImage), VideoURL: replaceMediaHost(old.VideoURL),
					Category: old.Category, Sort: old.SortOrder, Featured: old.Featured, Draft: old.Status != 1,
					Views: old.ViewCount, CreatedAt: old.CreatedAt,
				}
				err = tx.Create(&entry).Error
				if err == nil {
					result.Articles++
				}
			}
		}
		if err != nil {
			return nil, err
		}
		ids[old.ID] = entry.ID
	}
	return ids, nil
}

func importComments(tx *gorm.DB, rows []legacyComment, articles map[string]uint, result *Result) error {
	type commentLocation struct{ ID, ArticleID uint }
	ids := make(map[uint]commentLocation, len(rows))
	var inserted []legacyComment
	for _, old := range rows {
		article := articles[old.TargetID]
		if article == 0 {
			result.OrphanComments++
			continue
		}
		var entry models.ArticleComment
		err := tx.Select("id").Where("legacy_id = ?", old.ID).First(&entry).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			author := old.Author
			if author == "" {
				author = "原站用户"
			}
			entry = models.ArticleComment{LegacyID: &old.ID, ArticleID: article, Author: take(author, 128), Body: old.Content, Hidden: old.Status != 1, CreatedAt: old.CreatedAt}
			err = tx.Create(&entry).Error
			if err == nil {
				inserted = append(inserted, old)
				result.Comments++
			}
		}
		if err != nil {
			return err
		}
		ids[old.ID] = commentLocation{ID: entry.ID, ArticleID: article}
	}
	for _, old := range inserted {
		if old.ParentID == nil || *old.ParentID == old.ID {
			continue
		}
		parent := ids[*old.ParentID]
		child := ids[old.ID]
		if parent.ID != 0 && parent.ArticleID == child.ArticleID {
			if err := tx.Model(&models.ArticleComment{}).Where("id = ?", child.ID).Update("parent_id", parent.ID).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

var htmlBlock = regexp.MustCompile(`(?i)<\s*(p|div|h[1-6]|ul|ol|li|table|blockquote|pre|img|a|br|hr|span|strong|em)\b`)
var markdownBlock = regexp.MustCompile(`(?m)^\s*(#{1,6}\s|[-*+]\s|>\s|\x60{3})|\[[^\]]+\]\([^\n]+\)|\*\*[^*]+\*\*`)

func articleFormat(body string) string {
	if htmlBlock.MatchString(body) {
		return "html"
	}
	if markdownBlock.MatchString(body) {
		return "markdown"
	}
	return ""
}

func replaceMediaHost(value string) string {
	return strings.NewReplacer(
		"https://static.laoyou8.com/", "https://static.laoyou8.app/",
		"http://static.laoyou8.com/", "https://static.laoyou8.app/",
	).Replace(value)
}

func take(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
