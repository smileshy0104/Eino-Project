// Package main 演示 Eino ADK 多Agent组合模式的完整实现
// 本文件展示了三种核心的Agent协作模式，体现了现代AI系统中
// 多智能体协同工作的设计理念和最佳实践
package main

import (
	"context" // 上下文管理，用于控制Agent执行流程和超时处理
	"fmt"     // 格式化输出，用于演示结果展示
	"log"     // 日志记录，用于调试和错误追踪
	"strings" // 字符串处理，用于格式化显示和文本操作
	"time"    // 时间处理，用于事件时间戳和模拟延迟
)

// ============= Eino ADK 多Agent组合模式完整演示 =============
// 本演示基于 Eino ADK 官方文档，展示三种核心Agent组合模式：
//
// 🔧 模式1: Agent作为工具使用 (Tool Usage)
//    - Agent实现AgentTool接口，提供Invoke方法
//    - 支持同步调用，快速获取处理结果
//    - 适用于功能性调用和简单任务处理
//
// 🔄 模式2: Agent间任务转移 (Transfer)
//    - 通过AgentAction实现任务在Agent间的流转
//    - 保持执行上下文的连续性和状态传递
//    - 支持复杂工作流的分阶段执行
//
// 🏗️ 模式3: 层次化Agent结构 (Hierarchical)
//    - 主Agent协调多个专业子Agent
//    - 通过RunPath追踪调用层次和执行路径
//    - 实现复杂业务逻辑的模块化分解

// ============= 基础类型定义 =============

// Message 表示Agent间传递的消息结构
// 遵循标准的对话格式，支持角色区分和元数据扩展
type Message struct {
	Role     string                 `json:"role"`               // 消息角色：user(用户)、assistant(助手)、system(系统)
	Content  string                 `json:"content"`            // 消息内容：实际的文本信息
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 元数据：附加信息，如时间戳、来源等
}

// AgentInput 定义Agent执行时的输入参数
// 包含消息历史、流式处理配置和会话标识
type AgentInput struct {
	Messages        []*Message `json:"messages"`                   // 消息列表：对话历史和当前输入
	EnableStreaming bool       `json:"enable_streaming,omitempty"` // 流式处理：是否启用实时事件流
	SessionID       string     `json:"session_id,omitempty"`       // 会话ID：用于跟踪和关联多轮对话
}

// AgentEvent 表示Agent执行过程中产生的事件
// 提供完整的执行状态信息，支持实时监控和调试
type AgentEvent struct {
	AgentName string                 `json:"agent_name"`         // Agent名称：事件来源标识
	RunPath   []string               `json:"run_path,omitempty"` // 执行路径：显示调用层次结构
	Output    interface{}            `json:"output,omitempty"`   // 输出内容：处理结果或中间状态
	Action    *AgentAction           `json:"action,omitempty"`   // 执行动作：如任务转移、工具调用等
	Error     error                  `json:"error,omitempty"`    // 错误信息：异常情况的详细描述
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // 元数据：扩展信息和上下文数据
	Timestamp time.Time              `json:"timestamp"`          // 时间戳：事件发生的精确时间
}

// AgentAction 定义Agent可执行的动作类型
// 支持任务转移、工具调用、状态变更等操作
type AgentAction struct {
	Type   string      `json:"type"`             // 动作类型：transfer(转移)、invoke(调用)、exit(退出)等
	Target string      `json:"target,omitempty"` // 目标对象：转移的目标Agent或调用的工具名称
	Data   interface{} `json:"data,omitempty"`   // 动作数据：传递给目标的参数或状态信息
}

// AsyncIterator 异步迭代器，用于处理Agent执行过程中的事件流
// 基于Go泛型实现，支持类型安全的异步事件传递
// 核心特性：
// - 缓冲机制：避免生产者阻塞
// - 上下文感知：支持超时和取消操作
// - 线程安全：支持并发读写操作
type AsyncIterator[T any] struct {
	ch   chan T    // 事件通道：缓冲区大小为100，平衡内存使用和性能
	done chan bool // 完成信号：用于通知迭代器关闭
}

// NewAsyncIterator 创建新的异步迭代器实例
// 返回值：配置好缓冲区的迭代器，可立即使用
func NewAsyncIterator[T any]() *AsyncIterator[T] {
	return &AsyncIterator[T]{
		ch:   make(chan T, 100), // 100个事件的缓冲区，适合大多数场景
		done: make(chan bool),   // 无缓冲通道，确保关闭信号及时传递
	}
}

