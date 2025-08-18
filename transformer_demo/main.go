// Package main 提供了 Eino 框架中 Transformer、Indexer 和 Retriever 组件的演示。
// 这个示例展示了如何将一个长文档进行分割、向量化、索引，并最终根据查询检索出相关文档块的完整流程。
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	// Eino 框架的文档转换器组件，用于分割 Markdown 文档
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	// Eino 框架的 embedding 组件，这里使用火山方舟作为 embedding 服务
	embedder "github.com/cloudwego/eino-ext/components/embedding/ark"
	// Eino 框架的 indexer 组件，用于将文档存入 Milvus 向量数据库
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	// Eino 框架的 retriever 组件，用于从 Milvus 向量数据库中检索文档
	retriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	// Eino 框架的核心 embedding 接口定义
	"github.com/cloudwego/eino/components/embedding"
	// Eino 框架的核心数据结构定义
	"github.com/cloudwego/eino/schema"
	// Milvus Go SDK 客户端
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	// Milvus Go SDK 实体定义
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	// Viper 用于管理配置
	"github.com/spf13/viper"
)

// milvusSchema 定义了 Milvus 集合的结构。
// 包含 id, vector, content, 和 metadata 四个字段。
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
		TypeParams:  map[string]string{"dim": "81920"}, // 维度需与 embedding 模型匹配
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

// Config 存储应用程序配置
type Config struct {
	MilvusAddress    string
	MilvusCollection string
	ArkAPIKey        string
	EmbedderModel    string
}

// loadConfig 从配置文件 (config.yaml) 或环境变量中加载配置。
func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")  // 在当前目录查找
	viper.AddConfigPath("../") // 在上一级目录查找
	viper.AutomaticEnv()       // 允许从环境变量读取
	
	if err := viper.ReadInConfig(); err != nil {
		// 如果找不到配置文件，打印提示信息，程序将依赖环境变量
		fmt.Println("未找到 config.yaml 文件，将仅从环境变量读取配置。")
	}

	config := &Config{
		MilvusAddress:    viper.GetString("MILVUS_ADDRESS"),
		MilvusCollection: viper.GetString("MILVUS_COLLECTION"),
		ArkAPIKey:        viper.GetString("ARK_API_KEY"),
		EmbedderModel:    viper.GetString("EMBEDDER_MODEL"),
	}

	return config, validateConfig(config)
}

// validateConfig 验证配置是否完整
func validateConfig(config *Config) error {
	if config.MilvusAddress == "" {
		return errors.New("MILVUS_ADDRESS 必须设置")
	}
	if config.MilvusCollection == "" {
		return errors.New("MILVUS_COLLECTION 必须设置")
	}
	if config.ArkAPIKey == "" {
		return errors.New("ARK_API_KEY 必须设置")
	}
	if config.EmbedderModel == "" {
		return errors.New("EMBEDDER_MODEL 必须设置")
	}
	return nil
}

// prepareDocument 创建一个用于演示的原始 schema.Document 对象。
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
		MetaData: map[string]any{"source": "official-docs", "author": "yyds"},
	}
}

// splitDocument 使用 Markdown HeaderSplitter 将单个文档分割成多个小块。
func splitDocument(ctx context.Context, doc *schema.Document) ([]*schema.Document, error) {
	fmt.Println("\n--- 步骤 2: 使用 Transformer 分割文档 ---")
	// 基于 Markdown 的二级标题 "##" 进行分割
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{"##": "Header 2"},
	})
	if err != nil {
		return nil, fmt.Errorf("创建 HeaderSplitter 失败: %w", err)
	}

	// 执行分割操作
	chunks, err := splitter.Transform(ctx, []*schema.Document{doc})
	if err != nil {
		return nil, fmt.Errorf("转换文档失败: %w", err)
	}
	fmt.Printf("分割完成，原始文档被分割成 %d 个块。\n", len(chunks))
	return chunks, nil
}

// MilvusClient 封装 Milvus 客户端和相关操作
type MilvusClient struct {
	client cli.Client
}

// NewMilvusClient 创建新的 Milvus 客户端
func NewMilvusClient(ctx context.Context, address string) (*MilvusClient, error) {
	client, err := cli.NewClient(ctx, cli.Config{Address: address})
	if err != nil {
		return nil, fmt.Errorf("创建 Milvus 客户端失败: %w", err)
	}
	return &MilvusClient{client: client}, nil
}

// Close 关闭 Milvus 客户端连接
func (mc *MilvusClient) Close() error {
	return mc.client.Close()
}

