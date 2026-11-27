package api

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const internetRankPath = "/rank/monthly"
const maxRankEntries = 10000
const maxRankUpload = 4 << 20

func rankDomain(raw string) string {
	name := strings.TrimSuffix(strings.TrimSpace(raw), ".")
	ascii, err := idna.Lookup.ToASCII(name)
	if err != nil {
		return ""
	}
	domain := strings.TrimPrefix(strings.ToLower(ascii), "www.")
	if len(domain) > 253 || !hostName.MatchString(domain) {
		return ""
	}
	lastLabel := domain[strings.LastIndexByte(domain, '.')+1:]
	if _, icann := publicsuffix.PublicSuffix(lastLabel); !icann {
		return ""
	}
	if suffix, icann := publicsuffix.PublicSuffix(domain); icann && suffix == domain {
		return ""
	}
	return domain
}

func mountAdminInternetRank(r *gin.RouterGroup, db *gorm.DB) {
	r.GET("/internet-rank", func(c *gin.Context) {
		months := make([]models.InternetRankMonth, 0)
		blocked := make([]models.InternetRankBlock, 0)
		if err := db.Order("month desc").Find(&months).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取月榜失败"})
			return
		}
		if err := db.Order("domain asc").Find(&blocked).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取风险域名失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"months": months, "blocked": blocked})
	})
	r.POST("/internet-rank/preview", previewInternetRank(db))
	r.POST("/internet-rank", importInternetRank(db))
	r.POST("/internet-rank/blocked", func(c *gin.Context) {
		var req struct {
			Domain string `json:"domain"`
		}
		err := c.ShouldBindJSON(&req)
		domain := rankDomain(req.Domain)
		if err != nil || domain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入不带协议和路径的域名"})
			return
		}
		if err := blockRankDomain(db, domain); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存风险域名失败"})
			return
		}
		c.Status(http.StatusNoContent)
	})
	r.DELETE("/internet-rank/blocked/:domain", func(c *gin.Context) {
		domain := rankDomain(c.Param("domain"))
		if domain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "域名格式不正确"})
			return
		}
		if err := db.Where("domain = ?", domain).Delete(&models.InternetRankBlock{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "移除风险域名失败"})
			return
		}
		c.Status(http.StatusNoContent)
	})
}

func blockRankDomain(db *gorm.DB, domain string) error {
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.InternetRankBlock{Domain: domain}).Error
}

func parseRankUpload(c *gin.Context) (models.InternetRankMonth, []models.InternetRankEntry, error) {
	var snapshot models.InternetRankMonth
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRankUpload)
	if err := c.Request.ParseMultipartForm(maxRankUpload); err != nil {
		return snapshot, nil, errors.New("请选择不超过 4 MB 的 CSV 文件")
	}
	defer c.Request.MultipartForm.RemoveAll()
	month := strings.TrimSpace(c.PostForm("month"))
	source := strings.TrimSpace(c.PostForm("source"))
	sourceURL := strings.TrimSpace(c.PostForm("sourceUrl"))
	if date, err := time.Parse("2006-01", month); err != nil || date.Year() == 0 || month >= time.Now().In(time.FixedZone("Asia/Shanghai", 8*3600)).Format("2006-01") {
		return snapshot, nil, errors.New("请选择已结束的统计月份")
	}
	if !rankText(source, 128) || len(sourceURL) > 512 || !strings.HasPrefix(publicMedia(sourceURL), "https://") {
		return snapshot, nil, errors.New("请填写数据来源名称和 HTTPS 说明地址")
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		return snapshot, nil, errors.New("请选择 CSV 文件")
	}
	defer file.Close()
	entries, err := readInternetRankCSV(file, month)
	return models.InternetRankMonth{Month: month, Source: source, SourceURL: sourceURL, Entries: len(entries)}, entries, err
}

func rankVersion(db *gorm.DB, month string) (string, error) {
	var existing models.InternetRankMonth
	err := db.Where("month = ?", month).Take(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "new", nil
	}
	return existing.UpdatedAt.Format(time.RFC3339Nano), err
}

