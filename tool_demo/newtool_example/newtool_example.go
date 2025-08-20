// Package main 演示如何使用 NewTool 包装普通函数为 Eino Tool
package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// =============================================================================
//
//  文件: newtool_example.go
//  功能: 演示如何使用 NewTool 包装普通函数为 Eino Tool
//  说明: NewTool 提供了更简洁的方式将本地函数转换为工具
//
// =============================================================================

// --- 示例 1: 使用 NewTool 包装简单计算函数 ---

// AdditionRequest 加法运算的请求参数
type AdditionRequest struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

// AdditionResponse 加法运算的响应
type AdditionResponse struct {
	Result    float64 `json:"result"`
	Operation string  `json:"operation"`
	Timestamp string  `json:"timestamp"`
}

// addNumbers 执行加法运算的业务函数
// 参数:
//   - ctx: 上下文对象
//   - req: 包含两个数字的加法请求
// 返回:
//   - 包含运算结果、操作描述和时间戳的响应对象
//   - 错误信息（本例中总是返回 nil）
func addNumbers(ctx context.Context, req *AdditionRequest) (*AdditionResponse, error) {
	log.Printf("[AddNumbers] 执行加法: %f + %f", req.A, req.B)

	// 执行加法运算
	result := req.A + req.B

	// 构造并返回响应
	return &AdditionResponse{
		Result:    result,
		Operation: fmt.Sprintf("%.2f + %.2f", req.A, req.B),
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// --- 示例 2: 字符串处理函数 ---

type StringFormatRequest struct {
	Template string            `json:"template"`
	Values   map[string]string `json:"values"`
}

type StringFormatResponse struct {
	FormattedText    string `json:"formatted_text"`
	PlaceholderCount int    `json:"placeholder_count"`
}

// formatString 字符串模板格式化函数
// 参数:
//   - ctx: 上下文对象
//   - req: 包含模板字符串和替换值的请求
// 返回:
//   - 包含格式化后文本和替换占位符数量的响应对象
//   - 错误信息（本例中总是返回 nil）
func formatString(ctx context.Context, req *StringFormatRequest) (*StringFormatResponse, error) {
	log.Printf("[FormatString] 格式化模板: %s", req.Template)

	result := req.Template
	count := 0

	// 遍历所有替换值，查找并替换模板中的占位符
	for key, value := range req.Values {
		placeholder := fmt.Sprintf("{%s}", key)
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, value)
			count++
		}
	}

	return &StringFormatResponse{
		FormattedText:    result,
		PlaceholderCount: count,
	}, nil
}

// --- 示例 3: 数据验证函数 ---

type ValidationRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Age      int    `json:"age"`
	Username string `json:"username"`
}

type ValidationResponse struct {
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors"`
}

// validateUserData 用户数据验证函数
// 参数:
//   - ctx: 上下文对象
//   - req: 包含用户数据的验证请求
// 返回:
//   - 包含验证结果和错误信息的响应对象
//   - 错误信息（本例中总是返回 nil）
func validateUserData(ctx context.Context, req *ValidationRequest) (*ValidationResponse, error) {
	log.Printf("[ValidateUserData] 验证用户数据: %s", req.Username)

	var errors []string

	// 验证邮箱格式：必须包含 @ 和 . 符号
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		errors = append(errors, "邮箱格式不正确")
	}

	// 验证手机号长度：必须为11位
	if len(req.Phone) != 11 {
		errors = append(errors, "手机号长度不正确")
	}

	// 验证年龄范围：0-150岁
	if req.Age < 0 || req.Age > 150 {
		errors = append(errors, "年龄不在有效范围内")
	}

	// 验证用户名长度：3-20个字符
	if len(req.Username) < 3 || len(req.Username) > 20 {
		errors = append(errors, "用户名长度应在3-20个字符之间")
	}

	return &ValidationResponse{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}, nil
}

// --- 示例 4: 数据转换函数 ---

type ConversionRequest struct {
	Value    string `json:"value"`
	FromUnit string `json:"from_unit"`
	ToUnit   string `json:"to_unit"`
}

type ConversionResponse struct {
	OriginalValue  string  `json:"original_value"`
	ConvertedValue float64 `json:"converted_value"`
	FromUnit       string  `json:"from_unit"`
	ToUnit         string  `json:"to_unit"`
	Formula        string  `json:"formula"`
}

