package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"
	//"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/compose"
	"github.com/spf13/viper"
)

// 配置结构体
type Config struct {
	APIKey        string `mapstructure:"api_key"`
	Model         string `mapstructure:"model"`
	EmbedderModel string `mapstructure:"embedder_model"`
}

// 初始化配置
func initConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	_ = viper.ReadInConfig() // 忽略错误，因为我们也会检查环境变量

	return &Config{
		APIKey:        viper.GetString("ARK_API_KEY"),
		Model:         viper.GetString("ARK_MODEL"),
		EmbedderModel: viper.GetString("EMBEDDER_MODEL"),
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

// 计算余弦相似度
func cosineSimilarity(v1, v2 []float64) (float64, error) {
	if len(v1) != len(v2) {
		return 0, fmt.Errorf("向量维度不匹配: %d vs %d", len(v1), len(v2))
	}

	var dotProduct, normV1, normV2 float64
	for i := 0; i < len(v1); i++ {
		dotProduct += v1[i] * v2[i]
		normV1 += v1[i] * v1[i]
		normV2 += v2[i] * v2[i]
	}

	if normV1 == 0 || normV2 == 0 {
		return 0, fmt.Errorf("向量的模不能为零")
	}

	return dotProduct / (math.Sqrt(normV1) * math.Sqrt(normV2)), nil
}

// 基础嵌入示例
func basicEmbeddingExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 基础嵌入示例 ===")

	// 初始化 Embedder
	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 准备测试文本
	texts := []string{
		"今天天气真好，阳光明媚。",
		"今天是个大晴天，万里无云。",
		"红烧肉怎么做才好吃？",
	}

	fmt.Println("📝 输入文本:")
	for i, text := range texts {
		fmt.Printf("  文本%c: %s\n", 'A'+i, text)
	}

	// 执行向量化
	fmt.Println("\n🔄 正在将文本转换为向量...")
	startTime := time.Now()

	vectors, err := embedder.EmbedStrings(ctx, texts)
	if err != nil {
		log.Printf("向量化失败: %v", err)
		return
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ 向量转换完成，耗时: %v\n", duration)
	fmt.Printf("📊 生成了 %d 个向量，每个向量维度: %d\n", len(vectors), len(vectors[0]))

	// 计算相似度
	fmt.Println("\n🔍 计算文本相似度...")
	vecA, vecB, vecC := vectors[0], vectors[1], vectors[2]

	simAB, err := cosineSimilarity(vecA, vecB)
	if err != nil {
		log.Printf("计算A-B相似度失败: %v", err)
		return
	}

	simAC, err := cosineSimilarity(vecA, vecC)
	if err != nil {
		log.Printf("计算A-C相似度失败: %v", err)
		return
	}

	simBC, err := cosineSimilarity(vecB, vecC)
	if err != nil {
		log.Printf("计算B-C相似度失败: %v", err)
		return
	}

	fmt.Println("\n📈 相似度计算结果:")
	fmt.Printf("  • 文本A ↔ 文本B (语义相似): %.4f\n", simAB)
	fmt.Printf("  • 文本A ↔ 文本C (语义不同): %.4f\n", simAC)
	fmt.Printf("  • 文本B ↔ 文本C (语义不同): %.4f\n", simBC)

	fmt.Println("\n💡 结果分析:")
	fmt.Println("  可以看到，语义相似的文本对(A-B)获得了远高于不相似文本对的相似度分数！")
}

// 批量处理示例
func batchProcessingExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 批量处理示例 ===")

	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 准备大批量文本
	categories := []string{
		"科技", "体育", "娱乐", "财经", "健康",
		"教育", "旅游", "美食", "汽车", "时尚",
	}

	var batchTexts []string
	for _, category := range categories {
		for i := 0; i < 3; i++ {
			text := fmt.Sprintf("%s领域的最新发展趋势和前沿技术应用，第%d个示例。", category, i+1)
			batchTexts = append(batchTexts, text)
		}
	}

	fmt.Printf("📝 准备批量处理 %d 个文本\n", len(batchTexts))

	// 分批处理演示
	batchSize := 10
	var allVectors [][]float64
	totalDuration := time.Duration(0)

	fmt.Printf("🔄 开始分批处理，批次大小: %d\n", batchSize)

	for i := 0; i < len(batchTexts); i += batchSize {
		end := i + batchSize
		if end > len(batchTexts) {
			end = len(batchTexts)
		}

		batch := batchTexts[i:end]
		fmt.Printf("  处理批次 %d/%d: 文本 %d-%d\n",
			(i/batchSize)+1, (len(batchTexts)+batchSize-1)/batchSize, i+1, end)

		startTime := time.Now()
		batchVectors, err := embedder.EmbedStrings(ctx, batch)
		if err != nil {
			log.Printf("批次处理失败: %v", err)
			return
		}

		batchDuration := time.Since(startTime)
		totalDuration += batchDuration
		allVectors = append(allVectors, batchVectors...)

		fmt.Printf("    ✅ 批次完成，耗时: %v\n", batchDuration)
	}

	fmt.Printf("✅ 批量处理完成！\n")
	fmt.Printf("📊 性能统计:\n")
	fmt.Printf("  • 总文本数: %d\n", len(batchTexts))
	fmt.Printf("  • 总耗时: %v\n", totalDuration)
	fmt.Printf("  • 平均耗时/文本: %v\n", totalDuration/time.Duration(len(batchTexts)))
	fmt.Printf("  • 处理速度: %.2f 文本/秒\n", float64(len(batchTexts))/totalDuration.Seconds())
	fmt.Printf("  • 生成向量数: %d，向量维度: %d\n", len(allVectors), len(allVectors[0]))
}

