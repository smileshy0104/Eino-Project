package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// 配置结构体
type Config struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// 对话管理器
type ConversationManager struct {
	chatModel  model.BaseChatModel
	history    []*schema.Message
	maxHistory int
	mu         sync.RWMutex
}

func NewConversationManager(cm model.BaseChatModel) *ConversationManager {
	return &ConversationManager{
		chatModel:  cm,
		history:    make([]*schema.Message, 0),
		maxHistory: 10,
	}
}

func (cm *ConversationManager) Chat(ctx context.Context, userInput string) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 添加用户消息到历史
	userMessage := &schema.Message{
		Role:    schema.User,
		Content: userInput,
	}
	cm.history = append(cm.history, userMessage)

	// 构建完整的对话历史
	messages := make([]*schema.Message, 0, len(cm.history)+1)

	// 添加系统提示
	systemMessage := &schema.Message{
		Role:    schema.System,
		Content: "你是一个友好的AI助手，能够记住对话历史并提供连贯的回复。请用简洁明了的方式回答问题。",
	}
	messages = append(messages, systemMessage)

	// 添加历史对话（限制长度）
	startIdx := 0
	if len(cm.history) > cm.maxHistory {
		startIdx = len(cm.history) - cm.maxHistory
	}
	messages = append(messages, cm.history[startIdx:]...)

	// 生成回复
	response, err := cm.chatModel.Generate(ctx, messages,
		model.WithTemperature(0.7),
		model.WithMaxTokens(1000),
	)
	if err != nil {
		return "", err
	}

	// 添加AI回复到历史
	cm.history = append(cm.history, response)

	// 清理过长的历史
	if len(cm.history) > cm.maxHistory*2 {
		cm.history = cm.history[len(cm.history)-cm.maxHistory:]
	}

	return response.Content, nil
}

func (cm *ConversationManager) ClearHistory() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.history = make([]*schema.Message, 0)
}

func (cm *ConversationManager) GetHistoryLength() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.history)
}

// 智能客服机器人
type CustomerServiceBot struct {
	chatModel model.BaseChatModel
	knowledge []string
}

func NewCustomerServiceBot(cm model.BaseChatModel) *CustomerServiceBot {
	return &CustomerServiceBot{
		chatModel: cm,
		knowledge: []string{
			"我们的营业时间是周一到周五 9:00-18:00",
			"退货政策：7天无理由退货，商品需保持原包装",
			"配送时间：1-3个工作日，偏远地区可能需要5-7天",
			"支付方式：支持微信支付、支付宝、银行卡支付",
			"客服热线：400-123-4567",
		},
	}
}

func (bot *CustomerServiceBot) HandleCustomerQuery(ctx context.Context, query string) (string, error) {
	// 构建系统提示
	systemPrompt := fmt.Sprintf(`你是一个专业的客服助手。
知识库信息：
%s

请根据知识库信息回答用户问题，如果知识库中没有相关信息，请礼貌地告知用户联系人工客服。
回答要求：
1. 语气友好、专业
2. 回答简洁明了
3. 如需要可以提供具体的联系方式`,
		strings.Join(bot.knowledge, "\n"))

	messages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: query},
	}

	response, err := bot.chatModel.Generate(ctx, messages,
		model.WithTemperature(0.3), // 较低温度确保回答准确
		model.WithMaxTokens(500),
	)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// 代码生成助手
type CodeGenerator struct {
	chatModel model.BaseChatModel
}

func NewCodeGenerator(cm model.BaseChatModel) *CodeGenerator {
	return &CodeGenerator{chatModel: cm}
}

func (cg *CodeGenerator) GenerateCode(ctx context.Context, requirement string, language string) (string, error) {
	systemPrompt := fmt.Sprintf(`你是一个专业的%s程序员。
请根据用户需求生成高质量的代码，要求：
1. 代码结构清晰，注释完整
2. 遵循最佳实践和编码规范
3. 包含必要的错误处理
4. 提供使用示例
5. 代码要能够直接运行`, language)

	messages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: fmt.Sprintf("请用%s实现：%s", language, requirement)},
	}

	response, err := cg.chatModel.Generate(ctx, messages,
		model.WithTemperature(0.2), // 低温度确保代码准确性
		model.WithMaxTokens(2000),
	)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// 性能监控器 (简化版本)
type PerformanceMonitor struct {
	startTime time.Time
	metrics   map[string]interface{}
	mu        sync.RWMutex
}

func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics: make(map[string]interface{}),
	}
}

func (p *PerformanceMonitor) StartMonitoring(model string, messageCount int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.startTime = time.Now()
	p.metrics = make(map[string]interface{})
	p.metrics["model"] = model
	p.metrics["input_messages"] = messageCount
	fmt.Printf("[监控] 开始生成，模型: %s, 消息数: %d\n", model, messageCount)
}

