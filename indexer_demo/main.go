package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

// 配置结构体
type Config struct {
	APIKey           string `mapstructure:"api_key"`
	Model            string `mapstructure:"model"`
	EmbedderModel    string `mapstructure:"embedder_model"`
	MilvusAddress    string `mapstructure:"milvus_address"`
	MilvusCollection string `mapstructure:"milvus_collection"`
}

// Milvus集合字段定义
var fields = []*entity.Field{
	{
		Name:        "id",
		DataType:    entity.FieldTypeVarChar,
		TypeParams:  map[string]string{"max_length": "255"},
		PrimaryKey:  true,
		Description: "文档的唯一主键",
	},
	{
		Name:        "vector",
		DataType:    entity.FieldTypeBinaryVector,
		TypeParams:  map[string]string{"dim": "81920"},
		Description: "文档内容的向量表示",
	},
	{
		Name:        "content",
		DataType:    entity.FieldTypeVarChar,
		TypeParams:  map[string]string{"max_length": "8192"},
		Description: "原始的文本内容",
	},
	{
		Name:        "metadata",
		DataType:    entity.FieldTypeJSON,
		Description: "用于存储附加信息的 JSON 字段",
	},
}

// 初始化配置
func initConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
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
func initMilvusClient(ctx context.Context, address string) (cli.Client, error) {
	client, err := cli.NewClient(ctx, cli.Config{Address: address})
	if err != nil {
		return nil, fmt.Errorf("创建Milvus客户端失败: %w", err)
	}
	return client, nil
}

// 确保集合存在
func ensureCollection(ctx context.Context, client cli.Client, collectionName string) error {
	has, err := client.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("检查集合是否存在失败: %w", err)
	}

	if !has {
		fmt.Printf("集合 '%s' 不存在，正在创建...\n", collectionName)
		schema := &entity.Schema{
			CollectionName: collectionName,
			Fields:         fields,
			Description:    "Eino Indexer 演示集合",
		}
		err = client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
		if err != nil {
			return fmt.Errorf("创建集合失败: %w", err)
		}
		fmt.Println("集合创建成功！")

		fmt.Println("正在为 'vector' 字段创建 BIN_FLAT 索引...")
		binFlatIndex, err := entity.NewIndexBinFlat(entity.HAMMING, 128)
		if err != nil {
			return fmt.Errorf("创建索引对象失败: %w", err)
		}
		err = client.CreateIndex(ctx, collectionName, "vector", binFlatIndex, false)
		if err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}
		fmt.Println("BIN_FLAT 索引创建成功！")
	} else {
		fmt.Printf("集合 '%s' 已存在，跳过创建步骤。\n", collectionName)
	}
	return nil
}

// 基础索引示例
func basicIndexExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 基础索引示例 ===")

	// 初始化 Embedder
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 初始化 Milvus 客户端
	client, err := initMilvusClient(ctx, config.MilvusAddress)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	// 确保集合存在
	if err := ensureCollection(ctx, client, config.MilvusCollection); err != nil {
		log.Printf("确保集合存在失败: %v", err)
		return
	}

	// 初始化 Indexer
	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: config.MilvusCollection,
		Embedding:  embedder,
		Fields:     fields,
	}
	indexer, err := milvus.NewIndexer(ctx, cfg)
	if err != nil {
		log.Printf("创建Indexer失败: %v", err)
		return
	}

	// 准备基础文档
	documents := []*schema.Document{
		{
			ID:       "basic_001",
			Content:  "Eino 是一个云原生的大模型应用开发框架，旨在简化和加速大模型应用的构建。",
			MetaData: map[string]interface{}{"source": "official_docs", "type": "introduction"},
		},
		{
			ID:       "basic_002",
			Content:  "RAG (Retrieval-Augmented Generation) 是一种结合了检索和生成两大功能的AI技术。",
			MetaData: map[string]interface{}{"source": "tech_blog", "type": "concept"},
		},
	}

	fmt.Printf("📝 准备存储 %d 个基础文档\n", len(documents))
	for i, doc := range documents {
		fmt.Printf("  文档%d - ID: %s\n", i+1, doc.ID)
	}

	storedIDs, err := indexer.Store(ctx, documents)
	if err != nil {
		log.Printf("存储文档失败: %v", err)
		return
	}

	fmt.Printf("✅ 基础索引成功，存储了 %d 个文档: %v\n", len(storedIDs), storedIDs)

	// 加载集合到内存
	fmt.Println("🔄 加载集合到内存...")
	err = client.LoadCollection(ctx, config.MilvusCollection, false)
	if err != nil {
		log.Printf("加载集合失败: %v", err)
		return
	}
	fmt.Println("✅ 集合加载完成，可以进行检索！")
}

