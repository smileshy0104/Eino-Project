package main

import (
	"context"
	"fmt"
	"strings"
	"time"
	"log"
)

// ============= Eino ADK ChatModelAgent 实现真实演示 =============
// 基于 https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_implementation/chat_model_agent/
// 展示 ReAct 模式的"思考-行动-观察"循环

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
	RunPath   []string              `json:"run_path,omitempty"`
	Output    interface{}           `json:"output,omitempty"`
	Action    *AgentAction          `json:"action,omitempty"`
	Error     error                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time             `json:"timestamp"`
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

// ============= 工具接口定义 =============

// BaseTool 基础工具接口
type BaseTool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// ============= 具体工具实现 =============

// BookSearchTool 图书搜索工具
type BookSearchTool struct{}

func (b *BookSearchTool) Name() string {
	return "book_search"
}

func (b *BookSearchTool) Description() string {
	return "搜索相关主题的图书信息，支持按类型、作者、关键词搜索"
}

func (b *BookSearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少查询参数")
	}
	
	// 模拟图书搜索
	books := []map[string]interface{}{}
	
	if strings.Contains(query, "人工智能") || strings.Contains(query, "AI") {
		books = append(books, map[string]interface{}{
			"title":  "深度学习",
			"author": "Ian Goodfellow",
			"rating": 4.8,
			"category": "人工智能",
			"summary": "深度学习领域的经典教材，涵盖了神经网络的理论基础和实践应用",
		})
		books = append(books, map[string]interface{}{
			"title":  "统计学习方法",
			"author": "李航",
			"rating": 4.7,
			"category": "机器学习",
			"summary": "统计学习理论的入门经典，适合初学者系统学习机器学习基础",
		})
	}
	
	if strings.Contains(query, "编程") || strings.Contains(query, "程序设计") {
		books = append(books, map[string]interface{}{
			"title":  "Go语言程序设计",
			"author": "Alan Donovan",
			"rating": 4.6,
			"category": "编程语言",
			"summary": "Go语言权威指南，从基础语法到高级特性的全面介绍",
		})
	}
	
	if len(books) == 0 {
		books = append(books, map[string]interface{}{
			"title":  "通用推荐图书",
			"author": "推荐作者",
			"rating": 4.5,
			"category": "通用",
			"summary": "基于您的查询推荐的相关图书",
		})
	}
	
	return map[string]interface{}{
		"query": query,
		"results": books,
		"count": len(books),
	}, nil
}

// UserProfileTool 用户画像工具
type UserProfileTool struct{}

func (u *UserProfileTool) Name() string {
	return "user_profile"
}

func (u *UserProfileTool) Description() string {
	return "获取用户的兴趣偏好和阅读历史，用于个性化推荐"
}

func (u *UserProfileTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := params["user_id"].(string)
	if !ok {
		userID = "default_user"
	}
	
	// 模拟用户画像数据
	profile := map[string]interface{}{
		"user_id": userID,
		"interests": []string{"人工智能", "编程", "数据科学"},
		"reading_history": []map[string]interface{}{
			{"title": "Python机器学习", "rating": 5, "read_date": "2023-12-01"},
			{"title": "算法导论", "rating": 4, "read_date": "2023-11-15"},
		},
		"preferred_categories": []string{"技术", "科学", "编程"},
		"reading_level": "中高级",
		"language": "中文",
	}
	
	return profile, nil
}

// ============= ChatModel 模拟实现 =============

// SimpleChatModel 简化的聊天模型实现
type SimpleChatModel struct {
	name string
}

func NewSimpleChatModel() *SimpleChatModel {
	return &SimpleChatModel{name: "SimulatedChatModel"}
}

