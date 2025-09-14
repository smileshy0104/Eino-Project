package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	arkmodel "github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	//retrieverCm "github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

// 配置结构体
type Config struct {
	APIKey           string `mapstructure:"api_key"`
	Model            string `mapstructure:"model"`
	EmbedderModel    string `mapstructure:"embedder_model"`
	MilvusAddress    string `mapstructure:"milvus_address"`
	MilvusCollection string `mapstructure:"milvus_collection"`
}

// 初始化配置
func initConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig() // 忽略错误，因为我们也会检查环境变量

	return &Config{
		APIKey:           viper.GetString("ARK_API_KEY"),
		Model:            viper.GetString("ARK_MODEL"),
		EmbedderModel:    viper.GetString("EMBEDDER_MODEL"),
		MilvusAddress:    viper.GetString("MILVUS_ADDRESS"),
		MilvusCollection: viper.GetString("MILVUS_COLLECTION"),
	}, nil
}

// 初始化Embedder
func initEmbedder(ctx context.Context, config *Config) (*ark.Embedder, error) {
	timeout := 30 * time.Second
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  config.APIKey,
		Model:   config.EmbedderModel,
		Timeout: &timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化Embedder失败: %w", err)
	}
	return embedder, nil
}

// 初始化Milvus客户端
func initMilvusClient(ctx context.Context, config *Config) (cli.Client, error) {
	client, err := cli.NewClient(ctx, cli.Config{
		Address: config.MilvusAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化Milvus客户端失败: %w", err)
	}
	return client, nil
}

// 初始化Retriever
func initRetriever(ctx context.Context, config *Config, client cli.Client, embedder *ark.Embedder) (*milvus.Retriever, error) {
	cfg := &milvus.RetrieverConfig{
		Client:       client,
		Collection:   config.MilvusCollection,
		VectorField:  "vector",
		Embedding:    embedder,
		OutputFields: []string{"id", "content", "metadata"},
		TopK:         5,
	}

	retriever, err := milvus.NewRetriever(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化Retriever失败: %w", err)
	}
	return retriever, nil
}

// 初始化ChatModel
func initChatModel(ctx context.Context, config *Config) (*arkmodel.ChatModel, error) {
	model, err := arkmodel.NewChatModel(ctx, &arkmodel.ChatModelConfig{
		APIKey: config.APIKey,
		Model:  config.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化ChatModel失败: %w", err)
	}
	return model, nil
}

// 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// 1. 基础检索示例
func basicRetrievalExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 基础检索示例 ===")

	// 初始化组件
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	retriever, err := initRetriever(ctx, config, client, embedder)
	if err != nil {
		log.Printf("初始化Retriever失败: %v", err)
		return
	}

	// 准备查询
	queries := []string{
		"Eino框架是什么？",
		"如何使用向量数据库？",
		"RAG系统的核心组件",
	}

	fmt.Printf("📝 准备进行基础检索演示，测试 %d 个查询\n", len(queries))
	for i, query := range queries {
		fmt.Printf("  查询%d: %s\n", i+1, query)
	}

	// 执行检索
	for i, query := range queries {
		fmt.Printf("\n🔍 执行查询 %d: \"%s\"\n", i+1, query)
		startTime := time.Now()

		docs, err := retriever.Retrieve(ctx, query)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("  ❌ 查询失败: %v\n", err)
			continue
		}

		fmt.Printf("  ✅ 检索成功，耗时: %v，找到 %d 个相关文档\n", duration, len(docs))

		if len(docs) == 0 {
			fmt.Println("    📝 未找到相关文档，请确保已运行indexer_demo填充数据")
			continue
		}

		// 显示检索结果
		for j, doc := range docs {
			fmt.Printf("    📄 文档%d: ID=%s\n", j+1, doc.ID)
			if len(doc.Content) > 100 {
				fmt.Printf("       内容: %s...\n", doc.Content[:100])
			} else {
				fmt.Printf("       内容: %s\n", doc.Content)
			}
			if doc.MetaData != nil {
				fmt.Printf("       元数据: %v\n", doc.MetaData)
			}
		}
	}

	fmt.Println("✅ 基础检索示例完成！")
}

