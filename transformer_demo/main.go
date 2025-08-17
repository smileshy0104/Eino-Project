package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	embedder "github.com/cloudwego/eino-ext/components/embedding/ark"
	retriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: transformer_demo/main.go
//  功能: 演示一个完整的、端到端的 RAG 数据处理流水线，并采用良好的代码结构。
//
// =============================================================================

// milvusSchema 定义了 Milvus 集合的字段。
var milvusSchema = []*entity.Field{
	{
		Name: "id", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "255"}, PrimaryKey: true, Description: "文档块的唯一主键",
	},
	{
		// 最终修正：根据 EmbedStrings 的实际输出 ([][]float64)，将类型设置为 FloatVector。
		Name: "vector", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "1024"}, Description: "文档块内容的向量表示",
	},
	{
		Name: "content", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "8192"}, Description: "原始的文本内容块",
	},
	{
		Name: "metadata", DataType: entity.FieldTypeJSON, Description: "用于存储附加信息的 JSON 字段",
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
		MetaData: map[string]interface{}{"source": "official-docs"},
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

func setupMilvus(ctx context.Context, collectionName string) cli.Client {
	fmt.Printf("\n--- 步骤 3: 设置 Milvus (集合: %s) ---\n", collectionName)
	client, err := cli.NewClient(ctx, cli.Config{Address: viper.GetString("MILVUS_ADDRESS")})
	if err != nil {
		log.Fatalf("创建 Milvus 客户端失败: %v", err)
	}

	_ = client.DropCollection(ctx, collectionName)
	schema := &entity.Schema{CollectionName: collectionName, Fields: milvusSchema}
	err = client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		log.Fatalf("创建集合失败: %v", err)
	}
	// 最终修正：为 FloatVector 创建 HNSW 索引，使用 IP (内积) 作为距离度量。
	hnswIndex, _ := entity.NewIndexHNSW(entity.IP, 8, 16)
	err = client.CreateIndex(ctx, collectionName, "vector", hnswIndex, false)
	if err != nil {
		log.Fatalf("为 'vector' 字段创建索引失败: %v", err)
	}
	fmt.Println("集合与索引创建成功！")
	return client
}

// indexChunks 手动执行 embedding 并直接使用 Milvus SDK 写入数据。
func indexChunks(ctx context.Context, client cli.Client, embedderComponent embedding.Embedder, collectionName string, chunks []*schema.Document) {
	fmt.Println("\n--- 步骤 4: 索引文档块 (手动流程) ---")

	fmt.Println("正在为文档块生成向量...")
	var contents []string
	for _, chunk := range chunks {
		contents = append(contents, chunk.Content)
	}
	vectors, err := embedderComponent.EmbedStrings(ctx, contents)
	if err != nil {
		log.Fatalf("生成向量失败: %v", err)
	}

	if len(vectors) != len(chunks) {
		log.Fatalf("Embedder 返回的向量数量 (%d) 与文档块数量 (%d) 不匹配。", len(vectors), len(chunks))
	}
	fmt.Printf("向量生成成功，共 %d 个。\n", len(vectors))

	// 准备列式数据
	ids := make([]string, 0, len(chunks))
	contentsCol := make([]string, 0, len(chunks))
	vectorsCol := make([][]float32, 0, len(chunks)) // 最终修正：Milvus SDK 需要 [][]float32
	metadatasCol := make([][]byte, 0, len(chunks))

	for i, chunk := range chunks {
		ids = append(ids, chunk.ID)
		contentsCol = append(contentsCol, chunk.Content)

		// 最终修正：将 []float64 转换为 []float32
		float32Vector := make([]float32, len(vectors[i]))
		for j, v := range vectors[i] {
			float32Vector[j] = float32(v)
		}
		vectorsCol = append(vectorsCol, float32Vector)

		metaBytes, _ := json.Marshal(chunk.MetaData)
		metadatasCol = append(metadatasCol, metaBytes)
	}

	// 创建列
	idCol := entity.NewColumnVarChar("id", ids)
	contentCol := entity.NewColumnVarChar("content", contentsCol)
	vectorCol := entity.NewColumnFloatVector("vector", 1024, vectorsCol)
	metadataCol := entity.NewColumnJSONBytes("metadata", metadatasCol)

	// 插入数据
	fmt.Println("正在将数据直接插入 Milvus...")
	_, err = client.Insert(ctx, collectionName, "", idCol, contentCol, vectorCol, metadataCol)
	if err != nil {
		log.Fatalf("直接插入 Milvus 失败: %v", err)
	}
	fmt.Println("文档块存储成功！")

	err = client.LoadCollection(ctx, collectionName, false)
	if err != nil {
		log.Fatalf("加载集合失败: %v", err)
	}
	fmt.Println("集合加载成功！")
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
	client := setupMilvus(ctx, collectionName)
	indexChunks(ctx, client, embedderComponent, collectionName, chunks)
	retrieveChunks(ctx, client, embedderComponent, collectionName, "Transformer 是做什么的？")
}
