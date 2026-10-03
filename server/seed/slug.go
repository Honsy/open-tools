package seed

import (
	"net/url"
	"strings"

	"opentools/models"

	"gorm.io/gorm"
)

// SlugBase makes a stable path segment from a site address.
// Existing slugs are never rewritten, so the public URL stays put.
func SlugBase(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "/") {
		parts := strings.Split(strings.Trim(raw, "/"), "/")
		return cleanSlug(parts[len(parts)-1])
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" {
		return "site"
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
	labels := strings.Split(host, ".")
	if len(labels) >= 2 {
		labels = labels[:len(labels)-1]
	}
	return cleanSlug(strings.Join(labels, "-"))
}

func cleanSlug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "site"
	}
	if len(out) > 60 {
		out = strings.Trim(out[:60], "-")
	}
	return out
}

func UniqueSlug(db *gorm.DB, base string, selfID uint) string {
	if base == "" {
		base = "site"
	}
	candidate := base
	for i := 2; i < 50; i++ {
		var n int64
		q := db.Model(&models.Link{}).Where("slug = ?", candidate)
		if selfID != 0 {
			q = q.Where("id <> ?", selfID)
		}
		q.Count(&n)
		if n == 0 {
			return candidate
		}
		candidate = base + "-" + itoa(uint(i))
	}
	return base + "-" + itoa(selfID)
}

func FillSlug(db *gorm.DB, link *models.Link) {
	if link.Slug != "" {
		return
	}
	link.Slug = UniqueSlug(db, SlugBase(link.URL), link.ID)
}

func EnsureSlugs(db *gorm.DB) error {
	var links []models.Link
	if err := db.Where("slug = ? OR slug IS NULL", "").Find(&links).Error; err != nil {
		return err
	}
	for i := range links {
		FillSlug(db, &links[i])
		if err := db.Model(&links[i]).Update("slug", links[i].Slug).Error; err != nil {
			return err
		}
	}
	return nil
}

func itoa(n uint) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
