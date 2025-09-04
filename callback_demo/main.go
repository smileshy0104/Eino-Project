package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// 定义数据结构
type ProcessRequest struct {
	Text     string `json:"text"`
	UserID   string `json:"user_id"`
	Priority int    `json:"priority"`
}

type ProcessResponse struct {
	ProcessedText string        `json:"processed_text"`
	TokenCount    int           `json:"token_count"`
	ProcessTime   time.Duration `json:"process_time"`
	Quality       float64       `json:"quality"`
}

// 1. 简化的文本处理组件 - 模拟实际业务组件
type TextProcessor struct {
	name string
}

func (tp *TextProcessor) Process(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
	startTime := time.Now()

	// 触发开始回调
	ctx = callbacks.OnStart(ctx, &model.CallbackInput{
		Messages: []*schema.Message{{Content: req.Text}},
		Config:   &model.Config{},
	})

	// 模拟文本处理逻辑
	processedText := strings.ToUpper(req.Text)
	if req.Priority > 5 {
		processedText = "【高优先级】" + processedText
	}

	// 模拟处理时间
	processingTime := time.Duration(len(req.Text)) * time.Millisecond
	time.Sleep(processingTime)

	// 模拟token计数
	tokenCount := len(strings.Fields(req.Text))

	// 模拟质量评分
	quality := float64(len(req.Text)) / 100.0
	if quality > 1.0 {
		quality = 1.0
	}

	response := &ProcessResponse{
		ProcessedText: processedText,
		TokenCount:    tokenCount,
		ProcessTime:   time.Since(startTime),
		Quality:       quality,
	}

	// 触发成功完成回调
	ctx = callbacks.OnEnd(ctx, &model.CallbackOutput{
		Message: &schema.Message{Content: response.ProcessedText},
	})

	return response, nil
}

// 2. 自定义聊天模型演示 - 展示内置回调机制
type DemoChatModel struct {
	name string
}

func NewDemoChatModel(name string) *DemoChatModel {
	return &DemoChatModel{name: name}
}

func (m *DemoChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 1. 触发开始回调
	ctx = callbacks.OnStart(ctx, &model.CallbackInput{
		Messages: messages,
		Config: &model.Config{
			Model: m.name,
		},
	})

	// 2. 模拟处理逻辑
	var content strings.Builder
	content.WriteString("处理结果: ")
	for i, msg := range messages {
		if i > 0 {
			content.WriteString(" | ")
		}
		content.WriteString(strings.ToUpper(msg.Content))
	}

	// 模拟处理延迟
	time.Sleep(100 * time.Millisecond)

	response := &schema.Message{
		Content: content.String(),
		Role:    schema.Assistant,
	}

	// 3. 触发完成回调
	ctx = callbacks.OnEnd(ctx, &model.CallbackOutput{
		Message: response,
	})

	return response, nil
}

func (m *DemoChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 1. 触发开始回调
	ctx = callbacks.OnStart(ctx, &model.CallbackInput{
		Messages: messages,
		Config: &model.Config{
			Model: m.name,
		},
	})

	// 2. 创建流管道
	sr, sw := schema.Pipe[*model.CallbackOutput](1)

	// 3. 启动异步生成
	go func() {
		defer sw.Close()

		// 模拟流式输出
		chunks := []string{"处理", "流式", "输出", "数据"}
		for _, chunk := range chunks {
			time.Sleep(50 * time.Millisecond)
			output := &model.CallbackOutput{
				Message: &schema.Message{
					Content: chunk,
					Role:    schema.Assistant,
				},
			}
			if !sw.Send(output, nil) {
				return
			}
		}
	}()

	// 4. 触发流式完成回调
	_, nsr := callbacks.OnEndWithStreamOutput(ctx, sr)

	return schema.StreamReaderWithConvert(nsr, func(t *model.CallbackOutput) (*schema.Message, error) {
		return t.Message, nil
	}), nil
}

// 必须实现 model.ChatModel 接口
func (m *DemoChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return nil, fmt.Errorf("tools not implemented")
}

// 3. 演示各种回调使用方式
func demonstrateCallbackMethods() error {
	fmt.Println("🚀 === Eino 回调机制演示 ===\n")

	// 1. 直接使用 callbacks 包的演示
	fmt.Println("1️⃣  演示直接使用 callbacks 包")
	if err := demonstrateDirectCallback(); err != nil {
		return err
	}

	// 2. 自定义组件回调演示
	fmt.Println("\n2️⃣  演示自定义组件内置回调")
	if err := demonstrateComponentCallback(); err != nil {
		return err
	}

	// 3. 流式处理回调演示
	fmt.Println("\n3️⃣  演示流式处理回调")
	if err := demonstrateStreamCallback(); err != nil {
		return err
	}

	// 4. Chain 编排回调演示
	fmt.Println("\n4️⃣  演示 Chain 编排中的回调")
	if err := demonstrateChainCallback(); err != nil {
		return err
	}

	return nil
}

// 直接使用 callbacks 包演示
func demonstrateDirectCallback() error {
	fmt.Println("   使用 callbacks 包直接触发回调事件...")

	ctx := context.Background()

	// 创建文本处理器
	processor := &TextProcessor{name: "direct_processor"}

	// 测试请求
	req := &ProcessRequest{
		Text:     "Direct callback demonstration",
		UserID:   "direct_user",
		Priority: 5,
	}

	fmt.Printf("   🧪 处理请求: %s (优先级: %d)\n", req.Text, req.Priority)

	result, err := processor.Process(ctx, req)
	if err != nil {
		return fmt.Errorf("直接回调演示失败: %w", err)
	}

	fmt.Printf("✅ 直接回调演示完成: %s\n", result.ProcessedText)
	fmt.Printf("   📊 Token数量: %d, 处理时间: %v, 质量分数: %.2f\n",
		result.TokenCount, result.ProcessTime, result.Quality)

	return nil
}

