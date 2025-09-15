package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/schema"
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

// 工具函数: 截取字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// 工具函数: 获取第一行
func getFirstLine(content string) string {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) > 0 {
		return truncateString(lines[0], 50)
	}
	return "(空内容)"
}

// 生成大型内容
func generateLargeContent() string {
	content := "# 大型文档测试\n\n这是一个用于测试大型文档处理能力的文档。\n\n"
	for i := 1; i <= 50; i++ {
		content += fmt.Sprintf("## 章节 %d\n\n", i)
		content += fmt.Sprintf("这是第 %d 个章节的内容。", i)
		content += "包含了大量的文本内容用于测试处理性能。"
		content += "模拟真实场景中的长文档处理需求。"
		content += "确保系统能够稳定处理大量数据。\n\n"
	}
	return content
}

// 1. 基础转换示例
func basicTransformExample(ctx context.Context) {
	fmt.Println("\n=== 基础转换示例 ===")

	// 准备原始文档
	document := &schema.Document{
		ID: "eino-guide-001",
		Content: `# Eino 框架完全指南

Eino 是一个为简化和加速大模型应用构建而设计的云原生开发框架。

## 核心组件介绍

Eino 提供了多种核心组件，包括 Model、Retriever、Indexer 和 Transformer。这些组件可以帮助开发者快速构建强大的 RAG 应用。

### Model 组件

Model 组件是与大语言模型交互的核心接口，支持多种主流模型提供商。

### Retriever 组件

Retriever 组件负责从向量数据库中检索相关文档，支持多种检索策略。

## Transformer 组件详解

Transformer 组件负责文档的预处理工作。它可以将长文档分割成语义完整的小块，过滤无关信息，或进行格式转换。

### 分割策略

- Markdown 标题分割：按照标题层级进行分割
- 文本分割：按照字符数或句子进行分割
- 自定义分割：根据业务需求定制分割逻辑

## 快速开始

要开始使用 Eino，请参考我们的官方文档和示例代码。我们提供了丰富的教程和最佳实践指南。

### 安装和配置

1. 使用 go mod 安装 Eino
2. 配置 API 密钥和数据库连接
3. 初始化所需的组件

### 第一个示例

从一个简单的文档检索示例开始，逐步探索 Eino 的强大功能。`,
		MetaData: map[string]interface{}{
			"source":     "official_guide",
			"type":       "documentation",
			"language":   "zh-CN",
			"version":    "1.0",
			"word_count": 384,
			"sections":   4,
			"created_at": time.Now().Format(time.RFC3339),
		},
	}

	fmt.Printf("📝 准备转换文档: %s\n", document.ID)
	fmt.Printf("📊 文档字数: %d 字符\n", len(document.Content))
	fmt.Printf("🏷️ 元数据: %v\n", document.MetaData)

	// 初始化 Markdown Header Splitter
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"##":  "Header 2", // 使用二级标题作为主要分割点
			"###": "Header 3", // 使用三级标题作为子分割点
		},
	})
	if err != nil {
		log.Printf("初始化 HeaderSplitter 失败: %v", err)
		return
	}

	// 执行文档转换
	fmt.Println("\n🔄 正在执行文档转换...")
	startTime := time.Now()
	chunks, err := splitter.Transform(ctx, []*schema.Document{document})
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("❌ 转换失败: %v\n", err)
		return
	}

	// 显示转换结果
	fmt.Printf("\n✅ 转换完成，共生成 %d 个文档块（耗时: %v）\n", len(chunks), duration)
	for i, chunk := range chunks {
		fmt.Printf("\n--- 文档块 %d ---\n", i+1)
		fmt.Printf("ID: %s\n", chunk.ID)
		fmt.Printf("内容预览: %s...\n", truncateString(chunk.Content, 100))
		fmt.Printf("字符数: %d\n", len(chunk.Content))
		fmt.Printf("元数据: %v\n", chunk.MetaData)
	}

	fmt.Println("✅ 基础转换示例完成！")
}

