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

type BaseTool interface {
	Info(ctx context.Context) (*schema.ToolInfo, error)
}

type InvokableTool interface {
	BaseTool
	InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error)
}

// --- 工具实现 ---

// WeatherTool 天气查询工具
type WeatherTool struct{}

func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "查询指定城市的天气信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {Type: "string", Desc: "城市名称", Required: true},
			"date": {Type: "string", Desc: "查询日期 (YYYY-MM-DD)", Required: false},
		}),
	}, nil
}

func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		City string `json:"city"`
		Date string `json:"date"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.Date == "" {
		args.Date = time.Now().Format("2006-01-02")
	}

	log.Printf("[WeatherTool] 查询 %s 在 %s 的天气", args.City, args.Date)

	// 模拟天气数据
	weatherData := map[string]interface{}{
		"city":        args.City,
		"date":        args.Date,
		"temperature": 25,
		"humidity":    60,
		"condition":   "晴朗",
		"wind_speed":  "5 km/h",
		"description": fmt.Sprintf("%s 今天天气晴朗，温度适宜", args.City),
	}

	result, _ := json.Marshal(weatherData)
	return string(result), nil
}

// CalculatorTool 计算器工具
type CalculatorTool struct{}

func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行数学计算",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {Type: "string", Desc: "数学表达式", Required: true},
		}),
	}, nil
}

func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[CalculatorTool] 计算表达式: %s", args.Expression)

	// 简单的计算器实现（实际应该使用更安全的表达式解析器）
	result := evaluateSimpleExpression(args.Expression)

	response := map[string]interface{}{
		"expression": args.Expression,
		"result":     result,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	resultBytes, _ := json.Marshal(response)
	return string(resultBytes), nil
}

// TranslatorTool 翻译工具
type TranslatorTool struct{}

func (t *TranslatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "translator",
		Desc: "翻译文本",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"text":      {Type: "string", Desc: "要翻译的文本", Required: true},
			"from_lang": {Type: "string", Desc: "源语言", Required: false},
			"to_lang":   {Type: "string", Desc: "目标语言", Required: true},
		}),
	}, nil
}

func (t *TranslatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		Text     string `json:"text"`
		FromLang string `json:"from_lang"`
		ToLang   string `json:"to_lang"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.FromLang == "" {
		args.FromLang = "auto"
	}

	log.Printf("[TranslatorTool] 翻译 '%s' 从 %s 到 %s", args.Text, args.FromLang, args.ToLang)

	// 模拟翻译
	translatedText := simulateTranslation(args.Text, args.FromLang, args.ToLang)

	response := map[string]interface{}{
		"original_text":   args.Text,
		"translated_text": translatedText,
		"from_language":   args.FromLang,
		"to_language":     args.ToLang,
		"confidence":      0.95,
	}

	result, _ := json.Marshal(response)
	return string(result), nil
}

// FileManagerTool 文件管理工具
type FileManagerTool struct{}

func (f *FileManagerTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_manager",
		Desc: "管理文件和目录",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"action": {
				Type:     "string",
				Desc:     "操作类型",
				Required: true,
				Enum:     []string{"list", "create", "read", "delete", "info"},
			},
			"path":    {Type: "string", Desc: "文件路径", Required: true},
			"content": {Type: "string", Desc: "文件内容（创建文件时使用）", Required: false},
		}),
	}, nil
}

