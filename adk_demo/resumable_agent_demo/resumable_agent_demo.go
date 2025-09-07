/*
ResumableAgent 中断恢复演示

本文件演示了 Eino ADK 的企业级 Agent 扩展机制，重点展示：

🎯 核心特性：
  - ResumableAgent 接口：支持中断和恢复的 Agent 扩展
  - Runner 生命周期管理：统一的 Agent 执行和状态管理
  - CheckPointStore 机制：灵活的状态持久化存储
  - 中断驱动架构：基于事件的中断处理和恢复流程

🏗️ 架构设计：
  - 扩展标准 Agent 接口，增加 Resume() 方法
  - 通过 AgentAction.Type="interrupt" 标准化中断机制
  - 使用 Gob 序列化支持复杂状态数据的持久化
  - 基于 AsyncIterator 的统一事件流处理

💼 企业级应用价值：
  - 长时间运行任务的可靠性保障（如文档处理、数据分析）
  - 支持人机协作的中断决策（如敏感信息审核）
  - 跨系统重启的状态持久化（如服务器维护后恢复）
  - 复杂业务流程的断点续传（如多步骤工作流）

📋 演示场景：

	文档处理 Agent 在执行多步骤任务时，在敏感信息检测步骤触发中断，
	等待人工确认后从中断点恢复执行，完成剩余处理步骤。

基于 https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_extension/
严格按照官方扩展机制实现中断恢复功能
*/
package main

import (
	"bytes"         // 字节缓冲区操作，用于序列化数据
	"context"       // 上下文管理，支持取消和超时
	"encoding/gob"  // Go 二进制编码，用于复杂数据结构序列化
	"fmt"           // 格式化输出
	"log"           // 日志记录
	"os"            // 操作系统接口
	"path/filepath" // 文件路径操作
	"strings"       // 字符串操作
	"time"          // 时间处理
)

// init 初始化函数
// 注册 gob 类型以支持复杂数据结构的序列化和反序列化
// 这对于 CheckPointStore 的状态持久化至关重要
func init() {
	gob.Register(map[string]interface{}{})   // 注册通用映射类型
	gob.Register(&InterruptInfo{})           // 注册中断信息结构
	gob.Register(&DocumentProcessingState{}) // 注册文档处理状态结构
}

// ============= 核心扩展机制类型定义 =============

// Message 消息结构
// 标准化的消息格式，支持角色、内容和元数据
// 兼容多种 AI 模型的消息格式规范
type Message struct {
	Role     string                 `json:"role"`               // 消息角色："user", "assistant", "system"
	Content  string                 `json:"content"`            // 消息内容
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 扩展元数据，支持自定义字段
}

// AgentInput Agent 输入结构
// 封装了 Agent 执行所需的所有输入参数
// 支持流式处理和会话管理
type AgentInput struct {
	Messages        []*Message `json:"messages"`                   // 消息列表，支持多轮对话
	EnableStreaming bool       `json:"enable_streaming,omitempty"` // 是否启用流式输出
	SessionID       string     `json:"session_id,omitempty"`       // 会话标识，用于状态管理和恢复
}

// AgentEvent Agent 事件结构
// 统一的事件格式，承载 Agent 执行过程中的所有信息
// 支持输出、动作、错误和元数据的完整传递
type AgentEvent struct {
	AgentName string                 `json:"agent_name"`         // Agent 名称
	RunPath   []string               `json:"run_path,omitempty"` // 执行路径，支持嵌套 Agent 追踪
	Output    interface{}            `json:"output,omitempty"`   // 输出内容，可以是任意类型
	Action    *AgentAction           `json:"action,omitempty"`   // 动作指令，如退出、转移、中断
	Error     error                  `json:"error,omitempty"`    // 错误信息
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // 事件元数据
	Timestamp time.Time              `json:"timestamp"`          // 事件时间戳
}

// AgentAction Agent 动作结构
// 定义 Agent 可以执行的标准化动作类型
// 支持退出、转移和中断等核心操作
type AgentAction struct {
	Type   string      `json:"type"`             // 动作类型："exit", "transfer", "interrupt"
	Target string      `json:"target,omitempty"` // 目标 Agent（用于转移）
	Data   interface{} `json:"data,omitempty"`   // 动作携带的数据
}

// InterruptInfo 中断信息结构 - 官方接口
// 封装中断时需要保存的状态数据
// 支持任意复杂数据结构的序列化保存
type InterruptInfo struct {
	Data interface{} `json:"data"` // 中断时的状态数据，通常是 Agent 的内部状态
}

// CheckPointStore 检查点存储接口 - 官方接口
// 定义状态持久化的标准接口，支持多种存储后端
// 核心功能：Set(保存)、Get(读取)、Delete(清理)
type CheckPointStore interface {
	Set(ctx context.Context, key string, value []byte) error   // 保存检查点数据
	Get(ctx context.Context, key string) ([]byte, bool, error) // 读取检查点数据，返回数据、是否存在、错误
	Delete(ctx context.Context, key string) error              // 删除检查点数据
}

// AsyncIterator 异步迭代器
// 基于 Go 泛型的类型安全异步迭代器实现
// 核心特性：
//   - 支持任意类型 T 的异步数据流
//   - 内置缓冲机制，避免阻塞
//   - 上下文感知，支持取消操作
//   - 线程安全的并发访问
//
// 适用场景：Agent 事件流、实时数据传输、异步任务结果收集
type AsyncIterator[T any] struct {
	ch   chan T    // 数据通道，缓冲大小为100
	done chan bool // 完成信号通道
}

// NewAsyncIterator 创建新的异步迭代器
// 返回一个已初始化的 AsyncIterator 实例
// 内置100个元素的缓冲区，平衡内存使用和性能
func NewAsyncIterator[T any]() *AsyncIterator[T] {
	return &AsyncIterator[T]{
		ch:   make(chan T, 100), // 缓冲通道，避免发送阻塞
		done: make(chan bool),   // 完成信号通道
	}
}

