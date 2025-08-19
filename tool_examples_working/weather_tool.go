package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/schema"
)

// =============================================================================
//
//  文件: weather_tool.go
//  功能: 基于项目已有代码结构的可运行 Tool 示例
//  说明: 参考 tool_demo/main.go 的实现方式，创建实际可用的工具
//
// =============================================================================

// WeatherTool 天气查询工具
type WeatherTool struct{}

func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "查询指定城市的天气信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {Type: "string", Desc: "城市名称", Required: true},
			"date": {Type: "string", Desc: "查询日期", Required: false},
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
	result := map[string]interface{}{
		"city":        args.City,
		"date":        args.Date,
		"temperature": 25,
		"humidity":    60,
		"condition":   "晴朗",
		"description": fmt.Sprintf("%s今天天气晴朗，温度25°C", args.City),
	}
	
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// CalculatorTool 计算器工具
type CalculatorTool struct{}

func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行基本数学运算",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"operation": {Type: "string", Desc: "运算类型", Required: true, Enum: []string{"add", "subtract", "multiply", "divide"}},
			"a":         {Type: "number", Desc: "第一个数字", Required: true},
			"b":         {Type: "number", Desc: "第二个数字", Required: true},
		}),
	}, nil
}

func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		Operation string  `json:"operation"`
		A         float64 `json:"a"`
		B         float64 `json:"b"`
	}
	
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	
	log.Printf("[CalculatorTool] 执行运算: %s(%f, %f)", args.Operation, args.A, args.B)
	
	var result float64
	switch args.Operation {
	case "add":
		result = args.A + args.B
	case "subtract":
		result = args.A - args.B
	case "multiply":
		result = args.A * args.B
	case "divide":
		if args.B == 0 {
			return "", fmt.Errorf("除数不能为零")
		}
		result = args.A / args.B
	default:
		return "", fmt.Errorf("不支持的运算类型: %s", args.Operation)
	}
	
	response := map[string]interface{}{
		"operation": args.Operation,
		"a":         args.A,
		"b":         args.B,
		"result":    result,
	}
	
	responseBytes, _ := json.Marshal(response)
	return string(responseBytes), nil
}

// 模拟 ToolsNode（基于 tool_demo/main.go 的实现）
type SimpleToolsNode struct {
	tools map[string]Tool
}

type Tool interface {
	Info(ctx context.Context) (*schema.ToolInfo, error)
	InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error)
}

func NewSimpleToolsNode(tools []Tool) *SimpleToolsNode {
	toolMap := make(map[string]Tool)
	for _, t := range tools {
		info, err := t.Info(context.Background())
		if err != nil {
			log.Printf("警告: 获取工具 %T 的信息失败: %v", t, err)
			continue
		}
		toolMap[info.Name] = t
	}
	return &SimpleToolsNode{tools: toolMap}
}

func (n *SimpleToolsNode) Invoke(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
	if msg.Role != "assistant" {
		return nil, fmt.Errorf("输入消息的角色必须是 'assistant'")
	}
	if len(msg.ToolCalls) == 0 {
		return nil, nil
	}

	var results []*schema.Message
	for _, call := range msg.ToolCalls {
		toolToRun, ok := n.tools[call.Function.Name]
		if !ok {
			log.Printf("警告: 找不到名为 '%s' 的工具", call.Function.Name)
			continue
		}

		output, err := toolToRun.InvokableRun(ctx, call.Function.Arguments)
		if err != nil {
			log.Printf("警告: 执行工具 '%s' 失败: %v", call.Function.Name, err)
			output = fmt.Sprintf(`{"error": "%v"}`, err)
		}

		resultMsg := &schema.Message{
			Role:       "tool",
			Content:    output,
			Name:       call.Function.Name,
			ToolCallID: call.ID,
		}
		results = append(results, resultMsg)
	}
	return results, nil
}

func main() {
	ctx := context.Background()
	
	fmt.Println("=== 可运行的 Tool 演示 ===\n")
	
	// 创建工具实例
	weatherTool := &WeatherTool{}
	calculatorTool := &CalculatorTool{}
	
	// 测试单个工具
	fmt.Println("--- 测试天气工具 ---")
	weatherResult, err := weatherTool.InvokableRun(ctx, `{"city": "北京", "date": "2024-08-19"}`)
	if err != nil {
		log.Printf("天气工具执行失败: %v", err)
	} else {
		fmt.Printf("天气查询结果: %s\n\n", weatherResult)
	}
	
	fmt.Println("--- 测试计算器工具 ---")
	calcResult, err := calculatorTool.InvokableRun(ctx, `{"operation": "add", "a": 10, "b": 5}`)
	if err != nil {
		log.Printf("计算器工具执行失败: %v", err)
	} else {
		fmt.Printf("计算结果: %s\n\n", calcResult)
	}
	
	// 测试 ToolsNode
	fmt.Println("--- 测试 ToolsNode ---")
	toolsNode := NewSimpleToolsNode([]Tool{weatherTool, calculatorTool})
	
	// 模拟 LLM 的工具调用
	llmMessage := &schema.Message{
		Role: "assistant",
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_001",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"city": "上海"}`,
				},
			},
			{
				ID:   "call_002",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "calculator",
					Arguments: `{"operation": "multiply", "a": 8, "b": 7}`,
				},
			},
		},
	}
	
	results, err := toolsNode.Invoke(ctx, llmMessage)
	if err != nil {
		log.Printf("ToolsNode 执行失败: %v", err)
	} else {
		fmt.Printf("ToolsNode 执行成功，返回 %d 个结果:\n", len(results))
		for i, result := range results {
			fmt.Printf("  结果 %d: %s\n", i+1, result.Content)
		}
	}
}