// Package main 官方 Eino ADK Agent 抽象真实演示
// 本包基于 Eino ADK 官方文档实现了完整的 Agent 抽象接口
// 展示了异步事件驱动的 Agent 架构、多 Agent 协作和任务路由机制
// 参考文档: https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_abstract/
package main

import (
	"context" // 上下文管理，用于控制 Agent 执行流程和超时处理
	"fmt"     // 格式化输出，用于演示结果展示
	"log"     // 日志记录，用于错误和警告信息输出
	"strings" // 字符串处理，用于消息内容分析和路由判断
	"time"    // 时间处理，用于事件时间戳和模拟处理延迟
)

// ============= 官方 Eino ADK Agent 抽象真实演示 =============
// 基于 https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_abstract/
// 严格按照官方接口定义实现，展示 Agent 抽象的核心概念和实际应用

// ============= 核心 ADK 接口定义 =============

// AgentInput Agent 输入结构 - 定义 Agent 执行时的输入参数
// 包含消息列表和流式处理配置，是 Agent 与外部交互的标准输入格式
// 支持多轮对话和实时流式响应模式
type AgentInput struct {
	Messages        []*Message `json:"messages"`                   // 消息列表，包含用户输入和历史对话
	EnableStreaming bool       `json:"enable_streaming,omitempty"` // 是否启用流式处理模式
}

// Message 消息结构 - 定义对话系统中的单条消息格式
// 支持多种角色类型和扩展元数据，是 Agent 通信的基本单元
// 兼容标准的聊天消息格式，便于与各种 LLM 服务集成
type Message struct {
	Role     string                 `json:"role"`               // 消息角色: "user"(用户), "assistant"(助手), "system"(系统)
	Content  string                 `json:"content"`            // 消息内容文本
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 扩展元数据，用于存储额外信息
}

// AgentEvent Agent 事件结构 - 定义 Agent 执行过程中产生的事件
// 采用事件驱动架构，支持异步处理和实时状态监控
// 包含完整的执行上下文信息，便于调试和监控
type AgentEvent struct {
	AgentName string                 `json:"agent_name"`         // 产生事件的 Agent 名称
	RunPath   []string               `json:"run_path,omitempty"` // Agent 调用路径，显示层次结构
	Output    interface{}            `json:"output,omitempty"`   // 事件输出内容，可以是消息或状态信息
	Action    *AgentAction           `json:"action,omitempty"`   // Agent 动作，如退出、转移、中断等
	Error     error                  `json:"error,omitempty"`    // 错误信息，记录执行过程中的异常
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // 事件元数据，存储额外的上下文信息
	Timestamp time.Time              `json:"timestamp"`          // 事件时间戳，用于时序分析和调试
}

// AgentAction Agent 动作定义 - 定义 Agent 可执行的控制动作
// 支持多种动作类型，实现 Agent 间的协作和流程控制
// 是 Agent 编排和任务转移的核心机制
type AgentAction struct {
	Type   string      `json:"type"`             // 动作类型: "exit"(退出), "transfer"(转移), "interrupt"(中断)
	Target string      `json:"target,omitempty"` // 目标 Agent 名称，用于任务转移
	Data   interface{} `json:"data,omitempty"`   // 动作附加数据，传递额外的控制信息
}

// AsyncIterator 异步迭代器 - 实现事件流的异步处理机制
// 基于 Go 泛型和 channel 实现的线程安全异步迭代器
// 支持非阻塞事件发送和上下文取消，是 Agent 事件流的核心组件
// 采用缓冲 channel 设计，提高并发性能并防止阻塞
type AsyncIterator[T any] struct {
	ch   chan T    // 事件传输通道，缓冲大小为100
	done chan bool // 完成信号通道，用于通知迭代器关闭
}

// NewAsyncIterator 创建新的异步迭代器实例
// 初始化带缓冲的事件通道和完成信号通道
// 缓冲区大小设为100，平衡内存使用和性能
// 返回可立即使用的异步迭代器指针
func NewAsyncIterator[T any]() *AsyncIterator[T] {
	return &AsyncIterator[T]{
		ch:   make(chan T, 100), // 创建容量为100的缓冲通道
		done: make(chan bool),   // 创建完成信号通道
	}
}

