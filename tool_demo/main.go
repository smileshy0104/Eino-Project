package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// =============================================================================
//
//  文件: tool_demo/main.go
//  功能: 演示如何自定义 Tool 并使用一个模拟的 ToolsNode 来执行工具调用。
//  说明: 由于 eino v0.4.3 版本中尚未包含文档描述的 compose.ToolsNode 和 compose.Tool,
//        我们在此手动模拟其核心功能以完成演示。
//
// =============================================================================

// --- 接口定义 (根据文档模拟) ---

// Option 是一个空的接口，用于模拟 compose.Option
type Option interface{}

// Tool 定义了工具的核心接口
type Tool interface {
	Info(ctx context.Context) (*schema.ToolInfo, error)
	InvokableRun(ctx context.Context, argumentsInJSON string, opts ...Option) (string, error)
}

// --- Tool 实现: WeatherTool ---

type WeatherTool struct{}

func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "查询指定城市在特定日期的天气信息。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {Type: "string", Desc: "城市名称", Required: true},
			"date": {Type: "string", Desc: "日期", Required: true},
		}),
	}, nil
}

func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...Option) (string, error) {
	var args struct {
		City string `json:"city"`
		Date string `json:"date"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	log.Printf("[WeatherTool] 正在查询 %s 在 %s 的天气...\n", args.City, args.Date)
	return fmt.Sprintf(`{"weather":"晴朗", "temperature":"28°C"}`), nil
}

// --- ToolsNode 模拟实现 ---

// MockToolsNode 模拟了 eino.compose.ToolsNode 的功能
type MockToolsNode struct {
	tools map[string]Tool
}

// NewMockToolsNode 创建并注册工具
func NewMockToolsNode(tools []Tool) *MockToolsNode {
	toolMap := make(map[string]Tool)
	for _, t := range tools {
		info, err := t.Info(context.Background())
		if err != nil {
			log.Printf("警告: 获取工具 %T 的信息失败: %v", t, err)
			continue
		}
		toolMap[info.Name] = t
	}
	return &MockToolsNode{tools: toolMap}
}

// Invoke 手动实现了工具的查找和执行逻辑
func (n *MockToolsNode) Invoke(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
	if msg.Role != "assistant" {
		return nil, fmt.Errorf("输入消息的角色必须是 'assistant'")
	}
	if len(msg.ToolCalls) == 0 {
		return nil, nil // 没有工具调用，直接返回
	}

	var results []*schema.Message
	for _, call := range msg.ToolCalls {
		toolToRun, ok := n.tools[call.Function.Name]
		if !ok {
			log.Printf("警告: 找不到名为 '%s' 的工具", call.Function.Name)
			continue
		}

		// 执行工具
		output, err := toolToRun.InvokableRun(ctx, call.Function.Arguments)
		if err != nil {
			log.Printf("警告: 执行工具 '%s' 失败: %v", call.Function.Name, err)
			output = fmt.Sprintf("{\"error\": \"%v\"}", err) // 将错误信息作为输出
		}

		// 将结果封装成 Tool Message
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

	weatherTool := &WeatherTool{}
	fmt.Println("--- 自定义工具已实例化 ---")

	// 使用我们模拟的 ToolsNode
	toolsNode := NewMockToolsNode([]Tool{weatherTool})
	fmt.Println("--- MockToolsNode 已创建并注册了 1 个工具 ---")

	// 模拟 LLM 的 Tool Calling 请求
	llmOutput := &schema.Message{
		Role: "assistant",
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_12345",
				Type: "function", // 直接使用字符串
				Function: schema.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"city": "上海", "date": "2025-08-18"}`,
				},
			},
		},
	}
	fmt.Println("\n--- 模拟的 LLM Tool Calling 请求 ---")

	// 调用我们模拟的 Invoke 方法
	toolResults, err := toolsNode.Invoke(ctx, llmOutput)
	if err != nil {
		log.Fatalf("ToolsNode 执行失败: %v", err)
	}

	fmt.Println("\n--- ToolsNode 执行完成，返回结果 ---")
	if len(toolResults) > 0 {
		resultMsg := toolResults[0]
		fmt.Printf("返回的消息数量: %d\n", len(toolResults))
		fmt.Printf("消息角色: %s\n", resultMsg.Role)
		fmt.Printf("Tool Call ID: %s\n", resultMsg.ToolCallID)
		fmt.Printf("工具执行结果 (Content): %s\n", resultMsg.Content)
	}
	fmt.Println(strings.Repeat("-", 30))
}
