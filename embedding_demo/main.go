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
	fmt.Printf("  API Key: %s\n", config.APIKey[:10]+"...")
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
		default:
			fmt.Printf("未知示例: %s\n", exampleName)
			fmt.Println("可用示例: basic, batch, search, chain, performance, error")
			return
		}
	} else {
		// 运行所有示例
		//try("基础嵌入示例", basicEmbeddingExample)
		//try("批量处理示例", batchProcessingExample)
		//try("语义搜索示例", semanticSearchExample)
		try("Chain编排示例", chainEmbeddingExample)
		//try("性能测试示例", performanceTestExample)
		//try("错误处理示例", errorHandlingExample)
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
}