func (s *SimpleChatModel) GenerateResponse(ctx context.Context, messages []*Message, tools []BaseTool) *AsyncIterator[*Message] {
	iter := NewAsyncIterator[*Message]()
	
	go func() {
		defer iter.Close()
		
		// 模拟思考过程
		time.Sleep(300 * time.Millisecond)
		
		lastMessage := messages[len(messages)-1]
		content := lastMessage.Content
		
		// ReAct 模式：思考 -> 行动 -> 观察
		
		// 1. 思考阶段
		iter.Send(&Message{
			Role:    "assistant",
			Content: fmt.Sprintf("🤔 **思考**: 用户询问「%s」，我需要分析这个需求...", content),
			Metadata: map[string]interface{}{
				"step": "thinking",
				"type": "reasoning",
			},
		})
		
		time.Sleep(500 * time.Millisecond)
		
		// 2. 行动阶段 - 决定使用哪些工具
		var toolsToUse []string
		reasoning := ""
		
		if strings.Contains(content, "推荐") || strings.Contains(content, "书") {
			toolsToUse = append(toolsToUse, "user_profile", "book_search")
			reasoning = "这是一个图书推荐请求，我需要先了解用户画像，然后搜索相关图书"
		} else if strings.Contains(content, "搜索") {
			toolsToUse = append(toolsToUse, "book_search")
			reasoning = "这是一个搜索请求，我直接使用图书搜索工具"
		}
		
		iter.Send(&Message{
			Role:    "assistant",
			Content: fmt.Sprintf("🎯 **行动计划**: %s。我将使用工具: %v", reasoning, toolsToUse),
			Metadata: map[string]interface{}{
				"step": "action_planning",
				"tools": toolsToUse,
			},
		})
		
		time.Sleep(400 * time.Millisecond)
		
		// 3. 执行工具调用
		var toolResults []map[string]interface{}
		
		for _, toolName := range toolsToUse {
			for _, tool := range tools {
				if tool.Name() == toolName {
					iter.Send(&Message{
						Role:    "assistant",
						Content: fmt.Sprintf("🔧 **执行工具**: %s - %s", tool.Name(), tool.Description()),
						Metadata: map[string]interface{}{
							"step": "tool_execution",
							"tool": tool.Name(),
						},
					})
					
					time.Sleep(200 * time.Millisecond)
					
					var params map[string]interface{}
					if tool.Name() == "book_search" {
						params = map[string]interface{}{"query": content}
					} else if tool.Name() == "user_profile" {
						params = map[string]interface{}{"user_id": "demo_user"}
					}
					
					result, err := tool.Execute(ctx, params)
					if err != nil {
						iter.Send(&Message{
							Role:    "assistant",
							Content: fmt.Sprintf("❌ 工具执行失败: %v", err),
							Metadata: map[string]interface{}{
								"step": "tool_error",
								"tool": tool.Name(),
								"error": err.Error(),
							},
						})
					} else {
						toolResults = append(toolResults, map[string]interface{}{
							"tool": tool.Name(),
							"result": result,
						})
						
						iter.Send(&Message{
							Role:    "assistant",
							Content: fmt.Sprintf("✅ **工具结果**: %s 执行成功", tool.Name()),
							Metadata: map[string]interface{}{
								"step": "tool_result",
								"tool": tool.Name(),
								"result": result,
							},
						})
					}
					break
				}
			}
		}
		
		time.Sleep(300 * time.Millisecond)
		
		// 4. 观察和总结
		iter.Send(&Message{
			Role:    "assistant",
			Content: "🧐 **观察**: 正在分析工具执行结果，准备生成最终回答...",
			Metadata: map[string]interface{}{
				"step": "observation",
				"results_count": len(toolResults),
			},
		})
		
		time.Sleep(400 * time.Millisecond)
		
		// 5. 生成最终回答
		finalResponse := s.generateFinalResponse(content, toolResults)
		
		iter.Send(&Message{
			Role:    "assistant",
			Content: finalResponse,
			Metadata: map[string]interface{}{
				"step": "final_response",
				"reasoning_complete": true,
			},
		})
	}()
	
	return iter
}

func (s *SimpleChatModel) generateFinalResponse(query string, toolResults []map[string]interface{}) string {
	if len(toolResults) == 0 {
		return "抱歉，我无法处理您的请求，没有找到相关工具结果。"
	}
	
	response := "📚 **基于分析结果的回答**:\n\n"
	
	for _, result := range toolResults {
		toolName := result["tool"].(string)
		toolResult := result["result"]
		
		if toolName == "user_profile" {
			if profile, ok := toolResult.(map[string]interface{}); ok {
				interests := profile["interests"].([]string)
				response += fmt.Sprintf("根据您的兴趣偏好 %v，", interests)
			}
		}
		
		if toolName == "book_search" {
			if searchResult, ok := toolResult.(map[string]interface{}); ok {
				books := searchResult["results"].([]map[string]interface{})
				response += fmt.Sprintf("我为您找到了 %d 本相关图书:\n\n", len(books))
				
				for i, book := range books {
					response += fmt.Sprintf("%d. **%s** by %s\n", i+1, 
						book["title"], book["author"])
					response += fmt.Sprintf("   - 评分: %.1f/5.0\n", book["rating"])
					response += fmt.Sprintf("   - 类别: %s\n", book["category"])
					response += fmt.Sprintf("   - 简介: %s\n\n", book["summary"])
				}
			}
		}
	}
	
	response += "💡 **推荐理由**: 这些图书都是相关领域的优秀作品，结合您的兴趣和需求精心挑选。"
	
	return response
}

// ============= ChatModelAgent 实现 =============