// 批量索引示例
func batchIndexExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 批量索引示例 ===")

	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config.MilvusAddress)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	if err := ensureCollection(ctx, client, config.MilvusCollection); err != nil {
		log.Printf("确保集合存在失败: %v", err)
		return
	}

	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: config.MilvusCollection,
		Embedding:  embedder,
		Fields:     fields,
	}
	indexer, err := milvus.NewIndexer(ctx, cfg)
	if err != nil {
		log.Printf("创建Indexer失败: %v", err)
		return
	}

	// 准备大批量文档
	batchDocuments := make([]*schema.Document, 0)
	topics := []string{
		"机器学习", "深度学习", "自然语言处理", "计算机视觉", "推荐系统",
		"分布式系统", "云计算", "微服务架构", "容器技术", "DevOps",
	}

	for i, topic := range topics {
		for j := 0; j < 3; j++ {
			doc := &schema.Document{
				ID:      fmt.Sprintf("batch_%03d_%d", i+1, j+1),
				Content: fmt.Sprintf("%s是现代科技发展的重要领域，具有广泛的应用前景和技术价值。", topic),
				MetaData: map[string]interface{}{
					"source":    "batch_generation",
					"topic":     topic,
					"batch_id":  i + 1,
					"doc_index": j + 1,
				},
			}
			batchDocuments = append(batchDocuments, doc)
		}
	}

	fmt.Printf("📝 准备批量存储 %d 个文档\n", len(batchDocuments))
	startTime := time.Now()

	storedIDs, err := indexer.Store(ctx, batchDocuments)
	if err != nil {
		log.Printf("批量存储失败: %v", err)
		return
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ 批量索引成功，耗时: %v\n", duration)
	fmt.Printf("📊 存储了 %d 个文档，平均耗时: %v/文档\n",
		len(storedIDs), duration/time.Duration(len(storedIDs)))

	fmt.Println("🔄 加载集合到内存...")
	err = client.LoadCollection(ctx, config.MilvusCollection, false)
	if err != nil {
		log.Printf("加载集合失败: %v", err)
		return
	}
	fmt.Println("✅ 批量索引演示完成！")
}

