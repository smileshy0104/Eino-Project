// Package main 演示如何使用 Eino 框架构建基于 Chain 的智能 Agent
// Chain Agent 是最基础的 Agent 形式，采用线性的处理流程
// 相比 Graph Agent，Chain Agent 结构简单，适合顺序执行的任务场景
package main

import (
	"context"      // 上下文管理，控制请求生命周期和取消操作
	"encoding/json" // JSON 编解码，用于工具参数和返回结果的序列化
	"fmt"          // 格式化输出和字符串处理
	"log"          // 日志记录，用于调试和监控
	"time"         // 时间处理，用于时间戳和延时操作

	"github.com/cloudwego/eino-ext/components/model/ark" // ARK 大语言模型扩展组件
	"github.com/cloudwego/eino/components/tool"         // Eino 工具组件接口
	"github.com/cloudwego/eino/compose"                 // Eino 编排组件，用于构建 Chain
	"github.com/cloudwego/eino/schema"                  // Eino 模式定义和数据结构
	"github.com/spf13/viper"                           // 配置文件管理库
)

// =============================================================================
//
//  文件: main.go
//  功能: 演示如何使用 Chain 构建 Agent
//  说明: 展示 LLM + Tools 的完整 Agent 实现，包括任务管理工具集
//
// =============================================================================

// TaskItem 表示一个待办事项的完整信息
// 这是 Chain Agent 演示中的核心数据结构，用于任务管理
type TaskItem struct {
	ID          int       `json:"id"`          // 任务唯一标识符
	Title       string    `json:"title"`       // 任务标题
	Description string    `json:"description"` // 任务详细描述
	Completed   bool      `json:"completed"`   // 任务完成状态
	Priority    string    `json:"priority"`    // 任务优先级：high(高)/medium(中)/low(低)
	CreatedAt   time.Time `json:"created_at"`  // 任务创建时间
}

// 全局任务存储
// 在生产环境中，这些数据应该存储在数据库中
var (
	tasks  = make(map[int]*TaskItem) // 任务存储映射，键为任务ID，值为任务对象
	nextID = 1                       // 下一个可用的任务ID
)

// AddTodoTool 添加任务工具
// 这是 Chain Agent 工具集中的核心工具之一，负责创建新任务
// 演示了如何在 Chain 中集成自定义业务逻辑
type AddTodoTool struct{}

// Info 返回工具的元数据信息
// LLM 会根据这些信息决定何时以及如何调用此工具
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

// InvokableRun 执行添加任务的核心逻辑
// 当 LLM 决定调用此工具时，会执行这个方法
// 参数: argumentsInJSON - JSON 格式的参数字符串，opts - 工具选项
// 返回: JSON 格式的执行结果
func (t *AddTodoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体，用于解析 LLM 传入的 JSON 参数
	var args struct {
		Title       string `json:"title"`       // 任务标题
		Description string `json:"description"` // 任务描述
		Priority    string `json:"priority"`    // 任务优先级
	}

	// 解析 LLM 传入的 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 设置默认优先级为中等
	if args.Priority == "" {
		args.Priority = "medium"
	}

	// 记录工具调用日志
	log.Printf("[AddTodo] 添加任务: %s", args.Title)

	// 创建新任务对象
	task := &TaskItem{
		ID:          nextID,           // 分配唯一ID
		Title:       args.Title,       // 设置任务标题
		Description: args.Description, // 设置任务描述
		Completed:   false,            // 新任务默认未完成
		Priority:    args.Priority,    // 设置任务优先级
		CreatedAt:   time.Now(),       // 记录创建时间
	}

	// 保存任务到全局存储并递增ID计数器
	tasks[nextID] = task
	nextID++

	// 构造返回结果，包含成功状态、消息和任务信息
	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("任务 '%s' 添加成功", args.Title),
		"task":    task,
	}

	// 序列化结果为 JSON 返回给 LLM
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// ListTodosTool 列出待办事项工具
// 提供任务查询功能，支持按状态筛选
// 是 Chain Agent 中的查询类工具，展示数据检索能力
type ListTodosTool struct{}

// Info 返回工具的元数据信息
// 定义了工具的名称、描述和参数规范
func (t *ListTodosTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "list_todos",                    // 工具名称
		Desc:        "列出所有待办事项",                  // 工具功能描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}), // 无参数
	}, nil
}