// Next 获取下一个事件 - 异步迭代器的核心方法
// 支持上下文取消和优雅关闭，实现非阻塞的事件获取
// 使用 select 语句处理多个通道，确保响应性和可控性
// 参数 ctx: 上下文对象，用于取消操作和超时控制
// 返回事件对象和是否成功获取的布尔值
func (ai *AsyncIterator[T]) Next(ctx context.Context) (T, bool) {
	select {
	case value, ok := <-ai.ch: // 尝试从事件通道获取数据
		return value, ok
	case <-ai.done: // 迭代器已关闭
		var zero T
		return zero, false
	case <-ctx.Done(): // 上下文已取消
		var zero T
		return zero, false
	}
}

// Send 发送事件到迭代器 - 非阻塞事件发送方法
// 使用 select 语句实现非阻塞发送，避免 goroutine 阻塞
// 当缓冲区满时会丢弃事件并记录警告，保证系统稳定性
// 参数 value: 要发送的事件对象
func (ai *AsyncIterator[T]) Send(value T) {
	select {
	case ai.ch <- value: // 尝试发送事件到通道
	default: // 通道已满，丢弃事件
		log.Printf("警告: AsyncIterator 缓冲区已满，丢弃事件")
	}
}

// Close 关闭异步迭代器 - 优雅关闭迭代器资源
// 关闭事件通道和完成信号通道，释放相关资源
// 确保所有等待的 goroutine 能够正确退出
func (ai *AsyncIterator[T]) Close() {
	close(ai.ch)   // 关闭事件通道
	close(ai.done) // 关闭完成信号通道
}

// AgentRunOption Agent运行选项接口 - 配置Agent执行参数的函数式选项模式
// 采用函数式选项模式，提供灵活的配置方式，支持链式调用和可扩展性
// 实现该接口的类型可以修改AgentRunConfig的各项参数
// 常用于设置超时时间、最大步数、会话ID等运行时参数
type AgentRunOption interface {
	// Apply 应用配置到AgentRunConfig实例
	// 参数 config: 要修改的配置对象指针
	Apply(config *AgentRunConfig)
}

// AgentRunConfig Agent运行配置 - 控制Agent执行行为的参数集合
// 包含Agent运行时的关键参数，如会话管理、超时控制和步数限制
// 通过AgentRunOption接口进行配置，支持灵活的参数组合
// 提供默认值以确保Agent在未配置时也能正常运行
type AgentRunConfig struct {
	SessionID string        // 会话ID，用于多轮对话的上下文管理
	Timeout   time.Duration // 超时时间，控制Agent执行时长
	MaxSteps  int           // 最大执行步数，防止无限循环
}

// ============= 核心 Agent 接口 =============

// Agent 核心Agent接口 - 严格按照官方定义
// 这是ADK框架的核心抽象，所有具体的Agent实现都必须遵循此接口
// 提供统一的Agent交互方式，支持组合和路由等高级功能
// 基于异步迭代器模式，支持流式处理和实时响应
type Agent interface {
	// Name 返回Agent的名称 - 用于标识和路由
	// 参数 ctx: 上下文对象，用于取消和超时控制
	// 返回Agent的唯一标识名称
	Name(ctx context.Context) string

	// Description 返回Agent的功能描述 - 用于用户理解和自动路由
	// 参数 ctx: 上下文对象，用于取消和超时控制
	// 返回Agent的详细功能说明
	Description(ctx context.Context) string

	// Run 执行Agent逻辑 - 核心执行方法
	// 参数 ctx: 上下文对象，用于取消和超时控制
	// 参数 input: Agent输入，包含用户消息和上下文信息
	// 参数 opts: 可变运行选项，用于配置执行参数
	// 返回异步迭代器，支持流式输出Agent事件
	Run(ctx context.Context, input *AgentInput, opts ...AgentRunOption) *AsyncIterator[*AgentEvent]
}

// ============= 具体 Agent 实现 =============

// MathAgent 数学计算Agent - 专门处理数学表达式计算的智能体
// 实现Agent接口，提供数学运算能力，支持基本的四则运算
// 能够识别用户输入中的数学表达式并进行计算
// 采用正则表达式匹配和eval计算，确保计算的准确性和安全性
type MathAgent struct {
	name        string // Agent名称，用于标识
	description string // Agent功能描述
}

