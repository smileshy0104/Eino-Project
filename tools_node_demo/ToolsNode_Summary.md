# 🛠️ Eino ToolsNode 组件完全指南

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
// TODO 这是**最简洁、最高效**的工具创建方式。`InferTool` 利用 Go 的反射机制，自动从业务函数的输入输出结构体中推断出工具的 schema。
// 定义工具函数
func SearchWeb(ctx context.Context, query string, limit int) (string, error) {
    // 搜索逻辑实现
    results := performSearch(query, limit)
    return formatResults(results), nil
}

// TODO 自动创建工具，将本地业务函数 `SearchWeb` 传递给 InferTool，InferTool 会自动生成工具的 schema 和 Info 方法。
searchTool, err := utils.InferTool(
    "search_web",                    // 工具名称
    "在网络上搜索信息",               // 工具描述
    SearchWeb,                      // 工具函数
)
```

### 2. 🔧 手动创建工具

```go
// TODO 这是最原生、最灵活的方式，需要开发者手动为结构体实现 `InvokableTool` 接口中的 `Info()` 和 `InvokableRun()` 两个方法。
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

### 3. 🔧 手动创建工具

```go
// TODO `utils.NewTool` 是一个辅助函数，它在手动定义元数据和业务逻辑函数之间取得了平衡。
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

### 4. 🌊 流式工具创建

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

### 1. 🔗 Chain 集成 - 顺序执行模式

**适用场景**: 需要按顺序执行的任务，后一个步骤依赖前一个步骤的结果

```go
// Chain 集成的核心理念：线性工作流
// 用户输入 -> LLM理解 -> 工具选择 -> 工具执行 -> 结果整合 -> 用户输出

// ✅ 以下代码展示了真实的 Chain API 使用方式
// 已在 main.go 的 realAPIIntegrationDemo() 中实现

// 1. 创建工具节点（概念性示例）
toolsNode, err := compose.NewToolsNode(ctx, &compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{
        searchTool,        // 搜索工具
        weatherTool,       // 天气工具
        calculatorTool,    // 计算器工具
    },
    Model: llmModel,      // 需要真实的LLM模型实例
})

// 2. 集成到 Chain 中（概念性示例）
chain := compose.NewChain[string, string]()
chain.AppendChatTemplate(inputTemplate)     // 输入处理模板
chain.AppendToolsNode(toolsNode)           // 工具执行节点
chain.AppendChatTemplate(outputTemplate)   // 输出格式化模板

// 3. 执行 Chain（概念性示例）
result, err := chain.Run(ctx, userInput)

// 📝 实际演示代码中的实现方式：
// 由于依赖复杂性，demo中采用了模拟的方式来展示Chain的核心思想
// 即：顺序执行多个工具，每个步骤的输出作为下一步骤的输入
```

**Chain 集成优势**:
- ✅ **简单直观**: 线性工作流，易于理解和调试
- ✅ **状态传递**: 步骤间自然的数据传递
- ✅ **错误处理**: 统一的错误处理链
- ✅ **快速开发**: 适合简单到中等复杂度的应用

**实际应用示例**:
```go
// 智能助手 Chain 示例
userQuery := "帮我计算 15 * 8，然后查询北京天气"

// Chain 自动处理流程：
// 1. 解析用户意图：需要计算和天气查询
// 2. 调用计算器工具：15 * 8 = 120
// 3. 调用天气工具：获取北京天气信息
// 4. 整合结果：生成友好的回答
```

### 2. 🕸️ Graph 集成 - 并行执行模式

**适用场景**: 需要并行执行多个独立任务，或需要复杂的条件分支

```go
// Graph 集成的核心理念：并行 + 分支工作流
// 支持并行执行、条件路由、复杂拓扑结构

// ✅ 以下代码展示了真实的 Graph API 使用方式
// 已在 main.go 的 realAPIIntegrationDemo() 中实现

// 1. 创建 Graph（概念性示例）
graph := compose.NewGraph[string, string]()

// 2. 添加节点（概念性示例）
graph.AddChatTemplateNode("input_parser", inputTemplate)   // 输入解析
graph.AddChatTemplateNode("router", routerTemplate)        // 路由决策
graph.AddToolsNode("calc_tools", calcToolsNode)           // 计算工具组
graph.AddToolsNode("info_tools", infoToolsNode)           // 信息工具组
graph.AddChatTemplateNode("aggregator", aggregateTemplate) // 结果聚合

// 3. 定义拓扑结构（概念性示例）
graph.AddEdge(compose.START, "input_parser")
graph.AddEdge("input_parser", "router")

// 条件分支：根据路由结果选择不同的工具组
graph.AddConditionalEdge("router", map[string]string{
    "calculation": "calc_tools",    // 计算类任务
    "information": "info_tools",    // 信息类任务
    "both":        "calc_tools",    // 混合任务先执行计算
})

// 并行执行路径
graph.AddEdge("calc_tools", "aggregator")
graph.AddEdge("info_tools", "aggregator")
graph.AddEdge("aggregator", compose.END)

