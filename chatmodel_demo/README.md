# 🤖 Eino ChatModel 组件完全指南

本文档是对 Eino 框架中 `ChatModel` 组件的核心功能和使用方式的完整总结，结合官方文档和实际项目示例。

## 🚀 快速开始

### 配置环境
```bash
# 设置 API Key
export ARK_API_KEY="your-ark-api-key"

# 构建项目
go build -o chatmodel_demo main.go
```

### 运行示例
```bash
# 运行所有示例
./chatmodel_demo

# 运行特定示例
./chatmodel_demo basic        # 基础使用示例
./chatmodel_demo stream       # 流式生成示例
./chatmodel_demo conversation # 多轮对话示例
./chatmodel_demo service      # 智能客服示例
./chatmodel_demo code         # 代码生成示例
./chatmodel_demo performance  # 性能监控示例
```

### 配置文件
项目使用 `config.yaml` 配置文件，也可以通过环境变量设置：
```yaml
ark:
  api_key: "${ARK_API_KEY}"
  model: "doubao-pro-4k"
  base_url: ""
```

---

## 📖 基本介绍

`ChatModel` 组件是一个用于与**大语言模型交互**的核心组件。它的主要作用是将用户的输入消息发送给语言模型，并获取模型的响应。这个组件在 AI 应用开发中扮演着**"大脑"**的角色。

### 🎯 核心价值

在传统的应用开发中，我们只能处理结构化数据和预定义逻辑。而 ChatModel 组件让我们能够：

```
传统应用：固定逻辑 + 结构化数据  ❌
AI 应用：智能推理 + 自然语言理解  ✅
```

### 🚀 主要应用场景

- **💬 自然语言对话**: 构建智能聊天机器人和对话系统
- **📝 文本生成和补全**: 自动生成文章、代码、创意内容等
- **🛠️ 工具调用的参数生成**: 智能分析用户需求并调用相应工具
- **🎭 多模态交互**: 处理文本、图片、音频等多种输入形式
- **🤖 智能Agent系统**: 作为Agent的推理引擎，驱动复杂决策
- **📊 内容分析和理解**: 文本分类、情感分析、信息提取等

---

## 🔧 核心接口

`ChatModel` 组件提供了简洁而强大的接口设计：

### 基础接口

```go
type BaseChatModel interface {
    Generate(ctx context.Context, input []*schema.Message, opts ...Option) (*schema.Message, error)
    Stream(ctx context.Context, input []*schema.Message, opts ...Option) (
        *schema.StreamReader[*schema.Message], error)
}
```

### 工具调用接口

```go
type ToolCallingChatModel interface {
    BaseChatModel
    
    // WithTools 返回绑定了指定工具的新实例
    // 此方法不会修改当前实例，使并发使用更安全
    WithTools(tools []*schema.ToolInfo) (ToolCallingChatModel, error)
}
```

### 接口详解

#### 📤 Generate 方法
- **功能**: 生成完整的模型响应
- **输入**:
    - `ctx`: 上下文对象，用于传递请求级别信息和 Callback Manager
    - `input`: 输入消息列表 (`[]*schema.Message`)
    - `opts`: 可选参数，用于配置模型行为
- **输出**:
    - `*schema.Message`: 模型生成的响应消息
    - `error`: 生成过程中的错误信息

#### 🌊 Stream 方法
- **功能**: 以流式方式生成模型响应
- **参数**: 与 Generate 方法相同
- **输出**:
    - `*schema.StreamReader[*schema.Message]`: 模型响应的流式读取器
    - `error`: 生成过程中的错误信息

#### 🛠️ WithTools 方法
- **功能**: 为模型绑定可用的工具
- **输入**: `tools` - 工具信息列表
- **输出**:
    - `ToolCallingChatModel`: 绑定了工具后的新实例
    - `error`: 绑定过程中的错误信息

---

## 📨 Message 结构体

`Message` 是模型交互的基本数据结构，支持丰富的消息类型：

