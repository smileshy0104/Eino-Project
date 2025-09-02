# Eino React Agent Demo

## 🤖 项目简介
ReAct（Reasoning + Acting）Agent 结合了推理和行动能力，通过思考-行动-观察的循环来解决复杂问题。它能够在执行任务时进行深入的推理，并根据观察结果调整策略，特别适合需要多步推理的复杂场景。

这是一个基于 Eino 框架的 React Agent 演示项目，展示了如何使用"推理-行动"（Reasoning-Acting）循环构建智能代理。React Agent 是一种能够进行复杂多步推理的智能代理架构，它可以根据用户输入自动选择和执行合适的工具来完成任务。

> **注意**：本实现使用 Eino 框架的 `compose` API 构建，采用链式架构模拟 React Agent 的工作模式。

## 🏗️ React Agent 架构

### 🧠 React 逻辑循环
- **Reason**（推理）：分析用户需求，理解意图，决定是否需要使用工具
- **Act**（行动）：选择并调用合适的工具获取信息
- **Observe**（观察）：分析工具执行结果，整合信息
- **Response**（回应）：基于推理和工具结果生成最终回复

### 🔗 链式架构设计

```
用户输入 → [ChatModel 1] → [工具执行器] → [ChatModel 2] → 最终回复
           (意图分析)      (Lambda函数)    (响应生成)
           生成ToolCalls   执行工具调用    整合结果回复
```

### 🔧 核心组件
1. **ARK ChatModel**：支持工具调用的大语言模型（字节跳动火山方舟）
2. **InvokableTool**：标准化的工具接口
3. **Chain Composition**：链式编排组件
4. **Lambda Functions**：自定义消息处理逻辑

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

### 🛠️ 工具系统实现

#### 工具接口定义

每个工具都需要实现 `tool.InvokableTool` 接口：

```go
type WeatherTool struct{}

// Info 返回工具元数据，供 LLM 理解工具用途
func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "get_weather",                    // 工具名称
        Desc: "查询指定城市的天气情况",           // 功能描述
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "city": {
                Type:     "string",             // 参数类型
                Desc:     "要查询天气的城市名称", // 参数说明
                Required: true,                 // 是否必需
            },
        }),
    }, nil
}

// InvokableRun 实际执行工具逻辑
func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    // 解析 JSON 参数
    var args struct {
        City string `json:"city"`
    }
    json.Unmarshal([]byte(argumentsInJSON), &args)
    
    // 执行查询逻辑（这里是模拟数据）
    weather := "晴，25°C，微风"
    result := map[string]interface{}{
        "city":    args.City,
        "weather": weather,
        "message": fmt.Sprintf("🌤️ %s今天的天气：%s", args.City, weather),
    }
    
    resultJSON, _ := json.Marshal(result)
    return string(resultJSON), nil
}
```

### 🔗 React Agent 链构建

#### 第一步：工具绑定

```go
// 创建工具实例
tools := []tool.InvokableTool{
    &WeatherTool{},
    &CalculatorTool{}, 
    &TimeTool{},
}

// 收集工具元数据
toolInfos := make([]*schema.ToolInfo, 0, len(tools))
for _, tool := range tools {
    info, err := tool.Info(ctx)
    toolInfos = append(toolInfos, info)
}

// 🔑 关键步骤：将工具信息绑定到模型
chatModel.BindTools(toolInfos)
```

#### 第二步：构建处理链

```go
chain := compose.NewChain[[]*schema.Message, []*schema.Message]()

// 第一层：意图分析和工具调用生成
chain.AppendChatModel(chatModel, compose.WithNodeName("intent_analysis"))

// 第二层：工具执行器（Lambda 函数）
chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
    // 🎯 关键：msg.ToolCalls 来自上一层 ChatModel 的输出
    if len(msg.ToolCalls) > 0 {
        var allMessages []*schema.Message
        allMessages = append(allMessages, msg) // 保留原始消息
        
        // 遍历执行每个工具调用
        for _, toolCall := range msg.ToolCalls {
            var toolResponse string
            var err error
            
            // 根据工具名称分发执行
            switch toolCall.Function.Name {
            case "get_weather":
                weatherTool := &WeatherTool{}
                toolResponse, err = weatherTool.InvokableRun(ctx, toolCall.Function.Arguments)
            case "calculate":
                calcTool := &CalculatorTool{}
                toolResponse, err = calcTool.InvokableRun(ctx, toolCall.Function.Arguments)
            // ... 其他工具
            }
            
            // 创建工具响应消息
            toolMessage := &schema.Message{
                Role:       schema.Tool,        // 标记为工具消息
                Content:    toolResponse,       // 工具执行结果
                ToolCallID: toolCall.ID,        // 关联原始调用
            }
            allMessages = append(allMessages, toolMessage)
        }
        
        return allMessages, nil
    }
    
    return []*schema.Message{msg}, nil
}), compose.WithNodeName("tool_executor"))

// 第三层：响应生成
chain.AppendChatModel(chatModel, compose.WithNodeName("response_generator"))

// 编译链
compiledChain, err := chain.Compile(ctx)
```

