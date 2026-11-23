package main

import (
	"context"
	"encoding/json"
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
	"strings"
	"time"
)

type repoInfo struct {
	FullName        string   `json:"full_name"`
	Description     string   `json:"description"`
	HTMLURL         string   `json:"html_url"`
	Homepage        string   `json:"homepage"`
	Language        string   `json:"language"`
	StarCount       int      `json:"stargazers_count"`
	ForksCount      int      `json:"forks_count"`
	OpenIssuesCount int      `json:"open_issues_count"`
	WatchersCount   int      `json:"watchers_count"`
	Topics          []string `json:"topics"`
	License         *struct {
		Name string `json:"name"`
	} `json:"license"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	PushedAt  string `json:"pushed_at"`
}

// github-analyzer 命令行工具，用于分析 GitHub 开源项目，并生成技术分析文章
// 用法: github-analyzer -style intro -o github-analyzer.md -save
// 参数:
//
//	-proxy: HTTP 代理地址，如 http://127.0.0.1:6789
//	-style: 文章风格: deep(深度分析), intro(入门介绍), compare(对比同类)
//	-o: 输出文件路径（默认输出到终端）
//	-save: 是否保存到数据库（需要数据库连接）
//	-category: 文章分类（入库时使用）
//
// cd server && go run ./cmd/github-analyzer/main.go -proxy http://127.0.0.1:6789 -save langgenius/dify
func main() {
	proxy := flag.String("proxy", "", "HTTP 代理地址，如 http://127.0.0.1:6789")
	output := flag.String("o", "", "输出文件路径（默认输出到终端）")
	token := flag.String("token", "", "GitHub Personal Access Token（可选，提高 API 限额）")
	style := flag.String("style", "deep", "文章风格: deep(深度分析), intro(入门介绍), compare(对比同类)")
	save := flag.Bool("save", false, "保存到数据库")
	category := flag.String("category", "开源项目", "文章分类（入库时使用）")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "用法: github-analyzer [选项] <GitHub仓库地址或 owner/repo>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "示例:")
		fmt.Fprintln(os.Stderr, "  github-analyzer langgenius/dify")
		fmt.Fprintln(os.Stderr, "  github-analyzer https://github.com/langgenius/dify")
		fmt.Fprintln(os.Stderr, "  github-analyzer -proxy http://127.0.0.1:6789 langgenius/dify")
		fmt.Fprintln(os.Stderr, "  github-analyzer -style intro -o dify.md langgenius/dify")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "选项:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	repoArg := flag.Arg(0)
	ownerRepo := parseRepoArg(repoArg)
	if ownerRepo == "" {
		fmt.Fprintf(os.Stderr, "无法解析仓库地址: %s\n", repoArg)
		os.Exit(1)
	}

	if *proxy != "" {
		os.Setenv("HTTP_PROXY", *proxy)
		os.Setenv("HTTPS_PROXY", *proxy)
		fmt.Fprintf(os.Stderr, "[代理] 已设置: %s\n", *proxy)
	}

	if envToken := os.Getenv("GITHUB_TOKEN"); *token == "" && envToken != "" {
		*token = envToken
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	httpClient := buildHTTPClient(*proxy)

	fmt.Fprintf(os.Stderr, "[1/4] 正在获取仓库信息: %s ...\n", ownerRepo)
	info, err := fetchRepoInfo(httpClient, ownerRepo, *token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取仓库信息失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "[2/4] 正在获取 README ...")
	readme, err := fetchReadme(httpClient, ownerRepo, *token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[警告] README 获取失败: %v（将跳过）\n", err)
		readme = ""
	}

	fmt.Fprintln(os.Stderr, "[3/4] 正在获取最近提交记录 ...")
	commits, err := fetchRecentCommits(httpClient, ownerRepo, *token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[警告] 提交记录获取失败: %v\n", err)
		commits = ""
	}

	source := buildAnalysisContext(info, readme, commits)

	systemPrompt := getSystemPrompt(*style)
	userPrompt := fmt.Sprintf("请分析以下 GitHub 开源项目，生成一篇高质量的中文技术文章：\n\n%s", source)

	fmt.Fprintln(os.Stderr, "[4/4] 正在调用 AI 生成文章 ...")
	client := lab.FromConfig(cfg)
	article, err := client.Complete(context.Background(), systemPrompt, userPrompt, 4096)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AI 生成失败: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n=== 原始数据（降级输出） ===")
		fmt.Println(source)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "生成完成!")

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
		saveToDatabase(cfg, article, ownerRepo, *category)
	}
}

func parseRepoArg(arg string) string {
	arg = strings.TrimSpace(arg)
	arg = strings.TrimSuffix(arg, "/")
	arg = strings.TrimSuffix(arg, ".git")

	if strings.HasPrefix(arg, "https://github.com/") {
		arg = strings.TrimPrefix(arg, "https://github.com/")
	} else if strings.HasPrefix(arg, "http://github.com/") {
		arg = strings.TrimPrefix(arg, "http://github.com/")
	} else if strings.HasPrefix(arg, "github.com/") {
		arg = strings.TrimPrefix(arg, "github.com/")
	}

	parts := strings.SplitN(arg, "/", 3)
	if len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
		return parts[0] + "/" + parts[1]
	}
	return ""
}

func buildHTTPClient(proxy string) *http.Client {
	transport := &http.Transport{}
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	} else {
		transport.Proxy = http.ProxyFromEnvironment
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

func fetchRepoInfo(client *http.Client, ownerRepo, token string) (*repoInfo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s", ownerRepo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	setGitHubHeaders(req, token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		return nil, fmt.Errorf("API 返回 %d: %s", resp.StatusCode, string(body))
	}

	var info repoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}
	return &info, nil
}

func fetchReadme(client *http.Client, ownerRepo, token string) (string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/readme", ownerRepo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	setGitHubHeaders(req, token)
	req.Header.Set("Accept", "application/vnd.github.raw+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("返回 %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", err
	}

	content := string(body)
	if len(content) > 8000 {
		content = content[:8000] + "\n\n... (README 已截断)"
	}
	return content, nil
}

func fetchRecentCommits(client *http.Client, ownerRepo, token string) (string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/commits?per_page=10", ownerRepo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	setGitHubHeaders(req, token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("返回 %d", resp.StatusCode)
	}

	var commits []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name string `json:"name"`
				Date string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, c := range commits {
		msg := c.Commit.Message
		if idx := strings.Index(msg, "\n"); idx > 0 {
			msg = msg[:idx]
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", c.SHA[:7], msg, c.Commit.Author.Date))
	}
	return sb.String(), nil
}

func setGitHubHeaders(req *http.Request, token string) {
	req.Header.Set("User-Agent", "GitHub-Analyzer-CLI/1.0")
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func buildAnalysisContext(info *repoInfo, readme, commits string) string {
	var sb strings.Builder

	sb.WriteString("## 项目基本信息\n\n")
	sb.WriteString(fmt.Sprintf("- 名称: %s\n", info.FullName))
	sb.WriteString(fmt.Sprintf("- 描述: %s\n", info.Description))
	sb.WriteString(fmt.Sprintf("- 链接: %s\n", info.HTMLURL))
	if info.Homepage != "" {
		sb.WriteString(fmt.Sprintf("- 官网: %s\n", info.Homepage))
	}
	if info.Language != "" {
		sb.WriteString(fmt.Sprintf("- 主要语言: %s\n", info.Language))
	}
	sb.WriteString(fmt.Sprintf("- Star 数: %d\n", info.StarCount))
	sb.WriteString(fmt.Sprintf("- Fork 数: %d\n", info.ForksCount))
	sb.WriteString(fmt.Sprintf("- Open Issues: %d\n", info.OpenIssuesCount))
	if len(info.Topics) > 0 {
		sb.WriteString(fmt.Sprintf("- 标签: %s\n", strings.Join(info.Topics, ", ")))
	}
	if info.License != nil {
		sb.WriteString(fmt.Sprintf("- 许可证: %s\n", info.License.Name))
	}
	sb.WriteString(fmt.Sprintf("- 创建时间: %s\n", info.CreatedAt))
	sb.WriteString(fmt.Sprintf("- 最近更新: %s\n", info.UpdatedAt))

	if readme != "" {
		sb.WriteString("\n## README 内容\n\n")
		sb.WriteString(readme)
		sb.WriteString("\n")
	}

	if commits != "" {
		sb.WriteString("\n## 最近提交记录\n\n")
		sb.WriteString(commits)
	}

	return sb.String()
}

func saveToDatabase(cfg config.Config, article, ownerRepo, category string) {
	slug := "github-analysis-" + strings.ReplaceAll(ownerRepo, "/", "-")
	title := extractTitle(article)
	if title == "" {
		title = fmt.Sprintf("开源项目深度解析：%s", ownerRepo)
	}
	if err := editorial.SaveGenerated(cfg, slug, title, article, category); err != nil {
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

func getSystemPrompt(style string) string {
	base := `你是一位资深技术博主和开源项目评论家。请根据提供的 GitHub 项目信息，撰写一篇高质量的中文技术分析文章。

通用要求：
1. 使用 Markdown 格式
2. 文章开头要有一个吸引人的标题
3. 语言风格专业但通俗，适合开发者和技术爱好者阅读
4. 包含项目的 GitHub 链接
5. 不要输出任何与文章内容无关的文字
6. 如果 README 中有英文内容，请翻译并提炼关键信息，不要直接照搬`

	switch style {
	case "intro":
		return base + `

文章风格：入门介绍
额外要求：
1. 写一段引人入胜的导语，说明这个项目解决了什么问题
2. 详细介绍项目的核心功能和特性
3. 提供快速上手指南（基于 README）
4. 列出适用场景和目标用户
5. 给出对新手友好的评价和建议
6. 字数控制在 1500-2500 字`

	case "compare":
		return base + `

文章风格：对比分析
额外要求：
1. 分析这个项目在同类开源项目中的定位
2. 从功能、性能、社区活跃度、学习曲线等维度进行分析
3. 列出项目的优势和不足
4. 提及可能的替代方案（同类项目）
5. 给出选型建议：什么场景下应该选择这个项目
6. 字数控制在 2000-3000 字`

	default: // "deep"
		return base + `

文章风格：深度分析
额外要求：
1. 开头用一段话概括项目的核心价值和行业地位
2. 分析项目的技术架构和设计理念
3. 解读项目的发展趋势（基于提交记录和版本更新）
4. 评估项目的社区健康度（Star、Fork、Issue 活跃度）
5. 从技术选型角度给出推荐意见
6. 展望项目的未来发展方向
7. 字数控制在 2000-3000 字`
	}
}
