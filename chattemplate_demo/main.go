package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: chattemplate_demo/main.go (文档优化版)
//  功能: 演示如何使用 Eino 的编排工具 (compose.Chain) 来连接和运行组件。
//
// =============================================================================

// runTemplateChatWithChain 演示了如何使用 Chain 来编排 ChatTemplate 和 ChatModel。
func runTemplateChatWithChain() {
	ctx := context.Background()

	// --- 1. 定义聊天模板 ---
	// 这部分与之前相同，我们定义一个包含变量的模板。
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}"),
		schema.MessagesPlaceholder("history_key", false),
		&schema.Message{
			Role:    schema.User,
			Content: "请帮帮我，史瓦罗先生，{task}",
		},
	)

	// --- 2. 初始化模型 ---
	// 这部分也与之前相同，我们创建一个 ChatModel 实例。
	// 初始化 ChatModel
	timeout := 30 * time.Second
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("ARK_MODEL"),
		Timeout: &timeout,
	})

	if err != nil {
		log.Fatalf("创建聊天模型失败: %v", err)
	}

	// --- 3. 使用 Chain 进行编排 ---
	// 这是核心优化点：我们使用 Chain 来定义工作流，而不是手动编写编排器。
	// a. 创建一个 Chain，它接收 map[string]any (模板变量) 作为输入，
	//    并期望最终输出 *schema.Message (模型的回复)。
	chain := compose.NewChain[map[string]any, *schema.Message]()

	// b. 将聊天模板附加到 Chain 的第一步。
	//    这一步的输入是 map[string]any，输出是 []*schema.Message。
	chain.AppendChatTemplate(template)

	// c. 将聊天模型附加到 Chain 的第二步。
	//    这一步的输入是上一步的输出 ([]*schema.Message)，输出是 *schema.Message。
	chain.AppendChatModel(model)

	// --- 4. 编译并运行 Chain ---
	// a. 编译 Chain，将其转换为一个可运行的实例 (Runnable)。
	//    编译过程会检查链中各组件的输入输出类型是否匹配。
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译 Chain 失败: %v", err)
	}

	// b. 准备输入变量，这与之前手动编排时相同。
	variables := map[string]any{
		"role": "机器人史瓦罗先生",
		"task": "写一首关于星空的诗",
		"history_key": []*schema.Message{
			{Role: schema.User, Content: "告诉我油画是什么?"},
			{Role: schema.Assistant, Content: "油画是一种使用油性颜料在画布上创作的绘画形式。"},
		},
	}

	// c. 调用 Runnable 的 Invoke 方法来执行整个链。
	//    我们只需要提供最初的输入，Chain 会自动处理中间步骤的数据传递。
	fmt.Println("--- 开始通过 Chain 运行... ---")
	finalAnswer, err := runnable.Invoke(ctx, variables)
	if err != nil {
		log.Fatalf("运行 Chain 失败: %v", err)
	}

	// --- 5. 打印结果 ---
	fmt.Println("\n--- 最终答案 ---")
	fmt.Println(finalAnswer.Content)
}

// main 是程序的入口。
func main() {
	// --- 统一的配置加载 ---
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./") // 确保在项目根目录运行
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	runTemplateChatWithChain()
}
