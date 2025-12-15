package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ads-creative-gen-platform/config"
	"ads-creative-gen-platform/internal/tracing"
)

// QwenClient 调用千问 LLM 生成文案
type QwenClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
	tracer  *tracing.Tracer
}

// NewQwenClient 创建客户端
func NewQwenClient() *QwenClient {
	return &QwenClient{
		apiKey:  config.TongyiConfig.APIKey,
		model:   config.TongyiConfig.LLMModel,
		baseURL: "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		tracer: tracing.NewTracer(),
	}
}

// CopywritingRequest 请求体
type CopywritingRequest struct {
	Model      string                 `json:"model"`
	Input      CopywritingInput       `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type CopywritingInput struct {
	Prompt string `json:"prompt"`
}

// CopywritingResponse 响应体（兼容不同字段）
type CopywritingResponse struct {
	Output struct {
		Text    string `json:"text"`
		Message string `json:"message"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
}

// CopywritingResult 结构化结果
type CopywritingResult struct {
	CTAOptions          []string `json:"cta_options"`
	SellingPointOptions []string `json:"selling_point_options"`
	RawResponse         string   `json:"-"`
}

// GenerateCopywriting 调用 LLM 生成文案
func (c *QwenClient) GenerateCopywriting(productName string, language string) (*CopywritingResult, error) {
	if productName == "" {
		return nil, errors.New("product name is required")
	}

	lang := strings.ToLower(strings.TrimSpace(language))
	if lang != "en" {
		lang = "zh"
	}

	ctx := context.Background()
	ctx, _ = c.tracer.Start(ctx, "qwen-llm", c.model, productName, fmt.Sprintf("%s|%s", productName, lang), productName)

	req := CopywritingRequest{
		Model: c.model,
		Input: CopywritingInput{
			Prompt: c.buildPrompt(productName, lang),
		},
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	callStart := time.Now()
	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.tracer.Step(ctx, "llm_call", c.model, "failed", req.Input.Prompt, "", err.Error(), callStart, time.Now())
		c.tracer.Finish(ctx, "failed", "", err.Error())
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.tracer.Step(ctx, "llm_call", c.model, "failed", req.Input.Prompt, "", err.Error(), callStart, time.Now())
		c.tracer.Finish(ctx, "failed", "", err.Error())
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.tracer.Step(ctx, "llm_call", c.model, "failed", req.Input.Prompt, string(respBytes), fmt.Sprintf("status %d", resp.StatusCode), callStart, time.Now())
		c.tracer.Finish(ctx, "failed", "", fmt.Sprintf("status %d", resp.StatusCode))
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var result CopywritingResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		c.tracer.Step(ctx, "llm_call", c.model, "failed", req.Input.Prompt, string(respBytes), err.Error(), callStart, time.Now())
		c.tracer.Finish(ctx, "failed", "", err.Error())
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	parsed, err := c.parseResponse(result, string(respBytes))
	status := "success"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}
	c.tracer.Step(ctx, "llm_call", c.model, status, req.Input.Prompt, string(respBytes), errMsg, callStart, time.Now())
	c.tracer.Finish(ctx, status, string(respBytes), errMsg)
	return parsed, err
}

// buildPrompt 构造提示词
func (c *QwenClient) buildPrompt(productName, language string) string {
	langName := "中文"
	ctaRule := "- 生成恰好2个CTA (Call-to-Action) 选项，使用中文，长度 3-6 个汉字，行动导向（如：\"立即购买\"、\"马上抢购\"、\"了解更多\"）"
	spRule := "- 生成恰好3个核心卖点选项，使用中文，长度 8-15 个汉字，突出产品核心优势"

	if language == "en" {
		langName = "English"
		ctaRule = "- Generate exactly 2 CTA (Call-to-Action) options in English, concise and action-oriented (3-7 words, e.g., \"Buy now\", \"Shop today\", \"Learn more\")"
		spRule = "- Generate exactly 3 key selling points in English (6-12 words), focus on product benefits and clarity"
	}

	return fmt.Sprintf(`Generate ad copy for product: "%s"

Return ONLY valid JSON in the following shape:
{
  "cta_options": ["CTA option 1", "CTA option 2"],
  "selling_point_options": ["Selling point 1", "Selling point 2", "Selling point 3"]
}

Rules:
- Target language: %s (always output CTA and selling points in this language, even if product name is another language)
%s
%s
- Keep CTA and selling points consistent in the target language
- Strictly output JSON only, no extra text`, productName, langName, ctaRule, spRule)
}

// parseResponse 解析并校验
func (c *QwenClient) parseResponse(resp CopywritingResponse, raw string) (*CopywritingResult, error) {
	content := strings.TrimSpace(resp.Output.Text)
	if content == "" && len(resp.Output.Choices) > 0 {
		content = strings.TrimSpace(resp.Output.Choices[0].Message.Content)
	}
	if content == "" {
		return nil, errors.New("empty LLM response content")
	}

	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return nil, errors.New("failed to extract JSON from response")
	}

	var result CopywritingResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}
	result.RawResponse = raw

	// 容错：过滤空白，截断/补齐
	filteredCTA := make([]string, 0, len(result.CTAOptions))
	for _, v := range result.CTAOptions {
		if strings.TrimSpace(v) != "" {
			filteredCTA = append(filteredCTA, strings.TrimSpace(v))
		}
	}
	if len(filteredCTA) == 0 {
		return nil, errors.New("no CTA options")
	}
	if len(filteredCTA) > 2 {
		filteredCTA = filteredCTA[:2]
	}
	for len(filteredCTA) < 2 {
		filteredCTA = append(filteredCTA, filteredCTA[len(filteredCTA)-1])
	}

	filteredSP := make([]string, 0, len(result.SellingPointOptions))
	for _, v := range result.SellingPointOptions {
		if strings.TrimSpace(v) != "" {
			filteredSP = append(filteredSP, strings.TrimSpace(v))
		}
	}
	if len(filteredSP) == 0 {
		return nil, errors.New("no selling_point_options")
	}
	if len(filteredSP) > 3 {
		filteredSP = filteredSP[:3]
	}
	for len(filteredSP) < 3 {
		filteredSP = append(filteredSP, filteredSP[len(filteredSP)-1])
	}

	result.CTAOptions = filteredCTA
	result.SellingPointOptions = filteredSP

	return &result, nil
}

// extractJSON 从字符串中提取 JSON
func extractJSON(content string) string {
	// 去掉 ```json 包裹
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return content[start : end+1]
}
