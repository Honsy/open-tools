package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"opentools/config"
	"opentools/lab"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type labField struct {
	Name, Label, Type string
	Options           []string
	Optional          bool
}

type labTool struct {
	Slug, Name, Desc, Group string
	Fields                  []labField
}

var zodiacOptions = strings.Fields("白羊座 金牛座 双子座 巨蟹座 狮子座 处女座 天秤座 天蝎座 射手座 摩羯座 水瓶座 双鱼座")

var migratedTools = []labTool{
	{"poster", "AI 创意海报", "输入描述，生成不同尺寸的海报。", "图像", []labField{{Name: "prompt", Label: "海报描述", Type: "textarea"}, {Name: "size", Label: "画布尺寸", Options: strings.Fields("1280*720 720*1280 1280*1280 1024*1024 1440*720 720*1440")}}},
	{"bg-replace", "AI 图片换背景", "上传透明底 PNG，用文字描述新背景。", "图像", []labField{{Name: "image", Label: "透明底 PNG", Type: "file"}, {Name: "prompt", Label: "背景描述", Type: "textarea"}, {Name: "count", Label: "生成张数", Options: strings.Fields("1 2 3 4")}}},
	{"horoscope", "AI 星座运势", "按星座生成当天的娱乐运势。", "趣味", []labField{{Name: "zodiac", Label: "星座", Options: zodiacOptions}}},
	{"zodiac-match", "AI 星座匹配", "比较两个星座的互动特点。", "趣味", []labField{{Name: "zodiacA", Label: "第一个星座", Options: zodiacOptions}, {Name: "zodiacB", Label: "第二个星座", Options: zodiacOptions}}},
	{"fortune", "AI 运势签", "抽取事业、爱情、财运、学业或综合签文。", "趣味", []labField{{Name: "category", Label: "求签主题", Options: strings.Fields("综合 事业 爱情 财运 学业")}}},
	{"tarot", "AI 塔罗占卜", "生成三张牌及其含义。", "趣味", []labField{{Name: "question", Label: "想问的问题", Type: "textarea", Optional: true}}},
	{"dream", "AI 解梦", "输入梦境，获取娱乐性解读。", "趣味", []labField{{Name: "dream", Label: "梦境描述", Type: "textarea"}}},
	{"worldcup-predict", "世界杯预测", "选择两支球队，模拟对阵结果。", "趣味", []labField{{Name: "teamA", Label: "第一支球队", Type: "text"}, {Name: "teamB", Label: "第二支球队", Type: "text"}}},
	{"sbti", "SBTI 趣味测试", "根据回答查看人格类型和维度。", "趣味", nil},
	{"mortgage-calc", "房贷计算器", "计算商贷、公积金或组合贷款的还款明细。", "实用", nil},
	{"retirement-calc", "退休年龄计算器", "根据出生年月和原退休年龄估算退休日期。", "实用", nil},
}

func MountLab(router *gin.Engine, db *gorm.DB, cfg config.Config) {
	var calls labLimits
	var jobs imageJobs
	client := lab.FromConfig(cfg)
	tools := router.Group("/api/tools", func(c *gin.Context) {
		user, ok := requireReader(c, db)
		if !ok {
			c.Abort()
			return
		}
		c.Set("labUserID", user.ID)
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}
		if !sameOrigin(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "请从本站使用这个工具"})
			return
		}
		if client.Key == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "AI 服务暂未配置"})
			return
		}
		if !calls.allow(user.ID, time.Now()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "调用过于频繁，请稍后再试"})
			return
		}
		var maxBody int64 = 32 << 10
		if c.FullPath() == "/api/tools/bg-replace" {
			maxBody = 11 << 20
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBody)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 55*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	tools.POST("/horoscope", lab.AIHoroscope(client))
	tools.POST("/zodiac-match", lab.AIZodiacMatch(client))
	tools.POST("/fortune", lab.AIFortune(client))
	tools.POST("/tarot", lab.AITarot(client))
	tools.POST("/dream", lab.AIDream(client))
	tools.POST("/worldcup-predict", lab.AIWorldCupPredict(client))
	tools.POST("/poster", posterTool(client, &jobs))
	tools.POST("/bg-replace", backgroundTool(client, &jobs))
	tools.GET("/image-jobs/:id", jobs.result)
}

type labQuota struct {
	minute, day int64
	minuteCalls int
	dayCalls    int
}

