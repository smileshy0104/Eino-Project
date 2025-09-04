# Eino 调用选项(CallOption)能力详细指南

## 📖 **文档概述**

本文档基于 Eino 官方文档 `https://www.cloudwego.io/zh/docs/eino/core_modules/chain_and_graph_orchestration/call_option_capabilities/` 的内容进行详细总结，涵盖了调用选项能力的核心概念、使用方法、代码示例和最佳实践。

---

## 🎯 **CallOption 核心概念**

### **定义与作用**
CallOption 是 Eino 框架中一个强大的配置机制，它允许开发者在 Graph 编译产物调用时，直接传递数据给特定节点，实现**请求粒度的动态配置**。

### **核心特征**
- **请求级配置**：不同于节点的静态 Config，CallOption 是请求级别的动态配置
- **粒度灵活**：支持全局、组件类型、特定节点多个层级的配置
- **运行时调整**：可在运行时动态调整组件参数，无需重新编译
- **类型安全**：提供类型安全的配置接口，避免运行时错误

### **与 Config 的区别**

| 特性 | Config (配置) | CallOption (调用选项) |
|------|---------------|----------------------|
| **作用时机** | Graph 编译时 | Graph 调用时 |
| **配置范围** | 静态配置，影响所有请求 | 动态配置，单次请求生效 |
| **修改频率** | 相对固定，不常变更 | 灵活动态，每次调用可不同 |
| **使用场景** | 基础配置、连接参数 | 业务参数、请求特定配置 |
| **典型示例** | API Key、模型名称 | Temperature、MaxTokens |

```go
// Config 示例：编译时静态配置
chatModel := ark.NewChatModel(&ark.Config{
    APIKey: "your-api-key",     // 静态配置
    Model:  "doubao-pro",       // 不常变更
})

// CallOption 示例：调用时动态配置
result, err := graph.Invoke(ctx, input,
    WithTemperature(0.8),       // 这次调用使用 0.8
    WithMaxTokens(1000),        // 这次调用最多1000个token
)

result2, err := graph.Invoke(ctx, input2,
    WithTemperature(0.2),       // 下次调用使用 0.2
    WithMaxTokens(500),         // 下次调用最多500个token
)
```

---

## 🎨 **CallOption 的两种形态**

### **1. 组件抽象层面的统一 CallOption**

这是最————通用的配置方式，定义了所有同类组件都支持的通用选项。

```go
// 通用的 ChatModel 选项
type CommonChatModelOptions struct {
    Temperature   *float64  // 生成温度
    MaxTokens     *int      // 最大Token数
    TopP          *float64  // 核心采样参数
    TopK          *int      // Top-K采样
    StopWords     []string  // 停止词列表
    Stream        *bool     // 是否流式输出
}

// 通用选项构造函数
func WithTemperature(temp float64) CallOption {
    return &ChatModelOption{Temperature: &temp}
}

func WithMaxTokens(tokens int) CallOption {
    return &ChatModelOption{MaxTokens: &tokens}
}

// 使用示例：适用于任何 ChatModel 实现
result, err := graph.Invoke(ctx, input,
    WithTemperature(0.7),    // 通用选项
    WithMaxTokens(1000),     // 通用选项
    WithTopP(0.9),          // 通用选项
)
```

**优势：**
- ✅ **通用性强**：适用于所有同类组件
- ✅ **接口统一**：提供一致的配置接口
- ✅ **易于维护**：集中定义，统一管理
- ✅ **向前兼容**：新增组件自动支持

### **2. 组件实现层面的特定类型 CallOption**

针对特定组件实现的————专有配置选项，提供更精细的控制能力。

```go
// 火山方舟(Ark)特有选项
type ArkSpecificOptions struct {
    UseCache      *bool     // 是否使用缓存
    RetryCount    *int      // 重试次数
    Timeout       *time.Duration // 超时时间
    Region        *string   // 服务区域
    CustomHeader  map[string]string // 自定义头部
}

// OpenAI 特有选项
type OpenAISpecificOptions struct {
    Organization  *string   // 组织ID
    User          *string   // 用户标识
    LogitBias     map[string]float64 // 词汇偏置
    PresencePenalty *float64 // 存在惩罚
    FrequencyPenalty *float64 // 频率惩罚
}

// 特定选项构造函数
func WithArkCache(useCache bool) CallOption {
    return &ArkOption{UseCache: &useCache}
}

func WithOpenAIOrganization(org string) CallOption {
    return &OpenAIOption{Organization: &org}
}

// 使用示例：针对特定实现
result, err := graph.Invoke(ctx, input,
    WithTemperature(0.7),           // 通用选项
    WithArkCache(true),             // Ark 特有
    WithArkRetryCount(3),           // Ark 特有
)
```

