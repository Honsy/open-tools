package editorial

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"opentools/config"
	"opentools/lab"
	"opentools/models"
	"opentools/storage"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Exists(db *gorm.DB, slug string) (bool, error) {
	var total int64
	err := db.Model(&models.Article{}).Where("legacy_slug = ?", slug).Count(&total).Error
	return total > 0, err
}

func Save(db *gorm.DB, slug, title, summary, body, category, cover string) error {
	key := sha256.Sum256([]byte(slug))
	id := fmt.Sprintf("generated:%x", key[:20])
	// The unique source ID protects against concurrent runs of the same generator.
	if len([]rune(title)) > 255 {
		title = string([]rune(title)[:255])
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.Article{
		LegacyID: &id, LegacySlug: slug, Title: title, Summary: summary, Body: body,
		Format: "markdown", Category: category, CoverURL: cover, CreatedAt: time.Now(),
	}).Error
}

func SaveGenerated(cfg config.Config, slug, title, body, category string) error {
	if cfg.PostgresDB != "opentools" {
		return errors.New("destination database must be opentools")
	}
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return errors.New("cannot connect to opentools")
	}
	if found, err := Exists(db, slug); err != nil || found {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	provider := lab.FromConfig(cfg)
	summary, err := provider.Complete(ctx, "用一句话总结这篇文章，作为中文摘要。", body, 250)
	if err != nil {
		summary = title
	}
	cover, err := generatedCover(ctx, cfg, provider, title)
	if err != nil {
		log.Printf("article cover: %v", err)
	}
	if err := Save(db, slug, title, summary, body, category, cover); err != nil {
		return err
	}
	fmt.Printf("文章已保存：%s\n", title)
	return nil
}

func generatedCover(ctx context.Context, cfg config.Config, provider lab.Client, title string) (string, error) {
	objects := storage.New(cfg)
	if objects == nil || cfg.R2PublicBase == "" {
		return "", nil
	}
	images, err := provider.Poster(ctx, "为《"+title+"》生成一张横版科技插画封面。画面不加文字。", "1280*720")
	if err != nil {
		return "", err
	}
	uri, err := url.Parse(images[0])
	if err != nil || !isProviderCover(uri) {
		return "", errors.New("封面来源不受支持")
	}
	client := &http.Client{
		Timeout: 25 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 || !isProviderCover(req.URL) {
				return errors.New("封面跳转来源不受支持")
			}
			return nil
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri.String(), nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(req)
	if err != nil {
		return "", errors.New("封面下载失败")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("封面下载返回 HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, storage.MaxArticleImage+1))
	if err != nil {
		return "", err
	}
	return storage.PutArticleImage(ctx, objects, cfg.R2Bucket, cfg.R2PublicBase, data)
}

func isProviderCover(value *url.URL) bool {
	host := strings.ToLower(value.Hostname())
	return value.Scheme == "https" && value.User == nil && (value.Port() == "" || value.Port() == "443") &&
		strings.HasSuffix(host, ".aliyuncs.com") && strings.Contains(host, ".oss-")
}
