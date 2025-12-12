package config

import (
	"fmt"
	"log"

	"gopkg.in/ini.v1"
)

// 全局配置对象
var (
	AppConfig      *App
	MySQLConfig    *MySQL
	RabbitMQConfig *RabbitMQ
	EtcdConfig     *Etcd
	TongyiConfig   *Tongyi
)

// App 服务配置
type App struct {
	AppMode  string
	HttpPort string
}

// MySQL 数据库配置
type MySQL struct {
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

// LoadConfig 加载所有配置
func LoadConfig() {
	// 读取 ini 配置文件
	cfg, err := ini.Load("./config/config.ini")
	if err != nil {
		log.Fatalf("Failed to load config.ini: %v", err)
	}

	// 加载各模块配置
	loadAppConfig(cfg)
	loadMySQLConfig(cfg)
	loadRabbitMQConfig(cfg)
	loadEtcdConfig(cfg)
	loadTongyiConfig()

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

// loadMySQLConfig 加载MySQL配置
func loadMySQLConfig(cfg *ini.File) {
	MySQLConfig = &MySQL{
		Db:         cfg.Section("mysql").Key("Db").String(),
		DbHost:     cfg.Section("mysql").Key("DbHost").String(),
		DbPort:     cfg.Section("mysql").Key("DbPort").String(),
		DbUser:     cfg.Section("mysql").Key("DbUser").String(),
		DbPassWord: cfg.Section("mysql").Key("DbPassWord").String(),
		DbName:     cfg.Section("mysql").Key("DbName").String(),
		Charset:    cfg.Section("mysql").Key("Charset").String(),
	}

	if MySQLConfig.DbName == "" {
		log.Fatal("✗ MySQL DbName is required in config.ini")
	}

	log.Printf("✓ MySQL config loaded (Database: %s)", MySQLConfig.DbName)
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

// GetMySQLDSN 返回 MySQL DSN 连接字符串
func GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		MySQLConfig.DbUser,
		MySQLConfig.DbPassWord,
		MySQLConfig.DbHost,
		MySQLConfig.DbPort,
		MySQLConfig.DbName,
		MySQLConfig.Charset,
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
	// 这里需要先加载 .env 文件
	// 可以使用 godotenv 包
	return defaultValue
}