// 2. 批量检索示例
func batchRetrievalExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 批量检索示例 ===")

	// 初始化组件
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	retriever, err := initRetriever(ctx, config, client, embedder)
	if err != nil {
		log.Printf("初始化Retriever失败: %v", err)
		return
	}

	// 准备多个查询
	queries := []string{
		"AI框架的核心特性",
		"向量数据库的应用场景",
		"文本嵌入模型的选择",
		"RAG系统架构设计",
		"大语言模型集成方案",
		"知识库检索优化",
		"语义相似度计算方法",
		"智能问答系统构建",
		"文档索引最佳实践",
		"多模态AI应用开发",
	}

	fmt.Printf("📝 准备批量检索演示，处理 %d 个查询\n", len(queries))

	// 批量处理统计
	startTime := time.Now()
	totalDocs := 0
	successCount := 0

	fmt.Println("\n🚀 开始批量检索...")

	for i, query := range queries {
		queryStartTime := time.Now()
		docs, err := retriever.Retrieve(ctx, query)
		queryDuration := time.Since(queryStartTime)

		if err != nil {
			fmt.Printf("  ❌ 查询%d失败: %v\n", i+1, err)
			continue
		}

		totalDocs += len(docs)
		successCount++

		fmt.Printf("  ✅ 查询%d: 找到%d个文档，耗时:%v\n",
			i+1, len(docs), queryDuration)

		// 显示第一个结果作为示例
		if len(docs) > 0 {
			doc := docs[0]
			if len(doc.Content) > 50 {
				fmt.Printf("     📄 最相关: %s...\n", doc.Content[:50])
			} else {
				fmt.Printf("     📄 最相关: %s\n", doc.Content)
			}
		}

		// 模拟处理间隔
		time.Sleep(100 * time.Millisecond)
	}

	totalDuration := time.Since(startTime)

	// 批量处理统计
	fmt.Println("\n📊 批量检索统计:")
	fmt.Printf("  • 总查询数: %d\n", len(queries))
	fmt.Printf("  • 成功查询: %d\n", successCount)
	fmt.Printf("  • 总文档数: %d\n", totalDocs)
	fmt.Printf("  • 总耗时: %v\n", totalDuration)

	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(successCount)
		fmt.Printf("  • 平均耗时/查询: %v\n", avgDuration)

		if totalDuration.Seconds() > 0 {
			throughput := float64(successCount) / totalDuration.Seconds()
			fmt.Printf("  • 处理吞吐量: %.2f 查询/秒\n", throughput)
		}
	}

	fmt.Println("✅ 批量检索示例完成！")
}

// 3. 高级检索配置示例
func advancedRetrievalExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 高级检索配置示例 ===")

	// 初始化组件
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	// 演示不同的TopK配置
	topKValues := []int{1, 3, 5, 10}
	query := "Eino框架的主要特性"

	fmt.Printf("📝 测试不同TopK值对检索结果的影响\n")
	fmt.Printf("🔍 查询: \"%s\"\n", query)

	for _, topK := range topKValues {
		fmt.Printf("\n🔸 TopK = %d\n", topK)

		// 创建使用不同TopK的Retriever
		cfg := &milvus.RetrieverConfig{
			Client:       client,
			Collection:   config.MilvusCollection,
			VectorField:  "vector",
			Embedding:    embedder,
			OutputFields: []string{"id", "content", "metadata"},
			TopK:         topK,
		}

		retriever, err := milvus.NewRetriever(ctx, cfg)
		if err != nil {
			fmt.Printf("  ❌ 创建Retriever失败: %v\n", err)
			continue
		}

		startTime := time.Now()
		docs, err := retriever.Retrieve(ctx, query)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("  ❌ 检索失败: %v\n", err)
			continue
		}

		fmt.Printf("  ✅ 检索结果: %d个文档，耗时: %v\n", len(docs), duration)

		for i, doc := range docs {
			if len(doc.Content) > 60 {
				fmt.Printf("    %d. %s...\n", i+1, doc.Content[:60])
			} else {
				fmt.Printf("    %d. %s\n", i+1, doc.Content)
			}
		}
	}

	// 演示检索配置概念
	fmt.Println("\n🔸 高级配置选项说明:")
	fmt.Println("  📋 TopK配置:")
	fmt.Println("     • TopK=1: 只返回最相关的文档，精准但信息有限")
	fmt.Println("     • TopK=3-5: 平衡相关性和信息丰富度，适合大多数场景")
	fmt.Println("     • TopK=10+: 提供更多候选文档，适合需要全面信息的场景")

	fmt.Println("\n  🎯 其他重要配置:")
	fmt.Println("     • ScoreThreshold: 设置相似度阈值，过滤低质量结果")
	fmt.Println("     • OutputFields: 指定返回的字段，控制数据传输量")
	fmt.Println("     • VectorField: 指定向量字段名，确保查询正确的向量")
	fmt.Println("     • Collection: 指定查询的集合，支持多集合场景")

	fmt.Println("✅ 高级检索配置示例完成！")
}

