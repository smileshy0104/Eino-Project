# 🛠️ Eino ToolsNode 组件完全指南

本文档是对 Eino 框架中 `ToolsNode` 组件的核心功能和使用方式的完整总结，结合官方文档和实际项目示例。

## 🚀 快速开始

### 🛠️ 配置文件
项目使用 `config.yaml` 配置文件，也可以通过环境变量设置：
```yaml
ARK_API_KEY: "${ARK_API_KEY}"
LLM_MODEL: "doubao-seed-1-6-250615"
WEATHER_API_KEY: "${WEATHER_API_KEY}"
SEARCH_API_KEY: "${SEARCH_API_KEY}"
```

---

## 📖 基本介绍

`ToolsNode` 组件是一个专门用于**扩展模型能力**的智能组件。它的主要作用是允许大语言模型调用外部工具来完成特定任务，从而突破模型自身的知识和能力限制。这个组件在 AI 应用开发中扮演着**"能力扩展器"**的角色。

### 🎯 核心价值

在传统的 LLM 应用中，模型只能基于训练数据回答问题。而 ToolsNode 组件让我们能够：

```
传统模型：静态知识 + 有限能力 + 无法执行操作  ❌
ToolsNode：动态信息 + 能力扩展 + 实际操作执行  ✅
```

### 🚀 主要应用场景

- **🌐 实时信息获取**: 获取当前天气、新闻、股价等实时数据
- **🔍 智能搜索**: 调用搜索引擎获取最新信息和知识
- **📊 数据处理**: 执行计算、数据分析、格式转换等操作
- **🔗 系统集成**: 与数据库、API、第三方服务进行集成
- **🤖 智能助手**: 构建能够执行实际任务的AI助手
- **🧩 工作流增强**: 在复杂的AI工作流中提供关键能力扩展

---

## 🔧 核心接口

`ToolsNode` 组件提供了分层设计的接口架构：

### 基础接口层次

```go
// 基础工具接口
type BaseTool interface {
    Info() *ToolInfo
}

// 同步调用工具接口
type InvokableTool interface {
    BaseTool
    InvokableRun(ctx context.Context, input string, opts ...Option) (string, error)
}

// 流式调用工具接口
type StreamableTool interface {
    BaseTool
    StreamableRun(ctx context.Context, input string, opts ...Option) (*schema.StreamReader[string], error)
}
```

### ToolsNode 配置接口

```go
type ToolsNodeConfig struct {
    Tools []tool.BaseTool  // 工具列表
    Model llamaindex.LLM   // 使用的大语言模型
}

// 创建 ToolsNode
func NewToolsNode(ctx context.Context, config *ToolsNodeConfig, opts ...Option) (*ToolsNode, error)
```

### 接口详解

#### 🔧 BaseTool
- **功能**: 提供工具的基础信息描述
- **方法**: `Info()` 返回工具的元数据信息
- **用途**: 所有工具的基础接口，定义工具身份

#### ⚡ InvokableTool
- **功能**: 支持同步调用的工具
- **方法**: `InvokableRun()` 执行工具并返回完整结果
- **适用**: 快速响应、短时间执行的工具

#### 🌊 StreamableTool
- **功能**: 支持流式输出的工具
- **方法**: `StreamableRun()` 执行工具并流式返回结果
- **适用**: 长时间执行、大量数据输出的工具

---

## 📋 ToolInfo 结构体

`ToolInfo` 是工具描述的核心结构，包含工具的完整元数据：

```go
type ToolInfo struct {
    // Name 是工具的唯一标识符
    Name string
    // Description 是工具功能的详细说明
    Description string
    // ParamsOneOf 定义工具的参数规范
    ParamsOneOf interface{}
}
```

### 🎭 字段说明

