package api

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"opentools/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	hostName = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$`)
	iconLink = regexp.MustCompile(`(?i)<link[^>]+rel=["'][^"']*icon[^"']*["'][^>]*>`)
	hrefAttr = regexp.MustCompile(`(?i)href=["']([^"']+)["']`)
	iconOnce sync.Mutex
	iconBusy = map[string]*sync.Mutex{}
)

func mountIcon(r *gin.Engine, db *gorm.DB) {
	r.GET("/ico", func(c *gin.Context) {
		host := strings.ToLower(strings.TrimSpace(c.Query("host")))
		if !hostName.MatchString(host) {
			c.Status(http.StatusNotFound)
			return
		}
		body, kind, ok := cachedIcon(host)
		if ok && len(body) > 0 && iconUploaded(host) && r2Public != "" {
			c.Redirect(http.StatusFound, iconURL(host))
			return
		}
		if !ok {
			if !publicHost(host) || !hostInCatalog(db, host) {
				c.Status(http.StatusNotFound)
				return
			}
			body, kind, ok = fetchIcon(host)
			if ok {
				writeIconCache(host, body, kind)
			} else {
				writeIconCache(host, nil, "")
			}
		}
		if !ok || len(body) == 0 {
			c.Header("Cache-Control", "public, max-age=3600")
			c.Status(http.StatusNoContent)
			return
		}
		if err := putIcon(host, body, kind); err == nil && r2Public != "" {
			markIconUploaded(host)
			c.Redirect(http.StatusFound, iconURL(host))
			return
		}
		c.Header("Cache-Control", "public, max-age=604800")
		c.Data(http.StatusOK, kind, body)
	})
}

func hostInCatalog(db *gorm.DB, host string) bool {
	var n int64
	db.Model(&models.Link{}).
		Where("url = ? OR url = ? OR url LIKE ? OR url LIKE ?",
			"https://"+host, "http://"+host, "https://"+host+"/%", "http://"+host+"/%").
		Limit(1).Count(&n)
	return n > 0
}

func publicHost(host string) bool {
	if !hostName.MatchString(host) {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return false
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast() {
			return false
		}
	}
	return true
}

func iconLock(host string) *sync.Mutex {
	iconOnce.Lock()
	defer iconOnce.Unlock()
	if iconBusy[host] == nil {
		iconBusy[host] = &sync.Mutex{}
	}
	return iconBusy[host]
}

func iconDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "opentools-icons")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

func showIcon(host string) bool {
	if host == "" {
		return false
	}
	body, _, ok := cachedIcon(host)
	if !ok {
		return true
	}
	return len(body) > 0
}

func cachedIcon(host string) ([]byte, string, bool) {
	dir := iconDir()
	info, err := os.Stat(filepath.Join(dir, host+".bin"))
	if err != nil {
		return nil, "", false
	}
	kind, _ := os.ReadFile(filepath.Join(dir, host+".type"))
	body, err := os.ReadFile(filepath.Join(dir, host+".bin"))
	if err != nil {
		return nil, "", false
	}
	if len(body) == 0 && time.Since(info.ModTime()) > 6*time.Hour {
		return nil, "", false
	}
	if len(body) == 0 {
		return nil, "", true
	}
	return body, strings.TrimSpace(string(kind)), true
}

func writeIconCache(host string, body []byte, kind string) {
	dir := iconDir()
	_ = os.WriteFile(filepath.Join(dir, host+".bin"), body, 0o644)
	_ = os.WriteFile(filepath.Join(dir, host+".type"), []byte(kind), 0o644)
}

func iconUploaded(host string) bool {
	_, err := os.Stat(filepath.Join(iconDir(), host+".r2"))
	return err == nil
}

func markIconUploaded(host string) {
	_ = os.WriteFile(filepath.Join(iconDir(), host+".r2"), []byte("1"), 0o644)
}

func WarmIcons(db *gorm.DB) {
	if r2Client == nil {
		return
	}
	go warmIcons(db)
}