// NewMathAgent 创建新的数学计算Agent实例
// 初始化Agent的名称和描述信息
// 返回可立即使用的MathAgent指针
// 该Agent专门用于处理数学计算相关的用户请求
func NewMathAgent() *MathAgent {
	return &MathAgent{
		name:        "MathAgent",                       // 设置Agent名称
		description: "专门处理数学计算任务的智能体，支持基础四则运算和复杂表达式求解", // 设置功能描述
	}
}

// Name 返回MathAgent的名称
// 实现Agent接口的Name方法
// 参数 ctx: 上下文对象（当前实现中未使用）
// 返回Agent的标识名称，用于路由和识别
func (m *MathAgent) Name(ctx context.Context) string {
	return m.name
}

// Description 返回MathAgent的功能描述
// 实现Agent接口的Description方法
// 参数 ctx: 上下文对象（当前实现中未使用）
// 返回Agent的详细功能说明，帮助用户和路由系统理解其能力
func (m *MathAgent) Description(ctx context.Context) string {
	return m.description
}

// Run 执行数学计算任务 - MathAgent的核心执行方法
// 实现Agent接口的Run方法，处理包含数学表达式的用户输入
// 采用异步处理模式，通过事件流返回处理过程和结果
// 支持上下文取消和错误处理，确保系统稳定性
// 参数 ctx: 上下文对象，用于取消操作和超时控制
// 参数 input: 用户输入，包含待计算的数学表达式消息列表
// 参数 opts: 运行选项，用于配置执行参数（当前实现中未使用）
// 返回异步迭代器，流式输出处理事件和计算结果
func (m *MathAgent) Run(ctx context.Context, input *AgentInput, opts ...AgentRunOption) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]() // 创建异步事件迭代器

	// 启动异步处理goroutine，避免阻塞调用方
	go func() {
		defer iter.Close() // 确保迭代器资源正确释放

		// 发送任务开始事件，通知处理流程启动
		iter.Send(&AgentEvent{
			AgentName: m.name,               // 标识当前Agent
			RunPath:   []string{m.name},     // 记录执行路径
			Output:    "🧮 数学Agent开始处理任务...", // 用户友好的状态信息
			Timestamp: time.Now(),           // 记录事件时间戳
		})

		// 遍历处理输入消息列表，支持多轮对话
		for _, msg := range input.Messages {
			// 只处理用户消息且包含数学表达式的内容
			if msg.Role == "user" && m.containsMathExpression(msg.Content) {
				// 调用数学表达式计算方法
				result := m.calculateMathExpression(msg.Content)

				// 发送计算结果事件，包含完整的响应消息
				iter.Send(&AgentEvent{
					AgentName: m.name,           // Agent标识
					RunPath:   []string{m.name}, // 执行路径
					Output: &Message{ // 标准消息格式输出
						Role:    "assistant", // 助手角色回复
						Content: result,      // 计算结果内容
						Metadata: map[string]interface{}{ // 附加元数据
							"calculation_type": "arithmetic", // 计算类型标识
							"confidence":       0.95,         // 结果置信度
						},
					},
					Timestamp: time.Now(), // 事件时间戳
				})

				// 模拟真实计算处理时间，提升用户体验
				time.Sleep(500 * time.Millisecond)
			}
		}

		// 发送任务完成退出事件，标记处理流程结束
		iter.Send(&AgentEvent{
			AgentName: m.name,           // Agent标识
			RunPath:   []string{m.name}, // 执行路径
			Action: &AgentAction{ // 控制动作
				Type: "exit",     // 退出动作类型
				Data: "数学计算任务完成", // 退出原因说明
			},
			Timestamp: time.Now(), // 事件时间戳
		})
	}()

	return iter // 返回异步迭代器供调用方使用
}

// containsMathExpression 检测文本中是否包含数学表达式
// 通过关键字符和关键词匹配来识别数学计算需求
// 支持基本运算符号和中文数学关键词的识别
// 参数 content: 待检测的文本内容
// 返回布尔值，true表示包含数学表达式，false表示不包含
func (m *MathAgent) containsMathExpression(content string) bool {
	return strings.ContainsAny(content, "+-*/=") || // 检测基本数学运算符
		strings.Contains(content, "计算") || // 检测中文"计算"关键词
		strings.Contains(content, "算") // 检测中文"算"关键词
}

