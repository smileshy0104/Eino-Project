package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"Eini/retriever_demo/chain_example" // 使用 go.mod 中的模块路径导入

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: retriever_demo/main.go
//  功能: 演示如何独立使用 Retriever 组件从 Milvus 检索文档。
//        同时作为 retriever_demo 目录的统一程序入口。
//  前置: 运行此示例前，请确保已运行过 indexer_demo 来向 Milvus 中填充数据。
//
// =============================================================================

func runRetrieverExample() {
	ctx := context.Background()

	// --- 0. 初始化 Embedder ---
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
	address := viper.GetString("MILVUS_ADDRESS")
	collectionName := viper.GetString("MILVUS_COLLECTION")
	if address == "" || collectionName == "" {
		log.Fatal("Milvus 配置 (Address, Collection) 必须被设置！")
	}
	client, err := cli.NewClient(ctx, cli.Config{
		Address: address,
	})
	if err != nil {
		log.Fatalf("创建 Milvus 客户端失败: %v", err)
	}

	cfg := &milvus.RetrieverConfig{
		Client:       client,
		Collection:   collectionName,
		Embedding:    embedder,
		OutputFields: []string{"content", "metadata"},
	}
	retriever, err := milvus.NewRetriever(ctx, cfg)
	if err != nil {
		log.Fatalf("创建 Retriever 失败: %v", err)
	}
	fmt.Println("Retriever 初始化成功！")

	// --- 2. 调用 Retrieve 方法 ---
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

// main 是 retriever_demo 目录的唯一程序入口。
func main() {
	// 初始化 viper 以加载配置。
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")              // 在当前目录查找配置文件
	viper.AddConfigPath("../")             // 也在上一级目录查找
	viper.AddConfigPath("./chain_example") // 也在子目录查找
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("未找到 config.yaml 文件，将仅从环境变量读取配置。")
	}

	// --- 选择要运行的示例 ---
	exampleToRun := "rag" // 可选值: "standalone", "rag"

	switch exampleToRun {
	case "standalone":
		fmt.Println("--- 正在运行: 独立 Retriever 示例 ---")
		runRetrieverExample()
	case "rag":
		fmt.Println("\n--- 正在运行: RAG Chain 示例 ---")
		chain_example.Run()
	default:
		fmt.Println("无效的示例名称。请在 main.go 中设置 exampleToRun 为 'standalone' 或 'rag'。")
	}
}