- **🏷️ Name**: 工具的唯一标识符，用于调用时的识别
- **📝 Description**: 工具功能的详细说明，帮助模型理解何时使用
- **⚙️ ParamsOneOf**: 参数定义，支持多种格式：
  - `map[string]*ParameterInfo` - 简单参数定义
  - Go struct with tags - 结构化参数定义
  - OpenAPI3 Schema - 标准化参数定义

### 参数定义方式

#### 1. 使用 map[string]*ParameterInfo
```go
ParamsOneOf: map[string]*ParameterInfo{
    "query": {
        Type:        "string",
        Required:    true,
        Description: "搜索查询字符串",
    },
    "limit": {
        Type:        "integer",
        Required:    false,
        Description: "结果数量限制",
        Default:     10,
    },
}
```

#### 2. 使用 Go 结构体
```go
type SearchParams struct {
    Query string `json:"query" jsonschema:"description=搜索查询字符串,required"`
    Limit int    `json:"limit,omitempty" jsonschema:"description=结果数量限制"`
}

ParamsOneOf: SearchParams{}
```

#### 3. 使用 OpenAPI3 Schema
```go
ParamsOneOf: &openapi3.Schema{
    Type: "object",
    Properties: map[string]*openapi3.SchemaRef{
        "query": {
            Value: &openapi3.Schema{
                Type:        "string",
                Description: "搜索查询字符串",
            },
        },
    },
    Required: []string{"query"},
}
```

---

## 🏗️ Tool 创建方式

### 1. 🎯 自动推断创建（推荐）

使用 `utils.InferTool()` 从函数自动生成工具：

```go
// 定义工具函数
func SearchWeb(ctx context.Context, query string, limit int) (string, error) {
    // 搜索逻辑实现
    results := performSearch(query, limit)
    return formatResults(results), nil
}

// 自动创建工具
searchTool, err := utils.InferTool(
    "search_web",                    // 工具名称
    "在网络上搜索信息",               // 工具描述
    SearchWeb,                      // 工具函数
)
```

### 2. 🔧 手动创建工具

```go
// 实现 InvokableTool 接口
type WeatherTool struct {
    apiKey string
}

func (w *WeatherTool) Info() *tool.ToolInfo {
    return &tool.ToolInfo{
        Name:        "get_weather",
        Description: "获取指定城市的当前天气信息",
        ParamsOneOf: map[string]*tool.ParameterInfo{
            "city": {
                Type:        "string",
                Required:    true,
                Description: "城市名称",
            },
        },
    }
}

func (w *WeatherTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
    // 解析输入参数
    params := parseInput(input)
    city := params["city"]

    // 调用天气API
    weather := w.getWeatherData(city)
    return formatWeatherInfo(weather), nil
}
```

### 3. 🌊 流式工具创建

```go
// 实现 StreamableTool 接口
type DataProcessorTool struct {
    processor *DataProcessor
}

func (d *DataProcessorTool) Info() *tool.ToolInfo {
    return &tool.ToolInfo{
        Name:        "process_data",
        Description: "流式处理大量数据",
        ParamsOneOf: ProcessParams{},
    }
}

func (d *DataProcessorTool) StreamableRun(ctx context.Context, input string, opts ...tool.Option) (*schema.StreamReader[string], error) {
    // 创建流式处理器
    stream := d.processor.ProcessStream(input)
    return schema.NewStreamReader(stream), nil
}
```

---

## ⚙️ 配置选项与最佳实践

### Option 配置

```go
// 工具特定选项
type ToolSpecificOptions struct {
    Timeout     time.Duration
    RetryCount  int
    UseCache    bool
}

// 包装工具选项
toolOption := tool.WrapImplSpecificOptFn(func(o *ToolSpecificOptions) {
    o.Timeout = 30 * time.Second
    o.RetryCount = 3
    o.UseCache = true
})
```

### Callback 机制