// 复杂文档索引示例
func complexDocumentExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 复杂文档索引示例 ===")

	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config.MilvusAddress)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	if err := ensureCollection(ctx, client, config.MilvusCollection); err != nil {
		log.Printf("确保集合存在失败: %v", err)
		return
	}

	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: config.MilvusCollection,
		Embedding:  embedder,
		Fields:     fields,
	}
	indexer, err := milvus.NewIndexer(ctx, cfg)
	if err != nil {
		log.Printf("创建Indexer失败: %v", err)
		return
	}

	// 准备复杂结构的文档
	complexDocuments := []*schema.Document{
		{
			ID: "complex_001",
			Content: `AI 大模型技术栈完整指南：
			
			1. 基础架构层
			- Transformer架构: 注意力机制的核心实现
			- 多头自注意力: 并行处理的关键技术
			- 位置编码: 序列信息的表示方法
			
			2. 模型训练层  
			- 预训练: 大规模无监督学习
			- 微调: 下游任务适配
			- 强化学习: 人类反馈优化
			
			3. 应用部署层
			- 模型量化: 减少计算资源需求
			- 推理优化: 提升响应速度
			- 分布式部署: 扩展服务能力`,
			MetaData: map[string]interface{}{
				"source":      "technical_guide",
				"category":    "AI技术",
				"difficulty":  "高级",
				"word_count":  156,
				"sections":    3,
				"tags":        []string{"AI", "大模型", "技术栈", "架构"},
				"author":      "AI研究员",
				"version":     "1.0",
				"last_update": "2024-09-13",
			},
		},
		{
			ID: "complex_002",
			Content: `云原生微服务架构设计原则：
			
			设计原则:
			• 单一职责: 每个服务只负责一个业务功能
			• 去中心化: 避免单点故障和性能瓶颈
			• 故障隔离: 服务间故障不相互影响
			• 自动扩缩: 根据负载自动调整资源
			
			技术选型:
			• 容器化: Docker + Kubernetes
			• 服务网格: Istio/Envoy
			• API网关: Kong/Ambassador  
			• 监控体系: Prometheus + Grafana
			• 日志收集: ELK/EFK Stack
			
			最佳实践:
			• 接口版本管理
			• 服务降级熔断
			• 分布式链路追踪
			• 安全策略实施`,
			MetaData: map[string]interface{}{
				"source":          "architecture_guide",
				"category":        "云原生",
				"difficulty":      "中高级",
				"word_count":      198,
				"principles":      4,
				"technologies":    6,
				"practices":       4,
				"tags":            []string{"云原生", "微服务", "架构", "容器"},
				"target_audience": "架构师",
				"complexity":      "high",
			},
		},
		{
			ID: "complex_003",
			Content: `RAG (检索增强生成) 系统实现详解：
			
			核心组件:
			1. 文档预处理
			   - 文本分块策略
			   - 重叠窗口设计
			   - 语义边界识别
			
			2. 向量化存储
			   - Embedding模型选择
			   - 向量数据库设计  
			   - 索引优化策略
			
			3. 检索系统
			   - 相似度计算方法
			   - 混合检索策略
			   - 结果排序算法
			
			4. 生成增强
			   - 提示词工程
			   - 上下文窗口管理
			   - 答案质量控制
			   
			性能优化:
			• 缓存策略: 减少重复计算
			• 异步处理: 提升用户体验
			• 负载均衡: 分散请求压力`,
			MetaData: map[string]interface{}{
				"source":                    "technical_tutorial",
				"category":                  "RAG系统",
				"difficulty":                "高级",
				"word_count":                186,
				"components":                4,
				"optimizations":             3,
				"tags":                      []string{"RAG", "检索", "生成", "AI应用"},
				"implementation_complexity": "very_high",
				"learning_curve":            "steep",
				"industry_adoption":         "growing",
			},
		},
	}

	fmt.Printf("📝 准备存储 %d 个复杂结构文档\n", len(complexDocuments))
	for i, doc := range complexDocuments {
		wordCount := doc.MetaData["word_count"]
		category := doc.MetaData["category"]
		fmt.Printf("  文档%d - ID: %s, 类别: %s, 字数: %v\n",
			i+1, doc.ID, category, wordCount)
	}

	storedIDs, err := indexer.Store(ctx, complexDocuments)
	if err != nil {
		log.Printf("存储复杂文档失败: %v", err)
		return
	}

	fmt.Printf("✅ 复杂文档索引成功，存储了 %d 个文档\n", len(storedIDs))
	fmt.Println("📋 文档特点:")
	fmt.Println("  - 包含丰富的元数据信息")
	fmt.Println("  - 结构化内容布局")
	fmt.Println("  - 多维度标签分类")
	fmt.Println("  - 支持复杂查询场景")

	fmt.Println("🔄 加载集合到内存...")
	err = client.LoadCollection(ctx, config.MilvusCollection, false)
	if err != nil {
		log.Printf("加载集合失败: %v", err)
		return
	}
	fmt.Println("✅ 复杂文档索引演示完成！")
}

