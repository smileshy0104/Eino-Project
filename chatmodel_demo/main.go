package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
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

// 计算器工具实现
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行基本数学运算，支持加减乘除",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {
				Type:     "string",
				Desc:     "要计算的数学表达式，例如: '123 * 456'",
				Required: true,
			},
		}),
	}, nil
}

func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 简单的计算器实现（这里只是示例）
	result := fmt.Sprintf("计算结果: %s = [模拟计算结果]", argumentsInJSON)
	return result, nil
}

// 天气工具实现
type WeatherTool struct{}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "weather",
		Desc: "查询指定城市的天气信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",
				Desc:     "要查询天气的城市名称",
				Required: true,
			},
		}),
	}, nil
}

func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	result := fmt.Sprintf("城市天气信息: %s [模拟天气数据: 晴天，温度 25°C，湿度 60%%]", argumentsInJSON)
	return result, nil
}

// 工具调用示例
func toolCallingExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== 工具调用集成 ===")

	// 1. 创建工具
	tools := []tool.InvokableTool{
		NewCalculatorTool(),
		NewWeatherTool(),
	}

	// 2. 绑定工具到模型
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Printf("获取工具信息失败: %v", err)
			continue
		}
		toolInfos = append(toolInfos, info)
	}

	// 3. 发送需要工具调用的消息
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "你是一个智能助手，可以使用工具来帮助用户解决问题。",
		},
		{
			Role:    schema.User,
			Content: "请帮我计算 123 * 456 的结果，然后查询北京今天的天气。",
		},
	}

	// 4. 生成响应（可能包含工具调用）
	response, err := cm.Generate(ctx, messages, model.WithTools(toolInfos))
	if err != nil {
		log.Printf("生成响应失败: %v", err)
		return
	}

	// 5. 处理工具调用
	if len(response.ToolCalls) > 0 {
		fmt.Printf("模型请求调用 %d 个工具:\n", len(response.ToolCalls))

		for _, toolCall := range response.ToolCalls {
			fmt.Printf("- 工具: %s, 参数: %s\n",
				toolCall.Function.Name,
				toolCall.Function.Arguments)

			// 执行工具调用
			for _, t := range tools {
				info, _ := t.Info(ctx)
				if info.Name == toolCall.Function.Name {
					result, err := t.InvokableRun(ctx, toolCall.Function.Arguments)
					if err != nil {
						log.Printf("工具调用失败: %v", err)
						continue
					}
					fmt.Printf("  结果: %s\n", result)
				}
			}
		}
	} else {
		fmt.Printf("直接回复: %s\n", response.Content)
	}
}

// 自定义回调处理器
type ChatModelCallbackHandler struct {
	startTime time.Time
	requestID string
}

// 生成请求开始时的回调
func (h *ChatModelCallbackHandler) OnGenerateStart(ctx context.Context, messages []*schema.Message) {
	h.startTime = time.Now()
	h.requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	fmt.Printf("[回调] 🚀 开始生成 (ID: %s)\n", h.requestID)
	fmt.Printf("[回调] 📝 消息数量: %d\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("[回调]   消息%d [%s]: %s\n", i+1, msg.Role,
			truncateString(msg.Content, 50))
	}
}

// 生成请求成功时的回调
func (h *ChatModelCallbackHandler) OnGenerateSuccess(ctx context.Context, response *schema.Message) {
	duration := time.Since(h.startTime)
	fmt.Printf("[回调] ✅ 生成成功 (ID: %s)\n", h.requestID)
	fmt.Printf("[回调] ⏱️ 耗时: %v\n", duration)
	fmt.Printf("[回调] 📤 响应长度: %d 字符\n", len(response.Content))

	// 如果有使用统计信息，显示
	if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		usage := response.ResponseMeta.Usage
		fmt.Printf("[回调] 🔢 Token使用: 输入=%d, 输出=%d, 总计=%d\n",
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)
	}
}

// 生成请求失败时的回调
func (h *ChatModelCallbackHandler) OnGenerateError(ctx context.Context, err error) {
	duration := time.Since(h.startTime)
	fmt.Printf("[回调] ❌ 生成失败 (ID: %s)\n", h.requestID)
	fmt.Printf("[回调] ⏱️ 耗时: %v\n", duration)
	fmt.Printf("[回调] 💥 错误: %v\n", err)
}

// 流式数据接收时的回调
func (h *ChatModelCallbackHandler) OnStreamChunk(ctx context.Context, chunk string) {
	fmt.Printf("[回调] 🔄 流式数据: %s", chunk)
}