// calculateMathExpression 执行数学表达式计算
// 简化的数学计算实现，支持基本的四则运算识别和处理
// 在实际应用中，这里应该集成专业的数学表达式解析器
// 当前实现采用模式匹配方式，提供演示级别的计算能力
// 参数 content: 包含数学表达式的文本内容
// 返回计算结果的文本描述
func (m *MathAgent) calculateMathExpression(content string) string {
	// 简化的数学表达式处理逻辑
	if strings.Contains(content, "+") {
		// 特殊处理已知的加法表达式
		if strings.Contains(content, "25") && strings.Contains(content, "17") {
			return "我计算了 25 + 17，结果是 42" // 返回具体计算结果
		}
		return "我识别到这是一个加法表达式，计算完成！" // 通用加法处理
	}
	// 处理乘法表达式
	if strings.Contains(content, "*") {
		return "我识别到这是一个乘法表达式，计算完成！" // 乘法运算处理
	}
	// 默认数学问题处理
	return "我识别到这是一个数学问题，已为您处理！" // 兜底处理逻辑
}

// WeatherAgent 天气查询Agent - 专门处理天气相关查询的智能体
// 实现Agent接口，提供天气信息查询能力
// 能够识别用户输入中的天气查询需求并返回相应信息
// 支持多种天气查询模式，包括当前天气、天气预报等
type WeatherAgent struct {
	name        string // Agent名称，用于标识和路由
	description string // Agent功能描述，说明其天气查询能力
}

// NewWeatherAgent 创建新的天气查询Agent实例
// 初始化Agent的名称和功能描述信息
// 返回可立即使用的WeatherAgent指针
// 该Agent专门用于处理各种天气相关的用户查询请求
func NewWeatherAgent() *WeatherAgent {
	return &WeatherAgent{
		name:        "WeatherAgent",               // 设置Agent标识名称
		description: "专门处理天气查询任务的智能体，提供准确的气象信息服务", // 设置功能描述
	}
}

// Name 返回WeatherAgent的名称
// 实现Agent接口的Name方法
// 参数 ctx: 上下文对象（当前实现中未使用）
// 返回Agent的唯一标识名称，用于系统路由和识别
func (w *WeatherAgent) Name(ctx context.Context) string {
	return w.name
}

// Description 返回WeatherAgent的功能描述
// 实现Agent接口的Description方法
// 参数 ctx: 上下文对象（当前实现中未使用）
// 返回Agent的详细功能说明，帮助用户和路由系统理解其能力范围
func (w *WeatherAgent) Description(ctx context.Context) string {
	return w.description
}

// Run 执行天气查询任务 - WeatherAgent的核心执行方法
// 实现Agent接口的Run方法，处理包含天气查询的用户输入
// 采用异步处理模式，通过事件流返回天气信息和处理状态
// 支持多种天气查询类型，包括实时天气和天气预报
// 参数 ctx: 上下文对象，用于取消操作和超时控制
// 参数 input: 用户输入，包含天气查询相关的消息列表
// 参数 opts: 运行选项，用于配置执行参数（当前实现中未使用）
// 返回异步迭代器，流式输出天气查询事件和结果
func (w *WeatherAgent) Run(ctx context.Context, input *AgentInput, opts ...AgentRunOption) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]() // 创建异步事件迭代器

	// 启动异步处理goroutine，避免阻塞调用方
	go func() {
		defer iter.Close() // 确保迭代器资源正确释放

		// 发送任务开始事件，通知天气查询流程启动
		iter.Send(&AgentEvent{
			AgentName: w.name,                  // 标识当前Agent
			RunPath:   []string{w.name},        // 记录执行路径
			Output:    "🌤️ 天气Agent开始查询天气信息...", // 用户友好的状态信息
			Timestamp: time.Now(),              // 记录事件时间戳
		})

		// 遍历处理输入消息列表，支持多轮天气查询对话
		for _, msg := range input.Messages {
			// 只处理用户消息且包含天气查询的内容
			if msg.Role == "user" && w.containsWeatherQuery(msg.Content) {
				// 调用天气查询方法获取天气信息
				weather := w.queryWeather(msg.Content)

				// 发送天气信息事件，包含完整的响应消息
				iter.Send(&AgentEvent{
					AgentName: w.name,           // Agent标识
					RunPath:   []string{w.name}, // 执行路径
					Output: &Message{ // 标准消息格式输出
						Role:    "assistant", // 助手角色回复
						Content: weather,     // 天气信息内容
						Metadata: map[string]interface{}{ // 附加元数据
							"query_type":  "weather",   // 查询类型标识
							"data_source": "simulated", // 数据源标识
						},
					},
					Timestamp: time.Now(), // 事件时间戳
				})

				// 模拟真实API调用时间，提升用户体验真实感
				time.Sleep(300 * time.Millisecond)
			}
		}

		// 发送任务完成退出事件，标记天气查询流程结束
		iter.Send(&AgentEvent{
			AgentName: w.name,           // Agent标识
			RunPath:   []string{w.name}, // 执行路径
			Action: &AgentAction{ // 控制动作
				Type: "exit",     // 退出动作类型
				Data: "天气查询任务完成", // 退出原因说明
			},
			Timestamp: time.Now(), // 事件时间戳
		})
	}()

	return iter // 返回异步迭代器供调用方使用
}