// Next 获取下一个元素
// 参数：
//
//	ctx - 上下文，支持取消和超时
//
// 返回值：
//
//	T - 下一个元素
//	bool - 是否成功获取（false表示迭代器已关闭或上下文取消）
//
// 特性：
//   - 阻塞等待直到有数据或迭代器关闭
//   - 响应上下文取消信号
//   - 线程安全
func (ai *AsyncIterator[T]) Next(ctx context.Context) (T, bool) {
	select {
	case value, ok := <-ai.ch: // 接收到数据
		return value, ok
	case <-ai.done: // 迭代器已关闭
		var zero T
		return zero, false
	case <-ctx.Done(): // 上下文取消
		var zero T
		return zero, false
	}
}

// Send 发送元素到迭代器
// 参数：
//
//	value - 要发送的元素
//
// 特性：
//   - 非阻塞发送，缓冲区满时记录警告
//   - 线程安全
//   - 适用于高频数据发送场景
func (ai *AsyncIterator[T]) Send(value T) {
	select {
	case ai.ch <- value: // 成功发送
	default:
		log.Printf("警告: AsyncIterator 缓冲区已满") // 缓冲区满时记录警告
	}
}

// Close 关闭迭代器
// 关闭数据通道和完成信号通道，通知所有等待的 Next() 调用返回
// 特性：
//   - 释放所有相关资源
//   - 通知等待中的接收方
func (ai *AsyncIterator[T]) Close() {
	close(ai.ch)   // 关闭数据通道
	close(ai.done) // 关闭完成信号通道
}

// ============= 核心Agent接口 =============

// Agent 基础 Agent 接口
// 定义了所有 Agent 必须实现的核心方法
// 基于 Eino ADK 官方规范设计
// 核心方法：
//   - Name() - 返回 Agent 名称，用于标识和日志
//   - Description() - 返回 Agent 描述，用于文档和调试
//   - Run() - 执行 Agent 主要逻辑，返回事件流
type Agent interface {
	Name(ctx context.Context) string                                                             // 获取 Agent 名称
	Description(ctx context.Context) string                                                      // 获取 Agent 描述
	Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] // 执行 Agent 逻辑
}

// ResumableAgent 可恢复 Agent 接口 - 官方扩展接口
// 扩展标准 Agent 接口，增加中断恢复能力
// 核心特性：
//   - 继承 Agent 的所有基础功能
//   - 新增 Resume() 方法支持从中断点恢复执行
//   - 支持复杂状态的序列化和反序列化
//   - 适用于长时间运行的任务和人机协作场景
//
// 应用场景：
//   - 文档处理：支持大文件分步处理和中断恢复
//   - 数据分析：支持长时间计算任务的断点续传
//   - 工作流引擎：支持复杂业务流程的状态管理
type ResumableAgent interface {
	Agent                                                                                                                         // 继承基础 Agent 接口
	Resume(ctx context.Context, interruptInfo *InterruptInfo, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] // 从中断点恢复执行
}

// ============= 检查点存储实现 =============

// FileCheckPointStore 基于文件的检查点存储实现
// 提供简单可靠的文件系统持久化方案
// 核心特性：
//   - 基于文件系统的持久化存储
//   - 自动创建目录结构
//   - 支持并发读写操作
//   - 简单的键值存储模型
//
// 适用场景：
//   - 开发和测试环境
//   - 单机部署场景
//   - 小规模状态管理
type FileCheckPointStore struct {
	baseDir string // 检查点文件存储的基础目录
}

// NewFileCheckPointStore 创建新的文件检查点存储
// 参数：
//
//	baseDir - 检查点文件存储的基础目录
//
// 返回：
//
//	*FileCheckPointStore - 文件检查点存储实例
func NewFileCheckPointStore(baseDir string) *FileCheckPointStore {
	return &FileCheckPointStore{baseDir: baseDir}
}

// Set 保存检查点数据到文件
// 参数：
//
//	ctx - 上下文
//	key - 检查点键名
//	value - 要保存的数据（已序列化）
//
// 返回：
//
//	error - 保存过程中的错误
//
// 实现细节：
//   - 自动创建目录结构
//   - 文件名格式：{key}.checkpoint
//   - 文件权限：0644（用户读写，组和其他用户只读）
func (f *FileCheckPointStore) Set(ctx context.Context, key string, value []byte) error {
	dir := filepath.Join(f.baseDir, "checkpoints") // 构建检查点目录路径
	if err := os.MkdirAll(dir, 0755); err != nil { // 确保目录存在
		return err
	}
	return os.WriteFile(filepath.Join(dir, key+".checkpoint"), value, 0644) // 写入检查点文件
}

// Get 从文件读取检查点数据
// 参数：
//
//	ctx - 上下文
//	key - 检查点键名
//
// 返回：
//
//	[]byte - 检查点数据
//	bool - 是否存在该检查点
//	error - 读取过程中的错误
//
// 特性：
//   - 区分文件不存在和读取错误
//   - 返回原始字节数据，需要调用方反序列化
func (f *FileCheckPointStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	filePath := filepath.Join(f.baseDir, "checkpoints", key+".checkpoint") // 构建文件路径
	data, err := os.ReadFile(filePath)                                     // 读取文件
	if os.IsNotExist(err) {                                                // 文件不存在
		return nil, false, nil
	}
	return data, true, err // 返回数据、存在标志、错误
}

// Delete 删除检查点文件
// 参数：
//
//	ctx - 上下文
//	key - 检查点键名
//
// 返回：
//
//	error - 删除过程中的错误
//
// 用途：
//   - 清理已完成任务的检查点
//   - 释放存储空间
func (f *FileCheckPointStore) Delete(ctx context.Context, key string) error {
	filePath := filepath.Join(f.baseDir, "checkpoints", key+".checkpoint") // 构建文件路径
	return os.Remove(filePath)                                             // 删除文件
}

