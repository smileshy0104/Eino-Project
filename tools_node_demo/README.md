# 🛠️ Eino ToolsNode 组件完全指南

本文档是对 Eino 框架中 `ToolsNode` 组件的核心功能和使用方式的完整总结，结合官方文档和实际项目示例。

## 🚀 快速开始

### 配置环境
```bash
# 设置 API Key
export ARK_API_KEY="your-ark-api-key"
export WEATHER_API_KEY="your-weather-api-key"  # 可选
export SEARCH_API_KEY="your-search-api-key"    # 可选

# 构建项目
go build -o tools_node_demo main.go
```

### 运行示例
```bash
# 运行所有示例
./tools_node_demo

# 运行特定示例
./tools_node_demo basic         # 基础工具创建演示
./tools_node_demo manual        # 手动工具创建演示
./tools_node_demo config        # ToolsNode配置演示
./tools_node_demo chain         # 工具调用链演示
./tools_node_demo error         # 错误处理演示
./tools_node_demo performance   # 性能测试演示
```

### 配置文件
项目使用 `config.yaml` 配置文件，也可以通过环境变量设置：
```yaml
ARK_API_KEY: "${ARK_API_KEY}"
LLM_MODEL: "doubao-seed-1-6-250615"
WEATHER_API_KEY: "${WEATHER_API_KEY}"
SEARCH_API_KEY: "${SEARCH_API_KEY}"
```

---

## 📖 基本介绍

`ToolsNode` 组件是一个专门用于**扩展大语言模型能力**的智能组件。它的主要作用是允许大语言模型调用外部工具来完成特定任务，从而突破模型自身的知识和能力限制。这个组件在 AI 应用开发中扮演着**"能力扩展器"**的角色。

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

---

## 📚 演示示例详解

### 1. 🎯 基础工具创建演示 (`basic`)

**功能**: 演示使用 `utils.InferTool` 自动推断创建工具
```go
// 创建计算器工具
calculatorTool, err := utils.InferTool(
    "calculator",                    // 工具名称
    "执行数学计算和表达式求值",       // 工具描述
    CalculatorFunction,              // 工具函数
)

// 创建时间工具
timeTool, err := utils.InferTool(
    "get_time",
    "获取指定时区的当前时间",
    TimeFunction,
)
```

**输出示例**:
```
📝 创建计算器工具 (自动推断):
    工具名称: calculator
    工具描述: 执行数学计算和表达式求值
    参数类型: main.CalculatorParams

🧮 测试计算器工具:
🧮 执行计算: 10 + 5
✅ 计算完成: 计算结果: 10 + 5 = 15.00
    调用结果: 计算结果: 10 + 5 = 15.00
```

### 2. 🔧 手动工具创建演示 (`manual`)

**功能**: 展示手动实现工具接口的完整过程
```go
// 天气工具实现
type WeatherTool struct {
    apiKey string
}

func (w *WeatherTool) Info() *schema.ToolInfo {
    return &schema.ToolInfo{
        Name:        "get_weather",
        Description: "获取指定城市的当前天气信息",
        ParamsOneOf: WeatherParams{},
    }
}

func (w *WeatherTool) InvokableRun(ctx context.Context, input string, opts ...schema.ToolOption) (string, error) {
    // 工具实现逻辑
}
```

**特点**:
- 完整的接口实现
- 结构化参数定义
- 详细的错误处理
- 模拟真实API调用

### 3. 🏗️ ToolsNode 配置演示 (`config`)

**功能**: 演示如何配置和创建 ToolsNode
```go
// 收集所有工具
tools := []schema.BaseTool{
    calculatorTool,
    timeTool,
    weatherTool,
    searchTool,
}

// 创建 ToolsNode
toolsNode, err := compose.NewToolsNode(ctx, &compose.ToolsNodeConfig{
    Tools: tools,
    Model: mockLLM,
})
```

**配置要点**:
- 工具集合管理
- LLM模型集成
- 配置选项设置
- 错误处理机制

### 4. 🔗 工具调用链演示 (`chain`)

