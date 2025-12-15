package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"ads-creative-gen-platform/config"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// QiniuService 七牛云上传服务
type QiniuService struct {
	mac      *qbox.Mac
	cfg      *storage.Config
	bucket   string
	domain   string
	basePath string
}

// NewQiniuService 创建七牛云服务
func NewQiniuService() *QiniuService {
	if config.QiniuConfig == nil || config.QiniuConfig.AccessKey == "" {
		return nil
	}

	mac := qbox.NewMac(config.QiniuConfig.AccessKey, config.QiniuConfig.SecretKey)

	cfg := &storage.Config{
		UseHTTPS:      true,
		UseCdnDomains: false,
	}

	// 设置区域
	// 设置区域（支持全部七牛云 Region）
	switch config.QiniuConfig.Region {
	case "z0", "cn-east-1":
		cfg.Region = &storage.ZoneHuadong
	case "z1", "cn-north-1":
		cfg.Region = &storage.ZoneHuabei
	case "z2", "cn-south-1":
		cfg.Region = &storage.ZoneHuanan
	case "na0", "us-north-1":
		cfg.Region = &storage.ZoneBeimei
	case "as0", "ap-southeast-1":
		cfg.Region = &storage.ZoneXinjiapo
	case "as1":
		cfg.Region = &storage.ZoneXinjiapo
	default:
		// 强烈建议直接 panic，防止“悄悄传错区”
		panic(fmt.Sprintf("unsupported Qiniu region: %s", config.QiniuConfig.Region))
	}

	return &QiniuService{
		mac:      mac,
		cfg:      cfg,
		bucket:   config.QiniuConfig.Bucket,
		domain:   config.QiniuConfig.Domain,
		basePath: config.QiniuConfig.BasePath,
	}
}

// UploadFromURL 从 URL 下载并上传到七牛云
func (s *QiniuService) UploadFromURL(sourceURL string, fileName string) (string, error) {
	if s == nil {
		return sourceURL, fmt.Errorf("qiniu service not initialized")
	}

	// 下载图片
	resp, err := http.Get(sourceURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// 读取内容
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// 生成存储路径
	key := s.generateKey(fileName)

	// 上传到七牛云
	if err := s.uploadBytes(key, data); err != nil {
		return "", fmt.Errorf("failed to upload to qiniu: %w", err)
	}

	// 返回公开访问 URL
	publicURL := s.getPublicURL(key)
	return publicURL, nil
}

// uploadBytes 上传字节数据
func (s *QiniuService) uploadBytes(key string, data []byte) error {
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", s.bucket, key),
	}
	upToken := putPolicy.UploadToken(s.mac)

	formUploader := storage.NewFormUploader(s.cfg)
	ret := storage.PutRet{}

	dataLen := int64(len(data))
	err := formUploader.Put(
		context.Background(),
		&ret,
		upToken,
		key,
		io.NopCloser(bytes.NewReader(data)),
		dataLen,
		&storage.PutExtra{},
	)

	if err != nil {
		return err
	}

	return nil
}

// generateKey 生成存储路径
func (s *QiniuService) generateKey(fileName string) string {
	// 格式: s3/2024/01/15/uuid.png
	now := time.Now()
	dir := fmt.Sprintf("%s%d/%02d/%02d",
		s.basePath,
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

// getPublicURL 获取公开访问 URL
func (s *QiniuService) getPublicURL(key string) string {
	// 优先使用公共云访问域名
	if config.QiniuConfig.PublicCloudDomain != "" {
		// 移除末尾的斜杠
		publicDomain := config.QiniuConfig.PublicCloudDomain
		if publicDomain[len(publicDomain)-1] == '/' {
			publicDomain = publicDomain[:len(publicDomain)-1]
		}
		return fmt.Sprintf("%s/%s", publicDomain, key)
	}

	// 其次使用自定义域名
	if s.domain != "" {
		// 移除末尾的斜杠
		domain := s.domain
		if domain[len(domain)-1] == '/' {
			domain = domain[:len(domain)-1]
		}
		return fmt.Sprintf("%s/%s", domain, key)
	}

	// 如果都没有配置，返回七牛云默认域名格式
	// 格式: https://[bucket].s3.[region].qiniucs.com/[key]
	return fmt.Sprintf("https://%s.s3.%s.qiniucs.com/%s", s.bucket, config.QiniuConfig.Region, key)
}

// GetFullURL 根据文件键获取完整访问URL
func (s *QiniuService) GetFullURL(key string) string {
	return s.getPublicURL(key)
}

// UploadFile 上传本地文件
func (s *QiniuService) UploadFile(localPath string, fileName string) (string, error) {
	if s == nil {
		return "", fmt.Errorf("qiniu service not initialized")
	}

	key := s.generateKey(fileName)

	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", s.bucket, key),
	}
	upToken := putPolicy.UploadToken(s.mac)

	formUploader := storage.NewFormUploader(s.cfg)
	ret := storage.PutRet{}

	err := formUploader.PutFile(
		context.Background(),
		&ret,
		upToken,
		key,
		localPath,
		&storage.PutExtra{},
	)

	if err != nil {
		return "", err
	}

	publicURL := s.getPublicURL(key)
	return publicURL, nil
}
