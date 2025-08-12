# Eino ChatTemplate 核心要点总结

本文档是对 Eino 框架中 `ChatTemplate` 组件的核心功能和使用方式的总结。

---

## 1. 核心功能

`ChatTemplate` 是一个用于**处理和格式化提示（Prompt）**的强大组件。其主要作用是将包含**变量占位符**的模板，与用户提供的**具体值**相结合，最终生成可供大语言模型（LLM）使用的标准消息列表 (`[]*schema.Message`)。

**主要应用场景:**
- **构建结构化提示**: 创建包含动态内容（如角色、任务描述）的系统或用户提示。
- **处理多轮对话**: 通过 `MessagesPlaceholder` 轻松地将对话历史插入到提示中。
- **实现提示模式复用**: 将常用的提示结构定义为模板，方便在不同地方重复使用。

---

## 2. 核心接口

`ChatTemplate` 的核心是一个简单的接口：

```go
type ChatTemplate interface {
    Format(ctx context.Context, vs map[string]any, opts ...Option) ([]*schema.Message, error)
}
```

- **`Format` 方法**: 这是唯一需要调用的方法，它接收一个 `map[string]any` 类型的变量，并返回格式化后的消息列表。

---

## 3. 创建 ChatTemplate

创建模板最常用的方法是 `prompt.FromMessages()`。它接收一个**模板格式**和一系列**消息模板**作为参数。

### 3.1. 支持的模板格式

- **`schema.FString`**: 简单直观的 f-string 格式，使用 `{variable}` 作为占位符。最常用。
- **`schema.GoTemplate`**: 支持 Go 标准库的 `text/template` 语法，可实现条件、循环等复杂逻辑。
- **`schema.Jinja2`**: 支持 Jinja2 模板语法。

### 3.2. 常用的消息构建块

- **`schema.SystemMessage(text)`**: 创建系统角色的消息。
- **`schema.UserMessage(text)`**: 创建用户角色的消息。
- **`schema.AssistantMessage(text)`**: 创建助手角色的消息。
- **`schema.MessagesPlaceholder(key, optional)`**: **非常重要**。这是一个占位符，用于在运行时将一个 `[]*schema.Message` 类型的消息列表（通常是对话历史）插入到模板的指定位置。`key` 对应变量 map 中的键。

### 3.3. 示例代码

```go
import (
    "github.com/cloudwego/eino/components/prompt"
    "github.com/cloudwego/eino/schema"
)

// 创建一个模板
template := prompt.FromMessages(schema.FString,
    schema.SystemMessage("你是一个{role}。"),
    schema.MessagesPlaceholder("history_key", false), // 历史消息占位符
    schema.UserMessage("请帮我{task}。"),
)

// 准备变量
variables := map[string]any{
    "role": "专业的诗人",
    "task": "写一首关于海洋的诗",
    "history_key": []*schema.Message{
        {Role: schema.User, Content: "你好"},
        {Role: schema.Assistant, Content: "你好，有什么可以帮忙的吗？"},
    },
}

// 格式化模板
messages, err := template.Format(context.Background(), variables)
```

---

## 4. 最佳实践：在编排中使用

虽然可以单独使用 `ChatTemplate`，但官方**强烈推荐**将其作为编排工作流的一部分，与 `ChatModel` 等其他组件结合使用。这使得代码更具声明性、更健壮。

Eino 提供了 `compose` 包来实现编排，最常用的是 `compose.Chain`。

### 编排流程

1.  **创建 `Chain`**: 定义链的初始输入类型和最终输出类型。
2.  **附加 `ChatTemplate`**: 使用 `chain.AppendChatTemplate(template)` 将模板添加到链的第一步。
3.  **附加 `ChatModel`**: 使用 `chain.AppendChatModel(model)` 将模型添加到链的第二步。
4.  **编译 `Chain`**: 使用 `chain.Compile(ctx)` 将链编译成一个可运行的实例 (`Runnable`)。
5.  **调用 `Runnable`**: 使用 `runnable.Invoke(ctx, variables)` 来执行整个工作流。

### 示例代码

```go
import "github.com/cloudwego/eino/compose"

// 1. 创建 Chain
chain := compose.NewChain[map[string]any, *schema.Message]()

// 2. 附加组件
chain.AppendChatTemplate(template)
chain.AppendChatModel(model)

// 3. 编译并运行
runnable, err := chain.Compile(ctx)
if err != nil { /* ... */ }

finalAnswer, err := runnable.Invoke(ctx, variables)
if err != nil { /* ... */ }

fmt.Println(finalAnswer.Content)
```

通过这种方式，开发者无需手动管理组件之间的数据流，`Chain` 会自动将上一步的输出作为下一步的输入，极大地简化了代码。