func (p *PerformanceMonitor) EndMonitoring(success bool, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	duration := time.Since(p.startTime)
	p.metrics["duration_ms"] = duration.Milliseconds()

	if success {
		p.metrics["status"] = "success"
		fmt.Printf("[监控] 生成完成，耗时: %v\n", duration)
	} else {
		p.metrics["status"] = "failed"
		p.metrics["error"] = err.Error()
		fmt.Printf("[监控] 生成失败: %v\n", err)
	}
	p.sendMetrics()
}

func (p *PerformanceMonitor) sendMetrics() {
	// 发送指标到监控系统（这里只是打印）
	fmt.Printf("[性能指标] %+v\n", p.metrics)
}

// 初始化配置
func initConfig() (*Config, error) {
	// 为了能从 viper 加载配置，先进行初始化
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	_ = viper.ReadInConfig() // 忽略错误，因为我们也会检查环境变量
	return &Config{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("ARK_MODEL"),
		BaseURL: viper.GetString("BASE_URL"),
	}, nil
}

// 初始化ChatModel
func initChatModel(ctx context.Context, config *Config) (model.BaseChatModel, error) {
	timeout := 30 * time.Second
	chatModelConfig := &ark.ChatModelConfig{
		APIKey:  config.APIKey,
		Model:   config.Model,
		Timeout: &timeout,
	}

	if config.BaseURL != "" {
		chatModelConfig.BaseURL = config.BaseURL
	}

	// 创建ChatModel
	cm, err := ark.NewChatModel(ctx, chatModelConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化ChatModel失败: %w", err)
	}

	return cm, nil
}

// 基础使用示例
func basicChatModelExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== 基础ChatModel使用示例 ===")

	// 准备输入消息
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "你是一个有帮助的AI助手，请用简洁明了的方式回答问题。",
		},
		{
			Role:    schema.User,
			Content: "请简单解释什么是人工智能？",
		},
	}

	// 生成响应
	response, err := cm.Generate(ctx, messages,
		model.WithTemperature(0.7),
		model.WithMaxTokens(500),
	)
	if err != nil {
		log.Printf("生成响应失败: %v", err)
		return
	}

	fmt.Printf("AI回复: %s\n", response.Content)

	// 显示响应元信息
	if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		fmt.Printf("Token使用情况: 输入=%d, 输出=%d, 总计=%d\n",
			response.ResponseMeta.Usage.PromptTokens,
			response.ResponseMeta.Usage.CompletionTokens,
			response.ResponseMeta.Usage.TotalTokens)
	}
}

// 流式生成示例
func streamingExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== 流式生成示例 ===")

	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "你是一个创意写作助手，请写一个有趣的短故事。",
		},
		{
			Role:    schema.User,
			Content: "请写一个关于机器人学会做饭的有趣故事，大约100字。",
		},
	}

	streamResult, err := cm.Stream(ctx, messages,
		model.WithTemperature(0.8),
		model.WithMaxTokens(300),
	)
	if err != nil {
		log.Printf("流式生成失败: %v", err)
		return
	}
	defer streamResult.Close()

	fmt.Print("AI正在创作: ")
	for {
		chunk, err := streamResult.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("接收流式数据出错: %v", err)
			break
		}
		// 实时输出响应片段
		fmt.Print(chunk.Content)
		time.Sleep(50 * time.Millisecond) // 模拟打字效果
	}
	fmt.Println()
}

// 编排使用示例
func orchestrationExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== 编排使用示例 ===")

	// 在 Chain 中使用
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(cm)

	// 编译并运行
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译链失败: %v", err)
		return
	}

	messages := []*schema.Message{
		{Role: schema.User, Content: "你好！请介绍一下你自己。"},
	}

	result, err := runnable.Invoke(ctx, messages)
	if err != nil {
		log.Printf("执行链失败: %v", err)
		return
	}

	fmt.Printf("链式处理结果: %s\n", result.Content)

	// 3. 在 Graph 中使用
	graph := compose.NewGraph[[]*schema.Message, *schema.Message]()
	graph.AddChatModelNode("chat_model", cm)

	// 设置图的流程
	graph.AddEdge(compose.START, "chat_model")
	graph.AddEdge("chat_model", compose.END)

	graphRunnable, err := graph.Compile(ctx)
	if err != nil {
		log.Fatal("编译图失败:", err)
	}

	graphResult, err := graphRunnable.Invoke(ctx, messages)
	if err != nil {
		log.Fatal("执行图失败:", err)
	}

	fmt.Printf("图式处理结果: %s\n", graphResult.Content)
}

// 对话管理示例
func conversationExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== 多轮对话示例 ===")

	conversationManager := NewConversationManager(cm)

	// 模拟多轮对话
	queries := []string{
		"你好，我想了解一下Go语言",
		"Go语言有什么特点？",
		"能给我一个简单的Go代码示例吗？",
		"谢谢你的解释！",
	}

	for i, query := range queries {
		fmt.Printf("\n用户[%d]: %s\n", i+1, query)
		response, err := conversationManager.Chat(ctx, query)
		if err != nil {
			log.Printf("对话失败: %v", err)
			continue
		}
		fmt.Printf("AI[%d]: %s\n", i+1, response)
		fmt.Printf("(历史记录长度: %d)\n", conversationManager.GetHistoryLength())
	}
}

