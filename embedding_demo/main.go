package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: embedding_demo/main.go
//  功能: 演示如何使用 Embedding 组件将文本转换为向量，并计算其相似度。
//
// =============================================================================

// cosineSimilarity 计算两个向量之间的余弦相似度。
// 返回值范围在 -1 到 1 之间，越接近 1 表示越相似。
func cosineSimilarity(v1, v2 []float64) (float64, error) {
	if len(v1) != len(v2) {
		return 0, fmt.Errorf("向量维度不匹配")
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

func runEmbeddingExample() {
	ctx := context.Background()

	// --- 1. 初始化 Embedder ---
	// 我们使用 ARK 的 Embedding 组件。
	// Viper 会自动从环境变量或配置文件中读取配置。
	// 同样，模型名称也可以通过 viper 从配置中读取，这里为了示例清晰，我们硬编码一个默认值。
	// 在真实项目中，推荐做法是：viper.GetString("EMBEDDING_MODEL")
	timeout := 30 * time.Second
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("EMBEDDER_MODEL"),
		Timeout: &timeout,
	})

	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}

	// --- 2. 定义输入文本 ---
	// 我们准备了三段文本，其中前两段语义相似，第三段则不相关。
	texts := []string{
		"今天天气真好，阳光明媚。",  // 文本 A
		"今天是个大晴天，万里无云。", // 文本 B (与 A 相似)
		"红烧肉怎么做才好吃？",    // 文本 C (与 A, B 不相似)
	}
	fmt.Println("输入文本:")
	for i, t := range texts {
		fmt.Printf("  %c: %s\n", 'A'+i, t)
	}

	// --- 3. 调用 EmbedStrings 获取向量 ---
	fmt.Println("\n正在将文本转换为向量...")
	vectors, err := embedder.EmbedStrings(ctx, texts)
	if err != nil {
		log.Fatalf("EmbedStrings 调用失败: %v", err)
	}
	fmt.Println("向量结果:", vectors)
	fmt.Println("向量转换完成！")

	// --- 4. 计算并打印相似度 ---
	// 我们将计算 (A, B) 和 (A, C) 两对文本的相似度。
	// 预期结果：(A, B) 的相似度远高于 (A, C)。
	vecA, vecB, vecC := vectors[0], vectors[1], vectors[2]

	simAB, err := cosineSimilarity(vecA, vecB)
	if err != nil {
		log.Fatalf("计算 A 和 B 的相似度失败: %v", err)
	}

	simAC, err := cosineSimilarity(vecA, vecC)
	if err != nil {
		log.Fatalf("计算 A 和 C 的相似度失败: %v", err)
	}

	fmt.Println("\n--- 相似度计算结果 ---")
	fmt.Printf("  - 文本 A 和 B (相似) 的相似度: %.4f\n", simAB)
	fmt.Printf("  - 文本 A 和 C (不相似) 的相似度: %.4f\n", simAC)
	fmt.Println("\n可以看到，语义相似的文本对获得了远高于不相似文本对的得分。")
}

func main() {
	// 为了能从 viper 加载配置，先进行初始化
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	_ = viper.ReadInConfig() // 忽略错误，因为我们也会检查环境变量

	runEmbeddingExample()
}