```go
// 工具执行回调
toolCallback := func(ctx context.Context, req *tool.CallRequest, resp *tool.CallResponse) {
    log.Printf("工具调用: %s, 输入: %s, 输出: %s",
        req.ToolName, req.Input, resp.Output)
}

// 在 ToolsNode 中使用回调
toolsNode, err := compose.NewToolsNode(ctx, &compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{searchTool, weatherTool},
    Model: llmModel,
}, compose.WithToolsNodeCallback(toolCallback))
```

---

## 🏭 ToolsNode 集成模式

### 1. 🔗 Chain 集成

```go
// 创建工具节点
toolsNode, err := compose.NewToolsNode(ctx, &compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{
        searchTool,
        weatherTool,
        calculatorTool,
    },
    Model: llmModel,
})

// 集成到 Chain 中
chain := compose.NewChain[string, string]()
chain.AppendChatTemplate(chatTemplate)
chain.AppendToolsNode(toolsNode)
chain.AppendChatTemplate(finalTemplate)
```

### 2. 🕸️ Graph 集成

```go
// 创建 Graph
graph := compose.NewGraph[string, string]()

// 添加工具节点
graph.AddToolsNode("tools", toolsNode)
graph.AddChatTemplateNode("chat", chatTemplate)

// 定义执行流程
graph.AddEdge(compose.START, "chat")
graph.AddEdge("chat", "tools")
graph.AddEdge("tools", compose.END)
```

### 3. 🧬 复杂工作流集成

```go
// 多步骤工作流
graph := compose.NewGraph[string, string]()

// 添加多个专用工具节点
graph.AddToolsNode("search_tools", searchToolsNode)
graph.AddToolsNode("analysis_tools", analysisToolsNode)
graph.AddToolsNode("output_tools", outputToolsNode)

// 定义复杂的执行路径
graph.AddEdge(compose.START, "search_tools")
graph.AddEdge("search_tools", "analysis_tools")
graph.AddEdge("analysis_tools", "output_tools")
graph.AddEdge("output_tools", compose.END)
```

---

## 📊 常用工具类型示例

### 1. 🌐 网络搜索工具

```go
type WebSearchTool struct {
    apiKey string
    engine string
}

func (w *WebSearchTool) Info() *tool.ToolInfo {
    return &tool.ToolInfo{
        Name:        "web_search",
        Description: "在互联网上搜索最新信息",
        ParamsOneOf: map[string]*tool.ParameterInfo{
            "query": {
                Type:        "string",
                Required:    true,
                Description: "搜索查询内容",
            },
            "num_results": {
                Type:        "integer",
                Required:    false,
                Description: "返回结果数量",
                Default:     5,
            },
        },
    }
}
```

### 2. 🌤️ 天气查询工具

```go
type WeatherTool struct {
    apiKey string
}

func (w *WeatherTool) Info() *tool.ToolInfo {
    return &tool.ToolInfo{
        Name:        "get_weather",
        Description: "获取指定地点的实时天气信息",
        ParamsOneOf: map[string]*tool.ParameterInfo{
            "location": {
                Type:        "string",
                Required:    true,
                Description: "城市名称或地理位置",
            },
            "units": {
                Type:        "string",
                Required:    false,
                Description: "温度单位：celsius 或 fahrenheit",
                Default:     "celsius",
                Enum:        []string{"celsius", "fahrenheit"},
            },
        },
    }
}
```

### 3. 🧮 计算器工具

```go
func CalculatorFunction(ctx context.Context, expression string) (string, error) {
    // 数学表达式计算逻辑
    result, err := evaluateExpression(expression)
    if err != nil {
        return "", fmt.Errorf("计算错误: %w", err)
    }
    return fmt.Sprintf("计算结果: %v", result), nil
}

// 使用自动推断创建
calculatorTool, err := utils.InferTool(
    "calculator",
    "执行数学计算和表达式求值",
    CalculatorFunction,
)
```

### 4. 📊 数据分析工具

