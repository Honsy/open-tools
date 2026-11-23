package lab

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type DreamRequest struct {
	Dream string `json:"dream" binding:"required"`
}

type DreamResponse struct {
	Meaning       string `json:"meaning"`
	Psychological string `json:"psychological"`
	LifeAdvice    string `json:"lifeAdvice"`
	Score         int    `json:"score"`
	Tags          string `json:"tags"`
}

func AIDream(client Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DreamRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, "请描述你的梦境")
			return
		}
		dream := strings.TrimSpace(req.Dream)
		if dream == "" {
			badRequest(c, "梦境描述不能为空")
			return
		}
		if len([]rune(dream)) < 4 {
			badRequest(c, "梦境描述至少 4 个字符")
			return
		}
		systemPrompt := `你是资深梦境分析师，精通东西方解梦文化。用户会描述一个梦境，请做以下分析：
1. **梦境寓意**：50-80 字中文，解析梦境的核心象征意义
2. **心理分析**：40-60 字，从心理学角度分析梦境反映的潜意识状态
3. **生活启示**：30-50 字，给出对现实生活的启示或建议
4. **梦境指数**：1-100 的整数，代表梦境的积极程度（越高越正面）
5. **梦境标签**：2-4 个标签概括梦境类型，用 "、" 分隔，如"飞行、自由、焦虑"
语气温和有洞察力。请严格按以下 JSON 格式输出，不要包含任何其他文字：
{"meaning":"梦境寓意","psychological":"心理分析","lifeAdvice":"生活启示","score":65,"tags":"飞行、自由"}`
		userPrompt := "梦境描述：" + dream
		result, err := client.chat(c.Request.Context(), systemPrompt, userPrompt)
		if err != nil {
			modelError(c, "解梦失败，请稍后重试")
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
			Meaning       string `json:"meaning"`
			Psychological string `json:"psychological"`
			LifeAdvice    string `json:"lifeAdvice"`
			Score         int    `json:"score"`
			Tags          string `json:"tags"`
		}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			modelError(c, "解析梦境失败，请重试")
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
		resp := DreamResponse{
			Meaning: parsed.Meaning, Psychological: parsed.Psychological,
			LifeAdvice: parsed.LifeAdvice, Score: clamp(parsed.Score),
			Tags: parsed.Tags,
		}
		c.JSON(http.StatusOK, resp)
	}
}
