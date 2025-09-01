package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  多代理托管演示 (Multi-Agent Hosting Demo)
//  
//  架构说明:
//  1. Host Agent - 负责意图识别和路由决策
//  2. Specialist Agents - 专门处理特定任务的代理
//     - WeatherSpecialist: 天气查询专家
//     - CalculatorSpecialist: 数学计算专家  
//     - TimeSpecialist: 时间查询专家
//
//  工作流程:
//  用户输入 -> Host Agent 分析 -> 路由到对应专家 -> 返回结果
//
// =============================================================================

// 工具实现
// WeatherTool 天气查询工具
type WeatherTool struct{}

func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "查询指定城市的天气情况",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",
				Desc:     "要查询天气的城市名称，例如：北京、上海、广州等",
				Required: true,
			},
		}),
	}, nil
}

func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		City string `json:"city"`
	}
	
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	
	log.Printf("[WeatherTool] 查询城市: %s", args.City)
	
	// 模拟天气数据
	weatherData := map[string]string{
		"北京": "晴，25°C，微风",
		"上海": "多云，28°C，东南风", 
		"广州": "雷阵雨，30°C，南风",
		"深圳": "晴转多云，29°C，东风",
		"杭州": "小雨，22°C，北风",
		"成都": "阴，24°C，无风",
		"西安": "晴，26°C，西北风",
		"南京": "多云，23°C，东风",
	}
	
	weather, exists := weatherData[args.City]
	if !exists {
		weather = "晴，25°C，微风（默认天气）"
	}
	
	result := map[string]interface{}{
		"city":    args.City,
		"weather": weather,
		"message": fmt.Sprintf("🌤️ %s今天的天气：%s", args.City, weather),
	}
	
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// CalculatorTool 数学计算工具
type CalculatorTool struct{}

func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculate",
		Desc: "进行数学计算，支持基本的四则运算",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {
				Type:     "string",
				Desc:     "数学表达式，例如：10+5、20-8、6*7、15/3",
				Required: true,
			},
		}),
	}, nil
}

func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Expression string `json:"expression"`
	}
	
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	
	log.Printf("[CalculatorTool] 计算表达式: %s", args.Expression)
	
	resultValue, err := simpleCalculate(args.Expression)
	if err != nil {
		return "", fmt.Errorf("计算错误: %v", err)
	}
	
	result := map[string]interface{}{
		"expression": args.Expression,
		"result":     resultValue,
		"message":    fmt.Sprintf("🔢 计算结果：%s = %.2f", args.Expression, resultValue),
	}
	
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// TimeTool 时间查询工具
type TimeTool struct{}

func (t *TimeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_time",
		Desc: "获取当前的日期和时间信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"format": {
				Type:     "string", 
				Desc:     "时间格式：date（仅日期）、time（仅时间）、datetime（日期时间）、timestamp（时间戳）",
				Required: false,
				Enum:     []string{"date", "time", "datetime", "timestamp"},
			},
		}),
	}, nil
}

func (t *TimeTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Format string `json:"format"`
	}
	
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	
	if args.Format == "" {
		args.Format = "datetime"
	}
	
	log.Printf("[TimeTool] 查询时间格式: %s", args.Format)
	
	now := time.Now()
	var message string
	var value interface{}
	
	switch args.Format {
	case "date":
		value = now.Format("2006年01月02日")
		message = fmt.Sprintf("📅 今天是：%s", value)
	case "time":
		value = now.Format("15:04:05")
		message = fmt.Sprintf("⏰ 当前时间：%s", value)
	case "datetime":
		value = now.Format("2006年01月02日 15:04:05")
		message = fmt.Sprintf("📅⏰ 当前日期时间：%s", value)
	case "timestamp":
		value = now.Unix()
		message = fmt.Sprintf("🕐 时间戳：%d", value)
	default:
		value = now.Format("2006年01月02日 15:04:05")
		message = fmt.Sprintf("📅⏰ 当前日期时间：%s", value)
	}
	
	result := map[string]interface{}{
		"format":  args.Format,
		"value":   value,
		"message": message,
	}
	
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// 简单计算器实现
func simpleCalculate(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	
	// 处理加法
	if strings.Contains(expression, "+") {
		parts := strings.Split(expression, "+")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a + b, nil
			}
		}
	}
	
	// 处理减法
	if strings.Contains(expression, "-") && !strings.HasPrefix(expression, "-") {
		parts := strings.Split(expression, "-")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a - b, nil
			}
		}
	}
	
	// 处理乘法
	if strings.Contains(expression, "*") {
		parts := strings.Split(expression, "*")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a * b, nil
			}
		}
	}
	
	// 处理除法
	if strings.Contains(expression, "/") {
		parts := strings.Split(expression, "/")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				if b == 0 {
					return 0, fmt.Errorf("除数不能为零")
				}
				return a / b, nil
			}
		}
	}
	
	// 如果是单个数字
	if num, err := strconv.ParseFloat(expression, 64); err == nil {
		return num, nil
	}
	
	return 0, fmt.Errorf("不支持的表达式格式")
}