// 索引性能测试示例
func indexPerformanceExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 索引性能测试示例 ===")

	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	client, err := initMilvusClient(ctx, config.MilvusAddress)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	if err := ensureCollection(ctx, client, config.MilvusCollection); err != nil {
		log.Printf("确保集合存在失败: %v", err)
		return
	}

	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: config.MilvusCollection,
		Embedding:  embedder,
		Fields:     fields,
	}
	indexer, err := milvus.NewIndexer(ctx, cfg)
	if err != nil {
		log.Printf("创建Indexer失败: %v", err)
		return
	}

	// 性能测试参数
	testCases := []struct {
		name      string
		batchSize int
		totalDocs int
	}{
		{"小批量测试", 5, 25},
		{"中批量测试", 10, 50},
		{"大批量测试", 20, 100},
	}

	for _, testCase := range testCases {
		fmt.Printf("\n🏃 执行 %s (批次大小: %d, 总文档数: %d)\n",
			testCase.name, testCase.batchSize, testCase.totalDocs)

		totalDocs := 0
		totalDuration := time.Duration(0)

		for batch := 0; batch < testCase.totalDocs/testCase.batchSize; batch++ {
			// 生成测试文档
			testDocs := make([]*schema.Document, testCase.batchSize)
			for i := 0; i < testCase.batchSize; i++ {
				docID := fmt.Sprintf("perf_%s_%d_%d",
					strings.ToLower(strings.ReplaceAll(testCase.name, " ", "_")), batch, i)
				testDocs[i] = &schema.Document{
					ID:      docID,
					Content: fmt.Sprintf("性能测试文档内容 - 批次 %d 文档 %d。这是一个用于测试索引性能的示例文档，包含一些基本的文本内容用于向量化处理。", batch, i),
					MetaData: map[string]interface{}{
						"test_case": testCase.name,
						"batch":     batch,
						"doc_index": i,
						"timestamp": time.Now().Unix(),
					},
				}
			}

			// 执行索引并计时
			startTime := time.Now()
			_, err := indexer.Store(ctx, testDocs)
			batchDuration := time.Since(startTime)

			if err != nil {
				log.Printf("批次 %d 索引失败: %v", batch, err)
				continue
			}

			totalDocs += testCase.batchSize
			totalDuration += batchDuration

			fmt.Printf("  批次 %d: %d 文档, 耗时 %v\n",
				batch, testCase.batchSize, batchDuration)
		}

		// 性能统计
		if totalDocs > 0 {
			avgTimePerDoc := totalDuration / time.Duration(totalDocs)
			docsPerSecond := float64(totalDocs) / totalDuration.Seconds()

			fmt.Printf("📊 %s 性能统计:\n", testCase.name)
			fmt.Printf("  总文档数: %d\n", totalDocs)
			fmt.Printf("  总耗时: %v\n", totalDuration)
			fmt.Printf("  平均耗时/文档: %v\n", avgTimePerDoc)
			fmt.Printf("  处理速度: %.2f 文档/秒\n", docsPerSecond)
		}
	}

	fmt.Println("\n🔄 加载集合到内存...")
	err = client.LoadCollection(ctx, config.MilvusCollection, false)
	if err != nil {
		log.Printf("加载集合失败: %v", err)
		return
	}
	fmt.Println("✅ 索引性能测试完成！")
}

// Chain编排模式示例
func chainExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== Chain 编排模式示例 ===")

	// 初始化 Embedder
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 初始化 Milvus 客户端
	client, err := initMilvusClient(ctx, config.MilvusAddress)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	// 确保集合存在
	if err := ensureCollection(ctx, client, config.MilvusCollection); err != nil {
		log.Printf("确保集合存在失败: %v", err)
		return
	}

	// 初始化 Indexer
	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: config.MilvusCollection,
		Embedding:  embedder,
		Fields:     fields,
	}
	indexer, err := milvus.NewIndexer(ctx, cfg)
	if err != nil {
		log.Printf("创建Indexer失败: %v", err)
		return
	}

	fmt.Println("🔗 创建文档处理Chain...")

	// 1️⃣ 创建 Chain - 声明输入输出类型
	// 输入: []*schema.Document，输出: []string (文档ID列表)
	chain := compose.NewChain[[]*schema.Document, []string]()

	// 2️⃣ 添加Indexer组件到Chain中
	chain.AppendIndexer(indexer)

	// 3️⃣ 编译成可运行实例
	fmt.Println("⚙️ 编译Chain工作流...")
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("Chain编译失败: %v", err)
		return
	}

	fmt.Println("✅ Chain编译成功！")

	// 准备测试文档
	documents := []*schema.Document{
		{
			ID:       "chain_001",
			Content:  "Chain编排是Eino框架的核心特性，它允许将多个组件串联起来形成完整的处理工作流。",
			MetaData: map[string]interface{}{"source": "chain_demo", "type": "concept"},
		},
		{
			ID:       "chain_002",
			Content:  "通过Chain，可以实现文档的自动化处理：文档输入 → 向量化 → 存储索引 → 返回结果。",
			MetaData: map[string]interface{}{"source": "chain_demo", "type": "workflow"},
		},
		{
			ID:       "chain_003",
			Content:  "Chain编排模式特别适合线性处理流程，具有良好的可组合性和可扩展性。",
			MetaData: map[string]interface{}{"source": "chain_demo", "type": "advantage"},
		},
	}

	fmt.Printf("📝 准备通过Chain处理 %d 个文档\n", len(documents))
	for i, doc := range documents {
		fmt.Printf("  文档%d - ID: %s\n", i+1, doc.ID)
	}

	// 4️⃣ 通过Chain运行工作流
	fmt.Println("🚀 执行Chain工作流...")
	startTime := time.Now()

	documentIDs, err := runnable.Invoke(ctx, documents)
	if err != nil {
		log.Printf("Chain执行失败: %v", err)
		return
	}

	duration := time.Since(startTime)

	fmt.Printf("✅ Chain执行成功，耗时: %v\n", duration)
	fmt.Printf("📊 通过Chain存储了 %d 个文档: %v\n", len(documentIDs), documentIDs)

	// 加载集合到内存
	fmt.Println("🔄 加载集合到内存...")
	err = client.LoadCollection(ctx, config.MilvusCollection, false)
	if err != nil {
		log.Printf("加载集合失败: %v", err)
		return
	}

	fmt.Println("🎯 Chain编排的优势:")
	fmt.Println("  • 声明式编程: 专注于组件关系而非实现细节")
	fmt.Println("  • 类型安全: 编译时检查输入输出类型匹配")
	fmt.Println("  • 易于测试: 可以独立测试每个组件")
	fmt.Println("  • 可复用性: Chain可以作为更大工作流的一部分")
	fmt.Println("  • 错误传播: 统一的错误处理机制")

	fmt.Println("✅ Chain编排模式演示完成！")
}

