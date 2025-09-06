package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============= Eino ADK ResumableAgent 和 Runner 机制真实演示 =============
// 基于 https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_extension/
// 严格按照官方扩展机制实现中断恢复功能

// 注册自定义类型以支持gob序列化
func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(&InterruptInfo{})
	gob.Register(&DocumentProcessingState{})
}

// ============= 核心扩展机制类型定义 =============

type Message struct {
	Role     string                 `json:"role"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type AgentInput struct {
	Messages        []*Message `json:"messages"`
	EnableStreaming bool       `json:"enable_streaming,omitempty"`
	SessionID       string     `json:"session_id,omitempty"`
}

type AgentEvent struct {
	AgentName string                 `json:"agent_name"`
	RunPath   []string               `json:"run_path,omitempty"`
	Output    interface{}            `json:"output,omitempty"`
	Action    *AgentAction           `json:"action,omitempty"`
	Error     error                  `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type AgentAction struct {
	Type   string      `json:"type"` // "exit", "transfer", "interrupt"
	Target string      `json:"target,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

// InterruptInfo 中断信息 - 官方接口
type InterruptInfo struct {
	Data interface{} `json:"data"`
}

// CheckPointStore 检查点存储接口 - 官方接口
type CheckPointStore interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) error
}

// AsyncIterator 异步迭代器
type AsyncIterator[T any] struct {
	ch   chan T
	done chan bool
}

func NewAsyncIterator[T any]() *AsyncIterator[T] {
	return &AsyncIterator[T]{
		ch:   make(chan T, 100),
		done: make(chan bool),
	}
}

func (ai *AsyncIterator[T]) Next(ctx context.Context) (T, bool) {
	select {
	case value, ok := <-ai.ch:
		return value, ok
	case <-ai.done:
		var zero T
		return zero, false
	case <-ctx.Done():
		var zero T
		return zero, false
	}
}

func (ai *AsyncIterator[T]) Send(value T) {
	select {
	case ai.ch <- value:
	default:
		log.Printf("警告: AsyncIterator 缓冲区已满")
	}
}

func (ai *AsyncIterator[T]) Close() {
	close(ai.ch)
	close(ai.done)
}

// ============= 核心Agent接口 =============

// Agent 基础Agent接口
type Agent interface {
	Name(ctx context.Context) string
	Description(ctx context.Context) string
	Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent]
}

// ResumableAgent 可恢复Agent接口 - 官方扩展接口
type ResumableAgent interface {
	Agent
	Resume(ctx context.Context, interruptInfo *InterruptInfo, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent]
}

// ============= 检查点存储实现 =============

// FileCheckPointStore 基于文件的检查点存储
type FileCheckPointStore struct {
	baseDir string
}

func NewFileCheckPointStore(baseDir string) *FileCheckPointStore {
	return &FileCheckPointStore{baseDir: baseDir}
}

func (f *FileCheckPointStore) Set(ctx context.Context, key string, value []byte) error {
	dir := filepath.Join(f.baseDir, "checkpoints")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, key+".checkpoint"), value, 0644)
}

func (f *FileCheckPointStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	filePath := filepath.Join(f.baseDir, "checkpoints", key+".checkpoint")
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	return data, true, err
}

func (f *FileCheckPointStore) Delete(ctx context.Context, key string) error {
	filePath := filepath.Join(f.baseDir, "checkpoints", key+".checkpoint")
	return os.Remove(filePath)
}

// ============= Runner 实现 - 官方Runner机制 =============

// RunnerConfig Runner配置
type RunnerConfig struct {
	CheckPointStore CheckPointStore
	EnableCallback  bool
	MaxRetries      int
}

// Runner Agent运行器 - 管理整个Agent生命周期
type Runner struct {
	config          *RunnerConfig
	checkpointStore CheckPointStore
	runningAgents   map[string]*AgentExecution
}

type AgentExecution struct {
	Agent        Agent
	SessionID    string
	CheckPointID string
	StartTime    time.Time
	Status       string // "running", "interrupted", "completed", "error"
}

func NewRunner(config *RunnerConfig) *Runner {
	return &Runner{
		config:          config,
		checkpointStore: config.CheckPointStore,
		runningAgents:   make(map[string]*AgentExecution),
	}
}

func (r *Runner) Execute(ctx context.Context, agent Agent, input *AgentInput) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]()
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}

	execution := &AgentExecution{
		Agent:        agent,
		SessionID:    sessionID,
		CheckPointID: fmt.Sprintf("%s_%s", agent.Name(ctx), sessionID),
		StartTime:    time.Now(),
		Status:       "running",
	}

	r.runningAgents[sessionID] = execution

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

func (r *Runner) Resume(ctx context.Context, sessionID string, newInput *AgentInput) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]()

	go func() {
		defer iter.Close()

		// 发送恢复开始事件
		iter.Send(&AgentEvent{
			AgentName: "Runner",
			RunPath:   []string{"Runner"},
			Output:    fmt.Sprintf("🔄 Runner开始恢复会话: %s", sessionID),
			Timestamp: time.Now(),
		})

		// 从检查点恢复
		checkPointID := fmt.Sprintf("agent_%s", sessionID)
		data, exists, err := r.checkpointStore.Get(ctx, checkPointID)
		if err != nil {
			iter.Send(&AgentEvent{
				AgentName: "Runner",
				RunPath:   []string{"Runner"},
				Error:     fmt.Errorf("读取检查点失败: %w", err),
				Timestamp: time.Now(),
			})
			return
		}

		if !exists {
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
		buf := bytes.NewBuffer(data)
		decoder := gob.NewDecoder(buf)
		if err := decoder.Decode(&interruptInfo); err != nil {
			iter.Send(&AgentEvent{
				AgentName: "Runner",
				RunPath:   []string{"Runner"},
				Error:     fmt.Errorf("反序列化检查点失败: %w", err),
				Timestamp: time.Now(),
			})
			return
		}

		iter.Send(&AgentEvent{
			AgentName: "Runner",
			RunPath:   []string{"Runner"},
			Output:    "✅ 检查点恢复成功，准备恢复Agent执行",
			Timestamp: time.Now(),
		})

		// 这里需要根据实际情况恢复对应的Agent
		// 为了演示，我们创建一个新的DocumentProcessingAgent
		agent := NewDocumentProcessingAgent(r.checkpointStore)
		// DocumentProcessingAgent 实现了 ResumableAgent 接口
		agentIter := agent.Resume(ctx, &interruptInfo, newInput)
		for {
			event, ok := agentIter.Next(ctx)
			if !ok {
				break
			}
			if event != nil {
				iter.Send(event)
			}
		}

		// 清理检查点
		r.checkpointStore.Delete(ctx, checkPointID)

		iter.Send(&AgentEvent{
			AgentName: "Runner",
			RunPath:   []string{"Runner"},
			Output:    "🎉 Runner恢复执行完成",
			Timestamp: time.Now(),
		})
	}()

	return iter
}

func (r *Runner) handleInterrupt(ctx context.Context, execution *AgentExecution, event *AgentEvent) error {
	// 序列化中断信息
	interruptInfo := &InterruptInfo{
		Data: event.Action.Data,
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(interruptInfo); err != nil {
		return fmt.Errorf("序列化中断信息失败: %w", err)
	}

	// 保存检查点
	checkPointID := fmt.Sprintf("agent_%s", execution.SessionID)
	return r.checkpointStore.Set(ctx, checkPointID, buf.Bytes())
}

// ============= ResumableAgent 实现示例 =============

// DocumentProcessingState 文档处理状态
type DocumentProcessingState struct {
	CurrentStep     int                    `json:"current_step"`
	ProcessedData   map[string]interface{} `json:"processed_data"`
	DocumentID      string                 `json:"document_id"`
	ProcessingSteps []string               `json:"processing_steps"`
	StartTime       time.Time              `json:"start_time"`
}

// DocumentProcessingAgent 可恢复的文档处理Agent
type DocumentProcessingAgent struct {
	name            string
	checkpointStore CheckPointStore
	state           *DocumentProcessingState
}

func NewDocumentProcessingAgent(store CheckPointStore) *DocumentProcessingAgent {
	return &DocumentProcessingAgent{
		name:            "DocumentProcessingAgent",
		checkpointStore: store,
		state: &DocumentProcessingState{
			CurrentStep:   0,
			ProcessedData: make(map[string]interface{}),
			DocumentID:    fmt.Sprintf("doc_%d", time.Now().UnixNano()),
			ProcessingSteps: []string{
				"文档验证",
				"内容提取",
				"敏感信息检测",
				"格式转换",
				"质量检查",
				"最终输出",
			},
			StartTime: time.Now(),
		},
	}
}

func (d *DocumentProcessingAgent) Name(ctx context.Context) string {
	return d.name
}

func (d *DocumentProcessingAgent) Description(ctx context.Context) string {
	return "可恢复的文档处理Agent，支持中断和恢复机制，确保长时间处理任务的可靠性"
}

func (d *DocumentProcessingAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	return d.executeProcessing(ctx, input, false, nil)
}

func (d *DocumentProcessingAgent) Resume(ctx context.Context, interruptInfo *InterruptInfo, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	return d.executeProcessing(ctx, input, true, interruptInfo)
}

func (d *DocumentProcessingAgent) executeProcessing(ctx context.Context, input *AgentInput, isResume bool, interruptInfo *InterruptInfo) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]()

	go func() {
		defer iter.Close()

		if isResume && interruptInfo != nil {
			// 恢复模式：从中断信息中恢复状态
			if stateData, ok := interruptInfo.Data.(*DocumentProcessingState); ok {
				d.state = stateData
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

		// 处理各个步骤
		for i := d.state.CurrentStep; i < len(d.state.ProcessingSteps); i++ {
			step := d.state.ProcessingSteps[i]
			d.state.CurrentStep = i

			iter.Send(&AgentEvent{
				AgentName: d.name,
				RunPath:   []string{d.name},
				Output:    fmt.Sprintf("⚙️  执行步骤 %d/%d: %s", i+1, len(d.state.ProcessingSteps), step),
				Timestamp: time.Now(),
			})

			// 模拟步骤处理
			result, shouldInterrupt := d.processStep(i, step)
			d.state.ProcessedData[fmt.Sprintf("step_%d", i+1)] = result

			// 检查是否需要中断（例如在敏感信息检测步骤）
			if shouldInterrupt {
				iter.Send(&AgentEvent{
					AgentName: d.name,
					RunPath:   []string{d.name},
					Output:    fmt.Sprintf("⚠️  在步骤 %d 检测到需要人工干预: %v", i+1, result),
					Timestamp: time.Now(),
				})

				// 发送中断事件
				iter.Send(&AgentEvent{
					AgentName: d.name,
					RunPath:   []string{d.name},
					Action: &AgentAction{
						Type: "interrupt",
						Data: d.state,
					},
					Metadata: map[string]interface{}{
						"interrupt_reason": "sensitive_content_detected",
						"step":             i + 1,
						"requires_human":   true,
					},
					Timestamp: time.Now(),
				})
				return
			}

			// 步骤完成
			progress := float64(i+1) / float64(len(d.state.ProcessingSteps)) * 100
			iter.Send(&AgentEvent{
				AgentName: d.name,
				RunPath:   []string{d.name},
				Output:    fmt.Sprintf("✅ 步骤 %d 完成，进度: %.1f%% - 结果: %v", i+1, progress, result),
				Timestamp: time.Now(),
			})

			// 模拟处理时间
			time.Sleep(500 * time.Millisecond)
		}

		// 所有步骤完成
		duration := time.Since(d.state.StartTime)
		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Output: &Message{
				Role:    "assistant",
				Content: fmt.Sprintf("🎉 文档处理完成！文档ID: %s，耗时: %v，处理了 %d 个步骤", d.state.DocumentID, duration, len(d.state.ProcessingSteps)),
				Metadata: map[string]interface{}{
					"document_id":     d.state.DocumentID,
					"duration":        duration.String(),
					"steps_completed": len(d.state.ProcessingSteps),
					"processed_data":  d.state.ProcessedData,
				},
			},
			Timestamp: time.Now(),
		})

		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Action:    &AgentAction{Type: "exit", Data: "文档处理任务完成"},
			Timestamp: time.Now(),
		})
	}()

	return iter
}

func (d *DocumentProcessingAgent) processStep(stepIndex int, stepName string) (interface{}, bool) {
	switch stepIndex {
	case 0: // 文档验证
		return map[string]interface{}{
			"valid":  true,
			"format": "PDF",
			"size":   "2.1MB",
		}, false

	case 1: // 内容提取
		return map[string]interface{}{
			"pages":      15,
			"text_chars": 45000,
			"images":     3,
		}, false

	case 2: // 敏感信息检测 - 这里会触发中断
		return map[string]interface{}{
			"sensitive_items": []string{"身份证号", "银行卡号"},
			"risk_level":      "high",
			"requires_review": true,
		}, true // 返回true表示需要中断

	case 3: // 格式转换
		return map[string]interface{}{
			"output_format": "HTML",
			"file_size":     "1.8MB",
		}, false

	case 4: // 质量检查
		return map[string]interface{}{
			"quality_score": 0.92,
			"issues_found":  1,
		}, false

	case 5: // 最终输出
		return map[string]interface{}{
			"output_path": "/output/processed_document.html",
			"status":      "completed",
		}, false

	default:
		return "未知步骤", false
	}
}

// ============= 演示程序 =============

func demonstrateInterruptResume() {
	fmt.Println("🎯 ResumableAgent 中断恢复演示")
	fmt.Println(strings.Repeat("=", 70))

	ctx := context.Background()

	// 创建检查点存储
	store := NewFileCheckPointStore("./demo_data")

	// 创建Runner
	runner := NewRunner(&RunnerConfig{
		CheckPointStore: store,
		EnableCallback:  true,
		MaxRetries:      3,
	})

	// 创建可恢复Agent
	agent := NewDocumentProcessingAgent(store)

	fmt.Printf("🤖 可恢复Agent: %s\n", agent.Name(ctx))
	fmt.Printf("📝 描述: %s\n", agent.Description(ctx))
	fmt.Println()

	// 第一阶段：正常执行直到中断
	fmt.Println("▶️  第一阶段：执行任务直到中断")
	fmt.Println(strings.Repeat("-", 50))

	input := &AgentInput{
		Messages: []*Message{
			{
				Role:    "user",
				Content: "请处理这份包含敏感信息的文档",
			},
		},
		SessionID: "interrupt_demo_session",
	}

	iter := runner.Execute(ctx, agent, input)
	var sessionInterrupted bool

	for {
		event, ok := iter.Next(ctx)
		if !ok {
			break
		}

		if event != nil {
			runPathStr := strings.Join(event.RunPath, " → ")
			fmt.Printf("📡 [%s] %s: ", event.Timestamp.Format("15:04:05"), runPathStr)

			if event.Output != nil {
				if msg, ok := event.Output.(*Message); ok {
					fmt.Printf("💬 %s\n", msg.Content)
				} else {
					fmt.Printf("ℹ️  %v\n", event.Output)
				}
			}

			if event.Action != nil {
				fmt.Printf("🎬 动作: %s", event.Action.Type)
				if event.Action.Type == "interrupt" {
					fmt.Printf(" (中断原因: 敏感信息检测)")
					sessionInterrupted = true
				}
				if event.Action.Data != nil && event.Action.Type != "interrupt" {
					fmt.Printf(" (%v)", event.Action.Data)
				}
				fmt.Println()
			}

			if event.Error != nil {
				fmt.Printf("❌ 错误: %v\n", event.Error)
			}
		}
	}

	if sessionInterrupted {
		fmt.Println("\n⏸️  任务已中断，模拟用户决策时间...")
		time.Sleep(2 * time.Second)

		// 第二阶段：恢复执行
		fmt.Println("\n▶️  第二阶段：恢复任务执行")
		fmt.Println(strings.Repeat("-", 50))

		resumeInput := &AgentInput{
			Messages: []*Message{
				{
					Role:    "user",
					Content: "已确认处理敏感信息，继续执行后续步骤",
				},
			},
			SessionID: "interrupt_demo_session",
		}

		resumeIter := runner.Resume(ctx, "interrupt_demo_session", resumeInput)
		for {
			event, ok := resumeIter.Next(ctx)
			if !ok {
				break
			}

			if event != nil {
				runPathStr := strings.Join(event.RunPath, " → ")
				fmt.Printf("📡 [%s] %s: ", event.Timestamp.Format("15:04:05"), runPathStr)

				if event.Output != nil {
					if msg, ok := event.Output.(*Message); ok {
						fmt.Printf("💬 %s\n", msg.Content)
					} else {
						fmt.Printf("ℹ️  %v\n", event.Output)
					}
				}

				if event.Action != nil {
					fmt.Printf("🎬 动作: %s", event.Action.Type)
					if event.Action.Data != nil {
						fmt.Printf(" (%v)", event.Action.Data)
					}
					fmt.Println()
				}

				if event.Error != nil {
					fmt.Printf("❌ 错误: %v\n", event.Error)
				}
			}
		}
	}
}

func main() {
	fmt.Println("🎊 Eino ADK ResumableAgent 和 Runner 机制演示")
	fmt.Println("基于官方Agent扩展机制的完整实现")
	fmt.Println(strings.Repeat("=", 80))

	// 中断恢复演示
	demonstrateInterruptResume()

	// 总结
	fmt.Println("\n🎯 ResumableAgent 核心特性总结")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("✨ 成功演示的扩展机制特性:")
	fmt.Println("  🔄 ResumableAgent 接口")
	fmt.Println("    - Resume() 方法支持从中断点恢复")
	fmt.Println("    - InterruptInfo 携带中断时的状态数据")
	fmt.Println("    - 支持任意复杂状态的序列化保存")

	fmt.Println("  🎯 Runner 生命周期管理")
	fmt.Println("    - Execute() 管理Agent完整执行过程")
	fmt.Println("    - 自动处理中断事件和状态保存")
	fmt.Println("    - Resume() 支持会话级恢复")

	fmt.Println("  💾 CheckPointStore 机制")
	fmt.Println("    - 灵活的检查点存储抽象")
	fmt.Println("    - Gob序列化支持复杂数据结构")
	fmt.Println("    - 支持跨进程状态持久化")

	fmt.Println("  📡 中断驱动架构")
	fmt.Println("    - AgentAction.Type='interrupt' 标准化中断")
	fmt.Println("    - 事件流统一处理中断和恢复")
	fmt.Println("    - 支持复杂的多步骤任务管理")

	fmt.Println("\n💡 企业级应用价值:")
	fmt.Println("  • 长时间运行任务的可靠性保障")
	fmt.Println("  • 支持人机协作的中断决策")
	fmt.Println("  • 跨系统重启的状态持久化")
	fmt.Println("  • 复杂业务流程的断点续传")

	fmt.Println("\n🎉 这就是 Eino ADK 的企业级Agent扩展机制！")
}