// 创建聊天模型的辅助函数
func createChatModel(ctx context.Context, tools []tool.InvokableTool) (*ark.ChatModel, error) {
	config := &ark.ChatModelConfig{
		Model:  viper.GetString("ARK_MODEL"),
		APIKey: viper.GetString("ARK_API_KEY"),
	}
	
	chatModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建聊天模型失败: %v", err)
	}
	
	// 绑定工具到聊天模型
	if len(tools) > 0 {
		toolInfos := make([]*schema.ToolInfo, 0, len(tools))
		for _, tool := range tools {
			info, err := tool.Info(ctx)
			if err != nil {
				log.Printf("获取工具信息失败: %v", err)
				continue
			}
			toolInfos = append(toolInfos, info)
		}
		chatModel.BindTools(toolInfos)
	}
	
	return chatModel, nil
}

// 创建Host Agent - 负责意图识别和路由决策
func createHostAgent(ctx context.Context) (*ark.ChatModel, error) {
	log.Println("创建Host Agent...")
	
	// Host Agent 不需要工具，只负责分析用户意图
	hostModel, err := createChatModel(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("创建Host Agent失败: %v", err)
	}
	
	log.Println("Host Agent创建成功")
	return hostModel, nil
}

// 创建天气专家Agent
func createWeatherSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
	log.Println("创建Weather Specialist Agent...")
	
	tools := []tool.InvokableTool{&WeatherTool{}}
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, fmt.Errorf("创建Weather Specialist失败: %v", err)
	}
	
	// 创建专门的链
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(chatModel, compose.WithNodeName("weather_specialist"))
	
	specialist, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译Weather Specialist失败: %v", err)
	}
	
	log.Println("Weather Specialist Agent创建成功")
	return specialist, nil
}

// 创建计算器专家Agent
func createCalculatorSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
	log.Println("创建Calculator Specialist Agent...")
	
	tools := []tool.InvokableTool{&CalculatorTool{}}
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, fmt.Errorf("创建Calculator Specialist失败: %v", err)
	}
	
	// 创建专门的链
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(chatModel, compose.WithNodeName("calculator_specialist"))
	
	specialist, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译Calculator Specialist失败: %v", err)
	}
	
	log.Println("Calculator Specialist Agent创建成功")
	return specialist, nil
}

// 创建时间专家Agent
func createTimeSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
	log.Println("创建Time Specialist Agent...")
	
	tools := []tool.InvokableTool{&TimeTool{}}
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, fmt.Errorf("创建Time Specialist失败: %v", err)
	}
	
	// 创建专门的链
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(chatModel, compose.WithNodeName("time_specialist"))
	
	specialist, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译Time Specialist失败: %v", err)
	}
	
	log.Println("Time Specialist Agent创建成功")
	return specialist, nil
}

// MultiAgentRouter 多代理路由器 - 模拟多代理托管功能
type MultiAgentRouter struct {
	hostAgent           *ark.ChatModel
	weatherSpecialist   compose.Runnable[[]*schema.Message, *schema.Message]
	calculatorSpecialist compose.Runnable[[]*schema.Message, *schema.Message]
	timeSpecialist      compose.Runnable[[]*schema.Message, *schema.Message]
}

