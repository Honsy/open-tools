package editorial

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"opentools/lab"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type trendingRepo struct {
	Name        string
	URL         string
	Description string
	Language    string
	StarsToday  string
	TotalStars  string
}

// StartDaily publishes the migrated digests at 08:00 China time when enabled.
func StartDaily(db *gorm.DB, client lab.Client) {
	if client.Key == "" {
		log.Print("daily articles skipped: AI service is not configured")
		return
	}
	location := time.FixedZone("Asia/Shanghai", 8*3600)
	go func() {
		for {
			now := time.Now().In(location)
			next := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, location)
			if !next.After(now) {
				next = next.AddDate(0, 0, 1)
			}
			<-time.After(time.Until(next))
			for _, generate := range []func(context.Context, *gorm.DB, lab.Client, string) error{fetchAndGenerateArticle, fetchAndGenerateAINews} {
				ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
				if err := generate(ctx, db, client, next.Format("2006-01-02")); err != nil {
					log.Printf("daily article: %v", err)
				}
				cancel()
			}
		}
	}()
}

func fetchAndGenerateArticle(ctx context.Context, db *gorm.DB, client lab.Client, today string) error {
	repos, err := fetchTrendingRepos()
	if err != nil {
		return fmt.Errorf("抓取 trending 失败: %w", err)
	}
	if len(repos) == 0 {
		return fmt.Errorf("未抓取到任何 trending 仓库")
	}

	slug := "github-trending-" + today

	if exists, err := Exists(db, slug); err != nil || exists {
		return err
	}

	var sb strings.Builder
	for i, repo := range repos {
		sb.WriteString(fmt.Sprintf("%d. **%s** (%s)\n", i+1, repo.Name, repo.URL))
		if repo.Description != "" {
			sb.WriteString(fmt.Sprintf("   描述: %s\n", repo.Description))
		}
		if repo.Language != "" {
			sb.WriteString(fmt.Sprintf("   语言: %s\n", repo.Language))
		}
		if repo.TotalStars != "" {
			sb.WriteString(fmt.Sprintf("   总星数: %s\n", repo.TotalStars))
		}
		if repo.StarsToday != "" {
			sb.WriteString(fmt.Sprintf("   今日新增: %s\n", repo.StarsToday))
		}
		sb.WriteString("\n")
	}

	systemPrompt := `你是一位资深技术编辑，擅长将 GitHub 开源项目趋势整理成高质量的中文技术资讯文章。
请按照以下要求输出：
1. 文章标题要醒目，包含日期
2. 开头写一段简短的导语概述今日趋势亮点
3. 将项目按类别分组（如 AI/ML、开发工具、Web框架、基础设施 等），每个类别作为一个章节
4. 每个项目简要介绍：名称（带GitHub链接）、功能亮点、适用场景、今日热度
5. 末尾写一段总结，提炼今天的技术趋势关键词
6. 使用 Markdown 格式输出
7. 语言风格专业但通俗易懂，适合开发者阅读
8. 不要输出任何与文章内容无关的文字`

	userPrompt := fmt.Sprintf("以下是 %s GitHub Trending 热门开源项目列表，请整理成一篇技术资讯文章：\n\n%s", today, sb.String())

	content, err := client.Complete(ctx, systemPrompt, userPrompt, 4096)
	if err != nil {
		return fmt.Errorf("AI 生成文章失败: %w", err)
	}

	titleLine := extractTitle(content)
	if titleLine == "" {
		titleLine = fmt.Sprintf("GitHub Trending 每日精选 | %s", today)
	}

	summaryPrompt := `请用一句话（不超过100字）总结以下文章的核心内容，作为文章摘要：`
	summary, err := client.Complete(ctx, summaryPrompt, content, 250)
	if err != nil {
		summary = fmt.Sprintf("每日精选 GitHub 热门开源项目趋势分析 - %s", today)
	}

	return Save(db, slug, titleLine, summary, content, "资讯", "")
}

