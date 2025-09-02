# Eino Multi-Agent Hosting Demo

## 🏢 项目简介
Multi Agent 系统由多个协同工作的 Agent 组成，每个 Agent 都有其特定的职责和专长。通过 Agent 间的交互与协作，可以处理更复杂的任务，实现分工协作。这种方式特别适合需要多个专业领域知识结合的场景。

这是一个基于 Eino 框架的多代理托管（Multi-Agent Hosting）演示项目，展示了如何构建一个具有意图识别和任务路由功能的多代理系统。该架构通过 Host Agent 进行意图分析，然后将请求路由到专门的 Specialist Agents 来处理特定任务。

> **架构特点**：本实现使用自定义的 `MultiAgentRouter` 路由器，结合 Eino 的 `compose.Chain` 和 Lambda 函数构建了完整的多代理协作系统。

## 🏗️ 系统架构

### 🧠 核心概念

多代理托管是一种企业级的代理架构模式，实现了"分而治之"的设计理念：

- **Host Agent（主控智能体）**：作为系统的"大脑"和决策中心，专注于意图分析和路由决策
- **Specialist Agents（专家智能体）**：专注于特定领域的任务处理，每个专家配备专门的工具和执行链

### 🔗 核心组件架构

```
Multi-Agent Hosting System 技术架构
├── MultiAgentRouter (路由协调器)
│   ├── Host Agent - ARK ChatModel (意图分析)
│   └── Specialist Agents - Compose Chains (任务执行)
│       ├── WeatherSpecialist (天气专家)
│       ├── CalculatorSpecialist (计算专家)
│       └── TimeSpecialist (时间专家)
├── 每个 Specialist 内部架构:
│   ├── ChatModel (意图理解)
│   ├── Lambda (工具执行器)
│   └── ChatModel (响应生成)
└── InvokableTool 工具接口
    ├── WeatherTool
    ├── CalculatorTool
    └── TimeTool
```

### 架构图

```
用户输入
    ↓
┌─────────────────────┐
│   Host Agent        │ ← 意图识别与路由决策
│   (意图分析中心)      │
└─────────┬───────────┘
          │ 路由决策
          ↓
  ┌───────┴───────┐
  │               │
  ↓               ↓
┌──────────┐ ┌──────────┐ ┌──────────┐
│ Weather  │ │Calculator│ │   Time   │
│Specialist│ │Specialist│ │Specialist│
│(天气专家) │ │(计算专家) │ │(时间专家) │
└──────────┘ └──────────┘ └──────────┘
```

## 🛠️ 组件详解

### 1. Host Agent

**职责**：
- 分析用户输入的意图
- 决定调用哪个专家代理
- 协调多个专家的协作

**特点**：
- 不配置具体的工具
- 专注于自然语言理解和决策
- 作为整个系统的调度中心

### 2. Specialist Agents

#### 🌤️ Weather Specialist（天气专家）
- **专长**：天气信息查询
- **工具**：WeatherTool
- **支持城市**：北京、上海、广州、深圳、杭州、成都、西安、南京等
- **返回信息**：温度、天气状况、风力等

#### 🔢 Calculator Specialist（计算专家）  
- **专长**：数学运算
- **工具**：CalculatorTool
- **支持操作**：加法(+)、减法(-)、乘法(*)、除法(/)
- **返回信息**：计算过程和结果

#### ⏰ Time Specialist（时间专家）
- **专长**：时间和日期查询
- **工具**：TimeTool
- **支持格式**：
  - `date`：仅日期
  - `time`：仅时间
  - `datetime`：完整日期时间
  - `timestamp`：Unix时间戳

## 🚀 快速开始

### 环境要求

- Go 1.24.2+
- Eino v0.4.7+
- 有效的火山方舟 API Key

### 配置设置

确保项目目录包含 `config.yaml` 文件：

```yaml
ARK_API_KEY: "your-api-key-here"
ARK_MODEL: "doubao-seed-1-6-250615"
```

### 安装依赖

```bash
cd multi_agent_hosting_demo
go mod tidy
```

### 运行演示

```bash
go run main.go
```

## 📋 测试用例

Demo 包含以下测试场景：