// setupMilvus 初始化 Milvus 客户端，创建集合和索引（如果不存在），并使用 Indexer 组件将文档块存入 Milvus。
func setupMilvus(ctx context.Context, config *Config, embedderComponent *embedder.Embedder, chunkDocs []*schema.Document) (*MilvusClient, error) {
	fmt.Printf("\n--- 步骤 3 & 4: 设置 Milvus 并索引文档 (集合: %s) ---\n", config.MilvusCollection)
	
	// 1. 连接 Milvus
	milvusClient, err := NewMilvusClient(ctx, config.MilvusAddress)
	if err != nil {
		return nil, err
	}
	client := milvusClient.client

	// 2. 检查集合是否存在，如果不存在则创建
	has, err := client.HasCollection(ctx, config.MilvusCollection)
	if err != nil {
		return nil, fmt.Errorf("检查集合是否存在失败: %w", err)
	}
	if !has {
		fmt.Printf("集合 '%s' 不存在，正在创建...\n", config.MilvusCollection)
		schema := &entity.Schema{CollectionName: config.MilvusCollection, Fields: milvusSchema, Description: "Eino demo collection"}
		err = client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
		if err != nil {
			return nil, fmt.Errorf("创建集合失败: %w", err)
		}
		fmt.Println("集合创建成功！")

		// 3. 为 vector 字段创建索引
		fmt.Println("正在为 'vector' 字段创建 BIN_FLAT 索引...")
		// 注意：这里的参数需要根据 embedding 模型的特性来选择
		binFlatIndex, err := entity.NewIndexBinFlat(entity.HAMMING, 128)
		if err != nil {
			return nil, fmt.Errorf("创建 BIN_FLAT 索引对象失败: %w", err)
		}
		err = client.CreateIndex(ctx, config.MilvusCollection, "vector", binFlatIndex, false)
		if err != nil {
			return nil, fmt.Errorf("为 'vector' 字段创建索引失败: %w", err)
		}
		fmt.Println("BIN_FLAT 索引创建成功！")
	} else {
		fmt.Printf("集合 '%s' 已存在，跳过创建步骤。\n", config.MilvusCollection)
	}

	// 4. 初始化 Indexer 并存储文档
	indexerCfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: config.MilvusCollection,
		Embedding:  embedderComponent,
		Fields:     milvusSchema,
	}
	indexer, err := milvus.NewIndexer(ctx, indexerCfg)
	if err != nil {
		return nil, fmt.Errorf("创建 Indexer 失败: %w", err)
	}
	fmt.Println("Indexer 初始化成功！")

	fmt.Println("\n准备存储以下文档块:")
	for _, doc := range chunkDocs {
		fmt.Printf("  - ID: %s\n", doc.ID)
	}

	fmt.Println("\n正在调用 Store 方法将文档存入 Milvus...")
	storedIDs, err := indexer.Store(ctx, chunkDocs)
	if err != nil {
		return nil, fmt.Errorf("存储文档失败: %w", err)
	}

	fmt.Println("\n--- 存储成功 ---")
	fmt.Printf("返回的文档 IDs: %v\n", storedIDs)

	return milvusClient, nil
}

// retrieveChunks 使用 Retriever 组件从 Milvus 中检索与查询相关的文档块。
func retrieveChunks(ctx context.Context, milvusClient *MilvusClient, embedderComponent embedding.Embedder, collectionName string, query string) error {
	fmt.Println("\n--- 步骤 5: 检索文档块 ---")
	// 初始化 Retriever
	retrieverCfg := &retriever.RetrieverConfig{
		Client:       milvusClient.client,
		Collection:   collectionName,
		Embedding:    embedderComponent,
		OutputFields: []string{"content", "metadata"}, // 指定检索时需要返回的字段
	}
	retrieverComponent, err := retriever.NewRetriever(ctx, retrieverCfg)
	if err != nil {
		return fmt.Errorf("创建 Retriever 失败: %w", err)
	}

	// 执行检索
	fmt.Printf("正在使用查询: \"%s\"\n", query)
	retrievedDocs, err := retrieverComponent.Retrieve(ctx, query)
	if err != nil {
		return fmt.Errorf("检索文档失败: %w", err)
	}

	// 打印检索结果
	fmt.Println("\n--- 检索成功 ---")
	if len(retrievedDocs) == 0 {
		fmt.Println("未检索到相关文档。")
	} else {
		fmt.Printf("检索到 %d 个最相关的文档块:\n", len(retrievedDocs))
		for i, doc := range retrievedDocs {
			fmt.Printf("  - [%d] ID: %s\n", i+1, doc.ID)
			fmt.Printf("      内容: %s\n", doc.Content)
			fmt.Printf("      元数据: %v\n", doc.MetaData)
		}
	}
	return nil
}

// runRAGDemo 执行完整的 RAG 流程
func runRAGDemo(ctx context.Context, config *Config) error {
	// 初始化 embedding 组件
	timeout := 30 * time.Second
	embedderComponent, err := embedder.NewEmbedder(ctx, &embedder.EmbeddingConfig{
		APIKey:  config.ArkAPIKey,
		Model:   config.EmbedderModel,
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("创建 Embedder 失败: %w", err)
	}

	// 1. 准备文档
	originalDoc := prepareDocument()
	
	// 2. 分割文档
	chunks, err := splitDocument(ctx, originalDoc)
	if err != nil {
		return fmt.Errorf("分割文档失败: %w", err)
	}
	
	// 3. & 4. 设置 Milvus 并索引文档
	fmt.Println("正在索引文档...")
	milvusClient, err := setupMilvus(ctx, config, embedderComponent, chunks)
	if err != nil {
		return fmt.Errorf("设置 Milvus 失败: %w", err)
	}
	defer func() {
		if closeErr := milvusClient.Close(); closeErr != nil {
			fmt.Printf("关闭 Milvus 客户端失败: %v\n", closeErr)
		}
	}()
	
	// 5. 检索文档
	err = retrieveChunks(ctx, milvusClient, embedderComponent, config.MilvusCollection, "Transformer 是做什么的？")
	if err != nil {
		return fmt.Errorf("检索文档失败: %w", err)
	}
	
	return nil
}

// main 是程序的入口点，协调整个 RAG 流程。
func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	ctx := context.Background()
	
	// 执行 RAG 演示
	if err := runRAGDemo(ctx, config); err != nil {
		log.Fatalf("RAG 演示失败: %v", err)
	}
	
	fmt.Println("\n--- RAG 演示完成 ---")
}
