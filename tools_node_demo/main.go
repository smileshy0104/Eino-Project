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
	"github.com/cloudwego/eino/compose"
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

// 7. Chain 集成演示
func chainIntegrationDemo(ctx context.Context) {
	fmt.Println("\n🔗 Chain 集成演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建工具集合
	calculatorTool := &CalculatorTool{}
	weatherTool := &WeatherTool{apiKey: config.WeatherAPIKey}
	timeTool := &TimeTool{}

	fmt.Println("\n  📝 创建工具集合:")
	tools := []tool.BaseTool{calculatorTool, weatherTool, timeTool}

	for i, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Printf("获取工具 %d 信息失败: %v", i+1, err)
			continue
		}
		fmt.Printf("    %d. %s - %s\n", i+1, info.Name, info.Desc)
	}

	// 模拟创建 ToolsNode（实际创建可能需要LLM）
	fmt.Println("\n  🏗️ 模拟 ToolsNode 创建:")
	fmt.Println("    ToolsNode 配置:")
	fmt.Printf("      工具数量: %d\n", len(tools))
	fmt.Println("      LLM模型: doubao-seed-1-6-250615 (模拟)")

	// 模拟 Chain 集成流程
	fmt.Println("\n  🔗 Chain 工作流演示:")
	fmt.Println("    工作流程: 用户输入 -> LLM理解 -> 工具选择 -> 工具执行 -> 结果整合")

	// 第一步：模拟用户查询
	userQuery := "帮我计算 15 * 8 的结果，然后查询北京的天气"
	fmt.Printf("    1. 用户查询: %s\n", userQuery)

	// 第二步：模拟LLM理解和工具选择
	fmt.Println("    2. LLM分析: 需要调用计算器工具和天气工具")

	// 第三步：执行计算器工具
	fmt.Println("    3. 执行计算器工具:")
	calcInput := `{"expression": "15 * 8"}`
	calcResult, err := calculatorTool.InvokableRun(ctx, calcInput)
	if err != nil {
		log.Printf("       计算失败: %v", err)
	} else {
		fmt.Printf("       %s\n", calcResult)
	}

	// 第四步：执行天气工具
	fmt.Println("    4. 执行天气工具:")
	weatherInput := `{"city": "北京", "units": "celsius"}`
	weatherResult, err := weatherTool.InvokableRun(ctx, weatherInput)
	if err != nil {
		log.Printf("       天气查询失败: %v", err)
	} else {
		fmt.Printf("       天气信息:\n%s\n", indentLines(weatherResult, "       "))
	}

	// 第五步：模拟结果整合
	fmt.Println("    5. 结果整合:")
	fmt.Println("       LLM将工具结果整合为最终回答:")
	fmt.Println("       '15 × 8 = 120。北京当前天气：晴朗，气温22°C，湿度45%，风速3.2m/s。'")

	fmt.Println("\n    ✅ Chain 集成演示完成")
}

