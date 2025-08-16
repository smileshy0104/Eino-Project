package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: retriever_demo/main.go
//  功能: 演示如何独立使用 Retriever 组件从 Milvus 检索文档。
//  前置: 运行此示例前，请确保已运行过 indexer_demo 来向 Milvus 中填充数据。
//
// =============================================================================

func runRetrieverExample() {
	ctx := context.Background()

	// --- 0. 初始化 Embedder ---
	// Retriever 需要一个 Embedder 组件来将查询文本转换为向量。
	// 这里的配置必须与 indexer_demo 中使用的 Embedder 保持一致。
	timeout := 30 * time.Second
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("EMBEDDER_MODEL"),
		Timeout: &timeout,
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}

	// --- 1. 配置并初始化 Retriever ---
	// 从 viper 加载 Milvus 所需的配置。
	address := viper.GetString("MILVUS_ADDRESS")
	collectionName := viper.GetString("MILVUS_COLLECTION")

	if address == "" || collectionName == "" {
		log.Fatal("Milvus 配置 (Address, Collection) 必须被设置！")
	}

	// 创建一个 Milvus Go SDK 的客户端实例。
	client, err := cli.NewClient(ctx, cli.Config{
		Address: address,
	})
	if err != nil {
		log.Fatalf("创建 Milvus 客户端失败: %v", err)
	}

	// 创建 Retriever 的配置。
	cfg := &milvus.RetrieverConfig{
		Client:     client,
		Collection: collectionName,
		Embedding:  embedder,
		// 关键步骤：指定在检索结果中需要返回的字段。
		// Milvus 默认只返回 ID 和 score，我们需要显式要求它返回 content 和 metadata。
		// 这里的字段名必须与 indexer_demo 中定义的 Milvus schema 字段名完全匹配。
		OutputFields: []string{"content", "metadata"},
	}

	// 使用配置创建 Retriever 实例。
	retriever, err := milvus.NewRetriever(ctx, cfg)
	if err != nil {
		log.Fatalf("创建 Retriever 失败: %v", err)
	}
	fmt.Println("Retriever 初始化成功！")

	// --- 2. 调用 Retrieve 方法 ---
	// 准备一个查询，并调用 Retrieve 方法。
	// Retriever 内部会：
	// 1. 使用 Embedder 将查询文本转换为向量。
	// 2. 调用 Milvus Go SDK 的 `Search` 方法执行向量相似度搜索。
	// 3. 将搜索结果转换回 []*schema.Document 格式。
	query := "Eino 是什么？"
	fmt.Printf("\n正在使用查询 \"%s\" 调用 Retrieve 方法...\n", query)
	retrievedDocs, err := retriever.Retrieve(ctx, query)
	if err != nil {
		log.Fatalf("检索文档失败: %v", err)
	}

	// --- 3. 打印检索结果 ---
	fmt.Println("\n--- 检索成功 ---")
	if len(retrievedDocs) == 0 {
		fmt.Println("未检索到相关文档。请确保您已先运行 indexer_demo 来存储文档。")
		return
	}
	fmt.Printf("检索到 %d 个相关文档:\n", len(retrievedDocs))
	for _, doc := range retrievedDocs {
		fmt.Printf("  - ID: %s, 内容: %s, 元数据: %v\n", doc.ID, doc.Content, doc.MetaData)
	}
}

// main 是程序的入口点。
func main() {
	// 初始化 viper 以加载配置。
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./") // 在当前目录查找配置文件

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("未找到 config.yaml 文件，将仅从环境变量读取配置。")
	}

	runRetrieverExample()
}