```go
type Message struct {
    // Role 表示消息的角色（system/user/assistant/tool）
    Role RoleType
    // Content 是消息的文本内容
    Content string
    // MultiContent 是多模态内容，支持文本、图片、音频等
    MultiContent []ChatMessagePart
    // Name 是消息的发送者名称
    Name string
    // ToolCalls 是 assistant 消息中的工具调用信息
    ToolCalls []ToolCall
    // ToolCallID 是 tool 消息的工具调用 ID
    ToolCallID string
    // ResponseMeta 包含响应的元信息
    ResponseMeta *ResponseMeta
    // Extra 用于存储额外信息
    Extra map[string]any
}
```

### 🎭 消息角色类型

- **🔧 system**: 系统消息，用于设定AI的行为和角色
- **👤 user**: 用户消息，来自用户的输入
- **🤖 assistant**: AI助手消息，模型的回复
- **🛠️ tool**: 工具消息，工具执行的结果

### 🎨 多模态支持

`Message` 结构体支持多种内容类型：
- **📝 文本内容**: 通过 `Content` 字段
- **🖼️ 图片内容**: 通过 `MultiContent` 支持图像输入
- **🎵 音频内容**: 支持音频文件处理
- **📹 视频内容**: 支持视频文件分析
- **📎 文件内容**: 支持各种文档格式

---

## ⚙️ 配置选项 (Options)

`ChatModel` 组件提供了丰富的配置选项来控制模型行为：

### 通用选项

```go
type Options struct {
    // Temperature 控制输出的随机性 (0.0-2.0)
    Temperature *float32
    // MaxTokens 控制生成的最大 token 数量
    MaxTokens *int
    // Model 指定使用的模型名称
    Model *string
    // TopP 控制输出的多样性 (0.0-1.0)
    TopP *float32
    // Stop 指定停止生成的条件
    Stop []string
}
```

### 选项设置方法

```go
// 设置温度 - 控制创造性
model.WithTemperature(0.7) // 0.0=确定性, 1.0=平衡, 2.0=高创造性

// 设置最大 token 数 - 控制响应长度
model.WithMaxTokens(2000)

// 设置模型名称 - 选择特定模型
model.WithModel("gpt-4")

// 设置 top_p 值 - 控制词汇选择范围
model.WithTopP(0.9)

// 设置停止词 - 定义生成结束条件
model.WithStop([]string{"\n\n", "结束"})
```

### 参数调优指南

| 参数 | 推荐值 | 适用场景 |
|------|--------|----------|
| **Temperature** | 0.1-0.3 | 事实性回答、代码生成 |
| **Temperature** | 0.7-0.9 | 创意写作、头脑风暴 |
| **MaxTokens** | 500-1000 | 简短回答 |
| **MaxTokens** | 2000-4000 | 详细分析、长文生成 |
| **TopP** | 0.9-0.95 | 平衡质量和多样性 |

---

## 🛠️ 使用方式

### 1. 单独使用

这是最直接的使用方式，适合简单的对话和文本生成场景：

```go
import (
    "context"
    "fmt"
    "io"
    
    "github.com/cloudwego/eino-ext/components/model/ark"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/schema"
)

func basicChatModelExample() {
    ctx := context.Background()
    
    // 1. 初始化模型 (以ARK为例)
    cm, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
        APIKey: "YOUR_API_KEY",
        Model:  "doubao-pro-4k",
        Timeout: 30 * time.Second,
    })
    if err != nil {
        log.Fatal("初始化模型失败:", err)
    }
    
    // 2. 准备输入消息
    messages := []*schema.Message{
        {
            Role:    schema.System,
            Content: "你是一个有帮助的AI助手，请用简洁明了的方式回答问题。",
        },
        {
            Role:    schema.User,
            Content: "请解释什么是机器学习？",
        },
    }
    
    // 3. 生成响应
    response, err := cm.Generate(ctx, messages, 
        model.WithTemperature(0.7),
        model.WithMaxTokens(1000),
    )
    if err != nil {
        log.Fatal("生成响应失败:", err)
    }
    
    // 4. 处理响应
    fmt.Printf("AI回复: %s\n", response.Content)
    
    // 5. 流式生成示例
    fmt.Println("\n=== 流式生成 ===")
    streamResult, err := cm.Stream(ctx, messages)
    if err != nil {
        log.Fatal("流式生成失败:", err)
    }
    defer streamResult.Close()
    
    for {
        chunk, err := streamResult.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Printf("接收流式数据出错: %v", err)
            break
        }
        // 实时输出响应片段
        fmt.Print(chunk.Content)
    }
    fmt.Println()
}
```