// ChatModelAgentConfig 配置结构
type ChatModelAgentConfig struct {
	Name        string
	Description string
	Instruction string
	Model       *SimpleChatModel
	Tools       []BaseTool
}

// ChatModelAgent ReAct模式的智能Agent
type ChatModelAgent struct {
	name        string
	description string
	instruction string
	model       *SimpleChatModel
	tools       []BaseTool
}

func NewChatModelAgent(config *ChatModelAgentConfig) *ChatModelAgent {
	return &ChatModelAgent{
		name:        config.Name,
		description: config.Description,
		instruction: config.Instruction,
		model:       config.Model,
		tools:       config.Tools,
	}
}

func (c *ChatModelAgent) Name(ctx context.Context) string {
	return c.name
}

func (c *ChatModelAgent) Description(ctx context.Context) string {
	return c.description
}

func (c *ChatModelAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] {
	iter := NewAsyncIterator[*AgentEvent]()
	
	go func() {
		defer iter.Close()
		
		// 发送Agent启动事件
		iter.Send(&AgentEvent{
			AgentName: c.name,
			RunPath:   []string{c.name},
			Output:    fmt.Sprintf("🤖 %s 开始处理请求...", c.name),
			Metadata: map[string]interface{}{
				"agent_type": "ChatModelAgent",
				"react_mode": true,
				"tools_count": len(c.tools),
			},
			Timestamp: time.Now(),
		})
		
		// 准备消息列表（包含系统指令）
		messages := []*Message{
			{
				Role:    "system",
				Content: c.instruction,
			},
		}
		messages = append(messages, input.Messages...)
		
		// 调用模型生成响应
		modelIter := c.model.GenerateResponse(ctx, messages, c.tools)
		
		// 转发模型的流式响应
		for {
			msg, ok := modelIter.Next(ctx)
			if !ok {
				break
			}
			
			if msg != nil {
				// 将模型响应包装为AgentEvent
				iter.Send(&AgentEvent{
					AgentName: c.name,
					RunPath:   []string{c.name},
					Output:    msg,
					Metadata: map[string]interface{}{
						"source": "chat_model",
						"step": msg.Metadata["step"],
					},
					Timestamp: time.Now(),
				})
			}
		}
		
		// 发送Agent完成事件
		iter.Send(&AgentEvent{
			AgentName: c.name,
			RunPath:   []string{c.name},
			Action: &AgentAction{
				Type: "exit",
				Data: "ChatModelAgent处理完成",
			},
			Timestamp: time.Now(),
		})
	}()
	
	return iter
}

// ============= 演示程序 =============