// 客服机器人示例
func customerServiceExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== 智能客服示例 ===")

	bot := NewCustomerServiceBot(cm)

	// 模拟客户咨询
	queries := []string{
		"你们的营业时间是什么？",
		"我想退货，需要什么条件？",
		"订单什么时候能到？",
		"你们支持哪些支付方式？",
		"我想投诉，怎么联系你们？",
	}

	for i, query := range queries {
		fmt.Printf("\n客户[%d]: %s\n", i+1, query)
		response, err := bot.HandleCustomerQuery(ctx, query)
		if err != nil {
			log.Printf("客服回复失败: %v", err)
			continue
		}
		fmt.Printf("客服[%d]: %s\n", i+1, response)
	}
}

// 代码生成示例
func codeGenerationExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== 代码生成示例 ===")

	generator := NewCodeGenerator(cm)

	// 生成代码示例
	requirements := []struct {
		desc     string
		language string
	}{
		{"实现一个计算斐波那契数列的函数", "Go"},
		{"创建一个简单的HTTP服务器", "Python"},
	}

	for i, req := range requirements {
		fmt.Printf("\n需求[%d]: 用%s%s\n", i+1, req.language, req.desc)
		code, err := generator.GenerateCode(ctx, req.desc, req.language)
		if err != nil {
			log.Printf("代码生成失败: %v", err)
			continue
		}
		fmt.Printf("生成的代码[%d]:\n%s\n", i+1, code)
		fmt.Println(strings.Repeat("-", 50))
	}
}

// 性能监控示例
func performanceExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== 性能监控示例 ===")

	messages := []*schema.Message{
		{Role: schema.User, Content: "请解释一下什么是云计算？"},
	}

	// 创建性能监控器
	monitor := NewPerformanceMonitor()
	monitor.StartMonitoring("doubao-pro-4k", len(messages))

	// 生成响应并监控
	response, err := cm.Generate(ctx, messages,
		model.WithTemperature(0.7),
		model.WithMaxTokens(500),
	)

	if err != nil {
		monitor.EndMonitoring(false, err)
		log.Printf("生成失败: %v", err)
		return
	}

	monitor.EndMonitoring(true, nil)
	fmt.Printf("\n生成结果: %s\n", response.Content)
}

// 主函数
func main() {
	ctx := context.Background()

	fmt.Println("🤖 Eino ChatModel 组件完全示例")
	fmt.Println("================================")

	// 1. 初始化配置
	config, err := initConfig()
	if err != nil {
		log.Fatal("配置初始化失败:", err)
	}

	fmt.Printf("使用模型: %s\n", config.Model)

	// 2. 初始化ChatModel
	cm, err := initChatModel(ctx, config)
	if err != nil {
		log.Fatal("ChatModel初始化失败:", err)
	}

	fmt.Println("ChatModel 初始化成功！")

	// 3. 运行各种示例
	try := func(name string, fn func(context.Context, model.BaseChatModel)) {
		fmt.Printf("\n正在运行: %s\n", name)
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("示例 %s 运行出错: %v\n", name, r)
			}
		}()
		fn(ctx, cm)
	}

	// 检查是否有特定的示例要运行
	if len(os.Args) > 1 {
		exampleName := os.Args[1]
		switch strings.ToLower(exampleName) {
		case "basic":
			try("基础使用示例", basicChatModelExample)
		case "stream":
			try("流式生成示例", streamingExample)
		case "orchestration":
			try("编排使用示例", orchestrationExample)
		case "conversation":
			try("多轮对话示例", conversationExample)
		case "service":
			try("智能客服示例", customerServiceExample)
		case "code":
			try("代码生成示例", codeGenerationExample)
		case "performance":
			try("性能监控示例", performanceExample)
		default:
			fmt.Printf("未知示例: %s\n", exampleName)
			fmt.Println("可用示例: basic, stream, orchestration, conversation, service, code, performance")
			return
		}
	} else {
		// 运行所有示例
		try("基础使用示例", basicChatModelExample)
		try("流式生成示例", streamingExample)
		try("编排使用示例", orchestrationExample)
		try("多轮对话示例", conversationExample)
		try("智能客服示例", customerServiceExample)
		try("代码生成示例", codeGenerationExample)
		try("性能监控示例", performanceExample)
	}

	fmt.Println("\n🎉 所有示例运行完成！")
	fmt.Println("\n使用方法:")
	fmt.Println("  go run main.go              # 运行所有示例")
	fmt.Println("  go run main.go basic        # 运行基础示例")
	fmt.Println("  go run main.go stream       # 运行流式生成示例")
	fmt.Println("  go run main.go conversation # 运行多轮对话示例")
	fmt.Println("  go run main.go performance  # 运行性能监控示例")
	fmt.Println("  ... 等等")
}