### 🔄 msg.ToolCalls 详细流程

#### 1. 工具调用的生成过程

```go
// 用户输入："北京天气怎么样？"
userMessage := &schema.Message{
    Role:    schema.User,
    Content: "北京天气怎么样？",
}

// ChatModel 接收消息，基于绑定的工具信息进行推理
// LLM 知道有 get_weather 工具可用，决定调用它
// 输出消息包含 ToolCalls：
assistantMessage := &schema.Message{
    Role:    schema.Assistant,
    Content: "",                    // 内容可能为空
    ToolCalls: []schema.ToolCall{   // 🎯 这就是 msg.ToolCalls 的来源！
        {
            ID:   "call_xyz123",
            Type: "function",
            Function: schema.Function{
                Name:      "get_weather",
                Arguments: `{"city": "北京"}`,
            },
        },
    },
}
```

#### 2. 工具执行过程

```go
// Lambda 函数接收到包含 ToolCalls 的消息
func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
    // msg.ToolCalls 由上一层 ChatModel 生成
    for _, toolCall := range msg.ToolCalls {
        // toolCall.Function.Name = "get_weather"
        // toolCall.Function.Arguments = `{"city": "北京"}`
        
        // 执行对应工具
        weatherTool := &WeatherTool{}
        response, _ := weatherTool.InvokableRun(ctx, toolCall.Function.Arguments)
        
        // response = `{"city":"北京","weather":"晴，25°C，微风","message":"🌤️ 北京今天的天气：晴，25°C，微风"}`
    }
    
    // 返回包含工具响应的完整对话历史
    return []*schema.Message{原始消息, 工具响应}, nil
}
```

#### 3. 最终响应生成

```go
// 第三层 ChatModel 接收完整对话历史：
// 1. 用户消息："北京天气怎么样？"
// 2. 助手消息（带工具调用）
// 3. 工具响应消息：包含天气数据
// 
// 基于这些信息生成最终用户友好的回复：
finalResponse := &schema.Message{
    Role:    schema.Assistant,
    Content: "🌤️ 北京今天的天气：晴，25°C，微风。",
}
```

## 🎨 完整执行流程

### 📋 详细步骤解析

```
用户输入: "北京天气怎么样？"
    ↓
[步骤1] 构建输入消息
    ↓
[步骤2] 第一层 ChatModel (意图分析)
    ↓ 输出: msg.ToolCalls = [{name: "get_weather", args: {"city": "北京"}}]
    ↓
[步骤3] Lambda 工具执行器
    ├─ 识别工具调用：get_weather
    ├─ 执行 WeatherTool.InvokableRun()
    └─ 生成工具响应消息
    ↓ 输出: [原始消息, 工具响应消息]
    ↓
[步骤4] 第二层 ChatModel (响应生成)
    ├─ 接收完整对话历史
    ├─ 整合工具结果
    └─ 生成用户友好回复
    ↓
最终输出: "🌤️ 北京今天的天气：晴，25°C，微风。"
```

### 🔄 多工具调用流程

对于复杂查询如"上海天气如何，另外帮我算一下 15 * 8"：

```go
// 第一层输出可能包含多个工具调用
msg.ToolCalls = [
    {Function: {Name: "get_weather", Arguments: `{"city": "上海"}`}},
    {Function: {Name: "calculate", Arguments: `{"expression": "15*8"}`}},
]

// Lambda 函数并行或顺序执行所有工具
for _, toolCall := range msg.ToolCalls {
    // 执行每个工具并收集结果
}

// 第二层 ChatModel 整合所有工具结果生成综合回复
```

### 🔍 调试和监控

代码中包含详细的调试日志：

```go
log.Printf("工具执行阶段 - 消息角色: %s, 工具调用数量: %d", msg.Role, len(msg.ToolCalls))
log.Printf("执行工具: %s, 参数: %s", toolCall.Function.Name, toolCall.Function.Arguments)
log.Printf("工具执行成功: %s", toolResponse)
```