func demonstrateBasicChatModelAgent() {
	fmt.Println("🎯 基础 ChatModelAgent 演示")
	fmt.Println(strings.Repeat("=", 70))
	
	ctx := context.Background()
	
	// 创建工具
	tools := []BaseTool{
		&BookSearchTool{},
		&UserProfileTool{},
	}
	
	// 创建ChatModelAgent
	agent := NewChatModelAgent(&ChatModelAgentConfig{
		Name:        "BookRecommenderAgent",
		Description: "专业的图书推荐智能助手，基于用户需求和偏好推荐合适的图书",
		Instruction: `你是一个专业的图书推荐专家。你需要：
1. 仔细分析用户的需求和兴趣
2. 使用可用工具获取用户画像和搜索相关图书
3. 基于工具结果提供个性化的推荐
4. 遵循 ReAct 模式：思考 -> 行动 -> 观察 -> 回答

请始终保持专业、友好和有帮助的态度。`,
		Model: NewSimpleChatModel(),
		Tools: tools,
	})
	
	fmt.Printf("🤖 Agent: %s\n", agent.Name(ctx))
	fmt.Printf("📝 描述: %s\n", agent.Description(ctx))
	fmt.Printf("🔧 可用工具: %d 个\n", len(tools))
	fmt.Println()
	
	// 创建用户输入
	input := &AgentInput{
		Messages: []*Message{
			{
				Role:    "user",
				Content: "我对人工智能很感兴趣，能推荐一些相关的技术书籍吗？",
			},
		},
		EnableStreaming: true,
		SessionID:       "demo_session",
	}
	
	fmt.Println("▶️  开始 ReAct 推理过程...")
	fmt.Println()
	
	// 执行Agent
	iter := agent.Run(ctx, input)
	
	for {
		event, ok := iter.Next(ctx)
		if !ok {
			break
		}
		
		if event != nil {
			fmt.Printf("📡 [%s] %s: ", event.Timestamp.Format("15:04:05"), event.AgentName)
			
			if event.Output != nil {
				if msg, ok := event.Output.(*Message); ok {
					// 处理不同类型的消息
					if step, exists := msg.Metadata["step"]; exists {
						stepStr := step.(string)
						switch stepStr {
						case "thinking":
							fmt.Printf("💭 %s\n", msg.Content)
						case "action_planning":
							fmt.Printf("📋 %s\n", msg.Content)
						case "tool_execution":
							fmt.Printf("⚙️  %s\n", msg.Content)
						case "tool_result":
							fmt.Printf("📊 %s\n", msg.Content)
						case "observation":
							fmt.Printf("👀 %s\n", msg.Content)
						case "final_response":
							fmt.Printf("💬 %s\n", msg.Content)
						default:
							fmt.Printf("ℹ️  %s\n", msg.Content)
						}
					} else {
						fmt.Printf("💬 %s\n", msg.Content)
					}
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

func demonstrateMultiTurnConversation() {
	fmt.Println("\n🎯 多轮对话演示")
	fmt.Println(strings.Repeat("=", 70))
	
	ctx := context.Background()
	
	// 创建Agent
	agent := NewChatModelAgent(&ChatModelAgentConfig{
		Name:        "TechBookAdvisor",
		Description: "技术图书咨询顾问，帮助开发者选择合适的学习资料",
		Instruction: `你是一位资深的技术图书咨询顾问。你的任务是帮助用户找到最适合的技术书籍。
请使用 ReAct 推理模式，仔细分析用户需求，使用工具获取信息，然后提供专业建议。`,
		Model: NewSimpleChatModel(),
		Tools: []BaseTool{
			&BookSearchTool{},
			&UserProfileTool{},
		},
	})
	
	// 模拟多轮对话
	conversations := []*Message{
		{Role: "user", Content: "我是Go语言初学者，想要一些入门书籍推荐"},
		{Role: "user", Content: "有没有更偏向实战项目的编程书籍？"},
	}
	
	for i, msg := range conversations {
		fmt.Printf("\n💬 第 %d 轮对话\n", i+1)
		fmt.Printf("用户: %s\n", msg.Content)
		fmt.Println(strings.Repeat("-", 50))
		
		input := &AgentInput{
			Messages:        []*Message{msg},
			EnableStreaming: true,
			SessionID:       fmt.Sprintf("multi_turn_%d", i+1),
		}
		
		iter := agent.Run(ctx, input)
		for {
			event, ok := iter.Next(ctx)
			if !ok {
				break
			}
			
			if event != nil && event.Output != nil {
				if msg, ok := event.Output.(*Message); ok {
					if step, exists := msg.Metadata["step"]; exists && step == "final_response" {
						fmt.Printf("助手: %s\n", msg.Content)
					}
				}
			}
		}
	}
}

func main() {
	fmt.Println("🎊 Eino ADK ChatModelAgent 真实演示")
	fmt.Println("基于官方文档的 ReAct 模式智能Agent实现")
	fmt.Println(strings.Repeat("=", 80))
	
	// 基础演示
	demonstrateBasicChatModelAgent()
	
	// 多轮对话演示
	demonstrateMultiTurnConversation()
	
	// 总结
	fmt.Println("\n🎯 ChatModelAgent 核心特性总结")
	fmt.Println(strings.Repeat("=", 80))
	
	fmt.Println("✨ 成功演示的特性:")
	fmt.Println("  🧠 ReAct 推理模式")
	fmt.Println("    - 思考 (Thinking): 分析用户需求")
	fmt.Println("    - 行动 (Acting): 制定工具调用计划")
	fmt.Println("    - 观察 (Observing): 分析工具执行结果")
	fmt.Println("    - 回答 (Responding): 生成最终回答")
	
	fmt.Println("  🛠️  智能工具调用")
	fmt.Println("    - 动态选择合适的工具")
	fmt.Println("    - 基于上下文的参数传递")
	fmt.Println("    - 工具结果的智能分析")
	
	fmt.Println("  📡 流式交互体验")
	fmt.Println("    - 实时显示推理过程")
	fmt.Println("    - 透明的决策链条")
	fmt.Println("    - 结构化的输出格式")
	
	fmt.Println("  🎯 企业级应用价值")
	fmt.Println("    - 个性化推荐系统")
	fmt.Println("    - 智能客服助手")
	fmt.Println("    - 复杂问题分析")
	fmt.Println("    - 多步骤任务执行")
	
	fmt.Println("\n💡 技术优势:")
	fmt.Println("  • 高度可配置的Agent构建")
	fmt.Println("  • 标准化的工具集成接口") 
	fmt.Println("  • 可扩展的推理模式")
	fmt.Println("  • 完整的执行过程可观测性")
	
	fmt.Println("\n🎉 这就是 Eino ADK ChatModelAgent 的真正实力！")
}