// 2. Option配置示例
func optionConfigExample(ctx context.Context) {
	fmt.Println("\n=== Option配置示例 ===")

	// 准备测试文档
	document := &schema.Document{
		ID: "complex-doc-001",
		Content: `# 复杂文档示例

这是一个用于测试不同配置选项的复杂文档。

## 第一章：基础概念

这里介绍了一些基础概念。内容比较详细，包含了多个方面的介绍。

### 1.1 子章节 A

这是第一个子章节的内容。

### 1.2 子章节 B

这是第二个子章节的内容。

## 第二章：进阶内容

这里介绍了进阶内容。

### 2.1 高级特性

详细介绍高级特性。

## 第三章：实际应用

实际应用的案例和示例。`,
		MetaData: map[string]interface{}{
			"source":   "test_document",
			"category": "tutorial",
			"level":    "intermediate",
		},
	}

	fmt.Printf("📝 测试文档: %s\n", document.ID)

	// 测试不同配置
	configs := []struct {
		name    string
		headers map[string]string
		desc    string
	}{
		{
			name:    "简单配置",
			headers: map[string]string{"##": "Header 2"},
			desc:    "仅使用二级标题分割",
		},
		{
			name:    "标准配置",
			headers: map[string]string{"##": "Header 2", "###": "Header 3"},
			desc:    "使用二、三级标题分割",
		},
		{
			name:    "完整配置",
			headers: map[string]string{"#": "Header 1", "##": "Header 2", "###": "Header 3"},
			desc:    "使用一、二、三级标题分割",
		},
	}

	for i, cfg := range configs {
		fmt.Printf("\n🔹 配置 %d: %s (%s)\n", i+1, cfg.name, cfg.desc)

		splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
			Headers: cfg.headers,
		})
		if err != nil {
			log.Printf("初始化 Splitter 失败: %v", err)
			continue
		}

		startTime := time.Now()
		chunks, err := splitter.Transform(ctx, []*schema.Document{document})
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("  ❌ 转换失败: %v\n", err)
			continue
		}

		fmt.Printf("结果: 生成 %d 个文档块，耗时: %v\n", len(chunks), duration)
		for j, chunk := range chunks {
			fmt.Printf("  块 %d: %s (%d 字符)\n", j+1, getFirstLine(chunk.Content), len(chunk.Content))
		}
	}

	fmt.Println("\n💡 配置对比结果:")
	fmt.Printf("  - 简单配置: 较少但较大的块，适合结构简单的文档\n")
	fmt.Printf("  - 标准配置: 平衡的块大小，适合大多数情况\n")
	fmt.Printf("  - 完整配置: 较多但较小的块，适合结构复杂的文档\n")

	fmt.Println("✅ Option配置示例完成！")
}

// 3. 高级配置示例
func advancedConfigExample(ctx context.Context) {
	fmt.Println("\n=== 高级配置示例 ===")

	// 准备复杂文档
	document := &schema.Document{
		ID: "advanced-config-001",
		Content: `# 高级配置指南

这是一个用于测试高级配置的文档。

## 第一部分：基础配置

基础配置包括了最常用的参数设置。

### 1.1 基本参数

这里介绍基本参数的设置方法。

### 1.2 可选参数

可选参数提供了更灵活的配置方式。

## 第二部分：高级配置

高级配置适用于复杂的使用场景。

### 2.1 性能优化

性能优化配置可以提高处理效率。

### 2.2 安全设置

安全设置确保系统的安全性。

## 第三部分：最佳实践

最佳实践包括了多年来的经验总结。`,
		MetaData: map[string]interface{}{
			"source":      "advanced_guide",
			"complexity":  "high",
			"sections":    3,
			"subsections": 6,
		},
	}

	fmt.Printf("📝 测试高级配置: %s\n", document.ID)
	fmt.Printf("📊 文档复杂度: %v\n", document.MetaData["complexity"])

	// 测试不同的配置组合
	configs := []struct {
		name    string
		headers map[string]string
		desc    string
	}{
		{
			name:    "简单配置",
			headers: map[string]string{"##": "Header 2"},
			desc:    "仅使用二级标题分割",
		},
		{
			name:    "标准配置",
			headers: map[string]string{"##": "Header 2", "###": "Header 3"},
			desc:    "使用二、三级标题分割",
		},
		{
			name:    "完整配置",
			headers: map[string]string{"#": "Header 1", "##": "Header 2", "###": "Header 3"},
			desc:    "使用一、二、三级标题分割",
		},
	}

	for i, cfg := range configs {
		fmt.Printf("\n🔹 配置 %d: %s (%s)\n", i+1, cfg.name, cfg.desc)

		splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
			Headers: cfg.headers,
		})
		if err != nil {
			log.Printf("初始化 Splitter 失败: %v", err)
			continue
		}

		startTime := time.Now()
		chunks, err := splitter.Transform(ctx, []*schema.Document{document})
		duration := time.Since(startTime)

		if err != nil {
			log.Printf("转换失败: %v", err)
			continue
		}

		fmt.Printf("结果: 生成 %d 个块，耗时: %v\n", len(chunks), duration)

		// 统计信息
		totalChars := 0
		minChars := math.MaxInt32
		maxChars := 0
		for _, chunk := range chunks {
			chars := len(chunk.Content)
			totalChars += chars
			if chars < minChars {
				minChars = chars
			}
			if chars > maxChars {
				maxChars = chars
			}
		}

		avgChars := 0
		if len(chunks) > 0 {
			avgChars = totalChars / len(chunks)
		}

		fmt.Printf("统计: 平均 %d 字符/块，最小 %d，最大 %d\n", avgChars, minChars, maxChars)
	}

	fmt.Println("\n💡 配置选择建议:")
	fmt.Println("  - 简单配置: 适用于结构简单的文档")
	fmt.Println("  - 标准配置: 适用于大多数情况")
	fmt.Println("  - 完整配置: 适用于结构复杂的文档")

	fmt.Println("✅ 高级配置示例完成！")
}