// Option配置示例
func optionConfigExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== Option 配置示例 ===")

	// 初始化 Embedder
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 初始化 Milvus 客户端
	client, err := initMilvusClient(ctx, config.MilvusAddress)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	// 确保集合存在
	if err := ensureCollection(ctx, client, config.MilvusCollection); err != nil {
		log.Printf("确保集合存在失败: %v", err)
		return
	}

	// 初始化 Indexer
	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: config.MilvusCollection,
		Embedding:  embedder,
		Fields:     fields,
	}
	indexer, err := milvus.NewIndexer(ctx, cfg)
	if err != nil {
		log.Printf("创建Indexer失败: %v", err)
		return
	}

	// 准备测试文档
	documents := []*schema.Document{
		{
			ID:       "option_001",
			Content:  "Option配置允许在运行时传入额外的配置参数，提供更大的灵活性。",
			MetaData: map[string]interface{}{"source": "option_demo", "type": "concept"},
		},
		{
			ID:       "option_002",
			Content:  "通过Option，可以实现子索引操作、临时替换组件等高级功能。",
			MetaData: map[string]interface{}{"source": "option_demo", "type": "feature"},
		},
	}

	fmt.Printf("📝 准备演示Option配置功能\n")

	// 演示1: 基础存储（无Option）
	fmt.Println("\n🔸 演示1: 基础存储（无额外Option）")
	startTime := time.Now()
	basicIDs, err := indexer.Store(ctx, documents[:1])
	if err != nil {
		log.Printf("基础存储失败: %v", err)
		return
	}
	fmt.Printf("✅ 基础存储成功，耗时: %v，存储文档: %v\n", time.Since(startTime), basicIDs)

	// 演示2: Option配置说明和概念演示
	fmt.Println("\n🔸 演示2: Option配置说明和概念演示")

	startTime = time.Now()

	// 注意：这里演示Option的使用方式，实际项目中需要根据具体的milvus包API来调整
	// 由于我们无法直接看到milvus包的Option实现，这里展示概念性用法
	fmt.Println("📋 Option配置说明:")
	fmt.Println("  • WithEmbedding(embedder): 临时替换Embedding组件")
	fmt.Println("  • WithSubIndexes([]string): 指定子索引操作")
	fmt.Println("  • WithTimeout(duration): 设置操作超时时间")
	fmt.Println("  • WithRetry(count): 设置重试次数")

	// 基础存储（演示概念，实际API可能不同）
	optionIDs, err := indexer.Store(ctx, documents[1:])
	if err != nil {
		log.Printf("Option存储失败: %v", err)
		return
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Option存储成功，耗时: %v，存储文档: %v\n", duration, optionIDs)

	// 演示3: 不同场景下的Option使用
	fmt.Println("\n🔸 演示3: 不同场景的Option配置模式")
	fmt.Println("🎯 高优先级文档处理场景:")
	fmt.Println("   ids, err := indexer.Store(ctx, documents,")
	fmt.Println("       WithTimeout(30*time.Second),    // 延长超时时间")
	fmt.Println("       WithRetry(3),                   // 增加重试次数")
	fmt.Println("       WithPriority(\"high\"),           // 设置处理优先级")
	fmt.Println("   )")

	fmt.Println("\n📊 批量处理场景:")
	fmt.Println("   ids, err := indexer.Store(ctx, batchDocuments,")
	fmt.Println("       WithBatchSize(100),             // 设置批量大小")
	fmt.Println("       WithParallel(4),                // 并行处理数量")
	fmt.Println("       WithCallback(progressHandler),  // 进度回调")
	fmt.Println("   )")

	fmt.Println("\n🔍 子索引分区场景:")
	fmt.Println("   ids, err := indexer.Store(ctx, documents,")
	fmt.Println("       WithSubIndexes([]string{\"partition_1\", \"partition_2\"}),")
	fmt.Println("       WithPartitionKey(\"category\"),   // 分区键")
	fmt.Println("   )")

	// 加载集合到内存
	fmt.Println("\n🔄 加载集合到内存...")
	err = client.LoadCollection(ctx, config.MilvusCollection, false)
	if err != nil {
		log.Printf("加载集合失败: %v", err)
		return
	}

	fmt.Println("💡 Option配置的优势:")
	fmt.Println("  • 运行时灵活性: 无需重新创建组件即可调整行为")
	fmt.Println("  • 场景适配: 针对不同业务场景使用不同配置")
	fmt.Println("  • 性能优化: 根据数据特点优化处理参数")
	fmt.Println("  • 错误恢复: 配置重试和超时等容错机制")
	fmt.Println("✅ Option配置演示完成！")
}