**功能**: 展示复杂的工具调用序列
```go
// 模拟复杂任务执行流程
// 第一步：数学计算
calcResult, err := calculatorTool.InvokableRun(ctx, "25 * 4")

// 第二步：天气查询
weatherInput := `{"city": "上海", "units": "celsius"}`
weatherResult, err := weatherTool.InvokableRun(ctx, weatherInput)

// 第三步：信息搜索
searchInput := `{"query": "人工智能", "num_results": 2}`
searchResult, err := searchTool.InvokableRun(ctx, searchInput)
```

**工作流特点**:
- 多步骤任务编排
- 结果传递和处理
- 错误恢复机制
- 任务状态管理

### 5. ❌ 错误处理演示 (`error`)

**功能**: 全面的错误处理和故障排除演示
```go
// 测试各种错误情况
// 1. 计算器除零错误
_, err := calculatorTool.InvokableRun(ctx, "10 / 0")

// 2. 参数格式错误
_, err = weatherTool.InvokableRun(ctx, `{"invalid": "json"}`)

// 3. 缺少必需参数
_, err = weatherTool.InvokableRun(ctx, `{"units": "celsius"}`)
```

**错误类型**:
- 参数验证错误
- 业务逻辑错误
- JSON格式错误
- 必需参数缺失

### 6. 🚀 性能测试演示 (`performance`)

**功能**: 工具调用的性能测试和优化
```go
// 顺序调用测试
testCases := []string{"10 + 20", "50 - 15", "8 * 7", "100 / 4", "25 + 25"}
totalStart := time.Now()
for _, expr := range testCases {
    result, err := calculatorTool.InvokableRun(ctx, expr)
    // 记录性能数据
}

// 并发调用测试
for i, expr := range testCases {
    go func(index int, expression string) {
        start := time.Now()
        value, err := calculatorTool.InvokableRun(ctx, expression)
        duration := time.Since(start)
        // 收集并发结果
    }(i, expr)
}
```

**性能指标**:
- 单次调用延迟
- 吞吐量测试
- 并发性能对比
- 资源使用情况

---

## ⚙️ 配置说明

### 配置优先级
1. **环境变量** (最高优先级)
2. **config.yaml 文件**
3. **默认值**

### 配置项说明
| 配置项 | 环境变量 | 说明 | 默认值 |
|--------|----------|------|--------|
| API Key | `ARK_API_KEY` | 火山方舟 API 密钥 | - |
| LLM模型 | `LLM_MODEL` | 使用的大语言模型名称 | doubao-seed-1-6-250615 |
| 天气API | `WEATHER_API_KEY` | 天气服务API密钥 | - |
| 搜索API | `SEARCH_API_KEY` | 搜索服务API密钥 | - |

### 工具配置选项

```go
// 工具特定选项
type ToolSpecificOptions struct {
    Timeout     time.Duration  // 执行超时时间
    RetryCount  int           // 重试次数
    UseCache    bool          // 是否使用缓存
}

// 回调配置
toolCallback := func(ctx context.Context, req *tool.CallRequest, resp *tool.CallResponse) {
    log.Printf("工具调用: %s, 输入: %s, 输出: %s",
        req.ToolName, req.Input, resp.Output)
}
```

---

## 📁 代码结构

```
tools_node_demo/
├── main.go                    # 主程序入口
├── README.md                  # 项目文档（本文件）
├── ToolsNode_Summary.md       # ToolsNode 组件完全指南
├── config.yaml                # 配置文件
└── go.mod                     # Go 模块配置
```

### 核心函数说明

| 函数名 | 功能 | 技术特点 |
|--------|------|----------|
| `basicToolCreationDemo()` | 基础工具创建演示 | 自动推断工具生成 |
| `manualToolCreationDemo()` | 手动工具创建演示 | 完整接口实现 |
| `toolsNodeConfigDemo()` | ToolsNode配置演示 | 组件集成和配置 |
| `toolChainDemo()` | 工具调用链演示 | 复杂工作流编排 |
| `errorHandlingDemo()` | 错误处理演示 | 全面错误处理机制 |
| `performanceTestDemo()` | 性能测试演示 | 性能监控和优化 |

