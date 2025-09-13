# 🎨 Eino ChatTemplate 组件完全指南

本文档是对 Eino 框架中 `ChatTemplate` 组件的核心功能和使用方式的完整总结，结合官方文档和实际项目示例。

## 🚀 快速开始

### 配置环境
```bash
# 如需要使用 Chain 编排示例，请设置 API Key
export ARK_API_KEY="your-ark-api-key"

# 构建项目
go build -o chattemplate_demo main.go
```

### 运行示例
```bash
# 运行所有示例
./chattemplate_demo

# 运行特定示例
./chattemplate_demo basic        # 基础模板格式化
./chattemplate_demo complex      # 复杂模板处理
./chattemplate_demo conditional  # 条件模板逻辑
./chattemplate_demo jinja2       # Jinja2 高级模板 ⭐ 新增
./chattemplate_demo performance  # 性能基准测试
./chattemplate_demo chain        # Chain 编排集成
```

### 配置文件
项目使用 `config.yaml` 配置文件，也可以通过环境变量设置：
```yaml
ARK_API_KEY: "${ARK_API_KEY}"
ARK_MODEL: "doubao-pro-4k"
BASE_URL: "https://ark.cn-beijing.volces.com/api/v3"
```

---

## 📖 基本介绍

`ChatTemplate` 组件是一个专门用于**处理和格式化提示（Prompt）**的智能组件。它的主要作用是将包含**变量占位符**的模板与用户提供的**具体值**相结合，生成符合 Eino 标准的消息列表。这个组件在 AI 应用开发中扮演着**"智能模板引擎"**的角色。

### 🎯 核心价值

在传统的模板处理中，我们只能进行简单的字符串替换。而 ChatTemplate 组件让我们能够：

```
传统模板：简单字符串替换 + 静态内容  ❌
ChatTemplate：智能变量替换 + 动态消息构建 + 多格式支持  ✅
```

### 🚀 主要应用场景

- **🎭 角色定义模板**: 构建结构化的系统提示和角色设定
- **📚 任务描述模板**: 动态生成用户指令和任务说明
- **💬 多轮对话管理**: 处理对话历史和上下文连接
- **🔄 提示模式复用**: 模板封装和参数化配置
- **🧩 组件协作**: 与其他 Eino 组件无缝集成构建完整工作流
- **⚡ 动态内容插入**: 根据不同场景动态调整模板内容

---

## 🔧 核心接口

`ChatTemplate` 组件提供了简洁而强大的接口设计：

### 基础接口

```go
type ChatTemplate interface {
    Format(ctx context.Context, vs map[string]any, opts ...Option) ([]*schema.Message, error)
}
```

### 接口详解

#### 📝 Format 方法
- **功能**: 将模板格式化为消息列表
- **输入**:
    - `ctx`: 上下文对象，用于控制超时、取消等
    - `vs`: 模板变量映射表 (`map[string]any`)
    - `opts`: 可选配置参数
- **输出**:
    - `[]*schema.Message`: 格式化后的消息列表
    - `error`: 格式化过程中的错误信息

---

## 📨 Message 结构体

`Message` 是模板生成的基本数据结构，支持丰富的消息类型：

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

---

## 🎯 模板格式详解

ChatTemplate 支持三种不同的模板格式，适应不同复杂度的需求：

### 1. 🎯 FString 格式 (推荐)

**最常用的格式**，语法简单直观，适合大多数场景。

```go
// 基础语法
template := prompt.FromMessages(schema.FString,
    schema.SystemMessage("你是一个{role}。"),
    schema.UserMessage("请帮我{task}。")
)

variables := map[string]any{
    "role": "专业的写作助手",  
    "task": "写一篇关于AI的文章",
}
```

**使用场景**: 
- 🎭 简单变量替换
- 📝 基础模板构建
- 🚀 快速原型开发

### 2. 🔧 GoTemplate 格式 (高级)

支持 Go 标准库模板语法，可实现复杂逻辑。

