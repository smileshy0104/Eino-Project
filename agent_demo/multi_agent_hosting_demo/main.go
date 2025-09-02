/*
Multi-Agent Hosting Demo - 多智能体托管演示程序

本程序演示了基于 Eino 框架的多智能体协作系统，展示了如何构建和管理多个专业化的 AI 智能体。

系统架构：
1. Host Agent（主控智能体）：负责理解用户意图，决定将任务分配给哪个专业智能体
2. Weather Specialist（天气专家）：专门处理天气相关查询
3. Calculator Specialist（计算专家）：专门处理数学计算任务
4. Time Specialist（时间专家）：专门处理时间日期相关查询
5. Multi-Agent Router（多智能体路由器）：协调各个智能体之间的协作

核心特性：
- 智能体专业化分工：每个智能体专注于特定领域，提高处理效率和准确性
- 动态任务分配：Host Agent 根据用户查询内容智能选择合适的专业智能体
- 协作式处理：支持复杂查询的多智能体协同处理
- 统一接口：通过路由器提供统一的用户交互接口

技术实现：
- 使用 Eino 框架构建智能体链
- 集成字节跳动 ARK 大语言模型
- 实现工具调用和消息传递机制
- 支持上下文管理和状态维护
*/
package main

import (
	"context"       // 上下文管理，用于控制请求生命周期和取消操作
	"encoding/json" // JSON 编解码，用于工具参数和返回值的序列化
	"fmt"           // 格式化输入输出，用于控制台信息展示
	"log"           // 日志记录，用于错误处理和调试信息输出
	"os"            // 操作系统接口，用于环境变量设置
	"strconv"       // 字符串转换，用于数值类型转换操作
	"strings"       // 字符串处理，用于文本操作和格式化
	"time"          // 时间处理，用于时间相关功能和延迟控制

	// Eino 框架核心组件
	"github.com/cloudwego/eino-ext/components/model/ark" // ARK 模型组件，字节跳动大语言模型接口
	"github.com/cloudwego/eino/components/tool"          // 工具组件，提供工具调用能力
	"github.com/cloudwego/eino/compose"                  // 组合器，用于构建智能体处理链
	"github.com/cloudwego/eino/schema"                   // 模式定义，定义消息和工具的数据结构
	"github.com/spf13/viper"                             // 配置管理，用于读取配置文件和环境变量
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
// WeatherTool 天气查询工具 - 多智能体系统中的天气专业工具
// 为天气专家智能体提供城市天气信息查询能力
// 实现 tool.InvokableTool 接口，支持被 LLM 调用执行天气查询任务
type WeatherTool struct{}

// Info 方法 - 提供天气工具的元数据信息
// 返回工具的名称、功能描述和参数规范，帮助 LLM 理解如何使用此工具
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//
// 返回值:
//   - *schema.ToolInfo: 工具元数据信息
//   - error: 错误信息（如果有）
func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather", // 工具名称，LLM 调用时使用的标识符
		Desc: "查询指定城市的天气情况", // 工具功能描述，指导 LLM 何时使用此工具
		// 参数规范定义 - 指定工具接受的参数类型和要求
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",                  // 参数类型：字符串
				Desc:     "要查询天气的城市名称，例如：北京、上海、广州等", // 参数描述和示例
				Required: true,                      // 必需参数标识
			},
		}),
	}, nil
}

// InvokableRun 方法 - 执行天气查询的核心逻辑
// 接收 LLM 传递的参数，执行天气查询并返回结构化的天气信息
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//   - argumentsInJSON: JSON 格式的参数字符串，包含城市名称
//   - opts: 可选的工具执行选项
//
// 返回值:
//   - string: JSON 格式的天气信息结果
//   - error: 错误信息（如果有）
func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体 - 用于解析 JSON 参数
	var args struct {
		City string `json:"city"` // 城市名称字段
	}

	// 参数解析 - 将 JSON 字符串解析为 Go 结构体
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 记录查询日志 - 便于调试和监控
	log.Printf("[WeatherTool] 查询城市: %s", args.City)

	// 模拟天气数据库 - 预定义的城市天气信息
	// 在实际应用中，这里会调用真实的天气 API 服务
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

	// 天气信息查询 - 从数据库中获取对应城市的天气
	weather, exists := weatherData[args.City]
	if !exists {
		// 处理未知城市 - 提供默认天气信息
		weather = "晴，25°C，微风（默认天气）"
	}

	// 构建返回结果 - 创建结构化的天气信息响应
	result := map[string]interface{}{
		"city":    args.City,                                        // 查询的城市名称
		"weather": weather,                                          // 天气详细信息
		"message": fmt.Sprintf("🌤️ %s今天的天气：%s", args.City, weather), // 用户友好的描述信息
	}

	// JSON 序列化 - 将结果转换为 JSON 字符串返回给 LLM
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// CalculatorTool 数学计算工具 - 多智能体系统中的数学计算专业工具
// 为计算专家智能体提供数学运算能力，支持基本的四则运算
// 实现 tool.InvokableTool 接口，支持被 LLM 调用执行数学计算任务
type CalculatorTool struct{}

