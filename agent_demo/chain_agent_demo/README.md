# Chain Agent 演示

## 概述

这个演示展示了如何使用 Eino 框架的 Chain 组件构建一个完整的 AI Agent。Agent 集成了大语言模型（LLM）和多个工具，实现了智能的任务管理系统。Chain Agent 是最基础的 Agent 形式，采用线性的处理流程，结构简单，适合顺序执行的任务场景。

## 核心特性

- **Chain 架构**: 使用 `compose.NewChain()` 构建完整的 LLM + Tools 执行流水线
- **工具集成**: 包含 5 个任务管理工具，支持完整的 CRUD 操作
- **ARK 模型**: 使用火山方舟的 doubao 模型进行自然语言理解和工具调用
- **智能对话**: Agent 能理解自然语言并自动选择合适的工具执行任务
- **完整工具执行**: 真正执行工具调用，不是简单的模拟

## 架构设计

```
用户输入 → ChatModel (ARK) → 工具执行器 → 结果返回
         ↑________________Chain 编排________________↓

详细流程：
1. 用户输入自然语言请求
2. ChatModel 理解意图并生成 ToolCalls（如果需要）
3. 工具执行器检测并执行所有工具调用
4. 将工具执行结果整合到消息流中
5. 返回最终处理结果给用户
```

## 工具列表

所有工具都实现了 `tool.InvokableTool` 接口，支持真实的任务管理操作：

1. **AddTodoTool** - 添加待办事项
   - 参数：title (必需), description (可选), priority (可选)
   - 功能：创建新任务并分配唯一ID
   
2. **ListTodosTool** - 列出所有任务
   - 无参数
   - 功能：返回所有任务的详细信息和统计数据
   
3. **UpdateTodoTool** - 更新任务状态
   - 参数：id (必需), completed (必需)
   - 功能：修改指定任务的完成状态
   
4. **DeleteTodoTool** - 删除任务
   - 参数：id (必需)
   - 功能：从系统中移除指定的任务
   
5. **GetTaskStatsTool** - 获取统计信息
   - 无参数
   - 功能：生成任务统计报告（总数、已完成、未完成）

## 关键实现

### 1. 工具创建（InvokableTool 接口）

```go
type AddTodoTool struct{}

// Info 方法定义工具的元信息，供 LLM 理解工具功能
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

// InvokableRun 方法执行实际的工具逻辑
func (t *AddTodoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    var args struct {
        Title       string `json:"title"`
        Description string `json:"description"`
        Priority    string `json:"priority"`
    }

    if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
        return "", fmt.Errorf("参数解析失败: %v", err)
    }

    // 执行业务逻辑：创建任务、存储、返回结果
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
```

### 2. ChatModel 配置与工具绑定

```go
func createChatModel(ctx context.Context, tools []tool.InvokableTool) (*ark.ChatModel, error) {
    // 配置 ARK 模型
    config := &ark.ChatModelConfig{
        Model:  viper.GetString("ARK_MODEL"),    // doubao-seed-1-6-250615
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
    
    // BindTools 让 LLM 了解可用工具，能够在需要时生成工具调用
    chatModel.BindTools(toolInfos)
    return chatModel, nil
}
```

### 3. Chain 构建（完整工具执行流程）

