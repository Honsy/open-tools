package lab

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type WorldCupPredictRequest struct {
	TeamA string `json:"teamA" binding:"required"`
	TeamB string `json:"teamB" binding:"required"`
}

type WorldCupPredictResponse struct {
	TeamA          string `json:"teamA"`
	TeamB          string `json:"teamB"`
	TeamAWinProb   int    `json:"teamAWinProb"`
	TeamBWinProb   int    `json:"teamBWinProb"`
	DrawProb       int    `json:"drawProb"`
	PredictedScore string `json:"predictedScore"`
	Analysis       string `json:"analysis"`
	KeyFactors     string `json:"keyFactors"`
	Stage          string `json:"stage"`
	Confidence     string `json:"confidence"`
}

// AIWorldCupPredict 根据 2026 世界杯赛程和队伍实力，AI 预测两支队伍的比赛结果
func AIWorldCupPredict(client Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req WorldCupPredictRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, "请选择两支球队")
			return
		}

		teamA := strings.TrimSpace(req.TeamA)
		teamB := strings.TrimSpace(req.TeamB)

		if teamA == "" || teamB == "" {
			badRequest(c, "球队名称不能为空")
			return
		}
		if strings.EqualFold(teamA, teamB) {
			badRequest(c, "请选择两支不同的球队")
			return
		}

		systemPrompt := `你是资深足球数据分析师和世界杯预测专家，精通 2026 年 FIFA 世界杯全部赛程、各国家队实力分析、历史对战记录和最新动态。

用户会提供两支球队的名称，请做以下分析：

1. **预测比分**：给出最可能的具体比分
2. **胜率分布**：A队胜率、B队胜率、平局概率（均为 0-100 整数，三者之和为 100）
3. **关键分析**：100-150 字中文分析，涵盖实力对比、近期状态、阵容厚度、战术风格克制关系
4. **关键因素**：列出 3-5 个影响比赛走势的核心因素（如球星状态、伤病情况、战术克制等），用 "、" 分隔
5. **比赛阶段**：根据 2026 世界杯赛程推断这场比赛可能发生在哪个阶段（小组赛/16强/8强/半决赛/决赛）
6. **信心指数**：对预测结果的信心程度（高/中/低）

2026 世界杯由美国、加拿大、墨西哥联合举办，48 支球队参赛，赛制为 12 个小组每组 4 队。
请基于你的知识库中截至 2026 年 6 月的球队信息进行分析。

请严格按以下 JSON 格式输出，不要包含任何其他文字：
{"predictedScore":"2-1","teamAWinProb":55,"teamBWinProb":25,"drawProb":20,"analysis":"分析内容","keyFactors":"因素1、因素2、因素3","stage":"小组赛","confidence":"中"}`

		userPrompt := "A队：" + teamA + "\nB队：" + teamB

		result, err := client.chat(c.Request.Context(), systemPrompt, userPrompt)
		if err != nil {
			modelError(c, "预测失败，请稍后重试")
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
			PredictedScore string `json:"predictedScore"`
			TeamAWinProb   int    `json:"teamAWinProb"`
			TeamBWinProb   int    `json:"teamBWinProb"`
			DrawProb       int    `json:"drawProb"`
			Analysis       string `json:"analysis"`
			KeyFactors     string `json:"keyFactors"`
			Stage          string `json:"stage"`
			Confidence     string `json:"confidence"`
		}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			modelError(c, "解析预测结果失败，请重试")
			return
		}

		clampProb := func(v int) int {
			if v < 0 {
				return 0
			}
			if v > 100 {
				return 100
			}
			return v
		}
		parsed.TeamAWinProb = clampProb(parsed.TeamAWinProb)
		parsed.TeamBWinProb = clampProb(parsed.TeamBWinProb)
		parsed.DrawProb = clampProb(parsed.DrawProb)
		total := parsed.TeamAWinProb + parsed.TeamBWinProb + parsed.DrawProb
		if total == 0 {
			parsed.TeamAWinProb, parsed.TeamBWinProb, parsed.DrawProb = 40, 30, 30
		} else if total != 100 {
			parsed.TeamAWinProb = parsed.TeamAWinProb * 100 / total
			parsed.TeamBWinProb = parsed.TeamBWinProb * 100 / total
			parsed.DrawProb = 100 - parsed.TeamAWinProb - parsed.TeamBWinProb
			if parsed.DrawProb < 0 {
				parsed.DrawProb = 0
			}
		}

		confidence := parsed.Confidence
		if confidence != "高" && confidence != "中" && confidence != "低" {
			confidence = "中"
		}

		if parsed.PredictedScore == "" {
			parsed.PredictedScore = "-"
		}
		if parsed.Stage == "" {
			parsed.Stage = "待定"
		}

		resp := WorldCupPredictResponse{
			TeamA:          teamA,
			TeamB:          teamB,
			TeamAWinProb:   parsed.TeamAWinProb,
			TeamBWinProb:   parsed.TeamBWinProb,
			DrawProb:       parsed.DrawProb,
			PredictedScore: parsed.PredictedScore,
			Analysis:       parsed.Analysis,
			KeyFactors:     parsed.KeyFactors,
			Stage:          parsed.Stage,
			Confidence:     confidence,
		}
		c.JSON(http.StatusOK, resp)
	}
}