### 工具实现类型

| 工具类型 | 创建方式 | 特点 | 适用场景 |
|----------|----------|------|----------|
| `CalculatorFunction` | 自动推断 | 函数式，简单直接 | 数学计算、简单处理 |
| `TimeFunction` | 自动推断 | 参数化，可配置 | 时间查询、格式转换 |
| `WeatherTool` | 手动实现 | 结构体，完整接口 | 外部API调用 |
| `SearchTool` | 手动实现 | 复杂参数，多结果 | 搜索、信息聚合 |

---

## 📊 性能表现

### 工具调用性能测试结果
```
测试环境: macOS (Darwin 24.6.0)
Go版本: 1.24.2
工具类型: 计算器、天气、搜索

性能指标:
✅ 计算器工具: ~1-5ms (本地计算)
✅ 天气查询工具: ~10-50ms (模拟API)
✅ 搜索工具: ~20-100ms (模拟搜索)
✅ 并发调用: 2-5x 性能提升
```

### 性能优化建议
1. **工具设计**: 优化工具内部逻辑，减少不必要的计算
2. **并发处理**: 使用goroutine进行并发工具调用
3. **缓存策略**: 对频繁调用的工具结果进行缓存
4. **超时控制**: 设置合理的工具执行超时时间
5. **资源池**: 复用连接和资源，避免重复创建

---

## 💡 最佳实践

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

### 3. 错误处理策略
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

### 4. 资源管理
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

---

## ❓ 常见问题和解决方案

### Q1: 工具调用失败怎么办？

**常见错误**:
```
工具调用失败: 参数解析失败: invalid character 'i' looking for beginning of value
```

**解决方案**:
1. 检查输入参数的JSON格式是否正确
2. 验证所有必需参数是否提供
3. 确认参数类型是否匹配工具定义
4. 查看工具的参数规范和示例

### Q2: ToolsNode 创建失败

**常见错误**:
```
创建 ToolsNode 失败: model is required
```

**解决方案**:
- 确保提供了有效的LLM模型实例
- 检查模型配置和API密钥
- 验证工具列表不为空
- 确认所有工具都实现了正确的接口

### Q3: 工具执行超时

**常见错误**:
```
工具执行失败: context deadline exceeded
```

**解决方案**:
```go
// 设置合理的超时时间
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

// 或在工具配置中设置
toolOption := tool.WithTimeout(30 * time.Second)
```

### Q4: 工具参数类型不匹配

**优化建议**:
```go
// 使用强类型参数定义
type SearchParams struct {
    Query      string `json:"query" jsonschema:"description=搜索查询内容,required"`
    NumResults int    `json:"num_results,omitempty" jsonschema:"description=返回结果数量,minimum=1,maximum=100"`
}

// 添加参数验证
func (s *SearchTool) validateParams(params *SearchParams) error {
    if params.Query == "" {
        return errors.New("query 参数不能为空")
    }
    if params.NumResults < 1 || params.NumResults > 100 {
        return errors.New("num_results 必须在 1-100 之间")
    }
    return nil
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

### 💡 最佳实践总结
1. **清晰定义**: 工具名称和描述要清晰明确
2. **参数验证**: 实施完善的输入参数验证
3. **错误处理**: 提供有意义的错误信息和恢复机制
4. **性能优化**: 合理使用并发和缓存策略
5. **资源管理**: 正确管理连接、文件句柄等资源

### 🔗 相关资源
- 📚 [官方文档](https://www.cloudwego.io/zh/docs/eino/core_modules/components/tools_node_guide/)
- 📖 [工具创建指南](https://www.cloudwego.io/zh/docs/eino/core_modules/components/tools_node_guide/how_to_create_a_tool/)
- 💻 [示例代码](./main.go)
- 🎯 [ToolsNode 完全指南](./ToolsNode_Summary.md)
- 🌐 [GitHub 仓库](https://github.com/cloudwego/eino)

通过掌握 ToolsNode 组件的各种功能，你将能够构建出更加智能、实用和强大的AI应用系统！🚀