// 4. 批量处理示例
func batchProcessingExample(ctx context.Context) {
	fmt.Println("\n=== 批量处理示例 ===")

	// 准备多个文档用于批量处理
	documents := []*schema.Document{
		{
			ID: "batch-001",
			Content: `# AI 技术概述

AI 技术正在改变世界。

## 机器学习

机器学习是 AI 的子集。

## 深度学习

深度学习是机器学习的子集。`,
			MetaData: map[string]interface{}{"category": "tech", "level": "intro"},
		},
		{
			ID: "batch-002",
			Content: `# 云计算基础

云计算提供灵活的计算资源。

## IaaS 服务

基础设施即服务。

## PaaS 服务

平台即服务。`,
			MetaData: map[string]interface{}{"category": "cloud", "level": "basic"},
		},
		{
			ID: "batch-003",
			Content: `# 区块链技术

区块链是一种分布式账本技术。

## 共识机制

如何达成共识。

## 智能合约

可编程的合约。`,
			MetaData: map[string]interface{}{"category": "blockchain", "level": "advanced"},
		},
	}

	fmt.Printf("📝 准备批量处理 %d 个文档\n", len(documents))

	// 创建 Transformer
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"##": "Header 2",
		},
	})
	if err != nil {
		log.Printf("初始化 Splitter 失败: %v", err)
		return
	}

	// 批量转换文档
	fmt.Println("\n🔄 开始批量转换...")
	startTime := time.Now()

	allChunks, err := splitter.Transform(ctx, documents)
	if err != nil {
		log.Printf("批量转换失败: %v", err)
		return
	}

	duration := time.Since(startTime)

	// 显示批量处理结果
	fmt.Printf("\n✅ 批量处理完成\n")
	fmt.Printf("📊 处理统计:\n")
	fmt.Printf("  - 输入文档: %d 个\n", len(documents))
	fmt.Printf("  - 输出块: %d 个\n", len(allChunks))
	fmt.Printf("  - 处理时间: %v\n", duration)
	fmt.Printf("  - 平均每文档: %.2f 个块\n", float64(len(allChunks))/float64(len(documents)))

	// 按原文档统计结果
	fmt.Println("\n📈 分组统计:")
	docStats := make(map[string]int)
	for _, chunk := range allChunks {
		if originalID, ok := chunk.MetaData["original_id"].(string); ok {
			docStats[originalID]++
		}
	}

	for docID, count := range docStats {
		fmt.Printf("  - %s: %d 个块\n", docID, count)
	}

	// 显示部分结果示例
	fmt.Println("\n🔍 部分结果示例:")
	for i, chunk := range allChunks {
		if i >= 3 {
			fmt.Printf("  ... 其余 %d 个块\n", len(allChunks)-3)
			break
		}
		fmt.Printf("  块 %d: %s (%d 字符)\n", i+1, getFirstLine(chunk.Content), len(chunk.Content))
	}

	fmt.Println("✅ 批量处理示例完成！")
}