// Info 方法 - 提供计算工具的元数据信息
// 返回工具的名称、功能描述和参数规范，帮助 LLM 理解如何使用此工具
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//
// 返回值:
//   - *schema.ToolInfo: 工具元数据信息
//   - error: 错误信息（如果有）
func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculate",        // 工具名称，LLM 调用时使用的标识符
		Desc: "进行数学计算，支持基本的四则运算", // 工具功能描述，指导 LLM 何时使用此工具
		// 参数规范定义 - 指定工具接受的参数类型和要求
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {
				Type:     "string",                      // 参数类型：字符串
				Desc:     "数学表达式，例如：10+5、20-8、6*7、15/3", // 参数描述和示例
				Required: true,                          // 必需参数标识
			},
		}),
	}, nil
}

// InvokableRun 方法 - 执行数学计算的核心逻辑
// 接收 LLM 传递的数学表达式，执行计算并返回结果
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//   - argumentsInJSON: JSON 格式的参数字符串，包含数学表达式
//   - opts: 可选的工具执行选项
//
// 返回值:
//   - string: JSON 格式的计算结果
//   - error: 错误信息（如果有）
func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体 - 用于解析 JSON 参数
	var args struct {
		Expression string `json:"expression"` // 数学表达式字段
	}

	// 参数解析 - 将 JSON 字符串解析为 Go 结构体
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 记录计算日志 - 便于调试和监控
	log.Printf("[CalculatorTool] 计算表达式: %s", args.Expression)

	// 执行数学计算 - 调用简单计算函数处理表达式
	resultValue, err := simpleCalculate(args.Expression)
	if err != nil {
		return "", fmt.Errorf("计算错误: %v", err)
	}

	// 构建返回结果 - 创建结构化的计算结果响应
	result := map[string]interface{}{
		"expression": args.Expression,                                               // 原始数学表达式
		"result":     resultValue,                                                   // 计算结果
		"message":    fmt.Sprintf("🔢 计算结果：%s = %.2f", args.Expression, resultValue), // 用户友好的描述信息
	}

	// JSON 序列化 - 将结果转换为 JSON 字符串返回给 LLM
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// TimeTool 时间查询工具 - 多智能体系统中的时间专业工具
// 为时间专家智能体提供日期时间信息查询能力，支持多种时间格式
// 实现 tool.InvokableTool 接口，支持被 LLM 调用执行时间查询任务
type TimeTool struct{}

// Info 方法 - 提供时间工具的元数据信息
// 返回工具的名称、功能描述和参数规范，帮助 LLM 理解如何使用此工具
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//
// 返回值:
//   - *schema.ToolInfo: 工具元数据信息
//   - error: 错误信息（如果有）
func (t *TimeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_time",     // 工具名称，LLM 调用时使用的标识符
		Desc: "获取当前的日期和时间信息", // 工具功能描述，指导 LLM 何时使用此工具
		// 参数规范定义 - 指定工具接受的参数类型和要求
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"format": {
				Type:     "string",                                                 // 参数类型：字符串
				Desc:     "时间格式：date（仅日期）、time（仅时间）、datetime（日期时间）、timestamp（时间戳）", // 参数描述和可选值
				Required: false,                                                    // 可选参数，有默认值
				Enum:     []string{"date", "time", "datetime", "timestamp"},        // 枚举值限制
			},
		}),
	}, nil
}

