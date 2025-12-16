package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"ads-creative-gen-platform/config"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// QiniuClient 七牛云上传客户端
type QiniuClient struct {
	mac        *qbox.Mac
	cfg        *storage.Config
	bucket     string
	domain     string
	basePath   string
	httpClient *http.Client
}

// NewQiniuClient 创建七牛云客户端，如果未配置则返回 nil
func NewQiniuClient() *QiniuClient {
	if config.QiniuConfig == nil || config.QiniuConfig.AccessKey == "" {
		return nil
	}

	mac := qbox.NewMac(config.QiniuConfig.AccessKey, config.QiniuConfig.SecretKey)

	cfg := &storage.Config{
		UseHTTPS:      true,
		UseCdnDomains: false,
	}

	switch config.QiniuConfig.Region {
	case "z0", "cn-east-1":
		cfg.Region = &storage.ZoneHuadong
	case "z1", "cn-north-1":
		cfg.Region = &storage.ZoneHuabei
	case "z2", "cn-south-1":
		cfg.Region = &storage.ZoneHuanan
	case "na0", "us-north-1":
		cfg.Region = &storage.ZoneBeimei
	case "as0", "ap-southeast-1", "as1":
		cfg.Region = &storage.ZoneXinjiapo
	default:
		panic(fmt.Sprintf("unsupported Qiniu region: %s", config.QiniuConfig.Region))
	}

	return &QiniuClient{
		mac:        mac,
		cfg:        cfg,
		bucket:     config.QiniuConfig.Bucket,
		domain:     config.QiniuConfig.Domain,
		basePath:   config.QiniuConfig.BasePath,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// UploadFromURL 从 URL 下载并上传到七牛云
func (c *QiniuClient) UploadFromURL(ctx context.Context, sourceURL string, fileName string) (string, error) {
	if c == nil {
		return sourceURL, fmt.Errorf("qiniu client not initialized")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	key := c.GenerateKey(fileName)

	if err := c.uploadBytes(ctx, key, data); err != nil {
		return "", fmt.Errorf("failed to upload to qiniu: %w", err)
	}

	publicURL := c.getPublicURL(key)
	return publicURL, nil
}

// UploadFile 上传本地文件
func (c *QiniuClient) UploadFile(ctx context.Context, localPath string, fileName string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("qiniu client not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	key := c.GenerateKey(fileName)

	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", c.bucket, key),
	}
	upToken := putPolicy.UploadToken(c.mac)

	formUploader := storage.NewFormUploader(c.cfg)
	ret := storage.PutRet{}

	if err := formUploader.PutFile(ctx, &ret, upToken, key, localPath, &storage.PutExtra{}); err != nil {
		return "", err
	}

	publicURL := c.getPublicURL(key)
	return publicURL, nil
}

// GenerateKey 生成存储路径
func (c *QiniuClient) GenerateKey(fileName string) string {
	now := time.Now()
	dir := fmt.Sprintf("%s%d/%02d/%02d",
		c.basePath,
		now.Year(),
		now.Month(),
		now.Day(),
	)

	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".png"
	}

	return fmt.Sprintf("%s/%s%s", dir, fileName, ext)
}

func (c *QiniuClient) uploadBytes(ctx context.Context, key string, data []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}

	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", c.bucket, key),
	}
	upToken := putPolicy.UploadToken(c.mac)

	formUploader := storage.NewFormUploader(c.cfg)
	ret := storage.PutRet{}

	dataLen := int64(len(data))
	return formUploader.Put(
		ctx,
		&ret,
		upToken,
		key,
		io.NopCloser(bytes.NewReader(data)),
		dataLen,
		&storage.PutExtra{},
	)
}

func (c *QiniuClient) getPublicURL(key string) string {
	if config.QiniuConfig.PublicCloudDomain != "" {
		publicDomain := strings.TrimSuffix(config.QiniuConfig.PublicCloudDomain, "/")
		return fmt.Sprintf("%s/%s", publicDomain, key)
	}

	if c.domain != "" {
		domain := strings.TrimSuffix(c.domain, "/")
		return fmt.Sprintf("%s/%s", domain, key)
	}

	return fmt.Sprintf("https://%s.s3.%s.qiniucs.com/%s", c.bucket, config.QiniuConfig.Region, key)
}