// InvokableRun 执行列出任务的核心逻辑
// 查询并返回所有任务信息，支持状态筛选
// 参数: argumentsInJSON - JSON 格式的参数字符串（本工具无参数）
// 返回: JSON 格式的任务列表
func (t *ListTodosTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 记录工具调用日志
	log.Printf("[ListTodos] 列出所有任务，共 %d 个", len(tasks))

	// 构造返回结果，包含任务统计和详细列表
	result := map[string]interface{}{
		"total_tasks": len(tasks), // 任务总数
		"tasks":       tasks,      // 任务详细列表
	}

	// 序列化结果为 JSON 返回给 LLM
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// UpdateTodoTool 更新任务工具
// 提供任务修改功能，支持更新任务的完成状态
// 是 Chain Agent 中的修改类工具，展示数据更新能力
type UpdateTodoTool struct{}

// Info 返回工具的元数据信息
// 定义了更新任务所需的参数规范
func (t *UpdateTodoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_todo",           // 工具名称
		Desc: "更新待办事项的完成状态",      // 工具功能描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"id": {
				Type:     "integer",
				Desc:     "任务ID",      // 必填参数，用于定位要更新的任务
				Required: true,
			},
			"completed": {
				Type:     "boolean",
				Desc:     "是否已完成",   // 必填参数，新的完成状态
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行更新任务的核心逻辑
// 根据任务ID查找并更新任务的完成状态
// 参数: argumentsInJSON - JSON 格式的参数字符串，包含任务ID和新状态
// 返回: JSON 格式的更新结果
func (t *UpdateTodoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体，用于解析 LLM 传入的 JSON 参数
	var args struct {
		ID        int  `json:"id"`        // 要更新的任务ID
		Completed bool `json:"completed"` // 新的完成状态
	}

	// 解析 LLM 传入的 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 记录工具调用日志
	log.Printf("[UpdateTodo] 更新任务 ID: %d, 完成状态: %v", args.ID, args.Completed)

	// 查找指定ID的任务
	task, exists := tasks[args.ID]
	if !exists {
		return `{"success": false, "message": "任务不存在"}`, nil
	}

	// 更新任务的完成状态
	task.Completed = args.Completed

	// 构造返回结果，包含成功状态、消息和更新后的任务信息
	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("任务 '%s' 更新成功", task.Title),
		"task":    task, // 返回更新后的任务信息
	}

	// 序列化结果为 JSON 返回给 LLM
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// DeleteTodoTool 删除任务工具
// 提供任务删除功能，支持根据ID删除指定任务
// 是 Chain Agent 中的删除类工具，展示数据删除能力
type DeleteTodoTool struct{}

// Info 返回工具的元数据信息
// 定义了删除任务所需的参数规范
func (t *DeleteTodoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "delete_todo",        // 工具名称
		Desc: "删除指定的待办事项",      // 工具功能描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"id": {
				Type:     "integer",
				Desc:     "要删除的任务ID",  // 必填参数，用于定位要删除的任务
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行删除任务的核心逻辑
// 根据任务ID查找并删除指定任务
// 参数: argumentsInJSON - JSON 格式的参数字符串，包含任务ID
// 返回: JSON 格式的删除结果
func (t *DeleteTodoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体，用于解析 LLM 传入的 JSON 参数
	var args struct {
		ID int `json:"id"` // 要删除的任务ID
	}

	// 解析 LLM 传入的 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 记录工具调用日志
	log.Printf("[DeleteTodo] 删除任务 ID: %d", args.ID)

	// 查找指定ID的任务
	task, exists := tasks[args.ID]
	if !exists {
		return `{"success": false, "message": "任务不存在"}`, nil
	}

	// 从全局存储中删除任务
	delete(tasks, args.ID)

	// 构造返回结果，包含成功状态和消息
	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("任务 '%s' 删除成功", task.Title),
	}

	// 序列化结果为 JSON 返回给 LLM
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// GetTaskStatsTool 获取任务统计工具
// 提供任务统计分析功能，生成任务概览报告
// 是 Chain Agent 中的分析类工具，展示数据统计能力
type GetTaskStatsTool struct{}

// Info 返回工具的元数据信息
// 此工具无需参数，直接返回当前任务的统计信息
func (t *GetTaskStatsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "get_task_stats",  // 工具名称
		Desc:        "获取任务统计信息",     // 工具功能描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}), // 无参数
	}, nil
}

// InvokableRun 执行获取任务统计的核心逻辑
// 分析所有任务，计算各种统计指标
// 参数: argumentsInJSON - JSON 格式的参数字符串（本工具无参数）
// 返回: JSON 格式的统计报告
func (t *GetTaskStatsTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 记录工具调用日志
	log.Printf("[GetTaskStats] 获取任务统计")

	// 初始化统计计数器
	total := len(tasks)     // 任务总数
	completed := 0          // 已完成任务数
	pending := 0            // 未完成任务数

	// 遍历所有任务，统计完成和未完成数量
	for _, task := range tasks {
		if task.Completed {
			completed++
		} else {
			pending++
		}
	}

	// 构造统计结果，包含任务数量统计
	result := map[string]interface{}{
		"total":     total,     // 任务总数
		"completed": completed, // 已完成任务数
		"pending":   pending,   // 未完成任务数
	}

	// 序列化结果为 JSON 返回给 LLM
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// loadConfig 加载配置文件
// 从 YAML 配置文件或环境变量中读取 ARK 模型配置
// 配置包含模型名称、API 密钥等敏感信息
func loadConfig() {
	// 设置配置文件名称和类型
	viper.SetConfigName("config") // 配置文件名（不含扩展名）
	viper.SetConfigType("yaml")   // 配置文件类型
	
	// 添加配置文件搜索路径
	viper.AddConfigPath(".")        // 当前目录
	viper.AddConfigPath("./config") // config 子目录

	// 设置环境变量前缀，支持通过环境变量覆盖配置
	viper.SetEnvPrefix("EINO")
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v", err)
		log.Println("将使用环境变量")
	}
}

// createTools 创建工具集合
// 初始化 Chain Agent 所需的所有工具
// 这些工具构成了完整的任务管理功能集
// 返回: 工具接口切片，供 Chain Agent 使用
func createTools() []tool.InvokableTool {
	return []tool.InvokableTool{
		&AddTodoTool{},     // 添加任务工具
		&ListTodosTool{},   // 列出任务工具
		&UpdateTodoTool{},  // 更新任务工具
		&DeleteTodoTool{},  // 删除任务工具
		&GetTaskStatsTool{}, // 获取统计工具
	}
}

// createChatModel 创建聊天模型
// 初始化 ARK 大语言模型，这是 Chain Agent 的核心组件
// ARK 模型负责理解用户意图、选择合适工具并生成回复
// 参数: ctx - 上下文对象, tools - 可调用工具列表
// 返回: ChatModel 实例和可能的错误
func createChatModel(ctx context.Context, tools []tool.InvokableTool) (*ark.ChatModel, error) {
	// 配置 ARK 聊天模型参数
	// 从配置文件中读取模型名称和 API 密钥
	config := &ark.ChatModelConfig{
		Model:  viper.GetString("ARK_MODEL"),    // 从配置文件读取模型名称
		APIKey: viper.GetString("ARK_API_KEY"), // 从配置文件读取 API 密钥
	}

	// 创建 ARK 聊天模型实例
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

// createAgentChain 创建 Agent Chain
// 这是 Chain Agent 的核心构建函数，展示了 Chain 的简洁性
// Chain Agent 采用线性处理流程：用户输入 -> LLM处理 -> 工具调用 -> 结果返回
// 相比 Graph Agent，Chain 结构简单，适合顺序执行的场景
// 参数: ctx - 上下文对象
// 返回: 可运行的 Chain 实例和可能的错误
func createAgentChain(ctx context.Context) (compose.Runnable[[]*schema.Message, []*schema.Message], error) {
	// 1. 创建工具集合
	// 这些工具为 Chain Agent 提供任务管理能力
	tools := createTools()
	log.Printf("创建了 %d 个工具", len(tools))

	// 2. 创建聊天模型
	// ARK 模型是 Chain 的核心，负责理解意图和调用工具
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, err
	}
	log.Println("聊天模型创建成功")

	// 3. 构建 Chain - 使用Lambda来处理类型转换
	// Chain 是最基础的 Agent 形式，处理流程为线性序列
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.AppendChatModel(chatModel, compose.WithNodeName("chat_model"))
	
	// 添加Lambda来将单个Message转换为Message数组
	// 这个转换确保了 Chain 输入输出类型的一致性
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
		return []*schema.Message{msg}, nil
	}), compose.WithNodeName("message_wrapper"))

	// 4. 编译 Chain
	// 编译过程将 Chain 配置转换为可执行的 Agent 实例
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