```go
template := prompt.FromMessages(schema.GoTemplate,
    schema.SystemMessage(`你是一个{{.role}}。
{{if .expert_mode}}你需要提供专业级别的详细解答。{{end}}`),
    schema.UserMessage("任务: {{.task}}")
)

variables := map[string]any{
    "role": "AI专家",
    "task": "解释深度学习原理",
    "expert_mode": true,
}
```

**高级特性**:
- 🔄 条件判断 (`{{if .condition}}`)
- 🔁 循环处理 (`{{range .items}}`)
- 🧮 函数调用和管道操作
- 📊 复杂数据结构处理

### 3. 🌟 Jinja2 格式 (专业)

支持 Jinja2 模板语法，功能最强大。

#### 🎯 基础示例
```go
template := prompt.FromMessages(schema.Jinja2,
    schema.SystemMessage("你是{{ role }}{% if specialization %}，专长于{{ specialization }}{% endif %}。"),
    schema.UserMessage("{% for item in tasks %}任务{{ loop.index }}: {{ item }}{% if not loop.last %}\n{% endif %}{% endfor %}")
)
```

#### 🚀 完整演示案例

本项目的 `jinja2TemplateExample()` 函数展示了 Jinja2 的完整功能：

```go
template := prompt.FromMessages(schema.Jinja2,
    schema.SystemMessage(`你是{{ role }}{% if expertise %}，专长于{{ expertise }}{% endif %}。

你的服务特点:
{% for feature in service_features %}
- {{ feature }}
{% endfor %}

{% if strict_mode %}请严格按照专业标准回答。{% else %}请以友好的方式回答。{% endif %}`),
    schema.MessagesPlaceholder("conversation_history", true),
    schema.UserMessage(`当前任务类型: {{ task_type }}

具体要求:
{% for requirement in requirements %}
{{ loop.index }}. {{ requirement }}
{% endfor %}

用户问题: {{ user_question }}`),
)
```

**运行演示**:
```bash
go run main.go jinja2
```

**专业特性**:
- 🎨 丰富的过滤器系统
- 🧩 模板继承和包含
- 🔧 自定义函数注册
- 📚 完整的 Python Jinja2 兼容
- ✅ 条件判断: `{% if %}...{% endif %}`
- 🔄 循环处理: `{% for %}...{% endfor %}`
- 📊 循环计数: `{{ loop.index }}`
- 🎯 变量插值: `{{ variable_name }}`

---

## 📍 特殊占位符

### MessagesPlaceholder

**MessagesPlaceholder** 是处理对话历史的关键组件：

```go
schema.MessagesPlaceholder("history_key", optional)
```

| 参数 | 说明 |
|------|------|
| `"history_key"` | 🔑 变量映射表中的键名 |
| `optional` | 🔄 是否可选（true=可缺失，false=必须提供） |

**实际应用示例**:
```go
template := prompt.FromMessages(schema.FString,
    schema.SystemMessage("你是智能助手。"),
    schema.MessagesPlaceholder("chat_history", true),  // 可选的对话历史
    schema.UserMessage("新问题: {question}")
)

variables := map[string]any{
    "question": "什么是机器学习？",
    "chat_history": []*schema.Message{
        {Role: schema.User, Content: "你好"},
        {Role: schema.Assistant, Content: "你好！有什么可以帮你的吗？"},
    },
}
```

---

## 🏗️ 创建和使用模板

### 基础创建流程

```go
import (
    "github.com/cloudwego/eino/components/prompt"
    "github.com/cloudwego/eino/schema"
)

// 1️⃣ 创建模板
template := prompt.FromMessages(schema.FString,
    schema.SystemMessage("你是一个{role}。"),
    schema.MessagesPlaceholder("history", false),
    schema.UserMessage("请完成任务: {task}")
)

// 2️⃣ 准备变量
variables := map[string]any{
    "role": "专业助手",
    "task": "写一首诗",
    "history": []*schema.Message{
        // 对话历史...
    },
}

// 3️⃣ 格式化模板
messages, err := template.Format(context.Background(), variables)
if err != nil {
    log.Fatalf("模板格式化失败: %v", err)
}

// 4️⃣ 使用结果
for _, msg := range messages {
    fmt.Printf("角色: %s, 内容: %s\n", msg.Role, msg.Content)
}
```