type labLimits struct {
	sync.Mutex
	users map[uint]labQuota
	day   int64
}

func (l *labLimits) allow(id uint, now time.Time) bool {
	l.Lock()
	defer l.Unlock()
	minute, day := now.Unix()/60, now.Unix()/86400
	if l.day != day {
		l.users = make(map[uint]labQuota)
		l.day = day
	}
	quota := l.users[id]
	if quota.minute != minute {
		quota.minute, quota.minuteCalls = minute, 0
	}
	if quota.minuteCalls >= 5 || quota.dayCalls >= 1000 {
		return false
	}
	quota.minuteCalls++
	quota.dayCalls++
	l.users[id] = quota
	return true
}

func posterTool(client lab.Client, jobs *imageJobs) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Prompt string `json:"prompt"`
			Size   string `json:"size"`
		}
		if err := c.ShouldBindJSON(&input); err != nil || !validToolPrompt(input.Prompt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入 1 到 1200 个字的海报描述"})
			return
		}
		switch input.Size {
		case "1280*720", "720*1280", "1280*1280", "1024*1024", "1440*720", "720*1440":
		default:
			input.Size = "1280*720"
		}
		id, ok := jobs.add(c.MustGet("labUserID").(uint), func(ctx context.Context) ([]string, error) {
			return client.Poster(ctx, strings.TrimSpace(input.Prompt), input.Size)
		})
		sendImageJob(c, id, ok)
	}
}

func backgroundTool(client lab.Client, jobs *imageJobs) gin.HandlerFunc {
	return func(c *gin.Context) {
		if r2Client == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "图片上传服务暂未配置"})
			return
		}
		if err := c.Request.ParseMultipartForm(11 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请上传不超过 10 MB 的 PNG 图片"})
			return
		}
		defer c.Request.MultipartForm.RemoveAll()
		prompt := strings.TrimSpace(c.PostForm("prompt"))
		if !validToolPrompt(prompt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入 1 到 1200 个字的背景描述"})
			return
		}
		file, _, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择 PNG 图片"})
			return
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, (10<<20)+1))
		picture, pngErr := png.DecodeConfig(bytes.NewReader(data))
		if err != nil || pngErr != nil || len(data) > 10<<20 || picture.Width > 8192 || picture.Height > 8192 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请上传不超过 10 MB 的有效 PNG 图片"})
			return
		}
		count, _ := strconv.Atoi(c.PostForm("count"))
		if count < 1 || count > 4 {
			count = 1
		}
		id, ok := jobs.add(c.MustGet("labUserID").(uint), func(ctx context.Context) ([]string, error) {
			key := "lab/background/" + rand.Text() + ".png"
			_, err := r2Client.PutObject(ctx, &s3.PutObjectInput{
				Bucket: aws.String(r2Bucket), Key: aws.String(key), Body: bytes.NewReader(data), ContentType: aws.String("image/png"),
			})
			if err != nil {
				return nil, err
			}
			defer func() {
				cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				r2Client.DeleteObject(cleanup, &s3.DeleteObjectInput{Bucket: aws.String(r2Bucket), Key: aws.String(key)})
			}()
			signed, err := s3.NewPresignClient(r2Client).PresignGetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(r2Bucket), Key: aws.String(key),
			}, s3.WithPresignExpires(10*time.Minute))
			if err != nil {
				return nil, err
			}
			return client.Background(ctx, signed.URL, prompt, count)
		})
		sendImageJob(c, id, ok)
	}
}

func validToolPrompt(value string) bool {
	length := len([]rune(strings.TrimSpace(value)))
	return length > 0 && length <= 1200
}

func sendImageJob(c *gin.Context, id string, ok bool) {
	if !ok {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "图像工具繁忙，请稍后再试"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"job": id})
}

func toolIndex(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		render(c, db, http.StatusOK, shell{
			Title: "在线工具 - 开物录", Description: "AI 图像工具、趣味测试和实用计算器。",
			Canonical: "/tools/", Robots: "noindex,follow",
		}, "labIndex", migratedTools)
	}
}

func showLabTool(c *gin.Context, db *gorm.DB, tool labTool) {
	_, signedIn := currentUser(c, db)
	render(c, db, http.StatusOK, shell{
		Title: tool.Name + " - 开物录", Description: tool.Desc,
		Canonical: "/tools/" + tool.Slug, Robots: "noindex,follow",
	}, "labTool", struct {
		labTool
		SignedIn bool
	}{tool, signedIn})
}