// ============= Runner 实现 - 官方Runner机制 =============

// RunnerConfig Runner 配置结构
// 定义 Runner 的行为参数和依赖组件
// 配置项说明：
//   - CheckPointStore - 检查点存储后端
//   - EnableCallback - 是否启用回调机制
//   - MaxRetries - 最大重试次数
type RunnerConfig struct {
	CheckPointStore CheckPointStore // 检查点存储接口实现
	EnableCallback  bool            // 是否启用事件回调
	MaxRetries      int             // 执行失败时的最大重试次数
}

// Runner Agent 执行器 - 管理整个Agent生命周期
// 负责 Agent 的生命周期管理和中断恢复机制
// 核心功能：
//   - Agent 执行管理：启动、监控、停止
//   - 中断处理：检测中断事件，保存状态
//   - 恢复机制：从检查点恢复 Agent 执行
//   - 会话管理：支持多个并发 Agent 会话
//
// 设计特点：
//   - 基于事件驱动的异步架构
//   - 支持复杂状态的序列化保存
//   - 提供统一的错误处理和重试机制
type Runner struct {
	config          *RunnerConfig              // Runner 配置
	checkpointStore CheckPointStore            // 检查点存储实例
	runningAgents   map[string]*AgentExecution // 当前运行的 Agent 会话映射
}

// AgentExecution Agent 执行状态
// 记录单个 Agent 会话的执行信息和状态
// 状态字段说明：
//   - Agent - Agent 实例引用
//   - SessionID - 会话唯一标识
//   - CheckPointID - 检查点标识
//   - StartTime - 执行开始时间
//   - Status - 执行状态（running/interrupted/completed/error）
type AgentExecution struct {
	Agent        Agent     // Agent 实例
	SessionID    string    // 会话ID，用于状态管理和恢复
	CheckPointID string    // 检查点ID，用于状态持久化
	StartTime    time.Time // 执行开始时间
	Status       string    // 执行状态："running", "interrupted", "completed", "error"
}

// NewRunner 创建新的 Runner 实例
// 参数：
//
//	config - Runner 配置
//
// 返回：
//
//	*Runner - Runner 实例
//
// 初始化：
//   - 设置配置参数
//   - 初始化检查点存储
//   - 创建 Agent 执行状态映射
func NewRunner(config *RunnerConfig) *Runner {
	return &Runner{
		config:          config,                           // 保存配置
		checkpointStore: config.CheckPointStore,           // 设置检查点存储
		runningAgents:   make(map[string]*AgentExecution), // 初始化执行状态映射
	}
}

// Execute 执行 Agent
// 启动新的 Agent 会话，提供完整的生命周期管理
// 参数：
//
//	ctx - 执行上下文，用于取消和超时控制
//	agent - 要执行的 Agent 实例
//	input - Agent 输入参数
//
// 返回：
//
//	*AsyncIterator[*AgentEvent] - 异步事件迭代器，用于接收执行过程中的事件
//
// 执行流程：
//  1. 创建或使用现有会话ID
//  2. 初始化执行状态并注册到运行映射
//  3. 启动异步执行协程
//  4. 处理 Agent 事件流，包括中断检测
//  5. 管理会话完成和清理
func (r *Runner) Execute(ctx context.Context, agent Agent, input *AgentInput) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]() // 创建异步事件迭代器
	sessionID := input.SessionID            // 获取会话ID
	if sessionID == "" {
		// 如果没有提供会话ID，生成唯一会话ID
		sessionID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}

	// 创建执行状态记录
	execution := &AgentExecution{
		Agent:        agent,                                            // Agent 实例
		SessionID:    sessionID,                                        // 会话标识
		CheckPointID: fmt.Sprintf("%s_%s", agent.Name(ctx), sessionID), // 检查点标识
		StartTime:    time.Now(),                                       // 开始时间
		Status:       "running",                                        // 初始状态
	}

	r.runningAgents[sessionID] = execution // 注册到运行映射

	go func() {
		defer iter.Close()
		defer func() {
			delete(r.runningAgents, sessionID)
		}()

		// 发送Runner启动事件
		iter.Send(&AgentEvent{
			AgentName: "Runner",
			RunPath:   []string{"Runner"},
			Output:    fmt.Sprintf("🚀 Runner开始执行Agent: %s", agent.Name(ctx)),
			Timestamp: time.Now(),
		})

		// 执行Agent并处理中断
		agentIter := agent.Run(ctx, input)
		for {
			event, ok := agentIter.Next(ctx)
			if !ok {
				execution.Status = "completed"
				break
			}

			if event != nil {
				// 检查是否为中断事件
				if event.Action != nil && event.Action.Type == "interrupt" {
					execution.Status = "interrupted"

					// 处理中断逻辑
					if err := r.handleInterrupt(ctx, execution, event); err != nil {
						iter.Send(&AgentEvent{
							AgentName: "Runner",
							RunPath:   []string{"Runner"},
							Error:     fmt.Errorf("处理中断失败: %w", err),
							Timestamp: time.Now(),
						})
						execution.Status = "error"
						break
					}

					// 转发中断事件
					iter.Send(event)

					iter.Send(&AgentEvent{
						AgentName: "Runner",
						RunPath:   []string{"Runner"},
						Output:    "⏸️  Runner处理了Agent中断，状态已保存",
						Timestamp: time.Now(),
					})

					break
				}

				// 转发普通事件
				iter.Send(event)
			}
		}

		// 发送Runner完成事件
		iter.Send(&AgentEvent{
			AgentName: "Runner",
			RunPath:   []string{"Runner"},
			Output:    fmt.Sprintf("🏁 Runner执行完成，状态: %s", execution.Status),
			Timestamp: time.Now(),
		})
	}()

	return iter
}

