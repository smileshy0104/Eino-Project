package main

import (
	"context"
	"log"
	"os"

	"github.com/spf13/viper"

	"ai-doc-assistant/internal/config"
	"ai-doc-assistant/internal/service"
)

func main() {
	log.Println("🚀 启动AI文档助手演示")

	// 加载配置
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	log.Printf("✓ 配置加载成功")
	log.Printf("  Milvus: %s:%d", cfg.Database.Milvus.Host, cfg.Database.Milvus.Port)
	log.Printf("  AI模型: %s / %s", cfg.AI.Models.Embedding, cfg.AI.Models.Chat)

	ctx := context.Background()

	// 创建Eino服务
	einoService, err := service.NewEinoService(cfg)
	if err != nil {
		log.Fatalf("❌ Eino服务初始化失败: %v", err)
	}
	defer einoService.Close()

	log.Println("✅ AI文档助手初始化完成")

	// 健康检查
	if err := einoService.HealthCheck(ctx); err != nil {
		log.Printf("⚠️  健康检查失败: %v", err)
	} else {
		log.Println("✅ 系统健康检查通过")
	}

	// 演示文档处理
	err = demonstrateDocumentProcessing(ctx, einoService)
	if err != nil {
		log.Printf("❌ 文档处理演示失败: %v", err)
	}

	// 演示智能问答
	err = demonstrateSmartQA(ctx, einoService)
	if err != nil {
		log.Printf("❌ 智能问答演示失败: %v", err)
	}

	log.Println("🎉 AI文档助手演示完成！")
}

// demonstrateDocumentProcessing 演示文档处理功能
func demonstrateDocumentProcessing(ctx context.Context, einoService *service.EinoService) error {
	log.Println("\n=== 演示文档处理功能 ===")

	// 准备示例文档
	sampleDoc := &service.Document{
		ID:      "demo-doc-001",
		Title:   "AI文档助手使用指南",
		Content: `# AI文档助手使用指南

## 1. 功能介绍
AI文档助手基于Eino框架构建，提供智能文档问答功能。

## 2. 核心特性
- 文档智能分析和索引
- 自然语言问答
- 语义检索
- 多格式文档支持

## 3. 使用方法
1. 上传文档到系统
2. 系统自动进行向量化处理
3. 用户可以使用自然语言提问
4. 系统基于文档内容提供准确回答

## 4. 技术架构
- 使用火山方舟进行文本向量化
- Milvus作为向量数据库
- 基于RAG技术的智能问答`,
		Author:       "AI助手",
		Department:   "技术部",
		DocumentType: "指南",
		Version:      "v1.0",
		CreatedAt:    "2024-01-01T10:00:00Z",
	}

	// 处理文档
	if err := einoService.ProcessDocument(ctx, sampleDoc); err != nil {
		return err
	}

	log.Println("✓ 示例文档处理完成")
	return nil
}

// demonstrateSmartQA 演示智能问答功能
func demonstrateSmartQA(ctx context.Context, einoService *service.EinoService) error {
	log.Println("\n=== 演示智能问答功能 ===")

	// 准备测试问题
	questions := []string{
		"AI文档助手有什么核心特性？",
		"如何使用这个系统？",
		"技术架构是什么样的？",
	}

	for i, question := range questions {
		log.Printf("\n--- 问题 %d ---", i+1)
		log.Printf("Q: %s", question)

		// 执行问答
		response, err := einoService.QueryKnowledge(ctx, question)
		if err != nil {
			log.Printf("❌ 问答失败: %v", err)
			continue
		}

		log.Printf("A: %s", response.Answer)
		log.Printf("响应时间: %dms", response.ResponseTime)
		log.Printf("置信度: %.2f", response.Confidence)
		log.Printf("源文档数量: %d", len(response.Sources))
	}

	return nil
}

// loadConfig 加载配置
func loadConfig() (*config.Config, error) {
	// 设置配置文件
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../../config")
	viper.AddConfigPath(".")

	// 设置环境变量前缀
	viper.SetEnvPrefix("AI_DOC")
	viper.AutomaticEnv()

	// 设置默认值
	viper.SetDefault("database.milvus.host", "localhost")
	viper.SetDefault("database.milvus.port", 19530)
	viper.SetDefault("database.milvus.database", "ai_assistant")
	viper.SetDefault("ai.provider", "volcengine")
	viper.SetDefault("ai.models.embedding", "doubao-embedding")
	viper.SetDefault("ai.models.chat", "doubao-seed")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("⚠️  配置文件未找到，使用默认配置和环境变量")
		} else {
			return nil, err
		}
	}

	// 检查必需的环境变量
	apiKey := viper.GetString("ai.api_key")
	if apiKey == "" || apiKey == "your-volcengine-api-key-here" {
		if envKey := os.Getenv("AI_DOC_AI_API_KEY"); envKey != "" {
			viper.Set("ai.api_key", envKey)
		} else {
			log.Println("⚠️  API密钥未设置，请设置环境变量 AI_DOC_AI_API_KEY")
			log.Println("   export AI_DOC_AI_API_KEY=your-api-key")
		}
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}