### 🎯 实用模板示例

#### 📚 知识问答模板
```go
qaTemplate := prompt.FromMessages(schema.FString,
    schema.SystemMessage(`你是一个知识渊博的{domain}专家。
回答特点:
- 准确性: 提供准确可靠的信息
- 清晰性: 用简洁明了的语言解释
- 实用性: 包含实际应用建议
当前专业领域: {domain}`),
    schema.MessagesPlaceholder("conversation_history", true),
    schema.UserMessage("问题: {question}\n\n请提供详细的专业解答。")
)
```

#### 🎭 角色扮演模板  
```go
rolePlayTemplate := prompt.FromMessages(schema.FString,
    schema.SystemMessage(`请扮演{character}这个角色。

角色背景: {background}
性格特点: {personality}
说话风格: {speaking_style}

请始终保持角色一致性，用符合角色身份的方式回应。`),
    schema.UserMessage("{user_input}")
)
```

#### 🔄 任务处理模板
```go
taskTemplate := prompt.FromMessages(schema.GoTemplate,
    schema.SystemMessage(`你是任务处理助手。当前任务类型: {{.task_type}}

处理步骤:
{{range $i, $step := .steps}}
{{add $i 1}}. {{$step}}
{{end}}

质量要求: {{.quality_requirements}}`),
    schema.UserMessage("具体任务: {{.specific_task}}")
)
```

---

## 🚀 编排集成最佳实践

虽然可以直接使用 ChatTemplate，但**官方强烈推荐**将其集成到编排工作流中，与其他组件协同工作。

### 🔗 Chain 编排模式

Chain 是最常用的编排方式，适合线性处理流程：

```go
import "github.com/cloudwego/eino/compose"

// 1️⃣ 创建 Chain - 声明输入输出类型  
chain := compose.NewChain[map[string]any, *schema.Message]()

// 2️⃣ 添加组件 - 按处理顺序添加
chain.AppendChatTemplate(template)  // 模板 → 消息列表
chain.AppendChatModel(model)        // 消息列表 → AI回复

// 3️⃣ 编译执行
runnable, err := chain.Compile(ctx)
if err != nil {
    log.Fatalf("链编译失败: %v", err)
}

// 4️⃣ 运行工作流
result, err := runnable.Invoke(ctx, variables)
```

### 🔄 完整工作流示例

```go
func createAdvancedChatWorkflow() (*compose.Runnable, error) {
    ctx := context.Background()
    
    // 📝 创建高级模板
    template := prompt.FromMessages(schema.FString,
        schema.SystemMessage(`你是{assistant_type}，具有以下特长:
专业领域: {expertise_area}  
服务风格: {service_style}
回答深度: {response_depth}`),
        schema.MessagesPlaceholder("chat_history", true),
        schema.UserMessage(`用户问题: {user_question}
