# 🤖 Eino ChatModel 完全指南

> 💡 **核心价值**: ChatModel 是 Eino 框架中与大语言模型交互的**核心组件**，提供标准化、高性能的模型调用接口，支持多种模型平台和交互模式。

---

## 📋 目录
- [核心概念](#-核心概念)
- [功能特性](#-功能特性)
- [核心接口](#-核心接口)
- [消息结构体系](#-消息结构体系)
- [配置选项详解](#-配置选项详解)
- [基础使用方法](#-基础使用方法)
- [流式处理机制](#-流式处理机制)
- [工具调用支持](#-工具调用支持)
- [回调机制深入](#-回调机制深入)
- [多模型平台支持](#-多模型平台支持)
- [性能优化策略](#-性能优化策略)
- [错误处理最佳实践](#-错误处理最佳实践)
- [高级用法和技巧](#-高级用法和技巧)
- [常见问题解决](#-常见问题解决)

---

## 🎯 核心概念

**ChatModel** 是 Eino 框架中负责与大语言模型进行交互的**标准化组件**。它抽象了不同模型平台的差异，提供统一的接口和使用方式。

### 🌟 设计哲学

```
🎭 标准化抽象
├─ 🔌 统一接口设计
├─ 🔄 跨平台兼容
└─ 📊 一致的数据结构

🚀 高性能设计  
├─ ⚡ 异步处理机制
├─ 🌊 流式输出支持
└─ 🎯 资源优化管理

🧩 组件化集成
├─ 🔗 编排系统集成
├─ 🛠️ 工具调用支持
└─ 📈 回调机制扩展
```

### 🏗️ 核心价值

| 价值维度 | 说明 | 优势 |
|----------|------|------|
| 🔧 **抽象统一** | 统一不同模型平台的接口 | 降低平台切换成本 |
| 🚀 **高性能** | 异步流式处理机制 | 提升用户体验 |
| 🛠️ **扩展性** | 丰富的配置和回调机制 | 满足复杂业务需求 |
| 🔒 **可靠性** | 完善的错误处理和重试机制 | 保障系统稳定性 |

---

## ⚡ 功能特性

### 🎨 核心功能矩阵

| 功能类别 | 功能 | 描述 | 适用场景 |
|----------|------|------|----------|
| 🔄 **基础交互** | Generate | 单次完整响应生成 | 简单问答、文档生成 |
| 🌊 **流式处理** | Stream | 实时流式响应输出 | 长文本生成、实时对话 |
| 🛠️ **工具集成** | WithTools | 模型工具调用能力 | 复杂任务执行、外部API调用 |
| 📊 **多模态** | MultiModal | 支持文本、图片等多种输入 | 图文理解、视觉问答 |
| 🎛️ **参数控制** | Options | 丰富的生成参数配置 | 精确控制输出效果 |
| 📈 **监控扩展** | Callbacks | 生命周期回调机制 | 日志记录、性能监控 |

### 🌍 支持的模型平台

```
云端模型平台:
├─ 🌐 OpenAI (GPT系列)
├─ 🚀 字节跳动 ARK
├─ 🧠 Anthropic Claude  
├─ 🤖 Google Gemini
└─ 💫 其他云端API

本地部署平台:
├─ 🏠 Ollama (本地模型)
├─ 🔧 vLLM (高性能推理)
├─ 📦 LocalAI (本地API)
└─ 🛠️ 自定义实现
```

---

## 🔌 核心接口

ChatModel 的接口设计简洁而强大，主要包含三个核心方法：

```go
type ChatModel interface {
    // 🎯 生成完整响应
    Generate(ctx context.Context, messages []*schema.Message, opts ...Option) (*schema.Message, error)
    
    // 🌊 流式生成响应  
    Stream(ctx context.Context, messages []*schema.Message, opts ...Option) (*schema.StreamReader[*schema.Message], error)
    
    // 🛠️ 绑定工具调用
    WithTools(tools []*schema.ToolInfo) ChatModel
}
```

### 📋 接口详解

#### 1. Generate 方法 - 完整响应生成

```go
func (cm ChatModel) Generate(
    ctx context.Context,           // 🔄 上下文控制
    messages []*schema.Message,    // 📝 输入消息列表
    opts ...Option                 // ⚙️ 可选配置参数
) (*schema.Message, error)         // 📊 返回完整响应
```

**使用场景**:
- 📝 文档生成任务
- ❓ 简单问答交互
- 🎯 一次性任务处理
- 📊 批量处理场景

#### 2. Stream 方法 - 流式响应生成

```go
func (cm ChatModel) Stream(
    ctx context.Context,           // 🔄 上下文控制
    messages []*schema.Message,    // 📝 输入消息列表  
    opts ...Option                 // ⚙️ 可选配置参数
) (*schema.StreamReader[*schema.Message], error) // 🌊 返回流式读取器
```

**使用场景**:
- 💬 实时对话系统
- 📖 长文本逐步生成
- 🎬 打字机效果展示
- ⚡ 用户体验优化

#### 3. WithTools 方法 - 工具绑定

```go
func (cm ChatModel) WithTools(
    tools []*schema.ToolInfo       // 🛠️ 可用工具列表
) ChatModel                        // 🔄 返回增强的模型实例
```

**使用场景**:
- 🧮 数学计算工具
- 🌐 外部API调用
- 📊 数据查询工具
- 🔍 搜索引擎集成

---

## 💬 消息结构体系

Eino 定义了标准的消息结构，确保跨平台兼容性：

### 🏗️ Message 结构

```go
type Message struct {
    Role     MessageRole  `json:"role"`     // 🎭 消息角色
    Content  string       `json:"content"`  // 📝 文本内容
    Name     string       `json:"name"`     // 👤 发送者名称 (可选)
    ToolCalls []ToolCall  `json:"tool_calls"` // 🛠️ 工具调用列表
    ToolCallID string     `json:"tool_call_id"` // 🔗 工具调用ID
    Extra    map[string]any `json:"extra"`   // 📊 额外元数据
}
```

### 🎭 消息角色类型

```go
const (
    System    MessageRole = "system"    // 🤖 系统指令
    User      MessageRole = "user"      // 👤 用户输入
    Assistant MessageRole = "assistant" // 🤖 助手回复
    Tool      MessageRole = "tool"      // 🛠️ 工具响应
)
```

### 📋 角色使用说明

| 角色 | 用途 | 示例 |
|------|------|------|
| 🤖 **System** | 定义AI助手行为和身份 | "你是一个专业的编程助手" |
| 👤 **User** | 用户的问题和请求 | "请帮我写一个排序算法" |
| 🤖 **Assistant** | AI的回复和响应 | "我来为你介绍几种排序算法..." |
| 🛠️ **Tool** | 工具调用的结果 | "计算结果: 42" |

### 🌈 多模态内容支持

```go
// 📝 纯文本消息
textMessage := &schema.Message{
    Role:    schema.User,
    Content: "描述这张图片的内容",
}

// 🖼️ 多模态消息 (文本 + 图片)
multiModalMessage := &schema.Message{
    Role:    schema.User,
    Content: "请分析这张图片中的数据趋势",
    Extra: map[string]any{
        "images": []string{"data:image/jpeg;base64,..."},
        "image_urls": []string{"https://example.com/chart.png"},
    },
}
```

---

## ⚙️ 配置选项详解

ChatModel 提供了丰富的配置选项来精确控制模型行为：

### 🎛️ 核心参数

```go
// 🌡️ 温度控制 - 影响输出随机性
model.WithTemperature(0.7)  // 0.0 = 确定性输出, 1.0 = 高随机性

// 📏 最大Token数限制
model.WithMaxTokens(2000)   // 限制生成长度

// 🎯 Top-P 采样
model.WithTopP(0.9)         // 控制输出多样性

// 🛑 停止词设置
model.WithStop([]string{"```", "END"})  // 遇到指定词汇停止生成

// 🎭 指定模型名称
model.WithModel("gpt-4-turbo")  // 选择具体模型
```

### 📊 参数效果对比

| 参数 | 低值效果 | 高值效果 | 推荐场景 |
|------|----------|----------|----------|
| 🌡️ **Temperature** | 输出确定、一致 | 输出多样、创新 | 0.2(分析) / 0.8(创作) |
| 📏 **MaxTokens** | 简洁回答 | 详细回答 | 根据任务需求调整 |
| 🎯 **TopP** | 聚焦高概率词 | 包含低概率词 | 0.9(平衡) / 0.5(聚焦) |

### 🎨 高级配置示例

```go
// 📝 创意写作配置
creativeOptions := []model.Option{
    model.WithTemperature(0.9),    // 高创造性
    model.WithTopP(0.95),          // 高多样性
    model.WithMaxTokens(3000),     // 长篇输出
    model.WithPresencePenalty(0.6), // 减少重复内容
    model.WithFrequencyPenalty(0.3), // 鼓励词汇多样性
}

// 📊 数据分析配置  
analyticalOptions := []model.Option{
    model.WithTemperature(0.2),    // 高一致性
    model.WithTopP(0.8),           // 中等多样性
    model.WithMaxTokens(1500),     // 适中长度
    model.WithStop([]string{"结论:", "总结:"}), // 结构化停止
    model.WithLogitBias(map[string]float64{    // 词汇偏好
        "数据": 0.5,    // 提高"数据"出现概率
        "分析": 0.3,    // 提高"分析"出现概率
    }),
}

// 🤖 代码生成配置
codingOptions := []model.Option{
    model.WithTemperature(0.1),    // 高确定性
    model.WithTopP(0.7),           // 低多样性
    model.WithMaxTokens(2000),     // 代码长度
    model.WithStop([]string{"```", "// END"}), // 代码块停止
    model.WithResponseFormat(&schema.ResponseFormat{
        Type: "json_object",  // 强制JSON格式输出
    }),
}

// 🌐 多语言翻译配置
translationOptions := []model.Option{
    model.WithTemperature(0.3),    // 保持一致性
    model.WithMaxTokens(1000),     // 翻译长度限制
    model.WithTopK(40),            // 限制候选词数量
    model.WithRepetitionPenalty(1.1), // 避免重复翻译
}
```

### 📋 配置参数详解

| 参数名称 | 类型 | 范围 | 默认值 | 作用说明 |
|----------|------|------|--------|----------|
| 🌡️ **Temperature** | float64 | 0.0-2.0 | 1.0 | 控制输出随机性和创造性 |
| 🎯 **TopP** | float64 | 0.0-1.0 | 1.0 | 核采样，控制候选词范围 |
| 🔢 **TopK** | int | 1-100 | 50 | 限制每步采样的候选词数量 |
| 📏 **MaxTokens** | int | 1-8192 | 2048 | 限制生成文本的最大长度 |
| 🛑 **Stop** | []string | - | nil | 遇到指定词汇时停止生成 |
| 📊 **PresencePenalty** | float64 | -2.0-2.0 | 0.0 | 控制话题重复程度 |
| 🔄 **FrequencyPenalty** | float64 | -2.0-2.0 | 0.0 | 控制词汇重复频率 |
| 🎭 **LogitBias** | map[string]float64 | -100-100 | nil | 调整特定词汇出现概率 |
| 🎨 **ResponseFormat** | *schema.ResponseFormat | - | nil | 指定输出格式(JSON/XML) |

---

## 🏗️ 基础使用方法

### 1. 📦 模型初始化

```go
import (
    "context"
    "github.com/cloudwego/eino-ext/components/model/ark"
    "github.com/cloudwego/eino/components/model"
)

// 🚀 ARK模型初始化
func initARKModel(ctx context.Context) (model.ChatModel, error) {
    config := &ark.ChatModelConfig{
        APIKey:      "your-api-key",
        Model:       "doubao-pro-4k", 
        Timeout:     30 * time.Second,
        Temperature: 0.7,
        MaxTokens:   2000,
    }
    
    return ark.NewChatModel(ctx, config)
}

// 🌐 OpenAI模型初始化  
func initOpenAIModel(ctx context.Context) (model.ChatModel, error) {
    config := &openai.ChatModelConfig{
        APIKey: "your-openai-key",
        Model:  "gpt-4-turbo",
        BaseURL: "https://api.openai.com/v1", // 可选
    }
    
    return openai.NewChatModel(ctx, config)
}
```

### 2. 📝 基础对话示例

```go
func basicChatExample() {
    ctx := context.Background()
    
    // 🤖 初始化模型
    chatModel, err := initARKModel(ctx)
    if err != nil {
        log.Fatalf("模型初始化失败: %v", err)
    }
    
    // 📋 构建消息列表
    messages := []*schema.Message{
        {
            Role:    schema.System,
            Content: "你是一个专业的技术顾问，擅长解释复杂的技术概念。",
        },
        {
            Role:    schema.User,
            Content: "请解释什么是微服务架构，以及它的优缺点。",
        },
    }
    
    // 🎯 生成响应
    response, err := chatModel.Generate(ctx, messages,
        model.WithTemperature(0.7),
        model.WithMaxTokens(1500),
    )
    
    if err != nil {
        log.Fatalf("生成响应失败: %v", err)
    }
    
    fmt.Printf("🤖 AI回复: %s\n", response.Content)
}
```

### 3. 💬 多轮对话管理

```go
type ConversationManager struct {
    model    model.ChatModel
    messages []*schema.Message
}

func NewConversationManager(chatModel model.ChatModel, systemPrompt string) *ConversationManager {
    return &ConversationManager{
        model: chatModel,
        messages: []*schema.Message{
            {
                Role:    schema.System,
                Content: systemPrompt,
            },
        },
    }
}

func (cm *ConversationManager) AddUserMessage(content string) {
    cm.messages = append(cm.messages, &schema.Message{
        Role:    schema.User,
        Content: content,
    })
}

func (cm *ConversationManager) GenerateResponse(ctx context.Context) (*schema.Message, error) {
    // 🎯 生成AI回复
    response, err := cm.model.Generate(ctx, cm.messages,
        model.WithTemperature(0.7),
        model.WithMaxTokens(2000),
    )
    
    if err != nil {
        return nil, fmt.Errorf("生成回复失败: %w", err)
    }
    
    // 📝 将AI回复添加到对话历史
    cm.messages = append(cm.messages, response)
    
    return response, nil
}

// 使用示例
func multiTurnChatExample() {
    ctx := context.Background()
    chatModel, _ := initARKModel(ctx)
    
    // 🎭 创建对话管理器
    conversation := NewConversationManager(chatModel, 
        "你是一个友好的AI助手，喜欢用生动的比喻来解释概念。")
    
    // 🔄 进行多轮对话
    questions := []string{
        "什么是机器学习？",
        "那深度学习和机器学习有什么区别？", 
        "能给我一个实际的应用例子吗？",
    }
    
    for i, question := range questions {
        fmt.Printf("👤 用户问题 %d: %s\n", i+1, question)
        
        conversation.AddUserMessage(question)
        response, err := conversation.GenerateResponse(ctx)
        
        if err != nil {
            log.Printf("❌ 生成回复失败: %v", err)
            continue
        }
        
        fmt.Printf("🤖 AI回复 %d: %s\n\n", i+1, response.Content)
    }
}
```

---

## 🌊 流式处理机制

流式处理是现代AI应用的重要特性，能够显著提升用户体验：

### 🚀 流式生成基础

```go
func streamChatExample() {
    ctx := context.Background()
    chatModel, err := initARKModel(ctx)
    if err != nil {
        log.Fatalf("模型初始化失败: %v", err)
    }
    
    messages := []*schema.Message{
        {
            Role:    schema.System,
            Content: "你是一个创意写作助手。",
        },
        {
            Role:    schema.User,
            Content: "请写一个关于AI与人类协作的短故事。",
        },
    }
    
    // 🌊 启动流式生成
    streamReader, err := chatModel.Stream(ctx, messages,
        model.WithTemperature(0.8),
        model.WithMaxTokens(2000),
    )
    
    if err != nil {
        log.Fatalf("启动流式生成失败: %v", err)
    }
    defer streamReader.Close()
    
    // 📖 读取流式输出
    fmt.Print("🤖 AI正在创作: ")
    var fullContent strings.Builder
    
    for {
        chunk, err := streamReader.Recv()
        if err != nil {
            if err == io.EOF {
                break // 🏁 流结束
            }
            log.Printf("❌ 读取流数据失败: %v", err)
            break
        }
        
        // 🖨️ 实时打印内容片段
        fmt.Print(chunk.Content)
        fullContent.WriteString(chunk.Content)
        
        // 🎬 模拟打字机效果
        time.Sleep(50 * time.Millisecond)
    }
    
    fmt.Printf("\n\n📝 完整故事:\n%s\n", fullContent.String())
}
```

### 🔄 高级流式处理模式

流式处理的核心优势在于**实时反馈**和**用户体验优化**。以下是一个完整的流式处理器实现：

```go
type StreamProcessor struct {
    model      model.ChatModel
    onChunk    func(chunk *schema.Message)  // 📝 块处理回调
    onComplete func(fullText string)        // ✅ 完成回调
    onError    func(error)                  // ❌ 错误回调
    onProgress func(current, total int)     // 📊 进度回调
}

func NewStreamProcessor(chatModel model.ChatModel) *StreamProcessor {
    return &StreamProcessor{
        model: chatModel,
    }
}

func (sp *StreamProcessor) SetCallbacks(
    onChunk func(*schema.Message),
    onComplete func(string),
    onError func(error),
) {
    sp.onChunk = onChunk
    sp.onComplete = onComplete  
    sp.onError = onError
}

func (sp *StreamProcessor) ProcessStream(ctx context.Context, messages []*schema.Message, opts ...model.Option) {
    streamReader, err := sp.model.Stream(ctx, messages, opts...)
    if err != nil {
        if sp.onError != nil {
            sp.onError(fmt.Errorf("启动流式处理失败: %w", err))
        }
        return
    }
    defer streamReader.Close()
    
    var fullContent strings.Builder
    
    for {
        chunk, err := streamReader.Recv()
        if err != nil {
            if err == io.EOF {
                // 🎉 流处理完成
                if sp.onComplete != nil {
                    sp.onComplete(fullContent.String())
                }
                break
            }
            
            // ❌ 处理错误
            if sp.onError != nil {
                sp.onError(fmt.Errorf("读取流数据失败: %w", err))
            }
            break
        }
        
        // 📝 累积内容
        fullContent.WriteString(chunk.Content)
        
        // 🔄 触发块处理回调
        if sp.onChunk != nil {
            sp.onChunk(chunk)
        }
    }
}

// 🎬 实际应用示例
func advancedStreamExample() {
    ctx := context.Background()
    chatModel, _ := initARKModel(ctx)
    
    processor := NewStreamProcessor(chatModel)
    
    // 🎯 设置回调函数
    processor.SetCallbacks(
        // 📝 处理每个文本块
        func(chunk *schema.Message) {
            fmt.Print(chunk.Content)
        },
        // ✅ 处理完成
        func(fullText string) {
            fmt.Printf("\n\n📊 统计信息: 总共生成 %d 个字符\n", len(fullText))
        },
        // ❌ 处理错误
        func(err error) {
            log.Printf("❌ 流处理错误: %v", err)
        },
    )
    
    messages := []*schema.Message{
        {Role: schema.User, Content: "请详细解释量子计算的基本原理。"},
    }
    
    // 🚀 开始流式处理
    processor.ProcessStream(ctx, messages,
        model.WithTemperature(0.6),
        model.WithMaxTokens(3000),
    )
}
```

### 🎬 流式处理应用场景

#### 1. 💬 实时聊天界面
```go
type ChatInterface struct {
    processor  *StreamProcessor
    messageBox chan string
}

func (ci *ChatInterface) StartChat() {
    // 模拟UI更新
    ci.processor.SetCallbacks(
        func(chunk *schema.Message) {
            // 逐字显示到聊天界面
            ci.messageBox <- chunk.Content
        },
        func(fullText string) {
            // 显示发送完成状态
            ci.messageBox <- "\n[消息发送完成]"
        },
        func(err error) {
            ci.messageBox <- fmt.Sprintf("\n[错误: %v]", err)
        },
    )
}
```

#### 2. 📊 进度监控系统
```go
type ProgressTracker struct {
    estimatedTokens int
    currentTokens   int
    callback        func(percentage float64)
}

func (pt *ProgressTracker) TrackProgress(chunk *schema.Message) {
    pt.currentTokens += len(chunk.Content) / 3 // 估算token数
    
    percentage := float64(pt.currentTokens) / float64(pt.estimatedTokens) * 100
    if percentage > 100 {
        percentage = 100
    }
    
    if pt.callback != nil {
        pt.callback(percentage)
    }
}
```

---

## 🛠️ 工具调用支持

ChatModel 支持与外部工具集成，扩展AI的能力边界：

### 🔧 工具定义

```go
// 🧮 数学计算工具
func defineCalculatorTool() *schema.ToolInfo {
    return &schema.ToolInfo{
        Name:        "calculator",
        Description: "执行基本的数学计算",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "expression": map[string]any{
                    "type":        "string",
                    "description": "要计算的数学表达式，例如: '2 + 3 * 4'",
                },
            },
            "required": []string{"expression"},
        },
    }
}

// 🌐 天气查询工具
func defineWeatherTool() *schema.ToolInfo {
    return &schema.ToolInfo{
        Name:        "weather",
        Description: "查询指定城市的天气信息",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "city": map[string]any{
                    "type":        "string", 
                    "description": "城市名称，例如: '北京'",
                },
                "unit": map[string]any{
                    "type":        "string",
                    "description": "温度单位",
                    "enum":        []string{"celsius", "fahrenheit"},
                    "default":     "celsius",
                },
            },
            "required": []string{"city"},
        },
    }
}
```

### 🔄 工具调用完整流程

```go
type ToolCallHandler struct {
    model model.ChatModel
    tools map[string]func(args map[string]any) (string, error)
}

func NewToolCallHandler(chatModel model.ChatModel) *ToolCallHandler {
    handler := &ToolCallHandler{
        model: chatModel,
        tools: make(map[string]func(args map[string]any) (string, error)),
    }
    
    // 📝 注册工具实现
    handler.RegisterTool("calculator", func(args map[string]any) (string, error) {
        expression, ok := args["expression"].(string)
        if !ok {
            return "", fmt.Errorf("无效的表达式参数")
        }
        
        // 🧮 简单的表达式计算 (实际应用中应使用专业的表达式解析库)
        result, err := evaluateExpression(expression)
        if err != nil {
            return "", fmt.Errorf("计算失败: %w", err)
        }
        
        return fmt.Sprintf("计算结果: %s = %v", expression, result), nil
    })
    
    handler.RegisterTool("weather", func(args map[string]any) (string, error) {
        city, ok := args["city"].(string)
        if !ok {
            return "", fmt.Errorf("无效的城市参数")
        }
        
        // 🌤️ 模拟天气查询
        return fmt.Sprintf("%s今天天气晴朗，温度25°C，适合外出活动。", city), nil
    })
    
    return handler
}

func (tch *ToolCallHandler) RegisterTool(name string, handler func(map[string]any) (string, error)) {
    tch.tools[name] = handler
}

func (tch *ToolCallHandler) ProcessWithTools(ctx context.Context, userMessage string) (*schema.Message, error) {
    // 🛠️ 绑定工具到模型
    toolModel := tch.model.WithTools([]*schema.ToolInfo{
        defineCalculatorTool(),
        defineWeatherTool(),
    })
    
    messages := []*schema.Message{
        {
            Role:    schema.System,
            Content: "你是一个智能助手，可以使用计算器和天气查询工具来帮助用户。",
        },
        {
            Role:    schema.User,
            Content: userMessage,
        },
    }
    
    // 🎯 生成响应
    response, err := toolModel.Generate(ctx, messages,
        model.WithTemperature(0.3),
    )
    
    if err != nil {
        return nil, fmt.Errorf("生成响应失败: %w", err)
    }
    
    // 🔍 检查是否有工具调用
    if len(response.ToolCalls) > 0 {
        // 📝 处理工具调用
        for _, toolCall := range response.ToolCalls {
            toolResult, err := tch.executeTool(toolCall)
            if err != nil {
                log.Printf("❌ 工具调用失败: %v", err)
                continue
            }
            
            // 🔄 将工具结果添加到对话中
            messages = append(messages, response) // AI的工具调用消息
            messages = append(messages, &schema.Message{
                Role:       schema.Tool,
                Content:    toolResult,
                ToolCallID: toolCall.ID,
            })
        }
        
        // 🎯 基于工具结果生成最终回复
        finalResponse, err := toolModel.Generate(ctx, messages)
        if err != nil {
            return nil, fmt.Errorf("生成最终回复失败: %w", err)
        }
        
        return finalResponse, nil
    }
    
    return response, nil
}

func (tch *ToolCallHandler) executeTool(toolCall schema.ToolCall) (string, error) {
    handler, exists := tch.tools[toolCall.Function.Name]
    if !exists {
        return "", fmt.Errorf("未知工具: %s", toolCall.Function.Name)
    }
    
    // 🔧 解析工具参数
    var args map[string]any
    if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
        return "", fmt.Errorf("解析工具参数失败: %w", err)
    }
    
    // ⚡ 执行工具
    return handler(args)
}

// 📊 使用示例
func toolCallExample() {
    ctx := context.Background()
    chatModel, _ := initARKModel(ctx)
    
    handler := NewToolCallHandler(chatModel)
    
    testQueries := []string{
        "帮我计算 15 * 23 + 45 的结果",
        "查询一下北京今天的天气怎么样",
        "先帮我算一下 100 - 35 * 2，然后查询上海的天气",
    }
    
    for i, query := range testQueries {
        fmt.Printf("👤 问题 %d: %s\n", i+1, query)
        
        response, err := handler.ProcessWithTools(ctx, query)
        if err != nil {
            log.Printf("❌ 处理失败: %v", err)
            continue
        }
        
        fmt.Printf("🤖 回复 %d: %s\n\n", i+1, response.Content)
    }
}
```

---

## 📈 回调机制深入

回调机制是 ChatModel 的强大功能之一，它允许你在模型调用的各个生命周期阶段插入自定义逻辑。

### 🔄 回调生命周期

```
🚀 OnStart  → 🔄 OnChunk  → ✅ OnEnd
     ↓           ↓             ↓
   初始化    → 流式处理   →   完成处理
     ↓           ↓             ↓ 
   ❌ OnError ← ❌ OnError ← ❌ OnError
```

### 🔍 回调接口定义

```go
type ChatModelCallback interface {
    OnStart(ctx context.Context, input *schema.CallbackInput)
    OnEnd(ctx context.Context, output *schema.CallbackOutput) 
    OnError(ctx context.Context, err error)
    OnStreamChunk(ctx context.Context, chunk *schema.Message) // 流式专用
}

// 📊 性能监控回调
type PerformanceMonitor struct {
    startTime   time.Time
    metrics     map[string]any
    chunkCount  int
    totalTokens int
}

func (pm *PerformanceMonitor) OnStart(ctx context.Context, input *schema.CallbackInput) {
    pm.startTime = time.Now()
    pm.metrics = make(map[string]any)
    pm.chunkCount = 0
    pm.totalTokens = 0
    
    fmt.Printf("🚀 开始生成 - 消息数量: %d\n", len(input.Messages))
    
    // 记录请求参数
    pm.metrics["request_time"] = pm.startTime
    pm.metrics["message_count"] = len(input.Messages)
    pm.metrics["model"] = input.Model
    pm.metrics["temperature"] = input.Temperature
}

func (pm *PerformanceMonitor) OnStreamChunk(ctx context.Context, chunk *schema.Message) {
    pm.chunkCount++
    chunkTokens := len(chunk.Content) / 3 // 估算token数
    pm.totalTokens += chunkTokens
    
    fmt.Printf("📝 接收第 %d 个数据块 - 长度: %d 字符, 累计tokens: %d\n", 
        pm.chunkCount, len(chunk.Content), pm.totalTokens)
}

func (pm *PerformanceMonitor) OnEnd(ctx context.Context, output *schema.CallbackOutput) {
    duration := time.Since(pm.startTime)
    
    fmt.Printf("✅ 生成完成 - 耗时: %v\n", duration)
    fmt.Printf("📊 总数据块: %d, 总tokens: %d\n", pm.chunkCount, pm.totalTokens)
    
    if output.Message != nil {
        fmt.Printf("📝 最终输出长度: %d 字符\n", len(output.Message.Content))
        
        // 计算处理速度
        tokensPerSecond := float64(pm.totalTokens) / duration.Seconds()
        fmt.Printf("⚡ 处理速度: %.2f tokens/秒\n", tokensPerSecond)
    }
    
    // 记录完整指标
    pm.metrics["duration"] = duration
    pm.metrics["chunk_count"] = pm.chunkCount 
    pm.metrics["total_tokens"] = pm.totalTokens
    pm.metrics["tokens_per_second"] = float64(pm.totalTokens) / duration.Seconds()
}

func (pm *PerformanceMonitor) OnError(ctx context.Context, err error) {
    duration := time.Since(pm.startTime)
    fmt.Printf("❌ 生成失败 - 耗时: %v, 错误: %v\n", duration, err)
}

// 📝 日志记录回调
type DetailedLogger struct {
    logger *log.Logger
}

func NewDetailedLogger() *DetailedLogger {
    return &DetailedLogger{
        logger: log.New(os.Stdout, "[ChatModel] ", log.LstdFlags),
    }
}

func (dl *DetailedLogger) OnStart(ctx context.Context, input *schema.CallbackInput) {
    dl.logger.Printf("🎯 请求开始 - 模型: %s, 温度: %.2f", 
        input.Model, input.Temperature)
    
    for i, msg := range input.Messages {
        dl.logger.Printf("📝 消息 %d [%s]: %s", 
            i+1, msg.Role, truncateText(msg.Content, 100))
    }
}

func (dl *DetailedLogger) OnEnd(ctx context.Context, output *schema.CallbackOutput) {
    if output.Message != nil {
        dl.logger.Printf("✅ 响应完成 - 内容: %s", 
            truncateText(output.Message.Content, 200))
    }
}

func (dl *DetailedLogger) OnError(ctx context.Context, err error) {
    dl.logger.Printf("❌ 请求失败: %v", err)
}

func truncateText(text string, maxLen int) string {
    if len(text) <= maxLen {
        return text
    }
    return text[:maxLen] + "..."
}
```

### 🎛️ 回调使用示例

```go
func callbackExample() {
    ctx := context.Background()
    chatModel, _ := initARKModel(ctx)
    
    // 📊 创建监控器
    monitor := &PerformanceMonitor{}
    logger := NewDetailedLogger()
    
    messages := []*schema.Message{
        {
            Role:    schema.System,
            Content: "你是一个专业的数据分析师。",
        },
        {
            Role:    schema.User,
            Content: "请分析近年来人工智能技术的发展趋势。",
        },
    }
    
    // 🎯 使用回调生成响应
    response, err := chatModel.Generate(ctx, messages,
        model.WithTemperature(0.7),
        model.WithMaxTokens(2000),
        model.WithCallbacks(monitor, logger), // 📈 添加回调
    )
    
    if err != nil {
        log.Fatalf("生成失败: %v", err)
    }
    
    fmt.Printf("🎉 最终回复: %s\n", response.Content)
}
```

---

## 🌍 多模型平台支持

Eino 支持多种模型平台，提供统一的使用接口：

### 🚀 ARK 平台 (字节跳动)

```go
func initARKModel(ctx context.Context) (model.ChatModel, error) {
    config := &ark.ChatModelConfig{
        APIKey:      os.Getenv("ARK_API_KEY"),
        Model:       "doubao-pro-4k",
        Timeout:     30 * time.Second,
        Temperature: 0.7,
        MaxTokens:   2000,
        TopP:        0.9,
        BaseURL:     "https://ark.cn-beijing.volces.com/api/v3", // 可选
    }
    
    return ark.NewChatModel(ctx, config)
}
```

### 🌐 OpenAI 平台

```go
func initOpenAIModel(ctx context.Context) (model.ChatModel, error) {
    config := &openai.ChatModelConfig{
        APIKey:      os.Getenv("OPENAI_API_KEY"),
        Model:       "gpt-4-turbo",
        BaseURL:     "https://api.openai.com/v1",
        Temperature: 0.7,
        MaxTokens:   2000,
        TopP:        0.9,
        Organization: "your-org-id", // 可选
    }
    
    return openai.NewChatModel(ctx, config)
}
```

### 🏠 Ollama 平台 (本地部署)

```go
func initOllamaModel(ctx context.Context) (model.ChatModel, error) {
    config := &ollama.ChatModelConfig{
        Model:       "llama3:8b",
        BaseURL:     "http://localhost:11434",
        Temperature: 0.7,
        MaxTokens:   2000,
        KeepAlive:   "10m", // 模型保持活跃时间
    }
    
    return ollama.NewChatModel(ctx, config)
}
```

### 🔄 动态模型切换

```go
type MultiModelManager struct {
    models map[string]model.ChatModel
    defaultModel string
}

func NewMultiModelManager() *MultiModelManager {
    return &MultiModelManager{
        models: make(map[string]model.ChatModel),
        defaultModel: "ark",
    }
}

func (mmm *MultiModelManager) RegisterModel(name string, chatModel model.ChatModel) {
    mmm.models[name] = chatModel
}

func (mmm *MultiModelManager) SetDefaultModel(name string) {
    mmm.defaultModel = name
}

func (mmm *MultiModelManager) GetModel(name string) (model.ChatModel, error) {
    if name == "" {
        name = mmm.defaultModel
    }
    
    chatModel, exists := mmm.models[name]
    if !exists {
        return nil, fmt.Errorf("未找到模型: %s", name)
    }
    
    return chatModel, nil
}

func (mmm *MultiModelManager) GenerateWithModel(
    ctx context.Context,
    modelName string,
    messages []*schema.Message,
    opts ...model.Option,
) (*schema.Message, error) {
    chatModel, err := mmm.GetModel(modelName)
    if err != nil {
        return nil, err
    }
    
    return chatModel.Generate(ctx, messages, opts...)
}

// 🔄 使用示例
func multiModelExample() {
    ctx := context.Background()
    
    manager := NewMultiModelManager()
    
    // 📝 注册多个模型
    arkModel, _ := initARKModel(ctx)
    openaiModel, _ := initOpenAIModel(ctx) 
    ollamaModel, _ := initOllamaModel(ctx)
    
    manager.RegisterModel("ark", arkModel)
    manager.RegisterModel("openai", openaiModel)
    manager.RegisterModel("ollama", ollamaModel)
    manager.SetDefaultModel("ark")
    
    messages := []*schema.Message{
        {Role: schema.User, Content: "解释什么是区块链技术？"},
    }
    
    // 🎯 使用不同模型生成回复
    models := []string{"ark", "openai", "ollama"}
    
    for _, modelName := range models {
        fmt.Printf("🤖 使用 %s 模型:\n", modelName)
        
        response, err := manager.GenerateWithModel(ctx, modelName, messages,
            model.WithTemperature(0.7),
        )
        
        if err != nil {
            fmt.Printf("❌ %s 模型生成失败: %v\n\n", modelName, err)
            continue
        }
        
        fmt.Printf("📝 回复: %s\n\n", truncateText(response.Content, 200))
    }
}
```

---

## 🚀 性能优化策略

### 1. 🎯 连接池管理

```go
type ChatModelPool struct {
    models   []model.ChatModel
    current  int
    mutex    sync.Mutex
}

func NewChatModelPool(size int, factory func() (model.ChatModel, error)) (*ChatModelPool, error) {
    pool := &ChatModelPool{
        models: make([]model.ChatModel, size),
    }
    
    // 🔄 预创建模型实例
    for i := 0; i < size; i++ {
        chatModel, err := factory()
        if err != nil {
            return nil, fmt.Errorf("创建模型 %d 失败: %w", i, err)
        }
        pool.models[i] = chatModel
    }
    
    return pool, nil
}

func (cmp *ChatModelPool) GetModel() model.ChatModel {
    cmp.mutex.Lock()
    defer cmp.mutex.Unlock()
    
    chatModel := cmp.models[cmp.current]
    cmp.current = (cmp.current + 1) % len(cmp.models)
    
    return chatModel
}

func (cmp *ChatModelPool) Generate(
    ctx context.Context,
    messages []*schema.Message,
    opts ...model.Option,
) (*schema.Message, error) {
    chatModel := cmp.GetModel()
    return chatModel.Generate(ctx, messages, opts...)
}
```

### 2. 📦 请求批处理

```go
type BatchProcessor struct {
    model      model.ChatModel
    batchSize  int
    timeout    time.Duration
    requests   chan *BatchRequest
    processing sync.WaitGroup
}

type BatchRequest struct {
    Messages []*schema.Message
    Options  []model.Option
    Result   chan *BatchResult
}

type BatchResult struct {
    Response *schema.Message
    Error    error
}

func NewBatchProcessor(chatModel model.ChatModel, batchSize int, timeout time.Duration) *BatchProcessor {
    bp := &BatchProcessor{
        model:     chatModel,
        batchSize: batchSize,
        timeout:   timeout,
        requests:  make(chan *BatchRequest, batchSize*2),
    }
    
    go bp.processRequests()
    return bp
}

func (bp *BatchProcessor) processRequests() {
    batch := make([]*BatchRequest, 0, bp.batchSize)
    ticker := time.NewTicker(bp.timeout)
    defer ticker.Stop()
    
    for {
        select {
        case req := <-bp.requests:
            batch = append(batch, req)
            
            if len(batch) >= bp.batchSize {
                bp.processBatch(batch)
                batch = batch[:0]
                ticker.Reset(bp.timeout)
            }
            
        case <-ticker.C:
            if len(batch) > 0 {
                bp.processBatch(batch)
                batch = batch[:0]
            }
        }
    }
}

func (bp *BatchProcessor) processBatch(batch []*BatchRequest) {
    // 🔄 并行处理批次中的所有请求
    for _, req := range batch {
        bp.processing.Add(1)
        go func(r *BatchRequest) {
            defer bp.processing.Done()
            
            response, err := bp.model.Generate(context.Background(), r.Messages, r.Options...)
            r.Result <- &BatchResult{
                Response: response,
                Error:    err,
            }
            close(r.Result)
        }(req)
    }
}

func (bp *BatchProcessor) Generate(
    ctx context.Context,
    messages []*schema.Message,
    opts ...model.Option,
) (*schema.Message, error) {
    req := &BatchRequest{
        Messages: messages,
        Options:  opts,
        Result:   make(chan *BatchResult, 1),
    }
    
    select {
    case bp.requests <- req:
        // ✅ 请求已提交
    case <-ctx.Done():
        return nil, ctx.Err()
    }
    
    select {
    case result := <-req.Result:
        return result.Response, result.Error
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### 3. 💾 响应缓存

```go
type CachedChatModel struct {
    model model.ChatModel
    cache sync.Map
    ttl   time.Duration
}

type CacheEntry struct {
    Response  *schema.Message
    Timestamp time.Time
}

func NewCachedChatModel(chatModel model.ChatModel, ttl time.Duration) *CachedChatModel {
    return &CachedChatModel{
        model: chatModel,
        ttl:   ttl,
    }
}

func (ccm *CachedChatModel) generateCacheKey(messages []*schema.Message, opts []model.Option) string {
    // 🔑 生成缓存键
    hasher := sha256.New()
    
    for _, msg := range messages {
        hasher.Write([]byte(msg.Role))
        hasher.Write([]byte(msg.Content))
    }
    
    // 📊 包含配置参数
    optStr := fmt.Sprintf("%+v", opts)
    hasher.Write([]byte(optStr))
    
    return fmt.Sprintf("%x", hasher.Sum(nil))
}

func (ccm *CachedChatModel) Generate(
    ctx context.Context,
    messages []*schema.Message,
    opts ...model.Option,
) (*schema.Message, error) {
    cacheKey := ccm.generateCacheKey(messages, opts)
    
    // 🔍 检查缓存
    if cached, exists := ccm.cache.Load(cacheKey); exists {
        entry := cached.(*CacheEntry)
        
        // ⏰ 检查是否过期
        if time.Since(entry.Timestamp) < ccm.ttl {
            return entry.Response, nil
        }
        
        // 🗑️ 删除过期缓存
        ccm.cache.Delete(cacheKey)
    }
    
    // 🎯 生成新响应
    response, err := ccm.model.Generate(ctx, messages, opts...)
    if err != nil {
        return nil, err
    }
    
    // 💾 存储到缓存
    ccm.cache.Store(cacheKey, &CacheEntry{
        Response:  response,
        Timestamp: time.Now(),
    })
    
    return response, nil
}

// 🧹 清理过期缓存
func (ccm *CachedChatModel) CleanupExpiredEntries() {
    now := time.Now()
    ccm.cache.Range(func(key, value any) bool {
        entry := value.(*CacheEntry)
        if now.Sub(entry.Timestamp) > ccm.ttl {
            ccm.cache.Delete(key)
        }
        return true
    })
}
```

---

## ❌ 错误处理最佳实践

### 🛡️ 重试机制

```go
type RetryableChatModel struct {
    model      model.ChatModel
    maxRetries int
    backoff    time.Duration
}

func NewRetryableChatModel(chatModel model.ChatModel, maxRetries int, backoff time.Duration) *RetryableChatModel {
    return &RetryableChatModel{
        model:      chatModel,
        maxRetries: maxRetries,
        backoff:    backoff,
    }
}

func (rcm *RetryableChatModel) Generate(
    ctx context.Context,
    messages []*schema.Message,
    opts ...model.Option,
) (*schema.Message, error) {
    var lastErr error
    
    for attempt := 0; attempt <= rcm.maxRetries; attempt++ {
        if attempt > 0 {
            // 📈 指数退避
            backoffTime := time.Duration(attempt) * rcm.backoff
            
            select {
            case <-time.After(backoffTime):
                // 继续重试
            case <-ctx.Done():
                return nil, ctx.Err()
            }
            
            log.Printf("🔄 重试第 %d 次生成请求", attempt)
        }
        
        response, err := rcm.model.Generate(ctx, messages, opts...)
        if err == nil {
            if attempt > 0 {
                log.Printf("✅ 重试第 %d 次成功", attempt)
            }
            return response, nil
        }
        
        lastErr = err
        
        // 🔍 检查是否为可重试错误
        if !isRetryableError(err) {
            log.Printf("❌ 不可重试错误: %v", err)
            break
        }
        
        log.Printf("⚠️ 第 %d 次尝试失败: %v", attempt+1, err)
    }
    
    return nil, fmt.Errorf("达到最大重试次数 %d，最后错误: %w", rcm.maxRetries, lastErr)
}

func isRetryableError(err error) bool {
    if err == nil {
        return false
    }
    
    errStr := err.Error()
    
    // 🌐 网络相关错误
    retryableErrors := []string{
        "connection timeout",
        "network unreachable", 
        "service unavailable",
        "rate limit exceeded",
        "502 bad gateway",
        "503 service unavailable",
        "504 gateway timeout",
    }
    
    for _, retryable := range retryableErrors {
        if strings.Contains(strings.ToLower(errStr), retryable) {
            return true
        }
    }
    
    return false
}
```

### 🔧 降级处理

```go
type FallbackChatModel struct {
    primary   model.ChatModel
    secondary model.ChatModel
    fallbackThreshold time.Duration
}

func NewFallbackChatModel(primary, secondary model.ChatModel, threshold time.Duration) *FallbackChatModel {
    return &FallbackChatModel{
        primary:           primary,
        secondary:         secondary,
        fallbackThreshold: threshold,
    }
}

func (fcm *FallbackChatModel) Generate(
    ctx context.Context,
    messages []*schema.Message,
    opts ...model.Option,
) (*schema.Message, error) {
    // 🎯 尝试主模型
    startTime := time.Now()
    
    // ⏰ 设置主模型超时
    primaryCtx, cancel := context.WithTimeout(ctx, fcm.fallbackThreshold)
    defer cancel()
    
    response, err := fcm.primary.Generate(primaryCtx, messages, opts...)
    if err == nil {
        duration := time.Since(startTime)
        log.Printf("✅ 主模型响应成功 - 耗时: %v", duration)
        return response, nil
    }
    
    // ⚠️ 主模型失败，尝试备用模型
    log.Printf("⚠️ 主模型失败: %v，切换到备用模型", err)
    
    response, fallbackErr := fcm.secondary.Generate(ctx, messages, opts...)
    if fallbackErr == nil {
        duration := time.Since(startTime)
        log.Printf("✅ 备用模型响应成功 - 耗时: %v", duration)
        return response, nil
    }
    
    // ❌ 两个模型都失败了
    return nil, fmt.Errorf("主模型和备用模型都失败了 - 主模型错误: %w, 备用模型错误: %v", err, fallbackErr)
}
```

---

## 🎓 高级用法和技巧

### 1. 🎭 对话角色管理

```go
type RoleBasedChatModel struct {
    model       model.ChatModel
    rolePrompts map[string]string
    defaultRole string
}

func NewRoleBasedChatModel(chatModel model.ChatModel) *RoleBasedChatModel {
    rbcm := &RoleBasedChatModel{
        model:       chatModel,
        rolePrompts: make(map[string]string),
        defaultRole: "assistant",
    }
    
    // 🎭 预设角色
    rbcm.AddRole("技术专家", "你是一个资深的技术专家，擅长解释复杂的技术概念，并提供实用的解决方案。")
    rbcm.AddRole("创意写手", "你是一个富有创意的写手，善于创作引人入胜的故事和文案。")
    rbcm.AddRole("数据分析师", "你是一个专业的数据分析师，能够深入分析数据并提供有价值的洞察。")
    rbcm.AddRole("产品经理", "你是一个经验丰富的产品经理，能够从用户需求出发思考产品设计。")
    
    return rbcm
}

func (rbcm *RoleBasedChatModel) AddRole(roleName, prompt string) {
    rbcm.rolePrompts[roleName] = prompt
}

func (rbcm *RoleBasedChatModel) GenerateWithRole(
    ctx context.Context,
    roleName string,
    userMessage string,
    opts ...model.Option,
) (*schema.Message, error) {
    systemPrompt, exists := rbcm.rolePrompts[roleName]
    if !exists {
        systemPrompt = rbcm.rolePrompts[rbcm.defaultRole]
    }
    
    messages := []*schema.Message{
        {
            Role:    schema.System,
            Content: systemPrompt,
        },
        {
            Role:    schema.User,
            Content: userMessage,
        },
    }
    
    return rbcm.model.Generate(ctx, messages, opts...)
}

// 🎯 使用示例
func roleBasedExample() {
    ctx := context.Background()
    chatModel, _ := initARKModel(ctx)
    
    roleModel := NewRoleBasedChatModel(chatModel)
    
    question := "如何提高团队的工作效率？"
    roles := []string{"技术专家", "产品经理", "数据分析师"}
    
    for _, role := range roles {
        fmt.Printf("🎭 %s 的建议:\n", role)
        
        response, err := roleModel.GenerateWithRole(ctx, role, question,
            model.WithTemperature(0.7),
        )
        
        if err != nil {
            log.Printf("❌ 生成失败: %v", err)
            continue
        }
        
        fmt.Printf("💬 %s\n\n", response.Content)
    }
}
```

### 2. 📊 上下文窗口管理

```go
type ContextWindowManager struct {
    model         model.ChatModel
    maxTokens     int
    reserveTokens int // 为响应保留的token数
}

func NewContextWindowManager(chatModel model.ChatModel, maxTokens int) *ContextWindowManager {
    return &ContextWindowManager{
        model:         chatModel,
        maxTokens:     maxTokens,
        reserveTokens: maxTokens / 4, // 为响应保留25%的空间
    }
}

func (cwm *ContextWindowManager) estimateTokens(text string) int {
    // 🔢 简化的token估算 (实际应用中应使用专业的tokenizer)
    return len(text) / 3 // 粗略估算：平均每个token约3个字符
}

func (cwm *ContextWindowManager) trimMessages(messages []*schema.Message) []*schema.Message {
    if len(messages) == 0 {
        return messages
    }
    
    availableTokens := cwm.maxTokens - cwm.reserveTokens
    currentTokens := 0
    
    // 📝 从后往前累计token数（保持最近的对话）
    var result []*schema.Message
    
    // 🎯 始终保留系统消息
    systemMessages := make([]*schema.Message, 0)
    otherMessages := make([]*schema.Message, 0)
    
    for _, msg := range messages {
        if msg.Role == schema.System {
            systemMessages = append(systemMessages, msg)
        } else {
            otherMessages = append(otherMessages, msg)
        }
    }
    
    // 📊 计算系统消息的token
    for _, msg := range systemMessages {
        currentTokens += cwm.estimateTokens(msg.Content)
    }
    
    result = append(result, systemMessages...)
    
    // 🔄 从最新消息开始向前添加
    for i := len(otherMessages) - 1; i >= 0; i-- {
        msg := otherMessages[i]
        msgTokens := cwm.estimateTokens(msg.Content)
        
        if currentTokens+msgTokens > availableTokens {
            break
        }
        
        result = append([]*schema.Message{msg}, result...)
        currentTokens += msgTokens
    }
    
    log.Printf("📊 上下文管理: 估算使用 %d/%d tokens", currentTokens, availableTokens)
    
    return result
}

func (cwm *ContextWindowManager) Generate(
    ctx context.Context,
    messages []*schema.Message,
    opts ...model.Option,
) (*schema.Message, error) {
    trimmedMessages := cwm.trimMessages(messages)
    return cwm.model.Generate(ctx, trimmedMessages, opts...)
}
```

---

## ❓ 常见问题解决

### Q1: 如何处理模型响应超时？

**解决方案**: 使用上下文超时控制
```go
func generateWithTimeout(chatModel model.ChatModel, messages []*schema.Message, timeout time.Duration) (*schema.Message, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    response, err := chatModel.Generate(ctx, messages)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return nil, fmt.Errorf("⏰ 生成超时 (%.1fs)", timeout.Seconds())
        }
        return nil, err
    }
    
    return response, nil
}
```

### Q2: 如何处理API配额限制？

**解决方案**: 实现智能限流
```go
type RateLimitedChatModel struct {
    model     model.ChatModel
    limiter   *rate.Limiter
    waitTime  time.Duration
}

func NewRateLimitedChatModel(chatModel model.ChatModel, requestsPerSecond float64) *RateLimitedChatModel {
    return &RateLimitedChatModel{
        model:    chatModel,
        limiter:  rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
        waitTime: time.Minute,
    }
}

func (rlcm *RateLimitedChatModel) Generate(
    ctx context.Context,
    messages []*schema.Message,
    opts ...model.Option,
) (*schema.Message, error) {
    // ⏳ 等待限流器允许
    waitCtx, cancel := context.WithTimeout(ctx, rlcm.waitTime)
    defer cancel()
    
    if err := rlcm.limiter.Wait(waitCtx); err != nil {
        return nil, fmt.Errorf("🚫 限流等待超时: %w", err)
    }
    
    return rlcm.model.Generate(ctx, messages, opts...)
}
```

### Q3: 如何优化长文本处理性能？

**解决方案**: 文本分块处理
```go
type ChunkedProcessor struct {
    model       model.ChatModel
    chunkSize   int
    overlapSize int
}

func NewChunkedProcessor(chatModel model.ChatModel, chunkSize, overlapSize int) *ChunkedProcessor {
    return &ChunkedProcessor{
        model:       chatModel,
        chunkSize:   chunkSize,
        overlapSize: overlapSize,
    }
}

func (cp *ChunkedProcessor) ProcessLongText(
    ctx context.Context,
    systemPrompt string,
    longText string,
    task string,
) ([]string, error) {
    chunks := cp.splitText(longText)
    results := make([]string, 0, len(chunks))
    
    for i, chunk := range chunks {
        messages := []*schema.Message{
            {
                Role:    schema.System,
                Content: systemPrompt,
            },
            {
                Role:    schema.User,
                Content: fmt.Sprintf("文本块 %d/%d:\n%s\n\n任务: %s", i+1, len(chunks), chunk, task),
            },
        }
        
        response, err := cp.model.Generate(ctx, messages)
        if err != nil {
            return nil, fmt.Errorf("处理文本块 %d 失败: %w", i+1, err)
        }
        
        results = append(results, response.Content)
    }
    
    return results, nil
}

func (cp *ChunkedProcessor) splitText(text string) []string {
    if len(text) <= cp.chunkSize {
        return []string{text}
    }
    
    var chunks []string
    start := 0
    
    for start < len(text) {
        end := start + cp.chunkSize
        if end > len(text) {
            end = len(text)
        }
        
        chunk := text[start:end]
        chunks = append(chunks, chunk)
        
        start = end - cp.overlapSize
        if start <= 0 {
            start = end
        }
    }
    
    return chunks
}
```

---

## 🎉 总结

ChatModel 是 Eino 框架的**核心组件**，掌握其使用是构建高质量 AI 应用的关键：

### 🏆 核心价值
- 🔧 **标准化**: 统一的接口抽象不同模型平台
- 🚀 **高性能**: 流式处理和异步机制提升体验
- 🛠️ **扩展性**: 工具调用和回调机制支持复杂需求
- 🔒 **可靠性**: 完善的错误处理和重试机制

### 💡 最佳实践总结
1. **合理选择交互模式**: Generate适合简单任务，Stream适合长文本生成
2. **充分利用工具调用**: 扩展AI能力边界，处理复杂任务
3. **实现完善的错误处理**: 重试机制、降级处理确保系统稳定
4. **注重性能优化**: 连接池、批处理、缓存提升系统性能
5. **监控和日志**: 使用回调机制监控系统运行状态

### 🔗 相关资源
- 📚 [官方文档](https://www.cloudwego.io/zh/docs/eino/core_modules/components/chat_model_guide/)
- 💻 [示例代码](./model.go)
- 🌐 [GitHub 仓库](https://github.com/cloudwego/eino)

通过掌握 ChatModel 的各种特性和最佳实践，你将能够构建出高性能、可靠且功能强大的 AI 应用！🚀