### 2. 在编排中使用 (推荐)

与其他 Eino 组件结合使用，构建复杂的 AI 工作流：

```go
import (
    "github.com/cloudwego/eino/schema"
    "github.com/cloudwego/eino/compose"
)

func orchestrationExample() {
    ctx := context.Background()
    
    // 1. 初始化 ChatModel
    cm, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
        APIKey: "YOUR_API_KEY",
        Model:  "doubao-pro-4k",
    })
    if err != nil {
        log.Fatal("初始化模型失败:", err)
    }
    
    // 2. 在 Chain 中使用
    chain := compose.NewChain[[]*schema.Message, *schema.Message]()
    chain.AppendChatModel(cm)
    
    // 编译并运行
    runnable, err := chain.Compile(ctx)
    if err != nil {
        log.Fatal("编译链失败:", err)
    }
    
    messages := []*schema.Message{
        {Role: schema.User, Content: "你好！"},
    }
    
    result, err := runnable.Invoke(ctx, messages)
    if err != nil {
        log.Fatal("执行链失败:", err)
    }
    
    fmt.Printf("链式处理结果: %s\n", result.Content)
    
    // 3. 在 Graph 中使用
    graph := compose.NewGraph[[]*schema.Message, *schema.Message]()
    graph.AddChatModelNode("chat_model", cm)
    
    // 设置图的流程
    graph.AddEdge(compose.START, "chat_model")
    graph.AddEdge("chat_model", compose.END)
    
    graphRunnable, err := graph.Compile(ctx)
    if err != nil {
        log.Fatal("编译图失败:", err)
    }
    
    graphResult, err := graphRunnable.Invoke(ctx, messages)
    if err != nil {
        log.Fatal("执行图失败:", err)
    }
    
    fmt.Printf("图式处理结果: %s\n", graphResult.Content)
}
```

### 3. 工具调用集成

展示如何将 ChatModel 与工具系统集成：

```go
func toolCallingExample() {
    ctx := context.Background()
    
    // 1. 创建工具
    tools := []tool.InvokableTool{
        // 假设已经实现了计算器工具
        NewCalculatorTool(),
        NewWeatherTool(),
    }
    
    // 2. 初始化支持工具调用的模型
    cm, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
        APIKey: "YOUR_API_KEY",
        Model:  "doubao-pro-4k",
    })
    if err != nil {
        log.Fatal("初始化模型失败:", err)
    }
    
    // 3. 绑定工具到模型
    toolInfos := make([]*schema.ToolInfo, 0, len(tools))
    for _, tool := range tools {
        info, err := tool.Info(ctx)
        if err != nil {
            log.Printf("获取工具信息失败: %v", err)
            continue
        }
        toolInfos = append(toolInfos, info)
    }
    
    // 使用 BindTools 方法绑定工具
    cm.BindTools(toolInfos)
    
    // 4. 发送需要工具调用的消息
    messages := []*schema.Message{
        {
            Role:    schema.System,
            Content: "你是一个智能助手，可以使用工具来帮助用户解决问题。",
        },
        {
            Role:    schema.User,
            Content: "请帮我计算 123 * 456 的结果，然后查询北京今天的天气。",
        },
    }
    
    // 5. 生成响应（可能包含工具调用）
    response, err := cm.Generate(ctx, messages)
    if err != nil {
        log.Fatal("生成响应失败:", err)
    }
    
    // 6. 处理工具调用
    if len(response.ToolCalls) > 0 {
        fmt.Printf("模型请求调用 %d 个工具:\n", len(response.ToolCalls))
        
        for _, toolCall := range response.ToolCalls {
            fmt.Printf("- 工具: %s, 参数: %s\n", 
                toolCall.Function.Name, 
                toolCall.Function.Arguments)
            
            // 这里可以执行实际的工具调用
            // result := executeTool(toolCall)
        }
    } else {
        fmt.Printf("直接回复: %s\n", response.Content)
    }
}
```

---

## 📊 回调机制 (Callbacks)

