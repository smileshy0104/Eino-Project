# Chain Agent 演示

## 概述

这个演示展示了如何使用 Eino 框架的 Chain 组件构建一个完整的 AI Agent。Agent 集成了大语言模型（LLM）和多个工具，实现了智能的任务管理系统。

## 核心特性

- **Chain 架构**: 使用 `compose.NewChain()` 构建 LLM + Tools 的工作流
- **工具集成**: 包含 5 个任务管理工具
- **ARK 模型**: 使用火山方舟的 doubao 模型
- **智能对话**: Agent 能理解自然语言并自动选择合适的工具

## 架构设计

```
用户输入 → ChatModel (ARK) → Lambda 包装器 → 结果返回
         ↑_____________Chain 编排_____________↓
```

## 工具列表

1. **AddTodoTool** - 添加待办事项
   - 参数：title (必需), description (可选), priority (可选)
   
2. **ListTodosTool** - 列出所有任务
   - 无参数
   
3. **UpdateTodoTool** - 更新任务状态
   - 参数：id (必需), completed (必需)
   
4. **DeleteTodoTool** - 删除任务
   - 参数：id (必需)
   
5. **GetTaskStatsTool** - 获取统计信息
   - 无参数

## 关键实现

### 1. 工具创建（传统接口方式）

```go
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
            // 更多参数...
        }),
    }, nil
}

func (t *AddTodoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    // 工具执行逻辑
}
```

### 2. ChatModel 配置与工具绑定

```go
func createChatModel(ctx context.Context, tools []tool.InvokableTool) (*ark.ChatModel, error) {
    config := &ark.ChatModelConfig{
        Model:  viper.GetString("ARK_MODEL"),
        APIKey: viper.GetString("ARK_API_KEY"),
    }
    
    chatModel, err := ark.NewChatModel(ctx, config)
    if err != nil {
        return nil, err
    }

    // 绑定工具到聊天模型
    toolInfos := make([]*schema.ToolInfo, 0, len(tools))
    for _, tool := range tools {
        info, _ := tool.Info(ctx)
        toolInfos = append(toolInfos, info)
    }
    
    chatModel.BindTools(toolInfos)
    return chatModel, nil
}
```

### 3. Chain 构建（带类型转换）

```go
func createAgentChain(ctx context.Context) (compose.Runnable[[]*schema.Message, []*schema.Message], error) {
    tools := createTools()
    chatModel, err := createChatModel(ctx, tools)
    if err != nil {
        return nil, err
    }

    // 构建 Chain - 使用Lambda来处理类型转换
    chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
    chain.AppendChatModel(chatModel, compose.WithNodeName("chat_model"))
    
    // 添加Lambda来将单个Message转换为Message数组
    chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
        return []*schema.Message{msg}, nil
    }), compose.WithNodeName("message_wrapper"))

    agent, err := chain.Compile(ctx)
    return agent, nil
}
```

## 演示场景

1. **添加任务**: "请帮我添加一个学习 Eino 框架的高优先级任务"
2. **批量添加**: "再帮我添加两个任务：博客写作和复习编程"
3. **查看列表**: "显示我所有的任务列表"
4. **完成任务**: "请将 ID 为 1 的任务标记为已完成"
5. **查看统计**: "显示我的任务统计信息"

## 运行方式

```bash
cd chain_agent_demo
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
- 必须实现 `tool.InvokableTool` 接口
- `Info()` 方法返回工具元信息
- `InvokableRun()` 方法执行实际逻辑
- 参数类型必须是 `...tool.Option`

### 2. 类型转换处理
- ChatModel 输出单个 `*schema.Message`
- Chain 期望输入输出都是 `[]*schema.Message`
- 使用 Lambda 函数处理类型转换

### 3. 工具绑定机制
- 通过 `chatModel.BindTools()` 绑定工具信息
- LLM 根据工具信息决定是否调用工具
- 工具调用结果会自动集成到对话流程中

### 4. 错误处理
- 配置加载错误处理
- Chain 编译错误处理
- 工具执行错误处理

## 扩展建议

1. **添加更多工具**: 参考现有工具模式，创建新的业务工具
2. **完善错误处理**: 增加更详细的错误信息和恢复机制
3. **支持流式处理**: 使用 Eino 的流式API提升用户体验
4. **集成向量搜索**: 结合 RAG 功能提供知识增强能力

这个演示为构建基于 Chain 的 Agent 提供了完整的参考实现！