package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// 全局配置对象
var (
	AppConfig      *App
	DatabaseConfig *Database
	TongyiConfig   *Tongyi
	QiniuConfig    *Qiniu
	CacheConfig    *Cache
)

// App 服务配置
type App struct {
	AppMode  string
	HttpPort string
}

// Database 数据库配置
type Database struct {
	Db         string
	DbHost     string
	DbPort     string
	DbUser     string
	DbPassWord string
	DbName     string
	Charset    string
}

// Tongyi 通义API配置
type Tongyi struct {
	APIKey     string
	ImageModel string
	LLMModel   string
}

// Qiniu 七牛云配置
type Qiniu struct {
	AccessKey         string
	SecretKey         string
	Bucket            string
	Domain            string
	PublicCloudDomain string
	Region            string
	BasePath          string
}

// Cache 缓存配置
type Cache struct {
	Enabled           bool
	MaxEntries        int
	DefaultTTL        time.Duration
	DisableExperiment bool
	DisableCreative   bool
	DisableTracing    bool
}

// LoadConfig 加载所有配置
func LoadConfig() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// 加载各模块配置
	loadAppConfig()
	loadDatabaseConfig()
	loadTongyiConfig()
	loadQiniuConfig()
	loadCacheConfig()

	log.Println("✓ All configurations loaded successfully")
}

// loadAppConfig 加载服务配置
func loadAppConfig() {
	// 支持 Railway/Render 等平台的 PORT 环境变量
	port := getEnv("HTTP_PORT", "")
	if port == "" {
		// 如果 HTTP_PORT 未设置，尝试使用 PORT 环境变量（Railway/Render 等平台）
		if envPort := os.Getenv("PORT"); envPort != "" {
			port = ":" + envPort
		} else {
			port = ":4000"
		}
	}

	// 确保端口以冒号开头
	if port != "" && port[0] != ':' {
		port = ":" + port
	}

	AppConfig = &App{
		AppMode:  getEnv("APP_MODE", "debug"),
		HttpPort: port,
	}
	log.Printf("✓ App config loaded (Mode: %s, Port: %s)", AppConfig.AppMode, AppConfig.HttpPort)
}

// loadDatabaseConfig 加载数据库配置
func loadDatabaseConfig() {
	DatabaseConfig = &Database{
		Db:         getEnv("DB_TYPE", "postgres"),
		DbHost:     getEnv("DB_HOST", "localhost"),
		DbPort:     getEnv("DB_PORT", "5432"),
		DbUser:     getEnv("DB_USER", "postgres"),
		DbPassWord: getEnv("DB_PASSWORD", ""),
		DbName:     getEnv("DB_NAME", ""),
		Charset:    getEnv("DB_CHARSET", "utf8"),
	}

	if DatabaseConfig.DbName == "" {
		log.Fatal("✗ DB_NAME is required in environment variables")
	}

	log.Printf("✓ Database config loaded (Type: %s, Database: %s)", DatabaseConfig.Db, DatabaseConfig.DbName)
}

// loadTongyiConfig 加载通义API配置
func loadTongyiConfig() {
	TongyiConfig = &Tongyi{
		APIKey:     getEnv("TONGYI_API_KEY", ""),
		ImageModel: getEnv("TONGYI_IMAGE_MODEL", "wanx-v1"),
		LLMModel:   getEnv("TONGYI_LLM_MODEL", "qwen-turbo"),
	}

	if TongyiConfig.APIKey == "" {
		log.Fatal("✗ TONGYI_API_KEY is required in environment variables")
	}

	log.Printf("✓ Tongyi config loaded (Model: %s)", TongyiConfig.ImageModel)
}

// loadQiniuConfig 加载七牛云配置
func loadQiniuConfig() {
	QiniuConfig = &Qiniu{
		AccessKey:         getEnv("QINIU_ACCESS_KEY", ""),
		SecretKey:         getEnv("QINIU_SECRET_KEY", ""),
		Bucket:            getEnv("QINIU_BUCKET", "ads-creative-gen-platform"),
		Domain:            getEnv("QINIU_DOMAIN", ""),
		PublicCloudDomain: getEnv("QINIU_PUBLIC_CLOUD_DOMAIN", ""), // 新增：公共云访问域名
		Region:            getEnv("QINIU_REGION", "cn-south-1"),
		BasePath:          getEnv("QINIU_BASE_PATH", "s3/"),
	}

	if QiniuConfig.AccessKey == "" || QiniuConfig.SecretKey == "" {
		log.Println("⚠ Qiniu credentials not configured, image upload will be disabled")
		log.Println("💡 To enable Qiniu storage, set QINIU_ACCESS_KEY and QINIU_SECRET_KEY in your .env file")
		log.Println("💡 Also recommend setting QINIU_DOMAIN for custom domain access")
		return
	}

	log.Printf("✓ Qiniu config loaded (Bucket: %s, Region: %s)", QiniuConfig.Bucket, QiniuConfig.Region)

	if QiniuConfig.Domain == "" {
		log.Println("💡 QINIU_DOMAIN is not set, using default S3 domain format")
		log.Printf("💡 To set custom domain, configure CNAME for %s.s3.%s.qiniucs.com", QiniuConfig.Bucket, QiniuConfig.Region)
	}

	log.Println("💡 IMPORTANT: For public access, ensure your Qiniu bucket is set to 'Public Read' in Qiniu Console")
	log.Println("💡 If using 'Private' bucket, images will require authentication and may not be accessible")
}

// loadCacheConfig 加载缓存配置
func loadCacheConfig() {
	ttlSeconds := parseInt("CACHE_DEFAULT_TTL_SECONDS", 300)
	if ttlSeconds < 0 {
		ttlSeconds = 0
	}
	CacheConfig = &Cache{
		Enabled:           parseBool("CACHE_ENABLED", true),
		MaxEntries:        parseInt("CACHE_MAX_ENTRIES", 5000),
		DefaultTTL:        time.Duration(ttlSeconds) * time.Second,
		DisableExperiment: parseBool("CACHE_DISABLE_EXPERIMENT", false),
		DisableCreative:   parseBool("CACHE_DISABLE_CREATIVE", false),
		DisableTracing:    parseBool("CACHE_DISABLE_TRACING", false),
	}
	log.Printf("✓ Cache config loaded (enabled=%v, max_entries=%d, default_ttl=%s)", CacheConfig.Enabled, CacheConfig.MaxEntries, CacheConfig.DefaultTTL)
}

// GetDatabaseDSN 返回数据库 DSN 连接字符串
func GetDatabaseDSN() string {
	if DatabaseConfig.Db == "postgres" {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=public",
			DatabaseConfig.DbHost,
			DatabaseConfig.DbPort,
			DatabaseConfig.DbUser,
			DatabaseConfig.DbPassWord,
			DatabaseConfig.DbName,
		)
	}
	// MySQL fallback
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		DatabaseConfig.DbUser,
		DatabaseConfig.DbPassWord,
		DatabaseConfig.DbHost,
		DatabaseConfig.DbPort,
		DatabaseConfig.DbName,
		DatabaseConfig.Charset,
	)
}

// getEnv 从环境变量读取，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func parseBool(key string, defaultVal bool) bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if val == "" {
		return defaultVal
	}
	return val == "1" || val == "true" || val == "yes" || val == "on"
}

func parseInt(key string, defaultVal int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	if n, err := strconv.Atoi(val); err == nil {
		return n
	}
	return defaultVal
}
