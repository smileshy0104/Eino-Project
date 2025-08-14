package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: indexer_demo/main.go
//  功能: 演示如何使用 Indexer 组件将文档存储到 Milvus。
//
// =============================================================================

// fields 定义了 Milvus 集合的模式 (schema)。
// 每个字段都指定了名称、数据类型和（可选的）类型参数。
// - id: 文档的唯一标识符，是主键。
// - vector: 存储文档内容的向量表示。维度 (dim) 需要与 embedding 模型输出的维度匹配。
// - content: 原始的文档文本内容。
// - metadata: 一个 JSON 字段，用于存储任何附加的元数据。
var fields = []*entity.Field{
	{
		Name:       "id",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "255"},
		PrimaryKey: true,
	},
	{
		Name:       "vector", // 确保字段名匹配
		DataType:   entity.FieldTypeBinaryVector,
		TypeParams: map[string]string{"dim": "81920"},
	},
	{
		Name:       "content",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "8192"},
	},
	{
		Name:     "metadata",
		DataType: entity.FieldTypeJSON,
	},
}

// runIndexerExample 演示了如何配置和使用 Indexer 组件。
func runIndexerExample() {
	// 创建一个后台 context，用于控制 API 调用的生命周期。
	ctx := context.Background()

	// --- 0. 初始化 Embedder ---
	// Embedder 组件负责将文本内容转换为向量。
	// 这里我们使用火山方舟 (ark) 的 embedding 服务。
	timeout := 30 * time.Second
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),    // 从配置中读取 API Key
		Model:   viper.GetString("EMBEDDER_MODEL"), // 从配置中读取模型名称
		Timeout: &timeout,
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}

	// --- 1. 配置并初始化 Indexer ---
	// Indexer 组件负责将文档（包括其向量表示）存储到向量数据库中。
	// 这里我们使用 Milvus 作为向量数据库。

	// 从 viper 加载 Milvus 所需的配置。
	// 请确保这些值已在 config.yaml 或环境变量中设置。
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

	// 根据文档，创建一个 milvus.IndexerConfig。
	// 这个配置对象将 Milvus 客户端、集合名称、Embedder 和集合模式关联起来。
	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: collectionName,
		Embedding:  embedder,
		Fields:     fields,
	}

	// 使用配置创建 Indexer 实例。
	// NewIndexer 会检查集合是否存在，如果不存在，则会根据提供的 `Fields` 定义自动创建。
	indexer, err := milvus.NewIndexer(ctx, cfg)
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
		MetaData: map[string]interface{}{
			"source": "official_docs",
			"author": "CloudWeGo",
		},
	}

	doc2 := &schema.Document{
		ID:      "eino-doc-002",
		Content: "RAG (Retrieval-Augmented Generation) 是一种结合了检索和生成两大功能的AI技术。",
		MetaData: map[string]interface{}{
			"source": "tech_blog",
			"author": "AI_Researcher",
		},
	}

	docs := []*schema.Document{doc1, doc2}
	fmt.Println("\n准备存储以下文档:")
	for _, doc := range docs {
		fmt.Printf("  - ID: %s, 内容: %s\n", doc.ID, doc.Content)
	}

	// --- 3. 调用 Store 方法 ---
	// 将文档列表传入 Store 方法。
	// Indexer 内部会执行以下操作：
	// 1. 调用 Embedder 将每个文档的 `Content` 转换为向量。
	// 2. 将文档的 ID, 向量, 内容和元数据组织成符合 Milvus 格式的列。
	// 3. 调用 Milvus Go SDK 的 `Insert` 方法将数据存入指定的集合。
	fmt.Println("\n正在调用 Store 方法...")
	storedIDs, err := indexer.Store(ctx, docs)
	if err != nil {
		log.Fatalf("存储文档失败: %v", err)
	}

	// --- 4. 打印结果 ---
	// 打印返回的 ID 列表，确认文档已成功存储。
	// Store 方法返回的 ID 列表与传入的文档顺序一致。
	fmt.Println("\n--- 存储成功 ---")
	fmt.Printf("返回的文档 IDs: %v\n", storedIDs)
}

// main 是程序的入口点。
func main() {
	// 初始化 viper 以加载配置。
	// Viper 会首先尝试从 `config.yaml` 文件中读取配置。
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./") // 在当前目录查找配置文件

	// `AutomaticEnv` 允许 Viper 从环境变量中读取配置。
	// 环境变量的键名需要与配置文件中的键名匹配，但通常是大写的，例如 `MILVUS_ADDRESS`。
	viper.AutomaticEnv()

	// 尝试读取配置文件。如果文件不存在，我们会打印一条消息，
	// 但程序会继续运行，因为它仍然可以从环境变量中获取配置。
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("未找到 config.yaml 文件，将仅从环境变量读取配置。")
	}

	// 调用核心的示例函数。
	runIndexerExample()
}
