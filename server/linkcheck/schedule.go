package linkcheck

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm"
)

// Start checks catalog addresses in the background. Set LINK_CHECK=0 to turn
// it off. LINK_CHECK_INTERVAL is how long a result stays fresh, default 24h.
func Start(db *gorm.DB) {
	if os.Getenv("LINK_CHECK") == "0" {
		log.Printf("网址检查已关闭")
		return
	}
	interval := 24 * time.Hour
	if raw := os.Getenv("LINK_CHECK_INTERVAL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed < time.Hour {
			log.Printf("网址检查：LINK_CHECK_INTERVAL=%s 无效，仍按 24h", raw)
		} else {
			interval = parsed
		}
	}
	log.Printf("网址检查：每 %s 复查一次上架网址", interval)
	go loop(db, interval)
}

func loop(db *gorm.DB, interval time.Duration) {
	timer := time.NewTimer(3 * time.Minute)
	defer timer.Stop()
	for {
		<-timer.C
		Run(db, interval)
		timer.Reset(time.Hour)
	}
}