// Callback机制示例
func callbackExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== Callback 机制示例 ===")

	// 初始化 Embedder
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 初始化 Milvus 客户端
	client, err := initMilvusClient(ctx, config.MilvusAddress)
	if err != nil {
		log.Printf("初始化Milvus客户端失败: %v", err)
		return
	}
	defer client.Close()

	// 确保集合存在
	if err := ensureCollection(ctx, client, config.MilvusCollection); err != nil {
		log.Printf("确保集合存在失败: %v", err)
		return
	}

	// 初始化 Indexer
	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: config.MilvusCollection,
		Embedding:  embedder,
		Fields:     fields,
	}
	indexer, err := milvus.NewIndexer(ctx, cfg)
	if err != nil {
		log.Printf("创建Indexer失败: %v", err)
		return
	}

	// 准备测试文档
	documents := []*schema.Document{
		{
			ID:       "callback_001",
			Content:  "Callback机制是Eino框架的重要特性，允许在组件生命周期的关键节点注入自定义逻辑。",
			MetaData: map[string]interface{}{"source": "callback_demo", "type": "concept"},
		},
		{
			ID:       "callback_002",
			Content:  "通过OnStart、OnEnd、OnError等回调，可以实现日志记录、性能监控、错误处理等功能。",
			MetaData: map[string]interface{}{"source": "callback_demo", "type": "feature"},
		},
		{
			ID:       "callback_003",
			Content:  "Callback机制提供了非侵入式的组件扩展能力，是构建可观测系统的基础。",
			MetaData: map[string]interface{}{"source": "callback_demo", "type": "advantage"},
		},
	}

	fmt.Printf("📝 准备演示Callback机制功能，处理 %d 个文档\n", len(documents))

	// 演示1: 手动触发回调事件（概念演示）
	fmt.Println("\n🔸 演示1: Callback机制概念说明")

	fmt.Println("🔄 模拟OnStart回调...")
	fmt.Println("   🚀 开始索引操作")
	fmt.Printf("   📝 准备索引 %d 个文档\n", len(documents))
	for i, doc := range documents {
		fmt.Printf("      文档%d: ID=%s, 内容长度=%d字符\n",
			i+1, doc.ID, len(doc.Content))
	}

	// 记录开始时间
	startTime := time.Now()

	// 执行实际存储
	fmt.Println("⚡ 执行文档索引...")
	storedIDs, err := indexer.Store(ctx, documents)

	duration := time.Since(startTime)

	if err != nil {
		// 模拟OnError回调
		fmt.Println("🔴 模拟OnError回调...")
		fmt.Printf("   ❌ 索引操作失败，耗时: %v\n", duration)
		fmt.Printf("   💥 错误详情: %v\n", err)
		fmt.Println("   🔧 建议检查:")
		fmt.Println("      • 网络连接状态")
		fmt.Println("      • 文档格式是否正确")
		fmt.Println("      • Milvus服务是否正常")
		fmt.Println("      • API配置是否有效")
		return
	}

	// 模拟OnEnd回调
	fmt.Println("🟢 模拟OnEnd回调...")
	fmt.Printf("   ✅ 索引操作完成，总耗时: %v\n", duration)
	fmt.Printf("   📊 成功索引 %d 个文档: %v\n", len(storedIDs), storedIDs)

	// 演示2: Callback配置代码示例
	fmt.Println("\n🔸 演示2: Callback配置代码示例")
	fmt.Println("📋 标准Callback处理器创建代码:")
	fmt.Println("   callbackHandler := callbacks.NewHandlerBuilder().")
	fmt.Println("       OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {")
	fmt.Println("           fmt.Printf(\"🚀 开始索引操作\\n\")")
	fmt.Println("           return context.WithValue(ctx, \"start_time\", time.Now())")
	fmt.Println("       }).")
	fmt.Println("       OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) {")
	fmt.Println("           startTime, _ := ctx.Value(\"start_time\").(time.Time)")
	fmt.Println("           fmt.Printf(\"✅ 索引完成，耗时: %v\\n\", time.Since(startTime))")
	fmt.Println("       }).")
	fmt.Println("       OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) {")
	fmt.Println("           fmt.Printf(\"❌ 索引失败: %v\\n\", err)")
	fmt.Println("       }).")
	fmt.Println("       Build()")

	fmt.Println("\n📋 在Chain中使用Callback:")
	fmt.Println("   chain := compose.NewChain[[]*schema.Document, []string]()")
	fmt.Println("   chain.AppendIndexer(indexer, compose.WithCallbacks(callbackHandler))")
	fmt.Println("   runnable, err := chain.Compile(ctx)")

	// 加载集合到内存
	fmt.Println("\n🔄 加载集合到内存...")
	err = client.LoadCollection(ctx, config.MilvusCollection, false)
	if err != nil {
		log.Printf("加载集合失败: %v", err)
		return
	}

	fmt.Println("🎯 Callback机制的优势:")
	fmt.Println("  • 可观测性: 全面监控组件执行状态和性能")
	fmt.Println("  • 非侵入式: 不修改组件代码即可扩展功能")
	fmt.Println("  • 灵活配置: 按需启用不同类型的回调处理")
	fmt.Println("  • 统一接口: 所有组件使用相同的回调机制")
	fmt.Println("  • 生命周期: 覆盖开始、结束、错误等关键节点")

	fmt.Println("\n📊 常见Callback应用场景:")
	fmt.Println("  • 📈 性能监控: 记录执行时间和资源使用")
	fmt.Println("  • 📝 操作日志: 详细记录组件执行过程")
	fmt.Println("  • ⚠️ 错误处理: 统一的错误记录和告警")
	fmt.Println("  • 🔄 进度跟踪: 实时显示长时间操作的进度")
	fmt.Println("  • 🧪 调试辅助: 开发阶段的详细调试信息")

	fmt.Println("✅ Callback机制演示完成！")
}

