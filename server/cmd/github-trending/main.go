package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"opentools/config"
	"opentools/editorial"
	"opentools/lab"
	"os"
	"regexp"
	"strings"
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

// github-trending 命令行工具，用于抓取 GitHub Trending 热门开源项目，并生成技术资讯文章
// 用法: github-trending -lang go -since daily -o github-trending.md -save
// 参数:
//
//	-proxy: HTTP 代理地址，如 http://127.0.0.1:6789
//	-lang: 按编程语言筛选，如 go, python, rust
//	-since: 时间范围: daily, weekly, monthly
//	-o: 输出文件路径（默认输出到终端）
//	-save: 是否保存到数据库（需要数据库连接）
//
// cd server && go run ./cmd/github-trending/main.go -proxy http://127.0.0.1:6789 -save
func main() {
	proxy := flag.String("proxy", "", "HTTP 代理地址，如 http://127.0.0.1:6789")
	lang := flag.String("lang", "", "按编程语言筛选，如 go, python, rust")
	since := flag.String("since", "daily", "时间范围: daily, weekly, monthly")
	output := flag.String("o", "", "输出文件路径（默认输出到终端）")
	save := flag.Bool("save", false, "是否保存到数据库（需要数据库连接）")
	flag.Parse()

	if *proxy != "" {
		os.Setenv("HTTP_PROXY", *proxy)
		os.Setenv("HTTPS_PROXY", *proxy)
		fmt.Fprintf(os.Stderr, "[代理] 已设置: %s\n", *proxy)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stderr, "[1/3] 正在爬取 GitHub Trending ...")
	repos, err := fetchTrending(*lang, *since, *proxy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "爬取失败: %v\n", err)
		os.Exit(1)
	}
	if len(repos) == 0 {
		fmt.Fprintln(os.Stderr, "未找到任何 trending 项目")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "[1/3] 已抓取 %d 个项目\n", len(repos))

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

	today := time.Now().Format("2006-01-02")
	systemPrompt := `你是一位资深技术编辑，擅长将 GitHub 开源项目趋势整理成高质量的中文技术资讯文章。
请按照以下要求输出：
1. 文章标题要醒目，包含日期
2. 开头写一段简短的导语概述今日趋势亮点
3. 将项目按类别分组（如 AI/ML、开发工具、Web框架、基础设施等），每个类别作为一个章节
4. 每个项目简要介绍：名称（带 GitHub 链接）、功能亮点、适用场景、今日热度
5. 末尾写一段总结，提炼今天的技术趋势关键词
6. 使用 Markdown 格式输出
7. 语言风格专业但通俗易懂，适合开发者阅读
8. 不要输出任何与文章内容无关的文字`

	userPrompt := fmt.Sprintf("以下是 %s GitHub Trending 热门开源项目列表，请整理成一篇技术资讯文章：\n\n%s", today, sb.String())

	fmt.Fprintln(os.Stderr, "[2/3] 正在调用 AI 生成日报 ...")
	client := lab.FromConfig(cfg)
	article, err := client.Complete(context.Background(), systemPrompt, userPrompt, 4096)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AI 生成失败: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n=== 原始数据（降级输出） ===")
		fmt.Println(sb.String())
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "[3/3] 生成完成!")

	if *output != "" {
		if err := os.WriteFile(*output, []byte(article), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "写入文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "已保存到: %s\n", *output)
	} else if !*save {
		fmt.Println(article)
	}

	if *save {
		saveToDatabase(cfg, article, today)
	}
}

func saveToDatabase(cfg config.Config, article, today string) {
	title := extractTitle(article)
	if title == "" {
		title = "GitHub Trending 每日精选 | " + today
	}
	if err := editorial.SaveGenerated(cfg, "github-trending-"+today, title, article, "资讯"); err != nil {
		log.Fatal(err)
	}
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

func fetchTrending(lang, since, proxy string) ([]trendingRepo, error) {
	trendingURL := "https://github.com/trending"
	if lang != "" {
		trendingURL += "/" + url.PathEscape(lang)
	}
	params := url.Values{}
	if since != "" && since != "daily" {
		params.Set("since", since)
	}
	if len(params) > 0 {
		trendingURL += "?" + params.Encode()
	}

	transport := &http.Transport{}
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	} else {
		transport.Proxy = http.ProxyFromEnvironment
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	req, err := http.NewRequest("GET", trendingURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub 返回 %d", resp.StatusCode)
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
