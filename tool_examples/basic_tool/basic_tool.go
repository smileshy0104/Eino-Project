// package main 表明这是一个可执行程序
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
//  说明: 这是最基础的 Tool 实现方式，需要手动处理所有细节，包括参数的JSON解析和结果的JSON序列化。
//
// =============================================================================

// --- 示例 1: 计算器 Tool (手动实现接口) ---

// CalculatorTool 定义了一个简单的计算器工具结构体。
// 它通过实现 Eino 的 Tool 接口来提供加、减、乘、除功能。
type CalculatorTool struct{}

// Info 返回工具的元数据，包括名称、描述和参数定义。
// Eino 框架使用这些信息来理解工具的功能以及如何调用它。
func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",      // 工具的唯一名称
		Desc: "执行基本数学运算，支持加减乘除", // 工具功能的简要描述
		// 定义工具接受的参数
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"operation": { // 参数：运算类型
				Type:     "string",
				Desc:     "运算类型",
				Required: true,
				Enum:     []string{"add", "subtract", "multiply", "divide"}, // 枚举值，限制输入
			},
			"a": { // 参数：第一个操作数
				Type:     "number",
				Desc:     "第一个数字",
				Required: true,
			},
			"b": { // 参数：第二个操作数
				Type:     "number",
				Desc:     "第二个数字",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 是工具的核心执行逻辑。
// 它接收一个包含参数的 JSON 字符串，执行相应的计算，并返回一个包含结果的 JSON 字符串。
func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 定义一个匿名结构体来解析输入的 JSON 参数
	var args struct {
		Operation string  `json:"operation"`
		A         float64 `json:"a"`
		B         float64 `json:"b"`
	}

	// 将输入的 JSON 字符串反序列化到结构体中
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 打印日志，方便调试
	log.Printf("[CalculatorTool] 执行运算: %s(%f, %f)", args.Operation, args.A, args.B)

	var result float64
	var err error

	// 根据 'operation' 参数执行不同的计算
	switch args.Operation {
	case "add":
		result = args.A + args.B
	case "subtract":
		result = args.A - args.B
	case "multiply":
		result = args.A * args.B
	case "divide":
		// 处理除数为零的边界情况
		if args.B == 0 {
			return "", fmt.Errorf("除数不能为零")
		}
		result = args.A / args.B
	default:
		// 如果操作类型无效，则返回错误
		return "", fmt.Errorf("不支持的运算类型: %s", args.Operation)
	}

	// 构建一个 map 来存储响应结果
	response := map[string]interface{}{
		"result":    result,
		"operation": args.Operation,
		"operands":  []float64{args.A, args.B},
	}

	// 将响应结果序列化为 JSON 字节数组
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("结果序列化失败: %v", err)
	}

	// 将 JSON 字节数组转换为字符串并返回
	return string(responseBytes), nil
}

// --- 示例 2: 文本处理 Tool ---

// TextProcessorTool 定义了一个用于处理文本的工具。
type TextProcessorTool struct{}

// Info 返回文本处理工具的元数据。
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

// InvokableRun 执行文本处理逻辑。
func (t *TextProcessorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 定义用于解析参数的结构体
	var args struct {
		Action string `json:"action"`
		Text   string `json:"text"`
	}

	// 解析 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[TextProcessorTool] 处理文本: %s('%s')", args.Action, args.Text)

	var result interface{} // 使用 interface{} 因为不同操作的返回结构不同

	// 根据 'action' 参数执行不同的文本处理
	switch args.Action {
	case "uppercase":
		result = map[string]string{
			"original": args.Text,
			"result":   fmt.Sprintf("%s", args.Text), // 转换为大写
		}
	case "lowercase":
		result = map[string]string{
			"original": args.Text,
			"result":   fmt.Sprintf("%s", args.Text), // 转换为小写
		}
	case "length":
		result = map[string]interface{}{
			"text":   args.Text,
			"length": len(args.Text), // 计算字符串长度
		}
	case "reverse":
		// 反转字符串（支持 Unicode）
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

	// 将结果序列化为 JSON 并返回
	responseBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("结果序列化失败: %v", err)
	}

	return string(responseBytes), nil
}