// InvokableRun 方法 - 执行时间查询的核心逻辑
// 接收 LLM 传递的格式参数，获取当前时间并按指定格式返回
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期
//   - argumentsInJSON: JSON 格式的参数字符串，包含时间格式要求
//   - opts: 可选的工具执行选项
//
// 返回值:
//   - string: JSON 格式的时间信息结果
//   - error: 错误信息（如果有）
func (t *TimeTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体 - 用于解析 JSON 参数
	var args struct {
		Format string `json:"format"` // 时间格式字段
	}

	// 参数解析 - 将 JSON 字符串解析为 Go 结构体
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 设置默认格式 - 如果未指定格式，使用完整的日期时间格式
	if args.Format == "" {
		args.Format = "datetime"
	}

	// 记录查询日志 - 便于调试和监控
	log.Printf("[TimeTool] 查询时间格式: %s", args.Format)

	// 获取当前时间 - 系统当前时间
	now := time.Now()
	var message string    // 用户友好的描述信息
	var value interface{} // 格式化后的时间值

	// 时间格式处理 - 根据请求的格式返回相应的时间表示
	switch args.Format {
	case "date":
		// 仅日期格式 (中文格式)
		value = now.Format("2006年01月02日")
		message = fmt.Sprintf("📅 今天是：%s", value)
	case "time":
		// 仅时间格式 (24小时制)
		value = now.Format("15:04:05")
		message = fmt.Sprintf("⏰ 当前时间：%s", value)
	case "datetime":
		// 完整日期时间格式 (中文格式)
		value = now.Format("2006年01月02日 15:04:05")
		message = fmt.Sprintf("📅⏰ 当前日期时间：%s", value)
	case "timestamp":
		// Unix 时间戳格式
		value = now.Unix()
		message = fmt.Sprintf("🕐 时间戳：%d", value)
	default:
		// 默认情况 - 使用完整日期时间格式
		value = now.Format("2006年01月02日 15:04:05")
		message = fmt.Sprintf("📅⏰ 当前日期时间：%s", value)
	}

	// 构建返回结果 - 创建结构化的时间信息响应
	result := map[string]interface{}{
		"format":  args.Format, // 请求的时间格式
		"value":   value,       // 格式化后的时间值
		"message": message,     // 用户友好的描述信息
	}

	// JSON 序列化 - 将结果转换为 JSON 字符串返回给 LLM
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// simpleCalculate 简单数学计算函数 - 多智能体系统的计算引擎
// 解析和执行基本的数学表达式，支持四则运算（加减乘除）
// 这是一个简化的计算器实现，用于演示目的
// 参数:
//   - expression: 数学表达式字符串，如 "10+5"、"20*3"、"100/4" 等
//
// 返回值:
//   - float64: 计算结果
//   - error: 错误信息（如果有）
func simpleCalculate(expression string) (float64, error) {
	// 预处理 - 移除表达式中的所有空格，统一格式
	expression = strings.ReplaceAll(expression, " ", "")

	// 加法运算处理 - 检测并执行加法操作
	if strings.Contains(expression, "+") {
		// 按加号分割表达式
		parts := strings.Split(expression, "+")
		if len(parts) == 2 {
			// 解析操作数
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				// 执行加法运算
				return a + b, nil
			}
		}
	}

	// 减法运算处理 - 检测并执行减法操作（排除负数情况）
	if strings.Contains(expression, "-") && !strings.HasPrefix(expression, "-") {
		// 按减号分割表达式
		parts := strings.Split(expression, "-")
		if len(parts) == 2 {
			// 解析操作数
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				// 执行减法运算
				return a - b, nil
			}
		}
	}

	// 乘法运算处理 - 检测并执行乘法操作
	if strings.Contains(expression, "*") {
		// 按乘号分割表达式
		parts := strings.Split(expression, "*")
		if len(parts) == 2 {
			// 解析操作数
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				// 执行乘法运算
				return a * b, nil
			}
		}
	}

	// 除法运算处理 - 检测并执行除法操作
	if strings.Contains(expression, "/") {
		// 按除号分割表达式
		parts := strings.Split(expression, "/")
		if len(parts) == 2 {
			// 解析操作数
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				// 除零检查 - 防止除以零的数学错误
				if b == 0 {
					return 0, fmt.Errorf("除数不能为零")
				}
				// 执行除法运算
				return a / b, nil
			}
		}
	}

	// 单个数字处理 - 如果表达式是单个数字，直接返回
	if num, err := strconv.ParseFloat(expression, 64); err == nil {
		return num, nil
	}

	// 不支持的表达式格式 - 返回错误信息
	return 0, fmt.Errorf("不支持的表达式格式")
}

// createChatModel 创建 ARK 聊天模型实例 - 多智能体系统的核心推理引擎
// 为每个智能体创建独立的 ARK 大语言模型实例，支持工具绑定和并发安全
// 这是多智能体系统中所有智能体的基础组件，提供自然语言理解和生成能力
// 参数:
//   - ctx: 上下文对象，用于控制模型创建过程的生命周期
//   - tools: 可调用工具列表，为智能体提供外部能力扩展
//
// 返回值:
//   - *ark.ChatModel: 配置完成的 ARK 聊天模型实例
//   - error: 错误信息（如果有）
func createChatModel(ctx context.Context, tools []tool.InvokableTool) (*ark.ChatModel, error) {
	// 构建 ARK 模型配置 - 使用配置文件中的参数
	// ARK 是字节跳动提供的大语言模型服务，支持多种模型规格
	config := &ark.ChatModelConfig{
		Model:  viper.GetString("ARK_MODEL"),   // 模型名称，从配置文件获取
		APIKey: viper.GetString("ARK_API_KEY"), // API 密钥，从配置文件获取
	}

	// 创建 ARK 聊天模型实例 - 建立与 ARK 服务的连接
	chatModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建聊天模型失败: %v", err)
	}

	// 工具绑定处理 - 为智能体提供外部能力扩展
	// 只有专业智能体需要绑定工具，主控智能体通常不需要
	if len(tools) > 0 {
		// 收集工具元数据信息 - 获取每个工具的描述和参数规范
		toolInfos := make([]*schema.ToolInfo, 0, len(tools))
		for _, tool := range tools {
			// 获取工具的详细信息，包括名称、描述和参数定义
			info, err := tool.Info(ctx)
			if err != nil {
				// 工具信息获取失败时记录日志但继续处理其他工具
				log.Printf("获取工具信息失败: %v", err)
				continue
			}
			toolInfos = append(toolInfos, info)
		}
		// 将工具信息绑定到聊天模型，使 LLM 能够理解和调用这些工具
		chatModel.BindTools(toolInfos)
	}

	return chatModel, nil
}