func fetchTrendingRepos() ([]trendingRepo, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", "https://github.com/trending", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub Trending 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub Trending 返回 %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return parseTrendingHTML(string(body)), nil
}

func parseTrendingHTML(html string) []trendingRepo {
	reArticle := regexp.MustCompile(`(?s)<article class="Box-row">.*?</article>`)
	articles := reArticle.FindAllString(html, -1)

	var repos []trendingRepo
	reLink := regexp.MustCompile(`(?s)<h2[^>]*>.*?<a[^>]*href="(/[^"]+)"[^>]*>`)
	reDesc := regexp.MustCompile(`(?s)<p class="col-9[^"]*"[^>]*>(.*?)</p>`)
	reLang := regexp.MustCompile(`<span itemprop="programmingLanguage">(.*?)</span>`)
	reStarsTotal := regexp.MustCompile(`(?s)<a[^>]*href="/[^/]+/[^/]+/stargazers"[^>]*>\s*([\d,]+)\s*</a>`)
	reStarsToday := regexp.MustCompile(`([\d,]+)\s+stars today`)

	for _, art := range articles {
		repo := trendingRepo{}

		if m := reLink.FindStringSubmatch(art); len(m) > 1 {
			path := strings.TrimSpace(m[1])
			repo.Name = strings.TrimPrefix(path, "/")
			repo.URL = "https://github.com" + path
		}

		if m := reDesc.FindStringSubmatch(art); len(m) > 1 {
			repo.Description = strings.TrimSpace(m[1])
		}

		if m := reLang.FindStringSubmatch(art); len(m) > 1 {
			repo.Language = strings.TrimSpace(m[1])
		}

		if m := reStarsTotal.FindStringSubmatch(art); len(m) > 1 {
			repo.TotalStars = strings.TrimSpace(m[1])
		}

		if m := reStarsToday.FindStringSubmatch(art); len(m) > 1 {
			repo.StarsToday = strings.TrimSpace(m[1])
		}

		if repo.Name != "" {
			repos = append(repos, repo)
		}
	}

	return repos
}

// ==================== AI 每日新闻热点 ====================

type hnStory struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Score       int    `json:"score"`
	Descendants int    `json:"descendants"` // 评论数
	By          string `json:"by"`
	Time        int64  `json:"time"`
}

var aiKeywords = []string{
	"ai", "artificial intelligence", "machine learning", "deep learning",
	"llm", "gpt", "claude", "gemini", "openai", "anthropic", "mistral",
	"neural", "transformer", "diffusion", "stable diffusion", "midjourney",
	"copilot", "chatbot", "agent", "rag", "fine-tun", "embedding",
	"langchain", "llamaindex", "hugging face", "ollama", "llama",
	"deepseek", "qwen", "generative", "nlp", "computer vision",
	"reinforcement learning", "model", "inference", "training",
	"multimodal", "text-to-image", "text-to-video", "speech",
	"whisper", "sora", "mcp", "ai coding", "vibe coding",
}

