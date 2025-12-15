package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"ads-creative-gen-platform/config"
	"ads-creative-gen-platform/internal/tracing"
)

// TongyiClient 通义 API 客户端
type TongyiClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
	tracer  *tracing.Tracer
}

// NewTongyiClient 创建通义客户端
func NewTongyiClient() *TongyiClient {
	return &TongyiClient{
		apiKey:  config.TongyiConfig.APIKey,
		baseURL: "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		tracer: tracing.NewTracer(),
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

// GenerateImage 生成图片，traceID 可选（空则新建）
func (c *TongyiClient) GenerateImage(ctx context.Context, prompt string, size string, n int, source string, traceID string, productName string) (*ImageGenResponse, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if traceID == "" {
		ctx, traceID = c.tracer.Start(ctx, "tongyi-image", config.TongyiConfig.ImageModel, productName, prompt, productName)
	} else {
		ctx = context.WithValue(ctx, tracing.CtxKeyTraceID, traceID)
	}
	c.tracer.Step(ctx, "generate_image_start", "tongyi-image", "info", prompt, "", "", time.Now(), time.Now())
	log.Printf("[通义客户端] 开始生成图片, 提示词: %s, 尺寸: %s, 数量: %d", prompt, size, n)

	if size == "" {
		size = "1024*1024" // 使用API支持的标准尺寸
		log.Printf("[通义客户端] 尺寸为空，使用默认尺寸: %s", size)
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
		log.Printf("[通义客户端] 尺寸 %s 不被支持，回退到默认尺寸: 1024*1024", size)
		size = "1024*1024" // 如果不支持，回退到默认尺寸
	}

	if n <= 0 {
		n = 1
		log.Printf("[通义客户端] 数量为0或负数，设置为默认值: 1")
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

	log.Printf("[通义客户端] 发送请求到API，模型: %s, 尺寸: %s, 数量: %d", req.Model, req.Parameters.Size, req.Parameters.N)

	response, err := c.callAPI(ctx, req)
	if err != nil {
		log.Printf("[通义客户端] API调用失败: %v", err)
		c.tracer.Finish(ctx, "failed", "", err.Error())
		return nil, traceID, err
	}

	log.Printf("[通义客户端] API调用成功，任务ID: %s, 状态: %s", response.Output.TaskID, response.Output.TaskStatus)
	c.tracer.Step(ctx, "generate_image", "tongyi-image", "success", prompt, response.Output.TaskID, "", time.Now(), time.Now())
	return response, traceID, nil
}

// GenerateImageWithProduct 带商品图生成，traceID 可选（空则新建）
func (c *TongyiClient) GenerateImageWithProduct(ctx context.Context, prompt string, productImageURL string, size string, n int, source string, traceID string, productName string) (*ImageGenResponse, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if traceID == "" {
		ctx, traceID = c.tracer.Start(ctx, "tongyi-image", config.TongyiConfig.ImageModel, productName, prompt, productName)
	} else {
		ctx = context.WithValue(ctx, tracing.CtxKeyTraceID, traceID)
	}
	c.tracer.Step(ctx, "generate_image_with_product_start", "tongyi-image", "info", prompt, productImageURL, "", time.Now(), time.Now())
	log.Printf("[通义客户端] 开始带商品图生成, 提示词: %s, 商品图URL: %s, 尺寸: %s", prompt, productImageURL, size)

	if size == "" {
		size = "1024*1024" // 使用API支持的标准尺寸
		log.Printf("[通义客户端] 尺寸为空，使用默认尺寸: %s", size)
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
		log.Printf("[通义客户端] 尺寸 %s 不被支持，回退到默认尺寸: 1024*1024", size)
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
			Size:    size,
			N:       n,
			RefImg:  productImageURL,
			RefMode: "repaint",
		},
	}

	log.Printf("[通义客户端] 发送带商品图请求到API，模型: %s, 尺寸: %s, 商品图: %s", req.Model, req.Parameters.Size, req.Parameters.RefImg)

	response, err := c.callAPI(ctx, req)
	if err != nil {
		log.Printf("[通义客户端] 带商品图API调用失败: %v", err)
		c.tracer.Finish(ctx, "failed", "", err.Error())
		return nil, traceID, err
	}

	log.Printf("[通义客户端] 带商品图API调用成功，任务ID: %s, 状态: %s", response.Output.TaskID, response.Output.TaskStatus)
	c.tracer.Step(ctx, "generate_image_with_product", "tongyi-image", "success", prompt, response.Output.TaskID, "", time.Now(), time.Now())
	return response, traceID, nil
}

// callAPI 调用通义 API
func (c *TongyiClient) callAPI(ctx context.Context, req ImageGenRequest) (*ImageGenResponse, error) {
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
		c.tracer.Step(ctx, "http_call", "tongyi-image", "failed", string(body), "", err.Error(), time.Now(), time.Now())
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.tracer.Step(ctx, "http_call", "tongyi-image", "failed", string(body), "", err.Error(), time.Now(), time.Now())
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		c.tracer.Step(ctx, "http_call", "tongyi-image", "failed", string(body), string(respBody), fmt.Sprintf("status %d", resp.StatusCode), time.Now(), time.Now())
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result ImageGenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	return &result, nil
}

// QueryTask 查询任务状态，traceID 必须传递，复用同一条链路
func (c *TongyiClient) QueryTask(ctx context.Context, traceID string, taskID string, source string) (*ImageGenResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if traceID == "" {
		// 若未传 traceID，仍创建但建议上层复用同一 trace
		ctx, traceID = c.tracer.Start(ctx, "tongyi-image", config.TongyiConfig.ImageModel, source, taskID, "")
	} else {
		ctx = context.WithValue(ctx, tracing.CtxKeyTraceID, traceID)
	}
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
		c.tracer.Step(ctx, "query_task", "tongyi-task", "failed", taskID, string(respBody), err.Error(), time.Now(), time.Now())
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	c.tracer.Step(ctx, "query_task", "tongyi-task", "success", taskID, result.Output.TaskStatus, "", time.Now(), time.Now())
	return &result, nil
}