// 4. RAG Chain编排示例
func ragChainExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== RAG Chain编排示例 ===")

	// 初始化所有组件
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	retriever, err := initRetriever(ctx, config, client, embedder)
	if err != nil {
		log.Printf("初始化Retriever失败: %v", err)
		return
	}

	chatModel, err := initChatModel(ctx, config)
	if err != nil {
		log.Printf("初始化ChatModel失败: %v", err)
		return
	}

	fmt.Println("🔗 创建RAG Chain工作流...")

	// 使用闭包保存查询状态
	var currentQuery string

	// 创建Prompt构建函数
	createPromptFromDocs := func(ctx context.Context, docs []*schema.Document) ([]*schema.Message, error) {
		if len(docs) == 0 {
			prompt := fmt.Sprintf("背景知识库中没有与\"%s\"相关的信息。请基于你的知识回答问题。", currentQuery)
			messages := []*schema.Message{
				schema.SystemMessage("你是一个知识渊博的AI助手。"),
				schema.UserMessage(prompt),
			}
			return messages, nil
		}

		// 构建包含检索文档的Prompt
		prompt := "请根据以下背景知识来回答问题。\n\n=== 背景知识 ===\n"
		for i, doc := range docs {
			prompt += fmt.Sprintf("【文档%d】%s\n\n", i+1, doc.Content)
		}
		prompt += fmt.Sprintf("=== 问题 ===\n%s", currentQuery)

		messages := []*schema.Message{
			schema.SystemMessage("你是一个严谨的AI助手，请严格根据提供的背景知识回答问题。如果背景知识不足以回答问题，请说明情况。"),
			schema.UserMessage(prompt),
		}
		return messages, nil
	}

	// 创建Chain
	chain := compose.NewChain[string, *schema.Message]()

	// 步骤1: 保存查询并转换为map格式
	chain.AppendLambda(
		compose.InvokableLambda(func(ctx context.Context, query string) (map[string]any, error) {
			currentQuery = query // 保存查询到闭包变量
			return map[string]any{"query": query}, nil
		}),
	)

	// 步骤2: 添加Retriever节点
	chain.AppendRetriever(retriever, compose.WithInputKey("query"))

	// 步骤3: 构建Prompt
	chain.AppendLambda(compose.InvokableLambda(createPromptFromDocs))

	// 步骤4: 生成回答
	chain.AppendChatModel(chatModel)

	fmt.Println("⚙️ 编译Chain工作流...")
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译Chain失败: %v", err)
		return
	}
	fmt.Println("✅ Chain编译成功！")

	// 测试多个查询
	queries := []string{
		"Eino框架是什么？",
		"如何使用向量数据库进行语义检索？",
		"RAG系统的核心优势是什么？",
	}

	fmt.Printf("📝 准备通过RAG Chain处理 %d 个问题\n", len(queries))

	for i, query := range queries {
		fmt.Printf("\n🔍 问题%d: %s\n", i+1, query)
		fmt.Println("🚀 执行RAG Chain工作流...")

		startTime := time.Now()
		result, err := runnable.Invoke(ctx, query)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("❌ RAG Chain执行失败: %v\n", err)
			continue
		}

		fmt.Printf("✅ RAG Chain执行成功，总耗时: %v\n", duration)
		fmt.Println("\n💬 AI回答:")
		fmt.Println("----------------------------------------")
		fmt.Println(result.Content)
		fmt.Println("----------------------------------------")
	}

	fmt.Println("✅ RAG Chain编排示例完成！")
}