// Next 获取下一个事件，支持上下文控制
// 参数：
//   - ctx: 上下文，用于超时控制和取消操作
//
// 返回值：
//   - T: 事件数据
//   - bool: 是否成功获取（false表示迭代器已关闭或上下文取消）
func (ai *AsyncIterator[T]) Next(ctx context.Context) (T, bool) {
	select {
	case value, ok := <-ai.ch: // 从事件通道读取数据
		return value, ok
	case <-ai.done: // 迭代器已关闭
		var zero T
		return zero, false
	case <-ctx.Done(): // 上下文取消或超时
		var zero T
		return zero, false
	}
}

// Send 发送事件到迭代器
// 采用非阻塞模式，当缓冲区满时记录警告而不是阻塞
// 参数：
//   - value: 要发送的事件数据
func (ai *AsyncIterator[T]) Send(value T) {
	select {
	case ai.ch <- value: // 成功发送到缓冲区
	default: // 缓冲区已满，记录警告
		log.Printf("警告: AsyncIterator 缓冲区已满")
	}
}

// Close 关闭异步迭代器
// 释放所有资源，通知所有等待的Next调用返回
// 注意：关闭后不能再发送事件
func (ai *AsyncIterator[T]) Close() {
	close(ai.ch)   // 关闭事件通道
	close(ai.done) // 发送完成信号
}

// ============= 核心接口 =============

// Agent 定义了所有智能代理必须实现的基础接口
// 这是 Eino ADK 的核心抽象，支持统一的Agent管理和调用
// 设计原则：
// - 上下文感知：所有方法都接受context参数
// - 异步执行：Run方法返回事件流而非阻塞等待
// - 标准化：统一的接口便于Agent间的互操作
type Agent interface {
	// Name 返回Agent的唯一标识名称
	// 用于日志记录、事件追踪和调试
	// 参数：ctx - 执行上下文
	// 返回：Agent的名称字符串
	Name(ctx context.Context) string

	// Description 返回Agent的功能描述
	// 用于用户界面显示和自动化文档生成
	// 参数：ctx - 执行上下文
	// 返回：Agent功能的详细描述
	Description(ctx context.Context) string

	// Run 执行Agent的主要逻辑
	// 这是Agent的核心方法，处理输入并产生事件流
	// 参数：
	//   - ctx: 执行上下文，用于超时控制和取消操作
	//   - input: Agent输入，包含消息和配置
	//   - opts: 可选参数，支持扩展配置
	// 返回：异步事件迭代器，用于实时获取执行状态
	Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent]
}