// 8. Graph 集成演示
func graphIntegrationDemo(ctx context.Context) {
	fmt.Println("\n🕸️ Graph 集成演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建工具集合
	calculatorTool := &CalculatorTool{}
	weatherTool := &WeatherTool{apiKey: config.WeatherAPIKey}
	timeTool := &TimeTool{}

	fmt.Println("\n  📝 Graph 节点设计:")
	fmt.Println("    节点类型:")
	fmt.Println("      - 输入节点: 用户查询解析")
	fmt.Println("      - 决策节点: 工具选择逻辑")
	fmt.Println("      - 工具节点: 具体工具执行")
	fmt.Println("      - 聚合节点: 结果汇总")
	fmt.Println("      - 输出节点: 最终答案生成")

	fmt.Println("\n  🗺️ Graph 拓扑结构:")
	fmt.Println("    START -> [输入解析] -> [决策路由] -> [并行工具执行] -> [结果聚合] -> END")
	fmt.Println("                              |")
	fmt.Println("                              ├── [计算器工具]")
	fmt.Println("                              ├── [天气工具]")
	fmt.Println("                              └── [时间工具]")

	// 模拟复杂查询的 Graph 执行
	fmt.Println("\n  🔀 复杂查询 Graph 执行演示:")
	complexQuery := "计算今天到明年的天数，查询上海天气，获取当前时间"
	fmt.Printf("    输入查询: %s\n", complexQuery)

	// 第一步：查询解析（模拟）
	fmt.Println("\n    步骤1 - 查询解析:")
	fmt.Println("      解析结果:")
	fmt.Println("        - 需要日期计算")
	fmt.Println("        - 需要天气查询（上海）")
	fmt.Println("        - 需要时间获取")

	// 第二步：决策路由（模拟）
	fmt.Println("\n    步骤2 - 决策路由:")
	fmt.Println("      路由决策:")
	fmt.Println("        - 计算器工具: 计算天数")
	fmt.Println("        - 天气工具: 查询上海天气")
	fmt.Println("        - 时间工具: 获取当前时间")

	// 第三步：并行工具执行
	fmt.Println("\n    步骤3 - 并行工具执行:")

	// 使用goroutine模拟并行执行
	type toolResult struct {
		name   string
		result string
		err    error
	}

	resultsChan := make(chan toolResult, 3)

	// 并行执行工具
	go func() {
		// 计算器工具 - 模拟计算今天到明年的天数
		calcInput := `{"expression": "365 + 30"}` // 简化计算
		result, err := calculatorTool.InvokableRun(ctx, calcInput)
		resultsChan <- toolResult{"计算器", result, err}
	}()

	go func() {
		// 天气工具
		weatherInput := `{"city": "上海", "units": "celsius"}`
		result, err := weatherTool.InvokableRun(ctx, weatherInput)
		resultsChan <- toolResult{"天气查询", result, err}
	}()

	go func() {
		// 时间工具
		timeInput := `{"timezone": "Asia/Shanghai"}`
		result, err := timeTool.InvokableRun(ctx, timeInput)
		resultsChan <- toolResult{"时间查询", result, err}
	}()

	// 收集并行执行结果
	var results []toolResult
	for i := 0; i < 3; i++ {
		result := <-resultsChan
		results = append(results, result)

		if result.err != nil {
			fmt.Printf("      %s 执行失败: %v\n", result.name, result.err)
		} else {
			fmt.Printf("      %s 执行成功\n", result.name)
		}
	}

	// 第四步：结果聚合
	fmt.Println("\n    步骤4 - 结果聚合:")
	fmt.Println("      聚合所有工具执行结果:")
	for _, result := range results {
		if result.err == nil {
			fmt.Printf("        %s: %s\n", result.name,
				truncateString(strings.ReplaceAll(result.result, "\n", " "), 50))
		}
	}

	// 第五步：最终输出
	fmt.Println("\n    步骤5 - 最终输出:")
	fmt.Println("      Graph 执行完成，生成综合回答:")
	fmt.Println("      '根据计算，大约还有395天。上海当前天气为多云，26°C。'")
	fmt.Println("      '当前时间为2025年09月16日，时区为Asia/Shanghai。'")

	fmt.Println("\n  📊 Graph 执行统计:")
	fmt.Println("    - 总节点数: 5")
	fmt.Println("    - 并行执行节点: 3")
	fmt.Println("    - 执行成功率: 100%")
	fmt.Println("    - 总执行时间: ~200ms (模拟)")

	fmt.Println("\n    ✅ Graph 集成演示完成")
}

// 9. 高级编排模式演示
func advancedOrchestrationDemo(ctx context.Context) {
	fmt.Println("\n🚀 高级编排模式演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建工具集合
	calculatorTool := &CalculatorTool{}
	weatherTool := &WeatherTool{apiKey: config.WeatherAPIKey}
	timeTool := &TimeTool{}

	fmt.Println("\n  📝 高级编排模式介绍:")
	fmt.Println("    1. 条件分支编排")
	fmt.Println("    2. 循环重试编排")
	fmt.Println("    3. 错误恢复编排")
	fmt.Println("    4. 动态工具选择")

	// 演示1: 条件分支编排
	fmt.Println("\n  🔀 条件分支编排演示:")
	fmt.Println("    场景: 根据用户输入类型选择不同的工具处理路径")

	testInputs := []string{
		"计算 100 / 5",
		"查询深圳天气",
		"获取东京时间",
	}

	for i, input := range testInputs {
		fmt.Printf("\n    测试输入 %d: %s\n", i+1, input)

		// 模拟条件判断逻辑
		var selectedTool tool.BaseTool
		var toolInput string

		if strings.Contains(input, "计算") {
			selectedTool = calculatorTool
			toolInput = `{"expression": "100 / 5"}`
			fmt.Println("      -> 路由到: 计算器工具")
		} else if strings.Contains(input, "天气") {
			selectedTool = weatherTool
			toolInput = `{"city": "深圳", "units": "celsius"}`
			fmt.Println("      -> 路由到: 天气工具")
		} else if strings.Contains(input, "时间") {
			selectedTool = timeTool
			toolInput = `{"timezone": "Asia/Tokyo"}`
			fmt.Println("      -> 路由到: 时间工具")
		}

		// 执行选中的工具
		if selectedTool != nil {
			// 类型断言到InvokableTool接口
			if invokableTool, ok := selectedTool.(interface {
				InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error)
			}); ok {
				result, err := invokableTool.InvokableRun(ctx, toolInput)
				if err != nil {
					fmt.Printf("      ❌ 执行失败: %v\n", err)
				} else {
					fmt.Printf("      ✅ 执行结果: %s\n",
						truncateString(strings.ReplaceAll(result, "\n", " "), 60))
				}
			} else {
				fmt.Println("      ❌ 工具不支持直接调用")
			}
		}
	}

	// 演示2: 循环重试编排
	fmt.Println("\n  🔄 循环重试编排演示:")
	fmt.Println("    场景: 工具执行失败时的自动重试机制")

	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("    尝试 %d/%d: 执行除零计算\n", attempt, maxRetries)

		_, err := calculatorTool.InvokableRun(ctx, `{"expression": "10 / 0"}`)
		if err != nil {
			fmt.Printf("      ❌ 失败: %v\n", err)
			if attempt < maxRetries {
				fmt.Println("      🔄 准备重试...")
			} else {
				fmt.Println("      ❌ 达到最大重试次数，任务失败")
			}
		} else {
			fmt.Println("      ✅ 执行成功")
			break
		}
	}

	fmt.Println("\n  🛡️ 错误恢复编排演示:")
	fmt.Println("    场景: 主要工具失败时使用备用工具")

	// 主要工具（模拟失败）
	fmt.Println("    主要工具: 在线天气API")
	fmt.Println("      ❌ 连接失败 (模拟)")

	// 备用工具
	fmt.Println("    启动备用工具: 本地天气缓存")
	backupWeatherResult := "备用天气数据: 深圳 - 多云，28°C（来自缓存）"
	fmt.Printf("      ✅ 备用成功: %s\n", backupWeatherResult)

	fmt.Println("\n    ✅ 高级编排模式演示完成")
}