// convertUnits 单位转换函数（温度转换示例）
// 参数:
//   - ctx: 上下文对象
//   - req: 包含转换值和单位的请求
// 返回:
//   - 包含转换结果和转换公式的响应对象
//   - 错误信息（解析失败或不支持的转换时）
func convertUnits(ctx context.Context, req *ConversionRequest) (*ConversionResponse, error) {
	log.Printf("[ConvertUnits] 转换 %s from %s to %s", req.Value, req.FromUnit, req.ToUnit)

	// 解析输入的数值
	value, err := strconv.ParseFloat(req.Value, 64)
	if err != nil {
		return nil, fmt.Errorf("无法解析数值: %v", err)
	}

	var result float64
	var formula string

	// 简单的温度转换示例，支持摄氏度、华氏度、开尔文温度之间的转换
	switch {
	case req.FromUnit == "celsius" && req.ToUnit == "fahrenheit":
		result = value*9/5 + 32
		formula = "°F = °C × 9/5 + 32"
	case req.FromUnit == "fahrenheit" && req.ToUnit == "celsius":
		result = (value - 32) * 5 / 9
		formula = "°C = (°F - 32) × 5/9"
	case req.FromUnit == "celsius" && req.ToUnit == "kelvin":
		result = value + 273.15
		formula = "K = °C + 273.15"
	case req.FromUnit == "kelvin" && req.ToUnit == "celsius":
		result = value - 273.15
		formula = "°C = K - 273.15"
	default:
		return nil, fmt.Errorf("不支持从 %s 到 %s 的转换", req.FromUnit, req.ToUnit)
	}

	return &ConversionResponse{
		OriginalValue:  req.Value,
		ConvertedValue: result,
		FromUnit:       req.FromUnit,
		ToUnit:         req.ToUnit,
		Formula:        formula,
	}, nil
}

// --- 演示如何使用 NewTool 包装这些函数 ---

// demonstrateNewTool 演示如何使用 NewTool 包装普通函数为 Eino 工具
func demonstrateNewTool() {
	ctx := context.Background()

	fmt.Println("=== NewTool 包装函数演示 ===")

	// 1. 包装加法函数为工具
	// 定义工具的元信息：名称、描述、参数说明
	additionToolInfo := &schema.ToolInfo{
		Name: "add_numbers",
		Desc: "执行两个数字的加法运算",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"a": {Type: "number", Desc: "第一个数字", Required: true},
			"b": {Type: "number", Desc: "第二个数字", Required: true},
		}),
	}

	// 使用 NewTool 将普通函数包装为 Eino 工具
	additionTool := utils.NewTool(additionToolInfo, addNumbers)

	fmt.Println("--- 加法工具测试 ---")
	// 调用工具，传入 JSON 格式的参数
	addResult, err := additionTool.InvokableRun(ctx, `{"a": 15.5, "b": 24.3}`)
	if err != nil {
		log.Printf("加法工具执行失败: %v", err)
	} else {
		fmt.Printf("加法结果: %s\n\n", addResult)
	}

	// 2. 包装字符串格式化函数为工具
	stringToolInfo := &schema.ToolInfo{
		Name: "format_string",
		Desc: "使用提供的值格式化字符串模板",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"template": {Type: "string", Desc: "包含占位符的模板字符串", Required: true},
			"values":   {Type: "object", Desc: "替换值的键值对", Required: true},
		}),
	}

	stringTool := utils.NewTool(stringToolInfo, formatString)

	fmt.Println("--- 字符串格式化工具测试 ---")
	// 测试字符串模板格式化功能
	stringResult, err := stringTool.InvokableRun(ctx, `{
		"template": "Hello {name}, welcome to {city}!",
		"values": {
			"name": "Alice",
			"city": "Beijing"
		}
	}`)
	if err != nil {
		log.Printf("字符串工具执行失败: %v", err)
	} else {
		fmt.Printf("格式化结果: %s\n\n", stringResult)
	}

	// 3. 包装数据验证函数为工具
	validationToolInfo := &schema.ToolInfo{
		Name: "validate_user_data",
		Desc: "验证用户数据的有效性",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"email":    {Type: "string", Desc: "邮箱地址", Required: true},
			"phone":    {Type: "string", Desc: "手机号", Required: true},
			"age":      {Type: "integer", Desc: "年龄", Required: true},
			"username": {Type: "string", Desc: "用户名", Required: true},
		}),
	}

	validationTool := utils.NewTool(validationToolInfo, validateUserData)

	fmt.Println("--- 数据验证工具测试 ---")
	// 测试用户数据验证功能
	validationResult, err := validationTool.InvokableRun(ctx, `{
		"email": "user@example.com",
		"phone": "13812345678",
		"age": 25,
		"username": "alice"
	}`)
	if err != nil {
		log.Printf("验证工具执行失败: %v", err)
	} else {
		fmt.Printf("验证结果: %s\n\n", validationResult)
	}

	// 4. 包装单位转换函数为工具
	conversionToolInfo := &schema.ToolInfo{
		Name: "convert_temperature",
		Desc: "转换温度单位",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"value":     {Type: "string", Desc: "要转换的温度值", Required: true},
			"from_unit": {Type: "string", Desc: "原始单位", Required: true, Enum: []string{"celsius", "fahrenheit", "kelvin"}},
			"to_unit":   {Type: "string", Desc: "目标单位", Required: true, Enum: []string{"celsius", "fahrenheit", "kelvin"}},
		}),
	}

	conversionTool := utils.NewTool(conversionToolInfo, convertUnits)

	fmt.Println("--- 温度转换工具测试 ---")
	// 测试温度单位转换功能
	conversionResult, err := conversionTool.InvokableRun(ctx, `{
		"value": "25",
		"from_unit": "celsius",
		"to_unit": "fahrenheit"
	}`)
	if err != nil {
		log.Printf("转换工具执行失败: %v", err)
	} else {
		fmt.Printf("转换结果: %s\n\n", conversionResult)
	}
}

// main 函数：程序入口点，执行所有演示
func main() {
	demonstrateNewTool()
}