Eino 提供了强大的回调机制，允许开发者在组件执行的关键点注入自定义逻辑，用于**监控**、**日志记录**、**性能分析**和**错误处理**。

### 回调事件

Eino 使用 `callbacks.NewHandlerBuilder()` 创建回调处理器，支持以下事件：

- **`OnStartFn`**: 在组件开始执行时触发，接收输入数据
- **`OnEndFn`**: 在组件成功执行后触发，接收输出数据
- **`OnErrorFn`**: 在发生错误时触发，接收错误信息
- **`OnStartWithStreamInputFn`**: 处理流式输入的开始事件
- **`OnEndWithStreamOutputFn`**: 处理流式输出的结束事件

### 核心特性

- **通用性**: 支持所有 Eino 组件 (ChatModel, Chain, Graph, Tools 等)
- **标准化**: 统一的回调接口和数据结构
- **灵活性**: 可以组合多个回调处理器
- **性能友好**: 轻量级回调机制，不影响主要业务逻辑

### 使用示例

```go
import "github.com/cloudwego/eino/callbacks"

func callbackExample(ctx context.Context, cm model.BaseChatModel) {
    // 1. 创建 Eino 官方回调处理器
    callbackHandler := callbacks.NewHandlerBuilder().
        OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
            fmt.Printf("[官方回调] 🚀 组件开始执行\n")
            fmt.Printf("[官方回调] 📝 输入数据长度: %d\n", len(fmt.Sprintf("%v", input)))
            return ctx
        }).
        OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
            fmt.Printf("[官方回调] ✅ 组件执行完成\n")
            fmt.Printf("[官方回调] 📤 输出数据长度: %d\n", len(fmt.Sprintf("%v", output)))
            return ctx
        }).
        OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
            fmt.Printf("[官方回调] ❌ 组件执行失败: %v\n", err)
            return ctx
        }).
        Build()
    
    // 2. 在编排中使用回调
    chain := compose.NewChain[[]*schema.Message, *schema.Message]()
    chain.AppendChatModel(cm)
    
    runnable, err := chain.Compile(ctx)
    if err != nil {
        log.Printf("编译 Chain 失败: %v", err)
        return
    }
    
    messages := []*schema.Message{
        {Role: schema.User, Content: "你好！"},
    }
    
    // 3. 带回调执行
    result, err := runnable.Invoke(ctx, messages,
        compose.WithCallbacks(callbackHandler),
    )
    if err != nil {
        log.Printf("执行失败: %v", err)
        return
    }
    
    fmt.Printf("最终结果: %s\n", result.Content)
}
```

### 高级回调应用

```go
// 性能监控回调
type PerformanceMonitor struct {
    startTime time.Time
    metrics   map[string]interface{}
    mu        sync.RWMutex
}

func NewPerformanceMonitor() *PerformanceMonitor {
    return &PerformanceMonitor{
        metrics: make(map[string]interface{}),
    }
}

func (p *PerformanceMonitor) CreateHandler() callbacks.Handler {
    return callbacks.NewHandlerBuilder().
        OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
            p.mu.Lock()
            defer p.mu.Unlock()
            
            p.startTime = time.Now()
            p.metrics = make(map[string]interface{})
            p.metrics["component_type"] = "chatmodel"
            p.metrics["input_size"] = len(fmt.Sprintf("%v", input))
            
            fmt.Printf("[性能监控] 开始执行，输入大小: %d\n", p.metrics["input_size"])
            return ctx
        }).
        OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
            p.mu.Lock()
            defer p.mu.Unlock()
            
            duration := time.Since(p.startTime)
            p.metrics["duration_ms"] = duration.Milliseconds()
            p.metrics["output_size"] = len(fmt.Sprintf("%v", output))
            p.metrics["status"] = "success"
            
            // 计算处理速度
            if duration.Seconds() > 0 {
                p.metrics["chars_per_second"] = float64(p.metrics["output_size"].(int)) / duration.Seconds()
            }
            
            p.sendMetrics()
            return ctx
        }).
        OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
            p.mu.Lock()
            defer p.mu.Unlock()
            
            duration := time.Since(p.startTime)
            p.metrics["duration_ms"] = duration.Milliseconds()
            p.metrics["error"] = err.Error()
            p.metrics["status"] = "failed"
            
            p.sendMetrics()
            return ctx
        }).
        Build()
}

func (p *PerformanceMonitor) sendMetrics() {
    // 发送指标到监控系统（如 Prometheus、DataDog 等）
    fmt.Printf("[性能指标] %+v\n", p.metrics)
}

// 使用示例
func performanceMonitorExample(ctx context.Context, cm model.BaseChatModel) {
    monitor := NewPerformanceMonitor()
    handler := monitor.CreateHandler()
    
    chain := compose.NewChain[[]*schema.Message, *schema.Message]()
    chain.AppendChatModel(cm)
    
    runnable, _ := chain.Compile(ctx)
    
    messages := []*schema.Message{
        {Role: schema.User, Content: "解释一下人工智能的工作原理"},
    }
    
    // 使用性能监控执行
    _, err := runnable.Invoke(ctx, messages, compose.WithCallbacks(handler))
    if err != nil {
        log.Printf("执行失败: %v", err)
    }
}
```

