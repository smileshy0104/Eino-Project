# Eino 框架 Tool 组件使用总结

## 1. 概述

Eino 框架中的 `Tool` 是一个核心组件，它允许将任意 Go 函数或逻辑封装成可被大型语言模型（LLM）理解和调用的标准化工具。通过将业务逻辑工具化，可以极大地扩展 LLM 的能力，使其能够与外部世界交互、执行计算、查询数据等。

本文档基于 `tool_demo` 目录下的示例代码，全面总结了在 Eino 中创建和使用 `Tool` 的几种主要方式，从完全手动的原生实现到高度自动化的便捷方法，再到流式处理和多工具管理等高级用法。

---

## 2. Tool 的创建方式

Eino 提供了多种创建 `Tool` 的方式，以适应不同的开发需求和复杂度，核心目标是尽可能地减少模板代码，让开发者专注于业务逻辑。

### 方式一：基础实现 (手动实现接口)

这是最原生、最灵活的方式，需要开发者手动为结构体实现 `InvokableTool` 接口中的 `Info()` 和 `InvokableRun()` 两个方法。

-   **`Info(...)`**: 返回工具的元数据（名称、描述、参数定义等）。这些信息将被用于生成供 LLM 理解的工具 schema。
-   **`InvokableRun(...)`**: 工具的核心执行逻辑。它接收一个 JSON 格式的字符串作为参数，执行完毕后也必须返回一个 JSON 格式的字符串作为结果。开发者需要在此方法内手动处理 JSON 的反序列化和序列化。

**适用场景**:
- 需要对工具的定义和执行过程进行最精细的控制。
- 工具的输入输出逻辑非常复杂，难以用简单的结构体描述。

**示例**: [`basic_tool/basic_tool.go`](tool_demo/basic_tool/basic_tool.go)

### 方式二：使用 `utils.NewTool` 包装函数

`utils.NewTool` 是一个辅助函数，它在手动定义元数据和业务逻辑函数之间取得了平衡。开发者只需：
1.  手动创建一个 `schema.ToolInfo` 实例来定义工具的元数据。
2.  编写一个符合 `func(context.Context, *Request) (*Response, error)` 签名的业务逻辑函数。
3.  将两者传入 `utils.NewTool` 即可获得一个完整的 `Tool` 实例。

`NewTool` 内部会自动处理 JSON 的序列化和反序列化。

**适用场景**:
- 业务逻辑已经存在于一个独立的函数中。
- 相比 `InferTool`，希望更明确地手动控制工具的元数据定义。

**示例**: [`newtool_example/newtool_example.go`](tool_demo/newtool_example/newtool_example.go)

### 方式三：使用 `utils.InferTool` (推荐)

这是**最简洁、最高效**的工具创建方式。`InferTool` 利用 Go 的反射机制，自动从业务函数的输入输出结构体中推断出工具的 schema。

开发者只需：
1.  定义输入（Request）和输出（Response）的 Go 结构体，并使用 `jsonschema` 标签来描述字段的约束（如 `required`, `description`, `enum`, `minimum` 等）。
2.  编写一个符合 `func(context.Context, *Request) (*Response, error)` 签名的业务逻辑函数。
3.  将工具名称、描述和业务函数传入 `utils.InferTool`。

`InferTool` 会自动生成 `ToolInfo` 并处理所有 JSON 操作，极大地减少了模板代码。

**适用场景**:
- 绝大多数场景，特别是新建工具时。
- 追求开发效率，希望代码更简洁、更易于维护。

**示例**: [`infertool_example/infertool_example.go`](tool_demo/infertool_example/infertool_example.go)

---

## 3. 高级用法

### 流式工具 (StreamableTool)

对于需要长时间运行或逐步产生结果的任务（如流式生成文本、处理大量数据），可以实现 `StreamableTool` 接口。

-   **`StreamableRun(...)`**: 该方法是流式工具的核心。它不直接返回最终结果，而是返回一个 `*schema.StreamReader[string]`。调用方可以从这个 `StreamReader` 中以数据块（chunk）的形式逐步接收结果。
-   **`InvokableRun(...)`**: 通常也需要实现。其标准做法是调用 `StreamableRun`，然后完整地读取 `StreamReader` 中的所有数据块，最后将它们拼接成一个完整的结果返回。

**适用场景**:
- 与 LLM 进行流式交互。
- 监控耗时任务的实时进度。
- 处理无法一次性载入内存的大数据集。

**示例**: [`streamable_tool/streamable_tool.go`](tool_demo/streamable_tool/streamable_tool.go)

### 多工具管理 (ToolsNode)

当应用中包含多个工具时，需要一个组件来统一管理和调度它们。`ToolsNode` 正是为此设计的。

-   **功能**: `ToolsNode` 可以注册一个或多个 `Tool` 实例。它接收一个包含 `ToolCall` 请求的 `schema.Message`，然后根据 `ToolCall` 中的函数名自动路由到对应的工具去执行。
-   **并行执行**: 如果一个 `Message` 中包含多个 `ToolCall`，`ToolsNode` 能够**并行**执行它们，并将所有工具的执行结果汇总后返回。

**适用场景**:
- 构建复杂的 Agent 或 Chain，需要根据 LLM 的决策动态调用不同的工具。
- 需要在一个步骤中执行多个独立的工具调用。

**示例**: [`toolsnode_example/toolsnode_example.go`](tool_demo/toolsnode_example/toolsnode_example.go)

---

## 4. 总结

| 创建方式 | 优点 | 缺点 | 适用场景 |
| :--- | :--- | :--- | :--- |
| **基础实现** | 灵活性最高，完全控制 | 代码冗长，需手动处理JSON | 复杂或非标准化的工具 |
| **`utils.NewTool`** | 平衡了灵活性和简洁性 | 仍需手动定义`ToolInfo` | 包装已有的业务函数 |
| **`utils.InferTool`** | **开发效率极高，代码最简洁** | 依赖反射和结构体标签 | **绝大多数常规场景 (推荐)** |
| **`StreamableTool`** | 支持流式输出，处理长任务 | 实现相对复杂 | 耗时任务、流式交互 |
| **`ToolsNode`** | 统一管理、并行调度多工具 | - | 构建多工具 Agent 或 Chain |

通过组合使用这些方式，开发者可以高效地为 Eino 应用构建功能强大且易于维护的工具集。