// 5. 性能测试示例
func performanceTestExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 性能测试示例 ===")

	// 初始化组件
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	retriever, err := initRetriever(ctx, config, client, embedder)
	if err != nil {
		log.Printf("初始化Retriever失败: %v", err)
		return
	}

	// 生成测试查询
	testQueries := []string{
		"AI人工智能技术发展趋势",
		"机器学习算法优化方法",
		"深度学习框架选择指南",
		"自然语言处理最新进展",
		"计算机视觉应用案例",
		"大数据分析处理技术",
		"云计算架构设计方案",
		"分布式系统设计模式",
		"微服务架构最佳实践",
		"数据库性能优化策略",
	}

	fmt.Printf("📝 准备性能压测，使用 %d 个测试查询\n", len(testQueries))

	// 测试不同的并发场景
	testScenarios := []struct {
		name        string
		queries     []string
		description string
	}{
		{
			name:        "轻载测试",
			queries:     testQueries[:3],
			description: "少量查询，测试基础性能",
		},
		{
			name:        "中载测试",
			queries:     testQueries[:6],
			description: "中等查询量，测试稳定性",
		},
		{
			name:        "重载测试",
			queries:     testQueries,
			description: "大量查询，测试极限性能",
		},
	}

	for _, scenario := range testScenarios {
		fmt.Printf("\n🧪 执行 %s (%s)\n", scenario.name, scenario.description)

		startTime := time.Now()
		totalDocs := 0
		successCount := 0
		var minDuration, maxDuration time.Duration
		var totalDuration time.Duration

		for i, query := range scenario.queries {
			queryStart := time.Now()
			docs, err := retriever.Retrieve(ctx, query)
			queryDuration := time.Since(queryStart)

			totalDuration += queryDuration

			if err != nil {
				fmt.Printf("  ❌ 查询%d失败: %v\n", i+1, err)
				continue
			}

			successCount++
			totalDocs += len(docs)

			// 更新最小最大耗时
			if i == 0 || queryDuration < minDuration {
				minDuration = queryDuration
			}
			if queryDuration > maxDuration {
				maxDuration = queryDuration
			}

			fmt.Printf("  ✅ 查询%d: %d文档, %v\n", i+1, len(docs), queryDuration)
		}

		totalTestTime := time.Since(startTime)

		// 性能统计
		fmt.Printf("\n📊 %s 性能统计:\n", scenario.name)
		fmt.Printf("  • 总查询数: %d\n", len(scenario.queries))
		fmt.Printf("  • 成功查询: %d\n", successCount)
		fmt.Printf("  • 成功率: %.1f%%\n", float64(successCount)/float64(len(scenario.queries))*100)
		fmt.Printf("  • 总文档数: %d\n", totalDocs)
		fmt.Printf("  • 平均文档数/查询: %.1f\n", float64(totalDocs)/float64(successCount))
		fmt.Printf("  • 总测试时间: %v\n", totalTestTime)

		if successCount > 0 {
			avgDuration := totalDuration / time.Duration(successCount)
			fmt.Printf("  • 平均查询耗时: %v\n", avgDuration)
			fmt.Printf("  • 最快查询耗时: %v\n", minDuration)
			fmt.Printf("  • 最慢查询耗时: %v\n", maxDuration)

			if totalTestTime.Seconds() > 0 {
				qps := float64(successCount) / totalTestTime.Seconds()
				fmt.Printf("  • 查询吞吐量: %.2f QPS\n", qps)
			}
		}
	}

	fmt.Println("\n💡 性能优化建议:")
	fmt.Println("  • 🎯 TopK设置: 根据业务需求调整，避免返回过多不必要文档")
	fmt.Println("  • 📊 向量维度: 选择合适的向量维度平衡精度和性能")
	fmt.Println("  • 🔄 连接池: 使用连接池减少连接建立开销")
	fmt.Println("  • 💾 结果缓存: 对热点查询进行结果缓存")
	fmt.Println("  • 🚀 批量处理: 合并相似查询减少网络往返")

	fmt.Println("✅ 性能测试示例完成！")
}

