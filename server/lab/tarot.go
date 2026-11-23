package lab

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type TarotRequest struct {
	Question string `json:"question"`
}

type TarotCard struct {
	Name     string `json:"name"`
	Meaning  string `json:"meaning"`
	Position string `json:"position"`
}

type TarotResponse struct {
	Question string      `json:"question"`
	Cards    []TarotCard `json:"cards"`
	Overall  string      `json:"overall"`
	Advice   string      `json:"advice"`
}

func AITarot(client Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TarotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, "参数错误")
			return
		}
		question := strings.TrimSpace(req.Question)
		systemPrompt := `你是资深塔罗占卜师。用户可能会提出一个问题，请为其进行三张牌塔罗占卜（无牌阵）。
请从以下 22 张大阿卡纳中随机抽取 3 张不同的牌：
愚者、魔术师、女祭司、皇后、皇帝、教皇、恋人、战车、力量、隐士、命运之轮、正义、倒吊人、死神、节制、恶魔、高塔、星星、月亮、太阳、审判、世界
每张牌分别对应：过去（Past）、现在（Present）、未来（Future）
为每张牌给出 20-40 字的正位/逆位解读，贴合用户的问题（如有）。
最后给出 60-80 字的整体运势解读和 20-30 字的建议。
请严格按以下 JSON 格式输出，不要包含任何其他文字：
{"cards":[{"name":"恋人","meaning":"解读内容","position":"Past"},{"name":"星星","meaning":"解读内容","position":"Present"},{"name":"世界","meaning":"解读内容","position":"Future"}],"overall":"整体解读","advice":"建议"}`
		userPrompt := "我的问题"
		if question != "" {
			userPrompt = "我的问题：" + question
		}
		result, err := client.chat(c.Request.Context(), systemPrompt, userPrompt)
		if err != nil {
			modelError(c, "占卜失败，请稍后重试")
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
			Cards   []TarotCard `json:"cards"`
			Overall string      `json:"overall"`
			Advice  string      `json:"advice"`
		}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			modelError(c, "解析占卜结果失败，请重试")
			return
		}
		if len(parsed.Cards) == 0 {
			modelError(c, "占卜结果异常，请重试")
			return
		}
		resp := TarotResponse{
			Question: question,
			Cards:    parsed.Cards,
			Overall:  parsed.Overall,
			Advice:   parsed.Advice,
		}
		c.JSON(http.StatusOK, resp)
	}
}