// AgentTool 扩展Agent接口，支持Agent作为工具使用
// 这是多Agent组合模式中的关键接口，实现了"Agent作为工具"的设计模式
// 核心特性：
// - 同步调用：Invoke方法提供同步的工具调用接口
// - 参数化：支持灵活的参数传递
// - 双重身份：既可以作为独立Agent运行，也可以作为工具被调用
type AgentTool interface {
	Agent // 继承Agent的所有方法

	// Invoke 将Agent作为工具进行同步调用
	// 与Run方法不同，Invoke提供同步的、参数化的调用方式
	// 适用场景：
	// - 快速功能调用
	// - 嵌入到其他Agent的工作流中
	// - 需要立即获取结果的场景
	// 参数：
	//   - ctx: 执行上下文
	//   - params: 调用参数，键值对形式
	// 返回：
	//   - interface{}: 处理结果
	//   - error: 错误信息
	Invoke(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// ============= 专业Agent实现 =============

// DataAnalysisAgent 专业数据分析智能代理
// 实现了Agent和AgentTool双重接口，展示了"Agent作为工具"的设计模式
// 核心功能：
// - 统计分析：描述性统计、趋势分析
// - 数据处理：清洗、转换、聚合
// - 模式识别：异常检测、模式发现
// - 双重模式：既可独立运行，也可作为工具被调用
type DataAnalysisAgent struct {
	name string // Agent名称，用于标识和日志记录
}

// NewDataAnalysisAgent 创建新的数据分析Agent实例
// 返回值：配置完成的DataAnalysisAgent，可立即使用
func NewDataAnalysisAgent() *DataAnalysisAgent {
	return &DataAnalysisAgent{name: "DataAnalysisAgent"}
}

// Name 返回Agent的标识名称
// 实现Agent接口的Name方法
func (d *DataAnalysisAgent) Name(ctx context.Context) string {
	return d.name
}

// Description 返回Agent的功能描述
// 实现Agent接口的Description方法
// 提供详细的功能说明，便于用户理解和系统集成
func (d *DataAnalysisAgent) Description(ctx context.Context) string {
	return "专业数据分析Agent，提供统计分析、数据处理和模式识别服务"
}

// Run 执行数据分析的主要逻辑
// 实现Agent接口的Run方法，提供完整的异步数据分析流程
// 工作流程：
// 1. 数据加载和预处理
// 2. 描述性统计分析
// 3. 异常值检测
// 4. 模式识别和报告生成
// 特点：异步执行，实时事件反馈，支持上下文取消
func (d *DataAnalysisAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]()

	// 启动异步执行goroutine
	go func() {
		defer iter.Close() // 确保资源正确释放

		// 发送开始事件
		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Output:    "📊 数据分析Agent开始工作...",
			Timestamp: time.Now(),
		})

		// 模拟数据分析的各个阶段
		// 每个步骤都会发送进度事件，便于监控和调试
		steps := []string{
			"正在加载数据集...",   // 阶段1：数据加载
			"执行描述性统计分析...", // 阶段2：基础统计
			"检测数据异常值...",   // 阶段3：异常检测
			"生成分析报告...",    // 阶段4：报告生成
		}

		for _, step := range steps {
			time.Sleep(300 * time.Millisecond) // 模拟处理时间
			iter.Send(&AgentEvent{
				AgentName: d.name,
				RunPath:   []string{d.name},
				Output:    fmt.Sprintf("📈 %s", step),
				Timestamp: time.Now(),
			})
		}

		// 生成最终分析结果
		// 包含发现的模式、置信度和建议
		result := &Message{
			Role:    "assistant",
			Content: "数据分析完成：发现3个关键模式，置信度95%，建议进一步验证",
			Metadata: map[string]interface{}{
				"patterns_found": 3,                    // 发现的模式数量
				"confidence":     0.95,                 // 分析置信度
				"recommendation": "further_validation", // 后续建议
			},
		}

		// 发送分析结果事件
		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Output:    result,
			Timestamp: time.Now(),
		})

		// 发送任务完成事件
		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Action:    &AgentAction{Type: "exit", Data: "数据分析任务完成"},
			Timestamp: time.Now(),
		})
	}()

	return iter
}

// Invoke 实现AgentTool接口，提供同步的工具调用方式
// 这是"Agent作为工具"模式的核心实现
// 与Run方法的区别：
// - 同步执行：立即返回结果，不产生事件流
// - 参数化：接受结构化参数而非消息格式
// - 轻量级：适合快速调用和嵌入式使用
// 参数：
//   - ctx: 执行上下文
//   - params: 调用参数，期望包含"dataset"字段
//
// 返回：分析结果摘要和错误信息
func (d *DataAnalysisAgent) Invoke(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// 参数验证：检查必需的数据集参数
	dataset, ok := params["dataset"]
	if !ok {
		return "请提供数据集参数", nil
	}

	// 简化的工具调用响应
	// 在实际应用中，这里会执行真实的数据分析逻辑
	return map[string]interface{}{
		"status":   "success",                          // 执行状态
		"dataset":  dataset,                            // 处理的数据集
		"patterns": []string{"趋势上升", "季节性变化", "异常值检测"}, // 发现的模式
		"summary":  "数据分析完成，发现明显的上升趋势",                 // 分析摘要
	}, nil
}

// ReportAgent 专业报告生成智能代理
// 负责将分析结果转换为格式化的专业报告
// 核心功能：
// - 报告结构化：自动生成标准格式的报告
// - 内容整合：汇总多个数据源的分析结果
// - 格式支持：支持多种输出格式（Markdown、HTML等）
// - 可视化：集成图表和数据可视化元素
type ReportAgent struct {
	name string // Agent名称，用于标识和日志记录
}

// NewReportAgent 创建新的报告生成Agent实例
// 返回值：配置完成的ReportAgent，可立即使用
func NewReportAgent() *ReportAgent {
	return &ReportAgent{name: "ReportAgent"}
}

// Name 返回Agent的标识名称
// 实现Agent接口的Name方法
func (r *ReportAgent) Name(ctx context.Context) string {
	return r.name
}

// Description 返回Agent的功能描述
// 实现Agent接口的Description方法
// 提供详细的功能说明，便于用户理解和系统集成
func (r *ReportAgent) Description(ctx context.Context) string {
	return "专业报告生成Agent，基于分析结果生成格式化的专业报告"
}