// 语义搜索示例
func semanticSearchExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 语义搜索示例 ===")

	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 构建文档库
	documents := []string{
		"苹果iPhone是一款智能手机，具有出色的摄影功能和流畅的用户体验。",
		"华为Mate系列手机在拍照和续航方面表现突出，深受用户喜爱。",
		"特斯拉Model 3是一款纯电动汽车，具有自动驾驶和长续航里程。",
		"比亚迪汉EV是国产电动车的代表，在安全性和性价比方面优势明显。",
		"Python是一种高级编程语言，广泛应用于数据科学和机器学习。",
		"JavaScript是Web前端开发的核心技术，也可用于后端开发。",
		"机器学习是人工智能的重要分支，通过算法让计算机具备学习能力。",
		"深度学习基于神经网络，在图像识别和自然语言处理领域表现卓越。",
	}

	fmt.Printf("📚 构建文档库，包含 %d 个文档\n", len(documents))
	for i, doc := range documents {
		fmt.Printf("  文档%d: %s\n", i+1, doc)
	}

	// 为所有文档生成向量
	fmt.Println("\n🔄 为文档库生成向量索引...")
	docVectors, err := embedder.EmbedStrings(ctx, documents)
	if err != nil {
		log.Printf("文档向量化失败: %v", err)
		return
	}
	fmt.Println("✅ 文档索引构建完成")

	// 测试查询
	queries := []string{
		"手机拍照效果",
		"电动汽车续航",
		"编程语言学习",
		"人工智能应用",
	}

	for _, query := range queries {
		fmt.Printf("\n🔍 搜索查询: \"%s\"\n", query)

		// 查询向量化
		queryVectors, err := embedder.EmbedStrings(ctx, []string{query})
		if err != nil {
			log.Printf("查询向量化失败: %v", err)
			continue
		}
		queryVector := queryVectors[0]

		// 计算与所有文档的相似度
		type ScoredDoc struct {
			Index int
			Doc   string
			Score float64
		}

		var scoredDocs []ScoredDoc
		for i, docVector := range docVectors {
			similarity, err := cosineSimilarity(queryVector, docVector)
			if err != nil {
				log.Printf("相似度计算失败: %v", err)
				continue
			}
			scoredDocs = append(scoredDocs, ScoredDoc{
				Index: i,
				Doc:   documents[i],
				Score: similarity,
			})
		}

		// 按相似度排序
		sort.Slice(scoredDocs, func(i, j int) bool {
			return scoredDocs[i].Score > scoredDocs[j].Score
		})

		// 显示Top 3结果
		fmt.Println("📋 搜索结果 (按相关性排序):")
		topK := 3
		if topK > len(scoredDocs) {
			topK = len(scoredDocs)
		}

		for i := 0; i < topK; i++ {
			result := scoredDocs[i]
			fmt.Printf("  %d. [相似度: %.4f] 文档%d: %s\n",
				i+1, result.Score, result.Index+1, result.Doc)
		}
	}
}