func (f *FileManagerTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		Action  string `json:"action"`
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[FileManagerTool] 执行 %s 操作，路径: %s", args.Action, args.Path)

	// 模拟文件操作
	result := simulateFileOperation(args.Action, args.Path, args.Content)

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// --- ToolsNode 演示 ---

func demonstrateToolsNode() {
	ctx := context.Background()

	fmt.Println("=== ToolsNode 完整使用演示 ===\n")

	// 1. 创建工具实例
	weatherTool := &WeatherTool{}
	calculatorTool := &CalculatorTool{}
	translatorTool := &TranslatorTool{}
	fileManagerTool := &FileManagerTool{}

	// 2. 创建工具列表
	tools := []InvokableTool{
		weatherTool,
		calculatorTool,
		translatorTool,
		fileManagerTool,
	}

	// 3. 创建 ToolsNode (注意：这里使用模拟的 ToolsNode，实际使用时应该用 compose.NewToolsNode)
	toolsNode := NewMockToolsNode(tools)

	fmt.Printf("已创建 ToolsNode，注册了 %d 个工具:\n", len(tools))
	for _, tool := range tools {
		info, _ := tool.Info(ctx)
		fmt.Printf("  - %s: %s\n", info.Name, info.Desc)
	}
	fmt.Println()

	// 4. 演示单个工具调用
	fmt.Println("--- 演示单个工具调用 ---")

	// 模拟 LLM 想要调用天气工具
	weatherMessage := &schema.Message{
		Role: "assistant",
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_weather_001",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"city": "北京", "date": "2024-08-19"}`,
				},
			},
		},
	}

	weatherResults, err := toolsNode.Invoke(ctx, weatherMessage)
	if err != nil {
		log.Printf("天气工具调用失败: %v", err)
	} else {
		fmt.Printf("天气查询结果: %s\n\n", weatherResults[0].Content)
	}

	// 5. 演示多工具并行调用
	fmt.Println("--- 演示多工具并行调用 ---")

	multiToolMessage := &schema.Message{
		Role: "assistant",
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_calc_001",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "calculator",
					Arguments: `{"expression": "25 + 17"}`,
				},
			},
			{
				ID:   "call_translate_001",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "translator",
					Arguments: `{"text": "Hello World", "to_lang": "zh"}`,
				},
			},
			{
				ID:   "call_file_001",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "file_manager",
					Arguments: `{"action": "info", "path": "/tmp/example.txt"}`,
				},
			},
		},
	}

	multiResults, err := toolsNode.Invoke(ctx, multiToolMessage)
	if err != nil {
		log.Printf("多工具调用失败: %v", err)
	} else {
		fmt.Printf("并行调用了 %d 个工具:\n", len(multiResults))
		for i, result := range multiResults {
			fmt.Printf("  工具 %d - %s: %s\n", i+1, result.Name, result.Content)
		}
		fmt.Println()
	}

	// 6. 演示在 Chain 中使用 ToolsNode
	fmt.Println("--- 演示在 Chain 中使用 ToolsNode ---")
	demonstrateToolsNodeInChain(toolsNode)

	// 7. 演示错误处理
	fmt.Println("--- 演示错误处理 ---")

	errorMessage := &schema.Message{
		Role: "assistant",
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_error_001",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "nonexistent_tool",
					Arguments: `{"param": "value"}`,
				},
			},
		},
	}

	errorResults, err := toolsNode.Invoke(ctx, errorMessage)
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	} else if len(errorResults) == 0 {
		fmt.Println("未找到对应工具，跳过执行")
	}
}

// --- Chain 中使用 ToolsNode 的演示 ---

func demonstrateToolsNodeInChain(toolsNode *MockToolsNode) {
	ctx := context.Background()

	// 注意：这是一个模拟实现，实际使用时应该用 Eino 的 compose.NewChain()
	fmt.Println("模拟 Chain 工作流:")
	fmt.Println("  1. 用户消息 -> LLM 生成")
	fmt.Println("  2. LLM 生成 -> ToolsNode 执行")
	fmt.Println("  3. 工具结果 -> LLM 最终回复")

	// 模拟 LLM 决定调用工具
	llmDecision := &schema.Message{
		Role:    "assistant",
		Content: "我需要查询天气和进行一些计算来回答用户的问题。",
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_weather_chain",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"city": "上海"}`,
				},
			},
			{
				ID:   "call_calc_chain",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "calculator",
					Arguments: `{"expression": "20 + 5"}`,
				},
			},
		},
	}

	// 执行工具调用
	toolResults, err := toolsNode.Invoke(ctx, llmDecision)
	if err != nil {
		log.Printf("Chain 中工具调用失败: %v", err)
		return
	}

	fmt.Printf("Chain 执行完成，工具调用结果:\n")
	for _, result := range toolResults {
		fmt.Printf("  - %s: %s\n", result.Name, result.Content)
	}

	// 模拟将工具结果传递给 LLM 生成最终回复
	fmt.Println("  → 将工具结果传递给 LLM 生成最终用户回复")
	fmt.Println()
}