**优势：**
- ✅ **功能完整**：充分利用特定组件的所有能力
- ✅ **精细控制**：提供专业级的配置选项
- ✅ **性能优化**：针对特定实现的优化配置
- ✅ **扩展性强**：支持组件特有的高级功能

---

## 🎯 **CallOption 作用范围控制**

### **1. 全局作用范围**

配置对整个 Graph 中的所有相关组件生效。

```go
// 全局配置：影响所有 ChatModel 组件
result, err := graph.Invoke(ctx, input,
    WithTemperature(0.7),        // 所有ChatModel都使用0.7
    WithCallbacks(globalHandler), // 所有组件都使用这个回调
)
```

**使用场景：**
- 🌍 **统一参数**：整个请求使用相同的生成参数
- 📊 **全局监控**：整个请求使用相同的监控配置
- 🛡️ **安全策略**：整个请求使用相同的安全设置

### **2. 组件类型作用范围**

配置只对特定类型的组件生效。

```go
// 只对 ChatModel 类型的组件生效
result, err := graph.Invoke(ctx, input,
    WithChatModelOption(
        WithTemperature(0.8),     // 只影响ChatModel
        WithMaxTokens(1000),      // 只影响ChatModel
    ),
    WithRetrieverOption(
        WithTopK(5),              // 只影响Retriever
    ),
)
```

**使用场景：**
- 🎯 **精确控制**：不同类型组件使用不同参数
- ⚡ **性能优化**：针对不同组件的特定优化
- 🔧 **专业配置**：利用不同组件的专业能力

### **3. 特定节点作用范围**

配置只对指定名称的节点生效，提供最精确的控制。

```go
// 只对特定节点生效
result, err := graph.Invoke(ctx, input,
    WithTemperature(0.9).DesignateNode("creative_writer"),  // 创意写作节点使用高温度
    WithTemperature(0.1).DesignateNode("fact_checker"),     // 事实检查节点使用低温度
    WithCallbacks(debugHandler).DesignateNode("debug_node"), // 只调试特定节点
)
```

**使用场景：**
- 🎭 **角色差异**：不同角色的AI使用不同参数
- 🐛 **精准调试**：只对有问题的节点进行调试
- 📈 **A/B测试**：对特定节点进行实验配置

---

## 🔧 **CallOption 高级使用模式**

### **1. 选项组合模式**

将多个相关选项组合为一个配置单元，提高代码可读性和复用性。

```go
// 定义配置预设
type ConfigPreset struct {
    name string
    options []CallOption
}

// 创意写作预设
var CreativePreset = ConfigPreset{
    name: "creative",
    options: []CallOption{
        WithTemperature(0.9),
        WithMaxTokens(2000),
        WithTopP(0.95),
        WithPresencePenalty(0.6),
    },
}

// 技术写作预设
var TechnicalPreset = ConfigPreset{
    name: "technical",
    options: []CallOption{
        WithTemperature(0.2),
        WithMaxTokens(1500),
        WithTopP(0.8),
        WithFrequencyPenalty(0.3),
    },
}

// 使用预设
func usePreset(preset ConfigPreset) []CallOption {
    return preset.options
}

// 实际调用
result, err := graph.Invoke(ctx, creativeInput,
    usePreset(CreativePreset)..., // 展开创意预设
)

result2, err := graph.Invoke(ctx, technicalInput,
    usePreset(TechnicalPreset)..., // 展开技术预设
)
```

### **2. 条件选项模式**

根据输入数据或业务逻辑动态选择配置选项。

```go
// 动态选项选择函数
func selectOptionsForTask(taskType string, userLevel string) []CallOption {
    baseOptions := []CallOption{
        WithMaxTokens(1000),
    }
    
    // 根据任务类型调整
    switch taskType {
    case "creative":
        baseOptions = append(baseOptions, WithTemperature(0.8))
    case "analytical":
        baseOptions = append(baseOptions, WithTemperature(0.3))
    case "conversational":
        baseOptions = append(baseOptions, WithTemperature(0.7))
    }
    
    // 根据用户级别调整
    switch userLevel {
    case "beginner":
        baseOptions = append(baseOptions, WithMaxTokens(500))
    case "expert":
        baseOptions = append(baseOptions, WithMaxTokens(2000))
    }
    
    return baseOptions
}

// 使用示例
options := selectOptionsForTask("creative", "expert")
result, err := graph.Invoke(ctx, input, options...)
```