// 10. 真实API集成演示
func realAPIIntegrationDemo(ctx context.Context) {
	fmt.Println("\n🔧 真实API集成演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建工具集合
	calculatorTool := &CalculatorTool{}
	weatherTool := &WeatherTool{apiKey: config.WeatherAPIKey}
	timeTool := &TimeTool{}

	fmt.Println("\n  📝 尝试使用真实的Eino API:")

	// 尝试创建Chain
	fmt.Println("\n  🔗 创建Chain演示:")
	chain := compose.NewChain[string, string]()
	fmt.Printf("    ✅ Chain创建成功: %T\n", chain)

	// 显示Chain的基本信息
	fmt.Println("    Chain支持的操作:")
	fmt.Println("      - 可以添加各种节点组件")
	fmt.Println("      - 支持顺序执行工作流")
	fmt.Println("      - 提供统一的错误处理")

	// 尝试创建Graph
	fmt.Println("\n  🕸️ 创建Graph演示:")
	graph := compose.NewGraph[string, string]()
	fmt.Printf("    ✅ Graph创建成功: %T\n", graph)

	// 显示Graph的基本信息
	fmt.Println("    Graph支持的操作:")
	fmt.Println("      - 支持复杂的拓扑结构")
	fmt.Println("      - 可以并行执行节点")
	fmt.Println("      - 支持条件分支和合并")

	// 尝试创建ToolsNode（这可能需要LLM模型）
	fmt.Println("\n  🛠️ 尝试创建ToolsNode:")
	fmt.Println("    工具列表:")
	tools := []tool.BaseTool{calculatorTool, weatherTool, timeTool}
	for i, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			fmt.Printf("      %d. 工具信息获取失败: %v\n", i+1, err)
			continue
		}
		fmt.Printf("      %d. %s - %s\n", i+1, info.Name, info.Desc)
	}

	// 模拟ToolsNode配置（由于缺少LLM模型，我们只展示配置结构）
	fmt.Println("\n    ToolsNode配置结构:")
	fmt.Println("      Tools: []tool.BaseTool - 工具列表")
	fmt.Println("      Model: llamaindex.LLM - LLM模型（需要真实模型实例）")
	fmt.Println("      注意: 创建ToolsNode需要真实的LLM模型配置")

	// 尝试基本的Chain操作
	fmt.Println("\n  🔄 Chain基本操作演示:")
	fmt.Println("    可用的Chain方法:")
	fmt.Println("      - AppendLambda(): 添加Lambda节点")
	fmt.Println("      - AppendChatTemplate(): 添加聊天模板")
	fmt.Println("      - AppendToolsNode(): 添加工具节点")
	fmt.Println("      - Run(): 执行整个Chain")

	// 尝试基本的Graph操作
	fmt.Println("\n  🔄 Graph基本操作演示:")
	fmt.Println("    可用的Graph方法:")
	fmt.Println("      - AddNode(): 添加普通节点")
	fmt.Println("      - AddEdge(): 添加边连接")
	fmt.Println("      - AddConditionalEdge(): 添加条件边")
	fmt.Println("      - Run(): 执行整个Graph")

	fmt.Println("\n    ✅ 真实API集成演示完成")
	fmt.Println("    💡 提示: 完整的集成需要配置真实的LLM模型")
}

