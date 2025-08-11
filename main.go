package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Eini/examples" // 导入本地的 examples 包

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: main.go
//  功能: 作为项目的主入口，演示了 RAG (检索增强生成) 的编排流程，
//        并调用其他包中的示例函数。
//
// =============================================================================

// --- RAG 组件定义 ---

// Retriever 模拟一个文档检索器。
// 在真实应用中，它会连接到一个向量数据库或搜索引擎。
type Retriever struct {
	knowledgeBase map[string]string
}

// NewRetriever 创建一个带有预置知识的检索器实例。
func NewRetriever() *Retriever {
	return &Retriever{
		knowledgeBase: map[string]string{
			"eino": "Eino 是一个为简化和加速大模型应用构建而设计的云原生开发框架。",
			"ark":  "火山方舟（Ark）是字节跳动推出的一个模型即服务（MaaS）平台。",
		},
	}
}

// Retrieve 根据查询从知识库中查找相关文档。
func (r *Retriever) Retrieve(_ context.Context, query string) (string, error) {
	for keyword, doc := range r.knowledgeBase {
		if strings.Contains(strings.ToLower(query), keyword) {
			return doc, nil
		}
	}
	return "", fmt.Errorf("在知识库中未找到与 '%s' 相关的信息", query)
}

// Orchestrator 协调器，负责管理和执行 RAG 流程。
type Orchestrator struct {
	retriever *Retriever
	model     *ark.ChatModel
}

// NewOrchestrator 创建并初始化一个完整的编排器实例。
func NewOrchestrator(ctx context.Context) (*Orchestrator, error) {
	retriever := NewRetriever()
	timeout := 30 * time.Second
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("ARK_MODEL"),
		Timeout: &timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("在编排器中初始化 ChatModel 失败: %w", err)
	}
	return &Orchestrator{retriever: retriever, model: model}, nil
}

// Run 执行 RAG 流程：检索 -> 构建 Prompt -> 生成。
func (o *Orchestrator) Run(ctx context.Context, userQuery string) (string, error) {
	// 1. 调用检索器获取上下文
	contextDoc, err := o.retriever.Retrieve(ctx, userQuery)
	if err != nil {
		// 如果检索失败，提供一个默认的上下文，而不是让其为空
		contextDoc = "无相关背景知识"
	}

	// 2. 将检索到的上下文和用户问题组合成一个更丰富的 Prompt
	prompt := fmt.Sprintf("请根据以下背景知识来回答问题。\n\n背景知识：%s\n\n问题：%s", contextDoc, userQuery)

	// 3. 调用模型生成最终答案
	messages := []*schema.Message{
		schema.SystemMessage("你是一个严谨的问答助手，请严格根据提供的背景知识回答。如果知识不足，请说明情况。"),
		schema.UserMessage(prompt),
	}
	response, err := o.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("RAG 流程生成步骤失败: %w", err)
	}
	return response.Content, nil
}

// main 函数是程序的唯一入口。
func main() {
	// --- 1. 加载配置 ---
	// 从 config.yaml 文件中读取配置，如 API Key。
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w", err))
	}
	ctx := context.Background()

	// --- 2. 运行 RAG 编排示例 ---
	fmt.Println("--- 运行编排使用 (RAG) 示例 ---")
	orchestrator, err := NewOrchestrator(ctx)
	if err != nil {
		panic(err)
	}
	finalAnswer, err := orchestrator.Run(ctx, "Eino 和 Ark 分别是什么？")
	if err != nil {
		panic(err)
	}
	fmt.Println("--- RAG 最终答案 ---")
	fmt.Println(finalAnswer)

	// --- 3. 从 examples 包运行其他示例 ---
	// 调用 examples 包中的导出函数 (首字母大写)。
	examples.RunStandaloneExample()
	examples.RunOptionsExample()
}