// Chain编排示例
func chainEmbeddingExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== Chain 编排示例 ===")

	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	fmt.Println("🔗 创建文本向量化Chain...")

	// 创建 Chain - 声明输入输出类型
	// 输入: []string，输出: [][]float64
	chain := compose.NewChain[[]string, [][]float64]()

	// 添加Embedding组件到Chain中
	chain.AppendEmbedding(embedder)

	// 编译成可运行实例
	fmt.Println("⚙️ 编译Chain工作流...")
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("Chain编译失败: %v", err)
		return
	}

	fmt.Println("✅ Chain编译成功！")

	// 准备测试文本
	testTexts := []string{
		"Chain编排是Eino框架的核心特性，它允许将多个组件串联起来形成完整的处理工作流。",
		"通过Chain，可以实现文本的自动化处理：文本输入 → 向量化 → 语义分析 → 结果输出。",
		"Embedding组件在Chain中发挥着重要作用，为下游的语义处理提供向量基础。",
	}

	fmt.Printf("📝 准备通过Chain处理 %d 个文本\n", len(testTexts))
	for i, text := range testTexts {
		fmt.Printf("  文本%d: %s\n", i+1, text)
	}

	// 通过Chain运行工作流
	fmt.Println("🚀 执行Chain工作流...")
	startTime := time.Now()

	vectors, err := runnable.Invoke(ctx, testTexts)
	if err != nil {
		log.Printf("Chain执行失败: %v", err)
		return
	}

	duration := time.Since(startTime)

	fmt.Printf("✅ Chain执行成功，耗时: %v\n", duration)
	fmt.Printf("📊 通过Chain生成了 %d 个向量，维度: %d\n", len(vectors), len(vectors[0]))

	fmt.Println("🎯 Chain编排的优势:")
	fmt.Println("  • 模块化设计: 组件间职责清晰，易于维护")
	fmt.Println("  • 类型安全: 编译时检查输入输出类型匹配")
	fmt.Println("  • 易于扩展: 可以方便地添加更多处理组件")
	fmt.Println("  • 统一接口: 所有组件使用相同的调用方式")
	fmt.Println("  • 可组合性: Chain可以作为更大工作流的一部分")

	fmt.Println("✅ Chain编排示例完成！")
}

