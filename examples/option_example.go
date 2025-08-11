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

// =============================================================================
//
//  文件: option_example.go
//  功能: 演示如何使用功能选项 (Functional Options) 来配置 ChatModel。
//  注意: 本示例适配 eino v0.4.3。
//
// =============================================================================

// RunOptionsExample 展示了如何通过在初始化时传入配置结构体来定制 ChatModel 的行为。
// 在这个例子中，我们特别设置了 `Temperature` 参数。
func RunOptionsExample() {
	fmt.Println("\n\n--- 运行 Option 示例 ---")
	ctx := context.Background()

	// --- 1. 定义模型的可选参数 ---
	// Temperature 控制生成文本的随机性：
	// - 较高的值 (如 0.8) 会使输出更具创造性和随机性。
	// - 较低的值 (如 0.2) 会使输出更具确定性和一致性。
	// 注意：在 eino v0.4.3 中，该字段类型为 *float32。
	var temperature float32 = 0.2

	// --- 2. 初始化带有自定义选项的模型 ---
	// 我们在 NewChatModel 的配置中直接设置 Temperature 字段。
	// 这就是 eino v0.4.3 中实现“功能选项”的方式。
	timeout := 30 * time.Second
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:      viper.GetString("ARK_API_KEY"),
		Model:       viper.GetString("ARK_MODEL"),
		Timeout:     &timeout,
		Temperature: &temperature, // 将自定义参数传入配置
	})
	if err != nil {
		panic(fmt.Errorf("初始化带有 Option 的 ChatModel 失败: %w", err))
	}

	// --- 3. 使用配置好的模型进行调用 ---
	// 模型的行为将受到我们在初始化时设置的 Temperature 参数的影响。
	messages := []*schema.Message{
		schema.SystemMessage("你是一个严谨的科学家。"),
		schema.UserMessage("简单解释一下什么是黑洞。"),
	}

	fmt.Println("--- 开始流式生成 (使用自定义 Temperature=0.2) ---")
	stream, err := model.Stream(ctx, messages)
	if err != nil {
		panic(fmt.Errorf("调用流式生成失败: %w", err))
	}
	defer stream.Close()

	// --- 4. 处理流式返回 ---
	// 这部分逻辑与标准用法相同。
	var streamContent string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			// io.EOF 表示流已正常结束
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