// 6. 错误处理示例
func errorHandlingExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 错误处理示例 ===")

	// 演示1: 配置错误处理
	fmt.Println("🔸 演示1: 配置错误检测")

	// 检查必要配置
	configErrors := []string{}
	if config.APIKey == "" {
		configErrors = append(configErrors, "ARK_API_KEY未设置")
	}
	if config.EmbedderModel == "" {
		configErrors = append(configErrors, "EMBEDDER_MODEL未设置")
	}
	if config.MilvusAddress == "" {
		configErrors = append(configErrors, "MILVUS_ADDRESS未设置")
	}
	if config.MilvusCollection == "" {
		configErrors = append(configErrors, "MILVUS_COLLECTION未设置")
	}

	if len(configErrors) > 0 {
		fmt.Println("  ❌ 发现配置错误:")
		for _, err := range configErrors {
			fmt.Printf("     • %s\n", err)
		}
		fmt.Println("  🔧 请检查config.yaml文件或环境变量设置")
		return
	} else {
		fmt.Println("  ✅ 配置检查通过")
	}

	// 演示2: 组件初始化错误处理
	fmt.Println("\n🔸 演示2: 组件初始化错误处理")

	fmt.Println("  🔄 初始化Embedder...")
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		fmt.Printf("  ❌ Embedder初始化失败: %v\n", err)
		fmt.Println("  🔧 建议检查:")
		fmt.Println("     • API Key是否正确")
		fmt.Println("     • 模型名称是否有效")
		fmt.Println("     • 网络连接是否正常")
		return
	}
	fmt.Println("  ✅ Embedder初始化成功")

	fmt.Println("  🔄 初始化Milvus客户端...")
	client, err := initMilvusClient(ctx, config)
	if err != nil {
		fmt.Printf("  ❌ Milvus客户端初始化失败: %v\n", err)
		fmt.Println("  🔧 建议检查:")
		fmt.Println("     • Milvus服务是否启动")
		fmt.Println("     • 地址和端口是否正确")
		fmt.Println("     • 防火墙设置")
		return
	}
	defer client.Close()
	fmt.Println("  ✅ Milvus客户端初始化成功")

	fmt.Println("  🔄 初始化Retriever...")
	retriever, err := initRetriever(ctx, config, client, embedder)
	if err != nil {
		fmt.Printf("  ❌ Retriever初始化失败: %v\n", err)
		fmt.Println("  🔧 建议检查:")
		fmt.Println("     • 集合是否存在")
		fmt.Println("     • 字段配置是否正确")
		fmt.Println("     • 索引是否已创建")
		return
	}
	fmt.Println("  ✅ Retriever初始化成功")

	// 演示3: 检索错误处理与恢复
	fmt.Println("\n🔸 演示3: 检索错误处理与恢复")

	// 测试不同类型的查询
	testQueries := []struct {
		query       string
		expectError bool
		description string
	}{
		{
			query:       "正常的查询测试",
			expectError: false,
			description: "正常查询",
		},
		{
			query:       "",
			expectError: true,
			description: "空查询",
		},
		{
			query:       strings.Repeat("超长查询内容测试", 1000),
			expectError: true,
			description: "超长查询",
		},
		{
			query:       "Eino框架特性介绍",
			expectError: false,
			description: "正常技术查询",
		},
	}

	for i, test := range testQueries {
		fmt.Printf("\n  🧪 测试%d: %s\n", i+1, test.description)

		// 使用重试机制
		maxRetries := 3
		var docs []*schema.Document
		var lastErr error

		for retry := 0; retry < maxRetries; retry++ {
			if retry > 0 {
				fmt.Printf("    🔄 重试 %d/%d...\n", retry, maxRetries-1)
				time.Sleep(time.Duration(retry) * time.Second) // 递增延迟
			}

			startTime := time.Now()
			docs, lastErr = retriever.Retrieve(ctx, test.query)
			duration := time.Since(startTime)

			if lastErr == nil {
				fmt.Printf("    ✅ 检索成功，找到%d个文档，耗时:%v\n", len(docs), duration)
				break
			}

			fmt.Printf("    ⚠️ 第%d次尝试失败: %v\n", retry+1, lastErr)
		}

		if lastErr != nil {
			fmt.Printf("    ❌ 最终失败: %v\n", lastErr)

			// 错误分类和建议
			errorMsg := lastErr.Error()
			if strings.Contains(errorMsg, "timeout") {
				fmt.Println("    💡 这是超时错误，建议:")
				fmt.Println("       • 增加超时时间")
				fmt.Println("       • 检查网络连接")
				fmt.Println("       • 简化查询内容")
			} else if strings.Contains(errorMsg, "collection") {
				fmt.Println("    💡 这是集合相关错误，建议:")
				fmt.Println("       • 检查集合名称是否正确")
				fmt.Println("       • 确认集合是否存在")
				fmt.Println("       • 运行indexer_demo创建数据")
			} else if strings.Contains(errorMsg, "field") {
				fmt.Println("    💡 这是字段相关错误，建议:")
				fmt.Println("       • 检查OutputFields配置")
				fmt.Println("       • 确认字段名称是否正确")
			} else {
				fmt.Println("    💡 通用错误处理建议:")
				fmt.Println("       • 检查日志获取详细信息")
				fmt.Println("       • 验证输入数据格式")
				fmt.Println("       • 确认服务运行状态")
			}
		}
	}

	fmt.Println("\n💡 错误处理最佳实践:")
	fmt.Println("  • 🔧 预检查: 启动前验证所有必要配置")
	fmt.Println("  • 🔄 重试机制: 对临时错误实施指数退避重试")
	fmt.Println("  • 📝 日志记录: 记录详细错误信息便于排查")
	fmt.Println("  • 🛡️ 降级方案: 准备服务不可用时的备选策略")
	fmt.Println("  • ⏰ 超时控制: 设置合理的超时时间避免无限等待")

	fmt.Println("✅ 错误处理示例完成！")
}

