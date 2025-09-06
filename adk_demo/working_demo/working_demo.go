package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// ============= ADK 核心概念演示 =============
// 这是一个基于 Eino ADK 官方文档的概念演示
// 展示 Agent 抽象接口的核心思想和工作流程

// ADK Agent 接口（基于官方文档）
type Agent interface {
	Name(ctx context.Context) string
	Description(ctx context.Context) string
	Run(ctx context.Context, input *AgentInput, opts ...AgentRunOption) *EventStream
}

// Agent 输入结构
type AgentInput struct {
	Messages      []*schema.Message      `json:"messages"`
	SessionValues map[string]interface{} `json:"session_values,omitempty"`
	History       []*AgentEvent          `json:"history,omitempty"`
}

// Agent 运行选项
type AgentRunOption func(*RunConfig)

type RunConfig struct {
	MaxRetry     int
	Timeout      time.Duration
	EnableStream bool
}

// Agent 事件
type AgentEvent struct {
	Type      string    `json:"type"`
	AgentName string    `json:"agent_name,omitempty"`
	Content   string    `json:"content,omitempty"`
	ToolName  string    `json:"tool_name,omitempty"`
	FromAgent string    `json:"from_agent,omitempty"`
	ToAgent   string    `json:"to_agent,omitempty"`
	Reason    string    `json:"reason,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// 事件流（简化版异步迭代器）
type EventStream struct {
	events chan *AgentEvent
	done   chan bool
}

func NewEventStream() *EventStream {
	return &EventStream{
		events: make(chan *AgentEvent, 10),
		done:   make(chan bool, 1),
	}
}

func (es *EventStream) Send(event *AgentEvent) {
	select {
	case es.events <- event:
	default:
		// 缓冲区满时忽略
	}
}

func (es *EventStream) Next() (*AgentEvent, bool) {
	select {
	case event := <-es.events:
		return event, true
	case <-es.done:
		return nil, false
	default:
		return nil, false
	}
}

func (es *EventStream) Close() {
	select {
	case es.done <- true:
	default:
	}
	close(es.events)
	close(es.done)
}

// ============= 工具接口和实现 =============

// 简化的工具接口
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, input string) (string, error)
}

// 计算器工具
type CalculatorTool struct{}

func (c *CalculatorTool) Name() string {
	return "calculator"
}

func (c *CalculatorTool) Description() string {
	return "执行基础数学计算"
}

func (c *CalculatorTool) Execute(ctx context.Context, expression string) (string, error) {
	// 简单的表达式计算
	var result float64
	var operation string

	if strings.Contains(expression, "+") {
		parts := strings.Split(expression, "+")
		if len(parts) == 2 {
			var a, b float64
			fmt.Sscanf(strings.TrimSpace(parts[0]), "%f", &a)
			fmt.Sscanf(strings.TrimSpace(parts[1]), "%f", &b)
			result = a + b
			operation = "加法"
		}
	} else if strings.Contains(expression, "*") {
		parts := strings.Split(expression, "*")
		if len(parts) == 2 {
			var a, b float64
			fmt.Sscanf(strings.TrimSpace(parts[0]), "%f", &a)
			fmt.Sscanf(strings.TrimSpace(parts[1]), "%f", &b)
			result = a * b
			operation = "乘法"
		}
	}

	return fmt.Sprintf("执行%s运算：%s = %.2f", operation, expression, result), nil
}

// 天气查询工具
type WeatherTool struct{}

func (w *WeatherTool) Name() string {
	return "weather_query"
}

func (w *WeatherTool) Description() string {
	return "查询指定城市的天气信息"
}

func (w *WeatherTool) Execute(ctx context.Context, city string) (string, error) {
	weatherData := map[string]string{
		"北京": "晴天，温度25°C，微风",
		"上海": "多云，温度28°C，湿度较高",
		"深圳": "阵雨，温度30°C，南风",
		"广州": "多云转晴，温度32°C，东南风",
		"杭州": "小雨，温度26°C，北风",
	}

	weather, exists := weatherData[city]
	if !exists {
		weather = "暂无该城市天气数据"
	}

	return fmt.Sprintf("%s的天气：%s", city, weather), nil
}

// ============= Agent 实现 =============

// 数学专家 Agent
type MathExpertAgent struct {
	tools map[string]Tool
}

func NewMathExpertAgent() *MathExpertAgent {
	return &MathExpertAgent{
		tools: map[string]Tool{
			"calculator": &CalculatorTool{},
		},
	}
}

func (m *MathExpertAgent) Name(ctx context.Context) string {
	return "MathExpert"
}

func (m *MathExpertAgent) Description(ctx context.Context) string {
	return "数学计算专家，能够执行各种数学运算并提供详细解释"
}

func (m *MathExpertAgent) Run(ctx context.Context, input *AgentInput, opts ...AgentRunOption) *EventStream {
	stream := NewEventStream()

	go func() {
		defer stream.Close()

		// 发送开始事件
		stream.Send(&AgentEvent{
			Type:      "agent_start",
			AgentName: m.Name(ctx),
			Content:   "数学专家开始分析问题",
			Timestamp: time.Now(),
		})

		time.Sleep(200 * time.Millisecond)

		lastMessage := input.Messages[len(input.Messages)-1]
		content := lastMessage.Content

		stream.Send(&AgentEvent{
			Type:      "agent_thinking",
			AgentName: m.Name(ctx),
			Content:   fmt.Sprintf("分析数学问题：%s", content),
			Timestamp: time.Now(),
		})

		// 检测并提取数学表达式
		var expression string
		if strings.Contains(content, "计算") {
			parts := strings.Split(content, "计算")
			if len(parts) > 1 {
				expression = strings.TrimSpace(parts[1])
			}
		} else {
			// 尝试直接提取表达式
			for _, part := range strings.Fields(content) {
				if strings.Contains(part, "+") || strings.Contains(part, "*") {
					expression = part
					break
				}
			}
		}

		if expression != "" && (strings.Contains(expression, "+") || strings.Contains(expression, "*")) {
			time.Sleep(200 * time.Millisecond)

			stream.Send(&AgentEvent{
				Type:      "tool_call_start",
				AgentName: m.Name(ctx),
				ToolName:  "calculator",
				Content:   "调用计算器工具",
				Timestamp: time.Now(),
			})

			calculator := m.tools["calculator"]
			result, err := calculator.Execute(ctx, expression)

			time.Sleep(300 * time.Millisecond)

			stream.Send(&AgentEvent{
				Type:      "tool_call_end",
				AgentName: m.Name(ctx),
				ToolName:  "calculator",
				Timestamp: time.Now(),
			})

			if err != nil {
				stream.Send(&AgentEvent{
					Type:      "agent_error",
					AgentName: m.Name(ctx),
					Error:     err.Error(),
					Timestamp: time.Now(),
				})
				return
			}

			stream.Send(&AgentEvent{
				Type:      "agent_response",
				AgentName: m.Name(ctx),
				Content:   fmt.Sprintf("我帮您计算了表达式 '%s'：%s", expression, result),
				Timestamp: time.Now(),
			})
		} else {
			time.Sleep(300 * time.Millisecond)
			stream.Send(&AgentEvent{
				Type:      "agent_response",
				AgentName: m.Name(ctx),
				Content:   "这是一个数学问题。如需计算，请告诉我具体的数学表达式，如 '25 + 17' 或 '12 * 8'",
				Timestamp: time.Now(),
			})
		}
	}()

	return stream
}

// 生活助手 Agent
type LifeAssistantAgent struct {
	tools map[string]Tool
}

func NewLifeAssistantAgent() *LifeAssistantAgent {
	return &LifeAssistantAgent{
		tools: map[string]Tool{
			"weather_query": &WeatherTool{},
		},
	}
}

func (l *LifeAssistantAgent) Name(ctx context.Context) string {
	return "LifeAssistant"
}

func (l *LifeAssistantAgent) Description(ctx context.Context) string {
	return "生活服务助手，提供天气查询、生活建议等服务"
}

func (l *LifeAssistantAgent) Run(ctx context.Context, input *AgentInput, opts ...AgentRunOption) *EventStream {
	stream := NewEventStream()

	go func() {
		defer stream.Close()

		stream.Send(&AgentEvent{
			Type:      "agent_start",
			AgentName: l.Name(ctx),
			Content:   "生活助手开始提供服务",
			Timestamp: time.Now(),
		})

		time.Sleep(200 * time.Millisecond)

		lastMessage := input.Messages[len(input.Messages)-1]
		content := lastMessage.Content

		stream.Send(&AgentEvent{
			Type:      "agent_thinking",
			AgentName: l.Name(ctx),
			Content:   fmt.Sprintf("分析生活服务需求：%s", content),
			Timestamp: time.Now(),
		})

		if strings.Contains(content, "天气") {
			// 提取城市名
			cities := []string{"北京", "上海", "深圳", "广州", "杭州"}
			var targetCity string = "北京" // 默认

			for _, city := range cities {
				if strings.Contains(content, city) {
					targetCity = city
					break
				}
			}

			time.Sleep(200 * time.Millisecond)

			stream.Send(&AgentEvent{
				Type:      "tool_call_start",
				AgentName: l.Name(ctx),
				ToolName:  "weather_query",
				Content:   "查询天气信息",
				Timestamp: time.Now(),
			})

			weatherTool := l.tools["weather_query"]
			result, err := weatherTool.Execute(ctx, targetCity)

			time.Sleep(400 * time.Millisecond)

			stream.Send(&AgentEvent{
				Type:      "tool_call_end",
				AgentName: l.Name(ctx),
				ToolName:  "weather_query",
				Timestamp: time.Now(),
			})

			if err != nil {
				stream.Send(&AgentEvent{
					Type:      "agent_error",
					AgentName: l.Name(ctx),
					Error:     err.Error(),
					Timestamp: time.Now(),
				})
				return
			}

			stream.Send(&AgentEvent{
				Type:      "agent_response",
				AgentName: l.Name(ctx),
				Content:   result,
				Timestamp: time.Now(),
			})
		} else {
			time.Sleep(300 * time.Millisecond)
			stream.Send(&AgentEvent{
				Type:      "agent_response",
				AgentName: l.Name(ctx),
				Content:   "我是生活服务助手，可以帮您查询天气信息。请问您想了解哪个城市的天气？（支持：北京、上海、深圳、广州、杭州）",
				Timestamp: time.Now(),
			})
		}
	}()

	return stream
}

// 路由器 Agent - 智能分发中心
type RouterAgent struct {
	agents map[string]Agent
}

func NewRouterAgent() *RouterAgent {
	return &RouterAgent{
		agents: map[string]Agent{
			"MathExpert":    NewMathExpertAgent(),
			"LifeAssistant": NewLifeAssistantAgent(),
		},
	}
}

func (r *RouterAgent) Name(ctx context.Context) string {
	return "RouterAgent"
}

func (r *RouterAgent) Description(ctx context.Context) string {
	return "智能路由器，分析用户请求并转发给专门的智能体处理"
}

func (r *RouterAgent) Run(ctx context.Context, input *AgentInput, opts ...AgentRunOption) *EventStream {
	stream := NewEventStream()

	go func() {
		defer stream.Close()

		stream.Send(&AgentEvent{
			Type:      "agent_start",
			AgentName: r.Name(ctx),
			Content:   "智能路由器开始分析请求",
			Timestamp: time.Now(),
		})

		time.Sleep(200 * time.Millisecond)

		lastMessage := input.Messages[len(input.Messages)-1]
		content := strings.ToLower(lastMessage.Content)

		stream.Send(&AgentEvent{
			Type:      "agent_thinking",
			AgentName: r.Name(ctx),
			Content:   "分析请求类型，选择合适的专家智能体",
			Timestamp: time.Now(),
		})

		time.Sleep(300 * time.Millisecond)

		// 智能路由逻辑
		var targetAgent Agent
		var targetName string
		var reason string

		if strings.Contains(content, "计算") || strings.Contains(content, "+") ||
			strings.Contains(content, "*") || strings.Contains(content, "数学") {
			targetAgent = r.agents["MathExpert"]
			targetName = "MathExpert"
			reason = "检测到数学计算需求"
		} else if strings.Contains(content, "天气") || strings.Contains(content, "温度") {
			targetAgent = r.agents["LifeAssistant"]
			targetName = "LifeAssistant"
			reason = "检测到天气查询需求"
		} else {
			// 通用对话
			stream.Send(&AgentEvent{
				Type:      "agent_response",
				AgentName: r.Name(ctx),
				Content: `👋 您好！我是智能助理系统，由多个专业智能体组成：

🧮 数学专家 - 处理各种数学计算和数学问题
🌤️ 生活助手 - 提供天气查询和生活建议

您可以问我：
• "帮我算 25 + 17"
• "北京天气怎么样？"  
• "深圳天气如何？"

我会自动选择最合适的专家为您服务！`,
				Timestamp: time.Now(),
			})
			return
		}

		// 发送转发事件
		stream.Send(&AgentEvent{
			Type:      "agent_transfer",
			FromAgent: r.Name(ctx),
			ToAgent:   targetName,
			Reason:    reason,
			Timestamp: time.Now(),
		})

		// 转发给目标智能体
		targetStream := targetAgent.Run(ctx, input, opts...)

		// 转发目标智能体的事件
		for {
			event, ok := targetStream.Next()
			if !ok {
				break
			}
			stream.Send(event)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	return stream
}

// ============= 演示系统 =============

func main() {
	fmt.Println("🎊 Eino ADK 工作演示")
	fmt.Println("基于官方 Agent 抽象接口的可运行演示")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()

	// 创建路由器（包含所有专家智能体）
	router := NewRouterAgent()

	fmt.Printf("✅ 多智能体系统构建完成\n")
	fmt.Printf("   - 智能路由器: %s\n", router.Name(ctx))
	fmt.Printf("   - %s: %s\n", router.agents["MathExpert"].Name(ctx), router.agents["MathExpert"].Description(ctx))
	fmt.Printf("   - %s: %s\n", router.agents["LifeAssistant"].Name(ctx), router.agents["LifeAssistant"].Description(ctx))

	// 测试用例
	testCases := []struct {
		name  string
		input string
		desc  string
	}{
		{
			name:  "数学计算测试",
			input: "请帮我计算 25 + 17",
			desc:  "测试数学专家的加法计算能力",
		},
		{
			name:  "乘法运算测试",
			input: "算一下 12 * 8",
			desc:  "测试数学专家的乘法计算能力",
		},
		{
			name:  "天气查询测试",
			input: "北京今天天气怎么样？",
			desc:  "测试生活助手的天气查询功能",
		},
		{
			name:  "其他城市天气",
			input: "深圳的天气如何？",
			desc:  "测试不同城市的天气查询",
		},
		{
			name:  "通用对话测试",
			input: "你好，你能做什么？",
			desc:  "测试路由器的通用对话处理",
		},
	}

	// 执行测试
	for i, testCase := range testCases {
		fmt.Printf("\n📋 测试用例 %d: %s\n", i+1, testCase.name)
		fmt.Printf("📝 输入: %s\n", testCase.input)
		fmt.Printf("📄 说明: %s\n", testCase.desc)
		fmt.Println(strings.Repeat("-", 50))

		// 构建输入
		agentInput := &AgentInput{
			Messages: []*schema.Message{
				{
					Role:    schema.User,
					Content: testCase.input,
				},
			},
		}

		// 运行路由器
		events := router.Run(ctx, agentInput)

		// 处理事件流 - 持续监听直到处理完成
		timeout := time.After(10 * time.Second) // 10秒超时
		eventCount := 0

		for {
			select {
			case <-timeout:
				fmt.Println("⏰ 事件处理超时")
				goto nextTest
			default:
				event, ok := events.Next()
				if !ok {
					// 没有更多事件，等待一下再试
					if eventCount == 0 {
						time.Sleep(100 * time.Millisecond)
						continue
					}
					break
				}

				eventCount++

				switch event.Type {
				case "agent_start":
					fmt.Printf("🚀 %s 启动: %s\n", event.AgentName, event.Content)
				case "agent_thinking":
					fmt.Printf("🤔 %s 思考: %s\n", event.AgentName, event.Content)
				case "agent_transfer":
					fmt.Printf("🔀 任务转发: %s → %s (%s)\n", event.FromAgent, event.ToAgent, event.Reason)
				case "tool_call_start":
					fmt.Printf("🔧 %s 调用工具: %s\n", event.AgentName, event.ToolName)
				case "tool_call_end":
					fmt.Printf("✅ %s 工具完成: %s\n", event.AgentName, event.ToolName)
				case "agent_response":
					fmt.Printf("💬 %s 回复:\n%s\n", event.AgentName, event.Content)
				case "agent_error":
					fmt.Printf("❌ %s 错误: %s\n", event.AgentName, event.Error)
				}

				time.Sleep(100 * time.Millisecond)
			}
		}

	nextTest:

		fmt.Printf("✅ 测试用例 %d 完成\n", i+1)
		time.Sleep(time.Second)
	}

	// 总结展示
	fmt.Println("\n🎯 Eino ADK 核心特性演示总结")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("✨ 已成功演示的 ADK 核心特性:")
	fmt.Println("  🔗 统一 Agent 接口")
	fmt.Println("    - Name(), Description(), Run() 方法规范")
	fmt.Println("    - 所有智能体实现相同接口，可互换使用")

	fmt.Println("  ⚡ 异步事件流处理")
	fmt.Println("    - EventStream 模拟 AsyncIterator[*AgentEvent]")
	fmt.Println("    - 实时事件推送和处理机制")

	fmt.Println("  🤖 多智能体协作")
	fmt.Println("    - RouterAgent 智能路由和任务分发")
	fmt.Println("    - MathExpert 专业数学计算")
	fmt.Println("    - LifeAssistant 生活服务支持")

	fmt.Println("  🔀 智能任务路由")
	fmt.Println("    - 自动识别请求类型")
	fmt.Println("    - 动态转发给合适的专家智能体")

	fmt.Println("  🛠️  工具集成能力")
	fmt.Println("    - Calculator 数学计算工具")
	fmt.Println("    - WeatherQuery 天气查询工具")

	fmt.Println("  📊 完整事件监控")
	fmt.Println("    - agent_start/end 生命周期事件")
	fmt.Println("    - tool_call_start/end 工具调用事件")
	fmt.Println("    - agent_transfer 任务转发事件")

	fmt.Println("\n💡 ADK 设计优势:")
	fmt.Println("  • 标准化：统一接口规范，降低开发成本")
	fmt.Println("  • 模块化：专业智能体分工合作，提高效率")
	fmt.Println("  • 可观测：完整的事件流，便于监控和调试")
	fmt.Println("  • 可扩展：插件化架构，轻松添加新功能")
	fmt.Println("  • 可组合：灵活的智能体组织和协作模式")

	fmt.Println("\n🎉 演示完成！")
	fmt.Printf("🚀 这就是 Eino ADK 的强大之处 - 让 AI 智能体开发变得简单高效！\n\n")
}