---

## 🎯 实际应用示例

### 1. 智能客服系统

```go
type CustomerServiceBot struct {
    chatModel model.BaseChatModel
    knowledge []string // 知识库
}

func NewCustomerServiceBot(cm model.BaseChatModel) *CustomerServiceBot {
    return &CustomerServiceBot{
        chatModel: cm,
        knowledge: []string{
            "我们的营业时间是周一到周五 9:00-18:00",
            "退货政策：7天无理由退货",
            "配送时间：1-3个工作日",
        },
    }
}

func (bot *CustomerServiceBot) HandleCustomerQuery(ctx context.Context, query string) (string, error) {
    // 构建系统提示
    systemPrompt := fmt.Sprintf(`你是一个专业的客服助手。
知识库信息：
%s

请根据知识库信息回答用户问题，如果知识库中没有相关信息，请礼貌地告知用户联系人工客服。`,
        strings.Join(bot.knowledge, "\n"))
    
    messages := []*schema.Message{
        {Role: schema.System, Content: systemPrompt},
        {Role: schema.User, Content: query},
    }
    
    response, err := bot.chatModel.Generate(ctx, messages,
        model.WithTemperature(0.3), // 较低温度确保回答准确
        model.WithMaxTokens(500),
    )
    if err != nil {
        return "", err
    }
    
    return response.Content, nil
}
```

### 2. 代码生成助手

```go
type CodeGenerator struct {
    chatModel model.BaseChatModel
}

func NewCodeGenerator(cm model.BaseChatModel) *CodeGenerator {
    return &CodeGenerator{chatModel: cm}
}

func (cg *CodeGenerator) GenerateCode(ctx context.Context, requirement string, language string) (string, error) {
    systemPrompt := fmt.Sprintf(`你是一个专业的%s程序员。
请根据用户需求生成高质量的代码，要求：
1. 代码结构清晰，注释完整
2. 遵循最佳实践和编码规范
3. 包含必要的错误处理
4. 提供使用示例`, language)
    
    messages := []*schema.Message{
        {Role: schema.System, Content: systemPrompt},
        {Role: schema.User, Content: fmt.Sprintf("请用%s实现：%s", language, requirement)},
    }
    
    response, err := cg.chatModel.Generate(ctx, messages,
        model.WithTemperature(0.2), // 低温度确保代码准确性
        model.WithMaxTokens(2000),
    )
    if err != nil {
        return "", err
    }
    
    return response.Content, nil
}

// 使用示例
func codeGeneratorExample() {
    ctx := context.Background()
    
    cm, _ := ark.NewChatModel(ctx, &ark.ChatModelConfig{
        APIKey: "YOUR_API_KEY",
        Model:  "doubao-pro-4k",
    })
    
    generator := NewCodeGenerator(cm)
    
    code, err := generator.GenerateCode(ctx, 
        "实现一个线程安全的计数器", 
        "Go")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("生成的代码：")
    fmt.Println(code)
}
```

### 3. 多轮对话管理