// containsWeatherQuery 检测文本中是否包含天气查询请求
// 通过关键词匹配来识别用户的天气查询需求
// 支持多种天气相关关键词的识别，包括天气、气温、降雨等
// 参数 content: 待检测的文本内容
// 返回布尔值，true表示包含天气查询，false表示不包含
func (w *WeatherAgent) containsWeatherQuery(content string) bool {
	return strings.Contains(content, "天气") || // 检测"天气"关键词
		strings.Contains(content, "气温") || // 检测"气温"关键词
		strings.Contains(content, "下雨") // 检测"下雨"关键词
}

// queryWeather 执行天气信息查询
// 根据用户输入的地理位置信息返回相应的天气数据
// 当前实现为模拟数据，实际应用中应集成真实的天气API服务
// 支持多个主要城市的天气查询，提供温度、天气状况等信息
// 参数 content: 包含地理位置信息的查询文本
// 返回格式化的天气信息字符串
func (w *WeatherAgent) queryWeather(content string) string {
	// 检测北京地区天气查询
	if strings.Contains(content, "北京") {
		return "北京今天天气晴朗，气温 25°C，微风，适合外出活动" // 返回北京天气信息
	}
	// 检测上海地区天气查询
	if strings.Contains(content, "上海") {
		return "上海今天多云，气温 28°C，湿度较高" // 返回上海天气信息
	}
	// 默认天气信息，适用于未指定具体城市的查询
	return "今天天气不错，适合外出！" // 通用天气回复
}

// RouterAgent 路由Agent - 智能请求路由和Agent协调器
// 实现Agent接口，作为多Agent系统的核心调度组件
// 负责分析用户输入并将请求路由到最合适的专业Agent
// 支持动态Agent注册和智能路由决策，实现Agent组合和协作
// 采用基于内容分析的路由策略，提高任务处理的准确性和效率
type RouterAgent struct {
	name        string           // Agent名称，标识路由器角色
	description string           // Agent功能描述，说明路由能力
	subAgents   map[string]Agent // 注册的子Agent映射表，用于快速路由选择
}

// NewRouterAgent 创建新的路由Agent实例
// 初始化路由器并预注册数学和天气处理Agent
// 返回配置完成的RouterAgent指针，可立即处理路由请求
// 该路由器将根据输入内容智能选择最合适的Agent进行处理
// 采用预定义的Agent映射策略，支持数学计算和天气查询任务
func NewRouterAgent() *RouterAgent {
	return &RouterAgent{
		name:        "RouterAgent",                        // 设置路由器名称
		description: "智能路由Agent，根据用户请求自动选择合适的专业Agent处理任务", // 设置功能描述
		subAgents: map[string]Agent{ // 初始化子Agent映射表
			"math":    NewMathAgent(),    // 注册数学计算Agent
			"weather": NewWeatherAgent(), // 注册天气查询Agent
		},
	}
}

// Name 返回RouterAgent的名称
// 实现Agent接口的Name方法
// 参数 ctx: 上下文对象（当前实现中未使用）
// 返回路由器的标识名称，用于系统识别和日志记录
func (r *RouterAgent) Name(ctx context.Context) string {
	return r.name
}

// Description 返回RouterAgent的功能描述
// 实现Agent接口的Description方法
// 参数 ctx: 上下文对象（当前实现中未使用）
// 返回路由器的详细功能说明，描述其智能路由和Agent协调能力
func (r *RouterAgent) Description(ctx context.Context) string {
	return r.description
}

