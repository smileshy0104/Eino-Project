package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// =============================================================================
//
//  文件: toolsnode_example.go
//  功能: 演示如何创建和使用 ToolsNode 来管理多个工具
//  说明: ToolsNode 是 Eino 中管理工具调用的核心组件，可以集成到 Chain 或 Graph 中
//
// =============================================================================

// --- 工具接口和基础结构 ---

// BaseTool 定义了所有工具必须实现的基础接口
// 提供工具的元信息，包括名称、描述、参数定义等
type BaseTool interface {
	// Info 返回工具的基本信息和参数定义
	// 用于工具注册和参数验证
	Info(ctx context.Context) (*schema.ToolInfo, error)
}

// InvokableTool 定义了可调用工具的完整接口
// 继承自 BaseTool，并添加了实际执行功能
type InvokableTool interface {
	BaseTool
	// InvokableRun 执行工具的核心逻辑
	// 参数:
	//   ctx: 上下文对象，用于取消和超时控制
	//   argumentsInJSON: JSON 格式的工具参数
	//   opts: 可选的执行选项
	// 返回:
	//   string: 工具执行结果（JSON 格式）
	//   error: 执行过程中的错误
	InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error)
}

// --- 工具实现 ---

// WeatherTool 天气查询工具
// 实现了 InvokableTool 接口，提供城市天气信息查询功能
// 支持指定城市和日期的天气查询，如果不指定日期则查询当天天气
type WeatherTool struct{}

// Info 返回天气工具的元信息和参数定义
func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather", // 工具名称，用于工具调用时的标识
		Desc: "查询指定城市的天气信息", // 工具描述，帮助 LLM 理解工具功能
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",
				Desc:     "城市名称",
				Required: true, // 必需参数
			},
			"date": {
				Type:     "string",
				Desc:     "查询日期 (YYYY-MM-DD)",
				Required: false, // 可选参数，不提供时默认为当天
			},
		}),
	}, nil
}

// InvokableRun 执行天气查询逻辑
func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 定义参数结构体，用于解析 JSON 参数
	var args struct {
		City string `json:"city"` // 城市名称
		Date string `json:"date"` // 查询日期（可选）
	}

	// 解析 JSON 参数到结构体
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 如果未指定日期，使用当前日期
	if args.Date == "" {
		args.Date = time.Now().Format("2006-01-02")
	}

	log.Printf("[WeatherTool] 查询 %s 在 %s 的天气", args.City, args.Date)

	// 模拟天气数据生成（实际应用中这里会调用真实的天气 API）
	weatherData := map[string]interface{}{
		"city":        args.City,                                // 城市名称
		"date":        args.Date,                                // 查询日期
		"temperature": 25,                                       // 温度（摄氏度）
		"humidity":    60,                                       // 湿度（百分比）
		"condition":   "晴朗",                                     // 天气状况
		"wind_speed":  "5 km/h",                                 // 风速
		"description": fmt.Sprintf("%s 今天天气晴朗，温度适宜", args.City), // 天气描述
	}

	// 将结果序列化为 JSON 字符串返回
	result, _ := json.Marshal(weatherData)
	return string(result), nil
}

// CalculatorTool 计算器工具
// 实现简单的数学表达式计算功能，支持基本的算术运算
// 注意：这是一个演示实现，实际应用中应使用更安全和完整的表达式解析器
type CalculatorTool struct{}

// Info 返回计算器工具的元信息和参数定义
func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator", // 工具名称
		Desc: "执行数学计算",     // 工具描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {
				Type:     "string",
				Desc:     "数学表达式",
				Required: true, // 表达式是必需参数
			},
		}),
	}, nil
}

// InvokableRun 执行数学计算逻辑
func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 定义参数结构体
	var args struct {
		Expression string `json:"expression"` // 要计算的数学表达式
	}

	// 解析 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[CalculatorTool] 计算表达式: %s", args.Expression)

	// 执行表达式计算（简单的实现，仅支持基本加法运算）
	// 实际应用中应该使用更安全和完整的表达式解析器，如 govaluate 等
	result := evaluateSimpleExpression(args.Expression)

	// 构造响应结果
	response := map[string]interface{}{
		"expression": args.Expression,                 // 原始表达式
		"result":     result,                          // 计算结果
		"timestamp":  time.Now().Format(time.RFC3339), // 计算时间戳
	}

	// 序列化结果并返回
	resultBytes, _ := json.Marshal(response)
	return string(resultBytes), nil
}