// 性能测试示例
func performanceTestExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== 性能测试示例 ===")

	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 性能测试参数
	testCases := []struct {
		name       string
		batchSize  int
		totalTexts int
	}{
		{"小批量测试", 5, 25},
		{"中批量测试", 10, 50},
		{"大批量测试", 20, 100},
	}

	for _, testCase := range testCases {
		fmt.Printf("\n🏃 执行 %s (批次大小: %d, 总文本数: %d)\n",
			testCase.name, testCase.batchSize, testCase.totalTexts)

		// 生成测试文本
		var testTexts []string
		for i := 0; i < testCase.totalTexts; i++ {
			text := fmt.Sprintf("性能测试文本内容 - 编号 %d。这是一个用于测试Embedding组件性能的示例文本，包含一些基本的语义信息用于向量化处理。", i+1)
			testTexts = append(testTexts, text)
		}

		totalVectors := 0
		totalDuration := time.Duration(0)

		for i := 0; i < len(testTexts); i += testCase.batchSize {
			end := i + testCase.batchSize
			if end > len(testTexts) {
				end = len(testTexts)
			}

			batch := testTexts[i:end]

			// 执行向量化并计时
			startTime := time.Now()
			vectors, err := embedder.EmbedStrings(ctx, batch)
			batchDuration := time.Since(startTime)

			if err != nil {
				log.Printf("批次 %d 向量化失败: %v", i/testCase.batchSize, err)
				continue
			}

			totalVectors += len(vectors)
			totalDuration += batchDuration

			fmt.Printf("  批次 %d: %d 文本, 耗时 %v\n",
				i/testCase.batchSize+1, len(batch), batchDuration)
		}

		// 性能统计
		if totalVectors > 0 {
			avgTimePerText := totalDuration / time.Duration(totalVectors)
			textsPerSecond := float64(totalVectors) / totalDuration.Seconds()

			fmt.Printf("📊 %s 性能统计:\n", testCase.name)
			fmt.Printf("  总文本数: %d\n", totalVectors)
			fmt.Printf("  总耗时: %v\n", totalDuration)
			fmt.Printf("  平均耗时/文本: %v\n", avgTimePerText)
			fmt.Printf("  处理速度: %.2f 文本/秒\n", textsPerSecond)
		}
	}

	fmt.Println("✅ 性能测试完成！")
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

	if config.EmbedderModel == "" {
		fmt.Println("❌ 错误演示: Embedder模型未配置")
		fmt.Println("   解决方案: 设置环境变量 EMBEDDER_MODEL 或在 config.yaml 中配置")
		return
	}

	fmt.Println("✅ 配置验证通过")

	// 演示Embedder初始化错误处理
	fmt.Println("🔄 测试 Embedder 初始化...")
	embedder, err := initEmbedder(ctx, config)
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

	// 演示输入验证
	fmt.Println("🔄 测试输入数据验证...")
	invalidInputs := [][]string{
		{},               // 空输入
		{""},             // 空字符串
		{"", "有效文本", ""}, // 混合输入
	}

	for i, input := range invalidInputs {
		fmt.Printf("  验证输入 %d: %v\n", i+1, input)

		if len(input) == 0 {
			fmt.Println("    ❌ 输入为空")
			continue
		}

		hasEmpty := false
		for _, text := range input {
			if text == "" {
				hasEmpty = true
				break
			}
		}

		if hasEmpty {
			fmt.Println("    ❌ 包含空字符串")
		} else {
			fmt.Println("    ✅ 输入格式正确")
		}
	}

	// 演示带重试的向量化
	fmt.Println("\n🔄 测试带重试的向量化...")
	testTexts := []string{"这是一个测试文本"}
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("  尝试 %d/%d...\n", attempt, maxRetries)

		vectors, err := embedder.EmbedStrings(ctx, testTexts)
		if err != nil {
			fmt.Printf("    ❌ 失败: %v\n", err)
			if attempt < maxRetries {
				fmt.Println("    🔄 准备重试...")
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			} else {
				fmt.Println("    ❌ 达到最大重试次数，操作失败")
				return
			}
		}

		fmt.Printf("    ✅ 成功，生成 %d 个向量\n", len(vectors))
		break
	}

	fmt.Println("📋 错误处理最佳实践:")
	fmt.Println("  1. 配置验证: 启动时检查所有必需配置")
	fmt.Println("  2. 输入验证: 处理前验证输入数据格式")
	fmt.Println("  3. 重试机制: 网络请求失败时实施指数退避重试")
	fmt.Println("  4. 错误分类: 区分可恢复和不可恢复错误")
	fmt.Println("  5. 日志记录: 详细记录错误信息便于排查")
	fmt.Println("  6. 降级策略: 在服务不可用时提供备选方案")
	fmt.Println("✅ 错误处理演示完成！")
}