// --- 示例 3: 数学函数 Tool ---

// MathTool 定义了一个用于执行高级数学函数的工具。
type MathTool struct{}

// Info 返回数学函数工具的元数据。
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

// InvokableRun 执行数学函数计算。
func (m *MathTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 定义用于解析参数的结构体
	var args struct {
		Function string  `json:"function"`
		Value    float64 `json:"value"`
	}

	// 解析 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[MathTool] 执行函数: %s(%f)", args.Function, args.Value)

	var result float64
	var err error

	// 根据 'function' 参数调用相应的 math 包函数
	switch args.Function {
	case "sin":
		result = math.Sin(args.Value)
	case "cos":
		result = math.Cos(args.Value)
	case "sqrt":
		// 检查负数平方根的无效输入
		if args.Value < 0 {
			return "", fmt.Errorf("不能计算负数的平方根")
		}
		result = math.Sqrt(args.Value)
	case "log":
		// 检查对数函数的无效输入
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

	// 构建响应 map
	response := map[string]interface{}{
		"function": args.Function,
		"input":    args.Value,
		"result":   result,
	}

	// 序列化并返回结果
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("结果序列化失败: %v", err)
	}

	return string(responseBytes), nil
}

// --- 演示函数 ---

// demonstrateBasicTools 是一个演示函数，用于展示如何直接调用这些基础工具。
// 它模拟了 Eino 框架调用工具的过程，帮助理解工具的独立功能。
func demonstrateBasicTools() {
	ctx := context.Background()

	// 1. 创建所有工具的实例
	calculator := &CalculatorTool{}
	textProcessor := &TextProcessorTool{}
	mathTool := &MathTool{}

	fmt.Println("=== 基本 Tool 实现演示 ===")

	// 2. 演示计算器工具
	fmt.Println("\n--- 计算器工具 ---")
	// 获取并打印工具信息
	calcInfo, _ := calculator.Info(ctx)
	fmt.Printf("工具名称: %s\n", calcInfo.Name)
	fmt.Printf("工具描述: %s\n", calcInfo.Desc)

	// 调用工具的执行方法，并传入 JSON 格式的参数
	calcResult, err := calculator.InvokableRun(ctx, `{"operation":"add","a":10,"b":5}`)
	if err != nil {
		log.Printf("计算器执行失败: %v", err)
	} else {
		fmt.Printf("计算结果: %s\n", calcResult)
	}

	// 3. 演示文本处理工具
	fmt.Println("\n--- 文本处理工具 ---")
	textInfo, _ := textProcessor.Info(ctx)
	fmt.Printf("工具名称: %s\n", textInfo.Name)
	fmt.Printf("工具描述: %s\n", textInfo.Desc)

	// 调用并传入 "reverse" 动作
	textResult, err := textProcessor.InvokableRun(ctx, `{"action":"reverse","text":"Hello World"}`)
	if err != nil {
		log.Printf("文本处理执行失败: %v", err)
	} else {
		fmt.Printf("处理结果: %s\n", textResult)
	}

	// 4. 演示数学函数工具
	fmt.Println("\n--- 数学函数工具 ---")
	mathInfo, _ := mathTool.Info(ctx)
	fmt.Printf("工具名称: %s\n", mathInfo.Name)
	fmt.Printf("工具描述: %s\n", mathInfo.Desc)

	// 调用并计算 16 的平方根
	mathResult, err := mathTool.InvokableRun(ctx, `{"function":"sqrt","value":16}`)
	if err != nil {
		log.Printf("数学函数执行失败: %v", err)
	} else {
		fmt.Printf("计算结果: %s\n", mathResult)
	}
}

// main 是程序的入口点。
func main() {
	// 调用演示函数来执行所有工具的示例
	demonstrateBasicTools()
}