// Run 执行报告生成的主要逻辑
// 实现Agent接口的Run方法，提供完整的异步报告生成流程
// 工作流程：
// 1. 解析输入数据和分析结果
// 2. 构建标准化报告结构
// 3. 生成图表和可视化元素
// 4. 格式化并输出最终报告
// 特点：支持多种报告格式，自动化排版，实时进度反馈
func (r *ReportAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]()

	// 启动异步执行goroutine
	go func() {
		defer iter.Close() // 确保资源正确释放

		// 发送开始事件
		iter.Send(&AgentEvent{
			AgentName: r.name,
			RunPath:   []string{r.name},
			Output:    "📝 报告Agent开始生成报告...",
			Timestamp: time.Now(),
		})

		// 模拟报告生成的各个阶段
		// 每个步骤都会发送进度事件，便于监控处理进度
		steps := []string{
			"解析输入数据...",   // 阶段1：数据解析和预处理
			"构建报告结构...",   // 阶段2：报告框架构建
			"生成图表和可视化...", // 阶段3：可视化元素生成
			"格式化最终报告...",  // 阶段4：格式化和美化
		}

		for _, step := range steps {
			time.Sleep(200 * time.Millisecond) // 模拟处理时间
			iter.Send(&AgentEvent{
				AgentName: r.name,
				RunPath:   []string{r.name},
				Output:    fmt.Sprintf("📋 %s", step),
				Timestamp: time.Now(),
			})
		}

		// 生成最终报告
		// 使用Markdown格式，包含完整的分析结果和建议
		report := `
# 数据分析报告

## 执行摘要
- 分析时间: ` + time.Now().Format("2006-01-02 15:04:05") + `
- 数据质量: 优秀
- 关键发现: 发现3个重要模式

## 详细分析
1. 📈 趋势分析：数据呈现稳定上升趋势
2. 🔍 异常检测：识别出2个异常数据点
3. 📊 统计结果：均值显著提升，方差保持稳定

## 建议行动
- 继续监控关键指标
- 深入调查异常数据点
- 制定优化策略
`

		// 发送报告生成完成事件
		iter.Send(&AgentEvent{
			AgentName: r.name,
			RunPath:   []string{r.name},
			Output: &Message{
				Role:    "assistant",
				Content: report,
				Metadata: map[string]interface{}{
					"report_type": "data_analysis", // 报告类型标识
					"format":      "markdown",      // 报告格式
				},
			},
			Timestamp: time.Now(),
		})

		// 发送任务完成事件
		iter.Send(&AgentEvent{
			AgentName: r.name,
			RunPath:   []string{r.name},
			Action:    &AgentAction{Type: "exit", Data: "报告生成完成"},
			Timestamp: time.Now(),
		})
	}()

	return iter
}

// ============= 复合Agent - 展示三种组合模式 =============

// BusinessAnalysisAgent 高级业务分析智能代理
// 这是多Agent组合模式的完整演示，集成了三种核心组合模式：
//
// 🔧 模式1: Agent作为工具使用
//   - 通过toolAgents映射管理工具Agent
//   - 使用Invoke方法进行同步调用
//   - 快速获取处理结果用于决策
//
// 🔄 模式2: Agent间任务转移
//   - 通过AgentAction实现任务流转
//   - 保持执行上下文和状态连续性
//   - 支持复杂工作流的分阶段执行
//
// 🏗️ 模式3: 层次化Agent结构
//   - 作为主Agent协调多个专业子Agent
//   - 通过RunPath追踪调用层次
//   - 实现复杂业务逻辑的模块化分解
type BusinessAnalysisAgent struct {
	name        string               // Agent名称标识
	dataAgent   Agent                // 数据分析子Agent，用于任务转移
	reportAgent Agent                // 报告生成子Agent，用于任务转移
	toolAgents  map[string]AgentTool // 工具Agent映射，用于工具调用模式
}

// NewBusinessAnalysisAgent 创建新的业务分析Agent实例
// 初始化所有子Agent和工具映射，展示Agent组合的最佳实践
// 返回值：完全配置的BusinessAnalysisAgent，包含所有必要的子组件
func NewBusinessAnalysisAgent() *BusinessAnalysisAgent {
	// 创建数据分析Agent，同时用作子Agent和工具Agent
	dataAgent := NewDataAnalysisAgent()
	return &BusinessAnalysisAgent{
		name:        "BusinessAnalysisAgent",
		dataAgent:   dataAgent,        // 用于任务转移模式
		reportAgent: NewReportAgent(), // 用于任务转移模式
		toolAgents: map[string]AgentTool{ // 用于工具调用模式
			"data_analysis": dataAgent, // 同一个Agent的双重用途
		},
	}
}

// Name 返回Agent的标识名称
// 实现Agent接口的Name方法
func (b *BusinessAnalysisAgent) Name(ctx context.Context) string {
	return b.name
}