运行时你会看到类似输出：
```
2025/09/01 23:04:44 绑定工具: get_weather - 查询指定城市的天气情况
2025/09/01 23:04:44 绑定工具: calculate - 进行数学计算，支持基本的四则运算
工具执行阶段 - 消息角色: assistant, 工具调用数量: 1
执行工具: get_weather, 参数: {"city":"北京"}
工具执行成功: {"city":"北京","weather":"晴，25°C，微风","message":"🌤️ 北京今天的天气：晴，25°C，微风"}
```

## 📊 高级配置

### 🔧 模型配置优化

```go
// 创建带有优化配置的 ChatModel
config := &ark.ChatModelConfig{
    Model:       viper.GetString("ARK_MODEL"),
    APIKey:      viper.GetString("ARK_API_KEY"),
    Temperature: 0.1,    // 降低随机性，提高一致性
    MaxTokens:   2000,   // 控制输出长度
    TopP:        0.9,    // 控制采样策略
}

chatModel, err := ark.NewChatModel(ctx, config)
```

### ⚡ 性能优化建议

#### 工具执行优化

```go
// 添加超时控制
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

// 并发执行多个独立工具（如果适用）
var wg sync.WaitGroup
for _, toolCall := range msg.ToolCalls {
    wg.Add(1)
    go func(tc schema.ToolCall) {
        defer wg.Done()
        // 执行工具逻辑
    }(toolCall)
}
wg.Wait()
```

#### 缓存工具结果

```go
// 简单的内存缓存示例
var toolCache = make(map[string]string)

func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    // 检查缓存
    if cached, exists := toolCache[argumentsInJSON]; exists {
        return cached, nil
    }
    
    // 执行实际查询
    result := w.actualQuery(argumentsInJSON)
    
    // 缓存结果
    toolCache[argumentsInJSON] = result
    return result, nil
}
```

### 🔄 错误恢复和重试

```go
// 在 Lambda 函数中添加重试逻辑
func executeToolWithRetry(ctx context.Context, toolCall schema.ToolCall) (string, error) {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        result, err := executeTool(ctx, toolCall)
        if err == nil {
            return result, nil
        }
        
        log.Printf("工具执行失败，重试 %d/%d: %v", i+1, maxRetries, err)
        time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
    }
    
    return `{"error": "工具执行失败，已达到最大重试次数"}`, nil
}
```

## 🔧 自定义扩展

### 📝 添加新工具

#### 1. 实现 InvokableTool 接口

```go
type DatabaseTool struct{}

// 定义工具元数据
func (d *DatabaseTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "query_database",
        Desc: "查询数据库中的用户信息",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "user_id": {
                Type:     "string",
                Desc:     "用户ID",
                Required: true,
            },
            "fields": {
                Type:     "array",
                Desc:     "要查询的字段列表",
                Required: false,
            },
        }),
    }, nil
}

// 实现工具执行逻辑
func (d *DatabaseTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    var args struct {
        UserID string   `json:"user_id"`
        Fields []string `json:"fields"`
    }
    
    if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
        return "", fmt.Errorf("参数解析失败: %v", err)
    }
    
    // 执行数据库查询（示例）
    userData := map[string]interface{}{
        "user_id": args.UserID,
        "name":    "张三",
        "email":   "zhangsan@example.com",
        "status":  "active",
    }
    
    result, _ := json.Marshal(userData)
    return string(result), nil
}
```

#### 2. 集成新工具到链中

```go
// 在工具列表中添加新工具
tools := []tool.InvokableTool{
    &WeatherTool{},
    &CalculatorTool{},
    &TimeTool{},
    &DatabaseTool{},  // 新添加的工具
}

// 在 Lambda 函数中添加新工具的分发逻辑
switch toolCall.Function.Name {
case "get_weather":
    // ...
case "query_database":
    dbTool := &DatabaseTool{}
    toolResponse, err = dbTool.InvokableRun(ctx, toolCall.Function.Arguments)
default:
    toolResponse = `{"error": "未知工具"}`
}
```

### 🎨 自定义系统提示

```go
// 修改测试用例中的系统提示
systemMessage := &schema.Message{
    Role: schema.System,
    Content: `你是专业的AI助手，名叫小艾诺（Eino）。

你有以下能力：
🌤️ 天气查询 - 使用 get_weather 工具查询城市天气
🔢 数学计算 - 使用 calculate 工具进行四则运算
⏰ 时间查询 - 使用 get_time 工具获取当前时间
💾 数据查询 - 使用 query_database 工具查询用户数据

请根据用户需求选择合适的工具，并给出友好的回复。
如果需要使用多个工具，请按逻辑顺序依次调用。`,
}
```

### 🔀 自定义工具路由逻辑

