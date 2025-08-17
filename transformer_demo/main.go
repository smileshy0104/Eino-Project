package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	embedder "github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	retriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

var milvusSchema = []*entity.Field{
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

func loadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("未找到 config.yaml 文件，将仅从环境变量读取配置。")
	}
}

func prepareDocument() *schema.Document {
	fmt.Println("--- 步骤 1: 准备原始长文档 ---")
	return &schema.Document{
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
		MetaData: map[string]interface{}{"source": "official-docs", "author": "yyds"},
	}
}

func splitDocument(ctx context.Context, doc *schema.Document) []*schema.Document {
	fmt.Println("\n--- 步骤 2: 使用 Transformer 分割文档 ---")
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{"##": "Header 2"},
	})
	if err != nil {
		log.Fatalf("创建 HeaderSplitter 失败: %v", err)
	}
	chunks, err := splitter.Transform(ctx, []*schema.Document{doc})
	if err != nil {
		log.Fatalf("转换文档失败: %v", err)
	}
	fmt.Printf("分割完成，原始文档被分割成 %d 个块。\n", len(chunks))
	return chunks
}

func setupMilvus(ctx context.Context, collectionName string, embedder *embedder.Embedder, chunkDocs []*schema.Document) cli.Client {
	fmt.Printf("\n--- 步骤 3: 设置 Milvus (集合: %s) ---\n", collectionName)
	address := viper.GetString("MILVUS_ADDRESS")
	if address == "" || collectionName == "" {
		log.Fatal("Milvus 配置 (Address, Collection) 必须被设置！")
	}
	client, err := cli.NewClient(ctx, cli.Config{Address: address})
	if err != nil {
		log.Fatalf("创建 Milvus 客户端失败: %v", err)
	}
	has, err := client.HasCollection(ctx, collectionName)
	if err != nil {
		log.Fatalf("检查集合是否存在失败: %v", err)
	}
	if !has {
		fmt.Printf("集合 '%s' 不存在，正在创建...\n", collectionName)
		schema := &entity.Schema{CollectionName: collectionName, Fields: milvusSchema, Description: "Eino demo collection"}
		err = client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
		if err != nil {
			log.Fatalf("创建集合失败: %v", err)
		}
		fmt.Println("集合创建成功！")

		fmt.Println("正在为 'vector' 字段创建 BIN_FLAT 索引...")
		binFlatIndex, err := entity.NewIndexBinFlat(entity.HAMMING, 128)
		if err != nil {
			log.Fatalf("创建 BIN_FLAT 索引对象失败: %v", err)
		}
		err = client.CreateIndex(ctx, collectionName, "vector", binFlatIndex, false)
		if err != nil {
			log.Fatalf("为 'vector' 字段创建索引失败: %v", err)
		}
		fmt.Println("BIN_FLAT 索引创建成功！")
	} else {
		fmt.Printf("集合 '%s' 已存在，跳过创建步骤。\n", collectionName)
	}

	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: collectionName,
		Embedding:  embedder,
		Fields:     milvusSchema,
	}
	indexer, err := milvus.NewIndexer(ctx, cfg)
	if err != nil {
		log.Fatalf("创建 Indexer 失败: %v", err)
	}
	fmt.Println("Indexer 初始化成功！")

	docsToStore := chunkDocs

	fmt.Println("\n准备存储以下文档:")
	for _, doc := range docsToStore {
		fmt.Printf("  - ID: %s\n", doc.ID)
	}

	fmt.Println("\n正在调用 Store 方法...")
	storedIDs, err := indexer.Store(ctx, docsToStore)
	if err != nil {
		log.Fatalf("存储文档失败: %v", err)
	}

	fmt.Println("\n--- 存储成功 ---")
	fmt.Printf("返回的文档 IDs: %v\n", storedIDs)

	return client
}

func retrieveChunks(ctx context.Context, client cli.Client, embedderComponent embedding.Embedder, collectionName string, query string) {
	fmt.Println("\n--- 步骤 5: 检索文档块 ---")
	retrieverCfg := &retriever.RetrieverConfig{
		Client: client, Collection: collectionName, Embedding: embedderComponent, OutputFields: []string{"content", "metadata"},
	}
	retrieverComponent, err := retriever.NewRetriever(ctx, retrieverCfg)
	if err != nil {
		log.Fatalf("创建 Retriever 失败: %v", err)
	}

	fmt.Printf("正在使用查询: \"%s\"\n", query)
	retrievedDocs, err := retrieverComponent.Retrieve(ctx, query)
	if err != nil {
		log.Fatalf("检索文档失败: %v", err)
	}

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

func main() {
	loadConfig()
	ctx := context.Background()

	timeout := 30 * time.Second
	embedderComponent, err := embedder.NewEmbedder(ctx, &embedder.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("EMBEDDER_MODEL"),
		Timeout: &timeout,
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}
	collectionName := viper.GetString("MILVUS_COLLECTION")

	originalDoc := prepareDocument()
	chunks := splitDocument(ctx, originalDoc)
	fmt.Println("正在索引文档...")
	client := setupMilvus(ctx, collectionName, embedderComponent, chunks)
	retrieveChunks(ctx, client, embedderComponent, collectionName, "Transformer 是做什么的？")
}
