package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// MockEmbedder 是一个模拟的 Embedder 组件，用于在没有真实 embedding 服务时进行测试。
// 它实现了 eino/components/embedding/interface.go中定义的 Embedder 接口。
type MockEmbedder struct{}

// EmbedStrings 实现了 embedding.Embedder 接口中的方法。
// 它接收一个字符串切片，并返回一个模拟的向量切片。
// 注意：为了匹配接口定义，我们必须包含 ...embedding.Option 参数，即使在这个模拟中并未使用它。
func (m *MockEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	// 这是一个非常简化的模拟。在实际应用中，这里会调用一个真实的 embedding 模型服务。
	vectors := make([][]float64, len(texts))
	// 在这个模拟中，我们忽略输入文本的内容，总是返回一个固定的向量用于演示。
	if len(texts) > 0 {
		vectors[0] = []float64{0.8, 0.1, 0.1} // 模拟 "cat" 的向量
	}
	return vectors, nil
}

// EmbedQuery 实现了 embedding.Embedder 接口中的方法。
// 它接收单个查询字符串，并返回一个模拟的向量。
func (m *MockEmbedder) EmbedQuery(ctx context.Context, text string, opts ...embedding.Option) ([]float64, error) {
	// 复用 EmbedStrings 的逻辑来处理单个查询。
	vectors, err := m.EmbedStrings(ctx, []string{text}, opts...)
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

// MemoryRetriever 是一个基于内存的自定义 Retriever 实现。
// 它实现了 eino/components/retriever/interface.go 中定义的 Retriever 接口。
// 这个实现将文档存储在内存的一个 map 中，并执行简单的字符串匹配来模拟检索过程。
type MemoryRetriever struct {
	docs        map[string]*schema.Document // 用于存储文档的内存数据库
	embedder    embedding.Embedder          // 关联的 embedder 组件
	defaultTopK int                         // 默认返回的文档数量
}

// NewMemoryRetriever 创建一个新的 MemoryRetriever 实例。
// 它初始化了一些硬编码的示例文档。
func NewMemoryRetriever(embedder embedding.Embedder) *MemoryRetriever {
	// 初始化时，我们在内存中存储一些示例文档，用于后续的检索。
	docs := map[string]*schema.Document{
		"doc1": {ID: "doc1", Content: "A cat is a small carnivorous mammal.", MetaData: map[string]interface{}{"source": "wikipedia"}},
		"doc2": {ID: "doc2", Content: "A dog is a domestic animal.", MetaData: map[string]interface{}{"source": "dictionary"}},
		"doc3": {ID: "doc3", Content: "The fluffy cat is sleeping on the mat.", MetaData: map[string]interface{}{"source": "storybook"}},
	}
	return &MemoryRetriever{
		docs:        docs,
		embedder:    embedder,
		defaultTopK: 2, // 设置默认检索返回2个结果
	}
}

// Retrieve 实现了 Retriever 接口的核心方法。
// 这是执行文档检索的地方。
func (r *MemoryRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 1. 处理选项，设置默认值。
	// 这是 Option 机制的标准用法：先用默认值初始化，然后用用户传入的选项覆盖。
	options := &retriever.Options{
		TopK:      &r.defaultTopK,
		Embedding: r.embedder,
	}
	// GetCommonOptions 是一个辅助函数，它会解析用户传入的 opts 并更新 options 结构体。
	options = retriever.GetCommonOptions(options, opts...)

	log.Printf("开始检索，查询: '%s', TopK: %d\n", query, *options.TopK)

	// 2. 执行检索逻辑。
	// 在真实的场景中，这里会使用 options.Embedding 来将 query 字符串向量化，
	// 然后在向量数据库（如 Milvus, VikingDB）中进行相似度搜索。
	// 为了简化，我们这里仅作一个简单的基于关键词的过滤来模拟检索过程。
	var results []*schema.Document
	for _, doc := range r.docs {
		if strings.Contains(strings.ToLower(doc.Content), strings.ToLower(query)) {
			results = append(results, doc)
		}
	}

	// 3. 应用 TopK 限制。
	// 在获取所有匹配的文档后，根据 TopK 的值截取结果列表。
	if len(results) > *options.TopK {
		results = results[:*options.TopK]
	}

	return results, nil
}

// main 是程序的入口点。
func main() {
	// 创建一个后台 context。
	ctx := context.Background()

	// 1. 初始化我们自定义的 Retriever。
	// 首先创建一个模拟的 Embedder 实例。
	mockEmbedder := &MockEmbedder{}
	// 然后用这个 embedder 创建我们的 MemoryRetriever。
	memRetriever := NewMemoryRetriever(mockEmbedder)
	fmt.Println("自定义 Retriever 初始化成功。")

	// 2. 准备查询字符串。
	query := "cat"

	// 3. 执行第一次检索。
	// 这次调用不传递任何 Option，所以会使用在 NewMemoryRetriever 中设置的默认 TopK=2。
	fmt.Println("\n--- 第一次检索 (使用默认 TopK) ---")
	docs, err := memRetriever.Retrieve(ctx, query)
	if err != nil {
		log.Fatalf("检索失败: %v", err)
	}

	fmt.Printf("检索到 %d 个文档:\n", len(docs))
	for _, doc := range docs {
		fmt.Printf("  - ID: %s, 内容: %s\n", doc.ID, doc.Content)
	}

	// 4. 执行第二次检索。
	// 这次我们使用 retriever.WithTopK(1) 选项来覆盖默认的 TopK 值。
	fmt.Println("\n--- 第二次检索 (使用 WithTopK(1) 选项) ---")
	docs, err = memRetriever.Retrieve(ctx, query, retriever.WithTopK(1))
	if err != nil {
		log.Fatalf("检索失败: %v", err)
	}

	fmt.Printf("检索到 %d 个文档:\n", len(docs))
	for _, doc := range docs {
		fmt.Printf("  - ID: %s, 内容: %s\n", doc.ID, doc.Content)
	}
}