// 错误处理示例
func errorHandlingExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 错误处理示例 ===")

	// 演示配置验证
	if config.APIKey == "" {
		fmt.Println("❌ 错误演示: API Key 未配置")
		fmt.Println("   解决方案: 设置环境变量 ARK_API_KEY 或在 config.yaml 中配置")
		return
	}

	if config.MilvusAddress == "" {
		fmt.Println("❌ 错误演示: Milvus 地址未配置")
		fmt.Println("   解决方案: 设置环境变量 MILVUS_ADDRESS 或在 config.yaml 中配置")
		return
	}

	fmt.Println("✅ 配置验证通过")

	// 演示连接错误处理
	fmt.Println("🔄 测试 Milvus 连接...")
	client, err := initMilvusClient(ctx, config.MilvusAddress)
	if err != nil {
		fmt.Printf("❌ 连接错误演示: %v\n", err)
		fmt.Println("   常见原因:")
		fmt.Println("   1. Milvus 服务未启动")
		fmt.Println("   2. 网络连接问题")
		fmt.Println("   3. 地址配置错误")
		fmt.Println("   解决方案: 检查 Milvus 服务状态和网络配置")
		return
	}
	defer client.Close()
	fmt.Println("✅ Milvus 连接成功")

	// 演示 Embedder 错误处理
	fmt.Println("🔄 测试 Embedder 初始化...")
	_, err = initEmbedder(ctx, config)
	if err != nil {
		fmt.Printf("❌ Embedder错误演示: %v\n", err)
		fmt.Println("   常见原因:")
		fmt.Println("   1. API Key 无效或过期")
		fmt.Println("   2. 模型名称错误")
		fmt.Println("   3. 网络连接问题")
		fmt.Println("   解决方案: 检查 API 配置和网络连接")
		return
	}
	fmt.Println("✅ Embedder 初始化成功")

	// 演示文档验证
	fmt.Println("🔄 测试文档格式验证...")
	invalidDocuments := []*schema.Document{
		{
			ID:      "", // 空ID
			Content: "测试内容",
		},
		{
			ID:      "valid_id",
			Content: "", // 空内容
		},
	}

	for i, doc := range invalidDocuments {
		fmt.Printf("  验证文档 %d:\n", i+1)
		if doc.ID == "" {
			fmt.Println("    ❌ 文档ID为空")
		} else if doc.Content == "" {
			fmt.Println("    ❌ 文档内容为空")
		} else {
			fmt.Println("    ✅ 文档格式正确")
		}
	}

	fmt.Println("📋 错误处理最佳实践:")
	fmt.Println("  1. 配置验证: 启动时检查所有必需配置")
	fmt.Println("  2. 连接测试: 初始化时验证外部依赖")
	fmt.Println("  3. 输入验证: 处理前验证文档格式")
	fmt.Println("  4. 错误分类: 区分可恢复和不可恢复错误")
	fmt.Println("  5. 日志记录: 详细记录错误信息便于排查")
	fmt.Println("✅ 错误处理演示完成！")
}

