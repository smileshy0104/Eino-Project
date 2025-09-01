// =============================================================================
//
//  React Agent 演示程序 (React Agent Demo)
//  
//  本程序展示了如何使用 Eino 框架构建一个 React Agent 系统。
//  React Agent 是一种基于推理-行动循环的智能代理，能够：
//  1. 分析用户输入，理解任务需求
//  2. 选择合适的工具来执行任务
//  3. 根据工具执行结果进行推理
//  4. 决定下一步行动或给出最终答案
//
//  支持的功能:
//  - 天气查询：获取指定城市的天气信息
//  - 数学计算：执行基本的四则运算
//  - 时间查询：获取当前日期和时间信息
//
//  架构特点:
//  - 单一代理处理多种任务类型
//  - 工具驱动的任务执行模式
//  - 自然语言交互界面
//
// =============================================================================

package main

import (
	"context"     // 上下文管理，用于控制请求生命周期
	"encoding/json" // JSON 编解码，用于工具参数处理
	"fmt"        // 格式化输出
	"log"        // 日志记录
	"os"         // 操作系统接口，用于环境变量设置
	"strconv"    // 字符串转换，用于数值计算
	"strings"    // 字符串处理
	"time"       // 时间处理

	"github.com/cloudwego/eino-ext/components/model/ark" // ARK 大语言模型组件
	"github.com/cloudwego/eino/components/tool"         // 工具接口定义
	"github.com/cloudwego/eino/compose"                 // 组合器，用于构建代理链
	"github.com/cloudwego/eino/schema"                  // 消息和工具的模式定义
	"github.com/spf13/viper"                           // 配置文件管理
)

// =============================================================================
// 工具实现部分 - React Agent 的核心能力组件
// =============================================================================

// WeatherTool 天气查询工具
// 提供城市天气信息查询功能，支持多个主要城市的模拟天气数据
// 实现了 tool.InvokableTool 接口，可被 React Agent 调用
type WeatherTool struct{}

// Info 返回天气查询工具的元数据信息
// 这些信息帮助 LLM 理解工具的功能、参数要求和使用方式
// 返回值包含工具名称、描述和参数规范，供 React Agent 进行工具选择和调用
func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",                    // 工具标识符，LLM 通过此名称调用工具
		Desc: "查询指定城市的天气情况",              // 工具功能描述，帮助 LLM 理解使用场景
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {                           // 城市参数定义
				Type:     "string",              // 参数类型：字符串
				Desc:     "要查询天气的城市名称，例如：北京、上海、广州等", // 参数说明和示例
				Required: true,                  // 必需参数
			},
		}),
	}, nil
}

// InvokableRun 执行天气查询的核心逻辑
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//   - argumentsInJSON: JSON 格式的工具参数，包含要查询的城市名称
//   - opts: 可选的工具执行选项
// 返回值:
//   - string: JSON 格式的查询结果，包含城市、天气信息和友好的消息
//   - error: 执行过程中的错误信息
func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体，用于解析 JSON 参数
	var args struct {
		City string `json:"city"` // 城市名称参数
	}

	// 解析 JSON 参数到结构体
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 模拟天气数据库 - 在实际应用中，这里会调用真实的天气 API
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

	// 查找城市天气，如果城市不存在则使用默认天气
	weather, exists := weatherData[args.City]
	if !exists {
		weather = "晴，25°C，微风（默认天气）"
	}

	// 构建返回结果，包含结构化数据和用户友好的消息
	result := map[string]interface{}{
		"city":    args.City,                                                    // 查询的城市
		"weather": weather,                                                      // 天气信息
		"message": fmt.Sprintf("🌤️ %s今天的天气：%s", args.City, weather), // 友好的回复消息
	}

	// 将结果序列化为 JSON 格式返回给 React Agent
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// CalculatorTool 数学计算工具
// 提供基本的四则运算功能，支持加法、减法、乘法、除法运算
// 实现了 tool.InvokableTool 接口，可被 React Agent 调用执行数学计算任务
type CalculatorTool struct{}