```go
// 高级工具路由器
func advancedToolRouter(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
    if len(msg.ToolCalls) == 0 {
        return []*schema.Message{msg}, nil
    }
    
    var allMessages []*schema.Message
    allMessages = append(allMessages, msg)
    
    // 工具执行优先级（某些工具需要先执行）
    priorityTools := []string{"query_database", "get_weather"}
    normalTools := []string{"calculate", "get_time"}
    
    // 分组执行工具
    groups := [][]string{priorityTools, normalTools}
    
    for _, group := range groups {
        for _, toolCall := range msg.ToolCalls {
            if contains(group, toolCall.Function.Name) {
                response := executeSpecificTool(ctx, toolCall)
                toolMessage := &schema.Message{
                    Role:       schema.Tool,
                    Content:    response,
                    ToolCallID: toolCall.ID,
                }
                allMessages = append(allMessages, toolMessage)
            }
        }
    }
    
    return allMessages, nil
}
```

## 🎯 最佳实践

### 🛠️ 工具设计原则

1. **单一职责**：每个工具专注一个特定功能
   ```go
   // ✅ 好的设计
   type WeatherTool struct{}  // 只负责天气查询
   
   // ❌ 避免的设计
   type MultiTool struct{}    // 同时处理天气、计算、时间
   ```

2. **详细描述**：提供清晰的工具和参数说明
   ```go
   &schema.ParameterInfo{
       Type:     "string",
       Desc:     "城市名称，支持中国主要城市，如：北京、上海、广州", // 具体说明
       Required: true,
       Enum:     []string{"北京", "上海", "广州"},  // 提供枚举值
   }
   ```

3. **健壮的错误处理**：
   ```go
   func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
       // 参数验证
       if argumentsInJSON == "" {
           return "", fmt.Errorf("参数不能为空")
       }
       
       // JSON 解析错误处理
       var args struct { City string `json:"city"` }
       if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
           return "", fmt.Errorf("参数格式错误: %v", err)
       }
       
       // 业务逻辑验证
       if args.City == "" {
           return "", fmt.Errorf("城市名称不能为空")
       }
       
       // 返回结构化结果
       return w.formatResponse(result), nil
   }
   ```

### ⚡ 性能优化策略

1. **工具执行时间控制**：
   ```go
   // 为耗时工具添加超时
   ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
   defer cancel()
   ```

2. **避免无限工具调用循环**：
   ```go
   // 在 Lambda 函数中添加调用计数
   const MAX_TOOL_CALLS = 10
   if callCount > MAX_TOOL_CALLS {
       return []*schema.Message{msg}, fmt.Errorf("工具调用次数超限")
   }
   ```

3. **合理的模型参数**：
   ```go
   config := &ark.ChatModelConfig{
       Temperature: 0.1,    // 低温度保证一致性
       MaxTokens:   1500,   // 控制输出长度
   }
   ```

### 🎨 用户体验优化

1. **友好的响应格式**：
   ```go
   result := map[string]interface{}{
       "data":    actualData,
       "message": "🌤️ 北京今天天气：晴，25°C，微风", // 用户友好的消息
       "status":  "success",
   }
   ```

2. **渐进式信息展示**：
   ```go
   // 对于复杂查询，分步骤展示过程
   log.Printf("🔍 正在查询%s的天气信息...", city)
   log.Printf("✅ 查询完成，天气状况：%s", weather)
   ```

## 🔍 故障排除

### ⚠️ 常见问题及解决方案

#### 1. **API 相关问题**

**问题**: API Key 无效
```
Error: failed to create chat completion: Error code: 401
```

**解决方案**:
```bash
# 检查配置文件
cat config.yaml
# 确认 API Key 格式正确，无多余空格
ARK_API_KEY: "your-actual-api-key"
ARK_MODEL: "doubao-seed-1-6-250615"
```

#### 2. **工具调用问题**

**问题**: 工具不被调用或调用失败
```
工具执行阶段 - 消息角色: assistant, 工具调用数量: 0
```

**解决方案**:
```go
// 确保工具正确绑定
for _, tool := range tools {
    info, err := tool.Info(ctx)
    if err != nil {
        log.Printf("❌ 工具绑定失败: %v", err)
        continue
    }
    log.Printf("✅ 绑定工具: %s", info.Name)
}
chatModel.BindTools(toolInfos)
```

#### 3. **链编译错误**

**问题**: 类型不匹配
```
graph edge[node_1]-[node_2]: start node's output type[*Message] and end node's input type[[]*Message] mismatch
```