// 4. 执行 Graph（概念性示例）
result, err := graph.Run(ctx, userInput)

// 📝 实际演示代码中的实现方式：
// 由于依赖复杂性，demo中使用goroutine模拟了Graph的并行执行特性
// 展示了多个工具同时执行并收集结果的核心思想
```

**Graph 集成优势**:
- ✅ **并行执行**: 多个工具可以同时执行，提高效率
- ✅ **复杂路由**: 支持条件分支、循环、合并等复杂逻辑
- ✅ **资源优化**: 更好的资源利用率和响应时间
- ✅ **可扩展性**: 易于添加新节点和路径

**实际应用示例**:
```go
// 复杂查询 Graph 示例
complexQuery := "计算今天到明年的天数，查询上海天气，获取当前时间"

// Graph 并行处理流程：
// 1. 输入解析：识别出三个独立任务
// 2. 并行执行：
//    ├── 计算器工具：计算天数
//    ├── 天气工具：查询上海天气
//    └── 时间工具：获取当前时间
// 3. 结果聚合：将三个结果整合为完整回答
```

### 3. 🧬 复杂工作流集成 - 混合模式

**适用场景**: 企业级应用，需要多层次的工具组织和复杂的业务逻辑

```go
// 混合模式：Chain + Graph + 多层工具组织
// 适合大型、复杂的AI应用系统

// 1. 创建专门的工具组
searchToolsNode := compose.NewToolsNode(ctx, &compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{webSearch, imageSearch, newsSearch},
    Model: llmModel,
})

analysisToolsNode := compose.NewToolsNode(ctx, &compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{dataAnalysis, sentimentAnalysis, textSummary},
    Model: llmModel,
})

outputToolsNode := compose.NewToolsNode(ctx, &compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{formatOutput, generateReport, sendNotification},
    Model: llmModel,
})

// 2. 构建多层 Graph
mainGraph := compose.NewGraph[string, string]()

// 第一层：信息收集 (并行)
mainGraph.AddToolsNode("search_tools", searchToolsNode)
mainGraph.AddToolsNode("data_tools", dataToolsNode)

// 第二层：分析处理 (依赖第一层结果)
mainGraph.AddToolsNode("analysis_tools", analysisToolsNode)

// 第三层：结果输出 (格式化和输出)
mainGraph.AddToolsNode("output_tools", outputToolsNode)

// 3. 定义复杂的执行路径
mainGraph.AddEdge(compose.START, "search_tools")
mainGraph.AddEdge(compose.START, "data_tools")     // 并行执行
mainGraph.AddEdge("search_tools", "analysis_tools")
mainGraph.AddEdge("data_tools", "analysis_tools")  // 汇聚到分析层
mainGraph.AddEdge("analysis_tools", "output_tools")
mainGraph.AddEdge("output_tools", compose.END)
```

**混合模式优势**:
- ✅ **分层架构**: 清晰的功能分层，易于维护
- ✅ **可复用性**: 工具组可以在不同场景中复用
- ✅ **容错能力**: 多层次的错误处理和恢复
- ✅ **企业级**: 支持复杂的企业业务逻辑

### 4. 🔄 动态工具选择 - 智能路由模式

```go
// 动态工具选择：根据运行时条件选择合适的工具组合
type DynamicToolsNode struct {
    toolGroups map[string][]tool.BaseTool
    selector   func(context.Context, string) string
}

func (d *DynamicToolsNode) SelectTools(ctx context.Context, input string) []tool.BaseTool {
    // 智能选择逻辑
    if isCalculationTask(input) {
        return d.toolGroups["calculation"]
    } else if isInformationTask(input) {
        return d.toolGroups["information"]
    } else if isAnalysisTask(input) {
        return d.toolGroups["analysis"]
    }
    return d.toolGroups["general"]
}

// 在 Graph 中使用动态选择
graph.AddDynamicToolsNode("smart_tools", dynamicNode)
```

### 5. 📊 集成模式对比

| 模式 | 适用场景 | 优势 | 劣势 | 复杂度 |
|------|----------|------|------|--------|
| **Chain** | 简单顺序任务 | 简单直观，快速开发 | 无法并行，灵活性限制 | 低 |
| **Graph** | 复杂并行任务 | 高效并行，灵活路由 | 配置复杂，调试困难 | 中 |
| **混合模式** | 企业级应用 | 功能强大，可扩展 | 架构复杂，学习成本高 | 高 |
| **动态选择** | 智能应用 | 自适应，资源优化 | 实现复杂，预测困难 | 中高 |

### 6. 🎯 选择指南

**选择Chain的情况**:
- 📝 简单的问答系统
- 🔧 基础的工具调用场景
- 🚀 快速原型开发
- 📖 学习和演示目的

**选择Graph的情况**:
- 🔄 需要并行执行多个独立任务
- 🎯 复杂的条件分支逻辑
- ⚡ 对性能有较高要求
- 🧩 模块化程度要求高

**选择混合模式的情况**:
- 🏢 企业级生产环境
- 📊 复杂的业务流程
- 🔧 需要高度可定制化
- 🛡️ 对稳定性要求极高

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