// Resume 恢复中断的 Agent 执行
// 从检查点恢复之前中断的 Agent 会话
// 参数：
//
//	ctx - 执行上下文，用于取消和超时控制
//	sessionID - 要恢复的会话标识
//	newInput - 恢复时的新输入参数（可能包含用户决策）
//
// 返回：
//
//	*AsyncIterator[*AgentEvent] - 异步事件迭代器，用于接收恢复过程中的事件
//
// 恢复流程：
//  1. 从检查点存储中读取中断状态
//  2. 反序列化中断信息
//  3. 重新创建 Agent 实例
//  4. 调用 Agent 的 Resume 方法继续执行
//  5. 清理检查点并完成恢复
func (r *Runner) Resume(ctx context.Context, sessionID string, newInput *AgentInput) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]() // 创建异步事件迭代器

	go func() {
		defer iter.Close() // 确保迭代器关闭

		// 发送恢复开始事件
		iter.Send(&AgentEvent{
			AgentName: "Runner",
			RunPath:   []string{"Runner"},
			Output:    fmt.Sprintf("🔄 Runner开始恢复会话: %s", sessionID),
			Timestamp: time.Now(),
		})

		// 构建检查点ID并从存储中读取
		checkPointID := fmt.Sprintf("agent_%s", sessionID)
		data, exists, err := r.checkpointStore.Get(ctx, checkPointID)
		if err != nil {
			// 检查点读取失败
			iter.Send(&AgentEvent{
				AgentName: "Runner",
				RunPath:   []string{"Runner"},
				Error:     fmt.Errorf("读取检查点失败: %w", err),
				Timestamp: time.Now(),
			})
			return
		}

		if !exists {
			// 检查点不存在
			iter.Send(&AgentEvent{
				AgentName: "Runner",
				RunPath:   []string{"Runner"},
				Error:     fmt.Errorf("未找到会话 %s 的检查点", sessionID),
				Timestamp: time.Now(),
			})
			return
		}

		// 反序列化中断信息
		var interruptInfo InterruptInfo
		buf := bytes.NewBuffer(data)   // 创建字节缓冲区
		decoder := gob.NewDecoder(buf) // 创建 Gob 解码器
		if err := decoder.Decode(&interruptInfo); err != nil {
			// 反序列化失败
			iter.Send(&AgentEvent{
				AgentName: "Runner",
				RunPath:   []string{"Runner"},
				Error:     fmt.Errorf("反序列化检查点失败: %w", err),
				Timestamp: time.Now(),
			})
			return
		}

		// 检查点恢复成功
		iter.Send(&AgentEvent{
			AgentName: "Runner",
			RunPath:   []string{"Runner"},
			Output:    "✅ 检查点恢复成功，准备恢复Agent执行",
			Timestamp: time.Now(),
		})

		// 重新创建 Agent 实例
		// 注意：在实际应用中，这里应该根据检查点中的信息
		// 动态创建对应类型的 Agent 实例
		agent := NewDocumentProcessingAgent(r.checkpointStore)

		// 调用 ResumableAgent 的 Resume 方法恢复执行
		agentIter := agent.Resume(ctx, &interruptInfo, newInput)
		for {
			event, ok := agentIter.Next(ctx) // 获取 Agent 事件
			if !ok {
				break // Agent 执行完成
			}
			if event != nil {
				iter.Send(event) // 转发事件
			}
		}

		// 清理检查点文件
		r.checkpointStore.Delete(ctx, checkPointID)

		// 发送恢复完成事件
		iter.Send(&AgentEvent{
			AgentName: "Runner",
			RunPath:   []string{"Runner"},
			Output:    "🎉 Runner恢复执行完成",
			Timestamp: time.Now(),
		})
	}()

	return iter // 返回异步迭代器
}

// handleInterrupt 处理中断事件
// 当 Agent 执行过程中遇到需要用户决策的情况时调用
// 负责保存当前执行状态到检查点，以便后续恢复
// 参数：
//
//	ctx - 执行上下文
//	execution - 当前 Agent 执行状态
//	event - 中断事件，包含中断时的状态信息
//
// 返回：
//
//	error - 处理过程中的错误
//
// 处理流程：
//  1. 构建中断信息结构
//  2. 序列化中断状态
//  3. 保存到检查点存储
//  4. 更新执行状态
func (r *Runner) handleInterrupt(ctx context.Context, execution *AgentExecution, event *AgentEvent) error {
	// 构建中断信息结构，封装需要保存的状态数据
	interruptInfo := &InterruptInfo{
		Data: event.Action.Data, // 从中断事件中提取状态数据
	}

	// 使用 Gob 编码序列化中断信息
	var buf bytes.Buffer            // 创建字节缓冲区
	encoder := gob.NewEncoder(&buf) // 创建 Gob 编码器
	if err := encoder.Encode(interruptInfo); err != nil {
		// 序列化失败，返回详细错误信息
		return fmt.Errorf("序列化中断信息失败: %w", err)
	}

	// 构建检查点ID并保存到存储
	checkPointID := fmt.Sprintf("agent_%s", execution.SessionID) // 使用会话ID构建唯一检查点标识
	return r.checkpointStore.Set(ctx, checkPointID, buf.Bytes()) // 保存序列化后的检查点数据
}

// ============= ResumableAgent 实现示例 =============

// DocumentProcessingState 文档处理状态结构
// 封装文档处理过程中的所有状态信息，支持序列化和反序列化
// 用于中断恢复机制的状态持久化
// 状态字段说明：
//   - CurrentStep - 当前处理步骤索引，用于恢复时定位
//   - ProcessedData - 已处理的数据缓存，避免重复计算
//   - DocumentID - 文档唯一标识，用于追踪和日志
//   - ProcessingSteps - 完整的处理步骤列表
//   - StartTime - 处理开始时间，用于性能统计
type DocumentProcessingState struct {
	CurrentStep     int                    `json:"current_step"`     // 当前处理步骤索引（从0开始）
	ProcessedData   map[string]interface{} `json:"processed_data"`   // 已处理数据的缓存映射
	DocumentID      string                 `json:"document_id"`      // 文档唯一标识符
	ProcessingSteps []string               `json:"processing_steps"` // 处理步骤名称列表
	StartTime       time.Time              `json:"start_time"`       // 处理任务开始时间
}