// 7. Option配置示例
func optionConfigExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== Option配置示例 ===")

	// 初始化组件
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	// 基础Retriever配置
	retriever, err := initRetriever(ctx, config, client, embedder)
	if err != nil {
		log.Printf("初始化Retriever失败: %v", err)
		return
	}

	testQuery := "Eino框架的核心特性和优势"

	fmt.Printf("📝 演示Retriever的Option配置功能\n")
	fmt.Printf("🔍 测试查询: \"%s\"\n", testQuery)

	// 演示1: 基础检索（无Option）
	fmt.Println("\n🔸 演示1: 基础检索（使用默认配置）")
	startTime := time.Now()
	docs, err := retriever.Retrieve(ctx, testQuery)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("  ❌ 检索失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 基础检索完成，找到%d个文档，耗时:%v\n", len(docs), duration)

	// 演示2: Option配置概念说明
	fmt.Println("\n🔸 演示2: Retriever Option配置说明")
	fmt.Println("📋 常用Retriever Option配置:")
	fmt.Println("  • retriever.WithTopK(10): 设置返回的最大文档数量")
	fmt.Println("  • retriever.WithScoreThreshold(0.7): 设置相似度阈值")
	fmt.Println("  • retriever.WithMetricType(\"COSINE\"): 设置距离计算方式")
	fmt.Println("  • retriever.WithOutputFields([...]): 指定返回的字段")
	fmt.Println("  • retriever.WithSearchParams({...}): 设置搜索参数")

	// 演示3: TopK参数配置效果对比
	fmt.Println("\n🔸 演示3: TopK参数配置效果对比")
	topKValues := []int{1, 3, 5, 10}

	for _, topK := range topKValues {
		fmt.Printf("\n  📊 TopK = %d 的检索效果:\n", topK)

		// 创建新的Retriever配置
		cfg := &milvus.RetrieverConfig{
			Client:       client,
			Collection:   config.MilvusCollection,
			VectorField:  "vector",
			Embedding:    embedder,
			OutputFields: []string{"id", "content", "metadata"},
			TopK:         topK,
		}

		topKRetriever, err := milvus.NewRetriever(ctx, cfg)
		if err != nil {
			fmt.Printf("    ❌ 创建TopK=%d的Retriever失败: %v\n", topK, err)
			continue
		}

		// 注意：这里演示概念，实际API可能需要根据具体实现调整
		// docs, err := topKRetriever.Retrieve(ctx, testQuery, milvus.WithTopK(topK))
		docs, err := topKRetriever.Retrieve(ctx, testQuery)
		if err != nil {
			fmt.Printf("    ❌ TopK=%d检索失败: %v\n", topK, err)
			continue
		}

		fmt.Printf("    ✅ 返回%d个文档:\n", len(docs))
		for i, doc := range docs {
			if len(doc.Content) > 50 {
				fmt.Printf("      %d. %s...\n", i+1, doc.Content[:50])
			} else {
				fmt.Printf("      %d. %s\n", i+1, doc.Content)
			}
		}
	}

	// 演示4: 不同场景的Option配置模式
	fmt.Println("\n🔸 演示4: 不同场景的Option配置模式")

	fmt.Println("🎯 精准检索场景配置:")
	fmt.Println("   docs, err := retriever.Retrieve(ctx, query,")
	fmt.Println("       retriever.WithTopK(1),")
	fmt.Println("       retriever.WithScoreThreshold(0.9),")
	fmt.Println("   )")
	fmt.Println("   • 特点: 只返回最相关的文档，确保高质量结果")

	fmt.Println("\n📊 探索性检索场景:")
	fmt.Println("   docs, err := retriever.Retrieve(ctx, query,")
	fmt.Println("       retriever.WithTopK(10),")
	fmt.Println("       retriever.WithScoreThreshold(0.5),")
	fmt.Println("   )")
	fmt.Println("   • 特点: 返回更多候选文档，适合信息发现")

	fmt.Println("\n⚡ 高性能检索场景:")
	fmt.Println("   docs, err := retriever.Retrieve(ctx, query,")
	fmt.Println("       retriever.WithTopK(3),")
	fmt.Println("       retriever.WithOutputFields([\"id\", \"content\"]),")
	fmt.Println("   )")
	fmt.Println("   • 特点: 精简返回字段，提高检索速度")

	fmt.Println("\n🔍 调试分析场景:")
	fmt.Println("   docs, err := retriever.Retrieve(ctx, query,")
	fmt.Println("       retriever.WithTopK(5),")
	fmt.Println("       retriever.WithOutputFields([\"id\", \"content\", \"metadata\", \"score\"]),")
	fmt.Println("   )")
	fmt.Println("   • 特点: 返回详细信息，便于结果分析")

	fmt.Println("\n💡 Option配置的优势:")
	fmt.Println("  • 🎛️ 运行时调整: 根据查询类型动态配置参数")
	fmt.Println("  • 🎯 场景适配: 针对不同业务场景优化检索策略")
	fmt.Println("  • ⚡ 性能控制: 平衡检索质量和响应速度")
	fmt.Println("  • 🔧 灵活配置: 无需重新创建组件即可调整行为")
	fmt.Println("  • 📊 结果控制: 精确控制返回数据的数量和格式")

	fmt.Println("✅ Option配置示例完成！")
}