1. **单一专家调用**：
   - "我想知道北京今天的天气情况" → Weather Specialist
   - "帮我计算 156 + 89" → Calculator Specialist  
   - "现在几点了？" → Time Specialist

2. **复杂意图识别**：
   - "查询上海天气，然后告诉我现在的时间" → 多专家协作
   - "先算 25 * 4，然后告诉我今天的日期" → 跨专家任务

## 🔍 核心代码解析

### 🛠️ 工具系统实现

#### InvokableTool 接口实现

每个工具都实现了 `tool.InvokableTool` 接口：

```go
type WeatherTool struct{}

// Info 返回工具元数据
func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "get_weather",
        Desc: "查询指定城市的天气情况",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "city": {
                Type:     "string",
                Desc:     "要查询天气的城市名称，例如：北京、上海、广州等",
                Required: true,
            },
        }),
    }, nil
}

// InvokableRun 执行工具逻辑
func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    var args struct { City string `json:"city"` }
    json.Unmarshal([]byte(argumentsInJSON), &args)
    
    // 模拟天气数据库
    weatherData := map[string]string{
        "北京": "晴，25°C，微风",
        "上海": "多云，28°C，东南风",
        // ... 更多城市数据
    }
    
    weather, exists := weatherData[args.City]
    if !exists {
        weather = "晴，25°C，微风（默认天气）"
    }
    
    result := map[string]interface{}{
        "city":    args.City,
        "weather": weather,
        "message": fmt.Sprintf("🌤️ %s今天的天气：%s", args.City, weather),
    }
    
    resultJSON, _ := json.Marshal(result)
    return string(resultJSON), nil
}
```

### 🏗️ Host Agent 创建

Host Agent 专注于意图分析，不绑定任何工具：

```go
func createHostAgent(ctx context.Context) (*ark.ChatModel, error) {
    log.Println("创建Host Agent...")
    
    // Host Agent 专注于意图理解和决策，不需要直接调用外部工具
    hostModel, err := createChatModel(ctx, nil)  // nil 表示不绑定工具
    if err != nil {
        return nil, fmt.Errorf("创建Host Agent失败: %v", err)
    }
    
    log.Println("Host Agent创建成功")
    return hostModel, nil
}
```

### 🎯 Specialist Agent 创建

每个专家智能体都配备专门的工具和完整的执行链：

```go
func createWeatherSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
    log.Println("创建Weather Specialist Agent...")
    
    // 配置专业工具集
    tools := []tool.InvokableTool{&WeatherTool{}}
    chatModel, err := createChatModel(ctx, tools)
    if err != nil {
        return nil, fmt.Errorf("创建Weather Specialist失败: %v", err)
    }
    
    // 创建支持工具调用的处理链：ChatModel -> Tool Executor -> ChatModel
    chain := compose.NewChain[[]*schema.Message, *schema.Message]()
    
    // 第一层：理解意图并生成工具调用
    chain.AppendChatModel(chatModel, compose.WithNodeName("weather_intent_analysis"))
    
    // 第二层：执行工具调用
    chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
        if len(msg.ToolCalls) > 0 {
            var allMessages []*schema.Message
            allMessages = append(allMessages, msg)
            
            // 处理每个工具调用
            for _, toolCall := range msg.ToolCalls {
                var toolResponse string
                var err error
                
                if toolCall.Function.Name == "get_weather" {
                    weatherTool := &WeatherTool{}
                    toolResponse, err = weatherTool.InvokableRun(ctx, toolCall.Function.Arguments)
                } else {
                    toolResponse = `{"error": "未知工具"}`
                }
                
                // 创建工具响应消息
                toolMessage := &schema.Message{
                    Role:       schema.Tool,
                    Content:    toolResponse,
                    ToolCallID: toolCall.ID,
                }
                allMessages = append(allMessages, toolMessage)
            }
            
            return allMessages, nil
        }
        
        return []*schema.Message{msg}, nil
    }), compose.WithNodeName("weather_tool_executor"))
    
    // 第三层：基于工具结果生成最终回复
    chain.AppendChatModel(chatModel, compose.WithNodeName("weather_response_generator"))
    
    specialist, err := chain.Compile(ctx)
    if err != nil {
        return nil, fmt.Errorf("编译Weather Specialist失败: %v", err)
    }
    
    log.Println("Weather Specialist Agent创建成功")
    return specialist, nil
}
```

