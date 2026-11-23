package lab

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ZodiacMatchRequest struct {
	ZodiacA string `json:"zodiacA" binding:"required"`
	ZodiacB string `json:"zodiacB" binding:"required"`
}

type ZodiacMatchResponse struct {
	ZodiacA     string `json:"zodiacA"`
	ZodiacB     string `json:"zodiacB"`
	Score       int    `json:"score"`
	LoveScore   int    `json:"loveScore"`
	WorkScore   int    `json:"workScore"`
	FriendScore int    `json:"friendScore"`
	Summary     string `json:"summary"`
	Strengths   string `json:"strengths"`
	Challenges  string `json:"challenges"`
	Advice      string `json:"advice"`
}

func AIZodiacMatch(client Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ZodiacMatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, "请选择两个星座")
			return
		}

		zodiacA := strings.TrimSpace(req.ZodiacA)
		zodiacB := strings.TrimSpace(req.ZodiacB)

		if !validZodiacs[zodiacA] || !validZodiacs[zodiacB] {
			badRequest(c, "无效的星座")
			return
		}

		systemPrompt := `你是一位专业的星座分析师。用户会给出两个星座，请分析它们的匹配度。

要求：
1. 综合匹配度评分（1-100）
2. 爱情匹配度评分（1-100）
3. 事业匹配度评分（1-100）
4. 友情匹配度评分（1-100）
5. 一句话总结（20字以内，概括两者关系的核心特质）
6. 优势互补：30-50字描述两个星座在一起的优势
7. 潜在挑战：30-50字描述需要注意的摩擦点
8. 相处建议：30-50字给出一条实用的相处建议

语气活泼有趣，可以用比喻。请严格按以下 JSON 格式输出，不要包含任何其他文字：
{"score":85,"loveScore":80,"workScore":90,"friendScore":88,"summary":"一句话总结","strengths":"优势互补描述","challenges":"潜在挑战描述","advice":"相处建议"}`

		userPrompt := "星座A：" + zodiacA + "\n星座B：" + zodiacB

		result, err := client.chat(c.Request.Context(), systemPrompt, userPrompt)
		if err != nil {
			modelError(c, "匹配分析失败，请稍后重试")
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
			Score       int    `json:"score"`
			LoveScore   int    `json:"loveScore"`
			WorkScore   int    `json:"workScore"`
			FriendScore int    `json:"friendScore"`
			Summary     string `json:"summary"`
			Strengths   string `json:"strengths"`
			Challenges  string `json:"challenges"`
			Advice      string `json:"advice"`
		}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			modelError(c, "解析匹配结果失败，请重试")
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

		resp := ZodiacMatchResponse{
			ZodiacA:     zodiacA,
			ZodiacB:     zodiacB,
			Score:       clamp(parsed.Score),
			LoveScore:   clamp(parsed.LoveScore),
			WorkScore:   clamp(parsed.WorkScore),
			FriendScore: clamp(parsed.FriendScore),
			Summary:     parsed.Summary,
			Strengths:   parsed.Strengths,
			Challenges:  parsed.Challenges,
			Advice:      parsed.Advice,
		}
		c.JSON(http.StatusOK, resp)
	}
}