func previewInternetRank(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		snapshot, entries, err := parseRankUpload(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := db.WithContext(c.Request.Context())
		var blocks []string
		if err := query.Model(&models.InternetRankBlock{}).Pluck("domain", &blocks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取风险名单失败"})
			return
		}
		version, err := rankVersion(query, snapshot.Month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取已有榜单失败"})
			return
		}
		visible, excluded := rankPreviewEntries(entries, blocks)
		preview := make([]gin.H, 0, min(10, len(visible)))
		for _, entry := range visible[:min(10, len(visible))] {
			preview = append(preview, gin.H{"Domain": entry.Domain, "Name": entry.Name, "Category": entry.Category, "Visits": strconv.FormatUint(entry.Visits, 10)})
		}
		c.JSON(http.StatusOK, gin.H{
			"month": snapshot.Month, "total": len(entries), "excluded": len(entries) - len(visible),
			"version": version, "entries": preview, "excludedEntries": excluded,
		})
	}
}

type rankPreviewExclusion struct {
	Domain string `json:"domain"`
	Reason string `json:"reason"`
}

func rankPreviewEntries(entries []models.InternetRankEntry, blocks []string) ([]models.InternetRankEntry, []rankPreviewExclusion) {
	blocked := make(map[string]bool, len(blocks))
	for _, domain := range blocks {
		blocked[domain] = true
	}
	visible := make([]models.InternetRankEntry, 0, len(entries))
	excluded := make([]rankPreviewExclusion, 0, 10)
	for _, entry := range entries {
		if entry.Risky {
			if len(excluded) < 10 {
				excluded = append(excluded, rankPreviewExclusion{entry.Domain, "数据源风险标记"})
			}
			continue
		}
		domain := entry.Domain
		for domain != "" && !blocked[domain] {
			_, domain, _ = strings.Cut(domain, ".")
		}
		if domain == "" {
			visible = append(visible, entry)
		} else if len(excluded) < 10 {
			excluded = append(excluded, rankPreviewExclusion{entry.Domain, "命中风险名单"})
		}
	}
	sort.Slice(visible, func(i, j int) bool {
		if visible[i].Visits == visible[j].Visits {
			return visible[i].Domain < visible[j].Domain
		}
		return visible[i].Visits > visible[j].Visits
	})
	return visible, excluded
}

var errRankPreviewStale = errors.New("该月份已经被更新，请重新预览后再发布")

func importInternetRank(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		snapshot, entries, err := parseRankUpload(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		expectedVersion := c.PostForm("previewVersion")
		err = db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			if expectedVersion != "" {
				current, err := rankVersion(tx.Clauses(clause.Locking{Strength: "UPDATE"}), snapshot.Month)
				if err != nil {
					return err
				}
				if current != expectedVersion {
					return errRankPreviewStale
				}
			}
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&snapshot).Error; err != nil {
				return err
			}
			if err := tx.Where("month = ?", snapshot.Month).Delete(&models.InternetRankEntry{}).Error; err != nil {
				return err
			}
			return tx.CreateInBatches(entries, 500).Error
		})
		if errors.Is(err, errRankPreviewStale) {
			c.JSON(http.StatusConflict, gin.H{"error": errRankPreviewStale.Error()})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "导入失败，原有月榜已保留"})
			return
		}
		c.JSON(http.StatusOK, snapshot)
	}
}

func rankText(value string, max int) bool {
	return value != "" && utf8.ValidString(value) && utf8.RuneCountInString(value) <= max && !strings.ContainsAny(value, "\x00\r\n")
}