// Description 返回Agent的功能描述
// 实现Agent接口的Description方法
// 强调其作为复合Agent的协调能力
func (b *BusinessAnalysisAgent) Description(ctx context.Context) string {
	return "高级业务分析Agent，协调多个专业Agent完成复杂的业务分析任务"
}

// Run 执行业务分析任务，展示三种Agent组合模式的完整实现
// 这是多Agent协作的核心方法，按顺序演示：
// 1. Agent作为工具的同步调用模式
// 2. Agent间任务转移的异步协作模式
// 3. 层次化Agent结构的统一管理模式
//
// 参数:
//   - ctx: 执行上下文，支持取消和超时控制
//   - input: 输入参数，包含分析请求和执行路径
//   - opts: 可选参数，支持扩展配置
//
// 返回值:
//   - AsyncIterator[*AgentEvent]: 异步事件流，实时反馈执行进度
//
// 工作流程:
//  1. 🔧 工具模式: 快速获取初步分析结果
//  2. 🔄 转移模式: 深度分析任务委托
//  3. 🏗️ 层次模式: 报告生成和结果整合
func (b *BusinessAnalysisAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	// 创建异步事件迭代器，用于实时反馈执行状态
	iter := NewAsyncIterator[*AgentEvent]()

	// 启动异步执行goroutine，避免阻塞调用方
	go func() {
		defer iter.Close() // 确保资源正确释放

		// 发送任务开始事件，标记业务分析流程启动
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Output:    "🏢 业务分析Agent开始执行复合任务...",
			Timestamp: time.Now(),
		})

		// ========== 模式1: Agent作为工具使用 ==========
		// 特点: 同步调用、快速响应、轻量级处理
		// 适用: 简单查询、数据验证、快速计算
		// 优势: 立即获取结果，无需等待异步处理
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Output:    "🔧 阶段1: 调用数据分析工具...",
			Timestamp: time.Now(),
		})

		// 同步调用工具Agent，获取即时分析结果
		// 这里展示了Agent双重身份的使用：既是独立Agent，也是工具
		toolResult, err := b.toolAgents["data_analysis"].Invoke(ctx, map[string]interface{}{
			"dataset": "business_metrics_2024.csv", // 业务指标数据集
		})

		// 错误处理：工具调用失败时的优雅处理
		if err != nil {
			iter.Send(&AgentEvent{
				AgentName: b.name,
				RunPath:   []string{b.name},
				Error:     err,
				Timestamp: time.Now(),
			})
			return // 提前退出，避免后续处理
		}

		// 发送工具调用成功事件，包含分析结果摘要
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Output:    fmt.Sprintf("✅ 工具调用结果: %v", toolResult),
			Timestamp: time.Now(),
		})

		// ========== 模式2: Agent间任务转移 ==========
		// 特点: 异步执行、状态保持、上下文传递
		// 适用: 复杂处理、多步骤任务、专业化分工
		// 优势: 保持执行连续性，支持复杂工作流
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Action: &AgentAction{
				Type:   "transfer",            // 任务转移类型
				Target: b.dataAgent.Name(ctx), // 目标Agent标识
				Data:   "转移详细数据分析任务",          // 转移的任务描述
			},
			Timestamp: time.Now(),
		})

		// 构建数据分析Agent的输入参数
		// 包含具体的分析请求和上下文信息
		dataInput := &AgentInput{
			Messages: []*Message{
				{
					Role:    "user",       // 用户角色消息
					Content: "请进行详细的数据分析", // 具体分析请求
				},
			},
		}

		// 执行数据分析Agent并处理事件流
		// 这里展示了层次化调用：主Agent调用子Agent
		dataIter := b.dataAgent.Run(ctx, dataInput)
		for {
			// 从子Agent获取执行事件
			event, ok := dataIter.Next(ctx)
			if !ok {
				break // 子Agent执行完成
			}
			if event != nil {
				// 更新执行路径，显示层次化调用关系
				// 格式: BusinessAnalysisAgent → DataAnalysisAgent
				event.RunPath = append([]string{b.name}, event.RunPath...)
				iter.Send(event) // 转发事件到主事件流
			}
		}

		// ========== 模式3: 层次化Agent结构 - 报告生成阶段 ==========
		// 特点: 主Agent协调、统一管理、结果整合
		// 适用: 复杂业务流程、多Agent协同、结果汇总
		// 优势: 模块化分工，统一的事件管理和路径追踪
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Action: &AgentAction{
				Type:   "transfer",              // 继续使用任务转移模式
				Target: b.reportAgent.Name(ctx), // 转移到报告生成Agent
				Data:   "转移报告生成任务",              // 任务描述
			},
			Timestamp: time.Now(),
		})

		// 构建报告生成Agent的输入参数
		// 基于前面数据分析的结果生成专业报告
		reportInput := &AgentInput{
			Messages: []*Message{
				{
					Role:    "user",         // 用户角色
					Content: "基于分析结果生成业务报告", // 报告生成请求
				},
			},
		}

		// 执行报告生成Agent并处理其事件流
		// 展示层次化结构中的第二个子Agent调用
		reportIter := b.reportAgent.Run(ctx, reportInput)
		for {
			// 获取报告生成Agent的执行事件
			event, ok := reportIter.Next(ctx)
			if !ok {
				break // 报告生成完成
			}
			if event != nil {
				// 维护层次化的执行路径
				// 格式: BusinessAnalysisAgent → ReportAgent
				event.RunPath = append([]string{b.name}, event.RunPath...)
				iter.Send(event) // 转发到主事件流
			}
		}

		// ========== 工作流程总结和结果整合 ==========
		// 作为主Agent，负责整合所有子Agent的执行结果
		// 提供完整的执行摘要和元数据信息
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Output: &Message{
				Role:    "assistant",                   // 助手角色回复
				Content: "🎉 业务分析完成！已完成数据分析、报告生成和结果汇总。", // 成功完成消息
				Metadata: map[string]interface{}{
					// 记录使用的三种组合模式
					"workflow_stages": []string{"tool_usage", "task_transfer", "hierarchical_execution"},
					// 记录参与的所有Agent
					"agents_involved": []string{"DataAnalysisAgent", "ReportAgent"},
				},
			},
			Timestamp: time.Now(),
		})

		// 发送流程完成事件，标记整个业务分析任务结束
		// 使用exit动作类型，表示正常完成退出
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Action: &AgentAction{
				Type: "exit",     // 退出动作类型
				Data: "业务分析流程完成", // 完成状态描述
			},
			Timestamp: time.Now(),
		})
	}()

	return iter
}