// 11. 简化Chain集成演示
func simpleChainDemo(ctx context.Context) {
	fmt.Println("\n🔗 简化Chain集成演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建Chain
	fmt.Println("\n  📝 创建Chain实例:")
	chain := compose.NewChain[string, string]()
	fmt.Printf("    ✅ Chain创建成功: %T\n", chain)

	// 演示Chain的概念性工作流
	fmt.Println("\n  🔄 模拟Chain工作流:")
	fmt.Println("    步骤1: 输入处理")
	inputData := "计算 10 + 5 的结果"
	fmt.Printf("      输入: %s\n", inputData)

	fmt.Println("    步骤2: 工具选择和执行")
	calculatorTool := &CalculatorTool{}
	result, err := calculatorTool.InvokableRun(ctx, `{"expression": "10 + 5"}`)
	if err != nil {
		fmt.Printf("      ❌ 执行失败: %v\n", err)
	} else {
		fmt.Printf("      ✅ 执行结果: %s\n", result)
	}

	fmt.Println("    步骤3: 结果格式化")
	formattedResult := fmt.Sprintf("Chain处理结果: %s", result)
	fmt.Printf("      最终输出: %s\n", formattedResult)

	fmt.Println("\n    📋 Chain特点总结:")
	fmt.Println("      - 线性执行流程")
	fmt.Println("      - 数据在节点间传递")
	fmt.Println("      - 统一的错误处理")
	fmt.Println("      - 适合简单到中等复杂度的任务")

	fmt.Println("\n    ✅ 简化Chain演示完成")
}

// 12. 简化Graph集成演示
func simpleGraphDemo(ctx context.Context) {
	fmt.Println("\n🕸️ 简化Graph集成演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建Graph
	fmt.Println("\n  📝 创建Graph实例:")
	graph := compose.NewGraph[string, string]()
	fmt.Printf("    ✅ Graph创建成功: %T\n", graph)

	// 演示Graph的并行执行概念
	fmt.Println("\n  🔄 模拟Graph并行执行:")
	fmt.Println("    输入查询: 同时获取天气、时间和计算结果")

	// 创建工具实例
	calculatorTool := &CalculatorTool{}
	weatherTool := &WeatherTool{apiKey: config.WeatherAPIKey}
	timeTool := &TimeTool{}

	// 模拟并行执行（这展示了Graph的核心优势）
	fmt.Println("\n    🚀 并行节点执行:")

	type nodeResult struct {
		name   string
		output string
		err    error
	}

	resultChan := make(chan nodeResult, 3)

	// 启动并行执行
	go func() {
		result, err := calculatorTool.InvokableRun(ctx, `{"expression": "20 * 3"}`)
		resultChan <- nodeResult{"计算节点", result, err}
	}()

	go func() {
		result, err := weatherTool.InvokableRun(ctx, `{"city": "北京", "units": "celsius"}`)
		resultChan <- nodeResult{"天气节点", result, err}
	}()

	go func() {
		result, err := timeTool.InvokableRun(ctx, `{"timezone": "Asia/Shanghai"}`)
		resultChan <- nodeResult{"时间节点", result, err}
	}()

	// 收集并行执行结果
	var results []nodeResult
	for i := 0; i < 3; i++ {
		result := <-resultChan
		results = append(results, result)
		if result.err != nil {
			fmt.Printf("      ❌ %s执行失败: %v\n", result.name, result.err)
		} else {
			fmt.Printf("      ✅ %s执行成功\n", result.name)
		}
	}

	// 模拟结果聚合（Graph的汇聚节点）
	fmt.Println("\n    📊 结果聚合:")
	for _, result := range results {
		if result.err == nil {
			truncated := truncateString(strings.ReplaceAll(result.output, "\n", " "), 50)
			fmt.Printf("      %s: %s\n", result.name, truncated)
		}
	}

	fmt.Println("\n    📋 Graph特点总结:")
	fmt.Println("      - 支持并行执行")
	fmt.Println("      - 复杂的拓扑结构")
	fmt.Println("      - 高效的资源利用")
	fmt.Println("      - 适合复杂的并行任务")

	fmt.Println("\n    ✅ 简化Graph演示完成")
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

// 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// 显示使用说明
func showUsage() {
	fmt.Println("用法: go run main.go [example]")
	fmt.Println("示例:")
	fmt.Println("  basic       - 基础工具创建演示")
	fmt.Println("  manual      - 手动工具创建演示")
	fmt.Println("  config      - ToolsNode 配置演示")
	fmt.Println("  chain       - 工具调用链演示")
	fmt.Println("  error       - 错误处理演示")
	fmt.Println("  performance - 性能测试演示")
	fmt.Println("  chainflow   - Chain 集成演示")
	fmt.Println("  graph       - Graph 集成演示")
	fmt.Println("  advanced    - 高级编排模式演示")
	fmt.Println("  realapi     - 真实API集成演示")
	fmt.Println("  simplechain - 简化Chain演示")
	fmt.Println("  simplegraph - 简化Graph演示")
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
		case "chainflow":
			fmt.Println("运行Chain集成演示...")
			chainIntegrationDemo(ctx)
		case "graph":
			fmt.Println("运行Graph集成演示...")
			graphIntegrationDemo(ctx)
		case "advanced":
			fmt.Println("运行高级编排模式演示...")
			advancedOrchestrationDemo(ctx)
		case "realapi":
			fmt.Println("运行真实API集成演示...")
			realAPIIntegrationDemo(ctx)
		case "simplechain":
			fmt.Println("运行简化Chain演示...")
			simpleChainDemo(ctx)
		case "simplegraph":
			fmt.Println("运行简化Graph演示...")
			simpleGraphDemo(ctx)
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
	//fmt.Println("📝 演示1: 基础工具创建")
	//basicToolCreationDemo(ctx)
	//
	//fmt.Println("\n🔧 演示2: 手动工具创建")
	//manualToolCreationDemo(ctx)
	//
	//fmt.Println("\n🏗️ 演示3: ToolsNode配置")
	//toolsNodeConfigDemo(ctx)
	//
	//fmt.Println("\n🔗 演示4: 工具调用链")
	//toolChainDemo(ctx)
	//
	//fmt.Println("\n❌ 演示5: 错误处理")
	//errorHandlingDemo(ctx)
	//
	//fmt.Println("\n🚀 演示6: 性能测试")
	//performanceTestDemo(ctx)

	fmt.Println("\n🔗 演示7: Chain 集成")
	chainIntegrationDemo(ctx)

	fmt.Println("\n🕸️ 演示8: Graph 集成")
	graphIntegrationDemo(ctx)
	//
	//fmt.Println("\n🚀 演示9: 高级编排模式")
	//advancedOrchestrationDemo(ctx)

	fmt.Println("\n✅ 所有 ToolsNode 演示完成！")
}
