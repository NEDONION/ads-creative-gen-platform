package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ads-creative-gen-platform/config"
)

// QwenClient 调用千问 LLM 生成文案
type QwenClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
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
func (c *QwenClient) GenerateCopywriting(productName string) (*CopywritingResult, error) {
	if productName == "" {
		return nil, errors.New("product name is required")
	}

	req := CopywritingRequest{
		Model: c.model,
		Input: CopywritingInput{
			Prompt: c.buildPrompt(productName),
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

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var result CopywritingResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	return c.parseResponse(result, string(respBytes))
}

// buildPrompt 构造提示词
func (c *QwenClient) buildPrompt(productName string) string {
	return fmt.Sprintf(`生成产品广告文案: "%s"

请提供以下JSON格式的输出:
{
  "cta_options": ["CTA option 1", "CTA option 2"],
  "selling_point_options": ["Selling point 1", "Selling point 2", "Selling point 3"]
}

要求:
- 生成恰好2个CTA (Call-to-Action) 选项，使用中文
- 生成恰好3个核心卖点选项，使用中文
- CTA应简短有力（3-6个汉字），行动导向（例如: "立即购买", "马上抢购", "了解更多"）
- 卖点应简洁明了（8-15个汉字），突出产品核心优势
- 所有文本必须使用中文
- 只返回有效的JSON格式，不要包含其他文本`, productName)
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

	if len(result.CTAOptions) != 2 {
		return nil, fmt.Errorf("unexpected CTA options length: %d", len(result.CTAOptions))
	}
	if len(result.SellingPointOptions) != 3 {
		return nil, fmt.Errorf("unexpected selling_point_options length: %d", len(result.SellingPointOptions))
	}

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