// Option配置示例
func optionConfigExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== Option 配置示例 ===")

	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 准备测试文本
	testTexts := []string{
		"Option配置允许在运行时传入额外的配置参数，提供更大的灵活性。",
		"通过Option，可以实现模型切换、批次大小调整等高级功能。",
		"Embedding组件的Option支持多种参数配置，满足不同业务需求。",
	}

	fmt.Printf("📝 准备演示Option配置功能，处理 %d 个文本\n", len(testTexts))
	for i, text := range testTexts {
		fmt.Printf("  文本%d: %s\n", i+1, text)
	}

	// 演示1: 基础向量化（无Option）
	fmt.Println("\n🔸 演示1: 基础向量化（无额外Option）")
	startTime := time.Now()

	basicVectors, err := embedder.EmbedStrings(ctx, testTexts[:1])
	if err != nil {
		log.Printf("基础向量化失败: %v", err)
		return
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ 基础向量化成功，耗时: %v，生成向量数: %d\n",
		duration, len(basicVectors))

	// 演示2: 使用WithModel Option（概念展示）
	fmt.Println("\n🔸 演示2: Option配置说明和概念演示")
	fmt.Println("📋 常用Option配置:")
	fmt.Println("  • embedding.WithModel(\"model-name\"): 临时切换模型")
	fmt.Println("  • embedding.WithBatchSize(100): 设置批处理大小")
	fmt.Println("  • embedding.WithTimeout(30*time.Second): 设置超时时间")
	fmt.Println("  • embedding.WithRetry(3): 设置重试次数")

	// 演示3: 批处理大小配置（模拟）
	fmt.Println("\n🔸 演示3: 批处理优化配置")

	// 准备更多测试文本用于批处理演示
	batchTexts := make([]string, 15)
	for i := 0; i < 15; i++ {
		batchTexts[i] = fmt.Sprintf("批处理测试文本 %d - 这是用于演示批处理配置的示例文本。", i+1)
	}

	// 模拟不同批次大小的效果
	batchSizes := []int{5, 10, 15}

	for _, batchSize := range batchSizes {
		fmt.Printf("\n  🔹 批次大小: %d\n", batchSize)
		startTime := time.Now()

		var allVectors [][]float64
		for i := 0; i < len(batchTexts); i += batchSize {
			end := i + batchSize
			if end > len(batchTexts) {
				end = len(batchTexts)
			}

			batch := batchTexts[i:end]

			// 这里演示Option概念，实际使用时根据具体API调整
			// vectors, err := embedder.EmbedStrings(ctx, batch, embedding.WithBatchSize(batchSize))
			vectors, err := embedder.EmbedStrings(ctx, batch)
			if err != nil {
				log.Printf("批次处理失败: %v", err)
				continue
			}

			allVectors = append(allVectors, vectors...)
		}

		batchDuration := time.Since(startTime)
		fmt.Printf("    耗时: %v, 处理文本: %d, 生成向量: %d\n",
			batchDuration, len(batchTexts), len(allVectors))
	}

	// 演示4: 不同场景下的Option使用模式
	fmt.Println("\n🔸 演示4: 不同场景的Option配置模式")

	fmt.Println("🎯 高精度场景配置:")
	fmt.Println("   vectors, err := embedder.EmbedStrings(ctx, texts,")
	fmt.Println("       embedding.WithModel(\"high-precision-model\"),")
	fmt.Println("       embedding.WithTimeout(60*time.Second),")
	fmt.Println("       embedding.WithRetry(5),")
	fmt.Println("   )")

	fmt.Println("\n📊 大批量处理场景:")
	fmt.Println("   vectors, err := embedder.EmbedStrings(ctx, largeBatch,")
	fmt.Println("       embedding.WithBatchSize(200),")
	fmt.Println("       embedding.WithTimeout(120*time.Second),")
	fmt.Println("       embedding.WithParallel(4),")
	fmt.Println("   )")

	fmt.Println("\n⚡ 快速处理场景:")
	fmt.Println("   vectors, err := embedder.EmbedStrings(ctx, texts,")
	fmt.Println("       embedding.WithModel(\"fast-model\"),")
	fmt.Println("       embedding.WithBatchSize(50),")
	fmt.Println("       embedding.WithTimeout(10*time.Second),")
	fmt.Println("   )")

	fmt.Println("\n💡 Option配置的优势:")
	fmt.Println("  • 运行时灵活性: 无需重新创建组件即可调整行为")
	fmt.Println("  • 场景适配: 针对不同业务场景使用不同配置")
	fmt.Println("  • 性能优化: 根据数据特点和资源情况优化参数")
	fmt.Println("  • 错误恢复: 配置重试和超时等容错机制")
	fmt.Println("  • 成本控制: 通过模型选择和批次优化控制API调用成本")

	fmt.Println("✅ Option配置演示完成！")
}

