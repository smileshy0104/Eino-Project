package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// ============= Eino ADK 多Agent组合模式完整演示 =============
// 展示官方文档中提到的三种Agent组合模式：
// 1. Agent间任务转移 (Transfer)
// 2. Agent作为工具使用 (Tool Usage)
// 3. 层次化Agent结构 (Hierarchical)

// ============= 基础类型定义 =============

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
	Type   string      `json:"type"`
	Target string      `json:"target,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

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

// ============= 核心接口 =============

type Agent interface {
	Name(ctx context.Context) string
	Description(ctx context.Context) string
	Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent]
}

// AgentTool 接口 - Agent作为工具使用
type AgentTool interface {
	Agent
	Invoke(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// ============= 专业Agent实现 =============

// DataAnalysisAgent 数据分析Agent
type DataAnalysisAgent struct {
	name string
}

func NewDataAnalysisAgent() *DataAnalysisAgent {
	return &DataAnalysisAgent{name: "DataAnalysisAgent"}
}

func (d *DataAnalysisAgent) Name(ctx context.Context) string {
	return d.name
}

func (d *DataAnalysisAgent) Description(ctx context.Context) string {
	return "专业数据分析Agent，提供统计分析、数据处理和模式识别服务"
}

func (d *DataAnalysisAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]()

	go func() {
		defer iter.Close()

		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Output:    "📊 数据分析Agent开始工作...",
			Timestamp: time.Now(),
		})

		// 模拟数据分析过程
		steps := []string{
			"正在加载数据集...",
			"执行描述性统计分析...",
			"检测数据异常值...",
			"生成分析报告...",
		}

		for _, step := range steps {
			time.Sleep(300 * time.Millisecond)
			iter.Send(&AgentEvent{
				AgentName: d.name,
				RunPath:   []string{d.name},
				Output:    fmt.Sprintf("📈 %s", step),
				Timestamp: time.Now(),
			})
		}

		// 分析结果
		result := &Message{
			Role:    "assistant",
			Content: "数据分析完成：发现3个关键模式，置信度95%，建议进一步验证",
			Metadata: map[string]interface{}{
				"patterns_found": 3,
				"confidence":     0.95,
				"recommendation": "further_validation",
			},
		}

		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Output:    result,
			Timestamp: time.Now(),
		})

		iter.Send(&AgentEvent{
			AgentName: d.name,
			RunPath:   []string{d.name},
			Action:    &AgentAction{Type: "exit", Data: "数据分析任务完成"},
			Timestamp: time.Now(),
		})
	}()

	return iter
}

// Invoke 实现 AgentTool 接口
func (d *DataAnalysisAgent) Invoke(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dataset, ok := params["dataset"]
	if !ok {
		return "请提供数据集参数", nil
	}

	// 简化的工具调用响应
	return map[string]interface{}{
		"status":   "success",
		"dataset":  dataset,
		"patterns": []string{"趋势上升", "季节性变化", "异常值检测"},
		"summary":  "数据分析完成，发现明显的上升趋势",
	}, nil
}

// ReportAgent 报告生成Agent
type ReportAgent struct {
	name string
}

func NewReportAgent() *ReportAgent {
	return &ReportAgent{name: "ReportAgent"}
}

func (r *ReportAgent) Name(ctx context.Context) string {
	return r.name
}

func (r *ReportAgent) Description(ctx context.Context) string {
	return "专业报告生成Agent，基于分析结果生成格式化的专业报告"
}

func (r *ReportAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]()

	go func() {
		defer iter.Close()

		iter.Send(&AgentEvent{
			AgentName: r.name,
			RunPath:   []string{r.name},
			Output:    "📝 报告Agent开始生成报告...",
			Timestamp: time.Now(),
		})

		// 模拟报告生成过程
		steps := []string{
			"解析输入数据...",
			"构建报告结构...",
			"生成图表和可视化...",
			"格式化最终报告...",
		}

		for _, step := range steps {
			time.Sleep(200 * time.Millisecond)
			iter.Send(&AgentEvent{
				AgentName: r.name,
				RunPath:   []string{r.name},
				Output:    fmt.Sprintf("📋 %s", step),
				Timestamp: time.Now(),
			})
		}

		// 生成最终报告
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

		iter.Send(&AgentEvent{
			AgentName: r.name,
			RunPath:   []string{r.name},
			Output: &Message{
				Role:    "assistant",
				Content: report,
				Metadata: map[string]interface{}{
					"report_type": "data_analysis",
					"format":      "markdown",
				},
			},
			Timestamp: time.Now(),
		})

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

// BusinessAnalysisAgent 业务分析Agent - 展示层次化结构和任务转移
type BusinessAnalysisAgent struct {
	name        string
	dataAgent   Agent
	reportAgent Agent
	toolAgents  map[string]AgentTool
}

func NewBusinessAnalysisAgent() *BusinessAnalysisAgent {
	dataAgent := NewDataAnalysisAgent()
	return &BusinessAnalysisAgent{
		name:        "BusinessAnalysisAgent",
		dataAgent:   dataAgent,
		reportAgent: NewReportAgent(),
		toolAgents: map[string]AgentTool{
			"data_analysis": dataAgent,
		},
	}
}

func (b *BusinessAnalysisAgent) Name(ctx context.Context) string {
	return b.name
}

func (b *BusinessAnalysisAgent) Description(ctx context.Context) string {
	return "高级业务分析Agent，协调多个专业Agent完成复杂的业务分析任务"
}

func (b *BusinessAnalysisAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]()

	go func() {
		defer iter.Close()

		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Output:    "🏢 业务分析Agent开始执行复合任务...",
			Timestamp: time.Now(),
		})

		// 阶段1: 使用Agent作为工具
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Output:    "🔧 阶段1: 调用数据分析工具...",
			Timestamp: time.Now(),
		})

		toolResult, err := b.toolAgents["data_analysis"].Invoke(ctx, map[string]interface{}{
			"dataset": "business_metrics_2024.csv",
		})

		if err != nil {
			iter.Send(&AgentEvent{
				AgentName: b.name,
				RunPath:   []string{b.name},
				Error:     err,
				Timestamp: time.Now(),
			})
			return
		}

		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Output:    fmt.Sprintf("✅ 工具调用结果: %v", toolResult),
			Timestamp: time.Now(),
		})

		// 阶段2: 任务转移到数据分析Agent
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Action: &AgentAction{
				Type:   "transfer",
				Target: b.dataAgent.Name(ctx),
				Data:   "转移详细数据分析任务",
			},
			Timestamp: time.Now(),
		})

		// 执行数据分析并转发事件
		dataInput := &AgentInput{
			Messages: []*Message{
				{
					Role:    "user",
					Content: "请进行详细的数据分析",
				},
			},
		}

		dataIter := b.dataAgent.Run(ctx, dataInput)
		for {
			event, ok := dataIter.Next(ctx)
			if !ok {
				break
			}
			if event != nil {
				event.RunPath = append([]string{b.name}, event.RunPath...)
				iter.Send(event)
			}
		}

		// 阶段3: 转移到报告生成Agent
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Action: &AgentAction{
				Type:   "transfer",
				Target: b.reportAgent.Name(ctx),
				Data:   "转移报告生成任务",
			},
			Timestamp: time.Now(),
		})

		// 执行报告生成
		reportInput := &AgentInput{
			Messages: []*Message{
				{
					Role:    "user",
					Content: "基于分析结果生成业务报告",
				},
			},
		}

		reportIter := b.reportAgent.Run(ctx, reportInput)
		for {
			event, ok := reportIter.Next(ctx)
			if !ok {
				break
			}
			if event != nil {
				event.RunPath = append([]string{b.name}, event.RunPath...)
				iter.Send(event)
			}
		}

		// 最终总结
		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Output: &Message{
				Role:    "assistant",
				Content: "🎉 业务分析完成！已完成数据分析、报告生成和结果汇总。",
				Metadata: map[string]interface{}{
					"workflow_stages": []string{"tool_usage", "task_transfer", "hierarchical_execution"},
					"agents_involved": []string{"DataAnalysisAgent", "ReportAgent"},
				},
			},
			Timestamp: time.Now(),
		})

		iter.Send(&AgentEvent{
			AgentName: b.name,
			RunPath:   []string{b.name},
			Action:    &AgentAction{Type: "exit", Data: "业务分析流程完成"},
			Timestamp: time.Now(),
		})
	}()

	return iter
}

// ============= 演示函数 =============

func demonstrateAgentAsTools() {
	fmt.Println("🎯 模式1: Agent作为工具使用")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()
	dataAgent := NewDataAnalysisAgent()

	fmt.Printf("🔧 将 %s 作为工具调用\n", dataAgent.Name(ctx))

	// 直接作为工具调用
	result, err := dataAgent.Invoke(ctx, map[string]interface{}{
		"dataset": "sample_data.csv",
	})

	if err != nil {
		fmt.Printf("❌ 工具调用失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 工具调用成功:\n")
	if resultMap, ok := result.(map[string]interface{}); ok {
		for key, value := range resultMap {
			fmt.Printf("   %s: %v\n", key, value)
		}
	}
}

func demonstrateHierarchicalComposition() {
	fmt.Println("\n🎯 模式2&3: 层次化结构 + 任务转移")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()
	businessAgent := NewBusinessAnalysisAgent()

	fmt.Printf("🏢 执行业务分析Agent: %s\n", businessAgent.Name(ctx))
	fmt.Printf("📝 描述: %s\n", businessAgent.Description(ctx))
	fmt.Println()

	input := &AgentInput{
		Messages: []*Message{
			{
				Role:    "user",
				Content: "请执行完整的业务数据分析，包括数据处理和报告生成",
			},
		},
		EnableStreaming: true,
		SessionID:       "demo_session_001",
	}

	fmt.Println("▶️  开始执行复合Agent工作流...")
	iter := businessAgent.Run(ctx, input)

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
					// 对于长报告，只显示前几行
					lines := strings.Split(msg.Content, "\n")
					if len(lines) > 8 {
						fmt.Printf("💬 %s...\n", strings.Join(lines[:8], "\n"))
						fmt.Printf("    📄 (完整报告已生成，共%d行)\n", len(lines))
					} else {
						fmt.Printf("💬 %s\n", msg.Content)
					}
				} else {
					fmt.Printf("ℹ️  %v\n", event.Output)
				}
			}

			if event.Action != nil {
				fmt.Printf("🎬 动作: %s", event.Action.Type)
				if event.Action.Target != "" {
					fmt.Printf(" → %s", event.Action.Target)
				}
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

func main() {
	fmt.Println("🎊 Eino ADK 多Agent组合模式完整演示")
	fmt.Println("基于官方文档的三种Agent协作模式")
	fmt.Println(strings.Repeat("=", 80))

	// 演示Agent作为工具使用
	demonstrateAgentAsTools()

	time.Sleep(time.Second)

	// 演示层次化结构和任务转移
	demonstrateHierarchicalComposition()

	// 总结
	fmt.Println("\n🎯 多Agent组合模式总结")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("✨ 成功演示的组合模式:")

	fmt.Println("  🔧 模式1: Agent作为工具使用")
	fmt.Println("    - 直接调用Agent.Invoke()方法")
	fmt.Println("    - 同步返回处理结果")
	fmt.Println("    - 适用于简单的功能调用")

	fmt.Println("  🔄 模式2: Agent间任务转移")
	fmt.Println("    - 使用AgentAction.Type='transfer'")
	fmt.Println("    - 指定Target目标Agent")
	fmt.Println("    - 保持执行上下文连续性")

	fmt.Println("  🏗️  模式3: 层次化Agent结构")
	fmt.Println("    - 主Agent协调多个子Agent")
	fmt.Println("    - RunPath显示调用层次")
	fmt.Println("    - 事件流统一管理")

	fmt.Println("\n💡 关键技术特性:")
	fmt.Println("  • 异步事件流处理")
	fmt.Println("  • 灵活的Agent组合策略")
	fmt.Println("  • 统一的接口抽象")
	fmt.Println("  • 可追溯的执行路径")
	fmt.Println("  • 松耦合的模块化设计")

	fmt.Println("\n🎉 这就是Eino ADK的真正威力！")
}
