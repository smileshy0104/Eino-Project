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

// RunStandaloneExample 展示了如何“单独使用”ChatModel。
func RunStandaloneExample() {
	fmt.Println("\n--- 运行独立使用示例 ---")
	ctx := context.Background()
	timeout := 30 * time.Second

	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("ARK_MODEL"),
		Timeout: &timeout,
	})
	if err != nil {
		panic(fmt.Errorf("初始化 ChatModel 失败: %w", err))
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个助手"),
		schema.UserMessage("你好"),
	}

	// 标准生成
	println("--- 标准生成 ---")
	response, err := model.Generate(ctx, messages)
	if err != nil {
		panic(fmt.Errorf("标准生成失败: %w", err))
	}
	println(response.Content)

	// 流式生成
	println("\n--- 流式生成 ---")
	stream, err := model.Stream(ctx, messages)
	if err != nil {
		panic(fmt.Errorf("流式生成失败: %w", err))
	}
	defer stream.Close()
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		print(chunk.Content)
	}
	println()
}
