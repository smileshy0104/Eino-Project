package chain_example

import (
	"context"
	"fmt"
	"log"
	"time"

	embedder "github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"
)

// createPromptFromDocs 是一个自定义的 Lambda 函数。
func createPromptFromDocs(ctx context.Context, in map[string]any) (out []*schema.Message, err error) {
	query, _ := in["query"].(string)
	docs, _ := in["docs"].([]*schema.Document)

	if len(docs) == 0 {
		prompt := fmt.Sprintf("背景知识库中没有与“%s”相关的信息。请直接回答问题。", query)
		messages := []*schema.Message{
			schema.SystemMessage("你是一个知识渊博的问答助手。"),
			schema.UserMessage(prompt),
		}
		return messages, nil
	}

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

// Run 是此包的入口函数，用于执行 RAG Chain 示例。
func Run() {
	ctx := context.Background()

	// --- 1. 初始化所有组件 ---
	timeout := 30 * time.Second
	embedderComponent, err := embedder.NewEmbedder(ctx, &embedder.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("EMBEDDER_MODEL"),
		Timeout: &timeout,
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}

	client, err := cli.NewClient(ctx, cli.Config{
		Address: viper.GetString("MILVUS_ADDRESS"),
	})
	if err != nil {
		log.Fatalf("创建 Milvus 客户端失败: %v", err)
	}

	retrieverCfg := &milvus.RetrieverConfig{
		Client:       client,
		Collection:   viper.GetString("MILVUS_COLLECTION"),
		Embedding:    embedderComponent,
		OutputFields: []string{"content", "metadata"},
	}
	retriever, err := milvus.NewRetriever(ctx, retrieverCfg)
	if err != nil {
		log.Fatalf("创建 Milvus Retriever 失败: %v", err)
	}

	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: viper.GetString("ARK_API_KEY"),
		Model:  viper.GetString("ARK_MODEL"),
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}
	fmt.Println("所有 RAG 组件初始化成功！")

	// --- 2. 构建并编排 Chain ---
	chain := compose.NewChain[string, *schema.Message]()

	chain.AppendLambda(
		compose.InvokableLambda(func(ctx context.Context, query string) (map[string]any, error) {
			return map[string]any{"query": query}, nil
		}),
	)

	chain.AppendRetriever(retriever, compose.WithInputKey("query"), compose.WithOutputKey("docs"))
	chain.AppendLambda(compose.InvokableLambda(createPromptFromDocs))
	chain.AppendChatModel(model)

	// --- 3. 编译并运行 Chain ---
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译 Chain 失败: %v", err)
	}

	query := "Eino 是什么？"
	fmt.Printf("\n--- 开始运行 RAG Chain, 查询: \"%s\" ---\n", query)
	finalAnswer, err := runnable.Invoke(ctx, query)
	if err != nil {
		log.Fatalf("运行 Chain 失败: %v", err)
	}

	// --- 4. 打印结果 ---
	fmt.Println("\n--- RAG Chain 最终答案 ---")
	fmt.Println(finalAnswer.Content)
}