```go
func createAgentChain(ctx context.Context) (compose.Runnable[[]*schema.Message, []*schema.Message], error) {
    // 1. 创建工具集合和聊天模型
    tools := createTools()
    chatModel, err := createChatModel(ctx, tools)
    if err != nil {
        return nil, err
    }

    // 2. 构建 Chain - 包含完整的工具执行流程
    chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
    
    // 添加 ChatModel 节点：理解用户意图，可能生成工具调用
    chain.AppendChatModel(chatModel, compose.WithNodeName("chat_model"))
    
    // 添加工具执行器 Lambda - 关键组件！
    // 负责检测和执行 LLM 生成的工具调用
    chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
        // 检查消息是否包含工具调用
        if len(msg.ToolCalls) == 0 {
            return []*schema.Message{msg}, nil
        }

        log.Printf("[ToolExecutor] 检测到 %d 个工具调用", len(msg.ToolCalls))
        
        // 创建工具映射
        toolMap := make(map[string]tool.InvokableTool)
        for _, t := range tools {
            info, err := t.Info(ctx)
            if err != nil {
                continue
            }
            toolMap[info.Name] = t
        }

        // 执行所有工具调用
        toolResults := make([]*schema.Message, 0, len(msg.ToolCalls)+1)
        toolResults = append(toolResults, msg) // 添加原始消息
        
        for _, toolCall := range msg.ToolCalls {
            targetTool, exists := toolMap[toolCall.Function.Name]
            if !exists {
                toolResults = append(toolResults, &schema.Message{
                    Role: schema.Tool,
                    Content: fmt.Sprintf(`{"error": "工具 '%s' 不存在"}`, toolCall.Function.Name),
                    ToolCallID: toolCall.ID,
                })
                continue
            }

            // 执行工具并创建结果消息
            result, err := targetTool.InvokableRun(ctx, toolCall.Function.Arguments)
            if err != nil {
                toolResults = append(toolResults, &schema.Message{
                    Role: schema.Tool,
                    Content: fmt.Sprintf(`{"error": "工具执行失败: %v"}`, err),
                    ToolCallID: toolCall.ID,
                })
                continue
            }

            toolResults = append(toolResults, &schema.Message{
                Role:       schema.Tool,
                Content:    result,
                ToolCallID: toolCall.ID,
            })
        }

        return toolResults, nil
    }), compose.WithNodeName("tool_executor"))

    // 编译 Chain
    agent, err := chain.Compile(ctx)
    if err != nil {
        return nil, fmt.Errorf("编译 Chain 失败: %v", err)
    }

    return agent, nil
}
```

## 工具调用流程详解

### msg.ToolCalls 的生成和处理

1. **生成阶段**：
   ```go
   // ChatModel 根据用户输入和绑定的工具信息，决定是否生成工具调用
   // 例如用户说："添加一个学习任务"
   // LLM 会生成类似这样的工具调用：
   msg.ToolCalls = []*schema.ToolCall{
       {
           ID: "call_123",
           Function: &schema.ToolCallFunction{
               Name: "add_todo",
               Arguments: `{"title": "学习 Eino 框架", "priority": "high"}`,
           },
       },
   }
   ```

2. **执行阶段**：
   ```go
   // 工具执行器检测到 ToolCalls，逐一执行
   for _, toolCall := range msg.ToolCalls {
       // 1. 根据名称查找工具实例
       targetTool := toolMap[toolCall.Function.Name]
       
       // 2. 执行工具，传入 JSON 参数
       result, err := targetTool.InvokableRun(ctx, toolCall.Function.Arguments)
       
       // 3. 创建工具结果消息
       toolResultMsg := &schema.Message{
           Role:       schema.Tool,
           Content:    result,  // JSON 格式的执行结果
           ToolCallID: toolCall.ID,  // 关联到原始调用
       }
   }
   ```

3. **结果整合**：
   ```go
   // 工具执行器返回消息数组，包含：
   // - 原始助手消息（包含工具调用）
   // - 每个工具的执行结果消息
   return []*schema.Message{
       assistantMsg,    // 原始消息
       toolResult1,     // 工具1结果
       toolResult2,     // 工具2结果...
   }
   ```

## 演示场景

每个场景展示不同的工具使用模式：

1. **添加任务**: "请帮我添加一个学习 Eino 框架的高优先级任务，描述是：深入学习 Eino 的 Chain 和工具集成"
   - 触发：AddTodoTool
   - 展示：参数解析、任务创建、ID分配

2. **批量添加**: "再帮我添加两个任务：1. 写一个博客关于 AI Agent，优先级中等 2. 复习 Go 语言并发编程，优先级低"
   - 触发：多次 AddTodoTool 调用
   - 展示：LLM 理解复杂指令的能力

3. **查看列表**: "显示我所有的任务列表"
   - 触发：ListTodosTool
   - 展示：数据查询和格式化