// 字符串截断工具函数
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// 简化的回调执行函数
func executeWithCallback(handler *ChatModelCallbackHandler, ctx context.Context, messages []*schema.Message, cm model.BaseChatModel, opts ...model.Option) (*schema.Message, error) {
	// 调用开始回调
	handler.OnGenerateStart(ctx, messages)

	// 执行实际生成
	response, err := cm.Generate(ctx, messages, opts...)

	// 根据结果调用相应回调
	if err != nil {
		handler.OnGenerateError(ctx, err)
		return nil, err
	}

	handler.OnGenerateSuccess(ctx, response)
	return response, nil
}

// 简化的流式回调执行函数
func executeStreamWithCallback(handler *ChatModelCallbackHandler, ctx context.Context, messages []*schema.Message, cm model.BaseChatModel, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 调用开始回调
	handler.OnGenerateStart(ctx, messages)

	// 执行实际流式生成
	streamResult, err := cm.Stream(ctx, messages, opts...)
	if err != nil {
		handler.OnGenerateError(ctx, err)
		return nil, err
	}

	return streamResult, nil
}

// 回调监控示例
func callbackExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== Eino 官方回调系统示例 ===")
	fmt.Println("演示如何使用 Eino 官方回调机制监控ChatModel的调用过程")

	// 1. 创建增强版 Eino 回调处理器 - 集成性能监控、日志记录、错误统计
	startTime := time.Now()
	requestCount := 0
	errorCount := 0

	callbackHandler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			requestCount++
			startTime = time.Now()

			fmt.Printf("[增强回调] 🚀 第%d个请求开始执行\n", requestCount)
			fmt.Printf("[增强回调] ⏰ 开始时间: %s\n", startTime.Format("15:04:05.000"))

			// 分析输入数据
			inputStr := fmt.Sprintf("%v", input)
			inputSize := len(inputStr)
			fmt.Printf("[增强回调] 📊 输入分析: 大小=%d字符\n", inputSize)

			// 截断显示输入内容
			if inputSize > 150 {
				fmt.Printf("[增强回调] 📝 输入内容: %s...[截断]\n", inputStr[:150])
			} else {
				fmt.Printf("[增强回调] 📝 输入内容: %s\n", inputStr)
			}

			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			duration := time.Since(startTime)

			fmt.Printf("[增强回调] ✅ 请求成功完成\n")
			fmt.Printf("[增强回调] ⏱️ 执行耗时: %v\n", duration)
			fmt.Printf("[增强回调] 🏎️ 平均速度: %.2f ms/请求\n",
				float64(duration.Milliseconds()))

			// 分析输出数据
			outputStr := fmt.Sprintf("%v", output)
			outputSize := len(outputStr)
			fmt.Printf("[增强回调] 📊 输出分析: 大小=%d字符\n", outputSize)

			// 计算处理效率
			if duration.Seconds() > 0 {
				charPerSec := float64(outputSize) / duration.Seconds()
				fmt.Printf("[增强回调] 📈 处理效率: %.2f字符/秒\n", charPerSec)
			}

			// 截断显示输出内容
			if outputSize > 100 {
				fmt.Printf("[增强回调] 📤 输出内容: %s...[截断]\n", outputStr[:100])
			} else {
				fmt.Printf("[增强回调] 📤 输出内容: %s\n", outputStr)
			}

			// 统计信息
			fmt.Printf("[增强回调] 📊 会话统计: 总请求=%d, 成功率=%.1f%%\n",
				requestCount, float64(requestCount-errorCount)*100/float64(requestCount))

			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			duration := time.Since(startTime)
			errorCount++

			fmt.Printf("[增强回调] ❌ 请求执行失败\n")
			fmt.Printf("[增强回调] ⏱️ 失败耗时: %v\n", duration)
			fmt.Printf("[增强回调] 💥 错误详情: %v\n", err)

			// 错误分类
			errorMsg := err.Error()
			if strings.Contains(errorMsg, "timeout") {
				fmt.Printf("[增强回调] 🔍 错误类型: 超时错误\n")
			} else if strings.Contains(errorMsg, "token") {
				fmt.Printf("[增强回调] 🔍 错误类型: 认证/Token错误\n")
			} else if strings.Contains(errorMsg, "rate limit") {
				fmt.Printf("[增强回调] 🔍 错误类型: 频率限制\n")
			} else {
				fmt.Printf("[增强回调] 🔍 错误类型: 其他错误\n")
			}

			// 统计信息
			fmt.Printf("[增强回调] 📊 错误统计: 总请求=%d, 错误数=%d, 成功率=%.1f%%\n",
				requestCount, errorCount, float64(requestCount-errorCount)*100/float64(requestCount))

			return ctx
		}).
		Build()

	// 2. 基础生成示例 - 使用 Chain 集成回调
	fmt.Println("\n--- 使用 Chain 的回调示例 ---")
	messages := []*schema.Message{
		{Role: schema.System, Content: "你是一个友好的AI助手。"},
		{Role: schema.User, Content: "请简单介绍一下回调机制在软件开发中的作用。"},
	}

	// 创建带回调的 Chain
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(cm)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译 Chain 失败: %v", err)
		return
	}

	// 使用回调执行
	response, err := runnable.Invoke(ctx, messages,
		compose.WithCallbacks(callbackHandler),
	)

	if err != nil {
		log.Printf("Chain 执行失败: %v", err)
		// 继续执行自定义回调示例
		fmt.Println("\n--- 备用自定义回调示例 ---")
		customHandler := &ChatModelCallbackHandler{}
		executeWithCallback(customHandler, ctx, messages, cm,
			model.WithTemperature(0.7),
			model.WithMaxTokens(300),
		)
		return
	}

	fmt.Printf("\n📄 Chain 响应: %s\n", response.Content)

	// 3. Graph 回调示例
	fmt.Println("\n--- 使用 Graph 的回调示例 ---")
	graph := compose.NewGraph[[]*schema.Message, *schema.Message]()
	graph.AddChatModelNode("chat", cm)
	graph.AddEdge(compose.START, "chat")
	graph.AddEdge("chat", compose.END)

	graphRunnable, err := graph.Compile(ctx)
	if err != nil {
		log.Printf("编译 Graph 失败: %v", err)
		return
	}

	graphResponse, err := graphRunnable.Invoke(ctx, messages,
		compose.WithCallbacks(callbackHandler),
	)

	if err != nil {
		log.Printf("Graph 执行失败: %v", err)
		return
	}

	fmt.Printf("\n📄 Graph 响应: %s\n", graphResponse.Content)

	// 4. 演示错误回调
	fmt.Println("\n--- 错误处理回调示例 ---")
	errorMessages := []*schema.Message{
		{Role: schema.User, Content: strings.Repeat("测试超长文本", 500)},
	}

	_, err = runnable.Invoke(ctx, errorMessages,
		compose.WithCallbacks(callbackHandler),
	)

	if err != nil {
		fmt.Printf("✅ 错误回调成功触发，这是预期的行为\n")
	}

	fmt.Println("\n📝 Eino 官方回调机制总结:")
	fmt.Println("✅ 使用 callbacks.NewHandlerBuilder() 创建回调处理器")
	fmt.Println("✅ 支持 OnChatModelStart、OnChatModelEnd、OnChatModelError 等事件")
	fmt.Println("✅ 可以与 Chain 和 Graph 无缝集成")
	fmt.Println("✅ 提供标准化的回调接口和数据结构")
	fmt.Println("🎯 Eino 官方回调系统更加稳定和功能完整")
}

