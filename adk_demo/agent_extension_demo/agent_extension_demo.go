package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// ============= Agent 扩展机制演示 =============
// 基于 https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_extension/
// 演示中断、恢复、检查点等高级功能

// ============= 核心接口定义 =============

// 核心接口和结构定义
type InterruptInfo struct {
	Data interface{} `json:"data"`
}

type AgentEvent struct {
	Type      string                 `json:"type"`
	AgentName string                 `json:"agent_name,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type CheckPointStore interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) error
}

// ResumeInfo 恢复信息
type ResumeInfo struct {
	CheckPointData []byte                 `json:"checkpoint_data"`
	NewInput       *AgentInput            `json:"new_input"`
	ResumeReason   string                 `json:"resume_reason"`
	ResumeContext  map[string]interface{} `json:"resume_context"`
}

// AgentInput Agent 输入
type AgentInput struct {
	Messages      []*schema.Message      `json:"messages"`
	SessionValues map[string]interface{} `json:"session_values,omitempty"`
	History       []*AgentEvent          `json:"history,omitempty"`
}

// AsyncIterator 异步迭代器
type AsyncIterator[T any] struct {
	events chan T
	done   chan bool
}

func NewAsyncIterator[T any]() *AsyncIterator[T] {
	return &AsyncIterator[T]{
		events: make(chan T, 50),
		done:   make(chan bool, 1),
	}
}

func (it *AsyncIterator[T]) Send(event T) {
	select {
	case it.events <- event:
	default:
		// 缓冲区满时忽略
	}
}

func (it *AsyncIterator[T]) Next() (T, bool) {
	select {
	case event := <-it.events:
		return event, true
	case <-it.done:
		var zero T
		return zero, false
	default:
		var zero T
		return zero, false
	}
}

func (it *AsyncIterator[T]) Close() {
	select {
	case it.done <- true:
	default:
	}
	close(it.events)
	close(it.done)
}

// ============= 检查点存储接口 =============

// CheckPointStore 检查点存储接口 (已在上方定义)

// FileCheckPointStore 基于文件系统的检查点存储
type FileCheckPointStore struct {
	baseDir string
}

func NewFileCheckPointStore(baseDir string) *FileCheckPointStore {
	return &FileCheckPointStore{baseDir: baseDir}
}

func (f *FileCheckPointStore) Set(ctx context.Context, key string, value []byte) error {
	checkpointDir := filepath.Join(f.baseDir, "checkpoints")
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		return fmt.Errorf("创建检查点目录失败: %w", err)
	}

	filePath := filepath.Join(checkpointDir, key+".gob")
	return os.WriteFile(filePath, value, 0644)
}

func (f *FileCheckPointStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	filePath := filepath.Join(f.baseDir, "checkpoints", key+".gob")

	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (f *FileCheckPointStore) Delete(ctx context.Context, key string) error {
	filePath := filepath.Join(f.baseDir, "checkpoints", key+".gob")
	err := os.Remove(filePath)
	if os.IsNotExist(err) {
		return nil // 文件不存在也算删除成功
	}
	return err
}

// ============= 基础 Agent 接口 =============

// Agent 基础接口
type Agent interface {
	Name(ctx context.Context) string
	Description(ctx context.Context) string
	Run(ctx context.Context, input *AgentInput) *AsyncIterator[*AgentEvent]
}

// ResumableAgent 可恢复 Agent 接口
type ResumableAgent interface {
	Agent
	Resume(ctx context.Context, info *ResumeInfo) *AsyncIterator[*AgentEvent]
	SaveCheckpoint(ctx context.Context) ([]byte, error)
	LoadCheckpoint(ctx context.Context, data []byte) error
}

// ============= 文档处理 Agent 实现 =============

// 注册自定义类型以支持 gob 序列化
func init() {
	gob.RegisterName("DocumentProcessorState", DocumentProcessorState{})
	gob.RegisterName("TaskItem", TaskItem{})
	gob.Register(map[string]interface{}{})
}

// DocumentProcessorState 文档处理器状态
type DocumentProcessorState struct {
	CurrentStep     int                    `json:"current_step"`
	ProcessedData   map[string]interface{} `json:"processed_data"`
	DocumentPath    string                 `json:"document_path"`
	OutputFormat    string                 `json:"output_format"`
	StartTime       time.Time              `json:"start_time"`
	PauseReason     string                 `json:"pause_reason"`
	UserDecisions   []string               `json:"user_decisions"`
	ProcessingSteps []string               `json:"processing_steps"`
}

// TaskItem 任务项
type TaskItem struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	StartTime time.Time              `json:"start_time"`
	Duration  time.Duration          `json:"duration"`
	Result    map[string]interface{} `json:"result"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DocumentProcessorAgent 文档处理 Agent
type DocumentProcessorAgent struct {
	state           DocumentProcessorState
	checkpointStore CheckPointStore
	agentID         string
}

func NewDocumentProcessorAgent(checkpointStore CheckPointStore, agentID string) *DocumentProcessorAgent {
	return &DocumentProcessorAgent{
		state: DocumentProcessorState{
			CurrentStep:     1,
			ProcessedData:   make(map[string]interface{}),
			StartTime:       time.Now(),
			UserDecisions:   []string{},
			ProcessingSteps: []string{"文档验证", "OCR识别", "内容分析", "信息提取", "格式转换"},
		},
		checkpointStore: checkpointStore,
		agentID:         agentID,
	}
}

func (d *DocumentProcessorAgent) Name(ctx context.Context) string {
	return "DocumentProcessor"
}

func (d *DocumentProcessorAgent) Description(ctx context.Context) string {
	return "智能文档处理器，支持多格式文档的解析、分析和转换，具备中断恢复功能"
}

func (d *DocumentProcessorAgent) Run(ctx context.Context, input *AgentInput) *AsyncIterator[*AgentEvent] {
	stream := NewAsyncIterator[*AgentEvent]()

	go func() {
		defer stream.Close()

		// 从输入中获取文档路径和输出格式
		if input != nil && input.SessionValues != nil {
			if docPath, ok := input.SessionValues["document_path"].(string); ok {
				d.state.DocumentPath = docPath
			}
			if outputFormat, ok := input.SessionValues["output_format"].(string); ok {
				d.state.OutputFormat = outputFormat
			}
		}

		stream.Send(&AgentEvent{
			Type:      "agent_start",
			AgentName: d.Name(ctx),
			Content:   fmt.Sprintf("开始处理文档: %s", d.state.DocumentPath),
			Timestamp: time.Now(),
		})

		totalSteps := len(d.state.ProcessingSteps)

		for step := d.state.CurrentStep; step <= totalSteps; step++ {
			stepName := d.state.ProcessingSteps[step-1]

			stream.Send(&AgentEvent{
				Type:      "step_start",
				AgentName: d.Name(ctx),
				Content:   fmt.Sprintf("执行步骤 %d: %s", step, stepName),
				Data: map[string]interface{}{
					"step":      step,
					"step_name": stepName,
					"total":     totalSteps,
					"progress":  float64(step-1) / float64(totalSteps) * 100,
				},
				Timestamp: time.Now(),
			})

			// 模拟步骤执行
			result, shouldInterrupt, interruptReason := d.executeStep(ctx, step, stepName)

			// 更新状态
			d.state.CurrentStep = step
			d.state.ProcessedData[fmt.Sprintf("step_%d_result", step)] = result

			// 保存检查点
			if err := d.saveCheckpointToStore(ctx); err != nil {
				stream.Send(&AgentEvent{
					Type:      "error",
					AgentName: d.Name(ctx),
					Content:   fmt.Sprintf("保存检查点失败: %v", err),
					Timestamp: time.Now(),
				})
			}

			if shouldInterrupt {
				d.state.PauseReason = interruptReason

				stream.Send(&AgentEvent{
					Type:      "agent_interrupted",
					AgentName: d.Name(ctx),
					Content:   fmt.Sprintf("在步骤 %d 中断: %s", step, interruptReason),
					Data: map[string]interface{}{
						"interrupt_info": &InterruptInfo{
							Data: map[string]interface{}{
								"reason":       interruptReason,
								"current_step": step,
								"step_name":    stepName,
								"context":      d.state,
							},
						},
						"checkpoint_saved": true,
					},
					Timestamp: time.Now(),
				})
				return
			}

			// 步骤完成
			stream.Send(&AgentEvent{
				Type:      "step_completed",
				AgentName: d.Name(ctx),
				Content:   fmt.Sprintf("步骤 %d 完成: %s", step, stepName),
				Data: map[string]interface{}{
					"step":     step,
					"result":   result,
					"progress": float64(step) / float64(totalSteps) * 100,
				},
				Timestamp: time.Now(),
			})

			time.Sleep(500 * time.Millisecond) // 模拟处理时间
		}

		// 所有步骤完成
		d.state.CurrentStep = totalSteps + 1

		stream.Send(&AgentEvent{
			Type:      "agent_completed",
			AgentName: d.Name(ctx),
			Content:   "文档处理完成",
			Data: map[string]interface{}{
				"final_result":    d.state.ProcessedData,
				"document_path":   d.state.DocumentPath,
				"output_format":   d.state.OutputFormat,
				"total_duration":  time.Since(d.state.StartTime),
				"steps_completed": len(d.state.ProcessingSteps),
			},
			Timestamp: time.Now(),
		})

		// 清理检查点
		d.checkpointStore.Delete(ctx, d.agentID)
	}()

	return stream
}

func (d *DocumentProcessorAgent) Resume(ctx context.Context, info *ResumeInfo) *AsyncIterator[*AgentEvent] {
	stream := NewAsyncIterator[*AgentEvent]()

	go func() {
		defer stream.Close()

		// 恢复状态
		if err := d.LoadCheckpoint(ctx, info.CheckPointData); err != nil {
			stream.Send(&AgentEvent{
				Type:      "error",
				AgentName: d.Name(ctx),
				Content:   fmt.Sprintf("加载检查点失败: %v", err),
				Timestamp: time.Now(),
			})
			return
		}

		stream.Send(&AgentEvent{
			Type:      "agent_resume",
			AgentName: d.Name(ctx),
			Content:   fmt.Sprintf("从步骤 %d 恢复执行，原因: %s", d.state.CurrentStep, info.ResumeReason),
			Data: map[string]interface{}{
				"resume_reason":  info.ResumeReason,
				"resume_context": info.ResumeContext,
				"pause_reason":   d.state.PauseReason,
				"elapsed_time":   time.Since(d.state.StartTime),
			},
			Timestamp: time.Now(),
		})

		// 处理新的用户输入
		if info.NewInput != nil && len(info.NewInput.Messages) > 0 {
			userMessage := info.NewInput.Messages[len(info.NewInput.Messages)-1].Content
			d.state.UserDecisions = append(d.state.UserDecisions, userMessage)

			stream.Send(&AgentEvent{
				Type:      "user_input_received",
				AgentName: d.Name(ctx),
				Content:   fmt.Sprintf("收到用户指令: %s", userMessage),
				Timestamp: time.Now(),
			})
		}

		// 从当前步骤继续执行
		continueEvents := d.Run(ctx, info.NewInput)

		// 转发所有事件
		for {
			event, ok := continueEvents.Next()
			if !ok {
				break
			}
			stream.Send(event)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	return stream
}

func (d *DocumentProcessorAgent) SaveCheckpoint(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(d.state); err != nil {
		return nil, fmt.Errorf("序列化状态失败: %w", err)
	}
	return buf.Bytes(), nil
}

func (d *DocumentProcessorAgent) LoadCheckpoint(ctx context.Context, data []byte) error {
	var state DocumentProcessorState
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&state); err != nil {
		return fmt.Errorf("反序列化状态失败: %w", err)
	}

	d.state = state
	return nil
}

func (d *DocumentProcessorAgent) saveCheckpointToStore(ctx context.Context) error {
	data, err := d.SaveCheckpoint(ctx)
	if err != nil {
		return err
	}
	return d.checkpointStore.Set(ctx, d.agentID, data)
}

func (d *DocumentProcessorAgent) executeStep(ctx context.Context, step int, stepName string) (interface{}, bool, string) {
	// 模拟不同步骤的处理逻辑
	switch step {
	case 1: // 文档验证
		return map[string]interface{}{
			"valid":  true,
			"format": "PDF",
			"size":   "2.5MB",
			"pages":  15,
		}, false, ""

	case 2: // OCR识别
		return map[string]interface{}{
			"text_extracted": true,
			"confidence":     0.95,
			"word_count":     1250,
			"language":       "zh-CN",
		}, false, ""

	case 3: // 内容分析
		// 模拟检测到敏感信息需要用户确认
		return map[string]interface{}{
			"analysis_complete": true,
			"sections":          []string{"摘要", "正文", "结论"},
			"sensitive_info":    []string{"身份证号", "手机号码"},
		}, true, "检测到敏感信息，需要用户确认处理方式"

	case 4: // 信息提取
		return map[string]interface{}{
			"entities": []string{"人名", "地名", "机构名"},
			"keywords": []string{"AI", "机器学习", "深度学习"},
			"summary":  "这是一份关于AI技术应用的报告",
		}, false, ""

	case 5: // 格式转换
		return map[string]interface{}{
			"output_format":      d.state.OutputFormat,
			"output_path":        "/output/document.md",
			"conversion_success": true,
		}, false, ""

	default:
		return nil, false, ""
	}
}

// ============= Runner 实现 =============

type RunnerConfig struct {
	EnableStreaming     bool
	EnableInterrupt     bool
	EnableCheckpoint    bool
	CheckPointStore     CheckPointStore
	CheckPointInterval  time.Duration
	AutoSaveOnInterrupt bool
	AutoResumeOnStart   bool
	MaxResumeRetries    int
}

type Runner struct {
	config RunnerConfig
	ctx    context.Context
}

func NewRunner(ctx context.Context, config RunnerConfig) *Runner {
	return &Runner{
		config: config,
		ctx:    ctx,
	}
}

func (r *Runner) RunAgent(ctx context.Context, agent Agent, input *AgentInput) *AsyncIterator[*AgentEvent] {
	// 检查是否有未完成的检查点
	if r.config.AutoResumeOnStart && r.config.EnableCheckpoint {
		if resumableAgent, ok := agent.(ResumableAgent); ok {
			agentID := fmt.Sprintf("%s_%d", agent.Name(ctx), time.Now().Unix())
			if data, exists, err := r.config.CheckPointStore.Get(ctx, agentID); err == nil && exists {
				fmt.Printf("发现未完成的检查点，自动恢复执行...\n")
				resumeInfo := &ResumeInfo{
					CheckPointData: data,
					ResumeReason:   "auto_resume_on_start",
					ResumeContext:  map[string]interface{}{"auto_resume": true},
				}
				return resumableAgent.Resume(ctx, resumeInfo)
			}
		}
	}

	return agent.Run(ctx, input)
}

func (r *Runner) ResumeAgent(ctx context.Context, agent ResumableAgent, info *ResumeInfo) *AsyncIterator[*AgentEvent] {
	return agent.Resume(ctx, info)
}

// ============= 演示系统 =============

func demonstrateInterruptAndResume() {
	fmt.Println("🎯 Agent 中断与恢复演示")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()

	// 创建检查点存储
	checkpointStore := NewFileCheckPointStore("./demo_data")

	// 创建 Runner
	config := RunnerConfig{
		EnableStreaming:     true,
		EnableInterrupt:     true,
		EnableCheckpoint:    true,
		CheckPointStore:     checkpointStore,
		CheckPointInterval:  time.Second * 10,
		AutoSaveOnInterrupt: true,
		AutoResumeOnStart:   false,
	}

	runner := NewRunner(ctx, config)

	// 创建文档处理 Agent
	agentID := fmt.Sprintf("doc_processor_%d", time.Now().Unix())
	agent := NewDocumentProcessorAgent(checkpointStore, agentID)

	// 准备输入
	input := &AgentInput{
		Messages: []*schema.Message{
			{Role: schema.User, Content: "请处理这个PDF文档"},
		},
		SessionValues: map[string]interface{}{
			"document_path": "/demo/important_document.pdf",
			"output_format": "markdown",
		},
	}

	fmt.Println("📋 开始文档处理任务...")
	fmt.Println("💡 任务将在步骤3（内容分析）时中断，模拟需要用户确认")
	fmt.Println()

	// 运行 Agent
	events := runner.RunAgent(ctx, agent, input)

	var checkpointData []byte
	var interrupted = false

	// 处理事件流
	timeout := time.After(time.Second * 30)
	eventCount := 0

	for {
		select {
		case <-timeout:
			if eventCount == 0 {
				fmt.Println("⏰ 没有接收到任何事件，可能是异步处理问题")
			}
			goto handleInterruption
		default:
			event, ok := events.Next()
			if !ok {
				if eventCount == 0 {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				break
			}
			eventCount++

			switch event.Type {
			case "agent_start":
				fmt.Printf("🚀 %s: %s\n", event.AgentName, event.Content)

			case "step_start":
				step := event.Data["step"].(int)
				stepName := event.Data["step_name"].(string)
				progress := event.Data["progress"].(float64)
				fmt.Printf("📝 步骤 %d: %s (进度: %.1f%%)\n", step, stepName, progress)

			case "step_completed":
				step := event.Data["step"].(int)
				progress := event.Data["progress"].(float64)
				fmt.Printf("✅ 步骤 %d 完成 (进度: %.1f%%)\n", step, progress)

			case "agent_interrupted":
				fmt.Printf("⏸️  %s: %s\n", event.AgentName, event.Content)
				checkpointSaved := event.Data["checkpoint_saved"].(bool)
				if checkpointSaved {
					fmt.Println("💾 检查点已保存")
				}
				interrupted = true

				// 获取检查点数据用于恢复
				var err error
				checkpointData, _, err = checkpointStore.Get(ctx, agentID)
				if err != nil {
					fmt.Printf("❌ 获取检查点数据失败: %v\n", err)
				}

			case "agent_completed":
				fmt.Printf("🎉 %s: %s\n", event.AgentName, event.Content)
				totalDuration := event.Data["total_duration"].(time.Duration)
				stepsCompleted := event.Data["steps_completed"].(int)
				fmt.Printf("📊 总耗时: %v, 完成步骤: %d\n", totalDuration, stepsCompleted)

			case "error":
				fmt.Printf("❌ 错误: %s\n", event.Content)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

handleInterruption:

	// 如果任务被中断，演示恢复流程
	if interrupted && checkpointData != nil {
		fmt.Println("\n" + strings.Repeat("-", 60))
		fmt.Println("🔄 演示任务恢复流程")
		fmt.Println("💭 用户经过思考，决定继续处理并忽略敏感信息警告")
		fmt.Println()

		time.Sleep(time.Second * 2) // 模拟用户思考时间

		// 准备恢复信息
		resumeInfo := &ResumeInfo{
			CheckPointData: checkpointData,
			NewInput: &AgentInput{
				Messages: []*schema.Message{
					{Role: schema.User, Content: "继续处理，忽略敏感信息警告"},
				},
			},
			ResumeReason: "user_confirmed",
			ResumeContext: map[string]interface{}{
				"user_decision": "ignore_sensitive_info",
				"resume_time":   time.Now(),
			},
		}

		// 恢复执行
		fmt.Println("🔄 正在恢复任务执行...")
		resumeEvents := runner.ResumeAgent(ctx, agent, resumeInfo)

		// 处理恢复后的事件
		for {
			event, ok := resumeEvents.Next()
			if !ok {
				break
			}

			switch event.Type {
			case "agent_resume":
				fmt.Printf("🔄 %s: %s\n", event.AgentName, event.Content)
				resumeReason := event.Data["resume_reason"].(string)
				fmt.Printf("   恢复原因: %s\n", resumeReason)

			case "user_input_received":
				fmt.Printf("📥 %s: %s\n", event.AgentName, event.Content)

			case "step_start":
				step := event.Data["step"].(int)
				stepName := event.Data["step_name"].(string)
				fmt.Printf("📝 继续步骤 %d: %s\n", step, stepName)

			case "step_completed":
				step := event.Data["step"].(int)
				progress := event.Data["progress"].(float64)
				fmt.Printf("✅ 步骤 %d 完成 (进度: %.1f%%)\n", step, progress)

			case "agent_completed":
				fmt.Printf("🎉 %s: %s\n", event.AgentName, event.Content)
				fmt.Println("✨ 任务恢复并成功完成！")

			case "error":
				fmt.Printf("❌ 错误: %s\n", event.Content)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func demonstrateCheckpointManagement() {
	fmt.Println("\n🎯 检查点管理演示")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()
	checkpointStore := NewFileCheckPointStore("./demo_data")

	// 演示检查点的保存和加载
	fmt.Println("💾 演示检查点保存和加载...")

	// 创建测试状态
	testState := DocumentProcessorState{
		CurrentStep: 3,
		ProcessedData: map[string]interface{}{
			"step_1_result": map[string]interface{}{"valid": true, "format": "PDF"},
			"step_2_result": map[string]interface{}{"text_extracted": true, "confidence": 0.95},
		},
		DocumentPath:  "/test/document.pdf",
		OutputFormat:  "markdown",
		StartTime:     time.Now().Add(-time.Minute * 5),
		PauseReason:   "test_interruption",
		UserDecisions: []string{"continue_processing"},
	}

	// 序列化状态
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(testState); err != nil {
		fmt.Printf("❌ 序列化失败: %v\n", err)
		return
	}

	checkpointData := buf.Bytes()
	fmt.Printf("✅ 状态序列化完成，数据大小: %d bytes\n", len(checkpointData))

	// 保存到存储
	testKey := "test_checkpoint_demo"
	if err := checkpointStore.Set(ctx, testKey, checkpointData); err != nil {
		fmt.Printf("❌ 保存检查点失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 检查点已保存到存储，key: %s\n", testKey)

	// 从存储加载
	loadedData, exists, err := checkpointStore.Get(ctx, testKey)
	if err != nil {
		fmt.Printf("❌ 加载检查点失败: %v\n", err)
		return
	}
	if !exists {
		fmt.Println("❌ 检查点不存在")
		return
	}

	fmt.Printf("✅ 从存储加载检查点，数据大小: %d bytes\n", len(loadedData))

	// 反序列化状态
	var loadedState DocumentProcessorState
	buf = *bytes.NewBuffer(loadedData)
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&loadedState); err != nil {
		fmt.Printf("❌ 反序列化失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 状态反序列化完成\n")
	fmt.Printf("   当前步骤: %d\n", loadedState.CurrentStep)
	fmt.Printf("   文档路径: %s\n", loadedState.DocumentPath)
	fmt.Printf("   输出格式: %s\n", loadedState.OutputFormat)
	fmt.Printf("   暂停原因: %s\n", loadedState.PauseReason)
	fmt.Printf("   已处理数据条数: %d\n", len(loadedState.ProcessedData))

	// 清理测试数据
	if err := checkpointStore.Delete(ctx, testKey); err != nil {
		fmt.Printf("⚠️  清理测试数据失败: %v\n", err)
	} else {
		fmt.Println("🧹 测试数据清理完成")
	}
}

func main() {
	fmt.Println("🎊 Eino ADK Agent 扩展机制完整演示")
	fmt.Println("基于 https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_extension/")
	fmt.Println(strings.Repeat("=", 80))

	// 演示1: 中断与恢复功能
	demonstrateInterruptAndResume()

	time.Sleep(time.Second * 2)

	// 演示2: 检查点管理
	demonstrateCheckpointManagement()

	// 总结
	fmt.Println("\n🎯 Agent 扩展机制核心特性总结")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("✨ 已成功演示的扩展特性:")
	fmt.Println("  🔄 智能中断机制")
	fmt.Println("    - Agent 可以在执行过程中生成中断事件")
	fmt.Println("    - 支持多种中断原因（需要用户输入、资源不足等）")
	fmt.Println("    - 自动保存中断时的完整状态")

	fmt.Println("  💾 检查点状态管理")
	fmt.Println("    - 使用 gob 编码实现高效的状态序列化")
	fmt.Println("    - 支持复杂数据结构的完整保存和恢复")
	fmt.Println("    - 灵活的存储后端（文件系统、Redis 等）")

	fmt.Println("  🔄 无缝任务恢复")
	fmt.Println("    - ResumableAgent 接口支持断点续传")
	fmt.Println("    - 恢复时可以接收新的用户输入")
	fmt.Println("    - 保持完整的执行上下文和历史记录")

	fmt.Println("  🎛️  生命周期管理")
	fmt.Println("    - Runner 提供统一的 Agent 执行管理")
	fmt.Println("    - 支持自动恢复未完成的任务")
	fmt.Println("    - 完整的事件流监控和状态追踪")

	fmt.Println("\n💡 实际应用价值:")
	fmt.Println("  • 长时间任务的可靠执行")
	fmt.Println("  • 用户友好的交互体验")
	fmt.Println("  • 系统资源的优化利用")
	fmt.Println("  • 企业级应用的稳定性保障")

	fmt.Println("\n🚀 适用场景:")
	fmt.Println("  📄 文档处理：大批量文档的智能分析和转换")
	fmt.Println("  📊 数据分析：长时间运行的数据挖掘和机器学习任务")
	fmt.Println("  🔄 工作流自动化：复杂的业务流程自动化")
	fmt.Println("  🤖 智能客服：需要多轮交互的复杂对话场景")

	fmt.Println("\n📚 技术栈:")
	fmt.Println("  • Eino ADK 扩展框架")
	fmt.Println("  • Gob 序列化协议")
	fmt.Println("  • 异步事件流处理")
	fmt.Println("  • 可插拔的存储后端")

	fmt.Println("\n🎉 演示完成！")
	fmt.Printf("🚀 Eino ADK 扩展机制 - 让 AI 智能体具备企业级的可靠性和可扩展性！\n")
}