// Invoke 实现 Runnable 接口
func (m *MultiAgentRouter) Invoke(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
	// 1. Host Agent 分析用户意图
	hostPrompt := &schema.Message{
		Role: schema.System,
		Content: `你是一个智能助手调度中心。分析用户请求，确定需要调用哪些专家来处理任务。

可用专家：
- weather: 处理天气查询
- calculator: 处理数学计算  
- time: 处理时间查询

请只返回需要的专家名称，多个用逗号分隔。例如：
- 用户问天气 -> 返回: weather
- 用户要计算 -> 返回: calculator
- 用户问时间 -> 返回: time  
- 用户问天气和时间 -> 返回: weather,time

不要返回其他内容，只返回专家名称。`,
	}
	
	// 构建Host Agent的输入
	hostInput := append([]*schema.Message{hostPrompt}, input...)
	
	// Host Agent 分析意图
	hostResponse, err := m.hostAgent.Generate(ctx, hostInput)
	if err != nil {
		return nil, fmt.Errorf("Host Agent 分析失败: %v", err)
	}
	
	log.Printf("[Host Agent] 路由决策: %s", hostResponse.Content)
	
	// 2. 解析专家选择
	specialists := strings.Split(strings.TrimSpace(hostResponse.Content), ",")
	
	var results []string
	
	// 3. 依次调用选中的专家
	for _, specialist := range specialists {
		specialist = strings.TrimSpace(specialist)
		
		var result *schema.Message
		var err error
		
		// 为专家添加特定的系统提示
		specialistInput := make([]*schema.Message, len(input))
		copy(specialistInput, input)
		
		// 处理专家名称的变体（去除后缀）
		if strings.Contains(specialist, "weather") {
			log.Println("[Router] 调用 Weather Specialist")
			weatherPrompt := &schema.Message{
				Role: schema.System,
				Content: "你是天气查询专家。用户询问天气时，请使用天气工具获取信息并以友好的方式回复。",
			}
			specialistInput = append([]*schema.Message{weatherPrompt}, specialistInput...)
			result, err = m.weatherSpecialist.Invoke(ctx, specialistInput)
			
		} else if strings.Contains(specialist, "calculator") {
			log.Println("[Router] 调用 Calculator Specialist")  
			calcPrompt := &schema.Message{
				Role: schema.System,
				Content: "你是数学计算专家。用户需要计算时，请使用计算工具并清晰地展示计算过程和结果。",
			}
			specialistInput = append([]*schema.Message{calcPrompt}, specialistInput...)
			result, err = m.calculatorSpecialist.Invoke(ctx, specialistInput)
			
		} else if strings.Contains(specialist, "time") {
			log.Println("[Router] 调用 Time Specialist")
			timePrompt := &schema.Message{
				Role: schema.System,
				Content: "你是时间查询专家。用户询问时间时，请使用时间工具获取准确的时间信息。",
			}
			specialistInput = append([]*schema.Message{timePrompt}, specialistInput...)
			result, err = m.timeSpecialist.Invoke(ctx, specialistInput)
			
		} else {
			log.Printf("[Router] 未知专家: %s", specialist)
			continue
		}
		
		if err != nil {
			log.Printf("[Router] 专家 %s 执行失败: %v", specialist, err)
			continue
		}
		
		if result != nil && result.Content != "" {
			results = append(results, result.Content)
			log.Printf("[Router] 专家 %s 返回: %s", specialist, result.Content)
		}
	}
	
	// 4. 整合结果
	if len(results) == 0 {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "抱歉，我无法处理您的请求。",
		}, nil
	}
	
	finalContent := strings.Join(results, "\n\n")
	
	return &schema.Message{
		Role:    schema.Assistant,
		Content: finalContent,
	}, nil
}