// 8. Callback机制示例
func callbackExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== Callback机制示例 ===")

	// 初始化组件
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	retriever, err := initRetriever(ctx, config, client, embedder)
	if err != nil {
		log.Printf("初始化Retriever失败: %v", err)
		return
	}

	// 准备测试查询
	testQueries := []string{
		"Retriever组件是Eino框架中负责信息检索的核心组件",
		"通过Callback机制可以实现检索过程的全程监控和日志记录",
		"在RAG系统中，Retriever扮演着知识获取的重要角色",
	}

	fmt.Printf("📝 演示Retriever的Callback机制功能，处理 %d 个查询\n", len(testQueries))
	for i, query := range testQueries {
		fmt.Printf("  查询%d: %s\n", i+1, query)
	}

	// 演示1: 手动模拟回调事件（概念演示）
	fmt.Println("\n🔸 演示1: Callback机制概念说明")

	for i, query := range testQueries {
		fmt.Printf("\n🔍 处理查询 %d: \"%s\"\n", i+1, query)

		// 模拟OnStart回调
		fmt.Println("  🔄 模拟OnStart回调...")
		fmt.Printf("     🚀 开始检索操作\n")
		fmt.Printf("     📝 查询内容: %s\n", query)
		fmt.Printf("     🎯 目标集合: %s\n", config.MilvusCollection)

		// 记录开始时间
		startTime := time.Now()

		// 执行实际检索
		fmt.Println("  ⚡ 执行检索...")
		docs, err := retriever.Retrieve(ctx, query)

		duration := time.Since(startTime)

		if err != nil {
			// 模拟OnError回调
			fmt.Println("  🔴 模拟OnError回调...")
			fmt.Printf("     ❌ 检索失败，耗时: %v\n", duration)
			fmt.Printf("     💥 错误详情: %v\n", err)
			fmt.Println("     🔧 建议检查:")
			fmt.Println("        • Milvus服务状态")
			fmt.Println("        • 集合和索引配置")
			fmt.Println("        • 网络连接状况")
			fmt.Println("        • 查询参数设置")
			continue
		}

		// 模拟OnEnd回调
		fmt.Println("  🟢 模拟OnEnd回调...")
		fmt.Printf("     ✅ 检索完成，总耗时: %v\n", duration)
		fmt.Printf("     📊 成功检索到 %d 个相关文档\n", len(docs))

		if len(docs) > 0 {
			// 计算一些统计信息
			totalContentLength := 0
			for _, doc := range docs {
				totalContentLength += len(doc.Content)
			}
			avgContentLength := totalContentLength / len(docs)
			fmt.Printf("     📈 平均文档长度: %d 字符\n", avgContentLength)

			// 显示第一个文档作为示例
			firstDoc := docs[0]
			if len(firstDoc.Content) > 60 {
				fmt.Printf("     🏆 最相关文档: %s...\n", firstDoc.Content[:60])
			} else {
				fmt.Printf("     🏆 最相关文档: %s\n", firstDoc.Content)
			}
		}

		if duration.Seconds() > 0 {
			qps := 1.0 / duration.Seconds()
			fmt.Printf("     🚀 当前查询速率: %.2f QPS\n", qps)
		}
	}

	// 演示2: Callback配置代码示例
	fmt.Println("\n🔸 演示2: Callback配置代码示例")
	fmt.Println("📋 标准Callback处理器创建代码:")
	fmt.Println("   callbackHandler := callbacks.NewHandlerBuilder().")
	fmt.Println("       OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {")
	fmt.Println("           fmt.Printf(\"🚀 开始检索操作\\n\")")
	fmt.Println("           if query, ok := input.(string); ok {")
	fmt.Println("               fmt.Printf(\"📝 查询内容: %s\\n\", query)")
	fmt.Println("           }")
	fmt.Println("           return context.WithValue(ctx, \"start_time\", time.Now())")
	fmt.Println("       }).")
	fmt.Println("       OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) {")
	fmt.Println("           startTime, _ := ctx.Value(\"start_time\").(time.Time)")
	fmt.Println("           fmt.Printf(\"✅ 检索完成，耗时: %v\\n\", time.Since(startTime))")
	fmt.Println("           if docs, ok := output.([]*schema.Document); ok {")
	fmt.Println("               fmt.Printf(\"📊 检索到 %d 个文档\\n\", len(docs))")
	fmt.Println("           }")
	fmt.Println("       }).")
	fmt.Println("       OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) {")
	fmt.Println("           fmt.Printf(\"❌ 检索失败: %v\\n\", err)")
	fmt.Println("       }).")
	fmt.Println("       Build()")

	fmt.Println("\n📋 在Chain中使用Callback:")
	fmt.Println("   chain := compose.NewChain[string, []*schema.Document]()")
	fmt.Println("   chain.AppendRetriever(retriever, compose.WithCallbacks(callbackHandler))")
	fmt.Println("   runnable, err := chain.Compile(ctx)")
	fmt.Println("   docs, err := runnable.Invoke(ctx, query)")

	// 演示3: 高级Callback应用场景
	fmt.Println("\n🔸 演示3: 高级Callback应用场景")

	fmt.Println("📊 性能监控Callback:")
	fmt.Println("   - 记录每次检索的响应时间和吞吐量")
	fmt.Println("   - 统计不同查询类型的成功率")
	fmt.Println("   - 监控返回文档数量的分布情况")
	fmt.Println("   - 跟踪向量相似度计算性能")

	fmt.Println("\n🔍 调试分析Callback:")
	fmt.Println("   - 记录查询向量化的详细过程")
	fmt.Println("   - 输出检索参数和中间结果")
	fmt.Println("   - 分析文档匹配度和排序逻辑")
	fmt.Println("   - 追踪异常查询的详细堆栈")

	fmt.Println("\n💰 成本控制Callback:")
	fmt.Println("   - 统计Embedding API调用次数和成本")
	fmt.Println("   - 监控Milvus查询资源消耗")
	fmt.Println("   - 记录数据传输量和存储占用")
	fmt.Println("   - 成本预算和告警机制")

	fmt.Println("\n🛡️ 安全审计Callback:")
	fmt.Println("   - 记录所有检索操作的用户和时间")
	fmt.Println("   - 监控异常查询模式和频率")
	fmt.Println("   - 检测潜在的数据泄露风险")
	fmt.Println("   - 合规性检查和报告生成")

	fmt.Println("\n🎯 Callback机制的优势:")
	fmt.Println("  • 📈 可观测性: 全面监控检索过程的每个环节")
	fmt.Println("  • 🔧 非侵入式: 不修改核心逻辑即可扩展功能")
	fmt.Println("  • 🎛️ 灵活配置: 按需启用不同类型的回调处理")
	fmt.Println("  • 🌐 统一接口: 所有Retriever组件使用相同的回调机制")
	fmt.Println("  • ⏱️ 生命周期: 覆盖开始、结束、错误等关键节点")
	fmt.Println("  • 📊 数据洞察: 深入了解检索性能和质量")

	fmt.Println("\n📋 常见Callback应用场景:")
	fmt.Println("  • 📈 性能分析: 识别检索瓶颈和优化机会")
	fmt.Println("  • 📝 操作审计: 记录所有检索操作历史")
	fmt.Println("  • ⚠️ 异常告警: 及时发现和响应检索问题")
	fmt.Println("  • 🎯 质量监控: 评估检索结果的相关性")
	fmt.Println("  • 🧪 A/B测试: 对比不同检索策略的效果")
	fmt.Println("  • 🔄 自动重试: 在失败时实施智能重试策略")

	fmt.Println("✅ Callback机制示例完成！")
}

