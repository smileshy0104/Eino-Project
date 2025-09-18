// Package main 演示 Eino 框架中 ToolsNode 组件的各种用法
// ToolsNode 是用于扩展大语言模型能力的核心组件
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// 配置结构
type Config struct {
	ARKAPIKey     string `mapstructure:"ARK_API_KEY"`
	LLMModel      string `mapstructure:"LLM_MODEL"`
	WeatherAPIKey string `mapstructure:"WEATHER_API_KEY"`
	SearchAPIKey  string `mapstructure:"SEARCH_API_KEY"`
}

// 全局配置
var config Config

// ========== 工具实现部分 ==========

// 1. 计算器工具结构体
type CalculatorTool struct{}

// 实现 BaseTool 接口
func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行数学计算和表达式求值",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {
				Type:     "string",
				Desc:     "数学表达式，支持 +, -, *, / 运算",
				Required: true,
			},
		}),
	}, nil
}

// 实现 InvokableTool 接口
func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	fmt.Printf("🧮 执行计算: %s\n", argumentsInJSON)

	// 解析输入参数
	var args struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Expression == "" {
		return "", errors.New("表达式不能为空")
	}

	// 简单的数学表达式解析和计算
	result, err := evaluateExpression(args.Expression)
	if err != nil {
		return "", fmt.Errorf("计算错误: %w", err)
	}

	response := fmt.Sprintf("计算结果: %s = %.2f", args.Expression, result)
	fmt.Printf("✅ 计算完成: %s\n", response)
	return response, nil
}

// 简单的表达式计算器（支持基本四则运算）
func evaluateExpression(expr string) (float64, error) {
	expr = strings.ReplaceAll(expr, " ", "")

	// 简化版本，仅支持单个运算
	for _, op := range []string{"+", "-", "*", "/"} {
		if idx := strings.Index(expr, op); idx > 0 {
			left, err := strconv.ParseFloat(expr[:idx], 64)
			if err != nil {
				continue
			}
			right, err := strconv.ParseFloat(expr[idx+1:], 64)
			if err != nil {
				continue
			}

			switch op {
			case "+":
				return left + right, nil
			case "-":
				return left - right, nil
			case "*":
				return left * right, nil
			case "/":
				if right == 0 {
					return 0, errors.New("除零错误")
				}
				return left / right, nil
			}
		}
	}

	// 如果没有运算符，尝试解析为单个数字
	return strconv.ParseFloat(expr, 64)
}

// 2. 天气查询工具结构体
type WeatherTool struct {
	apiKey string
}

// 实现 BaseTool 接口
func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "获取指定城市的当前天气信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",
				Desc:     "城市名称",
				Required: true,
			},
			"units": {
				Type:     "string",
				Desc:     "温度单位(celsius/fahrenheit)",
				Required: false,
			},
		}),
	}, nil
}