// createHostAgent 创建主控智能体 - 多智能体系统的协调中心和决策大脑
// Host Agent 是整个多智能体系统的核心组件，负责理解用户意图、分析任务需求
// 并决定将任务分配给哪个或哪些专业智能体进行处理
// 它不直接执行具体的业务任务，而是作为智能路由器和任务协调器
// 参数:
//   - ctx: 上下文对象，用于控制智能体创建过程的生命周期
//
// 返回值:
//   - *ark.ChatModel: 配置完成的主控智能体模型实例
//   - error: 错误信息（如果有）
func createHostAgent(ctx context.Context) (*ark.ChatModel, error) {
	// 记录主控智能体创建过程 - 便于系统监控和调试
	log.Println("创建Host Agent...")

	// 创建主控智能体的聊天模型 - 不绑定任何工具
	// Host Agent 专注于意图理解和决策，不需要直接调用外部工具
	// 所有的工具调用都由专业智能体负责执行
	hostModel, err := createChatModel(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("创建Host Agent失败: %v", err)
	}

	// 记录创建成功状态 - 确认主控智能体已就绪
	log.Println("Host Agent创建成功")
	return hostModel, nil
}

// createWeatherSpecialist 创建天气专家智能体 - 专门处理天气相关查询的专业智能体
// 天气专家配备了 WeatherTool，能够查询各个城市的天气信息
// 它专注于天气领域，提供准确、专业的天气查询服务
// 参数:
//   - ctx: 上下文对象，用于控制智能体创建过程的生命周期
//
// 返回值:
//   - compose.Runnable[[]*schema.Message, *schema.Message]: 编译完成的天气专家智能体链
//   - error: 错误信息（如果有）
func createWeatherSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
	// 记录天气专家创建过程 - 便于系统监控和调试
	log.Println("创建Weather Specialist Agent...")

	// 配置天气专家的专业工具集 - 仅包含天气查询工具
	// 专业化分工确保智能体专注于特定领域，提高处理效率和准确性
	tools := []tool.InvokableTool{&WeatherTool{}}

	// 创建配备天气工具的聊天模型
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, fmt.Errorf("创建Weather Specialist失败: %v", err)
	}

	// 创建支持工具调用的处理链 - 构建天气专家的完整执行流程
	// 使用类似 React Agent 的模式：ChatModel -> Tool Executor -> ChatModel
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()

	// 第一层：理解意图并生成工具调用
	chain.AppendChatModel(chatModel, compose.WithNodeName("weather_intent_analysis"))

	// 第二层：执行工具调用
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
		// 检查是否有工具调用
		if len(msg.ToolCalls) > 0 {
			var allMessages []*schema.Message
			allMessages = append(allMessages, msg)

			// 处理每个工具调用
			for _, toolCall := range msg.ToolCalls {
				var toolResponse string
				var err error

				// 执行天气工具
				if toolCall.Function.Name == "get_weather" {
					weatherTool := &WeatherTool{}
					toolResponse, err = weatherTool.InvokableRun(ctx, toolCall.Function.Arguments)
				} else {
					toolResponse = `{"error": "未知工具"}`
					err = fmt.Errorf("未知工具: %s", toolCall.Function.Name)
				}

				if err != nil {
					toolResponse = fmt.Sprintf(`{"error": "工具执行失败", "details": "%s"}`, err.Error())
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

		return []*schema.Message{msg}, nil
	}), compose.WithNodeName("weather_tool_executor"))

	// 第三层：基于工具结果生成最终回复
	chain.AppendChatModel(chatModel, compose.WithNodeName("weather_response_generator"))

	// 编译处理链 - 将配置转换为可执行的智能体实例
	specialist, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译Weather Specialist失败: %v", err)
	}

	// 记录创建成功状态 - 确认天气专家已就绪
	log.Println("Weather Specialist Agent创建成功")
	return specialist, nil
}

