package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: main.go
//  功能: 演示如何使用 Chain 构建 Agent
//  说明: 展示 LLM + Tools 的完整 Agent 实现，包括任务管理工具集
//
// =============================================================================

// TaskItem 表示一个待办事项
type TaskItem struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
}

// 全局任务存储
var (
	tasks  = make(map[int]*TaskItem)
	nextID = 1
)

// AddTodoTool 添加任务工具
type AddTodoTool struct{}

func (t *AddTodoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "add_todo",
		Desc: "添加一个新的待办事项",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title": {
				Type:     "string",
				Desc:     "任务标题",
				Required: true,
			},
			"description": {
				Type:     "string",
				Desc:     "任务详细描述",
				Required: false,
			},
			"priority": {
				Type:     "string",
				Desc:     "任务优先级 (high/medium/low)",
				Required: false,
				Enum:     []string{"high", "medium", "low"},
			},
		}),
	}, nil
}

func (t *AddTodoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.Priority == "" {
		args.Priority = "medium"
	}

	log.Printf("[AddTodo] 添加任务: %s", args.Title)

	task := &TaskItem{
		ID:          nextID,
		Title:       args.Title,
		Description: args.Description,
		Completed:   false,
		Priority:    args.Priority,
		CreatedAt:   time.Now(),
	}

	tasks[nextID] = task
	nextID++

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("任务 '%s' 添加成功", args.Title),
		"task":    task,
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// ListTodosTool 列出任务工具
type ListTodosTool struct{}

func (t *ListTodosTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "list_todos",
		Desc:        "列出所有待办事项",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *ListTodosTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	log.Printf("[ListTodos] 列出所有任务，共 %d 个", len(tasks))

	result := map[string]interface{}{
		"total_tasks": len(tasks),
		"tasks":       tasks,
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// UpdateTodoTool 更新任务工具
type UpdateTodoTool struct{}

func (t *UpdateTodoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_todo",
		Desc: "更新待办事项的完成状态",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"id": {
				Type:     "integer",
				Desc:     "任务ID",
				Required: true,
			},
			"completed": {
				Type:     "boolean",
				Desc:     "是否已完成",
				Required: true,
			},
		}),
	}, nil
}

func (t *UpdateTodoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		ID        int  `json:"id"`
		Completed bool `json:"completed"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[UpdateTodo] 更新任务 ID: %d, 完成状态: %v", args.ID, args.Completed)

	task, exists := tasks[args.ID]
	if !exists {
		return `{"success": false, "message": "任务不存在"}`, nil
	}

	task.Completed = args.Completed

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("任务 '%s' 更新成功", task.Title),
		"task":    task,
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// DeleteTodoTool 删除任务工具
type DeleteTodoTool struct{}

func (t *DeleteTodoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "delete_todo",
		Desc: "删除指定的待办事项",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"id": {
				Type:     "integer",
				Desc:     "要删除的任务ID",
				Required: true,
			},
		}),
	}, nil
}

func (t *DeleteTodoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		ID int `json:"id"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[DeleteTodo] 删除任务 ID: %d", args.ID)

	task, exists := tasks[args.ID]
	if !exists {
		return `{"success": false, "message": "任务不存在"}`, nil
	}

	delete(tasks, args.ID)

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("任务 '%s' 删除成功", task.Title),
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// GetTaskStatsTool 获取统计工具
type GetTaskStatsTool struct{}

func (t *GetTaskStatsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "get_task_stats",
		Desc:        "获取任务统计信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *GetTaskStatsTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	log.Printf("[GetTaskStats] 获取任务统计")

	total := len(tasks)
	completed := 0
	pending := 0

	for _, task := range tasks {
		if task.Completed {
			completed++
		} else {
			pending++
		}
	}

	result := map[string]interface{}{
		"total":     total,
		"completed": completed,
		"pending":   pending,
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// 配置加载
func loadConfig() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
}

// 创建工具集合
func createTools() []tool.InvokableTool {
	return []tool.InvokableTool{
		&AddTodoTool{},
		&ListTodosTool{},
		&UpdateTodoTool{},
		&DeleteTodoTool{},
		&GetTaskStatsTool{},
	}
}

// 创建聊天模型
func createChatModel(ctx context.Context, tools []tool.InvokableTool) (*ark.ChatModel, error) {
	// 配置 ARK 聊天模型
	config := &ark.ChatModelConfig{
		Model:  viper.GetString("ARK_MODEL"),
		APIKey: viper.GetString("ARK_API_KEY"),
	}

	chatModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建聊天模型失败: %v", err)
	}

	// 绑定工具到聊天模型
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			log.Printf("获取工具信息失败: %v", err)
			continue
		}
		toolInfos = append(toolInfos, info)
		log.Printf("绑定工具: %s - %s", info.Name, info.Desc)
	}

	chatModel.BindTools(toolInfos)
	return chatModel, nil
}

