# Eino React Agent Demo

## 🤖 项目简介

这是一个基于 Eino 框架的 React Agent 演示项目，展示了如何使用 React 逻辑构建智能代理。React Agent 是一种能够进行复杂多步推理的智能代理架构，它可以根据用户输入自动选择和执行合适的工具来完成任务。

## 🏗️ React Agent 架构

React Agent 基于以下核心概念：

### 🧠 React 逻辑
- **Reason**（推理）：分析用户需求，制定解决方案
- **Act**（行动）：选择并执行合适的工具
- **Observe**（观察）：分析工具执行结果
- **循环**：根据结果决定是否需要进一步行动

### 🔧 核心组件
1. **ChatModel**：支持工具调用的大语言模型
2. **Tools**：可用的工具集合
3. **MessageModifier**：消息预处理器
4. **Graph**：底层编排机制

## 🛠️ 功能特性

本 Demo 集成了三个实用工具：

### 🌤️ 天气查询工具 (WeatherTool)
- **功能**：查询中国主要城市天气情况
- **支持城市**：北京、上海、广州、深圳、杭州、成都、西安、南京等
- **示例**："北京今天天气怎么样？"

### 🔢 数学计算工具 (CalculatorTool)  
- **功能**：基本四则运算
- **支持操作**：加法(+)、减法(-)、乘法(*)、除法(/)
- **示例**："帮我计算 125 + 237"

### ⏰ 时间查询工具 (TimeTool)
- **功能**：获取当前日期和时间
- **格式选项**：
  - `date`：仅日期
  - `time`：仅时间  
  - `datetime`：完整日期时间
  - `timestamp`：Unix时间戳
- **示例**："现在几点了？"

## 🚀 快速开始

### 环境要求

- Go 1.24.2+
- Eino v0.4.7+
- 有效的火山方舟 API Key

### 配置设置

确保项目根目录的 `config.yaml` 文件包含以下配置：

```yaml
ARK_API_KEY: "your-api-key-here"
ARK_MODEL: "doubao-seed-1-6-250615"
```

### 运行演示

```bash
cd react_agent_demo
go run main.go
```

### 🎯 测试用例

Demo 包含以下预设测试用例：

1. **单工具调用**：
   - "北京今天天气怎么样？"
   - "帮我计算 125 + 237"  
   - "现在几点了？"

2. **多工具组合**：
   - "上海的天气如何，另外帮我算一下 15 * 8"
   - "查询深圳天气，然后告诉我今天的日期"

## 🔍 代码详解

### Agent 初始化

```go
agent, err := react.NewAgent(ctx, &react.AgentConfig{
    ToolCallingModel: chatModel,      // 支持工具调用的模型
    ToolsConfig:      tools,          // 工具配置列表
    MaxStep:          10,             // 最大执行步骤
    MessageModifier:  messageModifier, // 消息修饰器
})
```

### 工具定义

每个工具都需要实现 `InvokeFunction` 方法：

```go
func (w *WeatherTool) InvokeFunction(ctx context.Context, params map[string]interface{}) (*schema.Message, error) {
    // 工具具体实现
    return schema.FunctionMessage(result, "weather_query"), nil
}
```

### 工具配置

```go
&tool.Config{
    Tool: &WeatherTool{},
    Info: &tool.Info{
        Name:        "weather_query",
        Description: "查询指定城市的天气情况",
        Parameters: &tool.Parameters{
            // 参数定义...
        },
    },
}
```

### 消息修饰器

```go
func messageModifier(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
    systemPrompt := &schema.Message{
        Role: schema.System,
        Content: `你是一个智能助手，名字叫小艾诺（Eino）...`,
    }
    
    if len(msgs) == 0 || msgs[0].Role != schema.System {
        return append([]*schema.Message{systemPrompt}, msgs...), nil
    }
    
    return msgs, nil
}
```

## 🎨 执行流程

1. **用户输入**：接收用户问题
2. **消息预处理**：通过 MessageModifier 添加系统提示
3. **模型推理**：LLM 分析问题并决定是否需要使用工具
4. **工具调用**：如果需要，执行相应工具
5. **结果整合**：将工具结果与推理结合，生成最终回答
6. **多轮交互**：如果需要，重复上述流程

## 📊 高级配置

### 流式处理

```go
// 使用流式处理
stream, err := agent.Stream(ctx, messages, &model.GenerateOptions{})
for {
    chunk, err := stream.Recv()
    if err != nil {
        break
    }
    fmt.Print(chunk.Content)
}
```

### 回调函数

```go
agent.WithCallbacks([]*schema.Callback{
    {
        Type: schema.CallbackTypeOnToolStart,
        Handler: func(ctx context.Context, data interface{}) error {
            fmt.Println("工具开始执行...")
            return nil
        },
    },
})
```

### 工具直接返回

某些工具可以配置为直接返回结果，不经过模型后处理：

```go
&react.AgentConfig{
    // ... 其他配置
    ToolReturnDirectly: []string{"calculator"},
}
```

## 🔧 自定义扩展

### 添加新工具

1. **实现工具接口**：
```go
type MyTool struct{}

func (m *MyTool) InvokeFunction(ctx context.Context, params map[string]interface{}) (*schema.Message, error) {
    // 工具逻辑
    return schema.FunctionMessage("结果", "my_tool"), nil
}
```

2. **配置工具信息**：
```go
&tool.Config{
    Tool: &MyTool{},
    Info: &tool.Info{
        Name:        "my_tool",
        Description: "我的自定义工具",
        Parameters: &tool.Parameters{
            // 参数定义
        },
    },
}
```

3. **添加到工具列表**：
```go
tools = append(tools, myToolConfig)
```

### 自定义消息修饰器

```go
func customMessageModifier(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
    // 自定义消息处理逻辑
    return processedMessages, nil
}
```

## 🎯 最佳实践

1. **工具设计**：
   - 保持工具功能单一明确
   - 提供详细的工具描述
   - 合理设置参数类型和必填项

2. **错误处理**：
   - 优雅处理工具执行错误
   - 提供用户友好的错误信息

3. **性能优化**：
   - 合理设置 MaxStep 避免无限循环
   - 考虑工具执行的时间复杂度

4. **用户体验**：
   - 设计自然的交互方式
   - 提供清晰的输出格式

## 🔍 故障排除

### 常见问题

1. **API Key 错误**：
   - 检查 config.yaml 中的 ARK_API_KEY 配置
   - 确认 API Key 有效且有足够余额

2. **工具调用失败**：
   - 检查工具参数定义是否正确
   - 确认工具实现逻辑无误

3. **消息格式错误**：
   - 确认 MessageModifier 返回正确格式
   - 检查系统提示是否合理

### 调试技巧

1. **启用详细日志**：
```go
// 在创建模型时启用调试模式
chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
    Model: viper.GetString("ARK_MODEL"),
    // 可以添加调试配置
})
```

2. **输出中间状态**：
   - 在工具执行前后添加日志
   - 监控 Agent 的执行步骤

## 📚 相关文档

- [Eino 官方文档](https://www.cloudwego.io/zh/docs/eino/)
- [React Agent 详细指南](https://www.cloudwego.io/zh/docs/eino/core_modules/flow_integration_components/react_agent_manual/)
- [工具系统文档](https://www.cloudwego.io/zh/docs/eino/core_modules/tool/)

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个演示项目！

## 📄 许可证

本项目遵循 MIT 许可证。