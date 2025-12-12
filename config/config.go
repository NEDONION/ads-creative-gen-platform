package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/ini.v1"
)

// 全局配置对象
var (
	AppConfig      *App
	DatabaseConfig *Database
	RabbitMQConfig *RabbitMQ
	EtcdConfig     *Etcd
	TongyiConfig   *Tongyi
	QiniuConfig    *Qiniu
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

// RabbitMQ 配置
type RabbitMQ struct {
	RabbitMQ         string
	RabbitMQUser     string
	RabbitMQPassWord string
	RabbitMQHost     string
	RabbitMQPort     string
}

// Etcd 配置
type Etcd struct {
	EtcdHost string
	EtcdPort string
}

// Tongyi 通义API配置（从环境变量读取）
type Tongyi struct {
	APIKey     string
	ImageModel string
	LLMModel   string
}

// Qiniu 七牛云配置
type Qiniu struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
	Region    string
	BasePath  string
}

// LoadConfig 加载所有配置
func LoadConfig() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 读取 ini 配置文件
	cfg, err := ini.Load("./config/config.ini")
	if err != nil {
		log.Fatalf("Failed to load config.ini: %v", err)
	}

	// 加载各模块配置
	loadAppConfig(cfg)
	loadDatabaseConfig(cfg)
	loadRabbitMQConfig(cfg)
	loadEtcdConfig(cfg)
	loadTongyiConfig()
	loadQiniuConfig()

	log.Println("✓ All configurations loaded successfully")
}

// loadAppConfig 加载服务配置
func loadAppConfig(cfg *ini.File) {
	AppConfig = &App{
		AppMode:  cfg.Section("service").Key("AppMode").String(),
		HttpPort: cfg.Section("service").Key("HttpPort").String(),
	}
	log.Printf("✓ App config loaded (Mode: %s, Port: %s)", AppConfig.AppMode, AppConfig.HttpPort)
}

// loadDatabaseConfig 加载数据库配置
func loadDatabaseConfig(cfg *ini.File) {
	DatabaseConfig = &Database{
		Db:         cfg.Section("database").Key("Db").String(),
		DbHost:     cfg.Section("database").Key("DbHost").String(),
		DbPort:     cfg.Section("database").Key("DbPort").String(),
		DbUser:     cfg.Section("database").Key("DbUser").String(),
		DbPassWord: cfg.Section("database").Key("DbPassWord").String(),
		DbName:     cfg.Section("database").Key("DbName").String(),
		Charset:    cfg.Section("database").Key("Charset").String(),
	}

	if DatabaseConfig.DbName == "" {
		log.Fatal("✗ Database DbName is required in config.ini")
	}

	log.Printf("✓ Database config loaded (Type: %s, Database: %s)", DatabaseConfig.Db, DatabaseConfig.DbName)
}

// loadRabbitMQConfig 加载RabbitMQ配置
func loadRabbitMQConfig(cfg *ini.File) {
	RabbitMQConfig = &RabbitMQ{
		RabbitMQ:         cfg.Section("rabbitmq").Key("RabbitMQ").String(),
		RabbitMQUser:     cfg.Section("rabbitmq").Key("RabbitMQUser").String(),
		RabbitMQPassWord: cfg.Section("rabbitmq").Key("RabbitMQPassWord").String(),
		RabbitMQHost:     cfg.Section("rabbitmq").Key("RabbitMQHost").String(),
		RabbitMQPort:     cfg.Section("rabbitmq").Key("RabbitMQPort").String(),
	}
	log.Printf("✓ RabbitMQ config loaded")
}

// loadEtcdConfig 加载Etcd配置
func loadEtcdConfig(cfg *ini.File) {
	EtcdConfig = &Etcd{
		EtcdHost: cfg.Section("etcd").Key("EtcdHost").String(),
		EtcdPort: cfg.Section("etcd").Key("EtcdPort").String(),
	}
	log.Printf("✓ Etcd config loaded")
}

// loadTongyiConfig 加载通义API配置（从环境变量）
func loadTongyiConfig() {
	// 先尝试从 .env 加载
	TongyiConfig = &Tongyi{
		APIKey:     getEnv("TONGYI_API_KEY", ""),
		ImageModel: getEnv("TONGYI_IMAGE_MODEL", "wanx-v1"),
		LLMModel:   getEnv("TONGYI_LLM_MODEL", "qwen-turbo"),
	}

	if TongyiConfig.APIKey == "" {
		log.Fatal("✗ TONGYI_API_KEY is required in .env file")
	}

	log.Printf("✓ Tongyi config loaded (Model: %s)", TongyiConfig.ImageModel)
}

// loadQiniuConfig 加载七牛云配置
func loadQiniuConfig() {
	QiniuConfig = &Qiniu{
		AccessKey: getEnv("QINIU_ACCESS_KEY", ""),
		SecretKey: getEnv("QINIU_SECRET_KEY", ""),
		Bucket:    getEnv("QINIU_BUCKET", "ads-creative-gen-platform"),
		Domain:    getEnv("QINIU_DOMAIN", ""),
		Region:    getEnv("QINIU_REGION", "cn-south-1"),
		BasePath:  getEnv("QINIU_BASE_PATH", "s3/"),
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

// GetDatabaseDSN 返回数据库 DSN 连接字符串
func GetDatabaseDSN() string {
	if DatabaseConfig.Db == "postgres" {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
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

// GetRabbitMQURL 返回 RabbitMQ 连接 URL
func GetRabbitMQURL() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/",
		RabbitMQConfig.RabbitMQ,
		RabbitMQConfig.RabbitMQUser,
		RabbitMQConfig.RabbitMQPassWord,
		RabbitMQConfig.RabbitMQHost,
		RabbitMQConfig.RabbitMQPort,
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
