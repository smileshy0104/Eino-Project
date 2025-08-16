package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	indexer "github.com/cloudwego/eino-ext/components/indexer/milvus"
	retriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/schema"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: transformer_demo/main.go
//  功能: 演示一个完整的、端到端的 RAG 数据处理流水线：
//        1. Transformer: 将长文档按标题分割成小块。
//        2. Indexer: 将分割后的文档块向量化并存入 Milvus。
//        3. Retriever: 从 Milvus 中精确检索与查询相关的文档块。
//
// =============================================================================

var fields = []*entity.Field{
	{
		Name:        "id",
		DataType:    entity.FieldTypeVarChar,
		TypeParams:  map[string]string{"max_length": "255"},
		PrimaryKey:  true,
		Description: "文档块的唯一主键",
	},
	{
		Name:        "vector",
		DataType:    entity.FieldTypeBinaryVector,
		TypeParams:  map[string]string{"dim": "8192"},
		Description: "文档块内容的向量表示",
	},
	{
		Name:        "content",
		DataType:    entity.FieldTypeVarChar,
		TypeParams:  map[string]string{"max_length": "8192"},
		Description: "原始的文本内容块",
	},
	{
		Name:        "metadata",
		DataType:    entity.FieldTypeJSON,
		Description: "用于存储附加信息的 JSON 字段",
	},
}

func main() {
	// --- 步骤 0: 加载配置 ---
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../") // 方便在 transformer_demo 目录内执行
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("未找到 config.yaml 文件，将仅从环境变量读取配置。")
	}

	ctx := context.Background()

	// --- 步骤 1: 准备原始长文档 ---
	longMarkdownDoc := &schema.Document{
		ID: "eino-intro-doc",
		Content: `
# Eino 框架介绍
Eino 是一个先进的大模型应用开发框架。
## 核心组件
Eino 提供了多种核心组件，包括 Model, Retriever, Indexer, 和 Transformer。这些组件可以帮助开发者快速构建强大的 RAG 应用。
## Transformer 详解
Transformer 组件负责文档的预处理。它可以将长文档分割成小块，过滤无关信息，或进行格式转换。这是确保检索质量的关键一步。
## 快速开始
要开始使用 Eino，请参考我们的官方文档和示例代码。`,
		MetaData: map[string]interface{}{"source": "official-docs"},
	}
	fmt.Println("--- 准备好原始长文档 ---")

	// --- 步骤 2: 使用 Transformer 分割文档 ---
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{"##": "Header 2"},
	})
	if err != nil {
		log.Fatalf("创建 HeaderSplitter 失败: %v", err)
	}
	docsToStore, err := splitter.Transform(ctx, []*schema.Document{longMarkdownDoc})
	if err != nil {
		log.Fatalf("转换文档失败: %v", err)
	}
	fmt.Printf("\n--- Transformer 分割完成，原始文档被分割成 %d 个块 ---\n", len(docsToStore))
	for i, d := range docsToStore {
		fmt.Printf("  块 %d ID: %s, 内容预览: %.30s...\n", i+1, d.ID, d.Content)
	}

	// --- 步骤 3: 初始化通用组件 (Embedder, Milvus Client) ---
	timeout := 30 * time.Second
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("EMBEDDER_MODEL"),
		Timeout: &timeout,
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}
	address := viper.GetString("MILVUS_ADDRESS")
	collectionName := viper.GetString("MILVUS_COLLECTION")
	client, err := cli.NewClient(ctx, cli.Config{Address: address})
	if err != nil {
		log.Fatalf("创建 Milvus 客户端失败: %v", err)
	}

	// --- 步骤 4: 索引流程 (检查集合 -> 创建 -> 索引 -> 加载) ---
	// **注意**: 为确保演示可重复运行，我们先删除可能存在的旧集合。
	// 在生产环境中，您可能不需要每次都删除。
	_ = client.DropCollection(ctx, collectionName) // 忽略错误
	fmt.Printf("\n--- 开始索引流程 (集合: %s) ---\n", collectionName)
	schema := &entity.Schema{CollectionName: collectionName, Fields: fields}
	err = client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		log.Fatalf("创建集合失败: %v", err)
	}
	binFlatIndex, _ := entity.NewIndexBinFlat(entity.HAMMING, 128)
	err = client.CreateIndex(ctx, collectionName, "vector", binFlatIndex, false)
	if err != nil {
		log.Fatalf("为 'vector' 字段创建索引失败: %v", err)
	}
	fmt.Println("集合与索引创建成功！")

	indexerCfg := &indexer.IndexerConfig{
		Client: client, Collection: collectionName, Embedding: embedder, Fields: fields,
	}
	indexerComponent, _ := indexer.NewIndexer(ctx, indexerCfg)
	_, err = indexerComponent.Store(ctx, docsToStore)
	if err != nil {
		log.Fatalf("存储文档块失败: %v", err)
	}
	fmt.Println("文档块存储成功！")
	err = client.LoadCollection(ctx, collectionName, false)
	if err != nil {
		log.Fatalf("加载集合失败: %v", err)
	}
	fmt.Println("集合加载成功！")

	// --- 步骤 5: 检索流程 ---
	fmt.Println("\n--- 开始检索流程 ---")
	retrieverCfg := &retriever.RetrieverConfig{
		Client:       client,
		Collection:   collectionName,
		Embedding:    embedder,
		OutputFields: []string{"content", "metadata"},
	}
	retrieverComponent, err := retriever.NewRetriever(ctx, retrieverCfg)
	if err != nil {
		log.Fatalf("创建 Retriever 失败: %v", err)
	}

	query := "Transformer 是做什么的？"
	fmt.Printf("正在使用查询: \"%s\"\n", query)
	retrievedDocs, err := retrieverComponent.Retrieve(ctx, query)
	if err != nil {
		log.Fatalf("检索文档失败: %v", err)
	}

	// --- 步骤 6: 打印最终结果 ---
	fmt.Println("\n--- 检索成功 ---")
	if len(retrievedDocs) == 0 {
		fmt.Println("未检索到相关文档。")
	} else {
		fmt.Printf("检索到 %d 个最相关的文档块:\n", len(retrievedDocs))
		for _, doc := range retrievedDocs {
			fmt.Printf("  - ID: %s\n", doc.ID)
			fmt.Printf("    内容: %s\n", doc.Content)
			fmt.Printf("    元数据: %v\n", doc.MetaData)
		}
	}
}
