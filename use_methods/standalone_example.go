package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// runStandaloneExample 展示了如何“单独使用”ChatModel。
// 它的功能是直接与大模型进行一次完整的对话交互。
func runStandaloneExample() {
	// --- 这部分代码在编排示例中会重复，因此可以考虑重构 ---
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // 路径已更正为当前目录
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	// ----------------------------------------------------

	// 创建一个上下文，用于控制请求的生命周期
	ctx := context.Background()

	// 设置请求超时时间
	timeout := 30 * time.Second

	// --- 初始化 ChatModel ---
	// 使用 ark.NewChatModel 创建一个模型实例。
	// 配置信息（如 API Key 和模型名称）从 viper 加载的 config.yaml 文件中读取。
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  viper.GetString("ARK_API_KEY"), // 从配置中获取 API Key
		Model:   viper.GetString("ARK_MODEL"),   // 从配置中获取模型名称
		Timeout: &timeout,                       // 设置超时
	})
	if err != nil {
		panic(fmt.Errorf("初始化 ChatModel 失败: %w", err))
	}

	// --- 准备对话消息 ---
	// 消息列表是一个对话历史，可以包含多种角色。
	messages := []*schema.Message{
		schema.SystemMessage("你是一个助手"), // 系统消息，用于设定模型的角色和行为
		schema.UserMessage("你好"),       // 用户消息，代表用户的输入
	}

	// --- 方式一: 标准生成 (Generate) ---
	// 一次性获取完整的模型回复。
	println("--- 标准生成 (Standalone) ---")
	response, err := model.Generate(ctx, messages)
	if err != nil {
		panic(fmt.Errorf("标准生成失败: %w", err))
	}

	// 打印模型生成的完整内容
	println(response.Content)

	// 打印本次调用的 Token 使用情况
	if usage := response.ResponseMeta.Usage; usage != nil {
		println("提示 Tokens:", usage.PromptTokens)
		println("生成 Tokens:", usage.CompletionTokens)
		println("总 Tokens:", usage.TotalTokens)
	}

	// --- 方式二: 流式生成 (Stream) ---
	// 逐块接收模型返回的内容，适用于打字机效果或长文本生成。
	println("\n--- 流式生成 (Standalone) ---")
	stream, err := model.Stream(ctx, messages)
	if err != nil {
		panic(fmt.Errorf("流式生成失败: %w", err))
	}
	// 确保在函数结束时关闭流
	defer stream.Close()

	// 循环接收数据块，直到流结束
	for {
		chunk, err := stream.Recv()
		if err != nil {
			// 当流结束时，stream.Recv() 会返回 io.EOF 错误，这是正常结束的标志。
			break
		}
		// 实时打印每个数据块的内容
		print(chunk.Content)
	}
	println() // 确保在流式输出后换行
}