func readInternetRankCSV(file io.Reader, month string) ([]models.InternetRankEntry, error) {
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, errors.New("无法读取 CSV 表头")
	}
	columns := make(map[string]int, len(header))
	for i, value := range header {
		key := strings.TrimSpace(strings.TrimPrefix(value, "\ufeff"))
		if _, duplicate := columns[key]; duplicate {
			return nil, fmt.Errorf("CSV 表头重复：%s", key)
		}
		columns[key] = i
	}
	for _, key := range []string{"domain", "category", "visits"} {
		if _, ok := columns[key]; !ok {
			return nil, fmt.Errorf("CSV 缺少 %s 列", key)
		}
	}
	var entries []models.InternetRankEntry
	seen := make(map[string]bool)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV 格式错误：%v", err)
		}
		if len(entries) == maxRankEntries {
			return nil, fmt.Errorf("每月最多导入 %d 个网站", maxRankEntries)
		}
		value := func(key string) string {
			if i, ok := columns[key]; ok {
				return strings.TrimSpace(record[i])
			}
			return ""
		}
		domain := rankDomain(value("domain"))
		category, name := value("category"), value("name")
		if name == "" {
			name = domain
		}
		visits, err := strconv.ParseUint(value("visits"), 10, 64)
		risk := strings.ToLower(value("risk"))
		line, _ := reader.FieldPos(0)
		if domain == "" || !rankText(category, 128) || !rankText(name, 128) || err != nil || visits == 0 ||
			(risk != "" && risk != "safe" && risk != "blocked") || (value("month") != "" && value("month") != month) {
			return nil, fmt.Errorf("第 %d 行的域名、分类、访问量、月份或风险值无效", line)
		}
		if seen[domain] {
			return nil, fmt.Errorf("第 %d 行的域名重复：%s", line, domain)
		}
		seen[domain] = true
		entries = append(entries, models.InternetRankEntry{Month: month, Domain: domain, Category: category, Name: name, Visits: visits, Risky: risk == "blocked"})
	}
	if len(entries) == 0 {
		return nil, errors.New("CSV 没有数据")
	}
	return entries, nil
}

func publicInternetRank(db *gorm.DB, month string) *gorm.DB {
	return db.Model(&models.InternetRankEntry{}).Where("month = ? AND risky = ?", month, false).
		Where(`NOT EXISTS (SELECT 1 FROM internet_rank_blocks AS b
			WHERE b.domain = internet_rank_entries.domain OR internet_rank_entries.domain LIKE CONCAT('%.', b.domain))`)
}

type internetRankView struct {
	Title      string
	Months     []models.InternetRankMonth
	Selected   *models.InternetRankMonth
	Category   string
	Categories []string
	Entries    []models.InternetRankEntry
}

func loadInternetRank(db *gorm.DB, requestedMonth, requestedCategory string) (internetRankView, error) {
	view := internetRankView{Title: "互联网月访问量排行榜"}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Order("month desc").Find(&view.Months).Error; err != nil || len(view.Months) == 0 {
			return err
		}
		view.Selected = &view.Months[0]
		for i := range view.Months {
			if view.Months[i].Month == requestedMonth {
				view.Selected = &view.Months[i]
				break
			}
		}
		base := publicInternetRank(tx, view.Selected.Month)
		if err := base.Session(&gorm.Session{}).Distinct("category").Order("category asc").Pluck("category", &view.Categories).Error; err != nil {
			return err
		}
		for _, name := range view.Categories {
			if name == requestedCategory {
				view.Category = name
				base = base.Where("category = ?", name)
				break
			}
		}
		return base.Order("visits desc, domain asc").Limit(100).Find(&view.Entries).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return view, err
}

func internetRankPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		view, err := loadInternetRank(db.WithContext(c.Request.Context()), c.Query("month"), c.Query("category"))
		if err != nil {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		meta := shell{Title: view.Title + " - 开物录", Description: "按分类查看全球网站每月访问量估算排行。", Canonical: internetRankPath}
		if len(view.Entries) == 0 {
			meta.Robots = "noindex,follow"
		}
		params := url.Values{}
		if view.Selected != nil && view.Selected.Month != view.Months[0].Month {
			params.Set("month", view.Selected.Month)
		}
		if view.Category != "" {
			params.Set("category", view.Category)
			meta.Title = view.Category + " · " + meta.Title
		}
		if len(params) > 0 {
			meta.Canonical += "?" + params.Encode()
		}
		sendSitePage(c, db, http.StatusOK, meta, "internetRank", view)
	}
}