### 🔄 MultiAgentRouter 路由器

自定义路由器实现了完整的多代理协作流程：

```go
type MultiAgentRouter struct {
    hostAgent            *ark.ChatModel                                       // 主控智能体
    weatherSpecialist    compose.Runnable[[]*schema.Message, *schema.Message] // 天气专家
    calculatorSpecialist compose.Runnable[[]*schema.Message, *schema.Message] // 计算专家
    timeSpecialist       compose.Runnable[[]*schema.Message, *schema.Message] // 时间专家
}

func (m *MultiAgentRouter) Invoke(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
    // 第一阶段：Host Agent 意图分析
    hostPrompt := &schema.Message{
        Role: schema.System,
        Content: `你是一个智能助手调度中心。分析用户请求，确定需要调用哪些专家来处理任务。

可用专家：
- weather: 处理天气查询
- calculator: 处理数学计算  
- time: 处理时间查询

请只返回需要的专家名称，多个用逗号分隔。`,
    }
    
    hostInput := append([]*schema.Message{hostPrompt}, input...)
    hostResponse, err := m.hostAgent.Generate(ctx, hostInput)
    if err != nil {
        return nil, fmt.Errorf("Host Agent 分析失败: %v", err)
    }
    
    log.Printf("[Host Agent] 路由决策: %s", hostResponse.Content)
    
    // 第二阶段：专业智能体任务执行
    specialists := strings.Split(strings.TrimSpace(hostResponse.Content), ",")
    var results []string
    
    for _, specialist := range specialists {
        specialist = strings.TrimSpace(specialist)
        
        // 路由到对应的专家
        if strings.Contains(specialist, "weather") {
            log.Println("[Router] 调用 Weather Specialist")
            result, err := m.weatherSpecialist.Invoke(ctx, input)
            if err == nil && result != nil && result.Content != "" {
                results = append(results, result.Content)
            }
        }
        // ... 其他专家的路由逻辑
    }
    
    // 第三阶段：结果整合
    if len(results) == 0 {
        return &schema.Message{
            Role:    schema.Assistant,
            Content: "抱歉，我无法处理您的请求。",
        }, nil
    }
    
    finalContent := strings.Join(results, "\n\n")
    return &schema.Message{
        Role:    schema.Assistant,
        Content: finalContent,
    }, nil
}
```

### 🚀 系统组装

```go
func createMultiAgentSystem(ctx context.Context) (*MultiAgentRouter, error) {
    log.Println("创建Multi-Agent Hosting系统...")
    
    // 创建 Host Agent
    hostAgent, err := createHostAgent(ctx)
    if err != nil {
        return nil, err
    }
    
    // 创建各个专家智能体
    weatherSpecialist, err := createWeatherSpecialist(ctx)
    if err != nil {
        return nil, err
    }
    
    calculatorSpecialist, err := createCalculatorSpecialist(ctx)
    if err != nil {
        return nil, err
    }
    
    timeSpecialist, err := createTimeSpecialist(ctx)
    if err != nil {
        return nil, err
    }
    
    // 组装路由器
    router := &MultiAgentRouter{
        hostAgent:            hostAgent,
        weatherSpecialist:    weatherSpecialist,
        calculatorSpecialist: calculatorSpecialist,
        timeSpecialist:       timeSpecialist,
    }
    
    log.Println("Multi-Agent Hosting系统创建成功")
    return router, nil
}
```

## 🎯 完整工作流程详解

### 📋 详细执行步骤

```
用户输入: "北京天气"
    ↓
[步骤1] 构建系统消息 + 用户消息
    ↓
[步骤2] Host Agent 意图分析
    ├─ 系统提示：分析用户意图，选择合适专家
    ├─ 用户输入：北京天气
    └─ 输出决策：weather
    ↓
[步骤3] MultiAgentRouter 路由分发
    ├─ 解析 Host Agent 决策结果
    ├─ 识别需要调用: Weather Specialist
    └─ 准备专家输入消息
    ↓
[步骤4] Weather Specialist 执行
    ├─ ChatModel: 理解天气查询意图
    ├─ 生成工具调用: get_weather({"city": "北京"})
    ├─ Lambda 执行器: 调用 WeatherTool
    ├─ 工具返回: {"city":"北京","weather":"晴，25°C，微风"}
    └─ ChatModel: 基于工具结果生成回复
    ↓
[步骤5] 结果整合
    ├─ 收集专家响应: "🌤️ 北京今天的天气：晴，25°C，微风。"
    └─ 返回最终结果
    ↓
最终输出: "🌤️ 北京今天的天气：晴，25°C，微风。"
```