**解决方案**:
```go
// 确保链中每层的输入输出类型匹配
chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
chain.AppendChatModel(chatModel)                                    // 输出: *Message
chain.AppendLambda(func(...*Message) ([]*Message, error) { ... })  // 输入: *Message, 输出: []*Message  
chain.AppendChatModel(chatModel)                                    // 输入: []*Message
```

#### 4. **响应不一致问题**

**问题**: 有时工具调用成功，有时失败

**调试方法**:
```go
// 添加详细日志
chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
    log.Printf("🔍 接收消息 - 角色: %s, 内容: %s, 工具调用数: %d", 
               msg.Role, msg.Content[:min(50, len(msg.Content))], len(msg.ToolCalls))
    
    if len(msg.ToolCalls) == 0 {
        log.Printf("⚠️  没有工具调用，原因可能：")
        log.Printf("   1. 用户问题不明确")
        log.Printf("   2. 系统提示不够清晰") 
        log.Printf("   3. 模型参数设置问题")
    }
    
    // 执行逻辑...
    return result, nil
}))
```

### 🔧 调试工具

#### 1. **消息流跟踪**

```go
// 创建消息跟踪器
func messageTracker(messages []*schema.Message) {
    fmt.Println("📋 当前消息流:")
    for i, msg := range messages {
        fmt.Printf("  [%d] %s: %s\n", i, msg.Role, 
                   msg.Content[:min(100, len(msg.Content))])
        if len(msg.ToolCalls) > 0 {
            fmt.Printf("      🔧 工具调用: %d 个\n", len(msg.ToolCalls))
            for j, tc := range msg.ToolCalls {
                fmt.Printf("        [%d] %s(%s)\n", j, tc.Function.Name, tc.Function.Arguments)
            }
        }
    }
}
```

#### 2. **性能分析**

```go
// 添加执行时间统计
start := time.Now()
defer func() {
    duration := time.Since(start)
    log.Printf("⏱️  执行时间: %v", duration)
    if duration > 10*time.Second {
        log.Printf("⚠️  执行时间过长，建议优化")
    }
}()
```

## 📊 技术架构总结

### 🔗 核心实现架构

```
React Agent Demo 技术栈
├── Eino Framework (v0.4.7+)
│   ├── compose.NewChain() - 链式编排
│   ├── schema.Message - 消息结构
│   └── tool.InvokableTool - 工具接口
├── ARK ChatModel (字节跳动)
│   ├── 支持工具调用 (BindTools)
│   └── doubao-seed-1-6-250615
└── Lambda Functions - 自定义逻辑
    ├── 工具执行器
    └── 消息处理器
```

### ⚡ 关键技术要点

1. **msg.ToolCalls 生成机制**: 
   - LLM 基于 `BindTools()` 绑定的工具信息自动生成
   - 包含工具名称、参数和调用ID的完整结构

2. **链式处理模式**:
   ```
   ChatModel → Lambda → ChatModel
   (生成调用)  (执行工具)  (整合回复)
   ```

3. **工具接口标准化**:
   - `Info()`: 返回工具元数据
   - `InvokableRun()`: 执行工具逻辑

4. **类型安全编排**:
   - 严格的输入输出类型匹配
   - 编译时错误检测

### 🎯 适用场景

- ✅ **智能客服系统**: 集成多种查询工具
- ✅ **数据分析助手**: 结合计算和可视化工具  
- ✅ **运维自动化**: 集成系统监控和操作工具
- ✅ **内容创作辅助**: 整合搜索、计算、格式化工具

## 📚 相关文档

- [Eino 官方文档](https://www.cloudwego.io/zh/docs/eino/)
- [React Agent 详细指南](https://www.cloudwego.io/zh/docs/eino/core_modules/flow_integration_components/react_agent_manual/)
- [工具系统文档](https://www.cloudwego.io/zh/docs/eino/core_modules/tool/)
- [Compose API 参考](https://www.cloudwego.io/zh/docs/eino/core_modules/compose/)
- [ARK 模型配置](https://www.volcengine.com/docs/82379/1099475)

## 📈 下一步扩展

1. **流式处理**: 实现实时响应流
2. **多模态支持**: 集成图片、语音工具
3. **持久化存储**: 添加对话历史管理
4. **分布式部署**: 支持工具的远程调用
5. **监控告警**: 完善的性能和错误监控

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个演示项目！

### 贡献方式
- 🐛 报告 Bug
- 💡 提出新功能建议  
- 🔧 提交代码改进
- 📖 完善文档

### 开发规范
- 遵循 Go 代码规范
- 添加充分的注释和文档
- 包含必要的测试用例

## 📄 许可证

本项目遵循 MIT 许可证。