// 实现 InvokableTool 接口
func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	fmt.Printf("🌤️ 查询天气: %s\n", argumentsInJSON)

	// 解析输入参数
	var args struct {
		City  string `json:"city"`
		Units string `json:"units"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.City == "" {
		return "", errors.New("城市名称不能为空")
	}

	if args.Units == "" {
		args.Units = "celsius"
	}

	// 模拟天气查询（实际应用中应调用真实的天气API）
	weather := w.simulateWeatherAPI(args.City, args.Units)

	response := fmt.Sprintf("城市: %s\n温度: %s\n天气: %s\n湿度: %s\n风速: %s",
		weather.City, weather.Temperature, weather.Description, weather.Humidity, weather.WindSpeed)

	fmt.Printf("✅ 天气查询完成: %s\n", args.City)
	return response, nil
}

// 天气信息结构
type WeatherInfo struct {
	City        string
	Temperature string
	Description string
	Humidity    string
	WindSpeed   string
}

// 模拟天气API调用
func (w *WeatherTool) simulateWeatherAPI(city, units string) *WeatherInfo {
	// 模拟不同城市的天气数据
	weatherData := map[string]*WeatherInfo{
		"北京": {
			City:        "北京",
			Temperature: "22°C",
			Description: "晴朗",
			Humidity:    "45%",
			WindSpeed:   "3.2 m/s",
		},
		"上海": {
			City:        "上海",
			Temperature: "26°C",
			Description: "多云",
			Humidity:    "65%",
			WindSpeed:   "2.8 m/s",
		},
		"深圳": {
			City:        "深圳",
			Temperature: "28°C",
			Description: "小雨",
			Humidity:    "75%",
			WindSpeed:   "4.1 m/s",
		},
	}

	// 转换温度单位
	if weather, exists := weatherData[city]; exists {
		if units == "fahrenheit" {
			// 简单转换示例
			weather.Temperature = strings.Replace(weather.Temperature, "°C", "°F", 1)
		}
		return weather
	}

	// 默认天气数据
	temp := "20°C"
	if units == "fahrenheit" {
		temp = "68°F"
	}

	return &WeatherInfo{
		City:        city,
		Temperature: temp,
		Description: "晴朗",
		Humidity:    "50%",
		WindSpeed:   "3.0 m/s",
	}
}

// 3. 时间工具结构体
type TimeTool struct{}

// 实现 BaseTool 接口
func (t *TimeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_time",
		Desc: "获取指定时区的当前时间",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"timezone": {
				Type:     "string",
				Desc:     "时区名称，如 Asia/Shanghai, UTC, local",
				Required: false,
			},
		}),
	}, nil
}

// 实现 InvokableTool 接口
func (t *TimeTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	fmt.Printf("🕐 获取时间: %s\n", argumentsInJSON)

	// 解析输入参数
	var args struct {
		Timezone string `json:"timezone"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	var loc *time.Location
	var err error

	if args.Timezone == "" || args.Timezone == "local" {
		loc = time.Local
	} else {
		loc, err = time.LoadLocation(args.Timezone)
		if err != nil {
			return "", fmt.Errorf("无效的时区: %w", err)
		}
	}

	now := time.Now().In(loc)
	response := fmt.Sprintf("当前时间: %s\n时区: %s\n格式化时间: %s",
		now.Format("2006-01-02 15:04:05"),
		loc.String(),
		now.Format("2006年01月02日 15时04分05秒"))

	fmt.Printf("✅ 时间获取完成\n")
	return response, nil
}

// ========== 演示函数部分 ==========

// 初始化配置
func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("警告: 无法读取配置文件: %v", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("配置解析失败: %w", err)
	}

	// 设置默认值
	if config.LLMModel == "" {
		config.LLMModel = "doubao-seed-1-6-250615"
	}

	return nil
}