### 🔄 多专家协作流程

对于复杂查询如"北京天气和计算10+5"：

```go
// Host Agent 分析结果
hostResponse.Content = "weather,calculator"

// Router 解析并分发
specialists := []string{"weather", "calculator"}

// 并行或序列执行多个专家
for _, specialist := range specialists {
    // Weather Specialist 执行
    weatherResult = "🌤️ 北京今天的天气：晴，25°C，微风。"
    
    // Calculator Specialist 执行  
    calcResult = "🔢 计算结果：10+5 = 15.00"
}

// 结果整合
finalResult = weatherResult + "\n\n" + calcResult
```

### 🔍 调试日志示例

运行时的详细日志输出：

```
2025/09/01 23:23:22 创建Multi-Agent Hosting系统...
2025/09/01 23:23:22 创建Host Agent...
2025/09/01 23:23:22 Host Agent创建成功
2025/09/01 23:23:22 创建Weather Specialist Agent...
2025/09/01 23:23:22 Weather Specialist Agent创建成功
2025/09/01 23:23:22 创建Calculator Specialist Agent...
2025/09/01 23:23:22 Calculator Specialist Agent创建成功
2025/09/01 23:23:22 创建Time Specialist Agent...
2025/09/01 23:23:22 Time Specialist Agent创建成功
2025/09/01 23:23:22 Multi-Agent Hosting系统创建成功

📝 测试用例 1: 天气查询测试
👤 用户: 北京天气
--------------------------------------------------
2025/09/01 23:23:24 [Host Agent] 路由决策: weather_specialist
2025/09/01 23:23:24 [Router] 调用 Weather Specialist
2025/09/01 23:23:28 [WeatherTool] 查询城市: 北京
2025/09/01 23:23:34 [Router] 专家 weather_specialist 返回: 🌤️ 北京今天的天气：晴，25°C，微风。
🤖 系统回复: 🌤️ 北京今天的天气：晴，25°C，微风。
```

## 🔧 自定义扩展

### 📝 添加新的专家代理

#### 1. 实现专业工具

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
            "table": {
                Type:     "string", 
                Desc:     "数据表名",
                Required: false,
            },
        }),
    }, nil
}

