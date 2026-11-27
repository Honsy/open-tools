package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"opentools/models"
	"opentools/seed"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var agentIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)
var agentUUIDPattern = regexp.MustCompile(`^[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}$`)
var errDiscoveryConflict = errors.New("catalog entry needs review")

type discoveryFinding struct {
	ID          string                `json:"id"`
	Status      string                `json:"status"`
	Body        discoveryPayload      `json:"payload"`
	Before      *models.Link          `json:"before,omitempty"`
	Publication *discoveryPublication `json:"publication,omitempty"`
}

type discoveryPublication struct {
	Summary      string `json:"summary"`
	Body         string `json:"body"`
	PreviousBody string `json:"previousBody"`
	PreviousDesc string `json:"previousDesc"`
}

type discoveryPayload struct {
	Kind         string   `json:"kind"`
	Name         string   `json:"name"`
	URL          string   `json:"official_url"`
	Evidence     []string `json:"evidence_urls"`
	PublishedOn  string   `json:"published_on"`
	CategorySlug string   `json:"category_slug"`
	Summary      string   `json:"summary"`
	Body         string   `json:"body"`
	ExistingSlug string   `json:"existing_slug"`
}

type discoveryClient struct {
	origin, key, agentID string
	client               *http.Client
	review               sync.Mutex
}

func newDiscoveryClient() *discoveryClient {
	origin := strings.TrimRight(strings.TrimSpace(os.Getenv("KAIGONG_ORIGIN")), "/")
	key := strings.TrimSpace(os.Getenv("KAIGONG_PROJECT_KEY"))
	agentID := strings.TrimSpace(os.Getenv("KAIGONG_AGENT_ID"))
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" ||
		key == "" || !agentIDPattern.MatchString(agentID) {
		return nil
	}
	loopback := u.Hostname() == "localhost" || (net.ParseIP(u.Hostname()) != nil && net.ParseIP(u.Hostname()).IsLoopback())
	if u.Scheme != "https" && !(u.Scheme == "http" && loopback) {
		return nil
	}
	return &discoveryClient{
		origin: origin, key: key, agentID: agentID,
		client: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}
}

func (client *discoveryClient) request(ctx context.Context, method, path string, data any, retryKey string) (*http.Response, error) {
	var body io.Reader
	if data != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, client.origin+"/api/v1"+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+client.key)
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if retryKey != "" {
		req.Header.Set("Idempotency-Key", retryKey)
	}
	return client.client.Do(req)
}

func (client *discoveryClient) json(ctx context.Context, method, path string, data any, retryKey string, result any) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	res, err := client.request(ctx, method, path, data, retryKey)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Kaigong returned HTTP %d", res.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(result)
}