```go
type ConversationManager struct {
    chatModel model.BaseChatModel
    history   []*schema.Message
    maxHistory int
}

func NewConversationManager(cm model.BaseChatModel) *ConversationManager {
    return &ConversationManager{
        chatModel: cm,
        history:   make([]*schema.Message, 0),
        maxHistory: 10, // 保持最近10轮对话
    }
}

func (cm *ConversationManager) Chat(ctx context.Context, userInput string) (string, error) {
    // 添加用户消息到历史
    userMessage := &schema.Message{
        Role:    schema.User,
        Content: userInput,
    }
    cm.history = append(cm.history, userMessage)
    
    // 构建完整的对话历史
    messages := make([]*schema.Message, 0, len(cm.history)+1)
    
    // 添加系统提示
    systemMessage := &schema.Message{
        Role:    schema.System,
        Content: "你是一个友好的AI助手，能够记住对话历史并提供连贯的回复。",
    }
    messages = append(messages, systemMessage)
    
    // 添加历史对话（限制长度）
    startIdx := 0
    if len(cm.history) > cm.maxHistory {
        startIdx = len(cm.history) - cm.maxHistory
    }
    messages = append(messages, cm.history[startIdx:]...)
    
    // 生成回复
    response, err := cm.chatModel.Generate(ctx, messages,
        model.WithTemperature(0.7),
        model.WithMaxTokens(1000),
    )
    if err != nil {
        return "", err
    }
    
    // 添加AI回复到历史
    cm.history = append(cm.history, response)
    
    // 清理过长的历史
    if len(cm.history) > cm.maxHistory*2 {
        cm.history = cm.history[len(cm.history)-cm.maxHistory:]
    }
    
    return response.Content, nil
}

func (cm *ConversationManager) ClearHistory() {
    cm.history = make([]*schema.Message, 0)
}

func (cm *ConversationManager) GetHistoryLength() int {
    return len(cm.history)
}
```

---

## 🔧 最佳实践

### 1. 性能优化

#### 连接池管理
```go
type ChatModelPool struct {
    models chan model.BaseChatModel
    config *ark.ChatModelConfig
}

func NewChatModelPool(size int, config *ark.ChatModelConfig) *ChatModelPool {
    pool := &ChatModelPool{
        models: make(chan model.BaseChatModel, size),
        config: config,
    }
    
    // 预创建模型实例
    for i := 0; i < size; i++ {
        cm, err := ark.NewChatModel(context.Background(), config)
        if err != nil {
            log.Printf("创建模型实例失败: %v", err)
            continue
        }
        pool.models <- cm
    }
    
    return pool
}

func (p *ChatModelPool) Get() model.BaseChatModel {
    return <-p.models
}

func (p *ChatModelPool) Put(cm model.BaseChatModel) {
    select {
    case p.models <- cm:
    default:
        // 池已满，丢弃实例
    }
}
```

#### 请求批处理
```go
type BatchProcessor struct {
    chatModel model.BaseChatModel
    batchSize int
    timeout   time.Duration
}

func (bp *BatchProcessor) ProcessBatch(ctx context.Context, requests []ChatRequest) ([]ChatResponse, error) {
    responses := make([]ChatResponse, len(requests))
    
    // 并发处理批次
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, bp.batchSize)
    
    for i, req := range requests {
        wg.Add(1)
        go func(index int, request ChatRequest) {
            defer wg.Done()
            semaphore <- struct{}{} // 获取信号量
            defer func() { <-semaphore }() // 释放信号量
            
            ctx, cancel := context.WithTimeout(ctx, bp.timeout)
            defer cancel()
            
            response, err := bp.chatModel.Generate(ctx, request.Messages)
            responses[index] = ChatResponse{
                Response: response,
                Error:    err,
            }
        }(i, req)
    }
    
    wg.Wait()
    return responses, nil
}
```

### 2. 错误处理和重试

