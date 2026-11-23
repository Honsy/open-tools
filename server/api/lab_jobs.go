package api

import (
	"context"
	"crypto/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type imageJob struct {
	user   uint
	until  time.Time
	done   bool
	images []string
	failed bool
}

type imageJobs struct {
	sync.Mutex
	items map[string]imageJob
}

func (j *imageJobs) add(user uint, work func(context.Context) ([]string, error)) (string, bool) {
	j.Lock()
	if j.items == nil {
		j.items = make(map[string]imageJob)
	}
	now := time.Now()
	running, personal := 0, 0
	for id, item := range j.items {
		if now.After(item.until) {
			delete(j.items, id)
		} else if !item.done {
			running++
			if item.user == user {
				personal++
			}
		}
	}
	if running >= 8 || personal >= 2 {
		j.Unlock()
		return "", false
	}
	id := rand.Text()
	j.items[id] = imageJob{user: user, until: now.Add(30 * time.Minute)}
	j.Unlock()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		defer cancel()
		images, err := work(ctx)
		j.Lock()
		j.items[id] = imageJob{user: user, until: time.Now().Add(30 * time.Minute), done: true, images: images, failed: err != nil}
		j.Unlock()
	}()
	return id, true
}

func (j *imageJobs) result(c *gin.Context) {
	j.Lock()
	job, ok := j.items[c.Param("id")]
	j.Unlock()
	c.Header("Cache-Control", "no-store")
	if !ok || job.user != c.MustGet("labUserID").(uint) || time.Now().After(job.until) {
		c.JSON(http.StatusNotFound, gin.H{"error": "没有这个图片任务"})
		return
	}
	if !job.done {
		c.JSON(http.StatusAccepted, gin.H{"pending": true})
		return
	}
	if job.failed {
		c.JSON(http.StatusBadGateway, gin.H{"error": "暂时无法生成图片，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"images": job.images})
}
