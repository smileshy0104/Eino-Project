package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/cloudwego/eino/schema"
)

// =============================================================================
//
//  文件: basic_tool.go
//  功能: 演示如何手动实现 Eino Tool 接口的基本示例
//  说明: 这是最基础的 Tool 实现方式，需要手动处理所有细节
//
// =============================================================================

// --- 示例 1: 计算器 Tool (手动实现接口) ---

type CalculatorTool struct{}

// Info 返回工具的基本信息
func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行基本数学运算，支持加减乘除",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"operation": {
				Type:     "string",
				Desc:     "运算类型",
				Required: true,
				Enum:     []string{"add", "subtract", "multiply", "divide"},
			},
			"a": {
				Type:     "number",
				Desc:     "第一个数字",
				Required: true,
			},
			"b": {
				Type:     "number",
				Desc:     "第二个数字",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具的具体逻辑
func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 解析 JSON 参数
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
	var err error
	
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
	
	// 返回结果为 JSON 格式
	response := map[string]interface{}{
		"result":    result,
		"operation": args.Operation,
		"operands":  []float64{args.A, args.B},
	}
	
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("结果序列化失败: %v", err)
	}
	
	return string(responseBytes), nil
}

// --- 示例 2: 文本处理 Tool ---

type TextProcessorTool struct{}

func (t *TextProcessorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "text_processor",
		Desc: "处理文本，支持转换大小写、计算长度、反转字符串",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"action": {
				Type:     "string",
				Desc:     "处理动作",
				Required: true,
				Enum:     []string{"uppercase", "lowercase", "length", "reverse"},
			},
			"text": {
				Type:     "string",
				Desc:     "要处理的文本",
				Required: true,
			},
		}),
	}, nil
}

func (t *TextProcessorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		Action string `json:"action"`
		Text   string `json:"text"`
	}
	
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	
	log.Printf("[TextProcessorTool] 处理文本: %s('%s')", args.Action, args.Text)
	
	var result interface{}
	
	switch args.Action {
	case "uppercase":
		result = map[string]string{
			"original": args.Text,
			"result":   fmt.Sprintf("%s", args.Text),
		}
	case "lowercase":
		result = map[string]string{
			"original": args.Text,
			"result":   fmt.Sprintf("%s", args.Text),
		}
	case "length":
		result = map[string]interface{}{
			"text":   args.Text,
			"length": len(args.Text),
		}
	case "reverse":
		runes := []rune(args.Text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = map[string]string{
			"original": args.Text,
			"result":   string(runes),
		}
	default:
		return "", fmt.Errorf("不支持的动作: %s", args.Action)
	}
	
	responseBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("结果序列化失败: %v", err)
	}
	
	return string(responseBytes), nil
}

// --- 示例 3: 数学函数 Tool ---

type MathTool struct{}

func (m *MathTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "math_functions",
		Desc: "执行高级数学函数，如sin、cos、sqrt、log等",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"function": {
				Type:     "string",
				Desc:     "数学函数名称",
				Required: true,
				Enum:     []string{"sin", "cos", "sqrt", "log", "exp", "abs"},
			},
			"value": {
				Type:     "number",
				Desc:     "输入值",
				Required: true,
			},
		}),
	}, nil
}

func (m *MathTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		Function string  `json:"function"`
		Value    float64 `json:"value"`
	}
	
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	
	log.Printf("[MathTool] 执行函数: %s(%f)", args.Function, args.Value)
	
	var result float64
	var err error
	
	switch args.Function {
	case "sin":
		result = math.Sin(args.Value)
	case "cos":
		result = math.Cos(args.Value)
	case "sqrt":
		if args.Value < 0 {
			return "", fmt.Errorf("不能计算负数的平方根")
		}
		result = math.Sqrt(args.Value)
	case "log":
		if args.Value <= 0 {
			return "", fmt.Errorf("对数函数的输入必须大于0")
		}
		result = math.Log(args.Value)
	case "exp":
		result = math.Exp(args.Value)
	case "abs":
		result = math.Abs(args.Value)
	default:
		return "", fmt.Errorf("不支持的函数: %s", args.Function)
	}
	
	response := map[string]interface{}{
		"function": args.Function,
		"input":    args.Value,
		"result":   result,
	}
	
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("结果序列化失败: %v", err)
	}
	
	return string(responseBytes), nil
}

// --- 演示函数 ---

func demonstrateBasicTools() {
	ctx := context.Background()
	
	// 创建工具实例
	calculator := &CalculatorTool{}
	textProcessor := &TextProcessorTool{}
	mathTool := &MathTool{}
	
	fmt.Println("=== 基本 Tool 实现演示 ===\n")
	
	// 演示计算器工具
	fmt.Println("--- 计算器工具 ---")
	calcInfo, _ := calculator.Info(ctx)
	fmt.Printf("工具名称: %s\n", calcInfo.Name)
	fmt.Printf("工具描述: %s\n", calcInfo.Desc)
	
	calcResult, err := calculator.InvokableRun(ctx, `{"operation":"add","a":10,"b":5}`)
	if err != nil {
		log.Printf("计算器执行失败: %v", err)
	} else {
		fmt.Printf("计算结果: %s\n\n", calcResult)
	}
	
	// 演示文本处理工具
	fmt.Println("--- 文本处理工具 ---")
	textInfo, _ := textProcessor.Info(ctx)
	fmt.Printf("工具名称: %s\n", textInfo.Name)
	fmt.Printf("工具描述: %s\n", textInfo.Desc)
	
	textResult, err := textProcessor.InvokableRun(ctx, `{"action":"reverse","text":"Hello World"}`)
	if err != nil {
		log.Printf("文本处理执行失败: %v", err)
	} else {
		fmt.Printf("处理结果: %s\n\n", textResult)
	}
	
	// 演示数学函数工具
	fmt.Println("--- 数学函数工具 ---")
	mathInfo, _ := mathTool.Info(ctx)
	fmt.Printf("工具名称: %s\n", mathInfo.Name)
	fmt.Printf("工具描述: %s\n", mathInfo.Desc)
	
	mathResult, err := mathTool.InvokableRun(ctx, `{"function":"sqrt","value":16}`)
	if err != nil {
		log.Printf("数学函数执行失败: %v", err)
	} else {
		fmt.Printf("计算结果: %s\n\n", mathResult)
	}
}

func main() {
	demonstrateBasicTools()
}