// 5. 性能测试示例
func performanceTestExample(ctx context.Context) {
	fmt.Println("\n=== 性能测试示例 ===")

	// 生成测试数据
	generateTestDoc := func(id string, sections int) *schema.Document {
		content := fmt.Sprintf("# 测试文档 %s\n\n这是一个用于性能测试的文档。\n\n", id)
		for i := 1; i <= sections; i++ {
			content += fmt.Sprintf("## 章节 %d\n\n", i)
			content += fmt.Sprintf("这是第 %d 个章节的内容。", i)
			content += "包含了详细的介绍和说明，用于测试文档处理的性能表现。\n\n"
		}
		return &schema.Document{
			ID:      fmt.Sprintf("perf-test-%s", id),
			Content: content,
			MetaData: map[string]interface{}{
				"test_type": "performance",
				"sections":  sections,
			},
		}
	}

	// 测试场景
	testScenarios := []struct {
		name        string
		documents   []*schema.Document
		description string
	}{
		{
			name:        "轻载测试",
			documents:   []*schema.Document{generateTestDoc("light-1", 3), generateTestDoc("light-2", 3)},
			description: "少量文档，测试基础性能",
		},
		{
			name: "中载测试",
			documents: []*schema.Document{
				generateTestDoc("medium-1", 5), generateTestDoc("medium-2", 5),
				generateTestDoc("medium-3", 5), generateTestDoc("medium-4", 5),
			},
			description: "中等文档量，测试稳定性",
		},
		{
			name: "重载测试",
			documents: []*schema.Document{
				generateTestDoc("heavy-1", 10), generateTestDoc("heavy-2", 10),
				generateTestDoc("heavy-3", 10), generateTestDoc("heavy-4", 10),
				generateTestDoc("heavy-5", 10), generateTestDoc("heavy-6", 10),
			},
			description: "大量文档，测试极限性能",
		},
	}

	// 创建 Transformer
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"##": "Header 2",
		},
	})
	if err != nil {
		log.Printf("初始化 Splitter 失败: %v", err)
		return
	}

	fmt.Printf("📝 准备性能压测，使用 %d 个测试场景\n", len(testScenarios))

	for _, scenario := range testScenarios {
		fmt.Printf("\n🧪 执行 %s (%s)\n", scenario.name, scenario.description)

		startTime := time.Now()
		totalChunks := 0
		successCount := 0
		var minDuration, maxDuration time.Duration
		var totalDuration time.Duration

		for i, doc := range scenario.documents {
			docStartTime := time.Now()
			chunks, err := splitter.Transform(ctx, []*schema.Document{doc})
			docDuration := time.Since(docStartTime)

			if err != nil {
				fmt.Printf("  ❌ 文档%d处理失败: %v\n", i+1, err)
				continue
			}

			totalChunks += len(chunks)
			successCount++
			totalDuration += docDuration

			if minDuration == 0 || docDuration < minDuration {
				minDuration = docDuration
			}
			if docDuration > maxDuration {
				maxDuration = docDuration
			}

			fmt.Printf("  ✅ 文档%d: %d个块，耗时:%v\n", i+1, len(chunks), docDuration)
		}

		scenarioDuration := time.Since(startTime)

		// 性能统计
		fmt.Printf("\n📊 %s 统计:\n", scenario.name)
		fmt.Printf("  • 总文档数: %d\n", len(scenario.documents))
		fmt.Printf("  • 成功处理: %d\n", successCount)
		fmt.Printf("  • 总块数: %d\n", totalChunks)
		fmt.Printf("  • 总耗时: %v\n", scenarioDuration)

		if successCount > 0 {
			avgDuration := totalDuration / time.Duration(successCount)
			fmt.Printf("  • 平均耗时/文档: %v\n", avgDuration)
			fmt.Printf("  • 最快: %v\n", minDuration)
			fmt.Printf("  • 最慢: %v\n", maxDuration)

			if scenarioDuration.Seconds() > 0 {
				throughput := float64(successCount) / scenarioDuration.Seconds()
				fmt.Printf("  • 处理吞吐量: %.2f 文档/秒\n", throughput)

				chunkThroughput := float64(totalChunks) / scenarioDuration.Seconds()
				fmt.Printf("  • 块吞吐量: %.2f 块/秒\n", chunkThroughput)
			}
		}
	}

	fmt.Println("✅ 性能测试示例完成！")
}