```go
type ResilientChatModel struct {
    chatModel   model.BaseChatModel
    maxRetries  int
    retryDelay  time.Duration
    backoffRate float64
}

func NewResilientChatModel(cm model.BaseChatModel) *ResilientChatModel {
    return &ResilientChatModel{
        chatModel:   cm,
        maxRetries:  3,
        retryDelay:  time.Second,
        backoffRate: 2.0,
    }
}

func (rcm *ResilientChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
    var lastErr error
    delay := rcm.retryDelay
    
    for attempt := 0; attempt <= rcm.maxRetries; attempt++ {
        if attempt > 0 {
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(delay):
                delay = time.Duration(float64(delay) * rcm.backoffRate)
            }
        }
        
        response, err := rcm.chatModel.Generate(ctx, messages, opts...)
        if err == nil {
            return response, nil
        }
        
        lastErr = err
        
        // 判断是否应该重试
        if !shouldRetry(err) {
            break
        }
        
        log.Printf("第 %d 次尝试失败: %v", attempt+1, err)
    }
    
    return nil, fmt.Errorf("重试 %d 次后仍然失败: %w", rcm.maxRetries, lastErr)
}

func shouldRetry(err error) bool {
    // 根据错误类型判断是否应该重试
    if strings.Contains(err.Error(), "timeout") {
        return true
    }
    if strings.Contains(err.Error(), "rate limit") {
        return true
    }
    if strings.Contains(err.Error(), "server error") {
        return true
    }
    return false
}
```

### 3. 监控和日志

```go
type MonitoredChatModel struct {
    chatModel model.BaseChatModel
    metrics   *Metrics
    logger    *log.Logger
}

type Metrics struct {
    TotalRequests    int64
    SuccessRequests  int64
    FailedRequests   int64
    TotalTokens      int64
    AverageLatency   time.Duration
    mu               sync.RWMutex
}

func (mcm *MonitoredChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
    startTime := time.Now()
    
    // 记录请求开始
    mcm.metrics.mu.Lock()
    mcm.metrics.TotalRequests++
    mcm.metrics.mu.Unlock()
    
    mcm.logger.Printf("[ChatModel] 开始处理请求，消息数: %d", len(messages))
    
    // 执行实际请求
    response, err := mcm.chatModel.Generate(ctx, messages, opts...)
    
    // 记录结果
    duration := time.Since(startTime)
    mcm.metrics.mu.Lock()
    if err != nil {
        mcm.metrics.FailedRequests++
        mcm.logger.Printf("[ChatModel] 请求失败，耗时: %v, 错误: %v", duration, err)
    } else {
        mcm.metrics.SuccessRequests++
        if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
            mcm.metrics.TotalTokens += int64(response.ResponseMeta.Usage.TotalTokens)
        }
        mcm.logger.Printf("[ChatModel] 请求成功，耗时: %v", duration)
    }
    
    // 更新平均延迟
    mcm.updateAverageLatency(duration)
    mcm.metrics.mu.Unlock()
    
    return response, err
}

func (mcm *MonitoredChatModel) updateAverageLatency(newLatency time.Duration) {
    // 简单的移动平均
    if mcm.metrics.AverageLatency == 0 {
        mcm.metrics.AverageLatency = newLatency
    } else {
        mcm.metrics.AverageLatency = (mcm.metrics.AverageLatency + newLatency) / 2
    }
}

func (mcm *MonitoredChatModel) GetMetrics() Metrics {
    mcm.metrics.mu.RLock()
    defer mcm.metrics.mu.RUnlock()
    return *mcm.metrics
}
```

### 4. 配置管理

```go
type ChatModelConfig struct {
    Provider    string        `yaml:"provider"`    // ark, openai, etc.
    Model       string        `yaml:"model"`       // 模型名称
    APIKey      string        `yaml:"api_key"`     // API密钥
    BaseURL     string        `yaml:"base_url"`    // 基础URL
    Temperature float32       `yaml:"temperature"` // 温度
    MaxTokens   int           `yaml:"max_tokens"`  // 最大token数
    Timeout     time.Duration `yaml:"timeout"`     // 超时时间
    Retries     int           `yaml:"retries"`     // 重试次数
}

type ChatModelFactory struct {
    configs map[string]*ChatModelConfig
}

func NewChatModelFactory(configFile string) (*ChatModelFactory, error) {
    data, err := ioutil.ReadFile(configFile)
    if err != nil {
        return nil, err
    }
    
    var configs map[string]*ChatModelConfig
    if err := yaml.Unmarshal(data, &configs); err != nil {
        return nil, err
    }
    
    return &ChatModelFactory{configs: configs}, nil
}

func (factory *ChatModelFactory) CreateChatModel(ctx context.Context, name string) (model.BaseChatModel, error) {
    config, exists := factory.configs[name]
    if !exists {
        return nil, fmt.Errorf("配置 %s 不存在", name)
    }
    
    switch config.Provider {
    case "ark":
        return ark.NewChatModel(ctx, &ark.ChatModelConfig{
            APIKey:      config.APIKey,
            Model:       config.Model,
            BaseURL:     config.BaseURL,
            Temperature: config.Temperature,
            MaxTokens:   config.MaxTokens,
            Timeout:     config.Timeout,
        })
    case "openai":
        // 实现 OpenAI 模型创建
        return nil, fmt.Errorf("OpenAI 提供商暂未实现")
    default:
        return nil, fmt.Errorf("不支持的提供商: %s", config.Provider)
    }
}
```