// DocumentProcessingAgent 可恢复的文档处理 Agent
// 实现 ResumableAgent 接口，支持复杂文档处理任务的中断和恢复
// 核心特性：
//   - 多步骤处理流程：文档验证、内容提取、敏感信息检测等
//   - 状态持久化：支持在任意步骤中断并从检查点恢复
//   - 进度跟踪：实时记录处理进度和中间结果
//   - 智能中断：在关键决策点自动触发中断等待人工干预
//
// 应用场景：
//   - 大文档批量处理：支持长时间运行任务的可靠执行
//   - 敏感信息审核：在检测到敏感内容时暂停等待人工确认
//   - 多阶段工作流：支持复杂业务流程的断点续传
//   - 系统维护恢复：服务重启后能够从中断点继续执行
type DocumentProcessingAgent struct {
	name            string                   // Agent 名称标识
	checkpointStore CheckPointStore          // 检查点存储接口，用于状态持久化
	state           *DocumentProcessingState // 当前处理状态，包含所有必要的状态信息
}

// NewDocumentProcessingAgent 创建新的文档处理 Agent 实例
// 参数：
//
//	store - 检查点存储实现，用于状态的持久化和恢复
//
// 返回：
//
//	*DocumentProcessingAgent - 已初始化的文档处理 Agent 实例
//
// 初始化内容：
//   - 设置 Agent 名称和检查点存储
//   - 创建初始处理状态，包含完整的处理步骤定义
//   - 生成唯一文档ID和记录开始时间
//   - 初始化数据缓存映射
func NewDocumentProcessingAgent(store CheckPointStore) *DocumentProcessingAgent {
	return &DocumentProcessingAgent{
		name:            "DocumentProcessingAgent", // 设置 Agent 名称
		checkpointStore: store,                     // 设置检查点存储
		state: &DocumentProcessingState{ // 初始化处理状态
			CurrentStep:   0,                                            // 从第一步开始
			ProcessedData: make(map[string]interface{}),                 // 初始化数据缓存
			DocumentID:    fmt.Sprintf("doc_%d", time.Now().UnixNano()), // 生成唯一文档ID
			ProcessingSteps: []string{ // 定义完整的处理步骤流程
				"文档验证",   // 步骤1：验证文档格式和完整性
				"内容提取",   // 步骤2：提取文档中的文本和结构化数据
				"敏感信息检测", // 步骤3：检测敏感信息，可能触发中断
				"格式转换",   // 步骤4：转换文档格式
				"质量检查",   // 步骤5：检查处理质量
				"最终输出",   // 步骤6：生成最终输出结果
			},
			StartTime: time.Now(), // 记录开始时间
		},
	}
}

// Name 返回 Agent 名称
// 实现 Agent 接口的 Name 方法
// 参数：
//
//	ctx - 上下文（此方法中未使用）
//
// 返回：
//
//	string - Agent 的名称标识
func (d *DocumentProcessingAgent) Name(ctx context.Context) string {
	return d.name // 返回 Agent 名称
}

// Description 返回 Agent 描述信息
// 实现 Agent 接口的 Description 方法
// 参数：
//
//	ctx - 上下文（此方法中未使用）
//
// 返回：
//
//	string - Agent 的功能描述
func (d *DocumentProcessingAgent) Description(ctx context.Context) string {
	return "可恢复的文档处理Agent，支持中断和恢复机制，确保长时间处理任务的可靠性" // 返回功能描述
}

// Run 启动 Agent 执行
// 实现 Agent 接口的 Run 方法，开始新的文档处理任务
// 参数：
//
//	ctx - 执行上下文，用于取消和超时控制
//	input - Agent 输入参数，包含要处理的文档信息
//	opts - 可选参数（此实现中未使用）
//
// 返回：
//
//	*AsyncIterator[*AgentEvent] - 异步事件迭代器，用于接收处理过程中的事件
//
// 执行流程：
//  1. 调用 executeProcessing 方法执行处理逻辑
//  2. 传入 false 表示非恢复模式
//  3. 传入 nil 表示没有中断信息
func (d *DocumentProcessingAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	return d.executeProcessing(ctx, input, false, nil) // 执行处理（非恢复模式）
}

// Resume 从中断点恢复 Agent 执行
// 实现 ResumableAgent 接口的 Resume 方法，从之前的中断点继续执行
// 参数：
//
//	ctx - 执行上下文，用于取消和超时控制
//	interruptInfo - 中断信息，包含中断时保存的状态数据
//	input - Agent 输入参数，可能包含用户的新决策
//	opts - 可选参数（此实现中未使用）
//
// 返回：
//
//	*AsyncIterator[*AgentEvent] - 异步事件迭代器，用于接收恢复过程中的事件
//
// 执行流程：
//  1. 调用 executeProcessing 方法执行处理逻辑
//  2. 传入 true 表示恢复模式
//  3. 传入 interruptInfo 包含中断时的状态信息
func (d *DocumentProcessingAgent) Resume(ctx context.Context, interruptInfo *InterruptInfo, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	return d.executeProcessing(ctx, input, true, interruptInfo) // 执行处理（恢复模式）
}