// 6. 错误处理示例
func errorHandlingExample(ctx context.Context) {
	fmt.Println("\n=== 错误处理示例 ===")

	// 测试各种错误情况
	errorCases := []struct {
		name     string
		document *schema.Document
		headers  map[string]string
		desc     string
	}{
		{
			name: "空文档处理",
			document: &schema.Document{
				ID:      "empty-doc",
				Content: "",
			},
			headers: map[string]string{"##": "Header 2"},
			desc:    "测试空内容文档的处理",
		},
		{
			name: "无标题文档",
			document: &schema.Document{
				ID:      "no-header-doc",
				Content: "这是一个没有任何标题的文档内容。\n只有纯文本内容，没有 Markdown 标题。",
			},
			headers: map[string]string{"##": "Header 2"},
			desc:    "测试没有匹配标题的文档",
		},
		{
			name: "大型文档处理",
			document: &schema.Document{
				ID:      "large-doc",
				Content: generateLargeContent(),
			},
			headers: map[string]string{"##": "Header 2"},
			desc:    "测试大型文档的处理能力",
		},
	}

	fmt.Printf("📝 准备测试 %d 种错误场景\n", len(errorCases))

	for i, testCase := range errorCases {
		fmt.Printf("\n🧪 测试场景 %d: %s\n", i+1, testCase.name)
		fmt.Printf("📝 描述: %s\n", testCase.desc)
		fmt.Printf("📊 文档信息: ID=%s，内容长度=%d\n", testCase.document.ID, len(testCase.document.Content))

		// 创建 Splitter
		splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
			Headers: testCase.headers,
		})
		if err != nil {
			fmt.Printf("❌ Splitter 初始化失败: %v\n", err)
			continue
		}

		// 执行转换并处理错误
		startTime := time.Now()
		chunks, err := splitter.Transform(ctx, []*schema.Document{testCase.document})
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("❌ 转换失败: %v（耗时: %v）\n", err, duration)
		} else {
			fmt.Printf("✅ 转换成功: 生成 %d 个块（耗时: %v）\n", len(chunks), duration)

			// 显示第一个块的信息
			if len(chunks) > 0 {
				firstChunk := chunks[0]
				fmt.Printf("📬 第一个块: ID=%s，内容长度=%d\n", firstChunk.ID, len(firstChunk.Content))
				if len(firstChunk.Content) > 0 {
					preview := truncateString(firstChunk.Content, 50)
					fmt.Printf("🔍 内容预览: %s\n", preview)
				}
			}
		}
	}

	fmt.Println("\n💡 错误处理最佳实践:")
	fmt.Println("  - 始终检查返回的错误")
	fmt.Println("  - 对空文档进行预处理")
	fmt.Println("  - 设置合理的超时时间")
	fmt.Println("  - 记录详细的错误日志")

	fmt.Println("✅ 错误处理示例完成！")
}

// 7. 转换策略对比示例
func transformStrategyExample(ctx context.Context) {
	fmt.Println("\n=== 转换策略对比示例 ===")

	// 准备测试文档
	document := &schema.Document{
		ID: "strategy-test-001",
		Content: `# 第一个文档

这是第一个用于测试不同策略的文档。

## 章节 A

章节 A 的内容。

## 章节 B

章节 B 的内容。`,
		MetaData: map[string]interface{}{"source": "test", "index": 1},
	}

	fmt.Printf("📝 准备转换 %d 个文档\n", 1)

	// 创建 Transformer
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"##": "Header 2",
		},
	})
	if err != nil {
		log.Printf("初始化 Splitter 失败: %v", err)
		return
	}

	// 执行转换
	fmt.Println("\n🛠️ 执行文档转换...")
	result, err := splitter.Transform(ctx, []*schema.Document{document})
	if err != nil {
		log.Printf("执行转换失败: %v", err)
		return
	}

	fmt.Printf("\n📊 最终结果统计: 输入 %d 个文档，输出 %d 个块\n", 1, len(result))

	// 显示转换结果
	for i, chunk := range result {
		fmt.Printf("\n块 %d:\n", i+1)
		fmt.Printf("  ID: %s\n", chunk.ID)
		fmt.Printf("  内容预览: %s\n", getFirstLine(chunk.Content))
		fmt.Printf("  字符数: %d\n", len(chunk.Content))
		fmt.Printf("  元数据: %v\n", chunk.MetaData)
	}

	fmt.Println("✅ 转换策略对比示例完成！")
}