// createCalculatorSpecialist 创建计算专家智能体 - 专门处理数学计算任务的专业智能体
// 计算专家配备了 CalculatorTool，能够执行各种数学运算
// 它专注于数学计算领域，提供准确、快速的计算服务
// 参数:
//   - ctx: 上下文对象，用于控制智能体创建过程的生命周期
//
// 返回值:
//   - compose.Runnable[[]*schema.Message, *schema.Message]: 编译完成的计算专家智能体链
//   - error: 错误信息（如果有）
func createCalculatorSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
	// 记录计算专家创建过程 - 便于系统监控和调试
	log.Println("创建Calculator Specialist Agent...")

	// 配置计算专家的专业工具集 - 仅包含数学计算工具
	// 专业化分工确保智能体专注于数学计算，提供精确的运算结果
	tools := []tool.InvokableTool{&CalculatorTool{}}

	// 创建配备计算工具的聊天模型
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, fmt.Errorf("创建Calculator Specialist失败: %v", err)
	}

	// 创建支持工具调用的处理链 - 构建计算专家的完整执行流程
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()

	// 第一层：理解意图并生成工具调用
	chain.AppendChatModel(chatModel, compose.WithNodeName("calculator_intent_analysis"))

	// 第二层：执行工具调用
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
		// 检查是否有工具调用
		if len(msg.ToolCalls) > 0 {
			var allMessages []*schema.Message
			allMessages = append(allMessages, msg)

			// 处理每个工具调用
			for _, toolCall := range msg.ToolCalls {
				var toolResponse string
				var err error

				// 执行计算工具
				if toolCall.Function.Name == "calculate" {
					calcTool := &CalculatorTool{}
					toolResponse, err = calcTool.InvokableRun(ctx, toolCall.Function.Arguments)
				} else {
					toolResponse = `{"error": "未知工具"}`
					err = fmt.Errorf("未知工具: %s", toolCall.Function.Name)
				}

				if err != nil {
					toolResponse = fmt.Sprintf(`{"error": "工具执行失败", "details": "%s"}`, err.Error())
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

		return []*schema.Message{msg}, nil
	}), compose.WithNodeName("calculator_tool_executor"))

	// 第三层：基于工具结果生成最终回复
	chain.AppendChatModel(chatModel, compose.WithNodeName("calculator_response_generator"))

	// 编译处理链 - 将配置转换为可执行的智能体实例
	specialist, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译Calculator Specialist失败: %v", err)
	}

	// 记录创建成功状态 - 确认计算专家已就绪
	log.Println("Calculator Specialist Agent创建成功")
	return specialist, nil
}

// createTimeSpecialist 创建时间专家智能体 - 专门处理时间日期查询的专业智能体
// 时间专家配备了 TimeTool，能够提供各种格式的时间信息
// 它专注于时间日期领域，提供准确、多样化的时间查询服务
// 参数:
//   - ctx: 上下文对象，用于控制智能体创建过程的生命周期
//
// 返回值:
//   - compose.Runnable[[]*schema.Message, *schema.Message]: 编译完成的时间专家智能体链
//   - error: 错误信息（如果有）
func createTimeSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
	// 记录时间专家创建过程 - 便于系统监控和调试
	log.Println("创建Time Specialist Agent...")

	// 配置时间专家的专业工具集 - 仅包含时间查询工具
	// 专业化分工确保智能体专注于时间处理，提供多格式的时间信息
	tools := []tool.InvokableTool{&TimeTool{}}

	// 创建配备时间工具的聊天模型
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, fmt.Errorf("创建Time Specialist失败: %v", err)
	}

	// 创建支持工具调用的处理链 - 构建时间专家的完整执行流程
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()

	// 第一层：理解意图并生成工具调用
	chain.AppendChatModel(chatModel, compose.WithNodeName("time_intent_analysis"))

	// 第二层：执行工具调用
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
		// 检查是否有工具调用
		if len(msg.ToolCalls) > 0 {
			var allMessages []*schema.Message
			allMessages = append(allMessages, msg)

			// 处理每个工具调用
			for _, toolCall := range msg.ToolCalls {
				var toolResponse string
				var err error

				// 执行时间工具
				if toolCall.Function.Name == "get_time" {
					timeTool := &TimeTool{}
					toolResponse, err = timeTool.InvokableRun(ctx, toolCall.Function.Arguments)
				} else {
					toolResponse = `{"error": "未知工具"}`
					err = fmt.Errorf("未知工具: %s", toolCall.Function.Name)
				}

				if err != nil {
					toolResponse = fmt.Sprintf(`{"error": "工具执行失败", "details": "%s"}`, err.Error())
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

		return []*schema.Message{msg}, nil
	}), compose.WithNodeName("time_tool_executor"))

	// 第三层：基于工具结果生成最终回复
	chain.AppendChatModel(chatModel, compose.WithNodeName("time_response_generator"))

	// 编译处理链 - 将配置转换为可执行的智能体实例
	specialist, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译Time Specialist失败: %v", err)
	}

	// 记录创建成功状态 - 确认时间专家已就绪
	log.Println("Time Specialist Agent创建成功")
	return specialist, nil
}