4. **完成任务**: "请将 ID 为 1 的任务标记为已完成"
   - 触发：UpdateTodoTool
   - 展示：状态修改操作

5. **查看统计**: "显示我的任务统计信息"
   - 触发：GetTaskStatsTool
   - 展示：数据分析和报告生成

## 运行方式

```bash
cd /Users/yuyansong/AiProject/Eino/agent_demo/chain_agent_demo
go run main.go
```

## 配置要求

确保 `config.yaml` 包含：
```yaml
ARK_API_KEY: "your_api_key_here"
ARK_MODEL: "doubao-seed-1-6-250615"
```

## 技术要点

### 1. 工具接口实现
- **必须实现** `tool.InvokableTool` 接口
- **Info()** 方法返回工具元信息，供 LLM 理解
- **InvokableRun()** 方法执行实际业务逻辑
- **参数格式** 必须是 JSON 字符串
- **返回格式** 必须是 JSON 字符串

### 2. 工具执行机制
- **BindTools()** 让 LLM 了解可用工具
- **ToolCalls** 由 LLM 根据用户输入自动生成
- **工具执行器** 负责实际调用工具并处理结果
- **错误处理** 每个工具调用都有独立的错误处理

### 3. Chain 架构优势
- **线性流程** 简单直观，易于理解和调试
- **类型安全** 编译期检查输入输出类型
- **可扩展性** 可以轻松添加更多处理节点
- **性能优秀** 顺序执行，延迟可预测

### 4. 消息流处理
```go
// 输入：用户消息
[]*schema.Message{
    {Role: schema.User, Content: "添加任务"}
}

// ChatModel 处理后：助手消息 + 工具调用
[]*schema.Message{
    {Role: schema.Assistant, Content: "", ToolCalls: [...]},
}

// 工具执行器处理后：完整的对话历史
[]*schema.Message{
    {Role: schema.Assistant, ToolCalls: [...]},    // 原始工具调用
    {Role: schema.Tool, Content: "...", ToolCallID: "..."}, // 工具结果
}
```

## 错误处理策略

1. **配置错误**：启动时检查配置文件和API密钥
2. **编译错误**：Chain 编译失败时提供详细错误信息
3. **工具执行错误**：每个工具调用独立处理，失败不影响其他工具
4. **类型转换错误**：Lambda 函数中处理消息格式转换异常

## 扩展建议

1. **添加更多工具**：
   ```go
   type SearchTodoTool struct{} // 搜索任务
   type ImportTodoTool struct{} // 导入任务
   type ExportTodoTool struct{} // 导出任务
   ```

2. **支持流式响应**：
   ```go
   // 使用 Stream API 提供实时反馈
   stream, err := agent.Stream(ctx, messages)
   ```

3. **集成向量数据库**：
   ```go
   // 添加 RAG 能力，支持知识检索
   type KnowledgeSearchTool struct{}
   ```

4. **完善监控和日志**：
   ```go
   // 添加详细的执行监控和性能分析
   chain.AppendLambda(monitoringLambda)
   ```

## Chain vs Graph 对比

| 特性 | Chain Agent | Graph Agent |
|-----|------------|-------------|
| 结构 | 线性序列 | 有向图 |
| 复杂度 | 简单 | 复杂 |
| 分支逻辑 | 不支持 | 支持 |
| 并行执行 | 不支持 | 支持 |
| 调试难度 | 简单 | 中等 |
| 适用场景 | 顺序任务 | 复杂工作流 |

## 总结

Chain Agent 演示展示了 Eino 框架构建线性 AI Agent 的完整方案。通过 ChatModel + 工具执行器的架构，实现了从自然语言理解到具体任务执行的完整流程。这为构建更复杂的 AI 应用奠定了坚实的技术基础。

关键成功要素：
1. **完整的工具执行机制** - 不仅能生成工具调用，还能真正执行
2. **类型安全的 Chain 构建** - 编译期保证数据流的正确性
3. **灵活的错误处理** - 确保系统的稳定性和可靠性
4. **清晰的架构设计** - 便于理解、维护和扩展