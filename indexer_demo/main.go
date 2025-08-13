package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/indexer/volc_vikingdb"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: indexer_demo/main.go
//  功能: 演示如何使用 Indexer 组件将文档存储到火山引擎 VikingDB。
//
// =============================================================================

func runIndexerExample() {
	ctx := context.Background()

	// --- 1. 配置并初始化 Indexer ---
	// 从 viper 加载 VikingDB 所需的配置。
	// 请确保这些值已在 config.yaml 或环境变量中设置。
	ak := viper.GetString("VIKINGDB_AK")
	sk := viper.GetString("VIKINGDB_SK")
	host := viper.GetString("VIKINGDB_HOST")
	region := viper.GetString("VIKINGDB_REGION")
	collectionName := viper.GetString("VIKINGDB_COLLECTION")

	if ak == "" || sk == "" || host == "" || region == "" || collectionName == "" {
		log.Fatal("VikingDB 配置 (AK, SK, Host, Region, Collection) 必须被设置！")
	}

	// 根据文档，创建一个 volc_vikingdb.IndexerConfig
	cfg := &volc_vikingdb.IndexerConfig{
		Host:       host,
		Region:     region,
		AK:         ak,
		SK:         sk,
		Scheme:     "https",
		Collection: collectionName,
		EmbeddingConfig: volc_vikingdb.EmbeddingConfig{
			UseBuiltin: true,           // 使用 VikingDB 内置的 Embedding 功能
			ModelName:  "bge-large-zh", // 指定内置模型
			UseSparse:  false,          // 本示例不使用稀疏向量
		},
	}

	// 创建 Indexer 实例
	indexer, err := volc_vikingdb.NewIndexer(ctx, cfg)
	if err != nil {
		log.Fatalf("创建 Indexer 失败: %v", err)
	}
	fmt.Println("Indexer 初始化成功！")

	// --- 2. 准备待存储的文档 ---
	// 我们创建两个 schema.Document 对象。
	// 每个文档都有唯一的 ID, 内容, 以及一些自定义的元数据字段。
	doc1 := &schema.Document{
		ID:      "eino-doc-001",
		Content: "Eino 是一个云原生的大模型应用开发框架，旨在简化和加速大模型应用的构建。",
	}
	// 使用辅助函数设置额外的元数据字段，这些字段需要在 VikingDB 的 collection 中预先定义。
	volc_vikingdb.SetExtraDataFields(doc1, map[string]interface{}{"source": "official_docs", "author": "CloudWeGo"})

	doc2 := &schema.Document{
		ID:      "eino-doc-002",
		Content: "RAG (Retrieval-Augmented Generation) 是一种结合了检索和生成两大功能的AI技术。",
	}
	volc_vikingdb.SetExtraDataFields(doc2, map[string]interface{}{"source": "tech_blog", "author": "AI_Researcher"})

	docs := []*schema.Document{doc1, doc2}
	fmt.Println("\n准备存储以下文档:")
	for _, doc := range docs {
		fmt.Printf("  - ID: %s, 内容: %s\n", doc.ID, doc.Content)
	}

	// --- 3. 调用 Store 方法 ---
	// 将文档列表传入 Store 方法，Indexer 会处理后续的向量化和存储流程。
	fmt.Println("\n正在调用 Store 方法...")
	storedIDs, err := indexer.Store(ctx, docs)
	if err != nil {
		log.Fatalf("存储文档失败: %v", err)
	}

	// --- 4. 打印结果 ---
	// 打印返回的 ID 列表，确认文档已成功存储。
	fmt.Println("\n--- 存储成功 ---")
	fmt.Printf("返回的文档 IDs: %v\n", storedIDs)
}

func main() {
	// 初始化 viper 以加载配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	viper.AutomaticEnv() // 允许从环境变量读取，例如 VIKINGDB_AK
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("未找到 config.yaml 文件，将仅从环境变量读取配置。")
	}

	runIndexerExample()
}