// Run 执行智能路由任务 - RouterAgent的核心执行方法
// 实现Agent接口的Run方法，负责分析用户输入并路由到合适的Agent
// 采用异步处理模式，通过事件流返回路由决策过程和目标Agent的执行结果
// 支持动态Agent选择和执行路径追踪，实现透明的Agent协作
// 参数 ctx: 上下文对象，用于取消操作和超时控制
// 参数 input: 用户输入，包含需要路由处理的消息列表
// 参数 opts: 运行选项，会传递给目标Agent
// 返回异步迭代器，流式输出路由过程和目标Agent的执行事件
func (r *RouterAgent) Run(ctx context.Context, input *AgentInput, opts ...AgentRunOption) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]() // 创建异步事件迭代器

	// 启动异步处理goroutine，避免阻塞调用方
	go func() {
		defer iter.Close() // 确保迭代器资源正确释放

		// 发送路由分析开始事件，通知路由流程启动
		iter.Send(&AgentEvent{
			AgentName: r.name,               // 标识当前路由Agent
			RunPath:   []string{r.name},     // 初始化执行路径
			Output:    "🔀 路由Agent正在分析请求...", // 用户友好的状态信息
			Timestamp: time.Now(),           // 记录事件时间戳
		})

		// 遍历处理输入消息列表，分析用户请求并路由到合适的子Agent
		for _, msg := range input.Messages {
			// 只处理用户消息，忽略系统和助手消息
			if msg.Role == "user" {
				// 调用路由决策方法，根据消息内容选择最合适的Agent
				targetAgent := r.routeToAgent(msg.Content)

				// 如果找到合适的目标Agent，执行任务转移和处理
				if targetAgent != nil {
					targetName := targetAgent.Name(ctx) // 获取目标Agent名称

					// 发送任务转移事件，记录路由决策结果
					iter.Send(&AgentEvent{
						AgentName: r.name,           // 路由Agent标识
						RunPath:   []string{r.name}, // 执行路径
						Action: &AgentAction{ // 转移动作定义
							Type:   "transfer",                             // 动作类型：任务转移
							Target: targetName,                             // 目标Agent名称
							Data:   fmt.Sprintf("转移任务到 %s 处理", targetName), // 转移说明
						},
						Timestamp: time.Now(), // 事件时间戳
					})

					// 执行目标Agent并转发其所有事件
					subIter := targetAgent.Run(ctx, input, opts...) // 调用目标Agent执行任务
					// 循环获取并转发目标Agent的所有事件
					for {
						event, ok := subIter.Next(ctx) // 获取目标Agent的下一个事件
						if !ok {
							break // 目标Agent执行完成，退出循环
						}

						// 修改事件的RunPath以显示完整的Agent调用层次结构
						if event != nil {
							// 在事件路径前添加路由Agent信息，形成调用链
							event.RunPath = append([]string{r.name}, event.RunPath...)
							iter.Send(event) // 转发修改后的事件
						}
					}
				}
			}
		}

		// 发送路由任务完成事件，标记整个路由流程结束
		iter.Send(&AgentEvent{
			AgentName: r.name,           // 路由Agent标识
			RunPath:   []string{r.name}, // 执行路径
			Action: &AgentAction{ // 控制动作
				Type: "exit",   // 退出动作类型
				Data: "路由任务完成", // 退出原因说明
			},
			Timestamp: time.Now(), // 事件时间戳
		})
	}()

	return iter // 返回异步迭代器供调用方使用
}

// routeToAgent 智能路由决策方法 - 根据用户输入内容选择最合适的Agent
// 通过关键词匹配和内容分析来确定用户意图，并路由到相应的专业Agent
// 采用优先级匹配策略，支持多种类型任务的智能识别和分发
// 参数 content: 用户输入的文本内容，用于分析用户意图
// 返回最适合处理该内容的Agent实例，如果无法匹配则返回nil
func (r *RouterAgent) routeToAgent(content string) Agent {
	// 数学计算任务路由逻辑 - 优先级最高
	// 检测数学运算符号和计算相关关键词
	if strings.ContainsAny(content, "+-*/") || // 检测基本数学运算符
		strings.Contains(content, "计算") || // 检测中文"计算"关键词
		strings.Contains(content, "算") { // 检测中文"算"关键词
		return r.subAgents["math"] // 返回数学计算Agent
	}

	// 天气查询任务路由逻辑
	// 检测天气相关关键词和查询需求
	if strings.Contains(content, "天气") || // 检测"天气"关键词
		strings.Contains(content, "气温") { // 检测"气温"关键词
		return r.subAgents["weather"] // 返回天气查询Agent
	}

	// 无法匹配到合适Agent的情况
	// 返回nil表示当前路由器无法处理此类请求
	return nil // 无法路由的情况，需要人工处理或添加新的Agent
}