// 实现工具执行逻辑
func (d *DatabaseTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    var args struct {
        UserID string `json:"user_id"`
        Table  string `json:"table"`
    }
    
    if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
        return "", fmt.Errorf("参数解析失败: %v", err)
    }
    
    // 模拟数据库查询
    userData := map[string]interface{}{
        "user_id": args.UserID,
        "name":    "张三",
        "email":   "zhangsan@example.com",
        "status":  "active",
        "message": fmt.Sprintf("🗄️ 用户 %s 的信息查询成功", args.UserID),
    }
    
    result, _ := json.Marshal(userData)
    return string(result), nil
}
```

#### 2. 创建专家智能体

```go
func createDatabaseSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
    log.Println("创建Database Specialist Agent...")
    
    // 配置专业工具集
    tools := []tool.InvokableTool{&DatabaseTool{}}
    chatModel, err := createChatModel(ctx, tools)
    if err != nil {
        return nil, fmt.Errorf("创建Database Specialist失败: %v", err)
    }
    
    // 创建处理链
    chain := compose.NewChain[[]*schema.Message, *schema.Message]()
    
    // 意图理解
    chain.AppendChatModel(chatModel, compose.WithNodeName("database_intent_analysis"))
    
    // 工具执行
    chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
        if len(msg.ToolCalls) > 0 {
            var allMessages []*schema.Message
            allMessages = append(allMessages, msg)
            
            for _, toolCall := range msg.ToolCalls {
                var toolResponse string
                var err error
                
                if toolCall.Function.Name == "query_database" {
                    dbTool := &DatabaseTool{}
                    toolResponse, err = dbTool.InvokableRun(ctx, toolCall.Function.Arguments)
                } else {
                    toolResponse = `{"error": "未知工具"}`
                }
                
                toolMessage := &schema.Message{
                    Role:       schema.Tool,
                    Content:    toolResponse,
                    ToolCallID: toolCall.ID,
                }
                allMessages = append(allMessages, toolMessage)
            }
            
            return allMessages, nil
        }
        
        return []*schema.Message{msg}, nil
    }), compose.WithNodeName("database_tool_executor"))
    
    // 响应生成
    chain.AppendChatModel(chatModel, compose.WithNodeName("database_response_generator"))
    
    specialist, err := chain.Compile(ctx)
    if err != nil {
        return nil, fmt.Errorf("编译Database Specialist失败: %v", err)
    }
    
    log.Println("Database Specialist Agent创建成功")
    return specialist, nil
}
```

#### 3. 扩展路由器

```go
type ExtendedMultiAgentRouter struct {
    hostAgent            *ark.ChatModel
    weatherSpecialist    compose.Runnable[[]*schema.Message, *schema.Message]
    calculatorSpecialist compose.Runnable[[]*schema.Message, *schema.Message] 
    timeSpecialist       compose.Runnable[[]*schema.Message, *schema.Message]
    databaseSpecialist   compose.Runnable[[]*schema.Message, *schema.Message] // 新增专家
}

// 在 Invoke 方法中添加新的路由逻辑
func (m *ExtendedMultiAgentRouter) Invoke(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
    // Host Agent 意图分析 - 更新系统提示
    hostPrompt := &schema.Message{
        Role: schema.System,
        Content: `你是智能助手调度中心。分析用户请求，选择合适的专家处理：

可用专家：
- weather: 处理天气查询
- calculator: 处理数学计算  
- time: 处理时间查询
- database: 处理数据库查询和用户信息查询

请只返回需要的专家名称，多个用逗号分隔。`,
    }
    
    // ... Host Agent 处理逻辑
    
    // 专家路由 - 添加新的分支
    for _, specialist := range specialists {
        specialist = strings.TrimSpace(specialist)
        
        if strings.Contains(specialist, "database") {
            log.Println("[Router] 调用 Database Specialist")
            dbPrompt := &schema.Message{
                Role:    schema.System,
                Content: "你是数据库查询专家。用户需要查询数据时，请使用数据库工具获取信息。",
            }
            specialistInput = append([]*schema.Message{dbPrompt}, specialistInput...)
            result, err = m.databaseSpecialist.Invoke(ctx, specialistInput)
            
        } else if strings.Contains(specialist, "weather") {
            // ... 现有逻辑
        }
        // ... 其他专家
    }
    
    // ... 结果整合逻辑
}
```

### 🎨 自定义 Host Agent 策略

#### 高级路由策略

```go
// 基于用户上下文的智能路由
func createSmartHostAgent(ctx context.Context) (*ark.ChatModel, error) {
    hostModel, err := createChatModel(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    return hostModel, nil
}

// 动态系统提示生成
func generateContextualPrompt(userContext map[string]interface{}) string {
    basePrompt := `你是智能助手调度中心。分析用户请求和上下文信息，选择最合适的专家处理任务。`
    
    // 根据用户历史偏好调整提示
    if userContext["preferred_weather_format"] == "detailed" {
        basePrompt += `\n用户偏好详细的天气信息，天气查询时选择 weather 专家。`
    }
    
    if userContext["calculation_precision"] == "high" {
        basePrompt += `\n用户需要高精度计算，数学任务优先使用 calculator 专家。`
    }
    
    return basePrompt
}
```

#### 专家能力注册机制

```go
type SpecialistRegistry struct {
    capabilities map[string]SpecialistInfo
    mu          sync.RWMutex
}

type SpecialistInfo struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Keywords    []string `json:"keywords"`
    Priority    int      `json:"priority"`
    Specialist  compose.Runnable[[]*schema.Message, *schema.Message]
}