// 8. 策略对比示例
func comparisonExample(ctx context.Context) {
	fmt.Println("\n=== 策略对比示例 ===")

	document := &schema.Document{
		ID: "comparison-test",
		Content: `# 策略对比文档

## 策略 A

这是策略 A 的内容。

### A.1 子策略

策略 A 的子策略。

## 策略 B

这是策略 B 的内容。

### B.1 子策略

策略 B 的子策略。`,
	}

	strategies := []struct {
		name    string
		headers map[string]string
	}{
		{"仅二级标题", map[string]string{"##": "Header 2"}},
		{"二三级标题", map[string]string{"##": "Header 2", "###": "Header 3"}},
	}

	for _, strategy := range strategies {
		fmt.Printf("\n📋 测试策略: %s\n", strategy.name)

		splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
			Headers: strategy.headers,
		})
		if err != nil {
			log.Printf("初始化失败: %v", err)
			continue
		}

		chunks, err := splitter.Transform(ctx, []*schema.Document{document})
		if err != nil {
			log.Printf("转换失败: %v", err)
			continue
		}

		fmt.Printf("结果: %d 个块\n", len(chunks))
	}

	fmt.Println("✅ 策略对比示例完成！")
}

func main() {
	// 1. 初始化基础信息
	ctx := context.Background()
	fmt.Println("🚀 Eino Transformer 组件演示程序")
	fmt.Println("==================================================")
	fmt.Printf("📁 工作目录: %s\n", getCurrentDir())
	fmt.Printf("🕗 启动时间: %s\n", time.Now().Format(time.RFC3339))

	// 2. 加载配置
	config, err := initConfig()
	if err != nil {
		fmt.Printf("⚠️  配置加载失败（忽略）: %v\n", err)
	} else {
		fmt.Println("\n🛠️  当前配置:")
		if config.APIKey != "" {
			if len(config.APIKey) > 10 {
				fmt.Printf("  API Key: %s\n", config.APIKey[:10]+"...")
			} else {
				fmt.Printf("  API Key: %s\n", config.APIKey)
			}
		} else {
			fmt.Println("  API Key: 未配置")
		}
		if config.Model != "" {
			fmt.Printf("  模型: %s\n", config.Model)
		}
		if config.EmbedderModel != "" {
			fmt.Printf("  Embedder模型: %s\n", config.EmbedderModel)
		}
	}

	// 3. 运行示例
	try := func(name string, fn func(context.Context)) {
		fmt.Printf("\n正在运行: %s\n", name)
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("示例 %s 运行出错: %v\n", name, r)
			}
		}()
		fn(ctx)
	}

	// 检查命令行参数
	if len(os.Args) > 1 {
		exampleName := os.Args[1]
		switch strings.ToLower(exampleName) {
		case "basic":
			try("基础转换示例", basicTransformExample)
		case "option":
			try("Option配置示例", optionConfigExample)
		case "strategy":
			try("转换策略对比示例", transformStrategyExample)
		case "advanced":
			try("高级配置示例", advancedConfigExample)
		case "batch":
			try("批量处理示例", batchProcessingExample)
		case "performance":
			try("性能测试示例", performanceTestExample)
		case "error":
			try("错误处理示例", errorHandlingExample)
		case "comparison":
			try("策略对比示例", comparisonExample)
		default:
			fmt.Printf("未知示例: %s\n", exampleName)
			fmt.Println("可用示例: basic, option, strategy, advanced, batch, performance, error, comparison")
			return
		}
	} else {
		// 运行所有示例
		try("基础转换示例", basicTransformExample)
		try("Option配置示例", optionConfigExample)
		try("转换策略对比示例", transformStrategyExample)
		try("高级配置示例", advancedConfigExample)
		try("批量处理示例", batchProcessingExample)
		try("性能测试示例", performanceTestExample)
		try("错误处理示例", errorHandlingExample)
		try("策略对比示例", comparisonExample)
	}

	fmt.Println("\n🎉 所有示例运行完成！")
	fmt.Println("\n使用方法:")
	fmt.Println("  go run main.go              # 运行所有示例")
	fmt.Println("  go run main.go basic        # 运行基础转换示例")
	fmt.Println("  go run main.go option       # 运行Option配置示例")
	fmt.Println("  go run main.go strategy     # 运行转换策略对比示例")
	fmt.Println("  go run main.go advanced     # 运行高级配置示例")
	fmt.Println("  go run main.go batch        # 运行批量处理示例")
	fmt.Println("  go run main.go performance  # 运行性能测试示例")
	fmt.Println("  go run main.go error        # 运行错误处理示例")
	fmt.Println("  go run main.go comparison   # 运行策略对比示例")
}

// 获取当前目录
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}