// Info 返回工具的元数据信息
// Info 返回数学计算工具的元数据信息
// 定义工具的名称、功能描述和参数规范，帮助 LLM 理解如何使用此工具
// 支持的运算类型包括：加法(+)、减法(-)、乘法(*)、除法(/)
func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculate",                           // 工具标识符，LLM 通过此名称调用计算功能
		Desc: "进行数学计算，支持基本的四则运算",          // 工具功能描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {                         // 数学表达式参数定义
				Type:     "string",                  // 参数类型：字符串格式的数学表达式
				Desc:     "数学表达式，例如：10+5、20-8、6*7、15/3", // 参数说明和使用示例
				Required: true,                      // 必需参数
			},
		}),
	}, nil
}

// InvokableRun 执行数学计算的核心逻辑
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//   - argumentsInJSON: JSON 格式的工具参数，包含要计算的数学表达式
//   - opts: 可选的工具执行选项
// 返回值:
//   - string: JSON 格式的计算结果，包含表达式、结果值和友好的消息
//   - error: 计算过程中的错误信息（如除零错误、格式错误等）
func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体，用于解析 JSON 参数
	var args struct {
		Expression string `json:"expression"` // 数学表达式参数
	}

	// 解析 JSON 参数到结构体
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 调用计算函数执行数学运算（仅支持基本四则运算）
	resultValue, err := simpleCalculate(args.Expression)
	if err != nil {
		return "", fmt.Errorf("计算错误: %v", err)
	}

	// 构建返回结果，包含原始表达式、计算结果和用户友好的消息
	result := map[string]interface{}{
		"expression": args.Expression,                                                      // 原始数学表达式
		"result":     resultValue,                                                         // 计算结果（数值）
		"message":    fmt.Sprintf("🔢 计算结果：%s = %.2f", args.Expression, resultValue), // 友好的回复消息
	}

	// 将结果序列化为 JSON 格式返回给 React Agent
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// simpleCalculate 简单计算器实现函数
// 支持基本的四则运算：加法(+)、减法(-)、乘法(*)、除法(/)
// 参数:
//   - expression: 数学表达式字符串，如 "10+5"、"20-8"、"6*7"、"15/3"
// 返回值:
//   - float64: 计算结果
//   - error: 计算错误（如格式错误、除零错误等）
// 注意: 这是一个简化版本，仅支持两个操作数的基本运算，不支持复杂表达式和运算优先级
func simpleCalculate(expression string) (float64, error) {
	// 移除表达式中的所有空格，统一格式
	expression = strings.ReplaceAll(expression, " ", "")

	// 处理加法运算：查找 "+" 符号并分割表达式
	if strings.Contains(expression, "+") {
		parts := strings.Split(expression, "+")
		if len(parts) == 2 { // 确保只有两个操作数
			// 将字符串转换为浮点数
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a + b, nil // 执行加法运算
			}
		}
	}

	// 处理减法运算：查找 "-" 符号并分割表达式
	// 注意：排除负数情况（表达式开头的 "-" 号）
	if strings.Contains(expression, "-") && !strings.HasPrefix(expression, "-") {
		parts := strings.Split(expression, "-")
		if len(parts) == 2 { // 确保只有两个操作数
			// 将字符串转换为浮点数
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a - b, nil // 执行减法运算
			}
		}
	}

	// 处理乘法运算：查找 "*" 符号并分割表达式
	if strings.Contains(expression, "*") {
		parts := strings.Split(expression, "*")
		if len(parts) == 2 { // 确保只有两个操作数
			// 将字符串转换为浮点数
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a * b, nil // 执行乘法运算
			}
		}
	}

	// 处理除法运算：查找 "/" 符号并分割表达式
	if strings.Contains(expression, "/") {
		parts := strings.Split(expression, "/")
		if len(parts) == 2 { // 确保只有两个操作数
			// 将字符串转换为浮点数
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				// 检查除零错误
				if b == 0 {
					return 0, fmt.Errorf("除数不能为零")
				}
				return a / b, nil // 执行除法运算
			}
		}
	}

	// 处理单个数字的情况（没有运算符）
	if num, err := strconv.ParseFloat(expression, 64); err == nil {
		return num, nil
	}

	// 如果所有格式都不匹配，返回错误
	return 0, fmt.Errorf("不支持的表达式格式")
}

// TimeTool 时间查询工具
// 提供当前日期和时间信息的查询功能，支持多种时间格式输出
// 支持的格式包括：仅日期、仅时间、完整日期时间、Unix时间戳
// 实现了 tool.InvokableTool 接口，可被 React Agent 调用
type TimeTool struct{}

