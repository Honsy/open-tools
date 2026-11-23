package lab

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var validFortuneCategories = map[string]bool{
	"事业": true, "爱情": true, "财运": true, "学业": true, "综合": true,
}

type FortuneRequest struct {
	Category string `json:"category"`
}

type FortuneResponse struct {
	Category    string `json:"category"`
	Verse       string `json:"verse"`
	Interpret   string `json:"interpret"`
	Score       int    `json:"score"`
	Advice      string `json:"advice"`
	LuckyColor  string `json:"luckyColor"`
	LuckyNumber int    `json:"luckyNumber"`
	LuckyDir    string `json:"luckyDir"`
}

func AIFortune(client Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req FortuneRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, "参数错误")
			return
		}
		category := strings.TrimSpace(req.Category)
		if category == "" {
			category = "综合"
		}
		if !validFortuneCategories[category] {
			badRequest(c, "无效的分类，可选：事业、爱情、财运、学业、综合")
			return
		}
		systemPrompt := `你是精通中国传统命理文化的解签大师。用户会选择求签类型，请为其生成一首运势签。
要求：
1. **签文**：一首原创七言或五言绝句，古风韵味，寓意丰富
2. **解签**：50-80 字逐句解读签文含义，贴合所选分类
3. **运势评分**：1-100 的整数
4. **建议**：20-30 字针对该分类的建议
5. **幸运色**：一个中文颜色名
6. **幸运数字**：1-99 的整数
7. **幸运方位**：如"东南"、"西北"等
签文需有古风韵味，避免现代词汇。请严格按以下 JSON 格式输出，不要包含任何其他文字：
{"verse":"签文诗句","interpret":"解签内容","score":75,"advice":"建议","luckyColor":"青色","luckyNumber":7,"luckyDir":"东南"}`
		userPrompt := "求签类型：" + category
		result, err := client.chat(c.Request.Context(), systemPrompt, userPrompt)
		if err != nil {
			modelError(c, "求签失败，请稍后重试")
			return
		}
		result = strings.TrimSpace(result)
		result = strings.Trim(result, "`")
		if idx := strings.Index(result, "{"); idx >= 0 {
			result = result[idx:]
		}
		if idx := strings.LastIndex(result, "}"); idx >= 0 {
			result = result[:idx+1]
		}
		var parsed struct {
			Verse       string `json:"verse"`
			Interpret   string `json:"interpret"`
			Score       int    `json:"score"`
			Advice      string `json:"advice"`
			LuckyColor  string `json:"luckyColor"`
			LuckyNumber int    `json:"luckyNumber"`
			LuckyDir    string `json:"luckyDir"`
		}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			modelError(c, "解析签文失败，请重试")
			return
		}
		clamp := func(v int) int {
			if v < 1 {
				return 50
			}
			if v > 100 {
				return 100
			}
			return v
		}
		ln := parsed.LuckyNumber
		if ln < 1 {
			ln = 7
		}
		if ln > 99 {
			ln = 99
		}
		resp := FortuneResponse{
			Category: category, Verse: parsed.Verse, Interpret: parsed.Interpret,
			Score: clamp(parsed.Score), Advice: parsed.Advice,
			LuckyColor: parsed.LuckyColor, LuckyNumber: ln, LuckyDir: parsed.LuckyDir,
		}
		c.JSON(http.StatusOK, resp)
	}
}
