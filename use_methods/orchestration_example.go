package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  本文件演示了如何“在编排中使用”ChatModel。
//  我们构建一个简单的“检索增强生成”(RAG)流程，它由多个组件协作完成。
//
// =============================================================================

// -----------------------------------------------------------------------------
// 组件 1: 文档检索器 (Retriever)
// -----------------------------------------------------------------------------
// Retriever 的作用是从知识库中查找与用户问题相关的信息。
// 在真实场景中，这可能是一个连接到向量数据库（如 Milvus, Pinecone）的复杂组件。
// 这里我们用一个简单的 map 来模拟一个小型知识库。
type Retriever struct {
	knowledgeBase map[string]string
}

// NewRetriever 创建并初始化一个检索器实例。
func NewRetriever() *Retriever {
	return &Retriever{
		knowledgeBase: map[string]string{
			"eino": "Eino 是一个云原生的大模型应用开发框架，旨在简化和加速大模型应用的构建。",
			"ark":  "火山方舟（Ark）是字节跳动推出的一个模型即服务（MaaS）平台，提供了多种先进的AI模型。",
		},
	}
}

// Retrieve 根据查询从知识库中查找相关文档。
// 它通过简单的关键字匹配来模拟检索过程。
func (r *Retriever) Retrieve(ctx context.Context, query string) (string, error) {
	for keyword, doc := range r.knowledgeBase {
		// 如果查询中包含知识库中的关键字，则返回对应的文档
		if strings.Contains(strings.ToLower(query), keyword) {
			return doc, nil
		}
	}
	return "", fmt.Errorf("在知识库中未找到与 '%s' 相关的信息", query)
}

// -----------------------------------------------------------------------------
// 组件 2: ChatModel (我们已经熟悉)
// -----------------------------------------------------------------------------
// ChatModel 在这个编排流程中扮演“大脑”的角色，负责根据提供的信息进行推理和生成文本。
// 为了保持编排器的整洁，我们将模型初始化放在编排器的构造函数中。

// -----------------------------------------------------------------------------
// 组件 3: 编排器 (Orchestrator)
// -----------------------------------------------------------------------------
// Orchestrator 是整个流程的核心，负责协调和调用其他组件。
type Orchestrator struct {
	retriever *Retriever
	model     *ark.ChatModel
}

// NewOrchestrator 创建并初始化编排器及其所有依赖的组件。
func NewOrchestrator(ctx context.Context) (*Orchestrator, error) {
	// 初始化检索器
	retriever := NewRetriever()

	// 初始化 ChatModel
	timeout := 30 * time.Second
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("ARK_MODEL"),
		Timeout: &timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 ChatModel 失败: %w", err)
	}

	// 返回一个包含所有已初始化组件的编排器实例
	return &Orchestrator{
		retriever: retriever,
		model:     model,
	}, nil
}

// Run 方法执行 RAG (Retrieval-Augmented Generation) 流程。
// 这是编排器定义的核心业务逻辑。
func (o *Orchestrator) Run(ctx context.Context, userQuery string) (string, error) {
	fmt.Printf("编排流程开始，用户问题: \"%s\"\n", userQuery)

	// 步骤 1: 调用【文档检索器】获取相关上下文
	fmt.Println("步骤 1: 调用【文档检索器】...")
	contextDoc, err := o.retriever.Retrieve(ctx, userQuery)
	if err != nil {
		// 如果在知识库中找不到相关信息，可以选择直接让模型回答，或返回错误。
		// 这里我们选择让模型在没有额外上下文的情况下尝试回答。
		fmt.Printf("检索失败: %v。将直接由模型回答。\n", err)
		contextDoc = "无相关背景知识" // 提供一个明确的“无信息”信号
	}
	fmt.Printf("检索到的上下文: \"%s\"\n", contextDoc)

	// 步骤 2: 动态构建包含上下文的提示词 (Prompt)
	// 这是 RAG 的核心思想：将检索到的知识注入到提示词中，为模型提供回答问题的依据。
	fmt.Println("步骤 2: 构建提示词...")
	prompt := fmt.Sprintf("请根据以下背景知识回答问题。\n\n背景知识：%s\n\n问题：%s", contextDoc, userQuery)
	fmt.Printf("构建的提示词: \"%s\"\n", prompt)

	// 步骤 3: 调用【ChatModel】进行推理和生成
	// 模型将基于我们提供的、包含上下文的提示词来生成答案。
	fmt.Println("步骤 3: 调用【ChatModel】...")
	messages := []*schema.Message{
		schema.SystemMessage("你是一个智能问答助手，请严格基于提供的背景知识来回答问题。如果背景知识没有提供相关信息，请直接说不知道。"),
		schema.UserMessage(prompt),
	}
	response, err := o.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("ChatModel 生成失败: %w", err)
	}

	fmt.Println("编排流程结束。")
	return response.Content, nil
}

// main 函数是程序的入口，它现在负责驱动编排器。
func main() {
	// --- 统一的配置加载 ---
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./") // 确保在项目根目录运行
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	ctx := context.Background()

	// --- 初始化并运行编排器 ---
	fmt.Println("--- 正在初始化编排器... ---")
	orchestrator, err := NewOrchestrator(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("--- 编排器初始化完成 ---")

	// 定义一个用户问题来驱动 RAG 流程
	userQuery := "请问 Eino 是什么？它和 Ark 有什么关系吗？"
	finalAnswer, err := orchestrator.Run(ctx, userQuery)
	if err != nil {
		panic(err)
	}

	// 打印由整个编排流程生成的最终答案
	fmt.Println("\n--- 最终答案 ---")
	fmt.Println(finalAnswer)

	// （可选）可以调用之前的独立使用示例
	// fmt.Println("\n\n--- 现在运行独立使用示例 ---")
	// runStandaloneExample()
}
