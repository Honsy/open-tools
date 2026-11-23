package lab

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var validZodiacs = map[string]bool{
	"白羊座": true, "金牛座": true, "双子座": true, "巨蟹座": true,
	"狮子座": true, "处女座": true, "天秤座": true, "天蝎座": true,
	"射手座": true, "摩羯座": true, "水瓶座": true, "双鱼座": true,
}

type HoroscopeRequest struct {
	Zodiac string `json:"zodiac" binding:"required"`
}

type HoroscopeResponse struct {
	Zodiac  string `json:"zodiac"`
	Date    string `json:"date"`
	Overall string `json:"overall"`
	Love    string `json:"love"`
	Career  string `json:"career"`
	Wealth  string `json:"wealth"`
	Lucky   string `json:"lucky"`
	Score   int    `json:"score"`
	Summary string `json:"summary"`
}

func AIHoroscope(client Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req HoroscopeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, "请选择星座")
			return
		}

		zodiac := strings.TrimSpace(req.Zodiac)
		if !validZodiacs[zodiac] {
			badRequest(c, "无效的星座")
			return
		}

		today := time.Now().Format("2006年01月02日")

		systemPrompt := `你是一位专业的星座占卜师。用户会告诉你一个星座，请为该星座生成今日运势。

要求：
1. 综合运势、爱情运势、事业运势、财富运势各用20-40字描述，语气温暖有趣
2. 幸运提示（如幸运颜色、幸运数字、幸运方位等组合一句话）
3. 综合评分（1-100的整数）
4. 一句话总结（15字以内，有哲理感或趣味感）

请严格按以下 JSON 格式输出，不要包含任何其他文字：
{"overall":"综合运势","love":"爱情运势","career":"事业运势","wealth":"财富运势","lucky":"幸运提示","score":85,"summary":"一句话总结"}`

		userPrompt := "星座：" + zodiac + "\n日期：" + today

		result, err := client.chat(c.Request.Context(), systemPrompt, userPrompt)
		if err != nil {
			modelError(c, "运势生成失败，请稍后重试")
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
			Overall string `json:"overall"`
			Love    string `json:"love"`
			Career  string `json:"career"`
			Wealth  string `json:"wealth"`
			Lucky   string `json:"lucky"`
			Score   int    `json:"score"`
			Summary string `json:"summary"`
		}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			modelError(c, "解析运势失败，请重试")
			return
		}

		if parsed.Score < 1 || parsed.Score > 100 {
			parsed.Score = 75
		}

		resp := HoroscopeResponse{
			Zodiac:  zodiac,
			Date:    today,
			Overall: parsed.Overall,
			Love:    parsed.Love,
			Career:  parsed.Career,
			Wealth:  parsed.Wealth,
			Lucky:   parsed.Lucky,
			Score:   parsed.Score,
			Summary: parsed.Summary,
		}
		c.JSON(http.StatusOK, resp)
	}
}