func mountDiscovery(g *gin.RouterGroup, db *gorm.DB) {
	group := g.Group("/agent-discovery")
	client := newDiscoveryClient()
	group.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"configured": client != nil})
	})
	group.Use(func(c *gin.Context) {
		if client == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "尚未配置开工 Agent 服务。"})
		}
	})
	group.GET("/findings", func(c *gin.Context) {
		status := c.DefaultQuery("status", "pending")
		if status != "pending" && status != "ignored" && status != "published" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "审核状态无效"})
			return
		}
		page := min(positivePage(c.DefaultQuery("page", "1")), 3999)
		values := url.Values{"status": {status}, "limit": {"26"}, "offset": {fmt.Sprint((page - 1) * 25)}}
		items := make([]discoveryFinding, 0)
		if err := client.json(c.Request.Context(), http.MethodGet, "/findings?"+values.Encode(), nil, "", &items); err != nil {
			discoveryUnavailable(c)
			return
		}
		more := len(items) > 25
		items = items[:min(25, len(items))]
		for i := range items {
			if items[i].Body.Kind != "update" {
				continue
			}
			var previous models.Link
			if db.WithContext(c.Request.Context()).Where("slug = ? AND status = ?", items[i].Body.ExistingSlug, "online").First(&previous).Error == nil {
				items[i].Before = &previous
			}
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "more": more})
	})
	group.POST("/runs", func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if !agentUUIDPattern.MatchString(key) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务编号"})
			return
		}
		task := "发现最近 7 天发布的 AIGC 新产品或已收录产品的重要更新，覆盖图像与设计、视频与短剧、音乐与语音。先查官方出处、产品入口和开物录已有条目，再生成待审候选。"
		var run struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		if err := client.json(c.Request.Context(), http.MethodPost, "/agents/"+client.agentID+"/runs", gin.H{"input": task}, key, &run); err != nil {
			discoveryUnavailable(c)
			return
		}
		c.JSON(http.StatusAccepted, run)
	})
	group.GET("/runs/:id/events", func(c *gin.Context) {
		id := c.Param("id")
		if !agentUUIDPattern.MatchString(id) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务编号"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Minute)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.origin+"/api/v1/runs/"+id+"/events", nil)
		if err != nil {
			discoveryUnavailable(c)
			return
		}
		req.Header.Set("Authorization", "Bearer "+client.key)
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Last-Event-ID", c.GetHeader("Last-Event-ID"))
		if req.Header.Get("Last-Event-ID") == "" {
			req.Header.Del("Last-Event-ID")
		}
		res, err := client.client.Do(req)
		if err != nil {
			discoveryUnavailable(c)
			return
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK || !strings.HasPrefix(res.Header.Get("Content-Type"), "text/event-stream") {
			discoveryUnavailable(c)
			return
		}
		c.Header("Content-Type", "text/event-stream; charset=utf-8")
		c.Header("Cache-Control", "no-store, no-transform")
		c.Header("X-Accel-Buffering", "no")
		c.Status(http.StatusOK)
		buf := make([]byte, 8192)
		for {
			n, readErr := res.Body.Read(buf)
			if n > 0 {
				if _, err := c.Writer.Write(buf[:n]); err != nil {
					return
				}
				c.Writer.Flush()
			}
			if readErr != nil {
				return
			}
		}
	})
	group.POST("/findings/:id/review", reviewDiscovery(client, db))
}

func discoveryUnavailable(c *gin.Context) {
	c.JSON(http.StatusBadGateway, gin.H{"error": "Agent 服务没有返回可用结果，请检查项目密钥、Agent 授权或服务状态。"})
}

func reviewDiscovery(client *discoveryClient, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var request struct {
			Action string `json:"action"`
			discoveryPublication
		}
		if !agentUUIDPattern.MatchString(id) || c.ShouldBindJSON(&request) != nil ||
			(request.Action != "published" && request.Action != "ignored") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "审核请求无效"})
			return
		}
		client.review.Lock()
		defer client.review.Unlock()
		path := "/findings/" + id
		ctx := c.Request.Context()
		var finding discoveryFinding
		if err := client.json(ctx, http.MethodGet, path, nil, "", &finding); err != nil {
			discoveryUnavailable(c)
			return
		}
		if finding.Status == request.Action {
			c.JSON(http.StatusOK, finding)
			return
		}
		if finding.ID != id || (finding.Status != "pending" && !(finding.Status == "publishing" && request.Action == "published")) {
			c.JSON(http.StatusConflict, gin.H{"error": "这条候选已审核或正在发布"})
			return
		}
		if request.Action == "published" {
			if finding.Status == "pending" {
				draft := request.discoveryPublication
				draft.Summary, draft.Body = strings.TrimSpace(draft.Summary), strings.TrimSpace(draft.Body)
				if utf8.RuneCountInString(draft.Summary) < 12 || utf8.RuneCountInString(draft.Summary) > 220 ||
					utf8.RuneCountInString(draft.Body) < 80 || utf8.RuneCountInString(draft.Body) > 3000 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "简介须为 12 至 220 字，详情须为 80 至 3000 字。"})
					return
				}
				if err := client.json(ctx, http.MethodPost, path+"/publication", draft, "", &finding); err != nil {
					discoveryUnavailable(c)
					return
				}
			}
			if finding.Publication == nil {
				discoveryUnavailable(c)
				return
			}
			finding.Body.Summary = finding.Publication.Summary
			finding.Body.Body = finding.Publication.Body
			err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				return applyDiscovery(tx, finding, finding.Publication.PreviousDesc, finding.Publication.PreviousBody)
			})
			if errors.Is(err, errDiscoveryConflict) {
				if clearErr := client.json(ctx, http.MethodPut, path+"/review", gin.H{"status": "pending"}, "", &finding); clearErr != nil {
					discoveryUnavailable(c)
					return
				}
				c.JSON(http.StatusConflict, gin.H{"error": "网址、分类或现有介绍有变化，请刷新候选再核对。"})
				return
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法写入开物录目录。"})
				return
			}
		}
		if err := client.json(ctx, http.MethodPut, path+"/review", gin.H{"status": request.Action}, "", &finding); err != nil {
			discoveryUnavailable(c)
			return
		}
		c.JSON(http.StatusOK, finding)
	}
}