// executeProcessing 执行处理逻辑
// 核心处理方法，支持正常执行和从中断点恢复两种模式
// 参数：
//
//	ctx - 执行上下文，用于取消和超时控制
//	input - Agent 输入参数，包含处理所需的数据
//	isResume - 是否为恢复模式
//	interruptInfo - 中断信息，仅在恢复模式下使用
//
// 返回：
//
//	*AsyncIterator[*AgentEvent] - 异步事件迭代器
//
// 处理流程：
//  1. 根据模式初始化状态（新建或恢复）
//  2. 逐步执行处理步骤
//  3. 检测中断条件并处理
//  4. 发送进度和完成事件
func (d *DocumentProcessingAgent) executeProcessing(ctx context.Context, input *AgentInput, isResume bool, interruptInfo *InterruptInfo) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]() // 创建异步事件迭代器

	go func() {
		defer iter.Close() // 确保迭代器关闭

		if isResume && interruptInfo != nil {
			// 恢复模式：从中断信息中恢复状态
			if stateData, ok := interruptInfo.Data.(*DocumentProcessingState); ok {
				d.state = stateData // 恢复完整状态
				iter.Send(&AgentEvent{
					AgentName: d.name,
					RunPath:   []string{d.name},
					Output:    fmt.Sprintf("📥 从步骤 %d 恢复文档处理: %s", d.state.CurrentStep+1, d.state.ProcessingSteps[d.state.CurrentStep]),
					Timestamp: time.Now(),
				})

				iter.Send(&AgentEvent{
					AgentName: d.name,
					RunPath:   []string{d.name},
					Output:    "✅ 用户确认继续处理，跳过敏感信息检测步骤",
					Timestamp: time.Now(),
				})

				// 恢复时跳过当前中断的步骤，继续下一步
				d.state.CurrentStep++
			}
		} else {
			// 正常模式：开始新的处理任务
			iter.Send(&AgentEvent{
				AgentName: d.name,
				RunPath:   []string{d.name},
				Output:    fmt.Sprintf("📄 开始文档处理任务: %s", d.state.DocumentID),
				Timestamp: time.Now(),
			})
		}

		// 处理各个步骤循环
		// 从当前步骤开始，逐个执行剩余的处理步骤
		for i := d.state.CurrentStep; i < len(d.state.ProcessingSteps); i++ {
			step := d.state.ProcessingSteps[i] // 获取当前步骤名称
			d.state.CurrentStep = i            // 更新状态中的当前步骤

			// 发送步骤开始事件
			iter.Send(&AgentEvent{
				AgentName: d.name,
				RunPath:   []string{d.name},
				Output:    fmt.Sprintf("⚙️  执行步骤 %d/%d: %s", i+1, len(d.state.ProcessingSteps), step),
				Timestamp: time.Now(),
			})

			// 执行具体的步骤处理逻辑
			result, shouldInterrupt := d.processStep(i, step)
			// 保存步骤处理结果到状态缓存
			d.state.ProcessedData[fmt.Sprintf("step_%d", i+1)] = result

			// 检查是否需要中断（例如在敏感信息检测步骤）
			if shouldInterrupt {
				// 发送中断警告事件
				iter.Send(&AgentEvent{
					AgentName: d.name,
					RunPath:   []string{d.name},
					Output:    fmt.Sprintf("⚠️  在步骤 %d 检测到需要人工干预: %v", i+1, result),
					Timestamp: time.Now(),
				})

				// 发送中断事件，包含完整的状态信息
				iter.Send(&AgentEvent{
					AgentName: d.name,
					RunPath:   []string{d.name},
					Action: &AgentAction{
						Type: "interrupt", // 中断动作类型
						Data: d.state,     // 保存当前完整状态
					},
					Metadata: map[string]interface{}{
						"interrupt_reason": "sensitive_content_detected", // 中断原因
						"step":             i + 1,                        // 中断步骤
						"requires_human":   true,                         // 需要人工干预
					},
					Timestamp: time.Now(),
				})
				return // 中断执行，等待恢复
			}

			// 步骤成功完成，计算并发送进度事件
			progress := float64(i+1) / float64(len(d.state.ProcessingSteps)) * 100
			iter.Send(&AgentEvent{
				AgentName: d.name,
				RunPath:   []string{d.name},
				Output:    fmt.Sprintf("✅ 步骤 %d 完成，进度: %.1f%% - 结果: %v", i+1, progress, result),
				Timestamp: time.Now(),
			})

			// 模拟处理时间（实际应用中为真实处理时间）
			time.Sleep(500 * time.Millisecond)
		}

		// 所有步骤完成，计算总耗时
		duration := time.Since(d.state.StartTime)
		// 发送处理完成事件，包含完整的结果信息
		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Output: &Message{
				Role:    "assistant",
				Content: fmt.Sprintf("🎉 文档处理完成！文档ID: %s，耗时: %v，处理了 %d 个步骤", d.state.DocumentID, duration, len(d.state.ProcessingSteps)),
				Metadata: map[string]interface{}{
					"document_id":     d.state.DocumentID,           // 文档标识
					"duration":        duration.String(),            // 处理耗时
					"steps_completed": len(d.state.ProcessingSteps), // 完成步骤数
					"processed_data":  d.state.ProcessedData,        // 处理结果数据
				},
			},
			Timestamp: time.Now(),
		})

		// 发送退出事件，标记任务完成
		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Action:    &AgentAction{Type: "exit", Data: "文档处理任务完成"},
			Timestamp: time.Now(),
		})
	}()

	return iter // 返回异步事件迭代器
}