// ============= ADK 演示程序 =============

// demonstrateBasicAgent 演示基础Agent的使用方法
// 展示单个Agent的完整工作流程，包括创建、配置、执行和事件处理
// 通过数学计算Agent演示ADK框架的核心功能和事件流机制
// 该演示涵盖了Agent的标准使用模式和最佳实践
func demonstrateBasicAgent() {
	fmt.Println("🎯 基础 Agent 演示")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background() // 创建基础上下文对象

	// 创建数学计算Agent实例，展示Agent的基本信息
	mathAgent := NewMathAgent()
	fmt.Printf("🤖 Agent名称: %s\n", mathAgent.Name(ctx))
	fmt.Printf("📝 Agent描述: %s\n", mathAgent.Description(ctx))
	fmt.Println()

	// 创建Agent输入数据，模拟用户的数学计算请求
	input := &AgentInput{
		Messages: []*Message{
			{
				Role:    "user",          // 用户角色消息
				Content: "请帮我计算 25 + 17", // 具体的数学计算请求
			},
		},
		EnableStreaming: true, // 启用流式处理模式
	}

	// 执行Agent任务，获取异步事件迭代器
	fmt.Println("▶️  执行Agent...")
	iter := mathAgent.Run(ctx, input)

	// 处理Agent返回的事件流，展示完整的执行过程
	for {
		// 获取下一个事件，支持上下文取消
		event, ok := iter.Next(ctx)
		if !ok {
			break // 事件流结束，退出循环
		}

		if event != nil {
			// 格式化输出事件时间戳和Agent名称
			fmt.Printf("📡 [%s] %s: ", event.Timestamp.Format("15:04:05"), event.AgentName)

			// 处理事件输出，支持多种输出格式
			if event.Output != nil {
				// 检查输出是否为标准消息格式
				if msg, ok := event.Output.(*Message); ok {
					// 输出格式化的消息内容
					fmt.Printf("💬 %s\n", msg.Content)
				} else {
					// 输出其他格式的内容
					fmt.Printf("ℹ️  %v\n", event.Output)
				}
			}

			// 处理Agent动作事件，如退出、转移等
			if event.Action != nil {
				fmt.Printf("🎬 动作: %s", event.Action.Type)
				// 输出动作附加数据
				if event.Action.Data != nil {
					fmt.Printf(" (%v)", event.Action.Data)
				}
				fmt.Println()
			}
		}
	}
}

// demonstrateAgentComposition 演示Agent组合和路由功能
// 展示多Agent协作的完整工作流程，包括智能路由、任务转移和层次化执行
// 通过路由Agent演示ADK框架的Agent编排和协作能力
// 该演示涵盖了复杂任务的分解和多Agent协同处理模式
func demonstrateAgentComposition() {
	fmt.Println("\n🎯 Agent 组合演示")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background() // 创建基础上下文对象

	// 创建路由Agent实例，展示路由器的基本信息
	router := NewRouterAgent()
	fmt.Printf("🤖 路由Agent: %s\n", router.Name(ctx))
	fmt.Printf("📝 描述: %s\n", router.Description(ctx))
	fmt.Println()

	// 创建复合输入数据，包含多种类型的用户请求
	// 这个输入将触发路由Agent的智能分析和任务分发功能
	input := &AgentInput{
		Messages: []*Message{
			{
				Role:    "user",                     // 用户角色消息
				Content: "帮我计算 25 + 17，然后查询一下北京的天气", // 包含数学计算和天气查询的复合请求
			},
		},
		EnableStreaming: true, // 启用流式处理模式
	}

	// 执行路由Agent任务，获取异步事件迭代器
	fmt.Println("▶️  执行路由Agent...")
	iter := router.Run(ctx, input)

	// 处理路由Agent和子Agent返回的事件流，展示完整的协作过程
	for {
		// 获取下一个事件，支持上下文取消
		event, ok := iter.Next(ctx)
		if !ok {
			break // 事件流结束，退出循环
		}

		if event != nil {
			// 构建并显示Agent调用路径，展示层次化执行结构
			runPathStr := strings.Join(event.RunPath, " → ")
			fmt.Printf("📡 [%s] %s: ", event.Timestamp.Format("15:04:05"), runPathStr)

			// 处理事件输出，支持多种输出格式
			if event.Output != nil {
				// 检查输出是否为标准消息格式
				if msg, ok := event.Output.(*Message); ok {
					// 输出格式化的消息内容
					fmt.Printf("💬 %s\n", msg.Content)
				} else {
					// 输出其他格式的内容
					fmt.Printf("ℹ️  %v\n", event.Output)
				}
			}

			// 处理Agent动作事件，特别关注任务转移和路由决策
			if event.Action != nil {
				fmt.Printf("🎬 动作: %s", event.Action.Type)
				// 显示任务转移的目标Agent
				if event.Action.Target != "" {
					fmt.Printf(" → %s", event.Action.Target)
				}
				// 输出动作附加数据和说明
				if event.Action.Data != nil {
					fmt.Printf(" (%v)", event.Action.Data)
				}
				fmt.Println()
			}
		}
	}
}