// 自定义组件回调演示
func demonstrateComponentCallback() error {
	fmt.Println("   创建自定义聊天模型组件...")

	ctx := context.Background()

	// 创建演示聊天模型
	chatModel := NewDemoChatModel("demo-chat-model")

	// 准备测试消息
	messages := []*schema.Message{
		{Content: "Hello", Role: schema.User},
		{Content: "How are you", Role: schema.User},
	}

	fmt.Printf("   🧪 发送消息到聊天模型: %d 条消息\n", len(messages))

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return fmt.Errorf("组件回调演示失败: %w", err)
	}

	fmt.Printf("✅ 组件回调演示完成\n")
	fmt.Printf("   💬 模型响应: %s\n", response.Content)

	return nil
}

// 流式处理回调演示
func demonstrateStreamCallback() error {
	fmt.Println("   创建流式聊天模型...")

	ctx := context.Background()

	// 创建演示聊天模型
	chatModel := NewDemoChatModel("demo-stream-model")

	// 准备测试消息
	messages := []*schema.Message{
		{Content: "Generate stream content", Role: schema.User},
	}

	fmt.Printf("   🧪 启动流式处理...\n")

	stream, err := chatModel.Stream(ctx, messages)
	if err != nil {
		return fmt.Errorf("流式回调演示失败: %w", err)
	}
	defer stream.Close()

	fmt.Printf("   📡 接收流式数据:\n")
	var fullContent strings.Builder

	for {
		msg, err := stream.Recv()
		if err != nil {
			// 检查是否是正常的流结束
			if err.Error() == "EOF" || strings.Contains(err.Error(), "closed") {
				break
			}
			return fmt.Errorf("读取流式数据失败: %w", err)
		}

		if msg != nil {
			fmt.Printf("      📦 数据块: %s\n", msg.Content)
			fullContent.WriteString(msg.Content)
			fullContent.WriteString(" ")
		}
	}

	fmt.Printf("✅ 流式回调演示完成\n")
	fmt.Printf("   📝 完整内容: %s\n", strings.TrimSpace(fullContent.String()))

	return nil
}

// Chain 编排回调演示
func demonstrateChainCallback() error {
	fmt.Println("   创建 Chain 编排...")

	ctx := context.Background()

	// 创建处理链
	chain := compose.NewChain[*ProcessRequest, *ProcessResponse]()

	// 创建处理器组件
	preprocessor := &TextProcessor{name: "preprocessor"}
	mainProcessor := &TextProcessor{name: "main_processor"}

	// 构建处理链
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, req *ProcessRequest) (*ProcessRequest, error) {
		fmt.Printf("   🔄 预处理步骤: %s\n", req.Text)

		// 预处理
		processed, err := preprocessor.Process(ctx, req)
		if err != nil {
			return nil, err
		}

		// 返回修改后的请求
		return &ProcessRequest{
			Text:     processed.ProcessedText,
			UserID:   req.UserID,
			Priority: req.Priority,
		}, nil
	}))

	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
		fmt.Printf("   ⚙️  主处理步骤: %s\n", req.Text)
		return mainProcessor.Process(ctx, req)
	}))

	// 测试请求
	testReq := &ProcessRequest{
		Text:     "Chain orchestration with callbacks",
		UserID:   "chain_user",
		Priority: 8,
	}

	fmt.Printf("   🧪 编译并执行 Chain 处理...\n")

	// 编译 Chain
	runnable, err := chain.Compile(ctx)
	if err != nil {
		return fmt.Errorf("Chain编译失败: %w", err)
	}

	// 执行 Chain
	result, err := runnable.Invoke(ctx, testReq)
	if err != nil {
		return fmt.Errorf("Chain回调演示失败: %w", err)
	}

	fmt.Printf("✅ Chain 回调演示完成\n")
	fmt.Printf("   🎉 最终结果: %s\n", result.ProcessedText)
	fmt.Printf("   📊 处理统计: Token=%d, 时间=%v, 质量=%.2f\n",
		result.TokenCount, result.ProcessTime, result.Quality)

	return nil
}

// 初始化配置
func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")

	// 设置默认值
	viper.SetDefault("model.provider", "demo")
	viper.SetDefault("model.model_name", "demo-model")

	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认值
		fmt.Println("⚠️  未找到配置文件，使用默认配置")
		return nil
	}

	fmt.Println("✅ 配置加载成功")
	return nil
}

// 主函数
func main() {
	fmt.Println("🎯 Eino 回调机制演示")
	fmt.Println("========================")

	// 初始化配置
	if err := initConfig(); err != nil {
		log.Fatalf("❌ 配置初始化失败: %v", err)
	}

	if err := demonstrateCallbackMethods(); err != nil {
		log.Fatalf("❌ 演示失败: %v", err)
	}

	fmt.Println("\n🎉 === 演示完成 ===")
	fmt.Println("📚 这个演示展示了:")
	fmt.Println("   • callbacks 包的直接使用")
	fmt.Println("   • 自定义组件的内置回调机制")
	fmt.Println("   • 流式处理的回调支持")
	fmt.Println("   • Chain 编排中的回调集成")
	fmt.Println("   • 回调在实际业务场景中的应用")

	fmt.Println("\n💡 关键特性:")
	fmt.Println("   🔧 无侵入式设计 - 不影响业务逻辑")
	fmt.Println("   📊 全程监控 - 从开始到结束的完整追踪")
	fmt.Println("   🌊 流式支持 - 支持实时流式数据处理")
	fmt.Println("   🧩 灵活集成 - 轻松集成到任何组件和编排中")
}