// processStep 处理单个步骤
// 执行具体的文档处理步骤，每个步骤都有特定的处理逻辑
// 参数：
//
//	stepIndex - 步骤索引（从0开始）
//	stepName - 步骤名称（用于日志和调试）
//
// 返回：
//
//	interface{} - 步骤处理结果，包含该步骤产生的数据
//	bool - 是否需要中断执行（true表示需要人工干预）
//
// 步骤说明：
//  0. 文档验证 - 检查文档格式、大小、完整性
//  1. 内容提取 - 提取文本、图片、表格等内容
//  2. 敏感信息检测 - 检测个人信息、机密数据（可能触发中断）
//  3. 格式转换 - 将文档转换为目标格式
//  4. 质量检查 - 评估处理质量和完整性
//  5. 最终输出 - 生成最终结果文件
func (d *DocumentProcessingAgent) processStep(stepIndex int, stepName string) (interface{}, bool) {
	switch stepIndex {
	case 0: // 文档验证步骤
		// 验证文档的基本属性：格式、大小、页数等
		return map[string]interface{}{
			"valid":  true,    // 文档是否有效
			"format": "PDF",   // 文档格式
			"size":   "2.1MB", // 文件大小
		}, false // 不需要中断

	case 1: // 内容提取步骤
		// 从文档中提取文本内容、图片等信息
		return map[string]interface{}{
			"pages":      15,    // 文档页数
			"text_chars": 45000, // 文本字符数
			"images":     3,     // 图片数量
		}, false // 不需要中断

	case 2: // 敏感信息检测步骤（关键步骤）
		// 检测文档中的敏感信息，如身份证号、银行卡号等
		// 这是一个关键步骤，检测到敏感信息时需要人工确认
		return map[string]interface{}{
			"sensitive_items": []string{"身份证号", "银行卡号"}, // 检测到的敏感信息类型
			"risk_level":      "high",                   // 风险等级
			"requires_review": true,                     // 是否需要人工审核
		}, true // 返回true表示需要中断，等待人工决策

	case 3: // 格式转换步骤
		// 将文档转换为目标格式（如HTML、JSON等）
		return map[string]interface{}{
			"output_format": "HTML",  // 输出格式
			"file_size":     "1.8MB", // 转换后文件大小
		}, false // 不需要中断

	case 4: // 质量检查步骤
		// 评估处理质量，检查是否有遗漏或错误
		return map[string]interface{}{
			"quality_score": 0.92, // 质量评分（0-1）
			"issues_found":  1,    // 发现的问题数量
		}, false // 不需要中断

	case 5: // 最终输出步骤
		// 生成最终的处理结果文件
		return map[string]interface{}{
			"output_path": "/output/processed_document.html", // 输出文件路径
			"status":      "completed",                       // 处理状态
		}, false // 不需要中断

	default:
		// 未知步骤，返回错误信息
		return "未知步骤", false
	}
}

// ============= 演示程序 =============

// demonstrateInterruptResume 演示中断和恢复功能
// 这是一个完整的演示函数，展示可恢复 Agent 的核心特性
// 演示流程：
//  1. 创建检查点存储和 Runner
//  2. 启动文档处理 Agent
//  3. 执行直到遇到中断（敏感信息检测步骤）
//  4. 模拟用户决策过程
//  5. 从中断点恢复执行
//  6. 完成剩余处理步骤
//
// 核心价值：
//   - 展示状态持久化机制
//   - 演示中断检测和处理
//   - 验证恢复机制的可靠性
//   - 模拟真实的业务场景
func demonstrateInterruptResume() {
	fmt.Println("🎯 ResumableAgent 中断恢复演示")
	fmt.Println(strings.Repeat("=", 70))

	ctx := context.Background()

	// 创建检查点存储
	// 使用文件系统作为持久化后端，实际生产环境可使用数据库
	store := NewFileCheckPointStore("./demo_data")

	// 创建 Runner 配置和实例
	// Runner 负责 Agent 的生命周期管理和中断恢复
	runner := NewRunner(&RunnerConfig{
		CheckPointStore: store, // 设置检查点存储
		EnableCallback:  true,  // 启用事件回调
		MaxRetries:      3,     // 最大重试次数
	})

	// 创建文档处理 Agent 实例
	// 该 Agent 实现了 ResumableAgent 接口，支持中断恢复
	agent := NewDocumentProcessingAgent(store)

	fmt.Printf("🤖 可恢复Agent: %s\n", agent.Name(ctx))
	fmt.Printf("📝 描述: %s\n", agent.Description(ctx))
	fmt.Println()

	// ========== 第一阶段：正常执行直到中断 ==========
	fmt.Println("▶️  第一阶段：执行任务直到中断")
	fmt.Println(strings.Repeat("-", 50))

	// 创建 Agent 输入参数
	// 包含要处理的文档信息和处理配置
	input := &AgentInput{
		Messages: []*Message{
			{
				Role:    "user",
				Content: "请处理这份包含敏感信息的文档",
			},
		},
		SessionID: "interrupt_demo_session", // 会话标识，用于状态管理
	}

	// 启动 Agent 执行，返回异步事件迭代器
	iter := runner.Execute(ctx, agent, input)
	var sessionInterrupted bool // 标记是否发生中断

	// 处理执行过程中的事件流
	// 监听 Agent 执行过程中产生的各种事件
	for {
		event, ok := iter.Next(ctx) // 获取下一个事件
		if !ok {
			break // 没有更多事件，执行结束
		}

		if event != nil {
			// 构建执行路径字符串，用于显示事件来源
			runPathStr := strings.Join(event.RunPath, " → ")
			fmt.Printf("📡 [%s] %s: ", event.Timestamp.Format("15:04:05"), runPathStr)

			// 处理事件输出内容
			if event.Output != nil {
				if msg, ok := event.Output.(*Message); ok {
					// 结构化消息输出
					fmt.Printf("💬 %s\n", msg.Content)
				} else {
					// 普通信息输出
					fmt.Printf("ℹ️  %v\n", event.Output)
				}
			}

			// 处理 Agent 动作事件
			if event.Action != nil {
				fmt.Printf("🎬 动作: %s", event.Action.Type)
				if event.Action.Type == "interrupt" {
					// 检测到中断事件，这是关键的业务逻辑点
					fmt.Printf(" (中断原因: 敏感信息检测)")
					sessionInterrupted = true // 标记发生中断
				}
				if event.Action.Data != nil && event.Action.Type != "interrupt" {
					// 显示非中断动作的附加数据
					fmt.Printf(" (%v)", event.Action.Data)
				}
				fmt.Println()
			}

			// 处理错误事件
			if event.Error != nil {
				fmt.Printf("❌ 错误: %v\n", event.Error)
			}
		}
	}

	// 检查是否发生了中断，如果是则进入恢复流程
	if sessionInterrupted {
		// ========== 模拟用户决策过程 ==========
		fmt.Println("\n⏸️  任务已中断，模拟用户决策时间...")
		time.Sleep(2 * time.Second) // 模拟用户思考和决策的时间

		// ========== 第二阶段：恢复执行 ==========
		fmt.Println("\n▶️  第二阶段：恢复任务执行")
		fmt.Println(strings.Repeat("-", 50))

		// 创建恢复输入参数
		// 包含用户的决策信息，表明已确认继续处理
		resumeInput := &AgentInput{
			Messages: []*Message{
				{
					Role:    "user",
					Content: "已确认处理敏感信息，继续执行后续步骤",
				},
			},
			SessionID: "interrupt_demo_session", // 使用相同的会话ID
		}

		// 从检查点恢复 Agent 执行
		// Runner 会自动加载之前保存的状态并继续执行
		resumeIter := runner.Resume(ctx, "interrupt_demo_session", resumeInput)
		// 处理恢复过程中的事件流
		// 监听从中断点恢复后的执行过程
		for {
			event, ok := resumeIter.Next(ctx) // 获取恢复过程中的事件
			if !ok {
				break // 恢复执行完成
			}

			if event != nil {
				// 构建执行路径字符串
				runPathStr := strings.Join(event.RunPath, " → ")
				fmt.Printf("📡 [%s] %s: ", event.Timestamp.Format("15:04:05"), runPathStr)

				// 处理恢复过程中的输出事件
				if event.Output != nil {
					if msg, ok := event.Output.(*Message); ok {
						// 结构化消息输出（如最终完成消息）
						fmt.Printf("💬 %s\n", msg.Content)
					} else {
						// 普通信息输出（如步骤进度）
						fmt.Printf("ℹ️  %v\n", event.Output)
					}
				}

				// 处理恢复过程中的动作事件
				if event.Action != nil {
					fmt.Printf("🎬 动作: %s", event.Action.Type)
					if event.Action.Data != nil {
						// 显示动作附加数据
						fmt.Printf(" (%v)", event.Action.Data)
					}
					fmt.Println()
				}

				// 处理恢复过程中的错误
				if event.Error != nil {
					fmt.Printf("❌ 错误: %v\n", event.Error)
				}
			}
		}
	}
}