func (sr *SpecialistRegistry) Register(info SpecialistInfo) {
    sr.mu.Lock()
    defer sr.mu.Unlock()
    sr.capabilities[info.Name] = info
}

func (sr *SpecialistRegistry) GetByKeyword(keyword string) []SpecialistInfo {
    sr.mu.RLock()
    defer sr.mu.RUnlock()
    
    var matches []SpecialistInfo
    for _, info := range sr.capabilities {
        for _, kw := range info.Keywords {
            if strings.Contains(strings.ToLower(keyword), strings.ToLower(kw)) {
                matches = append(matches, info)
                break
            }
        }
    }
    return matches
}
```

## 📊 优势特性

### 1. **模块化设计**
- 每个专家代理独立开发和维护
- 易于添加新的专业能力
- 代码结构清晰，职责分离

### 2. **智能路由**
- Host Agent 自动分析用户意图
- 支持复杂的多步骤任务分解
- 动态选择最合适的专家处理

### 3. **高可扩展性**
- 轻松添加新的专家代理
- 支持专家代理的热插拔
- 可配置的路由策略

### 4. **强大的协作能力**
- 支持多个专家协同工作
- 任务结果可以在专家间传递
- 统一的输入输出接口

## 🎨 使用场景

### 企业应用场景

1. **客服系统**：
   - Host Agent 分析客户问题
   - 路由到技术支持、销售咨询、售后服务等专家

2. **内容管理系统**：
   - 文档处理专家、图片处理专家、视频处理专家
   - 根据内容类型自动路由

3. **数据分析平台**：
   - 财务分析专家、市场分析专家、运营分析专家
   - 根据分析需求选择对应专家

### 开发应用场景

1. **代码助手**：
   - 前端开发专家、后端开发专家、数据库专家
   - 根据技术栈选择对应专家

2. **测试系统**：
   - 单元测试专家、集成测试专家、性能测试专家
   - 根据测试类型分配任务

## ⚡ 性能优化

### 1. **并行处理**
```go
// 支持多个专家并行处理独立任务
specialists := []string{"weather_specialist", "time_specialist"}
// 可以同时调用多个专家，提高响应速度
```

### 2. **缓存机制**
```go
// 可以为专家代理添加结果缓存
type CachedSpecialist struct {
    specialist compose.Runnable[[]*schema.Message, *schema.Message]
    cache      map[string]*schema.Message
}
```

### 3. **负载均衡**
```go
// 对于相同能力的专家，可以实现负载均衡
specialists := map[string][]compose.Runnable{
    "weather": {weatherSpecialist1, weatherSpecialist2},
}
```

## 🔍 故障排除

### 常见问题

1. **专家路由失败**：
   - 检查Host Agent的系统提示是否清晰
   - 确认专家名称与配置一致

2. **工具调用错误**：
   - 验证工具参数定义正确性
   - 检查工具实现逻辑

3. **多专家协作问题**：
   - 确保消息格式在专家间保持一致
   - 检查任务分解逻辑

### 调试技巧

1. **启用详细日志**：
```go
log.Printf("[Host] 分析意图: %s", userIntent)
log.Printf("[Specialist] 调用专家: %s", specialistName)
```

2. **输出中间状态**：
```go
fmt.Printf("Host分析结果: %+v\n", routingDecision)
fmt.Printf("专家执行结果: %+v\n", specialistResult)
```

## 🔮 高级功能

### 1. **动态专家注册**

```go
type DynamicMultiAgent struct {
    host        *ark.ChatModel
    specialists map[string]compose.Runnable[[]*schema.Message, *schema.Message]
    mu          sync.RWMutex
}