期望输出格式: {output_format}`)
    )
    
    // 🤖 初始化模型
    model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
        APIKey: viper.GetString("ARK_API_KEY"),
        Model:  "doubao-pro-4k",
        Temperature: 0.7,
    })
    if err != nil {
        return nil, fmt.Errorf("模型初始化失败: %w", err)
    }
    
    // 🔗 构建处理链
    chain := compose.NewChain[map[string]any, *schema.Message]()
    chain.AppendChatTemplate(template)
    chain.AppendChatModel(model)
    
    // ⚙️ 编译成可运行实例
    return chain.Compile(ctx)
}

// 使用示例
func runWorkflow() {
    runnable, err := createAdvancedChatWorkflow()
    if err != nil {
        log.Fatalf("工作流创建失败: %v", err)
    }
    
    variables := map[string]any{
        "assistant_type": "技术顾问",
        "expertise_area": "云计算和微服务架构",  
        "service_style": "专业严谨，注重实践",
        "response_depth": "深入分析，提供具体方案",
        "user_question": "如何设计高可用的微服务架构？",
        "output_format": "结构化回答，包含方案和实施步骤",
        "chat_history": []*schema.Message{
            {Role: schema.User, Content: "你好，我需要架构咨询"},
            {Role: schema.Assistant, Content: "你好！我是专业的技术顾问，很高兴为您提供架构咨询服务。"},
        },
    }
    
    result, err := runnable.Invoke(context.Background(), variables)
    if err != nil {
        log.Fatalf("工作流执行失败: %v", err)
    }
    
    fmt.Println("📋 AI回复:", result.Content)
}
```

---

## 🎓 高级用法和技巧

### 1. 📊 动态模板选择

根据不同场景动态选择合适的模板：

```go
type TemplateManager struct {
    templates map[string]*prompt.ChatTemplate
}

func (tm *TemplateManager) GetTemplate(templateType string) *prompt.ChatTemplate {
    switch templateType {
    case "qa":
        return tm.templates["knowledge_qa"]
    case "creative":  
        return tm.templates["creative_writing"]
    case "analysis":
        return tm.templates["data_analysis"] 
    default:
        return tm.templates["general_chat"]
    }
}
```

### 2. 🔄 模板组合模式

将多个小模板组合成复杂模板：

```go
func createComposedTemplate() *prompt.ChatTemplate {
    // 基础系统设定
    systemBase := "你是一个专业的{role}助手。"
    
    // 专业领域扩展
    domainExtension := `
专业领域: {domain}
核心技能: {skills}
服务标准: {service_standards}`
    
    // 输出格式要求
    formatRequirement := `
输出要求:
- 结构化回答
- 包含实例说明  
- 提供可行建议`
    
    return prompt.FromMessages(schema.FString,
        schema.SystemMessage(systemBase + domainExtension + formatRequirement),
        schema.MessagesPlaceholder("context_messages", true),
        schema.UserMessage("{user_request}")
    )
}
```

### 3. 📈 性能优化技巧

#### 模板缓存
```go
type CachedTemplateManager struct {
    cache sync.Map
}

func (ctm *CachedTemplateManager) GetOrCreateTemplate(key string, factory func() *prompt.ChatTemplate) *prompt.ChatTemplate {
    if cached, exists := ctm.cache.Load(key); exists {
        return cached.(*prompt.ChatTemplate)
    }
    
    template := factory()
    ctm.cache.Store(key, template)
    return template
}
```

#### 变量预处理
```go
func preprocessVariables(raw map[string]any) map[string]any {
    processed := make(map[string]any)
    
    for k, v := range raw {
        switch val := v.(type) {
        case string:
            processed[k] = strings.TrimSpace(val)  // 去除空白字符
        case []string:
            processed[k] = strings.Join(val, ", ") // 数组转字符串
        default:
            processed[k] = val
        }
    }
    
    return processed
}
```

---

## ❓ 常见问题和解决方案

### Q1: 模板变量缺失怎么办？

**问题**: 运行时提示变量未找到
```go
// ❌ 错误示例
variables := map[string]any{
    "role": "助手",
    // 缺少 "task" 变量
}
```

**解决方案**: 
```go
// ✅ 完整变量检查
func validateVariables(template *prompt.ChatTemplate, variables map[string]any) error {
    requiredVars := []string{"role", "task", "context"}
    
    for _, required := range requiredVars {
        if _, exists := variables[required]; !exists {
            return fmt.Errorf("缺少必需变量: %s", required)
        }
    }
    return nil
}
```

### Q2: MessagesPlaceholder 使用注意事项