// main 主函数 - 演示可恢复Agent的完整功能
// 这是整个演示程序的入口点，展示了企业级AI Agent的核心能力
// 主要功能：
//  1. 启动可恢复Agent演示
//  2. 展示中断恢复机制
//  3. 总结核心特性和企业价值
//
// 演示场景：
//   - 文档处理过程中的敏感信息检测中断
//   - 用户决策后的任务恢复
//   - 完整的状态持久化和恢复流程
//
// 企业价值：
//   - 保障长时间运行任务的可靠性
//   - 支持人工干预的业务流程
//   - 提升系统的容错能力和用户体验
func main() {
	// 程序启动信息 - 展示演示程序的主题和目标
	fmt.Println("🎊 Eino ADK ResumableAgent 和 Runner 机制演示")
	fmt.Println("基于官方Agent扩展机制的完整实现")
	fmt.Println(strings.Repeat("=", 80))

	// 运行核心演示功能
	// 这个演示展示了完整的中断-恢复生命周期
	demonstrateInterruptResume()

	// ========== 功能特性总结 ==========
	// 总结演示的核心技术特性和实现机制
	fmt.Println("\n🎯 ResumableAgent 核心特性总结")
	fmt.Println(strings.Repeat("=", 80))

	// 展示成功演示的扩展机制特性
	fmt.Println("✨ 成功演示的扩展机制特性:")

	// ResumableAgent 接口特性
	fmt.Println("  🔄 ResumableAgent 接口")
	fmt.Println("    - Resume() 方法支持从中断点恢复")
	fmt.Println("    - InterruptInfo 携带中断时的状态数据")
	fmt.Println("    - 支持任意复杂状态的序列化保存")

	// Runner 生命周期管理特性
	fmt.Println("  🎯 Runner 生命周期管理")
	fmt.Println("    - Execute() 管理Agent完整执行过程")
	fmt.Println("    - 自动处理中断事件和状态保存")
	fmt.Println("    - Resume() 支持会话级恢复")

	// CheckPointStore 机制特性
	fmt.Println("  💾 CheckPointStore 机制")
	fmt.Println("    - 灵活的检查点存储抽象")
	fmt.Println("    - Gob序列化支持复杂数据结构")
	fmt.Println("    - 支持跨进程状态持久化")

	// 中断驱动架构特性
	fmt.Println("  📡 中断驱动架构")
	fmt.Println("    - AgentAction.Type='interrupt' 标准化中断")
	fmt.Println("    - 事件流统一处理中断和恢复")
	fmt.Println("    - 支持复杂的多步骤任务管理")

	// ========== 企业级应用价值 ==========
	// 总结在企业环境中的实际应用价值和业务意义
	fmt.Println("\n💡 企业级应用价值:")
	// 长时间任务的可靠性保障
	fmt.Println("  • 长时间运行任务的可靠性保障")
	// 人机协作的中断决策支持
	fmt.Println("  • 支持人机协作的中断决策")
	// 跨系统重启的状态持久化能力
	fmt.Println("  • 跨系统重启的状态持久化")
	// 复杂业务流程的断点续传功能
	fmt.Println("  • 复杂业务流程的断点续传")

	// 演示结束信息
	fmt.Println("\n🎉 这就是 Eino ADK 的企业级Agent扩展机制！")
}