```go
type DataAnalysisTool struct {
    analyzer *DataAnalyzer
}

func (d *DataAnalysisTool) Info() *tool.ToolInfo {
    return &tool.ToolInfo{
        Name:        "analyze_data",
        Description: "分析数据集并生成统计报告",
        ParamsOneOf: map[string]*tool.ParameterInfo{
            "data": {
                Type:        "string",
                Required:    true,
                Description: "要分析的数据（JSON格式）",
            },
            "analysis_type": {
                Type:        "string",
                Required:    false,
                Description: "分析类型",
                Enum:        []string{"basic", "advanced", "statistical"},
                Default:     "basic",
            },
        },
    }
}
```

---

## 🎯 错误处理与调试

### 错误处理最佳实践

```go
func (t *MyTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
    // 1. 输入验证
    params, err := t.parseAndValidateInput(input)
    if err != nil {
        return "", fmt.Errorf("输入参数无效: %w", err)
    }

    // 2. 超时控制
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // 3. 错误分类处理
    result, err := t.executeLogic(ctx, params)
    if err != nil {
        switch {
        case errors.Is(err, context.DeadlineExceeded):
            return "", fmt.Errorf("工具执行超时: %w", err)
        case isRetryableError(err):
            return "", fmt.Errorf("可重试错误: %w", err)
        default:
            return "", fmt.Errorf("工具执行失败: %w", err)
        }
    }

    return result, nil
}
```

### 调试和监控

```go
// 工具执行监控
type ToolMonitor struct {
    metrics *Metrics
    logger  *Logger
}

func (m *ToolMonitor) WrapTool(tool tool.InvokableTool) tool.InvokableTool {
    return &MonitoredTool{
        original: tool,
        monitor:  m,
    }
}

type MonitoredTool struct {
    original tool.InvokableTool
    monitor  *ToolMonitor
}

func (m *MonitoredTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
    start := time.Now()
    toolName := m.original.Info().Name

    // 记录开始
    m.monitor.logger.Info("工具开始执行", "tool", toolName, "input", input)

    // 执行工具
    result, err := m.original.InvokableRun(ctx, input, opts...)

    // 记录结果
    duration := time.Since(start)
    if err != nil {
        m.monitor.logger.Error("工具执行失败", "tool", toolName, "error", err, "duration", duration)
        m.monitor.metrics.RecordToolError(toolName, err)
    } else {
        m.monitor.logger.Info("工具执行成功", "tool", toolName, "duration", duration)
        m.monitor.metrics.RecordToolSuccess(toolName, duration)
    }

    return result, err
}
```

---

## 🔄 高级用法和模式

### 1. 🎭 工具组合模式

```go
// 创建专门的工具组
type ToolGroup struct {
    name  string
    tools []tool.BaseTool
}

// 搜索工具组
searchGroup := &ToolGroup{
    name: "search_tools",
    tools: []tool.BaseTool{
        webSearchTool,
        imageSearchTool,
        newsSearchTool,
    },
}

// 分析工具组
analysisGroup := &ToolGroup{
    name: "analysis_tools",
    tools: []tool.BaseTool{
        dataAnalysisTool,
        textAnalysisTool,
        sentimentAnalysisTool,
    },
}
```

### 2. 🔄 动态工具加载

```go
type DynamicToolsNode struct {
    baseTools    []tool.BaseTool
    toolRegistry map[string]tool.BaseTool
    mutex        sync.RWMutex
}

func (d *DynamicToolsNode) RegisterTool(tool tool.BaseTool) {
    d.mutex.Lock()
    defer d.mutex.Unlock()
    d.toolRegistry[tool.Info().Name] = tool
}

func (d *DynamicToolsNode) GetAvailableTools() []tool.BaseTool {
    d.mutex.RLock()
    defer d.mutex.RUnlock()

    var allTools []tool.BaseTool
    allTools = append(allTools, d.baseTools...)
    for _, tool := range d.toolRegistry {
        allTools = append(allTools, tool)
    }
    return allTools
}
```

### 3. 🎯 条件工具选择