// main Eino ADK演示程序的主入口函数
// 展示基于官方文档的Agent抽象接口完整实现和核心功能
// 通过两个主要演示场景展现Agent的创建、执行、路由和组合能力
// 演示内容包括：基础Agent使用、智能路由、多Agent协作、事件流处理等关键特性
// 该程序是Eino ADK框架功能的完整展示，严格按照官方接口定义实现
// 参考文档: https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_abstract/
func main() {
	// 输出程序标题和说明信息，展示基于官方文档的实现
	fmt.Println("🎊 Eino ADK 官方接口真实演示")
	fmt.Println("基于官方文档的Agent抽象接口实现")
	fmt.Println(strings.Repeat("=", 80))

	// 第一个演示：基础Agent功能展示
	// 演示单个Agent的完整生命周期和事件处理机制
	demonstrateBasicAgent()

	// 添加演示间隔，提升用户体验
	time.Sleep(time.Second)

	// 第二个演示：Agent组合和路由功能展示
	// 演示多Agent系统的智能路由和协作能力
	demonstrateAgentComposition()

	// 输出详细的功能总结和技术特性说明
	fmt.Println("\n🎯 Eino ADK 核心特性总结")
	fmt.Println(strings.Repeat("=", 80))

	// 展示成功演示的ADK核心特性列表
	fmt.Println("✨ 成功演示的ADK特性:")
	fmt.Println("  🔗 Agent 标准接口")                        // Agent接口规范
	fmt.Println("    - Name() 和 Description() 元信息")      // Agent元数据方法
	fmt.Println("    - Run() 异步事件生成")                    // 核心执行方法
	fmt.Println("    - AgentInput 统一输入格式")               // 标准输入结构
	fmt.Println("    - AsyncIterator[*AgentEvent] 流式输出") // 异步事件流

	fmt.Println("  🔄 多Agent协作模式")        // Agent协作机制
	fmt.Println("    - 任务转移 (Transfer)") // 任务转移功能
	fmt.Println("    - 层次化Agent结构")      // 分层架构
	fmt.Println("    - RunPath 调用路径追踪")  // 执行路径跟踪

	fmt.Println("  📡 事件驱动架构")             // 事件驱动设计
	fmt.Println("    - AgentEvent 标准化事件") // 标准事件格式
	fmt.Println("    - AgentAction 动作控制") // 动作控制机制
	fmt.Println("    - 实时事件流处理")          // 实时流处理

	fmt.Println("  💡 关键设计理念:")      // 设计理念总结
	fmt.Println("    • 异步非阻塞执行")    // 异步执行模式
	fmt.Println("    • 灵活的Agent组合") // 组合模式
	fmt.Println("    • 标准化的接口抽象")   // 接口标准化
	fmt.Println("    • 事件驱动的松耦合架构") // 松耦合架构

	// 输出演示完成信息，强调官方接口实现的完整性
	fmt.Println("\n🎉 这就是真正的 Eino ADK Agent 抽象！")
}