// MultiAgentRouter 多智能体路由器 - 多智能体系统的核心协调组件
// 负责接收用户请求，通过 Host Agent 分析意图，然后将任务路由到合适的专业智能体
// 最后整合各个专业智能体的响应，为用户提供统一的服务接口
type MultiAgentRouter struct {
	hostAgent            *ark.ChatModel                                       // 主控智能体，负责意图分析和任务分发决策
	weatherSpecialist    compose.Runnable[[]*schema.Message, *schema.Message] // 天气专家智能体，处理天气相关查询
	calculatorSpecialist compose.Runnable[[]*schema.Message, *schema.Message] // 计算专家智能体，处理数学计算任务
	timeSpecialist       compose.Runnable[[]*schema.Message, *schema.Message] // 时间专家智能体，处理时间日期查询
}

// Invoke 处理用户请求的核心方法 - 多智能体协作的完整流程实现
// 该方法实现了多智能体系统的标准工作流程：意图分析 -> 任务路由 -> 专家执行 -> 结果整合
// 参数:
//   - ctx: 上下文对象，用于控制请求处理的生命周期
//   - input: 用户消息列表，包含完整的对话上下文
//
// 返回值:
//   - *schema.Message: 整合后的最终响应消息
//   - error: 错误信息（如果有）
func (m *MultiAgentRouter) Invoke(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
	// 第一阶段：Host Agent 意图分析和任务分发决策
	// 构建专门的提示词，指导 Host Agent 进行意图识别和专家选择
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

	// 为 Host Agent 构建专用的消息上下文
	hostInput := append([]*schema.Message{hostPrompt}, input...)

	// 调用 Host Agent 进行意图分析和决策
	hostResponse, err := m.hostAgent.Generate(ctx, hostInput)
	if err != nil {
		return nil, fmt.Errorf("Host Agent 分析失败: %v", err)
	}

	// 记录 Host Agent 的决策结果 - 便于系统监控和调试
	log.Printf("[Host Agent] 路由决策: %s", hostResponse.Content)

	// 解析 Host Agent 的决策结果 - 提取需要调用的专家类型
	specialists := strings.Split(strings.TrimSpace(hostResponse.Content), ",")

	// 收集所有专家的响应结果
	var results []string

	// 第二阶段：专业智能体任务执行
	// 根据 Host Agent 的决策，依次调用相应的专业智能体处理具体任务
	for _, specialist := range specialists {
		specialist = strings.TrimSpace(specialist) // 清理空格

		var result *schema.Message // 专家响应结果
		var err error              // 执行错误

		// 为专家智能体准备专用的输入消息 - 复制原始输入避免修改
		specialistInput := make([]*schema.Message, len(input))
		copy(specialistInput, input)

		// 专家智能体路由选择和执行 - 根据类型选择对应的专业智能体
		// 处理专家名称的变体（支持模糊匹配）
		if strings.Contains(specialist, "weather") {
			// 调用天气专家智能体
			log.Println("[Router] 调用 Weather Specialist")
			// 为天气专家添加专业化的系统提示
			weatherPrompt := &schema.Message{
				Role:    schema.System,
				Content: "你是天气查询专家。用户询问天气时，请使用天气工具获取信息并以友好的方式回复。",
			}
			specialistInput = append([]*schema.Message{weatherPrompt}, specialistInput...)
			result, err = m.weatherSpecialist.Invoke(ctx, specialistInput)

		} else if strings.Contains(specialist, "calculator") {
			// 调用计算专家智能体
			log.Println("[Router] 调用 Calculator Specialist")
			// 为计算专家添加专业化的系统提示
			calcPrompt := &schema.Message{
				Role:    schema.System,
				Content: "你是数学计算专家。用户需要计算时，请使用计算工具并清晰地展示计算过程和结果。",
			}
			specialistInput = append([]*schema.Message{calcPrompt}, specialistInput...)
			result, err = m.calculatorSpecialist.Invoke(ctx, specialistInput)

		} else if strings.Contains(specialist, "time") {
			// 调用时间专家智能体
			log.Println("[Router] 调用 Time Specialist")
			// 为时间专家添加专业化的系统提示
			timePrompt := &schema.Message{
				Role:    schema.System,
				Content: "你是时间查询专家。用户询问时间时，请使用时间工具获取准确的时间信息。",
			}
			specialistInput = append([]*schema.Message{timePrompt}, specialistInput...)
			result, err = m.timeSpecialist.Invoke(ctx, specialistInput)

		} else {
			// 处理未知的专家类型 - 记录日志并跳过
			log.Printf("[Router] 未知专家: %s", specialist)
			continue
		}

		// 专家执行结果处理
		if err != nil {
			// 专家调用失败时记录错误但继续处理其他专家
			log.Printf("[Router] 专家 %s 执行失败: %v", specialist, err)
			continue
		}

		// 收集有效的专家响应结果
		if result != nil && result.Content != "" {
			results = append(results, result.Content)
			log.Printf("[Router] 专家 %s 返回: %s", specialist, result.Content)
		}
	}

	// 第三阶段：结果整合和最终响应生成
	// 将所有专业智能体的响应整合成统一的最终回复
	if len(results) == 0 {
		// 处理没有任何专家成功响应的情况
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "抱歉，我无法处理您的请求。",
		}, nil
	}

	// 整合多个专家的响应 - 用换行分隔不同专家的回复
	finalContent := strings.Join(results, "\n\n")

	// 返回整合后的最终响应
	return &schema.Message{
		Role:    schema.Assistant, // 标识为助手回复
		Content: finalContent,     // 整合后的最终响应内容
	}, nil
}

