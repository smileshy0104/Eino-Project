package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/retriever/volc_vikingdb"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: chain_example.go
//  功能: 演示如何使用 compose.Chain 编排 Retriever 和 ChatModel，
//        构建一个完整的 RAG (检索增强生成) 应用。
//
// =============================================================================

// createPromptFromDocs 是一个自定义的 Lambda 函数，用于在 Chain 中承上启下。
// 它接收原始查询和检索到的文档，然后生成一个最终的 Prompt。
func createPromptFromDocs(ctx context.Context, in map[string]any) (out []*schema.Message, err error) {
	query, _ := in["query"].(string)
	docs, _ := in["docs"].([]*schema.Document)

	prompt := "请根据以下背景知识来回答问题。\n\n--- 背景知识 ---\n"
	for i, doc := range docs {
		prompt += fmt.Sprintf("[%d] %s\n", i+1, doc.Content)
	}
	prompt += fmt.Sprintf("\n--- 问题 ---\n%s", query)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个严谨的问答助手，请严格根据提供的背景知识回答。如果知识不足，请说明情况。"),
		schema.UserMessage(prompt),
	}
	return messages, nil
}

func runRAGChainExample() {
	ctx := context.Background()

	// --- 1. 初始化所有组件 ---
	// a. Retriever (与之前的例子相同)
	retrieverCfg := &volc_vikingdb.RetrieverConfig{
		Host:       viper.GetString("VIKINGDB_HOST"),
		Region:     viper.GetString("VIKINGDB_REGION"),
		AK:         viper.GetString("VIKINGDB_AK"),
		SK:         viper.GetString("VIKINGDB_SK"),
		Scheme:     "https",
		Collection: viper.GetString("VIKINGDB_COLLECTION"),
		Index:      viper.GetString("VIKINGDB_INDEX"),
		EmbeddingConfig: volc_vikingdb.EmbeddingConfig{
			UseBuiltin: true,
			ModelName:  "bge-large-zh",
		},
		TopK: of(3),
	}
	retriever, err := volc_vikingdb.NewRetriever(ctx, retrieverCfg)
	if err != nil {
		log.Fatalf("创建 Retriever 失败: %v", err)
	}

	// b. ChatModel
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: viper.GetString("ARK_API_KEY"),
		Model:  viper.GetString("ARK_MODEL"),
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}
	fmt.Println("所有组件初始化成功！")

	// --- 2. 构建并编排 Chain ---
	// a. 创建一个 Chain，输入为 string (query)，输出为 *schema.Message (answer)
	chain := compose.NewChain[string, *schema.Message]()

	// b. 步骤 1: Retriever
	//    输入: string (query)
	//    输出: []*schema.Document
	//    我们将此步骤的输入和输出都保留，以便后续步骤使用。
	chain.AppendRetriever(retriever, compose.WithInputKey("query"), compose.WithOutputKey("docs"))

	// c. 步骤 2: 自定义 Lambda 函数，用于构建 Prompt
	//    输入: map[string]any (包含 "query" 和 "docs")
	//    输出: []*schema.Message
	chain.AppendLambda(compose.InvokableLambda(createPromptFromDocs))

	// d. 步骤 3: ChatModel
	//    输入: []*schema.Message (来自上一步)
	//    输出: *schema.Message
	chain.AppendChatModel(model)

	// --- 3. 编译并运行 Chain ---
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译 Chain 失败: %v", err)
	}

	query := "Eino 框架是什么？"
	fmt.Printf("\n--- 开始运行 RAG Chain, 查询: \"%s\" ---\n", query)
	finalAnswer, err := runnable.Invoke(ctx, query)
	if err != nil {
		log.Fatalf("运行 Chain 失败: %v", err)
	}

	// --- 4. 打印结果 ---
	fmt.Println("\n--- RAG Chain 最终答案 ---")
	fmt.Println(finalAnswer.Content)
}

// of 是一个辅助函数，用于获取 int 或 float64 的指针。
// Eino 的配置中广泛使用指针来区分“未设置”和“零值”。
func of[T int | float64](v T) *T {
	return &v
}