```go
type ConditionalToolsNode struct {
    toolGroups map[string][]tool.BaseTool
    selector   func(context.Context, string) string
}

func (c *ConditionalToolsNode) SelectTools(ctx context.Context, input string) []tool.BaseTool {
    groupName := c.selector(ctx, input)
    return c.toolGroups[groupName]
}

// 使用示例
selector := func(ctx context.Context, input string) string {
    if strings.Contains(input, "天气") {
        return "weather_tools"
    } else if strings.Contains(input, "搜索") {
        return "search_tools"
    }
    return "general_tools"
}
```

---

## 💡 最佳实践总结

### 1. 工具设计原则

```go
// ✅ 好的实践：清晰的工具定义
tool := &MyTool{
    name:        "clear_descriptive_name",
    description: "详细说明工具的功能和使用场景",
    params:      wellDefinedParameters,
}

// ❌ 避免：模糊的工具定义
tool := &MyTool{
    name:        "tool1",
    description: "does something",
    params:      nil,
}
```

### 2. 参数验证模式

```go
// ✅ 完善的参数验证
func (t *MyTool) validateInput(input string) (*Params, error) {
    params := &Params{}
    if err := json.Unmarshal([]byte(input), params); err != nil {
        return nil, fmt.Errorf("参数解析失败: %w", err)
    }

    if params.Query == "" {
        return nil, errors.New("query 参数是必需的")
    }

    if params.Limit < 1 || params.Limit > 100 {
        return nil, errors.New("limit 参数必须在 1-100 之间")
    }

    return params, nil
}
```

### 3. 资源管理

```go
// ✅ 正确的资源管理
func (t *MyTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
    // 创建资源
    client := t.createClient()
    defer client.Close()  // 确保资源清理

    // 使用超时控制
    ctx, cancel := context.WithTimeout(ctx, t.timeout)
    defer cancel()

    // 执行逻辑
    return t.executeWithClient(ctx, client, input)
}
```

### 4. 错误处理策略

```go
// ✅ 分层错误处理
func (t *MyTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
    // 验证层
    params, err := t.validateInput(input)
    if err != nil {
        return "", fmt.Errorf("参数验证失败: %w", err)
    }

    // 执行层
    result, err := t.executeLogic(ctx, params)
    if err != nil {
        // 根据错误类型返回不同信息
        if isUserError(err) {
            return "", fmt.Errorf("用户输入错误: %w", err)
        }
        return "", fmt.Errorf("系统执行错误: %w", err)
    }

    return result, nil
}
```

---

## 🎉 总结

ToolsNode 是 Eino 框架中的**核心能力扩展组件**，掌握它的使用对于构建高质量的智能应用至关重要：

### 🏆 核心优势

- 🛠️ **能力扩展**: 让 LLM 具备执行实际任务的能力
- ⚡ **灵活集成**: 支持同步和异步两种调用模式
- 🧩 **模块化设计**: 工具可以独立开发、测试和部署
- 🔄 **动态配置**: 支持运行时工具注册和配置
- 📊 **完善监控**: 提供详细的执行监控和错误处理

### 💡 设计理念

1. **简单性**: 工具定义简洁明了，易于理解和使用
2. **可扩展性**: 支持各种类型工具的无缝集成
3. **可靠性**: 完善的错误处理和资源管理机制
4. **性能**: 支持并发执行和流式处理
5. **可观测性**: 提供全面的监控和调试支持

### 🚀 应用前景

- **🤖 智能助手**: 构建能够执行实际任务的AI助手
- **📊 数据分析**: 结合实时数据进行智能分析
- **🔗 系统集成**: 连接各种外部系统和服务
- **🌐 信息聚合**: 从多个来源获取和整合信息
- **⚡ 自动化**: 实现复杂业务流程的自动化

通过掌握 ToolsNode 组件的各种功能，你将能够构建出更加智能、实用和强大的AI应用系统！🚀