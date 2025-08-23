package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	AI       AIConfig       `mapstructure:"ai"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Log      LogConfig      `mapstructure:"log"`
	Feishu   FeishuConfig   `mapstructure:"feishu"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Milvus MilvusConfig `mapstructure:"milvus"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	MaxOpenConns int `mapstructure:"max_open_conns"`
}

// MilvusConfig Milvus配置
type MilvusConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// AIConfig AI服务配置
type AIConfig struct {
	Provider string      `mapstructure:"provider"` // volcengine, openai
	APIKey   string      `mapstructure:"api_key"`
	BaseURL  string      `mapstructure:"base_url"`
	Models   ModelsConfig `mapstructure:"models"`
}

// ModelsConfig 模型配置
type ModelsConfig struct {
	Embedding string `mapstructure:"embedding"`
	Chat      string `mapstructure:"chat"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type     string      `mapstructure:"type"` // local, oss, s3
	Local    LocalStorage `mapstructure:"local"`
	MaxFileSize string   `mapstructure:"max_file_size"`
	AllowedExts []string `mapstructure:"allowed_extensions"`
}

// LocalStorage 本地存储配置
type LocalStorage struct {
	UploadPath string `mapstructure:"upload_path"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	File       string `mapstructure:"file"`        // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"` // 备份文件数量
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	BaseURL   string `mapstructure:"base_url"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 设置环境变量前缀
	viper.SetEnvPrefix("AI_DOC")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，使用默认配置
			fmt.Println("⚠️  配置文件未找到，使用默认配置")
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")

	// 数据库默认配置
	viper.SetDefault("database.mysql.host", "localhost")
	viper.SetDefault("database.mysql.port", 3306)
	viper.SetDefault("database.mysql.username", "root")
	viper.SetDefault("database.mysql.password", "password")
	viper.SetDefault("database.mysql.database", "ai_assistant")
	viper.SetDefault("database.mysql.charset", "utf8mb4")
	viper.SetDefault("database.mysql.max_idle_conns", 10)
	viper.SetDefault("database.mysql.max_open_conns", 100)

	viper.SetDefault("database.milvus.host", "localhost")
	viper.SetDefault("database.milvus.port", 19530)
	viper.SetDefault("database.milvus.database", "ai_assistant")

	viper.SetDefault("database.redis.host", "localhost")
	viper.SetDefault("database.redis.port", 6379)
	viper.SetDefault("database.redis.db", 0)
	viper.SetDefault("database.redis.pool_size", 10)

	// AI服务默认配置
	viper.SetDefault("ai.provider", "volcengine")
	viper.SetDefault("ai.base_url", "https://ark.cn-beijing.volces.com/api/v3")
	viper.SetDefault("ai.models.embedding", "doubao-embedding")
	viper.SetDefault("ai.models.chat", "doubao-seed")

	// 存储默认配置
	viper.SetDefault("storage.type", "local")
	viper.SetDefault("storage.local.upload_path", "./uploads")
	viper.SetDefault("storage.max_file_size", "100MB")
	viper.SetDefault("storage.allowed_extensions", []string{".pdf", ".docx", ".txt", ".md"})

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "./logs/app.log")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 5)

	// 飞书默认配置
	viper.SetDefault("feishu.base_url", "https://open.feishu.cn")
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 检查必需的AI API密钥
	if config.AI.APIKey == "" {
		if apiKey := os.Getenv("AI_DOC_AI_API_KEY"); apiKey != "" {
			config.AI.APIKey = apiKey
		} else {
			return fmt.Errorf("AI API密钥未配置，请设置 ai.api_key 或环境变量 AI_DOC_AI_API_KEY")
		}
	}

	// 检查上传目录
	if err := os.MkdirAll(config.Storage.Local.UploadPath, 0755); err != nil {
		return fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 检查日志目录
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	return nil
}

// GetDSN 获取MySQL连接字符串
func (c *MySQLConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

// GetAddr 获取Redis连接地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetMilvusAddr 获取Milvus连接地址
func (c *MilvusConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}