// TranslatorTool 翻译工具
// 实现文本翻译功能，支持多种语言之间的转换
// 支持自动检测源语言或手动指定源语言
type TranslatorTool struct{}

// Info 返回翻译工具的元信息和参数定义
func (t *TranslatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "translator", // 工具名称
		Desc: "翻译文本",       // 工具描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"text": {
				Type:     "string",
				Desc:     "要翻译的文本",
				Required: true, // 必需参数
			},
			"from_lang": {
				Type:     "string",
				Desc:     "源语言",
				Required: false, // 可选参数，不指定时自动检测
			},
			"to_lang": {
				Type:     "string",
				Desc:     "目标语言",
				Required: true, // 必需参数
			},
		}),
	}, nil
}

// InvokableRun 执行文本翻译逻辑
func (t *TranslatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 定义参数结构体
	var args struct {
		Text     string `json:"text"`      // 待翻译的原文
		FromLang string `json:"from_lang"` // 源语言（可选）
		ToLang   string `json:"to_lang"`   // 目标语言（必需）
	}

	// 解析 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 如果未指定源语言，设置为自动检测
	if args.FromLang == "" {
		args.FromLang = "auto"
	}

	log.Printf("[TranslatorTool] 翻译 '%s' 从 %s 到 %s", args.Text, args.FromLang, args.ToLang)

	// 执行翻译逻辑（这里使用模拟翻译，实际应用中会调用翻译 API）
	translatedText := simulateTranslation(args.Text, args.FromLang, args.ToLang)

	// 构造翻译结果响应
	response := map[string]interface{}{
		"original_text":   args.Text,      // 原始文本
		"translated_text": translatedText, // 翻译后的文本
		"from_language":   args.FromLang,  // 源语言
		"to_language":     args.ToLang,    // 目标语言
		"confidence":      0.95,           // 翻译置信度（模拟值）
	}

	// 序列化结果并返回
	result, _ := json.Marshal(response)
	return string(result), nil
}

// FileManagerTool 文件管理工具
// 实现基本的文件系统操作功能，包括文件和目录的创建、读取、删除、列表等操作
// 注意：这是演示实现，实际应用中需要考虑安全性和权限控制
type FileManagerTool struct{}

// Info 返回文件管理工具的元信息和参数定义
func (f *FileManagerTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_manager", // 工具名称
		Desc: "管理文件和目录",      // 工具描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"action": {
				Type:     "string",
				Desc:     "操作类型",
				Required: true,
				Enum:     []string{"list", "create", "read", "delete", "info"}, // 支持的操作类型枚举
			},
			"path": {
				Type:     "string",
				Desc:     "文件路径",
				Required: true, // 文件路径是必需参数
			},
			"content": {
				Type:     "string",
				Desc:     "文件内容（创建文件时使用）",
				Required: false, // 仅在创建文件时需要
			},
		}),
	}, nil
}

