package examples

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// RunOptionsExample 展示了如何使用功能选项（Functional Options）来定制 ChatModel 的行为。
func RunOptionsExample() {
	fmt.Println("\n\n--- 运行 Option 示例 ---")
	ctx := context.Background()

	// 使用 Option 来设置模型的 "temperature" 参数。
	var temperature float32 = 0.5

	// 初始化带有 Option 的模型
	timeout := 30 * time.Second
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:      viper.GetString("ARK_API_KEY"),
		Model:       viper.GetString("ARK_MODEL"),
		Timeout:     &timeout,
		Temperature: &temperature,
	})
	if err != nil {
		panic(fmt.Errorf("初始化 ChatModel 失败: %w", err))
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个诗人。"),
		schema.UserMessage("写一句关于天空的诗。"),
	}

	fmt.Println("--- 开始流式生成 (使用自定义 Temperature) ---")
	stream, err := model.Stream(ctx, messages)
	if err != nil {
		panic(fmt.Errorf("调用流式生成失败: %w", err))
	}
	defer stream.Close()

	var streamContent string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(fmt.Errorf("接收流数据时发生错误: %w", err))
		}
		fmt.Print(chunk.Content)
		streamContent += chunk.Content
	}

	fmt.Println("\n--- 流式生成结束 ---")
	fmt.Printf("汇总内容: %s\n", streamContent)
	fmt.Println("--- Option 示例结束 ---")
}