### **3. 链式配置模式**

通过链式调用构建复杂的配置选项。

```go
// 链式配置构建器
type OptionBuilder struct {
    options []CallOption
}

func NewOptionBuilder() *OptionBuilder {
    return &OptionBuilder{options: make([]CallOption, 0)}
}

func (b *OptionBuilder) Temperature(temp float64) *OptionBuilder {
    b.options = append(b.options, WithTemperature(temp))
    return b
}

func (b *OptionBuilder) MaxTokens(tokens int) *OptionBuilder {
    b.options = append(b.options, WithMaxTokens(tokens))
    return b
}

func (b *OptionBuilder) ForNode(nodeName string) *OptionBuilder {
    if len(b.options) > 0 {
        // 给最后一个选项指定节点
        lastOption := b.options[len(b.options)-1]
        b.options[len(b.options)-1] = lastOption.DesignateNode(nodeName)
    }
    return b
}

func (b *OptionBuilder) Build() []CallOption {
    return b.options
}

// 使用链式构建器
options := NewOptionBuilder().
    Temperature(0.8).
    MaxTokens(1000).
    ForNode("writer").
    Temperature(0.2).
    ForNode("reviewer").
    Build()

result, err := graph.Invoke(ctx, input, options...)
```

---

## 💡 **实际应用场景**

### **场景一：智能写作助手**

```go
// 不同写作风格的配置
func createWritingOptions(style string, length string) []CallOption {
    options := []CallOption{}
    
    // 根据写作风格配置
    switch style {
    case "creative":
        options = append(options,
            WithTemperature(0.9),      // 高创意性
            WithTopP(0.95),            // 更多样性
            WithPresencePenalty(0.6),   // 鼓励新话题
        )
    case "formal":
        options = append(options,
            WithTemperature(0.3),      // 更保守
            WithTopP(0.8),             // 相对稳定
            WithFrequencyPenalty(0.3), // 减少重复
        )
    case "technical":
        options = append(options,
            WithTemperature(0.1),      // 非常准确
            WithMaxTokens(2000),       // 允许详细说明
            WithStopWords([]string{"简而言之", "总之"}), // 避免过早总结
        )
    }
    
    // 根据长度配置
    switch length {
    case "short":
        options = append(options, WithMaxTokens(300))
    case "medium":
        options = append(options, WithMaxTokens(800))
    case "long":
        options = append(options, WithMaxTokens(2000))
    }
    
    return options
}

// 使用示例
creativeOptions := createWritingOptions("creative", "long")
result, err := writingGraph.Invoke(ctx, prompt, creativeOptions...)
```

### **场景二：多角色AI协作**

```go
// 多角色协作配置
func setupMultiRoleOptions() []CallOption {
    return []CallOption{
        // 研究员角色：保守准确
        WithTemperature(0.2).DesignateNode("researcher"),
        WithMaxTokens(1000).DesignateNode("researcher"),
        
        // 创意总监：高创意
        WithTemperature(0.8).DesignateNode("creative_director"),
        WithTopP(0.95).DesignateNode("creative_director"),
        
        // 编辑：平衡
        WithTemperature(0.5).DesignateNode("editor"),
        WithMaxTokens(800).DesignateNode("editor"),
        
        // 全局监控
        WithCallbacks(multiRoleHandler),
    }
}

// 使用多角色配置
options := setupMultiRoleOptions()
result, err := collaborationGraph.Invoke(ctx, task, options...)
```

### **场景三：A/B测试配置**

```go
// A/B测试配置管理
type ABTestConfig struct {
    experimentID string
    variant      string
    userID       string
}

func createABTestOptions(config ABTestConfig) []CallOption {
    baseOptions := []CallOption{
        WithCallbacks(abTestHandler(config.experimentID, config.variant)),
    }
    
    switch config.variant {
    case "variant_a":
        baseOptions = append(baseOptions,
            WithTemperature(0.7),
            WithMaxTokens(1000),
        )
    case "variant_b":
        baseOptions = append(baseOptions,
            WithTemperature(0.5),
            WithMaxTokens(1200),
        )
    case "control":
        baseOptions = append(baseOptions,
            WithTemperature(0.6),
            WithMaxTokens(800),
        )
    }
    
    return baseOptions
}

// A/B测试执行
func runABTest(userID string, input interface{}) {
    variant := assignVariant(userID) // 分配实验组
    
    config := ABTestConfig{
        experimentID: "writing_temp_experiment",
        variant:      variant,
        userID:       userID,
    }
    
    options := createABTestOptions(config)
    result, err := graph.Invoke(ctx, input, options...)
    
    // 记录实验结果
    recordABTestResult(config, result, err)
}
```