// InvokableRun 执行文件管理操作逻辑
func (f *FileManagerTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 定义参数结构体
	var args struct {
		Action  string `json:"action"`  // 操作类型（list/create/read/delete/info）
		Path    string `json:"path"`    // 目标文件或目录路径
		Content string `json:"content"` // 文件内容（仅在创建文件时使用）
	}

	// 解析 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[FileManagerTool] 执行 %s 操作，路径: %s", args.Action, args.Path)

	// 执行文件操作（这里使用模拟实现，实际应用中会进行真实的文件系统操作）
	// 注意：实际应用中需要添加安全检查，防止路径遍历攻击等安全问题
	result := simulateFileOperation(args.Action, args.Path, args.Content)

	// 序列化结果并返回
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// --- ToolsNode 演示 ---

// demonstrateToolsNode 演示 ToolsNode 的完整使用流程
// 展示如何创建、注册工具，以及如何处理单个和多个工具调用
func demonstrateToolsNode() {
	ctx := context.Background()

	fmt.Println("=== ToolsNode 完整使用演示 ===")

	// 1. 创建各种工具实例
	// 每个工具都实现了 InvokableTool 接口，提供特定的功能
	weatherTool := &WeatherTool{}         // 天气查询工具
	calculatorTool := &CalculatorTool{}   // 数学计算工具
	translatorTool := &TranslatorTool{}   // 文本翻译工具
	fileManagerTool := &FileManagerTool{} // 文件管理工具

	// 2. 将所有工具组织成工具列表
	// ToolsNode 将基于这个列表来管理和调用工具
	tools := []InvokableTool{
		weatherTool,
		calculatorTool,
		translatorTool,
		fileManagerTool,
	}

	// 3. 创建 ToolsNode 实例
	// 注意：这里使用模拟的 ToolsNode 实现，实际项目中应该使用 compose.NewToolsNode()
	// ToolsNode 负责管理工具注册、工具调用路由、并行执行等功能
	toolsNode := NewMockToolsNode(tools)

	// 显示已注册的工具信息
	fmt.Printf("已创建 ToolsNode，注册了 %d 个工具:\n", len(tools))
	for _, tool := range tools {
		info, _ := tool.Info(ctx)
		fmt.Printf("  - %s: %s\n", info.Name, info.Desc)
	}
	fmt.Println()

	// 4. 演示单个工具调用
	fmt.Println("--- 演示单个工具调用 ---")

	// 创建一个模拟的 LLM 消息，表示 AI 助手决定调用天气工具
	// 在实际应用中，这种消息通常由 LLM 生成，包含工具调用的指令
	weatherMessage := &schema.Message{
		Role: "assistant", // 消息来自 AI 助手
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_weather_001", // 工具调用的唯一标识符
				Type: "function",         // 调用类型为函数调用
				Function: schema.FunctionCall{
					Name:      "get_weather",                          // 要调用的工具名称
					Arguments: `{"city": "北京", "date": "2024-08-19"}`, // JSON 格式的工具参数
				},
			},
		},
	}

	// 通过 ToolsNode 执行工具调用
	// ToolsNode 会根据工具名称找到对应的工具，并传递参数执行
	weatherResults, err := toolsNode.Invoke(ctx, weatherMessage)
	if err != nil {
		log.Printf("天气工具调用失败: %v", err)
	} else {
		// 显示工具执行结果
		// 结果是一个 Message 数组，每个 Message 包含一个工具的执行结果
		fmt.Printf("天气查询结果: %s\n\n", weatherResults[0].Content)
	}

	// 5. 演示多工具并行调用
	fmt.Println("--- 演示多工具并行调用 ---")

	// 创建包含多个工具调用的消息
	// 这展示了 ToolsNode 的一个重要特性：可以在一个调用中并行执行多个工具
	multiToolMessage := &schema.Message{
		Role: "assistant", // 消息来自 AI 助手
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_calc_001", // 计算器工具调用
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "calculator",
					Arguments: `{"expression": "25 + 17"}`, // 计算 25+17
				},
			},
			{
				ID:   "call_translate_001", // 翻译工具调用
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "translator",
					Arguments: `{"text": "Hello World", "to_lang": "zh"}`, // 将 "Hello World" 翻译成中文
				},
			},
			{
				ID:   "call_file_001", // 文件管理工具调用
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "file_manager",
					Arguments: `{"action": "info", "path": "/tmp/example.txt"}`, // 获取文件信息
				},
			},
		},
	}

	// 执行多工具调用
	// ToolsNode 会并行执行所有工具调用，提高效率
	multiResults, err := toolsNode.Invoke(ctx, multiToolMessage)
	if err != nil {
		log.Printf("多工具调用失败: %v", err)
	} else {
		// 显示所有工具的执行结果
		fmt.Printf("并行调用了 %d 个工具:\n", len(multiResults))
		for i, result := range multiResults {
			fmt.Printf("  工具 %d - %s: %s\n", i+1, result.Name, result.Content)
		}
		fmt.Println()
	}

	// 6. 演示在 Chain 中使用 ToolsNode
	fmt.Println("--- 演示在 Chain 中使用 ToolsNode ---")
	// 调用专门的函数来演示 ToolsNode 在工作流链中的使用
	demonstrateToolsNodeInChain(toolsNode)

	// 7. 演示错误处理
	fmt.Println("--- 演示错误处理 ---")

	// 创建一个调用不存在工具的消息，用于测试错误处理机制
	errorMessage := &schema.Message{
		Role: "assistant",
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_error_001",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "nonexistent_tool", // 故意使用不存在的工具名
					Arguments: `{"param": "value"}`,
				},
			},
		},
	}

	// 尝试调用不存在的工具
	// 这演示了 ToolsNode 如何处理错误情况
	errorResults, err := toolsNode.Invoke(ctx, errorMessage)
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	} else if len(errorResults) == 0 {
		fmt.Println("未找到对应工具，跳过执行")
	}
}