// main 主函数 - Chain Agent 演示程序入口
// 展示了完整的 Chain Agent 生命周期：
// 1. 配置加载 - 读取模型配置和 API 密钥
// 2. Agent 创建 - 构建 LLM + Tools 的 Chain Agent
// 3. 演示运行 - 执行多个任务管理场景
// 4. 状态展示 - 显示最终的任务状态
// Chain Agent 相比 Graph Agent 结构简单，适合线性处理流程
func main() {
	// 创建根上下文，用于控制整个程序的生命周期
	ctx := context.Background()

	// 1. 加载配置文件和环境变量
	// 配置包含 ARK 模型名称、API 密钥等关键信息
	loadConfig()
	log.Println("配置加载成功")

	// 2. 创建 Chain Agent
	// Chain Agent 是最基础的 Agent 形式，采用线性处理流程
	// 包含 LLM（ARK模型）+ Tools（任务管理工具集）
	agent, err := createAgentChain(ctx)
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	// 3. 运行 Chain Agent 演示
	// 演示包含添加任务、查看列表、更新状态、删除任务、统计分析等场景
	// 展示了 Chain Agent 在任务管理领域的完整能力
	runAgentDemo(ctx, agent)

	// 4. 显示最终任务状态
	// 演示结束后，展示所有任务的最终状态，便于验证 Agent 执行效果
	fmt.Println("\n=== 最终任务状态 ===")
	for id, task := range tasks {
		// 格式化任务状态显示
		status := "未完成"
		if task.Completed {
			status = "已完成"
		}
		// 输出任务详细信息：ID、标题、优先级、完成状态
		fmt.Printf("ID: %d | %s | 优先级: %s | 状态: %s\n",
			id, task.Title, task.Priority, status)
	}
}