// 创建 Agent Chain
func createAgentChain(ctx context.Context) (compose.Runnable[[]*schema.Message, []*schema.Message], error) {
	// 1. 创建工具集合
	tools := createTools()
	log.Printf("创建了 %d 个工具", len(tools))

	// 2. 创建聊天模型
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, err
	}
	log.Println("聊天模型创建成功")

	// 3. 构建 Chain - 使用Lambda来处理类型转换
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.AppendChatModel(chatModel, compose.WithNodeName("chat_model"))
	
	// 添加Lambda来将单个Message转换为Message数组
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
		return []*schema.Message{msg}, nil
	}), compose.WithNodeName("message_wrapper"))

	// 4. 编译 Chain
	agent, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译 Chain 失败: %v", err)
	}

	log.Println("Agent Chain 创建成功")
	return agent, nil
}

// 运行 Agent 演示
func runAgentDemo(ctx context.Context, agent compose.Runnable[[]*schema.Message, []*schema.Message]) {
	fmt.Println("\n=== Chain Agent 演示开始 ===")

	// 演示场景列表
	scenarios := []struct {
		name    string
		message string
	}{
		{
			name:    "场景1: 添加任务",
			message: "请帮我添加一个学习 Eino 框架的高优先级任务，描述是：深入学习 Eino 的 Chain 和工具集成",
		},
		{
			name:    "场景2: 添加多个任务",
			message: "再帮我添加两个任务：1. 写一个博客关于 AI Agent，优先级中等 2. 复习 Go 语言并发编程，优先级低",
		},
		{
			name:    "场景3: 查看任务列表",
			message: "显示我所有的任务列表",
		},
		{
			name:    "场景4: 完成任务",
			message: "请将 ID 为 1 的任务标记为已完成",
		},
		{
			name:    "场景5: 查看统计",
			message: "显示我的任务统计信息",
		},
	}

	// 执行每个演示场景
	for _, scenario := range scenarios {
		fmt.Printf("\n--- %s ---\n", scenario.name)
		fmt.Printf("用户: %s\n", scenario.message)

		// 创建用户消息
		messages := []*schema.Message{
			{
				Role:    schema.User,
				Content: scenario.message,
			},
		}

		// 调用 Agent
		resp, err := agent.Invoke(ctx, messages)
		if err != nil {
			log.Printf("Agent 调用失败: %v", err)
			continue
		}

		// 输出 Agent 响应
		for _, msg := range resp {
			if msg.Role == schema.Assistant {
				fmt.Printf("Agent: %s\n", msg.Content)
			}
		}

		// 添加间隔
		time.Sleep(time.Second)
	}

	fmt.Println("\n=== Chain Agent 演示结束 ===")
}

// 主函数
func main() {
	ctx := context.Background()

	// 加载配置
	loadConfig()
	log.Println("配置加载成功")

	// 创建 Agent Chain
	agent, err := createAgentChain(ctx)
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	// 运行演示
	runAgentDemo(ctx, agent)

	// 显示最终任务状态
	fmt.Println("\n=== 最终任务状态 ===")
	for id, task := range tasks {
		status := "未完成"
		if task.Completed {
			status = "已完成"
		}
		fmt.Printf("ID: %d | %s | 优先级: %s | 状态: %s\n",
			id, task.Title, task.Priority, status)
	}
}