// --- Chain 中使用 ToolsNode 的演示 ---

// demonstrateToolsNodeInChain 演示 ToolsNode 在工作流链（Chain）中的使用
// 展示了典型的 LLM + 工具调用的完整流程
func demonstrateToolsNodeInChain(toolsNode *MockToolsNode) {
	ctx := context.Background()

	// 说明：这是一个模拟实现，展示 ToolsNode 在真实工作流中的使用场景
	// 在实际项目中，应该使用 Eino 的 compose.NewChain() 来创建真正的工作流链
	fmt.Println("模拟 Chain 工作流:")
	fmt.Println("  1. 用户消息 -> LLM 生成")         // 用户输入传递给 LLM
	fmt.Println("  2. LLM 生成 -> ToolsNode 执行") // LLM 决定调用工具，ToolsNode 执行
	fmt.Println("  3. 工具结果 -> LLM 最终回复")       // 工具结果返回给 LLM，生成最终回复

	// 模拟 LLM 的决策输出
	// 在真实场景中，这个消息会由 LLM 根据用户输入自动生成
	llmDecision := &schema.Message{
		Role:    "assistant",
		Content: "我需要查询天气和进行一些计算来回答用户的问题。", // LLM 的推理过程
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_weather_chain",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"city": "上海"}`, // 查询上海天气
				},
			},
			{
				ID:   "call_calc_chain",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "calculator",
					Arguments: `{"expression": "20 + 5"}`, // 计算 20+5
				},
			},
		},
	}

	// ToolsNode 执行工具调用
	// 这是 Chain 中的关键步骤，将 LLM 的工具调用指令转换为实际的工具执行
	toolResults, err := toolsNode.Invoke(ctx, llmDecision)
	if err != nil {
		log.Printf("Chain 中工具调用失败: %v", err)
		return
	}

	// 显示工具执行结果
	fmt.Printf("Chain 执行完成，工具调用结果:\n")
	for _, result := range toolResults {
		fmt.Printf("  - %s: %s\n", result.Name, result.Content)
	}

	// 在真实的 Chain 中，这些工具结果会被传递回 LLM
	// LLM 会基于工具结果生成最终的用户回复
	fmt.Println("  → 将工具结果传递给 LLM 生成最终用户回复")
	fmt.Println()
}

// --- 模拟的 ToolsNode 实现 ---

// MockToolsNode 模拟的 ToolsNode 实现
// 这是为了演示而创建的简化版本，展示 ToolsNode 的核心功能
// 实际项目中应该使用 Eino 框架提供的官方 ToolsNode 实现
type MockToolsNode struct {
	tools map[string]InvokableTool // 工具名称到工具实例的映射
}

// NewMockToolsNode 创建新的模拟 ToolsNode 实例
// 接收工具列表，并将其注册到内部的工具映射中
func NewMockToolsNode(tools []InvokableTool) *MockToolsNode {
	// 创建工具名称到工具实例的映射
	toolMap := make(map[string]InvokableTool)

	// 遍历所有工具，获取其名称并建立映射
	for _, tool := range tools {
		info, err := tool.Info(context.Background())
		if err != nil {
			log.Printf("获取工具信息失败: %v", err)
			continue // 跳过有问题的工具
		}
		toolMap[info.Name] = tool // 使用工具名称作为键
	}

	return &MockToolsNode{tools: toolMap}
}