func (d *DynamicMultiAgent) RegisterSpecialist(name string, specialist compose.Runnable[[]*schema.Message, *schema.Message]) {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.specialists[name] = specialist
}
```

### 2. **专家能力描述**

```go
type SpecialistCapability struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Keywords    []string `json:"keywords"`
    Priority    int      `json:"priority"`
}
```

### 3. **上下文传递**

```go
type TaskContext struct {
    UserID      string                 `json:"user_id"`
    SessionID   string                 `json:"session_id"`
    History     []*schema.Message      `json:"history"`
    Metadata    map[string]interface{} `json:"metadata"`
}
```

## 📊 技术架构总结

### 🔗 核心实现架构

```
Multi-Agent Hosting Demo 技术栈
├── Eino Framework (v0.4.7+)
│   ├── compose.NewChain() - 链式编排
│   ├── schema.Message - 消息结构  
│   ├── tool.InvokableTool - 工具接口
│   └── compose.InvokableLambda - 自定义逻辑
├── ARK ChatModel (字节跳动)  
│   ├── 支持工具调用 (BindTools)
│   ├── doubao-seed-1-6-250615
│   └── 意图理解和响应生成
├── MultiAgentRouter (自定义路由器)
│   ├── Host Agent - 意图分析中心
│   └── Specialists - 专业执行团队
└── 三层专家架构
    ├── ChatModel (意图理解)
    ├── Lambda (工具执行)
    └── ChatModel (响应生成)
```

### ⚡ 关键技术要点

1. **分层协作架构**:
   ```
   用户请求 → Host Agent → Router → Specialists → 结果整合
   ```

2. **专家智能体模式**:
   - 每个专家都是独立的 `compose.Runnable` 链
   - 内置工具执行能力 (ChatModel + Lambda + ChatModel)
   - 专业化工具绑定和系统提示

3. **智能路由机制**:
   - Host Agent 基于自然语言进行意图识别
   - 支持多专家协作和任务分解
   - 动态专家选择和负载分发

4. **消息流处理**:
   - 标准化的 `schema.Message` 结构
   - 工具调用和响应的完整生命周期管理
   - 上下文保持和状态传递

### 🎯 适用场景

- ✅ **企业级客服系统**: 意图路由 + 专业服务团队
- ✅ **智能运维平台**: 故障分析 + 专业处理模块
- ✅ **教育辅导系统**: 学科识别 + 专业教师智能体
- ✅ **电商助手系统**: 需求分析 + 专业服务顾问

## 🔍 故障排除

### ⚠️ 常见问题及解决方案

#### 1. **模型配置错误**
```
Error: failed to create chat completion: Error code: 404
```
**解决方案**: 检查 `config.yaml` 中的模型名称是否正确
```yaml
ARK_MODEL: "doubao-seed-1-6-250615"  # 确保模型名称准确
```

#### 2. **专家路由失败** 
```
[Host Agent] 路由决策: unknown_specialist
[Router] 未知专家: unknown_specialist
```
**解决方案**: 优化 Host Agent 的系统提示，提供更清晰的专家描述

#### 3. **工具调用失败**
```
工具执行失败: 参数解析失败
```
**解决方案**: 检查工具参数定义和 JSON 解析逻辑

### 🔧 性能优化建议

1. **并行专家调用**: 对于独立任务，可以并行调用多个专家
2. **结果缓存**: 为频繁查询的结果添加缓存机制  
3. **连接池**: 复用 ARK 模型连接，减少创建开销
4. **超时控制**: 为长时间运行的工具添加超时机制

## 📚 相关文档

- [Eino 官方文档](https://www.cloudwego.io/zh/docs/eino/)
- [Multi-Agent Hosting 详细指南](https://www.cloudwego.io/zh/docs/eino/core_modules/flow_integration_components/multi_agent_hosting/)
- [Compose API 参考](https://www.cloudwego.io/zh/docs/eino/core_modules/compose/)
- [ARK 模型配置](https://www.volcengine.com/docs/82379/1099475)
- [工具系统文档](https://www.cloudwego.io/zh/docs/eino/core_modules/tool/)

## 📈 下一步扩展

1. **流式处理**: 实现实时的专家响应流
2. **持久化存储**: 添加对话历史和上下文管理
3. **负载均衡**: 支持多实例专家的负载分配
4. **监控告警**: 完善的专家性能和错误监控
5. **动态扩容**: 运行时动态添加和移除专家

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个演示项目！

### 贡献方式
- 🐛 报告 Bug 和问题
- 💡 提出新功能建议
- 🔧 提交代码改进
- 📖 完善文档和示例

### 开发规范
- 遵循 Go 官方编码规范
- 添加充分的注释和文档
- 包含必要的测试用例
- 保持代码的可读性和可维护性

## 📄 许可证

本项目遵循 MIT 许可证。