// createMultiAgentSystem 创建多智能体系统的核心构建函数
// 该函数负责初始化和组装完整的多智能体协作系统，包括主控智能体和所有专业智能体
// 系统采用分层架构：Host Agent（决策层） + 专业智能体（执行层） + 路由器（协调层）
// 参数:
//   - ctx: 上下文对象，用于控制系统初始化过程的生命周期
//
// 返回值:
//   - *MultiAgentRouter: 完整的多智能体路由器实例
//   - error: 初始化过程中的错误信息
func createMultiAgentSystem(ctx context.Context) (*MultiAgentRouter, error) {
	// 系统初始化开始标记 - 记录多智能体系统的构建过程
	log.Println("创建Multi-Agent Hosting系统...")

	// 第一阶段：创建 Host Agent（主控智能体）
	// Host Agent 是系统的大脑和决策中心，负责理解用户意图并做出智能路由决策
	// 它不直接执行具体任务，而是分析用户需求并选择合适的专业智能体
	hostAgent, err := createHostAgent(ctx)
	if err != nil {
		// Host Agent 创建失败会导致整个系统无法工作，直接返回错误
		return nil, err
	}

	// 第二阶段：创建专业智能体团队
	// 每个专业智能体都专注于特定领域，配备相应的专业工具，提供高质量的专业服务
	// 专业化分工确保了系统的高效性和准确性

	// 创建天气专家智能体 - 专门处理天气相关查询
	// 配备 WeatherTool，能够查询各个城市的实时天气信息
	weatherSpecialist, err := createWeatherSpecialist(ctx)
	if err != nil {
		// 天气专家创建失败时返回详细错误信息，便于问题定位
		return nil, err
	}

	// 创建计算专家智能体 - 专门处理数学计算任务
	// 配备 CalculatorTool，能够执行各种数学运算和表达式计算
	calculatorSpecialist, err := createCalculatorSpecialist(ctx)
	if err != nil {
		// 计算专家创建失败时返回详细错误信息，便于问题定位
		return nil, err
	}

	// 创建时间专家智能体 - 专门处理时间日期查询
	// 配备 TimeTool，能够提供多种格式的时间信息和日期计算
	timeSpecialist, err := createTimeSpecialist(ctx)
	if err != nil {
		// 时间专家创建失败时返回详细错误信息，便于问题定位
		return nil, err
	}

	// 第三阶段：组装多智能体路由器
	// 路由器是系统的协调中心和统一接口，整合所有智能体并提供完整的服务能力
	// 它实现了多智能体协作的完整工作流程：接收请求 -> 意图分析 -> 任务分发 -> 结果整合
	router := &MultiAgentRouter{
		hostAgent:            hostAgent,            // 主控智能体：负责意图分析和路由决策
		weatherSpecialist:    weatherSpecialist,    // 天气专家：处理天气查询任务
		calculatorSpecialist: calculatorSpecialist, // 计算专家：处理数学计算任务
		timeSpecialist:       timeSpecialist,       // 时间专家：处理时间日期查询任务
	}

	// 系统构建完成标记 - 确认多智能体系统已成功初始化并可以开始服务
	log.Println("Multi-Agent Hosting系统创建成功")

	// 返回完整的多智能体系统实例
	// 该实例可以直接处理用户请求，自动协调各个专业智能体完成复杂任务
	return router, nil
}