**问题**: 对话历史格式不正确
```go
// ❌ 错误格式
"history": []string{"用户消息", "助手回复"}
```

**解决方案**:
```go
// ✅ 正确格式  
"history": []*schema.Message{
    {Role: schema.User, Content: "用户消息"},
    {Role: schema.Assistant, Content: "助手回复"},
}
```

### Q3: 模板格式化失败处理

```go
func safeFormatTemplate(template *prompt.ChatTemplate, variables map[string]any) ([]*schema.Message, error) {
    // 🛡️ 添加容错处理
    defer func() {
        if r := recover(); r != nil {
            log.Printf("模板格式化 panic: %v", r)
        }
    }()
    
    // 📋 变量验证
    if err := validateVariables(template, variables); err != nil {
        return nil, fmt.Errorf("变量验证失败: %w", err)
    }
    
    // 🎯 执行格式化
    return template.Format(context.Background(), variables)
}
```

### Q4: 性能优化建议

```go
// 🚀 高性能模板使用模式
type OptimizedTemplateRunner struct {
    templateCache sync.Map
    variablePool  sync.Pool
}

func (otr *OptimizedTemplateRunner) RunTemplate(templateKey string, vars map[string]any) (*schema.Message, error) {
    // 1. 从缓存获取模板
    template := otr.getTemplateFromCache(templateKey)
    
    // 2. 从对象池获取变量容器
    variables := otr.variablePool.Get().(map[string]any)
    defer func() {
        // 清空并归还到池中
        for k := range variables {
            delete(variables, k)
        }
        otr.variablePool.Put(variables)
    }()
    
    // 3. 复制变量避免并发问题
    for k, v := range vars {
        variables[k] = v
    }
    
    // 4. 执行模板格式化
    messages, err := template.Format(context.Background(), variables)
    if err != nil {
        return nil, err
    }
    
    // 5. 假设后续还会调用模型，这里简化返回第一条消息
    return messages[0], nil
}
```

---

## 🎉 总结

ChatTemplate 是 Eino 框架中的**核心基础组件**，掌握它的使用对于构建高质量的 AI 应用至关重要：

### 🏆 核心优势
- 🎨 **灵活性**: 多种模板格式支持不同复杂度需求
- 🔄 **可复用性**: 模板可在多个场景中重复使用  
- 🧩 **组件化**: 与 Eino 生态系统深度集成
- 🚀 **高性能**: 优化的模板引擎保证快速响应
- 🛡️ **类型安全**: 编译时类型检查避免运行时错误

### 💡 最佳实践总结
1. **优先使用编排**: 将 ChatTemplate 与 Chain/Graph 结合使用
2. **合理选择格式**: FString 适合简单场景，GoTemplate 适合复杂逻辑，Jinja2 适合专业级应用
3. **注重性能优化**: 使用模板缓存和变量预处理
4. **完善错误处理**: 添加变量验证和异常捕获
5. **遵循规范**: 严格按照 schema.Message 格式处理消息
6. **学习完整示例**: 运行 `go run main.go jinja2` 体验 Jinja2 高级特性

### 🔗 相关资源
- 📚 [官方文档](https://www.cloudwego.io/zh/docs/eino/core_modules/components/chat_template_guide/)
- 💻 [示例代码](./main.go)
- 🎯 [项目文档](./README.md)
- 🌐 [GitHub 仓库](https://github.com/cloudwego/eino)

### 🚀 快速体验
```bash
# 基础模板演示
go run main.go basic

# 复杂模板演示  
go run main.go complex

# 条件模板演示
go run main.go conditional

# Jinja2 高级模板演示 ⭐ 新增
go run main.go jinja2

# 性能测试演示
go run main.go performance

# Chain 编排演示（需要 API Key）
go run main.go chain

# 运行所有演示
go run main.go
```

通过掌握 ChatTemplate 的各种格式（FString、GoTemplate、Jinja2），你将能够构建出更加智能、灵活和可维护的 AI 应用！🚀