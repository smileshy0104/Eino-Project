# Eino Multi-Agent Hosting Demo

## 🏢 项目简介

这是一个基于 Eino 框架的多代理托管（Multi-Agent Hosting）演示项目，展示了如何构建一个具有意图识别和任务路由功能的多代理系统。该架构通过 Host Agent 进行意图分析，然后将请求路由到专门的 Specialist Agents 来处理特定任务。

## 🏗️ 系统架构

### 核心概念

多代理托管是一种企业级的代理架构模式，其中：

- **Host Agent（主机代理）**：作为系统的"大脑"，负责理解用户意图并做出路由决策
- **Specialist Agents（专家代理）**：专注于特定领域的任务处理，具有专门的工具和能力

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

### Host Agent 创建

```go
func createHostAgent(ctx context.Context) (*ark.ChatModel, error) {
    // Host Agent 不需要工具，只负责分析用户意图
    hostModel, err := createChatModel(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("创建Host Agent失败: %v", err)
    }
    
    return hostModel, nil
}
```

### Specialist Agent 创建

```go
func createWeatherSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
    tools := []tool.InvokableTool{&WeatherTool{}}
    chatModel, err := createChatModel(ctx, tools)
    if err != nil {
        return nil, fmt.Errorf("创建Weather Specialist失败: %v", err)
    }
    
    // 创建专门的链
    chain := compose.NewChain[[]*schema.Message, *schema.Message]()
    chain.AppendChatModel(chatModel, compose.WithNodeName("weather_specialist"))
    
    specialist, err := chain.Compile(ctx)
    return specialist, err
}
```

### Multi-Agent 系统配置

```go
multiAgent, err := host.NewMultiAgent(ctx, &host.MultiAgentConfig{
    Host: hostAgent,
    Specialists: map[string]compose.Runnable[[]*schema.Message, *schema.Message]{
        "weather_specialist":    weatherSpecialist,
        "calculator_specialist": calculatorSpecialist,
        "time_specialist":       timeSpecialist,
    },
})
```

## 🎯 工作流程详解

### 1. 用户输入阶段
```
用户: "查询北京天气，然后计算 10+20"
```

### 2. Host Agent 分析阶段
```
Host Agent 分析:
- 识别到两个任务：天气查询 + 数学计算
- 决策：先调用 weather_specialist，再调用 calculator_specialist
```

### 3. Specialist 执行阶段
```
Weather Specialist: 
- 调用 WeatherTool
- 返回: "🌤️ 北京今天的天气：晴，25°C，微风"

Calculator Specialist:
- 调用 CalculatorTool  
- 返回: "🔢 计算结果：10+20 = 30.00"
```

### 4. 结果整合阶段
```
系统最终回复: 
"北京今天天气晴朗，温度25°C，微风。
计算结果：10+20 = 30"
```

## 🔧 自定义扩展

### 添加新的专家代理

1. **创建专门的工具**：
```go
type NewTool struct{}

func (n *NewTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "new_tool",
        Desc: "新工具描述",
        // 参数定义...
    }, nil
}

func (n *NewTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    // 工具实现逻辑
    return result, nil
}
```

2. **创建专家代理**：
```go
func createNewSpecialist(ctx context.Context) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
    tools := []tool.InvokableTool{&NewTool{}}
    chatModel, err := createChatModel(ctx, tools)
    if err != nil {
        return nil, err
    }
    
    chain := compose.NewChain[[]*schema.Message, *schema.Message]()
    chain.AppendChatModel(chatModel, compose.WithNodeName("new_specialist"))
    
    return chain.Compile(ctx)
}
```

3. **注册到多代理系统**：
```go
specialists["new_specialist"] = newSpecialist
```

### 自定义Host Agent逻辑

可以通过修改Host Agent的系统提示来改变路由策略：

```go
systemMessage := &schema.Message{
    Role: schema.System,
    Content: `你是智能助手调度中心。根据用户意图选择专家：
    
1. weather_specialist - 天气查询
2. calculator_specialist - 数学计算  
3. time_specialist - 时间查询
4. new_specialist - 新功能处理

请分析用户需求并选择合适的专家处理。`,
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

## 📚 相关文档

- [Eino 官方文档](https://www.cloudwego.io/zh/docs/eino/)
- [Multi-Agent Hosting 详细指南](https://www.cloudwego.io/zh/docs/eino/core_modules/flow_integration_components/multi_agent_hosting/)
- [Host Agent 配置](https://www.cloudwego.io/zh/docs/eino/core_modules/host/)

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个演示项目！

### 开发规范

1. **代码风格**：遵循 Go 官方编码规范
2. **注释要求**：核心函数必须有详细注释
3. **测试覆盖**：新功能需要包含测试用例
4. **文档更新**：功能变更需要同步更新文档

## 📄 许可证

本项目遵循 MIT 许可证。