func fetchAndGenerateAINews(ctx context.Context, db *gorm.DB, client lab.Client, today string) error {
	stories, err := fetchHNAIStories()
	if err != nil {
		return fmt.Errorf("抓取 HN AI 新闻失败: %w", err)
	}
	if len(stories) == 0 {
		return fmt.Errorf("未抓取到任何 AI 相关新闻")
	}

	slug := "ai-daily-news-" + today

	if exists, err := Exists(db, slug); err != nil || exists {
		return err
	}

	var sb strings.Builder
	for i, story := range stories {
		sb.WriteString(fmt.Sprintf("%d. 标题: %s\n", i+1, story.Title))
		if story.URL != "" {
			sb.WriteString(fmt.Sprintf("   链接: %s\n", story.URL))
		}
		sb.WriteString(fmt.Sprintf("   热度: %d 分 | 评论: %d 条\n", story.Score, story.Descendants))
		hnURL := fmt.Sprintf("https://news.ycombinator.com/item?id=%d", story.ID)
		sb.WriteString(fmt.Sprintf("   讨论: %s\n", hnURL))
		sb.WriteString("\n")
	}

	systemPrompt := `你是一位专注于人工智能领域的资深科技编辑。请根据提供的新闻列表，精选最重要的5条新闻，整理成一篇高质量的中文 AI 日报文章。

要求：
1. 文章标题要醒目，包含日期，例如"AI 日报 | 2025-01-01：xxx"
2. 开头用2-3句话概括今日 AI 领域最值得关注的动态
3. 精选 5 条最重要/最有影响力的新闻，每条包含：
   - 醒目的小标题
   - 100-200字的深度解读（不是简单翻译标题，要分析其意义和影响）
   - 原文链接
4. 末尾用"今日观察"做一段简短总结，点评 AI 行业趋势
5. 使用 Markdown 格式
6. 语言风格：专业、有深度、有观点，适合 AI 从业者和关注者阅读
7. 不要输出任何与文章内容无关的文字
8. 对英文标题和内容进行意译，不要生硬直译`

	userPrompt := fmt.Sprintf("以下是 %s 从 Hacker News 收集的 AI 相关热门新闻，请精选 5 条最重要的整理成 AI 日报：\n\n%s", today, sb.String())

	content, err := client.Complete(ctx, systemPrompt, userPrompt, 4096)
	if err != nil {
		return fmt.Errorf("AI 生成新闻文章失败: %w", err)
	}

	titleLine := extractTitle(content)
	if titleLine == "" {
		titleLine = fmt.Sprintf("AI 日报 | %s", today)
	}

	summaryPrompt := `请用一句话（不超过100字）总结以下 AI 日报的核心看点，作为文章摘要：`
	summary, err := client.Complete(ctx, summaryPrompt, content, 250)
	if err != nil {
		summary = fmt.Sprintf("每日 AI 领域重要新闻与深度解读 - %s", today)
	}

	return Save(db, slug, titleLine, summary, content, "AI日报", "")
}

func fetchHNAIStories() ([]hnStory, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if err != nil {
		return nil, fmt.Errorf("获取 HN top stories 失败: %w", err)
	}
	defer resp.Body.Close()

	var storyIDs []int
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&storyIDs); err != nil {
		return nil, fmt.Errorf("解析 HN top stories 失败: %w", err)
	}

	// 取前 200 条进行筛选
	limit := 200
	if len(storyIDs) < limit {
		limit = len(storyIDs)
	}
	storyIDs = storyIDs[:limit]

	// 并发获取 story 详情
	var mu sync.Mutex
	var wg sync.WaitGroup
	var allStories []hnStory

	sem := make(chan struct{}, 20) // 并发限制

	for _, id := range storyIDs {
		wg.Add(1)
		go func(sid int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			story, err := fetchHNStory(client, sid)
			if err != nil {
				return
			}
			if !isAIRelated(story.Title) {
				return
			}

			mu.Lock()
			allStories = append(allStories, story)
			mu.Unlock()
		}(id)
	}
	wg.Wait()

	// 按热度排序
	sort.Slice(allStories, func(i, j int) bool {
		return allStories[i].Score > allStories[j].Score
	})

	// 最多取前 15 条给 AI 挑选
	if len(allStories) > 15 {
		allStories = allStories[:15]
	}

	return allStories, nil
}

func fetchHNStory(client *http.Client, id int) (hnStory, error) {
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
	resp, err := client.Get(url)
	if err != nil {
		return hnStory{}, err
	}
	defer resp.Body.Close()

	var story hnStory
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&story); err != nil {
		return hnStory{}, err
	}
	return story, nil
}

func isAIRelated(title string) bool {
	lower := strings.ToLower(title)
	for _, kw := range aiKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "```") {
			if len([]rune(line)) > 100 {
				line = string([]rune(line)[:100])
			}
			return line
		}
	}
	return ""
}