func applyDiscovery(db *gorm.DB, finding discoveryFinding, previousDesc, previousBody string) error {
	p := finding.Body
	u, err := url.Parse(p.URL)
	if err != nil || u.Scheme != "https" || u.User != nil || !strings.Contains(u.Hostname(), ".") ||
		net.ParseIP(u.Hostname()) != nil || utf8.RuneCountInString(p.Summary) < 12 ||
		utf8.RuneCountInString(p.Summary) > 220 || utf8.RuneCountInString(p.Body) < 80 ||
		utf8.RuneCountInString(p.Body) > 3000 || utf8.RuneCountInString(p.Name) > 100 ||
		strings.TrimSpace(p.Name) == "" {
		return errDiscoveryConflict
	}
	body := p.Body
	if len(p.Evidence) == 0 || len(p.Evidence) > 5 {
		return errDiscoveryConflict
	}
	body += "\n\n### 官方出处\n"
	for _, address := range p.Evidence {
		evidence, err := url.Parse(address)
		if err != nil || evidence.Scheme != "https" || evidence.User != nil || evidence.Hostname() == "" ||
			strings.ContainsAny(address, "<>\r\n") {
			return errDiscoveryConflict
		}
		body += "\n- <" + address + ">"
	}
	var category models.Category
	err = db.Where("slug = ? AND parent_slug = ? AND kind = ? AND content_kind = ?", p.CategorySlug, "ai", "tab", "links").
		First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errDiscoveryConflict
	}
	if err != nil {
		return err
	}
	if p.Kind == "update" {
		var item models.Link
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("slug = ? AND status = ? AND category_slug = ?", p.ExistingSlug, "online", p.CategorySlug).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errDiscoveryConflict
		}
		if err != nil {
			return err
		}
		if item.Desc == p.Summary && item.Body == body {
			return nil
		}
		if item.Desc != previousDesc || item.Body != previousBody {
			return errDiscoveryConflict
		}
		return db.Model(&item).Updates(map[string]any{"desc": p.Summary, "body": body}).Error
	}
	if p.Kind != "new" {
		return errDiscoveryConflict
	}
	key := "kaigong:v1:" + finding.ID
	var already int64
	if err := db.Model(&models.Link{}).Where("imported_from = ?", key).Count(&already).Error; err != nil || already > 0 {
		return err
	}
	address := strings.TrimRight(p.URL, "/")
	var duplicates int64
	err = db.Model(&models.Link{}).Where("LOWER(name) = LOWER(?) OR url IN ?", p.Name, []string{address, address + "/"}).Count(&duplicates).Error
	if err != nil {
		return err
	}
	if duplicates > 0 {
		return errDiscoveryConflict
	}
	link := models.Link{
		Name: p.Name, URL: p.URL, CategorySlug: category.Slug, Desc: p.Summary, Body: body,
		Tags: "AIGC", Status: "online", ImportedFrom: &key,
	}
	seed.FillSlug(db, &link)
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "imported_from"}}, DoNothing: true}).Create(&link).Error
}