// 创建多代理托管系统
func createMultiAgentSystem(ctx context.Context) (*MultiAgentRouter, error) {
	log.Println("创建Multi-Agent Hosting系统...")
	
	// 1. 创建Host Agent
	hostAgent, err := createHostAgent(ctx)
	if err != nil {
		return nil, err
	}
	
	// 2. 创建专家Agents
	weatherSpecialist, err := createWeatherSpecialist(ctx)
	if err != nil {
		return nil, err
	}
	
	calculatorSpecialist, err := createCalculatorSpecialist(ctx)
	if err != nil {
		return nil, err
	}
	
	timeSpecialist, err := createTimeSpecialist(ctx)
	if err != nil {
		return nil, err
	}
	
	// 3. 创建多代理路由器
	router := &MultiAgentRouter{
		hostAgent:           hostAgent,
		weatherSpecialist:   weatherSpecialist,
		calculatorSpecialist: calculatorSpecialist,
		timeSpecialist:      timeSpecialist,
	}
	
	log.Println("Multi-Agent Hosting系统创建成功")
	return router, nil
}

func main() {
	// 初始化配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	
	// 设置环境变量
	os.Setenv("ARK_API_KEY", viper.GetString("ARK_API_KEY"))
	
	ctx := context.Background()
	
	// 创建多代理托管系统
	multiAgentSystem, err := createMultiAgentSystem(ctx)
	if err != nil {
		log.Fatalf("创建多代理系统失败: %v", err)
	}
	
	fmt.Println("🤖 多代理托管系统演示启动成功！")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("系统架构:")
	fmt.Println("🏢 Host Agent - 负责意图识别和路由决策")  
	fmt.Println("🌤️  Weather Specialist - 天气查询专家")
	fmt.Println("🔢 Calculator Specialist - 数学计算专家") 
	fmt.Println("⏰ Time Specialist - 时间查询专家")
	fmt.Println(strings.Repeat("=", 60))
	
	// 测试用例 - 简化版本专注于基本功能
	testCases := []struct {
		name    string
		message string
	}{
		{
			name:    "天气查询测试",
			message: "北京天气",
		},
		{
			name:    "数学计算测试", 
			message: "计算 10 + 5",
		},
	}
	
	for i, testCase := range testCases {
		fmt.Printf("\n📝 测试用例 %d: %s\n", i+1, testCase.name)
		fmt.Printf("👤 用户: %s\n", testCase.message)
		fmt.Println(strings.Repeat("-", 50))
		
		// 添加系统消息和用户消息
		systemMessage := &schema.Message{
			Role: schema.System,
			Content: `你是一个智能助手调度中心的Host Agent。你的任务是分析用户的意图，然后决定调用哪个专家代理来处理请求。

可用的专家代理:
1. weather_specialist - 处理天气查询相关问题
2. calculator_specialist - 处理数学计算相关问题  
3. time_specialist - 处理时间和日期查询相关问题

请根据用户的问题，选择合适的专家代理来处理。如果需要多个步骤，可以依次调用多个专家。

分析用户意图，然后给出你的路由决策。`,
		}
		
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: testCase.message,
		}
		
		// 执行多代理系统
		messages := []*schema.Message{systemMessage, userMessage}
		result, err := multiAgentSystem.Invoke(ctx, messages)
		if err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
			continue
		}
		
		// 输出结果
		if result != nil && result.Content != "" {
			fmt.Printf("🤖 系统回复: %s\n", result.Content)
		} else {
			fmt.Printf("🤖 系统回复: [没有收到回复内容]\n")
		}
		
		// 等待一秒，让输出更清晰
		time.Sleep(1 * time.Second)
	}
	
	fmt.Println("\n✅ 所有测试用例执行完成！")
	fmt.Println("\n💡 多代理托管系统演示:")
	fmt.Println("   - Host Agent 负责分析用户意图")
	fmt.Println("   - 根据意图路由到对应的专家代理")
	fmt.Println("   - 每个专家代理专注于特定领域的任务")
	fmt.Println("   - 支持复杂的多步骤任务处理")
	fmt.Println("📚 如需了解更多功能，请参考 README.md 文档。")
}