// Invoke 执行工具调用的核心方法
// 接收包含工具调用指令的消息，返回工具执行结果
func (n *MockToolsNode) Invoke(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
	// 只处理来自助手且包含工具调用的消息
	if msg.Role != "assistant" || len(msg.ToolCalls) == 0 {
		return nil, nil
	}

	var results []*schema.Message

	// 遍历所有工具调用请求
	for _, call := range msg.ToolCalls {
		// 根据工具名称查找对应的工具实例
		tool, exists := n.tools[call.Function.Name]
		if !exists {
			log.Printf("工具 '%s' 不存在", call.Function.Name)
			continue // 跳过不存在的工具
		}

		// 执行工具调用
		output, err := tool.InvokableRun(ctx, call.Function.Arguments)
		if err != nil {
			// 如果工具执行失败，记录错误并返回错误信息
			log.Printf("工具 '%s' 执行失败: %v", call.Function.Name, err)
			output = fmt.Sprintf(`{"error": "%s"}`, err.Error())
		}

		// 将工具执行结果包装成消息格式
		result := &schema.Message{
			Role:       "tool",             // 消息来源为工具
			Content:    output,             // 工具执行结果
			Name:       call.Function.Name, // 工具名称
			ToolCallID: call.ID,            // 对应的工具调用 ID
		}
		results = append(results, result)
	}

	return results, nil
}

// --- 辅助函数 ---

// evaluateSimpleExpression 执行简单的数学表达式计算
// 这是一个演示用的简化实现，仅支持加法运算
// 实际应用中应该使用更完整和安全的表达式解析库，如 govaluate
func evaluateSimpleExpression(expr string) float64 {
	// 移除表达式中的空格
	expr = strings.ReplaceAll(expr, " ", "")

	// 检查是否包含加法运算符
	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		if len(parts) == 2 {
			// 解析两个操作数并执行加法
			a := parseFloat(parts[0])
			b := parseFloat(parts[1])
			return a + b
		}
	}

	// 如果不是加法表达式，直接解析为数字
	return parseFloat(expr)
}

// parseFloat 将字符串解析为浮点数
// 使用 fmt.Sscanf 进行简单的数字解析
func parseFloat(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result) // 忽略错误，返回零值
	return result
}

// simulateTranslation 模拟文本翻译功能
// 使用预定义的翻译映射表来模拟翻译过程
// 实际应用中应该调用真实的翻译 API，如 Google Translate、百度翻译等
func simulateTranslation(text, fromLang, toLang string) string {
	// 预定义的翻译映射表
	// 键为原文，值为目标语言代码到翻译文本的映射
	translations := map[string]map[string]string{
		"Hello World": {
			"zh": "你好世界",             // 中文翻译
			"es": "Hola Mundo",       // 西班牙语翻译
			"fr": "Bonjour le Monde", // 法语翻译
		},
		"Good morning": {
			"zh": "早上好",
			"es": "Buenos días",
			"fr": "Bonjour",
		},
	}

	// 查找原文是否存在于翻译表中
	if langMap, exists := translations[text]; exists {
		// 查找目标语言的翻译
		if translation, exists := langMap[toLang]; exists {
			return translation
		}
	}

	// 如果没有找到对应的翻译，返回模拟的翻译格式
	return fmt.Sprintf("[模拟翻译: %s -> %s] %s", fromLang, toLang, text)
}

// simulateFileOperation 模拟文件系统操作
// 根据不同的操作类型返回相应的模拟结果
// 实际应用中应该执行真实的文件系统操作，并需要添加安全检查
func simulateFileOperation(action, path, content string) map[string]interface{} {
	switch action {
	case "list":
		// 模拟目录列表操作
		return map[string]interface{}{
			"action": "list",
			"path":   path,
			"files":  []string{"file1.txt", "file2.txt", "dir1/"}, // 模拟的文件列表
		}
	case "info":
		// 模拟文件信息查询
		return map[string]interface{}{
			"action":      "info",
			"path":        path,
			"exists":      true,                            // 模拟文件存在
			"size":        1024,                            // 模拟文件大小（字节）
			"modified":    time.Now().Format(time.RFC3339), // 模拟修改时间
			"permissions": "rw-r--r--",                     // 模拟文件权限
		}
	case "create":
		// 模拟文件创建操作
		return map[string]interface{}{
			"action":  "create",
			"path":    path,
			"success": true, // 模拟创建成功
			"message": fmt.Sprintf("文件 %s 创建成功", path),
		}
	default:
		// 不支持的操作类型
		return map[string]interface{}{
			"action": action,
			"path":   path,
			"error":  "不支持的操作",
		}
	}
}

// main 程序入口点，启动 ToolsNode 完整演示
func main() {
	demonstrateToolsNode()
}