func warmIcons(db *gorm.DB) {
	var links []models.Link
	db.Select("url").Where("status = ?", "online").Find(&links)
	seen := map[string]struct{}{}
	var hosts []string
	for _, link := range links {
		parsed, err := url.Parse(link.URL)
		if err != nil {
			continue
		}
		host := strings.ToLower(parsed.Hostname())
		if !hostName.MatchString(host) {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		hosts = append(hosts, host)
	}
	log.Printf("r2 icons %d", len(hosts))
	sem := make(chan struct{}, 6)
	var wg sync.WaitGroup
	for _, host := range hosts {
		sem <- struct{}{}
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			defer func() { <-sem }()
			ensureIcon(h)
		}(host)
	}
	wg.Wait()
	log.Printf("r2 icons done")
}

func ensureIcon(host string) {
	if iconUploaded(host) {
		return
	}
	body, kind, ok := cachedIcon(host)
	if !ok || len(body) == 0 {
		if !publicHost(host) {
			return
		}
		body, kind, ok = fetchIcon(host)
	}
	if !ok || len(body) == 0 || r2Client == nil {
		return
	}
	if err := putIcon(host, body, kind); err != nil {
		return
	}
	markIconUploaded(host)
}

func fetchIcon(host string) ([]byte, string, bool) {
	lock := iconLock(host)
	lock.Lock()
	defer lock.Unlock()
	if body, kind, ok := cachedIcon(host); ok {
		return body, kind, len(body) > 0
	}
	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 4 {
				return http.ErrUseLastResponse
			}
			if !publicHost(req.URL.Hostname()) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	if href := iconFromPage(client, "https://"+host+"/"); href != "" {
		if body, kind, ok := fetchImage(client, href); ok {
			return body, kind, true
		}
	}
	return fetchImage(client, "https://"+host+"/favicon.ico")
}

func iconFromPage(client *http.Client, page string) string {
	req, err := http.NewRequest(http.MethodGet, page, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	res, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ""
	}
	buf, _ := io.ReadAll(io.LimitReader(res.Body, 80*1024))
	html := string(buf)
	best := ""
	bestScore := 0
	for _, tag := range iconLink.FindAllString(html, 12) {
		m := hrefAttr.FindStringSubmatch(tag)
		if len(m) < 2 {
			continue
		}
		score := 1
		low := strings.ToLower(tag)
		if strings.Contains(low, "apple-touch") {
			score = 3
		} else if strings.Contains(low, "192") || strings.Contains(low, "180") || strings.Contains(low, "png") {
			score = 2
		}
		if score > bestScore {
			bestScore = score
			best = m[1]
		}
	}
	if best == "" {
		return ""
	}
	ref, err := url.Parse(best)
	if err != nil {
		return ""
	}
	base, err := url.Parse(page)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(ref)
	if abs.Scheme != "https" && abs.Scheme != "http" {
		return ""
	}
	if !publicHost(abs.Hostname()) {
		return ""
	}
	return abs.String()
}

func fetchImage(client *http.Client, raw string) ([]byte, string, bool) {
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return nil, "", false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	res, err := client.Do(req)
	if err != nil {
		return nil, "", false
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, "", false
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 256*1024))
	if err != nil || len(body) < 32 || len(body) >= 256*1024 {
		return nil, "", false
	}
	kind := sniffImage(body)
	if kind == "" {
		return nil, "", false
	}
	return body, kind, true
}

func sniffImage(body []byte) string {
	if len(body) >= 8 && body[0] == 0x89 && body[1] == 'P' && body[2] == 'N' && body[3] == 'G' {
		return "image/png"
	}
	if len(body) >= 3 && body[0] == 0xff && body[1] == 0xd8 && body[2] == 0xff {
		return "image/jpeg"
	}
	if len(body) >= 6 && string(body[:3]) == "GIF" {
		return "image/gif"
	}
	if len(body) >= 12 && string(body[:4]) == "RIFF" && string(body[8:12]) == "WEBP" {
		return "image/webp"
	}
	if len(body) >= 4 && body[0] == 0 && body[1] == 0 && body[2] == 1 && body[3] == 0 {
		return "image/x-icon"
	}
	return ""
}
