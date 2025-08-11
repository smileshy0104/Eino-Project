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

// Retriever 模拟文档检索器
type Retriever struct {
	knowledgeBase map[string]string
}

func NewRetriever() *Retriever {
	return &Retriever{
		knowledgeBase: map[string]string{
			"eino": "Eino 是一个云原生的大模型应用开发框架。",
			"ark":  "火山方舟（Ark）是一个模型即服务（MaaS）平台。",
		},
	}
}

func (r *Retriever) Retrieve(_ context.Context, query string) (string, error) {
	for keyword, doc := range r.knowledgeBase {
		if strings.Contains(strings.ToLower(query), keyword) {
			return doc, nil
		}
	}
	return "", fmt.Errorf("未找到相关信息")
}

// Orchestrator 协调器
type Orchestrator struct {
	retriever *Retriever
	model     *ark.ChatModel
}

func NewOrchestrator(ctx context.Context) (*Orchestrator, error) {
	retriever := NewRetriever()
	timeout := 30 * time.Second
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("ARK_MODEL"),
		Timeout: &timeout,
	})
	if err != nil {
		return nil, err
	}
	return &Orchestrator{retriever: retriever, model: model}, nil
}

func (o *Orchestrator) Run(ctx context.Context, userQuery string) (string, error) {
	// 1. 检索
	contextDoc, err := o.retriever.Retrieve(ctx, userQuery)
	if err != nil {
		contextDoc = "无相关背景知识"
	}
	// 2. 构建 Prompt
	prompt := fmt.Sprintf("背景知识：%s\n\n问题：%s", contextDoc, userQuery)
	// 3. 生成
	messages := []*schema.Message{
		schema.SystemMessage("你是一个问答助手，请根据背景知识回答问题。"),
		schema.UserMessage(prompt),
	}
	response, err := o.model.Generate(ctx, messages)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// main 函数是程序的唯一入口
func main() {
	// --- 配置加载 ---
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w", err))
	}
	ctx := context.Background()

	// --- 运行编排示例 ---
	fmt.Println("--- 运行编排使用 (RAG) 示例 ---")
	orchestrator, err := NewOrchestrator(ctx)
	if err != nil {
		panic(err)
	}
	finalAnswer, err := orchestrator.Run(ctx, "Eino 是什么？")
	if err != nil {
		panic(err)
	}
	fmt.Println("--- RAG 最终答案 ---")
	fmt.Println(finalAnswer)

	// --- 从 examples 包运行其他示例 ---
	// examples.RunStandaloneExample()
	examples.RunOptionsExample()
}