// 高级回调处理器 - 包含详细的监控和分析功能
func advancedCallbackExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== 高级回调处理器示例 ===")
	fmt.Println("演示带有详细监控、日志记录和性能分析的回调处理器")

	// 创建会话级别的统计数据
	sessionStats := struct {
		startTime      time.Time
		requestCount   int
		successCount   int
		errorCount     int
		totalDuration  time.Duration
		averageLatency time.Duration
		maxLatency     time.Duration
		minLatency     time.Duration
		mu             sync.RWMutex
	}{
		startTime:  time.Now(),
		minLatency: time.Hour, // 初始化为很大的值
	}

	// 创建高级回调处理器
	advancedHandler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			sessionStats.mu.Lock()
			sessionStats.requestCount++
			currentReq := sessionStats.requestCount
			sessionStats.mu.Unlock()

			fmt.Printf("\n[高级回调] 🎯 === 请求 #%d 开始 ===\n", currentReq)
			fmt.Printf("[高级回调] 📅 时间戳: %s\n", time.Now().Format("2006-01-02 15:04:05.000"))

			// 详细的输入分析
			inputStr := fmt.Sprintf("%v", input)
			inputSize := len(inputStr)
			wordCount := len(strings.Fields(inputStr))

			fmt.Printf("[高级回调] 📊 输入统计:\n")
			fmt.Printf("[高级回调]   - 总字符数: %d\n", inputSize)
			fmt.Printf("[高级回调]   - 单词数量: %d\n", wordCount)
			fmt.Printf("[高级回调]   - 平均单词长度: %.1f\n", float64(inputSize)/float64(wordCount))

			// 内容复杂度分析
			complexity := "简单"
			if inputSize > 500 {
				complexity = "复杂"
			} else if inputSize > 200 {
				complexity = "中等"
			}
			fmt.Printf("[高级回调]   - 内容复杂度: %s\n", complexity)

			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			endTime := time.Now()
			duration := endTime.Sub(sessionStats.startTime)

			sessionStats.mu.Lock()
			sessionStats.successCount++
			sessionStats.totalDuration += duration
			sessionStats.averageLatency = sessionStats.totalDuration / time.Duration(sessionStats.successCount)

			if duration > sessionStats.maxLatency {
				sessionStats.maxLatency = duration
			}
			if duration < sessionStats.minLatency {
				sessionStats.minLatency = duration
			}

			currentSuccess := sessionStats.successCount
			currentReq := sessionStats.requestCount
			sessionStats.mu.Unlock()

			fmt.Printf("\n[高级回调] ✨ === 请求 #%d 成功完成 ===\n", currentReq)
			fmt.Printf("[高级回调] ⏱️ 性能指标:\n")
			fmt.Printf("[高级回调]   - 执行耗时: %v\n", duration)
			fmt.Printf("[高级回调]   - 平均延迟: %v\n", sessionStats.averageLatency)
			fmt.Printf("[高级回调]   - 最大延迟: %v\n", sessionStats.maxLatency)
			fmt.Printf("[高级回调]   - 最小延迟: %v\n", sessionStats.minLatency)

			// 输出分析
			outputStr := fmt.Sprintf("%v", output)
			outputSize := len(outputStr)
			outputWords := len(strings.Fields(outputStr))

			fmt.Printf("[高级回调] 📊 输出统计:\n")
			fmt.Printf("[高级回调]   - 响应字符数: %d\n", outputSize)
			fmt.Printf("[高级回调]   - 响应单词数: %d\n", outputWords)

			// 处理效率
			if duration.Seconds() > 0 {
				wps := float64(outputWords) / duration.Seconds()
				cps := float64(outputSize) / duration.Seconds()
				fmt.Printf("[高级回调]   - 处理速度: %.1f单词/秒, %.1f字符/秒\n", wps, cps)
			}

			// 质量评估
			quality := "良好"
			if outputSize > 1000 {
				quality = "详细"
			} else if outputSize < 50 {
				quality = "简洁"
			}
			fmt.Printf("[高级回调]   - 响应质量: %s\n", quality)

			// 会话统计
			successRate := float64(currentSuccess) * 100 / float64(currentReq)
			fmt.Printf("[高级回调] 📈 会话统计:\n")
			fmt.Printf("[高级回调]   - 总请求数: %d\n", currentReq)
			fmt.Printf("[高级回调]   - 成功率: %.1f%%\n", successRate)
			fmt.Printf("[高级回调]   - 会话时长: %v\n", endTime.Sub(sessionStats.startTime))

			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			endTime := time.Now()
			duration := endTime.Sub(sessionStats.startTime)

			sessionStats.mu.Lock()
			sessionStats.errorCount++
			currentError := sessionStats.errorCount
			currentReq := sessionStats.requestCount
			sessionStats.mu.Unlock()

			fmt.Printf("\n[高级回调] 🚨 === 请求 #%d 执行失败 ===\n", currentReq)
			fmt.Printf("[高级回调] ⏱️ 失败耗时: %v\n", duration)

			// 详细的错误分析
			errorMsg := err.Error()
			fmt.Printf("[高级回调] 🔍 错误分析:\n")
			fmt.Printf("[高级回调]   - 错误信息: %v\n", err)

			// 错误分类和建议
			if strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "deadline") {
				fmt.Printf("[高级回调]   - 错误类型: ⏰ 超时错误\n")
				fmt.Printf("[高级回调]   - 建议: 增加超时时间或减少请求复杂度\n")
			} else if strings.Contains(errorMsg, "token") || strings.Contains(errorMsg, "auth") {
				fmt.Printf("[高级回调]   - 错误类型: 🔐 认证错误\n")
				fmt.Printf("[高级回调]   - 建议: 检查 API Key 或认证配置\n")
			} else if strings.Contains(errorMsg, "rate limit") || strings.Contains(errorMsg, "quota") {
				fmt.Printf("[高级回调]   - 错误类型: 🚦 频率限制\n")
				fmt.Printf("[高级回调]   - 建议: 降低请求频率或升级配额\n")
			} else if strings.Contains(errorMsg, "network") || strings.Contains(errorMsg, "connection") {
				fmt.Printf("[高级回调]   - 错误类型: 🌐 网络错误\n")
				fmt.Printf("[高级回调]   - 建议: 检查网络连接或重试请求\n")
			} else {
				fmt.Printf("[高级回调]   - 错误类型: ❓ 未知错误\n")
				fmt.Printf("[高级回调]   - 建议: 检查输入参数或联系技术支持\n")
			}

			// 错误统计和趋势
			errorRate := float64(currentError) * 100 / float64(currentReq)
			fmt.Printf("[高级回调] 📊 错误统计:\n")
			fmt.Printf("[高级回调]   - 错误次数: %d\n", currentError)
			fmt.Printf("[高级回调]   - 错误率: %.1f%%\n", errorRate)

			// 健康度评估
			if errorRate > 50 {
				fmt.Printf("[高级回调]   - 系统状态: 🔴 需要关注\n")
			} else if errorRate > 20 {
				fmt.Printf("[高级回调]   - 系统状态: 🟡 轻微问题\n")
			} else {
				fmt.Printf("[高级回调]   - 系统状态: 🟢 运行良好\n")
			}

			return ctx
		}).
		Build()

	// 运行高级回调示例
	fmt.Println("\n--- 高级回调测试序列 ---")

	testCases := []struct {
		name    string
		message string
		options []model.Option
	}{
		{
			name:    "简单问答",
			message: "你好，请问今天天气怎么样？",
			options: []model.Option{model.WithTemperature(0.7)},
		},
		{
			name:    "复杂分析",
			message: "请详细分析人工智能在医疗诊断领域的应用前景，包括技术挑战、伦理问题和未来发展趋势。",
			options: []model.Option{model.WithTemperature(0.5), model.WithMaxTokens(800)},
		},
		{
			name:    "错误测试",
			message: strings.Repeat("这是一个超长的测试消息，用于触发可能的错误。", 100),
			options: []model.Option{model.WithMaxTokens(1)}, // 故意设置很小的限制
		},
	}

	// 创建测试链
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(cm)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译测试链失败: %v", err)
		return
	}

	// 执行测试用例
	for i, testCase := range testCases {
		fmt.Printf("\n[高级回调] 🧪 测试用例 %d: %s\n", i+1, testCase.name)

		messages := []*schema.Message{
			{Role: schema.System, Content: "你是一个专业的AI助手，请提供准确和有用的回答。"},
			{Role: schema.User, Content: testCase.message},
		}

		// 合并选项
		allOptions := []compose.Option{compose.WithCallbacks(advancedHandler)}

		_, err := runnable.Invoke(ctx, messages, allOptions...)
		if err != nil {
			// 错误已经在回调中处理了，这里只做简单记录
			continue
		}

		// 在测试用例之间暂停一下
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n[高级回调] 🏁 === 测试序列完成 ===\n")
	fmt.Printf("[高级回调] 📊 最终统计: 总测试=%d, 成功=%d, 失败=%d\n",
		sessionStats.requestCount, sessionStats.successCount, sessionStats.errorCount)
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
		case "tool":
			try("工具调用示例", toolCallingExample)
		case "callback":
			try("回调监控示例", callbackExample)
		case "advanced":
			try("高级回调示例", advancedCallbackExample)
		default:
			fmt.Printf("未知示例: %s\n", exampleName)
			fmt.Println("可用示例: basic, stream, orchestration, conversation, service, code, performance, tool, callback, advanced")
			return
		}
	} else {
		// 运行所有示例
		//try("基础使用示例", basicChatModelExample)
		//try("流式生成示例", streamingExample)
		//try("编排使用示例", orchestrationExample)
		//try("多轮对话示例", conversationExample)
		//try("智能客服示例", customerServiceExample)
		//try("代码生成示例", codeGenerationExample)
		//try("性能监控示例", performanceExample)
		//try("工具调用示例", toolCallingExample)
		try("回调监控示例", callbackExample)
		//try("高级回调示例", advancedCallbackExample)
	}

	fmt.Println("\n🎉 所有示例运行完成！")
	fmt.Println("\n使用方法:")
	fmt.Println("  go run main.go              # 运行所有示例")
	fmt.Println("  go run main.go basic        # 运行基础示例")
	fmt.Println("  go run main.go stream       # 运行流式生成示例")
	fmt.Println("  go run main.go conversation # 运行多轮对话示例")
	fmt.Println("  go run main.go performance  # 运行性能监控示例")
	fmt.Println("  go run main.go tool         # 运行工具调用示例")
	fmt.Println("  go run main.go callback     # 运行回调监控示例")
	fmt.Println("  go run main.go advanced     # 运行高级回调示例")
	fmt.Println("  ... 等等")
}