// main 函数 - 多智能体托管系统的主入口点
// 该函数演示了如何构建和运行一个完整的多智能体协作系统
// 系统架构：Host Agent（意图分析） + 多个专业智能体（任务执行） + 路由器（协调管理）
func main() {
	// 第一阶段：系统初始化和配置加载
	// 使用 Viper 配置管理库加载系统配置文件，获取 API 密钥和服务端点等关键信息
	viper.SetConfigName("config") // 配置文件名称（不含扩展名）
	viper.SetConfigType("yaml")   // 配置文件类型：YAML 格式
	viper.AddConfigPath(".")      // 配置文件搜索路径：当前目录

	// 读取配置文件 - 加载系统运行所需的各项配置参数
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 设置 ARK API 环境变量 - 为所有智能体提供统一的 API 访问配置
	// ARK 是字节跳动提供的大语言模型服务，需要 API 密钥进行身份认证
	os.Setenv("ARK_API_KEY", viper.GetString("ARK_API_KEY"))

	// 创建执行上下文 - 用于控制所有智能体的执行环境和生命周期管理
	ctx := context.Background()

	// 第二阶段：多智能体系统构建
	// 创建完整的多智能体托管系统，包括主控智能体和所有专业智能体
	multiAgentSystem, err := createMultiAgentSystem(ctx)
	if err != nil {
		log.Fatalf("创建多代理系统失败: %v", err)
	}

	// 系统启动成功提示 - 向用户展示系统架构和组成
	fmt.Println("🤖 多代理托管系统演示启动成功！")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("系统架构:")
	fmt.Println("🏢 Host Agent - 负责意图识别和路由决策")       // 主控智能体功能说明
	fmt.Println("🌤️  Weather Specialist - 天气查询专家")  // 天气专家功能说明
	fmt.Println("🔢 Calculator Specialist - 数学计算专家") // 计算专家功能说明
	fmt.Println("⏰ Time Specialist - 时间查询专家")       // 时间专家功能说明
	fmt.Println(strings.Repeat("=", 60))

	// 第三阶段：测试用例设计和执行
	// 设计多样化的测试用例，验证多智能体系统的各种能力
	// 简化版本专注于基本功能演示，包括单一专家任务处理
	testCases := []struct {
		name    string // 测试用例名称，便于识别和调试
		message string // 用户输入消息，模拟真实用户请求
	}{
		{
			name:    "天气查询测试", // 测试天气专家智能体的功能
			message: "北京天气",   // 简单的天气查询请求
		},
		{
			name:    "数学计算测试",    // 测试计算专家智能体的功能
			message: "计算 10 + 5", // 基本的数学运算请求
		},
	}

	// 第四阶段：测试用例执行循环
	// 逐一执行测试用例，展示多智能体系统的完整工作流程
	for i, testCase := range testCases {
		// 测试用例开始标记 - 清晰地分隔不同的测试场景
		fmt.Printf("\n📝 测试用例 %d: %s\n", i+1, testCase.name)
		fmt.Printf("👤 用户: %s\n", testCase.message)
		fmt.Println(strings.Repeat("-", 50))

		// 构建系统消息 - 为 Host Agent 提供角色定义和专家信息
		// 这个消息指导 Host Agent 如何分析用户意图和选择合适的专业智能体
		systemMessage := &schema.Message{
			Role: schema.System, // 标识为系统消息
			Content: `你是一个智能助手调度中心的Host Agent。你的任务是分析用户的意图，然后决定调用哪个专家代理来处理请求。

可用的专家代理:
1. weather_specialist - 处理天气查询相关问题
2. calculator_specialist - 处理数学计算相关问题  
3. time_specialist - 处理时间和日期查询相关问题

请根据用户的问题，选择合适的专家代理来处理。如果需要多个步骤，可以依次调用多个专家。

分析用户意图，然后给出你的路由决策。`,
		}

		// 构建用户消息 - 将测试用例转换为标准的消息格式
		userMessage := &schema.Message{
			Role:    schema.User,      // 标识为用户消息
			Content: testCase.message, // 用户的具体请求内容
		}

		// 执行多智能体系统处理流程
		// 这里会触发完整的多智能体协作流程：意图分析 -> 任务路由 -> 专家执行 -> 结果整合
		messages := []*schema.Message{systemMessage, userMessage}
		result, err := multiAgentSystem.Invoke(ctx, messages)
		if err != nil {
			// 处理执行失败的情况 - 记录错误并继续下一个测试用例
			fmt.Printf("❌ 执行失败: %v\n", err)
			continue
		}

		// 展示系统的最终响应结果
		if result != nil && result.Content != "" {
			fmt.Printf("🤖 系统回复: %s\n", result.Content)
		} else {
			fmt.Printf("🤖 系统回复: [没有收到回复内容]\n")
		}

		// 添加执行间隔 - 便于观察和分析系统日志输出
		time.Sleep(1 * time.Second)
	}

	// 第五阶段：演示总结和系统说明
	// 向用户展示多智能体系统的核心特性和工作原理
	fmt.Println("\n✅ 所有测试用例执行完成！")
	fmt.Println("\n💡 多代理托管系统演示:")
	fmt.Println("   - Host Agent 负责分析用户意图")     // 主控智能体的核心职责
	fmt.Println("   - 根据意图路由到对应的专家代理")          // 智能路由机制
	fmt.Println("   - 每个专家代理专注于特定领域的任务")        // 专业化分工优势
	fmt.Println("   - 支持复杂的多步骤任务处理")            // 协作处理能力
	fmt.Println("📚 如需了解更多功能，请参考 README.md 文档。") // 进一步学习指引
}