// --- 模拟的 ToolsNode 实现 ---

type MockToolsNode struct {
	tools map[string]InvokableTool
}

func NewMockToolsNode(tools []InvokableTool) *MockToolsNode {
	toolMap := make(map[string]InvokableTool)
	for _, tool := range tools {
		info, err := tool.Info(context.Background())
		if err != nil {
			log.Printf("获取工具信息失败: %v", err)
			continue
		}
		toolMap[info.Name] = tool
	}
	return &MockToolsNode{tools: toolMap}
}

func (n *MockToolsNode) Invoke(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
	if msg.Role != "assistant" || len(msg.ToolCalls) == 0 {
		return nil, nil
	}

	var results []*schema.Message

	for _, call := range msg.ToolCalls {
		tool, exists := n.tools[call.Function.Name]
		if !exists {
			log.Printf("工具 '%s' 不存在", call.Function.Name)
			continue
		}

		output, err := tool.InvokableRun(ctx, call.Function.Arguments)
		if err != nil {
			log.Printf("工具 '%s' 执行失败: %v", call.Function.Name, err)
			output = fmt.Sprintf(`{"error": "%s"}`, err.Error())
		}

		result := &schema.Message{
			Role:       "tool",
			Content:    output,
			Name:       call.Function.Name,
			ToolCallID: call.ID,
		}
		results = append(results, result)
	}

	return results, nil
}

// --- 辅助函数 ---

func evaluateSimpleExpression(expr string) float64 {
	// 简单的表达式计算（实际应该使用更安全的解析器）
	expr = strings.ReplaceAll(expr, " ", "")

	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		if len(parts) == 2 {
			a := parseFloat(parts[0])
			b := parseFloat(parts[1])
			return a + b
		}
	}

	return parseFloat(expr)
}

func parseFloat(s string) float64 {
	// 简单的数字解析
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

func simulateTranslation(text, fromLang, toLang string) string {
	// 简单的翻译模拟
	translations := map[string]map[string]string{
		"Hello World": {
			"zh": "你好世界",
			"es": "Hola Mundo",
			"fr": "Bonjour le Monde",
		},
		"Good morning": {
			"zh": "早上好",
			"es": "Buenos días",
			"fr": "Bonjour",
		},
	}

	if langMap, exists := translations[text]; exists {
		if translation, exists := langMap[toLang]; exists {
			return translation
		}
	}

	return fmt.Sprintf("[模拟翻译: %s -> %s] %s", fromLang, toLang, text)
}

func simulateFileOperation(action, path, content string) map[string]interface{} {
	switch action {
	case "list":
		return map[string]interface{}{
			"action": "list",
			"path":   path,
			"files":  []string{"file1.txt", "file2.txt", "dir1/"},
		}
	case "info":
		return map[string]interface{}{
			"action":      "info",
			"path":        path,
			"exists":      true,
			"size":        1024,
			"modified":    time.Now().Format(time.RFC3339),
			"permissions": "rw-r--r--",
		}
	case "create":
		return map[string]interface{}{
			"action":  "create",
			"path":    path,
			"success": true,
			"message": fmt.Sprintf("文件 %s 创建成功", path),
		}
	default:
		return map[string]interface{}{
			"action": action,
			"path":   path,
			"error":  "不支持的操作",
		}
	}
}

func main() {
	demonstrateToolsNode()
}