// 主函数
func main() {
	ctx := context.Background()

	fmt.Println("🎯 Eino Retriever 组件完全示例")
	fmt.Println("====================================")

	// 1. 初始化配置
	config, err := initConfig()
	if err != nil {
		log.Fatal("配置初始化失败:", err)
	}

	fmt.Printf("使用配置:\n")
	if len(config.APIKey) > 10 {
		fmt.Printf("  API Key: %s\n", config.APIKey[:10]+"...")
	} else {
		fmt.Printf("  API Key: %s\n", config.APIKey)
	}
	fmt.Printf("  模型: %s\n", config.Model)
	fmt.Printf("  Embedder模型: %s\n", config.EmbedderModel)
	fmt.Printf("  Milvus地址: %s\n", config.MilvusAddress)
	fmt.Printf("  Milvus集合: %s\n", config.MilvusCollection)

	// 3. 运行示例
	try := func(name string, fn func(context.Context, *Config)) {
		fmt.Printf("\n正在运行: %s\n", name)
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("示例 %s 运行出错: %v\n", name, r)
			}
		}()
		fn(ctx, config)
	}

	// 检查命令行参数
	if len(os.Args) > 1 {
		exampleName := os.Args[1]
		switch strings.ToLower(exampleName) {
		case "basic":
			try("基础检索示例", basicRetrievalExample)
		case "batch":
			try("批量检索示例", batchRetrievalExample)
		case "advanced":
			try("高级检索配置示例", advancedRetrievalExample)
		case "rag":
			try("RAG Chain编排示例", ragChainExample)
		case "performance":
			try("性能测试示例", performanceTestExample)
		case "error":
			try("错误处理示例", errorHandlingExample)
		case "option":
			try("Option配置示例", optionConfigExample)
		case "callback":
			try("Callback机制示例", callbackExample)
		default:
			fmt.Printf("未知示例: %s\n", exampleName)
			fmt.Println("可用示例: basic, batch, advanced, rag, performance, error, option, callback")
			return
		}
	} else {
		// 运行所有示例
		try("基础检索示例", basicRetrievalExample)
		try("批量检索示例", batchRetrievalExample)
		try("高级检索配置示例", advancedRetrievalExample)
		try("RAG Chain编排示例", ragChainExample)
		try("性能测试示例", performanceTestExample)
		try("错误处理示例", errorHandlingExample)
		try("Option配置示例", optionConfigExample)
		try("Callback机制示例", callbackExample)
	}

	fmt.Println("\n🎉 所有示例运行完成！")
	fmt.Println("\n使用方法:")
	fmt.Println("  go run main.go              # 运行所有示例")
	fmt.Println("  go run main.go basic        # 运行基础检索示例")
	fmt.Println("  go run main.go batch        # 运行批量检索示例")
	fmt.Println("  go run main.go advanced     # 运行高级检索配置示例")
	fmt.Println("  go run main.go rag          # 运行RAG Chain编排示例")
	fmt.Println("  go run main.go performance  # 运行性能测试示例")
	fmt.Println("  go run main.go error        # 运行错误处理示例")
	fmt.Println("  go run main.go option       # 运行Option配置示例")
	fmt.Println("  go run main.go callback     # 运行Callback机制示例")
}