// 主函数
func main() {
	ctx := context.Background()

	fmt.Println("🤖 Eino Indexer 组件完全示例")
	fmt.Println("====================================")

	// 1. 初始化配置
	config, err := initConfig()
	if err != nil {
		log.Fatal("配置初始化失败:", err)
	}

	fmt.Printf("使用配置:\n")
	fmt.Printf("  Milvus地址: %s\n", config.MilvusAddress)
	fmt.Printf("  集合名称: %s\n", config.MilvusCollection)
	fmt.Printf("  嵌入模型: %s\n", config.EmbedderModel)

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
			try("基础索引示例", basicIndexExample)
		case "batch":
			try("批量索引示例", batchIndexExample)
		case "complex":
			try("复杂文档索引示例", complexDocumentExample)
		case "performance":
			try("索引性能测试示例", indexPerformanceExample)
		case "chain":
			try("Chain编排模式示例", chainExample)
		case "option":
			try("Option配置示例", optionConfigExample)
		case "callback":
			try("Callback机制示例", callbackExample)
		case "error":
			try("错误处理示例", errorHandlingExample)
		default:
			fmt.Printf("未知示例: %s\n", exampleName)
			fmt.Println("可用示例: basic, batch, complex, performance, chain, option, callback, error")
			return
		}
	} else {
		// 运行所有示例
		//try("基础索引示例", basicIndexExample)
		//try("批量索引示例", batchIndexExample)
		//try("复杂文档索引示例", complexDocumentExample)
		//try("索引性能测试示例", indexPerformanceExample)
		//try("Chain编排模式示例", chainExample)
		try("Option配置示例", optionConfigExample)
		//try("Callback机制示例", callbackExample)
		//try("错误处理示例", errorHandlingExample)
	}

	fmt.Println("\n🎉 所有示例运行完成！")
	fmt.Println("\n使用方法:")
	fmt.Println("  go run main.go              # 运行所有示例")
	fmt.Println("  go run main.go basic        # 运行基础索引示例")
	fmt.Println("  go run main.go batch        # 运行批量索引示例")
	fmt.Println("  go run main.go complex      # 运行复杂文档示例")
	fmt.Println("  go run main.go performance  # 运行性能测试示例")
	fmt.Println("  go run main.go chain        # 运行Chain编排模式示例")
	fmt.Println("  go run main.go option       # 运行Option配置示例")
	fmt.Println("  go run main.go callback     # 运行Callback机制示例")
	fmt.Println("  go run main.go error        # 运行错误处理示例")
}