// Info 返回工具的元数据信息
// Info 返回时间查询工具的元数据信息
// 定义工具的名称、功能描述和参数规范，帮助 LLM 理解如何使用此工具
// 支持多种时间格式，满足不同场景的时间查询需求
func (t *TimeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_time",                        // 工具标识符，LLM 通过此名称调用时间查询功能
		Desc: "获取当前的日期和时间信息",            // 工具功能描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"format": {                          // 时间格式参数定义
				Type:     "string",               // 参数类型：字符串
				Desc:     "时间格式：date（仅日期）、time（仅时间）、datetime（日期时间）、timestamp（时间戳）", // 参数说明
				Required: false,                  // 可选参数，默认为 datetime
				Enum:     []string{"date", "time", "datetime", "timestamp"}, // 枚举值，限制可选格式
			},
		}),
	}, nil
}

// InvokableRun 执行时间查询的核心逻辑
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//   - argumentsInJSON: JSON 格式的工具参数，包含时间格式要求
//   - opts: 可选的工具执行选项
// 返回值:
//   - string: JSON 格式的时间信息，包含格式、时间值和友好的消息
//   - error: 执行过程中的错误信息
func (t *TimeTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体，用于解析 JSON 参数
	var args struct {
		Format string `json:"format"` // 时间格式参数
	}

	// 解析 JSON 参数到结构体
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 如果未指定格式，使用默认的 datetime 格式
	if args.Format == "" {
		args.Format = "datetime" // 默认格式
	}

	// 获取当前时间
	now := time.Now()
	var message string     // 用户友好的消息
	var value interface{}  // 格式化后的时间值

	// 根据请求的格式生成相应的时间信息
	switch args.Format {
	case "date":  // 仅日期格式
		value = now.Format("2006年01月02日")           // 中文日期格式
		message = fmt.Sprintf("📅 今天是：%s", value)    // 友好的日期消息
	case "time":  // 仅时间格式
		value = now.Format("15:04:05")               // 24小时制时间格式
		message = fmt.Sprintf("⏰ 当前时间：%s", value)   // 友好的时间消息
	case "datetime":  // 完整日期时间格式
		value = now.Format("2006年01月02日 15:04:05")   // 中文日期时间格式
		message = fmt.Sprintf("📅⏰ 当前日期时间：%s", value) // 友好的日期时间消息
	case "timestamp":  // Unix时间戳格式
		value = now.Unix()                           // Unix时间戳（秒）
		message = fmt.Sprintf("🕐 时间戳：%d", value)    // 友好的时间戳消息
	default:  // 默认情况，使用 datetime 格式
		value = now.Format("2006年01月02日 15:04:05")
		message = fmt.Sprintf("📅⏰ 当前日期时间：%s", value)
	}

	// 构建返回结果，包含格式、时间值和用户友好的消息
	result := map[string]interface{}{
		"format":  args.Format, // 请求的时间格式
		"value":   value,       // 格式化后的时间值
		"message": message,     // 友好的回复消息
	}

	// 将结果序列化为 JSON 格式返回给 React Agent
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// =============================================================================
// React Agent 系统组件 - 消息处理和代理构建
// =============================================================================

// messageModifier 消息修改器函数
// 为输入的消息列表添加系统提示，定义 React Agent 的角色和能力
// 这个系统提示帮助 LLM 理解自己的身份、可用工具和工作方式
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//   - msgs: 原始消息列表，包含用户输入和历史对话
// 返回值:
//   - []*schema.Message: 添加了系统提示的完整消息列表
//   - error: 处理过程中的错误信息
func messageModifier(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
	// 定义系统提示消息，描述 React Agent 的角色、能力和行为准则
	systemPrompt := &schema.Message{
		Role: schema.System, // 系统角色消息，用于设定 AI 助手的行为模式
		Content: `你是一个智能助手，名字叫小艾诺（Eino）。你拥有以下能力：

1. 🌤️ 天气查询：可以查询中国主要城市的天气情况
2. 🔢 数学计算：可以进行基本的四则运算
3. ⏰ 时间查询：可以获取当前的日期和时间信息

请根据用户的问题，合理使用工具来提供准确的答案。如果用户的问题需要使用多个工具，请按步骤逐一使用。

注意：
- 回答要简洁明了
- 对于天气查询，目前只支持北京、上海、广州、深圳、杭州、成都、西安、南京等城市
- 对于数学计算，目前只支持简单的四则运算（+、-、*、/）
- 始终保持友好和乐于助人的态度`, // 详细的系统提示内容，定义助手的能力范围和行为准则
	}

	// 检查消息列表是否已包含系统消息，避免重复添加
	// 如果第一条消息不是系统消息，则添加系统消息到列表开头
	if len(msgs) == 0 || msgs[0].Role != schema.System {
		return append([]*schema.Message{systemPrompt}, msgs...), nil
	}

	// 如果已存在系统消息，直接返回原消息列表
	return msgs, nil
}

// =============================================================================
// 主函数 - React Agent 演示程序入口
// =============================================================================

// main 主函数，React Agent 演示程序的入口点
// 功能流程:
// 1. 加载配置文件，获取 ARK API 密钥等配置信息
// 2. 创建和配置聊天模型（ARK 大语言模型）
// 3. 初始化工具集合（天气、计算、时间查询工具）
// 4. 构建 React Agent 链，集成工具和消息处理器
// 5. 执行测试用例，演示 React Agent 的各项功能
func main() {
	// 配置文件初始化 - 使用 Viper 库管理应用配置
	viper.SetConfigName("config")  // 配置文件名（不含扩展名），默认查找 config.yaml
	viper.SetConfigType("yaml")    // 配置文件格式，支持 YAML 格式的配置文件
	viper.AddConfigPath(".")       // 配置文件搜索路径，当前目录

	// 读取配置文件，获取 API 密钥等必要配置信息
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err) // 配置文件读取失败时终止程序
	}

	// 环境变量设置 - 将配置文件中的 API 密钥设置到环境变量中
	// 这样做是为了确保 ARK SDK 能够正确获取到 API 密钥
	os.Setenv("ARK_API_KEY", viper.GetString("ARK_API_KEY"))

	// 创建上下文对象，用于控制整个程序的生命周期和请求管理
	ctx := context.Background()

	// ARK 聊天模型配置 - React Agent 的核心推理引擎配置
	// 配置 ARK 大语言模型的基本参数，包括模型名称和 API 密钥
	config := &ark.ChatModelConfig{
		Model:  viper.GetString("ARK_MODEL"),   // 模型名称，从配置文件获取（如 doubao-pro-4k）
		APIKey: viper.GetString("ARK_API_KEY"), // API 密钥，从配置文件获取，用于身份验证
	}

	// ARK 聊天模型创建 - React Agent 的核心推理引擎
	// 使用字节跳动的 ARK 大语言模型作为 React Agent 的推理核心
	// 该模型负责理解用户意图、选择合适工具、生成回复内容
	chatModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		log.Fatalf("创建聊天模型失败: %v", err)
	}

	// 工具集合初始化 - React Agent 的执行能力组件
	// 创建所有可用工具的实例，这些工具为 React Agent 提供具体的执行能力
	// 每个工具都实现了 tool.InvokableTool 接口，可被 LLM 调用执行特定任务
	tools := []tool.InvokableTool{
		&WeatherTool{},    // 天气查询工具，提供城市天气信息查询功能
		&CalculatorTool{}, // 数学计算工具，提供基本四则运算能力
		&TimeTool{},       // 时间查询工具，提供日期时间信息获取功能
	}

	// 工具信息收集和绑定 - 让 LLM 了解可用工具的功能和使用方法
	// 遍历所有工具，获取其元数据信息（名称、描述、参数规范等）
	// 这些信息帮助 LLM 理解何时以及如何使用每个工具
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, tool := range tools {
		// 获取工具的元数据信息，包括工具名称、功能描述和参数要求
		info, err := tool.Info(ctx)
		if err != nil {
			log.Printf("获取工具信息失败: %v", err)
			continue // 跳过有问题的工具，继续处理其他工具
		}
		toolInfos = append(toolInfos, info)
		// 记录成功绑定的工具信息，便于调试和监控
		log.Printf("绑定工具: %s - %s", info.Name, info.Desc)
	}

	// 工具绑定到聊天模型 - 使 LLM 能够感知和调用工具
	// 将收集到的工具信息绑定到聊天模型，使其能够在推理过程中选择和调用合适的工具
	// 这是 React Agent 实现工具调用能力的关键步骤
	chatModel.BindTools(toolInfos)

	// React Agent 链构建 - 创建支持多轮对话的工具调用处理链
	// 使用更强化的链结构确保工具调用和响应生成的可靠性
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	
	// 第一层：理解用户意图并决定工具调用
	chain.AppendChatModel(chatModel, compose.WithNodeName("intent_analysis"))
	
	// 第二层：执行工具调用并生成工具响应
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
		// 记录调试信息
		log.Printf("工具执行阶段 - 消息角色: %s, 工具调用数量: %d", msg.Role, len(msg.ToolCalls))
		
		// 检查是否有工具调用需要处理
		if msg.ToolCalls != nil && len(msg.ToolCalls) > 0 {
			var allMessages []*schema.Message
			
			// 保持原始助手消息（包含工具调用）
			allMessages = append(allMessages, msg)
			
			// 处理每个工具调用
			for _, toolCall := range msg.ToolCalls {
				var toolResponse string
				var err error
				
				log.Printf("执行工具: %s, 参数: %s", toolCall.Function.Name, toolCall.Function.Arguments)
				
				// 根据工具名称执行对应的工具
				switch toolCall.Function.Name {
				case "get_weather":
					weatherTool := &WeatherTool{}
					toolResponse, err = weatherTool.InvokableRun(ctx, toolCall.Function.Arguments)
				case "calculate":
					calcTool := &CalculatorTool{}
					toolResponse, err = calcTool.InvokableRun(ctx, toolCall.Function.Arguments)
				case "get_time":
					timeTool := &TimeTool{}
					toolResponse, err = timeTool.InvokableRun(ctx, toolCall.Function.Arguments)
				default:
					toolResponse = `{"error": "未知工具: ` + toolCall.Function.Name + `"}`
					err = fmt.Errorf("未知工具: %s", toolCall.Function.Name)
				}
				
				// 处理工具执行错误
				if err != nil {
					toolResponse = fmt.Sprintf(`{"error": "工具执行失败", "details": "%s"}`, err.Error())
					log.Printf("工具执行失败: %v", err)
				} else {
					log.Printf("工具执行成功: %s", toolResponse)
				}
				
				// 创建工具响应消息
				toolMessage := &schema.Message{
					Role:       schema.Tool,
					Content:    toolResponse,
					ToolCallID: toolCall.ID,
				}
				allMessages = append(allMessages, toolMessage)
			}
			
			return allMessages, nil
		}
		
		// 如果没有工具调用，保持消息不变
		log.Printf("无工具调用，直接传递消息")
		return []*schema.Message{msg}, nil
	}), compose.WithNodeName("tool_executor"))
	
	// 第三层：基于工具结果生成最终回复
	chain.AppendChatModel(chatModel, compose.WithNodeName("response_generator"))
	
	// 第四层：确保输出格式标准化
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
		log.Printf("最终输出 - 角色: %s, 内容长度: %d", msg.Role, len(msg.Content))
		
		// 确保我们有实际的响应内容
		if msg.Role == schema.Assistant && (msg.Content == "" || msg.Content == " ") {
			// 如果内容为空，尝试重新生成
			log.Printf("检测到空响应，尝试重新生成...")
			return []*schema.Message{
				{
					Role:    schema.Assistant,
					Content: "抱歉，我在处理您的请求时遇到了问题，请您稍后重试。",
				},
			}, nil
		}
		
		return []*schema.Message{msg}, nil
	}), compose.WithNodeName("output_validator"))

	// React Agent 编译 - 将链组装成可执行的代理
	// 编译过程会验证链的完整性、优化执行流程并生成最终的可执行代理
	// 编译后的代理可以处理用户请求并返回结果
	compiledChain, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译链失败: %v", err)
	}

	// 程序启动信息输出 - 向用户展示 React Agent 的功能和状态
	fmt.Println("🤖 小艾诺（Eino）Chain Agent Demo 启动成功！")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("我可以帮助您：")
	fmt.Println("🌤️  查询天气：例如 '北京今天天气怎么样？'")
	fmt.Println("🔢 数学计算：例如 '计算 25 + 17'")
	fmt.Println("⏰ 时间查询：例如 '现在几点了？'")
	fmt.Println("🔄 组合操作：例如 '查询上海天气，然后告诉我现在时间'")
	fmt.Println(strings.Repeat("=", 50))

	// 测试用例定义 - 演示 React Agent 的各项功能
	// 每个测试用例都设计来验证特定的工具调用能力
	// 包括单一工具使用和复合工具使用场景，全面测试 React Agent 的推理和执行能力
	testCases := []string{
		"北京今天天气怎么样？",                    // 天气查询测试：触发 get_weather 工具调用
		"帮我计算 125 + 237",                   // 数学计算测试：触发 calculate 工具调用
		"现在几点了？",                         // 时间查询测试：触发 get_time 工具调用
		"上海的天气如何，另外帮我算一下 15 * 8",      // 复合查询测试：同时触发天气和计算工具
		"查询深圳天气，然后告诉我今天的日期",         // 多步骤测试：天气查询 + 日期查询的组合操作
	}

	// 测试用例执行循环 - 逐一执行所有预定义的测试场景
	// 每个测试用例都会创建完整的消息上下文，调用 React Agent 并展示结果
	for i, testCase := range testCases {
		// 输出测试用例信息，便于跟踪执行进度和调试
		fmt.Printf("\n📝 测试用例 %d: %s\n", i+1, testCase)
		fmt.Println(strings.Repeat("-", 40))

		// 系统消息构建 - 为 React Agent 设定角色和能力范围
		// 系统消息定义了 AI 助手的身份、可用工具和行为准则
		// 这是 React Agent 理解自身能力和工作方式的重要指导信息
		systemMessage := &schema.Message{
			Role: schema.System, // 系统角色，用于设定 AI 助手的行为模式
			Content: `你是一个智能助手，名字叫小艾诺（Eino）。你拥有以下能力：

1. 🌤️ 天气查询：可以查询中国主要城市的天气情况
2. 🔢 数学计算：可以进行基本的四则运算
3. ⏰ 时间查询：可以获取当前的日期和时间信息

请根据用户的问题，合理使用工具来提供准确的答案。回答要简洁明了，保持友好和乐于助人的态度。`,
		}

		// 用户消息构建 - 将测试用例封装为标准的用户消息格式
		// 用户消息包含具体的查询内容，是 React Agent 需要处理的核心任务
		userMessage := &schema.Message{
			Role:    schema.User, // 用户角色，标识消息来源为用户输入
			Content: testCase,    // 测试用例内容，即用户的具体查询请求
		}

		// 消息上下文组装 - 创建完整的对话上下文
		// 将系统消息和用户消息组合成完整的消息列表，为 React Agent 提供完整的执行上下文
		messages := []*schema.Message{systemMessage, userMessage}
		
		// React Agent 执行调用 - 核心的推理和工具调用过程
		// 将消息传递给编译后的 React Agent，触发完整的推理流程：
		// 1. 理解用户意图 2. 选择合适工具 3. 执行工具调用 4. 生成最终回复
		results, err := compiledChain.Invoke(ctx, messages)
		if err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
			continue // 跳过失败的测试用例，继续执行后续测试
		}

		// 结果输出展示 - 向用户展示 React Agent 的处理结果
		// 检查返回结果的有效性，确保能够正确显示 Agent 的回复内容
		if results != nil && len(results) > 0 {
			// 取最后一条消息作为最终回复
			lastMessage := results[len(results)-1]
			if lastMessage != nil && lastMessage.Content != "" {
				// 输出 Agent 经过推理和工具调用后生成的最终回复内容
				fmt.Printf("🤖 小艾诺: %s\n", lastMessage.Content)
			} else {
				// 处理异常情况：Agent 没有返回有效内容时的提示信息
				fmt.Printf("🤖 小艾诺: [最后一条消息内容为空]\n")
				fmt.Printf("Debug: lastMessage=%+v\n", lastMessage) // 调试信息，帮助排查问题
			}
		} else {
			// 处理异常情况：Agent 没有返回消息时的提示信息
			fmt.Printf("🤖 小艾诺: [没有收到回复消息]\n")
			fmt.Printf("Debug: results=%+v\n", results) // 调试信息，帮助排查问题
		}

		// 执行间隔控制 - 在测试用例之间添加适当的延迟
		// 等待一秒钟，让输出更清晰，便于观察每个测试用例的执行结果
		time.Sleep(1 * time.Second)
	}

	// 程序结束信息输出 - 向用户展示测试完成状态和后续建议
	fmt.Println("\n✅ 所有测试用例执行完成！")
	fmt.Println("\n💡 您可以修改 testCases 数组来测试更多场景。")
	fmt.Println("📚 如需了解更多功能，请参考 README.md 文档。")
}