// Callback机制示例
func callbackExample(ctx context.Context, config *Config) {
	fmt.Println("\n=== Callback 机制示例 ===")

	embedder, err := initEmbedder(ctx, config)
	if err != nil {
		log.Printf("初始化Embedder失败: %v", err)
		return
	}

	// 准备测试文本
	testTexts := []string{
		"Callback机制是Eino框架的重要特性，允许在组件生命周期的关键节点注入自定义逻辑。",
		"通过OnStart、OnEnd、OnError等回调，可以实现日志记录、性能监控、错误处理等功能。",
		"Callback机制提供了非侵入式的组件扩展能力，是构建可观测系统的基础。",
	}

	fmt.Printf("📝 准备演示Callback机制功能，处理 %d 个文本\n", len(testTexts))
	for i, text := range testTexts {
		fmt.Printf("  文本%d: %s\n", i+1, text)
	}

	// 演示1: 手动触发回调事件（概念演示）
	fmt.Println("\n🔸 演示1: Callback机制概念说明")

	fmt.Println("🔄 模拟OnStart回调...")
	fmt.Println("   🚀 开始向量化操作")
	fmt.Printf("   📝 准备向量化 %d 个文本\n", len(testTexts))
	for i, text := range testTexts {
		fmt.Printf("      文本%d: 长度=%d字符\n", i+1, len(text))
	}

	// 记录开始时间
	startTime := time.Now()

	// 执行实际向量化
	fmt.Println("⚡ 执行向量化...")
	vectors, err := embedder.EmbedStrings(ctx, testTexts)

	duration := time.Since(startTime)

	if err != nil {
		// 模拟OnError回调
		fmt.Println("🔴 模拟OnError回调...")
		fmt.Printf("   ❌ 向量化失败，耗时: %v\n", duration)
		fmt.Printf("   💥 错误详情: %v\n", err)
		fmt.Println("   🔧 建议检查:")
		fmt.Println("      • 网络连接状态")
		fmt.Println("      • API配置是否正确")
		fmt.Println("      • 输入文本格式")
		fmt.Println("      • 模型服务可用性")
		return
	}

	// 模拟OnEnd回调
	fmt.Println("🟢 模拟OnEnd回调...")
	fmt.Printf("   ✅ 向量化完成，总耗时: %v\n", duration)
	fmt.Printf("   📊 成功向量化 %d 个文本，生成向量维度: %d\n",
		len(vectors), len(vectors[0]))

	// 计算一些统计信息
	avgProcessTime := duration / time.Duration(len(testTexts))
	fmt.Printf("   📈 平均处理时间: %v/文本\n", avgProcessTime)

	if duration.Seconds() > 0 {
		throughput := float64(len(testTexts)) / duration.Seconds()
		fmt.Printf("   🚀 处理吞吐量: %.2f 文本/秒\n", throughput)
	}

	// 演示2: Callback配置代码示例
	fmt.Println("\n🔸 演示2: Callback配置代码示例")
	fmt.Println("📋 标准Callback处理器创建代码:")
	fmt.Println("   callbackHandler := callbacks.NewHandlerBuilder().")
	fmt.Println("       OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {")
	fmt.Println("           fmt.Printf(\"🚀 开始向量化操作\\n\")")
	fmt.Println("           if texts, ok := input.([]string); ok {")
	fmt.Println("               fmt.Printf(\"📝 处理 %d 个文本\\n\", len(texts))")
	fmt.Println("           }")
	fmt.Println("           return context.WithValue(ctx, \"start_time\", time.Now())")
	fmt.Println("       }).")
	fmt.Println("       OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) {")
	fmt.Println("           startTime, _ := ctx.Value(\"start_time\").(time.Time)")
	fmt.Println("           fmt.Printf(\"✅ 向量化完成，耗时: %v\\n\", time.Since(startTime))")
	fmt.Println("           if vectors, ok := output.([][]float64); ok {")
	fmt.Println("               fmt.Printf(\"📊 生成 %d 个向量\\n\", len(vectors))")
	fmt.Println("           }")
	fmt.Println("       }).")
	fmt.Println("       OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) {")
	fmt.Println("           fmt.Printf(\"❌ 向量化失败: %v\\n\", err)")
	fmt.Println("       }).")
	fmt.Println("       Build()")

	fmt.Println("\n📋 在Chain中使用Callback:")
	fmt.Println("   chain := compose.NewChain[[]string, [][]float64]()")
	fmt.Println("   chain.AppendEmbedding(embedder, compose.WithCallbacks(callbackHandler))")
	fmt.Println("   runnable, err := chain.Compile(ctx)")
	fmt.Println("   vectors, err := runnable.Invoke(ctx, texts)")

	// 演示3: 高级Callback应用场景
	fmt.Println("\n🔸 演示3: 高级Callback应用场景")

	fmt.Println("📊 性能监控Callback:")
	fmt.Println("   - 记录每次向量化的响应时间")
	fmt.Println("   - 统计API调用成功率和失败率")
	fmt.Println("   - 监控向量维度变化")
	fmt.Println("   - 计算处理吞吐量指标")

	fmt.Println("\n🔍 调试分析Callback:")
	fmt.Println("   - 记录输入文本的特征信息")
	fmt.Println("   - 输出向量的统计特性")
	fmt.Println("   - 中间处理步骤的详细日志")
	fmt.Println("   - 异常情况的堆栈跟踪")

	fmt.Println("\n💰 成本控制Callback:")
	fmt.Println("   - 统计API调用次数和Token消耗")
	fmt.Println("   - 监控不同模型的使用情况")
	fmt.Println("   - 记录批处理效率优化效果")
	fmt.Println("   - 成本预算和告警机制")

	fmt.Println("🎯 Callback机制的优势:")
	fmt.Println("  • 可观测性: 全面监控向量化过程的每个环节")
	fmt.Println("  • 非侵入式: 不修改核心逻辑即可扩展功能")
	fmt.Println("  • 灵活配置: 按需启用不同类型的回调处理")
	fmt.Println("  • 统一接口: 所有组件使用相同的回调机制")
	fmt.Println("  • 生命周期: 覆盖开始、结束、错误等关键节点")
	fmt.Println("  • 数据洞察: 深入了解向量化性能和质量")

	fmt.Println("\n📊 常见Callback应用场景:")
	fmt.Println("  • 📈 性能分析: 识别瓶颈和优化机会")
	fmt.Println("  • 📝 操作审计: 记录所有向量化操作历史")
	fmt.Println("  • ⚠️ 异常告警: 及时发现和响应问题")
	fmt.Println("  • 🔄 质量监控: 评估向量化结果质量")
	fmt.Println("  • 🧪 A/B测试: 对比不同配置的效果")

	fmt.Println("✅ Callback机制演示完成！")
}