---

## 🚀 **最佳实践指南**

### **1. 选项设计原则**

#### **优先使用通用选项**
```go
// ✅ 推荐：优先使用通用选项
result, err := graph.Invoke(ctx, input,
    WithTemperature(0.7),    // 通用选项
    WithMaxTokens(1000),     // 通用选项
)

// ⚠️  谨慎使用：只有必要时才使用特定选项
result, err := graph.Invoke(ctx, input,
    WithTemperature(0.7),
    WithArkSpecificOption(value), // 只在必要时使用
)
```

#### **保持选项的内聚性**
```go
// ✅ 好的设计：相关选项组织在一起
type ContentGenerationOptions struct {
    creativity   float64  // 创意度
    length       int      // 长度
    style        string   // 风格
}

func WithContentGeneration(opts ContentGenerationOptions) []CallOption {
    return []CallOption{
        WithTemperature(opts.creativity),
        WithMaxTokens(opts.length),
        WithStyle(opts.style),
    }
}

// ❌ 避免：零散的不相关选项
func badOptionDesign() []CallOption {
    return []CallOption{
        WithTemperature(0.7),
        WithDatabaseTimeout(30*time.Second), // 不相关
        WithMaxTokens(1000),
        WithRedisKey("cache_key"),          // 不相关
    }
}
```

### **2. 性能优化最佳实践**

#### **选项复用和缓存**
```go
// 创建可复用的选项配置
var (
    fastOptions = []CallOption{
        WithTemperature(0.3),
        WithMaxTokens(500),
    }
    
    qualityOptions = []CallOption{
        WithTemperature(0.7),
        WithMaxTokens(1500),
        WithTopP(0.9),
    }
    
    creativeOptions = []CallOption{
        WithTemperature(0.9),
        WithMaxTokens(2000),
        WithTopP(0.95),
        WithPresencePenalty(0.6),
    }
)

// 根据需求选择预设配置
func selectPreset(requirement string) []CallOption {
    switch requirement {
    case "fast":
        return fastOptions
    case "quality":
        return qualityOptions
    case "creative":
        return creativeOptions
    default:
        return qualityOptions
    }
}
```

#### **避免选项冲突**
```go
// ✅ 正确：明确的选项优先级
func mergeOptions(base, override []CallOption) []CallOption {
    optionMap := make(map[string]CallOption)
    
    // 先应用基础选项
    for _, opt := range base {
        optionMap[opt.Key()] = opt
    }
    
    // 再应用覆盖选项
    for _, opt := range override {
        optionMap[opt.Key()] = opt // 覆盖同名选项
    }
    
    result := make([]CallOption, 0, len(optionMap))
    for _, opt := range optionMap {
        result = append(result, opt)
    }
    return result
}

// 使用示例
baseOptions := []CallOption{WithTemperature(0.5)}
userOptions := []CallOption{WithTemperature(0.8)} // 用户偏好
finalOptions := mergeOptions(baseOptions, userOptions)
```

### **3. 错误处理和验证**

#### **选项有效性验证**
```go
// 选项验证器
type OptionValidator struct {
    rules map[string]func(interface{}) error
}

func NewOptionValidator() *OptionValidator {
    return &OptionValidator{
        rules: map[string]func(interface{}) error{
            "temperature": validateTemperature,
            "max_tokens":  validateMaxTokens,
            "top_p":       validateTopP,
        },
    }
}

func validateTemperature(value interface{}) error {
    if temp, ok := value.(float64); ok {
        if temp < 0 || temp > 2 {
            return fmt.Errorf("temperature must be between 0 and 2, got %f", temp)
        }
        return nil
    }
    return fmt.Errorf("temperature must be a float64")
}

func validateMaxTokens(value interface{}) error {
    if tokens, ok := value.(int); ok {
        if tokens < 1 || tokens > 8192 {
            return fmt.Errorf("max_tokens must be between 1 and 8192, got %d", tokens)
        }
        return nil
    }
    return fmt.Errorf("max_tokens must be an int")
}

// 使用验证器
func safeInvoke(ctx context.Context, graph CompiledGraph, input interface{}, options ...CallOption) (interface{}, error) {
    validator := NewOptionValidator()
    
    // 验证所有选项
    for _, opt := range options {
        if err := validator.Validate(opt); err != nil {
            return nil, fmt.Errorf("invalid option: %w", err)
        }
    }
    
    return graph.Invoke(ctx, input, options...)
}
```