// ============= 演示函数 - 展示三种Agent组合模式的实际应用 =============

// demonstrateAgentAsTools 演示Agent作为工具的使用模式
// 这个函数展示了同一个Agent的双重用途：
// 1. 作为工具(AgentTool)进行同步调用 - 快速获取结果
// 2. 作为Agent进行异步执行 - 完整的事件流处理
//
// 核心特点:
// - 🔧 工具模式: 同步、轻量、即时响应
// - 🤖 Agent模式: 异步、完整、事件驱动
// - 📊 对比展示: 同一功能的不同调用方式
//
// 适用场景:
// - 简单查询时使用工具模式
// - 复杂处理时使用Agent模式
// - 需要事件监控时使用Agent模式
func demonstrateAgentAsTools() {
	fmt.Println("🎯 模式1: Agent作为工具使用")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()
	// 创建数据分析Agent实例
	// 注意：同一个Agent可以同时实现Agent和AgentTool接口
	dataAgent := NewDataAnalysisAgent()

	fmt.Printf("🔧 将 %s 作为工具调用\n", dataAgent.Name(ctx))

	// ========== 模式A: 作为工具使用 - 同步调用 ==========
	// 特点: 立即返回结果，无事件流，适合快速查询
	// 直接作为工具调用，获取即时分析结果
	result, err := dataAgent.Invoke(ctx, map[string]interface{}{
		"dataset": "sample_data.csv", // 示例数据集
	})

	// 错误处理：工具调用失败时的优雅处理
	if err != nil {
		fmt.Printf("❌ 工具调用失败: %v\n", err)
		return
	}

	// 展示工具调用的结构化结果
	fmt.Printf("✅ 工具调用成功:\n")
	if resultMap, ok := result.(map[string]interface{}); ok {
		for key, value := range resultMap {
			fmt.Printf("   %s: %v\n", key, value)
		}
	}

	// ========== 对比说明 ==========
	fmt.Println("\n💡 工具模式特点:")
	fmt.Println("   • 同步执行，立即返回结果")
	fmt.Println("   • 轻量级调用，适合简单任务")
	fmt.Println("   • 无事件流，无法监控执行过程")
	fmt.Println("   • 适用于快速查询和数据验证")
}

