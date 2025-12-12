package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ads-creative-gen-platform/config"
)

// TongyiClient 通义 API 客户端
type TongyiClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewTongyiClient 创建通义客户端
func NewTongyiClient() *TongyiClient {
	return &TongyiClient{
		apiKey:  config.TongyiConfig.APIKey,
		baseURL: "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ImageGenRequest 图像生成请求
type ImageGenRequest struct {
	Model      string         `json:"model"`
	Input      ImageGenInput  `json:"input"`
	Parameters ImageGenParams `json:"parameters"`
}

type ImageGenInput struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
}

type ImageGenParams struct {
	Size    string `json:"size,omitempty"`     // "1024*1024", "720*1280", "1280*720"
	N       int    `json:"n,omitempty"`        // 生成图片数量，默认1
	Seed    int    `json:"seed,omitempty"`     // 随机种子
	RefImg  string `json:"ref_img,omitempty"`  // 参考图URL
	RefMode string `json:"ref_mode,omitempty"` // "repaint"
}

// ImageGenResponse 图像生成响应
type ImageGenResponse struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"` // SUCCEEDED, FAILED, RUNNING
		Results    []struct {
			URL string `json:"url"`
		} `json:"results"`
		Message string `json:"message,omitempty"`
	} `json:"output"`
	RequestID string `json:"request_id"`
}

// GenerateImage 生成图片
func (c *TongyiClient) GenerateImage(prompt string, size string, n int) (*ImageGenResponse, error) {
	if size == "" {
		size = "1024*1024" // 使用API支持的标准尺寸
	}
	// 确保使用API支持的尺寸之一
	supportedSizes := []string{"1024*1024", "720*1280", "1280*720", "768*1152"}
	isSupported := false
	for _, supportedSize := range supportedSizes {
		if size == supportedSize {
			isSupported = true
			break
		}
	}

	if !isSupported {
		size = "1024*1024" // 如果不支持，回退到默认尺寸
	}

	if n <= 0 {
		n = 1
	}

	req := ImageGenRequest{
		Model: config.TongyiConfig.ImageModel,
		Input: ImageGenInput{
			Prompt: prompt,
		},
		Parameters: ImageGenParams{
			Size: size,
			N:    n,
		},
	}

	return c.callAPI(req)
}

// GenerateImageWithProduct 带商品图生成
func (c *TongyiClient) GenerateImageWithProduct(prompt string, productImageURL string, size string) (*ImageGenResponse, error) {
	if size == "" {
		size = "1024*1024" // 使用API支持的标准尺寸
	}
	// 确保使用API支持的尺寸之一
	supportedSizes := []string{"1024*1024", "720*1280", "1280*720", "768*1152"}
	isSupported := false
	for _, supportedSize := range supportedSizes {
		if size == supportedSize {
			isSupported = true
			break
		}
	}

	if !isSupported {
		size = "1024*1024" // 如果不支持，回退到默认尺寸
	}

	req := ImageGenRequest{
		Model: config.TongyiConfig.ImageModel,
		Input: ImageGenInput{
			Prompt: prompt,
		},
		Parameters: ImageGenParams{
			Size:    size,
			N:       1,
			RefImg:  productImageURL,
			RefMode: "repaint",
		},
	}

	return c.callAPI(req)
}

// callAPI 调用通义 API
func (c *TongyiClient) callAPI(req ImageGenRequest) (*ImageGenResponse, error) {
	// 序列化请求
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("X-DashScope-Async", "enable") // 异步模式

	// 发送请求
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result ImageGenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	return &result, nil
}

// QueryTask 查询任务状态
func (c *TongyiClient) QueryTask(taskID string) (*ImageGenResponse, error) {
	url := fmt.Sprintf("https://dashscope.aliyuncs.com/api/v1/tasks/%s", taskID)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result ImageGenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	return &result, nil
}
