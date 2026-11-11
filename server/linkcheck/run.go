package linkcheck

import (
	"log"
	"sync"
	"time"

	"opentools/models"

	"gorm.io/gorm"
)

// Run probes online http(s) links that have not been checked within interval.
// It rewrites a dead primary to a live spare, follows a permanent new address,
// and unpublishes a link only after repeated confirmed failures.
func Run(db *gorm.DB, interval time.Duration) {
	if interval < time.Hour {
		interval = 24 * time.Hour
	}
	cutoff := time.Now().Add(-interval)
	var links []models.Link
	err := db.Where("status = ? AND (url LIKE ? OR url LIKE ?)", "online", "https://%", "http://%").
		Where("checked_at IS NULL OR checked_at < ?", cutoff).
		Find(&links).Error
	if err != nil {
		log.Printf("网址检查：没有读到网址 %v", err)
		return
	}
	var swapped, retired int
	var mu sync.Mutex
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	for _, link := range links {
		wg.Add(1)
		sem <- struct{}{}
		go func(link models.Link) {
			defer wg.Done()
			defer func() { <-sem }()
			note, didSwap, didRetire := apply(db, link)
			if note != "" {
				log.Printf("网址检查：%s %s", link.Name, note)
			}
			mu.Lock()
			if didSwap {
				swapped++
			}
			if didRetire {
				retired++
			}
			mu.Unlock()
		}(link)
	}
	wg.Wait()
	log.Printf("网址检查：本轮 %d 条，改地址 %d，下架 %d", len(links), swapped, retired)
}

func apply(db *gorm.DB, link models.Link) (note string, swapped, retired bool) {
	primary := Check(link.URL)
	var spareProbe Result
	if primary.HardDead && (hasHTTP(link.Spare)) {
		spareProbe = Check(link.Spare)
	}
	decision := Decide(link.URL, link.Spare, link.Status, link.CheckFails, primary, spareProbe)
	now := time.Now()
	err := db.Model(&models.Link{}).Where("id = ? AND status = ?", link.ID, "online").Updates(map[string]any{
		"url":         decision.URL,
		"spare":       decision.Spare,
		"status":      decision.Status,
		"check_fails": decision.Fails,
		"checked_at":  now,
	}).Error
	if err != nil {
		log.Printf("网址检查：%s 没有写上 %v", link.Name, err)
		return "", false, false
	}
	swapped = decision.URL != link.URL
	retired = decision.Status == "off" && link.Status != "off"
	if !swapped && !retired && decision.Fails <= link.CheckFails {
		return "", false, false
	}
	return decision.Note, swapped, retired
}

func hasHTTP(raw string) bool {
	raw = trimURL(raw)
	return len(raw) >= 7 && (raw[:7] == "http://" || (len(raw) >= 8 && raw[:8] == "https://"))
}