---

## 🚨 常见问题和解决方案

### 1. Token 限制问题

**问题**: 输入或输出超过模型的 token 限制

**解决方案**:
```go
func truncateMessages(messages []*schema.Message, maxTokens int) []*schema.Message {
    // 简单的截断策略：保留系统消息和最近的用户消息
    if len(messages) <= 2 {
        return messages
    }
    
    result := make([]*schema.Message, 0)
    
    // 保留系统消息
    for _, msg := range messages {
        if msg.Role == schema.System {
            result = append(result, msg)
            break
        }
    }
    
    // 从后往前添加消息，直到接近 token 限制
    estimatedTokens := 0
    for i := len(messages) - 1; i >= 0; i-- {
        msg := messages[i]
        if msg.Role == schema.System {
            continue
        }
        
        // 粗略估算 token 数（1 token ≈ 4 字符）
        msgTokens := len(msg.Content) / 4
        if estimatedTokens+msgTokens > maxTokens {
            break
        }
        
        result = append([]*schema.Message{msg}, result...)
        estimatedTokens += msgTokens
    }
    
    return result
}
```

### 2. 速率限制处理

**问题**: API 调用频率过高导致限流

**解决方案**:
```go
type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    return rl.limiter.Wait(ctx)
}

type RateLimitedChatModel struct {
    chatModel   model.BaseChatModel
    rateLimiter *RateLimiter
}

func (rlcm *RateLimitedChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
    // 等待速率限制
    if err := rlcm.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }
    
    return rlcm.chatModel.Generate(ctx, messages, opts...)
}
```

### 3. 内存泄漏预防

**问题**: 长时间运行导致内存泄漏

**解决方案**:
```go
type ManagedChatModel struct {
    chatModel model.BaseChatModel
    cleanup   func()
}

func NewManagedChatModel(ctx context.Context, config *ark.ChatModelConfig) (*ManagedChatModel, error) {
    cm, err := ark.NewChatModel(ctx, config)
    if err != nil {
        return nil, err
    }
    
    managed := &ManagedChatModel{
        chatModel: cm,
    }
    
    // 设置清理函数
    managed.cleanup = func() {
        // 清理资源
        if closer, ok := cm.(io.Closer); ok {
            closer.Close()
        }
    }
    
    // 注册清理函数
    runtime.SetFinalizer(managed, (*ManagedChatModel).finalize)
    
    return managed, nil
}

func (mcm *ManagedChatModel) finalize() {
    if mcm.cleanup != nil {
        mcm.cleanup()
    }
}

func (mcm *ManagedChatModel) Close() {
    mcm.finalize()
    runtime.SetFinalizer(mcm, nil)
}
```

---

## 📚 相关资源

- **官方文档**: [Eino ChatModel 使用说明](https://www.cloudwego.io/zh/docs/eino/core_modules/components/chat_model_guide/)
- **GitHub 仓库**: [cloudwego/eino](https://github.com/cloudwego/eino)
- **示例代码**: 查看本项目的 `main.go` 文件获取完整示例
- **社区支持**: [CloudWeGo 社区](https://github.com/cloudwego/community)
- **相关组件**:
    - [ChatTemplate 组件指南](../chattemplate_demo/ChatTemplate_Summary.md)
    - [Tool 组件指南](../tool_demo/Tool_Summary.md)
    - [Embedding 组件指南](../embedding_demo/README.md)

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个示例项目！

---

*最后更新: 2024年12月*