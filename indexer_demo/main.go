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
//        此示例包含了处理复杂情况的最佳实践，例如：
//        1. 自动创建集合与向量索引。
//        2. 确保 Embedder 输出的向量类型与 Milvus Schema 定义一致。
//        3. 校验 Embedder 输出的向量数量与待索引的文档数量一致。
//
// =============================================================================

// fields 定义了 Milvus 集合的模式 (schema)。
// 每一个字段的定义都至关重要，必须与 Embedder 的输出和 Retriever 的需求严格对应。
// 每个字段都指定了名称、数据类型和（可选的）类型参数。
// - id: 文档的唯一标识符，是主键。
// - vector: 存储文档内容的向量表示。维度 (dim) 需要与 embedding 模型输出的维度匹配。
// - content: 原始的文档文本内容。
// - metadata: 一个 JSON 字段，用于存储任何附加的元数据。
var fields = []*entity.Field{
	{
		Name:        "id",
		DataType:    entity.FieldTypeVarChar,
		TypeParams:  map[string]string{"max_length": "255"},
		PrimaryKey:  true,
		Description: "文档的唯一主键",
	},
	{
		Name: "vector",
		// 关键设定 (1): 向量的数据类型。
		// 必须与 Embedder 组件返回的向量类型完全一致。
		// 经过调试，ark embedder 返回的是 []uint8 (byte slice)，因此这里必须使用 BinaryVector。
		// 如果使用其他 embedder 返回 []float32，则应使用 FloatVector。
		DataType: entity.FieldTypeBinaryVector,
		// 关键设定 (2): 向量的维度。
		// 必须与 embedding 模型输出的向量维度完全一致。
		// 例如，某些模型的二进制向量维度是 8192 (1024 * 8)。请根据您使用的模型进行精确设置。
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

// runIndexerExample 演示了配置和使用 Indexer 组件的完整流程。
func runIndexerExample() {
	ctx := context.Background()

	// --- 步骤 0: 初始化 Embedder ---
	// Embedder 负责将文本转换为向量。后续的 Indexer 和 Retriever 都依赖它。
	timeout := 30 * time.Second
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("EMBEDDER_MODEL"),
		Timeout: &timeout,
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}

	// --- 步骤 1: 配置并连接 Milvus ---
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
	client, err := cli.NewClient(ctx, cli.Config{Address: address})
	if err != nil {
		log.Fatalf("创建 Milvus 客户端失败: %v", err)
	}

	// --- 步骤 1a: 检查并创建集合与索引 (最佳实践) ---
	// 为保证检索的效率和准确性，必须为向量字段创建索引。
	// 此处我们手动检查集合是否存在，如果不存在，则创建集合并为其向量字段创建索引。
	has, err := client.HasCollection(ctx, collectionName)
	if err != nil {
		log.Fatalf("检查集合是否存在失败: %v", err)
	}
	if !has {
		fmt.Printf("集合 '%s' 不存在，正在创建...\n", collectionName)
		schema := &entity.Schema{CollectionName: collectionName, Fields: fields, Description: "Eino demo collection"}
		err = client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
		if err != nil {
			log.Fatalf("创建集合失败: %v", err)
		}
		fmt.Println("集合创建成功！")

		fmt.Println("正在为 'vector' 字段创建 BIN_FLAT 索引...")
		// 关键设定 (3): 向量索引类型。
		// 必须与向量的数据类型匹配。对于 BinaryVector，通常使用 BIN_FLAT 或 BIN_IVF_FLAT 索引。
		// 距离度量 (MetricType) 也需匹配，HAMMING 是二进制向量常用的距离计算方式。
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

	// --- 步骤 2: 配置并初始化 Indexer ---
	// Indexer 是 Eino 中负责将文档写入向量数据库的组件。
	cfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: collectionName,
		Embedding:  embedder,
		Fields:     fields,
	}
	indexer, err := milvus.NewIndexer(ctx, cfg)
	if err != nil {
		log.Fatalf("创建 Indexer 失败: %v", err)
	}
	fmt.Println("Indexer 初始化成功！")

	// --- 步骤 3: 准备待存储的文档 ---
	doc1 := &schema.Document{
		ID:       "1",
		Content:  "Eino 是一个云原生的大模型应用开发框架，旨在简化和加速大模型应用的构建。",
		MetaData: map[string]interface{}{"source": "official_docs", "author": "CloudWeGo"},
	}
	doc2 := &schema.Document{
		ID:       "2",
		Content:  "RAG (Retrieval-Augmented Generation) 是一种结合了检索和生成两大功能的AI技术。",
		MetaData: map[string]interface{}{"source": "tech_blog", "author": "AI_Researcher"},
	}
	doc3 := &schema.Document{
		ID:       "3",
		Content:  "Go语言微服务架构和gRPC框架的核心内容。",
		MetaData: map[string]interface{}{"source": "Aiyer0104_blog", "author": "yyds"},
	}
	docsToStore := []*schema.Document{doc1, doc2, doc3}

	// --- 步骤 4: 调用 Store 方法进行存储 ---
	// Store 方法是 Indexer 的核心，它会自动处理以下流程：
	// 1. 调用内部的 Embedder 组件，将每个文档的 Content 字段转换为向量。
	// 2. 将文档的 ID, Content, MetaData 以及生成的向量组装成符合 Milvus 格式的数据。
	// 3. 调用 Milvus SDK 的 Insert 方法将数据写入集合。
	fmt.Println("\n准备存储以下文档:")
	for _, doc := range docsToStore {
		fmt.Printf("  - ID: %s\n", doc.ID)
	}

	fmt.Println("\n正在调用 Store 方法...")
	storedIDs, err := indexer.Store(ctx, docsToStore)
	if err != nil {
		// 常见错误排查:
		// - "invalid type, expected [], got []": Embedder 输出类型与 Milvus Schema 不匹配。
		// - "num_rows of field is not equal to passed num_rows": Embedder 返回的向量数与文档数不匹配。
		// - "dimension is not match": Embedder 输出的向量维度与 Milvus Schema 不匹配。
		log.Fatalf("存储文档失败: %v", err)
	}

	// --- 步骤 5: 确认存储结果 ---
	fmt.Println("\n--- 存储成功 ---")
	fmt.Printf("返回的文档 IDs: %v\n", storedIDs)

	// --- 步骤 6: 加载集合到内存 (关键步骤) ---
	// 数据写入 Milvus 后，默认并不能立即被检索，需要先将集合或分区加载到内存中。
	// 这是确保后续 Retriever 能够查询到数据的关键一步。
	fmt.Println("\n正在加载集合到内存以便检索...")
	err = client.LoadCollection(ctx, collectionName, false)
	if err != nil {
		log.Fatalf("加载集合失败: %v", err)
	}
	fmt.Println("集合加载成功！现在可以进行检索了。")
}

// main 是程序的入口点，负责加载配置并执行示例。
func main() {
	// 使用 Viper 加载配置，可以从 config.yaml 或环境变量中读取。
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./") // 在当前目录查找配置文件
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("未找到 config.yaml 文件，将仅从环境变量读取配置。")
	}
	runIndexerExample()
}