### **4. 调试和监控**

#### **选项追踪和日志**
```go
// 选项追踪器
type OptionTracker struct {
    logger Logger
}

func (t *OptionTracker) TrackOptions(options []CallOption) {
    t.logger.Info("Applied CallOptions", "count", len(options))
    
    for i, opt := range options {
        t.logger.Debug("CallOption",
            "index", i,
            "type", opt.Type(),
            "value", opt.Value(),
            "node", opt.TargetNode(),
        )
    }
}

// 带追踪的调用
func invokeWithTracking(ctx context.Context, graph CompiledGraph, input interface{}, options ...CallOption) (interface{}, error) {
    tracker := &OptionTracker{logger: NewLogger()}
    tracker.TrackOptions(options)
    
    return graph.Invoke(ctx, input, options...)
}
```

---

## 📊 **CallOption 架构设计**

### **层次化设计**

```
                    CallOption 生态系统
                           |
        ┌─────────────────────────────────────────────┐
        |                                             |
    通用抽象层                                    实现特化层
        |                                             |
┌───────────────────┐                     ┌─────────────────────┐
│  CommonOptions    │                     │ Implementation      │
│  • Temperature    │                     │ Specific Options    │
│  • MaxTokens      │                     │ • ArkOptions        │
│  • TopP           │                     │ • OpenAIOptions     │
│  • StopWords      │                     │ • CustomOptions     │
└───────────────────┘                     └─────────────────────┘
        |                                             |
        └─────────────────────────────────────────────┘
                           |
                    作用范围控制
                           |
        ┌─────────────────────────────────────────────┐
        |                      |                      |
    全局范围                组件类型范围            特定节点范围
   Global Scope           Component Scope          Node Scope
```

### **接口设计**

```go
// CallOption 核心接口
type CallOption interface {
    // 选项类型标识
    Type() string
    
    // 选项值
    Value() interface{}
    
    // 目标节点（可选）
    TargetNode() string
    
    // 目标组件类型（可选）
    TargetComponentType() string
    
    // 应用选项到上下文
    Apply(ctx context.Context) context.Context
}

// 节点指定接口
type NodeDesignable interface {
    DesignateNode(nodeName string) CallOption
}

// 组件类型指定接口
type ComponentTypeDesignable interface {
    DesignateComponentType(componentType string) CallOption
}

// 选项合并接口
type OptionMergeable interface {
    Merge(other CallOption) (CallOption, error)
}
```

---

## 🔮 **未来发展方向**

### **1. 智能选项推荐**
- 基于历史使用数据推荐最佳选项组合
- 自动A/B测试不同选项配置
- 机器学习优化选项参数

### **2. 可视化配置**
- 图形化的选项配置界面
- 实时预览选项效果
- 拖拽式选项组合

### **3. 动态选项调整**
- 运行时自动调整选项
- 基于反馈的自适应配置
- 智能负载均衡选项

### **4. 选项模板生态**
- 社区共享的选项模板
- 行业最佳实践模板
- 自动化模板生成

---

## 🎯 **总结**

Eino 的 CallOption 能力提供了一套完整、灵活、类型安全的动态配置解决方案：

### **核心价值**
1. **动态性**：运行时灵活调整参数，适应不同业务需求
2. **精确性**：支持全局、组件类型、特定节点多层级配置
3. **类型安全**：编译时类型检查，避免运行时错误
4. **可组合性**：支持选项的灵活组合和复用

### **设计哲学**
- 🎯 **精确控制**：提供细粒度的配置能力
- 🔧 **易于使用**：直观的API设计
- ⚡ **高性能**：优化的选项处理机制
- 🛡️ **类型安全**：强类型系统保障

### **应用价值**
CallOption 使得 AI 应用能够：
- 动态适应不同的业务场景
- 实现精细化的参数控制
- 支持复杂的多角色AI协作
- 便于进行A/B测试和实验

这套机制为构建智能、灵活、可扩展的 AI 应用提供了强大的基础设施支持。

---

*文档基于 Eino 官方文档整理，如有更新请以官方文档为准。*