// demonstrateHierarchicalComposition 演示层次化Agent组合模式
// 这个函数展示了复杂的多Agent协作场景：
// 1. 主Agent协调多个专业子Agent
// 2. 任务在不同Agent间流转
// 3. 统一的事件管理和路径追踪
//
// 核心特点:
// - 🏗️ 层次结构: 主Agent -> 子Agent的调用层次
// - 🔄 任务流转: 通过AgentAction实现任务转移
// - 📡 事件聚合: 统一收集和处理所有子Agent事件
// - 🛤️ 路径追踪: 完整记录执行调用链
//
// 适用场景:
// - 复杂业务流程需要多个专业Agent协作
// - 需要统一管理和监控多Agent执行状态
// - 要求保持完整的执行上下文和调用链
func demonstrateHierarchicalComposition() {
	fmt.Println("\n🎯 模式2&3: 层次化结构 + 任务转移")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()
	// 创建业务分析Agent（集成了数据分析和报告生成子Agent）
	// 这是一个复合Agent，内部管理多个专业Agent
	businessAgent := NewBusinessAnalysisAgent()

	fmt.Printf("🏢 执行业务分析Agent: %s\n", businessAgent.Name(ctx))
	fmt.Printf("📝 描述: %s\n", businessAgent.Description(ctx))
	fmt.Println()

	// 构建复合任务输入
	// 这个任务将触发多个子Agent的协作执行
	input := &AgentInput{
		Messages: []*Message{
			{
				Role:    "user",
				Content: "请执行完整的业务数据分析，包括数据处理和报告生成",
			},
		},
		EnableStreaming: true,               // 启用流式处理，实时获取事件
		SessionID:       "demo_session_001", // 会话标识，用于跟踪
	}

	fmt.Println("▶️  开始执行复合Agent工作流...")
	fmt.Println("📊 将展示以下组合模式:")
	fmt.Println("   🔧 工具调用: 快速获取初步分析")
	fmt.Println("   🔄 任务转移: 深度分析委托")
	fmt.Println("   🏗️ 层次管理: 统一协调和监控")
	fmt.Println()

	// 启动层次化Agent执行
	// 这将触发：工具调用 -> 任务转移 -> 层次执行的完整流程
	iter := businessAgent.Run(ctx, input)

	// ========== 事件流处理循环 ==========
	// 实时监控和展示多Agent协作的完整执行过程
	// 每个事件都包含丰富的上下文信息：时间戳、执行路径、输出内容、动作类型等
	for {
		event, ok := iter.Next(ctx)
		if !ok {
			break // 所有Agent执行完成，退出监控循环
		}

		if event != nil {
			// 格式化显示执行路径，展示层次化调用关系
			// 例如: BusinessAnalysisAgent → DataAnalysisAgent
			runPathStr := strings.Join(event.RunPath, " → ")
			fmt.Printf("📡 [%s] %s: ", event.Timestamp.Format("15:04:05"), runPathStr)

			// ========== 处理输出内容 ==========
			if event.Output != nil {
				if msg, ok := event.Output.(*Message); ok {
					// 智能处理长文本输出（如报告）
					// 对于长报告，只显示前几行，避免输出过长
					lines := strings.Split(msg.Content, "\n")
					if len(lines) > 8 {
						fmt.Printf("💬 %s...\n", strings.Join(lines[:8], "\n"))
						fmt.Printf("    📄 (完整报告已生成，共%d行)\n", len(lines))
					} else {
						fmt.Printf("💬 %s\n", msg.Content)
					}
				} else {
					// 处理其他类型的输出（状态信息、进度更新等）
					fmt.Printf("ℹ️  %v\n", event.Output)
				}
			}

			// ========== 处理Agent动作 ==========
			// 显示Agent间的交互动作：任务转移、工具调用、退出等
			if event.Action != nil {
				fmt.Printf("🎬 动作: %s", event.Action.Type)
				if event.Action.Target != "" {
					// 显示动作目标（转移到哪个Agent）
					fmt.Printf(" → %s", event.Action.Target)
				}
				if event.Action.Data != nil {
					// 显示动作携带的数据
					fmt.Printf(" (%v)", event.Action.Data)
				}
				fmt.Println()
			}

			// ========== 错误处理 ==========
			// 显示执行过程中的任何错误信息
			if event.Error != nil {
				fmt.Printf("❌ 错误: %v\n", event.Error)
			}
		}
	}

	// ========== 执行完成总结 ==========
	fmt.Println("\n✅ 层次化Agent组合演示完成!")
	fmt.Println("💡 关键特性展示:")
	fmt.Println("   🔧 工具调用: 同步获取快速分析结果")
	fmt.Println("   🔄 任务转移: 异步委托专业Agent处理")
	fmt.Println("   🏗️ 层次管理: 统一协调多Agent协作")
	fmt.Println("   📡 事件流: 实时监控整个执行过程")
	fmt.Println("   🛤️ 路径追踪: 完整记录调用层次关系")
}