// 1. 基础工具创建演示
func basicToolCreationDemo(ctx context.Context) {
	fmt.Println("\n🎯 基础工具创建演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建工具实例
	fmt.Println("\n  📝 创建工具实例:")
	calculatorTool := &CalculatorTool{}
	timeTool := &TimeTool{}

	// 测试工具信息
	fmt.Println("\n  📋 工具信息:")
	calcInfo, err := calculatorTool.Info(ctx)
	if err != nil {
		log.Printf("获取计算器工具信息失败: %v", err)
		return
	}
	fmt.Printf("    工具名称: %s\n", calcInfo.Name)
	fmt.Printf("    工具描述: %s\n", calcInfo.Desc)

	timeInfo, err := timeTool.Info(ctx)
	if err != nil {
		log.Printf("获取时间工具信息失败: %v", err)
		return
	}
	fmt.Printf("    工具名称: %s\n", timeInfo.Name)
	fmt.Printf("    工具描述: %s\n", timeInfo.Desc)

	// 测试工具调用
	fmt.Println("\n  🧮 测试计算器工具:")
	calcInput := `{"expression": "10 + 5"}`
	result, err := calculatorTool.InvokableRun(ctx, calcInput)
	if err != nil {
		log.Printf("工具调用失败: %v", err)
		return
	}
	fmt.Printf("    调用结果: %s\n", result)

	// 测试时间工具
	fmt.Println("\n  🕐 测试时间工具:")
	timeInput := `{"timezone": "Asia/Shanghai"}`
	timeResult, err := timeTool.InvokableRun(ctx, timeInput)
	if err != nil {
		log.Printf("时间工具调用失败: %v", err)
		return
	}
	fmt.Printf("    时间信息:\n%s\n", indentLines(timeResult, "    "))
}

// 2. 手动工具创建演示
func manualToolCreationDemo(ctx context.Context) {
	fmt.Println("\n🔧 手动工具创建演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建天气工具
	fmt.Println("\n  📝 创建天气查询工具:")
	weatherTool := &WeatherTool{
		apiKey: config.WeatherAPIKey,
	}

	// 显示工具信息
	weatherInfo, err := weatherTool.Info(ctx)
	if err != nil {
		log.Printf("获取天气工具信息失败: %v", err)
		return
	}
	fmt.Printf("    工具名称: %s\n", weatherInfo.Name)
	fmt.Printf("    工具描述: %s\n", weatherInfo.Desc)

	// 测试天气工具
	fmt.Println("\n  🌤️ 测试天气查询工具:")
	weatherInput := `{"city": "北京", "units": "celsius"}`
	weatherResult, err := weatherTool.InvokableRun(ctx, weatherInput)
	if err != nil {
		log.Printf("天气工具调用失败: %v", err)
		return
	}
	fmt.Printf("    天气信息:\n%s\n", indentLines(weatherResult, "    "))
}

// 3. ToolsNode 配置演示
func toolsNodeConfigDemo(ctx context.Context) {
	fmt.Println("\n🏗️ ToolsNode 配置演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建所有工具
	fmt.Println("\n  📝 准备工具集合:")

	// 创建工具实例
	calculatorTool := &CalculatorTool{}
	timeTool := &TimeTool{}
	weatherTool := &WeatherTool{apiKey: config.WeatherAPIKey}

	// 收集所有工具
	tools := []tool.BaseTool{
		calculatorTool,
		timeTool,
		weatherTool,
	}

	fmt.Printf("    总计创建 %d 个工具:\n", len(tools))
	for i, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Printf("获取工具 %d 信息失败: %v", i+1, err)
			continue
		}
		fmt.Printf("      %d. %s - %s\n", i+1, info.Name, info.Desc)
	}

	fmt.Println("\n    ✅ 工具集合创建成功")
	fmt.Printf("    工具数量: %d\n", len(tools))
}

// 4. 工具调用链演示
func toolChainDemo(ctx context.Context) {
	fmt.Println("\n🔗 工具调用链演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建工具
	calculatorTool := &CalculatorTool{}
	weatherTool := &WeatherTool{apiKey: config.WeatherAPIKey}
	timeTool := &TimeTool{}

	// 模拟复杂的工具调用序列
	fmt.Println("\n  📝 模拟复杂任务: 计算 -> 天气查询 -> 时间查询")

	// 第一步：执行计算
	fmt.Println("\n    步骤1: 数学计算")
	calcInput := `{"expression": "25 * 4"}`
	calcResult, err := calculatorTool.InvokableRun(ctx, calcInput)
	if err != nil {
		log.Printf("计算失败: %v", err)
		return
	}
	fmt.Printf("      %s\n", calcResult)

	// 第二步：查询天气
	fmt.Println("\n    步骤2: 天气查询")
	weatherInput := `{"city": "上海", "units": "celsius"}`
	weatherResult, err := weatherTool.InvokableRun(ctx, weatherInput)
	if err != nil {
		log.Printf("天气查询失败: %v", err)
		return
	}
	fmt.Printf("      天气信息:\n%s\n", indentLines(weatherResult, "      "))

	// 第三步：获取时间
	fmt.Println("\n    步骤3: 时间查询")
	timeInput := `{"timezone": "Asia/Shanghai"}`
	timeResult, err := timeTool.InvokableRun(ctx, timeInput)
	if err != nil {
		log.Printf("时间查询失败: %v", err)
		return
	}
	fmt.Printf("      时间信息:\n%s\n", indentLines(timeResult, "      "))

	fmt.Println("\n    ✅ 工具调用链完成")
}

// 5. 错误处理演示
func errorHandlingDemo(ctx context.Context) {
	fmt.Println("\n❌ 错误处理演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建工具
	calculatorTool := &CalculatorTool{}
	weatherTool := &WeatherTool{apiKey: config.WeatherAPIKey}

	fmt.Println("\n  📝 测试各种错误情况:")

	// 1. 计算器错误
	fmt.Println("\n    测试1: 计算器除零错误")
	_, err := calculatorTool.InvokableRun(ctx, `{"expression": "10 / 0"}`)
	if err != nil {
		fmt.Printf("      ❌ 预期的错误: %v\n", err)
	}

	// 2. 天气工具参数错误
	fmt.Println("\n    测试2: 天气工具参数错误")
	_, err = weatherTool.InvokableRun(ctx, `{"invalid": "json"}`)
	if err != nil {
		fmt.Printf("      ❌ 预期的错误: %v\n", err)
	}

	// 3. 天气工具缺少必需参数
	fmt.Println("\n    测试3: 天气工具缺少必需参数")
	_, err = weatherTool.InvokableRun(ctx, `{"units": "celsius"}`)
	if err != nil {
		fmt.Printf("      ❌ 预期的错误: %v\n", err)
	}

	// 4. 无效JSON输入
	fmt.Println("\n    测试4: 无效JSON输入")
	_, err = weatherTool.InvokableRun(ctx, `invalid json`)
	if err != nil {
		fmt.Printf("      ❌ 预期的错误: %v\n", err)
	}

	fmt.Println("\n    📋 错误处理最佳实践:")
	fmt.Println("      1. 输入验证: 检查所有必需参数")
	fmt.Println("      2. 类型检查: 确保参数类型正确")
	fmt.Println("      3. 边界条件: 处理除零、空值等特殊情况")
	fmt.Println("      4. 清晰错误: 提供有意义的错误信息")
	fmt.Println("      5. 优雅降级: 在可能的情况下提供默认行为")
}

// 6. 性能测试演示
func performanceTestDemo(ctx context.Context) {
	fmt.Println("\n🚀 性能测试演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建工具
	calculatorTool := &CalculatorTool{}

	// 测试工具调用性能
	fmt.Println("\n  📝 工具调用性能测试:")

	testCases := []string{
		`{"expression": "10 + 20"}`,
		`{"expression": "50 - 15"}`,
		`{"expression": "8 * 7"}`,
		`{"expression": "100 / 4"}`,
		`{"expression": "25 + 25"}`,
	}

	totalStart := time.Now()
	for i, input := range testCases {
		start := time.Now()
		result, err := calculatorTool.InvokableRun(ctx, input)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("    测试 %d 失败: %v\n", i+1, err)
			continue
		}

		fmt.Printf("    测试 %d: %s -> 耗时: %v\n", i+1, result, duration)
	}

	totalDuration := time.Since(totalStart)
	fmt.Printf("\n  📊 性能统计:")
	fmt.Printf("    总测试数: %d\n", len(testCases))
	fmt.Printf("    总耗时: %v\n", totalDuration)
	fmt.Printf("    平均耗时: %v\n", totalDuration/time.Duration(len(testCases)))
}

// ========== 辅助函数 ==========

// 为文本添加缩进
func indentLines(text, indent string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder
	for _, line := range lines {
		if line != "" {
			result.WriteString(indent + line + "\n")
		} else {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// 显示使用说明
func showUsage() {
	fmt.Println("用法: go run main.go [example]")
	fmt.Println("示例:")
	fmt.Println("  basic      - 基础工具创建演示")
	fmt.Println("  manual     - 手动工具创建演示")
	fmt.Println("  config     - ToolsNode 配置演示")
	fmt.Println("  chain      - 工具调用链演示")
	fmt.Println("  error      - 错误处理演示")
	fmt.Println("  performance - 性能测试演示")
	fmt.Println("\n不带参数运行所有演示")
}

// ========== 主函数 ==========

func main() {
	// 初始化配置
	if err := initConfig(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	ctx := context.Background()

	fmt.Println("=== Eino ToolsNode 组件演示 ===")

	// 获取命令行参数决定运行哪个示例
	if len(os.Args) > 1 {
		example := os.Args[1]
		switch example {
		case "basic":
			fmt.Println("运行基础工具创建演示...")
			basicToolCreationDemo(ctx)
		case "manual":
			fmt.Println("运行手动工具创建演示...")
			manualToolCreationDemo(ctx)
		case "config":
			fmt.Println("运行ToolsNode配置演示...")
			toolsNodeConfigDemo(ctx)
		case "chain":
			fmt.Println("运行工具调用链演示...")
			toolChainDemo(ctx)
		case "error":
			fmt.Println("运行错误处理演示...")
			errorHandlingDemo(ctx)
		case "performance":
			fmt.Println("运行性能测试演示...")
			performanceTestDemo(ctx)
		case "help":
			showUsage()
			return
		default:
			fmt.Printf("未知示例: %s\n", example)
			showUsage()
			return
		}
	} else {
		// 运行所有演示
		runAllDemos(ctx)
	}
}

func runAllDemos(ctx context.Context) {
	fmt.Println("📝 演示1: 基础工具创建")
	basicToolCreationDemo(ctx)

	fmt.Println("\n🔧 演示2: 手动工具创建")
	manualToolCreationDemo(ctx)

	fmt.Println("\n🏗️ 演示3: ToolsNode配置")
	toolsNodeConfigDemo(ctx)

	fmt.Println("\n🔗 演示4: 工具调用链")
	toolChainDemo(ctx)

	fmt.Println("\n❌ 演示5: 错误处理")
	errorHandlingDemo(ctx)

	fmt.Println("\n🚀 演示6: 性能测试")
	performanceTestDemo(ctx)

	fmt.Println("\n✅ 所有 ToolsNode 演示完成！")
}
