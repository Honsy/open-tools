package lab

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"opentools/config"

	"github.com/gin-gonic/gin"
)

const dashscope = "https://dashscope.aliyuncs.com"

type Client struct {
	Key    string
	Models []string
}

var providerHTTP = &http.Client{Timeout: 120 * time.Second}

func FromConfig(cfg config.Config) Client {
	models := strings.FieldsFunc(cfg.AIBailianChatModels, func(r rune) bool { return r == ',' || r == ' ' })
	if len(models) == 0 {
		models = []string{"qwen-plus"}
	}
	return Client{Key: cfg.AIBailianAPIKey, Models: models}
}

func (c Client) request(ctx context.Context, address string, payload any, async bool, result any) (int, error) {
	method := http.MethodGet
	var body io.Reader
	if payload != nil {
		method = http.MethodPost
		data, err := json.Marshal(payload)
		if err != nil {
			return 0, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, dashscope+address, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if async {
		req.Header.Set("X-DashScope-Async", "enable")
	}
	response, err := providerHTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return response.StatusCode, fmt.Errorf("DashScope 返回 HTTP %d", response.StatusCode)
	}
	return response.StatusCode, json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(result)
}

func (c Client) chat(ctx context.Context, system, user string) (string, error) {
	return c.Complete(ctx, system, user, 2048)
}

func (c Client) Complete(ctx context.Context, system, user string, maxTokens int) (string, error) {
	if c.Key == "" || len(c.Models) == 0 {
		return "", errors.New("AI 服务暂未配置")
	}
	var lastErr error
	for _, model := range c.Models {
		var result struct {
			Choices []struct {
				Message struct{ Content string } `json:"message"`
			} `json:"choices"`
		}
		status, err := c.request(ctx, "/compatible-mode/v1/chat/completions", map[string]any{
			"model": model, "temperature": 0.7, "max_tokens": maxTokens,
			"messages": []map[string]string{{"role": "system", "content": system}, {"role": "user", "content": user}},
		}, false, &result)
		if err == nil && len(result.Choices) > 0 {
			return result.Choices[0].Message.Content, nil
		}
		if err == nil {
			return "", errors.New("模型没有返回结果")
		}
		lastErr = err
		if status != 400 && status != 403 && status != 429 && status < 500 {
			break
		}
	}
	return "", lastErr
}

func (c Client) Poster(ctx context.Context, prompt, size string) ([]string, error) {
	var result struct {
		Output struct {
			Choices []struct {
				Message struct {
					Content []struct{ Image string } `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		} `json:"output"`
	}
	_, err := c.request(ctx, "/api/v1/services/aigc/multimodal-generation/generation", map[string]any{
		"model": "wan2.6-t2i",
		"input": map[string]any{"messages": []map[string]any{
			{"role": "user", "content": []map[string]string{{"text": prompt}}},
		}},
		"parameters": map[string]any{"size": size, "n": 1, "prompt_extend": true, "watermark": false},
	}, false, &result)
	if err != nil {
		return nil, err
	}
	var images []string
	for _, choice := range result.Output.Choices {
		for _, part := range choice.Message.Content {
			images = append(images, part.Image)
		}
	}
	return imageLinks(images)
}

func (c Client) Background(ctx context.Context, imageURL, prompt string, count int) ([]string, error) {
	var task struct {
		Output struct {
			ID string `json:"task_id"`
		} `json:"output"`
	}
	_, err := c.request(ctx, "/api/v1/services/aigc/background-generation/generation/", map[string]any{
		"model":      "wanx-background-generation-v2",
		"input":      map[string]string{"base_image_url": imageURL, "ref_prompt": prompt},
		"parameters": map[string]any{"n": count, "model_version": "v3"},
	}, true, &task)
	if err != nil {
		return nil, err
	}
	if task.Output.ID == "" {
		return nil, errors.New("图像服务没有返回任务")
	}
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
		var result struct {
			Output struct {
				Status  string                 `json:"task_status"`
				Results []struct{ URL string } `json:"results"`
			} `json:"output"`
		}
		_, err := c.request(ctx, "/api/v1/tasks/"+url.PathEscape(task.Output.ID), nil, false, &result)
		if err != nil {
			return nil, err
		}
		switch result.Output.Status {
		case "SUCCEEDED":
			var images []string
			for _, image := range result.Output.Results {
				images = append(images, image.URL)
			}
			return imageLinks(images)
		case "FAILED", "CANCELED", "UNKNOWN":
			return nil, errors.New("图片背景替换失败")
		}
	}
}

func imageLinks(raw []string) ([]string, error) {
	var valid []string
	for _, image := range raw {
		parsed, err := url.Parse(image)
		if err == nil && parsed.Scheme == "https" && parsed.Hostname() != "" && parsed.User == nil {
			valid = append(valid, image)
		}
	}
	if len(valid) == 0 {
		return nil, errors.New("图像服务没有返回图片")
	}
	return valid, nil
}

func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func modelError(c *gin.Context, message string) {
	c.JSON(http.StatusBadGateway, gin.H{"error": message})
}