// main 多Agent组合模式演示程序入口
// 这是Eino ADK框架中Agent组合能力的完整展示程序
//
// 🎯 演示目标:
//
//	展示三种核心的Agent组合模式，帮助开发者理解如何构建
//	复杂的多Agent协作系统
//
// 📋 演示内容:
//  1. 🔧 Agent作为工具使用 - 同步调用，快速响应
//  2. 🔄 Agent间任务转移 - 异步协作，状态保持
//  3. 🏗️ 层次化Agent结构 - 统一管理，模块化分工
//
// 🏗️ 架构特点:
//   - 标准化接口: Agent和AgentTool接口的统一设计
//   - 事件驱动: 基于AsyncIterator的异步事件流
//   - 上下文感知: 完整的执行路径和状态追踪
//   - 灵活组合: 支持多种Agent组合和调用模式
//
// 💡 学习价值:
//
//	通过实际代码演示，理解如何设计和实现可扩展的
//	多Agent系统，掌握Agent组合的最佳实践
func main() {
	// ========== 程序启动和介绍 ==========
	fmt.Println("🎊 Eino ADK 多Agent组合模式完整演示")
	fmt.Println("基于官方文档的三种Agent协作模式")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("💡 本演示将展示Agent组合的核心技术和最佳实践")
	fmt.Println("📚 每种模式都有详细的实现和应用场景说明")
	fmt.Println()

	// ========== 演示1: Agent作为工具使用 ==========
	// 展示同一个Agent的双重身份使用方式
	// 重点：同步调用的便利性和即时响应特性
	demonstrateAgentAsTools()

	// 添加适当的间隔，便于观察演示效果
	time.Sleep(time.Second)

	// ========== 演示2: 层次化Agent组合 ==========
	// 展示复杂的多Agent协作场景
	// 重点：任务转移、层次管理、事件聚合的完整流程
	demonstrateHierarchicalComposition()

	// ========== 演示总结和技术要点分析 ==========
	fmt.Println("\n🎯 多Agent组合模式总结")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("✨ 成功演示的组合模式:")

	// 模式1总结：Agent作为工具的优势和适用场景
	fmt.Println("  🔧 模式1: Agent作为工具使用")
	fmt.Println("    - 直接调用Agent.Invoke()方法")
	fmt.Println("    - 同步返回处理结果")
	fmt.Println("    - 适用于简单的功能调用")
	fmt.Println("    - 轻量级，无事件流开销")

	// 模式2总结：任务转移的协作机制
	fmt.Println("  🔄 模式2: Agent间任务转移")
	fmt.Println("    - 使用AgentAction.Type='transfer'")
	fmt.Println("    - 指定Target目标Agent")
	fmt.Println("    - 保持执行上下文连续性")
	fmt.Println("    - 支持复杂工作流的分阶段执行")

	// 模式3总结：层次化结构的管理优势
	fmt.Println("  🏗️  模式3: 层次化Agent结构")
	fmt.Println("    - 主Agent协调多个子Agent")
	fmt.Println("    - RunPath显示调用层次")
	fmt.Println("    - 事件流统一管理")
	fmt.Println("    - 模块化分工，职责清晰")

	// 核心技术特性总结
	fmt.Println("\n💡 关键技术特性:")
	fmt.Println("  • 异步事件流处理 - 实时反馈执行状态")
	fmt.Println("  • 灵活的Agent组合策略 - 支持多种协作模式")
	fmt.Println("  • 统一的接口抽象 - Agent和AgentTool的无缝集成")
	fmt.Println("  • 可追溯的执行路径 - 完整的调用链记录")
	fmt.Println("  • 松耦合的模块化设计 - 易于扩展和维护")
	fmt.Println("  • 上下文感知机制 - 支持取消、超时等控制")
	fmt.Println("  • 类型安全的泛型设计 - 基于Go泛型的强类型支持")

	// 应用价值和发展前景
	fmt.Println("\n🚀 应用价值:")
	fmt.Println("  • 构建复杂的智能Agent系统")
	fmt.Println("  • 实现可扩展的多Agent协作平台")
	fmt.Println("  • 支持企业级的AI应用开发")
	fmt.Println("  • 提供标准化的Agent组合解决方案")

	fmt.Println("\n🎉 这就是Eino ADK的真正威力！")
	fmt.Println("💪 让我们一起构建更智能的Agent系统！")
}