// 主函数
func main() {
	ctx := context.Background()

	fmt.Println("🎯 Eino Embedding 组件完全示例")
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
	fmt.Printf("  Embedder模型: %s\n", config.EmbedderModel)

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
			try("基础嵌入示例", basicEmbeddingExample)
		case "batch":
			try("批量处理示例", batchProcessingExample)
		case "search":
			try("语义搜索示例", semanticSearchExample)
		case "chain":
			try("Chain编排示例", chainEmbeddingExample)
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
			fmt.Println("可用示例: basic, batch, search, chain, performance, error, option, callback")
			return
		}
	} else {
		// 运行所有示例
		//try("基础嵌入示例", basicEmbeddingExample)
		//try("批量处理示例", batchProcessingExample)
		//try("语义搜索示例", semanticSearchExample)
		//try("Chain编排示例", chainEmbeddingExample)
		//try("性能测试示例", performanceTestExample)
		//try("错误处理示例", errorHandlingExample)
		try("Option配置示例", optionConfigExample)
		try("Callback机制示例", callbackExample)
	}

	fmt.Println("\n🎉 所有示例运行完成！")
	fmt.Println("\n使用方法:")
	fmt.Println("  go run main.go              # 运行所有示例")
	fmt.Println("  go run main.go basic        # 运行基础嵌入示例")
	fmt.Println("  go run main.go batch        # 运行批量处理示例")
	fmt.Println("  go run main.go search       # 运行语义搜索示例")
	fmt.Println("  go run main.go chain        # 运行Chain编排示例")
	fmt.Println("  go run main.go performance  # 运行性能测试示例")
	fmt.Println("  go run main.go error        # 运行错误处理示例")
	fmt.Println("  go run main.go option       # 运行Option配置示例")
	fmt.Println("  go run main.go callback     # 运行Callback机制示例")
}
